package uri

import (
	"log"
	"testing"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/configs/proxy"
	"github.com/5vnetwork/vx-core/common/serial"
	"github.com/5vnetwork/vx-core/transport/protocols/websocket"
	"github.com/5vnetwork/vx-core/transport/security/tls"

	"github.com/google/go-cmp/cmp"
)

const id = "12345678-1234-1234-1234-123456789012"

func TestVmess(t *testing.T) {
	config := &configs.OutboundHandlerConfig{
		Tag:     "Test node",
		Port:    47845,
		Address: "localhost",
		Protocol: serial.ToTypedMessage(
			&proxy.VmessClientConfig{
				Id: id,
			},
		),
		Transport: &configs.TransportConfig{
			Protocol: &configs.TransportConfig_Websocket{
				Websocket: &websocket.WebsocketConfig{
					Path: "v",
				},
			},
			Security: &configs.TransportConfig_Tls{
				Tls: &tls.TlsConfig{
					NextProtocol: []string{"h2", "http/1.1"},
					Imitate:      "chrome",
					ServerName:   "a.b.com",
				},
			},
		},
	}
	s, err := ToUrl(config)
	if err != nil {
		t.Fatal(err)
	}
	log.Println("[", s, "]")
}

func TestShadowsocks(t *testing.T) {
	config := &configs.OutboundHandlerConfig{
		Tag:     "测试shadowsocks",
		Port:    12345,
		Address: "a.b.com",
		Protocol: serial.ToTypedMessage(
			&proxy.ShadowsocksClientConfig{
				Password:   id,
				CipherType: proxy.ShadowsocksCipherType_AES_128_GCM,
			},
		),
	}
	s, err := ToUrl(config)
	if err != nil {
		t.Fatal(err)
	}
	log.Println("[", s, "]")

	if s := cmp.Diff(s, `ss://YWVzLTEyOC1nY206MTIzNDU2NzgtMTIzNC0xMjM0LTEyMzQtMTIzNDU2Nzg5MDEy@a.b.com:12345#%E6%B5%8B%E8%AF%95shadowsocks`); s != "" {
		t.Fatal(s)
	}
}

func TestShadowsocks0(t *testing.T) {
	config := &configs.OutboundHandlerConfig{
		Tag:     "X",
		Port:    443,
		Address: "1.1.1.1",
		Protocol: serial.ToTypedMessage(
			&proxy.VlessClientConfig{
				Id:         id,
				Flow:       "xtls-rprx-vision",
				Encryption: "none",
			},
		),
		Transport: &configs.TransportConfig{
			Security: &configs.TransportConfig_Tls{
				Tls: &tls.TlsConfig{
					ServerName:   "www.adsf.site",
					NextProtocol: []string{"http/1.1"},
				},
			},
		},
	}
	s, err := ToUrl(config)
	if err != nil {
		t.Fatal(err)
	}
	log.Println("[", s, "]")
}

func TestTrojan(t *testing.T) {
	config := &configs.OutboundHandlerConfig{
		Tag:     "测试trojan",
		Port:    12345,
		Address: "a.b.com",
		Protocol: serial.ToTypedMessage(
			&proxy.TrojanClientConfig{
				Password: id,
			},
		),
		Transport: &configs.TransportConfig{
			Protocol: &configs.TransportConfig_Websocket{
				Websocket: &websocket.WebsocketConfig{
					Path: "/vvv",
				},
			},
			Security: &configs.TransportConfig_Tls{
				Tls: &tls.TlsConfig{
					NextProtocol:                     []string{"h2", "http/1.1"},
					Imitate:                          "chrome",
					ServerName:                       "a.b.com",
					AllowInsecure:                    true,
					PinnedPeerCertificateChainSha256: [][]byte{[]byte("000009cc900000510351942000078ace170904a3553727a200009dac74e30b6")},
				},
			},
		},
	}
	s, err := ToUrl(config)
	if err != nil {
		t.Fatal(err)
	}
	s1, err := toTrojan0(config)
	if err != nil {
		t.Fatal(err)
	}
	if s := cmp.Diff(s, s1); s != "" {
		t.Fatal(s)
	}
	t.Log(s)
}

func TestVless(t *testing.T) {
	config := &configs.OutboundHandlerConfig{
		Tag:     "测试vless",
		Port:    12345,
		Address: "a.b.com",
		Protocol: serial.ToTypedMessage(
			&proxy.VlessClientConfig{
				Id:         id,
				Flow:       "xtls-rprx-vision",
				Encryption: "none",
			},
		),
		Transport: &configs.TransportConfig{
			Protocol: &configs.TransportConfig_Websocket{
				Websocket: &websocket.WebsocketConfig{
					Path: "/vvv",
				},
			},
			Security: &configs.TransportConfig_Tls{
				Tls: &tls.TlsConfig{
					NextProtocol: []string{"h2", "http/1.1"},
					Imitate:      "chrome",
					ServerName:   "a.b.com",
				},
			},
		},
	}
	s, err := toVless0(config)
	if err != nil {
		t.Fatal(err)
	}
	log.Println("[", s, "]")
}

func TestHysteria(t *testing.T) {
	config := &configs.OutboundHandlerConfig{
		Tag:     "测试vless",
		Port:    12345,
		Address: "a.b.com",
		Protocol: serial.ToTypedMessage(
			&proxy.Hysteria2ClientConfig{
				Auth: id,
				TlsConfig: &tls.TlsConfig{
					ServerName: "a.b.com",
				},
				Obfs: &proxy.ObfsConfig{
					Obfs: &proxy.ObfsConfig_Salamander{
						Salamander: &proxy.SalamanderConfig{
							Password: "asdfwqr",
						},
					},
				},
			},
		),
	}
	s, err := ToUrl(config)
	if err != nil {
		t.Fatal(err)
	}
	log.Println("[", s, "]")
}
