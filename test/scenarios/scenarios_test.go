//go:build test

package scenarios

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/5vnetwork/vx-core/app/buildclient"
	"github.com/5vnetwork/vx-core/app/buildserver"
	"github.com/5vnetwork/vx-core/app/configs"
	proxyconfig "github.com/5vnetwork/vx-core/app/configs/proxy"
	"github.com/5vnetwork/vx-core/app/configs/server"
	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/serial"
	"github.com/5vnetwork/vx-core/test/servers/tcp"
	"github.com/5vnetwork/vx-core/test/servers/udp"
	"github.com/miekg/dns"
	"github.com/rs/zerolog"

	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var (
	udpDest net.Destination
	tcpDest net.Destination
	dnsPort uint16
)

func TestMain(m *testing.M) {
	zerolog.SetGlobalLevel(zerolog.Disabled)
	fmt.Println("TestMain")
	// dns server
	var dnsServer *dns.Server
	dnsServer, dnsPort = StartDnsServer()
	log.Printf("the selected dns server port is %d", dnsPort)
	defer dnsServer.Shutdown()
	// tcp server
	var tcpServer tcp.Server
	tcpServer, tcpDest = StartTcpServer()
	log.Print("tcpDest", tcpDest)
	defer tcpServer.Close()
	// udp server
	var udpServer *udp.Server
	udpServer, udpDest = StartUdpServer()
	log.Print("udpDest", udpDest)
	defer udpServer.Close()

	exitVal := m.Run()

	os.Exit(exitVal)
}

type testCommonConfig struct {
	serverProtocol, clientProtocol   protoreflect.ProtoMessage
	serverTransport, clientTransport *configs.TransportConfig
}

func testTcpCommon(config testCommonConfig) error {
	serverPort := tcp.PickPort()
	log.Print("serverPort ", serverPort)
	serverConfig := &server.ServerConfig{
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Address:   net.LocalHostIP.String(),
				Port:      uint32(serverPort),
				Transport: config.serverTransport,
				Protocol: serial.ToTypedMessage(
					config.serverProtocol,
				),
			},
		},
	}
	clientPort := tcp.PickPort()
	log.Print("clientPort ", clientPort)
	clientConfig := &configs.TmConfig{
		InboundManager: &configs.InboundManagerConfig{
			Handlers: []*configs.ProxyInboundConfig{
				{
					Address: net.LocalHostIP.String(),
					Port:    uint32(clientPort),
					Protocol: serial.ToTypedMessage(
						&proxyconfig.DokodemoConfig{
							Address:  "127.0.0.1",
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
					Protocol:  serial.ToTypedMessage(config.clientProtocol),
					Transport: config.clientTransport,
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
		errg.Go(TestTCPConn(clientPort, 10240*1024, Timeout))
	}

	if err := errg.Wait(); err != nil {
		return err
	}
	return nil
}

// one src to one dst
func testUdpFlow(config testCommonConfig) error {
	serverPort := tcp.PickPort()
	log.Print("serverPort ", serverPort)
	serverConfig := &server.ServerConfig{
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Address:   net.LocalHostIP.String(),
				Port:      uint32(serverPort),
				Transport: config.serverTransport,
				Protocol: serial.ToTypedMessage(
					config.serverProtocol,
				),
			},
		},
	}
	clientPort := udp.PickPort()
	log.Print("clientPort ", clientPort)
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
							Address:  "127.0.0.1",
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
					Protocol:  serial.ToTypedMessage(config.clientProtocol),
					Transport: config.clientTransport,
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
		errg.Go(TestUDPConnN(clientPort, 1024, Timeout, 1024))
	}

	if err := errg.Wait(); err != nil {
		return err
	}
	return nil
}
