package decode_test

import (
	"testing"

	"github.com/5vnetwork/vx-core/app/subscription/decode"
	"github.com/5vnetwork/vx-core/common"
)

var socksLink = "socks5://admin:abc567%40@23.95.165.11:22543"

func TestParseSocksFromLink(t *testing.T) {
	config, err := decode.ParseSocks5FromLink(socksLink)
	common.Must(err)
	t.Log(config)
}
