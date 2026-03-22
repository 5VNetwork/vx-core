package clash_test

import (
	"io/ioutil"
	"testing"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/util/sub/clash"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert/yaml"
)

func TestParseVmessProxy(t *testing.T) {
	// Example Mihomo/Clash vmess proxy configuration
	proxyMapping := map[string]any{
		"name":    "vmess-proxy",
		"type":    "vmess",
		"server":  "example.com",
		"port":    443,
		"uuid":    "b831381d-6324-4d53-ad4f-8cda48b30811",
		"alterId": 0,
		"cipher":  "auto",
		"tls":     true,
		"network": "ws",
		"ws-opts": map[string]any{
			"path": "/path",
			"headers": map[string]any{
				"Host": "example.com",
			},
		},
		"servername":  "example.com",
		"fingerprint": "chrome",
	}

	config, err := clash.ParseProxy(proxyMapping)
	if err != nil {
		t.Fatalf("Failed to parse vmess proxy: %v", err)
	}

	if config.Tag != "vmess-proxy" {
		t.Errorf("Expected tag 'vmess-proxy', got %s", config.Tag)
	}

	if config.Address != "example.com" {
		t.Errorf("Expected address 'example.com', got %s", config.Address)
	}

	if config.Port != 443 {
		t.Errorf("Expected port 443, got %d", config.Port)
	}

	// Check protocol is vmess
	vmessConfig := &configs.VmessClientConfig{}
	if err := config.Protocol.UnmarshalTo(vmessConfig); err != nil {
		t.Fatalf("Failed to unmarshal vmess config: %v", err)
	}

	if vmessConfig.Id != "b831381d-6324-4d53-ad4f-8cda48b30811" {
		t.Errorf("Expected UUID, got %s", vmessConfig.Id)
	}

	// Check transport config
	if config.Transport == nil {
		t.Fatal("Transport config is nil")
	}

	if config.Transport.GetWebsocket() == nil {
		t.Fatal("Expected websocket transport")
	}

	if config.Transport.GetWebsocket().Path != "/path" {
		t.Errorf("Expected path '/path', got %s", config.Transport.GetWebsocket().Path)
	}

	if config.Transport.GetTls() == nil {
		t.Fatal("Expected TLS config")
	}
}

func TestParseVlessProxy(t *testing.T) {
	proxyMapping := map[string]any{
		"name":    "vless-proxy",
		"type":    "vless",
		"server":  "example.com",
		"port":    443,
		"uuid":    "b831381d-6324-4d53-ad4f-8cda48b30811",
		"flow":    "xtls-rprx-vision",
		"tls":     true,
		"network": "tcp",
	}

	config, err := clash.ParseProxy(proxyMapping)
	if err != nil {
		t.Fatalf("Failed to parse vless proxy: %v", err)
	}

	if config.Tag != "vless-proxy" {
		t.Errorf("Expected tag 'vless-proxy', got %s", config.Tag)
	}

	vlessConfig := &configs.VlessClientConfig{}
	if err := config.Protocol.UnmarshalTo(vlessConfig); err != nil {
		t.Fatalf("Failed to unmarshal vless config: %v", err)
	}

	if vlessConfig.Id != "b831381d-6324-4d53-ad4f-8cda48b30811" {
		t.Errorf("Expected UUID, got %s", vlessConfig.Id)
	}

	if vlessConfig.Flow != "xtls-rprx-vision" {
		t.Errorf("Expected flow 'xtls-rprx-vision', got %s", vlessConfig.Flow)
	}
}

func TestParseTrojanProxy(t *testing.T) {
	proxyMapping := map[string]any{
		"name":     "trojan-proxy",
		"type":     "trojan",
		"server":   "example.com",
		"port":     443,
		"password": "password123",
		"tls":      true,
		"network":  "grpc",
		"grpc-opts": map[string]any{
			"grpc-service-name": "TrojanService",
		},
	}

	config, err := clash.ParseProxy(proxyMapping)
	if err != nil {
		t.Fatalf("Failed to parse trojan proxy: %v", err)
	}

	if config.Tag != "trojan-proxy" {
		t.Errorf("Expected tag 'trojan-proxy', got %s", config.Tag)
	}

	trojanConfig := &configs.TrojanClientConfig{}
	if err := config.Protocol.UnmarshalTo(trojanConfig); err != nil {
		t.Fatalf("Failed to unmarshal trojan config: %v", err)
	}

	if trojanConfig.Password != "password123" {
		t.Errorf("Expected password 'password123', got %s", trojanConfig.Password)
	}

	if config.Transport.GetGrpc() == nil {
		t.Fatal("Expected grpc transport")
	}

	if config.Transport.GetGrpc().ServiceName != "TrojanService" {
		t.Errorf("Expected service name 'TrojanService', got %s", config.Transport.GetGrpc().ServiceName)
	}
}

