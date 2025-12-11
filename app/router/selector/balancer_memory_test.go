package selector

import (
	"context"
	"net"
	"testing"

	"github.com/5vnetwork/vx-core/common/buf"
	mynet "github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/common/session"
	"github.com/5vnetwork/vx-core/i"
	"github.com/stretchr/testify/assert"
)

// MockMemoryHandler implements i.HandlerWith6Info for testing
type MockMemoryHandler struct {
	tag      string
	support6 bool
}

func (m *MockMemoryHandler) Tag() string {
	return m.tag
}

func (m *MockMemoryHandler) Support6() bool {
	return m.support6
}

func (m *MockMemoryHandler) HandleFlow(ctx context.Context, dst mynet.Destination, rw buf.ReaderWriter) error {
	return nil
}

func (m *MockMemoryHandler) HandlePacketConn(ctx context.Context, dst mynet.Destination, p udp.PacketReaderWriter) error {
	return nil
}

// MockIpToDomain implements ipToDomain interface for testing
type MockIpToDomain struct {
	mapping map[string]string // IP string to domain mapping
}

func NewMockIpToDomain() *MockIpToDomain {
	return &MockIpToDomain{
		mapping: make(map[string]string),
	}
}

func (m *MockIpToDomain) GetDomain(ip net.IP) string {
	return m.mapping[ip.String()]
}

func (m *MockIpToDomain) SetMapping(ip string, domain string) {
	m.mapping[ip] = domain
}

// Test helper functions
func createMockMemoryHandler(tag string, support6 bool) *MockMemoryHandler {
	return &MockMemoryHandler{
		tag:      tag,
		support6: support6,
	}
}

func createTestInfo() *session.Info {
	return &session.Info{}
}

func createTestInfoWithApp(appId string) *session.Info {
	return &session.Info{
		AppId: appId,
		Target: mynet.Destination{
			Address: mynet.DomainAddress("example.com"),
		},
	}
}

func createTestInfoWithDomain(domain string) *session.Info {
	return &session.Info{
		Target: mynet.Destination{
			Address: mynet.DomainAddress(domain),
		},
	}
}

func createTestInfoWithIP(ip string) *session.Info {
	return &session.Info{
		Target: mynet.Destination{
			Address: mynet.IPAddress(net.ParseIP(ip)),
		},
	}
}

func createTestInfoWithIPv6(ipv6 string) *session.Info {
	return &session.Info{
		Target: mynet.Destination{
			Address: mynet.IPAddress(net.ParseIP(ipv6)),
		},
	}
}

func createTestInfoWithFakeIPv6() *session.Info {
	fakeIP := net.ParseIP("2001:db8::1")
	return &session.Info{
		FakeIP: fakeIP,
		Target: mynet.Destination{
			Address: mynet.DomainAddress("fake.example.com"),
		},
	}
}

