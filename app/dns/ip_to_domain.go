package dns

import (
	sync "sync"
	"time"

	"github.com/5vnetwork/vx-core/common/cache"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/miekg/dns"
)

type IPToDomain struct {
	cache cache.Lru //key is net.Address, value is *ipToDomainEntry
}

func NewIPToDomain(size int) *IPToDomain {
	return &IPToDomain{
		cache: cache.NewLru(size),
	}
}

type ipToDomainEntry struct {
	lock               sync.RWMutex
	domainAndResolvers []DomainAndResolver
}

func (e *ipToDomainEntry) addDomain(d string, resolver net.Address, expireTime time.Time) {
	e.lock.Lock()
	defer e.lock.Unlock()

	// Remove expired entries
	now := time.Now()
	valid := e.domainAndResolvers[:0]
	for _, dr := range e.domainAndResolvers {
		if !dr.ExpireTime.Before(now) {
			valid = append(valid, dr)
		}
	}
	e.domainAndResolvers = valid

	// add new entry
	for i, dr := range e.domainAndResolvers {
		// if the domain and resolver already exist, update expireTime
		if dr.Domain == d && dr.Resolver == resolver {
			e.domainAndResolvers[i].ExpireTime = expireTime
			return
		}
	}
	entry := DomainAndResolver{
		Domain:     d,
		Resolver:   resolver,
		ExpireTime: expireTime,
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
	Domain     string
	Resolver   net.Address
	ExpireTime time.Time
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
		// Calculate expiration time based on TTL
		ttl := time.Duration(rr.Header().Ttl) * time.Second
		expireTime := time.Now().Add(ttl)

		if a, ok := rr.(*dns.A); ok {
			entryI, ok := i.cache.Get(net.IPAddress(a.A))
			if !ok {
				entryI = &ipToDomainEntry{
					domainAndResolvers: make([]DomainAndResolver, 0, 4),
				}
				i.cache.Put(net.IPAddress(a.A), entryI)
			}
			entry := entryI.(*ipToDomainEntry)
			entry.addDomain(UnFqdn(rr.Header().Name), src, expireTime)
		}
		if aaaa, ok := rr.(*dns.AAAA); ok {
			entryI, ok := i.cache.Get(net.IPAddress(aaaa.AAAA))
			if !ok {
				entryI = &ipToDomainEntry{
					domainAndResolvers: make([]DomainAndResolver, 0, 4),
				}
				i.cache.Put(net.IPAddress(aaaa.AAAA), entryI)
			}
			entry := entryI.(*ipToDomainEntry)
			entry.addDomain(UnFqdn(rr.Header().Name), src, expireTime)
		}
	}
}
