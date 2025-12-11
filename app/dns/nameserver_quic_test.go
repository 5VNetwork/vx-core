package dns

import (
	"context"
	"log"
	"testing"

	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/proxy/freedom"
	"github.com/5vnetwork/vx-core/test"
	"github.com/5vnetwork/vx-core/transport"
	d "github.com/miekg/dns"
)

func TestNameServerQuic(t *testing.T) {
	test.InitZeroLog()
	doh, _ := NewQUICNameServer(QuicNameServerOption{
		Name:        "test",
		Destination: net.UDPDestination(net.DomainAddress("dns.adguard.com"), 853),
		Handler: freedom.New(
			transport.DefaultDialer,
			transport.DefaultPacketListener,
			"", nil,
		),
		IPResolver: &DnsResolver{},
	})

	m := new(d.Msg)
	m.SetQuestion("www.apple.com.", d.TypeA)
	reply, err := doh.HandleQuery(context.Background(), m, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(reply.Answer) == 0 {
		t.Fatal("expected at least one answer")
	}
	log.Print("Reply:", reply)
}
