//go:build test

package scenarios

import (
	"bytes"
	"context"
	"crypto/rand"
	"log"
	"testing"
	"time"

	"github.com/5vnetwork/vx-core/app/buildclient"
	"github.com/5vnetwork/vx-core/app/buildserver"
	configs "github.com/5vnetwork/vx-core/app/configs"
	proxyconfig "github.com/5vnetwork/vx-core/app/configs/proxy"
	"github.com/5vnetwork/vx-core/app/configs/server"
	"github.com/5vnetwork/vx-core/app/create"
	"github.com/5vnetwork/vx-core/app/policy"
	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/net"
	netudp "github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/common/protocol"
	"github.com/5vnetwork/vx-core/common/serial"
	"github.com/5vnetwork/vx-core/common/session"
	"github.com/5vnetwork/vx-core/common/uuid"
	"github.com/5vnetwork/vx-core/test/servers/tcp"
	"github.com/5vnetwork/vx-core/test/servers/udp"
	"github.com/5vnetwork/vx-core/transport"

	"golang.org/x/sync/errgroup"
)

func TestShadowsocks2022ChaCha20Poly1305TCP(t *testing.T) {
	tcpServer := tcp.Server{
		MsgProcessor: Xor,
	}
	dest, err := tcpServer.Start()
	common.Must(err)
	defer tcpServer.Close()

	serverPort := tcp.PickPort()
	t.Log("server port", serverPort)
	serverConfig := &server.ServerConfig{
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Address: net.LocalHostIP.String(),
				Port:    uint32(serverPort),
				Protocol: serial.ToTypedMessage(
					&proxyconfig.Shadowsocks2022ServerConfig{
						User: &configs.UserConfig{
							Id:     protocol.NewID(uuid.New()).String(),
							Secret: secret.String(),
						},
						Method:   "2022-blake3-chacha20-poly1305",
						Networks: []net.Network{net.Network_TCP},
					},
				),
			},
		},
	}

	clientPort := tcp.PickPort()
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
					Protocol: serial.ToTypedMessage(&proxyconfig.Shadowsocks2022ClientConfig{
						Method: "2022-blake3-chacha20-poly1305",
						Key:    secret.String(),
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

	var errGroup errgroup.Group
	for i := 0; i < 16; i++ {
		errGroup.Go(TestTCPConn(clientPort, 10240*1024, Timeout))
	}
	if err := errGroup.Wait(); err != nil {
		t.Error(err)
	}
}

func TestShadowsocks2022AES256GCMTCP(t *testing.T) {
	tcpServer := tcp.Server{
		MsgProcessor: Xor,
	}
	dest, err := tcpServer.Start()
	common.Must(err)
	defer tcpServer.Close()

	serverPort := tcp.PickPort()
	t.Log("server port", serverPort)
	serverConfig := &server.ServerConfig{
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Address: net.LocalHostIP.String(),
				Port:    uint32(serverPort),
				Protocol: serial.ToTypedMessage(
					&proxyconfig.Shadowsocks2022ServerConfig{
						User: &configs.UserConfig{
							Id:     protocol.NewID(uuid.New()).String(),
							Secret: secret.String(),
						},
						Method: "2022-blake3-aes-256-gcm",
					},
				),
			},
		},
	}

	clientPort := tcp.PickPort()
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
					Protocol: serial.ToTypedMessage(&proxyconfig.Shadowsocks2022ClientConfig{
						Method: "2022-blake3-aes-256-gcm",
						Key:    secret.String(),
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

	var errGroup errgroup.Group
	for i := 0; i < 16; i++ {
		errGroup.Go(TestTCPConn(clientPort, 10240*1024, Timeout))
	}
	if err := errGroup.Wait(); err != nil {
		t.Error(err)
	}
}

func TestShadowsocks2022AES128GCMUDP(t *testing.T) {
	udpServer := udp.Server{
		MsgProcessor: Xor,
	}
	dest, err := udpServer.Start()
	common.Must(err)
	defer udpServer.Close()

	serverPort := 13245

	t.Log("server port", serverPort)
	serverConfig := &server.ServerConfig{
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Address: net.LocalHostIP.String(),
				Port:    uint32(serverPort),
				Protocol: serial.ToTypedMessage(
					&proxyconfig.Shadowsocks2022ServerConfig{
						User: &configs.UserConfig{
							Id:     protocol.NewID(uuid.New()).String(),
							Secret: secret.String(),
						},
						Method:   "2022-blake3-aes-128-gcm",
						Networks: []net.Network{net.Network_UDP},
					},
				),
			},
		},
	}

	clientPort := udp.PickPort()

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
					Protocol: serial.ToTypedMessage(&proxyconfig.Shadowsocks2022ClientConfig{
						Method: "2022-blake3-aes-128-gcm",
						Key:    secret.String(),
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

	var errGroup errgroup.Group
	for i := 0; i < 16; i++ {
		errGroup.Go(TestUDPConnN(clientPort, 1024, Timeout, 1024))
	}
	if err := errGroup.Wait(); err != nil {
		t.Error(err)
	}
}

func TestShadowsocks2022FullCone(t *testing.T) {
	userID := protocol.NewID(uuid.New())

	serverPort := tcp.PickPort()
	t.Log("server port", serverPort)
	serverConfig := &server.ServerConfig{
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Address: net.LocalHostIP.String(),
				Port:    uint32(serverPort),
				Protocol: serial.ToTypedMessage(
					&proxyconfig.Shadowsocks2022ServerConfig{
						User: &configs.UserConfig{
							Id:     protocol.NewID(uuid.New()).String(),
							Secret: userID.String(),
						},
						Method: "2022-blake3-aes-128-gcm",
					},
				),
			},
		},
	}
	server, err := buildserver.NewX(serverConfig)
	common.Must(err)

	common.Must(server.Start(context.Background()))
	defer server.Stop(context.Background())

	h, err := create.NewHandler(&create.HandlerConfig{
		Policy:        policy.DefaultPolicy,
		DialerFactory: transport.DefaultDialerFactory(),
		HandlerConfig: &configs.HandlerConfig{
			Type: &configs.HandlerConfig_Outbound{
				Outbound: &configs.OutboundHandlerConfig{
					Address: net.LocalHostIP.String(),
					Port:    uint32(serverPort),
					Protocol: serial.ToTypedMessage(&proxyconfig.Shadowsocks2022ClientConfig{
						Method: "2022-blake3-aes-128-gcm",
						Key:    userID.String(),
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

	expectedReceived := 2 * len(m)
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
	for i := 0; i < 2; i++ {
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
