//go:build test

package scenarios

import (
	"bytes"
	"context"
	"crypto/rand"
	gotls "crypto/tls"
	"log"
	"testing"
	"time"

	"github.com/5vnetwork/vx-core/app/buildclient"
	"github.com/5vnetwork/vx-core/app/buildserver"
	configs "github.com/5vnetwork/vx-core/app/configs"
	proxyconfig "github.com/5vnetwork/vx-core/app/configs/proxy"
	"github.com/5vnetwork/vx-core/app/configs/server"
	"github.com/5vnetwork/vx-core/app/outbound"
	"github.com/5vnetwork/vx-core/app/policy"
	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/net"
	netudp "github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/common/protocol"
	"github.com/5vnetwork/vx-core/common/protocol/tls/cert"
	"github.com/5vnetwork/vx-core/common/serial"
	"github.com/5vnetwork/vx-core/common/session"
	"github.com/5vnetwork/vx-core/common/uuid"
	"github.com/5vnetwork/vx-core/test/servers/tcp"
	"github.com/5vnetwork/vx-core/test/servers/udp"
	"github.com/5vnetwork/vx-core/transport"
	"github.com/5vnetwork/vx-core/transport/protocols/httpupgrade"
	"github.com/5vnetwork/vx-core/transport/protocols/websocket"
	"github.com/5vnetwork/vx-core/transport/security/tls"

	"golang.org/x/sync/errgroup"
)

func TestTrojanTCP(t *testing.T) {
	userID := protocol.NewID(uuid.New())
	serverPort := tcp.PickPort()
	t.Log("server port", serverPort)
	serverConfig := &server.ServerConfig{
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Address: net.LocalHostIP.String(),
				Port:    uint32(serverPort),
				Protocol: serial.ToTypedMessage(
					&proxyconfig.TrojanServerConfig{
						Users: []*configs.UserConfig{
							{
								Id:     userID.String(),
								Secret: userID.String(),
							},
						},
					},
				),
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
					Port:    uint32(serverPort),
					Protocol: serial.ToTypedMessage(&proxyconfig.TrojanClientConfig{
						Password: userID.String(),
					}),
				},
			},
		},
	}

	server, err := buildserver.NewX(serverConfig)
	common.Must(err)

	client, err := buildclient.NewX(clientConfig)
	common.Must(err)

	common.Must(server.Start(context.Background()))
	defer server.Stop(context.Background())
	common.Must(client.Start())
	defer client.Close()
	var errg errgroup.Group
	for i := 0; i < 16; i++ {
		errg.Go(TestTCPConn(clientPort, 10240*1024, Timeout))
	}

	if err := errg.Wait(); err != nil {
		t.Error(err)
	}
}

