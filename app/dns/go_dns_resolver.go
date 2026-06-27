// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package dns

import (
	"context"
	"errors"

	"github.com/5vnetwork/vx-core/common/dispatcher"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/proxy/freedom"
	"github.com/5vnetwork/vx-core/transport"
	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
)

type GoDnsResolver struct {
	*net.Resolver
	cache      *rrCache
	ipToDomain *IPToDomain
	rewriter   MsgRewriter
	name       string
}

type GoDnsResolverOption struct {
	Cache      *rrCache
	IpToDomain *IPToDomain
	Rewriter   MsgRewriter
	Name       string
}

func NewGoIpResolver(option GoDnsResolverOption) *GoDnsResolver {
	return &GoDnsResolver{
		cache:      option.Cache,
		ipToDomain: option.IpToDomain,
		rewriter:   option.Rewriter,
		name:       option.Name,
	}
}

func (d *GoDnsResolver) Name() string {
	return d.name
}

func (d *GoDnsResolver) Start() error {
	if d.cache != nil {
		d.cache.Start()
	}
	return nil
}

func (d *GoDnsResolver) Close() error {
	if d.cache != nil {
		d.cache.Close()
	}
	return nil
}

func (d *GoDnsResolver) HandleQuery(ctx context.Context, msg *dns.Msg, tcp bool) (*dns.Msg, error) {
	if len(msg.Question) == 0 {
		return nil, ErrNoQuestion
	}

	question := msg.Question[0]
	if question.Qtype != dns.TypeA && question.Qtype != dns.TypeAAAA {
		return nil, errors.New("only A and AAAA queries are supported")
	}

	if d.cache != nil {
		reply, ok := d.cache.Get(&question)
		if ok {
			return reply, nil
		}
	}

	var (
		ips []net.IP
		err error
	)
	switch question.Qtype {
	case dns.TypeA:
		ips, err = d.LookupIPv4(ctx, UnFqdn(question.Name))
	case dns.TypeAAAA:
		ips, err = d.LookupIPv6(ctx, UnFqdn(question.Name))
	}
	if err != nil {
		return nil, err
	}

	resp := new(dns.Msg)
	resp.SetReply(msg)
	resp.RecursionAvailable = true
	for _, ip := range ips {
		switch question.Qtype {
		case dns.TypeA:
			ip4 := ip.To4()
			if ip4 == nil {
				continue
			}
			resp.Answer = append(resp.Answer, &dns.A{
				Hdr: dns.RR_Header{
					Name:   question.Name,
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    60,
				},
				A: ip4,
			})
		case dns.TypeAAAA:
			ip6 := ip.To16()
			if ip6 == nil || ip.To4() != nil {
				continue
			}
			resp.Answer = append(resp.Answer, &dns.AAAA{
				Hdr: dns.RR_Header{
					Name:   question.Name,
					Rrtype: dns.TypeAAAA,
					Class:  dns.ClassINET,
					Ttl:    60,
				},
				AAAA: ip6,
			})
		}
	}

	if d.rewriter != nil {
		d.rewriter.Rewrite(resp)
	}
	if d.ipToDomain != nil && len(resp.Answer) > 0 {
		d.ipToDomain.SetDomain(resp, nil)
	}
	if d.cache != nil {
		d.cache.Set(resp)
	}

	log.Ctx(ctx).Debug().Str("host", UnFqdn(question.Name)).Int("answers", len(resp.Answer)).Msg("go dns resolver replied")
	return resp, nil

}

func (d *GoDnsResolver) LookupIPSpeed(ctx context.Context, host string) ([]net.IP, error) {
	return d.LookupIP(ctx, host)
}

func (d *GoDnsResolver) LookupIP(ctx context.Context, host string) ([]net.IP, error) {
	ips, err := d.Resolver.LookupIP(ctx, "ip", host)
	if err != nil {
		return nil, err
	}
	log.Ctx(ctx).Debug().Str("host", host).Int("ips", len(ips)).Msg("lookup ip")
	return ips, nil
}

func (d *GoDnsResolver) LookupIPv4(ctx context.Context, host string) ([]net.IP, error) {
	ips, err := d.Resolver.LookupIP(ctx, "ip4", host)
	if err != nil {
		return nil, err
	}
	log.Ctx(ctx).Debug().Str("host", host).Int("ips", len(ips)).Msg("lookup ip")
	return ips, nil
}

func (d *GoDnsResolver) LookupIPv6(ctx context.Context, host string) ([]net.IP, error) {
	ips, err := d.Resolver.LookupIP(ctx, "ip6", host)
	if err != nil {
		return nil, err
	}
	log.Ctx(ctx).Debug().Str("host", host).Int("ips", len(ips)).Msg("lookup ip")
	return ips, nil
}

func (d *GoDnsResolver) LookupIPPrefer4(ctx context.Context, host string) ([]net.IP, error) {
	ips, err := d.Resolver.LookupIP(ctx, "ip4", host)
	if err != nil {
		return nil, err
	}
	if len(ips) > 0 {
		return ips, nil
	}
	return d.Resolver.LookupIP(ctx, "ip6", host)
}

func DefaultCfResolver() i.DnsResolver {
	freedom := freedom.New(transport.DefaultDialer, transport.DefaultPacketListener, "freedom", nil)
	r := NewDnsServerConcurrent(
		DnsServerConcurrentOption{
			Name: "default_cf_resolver",
			NameserverAddrs: []net.AddressPort{
				{
					Address: net.CfDns4,
					Port:    53,
				},
			},
			Handler: freedom,
			Dispatcher: dispatcher.NewPacketDispatcher(
				context.Background(), freedom),
		},
	)
	return NewDnsServerToResolver(DnsServerToResolverOption{DnsServers: []DnsServer{r}})
}
