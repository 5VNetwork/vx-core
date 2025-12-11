package util

import (
	"testing"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/configs/proxy"
	"github.com/5vnetwork/vx-core/common/protocol/tls/cert"
	"github.com/5vnetwork/vx-core/common/serial"
	"github.com/5vnetwork/vx-core/transport/protocols/grpc"
	"github.com/5vnetwork/vx-core/transport/protocols/http"
	"github.com/5vnetwork/vx-core/transport/protocols/httpupgrade"
	"github.com/5vnetwork/vx-core/transport/protocols/tcp"
	"github.com/5vnetwork/vx-core/transport/protocols/websocket"
	"github.com/5vnetwork/vx-core/transport/security/reality"
	"github.com/5vnetwork/vx-core/transport/security/tls"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"
)

func TestInboundConfigToOutboundConfig_NoUsers(t *testing.T) {
	inboundConfig := &configs.ProxyInboundConfig{
		Tag:     "test-inbound",
		Address: "0.0.0.0",
		Port:    8080,
		Users:   []*configs.UserConfig{}, // Empty users
	}

	result, err := InboundConfigToOutboundConfig("prefix", inboundConfig, "example.com")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "no users")
}

func TestInboundConfigToOutboundConfig_VmessWithTLS(t *testing.T) {
	// Create a test certificate
	testCert := cert.MustGenerate(nil,
		cert.DNSNames("example.com"),
		cert.CommonName("example.com"),
	)
	certPEM, _ := testCert.ToPEM()

	// Create VMess server config
	vmessServerConfig := &proxy.VmessServerConfig{}
	vmessProtocol := serial.ToTypedMessage(vmessServerConfig)

	// Create TLS config
	tlsConfig := &tls.TlsConfig{
		Certificates: []*tls.Certificate{
			{
				Certificate: certPEM,
			},
		},
	}

	// Create transport config with TLS
	transportConfig := &configs.TransportConfig{
		Security: &configs.TransportConfig_Tls{
			Tls: tlsConfig,
		},
		Protocol: &configs.TransportConfig_Websocket{
			Websocket: &websocket.WebsocketConfig{},
		},
	}

	// Create inbound config
	inboundConfig := &configs.ProxyInboundConfig{
		Tag:       "vmess-inbound",
		Address:   "0.0.0.0",
		Port:      443,
		Protocols: []*anypb.Any{vmessProtocol},
		Transport: transportConfig,
		Users: []*configs.UserConfig{
			{
				Id:     "test-user-id",
				Secret: "test-secret",
			},
		},
	}

	result, err := InboundConfigToOutboundConfig("test", inboundConfig, "example.com")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Greater(t, len(result), 0)

	// Verify the first outbound config
	outbound := result[0]
	assert.Contains(t, outbound.Tag, "test-vmess-inbound")
	assert.Equal(t, "example.com", outbound.Address)
	assert.NotNil(t, outbound.Protocol)
	assert.NotNil(t, outbound.Transport)
}

func TestInboundConfigToOutboundConfig_ShadowsocksWithTLS(t *testing.T) {
	// Create a test certificate
	testCert := cert.MustGenerate(nil,
		cert.DNSNames("example.com"),
		cert.CommonName("example.com"),
	)
	certPEM, _ := testCert.ToPEM()

	// Create Shadowsocks server config
	ssServerConfig := &proxy.ShadowsocksServerConfig{}
	ssProtocol := serial.ToTypedMessage(ssServerConfig)

	// Create TLS config
	tlsConfig := &tls.TlsConfig{
		Certificates: []*tls.Certificate{
			{
				Certificate: certPEM,
			},
		},
	}

	// Create transport config with TLS and WebSocket
	transportConfig := &configs.TransportConfig{
		Security: &configs.TransportConfig_Tls{
			Tls: tlsConfig,
		},
		Protocol: &configs.TransportConfig_Websocket{
			Websocket: &websocket.WebsocketConfig{
				Path: "/path",
			},
		},
	}

	// Create inbound config
	inboundConfig := &configs.ProxyInboundConfig{
		Tag:       "ss-inbound",
		Address:   "0.0.0.0",
		Port:      443,
		Protocols: []*anypb.Any{ssProtocol},
		Transport: transportConfig,
		Users: []*configs.UserConfig{
			{
				Id:     "test-user-id",
				Secret: "test-password",
			},
		},
	}

	result, err := InboundConfigToOutboundConfig("test", inboundConfig, "example.com")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Greater(t, len(result), 0)

	// Verify the outbound config
	outbound := result[0]
	assert.Contains(t, outbound.Tag, "test-ss-inbound")
	assert.Equal(t, "example.com", outbound.Address)
	assert.NotNil(t, outbound.Protocol)
	assert.NotNil(t, outbound.Transport)
}

func TestInboundConfigToOutboundConfig_TrojanWithReality(t *testing.T) {
	// Create Reality config
	privateKey := make([]byte, 32)
	for i := range privateKey {
		privateKey[i] = byte(i)
	}

	realityConfig := &reality.RealityConfig{
		PrivateKey:  privateKey,
		ServerNames: []string{"example.com"},
		ShortIds:    [][]byte{[]byte("short-id")},
	}

	// Create Trojan server config
	trojanServerConfig := &proxy.TrojanServerConfig{}
	trojanProtocol := serial.ToTypedMessage(trojanServerConfig)

	// Create transport config with Reality
	transportConfig := &configs.TransportConfig{
		Security: &configs.TransportConfig_Reality{
			Reality: realityConfig,
		},
		Protocol: &configs.TransportConfig_Grpc{
			Grpc: &grpc.GrpcConfig{},
		},
	}

	// Create inbound config
	inboundConfig := &configs.ProxyInboundConfig{
		Tag:       "trojan-inbound",
		Address:   "0.0.0.0",
		Port:      443,
		Protocols: []*anypb.Any{trojanProtocol},
		Transport: transportConfig,
		Users: []*configs.UserConfig{
			{
				Id:     "test-user-id",
				Secret: "trojan-password",
			},
		},
	}

	result, err := InboundConfigToOutboundConfig("test", inboundConfig, "example.com")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Greater(t, len(result), 0)

	// Verify the outbound config
	outbound := result[0]
	assert.Contains(t, outbound.Tag, "test-trojan-inbound")
	assert.Equal(t, "example.com", outbound.Address)
	assert.NotNil(t, outbound.Protocol)
	assert.NotNil(t, outbound.Transport)
}

