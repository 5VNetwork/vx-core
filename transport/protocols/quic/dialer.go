package quic

import (
	"context"
	"crypto/tls"
	"sync"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/rs/zerolog/log"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/task"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/transport/security"
	mytls "github.com/5vnetwork/vx-core/transport/security/tls"
)

type connectionContext struct {
	rawConn *sysConn
	conn    *quic.Conn
}

var errConnectionClosed = errors.New("connection closed")

func (c *connectionContext) openStream(destAddr net.Addr) (*interConn, error) {
	if !isActive(c.conn) {
		return nil, errConnectionClosed
	}

	stream, err := c.conn.OpenStream()
	if err != nil {
		return nil, err
	}

	conn := &interConn{
		stream: stream,
		local:  c.conn.LocalAddr(),
		remote: destAddr,
	}

	return conn, nil
}

type clientConnections struct {
	access  sync.Mutex
	conns   map[net.Destination][]*connectionContext
	cleanup *task.Periodic
}

func isActive(s *quic.Conn) bool {
	select {
	case <-s.Context().Done():
		return false
	default:
		return true
	}
}

func removeInactiveConnections(conns []*connectionContext) []*connectionContext {
	activeConnections := make([]*connectionContext, 0, len(conns))
	for _, s := range conns {
		if isActive(s.conn) {
			activeConnections = append(activeConnections, s)
			continue
		}
		if err := s.conn.CloseWithError(0, ""); err != nil {
			log.Err(err).Msg("failed to close connection")
		}
		if err := s.rawConn.Close(); err != nil {
			log.Err(err).Msg("failed to close raw connection")
		}
	}

	if len(activeConnections) < len(conns) {
		return activeConnections
	}

	return conns
}

func openStream(conns []*connectionContext, destAddr net.Addr) *interConn {
	for _, s := range conns {
		if !isActive(s.conn) {
			continue
		}

		conn, err := s.openStream(destAddr)
		if err != nil {
			continue
		}

		return conn
	}

	return nil
}

func (s *clientConnections) cleanConnections() error {
	s.access.Lock()
	defer s.access.Unlock()

	if len(s.conns) == 0 {
		return nil
	}

	newConnMap := make(map[net.Destination][]*connectionContext)

	for dest, conns := range s.conns {
		conns = removeInactiveConnections(conns)
		if len(conns) > 0 {
			newConnMap[dest] = conns
		}
	}

	s.conns = newConnMap
	return nil
}

func (s *clientConnections) openConnection(destAddr net.Addr, config *QuicConfig, tlsConfig *tls.Config, sockopt i.PacketListener) (net.Conn, error) {
	s.access.Lock()
	defer s.access.Unlock()

	if s.conns == nil {
		s.conns = make(map[net.Destination][]*connectionContext)
	}

	dest := net.DestinationFromAddr(destAddr)

	var conns []*connectionContext
	if s, found := s.conns[dest]; found {
		conns = s
	}

	{
		conn := openStream(conns, destAddr)
		if conn != nil {
			return conn, nil
		}
	}

	conns = removeInactiveConnections(conns)

	// errors.New("dialing QUIC to ", dest).WriteToLog()

	rawConn, err := sockopt.ListenPacket(context.Background(), destAddr.Network(), "")
	if err != nil {
		return nil, err
	}

	quicConfig := &quic.Config{
		HandshakeIdleTimeout: time.Second * 8,
		MaxIdleTimeout:       time.Second * 30,
		KeepAlivePeriod:      time.Second * 15,
	}

	sysConn, err := wrapSysConn(rawConn.(*net.UDPConn), config)
	if err != nil {
		rawConn.Close()
		return nil, err
	}

	tr := quic.Transport{
		Conn:               sysConn,
		ConnectionIDLength: 12,
	}

	conn, err := tr.Dial(context.Background(), destAddr, tlsConfig, quicConfig)
	if err != nil {
		sysConn.Close()
		return nil, err
	}

	context := &connectionContext{
		conn:    conn,
		rawConn: sysConn,
	}
	s.conns[dest] = append(conns, context)
	return context.openStream(destAddr)
}

var client clientConnections

func init() {
	client.conns = make(map[net.Destination][]*connectionContext)
	client.cleanup = &task.Periodic{
		Interval: time.Minute,
		Execute:  client.cleanConnections,
	}
	common.Must(client.cleanup.Start())
}

type quicDialer struct {
	config       *QuicConfig
	engine       security.Engine
	socketConfig i.PacketListener
}

func NewQuicDialer(config *QuicConfig, engine security.Engine, socketConfig i.PacketListener) *quicDialer {
	return &quicDialer{
		config:       config,
		engine:       engine,
		socketConfig: socketConfig,
	}
}

func (d *quicDialer) Dial(ctx context.Context, dest net.Destination) (net.Conn, error) {
	return Dial(ctx, dest, d.config, d.engine, d.socketConfig)
}

func Dial(ctx context.Context, dest net.Destination, c *QuicConfig, e security.Engine, socketConfig i.PacketListener) (net.Conn, error) {
	var tlsConfig *tls.Config
	var err error
	if e == nil {
		c := &mytls.TlsConfig{
			ServerName:    internalDomain,
			AllowInsecure: true,
		}
		tlsConfig, err = c.GetTLSConfig(mytls.WithDestination(dest))
		if err != nil {
			return nil, err
		}
	} else {
		tlsConfig = e.GetTLSConfig(security.OptionWithDestination{Dest: dest})
	}

	var destAddr *net.UDPAddr
	if dest.Address.Family().IsIP() {
		destAddr = &net.UDPAddr{
			IP:   dest.Address.IP(),
			Port: int(dest.Port),
		}
	} else {
		addr, err := net.ResolveUDPAddr("udp", dest.NetAddr())
		if err != nil {
			return nil, err
		}
		destAddr = addr
	}

	return client.openConnection(destAddr, c, tlsConfig, socketConfig)
}
