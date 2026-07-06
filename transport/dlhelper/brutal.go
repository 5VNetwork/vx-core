package dlhelper

import (
	"net"

	"github.com/rs/zerolog/log"
)

type brutalListener struct {
	net.Listener
	sendBPS uint64
}

func (l *brutalListener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	if err := applyBrutalConn(conn, l.sendBPS); err != nil {
		log.Err(err).Msg("failed to apply brutal options")
	}
	return conn, nil
}

func maybeApplyBrutalConn(conn net.Conn, config *SocketSetting) error {
	if config == nil {
		return nil
	}
	return applyBrutalConn(conn, config.GetTcpBrutalSendRate())
}

func applyBrutalConn(conn net.Conn, rate uint64) error {
	if rate == 0 {
		return nil
	}
	if _, ok := conn.(*net.TCPConn); !ok {
		return nil
	}
	return SetBrutalOptions(conn, rate)
}

func wrapBrutalListener(l net.Listener, rate uint64) net.Listener {
	if rate == 0 {
		return l
	}
	return &brutalListener{Listener: l, sendBPS: rate}
}
