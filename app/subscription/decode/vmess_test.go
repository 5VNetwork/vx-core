package decode_test

import (
	"fmt"
	"testing"

	"github.com/5vnetwork/vx-core/app/configs/proxy"
	"github.com/5vnetwork/vx-core/app/subscription/decode"
	"github.com/5vnetwork/vx-core/common/serial"
)

func TestParseVmessFromLink(t *testing.T) {
	const vmessLink = `{
	"v": "2",
	"ps": " 备注或别名",
	"add": "111.111.111.111",
	"port": "32000",
	"id": "1111115e-1111-1111-1111-111111111111",
	"aid": "100",
	"scy": "zero",
	"net": "tcp",
	"type": "none",
	"host": "www.bbb.com",
	"path": "/",
	"tls": "tls",
	"sni": "www.ccc.com",
	"alpn": "h2",
	"fp": "chrome"
	}`

	vmessConfig, err := decode.ParseVmessFromJSON([]byte(vmessLink))
	if err != nil {
		t.Fatalf("failed to parse vmess link: %v", err)
	}

	fmt.Println(vmessConfig)
	handlerConfig, err := vmessConfig.ToProxyHandlerConfig()
	if err != nil {
		t.Fatalf("failed to convert vmess config to handler config: %v", err)
	}
	fmt.Println(handlerConfig)

	if handlerConfig.Address != "111.111.111.111" {
		t.Fatalf("address is not 111.111.111.111")
	}
	if handlerConfig.Ports[0].From != 32000 {
		t.Fatalf("port is not 32000")
	}
	if handlerConfig.Transport == nil {
		t.Fatalf("transport is nil")
	}
	if handlerConfig.Protocol.TypeUrl != "type.googleapis.com/x.proxy.VmessClientConfig" {
		t.Fatalf("protocol is not VmessClientConfig")
	}
	pm, err := serial.GetInstanceOf(handlerConfig.Protocol)
	if err != nil {
		t.Fatalf("failed to get vmess client config: %v", err)
	}
	vmessClientConfig, ok := pm.(*proxy.VmessClientConfig)
	if !ok {
		t.Fatalf("vmessClientConfig is not VmessClientConfig")
	}
	if vmessClientConfig.Id != "1111115e-1111-1111-1111-111111111111" {
		t.Fatalf("id is not 1111115e-1111-1111-1111-111111111111")
	}
}
