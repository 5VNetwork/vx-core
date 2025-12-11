//go:build test

package scenarios

import (
	"bytes"
	"context"
	"crypto/rand"
	"log"
	"testing"
	"time"

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
	"github.com/5vnetwork/vx-core/common/serial"
	"github.com/5vnetwork/vx-core/common/session"
	"github.com/5vnetwork/vx-core/common/uuid"
	"github.com/5vnetwork/vx-core/test/servers/tcp"
	"github.com/5vnetwork/vx-core/transport"
)

func TestSocksBridgeTCP(t *testing.T) {
	uid := uuid.New()
	psd := uuid.New()
	err := testTcpCommon(testCommonConfig{
		serverProtocol: &proxyconfig.SocksServerConfig{
			Address:    "127.0.0.1",
			UdpEnabled: true,
			AuthType:   proxyconfig.AuthType_PASSWORD,
			Accounts: []*configs.UserConfig{
				{
					Id:     uid.String(),
					Secret: psd.String(),
				},
			},
		},
		clientProtocol: &proxyconfig.SocksClientConfig{
			Name:     uid.String(),
			Password: psd.String(),
		},
	})
	if err != nil {
		t.Error(err)
	}
}

func TestSocksBridgeTCPNoPassword(t *testing.T) {
	err := testTcpCommon(testCommonConfig{
		serverProtocol: &proxyconfig.SocksServerConfig{
			Address:    "127.0.0.1",
			UdpEnabled: true,
			AuthType:   proxyconfig.AuthType_NO_AUTH,
		},
		clientProtocol: &proxyconfig.SocksClientConfig{},
	})
	if err != nil {
		t.Error(err)
	}
}

func TestSocksBridageUDP(t *testing.T) {
	uid := uuid.New()
	psd := uuid.New()
	err := testUdpFlow(testCommonConfig{
		serverProtocol: &proxyconfig.SocksServerConfig{
			Address:    "127.0.0.1",
			UdpEnabled: true,
			AuthType:   proxyconfig.AuthType_PASSWORD,
			Accounts: []*configs.UserConfig{
				{
					Id:     uid.String(),
					Secret: psd.String(),
				},
			},
		},
		clientProtocol: &proxyconfig.SocksClientConfig{
			Name:     uid.String(),
			Password: psd.String(),
		},
	})
	if err != nil {
		t.Error(err)
	}
}

func TestSocksBridageUDPMulti(t *testing.T) {
	userID := protocol.NewID(uuid.New())

	serverPort := tcp.PickPort()
	t.Log("server port", serverPort)
	serverConfig := &server.ServerConfig{
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Address: net.LocalHostIP.String(),
				Port:    uint32(serverPort),
				Protocol: serial.ToTypedMessage(
					&proxyconfig.SocksServerConfig{
						Address:    "127.0.0.1",
						UdpEnabled: true,
						AuthType:   proxyconfig.AuthType_PASSWORD,
						Accounts: []*configs.UserConfig{
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
					Protocol: serial.ToTypedMessage(&proxyconfig.SocksClientConfig{
						Name:     userID.String(),
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
