package congestion

import (
	"github.com/5vnetwork/vx-core/proxy/hysteria2/internal/internal/congestion/bbr"
	"github.com/5vnetwork/vx-core/proxy/hysteria2/internal/internal/congestion/brutal"
	"github.com/apernet/quic-go"
)

func UseBBR(conn *quic.Conn) {
	conn.SetCongestionControl(bbr.NewBbrSender(
		bbr.DefaultClock{},
		bbr.GetInitialPacketSize(conn.RemoteAddr()),
	))
}

func UseBrutal(conn *quic.Conn, tx uint64) {
	conn.SetCongestionControl(brutal.NewBrutalSender(tx))
}
