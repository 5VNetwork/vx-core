//go:build windows

package system

import (
	"log"
	"net"
	gonet "net"
	"net/netip"
	"os"
	"testing"

	"github.com/5vnetwork/vx-core/common/strings"
	"github.com/5vnetwork/vx-core/test"
	"github.com/5vnetwork/vx-core/test/mocks"
	"github.com/5vnetwork/vx-core/test/servers/tcp"
	"github.com/5vnetwork/vx-core/test/servers/udp"
	"github.com/5vnetwork/vx-core/tun"
)

func TestMain(m *testing.M) {
	if !test.IsAdmin() {
		return
	}

	// test.InitZeroLog()
	device, err := tun.NewTun(&tun.TunOption{
		Ip4:    netip.MustParsePrefix("172.23.27.1/24"),
		Ip6:    netip.MustParsePrefix("fc23::1/126"),
		Path:   "../../../tun/wintun",
		Name:   "xtun",
		Mtu:    1500,
		Route4: []netip.Prefix{netip.MustParsePrefix("1.0.0.0/24")},
		Route6: []netip.Prefix{netip.MustParsePrefix("2000::/64")},
	})
	if err != nil {
		log.Fatal(err)
	}
	system := New(WithHandler(
		&mocks.LoopbackHandler{}),
		WithTun(device),
		With4(netip.MustParseAddr("172.23.27.2").AsSlice(), netip.MustParseAddr("172.23.27.1").AsSlice(), 0),
		With6(netip.MustParseAddr("fc23::2").AsSlice(), netip.MustParseAddr("fc23::1").AsSlice(), 0),
		WithTag("tun"),
	)
	system.Start()
	defer system.Close()
	exitVal := m.Run()
	os.Exit(exitVal)
}

func TestTunInboundUdp4(t *testing.T) {
	udp.RunPeerAUdp(t, gonet.IPv4(1, 0, 0, 1), 10000)
}

func TestTunInboundUdp6(t *testing.T) {
	udp.RunPeerAUdp(t, gonet.ParseIP("2000::1"), 10000)
}

func TestTunInboundTcp4(t *testing.T) {
	tcp.RunSimpleTcpClient(t, gonet.IPv4(1, 0, 0, 1), 12345)
}

func TestTunInboundTcp6(t *testing.T) {
	tcp.RunSimpleTcpClient(t, gonet.ParseIP("2000::1"), 12345)
}

func TestTunInboundMultiTcp(t *testing.T) {
	t.Run("group", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			t.Run("ipv4 "+strings.ToString(i), func(t *testing.T) {
				t.Parallel()
				tcp.RunSimpleTcpClient(t, gonet.IPv4(1, 0, 0, byte(i)), 12345)
			})
			t.Run("ipv6 "+strings.ToString(i), func(t *testing.T) {
				t.Parallel()
				tcp.RunSimpleTcpClient(t, gonet.ParseIP("2000::"+strings.ToString(i)), 23456)
			})
		}
	})
}

func TestTunInboundMultiUdp(t *testing.T) {
	t.Run("group", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			t.Run("ipv4 "+strings.ToString(i), func(t *testing.T) {
				t.Parallel()
				udp.RunPeerAUdp(t, gonet.IPv4(1, 0, 0, byte(i)), 10000)
			})
			t.Run("ipv6 "+strings.ToString(i), func(t *testing.T) {
				t.Parallel()
				udp.RunPeerAUdp(t, gonet.ParseIP("2000::"+strings.ToString(i)), 10000)
			})
		}
	})

}

func TestTunInboundMultiUdp443(t *testing.T) {
	t.Run("group", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			t.Run("ipv4 "+strings.ToString(i), func(t *testing.T) {
				t.Parallel()
				udp.RunPeerAUdp(t, gonet.IPv4(1, 0, 0, byte(i)), 443)
			})
			t.Run("ipv6 "+strings.ToString(i), func(t *testing.T) {
				t.Parallel()
				udp.RunPeerAUdp(t, gonet.ParseIP("2000::"+strings.ToString(i)), 443)
			})
		}
	})

}

func TestTunInboundUdpPacketConn(t *testing.T) {
	t.Run("group", func(t *testing.T) {
		var ipv4Dsts []*net.UDPAddr
		var ipv6Dsts []*net.UDPAddr
		for i := 0; i < 10; i++ {
			ipv4Dsts = append(ipv4Dsts, &net.UDPAddr{IP: gonet.IPv4(1, 0, 0, byte(i)), Port: 10000 + i})
			ipv6Dsts = append(ipv6Dsts, &net.UDPAddr{IP: gonet.ParseIP("2000::" + strings.ToString(i)), Port: 10000 + i})
		}
		t.Run("ipv4 ", func(t *testing.T) {
			t.Parallel()
			udp.RunPeerAUdps(t, ipv4Dsts...)
		})
		t.Run("ipv6 ", func(t *testing.T) {
			t.Parallel()
			udp.RunPeerAUdps(t, ipv6Dsts...)
		})
	})

}
