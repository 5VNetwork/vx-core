//go:build test

package scenarios

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/5vnetwork/vx-core/app/buildclient"
	"github.com/5vnetwork/vx-core/app/buildserver"
	configs "github.com/5vnetwork/vx-core/app/configs"
	proxyconfig "github.com/5vnetwork/vx-core/app/configs/proxy"
	"github.com/5vnetwork/vx-core/app/configs/server"
	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/protocol"
	"github.com/5vnetwork/vx-core/common/serial"
	"github.com/5vnetwork/vx-core/common/uuid"
	"github.com/5vnetwork/vx-core/test"
	"github.com/5vnetwork/vx-core/test/servers/tcp"
	"github.com/5vnetwork/vx-core/test/servers/udp"
	"github.com/5vnetwork/vx-core/transport/protocols/kcp"

	"golang.org/x/sync/errgroup"
)

var clientPort = tcp.PickPort()

func getConfigs(security proxyconfig.SecurityType) (serverConfig *server.ServerConfig, clientConfig *configs.TmConfig) {
	userID := protocol.NewID(uuid.New())
	serverPort := tcp.PickPort()
	log.Printf("server port: %d", serverPort)
	serverConfig = &server.ServerConfig{
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Address: net.LocalHostIP.String(),
				Port:    uint32(serverPort),
				Protocol: serial.ToTypedMessage(
					&proxyconfig.VmessServerConfig{
						Accounts: []*configs.UserConfig{
							{
								Secret: userID.String(),
							},
						},
					},
				),
			},
		},
	}

	log.Printf("client port: %d", clientPort)
	clientConfig = &configs.TmConfig{
		// Log: &configs.LoggerConfig{
		// 	LogLevel: configs.Level_INFO,
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
					Port:    uint32(serverPort),
					Protocol: serial.ToTypedMessage(&proxyconfig.VmessClientConfig{
						Id:       userID.String(),
						Security: security,
					}),
				},
			},
		},
	}
	return
}

func TestVMessGCM(t *testing.T) {
	serverConfig, clientConfig := getConfigs(proxyconfig.SecurityType_SecurityType_AES128_GCM)

	server, err := buildserver.NewX(serverConfig)
	common.Must(err)
	client, err := buildclient.NewX(clientConfig)
	common.Must(err)
	common.Must(server.Start(context.Background()))
	defer server.Stop(context.Background())
	common.Must(client.Start())
	defer client.Close()

	var errg errgroup.Group
	for i := 0; i < 10; i++ {
		errg.Go(TestTCPConn(clientPort, 10240*1024, time.Second*4))
	}

	if err := errg.Wait(); err != nil {
		t.Error(err)
	}
}

func TestVMessGCMReady(t *testing.T) {
	serverConfig, clientConfig := getConfigs(proxyconfig.SecurityType_SecurityType_AES128_GCM)

	server, err := buildserver.NewX(serverConfig)
	common.Must(err)

	client, err := buildclient.NewX(clientConfig)
	common.Must(err)

	const envName = "V2RAY_BUF_READV"
	common.Must(os.Setenv(envName, "enable"))
	defer os.Unsetenv(envName)

	common.Must(server.Start(context.Background()))
	defer server.Stop(context.Background())

	common.Must(client.Start())
	defer client.Close()

	var errg errgroup.Group
	for i := 0; i < 10; i++ {
		errg.Go(TestTCPConn(clientPort, 10240*1024, Timeout))
	}

	if err := errg.Wait(); err != nil {
		t.Error(err)
	}
}

