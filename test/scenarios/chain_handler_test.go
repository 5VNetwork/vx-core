//go:build test

package scenarios

import (
	"context"
	"testing"

	"github.com/5vnetwork/vx-core/app/buildclient"
	"github.com/5vnetwork/vx-core/app/buildserver"
	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/configs/proxy"
	"github.com/5vnetwork/vx-core/app/configs/server"
	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/protocol/tls/cert"
	"github.com/5vnetwork/vx-core/common/serial"
	"github.com/5vnetwork/vx-core/common/uuid"
	"github.com/5vnetwork/vx-core/test/servers/tcp"
	"github.com/5vnetwork/vx-core/transport/protocols/websocket"
	"github.com/5vnetwork/vx-core/transport/security/tls"
	"golang.org/x/sync/errgroup"
)

func TestChainHanler(t *testing.T) {
	socksPort := net.PickTCPPort()
	ssPort := net.PickTCPPort()
	trojanPort := net.PickTCPPort()
	t.Log("socksPort", socksPort, "ssPort", ssPort, "trojanPort", trojanPort)
	secret := uuid.New()
	serverConfig := &server.ServerConfig{
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Tag:     "socks",
				Address: net.LocalHostIP.String(),
				Port:    uint32(socksPort),
				Protocol: serial.ToTypedMessage(&proxy.SocksServerConfig{
					AuthType: proxy.AuthType_NO_AUTH,
				}),
			},
			{
				Tag:     "trojan",
				Address: net.LocalHostIP.String(),
				Port:    uint32(trojanPort),
				Protocol: serial.ToTypedMessage(&proxy.TrojanServerConfig{
					Users: []*configs.UserConfig{
						{
							Id:     secret.String(),
							Secret: secret.String(),
						},
					},
				}),
			},
			{
				Tag:     "ss",
				Address: net.LocalHostIP.String(),
				Port:    uint32(ssPort),
				Protocol: serial.ToTypedMessage(
					&proxy.ShadowsocksServerConfig{
						User: &configs.UserConfig{
							Secret: secret.String(),
						},
						CipherType: proxy.ShadowsocksCipherType_CHACHA20_POLY1305,
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
						&proxy.DokodemoConfig{
							Address:  tcpDest.Address.String(),
							Port:     uint32(tcpDest.Port),
							Networks: []net.Network{net.Network_TCP},
						},
					),
				},
			},
		},
		Outbound: &configs.OutboundConfig{
			ChainHandlers: []*configs.ChainHandlerConfig{
				{
					Handlers: []*configs.OutboundHandlerConfig{
						{
							Address: net.LocalHostIP.String(),
							Port:    uint32(ssPort),
							Protocol: serial.ToTypedMessage(&proxy.ShadowsocksClientConfig{
								Password:   secret.String(),
								CipherType: proxy.ShadowsocksCipherType_CHACHA20_POLY1305,
							}),
						},
						{
							Address: net.LocalHostIP.String(),
							Port:    uint32(trojanPort),
							Protocol: serial.ToTypedMessage(&proxy.TrojanClientConfig{
								Password: secret.String(),
							}),
						},
						{
							Address:  net.LocalHostIP.String(),
							Port:     uint32(socksPort),
							Protocol: serial.ToTypedMessage(&proxy.SocksClientConfig{}),
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

	var errGroup errgroup.Group
	for i := 0; i < 10; i++ {
		errGroup.Go(TestTCPConn(clientPort, 10240*1024, Timeout))
	}
	if err := errGroup.Wait(); err != nil {
		t.Error(err)
	}
}

func TestChainHanlerTrojanTlsTrojanTls(t *testing.T) {
	trojan1 := net.PickTCPPort()
	trojan2 := net.PickTCPPort()
	trojanPort := net.PickTCPPort()
	t.Log("trojan1", trojan1, "trojan2", trojanPort)
	secret := uuid.New()
	serverConfig := &server.ServerConfig{
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Tag:     "trojan1",
				Address: net.LocalHostIP.String(),
				Port:    uint32(trojan1),
				Protocol: serial.ToTypedMessage(&proxy.TrojanServerConfig{
					Users: []*configs.UserConfig{
						{
							Id:     secret.String(),
							Secret: secret.String(),
						},
					},
				}),
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
			{
				Tag:     "trojan2",
				Address: net.LocalHostIP.String(),
				Port:    uint32(trojan2),
				Protocol: serial.ToTypedMessage(&proxy.TrojanServerConfig{
					Users: []*configs.UserConfig{
						{
							Id:     secret.String(),
							Secret: secret.String(),
						},
					},
				}),
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
						&proxy.DokodemoConfig{
							Address:  tcpDest.Address.String(),
							Port:     uint32(tcpDest.Port),
							Networks: []net.Network{net.Network_TCP},
						},
					),
				},
			},
		},
		Outbound: &configs.OutboundConfig{
			ChainHandlers: []*configs.ChainHandlerConfig{
				{
					Handlers: []*configs.OutboundHandlerConfig{
						{
							Address: net.LocalHostIP.String(),
							Port:    uint32(trojan1),
							Protocol: serial.ToTypedMessage(&proxy.TrojanClientConfig{
								Password: secret.String(),
							}),
							Transport: &configs.TransportConfig{
								Security: &configs.TransportConfig_Tls{
									Tls: &tls.TlsConfig{
										AllowInsecure: true,
									},
								},
							},
						},
						{
							Address: net.LocalHostIP.String(),
							Port:    uint32(trojan2),
							Protocol: serial.ToTypedMessage(&proxy.TrojanClientConfig{
								Password: secret.String(),
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

	var errGroup errgroup.Group
	for i := 0; i < 11; i++ {
		errGroup.Go(TestTCPConn(clientPort, 1024*1024, Timeout))
	}
	if err := errGroup.Wait(); err != nil {
		t.Error(err)
	}
}

func TestChainHanler1(t *testing.T) {
	socksPort := net.PickTCPPort()
	ssPort := net.PickTCPPort()
	trojanPort := net.PickTCPPort()
	t.Log("socksPort", socksPort, "ssPort", ssPort, "trojanPort", trojanPort)
	secret := uuid.New()
	serverConfig := &server.ServerConfig{
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Tag:     "socks",
				Address: net.LocalHostIP.String(),
				Port:    uint32(socksPort),
				Protocol: serial.ToTypedMessage(&proxy.SocksServerConfig{
					AuthType: proxy.AuthType_NO_AUTH,
				}),
			},
			{
				Tag:     "trojan",
				Address: net.LocalHostIP.String(),
				Port:    uint32(trojanPort),
				Protocol: serial.ToTypedMessage(&proxy.TrojanServerConfig{
					Users: []*configs.UserConfig{
						{
							Id:     secret.String(),
							Secret: secret.String(),
						},
					},
				}),
			},
			{
				Tag:     "ss",
				Address: net.LocalHostIP.String(),
				Port:    uint32(ssPort),
				Protocol: serial.ToTypedMessage(
					&proxy.ShadowsocksServerConfig{
						User: &configs.UserConfig{
							Secret: secret.String(),
						},
						CipherType: proxy.ShadowsocksCipherType_CHACHA20_POLY1305,
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
						&proxy.DokodemoConfig{
							Address:  tcpDest.Address.String(),
							Port:     uint32(tcpDest.Port),
							Networks: []net.Network{net.Network_TCP},
						},
					),
				},
			},
		},
		Outbound: &configs.OutboundConfig{
			ChainHandlers: []*configs.ChainHandlerConfig{
				{
					Handlers: []*configs.OutboundHandlerConfig{
						{
							Address:  net.LocalHostIP.String(),
							Port:     uint32(socksPort),
							Protocol: serial.ToTypedMessage(&proxy.SocksClientConfig{}),
						},
						{
							Address: net.LocalHostIP.String(),
							Port:    uint32(trojanPort),
							Protocol: serial.ToTypedMessage(&proxy.TrojanClientConfig{
								Password: secret.String(),
							}),
						},

						{
							Address: net.LocalHostIP.String(),
							Port:    uint32(ssPort),
							Protocol: serial.ToTypedMessage(&proxy.ShadowsocksClientConfig{
								Password:   secret.String(),
								CipherType: proxy.ShadowsocksCipherType_CHACHA20_POLY1305,
							}),
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

	var errGroup errgroup.Group
	for i := 0; i < 10; i++ {
		errGroup.Go(TestTCPConn(clientPort, 10240*1024, Timeout))
	}
	if err := errGroup.Wait(); err != nil {
		t.Error(err)
	}
}

func TestChainHanlerTransport(t *testing.T) {
	socksPort := net.PickTCPPort()
	ssPort := net.PickTCPPort()
	trojanPort := net.PickTCPPort()
	t.Log("socksPort", socksPort, "ssPort", ssPort, "trojanPort", trojanPort)
	secret := uuid.New()
	serverConfig := &server.ServerConfig{
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Tag:     "socks",
				Address: net.LocalHostIP.String(),
				Port:    uint32(socksPort),
				Protocol: serial.ToTypedMessage(&proxy.SocksServerConfig{
					AuthType: proxy.AuthType_NO_AUTH,
				}),
			},
			{
				Tag:     "ss",
				Address: net.LocalHostIP.String(),
				Port:    uint32(ssPort),
				Transport: &configs.TransportConfig{
					Protocol: &configs.TransportConfig_Websocket{
						Websocket: &websocket.WebsocketConfig{
							Path: "/ws",
						},
					},
				},
				Protocol: serial.ToTypedMessage(
					&proxy.ShadowsocksServerConfig{
						User: &configs.UserConfig{
							Secret: secret.String(),
						},
						CipherType: proxy.ShadowsocksCipherType_CHACHA20_POLY1305,
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
						&proxy.DokodemoConfig{
							Address:  tcpDest.Address.String(),
							Port:     uint32(tcpDest.Port),
							Networks: []net.Network{net.Network_TCP},
						},
					),
				},
			},
		},
		Outbound: &configs.OutboundConfig{
			ChainHandlers: []*configs.ChainHandlerConfig{
				{
					Handlers: []*configs.OutboundHandlerConfig{
						{
							Address: net.LocalHostIP.String(),
							Port:    uint32(ssPort),
							Protocol: serial.ToTypedMessage(&proxy.ShadowsocksClientConfig{
								Password:   secret.String(),
								CipherType: proxy.ShadowsocksCipherType_CHACHA20_POLY1305,
							}),
							Transport: &configs.TransportConfig{
								Protocol: &configs.TransportConfig_Websocket{
									Websocket: &websocket.WebsocketConfig{
										Path: "/ws",
									},
								},
							},
						},
						{
							Address:  net.LocalHostIP.String(),
							Port:     uint32(socksPort),
							Protocol: serial.ToTypedMessage(&proxy.SocksClientConfig{}),
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

	var errGroup errgroup.Group
	for i := 0; i < 1; i++ {
		errGroup.Go(TestTCPConn(clientPort, 10240*1024, Timeout))
	}
	if err := errGroup.Wait(); err != nil {
		t.Error(err)
	}
}

func TestChainHanlerTransport1(t *testing.T) {
	socksPort := net.PickTCPPort()
	ssPort := net.PickTCPPort()
	trojanPort := net.PickTCPPort()
	t.Log("socksPort", socksPort, "ssPort", ssPort, "trojanPort", trojanPort)
	secret := uuid.New()
	serverConfig := &server.ServerConfig{
		Policy: &configs.PolicyConfig{
			HandshakeTimeout:      10000,
			ConnectionIdleTimeout: 10000,
		},
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Tag:     "socks",
				Address: net.LocalHostIP.String(),
				Port:    uint32(socksPort),
				Protocol: serial.ToTypedMessage(&proxy.SocksServerConfig{
					AuthType: proxy.AuthType_NO_AUTH,
				}),
			},
			{
				Tag:     "ss",
				Address: net.LocalHostIP.String(),
				Port:    uint32(ssPort),
				Transport: &configs.TransportConfig{
					Protocol: &configs.TransportConfig_Websocket{
						Websocket: &websocket.WebsocketConfig{
							Path: "/ws",
						},
					},
				},
				Protocol: serial.ToTypedMessage(
					&proxy.ShadowsocksServerConfig{
						User: &configs.UserConfig{
							Secret: secret.String(),
						},
						CipherType: proxy.ShadowsocksCipherType_CHACHA20_POLY1305,
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
		Policy: &configs.PolicyConfig{
			HandshakeTimeout:      10000,
			ConnectionIdleTimeout: 10000,
		},
		InboundManager: &configs.InboundManagerConfig{
			Handlers: []*configs.ProxyInboundConfig{
				{
					Address: net.LocalHostIP.String(),
					Port:    uint32(clientPort),
					Protocol: serial.ToTypedMessage(
						&proxy.DokodemoConfig{
							Address:  tcpDest.Address.String(),
							Port:     uint32(tcpDest.Port),
							Networks: []net.Network{net.Network_TCP},
						},
					),
				},
			},
		},
		Outbound: &configs.OutboundConfig{
			ChainHandlers: []*configs.ChainHandlerConfig{
				{
					Handlers: []*configs.OutboundHandlerConfig{
						{
							Address:  net.LocalHostIP.String(),
							Port:     uint32(socksPort),
							Protocol: serial.ToTypedMessage(&proxy.SocksClientConfig{}),
						},
						{
							Address: net.LocalHostIP.String(),
							Port:    uint32(ssPort),
							Protocol: serial.ToTypedMessage(&proxy.ShadowsocksClientConfig{
								Password:   secret.String(),
								CipherType: proxy.ShadowsocksCipherType_CHACHA20_POLY1305,
							}),
							Transport: &configs.TransportConfig{
								Protocol: &configs.TransportConfig_Websocket{
									Websocket: &websocket.WebsocketConfig{
										Path: "/ws",
									},
								},
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

	var errGroup errgroup.Group
	for i := 0; i < 1; i++ {
		errGroup.Go(TestTCPConn(clientPort, 10240*1024, Timeout))
	}
	if err := errGroup.Wait(); err != nil {
		t.Error(err)
	}
}

func TestChainHanlerTransportTls(t *testing.T) {
	socksPort := net.PickTCPPort()
	ssPort := net.PickTCPPort()
	t.Log("socksPort", socksPort, "ssPort", ssPort)
	secret := uuid.New()
	serverConfig := &server.ServerConfig{
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Tag:     "socks",
				Address: net.LocalHostIP.String(),
				Port:    uint32(socksPort),
				Protocol: serial.ToTypedMessage(&proxy.SocksServerConfig{
					AuthType: proxy.AuthType_NO_AUTH,
				}),
			},
			{
				Tag:     "ss",
				Address: net.LocalHostIP.String(),
				Port:    uint32(ssPort),
				Transport: &configs.TransportConfig{
					Protocol: &configs.TransportConfig_Websocket{
						Websocket: &websocket.WebsocketConfig{
							Path: "/ws",
						},
					},
					Security: &configs.TransportConfig_Tls{
						Tls: &tls.TlsConfig{
							Certificates: []*tls.Certificate{
								tls.ParseCertificate(cert.MustGenerate(nil)),
							},
						},
					},
				},
				Protocol: serial.ToTypedMessage(
					&proxy.ShadowsocksServerConfig{
						User: &configs.UserConfig{
							Secret: secret.String(),
						},
						CipherType: proxy.ShadowsocksCipherType_CHACHA20_POLY1305,
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
						&proxy.DokodemoConfig{
							Address:  tcpDest.Address.String(),
							Port:     uint32(tcpDest.Port),
							Networks: []net.Network{net.Network_TCP},
						},
					),
				},
			},
		},
		Outbound: &configs.OutboundConfig{
			ChainHandlers: []*configs.ChainHandlerConfig{
				{
					Handlers: []*configs.OutboundHandlerConfig{

						{
							Address:  net.LocalHostIP.String(),
							Port:     uint32(socksPort),
							Protocol: serial.ToTypedMessage(&proxy.SocksClientConfig{}),
						},
						{
							Address: net.LocalHostIP.String(),
							Port:    uint32(ssPort),
							Protocol: serial.ToTypedMessage(&proxy.ShadowsocksClientConfig{
								Password:   secret.String(),
								CipherType: proxy.ShadowsocksCipherType_CHACHA20_POLY1305,
							}),
							Transport: &configs.TransportConfig{
								Protocol: &configs.TransportConfig_Websocket{
									Websocket: &websocket.WebsocketConfig{
										Path: "/ws",
									},
								},
								Security: &configs.TransportConfig_Tls{
									Tls: &tls.TlsConfig{
										AllowInsecure: true,
									},
								},
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

	var errGroup errgroup.Group
	for i := 0; i < 1; i++ {
		errGroup.Go(TestTCPConn(clientPort, 10240*1024, Timeout))
	}
	if err := errGroup.Wait(); err != nil {
		t.Error(err)
	}
}

func TestChainHanlerSSHys(t *testing.T) {
	hysPort := net.PickTCPPort()
	ssPort := net.PickTCPPort()
	t.Log("hysPort", hysPort, "ssPort", ssPort)
	secret := uuid.New()
	userID := uuid.New()

	serverConfig := &server.ServerConfig{
		Users: []*configs.UserConfig{
			{
				Id:     userID.String(),
				Secret: userID.String(),
			},
		},
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Tag:     "hysteria",
				Address: net.LocalHostIP.String(),
				Port:    uint32(hysPort),
				Protocol: serial.ToTypedMessage(&proxy.Hysteria2ServerConfig{
					IgnoreClientBandwidth: true,
					TlsConfig: &tls.TlsConfig{
						Certificates: []*tls.Certificate{
							tls.ParseCertificate(cert.MustGenerate(nil)),
						},
					},
				}),
			},
			{
				Tag:     "ss",
				Address: net.LocalHostIP.String(),
				Port:    uint32(ssPort),
				Protocol: serial.ToTypedMessage(
					&proxy.ShadowsocksServerConfig{
						User: &configs.UserConfig{
							Secret: secret.String(),
						},
						CipherType: proxy.ShadowsocksCipherType_CHACHA20_POLY1305,
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
						&proxy.DokodemoConfig{
							Address:  tcpDest.Address.String(),
							Port:     uint32(tcpDest.Port),
							Networks: []net.Network{net.Network_TCP},
						},
					),
				},
			},
		},
		Outbound: &configs.OutboundConfig{
			ChainHandlers: []*configs.ChainHandlerConfig{
				{
					Handlers: []*configs.OutboundHandlerConfig{

						{
							Address: net.LocalHostIP.String(),
							Port:    uint32(hysPort),
							Protocol: serial.ToTypedMessage(&proxy.Hysteria2ClientConfig{
								Auth: userID.String(),
								TlsConfig: &tls.TlsConfig{
									AllowInsecure: true,
								},
							}),
						},
						{
							Address: net.LocalHostIP.String(),
							Port:    uint32(ssPort),
							Protocol: serial.ToTypedMessage(&proxy.ShadowsocksClientConfig{
								Password:   secret.String(),
								CipherType: proxy.ShadowsocksCipherType_CHACHA20_POLY1305,
							}),
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

	var errGroup errgroup.Group
	for i := 0; i < 1; i++ {
		errGroup.Go(TestTCPConn(clientPort, 10240*1024, Timeout))
	}
	if err := errGroup.Wait(); err != nil {
		t.Error(err)
	}
}

func TestChainHanlerHysSS(t *testing.T) {
	hysPort := net.PickTCPPort()
	ssPort := net.PickTCPPort()
	trojanPort := net.PickTCPPort()
	t.Log("hysPort", hysPort, "ssPort", ssPort, "trojanPort", trojanPort)
	secret := uuid.New()
	userID := uuid.New()

	serverConfig := &server.ServerConfig{
		Users: []*configs.UserConfig{
			{
				Id:     userID.String(),
				Secret: userID.String(),
			},
		},
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Tag:     "hysteria",
				Address: net.LocalHostIP.String(),
				Port:    uint32(hysPort),
				Protocol: serial.ToTypedMessage(&proxy.Hysteria2ServerConfig{
					IgnoreClientBandwidth: true,
					TlsConfig: &tls.TlsConfig{
						Certificates: []*tls.Certificate{
							tls.ParseCertificate(cert.MustGenerate(nil)),
						},
					},
				}),
			},
			{
				Tag:     "ss",
				Address: net.LocalHostIP.String(),
				Port:    uint32(ssPort),
				Protocol: serial.ToTypedMessage(
					&proxy.ShadowsocksServerConfig{
						User: &configs.UserConfig{
							Secret: secret.String(),
						},
						CipherType: proxy.ShadowsocksCipherType_CHACHA20_POLY1305,
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
						&proxy.DokodemoConfig{
							Address:  tcpDest.Address.String(),
							Port:     uint32(tcpDest.Port),
							Networks: []net.Network{net.Network_TCP},
						},
					),
				},
			},
		},
		Outbound: &configs.OutboundConfig{
			ChainHandlers: []*configs.ChainHandlerConfig{
				{
					Handlers: []*configs.OutboundHandlerConfig{

						{
							Address: net.LocalHostIP.String(),
							Port:    uint32(ssPort),
							Protocol: serial.ToTypedMessage(&proxy.ShadowsocksClientConfig{
								Password:   secret.String(),
								CipherType: proxy.ShadowsocksCipherType_CHACHA20_POLY1305,
							}),
						},
						{
							Address: net.LocalHostIP.String(),
							Port:    uint32(hysPort),
							Protocol: serial.ToTypedMessage(&proxy.Hysteria2ClientConfig{
								Auth: userID.String(),
								TlsConfig: &tls.TlsConfig{
									AllowInsecure: true,
								},
							}),
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

	var errGroup errgroup.Group
	for i := 0; i < 1; i++ {
		errGroup.Go(TestTCPConn(clientPort, 10240*1024, Timeout))
	}
	if err := errGroup.Wait(); err != nil {
		t.Error(err)
	}
}

func TestChainHanlerTrojanHys(t *testing.T) {
	hysPort := net.PickTCPPort()
	trojanPort := net.PickTCPPort()
	t.Log("hysPort", hysPort, "trojanPort", trojanPort)
	secret := uuid.New()

	serverConfig := &server.ServerConfig{
		Users: []*configs.UserConfig{
			{
				Id:     secret.String(),
				Secret: secret.String(),
			},
		},
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Tag:     "hysteria",
				Address: net.LocalHostIP.String(),
				Port:    uint32(hysPort),
				Protocol: serial.ToTypedMessage(&proxy.Hysteria2ServerConfig{
					IgnoreClientBandwidth: true,
					TlsConfig: &tls.TlsConfig{
						Certificates: []*tls.Certificate{
							tls.ParseCertificate(cert.MustGenerate(nil)),
						},
					},
				}),
			},
			{
				Tag:     "trojan",
				Address: net.LocalHostIP.String(),
				Port:    uint32(trojanPort),
				Protocol: serial.ToTypedMessage(
					&proxy.TrojanServerConfig{
						Users: []*configs.UserConfig{
							{
								Id:     secret.String(),
								Secret: secret.String(),
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
						&proxy.DokodemoConfig{
							Address:  tcpDest.Address.String(),
							Port:     uint32(tcpDest.Port),
							Networks: []net.Network{net.Network_TCP},
						},
					),
				},
			},
		},
		Outbound: &configs.OutboundConfig{
			ChainHandlers: []*configs.ChainHandlerConfig{
				{
					Handlers: []*configs.OutboundHandlerConfig{

						{
							Address: net.LocalHostIP.String(),
							Port:    uint32(trojanPort),
							Protocol: serial.ToTypedMessage(&proxy.TrojanClientConfig{
								Password: secret.String(),
							}),
						},
						{
							Address: net.LocalHostIP.String(),
							Port:    uint32(hysPort),
							Protocol: serial.ToTypedMessage(&proxy.Hysteria2ClientConfig{
								Auth: secret.String(),
								TlsConfig: &tls.TlsConfig{
									AllowInsecure: true,
								},
							}),
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

	var errGroup errgroup.Group
	for i := 0; i < 1; i++ {
		errGroup.Go(TestTCPConn(clientPort, 1024000, Timeout))
	}
	if err := errGroup.Wait(); err != nil {
		t.Error(err)
	}
}

func TestChainHanlerTrojanTlsHys(t *testing.T) {
	hysPort := net.PickTCPPort()
	trojanPort := net.PickTCPPort()
	t.Log("hysPort", hysPort, "trojanPort", trojanPort)
	secret := uuid.New()

	serverConfig := &server.ServerConfig{
		Users: []*configs.UserConfig{
			{
				Id:     secret.String(),
				Secret: secret.String(),
			},
		},
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Tag:     "hysteria",
				Address: net.LocalHostIP.String(),
				Port:    uint32(hysPort),
				Protocol: serial.ToTypedMessage(&proxy.Hysteria2ServerConfig{
					IgnoreClientBandwidth: true,
					TlsConfig: &tls.TlsConfig{
						Certificates: []*tls.Certificate{
							tls.ParseCertificate(cert.MustGenerate(nil)),
						},
					},
				}),
			},
			{
				Tag:     "trojan",
				Address: net.LocalHostIP.String(),
				Port:    uint32(trojanPort),
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
					&proxy.TrojanServerConfig{
						Users: []*configs.UserConfig{
							{
								Id:     secret.String(),
								Secret: secret.String(),
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
						&proxy.DokodemoConfig{
							Address:  tcpDest.Address.String(),
							Port:     uint32(tcpDest.Port),
							Networks: []net.Network{net.Network_TCP},
						},
					),
				},
			},
		},
		Outbound: &configs.OutboundConfig{
			ChainHandlers: []*configs.ChainHandlerConfig{
				{
					Handlers: []*configs.OutboundHandlerConfig{

						{
							Address: net.LocalHostIP.String(),
							Port:    uint32(trojanPort),
							Protocol: serial.ToTypedMessage(&proxy.TrojanClientConfig{
								Password: secret.String(),
							}),
							Transport: &configs.TransportConfig{
								Security: &configs.TransportConfig_Tls{
									Tls: &tls.TlsConfig{
										AllowInsecure: true,
									},
								},
							},
						},
						{
							Address: net.LocalHostIP.String(),
							Port:    uint32(hysPort),
							Protocol: serial.ToTypedMessage(&proxy.Hysteria2ClientConfig{
								Auth: secret.String(),
								TlsConfig: &tls.TlsConfig{
									AllowInsecure: true,
								},
							}),
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

	var errGroup errgroup.Group
	for i := 0; i < 10; i++ {
		errGroup.Go(TestTCPConn(clientPort, 10240, Timeout))
	}
	if err := errGroup.Wait(); err != nil {
		t.Error(err)
	}
}
