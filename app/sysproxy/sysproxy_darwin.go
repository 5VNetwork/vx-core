package sysproxy

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"

	"github.com/5vnetwork/vx-core/i"

	"github.com/rs/zerolog/log"
)

// TODO: set proxy on all connected hardware ports? not just the
// one corresponding to the default interface

// Automatically set system proxy to the default interface
type SysProxy struct {
	sync.Mutex
	mon                 i.DefaultInterfaceChangeSubject
	info                i.DefaultInterfaceInfo
	proxySetting        *ProxySetting
	currentHardwarePort string
}

func NewSysProxy(ps *ProxySetting) *SysProxy {
	setting := &SysProxy{
		proxySetting: ps,
	}
	return setting
}

func (s *SysProxy) WithMon(mon i.DefaultInterfaceChangeSubject) {
	s.mon = mon
	s.info = mon.(i.DefaultInterfaceInfo)
}

func (s *SysProxy) Start() error {
	if s.mon == nil || s.proxySetting == nil {
		return errors.New("mon or proxySetting is nil")
	}
	s.mon.Register(s)
	s.OnDefaultInterfaceChanged()
	return nil
}

func (s *SysProxy) Close() error {
	s.Lock()
	defer s.Unlock()
	s.mon.Unregister(s)
	s.setProxyOff(s.currentHardwarePort)
	s.currentHardwarePort = ""
	return nil
}

// TODO: what about ipv6 default interface is different from that of ipv4?
func (s *SysProxy) OnDefaultInterfaceChanged() {
	s.do()
}

func (s *SysProxy) do() {
	s.Lock()
	defer s.Unlock()
	newHardwarePort := deviceNameToHardwarePort(s.info.DefaultInterfaceName4())

	if s.currentHardwarePort != "" && newHardwarePort != s.currentHardwarePort {
		s.setProxyOff(s.currentHardwarePort)
	}

	s.setProxy(newHardwarePort)
	log.Info().Str("current hardware port", newHardwarePort).Send()
	s.currentHardwarePort = newHardwarePort
}

// set proxy of hardwarePort
func (s *SysProxy) setProxy(hardwarePort string) {
	if hardwarePort == "" {
		return
	}
	// http
	if s.proxySetting.HttpProxySetting != nil {
		err := exec.Command("networksetup", "-setwebproxy", hardwarePort,
			s.proxySetting.HttpProxySetting.Address, fmt.Sprintf("%d", s.proxySetting.HttpProxySetting.Port)).
			Run()
		if err != nil {
			log.Error().Err(err).Msgf("failed to set web proxy, %v", err)
		}
	} else {
		err := exec.Command("networksetup", "-setwebproxystate", hardwarePort, "off").Run()
		if err != nil {
			log.Error().Err(err).Msgf("failed to set web proxy off, %v", err)
		}
	}
	// https
	if s.proxySetting.HttpsProxySetting != nil {
		err := exec.Command("networksetup", "-setsecurewebproxy", hardwarePort,
			s.proxySetting.HttpsProxySetting.Address, fmt.Sprintf("%d", s.proxySetting.HttpsProxySetting.Port)).Run()
		if err != nil {
			log.Error().Err(err).Msgf("failed to set secure web proxy, %v", err)
		}
	} else {
		err := exec.Command("networksetup", "-setsecurewebproxystate", hardwarePort, "off").Run()
		if err != nil {
			log.Error().Err(err).Msgf("failed to set secure web proxy off, %v", err)
		}
	}
	// socks
	if s.proxySetting.SocksProxySetting != nil {
		err := exec.Command("networksetup", "-setsocksfirewallproxy", hardwarePort,
			s.proxySetting.SocksProxySetting.Address, fmt.Sprintf("%d", s.proxySetting.SocksProxySetting.Port)).Run()
		if err != nil {
			log.Error().Err(err).Msgf("failed to set socks proxy, %v", err)
		}
	} else {
		err := exec.Command("networksetup", "-setsocksfirewallproxystate", hardwarePort, "off").Run()
		if err != nil {
			log.Error().Err(err).Msgf("failed to set socks proxy off, %v", err)
		}
	}
}

func (s *SysProxy) setProxyOff(hardwarePort string) {
	if hardwarePort == "" {
		return
	}
	err := exec.Command("networksetup", "-setwebproxystate", hardwarePort, "off").Run()
	if err != nil {
		log.Error().Err(err).Msgf("failed to set web proxy off, %v", err)
	}
	err = exec.Command("networksetup", "-setsecurewebproxystate", hardwarePort, "off").Run()
	if err != nil {
		log.Error().Err(err).Msgf("failed to set secure web proxy off, %v", err)
	}
	err = exec.Command("networksetup", "-setsocksfirewallproxystate", hardwarePort, "off").Run()
	if err != nil {
		log.Error().Err(err).Msgf("failed to set socks proxy off, %v", err)
	}
}

func deviceNameToHardwarePort(deviceName string) string {
	if deviceName == "" {
		return ""
	}

	out, err := exec.Command("networksetup", "-listallhardwareports").Output()
	if err != nil {
		log.Error().Err(err).Msg("failed to list hardware ports")
		return ""
	}

	// Output format is:
	// Hardware Port: <port_name>
	// Device: <device_name>
	// ...other info...
	// (repeats)

	lines := strings.Split(string(out), "\n")
	var currentPort string
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if strings.HasPrefix(line, "Hardware Port: ") {
			currentPort = strings.TrimPrefix(line, "Hardware Port: ")
		} else if strings.HasPrefix(line, "Device: ") {
			device := strings.TrimPrefix(line, "Device: ")
			if device == deviceName {
				return currentPort
			}
		}
	}

	log.Warn().Str("device", deviceName).Msg("device not found in hardware ports")
	return ""
}
