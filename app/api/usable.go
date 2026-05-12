// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package api

import (
	context "context"
	"net/http"
	"time"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/create/outbound"
	"github.com/5vnetwork/vx-core/app/policy"
	"github.com/5vnetwork/vx-core/app/util"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/session"

	"github.com/rs/zerolog/log"
)

func (a *Api) HandlerUsable(ctx context.Context, req *HandlerUsableRequest) (*HandlerUsableResponse, error) {
	log.Debug().Msgf("HandlerUsable for: %v", util.GetTag(req.Handler))
	for i := 0; i < 3; i++ {
		rsp := a.HandlerTest(ctx, req)
		if rsp.Ping > 0 {
			return &rsp, nil
		}
	}
	return &HandlerUsableResponse{
		Ping: -1,
	}, nil
}

// use TraceList. Get ip, usable of a handler.
func (a *Api) HandlerTest(ctx context.Context, req *HandlerUsableRequest) (ret HandlerUsableResponse) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	logger := log.With().Uint32("sid", uint32(session.NewID())).Logger()
	ctx = logger.WithContext(ctx)

	url := util.UsableTestUrlCf
	logger.Debug().Str("handler", configs.HandlerTag(req.Handler)).Str("test", "usable").Str("url", url).Send()

	var dest net.Address
	if req.Handler.GetOutbound() != nil {
		dest = net.ParseAddress(req.Handler.GetOutbound().Address)
	} else {
		dest = net.ParseAddress(req.Handler.GetChain().Handlers[len(req.Handler.GetChain().Handlers)-1].Address)
	}
	if dest.Family().IsDomain() {
		if a.mon != nil && a.mon.SupportIPv6() <= 0 {
			a.getIPResolver().LookupIPv4(ctx, dest.Domain())
		} else {
			a.getIPResolver().LookupIP(ctx, dest.Domain())
		}
	}

	h, err := outbound.NewHandler(&outbound.HandlerConfig{
		HandlerConfig:               req.Handler,
		DialerFactory:               a.getDialerFactory(),
		Policy:                      policy.New(),
		IPResolver:                  a.getIPResolver(),
		IPResolverForRequestAddress: a.getIPResolver(),
		EchResolver:                 a.echResolver,
	})
	if err != nil {
		logger.Debug().Msgf("Handler %s create handler err: %v", configs.HandlerTag(req.Handler), err)
		return
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		logger.Debug().Err(err).Msg("usable test failed")
		return
	}

	start := time.Now()
	httpClient := util.HandlerToHttpClient(h)
	defer httpClient.CloseIdleConnections()

	httpClient.Timeout = 4 * time.Second
	rsp, err := httpClient.Do(request)
	if err != nil {
		logger.Debug().Msgf("Handler %s get url err: %v", configs.HandlerTag(req.Handler), err)
		return
	}
	logger.Debug().Msg("response got")
	ping := time.Since(start).Milliseconds()
	rsp.Body.Close()

	ret.Ping = int32(ping)
	return
}
