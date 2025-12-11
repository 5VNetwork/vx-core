package dns

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/i"
	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
)

// for dialer to lookup ip when dialing
type InternalDns struct {
	StaticDns             *StaticDnsServer
	DnsServers            []DnsServer
	DnsServerToIPResolver *DnsServerToResolver
}

func NewInternalDns(staticDns *StaticDnsServer, DnsServers ...DnsServer) *InternalDns {
	return &InternalDns{
		StaticDns:             staticDns,
		DnsServers:            DnsServers,
		DnsServerToIPResolver: NewDnsServerToResolver(DnsServers...),
	}
}

func (d *InternalDns) Start() error {
	return common.StartAll(d.DnsServers)
}

func (d *InternalDns) Close() error {
	return common.CloseAll(d.DnsServers)
}

func (d *InternalDns) LookupIPv4(ctx context.Context, host string) ([]net.IP, error) {
	if d.StaticDns != nil {
		msg := new(dns.Msg)
		msg.SetQuestion(dns.Fqdn(host), dns.TypeA)
		if resp, ok := d.StaticDns.ReplyFor(msg); ok {
			ips := make([]net.IP, 0, len(resp.Answer))
			for _, answer := range resp.Answer {
				if a, ok := answer.(*dns.A); ok {
					ips = append(ips, net.IP(a.A))
				}
			}

			return ips, nil
		}
	}
	return d.DnsServerToIPResolver.LookupIPv4(ctx, host)
}

func (d *InternalDns) LookupIPv6(ctx context.Context, host string) ([]net.IP, error) {
	if d.StaticDns != nil {
		msg := new(dns.Msg)
		msg.SetQuestion(dns.Fqdn(host), dns.TypeAAAA)
		if resp, ok := d.StaticDns.ReplyFor(msg); ok {
			ips := make([]net.IP, 0, len(resp.Answer))
			for _, answer := range resp.Answer {
				if a, ok := answer.(*dns.AAAA); ok {
					ips = append(ips, net.IP(a.AAAA))
				}
			}
			return ips, nil
		}
	}
	return d.DnsServerToIPResolver.LookupIPv6(ctx, host)
}

func (d *InternalDns) LookupIP(ctx context.Context, host string) ([]net.IP, error) {
	// return d.LookupIPSpeed(ctx, host)
	var ipv6s []net.IP
	wait := sync.WaitGroup{}

	start := time.Now()

	wait.Add(1)
	go func() {
		defer wait.Done()
		var err error
		ipv6s, err = d.LookupIPv6(ctx, host)
		if err != nil {
			log.Ctx(ctx).Debug().Err(err).Dur("time", time.Since(start)).Msg("LookupIPv6 failed")
			return
		}
		log.Ctx(ctx).Debug().Dur("elapsed", time.Since(start)).Int("ipv6s", len(ipv6s)).Msg("LookupIPv6 finished")
	}()
	// }

	var ipv4s []net.IP
	var err error
	ipv4s, err = d.LookupIPv4(ctx, host)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Dur("time", time.Since(start)).Msg("LookupIPv4 failed")
	} else {
		log.Ctx(ctx).Debug().Dur("elapsed", time.Since(start)).Int("ipv4s", len(ipv4s)).Msg("LookupIPv4 finished")
	}

	wait.Wait()
	return append(ipv4s, ipv6s...), nil
}

func (d *InternalDns) LookupIPSpeed(ctx context.Context, host string) ([]net.IP, error) {
	var ipv6s []net.IP
	c := make(chan []net.IP, 2)
	endC := make(chan struct{}, 2)
	go func() {
		var err error
		ipv6s, err = d.LookupIPv6(ctx, host)
		if err != nil {
			log.Ctx(ctx).Debug().Err(err).Msg("lookup ipv6 failed")
			endC <- struct{}{}
			return
		}
		if len(ipv6s) > 0 {
			log.Ctx(ctx).Debug().Str("host", host).Int("ipv6s", len(ipv6s)).Msg("lookup ipv6 success")
			c <- ipv6s
		} else {
			endC <- struct{}{}
		}
	}()
	go func() {
		var ipv4s []net.IP
		var err error
		ipv4s, err = d.LookupIPv4(ctx, host)
		if err != nil {
			log.Ctx(ctx).Debug().Err(err).Msg("lookup ipv4 failed")
			endC <- struct{}{}
			return
		}
		if len(ipv4s) > 0 {
			log.Ctx(ctx).Debug().Str("host", host).Int("ipv4s", len(ipv4s)).Msg("lookup ipv4 success")
			c <- ipv4s
		} else {
			endC <- struct{}{}
		}
	}()

	i := 0
	for {
		select {
		case <-endC:
			i++
			if i == 2 {
				return nil, errors.New("both A and AAAA lookup failed")
			}
		case ips := <-c:
			return ips, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

type Prefer4IPResolver struct {
	i.IPResolver
}

func (d *Prefer4IPResolver) LookupIP(ctx context.Context, host string) ([]net.IP, error) {
	ips, err := d.LookupIPv4(ctx, host)
	if err != nil {
		return nil, err
	}
	if len(ips) > 0 {
		return ips, nil
	}
	return d.LookupIPv6(ctx, host)
}