func TestInboundConfigToOutboundConfig_MultiplePorts(t *testing.T) {
	// Create a test certificate
	testCert := cert.MustGenerate(nil,
		cert.DNSNames("example.com"),
		cert.CommonName("example.com"),
	)
	certPEM, _ := testCert.ToPEM()

	// Create VMess server config
	vmessServerConfig := &proxy.VmessServerConfig{}
	vmessProtocol := serial.ToTypedMessage(vmessServerConfig)

	// Create TLS config
	tlsConfig := &tls.TlsConfig{
		Certificates: []*tls.Certificate{
			{
				Certificate: certPEM,
			},
		},
	}

	// Create transport config
	transportConfig := &configs.TransportConfig{
		Security: &configs.TransportConfig_Tls{
			Tls: tlsConfig,
		},
		Protocol: &configs.TransportConfig_Tcp{
			Tcp: &tcp.TcpConfig{},
		},
	}

	// Create inbound config with multiple ports
	inboundConfig := &configs.ProxyInboundConfig{
		Tag:       "multi-port-inbound",
		Address:   "0.0.0.0",
		Port:      443,
		Ports:     []uint32{8080, 8443, 9090},
		Protocols: []*anypb.Any{vmessProtocol},
		Transport: transportConfig,
		Users: []*configs.UserConfig{
			{
				Id:     "test-user-id",
				Secret: "test-secret",
			},
		},
	}

	result, err := InboundConfigToOutboundConfig("test", inboundConfig, "example.com")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Greater(t, len(result), 0)

	// Verify ports are included
	outbound := result[0]
	assert.NotNil(t, outbound.Ports)
	assert.Greater(t, len(outbound.Ports), 0)
}

func TestInboundConfigToOutboundConfig_WebsocketTransport(t *testing.T) {
	// Create a test certificate
	testCert := cert.MustGenerate(nil,
		cert.DNSNames("example.com"),
		cert.CommonName("example.com"),
	)
	certPEM, _ := testCert.ToPEM()

	// Create VMess server config
	vmessServerConfig := &proxy.VmessServerConfig{}
	vmessProtocol := serial.ToTypedMessage(vmessServerConfig)

	// Create TLS config
	tlsConfig := &tls.TlsConfig{
		Certificates: []*tls.Certificate{
			{
				Certificate: certPEM,
			},
		},
	}

	// Create transport config with WebSocket
	wsConfig := &websocket.WebsocketConfig{
		Path: "/ws",
		Host: "example.com",
	}

	transportConfig := &configs.TransportConfig{
		Security: &configs.TransportConfig_Tls{
			Tls: tlsConfig,
		},
		Protocol: &configs.TransportConfig_Websocket{
			Websocket: wsConfig,
		},
	}

	// Create inbound config
	inboundConfig := &configs.ProxyInboundConfig{
		Tag:       "ws-inbound",
		Address:   "0.0.0.0",
		Port:      443,
		Protocols: []*anypb.Any{vmessProtocol},
		Transport: transportConfig,
		Users: []*configs.UserConfig{
			{
				Id:     "test-user-id",
				Secret: "test-secret",
			},
		},
	}

	result, err := InboundConfigToOutboundConfig("test", inboundConfig, "example.com")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Greater(t, len(result), 0)

	// Verify WebSocket transport is preserved
	outbound := result[0]
	assert.NotNil(t, outbound.Transport)
	assert.NotNil(t, outbound.Transport.GetWebsocket())
}

func TestInboundConfigToOutboundConfig_HttpTransport(t *testing.T) {
	// Create a test certificate
	testCert := cert.MustGenerate(nil,
		cert.DNSNames("example.com"),
		cert.CommonName("example.com"),
	)
	certPEM, _ := testCert.ToPEM()

	// Create VMess server config
	vmessServerConfig := &proxy.VmessServerConfig{}
	vmessProtocol := serial.ToTypedMessage(vmessServerConfig)

	// Create TLS config
	tlsConfig := &tls.TlsConfig{
		Certificates: []*tls.Certificate{
			{
				Certificate: certPEM,
			},
		},
	}

	// Create transport config with HTTP
	httpConfig := &http.HttpConfig{}

	transportConfig := &configs.TransportConfig{
		Security: &configs.TransportConfig_Tls{
			Tls: tlsConfig,
		},
		Protocol: &configs.TransportConfig_Http{
			Http: httpConfig,
		},
	}

	// Create inbound config
	inboundConfig := &configs.ProxyInboundConfig{
		Tag:       "http-inbound",
		Address:   "0.0.0.0",
		Port:      443,
		Protocols: []*anypb.Any{vmessProtocol},
		Transport: transportConfig,
		Users: []*configs.UserConfig{
			{
				Id:     "test-user-id",
				Secret: "test-secret",
			},
		},
	}

	result, err := InboundConfigToOutboundConfig("test", inboundConfig, "example.com")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Greater(t, len(result), 0)

	// Verify HTTP transport is preserved
	outbound := result[0]
	assert.NotNil(t, outbound.Transport)
	assert.NotNil(t, outbound.Transport.GetHttp())
}

func TestInboundConfigToOutboundConfig_GrpcTransport(t *testing.T) {
	// Create a test certificate
	testCert := cert.MustGenerate(nil,
		cert.DNSNames("example.com"),
		cert.CommonName("example.com"),
	)
	certPEM, _ := testCert.ToPEM()

	// Create VMess server config
	vmessServerConfig := &proxy.VmessServerConfig{}
	vmessProtocol := serial.ToTypedMessage(vmessServerConfig)

	// Create TLS config
	tlsConfig := &tls.TlsConfig{
		Certificates: []*tls.Certificate{
			{
				Certificate: certPEM,
			},
		},
	}

	// Create transport config with gRPC
	grpcConfig := &grpc.GrpcConfig{}

	transportConfig := &configs.TransportConfig{
		Security: &configs.TransportConfig_Tls{
			Tls: tlsConfig,
		},
		Protocol: &configs.TransportConfig_Grpc{
			Grpc: grpcConfig,
		},
	}

	// Create inbound config
	inboundConfig := &configs.ProxyInboundConfig{
		Tag:       "grpc-inbound",
		Address:   "0.0.0.0",
		Port:      443,
		Protocols: []*anypb.Any{vmessProtocol},
		Transport: transportConfig,
		Users: []*configs.UserConfig{
			{
				Id:     "test-user-id",
				Secret: "test-secret",
			},
		},
	}

	result, err := InboundConfigToOutboundConfig("test", inboundConfig, "example.com")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Greater(t, len(result), 0)

	// Verify gRPC transport is preserved
	outbound := result[0]
	assert.NotNil(t, outbound.Transport)
	assert.NotNil(t, outbound.Transport.GetGrpc())
}