func TestParseShadowsocksProxy(t *testing.T) {
	proxyMapping := map[string]any{
		"name":     "ss-proxy",
		"type":     "ss",
		"server":   "example.com",
		"port":     8388,
		"password": "password123",
		"cipher":   "aes-256-gcm",
	}

	config, err := clash.ParseProxy(proxyMapping)
	if err != nil {
		t.Fatalf("Failed to parse shadowsocks proxy: %v", err)
	}

	if config.Tag != "ss-proxy" {
		t.Errorf("Expected tag 'ss-proxy', got %s", config.Tag)
	}

	ssConfig := &configs.ShadowsocksClientConfig{}
	if err := config.Protocol.UnmarshalTo(ssConfig); err != nil {
		t.Fatalf("Failed to unmarshal shadowsocks config: %v", err)
	}

	if ssConfig.Password != "password123" {
		t.Errorf("Expected password 'password123', got %s", ssConfig.Password)
	}

	if ssConfig.CipherType != configs.ShadowsocksCipherType_AES_256_GCM {
		t.Errorf("Expected AES_256_GCM cipher, got %v", ssConfig.CipherType)
	}
}

func TestParseAnytlsProxy(t *testing.T) {
	proxyMapping := map[string]any{
		"name":        "anytls-proxy",
		"type":        "anytls",
		"server":      "example.com",
		"port":        443,
		"password":    "password123",
		"sni":         "example.com",
		"fingerprint": "chrome",
		"alpn":        []string{"h2", "http/1.1"},
	}

	config, err := clash.ParseProxy(proxyMapping)
	if err != nil {
		t.Fatalf("Failed to parse anytls proxy: %v", err)
	}

	if config.Tag != "anytls-proxy" {
		t.Errorf("Expected tag 'anytls-proxy', got %s", config.Tag)
	}

	anytlsConfig := &configs.AnytlsClientConfig{}
	if err := config.Protocol.UnmarshalTo(anytlsConfig); err != nil {
		t.Fatalf("Failed to unmarshal anytls config: %v", err)
	}

	if anytlsConfig.Password != "password123" {
		t.Errorf("Expected password 'password123', got %s", anytlsConfig.Password)
	}

	// AnyTLS should have TLS configured
	if config.Transport == nil {
		t.Fatal("Transport config is nil")
	}

	// Note: In the current implementation, TLS is parsed from explicit "tls: true"
	// AnyTLS might need special handling to always enable TLS
}

func TestParseVlessWithReality(t *testing.T) {
	proxyMapping := map[string]any{
		"name":    "vless-reality",
		"type":    "vless",
		"server":  "example.com",
		"port":    443,
		"uuid":    "b831381d-6324-4d53-ad4f-8cda48b30811",
		"flow":    "xtls-rprx-vision",
		"network": "tcp",
		"reality-opts": map[string]any{
			"public-key": "test-public-key",
			"short-id":   "0123456789abcdef",
		},
		"sni":         "example.com",
		"fingerprint": "chrome",
	}

	config, err := clash.ParseProxy(proxyMapping)
	if err != nil {
		t.Fatalf("Failed to parse vless with reality: %v", err)
	}

	if config.Tag != "vless-reality" {
		t.Errorf("Expected tag 'vless-reality', got %s", config.Tag)
	}

	// Check transport has Reality config
	if config.Transport == nil {
		t.Fatal("Transport config is nil")
	}

	realityConfig := config.Transport.GetReality()
	if realityConfig == nil {
		t.Fatal("Expected Reality config")
	}

	if realityConfig.Pbk != "test-public-key" {
		t.Errorf("Expected public key 'test-public-key', got %s", realityConfig.Pbk)
	}

	if realityConfig.Sid != "0123456789abcdef" {
		t.Errorf("Expected short ID '0123456789abcdef', got %s", realityConfig.Sid)
	}

	if realityConfig.ServerName != "example.com" {
		t.Errorf("Expected server name 'example.com', got %s", realityConfig.ServerName)
	}

	if realityConfig.Fingerprint != "chrome" {
		t.Errorf("Expected fingerprint 'chrome', got %s", realityConfig.Fingerprint)
	}
}

