//go:build test

package scenarios

import (
	"context"
	gotls "crypto/tls"
	"testing"

	"github.com/5vnetwork/vx-core/app/buildclient"
	"github.com/5vnetwork/vx-core/app/buildserver"
	configs "github.com/5vnetwork/vx-core/app/configs"
	proxyconfig "github.com/5vnetwork/vx-core/app/configs/proxy"
	"github.com/5vnetwork/vx-core/app/configs/server"
	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/protocol"
	"github.com/5vnetwork/vx-core/common/protocol/tls/cert"
	"github.com/5vnetwork/vx-core/common/serial"
	"github.com/5vnetwork/vx-core/common/uuid"
	"github.com/5vnetwork/vx-core/test/servers/tcp"
	"github.com/5vnetwork/vx-core/test/servers/udp"
	"github.com/5vnetwork/vx-core/transport/security/tls"

	"golang.org/x/sync/errgroup"
)

func TestVlessTCP(t *testing.T) {
	userID := protocol.NewID(uuid.New())
	serverPort := tcp.PickPort()
	t.Log("server port", serverPort)
	serverConfig := &server.ServerConfig{
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Address: net.LocalHostIP.String(),
				Port:    uint32(serverPort),
				Protocol: serial.ToTypedMessage(
					&proxyconfig.VlessServerConfig{
						Users: []*configs.UserConfig{
							{
								Id:     userID.String(),
								Secret: userID.String(),
							},
						},
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
					Protocol: serial.ToTypedMessage(&proxyconfig.VlessClientConfig{
						Id:         userID.String(),
						Flow:       "xtls-rprx-vision",
						Encryption: "none",
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
		errg.Go(TestTCPConn(clientPort, 10240*1024, Timeout))
	}

	if err := errg.Wait(); err != nil {
		t.Error(err)
	}
}

func TestVlessUdpFlow(t *testing.T) {
	userID := protocol.NewID(uuid.New())
	serverPort := tcp.PickPort()
	t.Log("server port", serverPort)
	serverConfig := &server.ServerConfig{
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Address: net.LocalHostIP.String(),
				Port:    uint32(serverPort),
				Protocol: serial.ToTypedMessage(
					&proxyconfig.VlessServerConfig{
						Users: []*configs.UserConfig{
							{
								Id:     userID.String(),
								Secret: userID.String(),
							},
						},
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
					Protocol: serial.ToTypedMessage(&proxyconfig.VlessClientConfig{
						Id:         userID.String(),
						Flow:       "xtls-rprx-vision",
						Encryption: "none",
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
	for i := 0; i < 16; i++ {
		errg.Go(TestUDPConnN(clientPort, 1024, Timeout, 1024))
	}

	if err := errg.Wait(); err != nil {
		t.Error(err)
	}
}

func TestVlessTCPTls(t *testing.T) {
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
					&proxyconfig.VlessServerConfig{
						Users: []*configs.UserConfig{
							{
								Id:     userID.String(),
								Secret: userID.String(),
							},
						},
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
					Protocol: serial.ToTypedMessage(&proxyconfig.VlessClientConfig{
						Id:         userID.String(),
						Flow:       "xtls-rprx-vision",
						Encryption: "none",
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