// verify that for a specific app or domain, returned handler is the same
func TestMemoryBalancerGetHandler(t *testing.T) {
	getBalancer := func() Balancer {
		balancer := NewMemoryBalancer()

		// Create test handlers
		handler4_1 := createMockMemoryHandler("handler4_1", false)
		handler4_2 := createMockMemoryHandler("handler4_2", false)
		handler6_1 := createMockMemoryHandler("handler6_1", true)
		handler6_2 := createMockMemoryHandler("handler6_2", true)

		handlers := []i.HandlerWith6Info{handler4_1, handler4_2, handler6_1, handler6_2}
		balancer.UpdateHandlers(handlers)
		return balancer
	}

	t.Run("Same App ID Returns Same Handler", func(t *testing.T) {
		// Test with a specific app ID
		appInfo := createTestInfoWithApp("com.example.testapp")
		balancer := getBalancer()
		// Get handler multiple times for the same app
		firstHandler := balancer.GetHandler(appInfo)
		assert.NotNil(t, firstHandler)

		// Verify that subsequent calls return the same handler
		for j := 0; j < 10; j++ {
			sameHandler := balancer.GetHandler(appInfo)
			assert.Equal(t, firstHandler, sameHandler, "Handler should be consistent for the same app ID")
			assert.Equal(t, firstHandler.Tag(), sameHandler.Tag(), "Handler tag should be consistent")
		}
	})

	t.Run("Different App IDs Can Get Different Handlers", func(t *testing.T) {
		// Test with different app IDs
		app1Info := createTestInfoWithApp("com.example.app1")
		app2Info := createTestInfoWithApp("com.example.app2")

		balancer := getBalancer()
		handler1 := balancer.GetHandler(app1Info)
		handler2 := balancer.GetHandler(app2Info)

		assert.NotNil(t, handler1)
		assert.NotNil(t, handler2)

		// They might be the same due to random selection, but each should be consistent
		// Verify consistency for each app
		for j := 0; j < 5; j++ {
			assert.Equal(t, handler1, balancer.GetHandler(app1Info), "App1 should get consistent handler")
			assert.Equal(t, handler2, balancer.GetHandler(app2Info), "App2 should get consistent handler")
		}
	})

	t.Run("Same Domain Returns Same Handler", func(t *testing.T) {
		// Test with a specific domain
		domainInfo := createTestInfoWithDomain("example.com")
		balancer := getBalancer()
		// Get handler multiple times for the same domain
		firstHandler := balancer.GetHandler(domainInfo)
		assert.NotNil(t, firstHandler)

		// Verify that subsequent calls return the same handler
		for j := 0; j < 10; j++ {
			sameHandler := balancer.GetHandler(domainInfo)
			assert.Equal(t, firstHandler, sameHandler, "Handler should be consistent for the same domain")
			assert.Equal(t, firstHandler.Tag(), sameHandler.Tag(), "Handler tag should be consistent")
		}
	})

	t.Run("Different Domains Can Get Different Handlers", func(t *testing.T) {
		// Test with different domains
		domain1Info := createTestInfoWithDomain("example1.com")
		domain2Info := createTestInfoWithDomain("example2.com")
		balancer := getBalancer()
		handler1 := balancer.GetHandler(domain1Info)
		handler2 := balancer.GetHandler(domain2Info)

		assert.NotNil(t, handler1)
		assert.NotNil(t, handler2)

		// Verify consistency for each domain
		for j := 0; j < 5; j++ {
			assert.Equal(t, handler1, balancer.GetHandler(domain1Info), "Domain1 should get consistent handler")
			assert.Equal(t, handler2, balancer.GetHandler(domain2Info), "Domain2 should get consistent handler")
		}
	})

	t.Run("IPv6 Targets Get IPv6 Handlers When Available", func(t *testing.T) {
		// Test with IPv6 target
		ipv6Info := createTestInfoWithIPv6("2001:db8::1")
		balancer := getBalancer()
		// Should get IPv6-capable handlers (may be different each time due to random selection)
		for j := 0; j < 5; j++ {
			handler := balancer.GetHandler(ipv6Info)
			assert.NotNil(t, handler)
			if h6, ok := handler.(i.HandlerWith6Info); ok {
				assert.True(t, h6.Support6(), "IPv6 target should get IPv6-capable handler")
			}
		}
	})

	t.Run("FakeIP IPv6 Gets IPv6 Handlers When Available", func(t *testing.T) {
		// Test with FakeIP IPv6 - should get consistent handler since it has a domain
		fakeIPv6Info := createTestInfoWithFakeIPv6()
		balancer := getBalancer()
		// Should get an IPv6-capable handler
		handler := balancer.GetHandler(fakeIPv6Info)
		assert.NotNil(t, handler)
		if h6, ok := handler.(i.HandlerWith6Info); ok {
			assert.True(t, h6.Support6(), "FakeIP IPv6 should get IPv6-capable handler")
		}

		// Verify consistency (should be consistent since it has a domain)
		for j := 0; j < 5; j++ {
			sameHandler := balancer.GetHandler(fakeIPv6Info)
			assert.Equal(t, handler, sameHandler, "FakeIP IPv6 should get consistent handler")
		}
	})

	t.Run("IPv4 Targets Can Get Any Handler", func(t *testing.T) {
		// Test with IPv4 target - may get different handlers each time since no AppID or domain
		ipv4Info := createTestInfoWithIP("192.168.1.1")
		balancer := getBalancer()
		// Just verify we get handlers (may be different each time)
		for j := 0; j < 5; j++ {
			handler := balancer.GetHandler(ipv4Info)
			assert.NotNil(t, handler, "IPv4 target should get some handler")
		}
	})

	t.Run("Cache Persistence Across Handler Updates", func(t *testing.T) {
		balancer := getBalancer()

		// Create test handlers
		handler4_1 := createMockMemoryHandler("handler4_1", false)
		handler4_2 := createMockMemoryHandler("handler4_2", false)
		handler6_1 := createMockMemoryHandler("handler6_1", true)
		handler6_2 := createMockMemoryHandler("handler6_2", true)

		handlers := []i.HandlerWith6Info{handler4_1, handler4_2, handler6_1, handler6_2}
		balancer.UpdateHandlers(handlers)

		// Get a handler for a specific app
		appInfo := createTestInfoWithApp("com.persistent.app")
		originalHandler := balancer.GetHandler(appInfo)
		assert.NotNil(t, originalHandler)

		// Update handlers with the same set (simulating refresh)
		balancer.UpdateHandlers(handlers)

		// Should still get the same handler (cache should persist)
		newHandler := balancer.GetHandler(appInfo)
		assert.Equal(t, originalHandler, newHandler, "Handler should persist across handler updates")
	})

	t.Run("No Handlers Available", func(t *testing.T) {
		emptyBalancer := NewMemoryBalancer()

		appInfo := createTestInfoWithApp("com.example.nohandlers")
		handler := emptyBalancer.GetHandler(appInfo)

		assert.Nil(t, handler, "Should return nil when no handlers are available")
	})
}

