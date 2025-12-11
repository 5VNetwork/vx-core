package dns

import (
	"fmt"
	"testing"
	"time"

	"github.com/5vnetwork/vx-core/common/net"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewIPToDomain(t *testing.T) {
	ipToDomain := NewIPToDomain(100)
	require.NotNil(t, ipToDomain)
	require.NotNil(t, ipToDomain.cache)
}

func TestIPToDomain_SetDomain_A_Record(t *testing.T) {
	ipToDomain := NewIPToDomain(100)

	// Create a DNS message with A record
	msg := &dns.Msg{
		Question: []dns.Question{
			{
				Name:   "example.com.",
				Qtype:  dns.TypeA,
				Qclass: dns.ClassINET,
			},
		},
		Answer: []dns.RR{
			&dns.A{
				Hdr: dns.RR_Header{
					Name:   "example.com.",
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    300,
				},
				A: net.ParseIP("192.168.1.1"),
			},
		},
	}

	resolver := "8.8.8.8"
	ipToDomain.SetDomain(msg, resolver)

	// Verify the domain was stored
	ip := net.ParseIP("192.168.1.1")
	domains := ipToDomain.GetDomain(ip)
	assert.Equal(t, []string{"example.com"}, domains)

	// Verify the resolver was stored
	resolvers := ipToDomain.GetResolvers("example.com", ip)
	assert.Equal(t, []string{resolver}, resolvers)
}

func TestIPToDomain_SetDomain_AAAA_Record(t *testing.T) {
	ipToDomain := NewIPToDomain(100)

	// Create a DNS message with AAAA record
	msg := &dns.Msg{
		Question: []dns.Question{
			{
				Name:   "example.com.",
				Qtype:  dns.TypeAAAA,
				Qclass: dns.ClassINET,
			},
		},
		Answer: []dns.RR{
			&dns.AAAA{
				Hdr: dns.RR_Header{
					Name:   "example.com.",
					Rrtype: dns.TypeAAAA,
					Class:  dns.ClassINET,
					Ttl:    300,
				},
				AAAA: net.ParseIP("2001:db8::1"),
			},
		},
	}

	resolver := "2001:4860:4860::8888"
	ipToDomain.SetDomain(msg, resolver)

	// Verify the domain was stored
	ip := net.ParseIP("2001:db8::1")
	domains := ipToDomain.GetDomain(ip)
	assert.Equal(t, []string{"example.com"}, domains)

	// Verify the resolver was stored
	resolvers := ipToDomain.GetResolvers("example.com", ip)
	assert.Equal(t, []string{resolver}, resolvers)
}

func TestIPToDomain_SetDomain_MultipleRecords(t *testing.T) {
	ipToDomain := NewIPToDomain(100)

	// Create a DNS message with multiple A records for the same domain
	msg := &dns.Msg{
		Question: []dns.Question{
			{
				Name:   "example.com.",
				Qtype:  dns.TypeA,
				Qclass: dns.ClassINET,
			},
		},
		Answer: []dns.RR{
			&dns.A{
				Hdr: dns.RR_Header{
					Name:   "example.com.",
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    300,
				},
				A: net.ParseIP("192.168.1.1"),
			},
			&dns.A{
				Hdr: dns.RR_Header{
					Name:   "example.com.",
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    300,
				},
				A: net.ParseIP("192.168.1.2"),
			},
		},
	}

	resolver := "8.8.8.8"
	ipToDomain.SetDomain(msg, resolver)

	// Verify both IPs map to the same domain
	ip1 := net.ParseIP("192.168.1.1")
	ip2 := net.ParseIP("192.168.1.2")

	assert.Equal(t, []string{"example.com"}, ipToDomain.GetDomain(ip1))
	assert.Equal(t, []string{"example.com"}, ipToDomain.GetDomain(ip2))

	// Verify resolvers are stored for both IPs
	assert.Equal(t, []string{resolver}, ipToDomain.GetResolvers("example.com", ip1))
	assert.Equal(t, []string{resolver}, ipToDomain.GetResolvers("example.com", ip2))
}

func TestIPToDomain_SetDomain_IgnoreNonARecords(t *testing.T) {
	ipToDomain := NewIPToDomain(100)

	// Create a DNS message with CNAME record (should be ignored)
	msg := &dns.Msg{
		Question: []dns.Question{
			{
				Name:   "example.com.",
				Qtype:  dns.TypeCNAME,
				Qclass: dns.ClassINET,
			},
		},
		Answer: []dns.RR{
			&dns.CNAME{
				Hdr: dns.RR_Header{
					Name:   "example.com.",
					Rrtype: dns.TypeCNAME,
					Class:  dns.ClassINET,
					Ttl:    300,
				},
				Target: "target.example.com.",
			},
		},
	}

	resolver := "8.8.8.8"
	ipToDomain.SetDomain(msg, resolver)

	// Should not store anything since it's not A or AAAA record
	ip := net.ParseIP("192.168.1.1")
	domains := ipToDomain.GetDomain(ip)
	assert.Nil(t, domains)
}

func TestIPToDomain_SetDomain_EmptyQuestion(t *testing.T) {
	ipToDomain := NewIPToDomain(100)

	// Create a DNS message with no questions
	msg := &dns.Msg{
		Question: []dns.Question{},
		Answer: []dns.RR{
			&dns.A{
				Hdr: dns.RR_Header{
					Name:   "example.com.",
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    300,
				},
				A: net.ParseIP("192.168.1.1"),
			},
		},
	}

	resolver := "8.8.8.8"
	ipToDomain.SetDomain(msg, resolver)

	// Should not store anything since there are no questions
	ip := net.ParseIP("192.168.1.1")
	domains := ipToDomain.GetDomain(ip)
	assert.Nil(t, domains)
}

func TestIPToDomain_GetDomain_NotFound(t *testing.T) {
	ipToDomain := NewIPToDomain(100)

	// Try to get domain for IP that doesn't exist in cache
	ip := net.ParseIP("192.168.1.1")
	domains := ipToDomain.GetDomain(ip)
	assert.Nil(t, domains)
}

func TestIPToDomain_GetDomain_SingleDomain(t *testing.T) {
	ipToDomain := NewIPToDomain(100)

	// Add a single domain-resolver pair
	msg := createDNSMessage("example.com.", "192.168.1.1")
	ipToDomain.SetDomain(msg, "8.8.8.8")

	ip := net.ParseIP("192.168.1.1")
	domains := ipToDomain.GetDomain(ip)
	assert.Equal(t, []string{"example.com"}, domains)
}

func TestIPToDomain_GetDomain_MultipleSameDomains(t *testing.T) {
	ipToDomain := NewIPToDomain(100)

	// Add same domain from multiple resolvers
	msg := createDNSMessage("example.com.", "192.168.1.1")
	ipToDomain.SetDomain(msg, "8.8.8.8")
	ipToDomain.SetDomain(msg, "1.1.1.1")

	ip := net.ParseIP("192.168.1.1")
	domains := ipToDomain.GetDomain(ip)
	// Should return all domains (same domain, multiple resolvers)
	assert.Equal(t, []string{"example.com", "example.com"}, domains)
}

func TestIPToDomain_GetDomain_MultipleDifferentDomains(t *testing.T) {
	ipToDomain := NewIPToDomain(100)

	// Add different domains for the same IP
	msg1 := createDNSMessage("example.com.", "192.168.1.1")
	msg2 := createDNSMessage("different.com.", "192.168.1.1")

	ipToDomain.SetDomain(msg1, "8.8.8.8")
	ipToDomain.SetDomain(msg2, "1.1.1.1")

	ip := net.ParseIP("192.168.1.1")
	domains := ipToDomain.GetDomain(ip)
	// Should return all domains (different domains map to same IP)
	assert.Contains(t, domains, "example.com")
	assert.Contains(t, domains, "different.com")
	assert.Len(t, domains, 2)
}

func TestIPToDomain_GetResolvers_NotFound(t *testing.T) {
	ipToDomain := NewIPToDomain(100)

	// Try to get resolvers for IP that doesn't exist
	ip := net.ParseIP("192.168.1.1")
	resolvers := ipToDomain.GetResolvers("example.com", ip)
	assert.Nil(t, resolvers)
}

func TestIPToDomain_GetResolvers_SingleResolver(t *testing.T) {
	ipToDomain := NewIPToDomain(100)

	msg := createDNSMessage("example.com.", "192.168.1.1")
	resolver := "8.8.8.8"
	ipToDomain.SetDomain(msg, resolver)

	ip := net.ParseIP("192.168.1.1")
	resolvers := ipToDomain.GetResolvers("example.com", ip)
	assert.Equal(t, []string{resolver}, resolvers)
}

func TestIPToDomain_GetResolvers_MultipleResolvers(t *testing.T) {
	ipToDomain := NewIPToDomain(100)

	msg := createDNSMessage("example.com.", "192.168.1.1")
	resolver1 := "8.8.8.8"
	resolver2 := "1.1.1.1"

	ipToDomain.SetDomain(msg, resolver1)
	ipToDomain.SetDomain(msg, resolver2)

	ip := net.ParseIP("192.168.1.1")
	resolvers := ipToDomain.GetResolvers("example.com", ip)

	// Should contain both resolvers
	assert.Len(t, resolvers, 2)
	assert.Contains(t, resolvers, resolver1)
	assert.Contains(t, resolvers, resolver2)
}

func TestIPToDomain_GetResolvers_DifferentDomains(t *testing.T) {
	ipToDomain := NewIPToDomain(100)

	// Add two different domains for the same IP
	msg1 := createDNSMessage("example.com.", "192.168.1.1")
	msg2 := createDNSMessage("different.com.", "192.168.1.1")

	ipToDomain.SetDomain(msg1, "8.8.8.8")
	ipToDomain.SetDomain(msg2, "1.1.1.1")

	ip := net.ParseIP("192.168.1.1")

	// Should only return resolvers for the specific domain requested
	resolvers1 := ipToDomain.GetResolvers("example.com", ip)
	resolvers2 := ipToDomain.GetResolvers("different.com", ip)

	assert.Equal(t, []string{"8.8.8.8"}, resolvers1)
	assert.Equal(t, []string{"1.1.1.1"}, resolvers2)
}

func TestIPToDomainEntry_AddDomain_Duplicate(t *testing.T) {
	entry := &ipToDomainEntry{
		domainAndResolvers: make([]DomainAndResolver, 0, 4),
	}

	expireTime := time.Now().Add(5 * time.Minute)

	// Add the same domain and resolver twice
	entry.addDomain("example.com", "8.8.8.8", expireTime)
	entry.addDomain("example.com", "8.8.8.8", expireTime)

	// Should only have one entry (duplicate updates expireTime)
	assert.Len(t, entry.domainAndResolvers, 1)
	assert.Equal(t, "example.com", entry.domainAndResolvers[0].Domain)
	assert.Equal(t, "8.8.8.8", entry.domainAndResolvers[0].Resolver)
}

func TestIPToDomainEntry_AddDomain_SameDomainDifferentResolver(t *testing.T) {
	entry := &ipToDomainEntry{
		domainAndResolvers: make([]DomainAndResolver, 0, 4),
	}

	expireTime := time.Now().Add(5 * time.Minute)

	// Add the same domain with different resolvers
	entry.addDomain("example.com", "8.8.8.8", expireTime)
	entry.addDomain("example.com", "1.1.1.1", expireTime)

	// Should have two entries
	assert.Len(t, entry.domainAndResolvers, 2)
	assert.Equal(t, "example.com", entry.domainAndResolvers[0].Domain)
	assert.Equal(t, "8.8.8.8", entry.domainAndResolvers[0].Resolver)
	assert.Equal(t, "example.com", entry.domainAndResolvers[1].Domain)
	assert.Equal(t, "1.1.1.1", entry.domainAndResolvers[1].Resolver)
}

func TestIPToDomainEntry_AddDomain_DifferentDomains(t *testing.T) {
	entry := &ipToDomainEntry{
		domainAndResolvers: make([]DomainAndResolver, 0, 4),
	}

	expireTime := time.Now().Add(5 * time.Minute)

	// Add different domains
	entry.addDomain("example.com", "8.8.8.8", expireTime)
	entry.addDomain("different.com", "1.1.1.1", expireTime)

	// Should have two entries
	assert.Len(t, entry.domainAndResolvers, 2)
	assert.Equal(t, "example.com", entry.domainAndResolvers[0].Domain)
	assert.Equal(t, "different.com", entry.domainAndResolvers[1].Domain)
}

func TestIPToDomainEntry_AddDomain_MaxEntriesLimit(t *testing.T) {
	entry := &ipToDomainEntry{
		domainAndResolvers: make([]DomainAndResolver, 0, 4),
	}

	expireTime := time.Now().Add(5 * time.Minute)

	// Add 5 entries (more than the limit of 4)
	entry.addDomain("domain1.com", "8.8.8.8", expireTime)
	entry.addDomain("domain2.com", "8.8.8.8", expireTime)
	entry.addDomain("domain3.com", "8.8.8.8", expireTime)
	entry.addDomain("domain4.com", "8.8.8.8", expireTime)
	entry.addDomain("domain5.com", "8.8.8.8", expireTime)

	// Should have exactly 4 entries (first one replaced by 5th)
	assert.Len(t, entry.domainAndResolvers, 4)
	// The 5th entry should replace the first one
	assert.Equal(t, "domain5.com", entry.domainAndResolvers[0].Domain)
	assert.Equal(t, "domain2.com", entry.domainAndResolvers[1].Domain)
	assert.Equal(t, "domain3.com", entry.domainAndResolvers[2].Domain)
	assert.Equal(t, "domain4.com", entry.domainAndResolvers[3].Domain)
}

func TestIPToDomain_UnFqdnHandling(t *testing.T) {
	ipToDomain := NewIPToDomain(100)

	// Create DNS message with FQDN (trailing dot)
	msg := &dns.Msg{
		Question: []dns.Question{
			{
				Name:   "example.com.",
				Qtype:  dns.TypeA,
				Qclass: dns.ClassINET,
			},
		},
		Answer: []dns.RR{
			&dns.A{
				Hdr: dns.RR_Header{
					Name:   "example.com.", // FQDN with trailing dot
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    300,
				},
				A: net.ParseIP("192.168.1.1"),
			},
		},
	}

	ipToDomain.SetDomain(msg, "8.8.8.8")

	// Should store domain without trailing dot
	ip := net.ParseIP("192.168.1.1")
	domains := ipToDomain.GetDomain(ip)
	assert.Equal(t, []string{"example.com"}, domains) // No trailing dot
}

func TestIPToDomain_ConcurrentAccess(t *testing.T) {
	ipToDomain := NewIPToDomain(100)

	// Test concurrent read/write access
	done := make(chan bool, 2)

	// Writer goroutine
	go func() {
		for i := 0; i < 100; i++ {
			msg := createDNSMessage("example.com.", "192.168.1.1")
			ipToDomain.SetDomain(msg, "8.8.8.8")
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 100; i++ {
			ip := net.ParseIP("192.168.1.1")
			_ = ipToDomain.GetDomain(ip)
			_ = ipToDomain.GetResolvers("example.com", ip)
		}
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	// Verify final state
	ip := net.ParseIP("192.168.1.1")
	domains := ipToDomain.GetDomain(ip)
	assert.Contains(t, domains, "example.com")
}

func TestIPToDomain_CacheLimit(t *testing.T) {
	ipToDomain := NewIPToDomain(100)

	// Add more entries than the cache limit (100)
	// This test verifies the LRU cache behavior
	for i := 0; i < 1100; i++ {
		ip := fmt.Sprintf("192.168.%d.%d", i/256, i%256)
		domain := fmt.Sprintf("example%d.com.", i)
		msg := createDNSMessage(domain, ip)
		ipToDomain.SetDomain(msg, "8.8.8.8")
	}

	// Early entries should be evicted due to LRU cache limit
	earlyIP := net.ParseIP("192.168.0.0")
	earlyDomains := ipToDomain.GetDomain(earlyIP)
	assert.Nil(t, earlyDomains, "Early entries should be evicted from cache")

	// Recent entries should still be present - check the last few entries
	lastIP := net.ParseIP("192.168.4.75") // 1099th entry (4*256 + 75 = 1099)
	lastDomains := ipToDomain.GetDomain(lastIP)
	assert.Contains(t, lastDomains, "example1099.com", "Last entry should still be in cache")

	// Also verify that some recent entries are still cached (not necessarily all due to LRU behavior)
	recentEntriesFound := 0
	for i := 1050; i < 1100; i++ {
		ip := fmt.Sprintf("192.168.%d.%d", i/256, i%256)
		testIP := net.ParseIP(ip)
		domains := ipToDomain.GetDomain(testIP)
		if len(domains) > 0 {
			recentEntriesFound++
		}
	}
	assert.Greater(t, recentEntriesFound, 0, "At least some recent entries should still be in cache")
}

func BenchmarkIPToDomain_SetDomain(b *testing.B) {
	ipToDomain := NewIPToDomain(100)
	msg := createDNSMessage("example.com.", "192.168.1.1")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ipToDomain.SetDomain(msg, "8.8.8.8")
	}
}

func BenchmarkIPToDomain_GetDomain(b *testing.B) {
	ipToDomain := NewIPToDomain(100)
	msg := createDNSMessage("example.com.", "192.168.1.1")
	ipToDomain.SetDomain(msg, "8.8.8.8")

	ip := net.ParseIP("192.168.1.1")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ipToDomain.GetDomain(ip)
	}
}

func BenchmarkIPToDomain_GetResolvers(b *testing.B) {
	ipToDomain := NewIPToDomain(100)
	msg := createDNSMessage("example.com.", "192.168.1.1")
	ipToDomain.SetDomain(msg, "8.8.8.8")

	ip := net.ParseIP("192.168.1.1")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ipToDomain.GetResolvers("example.com", ip)
	}
}