func TestParseProxies(t *testing.T) {
	// Example: parsing a list of proxies from a Clash config
	proxiesArray := []any{
		map[string]any{
			"name":   "proxy1",
			"type":   "vmess",
			"server": "server1.com",
			"port":   443,
			"uuid":   "uuid1",
			"cipher": "auto",
		},
		map[string]any{
			"name":     "proxy2",
			"type":     "trojan",
			"server":   "server2.com",
			"port":     443,
			"password": "pass123",
		},
	}

	configs, _, err := clash.ParseProxies(proxiesArray)
	if err != nil {
		t.Fatalf("Failed to parse proxies: %v", err)
	}

	if len(configs) != 2 {
		t.Fatalf("Expected 2 configs, got %d", len(configs))
	}

	if configs[0].Tag != "proxy1" {
		t.Errorf("Expected tag 'proxy1', got %s", configs[0].Tag)
	}

	if configs[1].Tag != "proxy2" {
		t.Errorf("Expected tag 'proxy2', got %s", configs[1].Tag)
	}
}

func TestParseProxyWithH2Transport(t *testing.T) {
	proxyMapping := map[string]any{
		"name":    "h2-proxy",
		"type":    "vmess",
		"server":  "example.com",
		"port":    443,
		"uuid":    "b831381d-6324-4d53-ad4f-8cda48b30811",
		"cipher":  "auto",
		"tls":     true,
		"network": "h2",
		"h2-opts": map[string]any{
			"path": "/path",
			"host": []string{"example.com", "example2.com"},
		},
		"alpn": []string{"h2", "http/1.1"},
	}

	config, err := clash.ParseProxy(proxyMapping)
	if err != nil {
		t.Fatalf("Failed to parse h2 proxy: %v", err)
	}

	if config.Transport.GetHttp() == nil {
		t.Fatal("Expected HTTP/2 transport")
	}

	if config.Transport.GetHttp().Path != "/path" {
		t.Errorf("Expected path '/path', got %s", config.Transport.GetHttp().Path)
	}

	if len(config.Transport.GetHttp().Host) != 2 {
		t.Fatalf("Expected 2 hosts, got %d", len(config.Transport.GetHttp().Host))
	}
}

