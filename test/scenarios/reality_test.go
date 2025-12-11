//go:build test

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
	"github.com/5vnetwork/vx-core/common/serial"
	"github.com/5vnetwork/vx-core/common/uuid"
	"github.com/5vnetwork/vx-core/test/servers/tcp"
	"github.com/5vnetwork/vx-core/transport/security/reality"
	"golang.org/x/sync/errgroup"
)

func TestTrojanReality(t *testing.T) {
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
					Security: &configs.TransportConfig_Reality{
						Reality: &reality.RealityConfig{
							ServerNames: []string{"www.nvidia.com"},
							Dest:        "www.nvidia.com:443",
							PrivateKey:  priByes,
							Show:        true,
							ShortIds:    [][]byte{shortId[:]},
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
						Security: &configs.TransportConfig_Reality{
							Reality: &reality.RealityConfig{
								ServerName: "www.nvidia.com",
								Pbk:        pub,
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

	// test.InitZeroLog()

	common.Must(server.Start(context.Background()))
	defer server.Stop(context.Background())
	common.Must(client.Start())
	defer client.Close()

	var errg errgroup.Group
	for i := 0; i < 1; i++ {
		errg.Go(TestTCPConn(clientPort, 10240*1024, Timeout))
	}

	if err := errg.Wait(); err != nil {
		t.Error(err)
	}
}
