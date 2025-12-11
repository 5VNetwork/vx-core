package xhttp_test

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/5vnetwork/vx-core/app/buildclient"
	"github.com/5vnetwork/vx-core/app/buildserver"
	"github.com/5vnetwork/vx-core/app/configs"
	proxyconfig "github.com/5vnetwork/vx-core/app/configs/proxy"
	"github.com/5vnetwork/vx-core/app/configs/server"
	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/protocol"
	"github.com/5vnetwork/vx-core/common/protocol/tls/cert"
	"github.com/5vnetwork/vx-core/common/serial"
	"github.com/5vnetwork/vx-core/common/uuid"
	"github.com/5vnetwork/vx-core/test/scenarios"
	"github.com/5vnetwork/vx-core/test/servers/tcp"
	"github.com/5vnetwork/vx-core/test/servers/udp"
	"github.com/5vnetwork/vx-core/transport/protocols/splithttp"
	"github.com/5vnetwork/vx-core/transport/security/tls"
)

var (
	udpDest net.Destination
	tcpDest net.Destination
	dnsPort uint16
)
var userID = protocol.NewID(uuid.New())
var noTls = tcp.PickPort()
var tlsPort = tcp.PickPort()
var realityPort = tcp.PickPort()

func TestMain(m *testing.M) {
	// tcp server
	var tcpServer tcp.Server
	tcpServer, tcpDest = scenarios.StartTcpServer()
	log.Print("tcpDest", tcpDest)
	defer tcpServer.Close()
	// udp server
	var udpServer *udp.Server
	udpServer, udpDest = scenarios.StartUdpServer()
	log.Print("udpDest", udpDest)
	defer udpServer.Close()

	serverConfig := &server.ServerConfig{
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Tag:     "no-tls",
				Address: net.LocalHostIP.String(),
				Port:    uint32(noTls),
				Transport: &configs.TransportConfig{
					Protocol: &configs.TransportConfig_Splithttp{
						Splithttp: &splithttp.SplitHttpConfig{
							Path: "/f",
							Host: "127.0.0.1",
							Mode: "auto",
						},
					},
				},
				Protocol: serial.ToTypedMessage(
					&proxyconfig.TrojanServerConfig{
						Users: []*configs.UserConfig{
							{
								Secret: userID.String(),
							},
						},
					},
				),
			},
			{
				Tag:     "tls",
				Address: net.LocalHostIP.String(),
				Port:    uint32(tlsPort),
				Transport: &configs.TransportConfig{
					Protocol: &configs.TransportConfig_Splithttp{
						Splithttp: &splithttp.SplitHttpConfig{
							Path: "/f",
							Host: "127.0.0.1",
							Mode: "auto",
						},
					},
					Security: &configs.TransportConfig_Tls{
						Tls: &tls.TlsConfig{
							Certificates: []*tls.Certificate{
								tls.ParseCertificate(cert.MustGenerate(nil)),
							},
						},
					},
				},
				Protocol: serial.ToTypedMessage(
					&proxyconfig.TrojanServerConfig{
						Users: []*configs.UserConfig{
							{
								Secret: userID.String(),
							},
						},
					},
				),
			},
		},
	}

	server, err := buildserver.NewX(serverConfig)
	common.Must(err)
	common.Must(server.Start(context.Background()))
	time.Sleep(1000 * time.Millisecond)
	exitVal := m.Run()
	common.Must(server.Stop(context.Background()))
	os.Exit(exitVal)
}

