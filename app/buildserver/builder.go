// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

//go:build server || test

package buildserver

import (
	"net"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/create"
	"github.com/5vnetwork/vx-core/app/dns"
	"github.com/5vnetwork/vx-core/app/inbound/monitor"
	"github.com/5vnetwork/vx-core/app/inbound/proxy"
	"github.com/5vnetwork/vx-core/app/memmon"
	"github.com/5vnetwork/vx-core/app/outbound"
	"github.com/5vnetwork/vx-core/app/user"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/proxy/freedom"
	"github.com/5vnetwork/vx-core/transport"
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
	fxOptions = append(fxOptions, fx.Supply(config.Outbound))

	// use default dialer factory for now
	fxOptions = append(fxOptions, fx.Provide(func() transport.DialerFactory {
		return transport.DefaultDialerFactory()
	}))
	fxOptions = append(fxOptions, fx.Provide(func() i.IPResolver {
		return &dns.GoDnsResolver{
			Resolver: net.DefaultResolver,
		}
	}))

	fxOptions = append(fxOptions, fx.Provide(NewInboundManager))
	fxOptions = append(fxOptions, fx.Provide(NewOutboundManager))
	fxOptions = append(fxOptions, fx.Provide(NewGeoHelper))
	fxOptions = append(fxOptions, fx.Provide(NewRouter))
	fxOptions = append(fxOptions, fx.Provide(fx.Annotate(
		NewDispatcher,
	)))
	fxOptions = append(fxOptions, fx.Provide(fx.Annotate(
		create.NewPolicy,
		fx.As(new(i.TimeoutSetting)),
	)))
	fxOptions = append(fxOptions, fx.Provide(NewUserManager))
	fxOptions = append(fxOptions, fx.Provide(monitor.NewInboundStats))
	fxOptions = append(fxOptions, fx.Provide(NewMonitor))

	// add users to inbounds
	fxOptions = append(fxOptions, fx.Decorate(func(im *proxy.InboundManager, um *user.Manager) *proxy.InboundManager {
		for _, user := range um.Users {
			for _, inbound := range im.GetInbounds() {
				inbound.AddUser(user)
			}
		}
		return im
	}))
	// add a freedom handler
	fxOptions = append(fxOptions, fx.Decorate(func(om *outbound.Manager, ipr i.IPResolver) *outbound.Manager {
		om.AddHandlers(freedom.New(
			&transport.Prefer4Dialer{
				Dialer:     transport.DefaultDialer,
				IpResolver: ipr,
			},
			transport.DefaultPacketListener,
			"direct",
			ipr,
		))
		return om
	}))

	fxOptions = append(fxOptions, fx.Invoke(func(im *proxy.InboundManager) {
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
