//go:build android

package x_android

import (
	"net/netip"
	"sync"

	"github.com/5vnetwork/vx-core/app/util"
	"github.com/5vnetwork/vx-core/common/slices"
	"github.com/5vnetwork/vx-core/tun"
	"github.com/rs/zerolog/log"
)

var monitor = nicMonitor{}

type nicMonitor struct {
	sync.RWMutex
	name       string
	index      int
	defaultDns []netip.Addr
	addresses  []netip.Addr
	// whether actually support ipv6
	supportIpv6 int
	tun.DefaultInterfaceChangeNotifier
}

func (n *nicMonitor) Start() error {
	return nil
}

func (n *nicMonitor) Close() error {
	n.Lock()
	defer n.Unlock()
	n.defaultDns = nil
	n.DefaultInterfaceChangeNotifier = tun.DefaultInterfaceChangeNotifier{}
	return nil
}

// called when default nic changed
func UpdateDefaultNICInfo(name string, addresess StringList, s StringList) error {
	monitor.Lock()
	defer monitor.Unlock()
	var dns []netip.Addr
	for i := range s.Len() {
		ip, err := netip.ParseAddr(s.Get(i))
		if err != nil {
			return err
		}
		dns = append(dns, ip)
	}
	var addresses []netip.Addr
	for i := range addresess.Len() {
		ip, err := netip.ParseAddr(addresess.Get(i))
		if err != nil {
			return err
		}
		addresses = append(addresses, ip)
	}

	changed := false
	if monitor.name != name {
		changed = true
		monitor.name = name
	}
	if !slices.CompareSlices(monitor.defaultDns, dns) {
		changed = true
		monitor.defaultDns = dns
	}
	if !slices.CompareSlices(monitor.addresses, addresses) {
		changed = true
		monitor.addresses = addresses
	}
	newSupportIpv6 := util.NICSupportIPv6Name(name)
	if newSupportIpv6 && monitor.supportIpv6 != 1 {
		changed = true
		monitor.supportIpv6 = 1
	} else if !newSupportIpv6 && monitor.supportIpv6 != -1 {
		changed = true
		monitor.supportIpv6 = -1
	}

	if changed {
		monitor.Notify()
		log.Info().Str("name", name).Any("addresses", addresses).Any("dns", dns).
			Bool("supportIpv6", newSupportIpv6).Msg("default nic changed")
	}

	return nil
}

func (t *nicMonitor) DefaultInterface4() uint32 {
	return 0
}

func (t *nicMonitor) DefaultInterface6() uint32 {
	return 0
}

func (t *nicMonitor) DefaultInterfaceName4() string {
	t.RLock()
	defer t.RUnlock()
	return t.name
}

func (t *nicMonitor) DefaultInterfaceName6() string {
	t.RLock()
	defer t.RUnlock()
	return t.name
}

func (t *nicMonitor) DefaultDns4() []netip.Addr {
	t.RLock()
	defer t.RUnlock()
	return t.defaultDns
}

func (t *nicMonitor) DefaultDns6() []netip.Addr {
	t.RLock()
	defer t.RUnlock()
	return t.defaultDns
}

func (t *nicMonitor) SupportIPv6() int {
	t.RLock()
	defer t.RUnlock()
	return t.supportIpv6
}

func (t *nicMonitor) HasGlobalIPv6() (bool, error) {
	t.RLock()
	defer t.RUnlock()
	for _, addr := range t.addresses {
		if addr.Is6() && !addr.Is4In6() && addr.IsGlobalUnicast() {
			return true, nil
		}
	}
	return false, nil
}
