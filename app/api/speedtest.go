package api

type SpeedTestResult struct {
	Ping uint32 //ms
	Down uint64 //bytes/s
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