func TestParseProxyErrorCases(t *testing.T) {
	tests := []struct {
		name    string
		mapping map[string]any
		wantErr bool
	}{
		{
			name: "missing type",
			mapping: map[string]any{
				"name":   "test",
				"server": "example.com",
			},
			wantErr: true,
		},
		{
			name: "unsupported type",
			mapping: map[string]any{
				"name":   "test",
				"type":   "unsupported",
				"server": "example.com",
			},
			wantErr: true,
		},
		{
			name: "missing uuid for vmess",
			mapping: map[string]any{
				"name":   "test",
				"type":   "vmess",
				"server": "example.com",
				"port":   443,
			},
			wantErr: true,
		},
		{
			name: "invalid port",
			mapping: map[string]any{
				"name":   "test",
				"type":   "vmess",
				"server": "example.com",
				"uuid":   "test-uuid",
			},
			wantErr: true,
		},
		{
			name: "missing password for anytls",
			mapping: map[string]any{
				"name":   "test",
				"type":   "anytls",
				"server": "example.com",
				"port":   443,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := clash.ParseProxy(tt.mapping)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseProxy() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseClashConfig(t *testing.T) {
	t.Skip()
	configPath := "example-clash-config.yaml"
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	var clashConfig clash.ClashConfig
	if err := yaml.Unmarshal(data, &clashConfig); err != nil {
		t.Fatalf("Failed to parse YAML: %v", err)
	}

	outboundConfigs, failedReasons, err := clash.ParseProxies(clashConfig.Proxies)
	if err != nil {
		t.Fatalf("Failed to parse proxies: %v", err)
	}
	log.Debug().Msgf("failedReasons: %v", failedReasons)

	// Note: The test config has 2 proxies with unsupported cipher (aes-256-cfb)
	// which is a legacy cipher not supported in modern shadowsocks
	expectedFailed := 2
	expectedSuccess := len(clashConfig.Proxies) - expectedFailed

	if len(outboundConfigs) != expectedSuccess {
		t.Fatalf("Expected %d outbound configs (out of %d total, %d failed), got %d",
			expectedSuccess, len(clashConfig.Proxies), expectedFailed, len(outboundConfigs))
	}

	if len(failedReasons) != expectedFailed {
		t.Errorf("Expected %d failed proxies, got %d: %v", expectedFailed, len(failedReasons), failedReasons)
	}
}

// TestParseSocksProxy tests SOCKS5 proxy parsing
func TestParseSocksProxy(t *testing.T) {
	tests := []struct {
		name    string
		mapping map[string]any
		wantErr bool
	}{
		{
			name: "socks with auth",
			mapping: map[string]any{
				"name":     "socks-proxy",
				"type":     "socks5",
				"server":   "example.com",
				"port":     1080,
				"username": "user",
				"password": "pass",
			},
			wantErr: false,
		},
		{
			name: "socks without auth",
			mapping: map[string]any{
				"name":   "socks-proxy",
				"type":   "socks5",
				"server": "example.com",
				"port":   1080,
			},
			wantErr: false,
		},
		{
			name: "socks missing port",
			mapping: map[string]any{
				"name":   "socks-proxy",
				"type":   "socks5",
				"server": "example.com",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := clash.ParseProxy(tt.mapping)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseProxy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				socksConfig := &configs.SocksClientConfig{}
				if err := config.Protocol.UnmarshalTo(socksConfig); err != nil {
					t.Fatalf("Failed to unmarshal socks config: %v", err)
				}
				if username, ok := tt.mapping["username"].(string); ok {
					if socksConfig.Name != username {
						t.Errorf("Expected username '%s', got '%s'", username, socksConfig.Name)
					}
				}
			}
		})
	}
}

// TestParseHTTPProxy tests HTTP proxy parsing
func TestParseHTTPProxy(t *testing.T) {
	tests := []struct {
		name    string
		mapping map[string]any
		wantErr bool
	}{
		{
			name: "http with auth",
			mapping: map[string]any{
				"name":     "http-proxy",
				"type":     "http",
				"server":   "example.com",
				"port":     8080,
				"username": "user",
				"password": "pass",
			},
			wantErr: false,
		},
		{
			name: "http without auth",
			mapping: map[string]any{
				"name":   "http-proxy",
				"type":   "http",
				"server": "example.com",
				"port":   8080,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := clash.ParseProxy(tt.mapping)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseProxy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				httpConfig := &configs.HttpClientConfig{}
				if err := config.Protocol.UnmarshalTo(httpConfig); err != nil {
					t.Fatalf("Failed to unmarshal http config: %v", err)
				}
			}
		})
	}
}

// TestParseHysteriaProxy tests Hysteria2 proxy parsing
func TestParseHysteriaProxy(t *testing.T) {
	proxyMapping := map[string]any{
		"name":          "hysteria-proxy",
		"type":          "hysteria",
		"server":        "example.com",
		"port":          "443",
		"password":      "password123",
		"up":            "100 Mbps",
		"down":          "100 Mbps",
		"obfs":          "salamander",
		"obfs-password": "obfs123",
		"sni":           "example.com",
	}

	config, err := clash.ParseProxy(proxyMapping)
	if err != nil {
		t.Fatalf("Failed to parse hysteria proxy: %v", err)
	}

	if config.Tag != "hysteria-proxy" {
		t.Errorf("Expected tag 'hysteria-proxy', got %s", config.Tag)
	}

	hysteriaConfig := &configs.Hysteria2ClientConfig{}
	if err := config.Protocol.UnmarshalTo(hysteriaConfig); err != nil {
		t.Fatalf("Failed to unmarshal hysteria config: %v", err)
	}

	if hysteriaConfig.Auth != "password123" {
		t.Errorf("Expected auth 'password123', got %s", hysteriaConfig.Auth)
	}

	// Check obfs
	if hysteriaConfig.Obfs == nil {
		t.Fatal("Expected obfs config")
	}

	if hysteriaConfig.Obfs.GetSalamander().Password != "obfs123" {
		t.Errorf("Expected obfs password 'obfs123', got %s", hysteriaConfig.Obfs.GetSalamander().Password)
	}

	// Check bandwidth
	if hysteriaConfig.Bandwidth == nil {
		t.Fatal("Expected bandwidth config")
	}
}

// TestGetPort tests port parsing from various types
func TestGetPort(t *testing.T) {
	tests := []struct {
		name     string
		portVal  any
		expected uint32
	}{
		{"int port", 443, 443},
		{"float64 port", 443.0, 443},
		{"string port", "443", 443},
		{"invalid string", "invalid", 0},
		{"nil", nil, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We need to test this indirectly through ParseProxy
			proxyMapping := map[string]any{
				"name":     "test",
				"type":     "ss",
				"server":   "example.com",
				"port":     tt.portVal,
				"password": "pass",
				"cipher":   "aes-256-gcm",
			}

			config, err := clash.ParseProxy(proxyMapping)
			if tt.expected == 0 {
				if err == nil {
					t.Error("Expected error for invalid port")
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if config.Port != tt.expected {
					t.Errorf("Expected port %d, got %d", tt.expected, config.Port)
				}
			}
		})
	}
}

// TestVmessCipherTypes tests different vmess cipher types
func TestVmessCipherTypes(t *testing.T) {
	tests := []struct {
		cipher   string
		expected configs.SecurityType
	}{
		{"auto", configs.SecurityType_SecurityType_AUTO},
		{"aes-128-gcm", configs.SecurityType_SecurityType_AES128_GCM},
		{"chacha20-poly1305", configs.SecurityType_SecurityType_CHACHA20_POLY1305},
		{"none", configs.SecurityType_SecurityType_NONE},
		{"zero", configs.SecurityType_SecurityType_ZERO},
		{"unknown", configs.SecurityType_SecurityType_AUTO}, // defaults to AUTO
	}

	for _, tt := range tests {
		t.Run(tt.cipher, func(t *testing.T) {
			proxyMapping := map[string]any{
				"name":   "vmess-test",
				"type":   "vmess",
				"server": "example.com",
				"port":   443,
				"uuid":   "test-uuid",
				"cipher": tt.cipher,
			}

			config, err := clash.ParseProxy(proxyMapping)
			if err != nil {
				t.Fatalf("Failed to parse vmess: %v", err)
			}

			vmessConfig := &configs.VmessClientConfig{}
			if err := config.Protocol.UnmarshalTo(vmessConfig); err != nil {
				t.Fatalf("Failed to unmarshal vmess config: %v", err)
			}

			if vmessConfig.Security != tt.expected {
				t.Errorf("Expected security %v, got %v", tt.expected, vmessConfig.Security)
			}
		})
	}
}

// TestVmessAlterID tests alterID parsing from different types
func TestVmessAlterID(t *testing.T) {
	tests := []struct {
		name     string
		alterId  any
		expected uint32
	}{
		{"int alterId", 64, 64},
		{"float64 alterId", 64.0, 64},
		{"string alterId", "64", 64},
		{"zero alterId", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxyMapping := map[string]any{
				"name":    "vmess-test",
				"type":    "vmess",
				"server":  "example.com",
				"port":    443,
				"uuid":    "test-uuid",
				"alterId": tt.alterId,
			}

			config, err := clash.ParseProxy(proxyMapping)
			if err != nil {
				t.Fatalf("Failed to parse vmess: %v", err)
			}

			vmessConfig := &configs.VmessClientConfig{}
			if err := config.Protocol.UnmarshalTo(vmessConfig); err != nil {
				t.Fatalf("Failed to unmarshal vmess config: %v", err)
			}

			if vmessConfig.AlterId != tt.expected {
				t.Errorf("Expected alterId %d, got %d", tt.expected, vmessConfig.AlterId)
			}
		})
	}
}

// TestShadowsocksCipherTypes tests different shadowsocks cipher types
func TestShadowsocksCipherTypes(t *testing.T) {
	tests := []struct {
		cipher   string
		expected configs.ShadowsocksCipherType
		wantErr  bool
	}{
		{"aes-128-gcm", configs.ShadowsocksCipherType_AES_128_GCM, false},
		{"AEAD_AES_128_GCM", configs.ShadowsocksCipherType_AES_128_GCM, false},
		{"aes-256-gcm", configs.ShadowsocksCipherType_AES_256_GCM, false},
		{"AEAD_AES_256_GCM", configs.ShadowsocksCipherType_AES_256_GCM, false},
		{"chacha20-poly1305", configs.ShadowsocksCipherType_CHACHA20_POLY1305, false},
		{"chacha20-ietf-poly1305", configs.ShadowsocksCipherType_CHACHA20_POLY1305, false},
		{"AEAD_CHACHA20_POLY1305", configs.ShadowsocksCipherType_CHACHA20_POLY1305, false},
		{"none", configs.ShadowsocksCipherType_NONE, false},
		{"plain", configs.ShadowsocksCipherType_NONE, false},
		{"unsupported-cipher", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.cipher, func(t *testing.T) {
			proxyMapping := map[string]any{
				"name":     "ss-test",
				"type":     "ss",
				"server":   "example.com",
				"port":     8388,
				"password": "password",
				"cipher":   tt.cipher,
			}

			config, err := clash.ParseProxy(proxyMapping)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseProxy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				ssConfig := &configs.ShadowsocksClientConfig{}
				if err := config.Protocol.UnmarshalTo(ssConfig); err != nil {
					t.Fatalf("Failed to unmarshal ss config: %v", err)
				}

				if ssConfig.CipherType != tt.expected {
					t.Errorf("Expected cipher %v, got %v", tt.expected, ssConfig.CipherType)
				}
			}
		})
	}
}

