//go:build server || test

package buildserver

import (
	"time"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/dispatcher"
	"github.com/5vnetwork/vx-core/app/dispatcher/hooks"
	"github.com/5vnetwork/vx-core/app/sniff"
	"github.com/5vnetwork/vx-core/i"
	"go.uber.org/fx"
)

type DispatcherParams struct {
	fx.In
	Config  *configs.DispatcherConfig
	Timeout i.TimeoutSetting
}
type DispatcherResult struct {
	fx.Out
	Handler    i.Handler `name:"dispatcher"`
	Dispatcher *dispatcher.Dispatcher
}

func NewDispatcher(params DispatcherParams) (DispatcherResult, error) {
	dp := &dispatcher.Dispatcher{
		SessionStats: params.Config.GetSessionStats(),
	}
	if len(params.Config.GetDestinationOverride()) > 0 || params.Config.GetSniff() {
		rewriteDestination := &hooks.RewriteDestinationHook{
			DestinationOverride: params.Config.GetDestinationOverride(),
			Sniff:               params.Config.GetSniff(),
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
		dp.AddBeforeHandlerSelectionHook(rewriteDestination)
	}

	dp.AddBeforeHandlerSelectionHook(&hooks.IdleHook{
		TimeoutPolicy: params.Timeout,
	})
	return DispatcherResult{Handler: dp, Dispatcher: dp}, nil
}
