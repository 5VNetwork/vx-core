//go:build windows

package tun

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"strings"
	"sync"
	"time"

	"github.com/5vnetwork/vx-core/app/util"
	"github.com/rs/zerolog/log"
	"golang.org/x/sys/windows"
	"golang.zx2c4.com/wireguard/windows/tunnel/winipcfg"
)

// golang.zx2c4.com/wireguard/windows/tunnel/winipcfg
// Monitor changes of the default primary interface
type DefaultInterfaceMonitor struct {
	sync.RWMutex
	// functions used to unregister change callbacks
	// called in Close()
	unregisters []ChangeCallbackUnregister
	tunName     string
	name4       string
	idx4        uint32
	idx6        uint32
	name6       string
	// dns nameservers of the default interface
	dnsAddrs4 []netip.Addr
	dnsAddrs6 []netip.Addr
	// idx6 actually support ipv6
	supportIPv6 int
	DefaultInterfaceChangeNotifier
}

func NewInterfaceMonitor(tunName string) (*DefaultInterfaceMonitor, error) {
	m := &DefaultInterfaceMonitor{
		tunName: tunName,
	}

	err := m.update()
	if err != nil {
		return nil, err
	}

	cbr, err := winipcfg.RegisterRouteChangeCallback(func(notificationType winipcfg.MibNotificationType, route *winipcfg.MibIPforwardRow2) {
		if route != nil && route.DestinationPrefix.PrefixLength == 0 {
			err := m.update()
			if err != nil {
				log.Error().Err(err).Msg("failed to update default interface")
			}
		}
	})
	if err != nil {
		return nil, fmt.Errorf("failed to register route change callback: %w", err)
	}
	cbi, err := winipcfg.RegisterInterfaceChangeCallback(func(notificationType winipcfg.MibNotificationType, iface *winipcfg.MibIPInterfaceRow) {
		if notificationType == winipcfg.MibParameterNotification {
			err := m.update()
			if err != nil {
				log.Error().Err(err).Msg("failed to update default interface")
			}
		}
	})
	if err != nil {
		cbr.Unregister()
		return nil, fmt.Errorf("failed to register interface change callback: %w", err)
	}
	m.unregisters = []ChangeCallbackUnregister{cbr, cbi}
	return m, nil
}

func (m *DefaultInterfaceMonitor) Start() error {
	return nil
}

func (m *DefaultInterfaceMonitor) SetTunName(tunName string) {
	m.Lock()
	m.tunName = tunName
	m.Unlock()
	m.update()
}

func (m *DefaultInterfaceMonitor) SupportIPv6() int {
	m.RLock()
	defer m.RUnlock()
	return m.supportIPv6
}

func (m *DefaultInterfaceMonitor) update() error {
	var myTunIndex int
	if m.tunName == "" {
		myTunIndex = 0
	} else {
		i, err := net.InterfaceByName(m.tunName)
		if err != nil {
			if !strings.Contains(err.Error(), "no such network interface") {
				return fmt.Errorf("failed to get interface by its name %s: %w", m.tunName, err)
			}
		} else {
			myTunIndex = i.Index
		}
	}

	n4, l, idx4, err := FindDefaultLUID(windows.AF_INET, myTunIndex)
	if err != nil {
		return fmt.Errorf("failed to find default IPv4 interface: %w", err)
	}
	m.Lock()
	var shouldNotify bool
	if n4 != m.name4 || idx4 != m.idx4 {
		log.Info().Msgf("default interface 4 changed: %s -> %s, %d -> %d", m.name4, n4, m.idx4, idx4)
		m.name4 = n4
		m.idx4 = idx4
		dnsAddrs, err := l.DNS()
		if err != nil {
			m.Unlock()
			return fmt.Errorf("failed to get all DNS nameservers for interface %s: %w", n4, err)
		}
		m.dnsAddrs4 = dnsAddrs
		shouldNotify = true
	}
	n6, l, idx6, err := FindDefaultLUID(windows.AF_INET6, myTunIndex)
	if err != nil {
		m.Unlock()
		return fmt.Errorf("failed to find default IPv6 interface: %w", err)
	}
	if n6 != m.name6 || idx6 != m.idx6 {
		log.Info().Msgf("default interface 6 changed: %s -> %s, %d -> %d", m.name6, n6, m.idx6, idx6)
		m.name6 = n6
		m.idx6 = idx6
		dnsAddrs, err := l.DNS()
		if err != nil {
			m.Unlock()
			return fmt.Errorf("failed to get all DNS nameservers for interface %s: %w", n6, err)
		}
		m.dnsAddrs6 = dnsAddrs
		shouldNotify = true
		if m.idx6 != 0 {
			m.supportIPv6 = 0
			go func() {
				// wait for route changes when tun ipv6 depends on default nic
				time.Sleep(1 * time.Second)
				m.setSupportIPv6(m.idx6)
			}()
		} else {
			m.supportIPv6 = -1
		}
	}
	m.Unlock()
	if shouldNotify {
		m.log()
		m.Notify()
	}
	return nil
}