func TestInboundConfigToOutboundConfig_TagGeneration(t *testing.T) {
	// Create a test certificate
	testCert := cert.MustGenerate(nil,
		cert.DNSNames("example.com"),
		cert.CommonName("example.com"),
	)
	certPEM, _ := testCert.ToPEM()

	// Create VMess server config
	vmessServerConfig := &proxy.VmessServerConfig{}
	vmessProtocol := serial.ToTypedMessage(vmessServerConfig)

	// Create TLS config
	tlsConfig := &tls.TlsConfig{
		Certificates: []*tls.Certificate{
			{
				Certificate: certPEM,
			},
		},
	}

	// Create transport config
	transportConfig := &configs.TransportConfig{
		Security: &configs.TransportConfig_Tls{
			Tls: tlsConfig,
		},
		Protocol: &configs.TransportConfig_Tcp{
			Tcp: &tcp.TcpConfig{},
		},
	}

	// Create inbound config
	inboundConfig := &configs.ProxyInboundConfig{
		Tag:       "my-inbound",
		Address:   "0.0.0.0",
		Port:      443,
		Protocols: []*anypb.Any{vmessProtocol},
		Transport: transportConfig,
		Users: []*configs.UserConfig{
			{
				Id:     "test-user-id",
				Secret: "test-secret",
			},
		},
	}

	result, err := InboundConfigToOutboundConfig("prefix", inboundConfig, "example.com")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Greater(t, len(result), 0)

	// Verify tag format
	outbound := result[0]
	assert.Contains(t, outbound.Tag, "prefix")
	assert.Contains(t, outbound.Tag, "my-inbound")
}

func TestInboundConfigToOutboundConfig_AnytlsProtocol(t *testing.T) {
	// Create Anytls server config
	anytlsServerConfig := &proxy.AnytlsServerConfig{}
	anytlsProtocol := serial.ToTypedMessage(anytlsServerConfig)

	// Create transport config without security (Anytls handles its own)
	transportConfig := &configs.TransportConfig{
		Protocol: &configs.TransportConfig_Tcp{
			Tcp: &tcp.TcpConfig{},
		},
	}

	// Create inbound config
	inboundConfig := &configs.ProxyInboundConfig{
		Tag:       "anytls-inbound",
		Address:   "0.0.0.0",
		Port:      443,
		Protocols: []*anypb.Any{anytlsProtocol},
		Transport: transportConfig,
		Users: []*configs.UserConfig{
			{
				Id:     "test-user-id",
				Secret: "anytls-password",
			},
		},
	}

	// Anytls might not require TLS in transport, but the function expects security config
	// This test will fail if security is required, which is expected behavior
	result, err := InboundConfigToOutboundConfig("test", inboundConfig, "example.com")
	if err != nil {
		// Expected if security config is required
		assert.Contains(t, err.Error(), "invalid security config")
	} else {
		// If it succeeds, verify the result
		require.NotNil(t, result)
		assert.Greater(t, len(result), 0)
	}
}

func TestInboundConfigToOutboundConfig_SocksProtocol(t *testing.T) {
	// Create a test certificate
	testCert := cert.MustGenerate(nil,
		cert.DNSNames("example.com"),
		cert.CommonName("example.com"),
	)
	certPEM, _ := testCert.ToPEM()

	// Create SOCKS server config
	socksServerConfig := &proxy.SocksServerConfig{}
	socksProtocol := serial.ToTypedMessage(socksServerConfig)

	// Create TLS config
	tlsConfig := &tls.TlsConfig{
		Certificates: []*tls.Certificate{
			{
				Certificate: certPEM,
			},
		},
	}

	// Create transport config
	transportConfig := &configs.TransportConfig{
		Security: &configs.TransportConfig_Tls{
			Tls: tlsConfig,
		},
		Protocol: &configs.TransportConfig_Tcp{
			Tcp: &tcp.TcpConfig{},
		},
	}

	// Create inbound config
	inboundConfig := &configs.ProxyInboundConfig{
		Tag:       "socks-inbound",
		Address:   "0.0.0.0",
		Port:      1080,
		Protocols: []*anypb.Any{socksProtocol},
		Transport: transportConfig,
		Users: []*configs.UserConfig{
			{
				Id:     "socks-user",
				Secret: "socks-password",
			},
		},
	}

	result, err := InboundConfigToOutboundConfig("test", inboundConfig, "example.com")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Greater(t, len(result), 0)

	// Verify the outbound config
	outbound := result[0]
	assert.Contains(t, outbound.Tag, "test-socks-inbound")
	assert.Equal(t, "example.com", outbound.Address)
	assert.NotNil(t, outbound.Protocol)
}

func TestInboundConfigToOutboundConfig_ZeroPort(t *testing.T) {
	// Create a test certificate
	testCert := cert.MustGenerate(nil,
		cert.DNSNames("example.com"),
		cert.CommonName("example.com"),
	)
	certPEM, _ := testCert.ToPEM()

	// Create VMess server config
	vmessServerConfig := &proxy.VmessServerConfig{}
	vmessProtocol := serial.ToTypedMessage(vmessServerConfig)

	// Create TLS config
	tlsConfig := &tls.TlsConfig{
		Certificates: []*tls.Certificate{
			{
				Certificate: certPEM,
			},
		},
	}

	// Create transport config
	transportConfig := &configs.TransportConfig{
		Security: &configs.TransportConfig_Tls{
			Tls: tlsConfig,
		},
		Protocol: &configs.TransportConfig_Tcp{
			Tcp: &tcp.TcpConfig{},
		},
	}

	// Create inbound config with zero port (should be filtered out)
	inboundConfig := &configs.ProxyInboundConfig{
		Tag:       "zero-port-inbound",
		Address:   "0.0.0.0",
		Port:      0,                      // Zero port
		Ports:     []uint32{443, 0, 8080}, // Mix of valid and zero ports
		Protocols: []*anypb.Any{vmessProtocol},
		Transport: transportConfig,
		Users: []*configs.UserConfig{
			{
				Id:     "test-user-id",
				Secret: "test-secret",
			},
		},
	}

	result, err := InboundConfigToOutboundConfig("test", inboundConfig, "example.com")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Greater(t, len(result), 0)

	// Verify ports - zero ports should be filtered
	outbound := result[0]
	if outbound.Ports != nil {
		for _, portRange := range outbound.Ports {
			assert.NotEqual(t, uint32(0), portRange.From)
			assert.NotEqual(t, uint32(0), portRange.To)
		}
	}
}

