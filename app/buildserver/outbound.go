//go:build server || test

package buildserver

import (
	"context"
	"fmt"

	"github.com/5vnetwork/vx-core/app/configs"
	outboundcreate "github.com/5vnetwork/vx-core/app/create/outbound"
	"github.com/5vnetwork/vx-core/app/outbound"
	"github.com/5vnetwork/vx-core/common/serial"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/transport"
	"go.uber.org/fx"
)

type OutboundManagerParams struct {
	fx.In
	Configs          []*configs.OutboundHandlerConfig
	DialerFactory    transport.DialerFactory
	IpResolver       i.IPResolver `name:"internal_resolver"`
	Policy           i.TimeoutSetting
	HijackDnsHandler i.Handler `name:"hijack_dns_handler" optional:"true"`
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

	if len(params.Configs) == 0 {
		params.Configs = []*configs.OutboundHandlerConfig{
			{
				Tag:      "direct",
				Protocol: serial.ToTypedMessage(&configs.FreedomConfig{}),
			},
		}
	}

	for _, handlerConfig := range params.Configs {
		h, err := outboundcreate.NewOutHandler(&outboundcreate.Config{
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

	if params.HijackDnsHandler != nil {
		om.AddHandlers(params.HijackDnsHandler.(i.Outbound))
	}

	return OutboundManagerResult{OutboundManager: om}, nil
}
