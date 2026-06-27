package create

import (
	"time"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/proxy/hysteria2"
)

func HysteriaRealmConfig(cfg *configs.RealmConfig, resolver i.IPResolver) hysteria2.RealmConfig {
	if cfg == nil {
		return hysteria2.RealmConfig{}
	}
	out := hysteria2.RealmConfig{
		RealmAddr: cfg.GetRealmAddr(),
		LocalPort: uint16(cfg.GetLocalPort()),
		Insecure:  cfg.GetInsecure(),
		Resolver:  resolver,
	}
	out.STUNServers = append([]string(nil), cfg.GetStunServers()...)
	if sec := cfg.GetStunTimeout(); sec != 0 {
		out.STUNTimeout = time.Duration(sec) * time.Second
	}
	if sec := cfg.GetPunchTimeout(); sec != 0 {
		out.PunchTimeout = time.Duration(sec) * time.Second
	}
	if sec := cfg.GetHeartbeatInterval(); sec != 0 {
		out.HeartbeatInterval = time.Duration(sec) * time.Second
	}
	if cfg.GetInsecure() {
		out.Insecure = true
	}
	out.IPMode = cfg.GetIpMode()
	if pm := cfg.GetPortMapping(); pm != nil {
		out.PortMapping = hysteria2.RealmPortMappingConfig{
			Enabled: pm.GetEnabled(),
		}
		if sec := pm.GetTimeout(); sec != 0 {
			out.PortMapping.Timeout = time.Duration(sec) * time.Second
		}
		if sec := pm.GetLifetime(); sec != 0 {
			out.PortMapping.Lifetime = time.Duration(sec) * time.Second
		}
	}
	return out
}