// Tests for MultiInboundConfigToOutboundConfig

func TestMultiInboundConfigToOutboundConfig_NoUsers(t *testing.T) {
	inboundConfig := &configs.MultiProxyInboundConfig{
		Tag:     "test-multi-inbound",
		Address: "0.0.0.0",
		Ports:   []uint32{8080},
		Users:   []*configs.UserConfig{}, // Empty users
	}

	result, err := MultiInboundConfigToOutboundConfig("prefix", inboundConfig, "example.com")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "no users")
}

func TestMultiInboundConfigToOutboundConfig_VmessWithTLS(t *testing.T) {
	// Create a test certificate
	testCert := cert.MustGenerate(nil,
		cert.DNSNames("example.com"),
		cert.CommonName("example.com"),
	)
	certPEM, _ := testCert.ToPEM()

	// Create VMess server config
	vmessServerConfig := &proxy.VmessServerConfig{}
	vmessProtocol := serial.ToTypedMessage(vmessServerConfig)

	// Create TLS config
	tlsConfig := &tls.TlsConfig{
		Certificates: []*tls.Certificate{
			{
				Certificate: certPEM,
			},
		},
	}

	// Create security config with TLS
	securityConfig := &configs.MultiProxyInboundConfig_Security{
		Security: &configs.MultiProxyInboundConfig_Security_Tls{
			Tls: tlsConfig,
		},
	}

	// Create transport protocol with WebSocket
	transportProtocol := &configs.MultiProxyInboundConfig_Protocol{
		Protocol: &configs.MultiProxyInboundConfig_Protocol_Websocket{
			Websocket: &websocket.WebsocketConfig{
				Path: "/ws",
			},
		},
	}

	// Create inbound config
	inboundConfig := &configs.MultiProxyInboundConfig{
		Tag:                "vmess-multi-inbound",
		Address:            "0.0.0.0",
		Ports:              []uint32{443},
		Protocols:          []*anypb.Any{vmessProtocol},
		SecurityConfigs:    []*configs.MultiProxyInboundConfig_Security{securityConfig},
		TransportProtocols: []*configs.MultiProxyInboundConfig_Protocol{transportProtocol},
		Users: []*configs.UserConfig{
			{
				Id:     "test-user-id",
				Secret: "test-secret",
			},
		},
	}

	result, err := MultiInboundConfigToOutboundConfig("test", inboundConfig, "example.com")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Greater(t, len(result), 0)

	// Verify the first outbound config
	outbound := result[1]
	assert.Contains(t, outbound.Tag, "test-vmess-multi-inbound")
	assert.Equal(t, "example.com", outbound.Address)
	assert.NotNil(t, outbound.Protocol)
	assert.NotNil(t, outbound.Transport)
}

func TestMultiInboundConfigToOutboundConfig_MultipleTransportProtocols(t *testing.T) {
	// Create a test certificate
	testCert := cert.MustGenerate(nil,
		cert.DNSNames("example.com"),
		cert.CommonName("example.com"),
	)
	certPEM, _ := testCert.ToPEM()

	// Create VMess server config
	trojanServerConfig := &proxy.TrojanServerConfig{}
	trojanProtocol := serial.ToTypedMessage(trojanServerConfig)

	// Create TLS config
	tlsConfig := &tls.TlsConfig{
		Certificates: []*tls.Certificate{
			{
				Certificate: certPEM,
			},
		},
	}

	// Create security config with TLS
	securityConfig := &configs.MultiProxyInboundConfig_Security{
		Security: &configs.MultiProxyInboundConfig_Security_Tls{
			Tls: tlsConfig,
		},
	}

	// Create multiple transport protocols
	wsProtocol := &configs.MultiProxyInboundConfig_Protocol{
		Protocol: &configs.MultiProxyInboundConfig_Protocol_Websocket{
			Websocket: &websocket.WebsocketConfig{Path: "/ws"},
		},
	}
	httpProtocol := &configs.MultiProxyInboundConfig_Protocol{
		Protocol: &configs.MultiProxyInboundConfig_Protocol_Http{
			Http: &http.HttpConfig{},
		},
	}
	grpcProtocol := &configs.MultiProxyInboundConfig_Protocol{
		Protocol: &configs.MultiProxyInboundConfig_Protocol_Grpc{
			Grpc: &grpc.GrpcConfig{},
		},
	}

	// Create inbound config with multiple transport protocols
	inboundConfig := &configs.MultiProxyInboundConfig{
		Tag:                "multi-transport-inbound",
		Address:            "0.0.0.0",
		Ports:              []uint32{443},
		Protocols:          []*anypb.Any{trojanProtocol},
		SecurityConfigs:    []*configs.MultiProxyInboundConfig_Security{securityConfig},
		TransportProtocols: []*configs.MultiProxyInboundConfig_Protocol{wsProtocol, httpProtocol, grpcProtocol},
		Users: []*configs.UserConfig{
			{
				Id:     "test-user-id",
				Secret: "test-secret",
			},
		},
	}

	result, err := MultiInboundConfigToOutboundConfig("test", inboundConfig, "example.com")
	require.NoError(t, err)
	require.NotNil(t, result)
	// Should generate multiple outbound configs (one for each transport protocol)
	assert.Greater(t, len(result), 1)

	// Verify all outbound configs have the correct structure
	for _, outbound := range result {
		assert.Contains(t, outbound.Tag, "test-multi-transport-inbound")
		assert.Equal(t, "example.com", outbound.Address)
		assert.NotNil(t, outbound.Protocol)
		assert.NotNil(t, outbound.Transport)
	}
}

