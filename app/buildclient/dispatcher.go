// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package buildclient

import (
	"fmt"
	"time"

	"github.com/5vnetwork/vx-core/app/client"
	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/dispatcher"
	"github.com/5vnetwork/vx-core/app/dns"
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
		SessinoErrorLogger: &dispatcher.NullSessionErrorLogger{},
	}

	if config.Log.LogLevel == configs.Level_DEBUG {
		debugHook := &dispatcher.DebugHook{}
		d.AddBeforeHandlerSelectionHook(debugHook)
		d.AddAfterHandlerSelectionHook(debugHook)
	}

	rewriteDestination := &dispatcher.RewriteDestinationHook{
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

	fallback := &dispatcher.Fallbacker{
		FallbackToProxy:  config.GetDispatcher().GetFallbackToProxy(),
		FallbackToDomain: config.GetDispatcher().GetFallbackToDomain(),
		Sm:               selectors,
	}
	d.Fallback = fallback

	if config.GetDispatcher().GetIpv6UseDomain() {
		d.AddAfterHandlerSelectionHook(&dispatcher.RewriteIPv6ToDomainHook{})
	}

	cc.Dispatcher = d
	fc.requireFeature(func(om *outbound.Manager, p *policy.Policy,
		ul *userlogger.UserLogger) {
		d.SessinoErrorLogger = ul
		fallback.Om = om
		fallback.Logger = ul
		d.AddAfterHandlerSelectionHook(ul)
		d.AddBeforeHandlerSelectionHook(&dispatcher.IdleHook{
			TimeoutPolicy: p,
		})
		d.AddAfterHandlerSelectionHook(&dispatcher.StatsHook{
			StatsPolicy: p,
			OutStats:    cc.OutStats,
		})
	})
	fc.requireOptionalFeatures(func(id *dns.Dns) {
		rewriteDestination.FakeDns = cc.AllFakeDns
		rewriteDestination.Dns = cc.IPResolverForRequestAddress
	})

	if len(config.GetSelectors().GetSelectors()) > 0 {
		for _, selectorConfig := range config.Selectors.Selectors {
			if selectorConfig.SelectFromOm {
				err := fc.requireFeature(func(dispatcher *dispatcher.Dispatcher, tester *tester.Tester,
					om *outbound.Manager) error {
					filter := selector.NewOmFilter(selectorConfig.GetFilter(), om)
					selectors.AddSelector(selector.NewSelector(selector.SelectorConfig{
						SelectorConfig:            selectorConfig,
						CreateHandler:             cc.CreateHandlerWithLandHandlers,
						Tester:                    tester,
						HandlerErrorChangeSubject: d,
						Filter:                    filter,
					}))
					return nil
				})
				if err != nil {
					return err
				}
			} else {
				err := fc.requireFeature(func(dispatcher *dispatcher.Dispatcher, tester *tester.Tester,
					db client.Db) error {
					landHandlers := make([]*xsqlite.OutboundHandler, 0, len(selectorConfig.LandHandlers))
					for _, landHandlerId := range selectorConfig.LandHandlers {
						handler := db.GetHandler(int(landHandlerId))
						if handler == nil {
							return fmt.Errorf("land handler %d not found", landHandlerId)
						}
						landHandlers = append(landHandlers, handler)
					}
					filter := selector.NewDbFilter(db, selectorConfig.GetFilter(),
						landHandlers, cc.CreateHandlerWithLandHandlers)
					selectors.AddSelector(selector.NewSelector(selector.SelectorConfig{
						SelectorConfig:            selectorConfig,
						CreateHandler:             cc.CreateHandlerWithLandHandlers,
						Tester:                    tester,
						HandlerErrorChangeSubject: d,
						Filter:                    filter,
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
		ipr i.IPResolver) error {
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
