package dns_test

import (
	"net"
	"strconv"
	"testing"

	"github.com/5vnetwork/vx-core/app/configs"
	. "github.com/5vnetwork/vx-core/app/dns"
	"github.com/5vnetwork/vx-core/common"
	net1 "github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/uuid"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
)

func TestPoolBasic(t *testing.T) {
	pool, err := NewPool("198.18.0.0/15", 65535)
	common.Must(err)
	addr := pool.GetFakeIPForDomain("a.b", true)
	assert.Equal(t, "198.18.0.0", addr.String())

	addr0 := pool.GetFakeIPForDomain("a.c", true)
	assert.Equal(t, "198.18.0.1", addr0.String())

	domain := pool.GetDomainFromFakeIP(net1.ParseAddress("198.18.0.0"))
	assert.Equal(t, "a.b", domain)

	domain0 := pool.GetDomainFromFakeIP(net1.ParseAddress("198.18.0.1"))
	assert.Equal(t, "a.c", domain0)

	addr = pool.GetFakeIPForDomain("a.b", true)
	assert.Equal(t, "198.18.0.0", addr.String())

}

func TestGetFakeIPForDomainConcurrently(t *testing.T) {
	pool, err := NewPool("198.18.0.0/15", 65535)
	common.Must(err)
	total := 200
	addr := make([]net1.Address, total+1)
	var errg errgroup.Group
	for i := 0; i < total; i++ {
		errg.Go(testGetFakeIP(i, addr, pool))
	}
	errg.Wait()
	for i := 0; i < total; i++ {
		for j := i + 1; j < total; j++ {
			assert.NotEqual(t, addr[i].IP().String(), addr[j].IP().String())
		}
	}
}

func testGetFakeIP(index int, addr []net1.Address, p *Pool) func() error {
	return func() error {
		addr[index] = p.GetFakeIPForDomain("fakednstest"+strconv.Itoa(index)+".example.com", true)
		return nil
	}
}

func TestRollingOver(t *testing.T) {
	pool, err := NewPool("240.0.0.0/12", 256)
	common.Must(err)

	addr := pool.GetFakeIPForDomain("a.b", true)
	assert.Equal(t, "240.0.0.0", addr.String())

	addr0 := pool.GetFakeIPForDomain("a.c", true)
	assert.Equal(t, "240.0.0.1", addr0.String())

	for i := 0; i <= 8192; i++ {
		{
			result := pool.GetDomainFromFakeIP(net1.ParseAddress("240.0.0.0"))
			assert.Equal(t, "a.b", result)
		}

		{
			result := pool.GetDomainFromFakeIP(net1.ParseAddress("240.0.0.1"))
			assert.Equal(t, "a.c", result)
		}

		{
			uuid := uuid.New()
			domain := uuid.String() + ".a.b"
			addr := pool.GetFakeIPForDomain(domain, true)
			rsaddr := addr.IP().String()

			result := pool.GetDomainFromFakeIP(net1.ParseAddress(rsaddr))
			assert.Equal(t, domain, result)
		}
	}
}

func TestPools(t *testing.T) {
	pools, err := NewPools([]*configs.FakeDnsServer_PoolConfig{
		{
			Cidr:    "240.0.0.0/12",
			LruSize: 256,
		},
		{
			Cidr:    "fddd:c5b4:ff5f:f4f0::/64",
			LruSize: 256,
		},
	})
	common.Must(err)
	t.Run("checkInRange", func(t *testing.T) {
		t.Run("ipv4", func(t *testing.T) {
			inPool := pools.IsIPInIPPool(net1.IPAddress([]byte{240, 0, 0, 5}))
			assert.True(t, inPool)
		})
		t.Run("ipv6", func(t *testing.T) {
			ip, err := net.ResolveIPAddr("ip", "fddd:c5b4:ff5f:f4f0::5")
			assert.Nil(t, err)
			inPool := pools.IsIPInIPPool(net1.IPAddress(ip.IP))
			assert.True(t, inPool)
		})
		t.Run("ipv4_inverse", func(t *testing.T) {
			inPool := pools.IsIPInIPPool(net1.IPAddress([]byte{241, 0, 0, 5}))
			assert.False(t, inPool)
		})
		t.Run("ipv6_inverse", func(t *testing.T) {
			ip, err := net.ResolveIPAddr("ip", "fcdd:c5b4:ff5f:f4f0::5")
			assert.Nil(t, err)
			inPool := pools.IsIPInIPPool(net1.IPAddress(ip.IP))
			assert.False(t, inPool)
		})
	})

	// t.Run("allocateTwoAddressForTwoPool", func(t *testing.T) {
	// 	address := pools.GetFakeIPForDomain("fakednstest.v2fly.org", true)
	// 	assert.True(t, address.Family().IsIPv4(), "should be ipv4 address")
	// 	address0 := pools.GetFakeIPForDomain("fakednstest.v2fly.org", false)
	// 	assert.True(t, address0.Family().IsIPv6(), "should be ipv6 address")
	// 	t.Run("eachOfThemShouldResolve:0", func(t *testing.T) {
	// 		domain := pools.GetDomainFromFakeDNS(address)
	// 		assert.Equal(t, "fakednstest.v2fly.org", domain)
	// 	})
	// 	t.Run("eachOfThemShouldResolve:1", func(t *testing.T) {
	// 		domain := pools.GetDomainFromFakeDNS(address0)
	// 		assert.Equal(t, "fakednstest.v2fly.org", domain)
	// 	})
	// })

}

// func TestPoolsAdd(t *testing.T) {
// 	pools, err := NewPools([]*PoolConfig{
// 		{
// 			Cidr:    "240.0.0.0/12",
// 			LruSize: 256,
// 		},
// 		{
// 			Cidr:    "fddd:c5b4:ff5f:f4f0::/64",
// 			LruSize: 256,
// 		},
// 	})
// 	common.Must(err)

// 	t.Run("ipv4_return_existing", func(t *testing.T) {
// 		pool, err := pools.AddPool("240.0.0.0/12", 256)
// 		common.Must(err)
// 		if pool != pools[0] {
// 			t.Error("HolderMulti.AddPool not returning same holder for existing IPv4 pool")
// 		}
// 	})
// 	t.Run("ipv6_return_existing", func(t *testing.T) {
// 		pool, err := pools.AddPool("fddd:c5b4:ff5f:f4f0::1/64", 256)
// 		common.Must(err)
// 		if pool != pools[1] {
// 			t.Error("HolderMulti.AddPool not returning same holder for existing IPv4 pool")
// 		}
// 	})
// 	t.Run("ipv4_reject_overlap", func(t *testing.T) {
// 		_, err := pools.AddPool("240.8.0.0/13", 256)
// 		if err == nil {
// 			t.Error("not rejecting IPv4 pool that is subnet of existing ones")
// 		}
// 		_, err = pools.AddPool("240.0.0.0/11", 256)
// 		if err == nil {
// 			t.Error("not rejecting IPv4 pool that is subnet of existing ones")
// 		}
// 	})
// 	t.Run("new_pool", func(t *testing.T) {
// 		pool, err := pools.AddPool("192.168.168.0/16", 256)
// 		common.Must(err)

// 		if pool != pools[2] {
// 			t.Error("not creating new holder for new IPv4 pool")
// 		}
// 	})
// }
