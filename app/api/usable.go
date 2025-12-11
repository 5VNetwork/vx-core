package api

import (
	context "context"
	"io"
	"net/http"
	"time"

	"github.com/5vnetwork/vx-core/app/outbound"
	"github.com/5vnetwork/vx-core/app/policy"
	"github.com/5vnetwork/vx-core/app/util"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/session"

	"github.com/rs/zerolog/log"
)

// use TraceList. Get ip, usable of a handler.
func (a *Api) HandlerTest(ctx context.Context, req *HandlerUsableRequest, url string) (ret HandlerUsableResponse) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	logger := log.With().Uint32("sid", uint32(session.NewID())).Logger()
	ctx = logger.WithContext(ctx)

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

	h, err := outbound.NewHandler(&outbound.HandlerConfig{
		HandlerConfig: req.Handler,
		DialerFactory: a.getDialerFactory(),
		Policy:        policy.New(),
		IPResolver:    a.getIPResolver(),
		DnsServer:     a.dnsServer,
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

	pairs := util.ParseKeyValueText(string(data))

	ret.Ping = int32(ping)
	ret.Ip = pairs["ip"]
	return
}
