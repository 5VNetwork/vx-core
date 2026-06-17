//go:build test

package scenarios

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/5vnetwork/vx-core/app/buildclient"
	"github.com/5vnetwork/vx-core/app/buildserver"
	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/protocol"
	"github.com/5vnetwork/vx-core/common/serial"
	"github.com/5vnetwork/vx-core/common/uuid"
	"github.com/5vnetwork/vx-core/test"
	"github.com/5vnetwork/vx-core/test/servers/tcp"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

func TestTrojanTCPMuxManySessions(t *testing.T) {
	tcpServer := tcp.Server{
		MsgProcessor: Xor,
	}
	dest, err := tcpServer.Start()
	common.Must(err)
	defer tcpServer.Close()

	userID := protocol.NewID(uuid.New())
	serverPort := tcp.PickPort()
	clientPort := tcp.PickPort()

	serverConfig := &configs.ServerConfig{
		Log: &configs.LoggerConfig{
			LogLevel:      configs.Level_DISABLED,
			ConsoleWriter: false,
		},
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Address: net.LocalHostIP.String(),
				Port:    uint32(serverPort),
				Protocol: serial.ToTypedMessage(
					&configs.TrojanServerConfig{
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
		Policy: &configs.PolicyConfig{
			ConnectionIdleTimeout: 0,
		},
	}

	clientConfig := &configs.TmConfig{
		Log: &configs.LoggerConfig{
			LogLevel:      configs.Level_DISABLED,
			ConsoleWriter: false,
		},
		InboundManager: &configs.InboundManagerConfig{
			Handlers: []*configs.ProxyInboundConfig{
				{
					Address: net.LocalHostIP.String(),
					Port:    uint32(clientPort),
					Protocol: serial.ToTypedMessage(
						&configs.DokodemoConfig{
							Address:  dest.Address.String(),
							Port:     uint32(dest.Port),
							Networks: []net.Network{net.Network_TCP},
						},
					),
				},
			},
		},
		Policy: &configs.PolicyConfig{
			ConnectionIdleTimeout: 0,
		},
		Outbound: &configs.OutboundConfig{
			OutboundHandlers: []*configs.OutboundHandlerConfig{
				{
					Address:   net.LocalHostIP.String(),
					Port:      uint32(serverPort),
					EnableMux: true,
					MuxConfig: &configs.MuxConfig{
						MaxConcurrency: 1024,
						MaxConnection:  1024,
					},
					Protocol: serial.ToTypedMessage(
						&configs.TrojanClientConfig{
							Password: userID.String(),
						},
					),
				},
			},
		},
	}

	server, err := buildserver.NewX(serverConfig)
	common.Must(err)
	client, err := buildclient.NewX(clientConfig)
	common.Must(err)

	test.InitZeroLog()

	common.Must(server.Start(context.Background()))
	defer server.Stop(context.Background())
	common.Must(client.Start())
	defer client.Close()

	const (
		totalSessions = 284
		payloadSize   = 4 * 1024
	)

	var remaining atomic.Int32
	remaining.Store(284)
	var errg errgroup.Group
	for i := 0; i < totalSessions; i++ {
		time.Sleep(1 * time.Millisecond)
		errg.Go(func() error {
			err := TestTCPConn(clientPort, payloadSize, Timeout*20)()
			if err != nil {
				return err
			}
			remaining.Add(-1)
			log.Debug().Int32("remaining", remaining.Load()).Msg("test tcp conn")
			return nil
		})
	}
	if err := errg.Wait(); err != nil {
		t.Fatalf("mux stress failed, remaining=%d, err=%v", remaining.Load(), fmt.Errorf("%w", err))
	}
}

func TestTrojanTCPMuxLargePayload(t *testing.T) {
	userID := protocol.NewID(uuid.New())
	serverPort := tcp.PickPort()
	t.Log("server port", serverPort)
	serverConfig := &configs.ServerConfig{
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Address: net.LocalHostIP.String(),
				Port:    uint32(serverPort),
				Protocol: serial.ToTypedMessage(
					&configs.TrojanServerConfig{
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
						&configs.DokodemoConfig{
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
					EnableMux: true,
					MuxConfig: &configs.MuxConfig{
						MaxConcurrency: 1,
						MaxConnection:  16,
					},
					Address: net.LocalHostIP.String(),
					Port:    uint32(serverPort),
					Protocol: serial.ToTypedMessage(&configs.TrojanClientConfig{
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

	test.InitZeroLog()

	common.Must(server.Start(context.Background()))
	defer server.Stop(context.Background())
	common.Must(client.Start())
	defer client.Close()
	var errg errgroup.Group
	for i := 0; i < 16; i++ {
		errg.Go(TestTCPConn(clientPort, 10240*10240, Timeout*4))
	}

	if err := errg.Wait(); err != nil {
		t.Error(err)
	}
}
