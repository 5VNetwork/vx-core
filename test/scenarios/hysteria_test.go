//go:build test

package scenarios

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	stdhttp "net/http"
	"testing"
	"time"

	"github.com/5vnetwork/vx-core/app/buildclient"
	"github.com/5vnetwork/vx-core/app/buildserver"
	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/util"
	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/protocol/tls/cert"
	"github.com/5vnetwork/vx-core/common/serial"
	"github.com/5vnetwork/vx-core/common/uuid"
	"github.com/5vnetwork/vx-core/test"
	"github.com/5vnetwork/vx-core/test/realmtest"
	httptest "github.com/5vnetwork/vx-core/test/servers/http"
	"github.com/5vnetwork/vx-core/test/servers/tcp"
	"github.com/5vnetwork/vx-core/test/servers/udp"
	"github.com/5vnetwork/vx-core/transport/security/tls"

	"golang.org/x/sync/errgroup"
)

func TestHysteriaTCP(t *testing.T) {
	userID := uuid.New()
	serverPort := net.PickUDPPort()
	t.Logf("server port: %d", serverPort)

	serverConfig := &configs.ServerConfig{
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Port: uint32(serverPort),
				Protocol: serial.ToTypedMessage(&configs.Hysteria2ServerConfig{
					IgnoreClientBandwidth: true,
					TlsConfig: &tls.TlsConfig{
						Certificates: []*tls.Certificate{
							tls.ParseCertificate(cert.MustGenerate(nil)),
						},
					},
				}),
			},
		},
		// Router: &configs.RouterConfig{
		// 	Rules: []*configs.RuleConfig{
		// 		{
		// 			OutboundTag: "direct",
		// 		},
		// 	},
		// },
		Users: []*configs.UserConfig{
			{
				Id:     userID.String(),
				Secret: userID.String(),
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
					Address: "127.0.0.1",
					Port:    uint32(serverPort),
					Protocol: serial.ToTypedMessage(&configs.Hysteria2ClientConfig{
						Auth: userID.String(),
						TlsConfig: &tls.TlsConfig{
							AllowInsecure: true,
							ServerName:    "example.com",
						},
					}),
				},
			},
		},
	}

	server, err := buildserver.NewX(serverConfig)
	common.Must(err)
	common.Must(server.Start(context.Background()))
	defer server.Stop(context.Background())

	client, err := buildclient.NewX(clientConfig)
	common.Must(err)
	common.Must(client.Start())
	defer client.Close()

	var errg errgroup.Group
	for i := 0; i < 1; i++ {
		errg.Go(TestTCPConn(clientPort, 10*1024, time.Second*400))
	}

	if err := errg.Wait(); err != nil {
		t.Error(err)
	}
}

const hysteriaLargeUploadTimeout = 60 * time.Second

func TestHysteriaTCPLarge(t *testing.T) {
	clientPort, cleanup := startHysteriaDokodemo(t, tcpDest)
	defer cleanup()

	var errg errgroup.Group
	errg.Go(TestTCPConn(clientPort, test.TenMB, hysteriaLargeUploadTimeout))
	if err := errg.Wait(); err != nil {
		t.Error(err)
	}
}

func TestHysteriaTCPLargeConcurrent(t *testing.T) {
	clientPort, cleanup := startHysteriaDokodemo(t, tcpDest)
	defer cleanup()

	var errg errgroup.Group
	for i := 0; i < 2; i++ {
		errg.Go(TestTCPConn(clientPort, test.OneMB, hysteriaLargeUploadTimeout))
	}
	if err := errg.Wait(); err != nil {
		t.Error(err)
	}
}

func TestHysteriaHTTPUpload(t *testing.T) {
	httpPort, httpServer := startHysteriaUploadHTTP(t)
	defer httpServer.Close()

	clientPort, cleanup := startHysteriaDokodemo(t, net.TCPDestination(net.LocalHostIP, httpPort))
	defer cleanup()

	if err := hysteriaHTTPPut(fmt.Sprintf("http://127.0.0.1:%d/upload", clientPort), test.OneMB, hysteriaLargeUploadTimeout); err != nil {
		t.Fatal(err)
	}
}

func TestHysteriaHTTPUploadConcurrent(t *testing.T) {
	httpPort, httpServer := startHysteriaUploadHTTP(t)
	defer httpServer.Close()

	clientPort, cleanup := startHysteriaDokodemo(t, net.TCPDestination(net.LocalHostIP, httpPort))
	defer cleanup()

	url := fmt.Sprintf("http://127.0.0.1:%d/upload", clientPort)
	var errg errgroup.Group
	for i := 0; i < 2; i++ {
		errg.Go(func() error {
			return hysteriaHTTPPut(url, test.OneMB, hysteriaLargeUploadTimeout)
		})
	}
	if err := errg.Wait(); err != nil {
		t.Fatal(err)
	}
}

