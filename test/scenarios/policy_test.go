//go:build test

package scenarios

import (
	"context"
	"io"
	"testing"

	"github.com/5vnetwork/vx-core/app/buildclient"
	"github.com/5vnetwork/vx-core/app/buildserver"
	"github.com/5vnetwork/vx-core/app/configs"
	proxyconfig "github.com/5vnetwork/vx-core/app/configs/proxy"
	"github.com/5vnetwork/vx-core/app/configs/server"
	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/protocol"
	"github.com/5vnetwork/vx-core/common/serial"
	"github.com/5vnetwork/vx-core/common/uuid"
	"github.com/5vnetwork/vx-core/test/servers/tcp"

	"golang.org/x/sync/errgroup"
)

func startQuickClosingTCPServer() (net.Listener, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				break
			}
			b := make([]byte, 1024)
			conn.Read(b)
			conn.Close()
		}
	}()
	return listener, nil
}

func TestVMessClosing(t *testing.T) {
	tcpServer, err := startQuickClosingTCPServer()
	common.Must(err)
	defer tcpServer.Close()

	dest := net.DestinationFromAddr(tcpServer.Addr())

	userID := protocol.NewID(uuid.New())
	serverPort := tcp.PickPort()
	t.Log("server port", serverPort)
	serverConfig := &server.ServerConfig{
		Policy: &configs.PolicyConfig{
			UpLinkOnlyTimeout:   0,
			DownLinkOnlyTimeout: 0,
		},
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
		Policy: &configs.PolicyConfig{
			UpLinkOnlyTimeout:   0,
			DownLinkOnlyTimeout: 0,
		},
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
					Port:    uint32(serverPort),
					Address: net.LocalHostIP.String(),
					Protocol: serial.ToTypedMessage(&proxyconfig.VmessClientConfig{
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

	if err := TestTCPConn(clientPort, 1024, Timeout*2)(); err != io.EOF {
		t.Error(err)
	}
}

func TestZeroBuffer(t *testing.T) {
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
		Policy: &configs.PolicyConfig{
			UpLinkOnlyTimeout:   0,
			DownLinkOnlyTimeout: 0,
			DefaultBufferSize:   0,
		},
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
	for i := 0; i < 10; i++ {
		errg.Go(TestTCPConn(clientPort, 10240*1024, Timeout))
	}
	if err := errg.Wait(); err != nil {
		t.Error(err)
	}
}
