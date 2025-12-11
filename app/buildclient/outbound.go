package buildclient

import (
	"reflect"

	"github.com/5vnetwork/vx-core/app/client"
	"github.com/5vnetwork/vx-core/app/configs"
	proxyconfigs "github.com/5vnetwork/vx-core/app/configs/proxy"
	"github.com/5vnetwork/vx-core/app/dns"
	"github.com/5vnetwork/vx-core/app/outbound"
	"github.com/5vnetwork/vx-core/app/policy"
	"github.com/5vnetwork/vx-core/app/util"
	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/serial"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/proxy/freedom"
	"github.com/5vnetwork/vx-core/transport"
	"github.com/rs/zerolog/log"
)

func buildOutbound(config *configs.TmConfig, builder *Builder, client *client.Client) (*outbound.Manager, error) {
	om := outbound.NewManager()
	client.OutboundManager = om
	common.Must(builder.addComponent(om))
	err := builder.requireFeature(func(df transport.DialerFactory,
		policy *policy.Policy, ipr i.IPResolver) error {
		var singleHandlers []*configs.HandlerConfig
		var chainHandlers []*configs.HandlerConfig

		for _, handlerConfig := range config.Outbound.GetHandlers() {
			if handlerConfig.GetOutbound() != nil {
				singleHandlers = append(singleHandlers, handlerConfig)
			} else if handlerConfig.GetChain() != nil {
				chainHandlers = append(chainHandlers, handlerConfig)
			}
		}
		for _, handlerConfig := range config.Outbound.GetOutboundHandlers() {
			singleHandlers = append(singleHandlers, &configs.HandlerConfig{
				Type: &configs.HandlerConfig_Outbound{
					Outbound: handlerConfig,
				},
			})
		}
		if len(singleHandlers) == 0 && len(chainHandlers) == 0 {
			singleHandlers = append(singleHandlers, &configs.HandlerConfig{
				Type: &configs.HandlerConfig_Outbound{
					Outbound: &configs.OutboundHandlerConfig{
						Protocol: serial.ToTypedMessage(&proxyconfigs.FreedomConfig{}),
					},
				},
			})
		}
		handlers := make([]i.Outbound, 0, len(singleHandlers)+len(chainHandlers))
		for _, handlerConfig := range singleHandlers {
			handler, err := outbound.NewOutHandler(&outbound.Config{
				OutboundHandlerConfig:       handlerConfig.GetOutbound(),
				DialerFactory:               df,
				Policy:                      policy,
				IPResolver:                  client.IPResolver,
				IPResolverForRequestAddress: client.IPResolverForRequestAddress,
				RejectQuic:                  config.Hysteria2RejectQuic,
			})
			if err != nil {
				return err
			}
			if _, ok := handler.(*freedom.FreedomHandler); ok {
				nicMonIntf := builder.getFeature(reflect.TypeOf((*i.DefaultInterfaceInfo)(nil)).Elem())
				if nicMonIntf != nil {
					nicMon := nicMonIntf.(i.DefaultInterfaceInfo)
					freedomHandlerWithSupport6Info := &outbound.HandlerWithSupport6Info{
						Outbound:                  handler,
						IPv6SupportChangeNotifier: util.IPv6SupportChangeNotifier{},
					}
					freedomHandlerWithSupport6Info.SetSupport6(nicMon.SupportIPv6() > 0)
					nicMon.Register(i.OnDefaultInterfaceChanged(func() {
						freedomHandlerWithSupport6Info.SetSupport6(nicMon.SupportIPv6() > 0)
						log.Info().Bool("support6", freedomHandlerWithSupport6Info.Support6()).Msg("freedom handler support6 changed")
					}))
					handler = freedomHandlerWithSupport6Info
				}
			}
			if handlerConfig.SupportIpv6 != nil {
				handler = &outbound.HandlerWithSupport6Info{
					Outbound: handler,
				}
				handler.(*outbound.HandlerWithSupport6Info).SetSupport6(*handlerConfig.SupportIpv6)
			}
			handlers = append(handlers, handler)
		}
		for _, chainHandlerConfig := range chainHandlers {
			chainHandler, err := outbound.NewChainHandler(&outbound.ChainHandlerConfig{
				ChainHandlerConfig:          chainHandlerConfig.GetChain(),
				Policy:                      policy,
				IPResolver:                  client.IPResolver,
				IPResolverForRequestAddress: client.IPResolverForRequestAddress,
				DF:                          df,
				RejectQuic:                  config.Hysteria2RejectQuic,
				DnsServer:                   dns.NewDnsServerToResolver(&dns.DnsToDnsServer{Dns: client.Dns}),
			})
			if err != nil {
				return err
			}
			handlers = append(handlers, chainHandler)
		}

		om.AddHandlers(handlers...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return om, err
}