func TestMemoryBalancerIPv6OnlyScenario(t *testing.T) {
	balancer := NewMemoryBalancer()

	// Only IPv6 handlers
	handler6_1 := createMockMemoryHandler("handler6_1", true)
	handler6_2 := createMockMemoryHandler("handler6_2", true)

	handlers := []i.HandlerWith6Info{handler6_1, handler6_2}
	balancer.UpdateHandlers(handlers)

	t.Run("IPv4 Targets Fall Back to IPv6 Handlers", func(t *testing.T) {
		ipv4Info := createTestInfoWithIP("192.168.1.1")

		// Verify IPv6 handlers are returned (may be different each time due to random selection)
		for n := 0; n < 5; n++ {
			handler := balancer.GetHandler(ipv4Info)
			assert.NotNil(t, handler)
			if h6, ok := handler.(i.HandlerWith6Info); ok {
				assert.True(t, h6.Support6(), "Should get IPv6 handler even for IPv4 target")
			}
		}
	})
}

func TestMemoryBalancerIPv4OnlyScenario(t *testing.T) {
	balancer := NewMemoryBalancer()

	// Only IPv4 handlers
	handler4_1 := createMockMemoryHandler("handler4_1", false)
	handler4_2 := createMockMemoryHandler("handler4_2", false)

	handlers := []i.HandlerWith6Info{handler4_1, handler4_2}
	balancer.UpdateHandlers(handlers)

	t.Run("IPv6 Targets Fall Back to IPv4 Handlers", func(t *testing.T) {
		ipv6Info := createTestInfoWithIPv6("2001:db8::1")

		// Verify IPv4 handlers are returned (may be different each time due to random selection)
		for j := 0; j < 5; j++ {
			handler := balancer.GetHandler(ipv6Info)
			assert.NotNil(t, handler)
			if h6, ok := handler.(i.HandlerWith6Info); ok {
				assert.False(t, h6.Support6(), "Should get IPv4 handler when no IPv6 handlers available")
			}
		}
	})
}