// TTL Expiration Tests

func TestIPToDomain_TTL_SetAndGet(t *testing.T) {
	ipToDomain := NewIPToDomain(100)

	// Create DNS message with 300 second TTL
	msg := &dns.Msg{
		Question: []dns.Question{
			{
				Name:   "example.com.",
				Qtype:  dns.TypeA,
				Qclass: dns.ClassINET,
			},
		},
		Answer: []dns.RR{
			&dns.A{
				Hdr: dns.RR_Header{
					Name:   "example.com.",
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    300, // 5 minutes
				},
				A: net.ParseIP("192.168.1.1"),
			},
		},
	}

	ipToDomain.SetDomain(msg, "8.8.8.8")

	// Verify ExpireTime was set correctly
	ip := net.ParseIP("192.168.1.1")
	v, ok := ipToDomain.cache.Get(net.IPAddress(ip))
	require.True(t, ok)
	entry := v.(*ipToDomainEntry)

	entry.lock.RLock()
	require.Len(t, entry.domainAndResolvers, 1)
	expireTime := entry.domainAndResolvers[0].ExpireTime
	entry.lock.RUnlock()

	// Verify expire time is approximately 300 seconds in the future
	expectedExpire := time.Now().Add(300 * time.Second)
	timeDiff := expireTime.Sub(expectedExpire)
	assert.Less(t, timeDiff.Abs().Seconds(), 2.0, "ExpireTime should be ~300 seconds from now")
}

