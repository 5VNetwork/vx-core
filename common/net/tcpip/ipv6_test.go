package tcpip

import (
	"net"
	"testing"

	"github.com/5vnetwork/vx-core/common/buf"

	"golang.org/x/net/ipv6"
)

func TestSetFields(t *testing.T) {
	b := buf.NewWithSize(1500)
	p := IPv6Packet(b.Extend(1500))
	src := net.ParseIP("240e:83:205:58:0:ff:b09f:36bf")
	dst := net.ParseIP("240e:83:205:58:0:ff:b09f:1234")
	p.SetFields(src, dst, TCP, 0)
	h, err := ipv6.ParseHeader(p[:ipv6.HeaderLen])
	if err != nil {
		t.Fatal(err)
	}
	if h.Src.String() != src.String() {
		t.Fatalf("expected src %v, got %v", src, h.Src)
	}
	if h.Dst.String() != dst.String() {
		t.Fatalf("expected dst %v, got %v", dst, h.Dst)
	}
}
