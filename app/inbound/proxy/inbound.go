package proxy

import "github.com/5vnetwork/vx-core/i"

type withHandler interface {
	WithHandler(h i.Handler)
}

type withTimeoutPolicy interface {
	WithTimeoutPolicy(tp i.TimeoutSetting)
}

type withOnUnauthorizedRequest interface {
	WithOnUnauthorizedRequest(f i.UnauthorizedReport)
}
