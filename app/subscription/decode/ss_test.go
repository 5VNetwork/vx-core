package decode_test

import (
	"fmt"
	"testing"

	"github.com/5vnetwork/vx-core/app/configs/proxy"
	. "github.com/5vnetwork/vx-core/app/subscription/decode"
	"github.com/5vnetwork/vx-core/common/serial"
)

// func Test
func TestParseSsFromLink(t *testing.T) {
	const ssLink = "ss://Y2hhY2hhMjAtaWV0Zi1wb2x5MTMwNTp0ZXN0@test.test:4621#test%E5%95%8A"

	config, err := ParseSsFromLink(ssLink)
	if err != nil {
		t.Fatalf("failed to parse ss link: %v", err)
	}
	ssConfig0, err := serial.GetInstanceOf(config.Protocol)
	if err != nil {
		t.Fatalf("failed to get ss config: %v", err)
	}
	ssConfig := ssConfig0.(*proxy.ShadowsocksClientConfig)

	fmt.Println(ssConfig)
	if ssConfig.CipherType != proxy.ShadowsocksCipherType_CHACHA20_POLY1305 {
		t.Fatalf("cipher is not chacha20-ietf-poly1305")
	}
	if ssConfig.Password != "test" {
		t.Fatalf("password is not test")
	}
	if config.Address != "test.test" {
		t.Fatalf("address is not test.test")
	}
	if config.Ports[0].From != 4621 {
		t.Fatalf("port is not 4621")
	}
	if config.Tag != "test啊" {
		t.Fatalf("remark is not test啊")
	}
}
