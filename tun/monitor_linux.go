package tun

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"os"
	"os/exec"
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
		go info.setSupportIPv6(iffInfo.Name)
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
			Any("defaultDns", m.defaultDns).Msg("nic monitor")
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
			log.Error().Str("defaultInterface", change.New.DefaultRouteInterface).Msg("failed to get primary physical interface")
			return
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
			if t.supportIPv6 == 0 && iffName != "" {
				// t.supportIPv6 = util.NICSupportIPv6(t.defaultInterface)
				go t.setSupportIPv6(iffName)
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

func (t *DefaultInterfaceInfo) setSupportIPv6(name string) {
	supportIPv6 := util.NICSupportIPv6Name(t.defaultInterfaceName)

	t.Lock()
	if t.defaultInterfaceName == name {
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

type InterfaceInfo struct {
	Index int
	Name  string
}

func GetPrimaryPhysicalInterface() (info *InterfaceInfo, err error) {
	state := netmon.NewStatic().InterfaceState()
	if state != nil {
		if !strings.Contains(state.DefaultRouteInterface, "utun") {
			return &InterfaceInfo{
				Index: state.Interface[state.DefaultRouteInterface].Index,
				Name:  state.DefaultRouteInterface}, nil
		} else {
			log.Debug().Str("state.DefaultRouteInterface", state.DefaultRouteInterface).Msg("netmon.DefaultRoute return utun")
		}
	}
	return nil, errors.New("failed to get primary physical interface")
}

func DnsServers(index int) ([]netip.Addr, error) {
	// Get interface by index
	iface, err := net.InterfaceByIndex(index)
	if err != nil {
		return nil, fmt.Errorf("failed to get interface by index %d: %v", index, err)
	}

	// Try multiple methods to get DNS servers for the interface
	// Method 1: Try systemd-resolved (resolvectl)
	if servers, err := getDnsFromSystemdResolved(iface.Name); err == nil && len(servers) > 0 {
		return servers, nil
	}

	// Method 2: Try NetworkManager (nmcli)
	if servers, err := getDnsFromNetworkManager(iface.Name); err == nil && len(servers) > 0 {
		return servers, nil
	}

	// Method 3: Try reading from /etc/resolv.conf as fallback
	if servers, err := getDnsFromResolvConf(); err == nil && len(servers) > 0 {
		return servers, nil
	}

	return nil, fmt.Errorf("failed to get DNS servers for interface %s", iface.Name)
}

// getDnsFromSystemdResolved tries to get DNS servers using systemd-resolved (resolvectl)
func getDnsFromSystemdResolved(ifaceName string) ([]netip.Addr, error) {
	cmd := exec.Command("resolvectl", "status", ifaceName)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("resolvectl failed: %v", err)
	}

	var servers []netip.Addr
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "DNS Servers:") {
			// Extract DNS servers from the line
			dnsPart := strings.TrimPrefix(line, "DNS Servers:")
			dnsPart = strings.TrimSpace(dnsPart)
			if dnsPart == "" {
				continue
			}

			// Split by space and parse each DNS server
			dnsServers := strings.Fields(dnsPart)
			for _, server := range dnsServers {
				if addr, err := netip.ParseAddr(server); err == nil {
					servers = append(servers, addr)
				}
			}
			break
		}
	}

	return servers, nil
}

// getDnsFromNetworkManager tries to get DNS servers using NetworkManager (nmcli)
func getDnsFromNetworkManager(ifaceName string) ([]netip.Addr, error) {
	cmd := exec.Command("nmcli", "device", "show", ifaceName)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("nmcli failed: %v", err)
	}

	var servers []netip.Addr
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "IP4.DNS[") || strings.HasPrefix(line, "IP6.DNS[") {
			// Extract DNS server from the line
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				server := strings.TrimSpace(parts[1])
				if addr, err := netip.ParseAddr(server); err == nil {
					servers = append(servers, addr)
				}
			}
		}
	}

	return servers, nil
}

// getDnsFromResolvConf reads DNS servers from /etc/resolv.conf as fallback
func getDnsFromResolvConf() ([]netip.Addr, error) {
	file, err := os.Open("/etc/resolv.conf")
	if err != nil {
		return nil, fmt.Errorf("failed to open /etc/resolv.conf: %v", err)
	}
	defer file.Close()

	var servers []netip.Addr
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "nameserver") {
			// Extract nameserver IP
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				server := fields[1]
				if addr, err := netip.ParseAddr(server); err == nil {
					servers = append(servers, addr)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading /etc/resolv.conf: %v", err)
	}

	return servers, nil
}