func TestIPToDomain_TTL_UpdateExpireTime(t *testing.T) {
	ipToDomain := NewIPToDomain(100)

	// Add entry with short TTL
	msg1 := &dns.Msg{
		Question: []dns.Question{
			{
				Name:   "example.com.",
				Qtype:  dns.TypeA,
				Qclass: dns.ClassINET,
			},
		},
		Answer: []dns.RR{
			&dns.A{
				Hdr: dns.RR_Header{
					Name:   "example.com.",
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    2, // 2 seconds
				},
				A: net.ParseIP("192.168.1.1"),
			},
		},
	}
	ipToDomain.SetDomain(msg1, "8.8.8.8")

	ip := net.ParseIP("192.168.1.1")

	// Wait 1 second
	time.Sleep(1 * time.Second)

	// Update with new TTL from same resolver
	msg2 := &dns.Msg{
		Question: []dns.Question{
			{
				Name:   "example.com.",
				Qtype:  dns.TypeA,
				Qclass: dns.ClassINET,
			},
		},
		Answer: []dns.RR{
			&dns.A{
				Hdr: dns.RR_Header{
					Name:   "example.com.",
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    300, // 5 minutes - refresh
				},
				A: net.ParseIP("192.168.1.1"),
			},
		},
	}
	ipToDomain.SetDomain(msg2, "8.8.8.8")

	// Should only have one entry (updated)
	v, ok := ipToDomain.cache.Get(net.IPAddress(ip))
	require.True(t, ok)
	entry := v.(*ipToDomainEntry)
	entry.lock.RLock()
	assert.Len(t, entry.domainAndResolvers, 1)
	entry.lock.RUnlock()

	// Wait for original TTL to expire
	time.Sleep(1100 * time.Millisecond)

	// Should still be available due to refresh
	domains := ipToDomain.GetDomain(ip)
	assert.Contains(t, domains, "example.com")
}