// TestShadowsocksUDPOverTCP tests UDP-over-TCP flag
func TestShadowsocksUDPOverTCP(t *testing.T) {
	tests := []struct {
		name        string
		udpOverTCP  any
		expectedUot bool
	}{
		{"udp-over-tcp enabled", true, true},
		{"udp-over-tcp disabled", false, false},
		{"udp-over-tcp not set", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxyMapping := map[string]any{
				"name":     "ss-test",
				"type":     "ss",
				"server":   "example.com",
				"port":     8388,
				"password": "password",
				"cipher":   "aes-256-gcm",
			}

			if tt.udpOverTCP != nil {
				proxyMapping["udp-over-tcp"] = tt.udpOverTCP
			}

			config, err := clash.ParseProxy(proxyMapping)
			if err != nil {
				t.Fatalf("Failed to parse ss: %v", err)
			}

			if config.Uot != tt.expectedUot {
				t.Errorf("Expected Uot %v, got %v", tt.expectedUot, config.Uot)
			}
		})
	}
}

// TestParseProxiesWithFailures tests the failed nodes handling with yaml.Marshal
func TestParseProxiesWithFailures(t *testing.T) {
	proxiesArray := []any{
		// Valid proxy
		map[string]any{
			"name":     "valid-proxy",
			"type":     "ss",
			"server":   "example.com",
			"port":     8388,
			"password": "password",
			"cipher":   "aes-256-gcm",
		},
		// Invalid: missing required field
		map[string]any{
			"name":   "missing-password",
			"type":   "ss",
			"server": "example.com",
			"port":   8388,
			"cipher": "aes-256-gcm",
		},
		// Invalid: unsupported type
		map[string]any{
			"name":   "unsupported",
			"type":   "wireguard",
			"server": "example.com",
			"port":   51820,
		},
		// Not a map (will fail type assertion)
		"invalid-string-proxy",
	}

	configs, failedNodes, err := clash.ParseProxies(proxiesArray)
	if err != nil {
		t.Fatalf("ParseProxies returned error: %v", err)
	}

	if len(configs) != 1 {
		t.Errorf("Expected 1 valid config, got %d", len(configs))
	}

	if len(failedNodes) != 3 {
		t.Errorf("Expected 3 failed nodes, got %d", len(failedNodes))
	}

	// Check that failed nodes are formatted (contain yaml or string representation)
	for i, failedNode := range failedNodes {
		if failedNode == "" {
			t.Errorf("Failed node %d is empty", i)
		}
		t.Logf("Failed node %d: %s", i, failedNode)
	}
}

