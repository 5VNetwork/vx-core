package scenarios

// func TestDnsFakeDns(t *testing.T) {
// 	serverPort := tcp.PickPort()
// 	serverConfig := &x.Config{
// 		 &configs.DnsConfig{},
// 		 &configs.RouterConfig{
// 			DomainStrategy: domain.DomainStrategy_IpIfNonMatch,
// 			Rules: []*configs.RuleConfig{
// 				{
// 					OutboundTag: "direct",
// 					DstCidrs: []string{
// 						"127.0.0.0/8",
// 					},
// 				},
// 			},
// 		},
// 		 &configs.InboundManagerConfig{
// 			Handlers: []*anypb.Any{
// 				serial.ToTypedMessage(&configs.ProxyInboundConfig{
// 					Address: net.LocalHostIP.String(),
// 					Port:    uint32(serverPort),
// 					Protocol: serial.ToTypedMessage(&configs.SocksServerConfig{
// 						AuthType: socks.AuthType_NO_AUTH,
// 						Accounts: []*user.User{
// 							{
// 								Id:     protocol.NewID(uuid.New()).String(),
// 								Secret: protocol.NewID(uuid.New()).String(),
// 							},
// 						},
// 						Address: net.LocalHostIP.String(),
// 					}),
// 				}),
// 			},
// 		},
// 		 &configs.OutboundManagerConfig{
// 			OutboundHandlers: []*configs.OutboundHandlerConfig{
// 				{
// 					Protocol: serial.ToTypedMessage(&configs.BlackholeConfig{}),
// 				},
// 				{
// 					Tag:            "direct",
// 					Protocol:    serial.ToTypedMessage(&configs.FreedomConfig{}),
// 					DomainStrategy: domain.DomainStrategy_UseIp,
// 				},
// 			},
// 		},
// 	}

// 	server, err := x.NewInstance(serverConfig)
// 	common.Must(err)
// 	common.Must(server.Start())
// 	defer server.Close()

// }
