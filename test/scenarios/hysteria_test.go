//go:build test

package scenarios

import (
	"context"
	"testing"
	"time"

	"github.com/5vnetwork/vx-core/app/buildclient"
	"github.com/5vnetwork/vx-core/app/buildserver"
	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/configs/proxy"
	"github.com/5vnetwork/vx-core/app/configs/server"
	"github.com/5vnetwork/vx-core/app/util"
	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/protocol/tls/cert"
	"github.com/5vnetwork/vx-core/common/serial"
	"github.com/5vnetwork/vx-core/common/uuid"
	"github.com/5vnetwork/vx-core/test/servers/tcp"
	"github.com/5vnetwork/vx-core/test/servers/udp"
	"github.com/5vnetwork/vx-core/transport/security/tls"

	"golang.org/x/sync/errgroup"
)

func TestHysteriaTCP(t *testing.T) {
	userID := uuid.New()
	serverPort := net.PickUDPPort()
	t.Logf("server port: %d", serverPort)

	serverConfig := &server.ServerConfig{
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Port: uint32(serverPort),
				Protocol: serial.ToTypedMessage(&proxy.Hysteria2ServerConfig{
					IgnoreClientBandwidth: true,
					TlsConfig: &tls.TlsConfig{
						Certificates: []*tls.Certificate{
							tls.ParseCertificate(cert.MustGenerate(nil)),
						},
					},
				}),
			},
		},
		// Router: &configs.RouterConfig{
		// 	Rules: []*configs.RuleConfig{
		// 		{
		// 			OutboundTag: "direct",
		// 		},
		// 	},
		// },
		Users: []*configs.UserConfig{
			{
				Id:     userID.String(),
				Secret: userID.String(),
			},
		},
	}

	clientPort := tcp.PickPort()
	t.Log("client port", clientPort)
	clientConfig := &configs.TmConfig{
		// Log: &configs.LoggerConfig{
		// 	LogLevel:      configs.Level_DEBUG,
		// 	ConsoleWriter: true,
		// 	ShowColor:     true,
		// 	ShowCaller:    true,
		// },
		InboundManager: &configs.InboundManagerConfig{
			Handlers: []*configs.ProxyInboundConfig{
				{
					Address: net.LocalHostIP.String(),
					Port:    uint32(clientPort),
					Protocol: serial.ToTypedMessage(
						&proxy.DokodemoConfig{
							Address:  tcpDest.Address.String(),
							Port:     uint32(tcpDest.Port),
							Networks: []net.Network{net.Network_TCP},
						},
					),
				},
			},
		},
		Outbound: &configs.OutboundConfig{
			OutboundHandlers: []*configs.OutboundHandlerConfig{
				{
					Address: "127.0.0.1",
					Port:    uint32(serverPort),
					Protocol: serial.ToTypedMessage(&proxy.Hysteria2ClientConfig{
						Auth: userID.String(),
						TlsConfig: &tls.TlsConfig{
							AllowInsecure: true,
							ServerName:    "example.com",
						},
					}),
				},
			},
		},
	}

	server, err := buildserver.NewX(serverConfig)
	common.Must(err)
	common.Must(server.Start(context.Background()))
	defer server.Stop(context.Background())

	client, err := buildclient.NewX(clientConfig)
	common.Must(err)
	common.Must(client.Start())
	defer client.Close()

	var errg errgroup.Group
	for i := 0; i < 1; i++ {
		errg.Go(TestTCPConn(clientPort, 10*1024, time.Second*400))
	}

	if err := errg.Wait(); err != nil {
		t.Error(err)
	}
}

func TestHysteriaUDP(t *testing.T) {
	userID := uuid.New()
	serverPort := net.PickUDPPort()
	serverConfig := &server.ServerConfig{
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Port: uint32(serverPort),
				Protocol: serial.ToTypedMessage(&proxy.Hysteria2ServerConfig{
					IgnoreClientBandwidth: true,
					TlsConfig: &tls.TlsConfig{
						Certificates: []*tls.Certificate{
							tls.ParseCertificate(cert.MustGenerate(nil)),
						},
					},
				}),
			},
		},

		Users: []*configs.UserConfig{
			{
				Id:     userID.String(),
				Secret: userID.String(),
			},
		},
	}

	clientPort := udp.PickPort()
	t.Log("client port", clientPort)
	clientConfig := &configs.TmConfig{
		InboundManager: &configs.InboundManagerConfig{
			Handlers: []*configs.ProxyInboundConfig{
				{
					Address: net.LocalHostIP.String(),
					Port:    uint32(clientPort),
					Protocol: serial.ToTypedMessage(
						&proxy.DokodemoConfig{
							Address:  udpDest.Address.String(),
							Port:     uint32(udpDest.Port),
							Networks: []net.Network{net.Network_UDP},
						},
					),
				},
			},
		},
		Outbound: &configs.OutboundConfig{
			OutboundHandlers: []*configs.OutboundHandlerConfig{
				{
					Address: "127.0.0.1",
					Port:    uint32(serverPort),
					Protocol: serial.ToTypedMessage(&proxy.Hysteria2ClientConfig{
						Auth: userID.String(),
						TlsConfig: &tls.TlsConfig{
							AllowInsecure: true,
						},
					}),
				},
			},
		},
	}

	server, err := buildserver.NewX(serverConfig)
	common.Must(err)
	common.Must(server.Start(context.Background()))
	defer server.Stop(context.Background())

	client, err := buildclient.NewX(clientConfig)
	common.Must(err)
	common.Must(client.Start())
	defer client.Close()

	var errg errgroup.Group
	for i := 0; i < 10; i++ {
		errg.Go(TestUDPConnN(clientPort, 1024, time.Second*4, 10))
	}

	if err := errg.Wait(); err != nil {
		t.Error(err)
	}
}

