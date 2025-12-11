package api

// func TestApiHandlerUsable(t *testing.T) {
// 	uri := ""
// 	decoded, err := decode.Decode(uri)
// 	common.Must(err)
// 	h, err := outbound.NewOutHandler(&outbound.Config{
// 		OutboundHandlerConfig: decoded.Configs[0],
// 		DialerFactory:         transport.DefaultDialerFactory(),
// 		IPResolver:            &dns.DnsResolver{},
// 		Policy:                policy.DefaultPolicy,
// 	})
// 	common.Must(err)

// 	response, err := ApiHandlerUsable1(context.Background(), h, TraceList[0])
// 	if err != nil {
// 		common.Must(err)
// 		return
// 	}
// 	log.Println(response)
// }
