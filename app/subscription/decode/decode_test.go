package decode_test

import (
	"log"
	"testing"

	"github.com/5vnetwork/vx-core/app/subscription/decode"
	"github.com/5vnetwork/vx-core/common"
)

//TODO: support this
// vless://[uuid]@1.1.1.1:443?encryption=none&security=tls&sni=a.b.com&fp
// =random&type=ws&host=a.b.com&path=%3&allowInsecure=1&fragment=1,40-60,30-50,tlshello#1.1.1.1

func TestDecode(t *testing.T) {
	r, err := decode.Decode(a)
	common.Must(err)
	log.Println(r)
}

const a = ""
