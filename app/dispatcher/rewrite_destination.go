// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package dispatcher

import (
	"context"

	"github.com/5vnetwork/vx-core/app/sniff"
	mynet "github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/common/session"
	"github.com/5vnetwork/vx-core/common/strmatcher"
	"github.com/5vnetwork/vx-core/i"
	"github.com/rs/zerolog/log"
)

type RewriteDestinationHook struct {
	Sniff               bool
	Sniffer             *sniff.Sniffer
	DestinationOverride []string
	FakeDns             i.FakeDnsPool
	Dns                 i.IPResolver
}

func (p *RewriteDestinationHook) BeforeHandlerSelection(ctx context.Context, si *session.Info,
	rw any) (context.Context, any, error) {
	var fakeButNotFound bool
	// change fake ip to real domain
	fd := p.FakeDns
	if fd != nil && si.Target.IsValid() && si.Target.Address.Family().IsIP() &&
		fd.IsIPInIPPool(si.Target.Address) {
		if s := fd.GetDomainFromFakeDNS(si.Target.Address); s != "" {
			log.Ctx(ctx).Debug().Str("domain", s).Msg("fake ip found")
			domain, err := strmatcher.ToDomain(s)
			if err != nil {
				log.Ctx(ctx).Error().Err(err).Str("domain", s).Msg("failed to convert to domain")
				domain = s
			}
			si.FakeIP = si.Target.Address.IP()
			si.Target = mynet.Destination{
				Address: mynet.ParseAddress(domain),
				Port:    si.Target.Port,
				Network: si.Target.Network,
			}
		} else {
			log.Ctx(ctx).Debug().IPAddr("ip", si.Target.Address.IP()).Msg("fake ip but domain not found")
			fakeButNotFound = true
		}
	}

	shouldSniff := !si.Sniffed && ((si.Target.IsValid() && si.Target.Address.Family().IsIP() && len(p.DestinationOverride) > 0) ||
		fakeButNotFound || (p.Sniff && si.Target.IsValid() && si.Target.Address.Family().IsIP())) && si.Target.Port != 53
	// sniff
	if shouldSniff {
		rw0, err := p.Sniffer.Sniff(ctx, si, rw)
		if err == nil && si.Protocol != "" {
			log.Ctx(ctx).Debug().Str("protocol", si.Protocol).Str("domain", si.SniffedDomain).Msg("sniff result")
		} else {
			log.Ctx(ctx).Debug().Err(err).Msg("sniff failed")
		}
		rw = rw0
	}
	if shouldOverride(si, p.DestinationOverride, fakeButNotFound) {
		log.Ctx(ctx).Debug().Str("original dest", si.Target.String()).
			Str("sniff domain", si.SniffedDomain).Msg("replace destination")
		si.Target.Address = mynet.ParseAddress(si.SniffedDomain)
	}

	if pc, ok := rw.(udp.PacketReaderWriter); ok && fd != nil && p.Dns != nil {
		var addressFamily mynet.AddressFamily
		if si.Source.Address != nil {
			addressFamily = si.Source.Address.Family()
		} else {
			addressFamily = mynet.AddressFamilyIPv4
		}
		rw = &RealIpPacketConn{
			m:                  map[mynet.Address]mynet.Address{},
			PacketReaderWriter: pc,
			fakeDns:            fd,
			dns:                p.Dns,
			ctx:                ctx,
			addressFamily:      addressFamily,
		}
	}
	return ctx, rw, nil
}

type RewriteIPv6ToDomainHook struct{}

func (p *RewriteIPv6ToDomainHook) AfterHandlerSelection(ctx context.Context, info *session.Info, rw any,
	handler i.Outbound) (context.Context, any, error) {
	if info.Target.Address.Family().IsIP() && info.Target.Address.Family().IsIPv6() {
		if handlerSupport6, ok := handler.(i.HandlerWith6Info); ok && !handlerSupport6.Support6() && info.SniffedDomain != "" {
			log.Ctx(ctx).Debug().Str("handler", handler.Tag()).Str("dst", info.Target.String()).Msg("ipv6 not supported, replace it with the domain")
			info.Target.Address = mynet.DomainAddress(info.SniffedDomain)
		}
	}
	return ctx, rw, nil
}
