package dns

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"net"
	"sync"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/common/cache"
	nethelper "github.com/5vnetwork/vx-core/common/net"
	"github.com/rs/zerolog/log"
)

func NewPools(poolConfigs []*configs.FakeDnsServer_PoolConfig) (Pools, error) {
	p := Pools(make([]*Pool, 0, len(poolConfigs)))
	for _, config := range poolConfigs {
		if err := p.addPool(config.Cidr, uint16(config.LruSize)); err != nil {
			return nil, err
		}
	}
	return p, nil
}

type Pools []*Pool

type PoolConfig struct {
	Cidr    string
	LruSize uint16
}

// Add a pool to h
//
// firstly check if any exsiting pool's cidr overlaps with the new pool's cidr or
// same as the new pool's cidr. If same, return the existing pool. If overlaps, return error.
// If none of the above, create a new pool and add it to the pools list.
func (h *Pools) addPool(cidr string, lruSize uint16) error {
	_, ipRange, err := net.ParseCIDR(cidr)
	if err != nil {
		return fmt.Errorf("cannot parseCIDR, %w", err)
	}
	alreadyExisted := false
	for _, existingPool := range *h {
		if existingPool.ipRange.String() == ipRange.String() {
			alreadyExisted = true
			break
		}
		if existingPool.ipRange.Contains(ipRange.IP) || ipRange.Contains(existingPool.ipRange.IP) {
			return errors.New("Trying to add ip pool " + ipRange.String() + " that overlaps with existing ip pool " + existingPool.ipRange.String())
		}
	}
	if !alreadyExisted {
		pool, err := NewPool(cidr, int(lruSize))
		if err != nil {
			return fmt.Errorf("cannot create a pool, %w", err)
		}
		*h = append(*h, pool)
	}
	return nil
}

func (h Pools) IsIPInIPPool(ip nethelper.Address) bool {
	for _, v := range h {
		if v.IsIPInIPPool(ip) {
			return true
		}
	}
	return false
}

func (h Pools) GetDomainFromFakeDNS(ip nethelper.Address) string {
	for _, pool := range h {
		if domain := pool.GetDomainFromFakeIP(ip); domain != "" {
			return domain
		}
	}
	return ""
}

func (h Pools) GetFakeIP(domain string) ([]net.IP, error) {
	ips := make([]net.IP, 0, 2)
	ipv4, _ := h.GetFakeIPv4(domain)
	if ipv4 != nil {
		ips = append(ips, ipv4)
	}
	ipv6, _ := h.GetFakeIPv6(domain)
	if ipv6 != nil {
		ips = append(ips, ipv6)
	}
	if len(ips) == 0 {
		return nil, errors.New("cannot create a fake ip for" + domain)
	}
	log.Debug().Str("domain", domain).Msgf("fake ip %s", ips)
	return ips, nil
}
func (h Pools) GetFakeIPv4(domain string) (net.IP, error) {
	for _, p := range h {
		ip := p.GetFakeIPForDomain(domain, true)
		if ip != nil {
			log.Debug().Str("domain", domain).Str("fake_ip", ip.String()).Msg("fake ip 4")
			return ip.IP(), nil
		}
	}
	return nil, errors.New("cannot create a fake ipv4 for" + domain)
}

func (h Pools) GetFakeIPv6(domain string) (net.IP, error) {
	for _, p := range h {
		ip := p.GetFakeIPForDomain(domain, false)
		if ip != nil {
			log.Debug().Str("domain", domain).Str("fake_ip", ip.String()).Msg("fake ip 6")
			return ip.IP(), nil
		}
	}
	return nil, errors.New("cannot create a fake ipv6 for" + domain)
}

// func (h Pools) GetFakeIPForDomain(domain string, ipv4 bool) nethelper.Address {
// 	for _, p := range h {
// 		addr := p.GetFakeIPForDomain(domain, ipv4)
// 		if addr != nil {
// 			return addr
// 		}
// 	}
// 	return nil
// }

type Pool struct {
	domainToIP cache.Lru
	nextIP     *big.Int
	mu         *sync.Mutex
	isV4Pool   bool
	ipRange    *net.IPNet
}

func (fkdns *Pool) IsIPInIPPool(ip nethelper.Address) bool {
	if ip.Family().IsDomain() {
		return false
	}
	return fkdns.ipRange.Contains(ip.IP())
}

func NewPool(cidr string, lruSize int) (*Pool, error) {
	pool := &Pool{}

	err := pool.initialize(cidr, lruSize)
	if err != nil {
		return nil, err
	}

	return pool, nil
}

func (fkdns *Pool) initialize(ipPoolCidr string, lruSize int) error {
	var ipRange *net.IPNet
	var ipaddr net.IP
	var currentIP *big.Int
	var err error
	if ipaddr, ipRange, err = net.ParseCIDR(ipPoolCidr); err != nil {
		return fmt.Errorf("unable to parse CIDR for Fake DNS IP assignment, %w", err)
	}
	if ipaddr.To4() != nil {
		fkdns.isV4Pool = true
	}

	currentIP = big.NewInt(0).SetBytes(ipaddr)
	if ipaddr.To4() != nil {
		currentIP = big.NewInt(0).SetBytes(ipaddr.To4())
	}

	ones, bits := ipRange.Mask.Size()
	rooms := bits - ones
	if math.Log2(float64(lruSize)) >= float64(rooms) {
		return errors.New("LRU size is bigger than subnet size")
	}
	fkdns.domainToIP = cache.NewLru(lruSize)
	fkdns.ipRange = ipRange
	fkdns.nextIP = currentIP
	fkdns.mu = new(sync.Mutex)
	return nil
}

// GetFakeIPForDomain checks if there is already an ip for domain first, if so returns the ip; if not
// generate a new fake IP for domain
func (fkdns *Pool) GetFakeIPForDomain(domain string, ipv4 bool) nethelper.Address {
	if ipv4 != fkdns.isV4Pool {
		return nil
	}
	fkdns.mu.Lock()
	defer fkdns.mu.Unlock()
	if v, ok := fkdns.domainToIP.Get(domain); ok {
		return v.(nethelper.Address)
	}
	var ip nethelper.Address
	for {
		ip = nethelper.IPAddress(fkdns.nextIP.Bytes())

		fkdns.nextIP = fkdns.nextIP.Add(fkdns.nextIP, big.NewInt(1))
		if !fkdns.ipRange.Contains(fkdns.nextIP.Bytes()) {
			fkdns.nextIP = big.NewInt(0).SetBytes(fkdns.ipRange.IP)
		}

		// if we run for a long time, we may go back to beginning and start seeing the IP in use
		// if ok, it means that there is already a domain with that ip
		if _, ok := fkdns.domainToIP.GetKeyFromValue(ip); !ok {
			break
		}
	}
	fkdns.domainToIP.Put(domain, ip)
	return ip
}

// GetDomainFromFakeIP checks if an IP is a fake IP and have corresponding domain name
func (fkdns *Pool) GetDomainFromFakeIP(ip nethelper.Address) string {
	if !fkdns.IsIPInIPPool(ip) {
		return ""
	}
	if k, ok := fkdns.domainToIP.GetKeyFromValue(ip); ok {
		return k.(string)
	}
	return ""
}