// TestVlessEncryptionDefault tests vless encryption defaults to "none"
func TestVlessEncryptionDefault(t *testing.T) {
	proxyMapping := map[string]any{
		"name":   "vless-test",
		"type":   "vless",
		"server": "example.com",
		"port":   443,
		"uuid":   "test-uuid",
		// encryption not specified
	}

	config, err := clash.ParseProxy(proxyMapping)
	if err != nil {
		t.Fatalf("Failed to parse vless: %v", err)
	}

	vlessConfig := &configs.VlessClientConfig{}
	if err := config.Protocol.UnmarshalTo(vlessConfig); err != nil {
		t.Fatalf("Failed to unmarshal vless config: %v", err)
	}

	if vlessConfig.Encryption != "none" {
		t.Errorf("Expected encryption 'none', got '%s'", vlessConfig.Encryption)
	}
}

// TestAnytlsSessionParameters tests anytls session parameters
func TestAnytlsSessionParameters(t *testing.T) {
	proxyMapping := map[string]any{
		"name":                        "anytls-test",
		"type":                        "anytls",
		"server":                      "example.com",
		"port":                        443,
		"password":                    "password123",
		"idle-session-check-interval": 30,
		"idle-session-timeout":        60,
		"min-idle-session":            4,
	}

	config, err := clash.ParseProxy(proxyMapping)
	if err != nil {
		t.Fatalf("Failed to parse anytls: %v", err)
	}

	anytlsConfig := &configs.AnytlsClientConfig{}
	if err := config.Protocol.UnmarshalTo(anytlsConfig); err != nil {
		t.Fatalf("Failed to unmarshal anytls config: %v", err)
	}

	if anytlsConfig.IdleSessionCheckInterval != 30 {
		t.Errorf("Expected IdleSessionCheckInterval 30, got %d", anytlsConfig.IdleSessionCheckInterval)
	}

	if anytlsConfig.IdleSessionTimeout != 60 {
		t.Errorf("Expected IdleSessionTimeout 60, got %d", anytlsConfig.IdleSessionTimeout)
	}

	if anytlsConfig.MinIdleSession != 4 {
		t.Errorf("Expected MinIdleSession 4, got %d", anytlsConfig.MinIdleSession)
	}
}

