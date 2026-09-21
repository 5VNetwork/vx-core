// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package dns

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/i"
	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
)

type AllDnsServers struct {
	lock       sync.RWMutex
	dnsServers []DnsServer
}

func NewAllDnsServers(dnsServers []DnsServer) *AllDnsServers {
	d := &AllDnsServers{
		dnsServers: dnsServers,
	}
	return d
}

func (dsp *AllDnsServers) Start() error {
	for _, client := range dsp.dnsServers {
		if err := client.Start(); err != nil {
			return err
		}
	}
	return nil
}

func (dsp *AllDnsServers) Close() error {
	for _, dnsServer := range dsp.dnsServers {
		if err := dnsServer.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (dsp *AllDnsServers) UpdateDnsServers(dnsServers []DnsServer) {
	for _, dnsServer := range dnsServers {
		if err := dnsServer.Start(); err != nil {
			log.Fatal().Err(err).Msg("failed to start dns server")
		}
	}
	dsp.lock.Lock()
	oldDnsServers := dsp.dnsServers
	dsp.dnsServers = dnsServers
	dsp.lock.Unlock()

	for _, dnsServer := range oldDnsServers {
		if err := dnsServer.Close(); err != nil {
			log.Warn().Err(err).Msg("failed to close dns server")
		}
	}
}

func (dsp *AllDnsServers) IsIPInIPPool(ip net.Address) bool {
	dsp.lock.RLock()
	defer dsp.lock.RUnlock()

	for _, dnsServer := range dsp.dnsServers {
		if isFakeDns(dnsServer) {
			if fakeDns, ok := dnsServer.(*FakeDns); ok {
				if fakeDns.IsIPInIPPool(ip) {
					return true
				}
			}
		}
	}
	return false
}

func (dsp *AllDnsServers) GetDomainFromFakeDNS(ip net.Address) string {
	dsp.lock.RLock()
	defer dsp.lock.RUnlock()

	for _, dnsServer := range dsp.dnsServers {
		if isFakeDns(dnsServer) {
			if fakeDns, ok := dnsServer.(*FakeDns); ok {
				if d := fakeDns.GetDomainFromFakeDNS(ip); d != "" {
					return d
				}
			}
		}
	}
	return ""
}

func isFakeDns(dnsServer DnsServer) bool {
	_, ok := dnsServer.(*FakeDns)
	return ok
}

func addClientIP(msg *dns.Msg, clientIp net.IP) {
	o := new(dns.OPT)
	o.Hdr.Name = "."
	o.Hdr.Rrtype = dns.TypeOPT

	subnet := &dns.EDNS0_SUBNET{
		Code:          dns.EDNS0SUBNET,
		Family:        1, // IPv4
		SourceNetmask: 24,
		SourceScope:   0,
		Address:       clientIp.To4(),
	}
	if clientIp.To4() == nil && clientIp.To16() != nil {
		subnet.Family = 2 // IPv6
		subnet.SourceNetmask = 64
		subnet.Address = clientIp.To16()
	}
	o.Option = append(o.Option, subnet)
	msg.Extra = append(msg.Extra, o)
}

type IPResolverWrapper struct {
	atomic.Value
}

func (r *IPResolverWrapper) GetIPResolver() i.IPResolver {
	resolver := r.Value.Load()
	if resolver == nil {
		return nil
	}
	return resolver.(i.IPResolver)
}

func (r *IPResolverWrapper) UpdateIPResolver(resolver i.IPResolver) {
	r.Value.Store(resolver)
}

func (r *IPResolverWrapper) LookupIP(ctx context.Context, domain string) ([]net.IP, error) {
	resolver := r.Value.Load()
	if resolver == nil {
		return nil, errors.New("ip resolver not found")
	}
	return resolver.(i.IPResolver).LookupIP(ctx, domain)
}

func (r *IPResolverWrapper) LookupIPv4(ctx context.Context, domain string) ([]net.IP, error) {
	resolver := r.Value.Load()
	if resolver == nil {
		return nil, errors.New("ip resolver not found")
	}
	return resolver.(i.IPResolver).LookupIPv4(ctx, domain)
}

func (r *IPResolverWrapper) LookupIPv6(ctx context.Context, domain string) ([]net.IP, error) {
	resolver := r.Value.Load()
	if resolver == nil {
		return nil, errors.New("ip resolver not found")
	}
	return resolver.(i.IPResolver).LookupIPv6(ctx, domain)
}

func (r *IPResolverWrapper) LookupIPSpeed(ctx context.Context, domain string) ([]net.IP, error) {
	resolver := r.Value.Load()
	if resolver == nil {
		return nil, errors.New("ip resolver not found")
	}
	return resolver.(i.IPResolver).LookupIPSpeed(ctx, domain)
}

type ECHResolverWrapper struct {
	atomic.Value
}

func (r *ECHResolverWrapper) GetECHResolver() i.ECHResolver {
	resolver := r.Value.Load()
	if resolver == nil {
		return nil
	}
	return resolver.(i.ECHResolver)
}

func (r *ECHResolverWrapper) UpdateECHResolver(resolver i.ECHResolver) {
	r.Value.Store(resolver)
}

func (r *ECHResolverWrapper) LookupECH(ctx context.Context, domain string) ([]byte, error) {
	resolver := r.Value.Load()
	if resolver == nil {
		return nil, errors.New("ech resolver not found")
	}
	return resolver.(i.ECHResolver).LookupECH(ctx, domain)
}
