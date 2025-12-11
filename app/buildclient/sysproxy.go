package buildclient

import (
	configs "github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/sysproxy"
	"github.com/5vnetwork/vx-core/i"
)

func NewSysProxy(cfg *configs.SysProxyConfig, f *Builder) error {
	ps := &sysproxy.ProxySetting{}
	if cfg.HttpProxyAddress != "" {
		ps.HttpProxySetting = &sysproxy.HttpProxySetting{
			Address: cfg.HttpProxyAddress,
			Port:    uint16(cfg.HttpProxyPort),
		}
	}
	if cfg.HttpsProxyAddress != "" {
		ps.HttpsProxySetting = &sysproxy.HttpsProxySetting{
			Address: cfg.HttpsProxyAddress,
			Port:    uint16(cfg.HttpsProxyPort),
		}
	}
	if cfg.SocksProxyAddress != "" {
		ps.SocksProxySetting = &sysproxy.SocksProxySetting{
			Address: cfg.SocksProxyAddress,
			Port:    uint16(cfg.SocksProxyPort),
		}
	}
	sys := sysproxy.NewSysProxy(ps)
	f.requireFeature(func(mon i.DefaultInterfaceChangeSubject) {
		sys.WithMon(mon)
	})
	return f.addComponent(sys)
}
