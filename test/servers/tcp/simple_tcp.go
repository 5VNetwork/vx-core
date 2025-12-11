package tcp

import (
	"crypto/rand"
	"testing"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/net"

	"github.com/google/go-cmp/cmp"
)

func RunSimpleTcpClient(t *testing.T, ip net.IP, port uint16) {
	conn, err := net.DialTCP("tcp", nil, &net.TCPAddr{IP: ip, Port: int(port)})
	common.Must(err)
	buf := make([]byte, 2048)
	received := make([]byte, 2048)
	defer conn.Close()

	for i := 0; i < 100; i++ {
		rand.Read(buf[:200])
		_, err := conn.Write(buf[:200])
		if err != nil {
			t.Fatal("PeerA cannotd write", err)
		}
		n, err := conn.Read(received)
		if err != nil || n != 200 {
			t.Fatalf("PeerA should get 200 bytes, but got %v", buf[:n])
		}
		if cmp.Diff(buf[:200], received[:n]) != "" {
			t.Fatalf("PeerA expected %v but got %v", buf[:200], received[:n])
		}
	}
}

func RunSimpleTcpServer(t *testing.T) net.Listener {
	l, err := net.ListenTCP("tcp", &net.TCPAddr{
		Port: 10000,
	})
	common.Must(err)
	t.Log("PeerB listening at", l.Addr())

	go func() {
		c, err := l.Accept()
		common.Must(err)
		t.Log("PeerB got new connection. src:", c.RemoteAddr())
		defer c.Close()
		for {
			buf := make([]byte, 2048)
			for {
				n, err := c.Read(buf)
				if err != nil {
					t.Log("PeerB cannot read", err)
					return
				}
				_, err = c.Write(buf[:n])
				if err != nil {
					t.Log("PeerB cannot write", err)
					return
				}
			}
		}
	}()

	return l
}
