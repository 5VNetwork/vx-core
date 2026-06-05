//go:build server || test

package buildserver

import (
	"context"
	"fmt"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/create/inbound"
	"github.com/5vnetwork/vx-core/app/inbound/proxy"
	"github.com/5vnetwork/vx-core/app/inbound/proxy/multi"
	"github.com/5vnetwork/vx-core/i"
	"go.uber.org/fx"
)

type InboundManagerParams struct {
	fx.In
	Configs      []*configs.ProxyInboundConfig
	MultiConfigs []*configs.MultiProxyInboundConfig
	Handler      i.Handler `name:"dispatcher"`
	Policy       i.TimeoutSetting
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
		h, err := inbound.NewInboundServer(config, params.Handler, params.Policy,
			params.OnUnauth)
		if err != nil {
			return InboundManagerResult{}, fmt.Errorf("failed to create inbound proxy handler: %w", err)
		}
		im.AddInbound(h)
	}
	for _, config := range params.MultiConfigs {
		h, err := multi.NewMultiInboundServer(config, params.Handler, params.Policy,
			params.OnUnauth)
		if err != nil {
			return InboundManagerResult{}, fmt.Errorf("failed to create inbound proxy handler: %w", err)
		}
		im.AddInbound(h)
	}
	return InboundManagerResult{InboundManager: im}, nil
}
