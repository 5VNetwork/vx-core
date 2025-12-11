package sysproxy

import (
	"testing"
	"time"

	"github.com/5vnetwork/vx-core/tun"
)

func TestSysProxy(t *testing.T) {
	t.Skip("skipping test")
	mon, _ := tun.NewInterfaceMonitor("")
	mon.Start()
	defer mon.Close()

	proxySetting := &ProxySetting{
		HttpProxySetting: &HttpProxySetting{
			Address: "127.0.0.1",
			Port:    8080,
		},
	}
	sysProxy := NewSysProxy(proxySetting)
	sysProxy.WithMon(mon)
	sysProxy.Start()
	time.Sleep(10 * time.Second)
	sysProxy.Close()
}
