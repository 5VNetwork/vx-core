package scenarios

// var DefaultTunConfig = &configs.TunDeviceConfig{
// 	Cidr4:   "172.23.27.1/24",
// 	Cidr6:   "fd00::1/64",
// 	Name:    "tun0",
// 	Mtu:     1500,
// 	Dns4:    []string{"172.23.27.2"},
// 	Dns6:    []string{"fd00::2"},
// 	Routes4: []string{"0.0.0.0/0"},
// 	Routes6: []string{"::/0"},
// 	Path:    "../../tun/wintun",
// }

// func TestTunTCP(t *testing.T) {
// 	tcpServer := tcp.Server{
// 		MsgProcessor: xor,
// 	}
// 	dest, err := tcpServer.Start()
// 	common.Must(err)
// 	defer tcpServer.Close()

// 	t.Log(os.Getwd())

// 	//TODO
// 	clientConfig := &configs.TmConfig{
// 		Tun: &configs.TunConfig{
// 			Inbound: &configs.TunInboundConfig{
// 				Tag:  "tun",
// 				Mode: configs.TunInboundConfig_MODE_SYSTEM,
// 				Tun:  DefaultTunConfig,
// 			},
// 			Info: &configs.TunInfoConfig{
// 				Name: "tun",
// 			},
// 			Monitor: &configs.TunMonitorConfig{
// 				Name: "tun",
// 			},
// 		},
// 	}

// 	client, err := builder.NewInstanceTM(clientConfig)
// 	common.Must(err)
// 	common.Must(client.Start())
// 	defer client.Close()

// 	var errg errgroup.Group
// 	for i := 0; i < 1; i++ {
// 		errg.Go(testTCPConn(dest.Port, 10240*1024, timeout40))
// 	}

// 	if err := errg.Wait(); err != nil {
// 		t.Error(err)
// 	}
// }
