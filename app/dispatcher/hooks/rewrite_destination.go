// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package hooks

import (
	"context"
	"strings"
	"sync"

	"github.com/5vnetwork/vx-core/app/sniff"
	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/5vnetwork/vx-core/common/net"
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
	ChangeDomainToIpv6  bool
	ChangeDomainToIpv4  bool
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

	if p.ChangeDomainToIpv6 && si.Target.Address.Family().IsDomain() {
		ips, err := p.Dns.LookupIPv6(ctx, si.Target.Address.Domain())
		if err == nil && len(ips) > 0 {
			si.Target.Address = mynet.IPAddress(ips[0])
		}
	} else if p.ChangeDomainToIpv4 && si.Target.Address.Family().IsDomain() {
		ips, err := p.Dns.LookupIPv4(ctx, si.Target.Address.Domain())
		if err == nil && len(ips) > 0 {
			si.Target.Address = mynet.IPAddress(ips[0])
		}
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

func shouldOverride(info *session.Info, domainOverride []string, fakeIPNotFound bool) bool {
	if info.SniffedDomain == "" {
		return false
	}
	if fakeIPNotFound {
		return true
	}
	protocolString := info.Protocol
	if protocolString == "" {
		return false
	}
	for _, p := range domainOverride {
		if strings.HasPrefix(protocolString, p) || strings.HasSuffix(protocolString, p) {
			return true
		}
	}
	return false
}

// change fake ip to real ip
type RealIpPacketConn struct {
	udp.PacketReaderWriter
	m              map[net.Address]net.Address // fake ip to real ip, real ip to real ip
	realIpToFakeIp sync.Map                    // real ip (net.Address) to fake ip (net.Address)
	domainToRealIp sync.Map                    // domain (net.Address) to real ip (net.Address)

	fakeDns i.FakeDnsPool
	dns     i.IPResolver
	ctx     context.Context
	// either ipv4 or ipv6
	addressFamily net.AddressFamily
}

// should be called sequentially
func (p *RealIpPacketConn) ReadPacket() (*udp.Packet, error) {
	packet, err := p.PacketReaderWriter.ReadPacket()
	if err != nil {
		return nil, err
	}
	originalTarget := packet.Target.Address
	if originalTarget.Family().IsIP() {
		if v, ok := p.m[originalTarget]; ok {
			packet.Target.Address = v
			return packet, nil
		}
		if p.fakeDns.IsIPInIPPool(originalTarget) {
			if d := p.fakeDns.GetDomainFromFakeDNS(originalTarget); d != "" {
				var ips []net.IP
				if p.addressFamily == net.AddressFamilyIPv4 {
					ips, err = p.dns.LookupIPv4(p.ctx, d)
				} else {
					ips, err = p.dns.LookupIPv6(p.ctx, d)
				}
				if err != nil {
					return nil, err
				}
				if len(ips) == 0 {
					return nil, errors.New("failed to find ip for a domain")
				}
				newTarget := net.IPAddress(ips[0])
				log.Ctx(p.ctx).Debug().IPAddr("original target", originalTarget.IP()).
					IPAddr("new target", newTarget.IP()).Msg("rewrite destination")
				packet.Target.Address = newTarget
				p.m[originalTarget] = newTarget
				p.realIpToFakeIp.Store(newTarget, originalTarget)
				if len(p.m) > 1024 {
					log.Ctx(p.ctx).Warn().Int("len", len(p.m)).Msg("rewrite destination cache grew beyond threshold")
				}
			} else {
				return nil, errors.New("failed to find domain for a fake ip")
			}
		} else {
			p.m[originalTarget] = originalTarget
		}
	}
	return packet, nil
}

func (p *RealIpPacketConn) WritePacket(packet *udp.Packet) error {
	if v, ok := p.realIpToFakeIp.Load(packet.Source.Address); ok {
		packet.Source.Address = v.(net.Address)
	} else if packet.Source.Address.Family().IsDomain() {
		if v, ok := p.domainToRealIp.Load(packet.Source.Address); ok {
			packet.Source.Address = v.(net.Address)
		} else {
			var ips []net.IP
			var err error
			if p.addressFamily == net.AddressFamilyIPv4 {
				ips, err = p.dns.LookupIPv4(p.ctx, packet.Source.Address.Domain())
			} else {
				ips, err = p.dns.LookupIPv6(p.ctx, packet.Source.Address.Domain())
			}
			if err != nil && len(ips) == 0 {
				packet.Release()
				log.Warn().Ctx(p.ctx).Err(err).
					Str("domain", packet.Source.Address.Domain()).
					Msg("failed to lookup ip for a domain")
				return nil
			}
			newSource := net.IPAddress(ips[0])
			p.domainToRealIp.Store(packet.Source.Address, newSource)
			packet.Source.Address = newSource
		}
	}
	return p.PacketReaderWriter.WritePacket(packet)
}