func TestXhttp(t *testing.T) {
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
						&proxyconfig.DokodemoConfig{
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
					Address: net.LocalHostIP.String(),
					Port:    uint32(noTls),
					Protocol: serial.ToTypedMessage(&proxyconfig.TrojanClientConfig{
						Password: userID.String(),
					}),
					Transport: &configs.TransportConfig{
						Protocol: &configs.TransportConfig_Splithttp{
							Splithttp: &splithttp.SplitHttpConfig{
								Path: "/f",
								Host: "127.0.0.1",
								Mode: "stream-one",
								Xmux: &splithttp.XmuxConfig{
									MaxConcurrency: &splithttp.RangeConfig{
										From: 16,
										To:   32,
									},
									HMaxRequestTimes: &splithttp.RangeConfig{
										From: 600,
										To:   900,
									},
									HMaxReusableSecs: &splithttp.RangeConfig{
										From: 1800,
										To:   3000,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	client, err := buildclient.NewX(clientConfig)
	common.Must(err)
	common.Must(client.Start())
	defer client.Close()

	var errg errgroup.Group
	for i := 0; i < 1; i++ {
		errg.Go(scenarios.TestTCPConn(clientPort, 10240*1024, scenarios.Timeout))
	}

	if err := errg.Wait(); err != nil {
		t.Error(err)
	}
}

func TestXhttpPacketUp(t *testing.T) {
	clientPort := tcp.PickPort()
	t.Log("client port", clientPort)
	clientConfig := &configs.TmConfig{
		InboundManager: &configs.InboundManagerConfig{
			Handlers: []*configs.ProxyInboundConfig{
				{
					Address: net.LocalHostIP.String(),
					Port:    uint32(clientPort),
					Protocol: serial.ToTypedMessage(
						&proxyconfig.DokodemoConfig{
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
					Address: net.LocalHostIP.String(),
					Port:    uint32(noTls),
					Protocol: serial.ToTypedMessage(&proxyconfig.TrojanClientConfig{
						Password: userID.String(),
					}),
					Transport: &configs.TransportConfig{
						Protocol: &configs.TransportConfig_Splithttp{
							Splithttp: &splithttp.SplitHttpConfig{
								Path: "/f",
								Host: "127.0.0.1",
								Mode: "packet-up",
							},
						},
					},
				},
			},
		},
	}

	client, err := buildclient.NewX(clientConfig)
	common.Must(err)
	common.Must(client.Start())
	defer client.Close()

	var errg errgroup.Group
	for i := 0; i < 11; i++ {
		errg.Go(scenarios.TestTCPConn(clientPort, 10240*1024, scenarios.Timeout))
	}

	if err := errg.Wait(); err != nil {
		t.Error(err)
	}
}
func TestXhttpSplitTls(t *testing.T) {
	clientPort := tcp.PickPort()
	t.Log("client port", clientPort)
	clientConfig := &configs.TmConfig{
		Log: &configs.LoggerConfig{
			LogLevel:      configs.Level_DEBUG,
			ConsoleWriter: true,
			ShowColor:     true,
			ShowCaller:    true,
		},
		InboundManager: &configs.InboundManagerConfig{
			Handlers: []*configs.ProxyInboundConfig{
				{
					Address: net.LocalHostIP.String(),
					Port:    uint32(clientPort),
					Protocol: serial.ToTypedMessage(
						&proxyconfig.DokodemoConfig{
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
					Address: net.LocalHostIP.String(),
					Port:    uint32(tlsPort),
					Protocol: serial.ToTypedMessage(&proxyconfig.TrojanClientConfig{
						Password: userID.String(),
					}),
					Transport: &configs.TransportConfig{
						Security: &configs.TransportConfig_Tls{
							Tls: &tls.TlsConfig{
								AllowInsecure: true,
							},
						},
						Protocol: &configs.TransportConfig_Splithttp{

							Splithttp: &splithttp.SplitHttpConfig{
								Path: "/f",
								Host: "127.0.0.1",
								Mode: "stream-up",
								Xmux: &splithttp.XmuxConfig{
									MaxConcurrency: &splithttp.RangeConfig{
										From: 16,
										To:   32,
									},
									HMaxRequestTimes: &splithttp.RangeConfig{
										From: 600,
										To:   900,
									},
									HMaxReusableSecs: &splithttp.RangeConfig{
										From: 1800,
										To:   3000,
									},
								},
								DownloadSettings: &splithttp.DownConfig{
									Address: "127.0.0.1",
									Port:    uint32(tlsPort),
									Security: &splithttp.DownConfig_Tls{
										Tls: &tls.TlsConfig{
											AllowInsecure: true,
										},
									},
									XhttpConfig: &splithttp.SplitHttpConfig{
										Path: "/f",
										Host: "127.0.0.1",
										Xmux: &splithttp.XmuxConfig{
											MaxConcurrency: &splithttp.RangeConfig{
												From: 16,
												To:   32,
											},
											HMaxRequestTimes: &splithttp.RangeConfig{
												From: 600,
												To:   900,
											},
											HMaxReusableSecs: &splithttp.RangeConfig{
												From: 1800,
												To:   3000,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	client, err := buildclient.NewX(clientConfig)
	common.Must(err)
	common.Must(client.Start())
	defer client.Close()

	var errg errgroup.Group
	for i := 0; i < 1; i++ {
		errg.Go(scenarios.TestTCPConn(clientPort, 10240*1024, scenarios.Timeout))
	}

	if err := errg.Wait(); err != nil {
		t.Error(err)
	}
}

func TestXhttpTls(t *testing.T) {
	clientPort := tcp.PickPort()
	t.Log("client port", clientPort)
	clientConfig := &configs.TmConfig{
		InboundManager: &configs.InboundManagerConfig{
			Handlers: []*configs.ProxyInboundConfig{
				{
					Address: net.LocalHostIP.String(),
					Port:    uint32(clientPort),
					Protocol: serial.ToTypedMessage(
						&proxyconfig.DokodemoConfig{
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
					Address: net.LocalHostIP.String(),
					Port:    uint32(tlsPort),
					Protocol: serial.ToTypedMessage(&proxyconfig.TrojanClientConfig{
						Password: userID.String(),
					}),
					Transport: &configs.TransportConfig{
						Protocol: &configs.TransportConfig_Splithttp{
							Splithttp: &splithttp.SplitHttpConfig{
								Path: "/f",
								Host: "127.0.0.1",
								Mode: "stream-one",
							},
						},
						Security: &configs.TransportConfig_Tls{
							Tls: &tls.TlsConfig{
								AllowInsecure: true,
							},
						},
					},
				},
			},
		},
	}

	client, err := buildclient.NewX(clientConfig)
	common.Must(err)
	common.Must(client.Start())
	defer client.Close()

	var errg errgroup.Group
	for i := 0; i < 1; i++ {
		errg.Go(scenarios.TestTCPConn(clientPort, 10240*1024, scenarios.Timeout*2))
	}

	if err := errg.Wait(); err != nil {
		t.Error(err)
	}
}
