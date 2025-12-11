package scenarios

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/5vnetwork/vx-core/app/buildclient"
	"github.com/5vnetwork/vx-core/app/buildserver"
	"github.com/5vnetwork/vx-core/app/configs"
	proxyconfig "github.com/5vnetwork/vx-core/app/configs/proxy"
	"github.com/5vnetwork/vx-core/app/configs/server"
	"github.com/5vnetwork/vx-core/app/util"
	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/protocol"
	"github.com/5vnetwork/vx-core/common/protocol/tls/cert"
	"github.com/5vnetwork/vx-core/common/serial"
	"github.com/5vnetwork/vx-core/common/uuid"
	"github.com/5vnetwork/vx-core/test"
	"github.com/5vnetwork/vx-core/test/servers/tcp"
	"github.com/5vnetwork/vx-core/transport/protocols/grpc"
	"github.com/5vnetwork/vx-core/transport/protocols/websocket"
	"github.com/5vnetwork/vx-core/transport/security/reality"
	"github.com/5vnetwork/vx-core/transport/security/tls"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/anypb"
)

func TestMultiInbound(t *testing.T) {
	pub, pri, err := util.Curve25519Genkey(false, "")
	if err != nil {
		t.Fatalf("failed to generate curve25519 key: %v", err)
	}
	priByes, err := base64.RawURLEncoding.DecodeString(pri)
	if err != nil {
		t.Fatalf("failed to decode private key: %v", err)
	}

	userID := protocol.NewID(uuid.New())
	serverPort := tcp.PickPort()
	t.Log("server port", serverPort)

	shortId := [8]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	serverConfig := &server.ServerConfig{
		MultiInbounds: []*configs.MultiProxyInboundConfig{
			{
				Address: net.LocalHostIP.String(),
				Ports:   []uint32{uint32(serverPort)},
				Protocols: []*anypb.Any{
					serial.ToTypedMessage(
						&proxyconfig.TrojanServerConfig{
							Users: []*configs.UserConfig{
								{
									Id:     userID.String(),
									Secret: userID.String(),
								},
							},
						},
					),
					serial.ToTypedMessage(
						&proxyconfig.VmessServerConfig{
							Accounts: []*configs.UserConfig{
								{
									Id:     userID.String(),
									Secret: userID.String(),
								},
							},
						},
					),
				},
				TransportProtocols: []*configs.MultiProxyInboundConfig_Protocol{
					{
						Protocol: &configs.MultiProxyInboundConfig_Protocol_Websocket{
							Websocket: &websocket.WebsocketConfig{
								Path: "/ws",
								Host: "www.websocket.com",
							},
						},
						Path: "/ws",
					},
					{
						Protocol: &configs.MultiProxyInboundConfig_Protocol_Grpc{
							Grpc: &grpc.GrpcConfig{
								ServiceName: "grpc",
								Authority:   "www.grpc.com",
							},
						},
						H2: true,
					},
				},
				SecurityConfigs: []*configs.MultiProxyInboundConfig_Security{
					{
						Domains: []string{"www.nvidia.com"},
						Security: &configs.MultiProxyInboundConfig_Security_Reality{
							Reality: &reality.RealityConfig{
								ServerNames: []string{"www.nvidia.com"},
								Dest:        "www.nvidia.com:443",
								PrivateKey:  priByes,
								Show:        true,
								ShortIds:    [][]byte{shortId[:]},
							},
						},
					},
					{
						Domains: []string{"www.test.com"},
						Security: &configs.MultiProxyInboundConfig_Security_Tls{
							Tls: &tls.TlsConfig{
								Certificates: []*tls.Certificate{
									tls.ParseCertificate(cert.MustGenerate(nil, cert.CommonName("www.test.com"),
										cert.DNSNames("www.test.com"))),
								},
							},
						},
					},
				},
			},
		},
	}

	server, err := buildserver.NewX(serverConfig)
	common.Must(err)

	common.Must(server.Start(context.Background()))
	defer server.Stop(context.Background())

	testServer(t, uint32(serverPort), serial.ToTypedMessage(&proxyconfig.VmessClientConfig{
		Id: userID.String(),
	}), nil)

	testServer(t, uint32(serverPort), serial.ToTypedMessage(&proxyconfig.TrojanClientConfig{
		Password: userID.String(),
	}), &configs.TransportConfig{
		Security: &configs.TransportConfig_Reality{
			Reality: &reality.RealityConfig{
				ServerName: "www.nvidia.com",
				Pbk:        pub,
			},
		},
	})

	testServer(t, uint32(serverPort), serial.ToTypedMessage(&proxyconfig.TrojanClientConfig{
		Password: userID.String(),
	}), &configs.TransportConfig{
		Protocol: &configs.TransportConfig_Grpc{
			Grpc: &grpc.GrpcConfig{
				ServiceName: "grpc",
				Authority:   "www.grpc.com",
			},
		},
		Security: &configs.TransportConfig_Tls{
			Tls: &tls.TlsConfig{
				ServerName:    "www.test.com",
				AllowInsecure: true,
			},
		},
	})

	testServer(t, uint32(serverPort), serial.ToTypedMessage(&proxyconfig.TrojanClientConfig{
		Password: userID.String(),
	}), &configs.TransportConfig{
		Protocol: &configs.TransportConfig_Websocket{
			Websocket: &websocket.WebsocketConfig{
				Path: "/ws",
				Host: "www.websocket.com",
			},
		},
		Security: &configs.TransportConfig_Tls{
			Tls: &tls.TlsConfig{
				ServerName:    "www.test.com",
				AllowInsecure: true,
			},
		},
	})
}

func testServer(t *testing.T, serverPort uint32,
	protocol *anypb.Any, transport *configs.TransportConfig) {
	clientPort := tcp.PickPort()
	t.Log("client port", clientPort)
	clientConfig := &configs.TmConfig{
		InboundManager: &configs.InboundManagerConfig{
			Handlers: []*configs.ProxyInboundConfig{
				{
					Address: net.LocalHostIP.String(),
					Ports:   []uint32{uint32(clientPort)},
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
					Address:   net.LocalHostIP.String(),
					Port:      uint32(serverPort),
					Protocol:  protocol,
					Transport: transport,
				},
			},
		},
	}
	client, err := buildclient.NewX(clientConfig)
	common.Must(err)

	test.InitZeroLog()

	common.Must(client.Start())
	defer client.Close()

	var errg errgroup.Group
	for i := 0; i < 4; i++ {
		errg.Go(TestTCPConn(clientPort, 10240*1024, Timeout*2))
	}

	if err := errg.Wait(); err != nil {
		t.Error(err)
	}
}