func TestHysteriaTCPSalamander(t *testing.T) {
	userID := uuid.New()
	serverPort := net.PickUDPPort()
	secret := "1234567890"
	serverConfig := &server.ServerConfig{
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Port: uint32(serverPort),
				Protocol: serial.ToTypedMessage(&proxy.Hysteria2ServerConfig{
					IgnoreClientBandwidth: true,
					TlsConfig: &tls.TlsConfig{
						Certificates: []*tls.Certificate{
							tls.ParseCertificate(cert.MustGenerate(nil)),
						},
					},
					Obfs: &proxy.ObfsConfig{
						Obfs: &proxy.ObfsConfig_Salamander{
							Salamander: &proxy.SalamanderConfig{
								Password: secret,
							},
						},
					},
				}),
			},
		},
		// Router: &configs.RouterConfig{
		// 	Rules: []*configs.RuleConfig{
		// 		{
		// 			OutboundTag: "direct",
		// 		},
		// 	},
		// },
		Users: []*configs.UserConfig{
			{
				Id:     userID.String(),
				Secret: userID.String(),
			},
		},
	}

	clientPort := tcp.PickPort()
	t.Log("client port", clientPort)
	clientConfig := &configs.TmConfig{
		// Log: &configs.LoggerConfig{
		// 	LogLevel:      configs.Level_DEBUG,
		// 	ConsoleWriter: true,
		// 	ShowColor:     true,
		// 	ShowCaller:    true,
		// },
		InboundManager: &configs.InboundManagerConfig{
			Handlers: []*configs.ProxyInboundConfig{
				{
					Address: net.LocalHostIP.String(),
					Port:    uint32(clientPort),
					Protocol: serial.ToTypedMessage(
						&proxy.DokodemoConfig{
							Address:  tcpDest.Address.String(),
							Port:     uint32(tcpDest.Port),
							Networks: []net.Network{net.Network_TCP},
						},
					),
				},
			},
		},
		Outbound: &configs.OutboundConfig{
			OutboundHandlers: []*configs.OutboundHandlerConfig{
				{
					Address: "127.0.0.1",
					Port:    uint32(serverPort),
					Protocol: serial.ToTypedMessage(&proxy.Hysteria2ClientConfig{
						Auth: userID.String(),
						TlsConfig: &tls.TlsConfig{
							AllowInsecure: true,
						},
						Obfs: &proxy.ObfsConfig{
							Obfs: &proxy.ObfsConfig_Salamander{
								Salamander: &proxy.SalamanderConfig{
									Password: secret,
								},
							},
						},
					}),
				},
			},
		},
	}

	server, err := buildserver.NewX(serverConfig)
	common.Must(err)
	common.Must(server.Start(context.Background()))
	defer server.Stop(context.Background())

	client, err := buildclient.NewX(clientConfig)
	common.Must(err)
	common.Must(client.Start())
	defer client.Close()

	var errg errgroup.Group
	for i := 0; i < 1; i++ {
		errg.Go(TestTCPConn(clientPort, 10*1024, time.Second*4))
	}

	if err := errg.Wait(); err != nil {
		t.Error(err)
	}
}

func TestHysteriaECH(t *testing.T) {
	userID := uuid.New()
	serverPort := net.PickUDPPort()

	t.Logf("server  port: %d", serverPort)

	echConfig, echKey, err := util.ExecuteECH("asdf.a")
	common.Must(err)

	serverConfig := &server.ServerConfig{
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Port: uint32(serverPort),
				Protocol: serial.ToTypedMessage(&proxy.Hysteria2ServerConfig{
					IgnoreClientBandwidth: true,
					TlsConfig: &tls.TlsConfig{
						Certificates: []*tls.Certificate{
							tls.ParseCertificate(cert.MustGenerate(nil)),
						},
						EchKey: echKey,
					},
				}),
			},
		},
		Users: []*configs.UserConfig{
			{
				Id:     userID.String(),
				Secret: userID.String(),
			},
		},
	}

	clientPort := tcp.PickPort()
	t.Log("client port", clientPort)
	clientConfig := &configs.TmConfig{
		InboundManager: &configs.InboundManagerConfig{
			Handlers: []*configs.ProxyInboundConfig{
				{
					Address: net.LocalHostIP.String(),
					Port:    uint32(clientPort),
					Protocol: serial.ToTypedMessage(
						&proxy.DokodemoConfig{
							Address:  tcpDest.Address.String(),
							Port:     uint32(tcpDest.Port),
							Networks: []net.Network{net.Network_TCP},
						},
					),
				},
			},
		},
		Outbound: &configs.OutboundConfig{
			OutboundHandlers: []*configs.OutboundHandlerConfig{
				{
					Address: "127.0.0.1",
					Port:    uint32(serverPort),
					Protocol: serial.ToTypedMessage(&proxy.Hysteria2ClientConfig{
						Auth: userID.String(),
						TlsConfig: &tls.TlsConfig{
							AllowInsecure: true,
							EchConfig:     echConfig,
							ServerName:    "example.com",
						},
					}),
				},
			},
		},
	}

	server, err := buildserver.NewX(serverConfig)
	common.Must(err)
	common.Must(server.Start(context.Background()))
	defer server.Stop(context.Background())

	client, err := buildclient.NewX(clientConfig)
	common.Must(err)
	common.Must(client.Start())
	defer client.Close()

	// test.InitZeroLog()

	var errg errgroup.Group
	for i := 0; i < 1; i++ {
		errg.Go(TestTCPConn(clientPort, 10*1024, time.Second*400))
	}

	if err := errg.Wait(); err != nil {
		t.Error(err)
	}
}
