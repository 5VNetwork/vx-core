// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package dns

import (
	"context"
	"errors"
	sync "sync"
	"time"

	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/i"
	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
)

type DnsServerToResolver struct {
	DnsServers []DnsServer
	Interval   time.Duration
}

type DnsServerToResolverOption struct {
	DnsServers []DnsServer
	Interval   time.Duration
}

func NewDnsServerToResolver(opts DnsServerToResolverOption) *DnsServerToResolver {
	if opts.Interval == 0 {
		opts.Interval = time.Second * 4
	}
	return &DnsServerToResolver{
		DnsServers: opts.DnsServers,
		Interval:   opts.Interval,
	}
}

func (d *DnsServerToResolver) LookupIP(ctx context.Context, host string) ([]net.IP, error) {
	var ipv6s []net.IP
	wait := sync.WaitGroup{}

	wait.Go(func() {
		var err error
		ipv6s, err = d.LookupIPv6(ctx, host)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("LookupIPv6 failed")
			return
		}
	})

	var ipv4s []net.IP
	var err error
	ipv4s, err = d.LookupIPv4(ctx, host)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("LookupIPv4 failed")
	}

	wait.Wait()
	log.Ctx(ctx).Debug().Str("host", host).Int("ipv4s", len(ipv4s)).Int("ipv6s", len(ipv6s)).Msg("lookup ip")
	return append(ipv4s, ipv6s...), nil
}

func (d *DnsServerToResolver) LookupIPSpeed(ctx context.Context, host string) ([]net.IP, error) {
	return d.LookupIP(ctx, host)
}

func (d *DnsServerToResolver) LookupIPPrefer4(ctx context.Context, host string) ([]net.IP, error) {
	ips, err := d.LookupIPv4(ctx, host)
	if err != nil {
		return nil, err
	}
	if len(ips) > 0 {
		return ips, nil
	}
	return d.LookupIPv6(ctx, host)
}

func (d *DnsServerToResolver) LookupIPv4(ctx context.Context, host string) ([]net.IP, error) {
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(host), dns.TypeA)

	ctx, cancel := context.WithTimeout(ctx, time.Second*4*time.Duration(len(d.DnsServers)))
	defer cancel()

	if len(d.DnsServers) == 1 {
		ips, err := d.lookupIPv4FromServer(ctx, d.DnsServers[0], msg)
		if err != nil {
			return nil, err
		}
		if len(ips) == 0 {
			return nil, ErrAllServersFailed
		}
		return ips, nil
	}

	resultChan := make(chan resolveIPResult, len(d.DnsServers))
	timer := time.NewTimer(d.Interval)
	defer timer.Stop()