func startHysteriaDokodemo(t *testing.T, dest net.Destination) (net.Port, func()) {
	t.Helper()
	userID := uuid.New()
	serverPort := net.PickUDPPort()
	serverConfig := &configs.ServerConfig{
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Port: uint32(serverPort),
				Protocol: serial.ToTypedMessage(&configs.Hysteria2ServerConfig{
					IgnoreClientBandwidth: true,
					TlsConfig: &tls.TlsConfig{
						Certificates: []*tls.Certificate{
							tls.ParseCertificate(cert.MustGenerate(nil)),
						},
					},
				}),
			},
		},
		Users: []*configs.UserConfig{
			{
				Id:     userID.String(),
				Secret: userID.String(),
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
					Address: "127.0.0.1",
					Port:    uint32(serverPort),
					Protocol: serial.ToTypedMessage(&configs.Hysteria2ClientConfig{
						Auth: userID.String(),
						TlsConfig: &tls.TlsConfig{
							AllowInsecure: true,
							ServerName:    "example.com",
						},
					}),
				},
			},
		},
	}

	server, err := buildserver.NewX(serverConfig)
	common.Must(err)
	common.Must(server.Start(context.Background()))

	client, err := buildclient.NewX(clientConfig)
	common.Must(err)
	common.Must(client.Start())

	return clientPort, func() {
		client.Close()
		server.Stop(context.Background())
	}
}

func startHysteriaUploadHTTP(t *testing.T) (net.Port, *httptest.Server) {
	t.Helper()
	httpPort := tcp.PickPort()
	httpServer := &httptest.Server{
		Port: httpPort,
		PathHandler: map[string]stdhttp.HandlerFunc{
			"/upload": func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
				n, err := io.Copy(io.Discard, r.Body)
				r.Body.Close()
				if err != nil {
					w.WriteHeader(stdhttp.StatusInternalServerError)
					fmt.Fprintf(w, "read body: %v", err)
					return
				}
				w.Header().Set("Content-Type", "text/plain")
				fmt.Fprintf(w, "ok %d", n)
			},
		},
	}
	_, err := httpServer.Start()
	common.Must(err)
	return httpPort, httpServer
}

