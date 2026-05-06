//go:build server || test

package buildserver

import (
	"context"
	"time"

	"github.com/5vnetwork/vx-core/app/dispatcher"
	"github.com/5vnetwork/vx-core/app/memmon"
	"go.uber.org/fx"
)

type MonitorParams struct {
	fx.In
	Dispatcher *dispatcher.Dispatcher
}

type MonitorResult struct {
	fx.Out
	Monitor *memmon.Monitor
}

func NewMonitor(lc fx.Lifecycle, params MonitorParams) (MonitorResult, error) {
	monitor := memmon.NewMonitor(&memmon.MonitorConfig{
		Interval:      time.Second * 1,
		ListenAddress: "127.0.0.1:6060",
	})
	monitor.Dispatcher = params.Dispatcher
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return monitor.Start()
		},
		OnStop: func(ctx context.Context) error {
			return monitor.Close()
		},
	})
	return MonitorResult{Monitor: monitor}, nil
}
