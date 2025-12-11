package domain

import (
	"context"
	"math/rand"
	"net"
	"time"

	"github.com/5vnetwork/vx-core/i"
	"github.com/rs/zerolog/log"
)

type DomainStrategy int32

const (
	DomainStrategy_PreferIPv4 DomainStrategy = iota
	DomainStrategy_PreferIPv6
	DomainStrategy_IPv4Only
	DomainStrategy_IPv6Only
)

func GetIPs(ctx context.Context, domain string,
	stategy DomainStrategy, ipr i.IPResolver) []net.IP {
	switch stategy {
	case DomainStrategy_PreferIPv4, DomainStrategy_PreferIPv6:
		now := time.Now()
		ips, err := ipr.LookupIP(ctx, domain)
		log.Ctx(ctx).Debug().Str("domain", domain).
			Dur("cost", time.Since(now)).
			Msg("DnsDialer look up domain")
		if err != nil {
			return nil
		}
		ipv4s := make([]net.IP, 0, len(ips))
		ipv6s := make([]net.IP, 0, len(ips))
		for _, ip := range ips {
			if ip.To4() != nil {
				ipv4s = append(ipv4s, ip)
			} else {
				ipv6s = append(ipv6s, ip)
			}
		}
		if stategy == DomainStrategy_PreferIPv4 {
			var ips []net.IP
			if len(ipv4s) > 0 {
				ips = append(ips, ipv4s[rand.Intn(len(ipv4s))])
			}
			if len(ipv6s) > 0 {
				ips = append(ips, ipv6s[rand.Intn(len(ipv6s))])
			}
			return ips
		} else {
			var ips []net.IP
			if len(ipv6s) > 0 {
				ips = append(ips, ipv6s[rand.Intn(len(ipv6s))])
			}
			if len(ipv4s) > 0 {
				ips = append(ips, ipv4s[rand.Intn(len(ipv4s))])
			}
			return ips
		}
	case DomainStrategy_IPv4Only:
		ips, err := ipr.LookupIPv4(ctx, domain)
		if err != nil {
			return nil
		}
		return ips
	case DomainStrategy_IPv6Only:
		ips, err := ipr.LookupIPv6(ctx, domain)
		if err != nil {
			return nil
		}
		return ips
	}
	return nil
}