// TestParseProxyWithMultipleHostHeaders tests websocket with multiple host headers
func TestParseProxyWithMultipleHostHeaders(t *testing.T) {
	proxyMapping := map[string]any{
		"name":    "ws-multi-host",
		"type":    "vmess",
		"server":  "example.com",
		"port":    443,
		"uuid":    "test-uuid",
		"network": "ws",
		"ws-opts": map[string]any{
			"path": "/path",
			"headers": map[string]any{
				"Host":       "host1.com",
				"User-Agent": "Mozilla/5.0",
			},
		},
	}

	config, err := clash.ParseProxy(proxyMapping)
	if err != nil {
		t.Fatalf("Failed to parse proxy: %v", err)
	}

	if config.Transport.GetWebsocket() == nil {
		t.Fatal("Expected websocket transport")
	}

	ws := config.Transport.GetWebsocket()
	if ws.Path != "/path" {
		t.Errorf("Expected path '/path', got %s", ws.Path)
	}

	// Check headers were parsed
	if len(ws.Header) == 0 {
		t.Error("Expected headers to be set")
	}
}

// TestParseProxyMissingName tests proxy without name field
func TestParseProxyMissingName(t *testing.T) {
	proxyMapping := map[string]any{
		"type":     "ss",
		"server":   "example.com",
		"port":     8388,
		"password": "password",
		"cipher":   "aes-256-gcm",
	}

	config, err := clash.ParseProxy(proxyMapping)
	if err != nil {
		t.Fatalf("Failed to parse proxy: %v", err)
	}

	// Tag should be empty string when name is not provided
	if config.Tag != "" {
		t.Errorf("Expected empty tag, got '%s'", config.Tag)
	}
}