func hysteriaHTTPPut(url string, payloadSize int, timeout time.Duration) error {
	payload := make([]byte, payloadSize)
	common.Must2(rand.Read(payload))

	req, err := stdhttp.NewRequest(stdhttp.MethodPut, url, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.ContentLength = int64(len(payload))
	req.Header.Set("Expect", "100-continue")
	req.Header.Set("Content-Type", "application/octet-stream")

	client := &stdhttp.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != stdhttp.StatusOK {
		return fmt.Errorf("status %d: %s", resp.StatusCode, body)
	}
	want := fmt.Sprintf("ok %d", payloadSize)
	if string(body) != want {
		return fmt.Errorf("unexpected body %q, want %q", body, want)
	}
	return nil
}

func TestHysteriaUDP(t *testing.T) {
	userID := uuid.New()
	serverPort := net.PickUDPPort()
	serverConfig := &configs.ServerConfig{
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Port: uint32(serverPort),
				Protocol: serial.ToTypedMessage(&configs.Hysteria2ServerConfig{
					IgnoreClientBandwidth: true,
					TlsConfig: &tls.TlsConfig{
						Certificates: []*tls.Certificate{
							tls.ParseCertificate(cert.MustGenerate(nil)),
						},
					},
				}),
			},
		},

		Users: []*configs.UserConfig{
			{
				Id:     userID.String(),
				Secret: userID.String(),
			},
		},
	}

	clientPort := udp.PickPort()
	t.Log("client port", clientPort)
	clientConfig := &configs.TmConfig{
		InboundManager: &configs.InboundManagerConfig{
			Handlers: []*configs.ProxyInboundConfig{
				{
					Address: net.LocalHostIP.String(),
					Port:    uint32(clientPort),
					Protocol: serial.ToTypedMessage(
						&configs.DokodemoConfig{
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
					Address: "127.0.0.1",
					Port:    uint32(serverPort),
					Protocol: serial.ToTypedMessage(&configs.Hysteria2ClientConfig{
						Auth: userID.String(),
						TlsConfig: &tls.TlsConfig{
							AllowInsecure: true,
						},
					}),
				},
			},
		},
	}

	server, err := buildserver.NewX(serverConfig)
	common.Must(err)
	common.Must(server.Start(context.Background()))
	defer server.Stop(context.Background())

	client, err := buildclient.NewX(clientConfig)
	common.Must(err)
	common.Must(client.Start())
	defer client.Close()

	var errg errgroup.Group
	for i := 0; i < 10; i++ {
		errg.Go(TestUDPConnN(clientPort, 1024, time.Second*4, 10))
	}

	if err := errg.Wait(); err != nil {
		t.Error(err)
	}
}

func TestHysteriaTCPSalamander(t *testing.T) {
	userID := uuid.New()
	serverPort := net.PickUDPPort()
	secret := "1234567890"
	serverConfig := &configs.ServerConfig{
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Port: uint32(serverPort),
				Protocol: serial.ToTypedMessage(&configs.Hysteria2ServerConfig{
					IgnoreClientBandwidth: true,
					TlsConfig: &tls.TlsConfig{
						Certificates: []*tls.Certificate{
							tls.ParseCertificate(cert.MustGenerate(nil)),
						},
					},
					Obfs: &configs.ObfsConfig{
						Obfs: &configs.ObfsConfig_Salamander{
							Salamander: &configs.SalamanderConfig{
								Password: secret,
							},
						},
					},
				}),
			},
		},
		// Router: &configs.RouterConfig{
		// 	Rules: []*configs.RuleConfig{
		// 		{
		// 			OutboundTag: "direct",
		// 		},
		// 	},
		// },
		Users: []*configs.UserConfig{
			{
				Id:     userID.String(),
				Secret: userID.String(),
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
					Address: "127.0.0.1",
					Port:    uint32(serverPort),
					Protocol: serial.ToTypedMessage(&configs.Hysteria2ClientConfig{
						Auth: userID.String(),
						TlsConfig: &tls.TlsConfig{
							AllowInsecure: true,
						},
						Obfs: &configs.ObfsConfig{
							Obfs: &configs.ObfsConfig_Salamander{
								Salamander: &configs.SalamanderConfig{
									Password: secret,
								},
							},
						},
					}),
				},
			},
		},
	}

	server, err := buildserver.NewX(serverConfig)
	common.Must(err)
	common.Must(server.Start(context.Background()))
	defer server.Stop(context.Background())

	client, err := buildclient.NewX(clientConfig)
	common.Must(err)
	common.Must(client.Start())
	defer client.Close()

	var errg errgroup.Group
	for i := 0; i < 1; i++ {
		errg.Go(TestTCPConn(clientPort, 10*1024, time.Second*4))
	}

	if err := errg.Wait(); err != nil {
		t.Error(err)
	}
}

func TestHysteriaECH(t *testing.T) {
	userID := uuid.New()
	serverPort := net.PickUDPPort()

	t.Logf("server  port: %d", serverPort)

	echConfig, echKey, err := util.ExecuteECH("asdf.a")
	common.Must(err)

	serverConfig := &configs.ServerConfig{
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Port: uint32(serverPort),
				Protocol: serial.ToTypedMessage(&configs.Hysteria2ServerConfig{
					IgnoreClientBandwidth: true,
					TlsConfig: &tls.TlsConfig{
						Certificates: []*tls.Certificate{
							tls.ParseCertificate(cert.MustGenerate(nil)),
						},
						EchKey: echKey,
					},
				}),
			},
		},
		Users: []*configs.UserConfig{
			{
				Id:     userID.String(),
				Secret: userID.String(),
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
					Address: "127.0.0.1",
					Port:    uint32(serverPort),
					Protocol: serial.ToTypedMessage(&configs.Hysteria2ClientConfig{
						Auth: userID.String(),
						TlsConfig: &tls.TlsConfig{
							AllowInsecure: true,
							EchConfig:     echConfig,
							ServerName:    "example.com",
						},
					}),
				},
			},
		},
	}

	server, err := buildserver.NewX(serverConfig)
	common.Must(err)
	common.Must(server.Start(context.Background()))
	defer server.Stop(context.Background())

	client, err := buildclient.NewX(clientConfig)
	common.Must(err)
	common.Must(client.Start())
	defer client.Close()

	// test.InitZeroLog()

	var errg errgroup.Group
	for i := 0; i < 1; i++ {
		errg.Go(TestTCPConn(clientPort, 10*1024, time.Second*400))
	}

	if err := errg.Wait(); err != nil {
		t.Error(err)
	}
}