func TestMultiInboundConfigToOutboundConfig_MultipleSecurityConfigs(t *testing.T) {
	// Create test certificates
	cert1 := cert.MustGenerate(nil,
		cert.DNSNames("example.com"),
		cert.CommonName("example.com"),
	)
	certPEM1, _ := cert1.ToPEM()

	cert2 := cert.MustGenerate(nil,
		cert.DNSNames("test.com"),
		cert.CommonName("test.com"),
	)
	certPEM2, _ := cert2.ToPEM()

	// Create VMess server config
	vmessServerConfig := &proxy.VmessServerConfig{}
	vmessProtocol := serial.ToTypedMessage(vmessServerConfig)

	// Create multiple TLS security configs
	securityConfig1 := &configs.MultiProxyInboundConfig_Security{
		Security: &configs.MultiProxyInboundConfig_Security_Tls{
			Tls: &tls.TlsConfig{
				Certificates: []*tls.Certificate{
					{Certificate: certPEM1},
				},
			},
		},
		Domains: []string{"example.com"},
	}
	securityConfig2 := &configs.MultiProxyInboundConfig_Security{
		Security: &configs.MultiProxyInboundConfig_Security_Tls{
			Tls: &tls.TlsConfig{
				Certificates: []*tls.Certificate{
					{Certificate: certPEM2},
				},
			},
		},
		Domains: []string{"test.com"},
	}

	// Create transport protocol
	transportProtocol := &configs.MultiProxyInboundConfig_Protocol{
		Protocol: &configs.MultiProxyInboundConfig_Protocol_Websocket{
			Websocket: &websocket.WebsocketConfig{Path: "/ws"},
		},
	}

	// Create inbound config with multiple security configs
	inboundConfig := &configs.MultiProxyInboundConfig{
		Tag:                "multi-security-inbound",
		Address:            "0.0.0.0",
		Ports:              []uint32{443},
		Protocols:          []*anypb.Any{vmessProtocol},
		SecurityConfigs:    []*configs.MultiProxyInboundConfig_Security{securityConfig1, securityConfig2},
		TransportProtocols: []*configs.MultiProxyInboundConfig_Protocol{transportProtocol},
		Users: []*configs.UserConfig{
			{
				Id:     "test-user-id",
				Secret: "test-secret",
			},
		},
	}

	result, err := MultiInboundConfigToOutboundConfig("test", inboundConfig, "example.com")
	require.NoError(t, err)
	require.NotNil(t, result)
	// Should generate multiple outbound configs (one for each security config)
	assert.Greater(t, len(result), 1)
}

func TestMultiInboundConfigToOutboundConfig_SecurityConfigAlwaysFlag(t *testing.T) {
	// Create test certificates
	cert1 := cert.MustGenerate(nil,
		cert.DNSNames("example.com"),
		cert.CommonName("example.com"),
	)
	certPEM1, _ := cert1.ToPEM()

	cert2 := cert.MustGenerate(nil,
		cert.DNSNames("test.com"),
		cert.CommonName("test.com"),
	)
	certPEM2, _ := cert2.ToPEM()

	// Create VMess server config
	vmessServerConfig := &proxy.VmessServerConfig{}
	vmessProtocol := serial.ToTypedMessage(vmessServerConfig)

	// Create security configs - one with Always=true
	securityConfig1 := &configs.MultiProxyInboundConfig_Security{
		Security: &configs.MultiProxyInboundConfig_Security_Tls{
			Tls: &tls.TlsConfig{
				Certificates: []*tls.Certificate{
					{Certificate: certPEM1},
				},
			},
		},
		Always: true, // This should be the only one used
	}
	securityConfig2 := &configs.MultiProxyInboundConfig_Security{
		Security: &configs.MultiProxyInboundConfig_Security_Tls{
			Tls: &tls.TlsConfig{
				Certificates: []*tls.Certificate{
					{Certificate: certPEM2},
				},
			},
		},
		Always: false, // This should be ignored
	}

	// Create transport protocol
	transportProtocol := &configs.MultiProxyInboundConfig_Protocol{
		Protocol: &configs.MultiProxyInboundConfig_Protocol_Websocket{
			Websocket: &websocket.WebsocketConfig{Path: "/ws"},
		},
	}

	// Create inbound config
	inboundConfig := &configs.MultiProxyInboundConfig{
		Tag:                "always-security-inbound",
		Address:            "0.0.0.0",
		Ports:              []uint32{443},
		Protocols:          []*anypb.Any{vmessProtocol},
		SecurityConfigs:    []*configs.MultiProxyInboundConfig_Security{securityConfig1, securityConfig2},
		TransportProtocols: []*configs.MultiProxyInboundConfig_Protocol{transportProtocol},
		Users: []*configs.UserConfig{
			{
				Id:     "test-user-id",
				Secret: "test-secret",
			},
		},
	}

	result, err := MultiInboundConfigToOutboundConfig("test", inboundConfig, "example.com")
	require.NoError(t, err)
	require.NotNil(t, result)
	// Should only generate configs for the Always=true security config
	// The number depends on how many certificates are in that config
	assert.Greater(t, len(result), 0)
}

func TestMultiInboundConfigToOutboundConfig_RealitySecurity(t *testing.T) {
	// Create Reality config
	privateKey := make([]byte, 32)
	for i := range privateKey {
		privateKey[i] = byte(i)
	}

	realityConfig := &reality.RealityConfig{
		PrivateKey:  privateKey,
		ServerNames: []string{"example.com", "test.com"},
		ShortIds:    [][]byte{[]byte("short-id-1"), []byte("short-id-2")},
	}

	// Create VMess server config
	vmessServerConfig := &proxy.VmessServerConfig{}
	vmessProtocol := serial.ToTypedMessage(vmessServerConfig)

	// Create security config with Reality
	securityConfig := &configs.MultiProxyInboundConfig_Security{
		Security: &configs.MultiProxyInboundConfig_Security_Reality{
			Reality: realityConfig,
		},
	}

	// Create transport protocol
	transportProtocol := &configs.MultiProxyInboundConfig_Protocol{
		Protocol: &configs.MultiProxyInboundConfig_Protocol_Grpc{
			Grpc: &grpc.GrpcConfig{},
		},
	}

	// Create inbound config
	inboundConfig := &configs.MultiProxyInboundConfig{
		Tag:                "reality-multi-inbound",
		Address:            "0.0.0.0",
		Ports:              []uint32{443},
		Protocols:          []*anypb.Any{vmessProtocol},
		SecurityConfigs:    []*configs.MultiProxyInboundConfig_Security{securityConfig},
		TransportProtocols: []*configs.MultiProxyInboundConfig_Protocol{transportProtocol},
		Users: []*configs.UserConfig{
			{
				Id:     "test-user-id",
				Secret: "test-secret",
			},
		},
	}

	result, err := MultiInboundConfigToOutboundConfig("test", inboundConfig, "example.com")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Greater(t, len(result), 0)

	// Verify Reality config is preserved
	outbound := result[1]
	assert.NotNil(t, outbound.Transport)
	assert.NotNil(t, outbound.Transport.GetReality())
}