nextServer:
	for i, dnsServer := range d.DnsServers {
		newMsg := msg
		if i != len(d.DnsServers)-1 {
			newMsg = msg.Copy()
		}
		go d.resolveIP(ctx, newMsg, i, dnsServer, resultChan, true)
		for {
			select {
			// dns server failed
			case <-timer.C:
				timer.Reset(d.Interval)
				continue nextServer
			case result := <-resultChan:
				if len(result.ips) > 0 {
					return result.ips, nil
				}
				// this server failed, break the loop to next server
				if result.index == i {
					continue nextServer
				}
				// previous server failed, continue to wait for result of this server
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}
	return nil, ErrAllServersFailed
}

type resolveIPResult struct {
	index int
	ips   []net.IP
}

func (d *DnsServerToResolver) resolveIP(ctx context.Context, msg *dns.Msg, index int,
	dnsServer DnsServer, resultChan chan resolveIPResult, ipv4 bool) {
	var ips []net.IP
	var err error
	if ipv4 {
		ips, err = d.lookupIPv4FromServer(ctx, dnsServer, msg)
	} else {
		ips, err = d.lookupIPv6FromServer(ctx, dnsServer, msg)
	}
	if (err != nil && !errors.Is(err, context.Canceled)) || len(ips) == 0 {
		log.Ctx(ctx).Debug().Err(err).Msg("one dns server lookup failed")
	}
	resultChan <- resolveIPResult{index: index, ips: ips}
}

func findNextCNAMETarget(answers []dns.RR, qName string) (string, bool) {
	for _, rr := range answers {
		cname, ok := rr.(*dns.CNAME)
		if !ok {
			continue
		}
		if !dns.IsFqdn(cname.Hdr.Name) || !dns.IsFqdn(cname.Target) {
			continue
		}
		if cname.Hdr.Name != qName {
			continue
		}
		return cname.Target, true
	}
	return "", false
}

func findCNAMEChainTail(answers []dns.RR, start string, maxHops int) string {
	current := dns.Fqdn(start)
	if current == "." {
		return ""
	}
	for range maxHops {
		next, ok := findNextCNAMETarget(answers, current)
		if !ok {
			return current
		}
		if next == current {
			return current
		}
		current = next
	}
	return current
}

// parseAnswerIPs extracts IPs from DNS answer RRs and CNAME chain tail.
// ipv4=true collects A records; ipv4=false collects AAAA records.
func parseAnswerIPs(answers []dns.RR, ipv4 bool, qName string) ([]net.IP, string) {
	var ips []net.IP
	hasCNAME := false
	for _, rr := range answers {
		switch v := rr.(type) {
		case *dns.A:
			if ipv4 {
				ips = append(ips, net.IP(v.A))
			}
		case *dns.AAAA:
			if !ipv4 {
				ips = append(ips, net.IP(v.AAAA))
			}
		case *dns.CNAME:
			hasCNAME = true
		}
	}
	if len(ips) == 0 && hasCNAME {
		tail := findCNAMEChainTail(answers, qName, 8)
		log.Warn().Str("qname", dns.Fqdn(qName)).Str("cname_tail", tail).Msg("no ips found but cname present")
		return nil, tail
	}
	return ips, ""
}

// fetchResponse queries dnsServer over UDP, falling back to TCP when the response
// is truncated and no usable records were found in the partial UDP reply.
func (d *DnsServerToResolver) fetchResponse(ctx context.Context, dnsServer DnsServer, msg *dns.Msg, ipv4 bool) ([]net.IP, error) {
	resp, err := dnsServer.HandleQuery(ctx, msg, false)
	if err != nil || resp == nil {
		return nil, err
	}
	ips, cname := parseAnswerIPs(resp.Answer, ipv4, msg.Question[0].Name)
	if len(ips) > 0 {
		return ips, nil
	}
	if cname != "" {
		if ipv4 {
			return d.LookupIPv4(ctx, cname)
		} else {
			return d.LookupIPv6(ctx, cname)
		}
	}
	if resp.Truncated {
		log.Ctx(ctx).Debug().Any("resp", resp).Msg("ip resolver truncated response")
		resp, err = dnsServer.HandleQuery(ctx, msg, true)
		if err != nil || resp == nil {
			return nil, err
		}
		ips, cname = parseAnswerIPs(resp.Answer, ipv4, msg.Question[0].Name)
		if len(ips) > 0 {
			return ips, nil
		}
		if cname != "" {
			if ipv4 {
				return d.LookupIPv4(ctx, cname)
			} else {
				return d.LookupIPv6(ctx, cname)
			}
		}
	}
	return nil, nil
}

func (d *DnsServerToResolver) lookupIPv4FromServer(ctx context.Context, dnsServer DnsServer, msg *dns.Msg) ([]net.IP, error) {
	if ipResolver, ok := dnsServer.(i.IPResolver); ok {
		return ipResolver.LookupIPv4(ctx, UnFqdn(msg.Question[0].Name))
	}
	return d.fetchResponse(ctx, dnsServer, msg, true)
}

func (d *DnsServerToResolver) LookupIPv6(ctx context.Context, host string) ([]net.IP, error) {
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(host), dns.TypeAAAA)

	ctx, cancel := context.WithTimeout(ctx, time.Second*4*time.Duration(len(d.DnsServers)))
	defer cancel()

	if len(d.DnsServers) == 1 {
		ips, err := d.lookupIPv6FromServer(ctx, d.DnsServers[0], msg)
		if err != nil {
			return nil, err
		}
		if len(ips) == 0 {
			return nil, ErrAllServersFailed
		}
		return ips, nil
	}

	resultChan := make(chan resolveIPResult, len(d.DnsServers))
	timer := time.NewTimer(d.Interval)
	defer timer.Stop()
nextServer:
	for i, dnsServer := range d.DnsServers {
		newMsg := msg
		if i != len(d.DnsServers)-1 {
			newMsg = msg.Copy()
		}
		go d.resolveIP(ctx, newMsg, i, dnsServer, resultChan, false)
		for {
			select {
			case <-timer.C:
				timer.Reset(d.Interval)
				continue nextServer
			case result := <-resultChan:
				if len(result.ips) > 0 {
					return result.ips, nil
				}
				// this server failed, break the loop to next server
				if result.index == i {
					continue nextServer
				}
				// previous server failed, continue to wait for result of this server
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}

	return nil, ErrAllServersFailed
}

func (d *DnsServerToResolver) lookupIPv6FromServer(ctx context.Context, dnsServer DnsServer, msg *dns.Msg) ([]net.IP, error) {
	if ipResolver, ok := dnsServer.(i.IPResolver); ok {
		return ipResolver.LookupIPv6(ctx, UnFqdn(msg.Question[0].Name))
	}
	return d.fetchResponse(ctx, dnsServer, msg, false)
}

func (d *DnsServerToResolver) LookupECH(ctx context.Context, domain string) ([]byte, error) {
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(domain), dns.TypeHTTPS)

	ctx, cancel := context.WithTimeout(ctx, time.Second*4*time.Duration(len(d.DnsServers)))
	defer cancel()

	if len(d.DnsServers) == 1 {
		ech, err := d.lookupECHFromServer(ctx, d.DnsServers[0], msg, domain)
		if err != nil {
			return nil, err
		}
		if len(ech) == 0 {
			return nil, ErrAllServersFailed
		}
		return ech, nil
	}

	resultChan := make(chan resolveECHResult, len(d.DnsServers))
	timer := time.NewTimer(d.Interval)
	defer timer.Stop()
nextServer:
	for i, dnsServer := range d.DnsServers {
		go d.resolveECH(ctx, msg, domain, i, dnsServer, resultChan)
		for {
			select {
			case <-timer.C:
				timer.Reset(d.Interval)
				continue nextServer
			case result := <-resultChan:
				if len(result.ech) > 0 {
					return result.ech, nil
				}
				// this server failed, break the loop to next server
				if result.index == i {
					continue nextServer
				}
				// previous server failed, continue to wait for result of this server
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}

	return nil, ErrAllServersFailed
}

type resolveECHResult struct {
	index int
	ech   []byte
}

func (d *DnsServerToResolver) resolveECH(ctx context.Context, msg *dns.Msg, domain string, index int,
	dnsServer DnsServer, resultChan chan resolveECHResult) {
	ech, err := d.lookupECHFromServer(ctx, dnsServer, msg, domain)
	if err != nil || len(ech) == 0 {
		log.Ctx(ctx).Debug().Err(err).Msg("one dns server lookup ech failed")
	}
	resultChan <- resolveECHResult{index: index, ech: ech}
}

func (d *DnsServerToResolver) lookupECHFromServer(ctx context.Context, dnsServer DnsServer, msg *dns.Msg, domain string) ([]byte, error) {
	resp, err := dnsServer.HandleQuery(ctx, msg, false)
	if err != nil {
		return nil, err
	}
	fqdn := dns.Fqdn(domain)
	for _, answer := range resp.Answer {
		https, ok := answer.(*dns.HTTPS)
		if !ok || https.Hdr.Name != fqdn {
			continue
		}
		for _, v := range https.Value {
			if echConfig, ok := v.(*dns.SVCBECHConfig); ok {
				return echConfig.ECH, nil
			}
		}
	}
	return nil, errors.New("no ech found in response")
}
