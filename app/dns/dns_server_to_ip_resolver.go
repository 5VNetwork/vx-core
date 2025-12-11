package dns

import (
	"context"
	sync "sync"
	"time"

	"github.com/5vnetwork/vx-core/common/net"
	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
)

type DnsServerToResolver struct {
	DnsServers []DnsServer
}

func NewDnsServerToResolver(dnsServers ...DnsServer) *DnsServerToResolver {
	return &DnsServerToResolver{
		// IPOption:   ipOption,
		DnsServers: dnsServers,
	}
}

func (d *DnsServerToResolver) LookupIP(ctx context.Context, host string) ([]net.IP, error) {
	var ipv6s []net.IP
	wait := sync.WaitGroup{}

	wait.Add(1)
	go func() {
		defer wait.Done()
		var err error
		ipv6s, err = d.LookupIPv6(ctx, host)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("LookupIPv6 failed")
			return
		}
	}()

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

	for _, dnsServer := range d.DnsServers {
		subCtx, cancel := context.WithTimeout(ctx, time.Second*4)
		defer cancel()
		resp, err := dnsServer.HandleQuery(subCtx, msg, false)
		if err != nil {
			log.Ctx(subCtx).Error().Err(err).Msg("one dns server lookup ipv4 failed")
		}
		if resp != nil {
			hasCname := false
			ips := make([]net.IP, 0, len(resp.Answer))
			for _, answer := range resp.Answer {
				if a, ok := answer.(*dns.A); ok {
					ips = append(ips, net.IP(a.A))
				} else if _, ok := answer.(*dns.CNAME); ok {
					// if answer is a cname, resolve the cname
					hasCname = true
				}
			}
			if len(ips) == 0 && hasCname {
				for _, answer := range resp.Answer {
					if cname, ok := answer.(*dns.CNAME); ok {
						return d.LookupIPv4(subCtx, cname.Target)
					}
				}
			}
			if resp.Truncated {
				log.Ctx(subCtx).Warn().Any("resp", resp).Msg("ip resolver truncated response")
			}
			return ips, nil
		}
	}

	return nil, ErrAllServersFailed
}

func (d *DnsServerToResolver) LookupIPv6(ctx context.Context, host string) ([]net.IP, error) {
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(host), dns.TypeAAAA)

	for _, dnsServer := range d.DnsServers {
		subCtx, cancel := context.WithTimeout(ctx, time.Second*4)
		defer cancel()
		resp, err := dnsServer.HandleQuery(subCtx, msg, false)
		if err != nil {
			log.Ctx(subCtx).Debug().Err(err).Msg("lookup ipv6 failed")
		}
		if resp != nil {
			hasCname := false
			ips := make([]net.IP, 0, len(resp.Answer))
			for _, answer := range resp.Answer {
				if a, ok := answer.(*dns.AAAA); ok {
					ips = append(ips, net.IP(a.AAAA))
				} else if _, ok := answer.(*dns.CNAME); ok {
					// if answer is a cname, resolve the cname
					hasCname = true
				}
			}
			if len(ips) == 0 && hasCname {
				for _, answer := range resp.Answer {
					if cname, ok := answer.(*dns.CNAME); ok {
						return d.LookupIPv6(subCtx, cname.Target)
					}
				}
			}
			if resp.Truncated {
				log.Ctx(subCtx).Warn().Any("resp", resp).Msg("ip resolver truncated response")
			}
			return ips, nil
		}
	}

	return nil, ErrAllServersFailed
}

func (d *DnsServerToResolver) LookupECH(ctx context.Context, domain string) ([]byte, error) {
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(domain), dns.TypeHTTPS)
	for _, dnsServer := range d.DnsServers {
		resp, err := dnsServer.HandleQuery(ctx, msg, false)
		if err != nil {
			return nil, err
		}
		if len(resp.Answer) > 0 {
			for _, answer := range resp.Answer {
				if https, ok := answer.(*dns.HTTPS); ok && https.Hdr.Name == dns.Fqdn(domain) {
					for _, v := range https.Value {
						if echConfig, ok := v.(*dns.SVCBECHConfig); ok {
							return echConfig.ECH, nil
						}
					}
				}
			}
		}
	}
	return nil, ErrAllServersFailed
}