func TestHysteriaRealm(t *testing.T) {
	const realmToken = "test-token"
	const realmID = "test-realm"

	stunAddr, stopSTUN := realmtest.StartReflectiveSTUN(t)
	defer stopSTUN()
	t.Logf("stun server: %s", stunAddr)

	rdvAddr, stopRendezvous := realmtest.StartRendezvous(t, realmToken)
	defer stopRendezvous()
	t.Logf("rendezvous server: %s", rdvAddr)

	realmURL := fmt.Sprintf("realm+http://%s@%s/%s?stun=%s", realmToken, rdvAddr, realmID, stunAddr)
	runHysteriaRealmTCPTest(t, realmURL, 30*time.Second)
}

const random = ""

func TestHysteriaRealmPublic(t *testing.T) {
	t.Skip()
	if testing.Short() {
		t.Skip("skipping public realm integration test in short mode")
	}

	realmURL := fmt.Sprintf("realm://public@realm.hy2.io/%s", random)
	t.Logf("public realm URL: %s", realmURL)
	runHysteriaRealmTCPTest(t, realmURL, 2*time.Minute)
}

func runHysteriaRealmTCPTest(t *testing.T, realmURL string, connTimeout time.Duration) {
	t.Helper()
	t.Logf("realm URL: %s", realmURL)

	userID := uuid.New()
	serverPort := net.PickUDPPort()
	t.Logf("server port: %d", serverPort)

	serverConfig := &configs.ServerConfig{
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Port: uint32(serverPort),
				Protocol: serial.ToTypedMessage(&configs.Hysteria2ServerConfig{
					IgnoreClientBandwidth: true,
					Realm: &configs.RealmConfig{
						RealmAddr: realmURL,
					},
					TlsConfig: &tls.TlsConfig{
						Certificates: []*tls.Certificate{
							tls.ParseCertificate(cert.MustGenerate(nil)),
						},
					},
				}),
			},
		},
		Users: []*configs.UserConfig{
			{
				Id:     userID.String(),
				Secret: userID.String(),
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
					Protocol: serial.ToTypedMessage(&configs.Hysteria2ClientConfig{
						Auth: userID.String(),
						TlsConfig: &tls.TlsConfig{
							AllowInsecure: true,
							ServerName:    "example.com",
						},
						Realm: &configs.RealmConfig{
							RealmAddr: realmURL,
						},
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
	errg.Go(TestTCPConn(clientPort, 10*1024, connTimeout))

	if err := errg.Wait(); err != nil {
		t.Error(err)
	}
}

func TestHysteriaRealmPortMap(t *testing.T) {
	t.Skip()
	if testing.Short() {
		t.Skip("skipping public realm integration test in short mode")
	}

	realmURL := fmt.Sprintf("realm://public@realm.hy2.io/%s", random)
	t.Logf("public realm URL: %s", realmURL)
	runHysteriaRealmPortMapTest(t, realmURL, 2*time.Minute)
}

func runHysteriaRealmPortMapTest(t *testing.T, realmURL string, connTimeout time.Duration) {
	t.Helper()
	t.Logf("realm URL: %s", realmURL)

	userID := uuid.New()
	serverPort := net.PickUDPPort()
	t.Logf("server port: %d", serverPort)

	serverConfig := &configs.ServerConfig{
		Inbounds: []*configs.ProxyInboundConfig{
			{
				Port: uint32(serverPort),
				Protocol: serial.ToTypedMessage(&configs.Hysteria2ServerConfig{
					IgnoreClientBandwidth: true,
					Realm: &configs.RealmConfig{
						RealmAddr: realmURL,
						PortMapping: &configs.RealmPortMappingConfig{
							Enabled: true,
						},
					},
					TlsConfig: &tls.TlsConfig{
						Certificates: []*tls.Certificate{
							tls.ParseCertificate(cert.MustGenerate(nil)),
						},
					},
				}),
			},
		},
		Users: []*configs.UserConfig{
			{
				Id:     userID.String(),
				Secret: userID.String(),
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
					Protocol: serial.ToTypedMessage(&configs.Hysteria2ClientConfig{
						Auth: userID.String(),
						TlsConfig: &tls.TlsConfig{
							AllowInsecure: true,
							ServerName:    "example.com",
						},
						Realm: &configs.RealmConfig{
							RealmAddr: realmURL,
							PortMapping: &configs.RealmPortMappingConfig{
								Enabled: true,
							},
						},
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
	errg.Go(TestTCPConn(clientPort, 10*1024, connTimeout))

	if err := errg.Wait(); err != nil {
		t.Error(err)
	}
}
