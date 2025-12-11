package api

import (
	context "context"

	"github.com/5vnetwork/vx-core/app/sysproxy"
)

// This function is for Mac only
func (a *Api) StartMacSystemProxy(ctx context.Context, req *StartMacSystemProxyRequest) (*Receipt, error) {
	setting := &sysproxy.ProxySetting{}
	if req.HttpsProxyAddress != "" {
		setting.HttpsProxySetting = &sysproxy.HttpsProxySetting{
			Address: req.HttpsProxyAddress,
			Port:    uint16(req.HttpsProxyPort),
		}
	}
	if req.HttpProxyAddress != "" {
		setting.HttpProxySetting = &sysproxy.HttpProxySetting{
			Address: req.HttpProxyAddress,
			Port:    uint16(req.HttpProxyPort),
		}
	}
	if req.SocksProxyAddress != "" {
		setting.SocksProxySetting = &sysproxy.SocksProxySetting{
			Address: req.SocksProxyAddress,
			Port:    uint16(req.SocksProxyPort),
		}
	}
	a.sysProxy = sysproxy.NewSysProxy(setting)

	a.sysProxy.WithMon(a.mon)
	a.sysProxy.Start()

	return &Receipt{}, nil
}

func (a *Api) StopMacSystemProxy(ctx context.Context, req *StopMacSystemProxyRequest) (*Receipt, error) {
	a.sysProxy.Close()
	a.sysProxy = nil
	return &Receipt{}, nil
}
