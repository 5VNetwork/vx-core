package tun

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"os/exec"
	"runtime"
	"strings"

	sync "sync"

	"github.com/5vnetwork/vx-core/app/util"
	"github.com/5vnetwork/vx-core/common/slices"
	"github.com/5vnetwork/vx-core/tun/netmon"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"tailscale.com/types/logger"
)

// TODO: this package assumes that the default interface for ipv4 and ipv6 are the same

// implements i.DefaultInterfaceInfo
// Okay to not start.
type DefaultInterfaceInfo struct {
	sync.RWMutex
	tunName              string
	logf                 logger.Logf
	monitor              *netmon.Monitor
	state                *netmon.State
	defaultInterface     uint32
	defaultInterfaceName string
	defaultDns           []netip.Addr
	supportIPv6          int
	DefaultInterfaceChangeNotifier
}

func NewInterfaceMonitor(tunName string) (*DefaultInterfaceInfo, error) {
	info := &DefaultInterfaceInfo{
		tunName: tunName,
		logf: func(format string, args ...any) {
			// log.Printf(format, args...)
		},
	}
	// TODO this might failed. use another way
	iffInfo, err := GetPrimaryPhysicalInterface()
	if err == nil {
		info.defaultInterfaceName = iffInfo.Name
		info.defaultInterface = uint32(iffInfo.Index)
		dnsServers, err := DnsServers(int(iffInfo.Index))
		if err == nil {
			info.defaultDns = dnsServers
		}
		go info.setSupportIPv6(int(iffInfo.Index))
		info.log()
	} else {
		log.Err(err).Msg("netmon failed to get primary physical interface")
	}
	return info, nil
}

func (m *DefaultInterfaceInfo) log() {
	if zerolog.GlobalLevel() <= zerolog.InfoLevel {
		iff, err := net.InterfaceByIndex(int(m.defaultInterface))
		if err == nil {
			addrs, err := iff.Addrs()
			if err == nil {
				l := log.Info()
				for i, addr := range addrs {
					l.Str(fmt.Sprintf("addr%d", i), addr.String())
				}
				l.Msg("default interface addresses")
			}
		}
		m.RLock()
		defer m.RUnlock()
		log.Info().Str("tunName", m.tunName).
			Str("name", m.defaultInterfaceName).
			Uint32("index", m.defaultInterface).
			Int("supportIPv6", m.supportIPv6).
			Any("defaultDns", m.defaultDns).Send()
	}
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

func (t *DefaultInterfaceInfo) Start() error {
	mon, err := netmon.New(t.logf)
	if err != nil {
		return err
	}
	mon.SetTailscaleInterfaceName(t.tunName)
	t.monitor = mon
	t.monitor.RegisterChangeCallback(func(change *netmon.ChangeDelta) {
		log.Info().Msg("default interface changed")
		// populate default interface info
		var iffIndex int
		var iffName string

		if defaultInterface, f := change.New.Interface[change.New.DefaultRouteInterface]; f &&
			!strings.HasPrefix(defaultInterface.Name, t.tunName) {
			// log.Debug().Str("tunName", t.tunName).Str("defaultInterface", defaultInterface.Name).Msg("default interface")
			iffIndex = defaultInterface.Index
			iffName = defaultInterface.Name
		} else {
			iffInfo, err := GetPrimaryPhysicalInterface()
			if err != nil {
				log.Err(err).Msg("failed to get primary physical interface")
				return
			} else {
				iffIndex = iffInfo.Index
				iffName = iffInfo.Name
			}
		}
		// TODO: When default interfaces for v4 and v6 are different, there will be a problem:
		// the returned defaultInterface might be for v4, but it might be set to v6
		t.Lock()
		t.state = change.New

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
		if iffIndex != 0 {
			servers, err = DnsServers(int(iffIndex))
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
				// t.supportIPv6 = util.NICSupportIPv6(t.defaultInterface)
				go t.setSupportIPv6(iffIndex)
			}
			t.log()
			t.Notify()
		}

	})
	t.monitor.Start()
	t.monitor.InjectEvent()
	return nil
}

func (t *DefaultInterfaceInfo) Close() error {
	if t.monitor != nil {
		return t.monitor.Close()
	}
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

type IOSInterfaceDnsServersGetter func(string) ([]netip.Addr, error)

var IosInterfaceGetter IOSInterfaceDnsServersGetter

// returns dns servers of an interface of [index]
func DnsServers(index int) ([]netip.Addr, error) {
	// Get interface by index
	iface, err := net.InterfaceByIndex(index)
	if err != nil {
		return nil, fmt.Errorf("failed to get interface by index %d: %v", index, err)
	}
	if runtime.GOOS == "ios" {
		if IosInterfaceGetter != nil {
			return IosInterfaceGetter(iface.Name)
		}
		return nil, errors.New("IosInterfaceGetter not set")
	}

	// Get DNS servers using scutil
	cmd := exec.Command("scutil", "--dns")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get DNS configuration: %v", err)
	}

	var servers []netip.Addr
	blocks := strings.Split(string(output), "\n\n")

	for _, block := range blocks {
		if !strings.Contains(block, fmt.Sprintf("if_index : %d (%s)", index, iface.Name)) {
			continue
		}
		lines := strings.Split(block, "\n")
		for _, line := range lines {
			if strings.Contains(line, "nameserver[") {
				parts := strings.Split(line, "] : ")
				if len(parts) != 2 {
					continue
				}
				serverStr := strings.TrimSpace(parts[1])
				addr, err := netip.ParseAddr(serverStr)
				if err != nil {
					continue
				}
				servers = append(servers, addr)
			}
		}
		break
	}

	return servers, nil
}

