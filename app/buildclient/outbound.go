// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package buildclient

import (
	"reflect"
	"sync/atomic"

	"github.com/5vnetwork/vx-core/app/client"
	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/dns"
	"github.com/5vnetwork/vx-core/app/handlerfactory"
	"github.com/5vnetwork/vx-core/app/outbound"
	outboundstats "github.com/5vnetwork/vx-core/app/outbound/stats"
	"github.com/5vnetwork/vx-core/app/policy"
	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/serial"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/proxy/freedom"
	"github.com/5vnetwork/vx-core/transport"
	"github.com/rs/zerolog/log"
)

func buildOutbound(config *configs.OutboundConfig, builder *Builder, client *client.Client) (*outbound.Manager, error) {
	om := outbound.NewManager()
	client.OutboundManager = om
	common.Must(builder.addComponent(om))
	err := builder.requireFeature(func(df transport.DialerFactory,
		policy *policy.Policy, _ *dns.HijackDns, outStats *outboundstats.OutStats) error {
		handlerFactory := &handlerfactory.HandlerFactory{
			DialerFactory:               df,
			Policy:                      policy,
			IPResolver:                  client.IPResolver,
			IPResolverForRequestAddress: client.IPResolverForRequestAddress,
			EchResolver:                 client.EchResolver,
			Hysteria2RejectQuic:         config.GetHysteriaRejectQuic(),
			HandlerLinkStats:            config.GetHandlerLinkStats(),
			HandlerMeter:                config.GetHandlerMeter(),
			OutStats:                    outStats,
		}
		if config.GetTotalCounter() {
			handlerFactory.TotalUpCounter = &atomic.Uint64{}
			handlerFactory.TotalDownCounter = &atomic.Uint64{}
		}
		client.HandlerFactory = handlerFactory
		common.Must(builder.addComponent(handlerFactory))

		var handlerConfigs []*configs.HandlerConfig
		for _, handlerConfig := range config.GetOutboundHandlers() {
			handlerConfigs = append(handlerConfigs, &configs.HandlerConfig{
				Type: &configs.HandlerConfig_Outbound{
					Outbound: handlerConfig,
				},
			})
		}
		for _, handlerConfig := range config.GetChainHandlers() {
			handlerConfigs = append(handlerConfigs, &configs.HandlerConfig{
				Type: &configs.HandlerConfig_Chain{
					Chain: handlerConfig,
				},
			})
		}
		handlerConfigs = append(handlerConfigs, config.GetHandlers()...)
		if len(handlerConfigs) == 0 {
			handlerConfigs = []*configs.HandlerConfig{
				{
					Type: &configs.HandlerConfig_Outbound{
						Outbound: &configs.OutboundHandlerConfig{
							Tag:      "direct",
							Protocol: serial.ToTypedMessage(&configs.FreedomConfig{}),
						},
					},
				},
			}
		}

		handlers := make([]i.Outbound, 0, len(handlerConfigs))
		for _, handlerConfig := range handlerConfigs {
			handler, err := handlerFactory.CreateHandler(handlerConfig)
			if err != nil {
				return err
			}
			if _, ok := handler.(*freedom.FreedomHandler); ok {
				nicMonIntf := builder.getFeature(reflect.TypeOf((*i.DefaultInterfaceInfo)(nil)).Elem())
				if nicMonIntf != nil {
					nicMon := nicMonIntf.(i.DefaultInterfaceInfo)
					freedomHandlerWithSupport6Info := outbound.NewHandlerWithSupport6Info(handler,
						nicMon.SupportIPv6() > 0)
					nicMon.Register(i.OnDefaultInterfaceChanged(func() {
						freedomHandlerWithSupport6Info.SetSupport6(nicMon.SupportIPv6() > 0)
						log.Info().Bool("support6", freedomHandlerWithSupport6Info.Support6()).Msg("freedom handler support6 changed")
					}))
					handler = freedomHandlerWithSupport6Info
				}
			}
			if handlerConfig.SupportIpv6 != nil {
				handler = outbound.NewHandlerWithSupport6Info(handler, *handlerConfig.SupportIpv6)
			}
			handlers = append(handlers, handler)
		}
		om.AddHandlers(handlers...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return om, err
}
