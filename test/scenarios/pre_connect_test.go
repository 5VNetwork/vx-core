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
	"github.com/5vnetwork/vx-core/common/protocol"
	"github.com/5vnetwork/vx-core/common/serial"
	"github.com/5vnetwork/vx-core/common/uuid"
	"github.com/5vnetwork/vx-core/test/servers/tcp"

	"golang.org/x/sync/errgroup"
)

func TestPreconnectTCP(t *testing.T) {
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
					// PreConnect: true,
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
