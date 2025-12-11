package decode

import (
	"bytes"
	"testing"

	"github.com/5vnetwork/vx-core/app/configs/proxy"
	"github.com/5vnetwork/vx-core/common/serial"
)

func TestParseHysteriaFromLink(t *testing.T) {
	config, err := ParseHysteriaFromLink("hysteria2://12345678-1234-1234-1234-123456789012@a.b.com:12345?insecure=0&obfs=salamander&obfs-password=asdfwqr&sni=a.b.com#%E6%B5%8B%E8%AF%95vless")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(config)
	hysteriaConfig, err := serial.GetInstanceOf(config.Protocol)
	if err != nil {
		t.Fatal(err)
	}
	hysteria := hysteriaConfig.(*proxy.Hysteria2ClientConfig)
	if hysteria.Auth != "12345678-1234-1234-1234-123456789012" {
		t.Fatal("auth mismatch")
	}
	if hysteria.TlsConfig.ServerName != "a.b.com" {
		t.Fatal("sni mismatch")
	}
	if hysteria.Obfs.Obfs.(*proxy.ObfsConfig_Salamander).Salamander.Password != "asdfwqr" {
		t.Fatal("obfs password mismatch")
	}
	if config.Address != "a.b.com" {
		t.Fatal("addr mismatch")
	}
	if config.Ports[0].From != 12345 || config.Ports[0].To != 12345 {
		t.Fatal("port mismatch")
	}
	if config.Tag != "测试vless" {
		t.Fatal("tag mismatch")
	}
}

func TestParseHysteriaFromLink1(t *testing.T) {
	config, err := ParseHysteriaFromLink("hysteria2://username:password@a.b.com:123,5000-6000?insecure=0&obfs=salamander&obfs-password=asdfwqr&sni=a.b.com#%E6%B5%8B%E8%AF%95vless")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(config)
	hysteriaConfig, err := serial.GetInstanceOf(config.Protocol)
	if err != nil {
		t.Fatal(err)
	}
	hysteria := hysteriaConfig.(*proxy.Hysteria2ClientConfig)
	if hysteria.Auth != "username:password" {
		t.Fatal("auth mismatch")
	}
	if hysteria.TlsConfig.ServerName != "a.b.com" {
		t.Fatal("sni mismatch")
	}
	if hysteria.Obfs.Obfs.(*proxy.ObfsConfig_Salamander).Salamander.Password != "asdfwqr" {
		t.Fatal("obfs password mismatch")
	}
	if config.Address != "a.b.com" {
		t.Fatal("addr mismatch")
	}
	if config.Ports[0].From != 123 || config.Ports[0].To != 123 {
		t.Fatal("port mismatch")
	}
	if config.Ports[1].From != 5000 || config.Ports[1].To != 6000 {
		t.Fatal("port mismatch")
	}
	if config.Tag != "测试vless" {
		t.Fatal("tag mismatch")
	}
}

// no port, username:password
func TestParseHysteriaFromLink2(t *testing.T) {
	config, err := ParseHysteriaFromLink("hysteria2://username:password@a.b.com:?insecure=0&obfs=salamander&obfs-password=asdfwqr&sni=a.b.com#%E6%B5%8B%E8%AF%95vless")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(config)
	hysteriaConfig, err := serial.GetInstanceOf(config.Protocol)
	if err != nil {
		t.Fatal(err)
	}
	hysteria := hysteriaConfig.(*proxy.Hysteria2ClientConfig)
	if hysteria.Auth != "username:password" {
		t.Fatal("auth mismatch")
	}
	if hysteria.TlsConfig.ServerName != "a.b.com" {
		t.Fatal("sni mismatch")
	}
	if hysteria.Obfs.Obfs.(*proxy.ObfsConfig_Salamander).Salamander.Password != "asdfwqr" {
		t.Fatal("obfs password mismatch")
	}
	if config.Address != "a.b.com" {
		t.Fatal("addr mismatch")
	}
	if config.Ports[0].From != 443 || config.Ports[0].To != 443 {
		t.Fatal("port mismatch")
	}
	if config.Tag != "测试vless" {
		t.Fatal("tag mismatch")
	}
}