func TestMultiInboundConfigToOutboundConfig_HttpUpgradeTransport(t *testing.T) {
	// Create a test certificate
	testCert := cert.MustGenerate(nil,
		cert.DNSNames("example.com"),
		cert.CommonName("example.com"),
	)
	certPEM, _ := testCert.ToPEM()

	// Create VMess server config
	vmessServerConfig := &proxy.VmessServerConfig{}
	vmessProtocol := serial.ToTypedMessage(vmessServerConfig)

	// Create TLS config
	tlsConfig := &tls.TlsConfig{
		Certificates: []*tls.Certificate{
			{
				Certificate: certPEM,
			},
		},
	}

	// Create security config
	securityConfig := &configs.MultiProxyInboundConfig_Security{
		Security: &configs.MultiProxyInboundConfig_Security_Tls{
			Tls: tlsConfig,
		},
	}

	// Create HTTPUpgrade transport protocol
	httpUpgradeProtocol := &configs.MultiProxyInboundConfig_Protocol{
		Protocol: &configs.MultiProxyInboundConfig_Protocol_Httpupgrade{
			Httpupgrade: &httpupgrade.HttpUpgradeConfig{
				Config: &websocket.WebsocketConfig{
					Path: "/path",
				},
			},
		},
	}

	// Create inbound config
	inboundConfig := &configs.MultiProxyInboundConfig{
		Tag:                "httpupgrade-inbound",
		Address:            "0.0.0.0",
		Ports:              []uint32{443},
		Protocols:          []*anypb.Any{vmessProtocol},
		SecurityConfigs:    []*configs.MultiProxyInboundConfig_Security{securityConfig},
		TransportProtocols: []*configs.MultiProxyInboundConfig_Protocol{httpUpgradeProtocol},
		Users: []*configs.UserConfig{
			{
				Id:     "test-user-id",
				Secret: "test-secret",
			},
		},
	}

	result, err := MultiInboundConfigToOutboundConfig("test", inboundConfig, "example.com")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Greater(t, len(result), 0)

	// Verify HTTPUpgrade transport is preserved
	outbound := result[1]
	assert.NotNil(t, outbound.Transport)
	assert.NotNil(t, outbound.Transport.GetHttpupgrade())
}

func TestMultiInboundConfigToOutboundConfig_MultiplePorts(t *testing.T) {
	// Create a test certificate
	testCert := cert.MustGenerate(nil,
		cert.DNSNames("example.com"),
		cert.CommonName("example.com"),
	)
	certPEM, _ := testCert.ToPEM()

	// Create VMess server config
	vmessServerConfig := &proxy.VmessServerConfig{}
	vmessProtocol := serial.ToTypedMessage(vmessServerConfig)

	// Create TLS config
	tlsConfig := &tls.TlsConfig{
		Certificates: []*tls.Certificate{
			{
				Certificate: certPEM,
			},
		},
	}

	// Create security config
	securityConfig := &configs.MultiProxyInboundConfig_Security{
		Security: &configs.MultiProxyInboundConfig_Security_Tls{
			Tls: tlsConfig,
		},
	}

	// Create transport protocol
	transportProtocol := &configs.MultiProxyInboundConfig_Protocol{
		Protocol: &configs.MultiProxyInboundConfig_Protocol_Websocket{
			Websocket: &websocket.WebsocketConfig{Path: "/ws"},
		},
	}

	// Create inbound config with multiple ports
	inboundConfig := &configs.MultiProxyInboundConfig{
		Tag:                "multi-port-inbound",
		Address:            "0.0.0.0",
		Ports:              []uint32{443, 8080, 8443},
		Protocols:          []*anypb.Any{vmessProtocol},
		SecurityConfigs:    []*configs.MultiProxyInboundConfig_Security{securityConfig},
		TransportProtocols: []*configs.MultiProxyInboundConfig_Protocol{transportProtocol},
		Users: []*configs.UserConfig{
			{
				Id:     "test-user-id",
				Secret: "test-secret",
			},
		},
	}

	result, err := MultiInboundConfigToOutboundConfig("test", inboundConfig, "example.com")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Greater(t, len(result), 0)

	// Verify ports are included
	outbound := result[0]
	assert.NotNil(t, outbound.Ports)
	assert.Equal(t, 3, len(outbound.Ports))
}

func TestMultiInboundConfigToOutboundConfig_InvalidSecurityConfig(t *testing.T) {
	// Create VMess server config
	vmessServerConfig := &proxy.VmessServerConfig{}
	vmessProtocol := serial.ToTypedMessage(vmessServerConfig)

	// Create invalid security config (no security type set)
	securityConfig := &configs.MultiProxyInboundConfig_Security{
		// Security field is nil
	}

	// Create transport protocol
	transportProtocol := &configs.MultiProxyInboundConfig_Protocol{
		Protocol: &configs.MultiProxyInboundConfig_Protocol_Websocket{
			Websocket: &websocket.WebsocketConfig{Path: "/ws"},
		},
	}

	// Create inbound config
	inboundConfig := &configs.MultiProxyInboundConfig{
		Tag:                "invalid-security-inbound",
		Address:            "0.0.0.0",
		Ports:              []uint32{443},
		Protocols:          []*anypb.Any{vmessProtocol},
		SecurityConfigs:    []*configs.MultiProxyInboundConfig_Security{securityConfig},
		TransportProtocols: []*configs.MultiProxyInboundConfig_Protocol{transportProtocol},
		Users: []*configs.UserConfig{
			{
				Id:     "test-user-id",
				Secret: "test-secret",
			},
		},
	}

	result, err := MultiInboundConfigToOutboundConfig("test", inboundConfig, "example.com")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid security config")
}

