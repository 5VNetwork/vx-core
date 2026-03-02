// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package x_darwin

import (
	"errors"
	"fmt"
	"net"
	"net/netip"

	sync "sync"

	"github.com/5vnetwork/vx-core/app/util"
	"github.com/5vnetwork/vx-core/common/slices"
	"github.com/5vnetwork/vx-core/nic"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var nicMon DefaultInterfaceInfo

// implements i.DefaultInterfaceInfo
// Okay to not start.
type DefaultInterfaceInfo struct {
	sync.RWMutex
	defaultInterface     uint32
	defaultInterfaceName string
	defaultDns           []netip.Addr
	supportIPv6          int
	nic.DefaultInterfaceChangeNotifier
}

func setDefaultInterfaceInfo() {
	iffInfo, err := nic.GetPrimaryPhysicalInterface()
	if err == nil && nicMon.DefaultInterface4() == 0 {
		nicMon.SetDefaultInterface(int(iffInfo.Index), iffInfo.Name)
	} else {
		log.Err(err).Msg("netmon failed to get primary physical interface")
	}
}

func (m *DefaultInterfaceInfo) log() {
	m.RLock()
	l := log.Debug().Str("name", m.defaultInterfaceName).
		Uint32("index", m.defaultInterface).
		Int("supportIPv6", m.supportIPv6).
		Any("defaultDns", m.defaultDns).Func(func(e *zerolog.Event) {
		iff, err := net.InterfaceByIndex(int(m.defaultInterface))
		if err == nil {
			addrs, err := iff.Addrs()
			if err == nil {
				for i, addr := range addrs {
					e.Str(fmt.Sprintf("addr%d", i), addr.String())
				}
			}
		}
	})
	m.RUnlock()
	l.Msg("default interface info")
}

func (t *DefaultInterfaceInfo) DefaultInterface4() uint32 {
	t.RLock()
	defer t.RUnlock()
	return t.defaultInterface
}

func (t *DefaultInterfaceInfo) DefaultInterface6() uint32 {
	t.RLock()
	defer t.RUnlock()
	return t.defaultInterface
}

func (t *DefaultInterfaceInfo) DefaultInterfaceName4() string {
	t.RLock()
	defer t.RUnlock()
	return t.defaultInterfaceName
}

func (t *DefaultInterfaceInfo) DefaultInterfaceName6() string {
	t.RLock()
	defer t.RUnlock()
	return t.defaultInterfaceName
}

func (t *DefaultInterfaceInfo) DefaultDns4() []netip.Addr {
	t.RLock()
	defer t.RUnlock()
	return t.defaultDns
}

func (t *DefaultInterfaceInfo) DefaultDns6() []netip.Addr {
	t.RLock()
	defer t.RUnlock()
	return t.defaultDns
}

func (t *DefaultInterfaceInfo) SupportIPv6() int {
	return t.supportIPv6
}

func (t *DefaultInterfaceInfo) SetDefaultInterface(iffIndex int, iffName string) {
	log.Debug().Int("index", iffIndex).Str("name", iffName).Msg("SetDefaultInterface")

	t.Lock()
	changed := false

	if t.defaultInterface != uint32(iffIndex) {
		changed = true
		t.supportIPv6 = 0
		t.defaultInterface = uint32(iffIndex)
	}
	if t.defaultInterfaceName != iffName {
		changed = true
		t.supportIPv6 = 0
		t.defaultInterfaceName = iffName
	}
	// set dns servers
	var servers []netip.Addr
	var err error
	if iffIndex != 0 {
		servers, err = nic.DnsServers(int(iffIndex))
		if err != nil {
			log.Err(err).Msg("failed to get dns servers")
		}
	}
	if !slices.CompareSlices(t.defaultDns, servers) {
		changed = true
		t.defaultDns = servers
	}
	t.Unlock()

	if changed {
		if t.supportIPv6 == 0 && iffIndex != 0 {
			go t.setSupportIPv6(iffIndex)
		}
		t.log()
		t.Notify()
	}
}

func (t *DefaultInterfaceInfo) Close() error {
	return nil
}

func (t *DefaultInterfaceInfo) setSupportIPv6(index int) {
	supportIPv6 := util.NICSupportIPv6Index(uint32(index))

	t.Lock()
	if t.defaultInterface == uint32(index) {
		if supportIPv6 {
			t.supportIPv6 = 1
		} else {
			t.supportIPv6 = -1
		}
		t.Unlock()
		t.log()
		t.Notify()
	} else {
		t.Unlock()
	}
}

func (t *DefaultInterfaceInfo) HasGlobalIPv6() (bool, error) {
	t.RLock()
	index := t.defaultInterface
	if t.supportIPv6 > 0 {
		t.RUnlock()
		return true, nil
	}
	t.RUnlock()

	if index == 0 {
		return false, errors.New("default interface unknown")
	}

	has, err := util.NICHasGlobalIPv6Address(index)
	if err != nil {
		return false, err
	}
	return has, nil
}
