// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

//go:build server || test

package buildserver

import (
	"context"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/create"
	"github.com/5vnetwork/vx-core/app/dispatcher"
	"github.com/5vnetwork/vx-core/app/dispatcher/hooks"
	"github.com/5vnetwork/vx-core/app/geo/geosync"
	"github.com/5vnetwork/vx-core/app/inbound/monitor"
	"github.com/5vnetwork/vx-core/app/inbound/proxy"
	"github.com/5vnetwork/vx-core/app/memmon"
	"github.com/5vnetwork/vx-core/app/user"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/nic"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

func NewX(config *configs.ServerConfig) (*fx.App, error) {
	var fxOptions []fx.Option
	fxOptions = append(fxOptions, fx.Supply(config.Inbounds))
	fxOptions = append(fxOptions, fx.Supply(config.MultiInbounds))
	fxOptions = append(fxOptions, fx.Supply(config.Outbounds))
	fxOptions = append(fxOptions, fx.Supply(config.Users))
	fxOptions = append(fxOptions, fx.Supply(config.Geo))
	fxOptions = append(fxOptions, fx.Supply(config.Router))
	fxOptions = append(fxOptions, fx.Supply(config.Policy))
	fxOptions = append(fxOptions, fx.Supply(config.Dns))
	fxOptions = append(fxOptions, fx.Supply(config.Dispatcher))
	fxOptions = append(fxOptions, fx.Supply(config.DialerFactory))

	fxOptions = append(fxOptions, DialerFactoryOption(config.GetDialerFactory()))
	fxOptions = append(fxOptions, fx.Provide(NewInboundManager))
	fxOptions = append(fxOptions, fx.Provide(NewOutboundManager))
	fxOptions = append(fxOptions, fx.Provide(NewGeoHelper))
	fxOptions = append(fxOptions, fx.Provide(NewRouter))
	fxOptions = append(fxOptions, fx.Provide(NewDispatcher))
	fxOptions = append(fxOptions, fx.Provide(fx.Annotate(
		create.NewPolicy,
		fx.As(new(i.TimeoutSetting)),
	)))
	fxOptions = append(fxOptions, fx.Provide(NewUserManager))
	fxOptions = append(fxOptions, fx.Provide(monitor.NewInboundStats))
	fxOptions = append(fxOptions, fx.Provide(NewMonitor))
	fxOptions = append(fxOptions, fx.Provide(NewDNS))
	fxOptions = append(fxOptions, fx.Provide(NewGeoSync))
	fxOptions = append(fxOptions, fx.Provide(
		func(lc fx.Lifecycle) i.DefaultInterfaceInfo {
			m, err := nic.NewInterfaceMonitor("")
			if err != nil {
				panic(err)
			}
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					return m.Start()
				},
				OnStop: func(ctx context.Context) error {
					return m.Close()
				},
			})
			return m
		}))

	// add users to inbounds
	fxOptions = append(fxOptions, fx.Decorate(func(im *proxy.InboundManager,
		um *user.Manager) *proxy.InboundManager {
		for _, user := range um.Users {
			for _, inbound := range im.GetInbounds() {
				inbound.AddUser(user)
			}
		}
		return im
	}))
	// add router to dispatcher to break circular dependency (single Decorate per type)
	fxOptions = append(fxOptions,
		fx.Decorate(func(dp *dispatcher.Dispatcher, router i.Router) *dispatcher.Dispatcher {
			dp.Router = router
			if config.GetLog().GetLogLevel() == configs.Level_DEBUG {
				debugHook := &hooks.DebugHook{}
				dp.AddBeforeHandlerSelectionHook(debugHook)
				dp.AddAfterHandlerSelectionHook(debugHook)
				dp.AddSessionEndHook(debugHook)
			}
			return dp
		}))

	fxOptions = append(fxOptions, fx.Invoke(func(im *proxy.InboundManager) {
	}))
	fxOptions = append(fxOptions, fx.Invoke(func(gs *geosync.GeoSync) {
	}))
	fxOptions = append(fxOptions, fx.Invoke(func(dp *dispatcher.Dispatcher) {
	}))

	if config.GetLog().GetLogLevel() != configs.Level_DEBUG {
		fxOptions = append(fxOptions, fx.WithLogger(func() fxevent.Logger {
			return fxevent.NopLogger
		}))
	} else {
		fxOptions = append(fxOptions, fx.Invoke(func(monitor *memmon.Monitor) {
		}))
	}

	return fx.New(
		fxOptions...,
	), nil
}
