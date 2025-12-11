package dns

import (
	"context"
	"net"

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
