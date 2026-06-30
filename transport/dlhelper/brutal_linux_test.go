//go:build linux

package dlhelper_test

import (
	"context"
	"net"
	"testing"

	"github.com/5vnetwork/vx-core/common"
	net1 "github.com/5vnetwork/vx-core/common/net"
	. "github.com/5vnetwork/vx-core/transport/dlhelper"
)

func TestBrutalOnEstablishedConn(t *testing.T) {
	const rate = 10 * 1024 * 1024

	ln, err := (&DefaultListener{}).Listen(context.Background(), &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1)}, &SocketSetting{
		TcpBrutalSendRate: rate,
	})
	common.Must(err)
	defer ln.Close()

	dest := net1.TCPDestination(net1.LocalHostIP, net1.Port(ln.Addr().(*net.TCPAddr).Port))

	go func() {
		conn, err := ln.Accept()
		common.Must(err)
		defer conn.Close()
	}()

	conn, err := (&DefaultSystemDialer{}).DialConn(context.Background(), dest, &SocketSetting{
		TcpBrutalSendRate: rate,
	})
	common.Must(err)
	defer conn.Close()
}
