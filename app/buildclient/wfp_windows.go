package buildclient

import (
	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/dispatcher"
	"github.com/5vnetwork/vx-core/app/dns"
	"github.com/5vnetwork/vx-core/app/inbound/wfp"
)

func Wfp(config *configs.WfpConfig, f *Builder) error {
	return f.requireFeature(func(d *dispatcher.Dispatcher, dns *dns.Dns) error {
		w := wfp.NewWfpHandler().WithDispatcher(d).
			WithDns(dns, nil).WithTcpPort(uint16(config.TcpPort)).
			WithUdpPort(uint16(config.UdpPort))
		return f.addComponent(w)
	})
}
