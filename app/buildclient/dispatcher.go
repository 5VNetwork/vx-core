// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package buildclient

import (
	"fmt"
	"time"

	"github.com/5vnetwork/vx-core/app/client"
	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/dispatcher"
	"github.com/5vnetwork/vx-core/app/dispatcher/hooks"
	"github.com/5vnetwork/vx-core/app/dns"
	idns "github.com/5vnetwork/vx-core/app/dns"
	"github.com/5vnetwork/vx-core/app/handlerfactory"
	"github.com/5vnetwork/vx-core/app/outbound"
	"github.com/5vnetwork/vx-core/app/policy"
	"github.com/5vnetwork/vx-core/app/router"
	"github.com/5vnetwork/vx-core/app/router/selector"
	"github.com/5vnetwork/vx-core/app/sniff"
	"github.com/5vnetwork/vx-core/app/tester"
	"github.com/5vnetwork/vx-core/app/userlogger"
	"github.com/5vnetwork/vx-core/app/xsqlite"
	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/i"
)

func Handler(config *configs.TmConfig, fc *Builder, cc *client.Client) error {
	selectors := selector.NewSelectors()
	common.Must(fc.addComponent(selectors))
	cc.Selectors = selectors

	d := &dispatcher.Dispatcher{
		FallbackTimeout:     time.Duration(config.GetDispatcher().GetFallbackTimeout()) * time.Second,
		SessionStats:        config.GetDispatcher().GetSessionStats(),
		RewriteIpv6ToDomain: config.GetDispatcher().GetIpv6UseDomain(),
	}

	if config.Log.LogLevel == configs.Level_DEBUG {
		debugHook := &hooks.DebugHook{}
		d.AddBeforeHandlerSelectionHook(debugHook)
		d.AddAfterHandlerSelectionHook(debugHook)
		d.AddSessionEndHook(debugHook)
	}

	rewriteDestination := &hooks.RewriteDestinationHook{
		DestinationOverride: config.GetDispatcher().GetDestinationOverride(),
		Sniff:               config.GetDispatcher().GetSniff(),
		Sniffer: sniff.NewSniffer(sniff.SniffSetting{
			Interval: 10 * time.Millisecond,
			Sniffers: []sniff.ProtocolSnifferWithNetwork{
				sniff.TlsSniff,
				sniff.HTTP1Sniff,
				sniff.QUICSniff,
				sniff.BTScniff,
				sniff.UTPSniff,
			},
		}),
	}
	d.AddBeforeHandlerSelectionHook(rewriteDestination)

	cc.Dispatcher = d
	fc.requireFeature(func(om *outbound.Manager, p *policy.Policy,
		ul *userlogger.UserLogger) {
		d.AddAfterHandlerSelectionHook(ul)
		d.AddBeforeHandlerSelectionHook(&hooks.IdleHook{
			TimeoutPolicy: p,
		})
		d.AddSessionEndHook(ul)
		d.AddOnFallback(ul)
	})
	fc.requireOptionalFeatures(func(id *dns.AllDnsServers) {
		rewriteDestination.FakeDns = id
		rewriteDestination.Dns = cc.IPResolverForRequestAddress
	})

	if len(config.GetSelectors().GetSelectors()) > 0 {
		for _, selectorConfig := range config.Selectors.Selectors {
			if selectorConfig.SelectFromOm {
				err := fc.requireFeature(func(dispatcher *dispatcher.Dispatcher, tester *tester.Tester,
					om *outbound.Manager, handlerFactory *handlerfactory.HandlerFactory) error {
					filter := selector.NewOmFilter(selectorConfig.GetFilter(), om)
					selectors.AddSelector(selector.NewSelector(selector.SelectorConfig{
						SelectorConfig:            selectorConfig,
						CreateHandler:             handlerFactory,
						Tester:                    tester,
						HandlerErrorChangeSubject: d,
						Filter:                    filter,
						HandlerInfo:               cc.OutStats,
					}))
					return nil
				})
				if err != nil {
					return err
				}
			} else {
				err := fc.requireFeature(func(tester *tester.Tester,
					db client.Db, handlerFactory *handlerfactory.HandlerFactory) error {
					landHandlers, err := selector.ResolveLandHandlers(db, selectorConfig.LandHandlers, cc.HandlerStore)
					if err != nil {
						return err
					}
					var filter selector.Filter
					if config.GetDbPath() == "" && config.GetServicePort() != 0 && runtime.GOOS == "android" {
						filter = selector.NewDbOrOmFilter(db, selectorConfig.GetFilter(),
							landHandlers, handlerFactory, cc.HandlerStore)
					} else {
						filter = selector.NewDbFilter(db, selectorConfig.GetFilter(),
							landHandlers, handlerFactory)
					}
					selectors.AddSelector(selector.NewSelector(selector.SelectorConfig{
						SelectorConfig:            selectorConfig,
						CreateHandler:             handlerFactory,
						Tester:                    tester,
						HandlerErrorChangeSubject: d,
						Filter:                    filter,
						HandlerInfo:               cc.OutStats,
					}))
					return nil
				})
				if err != nil {
					return err
				}
			}
		}
	}

	// router
	err := fc.requireFeature(func(om *outbound.Manager, g i.GeoHelper,
		_ *idns.AllDnsServers) error {
		r, err := router.NewRouter(&router.RouterConfig{
			RouterConfig:    config.Router,
			GeoHelper:       g,
			OutboundManager: om,
			Selectors:       selectors,
			IpResolver:      cc.IPResolverForRequestAddress,
		})
		if err != nil {
			return err
		}
		routerWrapper := &router.RouterWrapper{}
		routerWrapper.UpdateRouter(r)
		d.Router = routerWrapper
		cc.Router = routerWrapper
		fc.addComponent(routerWrapper)
		return nil
	})
	if err != nil {
		return err
	}

	common.Must(fc.addComponent(d))
	return nil
}