// no port, sha256 pin
func TestParseHysteriaFromLink3(t *testing.T) {
	config, err := ParseHysteriaFromLink("hysteria2://username:password@a.b.com?insecure=0&obfs=salamander&obfs-password=asdfwqr&sni=a.b.com&pinSHA256=0974c9cc9c2f97510351942356978ace170904a3553727a29248a9dac74e30b6#%E6%B5%8B%E8%AF%95vless")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(config)
	hysteriaConfig, err := serial.GetInstanceOf(config.Protocol)
	if err != nil {
		t.Fatal(err)
	}
	hysteria := hysteriaConfig.(*proxy.Hysteria2ClientConfig)
	if hysteria.Auth != "username:password" {
		t.Fatal("auth mismatch")
	}
	if hysteria.TlsConfig.ServerName != "a.b.com" {
		t.Fatal("sni mismatch")
	}
	if hysteria.Obfs.Obfs.(*proxy.ObfsConfig_Salamander).Salamander.Password != "asdfwqr" {
		t.Fatal("obfs password mismatch")
	}
	if len(hysteria.TlsConfig.PinnedPeerCertificateChainSha256) != 1 {
		t.Fatal("pinSHA256 mismatch")
	}
	if !bytes.Equal(hysteria.TlsConfig.PinnedPeerCertificateChainSha256[0], []byte{0x09, 0x74, 0xc9, 0xcc, 0x9c, 0x2f, 0x97, 0x51, 0x03, 0x51, 0x94, 0x23, 0x56, 0x97, 0x8a, 0xce, 0x17, 0x09, 0x04, 0xa3, 0x55, 0x37, 0x27, 0xa2, 0x92, 0x48, 0xa9, 0xda, 0xc7, 0x4e, 0x30, 0xb6}) {
		t.Fatal("pinSHA256 mismatch")
	}
	if config.Address != "a.b.com" {
		t.Fatal("addr mismatch")
	}
	if config.Ports[0].From != 443 || config.Ports[0].To != 443 {
		t.Fatal("port mismatch")
	}
	if config.Tag != "测试vless" {
		t.Fatal("tag mismatch")
	}
}

func TestExtractPortFromURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "standard port",
			input:    "http://example.com:8080/path",
			expected: "",
			wantErr:  false,
		},
		{
			name:     "port with authentication",
			input:    "http://user:pass@example.com:8080/path",
			expected: "",
			wantErr:  false,
		},
		{
			name:     "multiple ports",
			input:    "hysteria2://example.com:123,5000-6000/?insecure=1",
			expected: "123,5000-6000",
			wantErr:  false,
		},
		{
			name:     "port range",
			input:    "example.com:5000-6000/",
			expected: "5000-6000",
			wantErr:  false,
		},
		{
			name:     "no port",
			input:    "example.com",
			expected: "",
			wantErr:  false,
		},
		{
			name:     "no port with path",
			input:    "example.com/",
			expected: "",
			wantErr:  false,
		},
		{
			name:     "no port with query",
			input:    "example.com?param=value",
			expected: "",
			wantErr:  false,
		},
		{
			name:     "port with query",
			input:    "example.com:8080?param=value",
			expected: "",
			wantErr:  false,
		},
		{
			name:     "port with path and query",
			input:    "example.com:8080/path?param=value",
			expected: "",
			wantErr:  false,
		},
		{
			name:     "complex hysteria2 url",
			input:    "hysteria2://letmein@example.com:123,5000-6000/?insecure=1&obfs=salamander&obfs-password=gawrgura&pinSHA256=deadbeef&sni=real.example.com",
			expected: "123,5000-6000",
			wantErr:  false,
		},
		{
			name:     "empty port after colon",
			input:    "example.com:/path",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "port with special characters",
			input:    "example.com:8080,9090-10000/path",
			expected: "8080,9090-10000",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractHysteriaPortFromURL(tt.input)
			if got != tt.expected {
				t.Errorf("extractPortFromURL() = %v, want %v", got, tt.expected)
			}
		})
	}
}
