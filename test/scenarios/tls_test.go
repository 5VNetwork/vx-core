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
	"github.com/5vnetwork/vx-core/app/util"
	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/protocol"
	"github.com/5vnetwork/vx-core/common/protocol/tls/cert"
	"github.com/5vnetwork/vx-core/common/serial"
	"github.com/5vnetwork/vx-core/common/uuid"
	"github.com/5vnetwork/vx-core/test/servers/tcp"
	"github.com/5vnetwork/vx-core/transport/protocols/websocket"
	"github.com/5vnetwork/vx-core/transport/security/tls"
)

func TestSimpleTLSConnection(t *testing.T) {
	tcpServer := tcp.Server{
		MsgProcessor: Xor,
	}
	dest, err := tcpServer.Start()
	common.Must(err)
	defer tcpServer.Close()

	userID := protocol.NewID(uuid.New())
	serverPort := tcp.PickPort()
	serverConfig := &server.ServerConfig{
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Address: net.LocalHostIP.String(),
				Port:    uint32(serverPort),
				Transport: &configs.TransportConfig{
					Security: &configs.TransportConfig_Tls{
						Tls: &tls.TlsConfig{
							Certificates: []*tls.Certificate{
								tls.ParseCertificate(cert.MustGenerate(nil)),
							},
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

						Id: userID.String(),
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

	if err := TestTCPConn(clientPort, 1024, Timeout)(); err != nil {
		t.Fatal(err)
	}
}

func TestSimpleTLSConnectionPinned(t *testing.T) {
	certificateDer := cert.MustGenerate(nil)
	certificate := tls.ParseCertificate(certificateDer)
	certHash := tls.GenerateCertChainHash([][]byte{certificateDer.Certificate})

	userID := protocol.NewID(uuid.New())
	serverPort := tcp.PickPort()
	serverConfig := &server.ServerConfig{
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Port:    uint32(serverPort),
				Address: net.LocalHostIP.String(),
				Transport: &configs.TransportConfig{
					Protocol: &configs.TransportConfig_Websocket{
						Websocket: &websocket.WebsocketConfig{},
					},
					Security: &configs.TransportConfig_Tls{
						Tls: &tls.TlsConfig{
							Certificates: []*tls.Certificate{certificate},
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
					Port:    uint32(clientPort),
					Address: net.LocalHostIP.String(),
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
						Id: userID.String(),
					}),
					Transport: &configs.TransportConfig{
						Protocol: &configs.TransportConfig_Websocket{
							Websocket: &websocket.WebsocketConfig{},
						},
						Security: &configs.TransportConfig_Tls{
							Tls: &tls.TlsConfig{
								AllowInsecure:                    true,
								PinnedPeerCertificateChainSha256: [][]byte{certHash},
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

	if err := TestTCPConn(clientPort, 10240, Timeout)(); err != nil {
		t.Fatal(err)
	}
}

func TestTlsUtls(t *testing.T) {
	tcpServer := tcp.Server{
		MsgProcessor: Xor,
	}
	dest, err := tcpServer.Start()
	common.Must(err)
	defer tcpServer.Close()

	userID := protocol.NewID(uuid.New())
	serverPort := tcp.PickPort()
	serverConfig := &server.ServerConfig{
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Address: net.LocalHostIP.String(),
				Port:    uint32(serverPort),
				Transport: &configs.TransportConfig{
					Security: &configs.TransportConfig_Tls{
						Tls: &tls.TlsConfig{
							Certificates: []*tls.Certificate{
								tls.ParseCertificate(cert.MustGenerate(nil)),
							},
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
						Id: userID.String(),
					}),
					Transport: &configs.TransportConfig{
						Security: &configs.TransportConfig_Tls{
							Tls: &tls.TlsConfig{
								Imitate:       "chrome",
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

	if err := TestTCPConn(clientPort, 1024, Timeout)(); err != nil {
		t.Fatal(err)
	}
}

func TestTlsEch(t *testing.T) {
	tcpServer := tcp.Server{
		MsgProcessor: Xor,
	}
	dest, err := tcpServer.Start()
	common.Must(err)
	defer tcpServer.Close()

	echConfig, echKey, err := util.ExecuteECH("a.c")
	common.Must(err)

	userID := protocol.NewID(uuid.New())
	serverPort := tcp.PickPort()

	t.Logf("server port: %d", serverPort)
	serverConfig := &server.ServerConfig{
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Address: net.LocalHostIP.String(),
				Port:    uint32(serverPort),
				Transport: &configs.TransportConfig{
					Security: &configs.TransportConfig_Tls{
						Tls: &tls.TlsConfig{
							Certificates: []*tls.Certificate{
								tls.ParseCertificate(cert.MustGenerate(nil)),
							},
							EchKey: echKey,
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

						Id: userID.String(),
					}),
					Transport: &configs.TransportConfig{
						Security: &configs.TransportConfig_Tls{
							Tls: &tls.TlsConfig{
								AllowInsecure: true,
								EchConfig:     echConfig,
								ServerName:    "asdfexample.com",
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

	if err := TestTCPConn(clientPort, 1024, Timeout)(); err != nil {
		t.Fatal(err)
	}
}
