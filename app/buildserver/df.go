package buildserver

import (
	"time"

	vxdialerfactory "buf.build/gen/go/vvvvv/vx/protocolbuffers/go/vx/dialerfactory"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/transport"
	"go.uber.org/fx"
)

type DialerFactoryParams struct {
	fx.In
	DialerFactory *vxdialerfactory.DialerFactoryConfig
	IpResolver    i.IPResolver `name:"internal_resolver"`
}

type DialerFactoryResult struct {
	fx.Out
	DialerFactory transport.DialerFactory
}

func DialerFactoryOption(config *vxdialerfactory.DialerFactoryConfig) fx.Option {
	if config == nil || !config.ShouldBindDevice {
		return fx.Provide(func() transport.DialerFactory {
			return transport.NewDialerFactoryImp(
				transport.DialerFactoryOption{
					DialTimeout: time.Duration(config.GetDialTimeout()) * time.Second,
				},
			)
		})
	}
	if config != nil {
		if config.ShouldBindDevice && !config.ResolveDomain {
			return fx.Provide(func(defaultNICMonitor i.DefaultInterfaceInfo) transport.DialerFactory {
				return transport.NewDialerFactoryImp(
					transport.DialerFactoryOption{
						DefaultInterfaceMonitor: defaultNICMonitor,
						DialTimeout:             time.Duration(config.GetDialTimeout()) * time.Second,
					},
				)
			})
		} else if config.ShouldBindDevice && config.ResolveDomain {
			return fx.Provide(func(ipResolver i.IPResolver, defaultNICMonitor i.DefaultInterfaceInfo) transport.DialerFactory {
				return transport.NewDialerFactoryImp(
					transport.DialerFactoryOption{
						IpResolver:              ipResolver,
						DefaultInterfaceMonitor: defaultNICMonitor,
						DialTimeout:             time.Duration(config.GetDialTimeout()) * time.Second,
					},
				)
			})
		} else if !config.ShouldBindDevice && config.ResolveDomain {
			return fx.Provide(func(ipResolver i.IPResolver) transport.DialerFactory {
				return transport.NewDialerFactoryImp(
					transport.DialerFactoryOption{
						IpResolver:  ipResolver,
						DialTimeout: time.Duration(config.GetDialTimeout()) * time.Second,
					},
				)
			})
		} else {
			return fx.Provide(func() transport.DialerFactory {
				return transport.NewDialerFactoryImp(
					transport.DialerFactoryOption{
						DialTimeout: time.Duration(config.GetDialTimeout()) * time.Second,
					},
				)
			})
		}
	}
	return nil
}
