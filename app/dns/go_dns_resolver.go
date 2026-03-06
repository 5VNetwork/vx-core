// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package dns

import (
	"context"

	"github.com/5vnetwork/vx-core/common/dispatcher"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/proxy/freedom"
	"github.com/5vnetwork/vx-core/transport"
	"github.com/rs/zerolog/log"
)

type DnsResolver struct {
	*net.Resolver
}

func NewGoIpResolver() *DnsResolver {
	return &DnsResolver{}
}

func (d *DnsResolver) LookupIPSpeed(ctx context.Context, host string) ([]net.IP, error) {
	return d.LookupIP(ctx, host)
}

func (d *DnsResolver) LookupIP(ctx context.Context, host string) ([]net.IP, error) {
	ips, err := d.Resolver.LookupIP(ctx, "ip", host)
	if err != nil {
		return nil, err
	}
	log.Ctx(ctx).Debug().Str("host", host).Int("ips", len(ips)).Msg("lookup ip")
	return ips, nil
}

func (d *DnsResolver) LookupIPv4(ctx context.Context, host string) ([]net.IP, error) {
	ips, err := d.Resolver.LookupIP(ctx, "ip4", host)
	if err != nil {
		return nil, err
	}
	log.Ctx(ctx).Debug().Str("host", host).Int("ips", len(ips)).Msg("lookup ip")
	return ips, nil
}

func (d *DnsResolver) LookupIPv6(ctx context.Context, host string) ([]net.IP, error) {
	ips, err := d.Resolver.LookupIP(ctx, "ip6", host)
	if err != nil {
		return nil, err
	}
	log.Ctx(ctx).Debug().Str("host", host).Int("ips", len(ips)).Msg("lookup ip")
	return ips, nil
}

func (d *DnsResolver) LookupIPPrefer4(ctx context.Context, host string) ([]net.IP, error) {
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
	return NewDnsServerToResolver(r)
}
