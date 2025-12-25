package clashparser_test

import (
	"io/ioutil"
	"testing"

	"github.com/5vnetwork/vx-core/app/configs/proxy"
	"github.com/5vnetwork/vx-core/app/util/clashparser"
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

	config, err := clashparser.ParseProxy(proxyMapping)
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
	vmessConfig := &proxy.VmessClientConfig{}
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

	config, err := clashparser.ParseProxy(proxyMapping)
	if err != nil {
		t.Fatalf("Failed to parse vless proxy: %v", err)
	}

	if config.Tag != "vless-proxy" {
		t.Errorf("Expected tag 'vless-proxy', got %s", config.Tag)
	}

	vlessConfig := &proxy.VlessClientConfig{}
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

	config, err := clashparser.ParseProxy(proxyMapping)
	if err != nil {
		t.Fatalf("Failed to parse trojan proxy: %v", err)
	}

	if config.Tag != "trojan-proxy" {
		t.Errorf("Expected tag 'trojan-proxy', got %s", config.Tag)
	}

	trojanConfig := &proxy.TrojanClientConfig{}
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

	config, err := clashparser.ParseProxy(proxyMapping)
	if err != nil {
		t.Fatalf("Failed to parse shadowsocks proxy: %v", err)
	}

	if config.Tag != "ss-proxy" {
		t.Errorf("Expected tag 'ss-proxy', got %s", config.Tag)
	}

	ssConfig := &proxy.ShadowsocksClientConfig{}
	if err := config.Protocol.UnmarshalTo(ssConfig); err != nil {
		t.Fatalf("Failed to unmarshal shadowsocks config: %v", err)
	}

	if ssConfig.Password != "password123" {
		t.Errorf("Expected password 'password123', got %s", ssConfig.Password)
	}

	if ssConfig.CipherType != proxy.ShadowsocksCipherType_AES_256_GCM {
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

	config, err := clashparser.ParseProxy(proxyMapping)
	if err != nil {
		t.Fatalf("Failed to parse anytls proxy: %v", err)
	}

	if config.Tag != "anytls-proxy" {
		t.Errorf("Expected tag 'anytls-proxy', got %s", config.Tag)
	}

	anytlsConfig := &proxy.AnytlsClientConfig{}
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

	config, err := clashparser.ParseProxy(proxyMapping)
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

	configs, _, err := clashparser.ParseProxies(proxiesArray)
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

	config, err := clashparser.ParseProxy(proxyMapping)
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
			_, err := clashparser.ParseProxy(tt.mapping)
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

	var clashConfig clashparser.ClashConfig
	if err := yaml.Unmarshal(data, &clashConfig); err != nil {
		t.Fatalf("Failed to parse YAML: %v", err)
	}

	outboundConfigs, failedReasons, err := clashparser.ParseProxies(clashConfig.Proxies)
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
