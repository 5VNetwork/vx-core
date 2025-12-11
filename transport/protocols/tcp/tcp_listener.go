package tcp

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	mynet "github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/i"

	"github.com/rs/zerolog/log"
)

// implement net.Listener
type Listener struct {
	net.Listener
	h func(net.Conn)
}

func Listen(ctx context.Context, addr mynet.Destination,
	config *TcpConfig, listener i.Listener, h func(net.Conn)) (*Listener, error) {

	netListener, err := listener.Listen(ctx, addr.Addr())
	if err != nil {
		return nil, fmt.Errorf("cannot create a tcp.Listener, %w", err)
	}

	l := &Listener{
		h:        h,
		Listener: netListener,
	}

	go l.keepAccept()
	return l, nil
}

func (l *Listener) keepAccept() {
	for {
		conn, err := l.Listener.Accept()
		if err != nil {
			errStr := err.Error()
			if strings.Contains(errStr, "too many") {
				log.Warn().Err(err).Msg("too many connections")
				time.Sleep(time.Millisecond * 500)
				continue
			}
			if strings.Contains(errStr, "use of closed network connection") {
				return
			}
			log.Error().Err(err).Msg("failed to accepted")
			return
		}

		l.h(conn)
	}
}
