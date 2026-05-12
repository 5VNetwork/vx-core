// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package api

import (
	"github.com/5vnetwork/vx-core/app/create/outbound"
	"github.com/5vnetwork/vx-core/app/policy"
	"github.com/5vnetwork/vx-core/app/util"
	"github.com/5vnetwork/vx-core/common/session"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

type SpeedTestResult struct {
	Ping uint32 //ms
	Down uint64 //bytes/s
}

func (a *Api) SpeedTest(req *SpeedTestRequest, in Api_SpeedTestServer) error {
	const batchSize = 100
	handlers := req.GetHandlers()
	url := util.SpeedtestURL1

	for start := 0; start < len(handlers); start += batchSize {
		end := start + batchSize
		if end > len(handlers) {
			end = len(handlers)
		}

		wg := new(errgroup.Group)
		for _, handler := range handlers[start:end] {
			t := handler
			wg.Go(func() error {
				log.Debug().Msgf("SpeedTest for: %v", util.GetTag(t))

				rsp := &SpeedTestResponse{
					Tag: util.GetTag(t),
				}
				h, err := outbound.NewHandler(&outbound.HandlerConfig{
					HandlerConfig:               t,
					DialerFactory:               a.getDialerFactory(),
					Policy:                      policy.New(),
					IPResolver:                  a.getIPResolver(),
					EchResolver:                 a.echResolver,
					IPResolverForRequestAddress: a.getIPResolver(),
				})
				if err != nil {
					log.Debug().Err(err).Str("tag", util.GetTag(t)).Msg("failed to create outbound handler")
					rsp.Down = -1
				} else {
					logger := log.With().Uint32("sid", uint32(session.NewID())).
						Str("handler", util.GetTag(t)).Logger()
					logger.Debug().Msg("speed test start")
					ctx := logger.WithContext(in.Context())
					rst, err := util.Speedtest(ctx, url, h)
					if err != nil {
						log.Err(err).Msg("failed to speed test")
						rsp.Down = -1
					} else {
						rsp.Down = int32(rst)
					}
				}
				if err := in.Send(rsp); err != nil {
					log.Err(err).Msg("failed to send speed test response")
					return err
				}
				return nil
			})
		}
		if err := wg.Wait(); err != nil {
			return err
		}
	}
	return nil
}

// use speedtest.net
// func SpeedTest0(ctx context.Context, t *configs.OutboundHandlerConfig) (*SpeedTestResponse, error) {
// 	h, err := outbound.NewOutHandler(t, transport.NewDefaultDialerFactory(),
// 		policy.New(), nil)
// 	if err != nil {
// 		return nil, fmt.Errorf("speedtest create outbound handler err: %v", err)
// 	}
// 	rst, err := st.Run(ctx, speedtest.WithDoer(outbound.HandlerToHttpClient(h)))
// 	if err != nil {
// 		return nil, fmt.Errorf("speedtest run err: %v", err)
// 	}
// 	return &SpeedTestResponse{
// 		Ok:   true,
// 		Tag:  h.Tag(),
// 		Up:   uint64(rst.Upload),
// 		Down: uint64(rst.Download),
// 		Ping: uint32(rst.Latency.Milliseconds()),
// 	}, nil
// }
