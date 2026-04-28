package api

import (
	"bufio"
	context "context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/create"
	"github.com/5vnetwork/vx-core/app/policy"
	"github.com/5vnetwork/vx-core/app/util"
	"github.com/5vnetwork/vx-core/common/session"
	"github.com/rs/zerolog/log"
)

func (a *Api) HandlerCountryTest(ctx context.Context, req *HandlerCountryTestRequest) (*HandlerCountryTestResponse, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	logger := log.With().Uint32("sid", uint32(session.NewID())).Logger()
	ctx = logger.WithContext(ctx)

	url := util.TraceList[0]
	logger.Debug().Str("handler", configs.HandlerTag(req.Handler)).Str("test", "country").Str("url", url).Send()

	h, err := create.NewHandler(&create.HandlerConfig{
		HandlerConfig:               req.Handler,
		DialerFactory:               a.getDialerFactory(),
		Policy:                      policy.New(),
		IPResolver:                  a.getIPResolver(),
		IPResolverForRequestAddress: a.getIPResolver(),
		EchResolver:                 a.echResolver,
	})
	if err != nil {
		logger.Debug().Msgf("Handler %s create handler err: %v", configs.HandlerTag(req.Handler), err)
		return nil, err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		logger.Debug().Err(err).Msg("country test failed")
		return nil, err
	}

	httpClient := util.HandlerToHttpClient(h)
	defer httpClient.CloseIdleConnections()

	httpClient.Timeout = 10 * time.Second
	rsp, err := httpClient.Do(request)
	if err != nil {
		logger.Debug().Msgf("Handler %s get url err: %v", configs.HandlerTag(req.Handler), err)
		return nil, err
	}
	logger.Debug().Msg("response got")
	data, err := io.ReadAll(rsp.Body)
	if err != nil {
		logger.Debug().Msgf("Handler %s read body err: %v", configs.HandlerTag(req.Handler), err)
		return nil, err
	}
	logger.Debug().Msg("body read")
	rsp.Body.Close()

	pairs := ParseKeyValueText(string(data))

	ip := pairs["ip"]
	country := pairs["loc"]
	return &HandlerCountryTestResponse{Ip: ip, Country: country}, nil
}

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