func TestIPToDomain_TTL_ConcurrentAccessWithExpiration(t *testing.T) {
	ipToDomain := NewIPToDomain(100)

	// Add entry with 2 second TTL
	msg := &dns.Msg{
		Question: []dns.Question{
			{
				Name:   "example.com.",
				Qtype:  dns.TypeA,
				Qclass: dns.ClassINET,
			},
		},
		Answer: []dns.RR{
			&dns.A{
				Hdr: dns.RR_Header{
					Name:   "example.com.",
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    2, // 2 seconds
				},
				A: net.ParseIP("192.168.1.1"),
			},
		},
	}
	ipToDomain.SetDomain(msg, "8.8.8.8")

	done := make(chan bool, 3)
	ip := net.ParseIP("192.168.1.1")

	// Reader goroutine 1
	go func() {
		for i := 0; i < 50; i++ {
			_ = ipToDomain.GetDomain(ip)
			time.Sleep(10 * time.Millisecond)
		}
		done <- true
	}()

	// Reader goroutine 2
	go func() {
		for i := 0; i < 50; i++ {
			_ = ipToDomain.GetResolvers("example.com", ip)
			time.Sleep(10 * time.Millisecond)
		}
		done <- true
	}()

	// Writer goroutine - refresh entry
	go func() {
		time.Sleep(1 * time.Second)
		refreshMsg := &dns.Msg{
			Question: []dns.Question{
				{
					Name:   "example.com.",
					Qtype:  dns.TypeA,
					Qclass: dns.ClassINET,
				},
			},
			Answer: []dns.RR{
				&dns.A{
					Hdr: dns.RR_Header{
						Name:   "example.com.",
						Rrtype: dns.TypeA,
						Class:  dns.ClassINET,
						Ttl:    300,
					},
					A: net.ParseIP("192.168.1.1"),
				},
			},
		}
		ipToDomain.SetDomain(refreshMsg, "8.8.8.8")
		done <- true
	}()

	// Wait for all goroutines
	<-done
	<-done
	<-done

	// After all operations, entry should still be valid (refreshed)
	domains := ipToDomain.GetDomain(ip)
	assert.Contains(t, domains, "example.com")
}

// Helper function to create DNS messages for testing
func createDNSMessage(domain string, ipStr string) *dns.Msg {
	return &dns.Msg{
		Question: []dns.Question{
			{
				Name:   domain,
				Qtype:  dns.TypeA,
				Qclass: dns.ClassINET,
			},
		},
		Answer: []dns.RR{
			&dns.A{
				Hdr: dns.RR_Header{
					Name:   domain,
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    300,
				},
				A: net.ParseIP(ipStr),
			},
		},
	}
}
