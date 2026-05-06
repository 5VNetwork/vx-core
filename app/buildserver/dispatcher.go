//go:build server || test

package buildserver

import (
	"github.com/5vnetwork/vx-core/app/dispatcher"
	"github.com/5vnetwork/vx-core/app/dispatcher/hooks"
	"github.com/5vnetwork/vx-core/i"
	"go.uber.org/fx"
)

type DispatcherParams struct {
	fx.In
	Timeout i.TimeoutSetting
	Router  i.Router
}
type DispatcherResult struct {
	fx.Out
	Handler    i.Handler
	Dispatcher *dispatcher.Dispatcher
}

func NewDispatcher(params DispatcherParams) (DispatcherResult, error) {
	dp := &dispatcher.Dispatcher{}
	dp.AddBeforeHandlerSelectionHook(&hooks.IdleHook{
		TimeoutPolicy: params.Timeout,
	})
	dp.Router = params.Router
	return DispatcherResult{Handler: dp, Dispatcher: dp}, nil
}
