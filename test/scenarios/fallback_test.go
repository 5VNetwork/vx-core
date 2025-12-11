//go:build test

package scenarios

import (
	"context"
	"testing"

	"github.com/5vnetwork/vx-core/app/buildclient"
	"github.com/5vnetwork/vx-core/app/buildserver"
	configs "github.com/5vnetwork/vx-core/app/configs"
	proxyconfig "github.com/5vnetwork/vx-core/app/configs/proxy"
	"github.com/5vnetwork/vx-core/app/configs/server"
	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/serial"
	"github.com/5vnetwork/vx-core/test/servers/tcp"
	"github.com/5vnetwork/vx-core/test/servers/udp"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/anypb"
)

func TestFallbackSocksHttp(t *testing.T) {
	serverPort := tcp.PickPort()
	t.Log("http,socks server port", serverPort)
	serverConfig := &server.ServerConfig{
		Users: []*configs.UserConfig{
			{
				Secret: secret.String(),
			},
		},
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Tag:     "fallback",
				Address: net.LocalHostIP.String(),
				Port:    uint32(serverPort),
				Protocols: []*anypb.Any{
					serial.ToTypedMessage(
						&proxyconfig.HttpServerConfig{},
					),
					serial.ToTypedMessage(
						&proxyconfig.TrojanServerConfig{},
					),
					serial.ToTypedMessage(
						&proxyconfig.SocksServerConfig{
							UdpEnabled: true,
							Address:    "127.0.0.1",
							AuthType:   proxyconfig.AuthType_NO_AUTH,
						},
					),
					serial.ToTypedMessage(
						&proxyconfig.ShadowsocksServerConfig{
							User: &configs.UserConfig{
								Secret: secret.String(),
							},
							CipherType: proxyconfig.ShadowsocksCipherType_CHACHA20_POLY1305,
						},
					),
					serial.ToTypedMessage(
						&proxyconfig.VmessServerConfig{},
					),
				},
			},
		},
	}

	clientPortTCPSocks := tcp.PickPort()
	t.Log("client port tcp socks", clientPortTCPSocks)
	clientPortTCPHttp := tcp.PickPort()
	t.Log("client port tcp Http", clientPortTCPHttp)
	clientPortUDP := udp.PickPort()
	t.Log("client port udp", clientPortUDP)
	clientPortVmess := tcp.PickPort()
	t.Log("client port vmess", clientPortVmess)
	clientPortTrojan := tcp.PickPort()
	t.Log("client port trojan", clientPortTrojan)
	clientPortShadowsocks := tcp.PickPort()
	t.Log("client port shadowsocks", clientPortShadowsocks)

	clientConfig := &configs.TmConfig{
		InboundManager: &configs.InboundManagerConfig{
			Handlers: []*configs.ProxyInboundConfig{
				{
					Tag:     "tcp doko http",
					Address: net.LocalHostIP.String(),
					Port:    uint32(clientPortTCPHttp),
					Protocol: serial.ToTypedMessage(
						&proxyconfig.DokodemoConfig{
							Address:  "127.0.0.1",
							Port:     uint32(tcpDest.Port),
							Networks: []net.Network{net.Network_TCP},
						},
					),
				},
				{
					Tag:     "tcp doko socks",
					Address: net.LocalHostIP.String(),
					Port:    uint32(clientPortTCPSocks),
					Protocol: serial.ToTypedMessage(
						&proxyconfig.DokodemoConfig{
							Address:  "127.0.0.1",
							Port:     uint32(tcpDest.Port),
							Networks: []net.Network{net.Network_TCP},
						},
					),
				},
				{
					Tag:     "tcp doko vmess",
					Address: net.LocalHostIP.String(),
					Port:    uint32(clientPortVmess),
					Protocol: serial.ToTypedMessage(
						&proxyconfig.DokodemoConfig{
							Address:  "127.0.0.1",
							Port:     uint32(tcpDest.Port),
							Networks: []net.Network{net.Network_TCP},
						},
					),
				},
				{
					Tag:     "tcp doko trojan",
					Address: net.LocalHostIP.String(),
					Port:    uint32(clientPortTrojan),
					Protocol: serial.ToTypedMessage(
						&proxyconfig.DokodemoConfig{
							Address:  "127.0.0.1",
							Port:     uint32(tcpDest.Port),
							Networks: []net.Network{net.Network_TCP},
						},
					),
				},
				{
					Tag:     "tcp doko shadowsocks",
					Address: net.LocalHostIP.String(),
					Port:    uint32(clientPortShadowsocks),
					Protocol: serial.ToTypedMessage(
						&proxyconfig.DokodemoConfig{
							Address:  "127.0.0.1",
							Port:     uint32(tcpDest.Port),
							Networks: []net.Network{net.Network_TCP},
						},
					),
				},
				{
					Tag:     "udp doko",
					Address: net.LocalHostIP.String(),
					Port:    uint32(clientPortUDP),
					Protocol: serial.ToTypedMessage(
						&proxyconfig.DokodemoConfig{
							Address:  "127.0.0.1",
							Port:     uint32(udpDest.Port),
							Networks: []net.Network{net.Network_UDP},
						},
					),
				},
			},
		},
		Router: &configs.RouterConfig{
			Rules: []*configs.RuleConfig{
				{
					InboundTags: []string{"tcp doko http"},
					OutboundTag: "http",
				},
				{
					InboundTags: []string{"tcp doko socks"},
					OutboundTag: "socks",
				},
				{
					InboundTags: []string{"tcp doko vmess"},
					OutboundTag: "vmess",
				},
				{
					InboundTags: []string{"tcp doko trojan"},
					OutboundTag: "trojan",
				},
				{
					InboundTags: []string{"tcp doko shadowsocks"},
					OutboundTag: "shadowsocks",
				},
				{
					Networks: []net.Network{
						net.Network_UDP,
					},
					OutboundTag: "socks",
				},
			},
		},
		Outbound: &configs.OutboundConfig{
			OutboundHandlers: []*configs.OutboundHandlerConfig{
				{
					Tag:      "http",
					Address:  net.LocalHostIP.String(),
					Port:     uint32(serverPort),
					Protocol: serial.ToTypedMessage(&proxyconfig.HttpClientConfig{}),
				},
				{
					Tag:      "socks",
					Address:  net.LocalHostIP.String(),
					Port:     uint32(serverPort),
					Protocol: serial.ToTypedMessage(&proxyconfig.SocksClientConfig{}),
				},
				{
					Tag:     "vmess",
					Address: net.LocalHostIP.String(),
					Port:    uint32(serverPort),
					Protocol: serial.ToTypedMessage(&proxyconfig.VmessClientConfig{
						Id: secret.String(),
					}),
				},
				{
					Tag:     "trojan",
					Address: net.LocalHostIP.String(),
					Port:    uint32(serverPort),
					Protocol: serial.ToTypedMessage(&proxyconfig.TrojanClientConfig{
						Password: secret.String(),
					}),
				},
				{
					Tag:     "shadowsocks",
					Address: net.LocalHostIP.String(),
					Port:    uint32(serverPort),
					Protocol: serial.ToTypedMessage(&proxyconfig.ShadowsocksClientConfig{
						Password:   secret.String(),
						CipherType: proxyconfig.ShadowsocksCipherType_CHACHA20_POLY1305,
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

	// test.InitZeroLog()

	var errg errgroup.Group
	for i := 0; i < 2; i++ {
		errg.Go(TestTCPConn(clientPortTCPSocks, 10240*1024, Timeout))
	}
	for i := 0; i < 2; i++ {
		errg.Go(TestTCPConn(clientPortTCPHttp, 10240*1024, Timeout))
	}
	for i := 0; i < 2; i++ {
		errg.Go(TestUDPConnN(clientPortUDP, 1024, Timeout, 1024))
	}
	for i := 0; i < 2; i++ {
		errg.Go(TestTCPConn(clientPortVmess, 10240*1024, Timeout))
	}
	for i := 0; i < 2; i++ {
		errg.Go(TestTCPConn(clientPortTrojan, 10240*1024, Timeout))
	}
	for i := 0; i < 2; i++ {
		errg.Go(TestTCPConn(clientPortShadowsocks, 10240*1024, Timeout))
	}
	if err := errg.Wait(); err != nil {
		t.Error(err)
	}
}