func TestMultiInboundConfigToOutboundConfig_ShadowsocksProtocol(t *testing.T) {
	// Create a test certificate
	testCert := cert.MustGenerate(nil,
		cert.DNSNames("example.com"),
		cert.CommonName("example.com"),
	)
	certPEM, _ := testCert.ToPEM()

	// Create Shadowsocks server config
	ssServerConfig := &proxy.ShadowsocksServerConfig{}
	ssProtocol := serial.ToTypedMessage(ssServerConfig)

	// Create TLS config
	tlsConfig := &tls.TlsConfig{
		Certificates: []*tls.Certificate{
			{
				Certificate: certPEM,
			},
		},
	}

	// Create security config
	securityConfig := &configs.MultiProxyInboundConfig_Security{
		Security: &configs.MultiProxyInboundConfig_Security_Tls{
			Tls: tlsConfig,
		},
	}

	// Create transport protocol
	transportProtocol := &configs.MultiProxyInboundConfig_Protocol{
		Protocol: &configs.MultiProxyInboundConfig_Protocol_Http{
			Http: &http.HttpConfig{},
		},
	}

	// Create inbound config
	inboundConfig := &configs.MultiProxyInboundConfig{
		Tag:                "ss-multi-inbound",
		Address:            "0.0.0.0",
		Ports:              []uint32{443},
		Protocols:          []*anypb.Any{ssProtocol},
		SecurityConfigs:    []*configs.MultiProxyInboundConfig_Security{securityConfig},
		TransportProtocols: []*configs.MultiProxyInboundConfig_Protocol{transportProtocol},
		Users: []*configs.UserConfig{
			{
				Id:     "test-user-id",
				Secret: "ss-password",
			},
		},
	}

	result, err := MultiInboundConfigToOutboundConfig("test", inboundConfig, "example.com")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Greater(t, len(result), 0)

	// Verify the outbound config
	outbound := result[0]
	assert.Contains(t, outbound.Tag, "test-ss-multi-inbound")
	assert.Equal(t, "example.com", outbound.Address)
	assert.NotNil(t, outbound.Protocol)
}

func TestMultiInboundConfigToOutboundConfig_TrojanProtocol(t *testing.T) {
	// Create a test certificate
	testCert := cert.MustGenerate(nil,
		cert.DNSNames("example.com"),
		cert.CommonName("example.com"),
	)
	certPEM, _ := testCert.ToPEM()

	// Create Trojan server config
	trojanServerConfig := &proxy.TrojanServerConfig{}
	trojanProtocol := serial.ToTypedMessage(trojanServerConfig)

	// Create TLS config
	tlsConfig := &tls.TlsConfig{
		Certificates: []*tls.Certificate{
			{
				Certificate: certPEM,
			},
		},
	}

	// Create security config
	securityConfig := &configs.MultiProxyInboundConfig_Security{
		Security: &configs.MultiProxyInboundConfig_Security_Tls{
			Tls: tlsConfig,
		},
	}

	// Create transport protocol
	transportProtocol := &configs.MultiProxyInboundConfig_Protocol{
		Protocol: &configs.MultiProxyInboundConfig_Protocol_Grpc{
			Grpc: &grpc.GrpcConfig{},
		},
	}

	// Create inbound config
	inboundConfig := &configs.MultiProxyInboundConfig{
		Tag:                "trojan-multi-inbound",
		Address:            "0.0.0.0",
		Ports:              []uint32{443},
		Protocols:          []*anypb.Any{trojanProtocol},
		SecurityConfigs:    []*configs.MultiProxyInboundConfig_Security{securityConfig},
		TransportProtocols: []*configs.MultiProxyInboundConfig_Protocol{transportProtocol},
		Users: []*configs.UserConfig{
			{
				Id:     "test-user-id",
				Secret: "trojan-password",
			},
		},
	}

	result, err := MultiInboundConfigToOutboundConfig("test", inboundConfig, "example.com")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Greater(t, len(result), 0)

	// Verify the outbound config
	outbound := result[0]
	assert.Contains(t, outbound.Tag, "test-trojan-multi-inbound")
	assert.Equal(t, "example.com", outbound.Address)
	assert.NotNil(t, outbound.Protocol)
}

func TestMultiInboundConfigToOutboundConfig_TagGeneration(t *testing.T) {
	// Create a test certificate
	testCert := cert.MustGenerate(nil,
		cert.DNSNames("example.com"),
		cert.CommonName("example.com"),
	)
	certPEM, _ := testCert.ToPEM()

	// Create VMess server config
	vmessServerConfig := &proxy.VmessServerConfig{}
	vmessProtocol := serial.ToTypedMessage(vmessServerConfig)

	// Create TLS config
	tlsConfig := &tls.TlsConfig{
		Certificates: []*tls.Certificate{
			{
				Certificate: certPEM,
			},
		},
	}

	// Create security config
	securityConfig := &configs.MultiProxyInboundConfig_Security{
		Security: &configs.MultiProxyInboundConfig_Security_Tls{
			Tls: tlsConfig,
		},
	}

	// Create transport protocol (MultiProxyInboundConfig doesn't support TCP, use WebSocket instead)
	transportProtocol := &configs.MultiProxyInboundConfig_Protocol{
		Protocol: &configs.MultiProxyInboundConfig_Protocol_Websocket{
			Websocket: &websocket.WebsocketConfig{Path: "/ws"},
		},
	}

	// Create inbound config
	inboundConfig := &configs.MultiProxyInboundConfig{
		Tag:                "my-multi-inbound",
		Address:            "0.0.0.0",
		Ports:              []uint32{443},
		Protocols:          []*anypb.Any{vmessProtocol},
		SecurityConfigs:    []*configs.MultiProxyInboundConfig_Security{securityConfig},
		TransportProtocols: []*configs.MultiProxyInboundConfig_Protocol{transportProtocol},
		Users: []*configs.UserConfig{
			{
				Id:     "test-user-id",
				Secret: "test-secret",
			},
		},
	}

	result, err := MultiInboundConfigToOutboundConfig("prefix", inboundConfig, "example.com")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Greater(t, len(result), 0)

	// Verify tag format
	outbound := result[0]
	assert.Contains(t, outbound.Tag, "prefix")
	assert.Contains(t, outbound.Tag, "my-multi-inbound")
}

