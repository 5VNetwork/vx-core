//go:build test

package realmtest

import (
	"net"
	"net/netip"
	"testing"

	"github.com/pion/stun/v3"
	"github.com/stretchr/testify/require"
)

type reflectiveSTUN struct {
	conn net.PacketConn
}

func StartReflectiveSTUN(t *testing.T) (addr string, cleanup func()) {
	t.Helper()
	conn, err := net.ListenPacket("udp4", "127.0.0.1:0")
	require.NoError(t, err)
	s := &reflectiveSTUN{conn: conn}
	go s.serve()
	return conn.LocalAddr().String(), func() { _ = conn.Close() }
}

func (s *reflectiveSTUN) serve() {
	buf := make([]byte, 1500)
	for {
		n, addr, err := s.conn.ReadFrom(buf)
		if err != nil {
			return
		}
		msg := stun.New()
		if err := stun.Decode(buf[:n], msg); err != nil || msg.Type != stun.BindingRequest {
			continue
		}
		host, portStr, err := net.SplitHostPort(addr.String())
		if err != nil {
			continue
		}
		port, err := netip.ParseAddrPort(net.JoinHostPort(host, portStr))
		if err != nil {
			continue
		}
		packet, err := buildSTUNBindingResponse(msg.TransactionID, port)
		if err != nil {
			continue
		}
		_, _ = s.conn.WriteTo(packet, addr)
	}
}

func buildSTUNBindingResponse(txID [stun.TransactionIDSize]byte, xorMapped netip.AddrPort) ([]byte, error) {
	ip := xorMapped.Addr().AsSlice()
	if xorMapped.Addr().Is4() {
		v4 := xorMapped.Addr().As4()
		ip = net.IPv4(v4[0], v4[1], v4[2], v4[3])
	}
	msg, err := stun.Build(
		stun.BindingSuccess,
		stun.NewTransactionIDSetter(txID),
		&stun.XORMappedAddress{
			IP:   ip,
			Port: int(xorMapped.Port()),
		},
	)
	if err != nil {
		return nil, err
	}
	return msg.Raw, nil
}
