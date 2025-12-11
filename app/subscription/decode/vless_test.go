package decode

import (
	"testing"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/configs/proxy"
	"github.com/5vnetwork/vx-core/common/serial"
)

func TestParseVlessFromLink(t *testing.T) {
	link := "vless://12345678-1234-1234-1234-123456789012@a.b.com:12345?encryption=none&flow=xtls-rprx-vision&fp=chrome&network=ws&path=%2Fvvv&security=tls&sni=a.b.com#%E6%B5%8B%E8%AF%95vless"
	cfg, err := ParseVlessFromLink(link)
	if err != nil {
		t.Fatal(err)
	}
	vlessConfig, err := serial.GetInstanceOf(cfg.Protocol)
	if err != nil {
		t.Fatal(err)
	}
	vless := vlessConfig.(*proxy.VlessClientConfig)
	if vless.Id != "12345678-1234-1234-1234-123456789012" {
		t.Fatal("id mismatch")
	}
	if vless.Encryption != "none" {
		t.Fatal("encryption mismatch")
	}
	if vless.Flow != "xtls-rprx-vision" {
		t.Fatal("flow mismatch")
	}
	if cfg.Address != "a.b.com" {
		t.Fatal("addr wrong")
	}
	if cfg.Ports[0].To != 12345 {
		t.Fatal("port wrong")
	}
	if cfg.Tag != "测试vless" {
		t.Fatal("tag wrong")
	}
	transportConfig := cfg.Transport
	if transportConfig.Protocol == nil {
		t.Fatal("transport config is nil")
	}
	if transportConfig.Protocol.(*configs.TransportConfig_Websocket).Websocket.Path != "/vvv" {
		t.Fatal("path wrong")
	}
}

func TestDecodeVless(t *testing.T) {
	sub, _ := Decode("vless://12345678-1234-1234-1234-123456789012@1.1.1.1:443?encryption=none&flow=xtls-rprx-vision&security=tls&sni=asdf.com&alpn=h2%2Chttp%2F1.1&fp=chrome&allowInsecure=1&type=tcp&headerType=none#a")
	if len(sub.Configs) != 1 {
		t.Errorf("Expected 1 config, got %d", len(sub.Configs))
	}
	config := sub.Configs[0]
	if config.Address != "1.1.1.1" {
		t.Fatal("address wrong")
	}
	if config.Ports[0].From != 443 {
		t.Fatal("port wrong")
	}
	vlessConfig, err := serial.GetInstanceOf(config.Protocol)
	if err != nil {
		t.Fatal(err)
	}
	vless := vlessConfig.(*proxy.VlessClientConfig)
	if vless.Id != "12345678-1234-1234-1234-123456789012" {
		t.Fatal("id wrong")
	}
	if vless.Encryption != "none" {
		t.Fatal("encryption wrong")
	}
	if vless.Flow != "xtls-rprx-vision" {
		t.Fatal("flow wrong")
	}
	tlsConfig := config.Transport.GetTls()
	if tlsConfig.ServerName != "asdf.com" {
		t.Fatal("sni wrong")
	}
	if len(tlsConfig.NextProtocol) != 2 {
		t.Fatal("alpn wrong")
	}
	if tlsConfig.Imitate != "chrome" {
		t.Fatal("fp wrong")
	}
	if !tlsConfig.AllowInsecure {
		t.Fatal("allowInsecure wrong")
	}
}