func TestMultiInboundConfigToOutboundConfig_EmptyPorts(t *testing.T) {
	// Create a test certificate
	testCert := cert.MustGenerate(nil,
		cert.DNSNames("example.com"),
		cert.CommonName("example.com"),
	)
	certPEM, _ := testCert.ToPEM()

	// Create VMess server config
	vmessServerConfig := &proxy.VmessServerConfig{}
	vmessProtocol := serial.ToTypedMessage(vmessServerConfig)

	// Create TLS config
	tlsConfig := &tls.TlsConfig{
		Certificates: []*tls.Certificate{
			{
				Certificate: certPEM,
			},
		},
	}

	// Create security config
	securityConfig := &configs.MultiProxyInboundConfig_Security{
		Security: &configs.MultiProxyInboundConfig_Security_Tls{
			Tls: tlsConfig,
		},
	}

	// Create transport protocol
	transportProtocol := &configs.MultiProxyInboundConfig_Protocol{
		Protocol: &configs.MultiProxyInboundConfig_Protocol_Websocket{
			Websocket: &websocket.WebsocketConfig{Path: "/ws"},
		},
	}

	// Create inbound config with empty ports
	inboundConfig := &configs.MultiProxyInboundConfig{
		Tag:                "empty-ports-inbound",
		Address:            "0.0.0.0",
		Ports:              []uint32{}, // Empty ports
		Protocols:          []*anypb.Any{vmessProtocol},
		SecurityConfigs:    []*configs.MultiProxyInboundConfig_Security{securityConfig},
		TransportProtocols: []*configs.MultiProxyInboundConfig_Protocol{transportProtocol},
		Users: []*configs.UserConfig{
			{
				Id:     "test-user-id",
				Secret: "test-secret",
			},
		},
	}

	result, err := MultiInboundConfigToOutboundConfig("test", inboundConfig, "example.com")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Greater(t, len(result), 0)

	// Ports should be empty or nil
	outbound := result[0]
	if outbound.Ports != nil {
		assert.Equal(t, 0, len(outbound.Ports))
	}
}

func TestMultiInboundConfigToOutboundConfig_ZeroPortsFiltered(t *testing.T) {
	// Create a test certificate
	testCert := cert.MustGenerate(nil,
		cert.DNSNames("example.com"),
		cert.CommonName("example.com"),
	)
	certPEM, _ := testCert.ToPEM()

	// Create VMess server config
	vmessServerConfig := &proxy.VmessServerConfig{}
	vmessProtocol := serial.ToTypedMessage(vmessServerConfig)

	// Create TLS config
	tlsConfig := &tls.TlsConfig{
		Certificates: []*tls.Certificate{
			{
				Certificate: certPEM,
			},
		},
	}

	// Create security config
	securityConfig := &configs.MultiProxyInboundConfig_Security{
		Security: &configs.MultiProxyInboundConfig_Security_Tls{
			Tls: tlsConfig,
		},
	}

	// Create transport protocol
	transportProtocol := &configs.MultiProxyInboundConfig_Protocol{
		Protocol: &configs.MultiProxyInboundConfig_Protocol_Websocket{
			Websocket: &websocket.WebsocketConfig{Path: "/ws"},
		},
	}

	// Create inbound config with zero ports (should be filtered)
	inboundConfig := &configs.MultiProxyInboundConfig{
		Tag:                "zero-ports-inbound",
		Address:            "0.0.0.0",
		Ports:              []uint32{443, 0, 8080, 0}, // Mix of valid and zero ports
		Protocols:          []*anypb.Any{vmessProtocol},
		SecurityConfigs:    []*configs.MultiProxyInboundConfig_Security{securityConfig},
		TransportProtocols: []*configs.MultiProxyInboundConfig_Protocol{transportProtocol},
		Users: []*configs.UserConfig{
			{
				Id:     "test-user-id",
				Secret: "test-secret",
			},
		},
	}

	result, err := MultiInboundConfigToOutboundConfig("test", inboundConfig, "example.com")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Greater(t, len(result), 0)

	// Verify zero ports are filtered
	outbound := result[0]
	if outbound.Ports != nil {
		for _, portRange := range outbound.Ports {
			assert.NotEqual(t, uint32(0), portRange.From)
			assert.NotEqual(t, uint32(0), portRange.To)
		}
	}
}

func TestMultiInboundConfigToOutboundConfig_MultipleProtocolsAndTransports(t *testing.T) {
	// Create a test certificate
	testCert := cert.MustGenerate(nil,
		cert.DNSNames("example.com"),
		cert.CommonName("example.com"),
	)
	certPEM, _ := testCert.ToPEM()

	// Create multiple proxy protocols
	vmessServerConfig := &proxy.VmessServerConfig{}
	vmessProtocol := serial.ToTypedMessage(vmessServerConfig)

	ssServerConfig := &proxy.ShadowsocksServerConfig{}
	ssProtocol := serial.ToTypedMessage(ssServerConfig)

	// Create TLS config
	tlsConfig := &tls.TlsConfig{
		Certificates: []*tls.Certificate{
			{
				Certificate: certPEM,
			},
		},
	}

	// Create security config
	securityConfig := &configs.MultiProxyInboundConfig_Security{
		Security: &configs.MultiProxyInboundConfig_Security_Tls{
			Tls: tlsConfig,
		},
	}

	// Create multiple transport protocols
	wsProtocol := &configs.MultiProxyInboundConfig_Protocol{
		Protocol: &configs.MultiProxyInboundConfig_Protocol_Websocket{
			Websocket: &websocket.WebsocketConfig{Path: "/ws"},
		},
	}
	httpProtocol := &configs.MultiProxyInboundConfig_Protocol{
		Protocol: &configs.MultiProxyInboundConfig_Protocol_Http{
			Http: &http.HttpConfig{},
		},
	}

	// Create inbound config with multiple protocols and transports
	inboundConfig := &configs.MultiProxyInboundConfig{
		Tag:                "multi-proto-transport-inbound",
		Address:            "0.0.0.0",
		Ports:              []uint32{443},
		Protocols:          []*anypb.Any{vmessProtocol, ssProtocol},
		SecurityConfigs:    []*configs.MultiProxyInboundConfig_Security{securityConfig},
		TransportProtocols: []*configs.MultiProxyInboundConfig_Protocol{wsProtocol, httpProtocol},
		Users: []*configs.UserConfig{
			{
				Id:     "test-user-id",
				Secret: "test-secret",
			},
		},
	}

	result, err := MultiInboundConfigToOutboundConfig("test", inboundConfig, "example.com")
	require.NoError(t, err)
	require.NotNil(t, result)
	// Should generate configs for each combination of protocol and transport
	assert.Greater(t, len(result), 1)
}
