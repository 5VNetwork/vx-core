// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package dns

import (
	sync "sync"
	"time"

	"github.com/5vnetwork/vx-core/common/cache"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/miekg/dns"
)

type IPToDomain struct {
	cache                      cache.Lru //key is net.Address, value is *ipToDomainEntry
	maxDomainAndResolversPerIp int
}

func NewIPToDomain(size int, maxDomainAndResolversPerIp int) *IPToDomain {
	return &IPToDomain{
		cache:                      cache.NewLru1(size),
		maxDomainAndResolversPerIp: maxDomainAndResolversPerIp,
	}
}

type ipToDomainEntry struct {
	lock               sync.RWMutex
	domainAndResolvers []DomainAndResolver
}

func (e *ipToDomainEntry) addDomain(d string, resolver net.Address, expireAt int64) {
	e.lock.Lock()
	defer e.lock.Unlock()

	// Remove expired entries
	now := time.Now().Unix()
	valid := e.domainAndResolvers[:0]
	for _, dr := range e.domainAndResolvers {
		if dr.ExpireAt >= now {
			valid = append(valid, dr)
		}
	}
	e.domainAndResolvers = valid

	// add new entry
	for i, dr := range e.domainAndResolvers {
		// if the domain and resolver already exist, update expireAt
		if dr.Domain == d && dr.Resolver == resolver {
			e.domainAndResolvers[i].ExpireAt = expireAt
			return
		}
	}
	entry := DomainAndResolver{
		Domain:   d,
		Resolver: resolver,
		ExpireAt: expireAt,
	}
	if len(e.domainAndResolvers) == 0 {
		e.domainAndResolvers = append(e.domainAndResolvers, entry)
	} else {
		if len(e.domainAndResolvers) < cap(e.domainAndResolvers) {
			e.domainAndResolvers = append(e.domainAndResolvers, entry)
		}
		copy(e.domainAndResolvers[1:], e.domainAndResolvers[:len(e.domainAndResolvers)-1])
		e.domainAndResolvers[0] = entry
	}
}

type DomainAndResolver struct {
	Domain   string
	Resolver net.Address
	ExpireAt int64 // Unix timestamp (saves 16 bytes vs time.Time)
}

func (i *IPToDomain) GetDomain(ip net.IP) []string {
	v, ok := i.cache.Get(net.IPAddress(ip))
	if !ok {
		return nil
	}
	entry := v.(*ipToDomainEntry)

	entry.lock.RLock()
	defer entry.lock.RUnlock()

	domains := make([]string, 0, len(entry.domainAndResolvers))
	for _, dr := range entry.domainAndResolvers {
		domains = append(domains, dr.Domain)
	}
	return domains
}

func (i *IPToDomain) GetResolvers(domain string, ip net.IP) []net.Address {
	v, ok := i.cache.Get(net.IPAddress(ip))
	if !ok {
		return nil
	}
	entry := v.(*ipToDomainEntry)

	entry.lock.RLock()
	defer entry.lock.RUnlock()
	var resolvers []net.Address
	for _, dr := range entry.domainAndResolvers {
		if dr.Domain == domain {
			resolvers = append(resolvers, dr.Resolver)
		}
	}
	return resolvers
}

func (i *IPToDomain) SetDomain(reply *dns.Msg, src net.Address) {
	if len(reply.Question) == 0 {
		return
	}
	question := reply.Question[0]
	if question.Qtype != dns.TypeA && question.Qtype != dns.TypeAAAA {
		return
	}

	for _, rr := range reply.Answer {
		// Calculate expiration time based on TTL (unix timestamp)
		expireAt := time.Now().Unix() + int64(rr.Header().Ttl)

		if a, ok := rr.(*dns.A); ok {
			addr := net.IPAddress(a.A)
			entry, ok := i.cache.Get(addr)
			if !ok {
				entry = &ipToDomainEntry{
					domainAndResolvers: make([]DomainAndResolver, 0, i.maxDomainAndResolversPerIp),
				}
				i.cache.Put(addr, entry)
			}
			entry.(*ipToDomainEntry).addDomain(UnFqdn(rr.Header().Name), src, expireAt)
		}
		if aaaa, ok := rr.(*dns.AAAA); ok {
			addr := net.IPAddress(aaaa.AAAA)
			entry, ok := i.cache.Get(addr)
			if !ok {
				entry = &ipToDomainEntry{
					domainAndResolvers: make([]DomainAndResolver, 0, i.maxDomainAndResolversPerIp),
				}
				i.cache.Put(addr, entry)
			}
			entry.(*ipToDomainEntry).addDomain(UnFqdn(rr.Header().Name), src, expireAt)
		}
	}
}
