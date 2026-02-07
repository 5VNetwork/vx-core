// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package api

import (
	"bufio"
	context "context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/5vnetwork/vx-core/app/create"
	"github.com/5vnetwork/vx-core/app/policy"
	"github.com/5vnetwork/vx-core/app/util"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/session"

	"github.com/rs/zerolog/log"
)

// use TraceList. Get ip, usable of a handler.
func (a *Api) HandlerTest(ctx context.Context, req *HandlerUsableRequest) (ret HandlerUsableResponse) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	logger := log.With().Uint32("sid", uint32(session.NewID())).Logger()
	ctx = logger.WithContext(ctx)

	url := util.TraceList[0]
	logger.Debug().Str("handler", req.Handler.GetTag()).Str("test", "usable").Str("url", url).Send()

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

	h, err := create.NewHandler(&create.HandlerConfig{
		HandlerConfig: req.Handler,
		DialerFactory: a.getDialerFactory(),
		Policy:        policy.New(),
		IPResolver:    a.getIPResolver(),
		EchResolver:   a.echResolver,
	})
	if err != nil {
		logger.Debug().Msgf("Handler %s create handler err: %v", req.Handler.GetTag(), err)
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
		logger.Debug().Msgf("Handler %s get url err: %v", req.Handler.GetTag(), err)
		return
	}
	logger.Debug().Msg("response got")
	ping := time.Since(start).Milliseconds()
	data, err := io.ReadAll(rsp.Body)
	if err != nil {
		logger.Debug().Msgf("Handler %s read body err: %v", req.Handler.GetTag(), err)
		return
	}
	logger.Debug().Msg("body read")
	rsp.Body.Close()

	pairs := ParseKeyValueText(string(data))

	ret.Ping = int32(ping)
	ret.Ip = pairs["ip"]
	ret.Country = pairs["loc"]
	return
}

const testUrl = "https://blog.cloudflare.com/cdn-cgi/trace"

func ParseKeyValueText(text string) map[string]string {
	result := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(text))

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			result[key] = value
		}
	}

	return result
}