type InterfaceInfo struct {
	Index int
	Name  string
}

func GetPrimaryPhysicalInterface() (info *InterfaceInfo, err error) {
	//  netmon.DefaultRoute first
	route, err := netmon.DefaultRoute()
	if err == nil {
		if !strings.Contains(route.InterfaceName, "utun") {
			return &InterfaceInfo{Index: route.InterfaceIndex, Name: route.InterfaceName}, nil
		} else {
			log.Debug().Str("route.InterfaceName", route.InterfaceName).Msg("netmon.DefaultRoute return utun")
		}
	} else {
		// log.Debug().Err(err).Msg("netmon.DefaultRoute failed")
	}
	// route -n get 169.254.0.1
	info, err = GetPrimaryPhysicalInterface1()
	if err == nil {
		if !strings.Contains(info.Name, "utun") {
			return info, nil
		} else {
			log.Debug().Str("info.Name", info.Name).Msg("GetPrimaryPhysicalInterface1 return utun")
		}
	} else {
		// log.Debug().Err(err).Msg("GetPrimaryPhysicalInterface1 failed")
	}
	//  netstat -nr
	idx, err := DefaultRouteInterfaceIndex()
	if err == nil {
		iface, err := net.InterfaceByIndex(idx)
		if err == nil && !strings.Contains(iface.Name, "utun") {
			return &InterfaceInfo{Index: idx, Name: iface.Name}, nil
		} else {
			// log.Debug().Int("idx", idx).Str("name", iface.Name).Msg("DefaultRouteInterfaceIndex return utun")
		}
	} else {
		// log.Err(err).Msg("DefaultRouteInterfaceIndex failed")
	}
	// fallback to networksetup
	n, err := GetPrimaryPhysicalInterface0()
	if err == nil {
		iface, err := net.InterfaceByName(n)
		if err == nil && !strings.Contains(iface.Name, "utun") {
			return &InterfaceInfo{Index: iface.Index, Name: n}, nil
		}
	} else {
		// log.Err(err).Msg("GetPrimaryPhysicalInterface0 failed")
	}
	return nil, errors.New("no primary physical interface found")
}

func GetPrimaryPhysicalInterface0() (string, error) {
	// Run networksetup -listallhardwareports to get interface info
	cmd := exec.Command("networksetup", "-listallhardwareports")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to list hardware ports: %v", err)
	}
	// Parse output to find primary interface
	lines := strings.Split(string(output), "\n")
	var deviceName string
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if strings.HasPrefix(line, "Hardware Port:") {
			// Check if this is a physical interface (Wi-Fi or Ethernet)
			if strings.Contains(line, "Wi-Fi") || strings.Contains(line, "Ethernet") {
				// Get the device name from next line
				if i+1 < len(lines) {
					deviceLine := strings.TrimSpace(lines[i+1])
					if strings.HasPrefix(deviceLine, "Device:") {
						deviceName = strings.TrimSpace(strings.TrimPrefix(deviceLine, "Device:"))
						// Get interface status
						statusCmd := exec.Command("ifconfig", deviceName)
						statusOutput, err := statusCmd.Output()
						if err == nil && strings.Contains(string(statusOutput), "status: active") {
							return deviceName, nil
						}
					}
				}
			}
		}
	}
	if deviceName == "" {
		return "", errors.New("no active physical interface found")
	}
	return deviceName, nil
}

// This function assumes 169.254.0.1 does not route to utun
func GetPrimaryPhysicalInterface1() (*InterfaceInfo, error) {
	// Get the default route interface using "route" command
	cmd := exec.Command("route", "-n", "get", "169.254.0.1")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get default route: %v", err)
	}

	// Parse the output to find interface
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "interface:") {
			ifaceName := strings.TrimSpace(strings.TrimPrefix(line, "interface:"))
			// Verify it's a physical interface
			iface, err := net.InterfaceByName(ifaceName)
			if err != nil {
				return nil, fmt.Errorf("failed to get interface %s: %v", ifaceName, err)
			}
			// Check if it's a physical interface (not loopback, not virtual)
			if iface.Flags&net.FlagLoopback == 0 && iface.Flags&net.FlagUp != 0 {
				return &InterfaceInfo{Index: iface.Index, Name: iface.Name}, nil
			}
		}
	}

	return nil, errors.New("GetPrimaryPhysicalInterface1 failed")
}