func TestTrojanUdpFlow(t *testing.T) {
	userID := protocol.NewID(uuid.New())
	serverPort := tcp.PickPort()
	t.Log("server port", serverPort)
	serverConfig := &server.ServerConfig{
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Address: net.LocalHostIP.String(),
				Port:    uint32(serverPort),
				Protocol: serial.ToTypedMessage(
					&proxyconfig.TrojanServerConfig{
						Users: []*configs.UserConfig{
							{
								Id:     userID.String(),
								Secret: userID.String(),
							},
						},
					},
				),
			},
		},
	}

	clientPort := udp.PickPort()
	t.Log("client port", clientPort)
	clientConfig := &configs.TmConfig{
		Policy: &configs.PolicyConfig{
			ConnectionIdleTimeout: 500,
		},
		// Log: &configs.LoggerConfig{
		// 	LogLevel:      configs.Level_DEBUG,
		// 	ConsoleWriter: true,
		// 	ShowCaller:    true,
		// },
		InboundManager: &configs.InboundManagerConfig{
			Handlers: []*configs.ProxyInboundConfig{
				{
					Address: net.LocalHostIP.String(),
					Port:    uint32(clientPort),
					Protocol: serial.ToTypedMessage(
						&proxyconfig.DokodemoConfig{
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
					Address: net.LocalHostIP.String(),
					Port:    uint32(serverPort),
					Protocol: serial.ToTypedMessage(&proxyconfig.TrojanClientConfig{
						Password: userID.String(),
					}),
				},
			},
		},
	}

	server, err := buildserver.NewX(serverConfig)
	common.Must(err)

	client, err := buildclient.NewX(clientConfig)
	common.Must(err)

	common.Must(server.Start(context.Background()))
	defer server.Stop(context.Background())
	common.Must(client.Start())
	defer client.Close()
	var errg errgroup.Group
	for i := 0; i < 16; i++ {
		errg.Go(TestUDPConnN(clientPort, 1024, Timeout, 1024))
	}

	if err := errg.Wait(); err != nil {
		t.Error(err)
	}
}

func TestTrojanFullCone(t *testing.T) {
	userID := protocol.NewID(uuid.New())

	serverPort := tcp.PickPort()
	t.Log("server port", serverPort)
	serverConfig := &server.ServerConfig{
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Address: net.LocalHostIP.String(),
				Port:    uint32(serverPort),
				Protocol: serial.ToTypedMessage(
					&proxyconfig.TrojanServerConfig{
						Users: []*configs.UserConfig{
							{
								Id:     userID.String(),
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
	defer server.Stop(context.Background())

	h, err := outbound.NewHandler(&outbound.HandlerConfig{
		Policy:        policy.DefaultPolicy,
		DialerFactory: transport.DefaultDialerFactory(),
		HandlerConfig: &configs.HandlerConfig{
			Type: &configs.HandlerConfig_Outbound{
				Outbound: &configs.OutboundHandlerConfig{
					Address: net.LocalHostIP.String(),
					Port:    uint32(serverPort),
					Protocol: serial.ToTypedMessage(&proxyconfig.TrojanClientConfig{
						Password: userID.String(),
					}),
				},
			},
		},
	})
	common.Must(err)

	m := StartUdpServers()

	linkA, linkB := netudp.NewLink(100)
	go func() {
		_, ctx, _ := session.NewInfoInbound(session.WithTarget(net.Destination{
			Address: net.AnyIP,
			Network: net.Network_UDP,
		}))
		err := h.HandlePacketConn(ctx, net.Destination{
			Address: net.AnyIP,
			Network: net.Network_UDP,
		}, linkB)
		if err != nil {
			t.Error(err)
		}
	}()

	dstToBuffer := make(map[net.Destination][]byte)
	for _, serverDest := range m {
		payload := make([]byte, 1024)
		common.Must2(rand.Read(payload))
		dstToBuffer[serverDest] = payload
	}

	expectedReceived := 10 * len(m)
	received := 0
	destinationToCount := make(map[net.Destination]int)
	for _, serverDest := range m {
		destinationToCount[serverDest] = 0
	}

	ch := make(chan struct{})
	go func() {
		defer close(ch)
		for {
			packet, err := linkA.ReadPacket()
			if err != nil {
				log.Println("read packet error", err)
				break
			}
			original, ok := dstToBuffer[packet.Source]
			if !ok {
				log.Println("original not found")
				break
			}
			if r := bytes.Compare(original, Xor(packet.Payload.Bytes())); r != 0 {
				log.Println("xor not match")
				break
			}
			destinationToCount[packet.Source]++
			received++
			if received == expectedReceived {
				break
			}
			packet.Payload.Release()
		}
	}()

	sent := 0
	for i := 0; i < 10; i++ {
		for _, serverDest := range m {
			b := buf.New()
			b.Write(dstToBuffer[serverDest])
			err := linkA.WritePacket(&netudp.Packet{
				Target:  serverDest,
				Payload: b,
			})
			if err != nil {
				t.Error(err)
			}
			sent++
			time.Sleep(10 * time.Millisecond)
		}
	}

	select {
	case <-ch:
	case <-time.After(Timeout * 5):
		t.Error("timeout")
	}

	for server := range m {
		server.Close()
	}

	t.Log("sent", sent, "packets", "received", received, "packets")
	for dest, count := range destinationToCount {
		t.Log("destination", dest, "count", count)
	}
}

func TestTrojanTCPTls(t *testing.T) {
	tlsCert := tls.ParseCertificate(cert.MustGenerate(nil))
	goTlsCert, err := gotls.X509KeyPair(tlsCert.Certificate, tlsCert.Key)
	if err != nil {
		t.Fatal(err)
	}
	tcpServer := tcp.Server{
		MsgProcessor: Xor,
		TlsConfig: &gotls.Config{
			Certificates: []gotls.Certificate{
				goTlsCert,
			},
		},
	}
	dest, err := tcpServer.Start()
	t.Log("tcp server listeninsg", dest)
	common.Must(err)
	defer tcpServer.Close()

	userID := protocol.NewID(uuid.New())
	serverPort := tcp.PickPort()
	t.Log("server port", serverPort)
	serverConfig := &server.ServerConfig{
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Address: net.LocalHostIP.String(),
				Port:    uint32(serverPort),
				Protocol: serial.ToTypedMessage(
					&proxyconfig.TrojanServerConfig{
						Users: []*configs.UserConfig{
							{
								Id:     userID.String(),
								Secret: userID.String(),
							},
						},
						Vision: true,
					},
				),
				Transport: &configs.TransportConfig{
					Security: &configs.TransportConfig_Tls{
						Tls: &tls.TlsConfig{
							Certificates: []*tls.Certificate{
								tls.ParseCertificate(cert.MustGenerate(nil)),
							},
						},
					},
				},
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
						&proxyconfig.DokodemoConfig{
							Address:  dest.Address.String(),
							Port:     uint32(dest.Port),
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
					Port:    uint32(serverPort),
					Protocol: serial.ToTypedMessage(&proxyconfig.TrojanClientConfig{
						Password: userID.String(),
						Vision:   true,
					}),
					Transport: &configs.TransportConfig{
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

	server, err := buildserver.NewX(serverConfig)
	common.Must(err)

	client, err := buildclient.NewX(clientConfig)
	common.Must(err)

	common.Must(server.Start(context.Background()))
	defer server.Stop(context.Background())
	common.Must(client.Start())
	defer client.Close()
	var errg errgroup.Group
	for i := 0; i < 1; i++ {
		errg.Go(TestTCPConnTls(clientPort, 10240*1024, Timeout))
	}

	if err := errg.Wait(); err != nil {
		t.Error(err)
	}
}

func TestTrojanWebsocketTCP(t *testing.T) {
	userID := protocol.NewID(uuid.New())
	serverPort := tcp.PickPort()
	t.Log("server port", serverPort)
	serverConfig := &server.ServerConfig{
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Address: net.LocalHostIP.String(),
				Port:    uint32(serverPort),
				Protocol: serial.ToTypedMessage(
					&proxyconfig.TrojanServerConfig{
						Users: []*configs.UserConfig{
							{
								Id:     userID.String(),
								Secret: userID.String(),
							},
						},
					},
				),
				Transport: &configs.TransportConfig{
					Protocol: &configs.TransportConfig_Websocket{
						Websocket: &websocket.WebsocketConfig{
							Path: "/ws",
						},
					},
				},
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
					Port:    uint32(serverPort),
					Protocol: serial.ToTypedMessage(&proxyconfig.TrojanClientConfig{
						Password: userID.String(),
					}),
					Transport: &configs.TransportConfig{
						Protocol: &configs.TransportConfig_Websocket{
							Websocket: &websocket.WebsocketConfig{
								Path: "/ws",
							},
						},
					},
				},
			},
		},
	}

	server, err := buildserver.NewX(serverConfig)
	common.Must(err)

	client, err := buildclient.NewX(clientConfig)
	common.Must(err)

	common.Must(server.Start(context.Background()))
	defer server.Stop(context.Background())
	common.Must(client.Start())
	defer client.Close()
	var errg errgroup.Group
	for i := 0; i < 16; i++ {
		errg.Go(TestTCPConn(clientPort, 10240*1024, Timeout))
	}

	if err := errg.Wait(); err != nil {
		t.Error(err)
	}
}

func TestTrojanHttpUpgradeTCP(t *testing.T) {
	userID := protocol.NewID(uuid.New())
	serverPort := tcp.PickPort()
	t.Log("server port", serverPort)
	serverConfig := &server.ServerConfig{
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Address: net.LocalHostIP.String(),
				Port:    uint32(serverPort),
				Protocol: serial.ToTypedMessage(
					&proxyconfig.TrojanServerConfig{
						Users: []*configs.UserConfig{
							{
								Id:     userID.String(),
								Secret: userID.String(),
							},
						},
					},
				),
				Transport: &configs.TransportConfig{
					Protocol: &configs.TransportConfig_Httpupgrade{
						Httpupgrade: &httpupgrade.HttpUpgradeConfig{
							Config: &websocket.WebsocketConfig{
								Path: "/ws",
							},
						},
					},
				},
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
					Port:    uint32(serverPort),
					Protocol: serial.ToTypedMessage(&proxyconfig.TrojanClientConfig{
						Password: userID.String(),
					}),
					Transport: &configs.TransportConfig{
						Protocol: &configs.TransportConfig_Httpupgrade{
							Httpupgrade: &httpupgrade.HttpUpgradeConfig{
								Config: &websocket.WebsocketConfig{
									Path: "/ws",
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

	client, err := buildclient.NewX(clientConfig)
	common.Must(err)

	common.Must(server.Start(context.Background()))
	defer server.Stop(context.Background())
	common.Must(client.Start())
	defer client.Close()
	var errg errgroup.Group
	for i := 0; i < 16; i++ {
		errg.Go(TestTCPConn(clientPort, 10240*1024, Timeout))
	}

	if err := errg.Wait(); err != nil {
		t.Error(err)
	}
}
