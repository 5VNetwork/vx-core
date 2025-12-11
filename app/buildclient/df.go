package buildclient

import (
	"fmt"
	"reflect"
	"runtime"

	"github.com/5vnetwork/vx-core/app/client"
	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/transport"
)

func DialerFactory(config *configs.TmConfig, fc *Builder, client *client.Client) error {
	// dialer factory
	if config.Tun.GetShouldBindDevice() {
		err := fc.requireFeature(func(bdl i.DefaultInterfaceInfo, ipResolver i.IPResolver) error {
			opt := transport.DialerFactoryOption{
				BindToDefaultNIC: runtime.GOOS != "android",
				// PreventRouteLoop: !config.Tun.GetShouldBindDevice(),
				IpResolver:              ipResolver,
				DefaultInterfaceMonitor: bdl,
			}
			if runtime.GOOS == "android" {
				fdFunc := fc.getFeature(reflect.TypeOf((*transport.FdFunc)(nil)).Elem())
				opt.FdFunc = fdFunc.(transport.FdFunc)
			}

			df := transport.NewDialerFactoryImp(opt)
			client.DialerFactory = df
			return fc.addComponent(df)
		})
		if err != nil {
			return fmt.Errorf("failed to require features: %w", err)
		}
	} else {
		df := transport.DefaultDialerFactory()
		client.DialerFactory = df
		err := fc.addComponent(df)
		if err != nil {
			return fmt.Errorf("failed to add dialer factory: %w", err)
		}
	}

	return nil
}
