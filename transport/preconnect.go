package transport

import (
	"context"
	"sync"
	"time"

	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/i"
	"github.com/rs/zerolog/log"
)

type PreConnectDialer struct {
	i.Dialer
	addr         net.Address
	PortSelector i.PortSelector

	ctx context.Context
	// Connection pool
	mu                sync.RWMutex
	closed            bool
	pool              map[net.Conn]*time.Timer
	poolSize          int32 // current pool size
	targetSize        int32 // target pool size
	lastEmptyPoolTime time.Time
}

// NewPreConnectDialer creates a new PreConnectDialer with connection pooling
func NewPreConnectDialer(ctx context.Context, dialer i.Dialer, addr net.Address,
	portSelector i.PortSelector) *PreConnectDialer {
	d := &PreConnectDialer{
		ctx:          ctx,
		Dialer:       dialer,
		addr:         addr,
		PortSelector: portSelector,
		pool:         make(map[net.Conn]*time.Timer), // buffer for up to 100 connections
	}
	return d
}

// fillPool fills the pool to target size
func (d *PreConnectDialer) fillPool() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed {
		return
	}

	log.Debug().Int32("poolSize", d.poolSize).Int32("targetSize", d.targetSize).Msg("fillPool")

	if !d.lastEmptyPoolTime.IsZero() &&
		time.Since(d.lastEmptyPoolTime) > time.Second*10 {
		d.targetSize--
		d.lastEmptyPoolTime = time.Time{}
	}

	needed := d.targetSize - d.poolSize
	if needed <= 0 {
		return
	}
	d.poolSize += needed
	// Create connections asynchronously
	for i := 0; i < int(needed); i++ {
		go d.createConnection()
	}
}

// createConnection creates a new connection and adds it to the pool
func (d *PreConnectDialer) createConnection() {
	dst := net.Destination{
		Address: d.addr,
		Port:    net.Port(d.PortSelector.SelectPort()),
		Network: net.Network_TCP,
	}

	conn, err := d.Dialer.Dial(d.ctx, dst)
	d.mu.Lock()
	defer d.mu.Unlock()
	if err != nil {
		log.Debug().Err(err).Msg("preconnect: failed to create pre-connection")
		d.poolSize--
		return
	}

	d.pool[conn] = time.AfterFunc(
		time.Second*5,
		func() {
			d.mu.Lock()
			_, ok := d.pool[conn]
			if ok {
				log.Debug().Msg("preconnect: connection not used")
				delete(d.pool, conn)
				d.poolSize--
				d.targetSize--
			}
			d.mu.Unlock()
			if ok {
				conn.Close()
			}
		},
	)
}

// isConnAlive checks if a connection is still alive using a non-intrusive method
// func isConnAlive(conn net.Conn) bool {
// 	if err := conn.SetReadDeadline(time.Now().Add(1 * time.Microsecond)); err != nil {
// 		log.Debug().Err(err).Msg("failed to set read deadline")
// 		return false
// 	}
// 	_, err := conn.Read([]byte{0})
// 	if err1 := conn.SetReadDeadline(time.Time{}); err1 != nil {
// 		log.Debug().Err(err1).Msg("failed to clear read deadline")
// 		return false
// 	}
// 	log.Debug().Err(err).Msg("preconnect read error")
// 	return errors.Is(err, os.ErrDeadlineExceeded)
// }

// Dial returns a pre-established connection from the pool, or creates a new one if pool is empty
func (d *PreConnectDialer) Dial(ctx context.Context, dst net.Destination) (net.Conn, error) {
	if dst.Network == net.Network_UDP {
		return d.Dialer.Dial(ctx, dst)
	}

	d.mu.Lock()
	var conn net.Conn
	for c, timer := range d.pool {
		timer.Stop()
		conn = c
		delete(d.pool, c)
		d.poolSize--
	}
	if conn == nil {
		log.Ctx(ctx).Debug().Msg("preconnect pool is empty")
		d.targetSize++
		d.lastEmptyPoolTime = time.Now()
	}
	d.mu.Unlock()

	d.fillPool()

	if conn != nil {
		log.Ctx(ctx).Debug().Msg("preconnect: use existing connection")
		return conn, nil
		// if isConnAlive(conn) {
		// 	log.Ctx(ctx).Debug().Msg("preconnect: connection is still alive")
		// 	return conn, nil
		// } else {
		// 	log.Ctx(ctx).Debug().Msg("preconnect: connection is not alive")
		// 	conn.Close()
		// }
	}

	return d.Dialer.Dial(ctx, dst)
}
