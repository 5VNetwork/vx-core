//go:build server || test

package buildserver

import (
	"context"
	"fmt"
	"net"

	configs "github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/configs/server"
	"github.com/5vnetwork/vx-core/app/create"
	"github.com/5vnetwork/vx-core/app/dispatcher"
	"github.com/5vnetwork/vx-core/app/dns"
	"github.com/5vnetwork/vx-core/app/geo"
	"github.com/5vnetwork/vx-core/app/inbound/monitor"
	"github.com/5vnetwork/vx-core/app/inbound/proxy"
	"github.com/5vnetwork/vx-core/app/outbound"
	"github.com/5vnetwork/vx-core/app/router"
	"github.com/5vnetwork/vx-core/app/user"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/proxy/freedom"
	"github.com/5vnetwork/vx-core/transport"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

func NewX(config *server.ServerConfig) (*fx.App, error) {
	var fxOptions []fx.Option
	fxOptions = append(fxOptions, fx.Supply(config.Inbounds))
	fxOptions = append(fxOptions, fx.Supply(config.MultiInbounds))
	fxOptions = append(fxOptions, fx.Supply(config.Outbounds))
	fxOptions = append(fxOptions, fx.Supply(config.Users))
	fxOptions = append(fxOptions, fx.Supply(config.Geo))
	fxOptions = append(fxOptions, fx.Supply(config.Router))
	fxOptions = append(fxOptions, fx.Supply(config.Policy))

	fxOptions = append(fxOptions, fx.Provide(func() i.IPResolver {
		return &dns.DnsResolver{
			Resolver: net.DefaultResolver,
		}
	}))
	fxOptions = append(fxOptions, fx.Provide(func(ipr i.IPResolver) transport.DialerFactory {
		return transport.NewDialerFactoryImp(transport.DialerFactoryOption{
			IpResolver: ipr,
		})
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
		fx.As(new(i.StatsSetting)),
	)))
	fxOptions = append(fxOptions, fx.Provide(NewUserManager))
	fxOptions = append(fxOptions, fx.Provide(monitor.NewInboundStats))

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
	}
	return fx.New(
		fxOptions...,
	), nil
}

type UserManagerParams struct {
	fx.In
	Configs []*configs.UserConfig
}
type UserManagerResult struct {
	fx.Out
	UserManager *user.Manager
}

func NewUserManager(lc fx.Lifecycle, params UserManagerParams) (UserManagerResult, error) {
	um := user.NewManager()
	for _, userConfig := range params.Configs {
		u := user.NewUser(userConfig.Id, 0, userConfig.Secret)
		um.AddUser(u)
	}
	return UserManagerResult{UserManager: um}, nil
}

type OutboundManagerParams struct {
	fx.In
	Configs       []*configs.OutboundHandlerConfig
	DialerFactory transport.DialerFactory
	IpResolver    i.IPResolver
	Policy        i.TimeoutSetting
}
type OutboundManagerResult struct {
	fx.Out
	OutboundManager *outbound.Manager
}

func NewOutboundManager(lc fx.Lifecycle, params OutboundManagerParams) (OutboundManagerResult, error) {
	om := outbound.NewManager()
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return om.Start()
		},
		OnStop: func(ctx context.Context) error {
			return om.Close()
		},
	})
	for _, handlerConfig := range params.Configs {
		h, err := outbound.NewOutHandler(&outbound.Config{
			OutboundHandlerConfig: handlerConfig,
			DialerFactory:         params.DialerFactory,
			Policy:                params.Policy,
			IPResolver:            params.IpResolver,
		})
		if err != nil {
			return OutboundManagerResult{}, fmt.Errorf("failed to create outbound proxy handler: %w", err)
		}
		om.AddHandlers(h)
	}
	return OutboundManagerResult{OutboundManager: om}, nil
}

type GeoHelperParams struct {
	fx.In
	Config *configs.GeoConfig
}
type GeoHelperResult struct {
	fx.Out
	GeoHelper *geo.Geo
}

func NewGeoHelper(params GeoHelperParams) (GeoHelperResult, error) {
	geoHelper, err := geo.NewGeo(params.Config)
	if err != nil {
		return GeoHelperResult{}, fmt.Errorf("failed to create geo helper: %w", err)
	}
	return GeoHelperResult{GeoHelper: geoHelper}, nil
}

type RouterParams struct {
	fx.In
	Config          *configs.RouterConfig
	OutboundManager *outbound.Manager
	GeoHelper       *geo.Geo
	IpResolver      i.IPResolver
}
type RouterResult struct {
	fx.Out
	Router i.Router
}

func NewRouter(lc fx.Lifecycle, params RouterParams) (RouterResult, error) {
	router, err := router.NewRouter(&router.RouterConfig{
		RouterConfig:    params.Config,
		OutboundManager: params.OutboundManager,
		GeoHelper:       params.GeoHelper,
		IpResolver:      params.IpResolver,
	})
	if err != nil {
		return RouterResult{}, fmt.Errorf("failed to create router: %w", err)
	}

	return RouterResult{Router: router}, nil
}

type DispatcherParams struct {
	fx.In
	Policy  i.StatsSetting
	Timeout i.TimeoutSetting
	InStats *monitor.InboundStats
	Router  i.Router
}
type DispatcherResult struct {
	fx.Out
	Handler    i.Handler
	Dispatcher *dispatcher.Dispatcher
}

func NewDispatcher(params DispatcherParams) (DispatcherResult, error) {
	dp := dispatcher.New()
	dp.TimeoutPolicy = params.Timeout
	dp.StatsPolicy = params.Policy
	dp.InboundStats = params.InStats
	dp.Router = params.Router
	return DispatcherResult{Handler: dp, Dispatcher: dp}, nil
}

type InboundManagerParams struct {
	fx.In
	Configs      []*configs.ProxyInboundConfig
	MultiConfigs []*configs.MultiProxyInboundConfig
	Handler      i.Handler
	Router       i.Router
	Policy       i.TimeoutSetting
	Stats        *monitor.InboundStats
	OnUnauth     i.UnauthorizedReport `optional:"true"`
}
type InboundManagerResult struct {
	fx.Out
	InboundManager *proxy.InboundManager
}

func NewInboundManager(lc fx.Lifecycle, params InboundManagerParams) (InboundManagerResult, error) {
	im := proxy.NewManager()
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return im.Start()
		},
		OnStop: func(ctx context.Context) error {
			return im.Close()
		},
	})
	for _, config := range params.Configs {
		h, err := proxy.NewInboundServer(config, params.Handler, params.Router, params.Policy,
			params.Stats.Get(config.Tag), params.OnUnauth)
		if err != nil {
			return InboundManagerResult{}, fmt.Errorf("failed to create inbound proxy handler: %w", err)
		}
		im.AddInbound(h)
	}
	for _, config := range params.MultiConfigs {
		h, err := proxy.NewMultiInboundServer(config, params.Handler, params.Router, params.Policy,
			params.Stats.Get(config.Tag), params.OnUnauth)
		if err != nil {
			return InboundManagerResult{}, fmt.Errorf("failed to create inbound proxy handler: %w", err)
		}
		im.AddInbound(h)
	}
	return InboundManagerResult{InboundManager: im}, nil
}
