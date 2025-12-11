//go:build windows
// +build windows

package gvisor

import (
	"log"
	"net"
	"net/netip"
	"testing"

	"github.com/5vnetwork/vx-core/app/inbound/reject"
	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/test"
	"github.com/5vnetwork/vx-core/test/mocks"
	"github.com/5vnetwork/vx-core/tun"
	"gvisor.dev/gvisor/pkg/tcpip/header"
)

var tunOption2 = &tun.TunOption{
	Mtu:  1500,
	Name: "tesaf",
	Ip4:  netip.MustParsePrefix("172.13.17.1/24"),
	Ip6:  netip.MustParsePrefix("fc12::1/64"),
	Route4: []netip.Prefix{
		netip.MustParsePrefix("1.2.3.4/32"),
	},
	Route6: []netip.Prefix{
		netip.MustParsePrefix("2001:db8::1/128"),
	},
	Path: "../../../tun/wintun",
}

func TestFilterLinkEndpoint(t *testing.T) {
	if !test.IsAdmin() {
		t.Skip("Skipping test because it requires admin privileges")
	}
	tn, err := tun.NewTun(tunOption2)
	if err != nil {
		t.Errorf("Failed to create tun: %v", err)
	}

	td := NewTunLinkEndpoint(tn, 1500)
	common.Must(td.Start())
	defer td.Close()

	opt := GvisorInboundOption{
		TcpOnly:      true,
		Handler:      &mocks.LoopbackHandler{},
		LinkEndpoint: NewFilterLinkEndpoint(td, &mockRejector{}, false),
	}
	gi, err := NewGvisorInbound(&opt)
	if err != nil {
		t.Errorf("Failed to init gvisor: %v", err)
	}
	common.Must(gi.Start())
	defer gi.Close()

	_, err = net.Dial("tcp6", "[2001:db8::1]:80")
	if err == nil {
		t.Fatal("dial success, which should not happen")
	}
	log.Print(err)
	t.Log(err)

	conn, err := net.Dial("tcp4", "1.2.3.4:80")
	if err != nil {
		t.Fatal("dial failed, which should not happen")
	}
	conn.Close()
}

type mockRejector struct{}

func (m *mockRejector) Reject(p []byte) *buf.Buffer {
	if header.IPVersion(p) == header.IPv6Version {
		ipv6 := header.IPv6(p)
		if ipv6.TransportProtocol() == header.TCPProtocolNumber {
			tcp := header.TCP(p[header.IPv6MinimumSize:])
			return reject.GenerateRstForTcpSynIPv60(ipv6, tcp)
		}
	}
	return nil
}
