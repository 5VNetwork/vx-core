//go:build windows
// +build windows

package gvisor

import (
	"bytes"
	"crypto/rand"
	"net"
	"net/netip"
	"testing"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/test"
	"github.com/5vnetwork/vx-core/test/mocks"
	"github.com/5vnetwork/vx-core/tun"

	"gvisor.dev/gvisor/pkg/tcpip/link/channel"
)

func TestGvisorInit(t *testing.T) {
	opt := GvisorInboundOption{
		Handler:      &mocks.LoopbackHandler{},
		LinkEndpoint: channel.New(1024, 1500, ""),
	}
	_, err := NewGvisorInbound(&opt)
	if err != nil {
		t.Errorf("Failed to init gvisor: %v", err)
	}
}

var tunOption = &tun.TunOption{
	Mtu:  1500,
	Name: "test",
	Ip4:  netip.MustParsePrefix("172.13.17.1/24"),
	Route4: []netip.Prefix{
		netip.MustParsePrefix("1.2.3.4/32"),
	},
	Path: "../../../tun/wintun",
}

func TestGvisorTCPConn(t *testing.T) {
	if !test.IsAdmin() {
		t.Skip("Skipping test because it requires admin privileges")
	}

	// test.InitZeroLog()
	tn, err := tun.NewTun(tunOption)
	if err != nil {
		t.Errorf("Failed to create tun: %v", err)
	}
	opt := GvisorInboundOption{
		TcpOnly:      true,
		Handler:      &mocks.LoopbackHandler{},
		LinkEndpoint: NewTunLinkEndpoint(tn, 1500),
	}
	gi, err := NewGvisorInbound(&opt)
	if err != nil {
		t.Errorf("Failed to init gvisor: %v", err)
	}
	common.Must(gi.Start())
	defer gi.Close()

	conn, err := net.Dial("tcp", "1.2.3.4:80")
	if err != nil {
		t.Errorf("Failed to dial: %v", err)
	}
	defer conn.Close()

	b := make([]byte, 1024)
	readBuf := make([]byte, 1024)
	for i := 0; i < 10; i++ {
		// randomize b
		rand.Read(b)
		_, err := conn.Write(b)
		if err != nil {
			t.Errorf("Failed to write: %v", err)
		}
		_, err = conn.Read(readBuf)
		if err != nil {
			t.Errorf("Failed to read: %v", err)
		}
		if !bytes.Equal(b, readBuf) {
			t.Errorf("Read data is not equal to written data")
		}
	}
}

var tunOption1 = &tun.TunOption{
	Mtu:  1500,
	Name: "test1",
	Ip4:  netip.MustParsePrefix("172.13.17.11/24"),
	Route4: []netip.Prefix{
		netip.MustParsePrefix("1.2.3.4/32"),
	},
	Path: "../../../tun/wintun",
}

func TestGvisorUdp(t *testing.T) {
	if !test.IsAdmin() {
		t.Skip("Skipping test because it requires admin privileges")
	}

	// test.InitZeroLog()
	tn, err := tun.NewTun(tunOption1)
	if err != nil {
		t.Errorf("Failed to create tun: %v", err)
	}
	opt := GvisorInboundOption{
		UdpOnly:      true,
		Handler:      &mocks.LoopbackHandler{},
		LinkEndpoint: NewTunLinkEndpoint(tn, 1500),
	}
	gi, err := NewGvisorInbound(&opt)
	if err != nil {
		t.Errorf("Failed to init gvisor: %v", err)
	}
	common.Must(gi.Start())
	defer gi.Close()

	conn, err := net.Dial("udp", "1.2.3.4:80")
	if err != nil {
		t.Errorf("Failed to dial: %v", err)
	}

	b := make([]byte, 1024)
	readBuf := make([]byte, 1024)
	for i := 0; i < 10; i++ {
		// randomize b
		rand.Read(b)
		_, err := conn.Write(b)
		if err != nil {
			t.Errorf("Failed to write: %v", err)
		}
		_, err = conn.Read(readBuf)
		if err != nil {
			t.Errorf("Failed to read: %v", err)
		}
		if !bytes.Equal(b, readBuf) {
			t.Errorf("Read data is not equal to written data")
		}
	}
}