func (t *DefaultInterfaceMonitor) HasGlobalIPv6() (bool, error) {
	t.RLock()
	index6 := t.idx6
	if index6 == 0 {
		t.RUnlock()
		return false, nil
	}
	if t.supportIPv6 > 0 {
		t.RUnlock()
		return true, nil
	}
	t.RUnlock()

	has, err := util.NICHasGlobalIPv6Address(index6)
	if err != nil {
		return false, err
	}
	return has, nil
}

func (m *DefaultInterfaceMonitor) setSupportIPv6(index uint32) {
	supportIPv6 := util.NICSupportIPv6Index(index)

	m.Lock()
	if m.idx6 == index {
		if supportIPv6 {
			m.supportIPv6 = 1
		} else {
			m.supportIPv6 = -1
		}
		m.log()
		m.Notify()
	}
	m.Unlock()
}

func (m *DefaultInterfaceMonitor) log() {
	log.Info().
		Str("4Name", m.name4).
		Uint32("4Index", m.idx4).
		Any("defaultDns4", m.dnsAddrs4).
		Str("6Name", m.name6).
		Uint32("6Index", m.idx6).
		Int("supportIPv6", m.supportIPv6).
		Any("defaultDns6", m.dnsAddrs6).Msg("default nic changed")
}

func (m *DefaultInterfaceMonitor) DefaultInterface4() uint32 {
	m.RLock()
	defer m.RUnlock()
	return m.idx4
}

func (m *DefaultInterfaceMonitor) DefaultInterface6() uint32 {
	m.RLock()
	defer m.RUnlock()
	return m.idx6
}

func (m *DefaultInterfaceMonitor) DefaultInterfaceName4() string {
	m.RLock()
	defer m.RUnlock()
	return m.name4
}

func (m *DefaultInterfaceMonitor) DefaultInterfaceName6() string {
	m.RLock()
	defer m.RUnlock()
	return m.name6
}

// func (m *DefaultInterfaceMonitor) DefaultInterface6() uint32 {
// 	m.RLock()
// 	defer m.RUnlock()
// 	return m.idx6
// }

// func (m *DefaultInterfaceMonitor) DefaultInterface6Name() string {
// 	m.RLock()
// 	defer m.RUnlock()
// 	return m.name6
// }

func (m *DefaultInterfaceMonitor) DefaultDns4() []netip.Addr {
	m.RLock()
	defer m.RUnlock()
	return m.dnsAddrs4
}

func (m *DefaultInterfaceMonitor) DefaultDns6() []netip.Addr {
	m.RLock()
	defer m.RUnlock()
	return m.dnsAddrs6
}

/*
https://learn.microsoft.com/en-us/windows/win32/api/netioapi/nf-netioapi-notifyroutechange2
If the application terminates, the system will automatically deregister
any registration for change notifications. It is still recommended that an
application explicitly deregister for change notifications before it terminates.
*/
func (m *DefaultInterfaceMonitor) Close() error {
	m.Lock()
	registers := m.unregisters
	m.Unlock()
	log.Debug().Msg("close NICMonitor")
	var err error
	for _, cb := range registers {
		log.Debug().Type("type", cb).Msg("removing cb")
		err = errors.Join(err, cb.Unregister())
	}
	log.Debug().Msg("closed NICMonitor")
	return nil
}

// find the default interface with the lowest metric that is not the interface with index idx
func FindDefaultLUID(family winipcfg.AddressFamily, idx int) (string, winipcfg.LUID, uint32, error) {
	r, err := winipcfg.GetIPForwardTable2(family)
	if err != nil {
		return "", 0, 0, err
	}
	lowestMetric := ^uint32(0)
	index := uint32(0)
	luid := winipcfg.LUID(0)
	name := ""
	for i := range r {
		if r[i].DestinationPrefix.PrefixLength != 0 || r[i].InterfaceIndex == uint32(idx) {
			continue
		}
		ifrow, err := r[i].InterfaceLUID.Interface()
		if err != nil || ifrow.OperStatus != winipcfg.IfOperStatusUp {
			continue
		}

		iface, err := r[i].InterfaceLUID.IPInterface(family)
		if err != nil {
			continue
		}

		if r[i].Metric+iface.Metric < lowestMetric {
			lowestMetric = r[i].Metric + iface.Metric
			index = r[i].InterfaceIndex
			luid = r[i].InterfaceLUID
			name = ifrow.Alias()
		}
	}

	return name, luid, index, nil
}
