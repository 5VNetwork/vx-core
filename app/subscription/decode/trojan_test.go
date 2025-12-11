package decode

import (
	"fmt"
	"testing"

	"github.com/5vnetwork/vx-core/app/configs/proxy"
	"github.com/5vnetwork/vx-core/common/serial"
)

func TestParseTrojanFromLink(t *testing.T) {
	trojan := "trojan://12345678-1234-1234-1234-123456789012@123.com:17919?allowInsecure=1&peer=&sni=&udp=1&type=ws#test"

	config, err := ParseTrojan(trojan)
	if err != nil {
		t.Fatalf("failed to parse trojan link: %v", err)
	}

	fmt.Println(config)

	trojanConfig0, err := serial.GetInstanceOf(config.Protocol)
	if err != nil {
		t.Fatalf("failed to get trojan client config: %v", err)
	}

	trojanConfig := trojanConfig0.(*proxy.TrojanClientConfig)

	if trojanConfig.Password != "12345678-1234-1234-1234-123456789012" {
		t.Fatal("secret is not 12345678-1234-1234-1234-123456789012")
	}
	if config.Address != "123.com" {
		t.Fatal("address is not 123.com")
	}
	if config.Ports[0].From != 17919 {
		t.Fatal("port is not 17919")
	}
	if config.Tag != "test" {
		t.Fatal("remark is not test")
	}

	if config.Transport.Security == nil {
		t.Fatal("transport security is nil")
	}
	tlsConfig := config.Transport.GetTls()
	if !tlsConfig.AllowInsecure {
		t.Fatal("allowInsecure is not true")
	}

	wsConfig := config.Transport.GetWebsocket()
	if wsConfig == nil {
		t.Fatal("websocket config is nil")
	}
}