func TestVMessGCMUDP(t *testing.T) {
	userID := protocol.NewID(uuid.New())
	serverPort := tcp.PickPort()
	t.Log("server port", serverPort)
	serverConfig := &server.ServerConfig{
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Address: net.LocalHostIP.String(),
				Port:    uint32(serverPort),
				Protocol: serial.ToTypedMessage(
					&proxyconfig.VmessServerConfig{
						Accounts: []*configs.UserConfig{
							{
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
					Protocol: serial.ToTypedMessage(&proxyconfig.VmessClientConfig{
						Special:  true,
						Id:       userID.String(),
						Security: proxyconfig.SecurityType_SecurityType_AES128_GCM,
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
	for i := 0; i < 20; i++ {
		errg.Go(TestUDPConn(clientPort, 1024, Timeout))
	}

	if err := errg.Wait(); err != nil {
		t.Error(err)
	}
}

func TestVMessChacha20(t *testing.T) {
	tcpServer := tcp.Server{
		MsgProcessor: Xor,
	}
	dest, err := tcpServer.Start()
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
					&proxyconfig.VmessServerConfig{
						Accounts: []*configs.UserConfig{
							{
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
					Protocol: serial.ToTypedMessage(&proxyconfig.VmessClientConfig{
						Special:  true,
						Id:       userID.String(),
						Security: proxyconfig.SecurityType_SecurityType_CHACHA20_POLY1305,
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
	for i := 0; i < 10; i++ {
		errg.Go(TestTCPConn(clientPort, 10240*1024, Timeout))
	}

	if err := errg.Wait(); err != nil {
		t.Error(err)
	}
}

func TestVMessNone(t *testing.T) {
	tcpServer := tcp.Server{
		MsgProcessor: Xor,
	}
	dest, err := tcpServer.Start()
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
					&proxyconfig.VmessServerConfig{
						Accounts: []*configs.UserConfig{
							{
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
					Protocol: serial.ToTypedMessage(&proxyconfig.VmessClientConfig{
						Special:  true,
						Id:       userID.String(),
						Security: proxyconfig.SecurityType_SecurityType_NONE,
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
	for i := 0; i < 10; i++ {
		errg.Go(TestTCPConn(clientPort, 10240*1024, Timeout))
	}

	if err := errg.Wait(); err != nil {
		t.Error(err)
	}
}

func TestVMessKCP(t *testing.T) {
	tcpServer := tcp.Server{
		MsgProcessor: Xor,
	}
	dest, err := tcpServer.Start()
	common.Must(err)
	defer tcpServer.Close()

	userID := protocol.NewID(uuid.New())
	serverPort := udp.PickPort()
	t.Log("server port", serverPort)
	serverConfig := &server.ServerConfig{
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Address: net.LocalHostIP.String(),
				Port:    uint32(serverPort),
				Transport: &configs.TransportConfig{
					Protocol: &configs.TransportConfig_Kcp{
						Kcp: &kcp.KcpConfig{},
					},
				},
				Protocol: serial.ToTypedMessage(
					&proxyconfig.VmessServerConfig{
						Accounts: []*configs.UserConfig{
							{
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
					Protocol: serial.ToTypedMessage(&proxyconfig.VmessClientConfig{
						Special:  true,
						Id:       userID.String(),
						Security: proxyconfig.SecurityType_SecurityType_NONE,
					}),
					Transport: &configs.TransportConfig{
						Protocol: &configs.TransportConfig_Kcp{
							Kcp: &kcp.KcpConfig{},
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
	for i := 0; i < 5; i++ {
		errg.Go(TestTCPConn(clientPort, 1024*1024, Timeout))
	}

	if err := errg.Wait(); err != nil {
		t.Error(err)
	}
}

func TestVMessKCPLarge(t *testing.T) {
	tcpServer := tcp.Server{
		MsgProcessor: Xor,
	}
	dest, err := tcpServer.Start()
	common.Must(err)
	defer tcpServer.Close()

	userID := protocol.NewID(uuid.New())
	serverPort := udp.PickPort()
	t.Log("server port", serverPort)
	serverConfig := &server.ServerConfig{
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Address: net.LocalHostIP.String(),
				Port:    uint32(serverPort),
				Transport: &configs.TransportConfig{
					Protocol: &configs.TransportConfig_Kcp{
						Kcp: &kcp.KcpConfig{
							ReadBuffer:       512 * 1024,
							WriteBuffer:      512 * 1024,
							UplinkCapacity:   20,
							DownlinkCapacity: 20,
						},
					},
				},
				Protocol: serial.ToTypedMessage(
					&proxyconfig.VmessServerConfig{
						Accounts: []*configs.UserConfig{
							{
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
					Protocol: serial.ToTypedMessage(&proxyconfig.VmessClientConfig{

						Special:  true,
						Id:       userID.String(),
						Security: proxyconfig.SecurityType_SecurityType_NONE,
					}),
					Transport: &configs.TransportConfig{
						Protocol: &configs.TransportConfig_Kcp{
							Kcp: &kcp.KcpConfig{
								ReadBuffer:       512 * 1024,
								WriteBuffer:      512 * 1024,
								UplinkCapacity:   20,
								DownlinkCapacity: 20,
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
	for i := 0; i < 2; i++ {
		errg.Go(TestTCPConn(clientPort, 10240*1024, Timeout*2))
	}

	if err := errg.Wait(); err != nil {
		t.Error(err)
	}
}

func TestVMessGCMMux(t *testing.T) {
	tcpServer := tcp.Server{
		MsgProcessor: Xor,
	}
	dest, err := tcpServer.Start()
	common.Must(err)
	defer tcpServer.Close()

	udpServer := udp.Server{
		MsgProcessor: Xor,
	}
	udpDest, err := udpServer.Start()
	common.Must(err)
	defer udpServer.Close()

	userID := protocol.NewID(uuid.New())
	serverPort := tcp.PickPort()
	t.Log("server port", serverPort)
	serverConfig := &server.ServerConfig{
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Address: net.LocalHostIP.String(),
				Port:    uint32(serverPort),
				Protocol: serial.ToTypedMessage(
					&proxyconfig.VmessServerConfig{
						Accounts: []*configs.UserConfig{
							{
								Secret: userID.String(),
							},
						},
					},
				),
			},
		},
	}

	clientPort := tcp.PickPort()
	clientUDPPort := udp.PickPort()
	t.Log("client port", clientPort)
	t.Log("client udp port", clientUDPPort)
	clientConfig := &configs.TmConfig{
		InboundManager: &configs.InboundManagerConfig{
			Handlers: []*configs.ProxyInboundConfig{
				{
					Tag:     "docodemo-tcp",
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
				{
					Tag:     "docodemo-udp",
					Address: net.LocalHostIP.String(),
					Port:    uint32(clientUDPPort),
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
					Address:   net.LocalHostIP.String(),
					Port:      uint32(serverPort),
					EnableMux: true,
					Protocol: serial.ToTypedMessage(&proxyconfig.VmessClientConfig{

						Special:  true,
						Id:       userID.String(),
						Security: proxyconfig.SecurityType_SecurityType_AES128_GCM,
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

	for range "aBCD" {
		var errg errgroup.Group
		for i := 0; i < 16; i++ {
			errg.Go(TestTCPConn(clientPort, 10240, Timeout))
			errg.Go(TestUDPConnN(clientUDPPort, 1024, Timeout, 10))
		}
		if err := errg.Wait(); err != nil {
			t.Fatal(err)
		}
	}
	<-time.After(1 * time.Second)
}

func TestVMessZero(t *testing.T) {
	tcpServer := tcp.Server{
		MsgProcessor: Xor,
	}
	dest, err := tcpServer.Start()
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
					&proxyconfig.VmessServerConfig{
						Accounts: []*configs.UserConfig{
							{
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
					Protocol: serial.ToTypedMessage(&proxyconfig.VmessClientConfig{

						Special:  true,
						Id:       userID.String(),
						Security: proxyconfig.SecurityType_SecurityType_ZERO,
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
	for i := 0; i < 10; i++ {
		errg.Go(TestTCPConn(clientPort, 1024*1024, Timeout))
	}

	if err := errg.Wait(); err != nil {
		t.Error(err)
	}
}

// TODO
func TestVMessGCMLengthAuth(t *testing.T) {
	tcpServer := tcp.Server{
		MsgProcessor: Xor,
	}
	dest, err := tcpServer.Start()
	common.Must(err)
	defer tcpServer.Close()

	userID := protocol.NewID(uuid.New())
	serverPort := tcp.PickPort()
	t.Log("server port", serverPort)
	serverConfig := &server.ServerConfig{
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Port:    uint32(serverPort),
				Address: net.LocalHostIP.String(),
				Protocol: serial.ToTypedMessage(
					&proxyconfig.VmessServerConfig{
						Accounts: []*configs.UserConfig{
							{
								Secret: userID.String(),
								// TestsEnabled: "AuthenticatedLength|NoTerminationSignal",
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
					Port:    uint32(clientPort),
					Address: net.LocalHostIP.String(),
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
					Protocol: serial.ToTypedMessage(&proxyconfig.VmessClientConfig{
						Special:  true,
						Id:       userID.String(),
						Security: proxyconfig.SecurityType_SecurityType_AES128_GCM,
						// TestsEnabled: "AuthenticatedLength|NoTerminationSignal",
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
	for i := 0; i < 10; i++ {
		errg.Go(TestTCPConn(clientPort, test.OneMB, Timeout))
	}

	if err := errg.Wait(); err != nil {
		t.Error(err)
	}
}
