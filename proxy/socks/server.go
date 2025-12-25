package socks

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	pc "github.com/5vnetwork/vx-core/app/configs/proxy"
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/errors"
	nethelper "github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/common/protocol"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/proxy"

	"github.com/rs/zerolog/log"
)

// Server is a SOCKS 5 proxy server
type Server struct {
	address    nethelper.Address
	udpEnabled bool
	authType   pc.AuthType
	policy     i.TimeoutSetting

	usersLock sync.RWMutex
	users     map[string]string // username to password. username is uuid in string format, password is uuid in string format
	handler   i.Handler
}

// NewServer creates a new Server object.
func NewServer(config *SocksServerConfig) *Server {
	s := &Server{
		udpEnabled: config.UdpEnabled,
		authType:   config.AuthType,
		policy:     config.Policy,
		handler:    config.Handler,
		address:    config.Address,
		users:      make(map[string]string),
	}
	return s
}

type SocksServerConfig struct {
	Address    nethelper.Address
	UdpEnabled bool
	AuthType   pc.AuthType
	Policy     i.TimeoutSetting
	Handler    i.Handler
}

func (s *Server) AddUser(user i.User) {
	s.usersLock.Lock()
	defer s.usersLock.Unlock()
	s.users[user.Uid()] = user.Secret()
}

func (s *Server) RemoveUser(uid, secret string) {
	s.usersLock.Lock()
	defer s.usersLock.Unlock()
	delete(s.users, uid)
}

func (s *Server) Network() []nethelper.Network {
	list := []nethelper.Network{nethelper.Network_TCP}
	if s.udpEnabled {
		list = append(list, nethelper.Network_UDP)
	}
	return list
}

func (s *Server) FallbackProcess(ctx context.Context, conn net.Conn) (bool, buf.MultiBuffer, error) {
	switch nethelper.NetworkFromAddr(conn.LocalAddr()) {
	case nethelper.Network_TCP:
		return s.processTcp(ctx, conn)
	case nethelper.Network_UDP:
		return false, nil, s.relayUDP(ctx, conn)
	default:
		return true, nil, errors.New("unknown network")
	}
}

func (s *Server) Process(ctx context.Context, conn net.Conn) error {
	switch nethelper.NetworkFromAddr(conn.LocalAddr()) {
	case nethelper.Network_TCP:
		_, b, err := s.processTcp(ctx, conn)
		if b != nil {
			buf.ReleaseMulti(b)
		}
		return err
	case nethelper.Network_UDP:
		return s.relayUDP(ctx, conn)
	default:
		return errors.New("unknown network")
	}
}

// handshake, if request is of TCP, handle it; if the tcp connecrtion is for UDP, then maintain the
// connection until client closes it
func (s *Server) processTcp(ctx context.Context, conn net.Conn) (bool, buf.MultiBuffer, error) {
	if err := conn.SetReadDeadline(time.Now().Add(time.Duration(s.policy.HandshakeTimeout()))); err != nil { //todo
		return false, nil, errors.New("failed to set deadline on tcp Conn")
	}

	gateway := nethelper.DestinationFromAddr(conn.LocalAddr())
	if !gateway.IsValid() {
		return true, nil, errors.New("gateway not specified")
	}

	cacheReader := buf.NewMemoryReader(conn)

	reader := &buf.BufferedReader{Reader: buf.NewReader(cacheReader)}
	socksSession := &ServerSession{
		serverConfig:   s,
		gatewayAddress: gateway.Address,
		gatewayPort:    gateway.Port,
		clientAddress:  nethelper.DestinationFromAddr(conn.RemoteAddr()).Address,
	}
	request, err := socksSession.Handshake(reader, conn) // request is protocol.RequestHeader
	if err1 := conn.SetReadDeadline(time.Time{}); err1 != nil {
		return false, nil, fmt.Errorf("failed to clear deadline on tcp Conn, %w", err)
	}
	if err != nil {
		// if errors.As(err, &errors.AuthError{}) {
		return true, cacheReader.History(), err
		// }
		// buf.ReleaseMulti(cacheReader.Cache())
		// return false, nil, fmt.Errorf("failed to read request during handshake, %w", err)
	}
	cacheReader.StopMemorize()

	if request.Command == protocol.RequestCommandTCP {
		log.Ctx(ctx).Debug().Str("dest", request.Destination().String()).Msg("socks tcp")
		if request.User != "" {
			ctx = proxy.ContextWithUser(ctx, request.User)
		}
		if err = s.handler.HandleFlow(ctx, request.Destination(), buf.NewRWD(reader, buf.NewWriter(conn), conn)); err != nil {
			return false, nil, fmt.Errorf("failed to dispatch reader writer, %w", err)
		}
		return false, nil, nil
	}

	if request.Command == protocol.RequestCommandUDP {
		return false, nil, maintainHandshakeConnectionForUdp(conn)
	}

	return false, nil, errors.New("upsupported command")
}

// TODO when will it close
func maintainHandshakeConnectionForUdp(conn net.Conn) error {
	_, err := io.Copy(buf.DiscardBytes, conn)
	return err
}

func (s *Server) relayUDP(ctx context.Context, conn net.Conn) error {
	udpConn := &UDPConn{
		ctx:    ctx,
		reader: buf.NewReader(conn),
		writer: buf.NewWriter(conn),
	}

	packet, err := udpConn.ReadPacket()
	if err != nil {
		return fmt.Errorf("failed to read UDP packet, %w", err)
	}
	udpConn.firstPacket = packet

	if err := s.handler.HandlePacketConn(ctx, packet.Target, udpConn); err != nil {
		return fmt.Errorf("failed to dispatch UDP, %w", err)
	}
	return nil
}

type UDPConn struct {
	ctx         context.Context
	firstPacket *udp.Packet
	reader      buf.Reader
	writer      buf.Writer
}

func (c *UDPConn) ReadPacket() (*udp.Packet, error) {
	if c.firstPacket != nil {
		packet := c.firstPacket
		c.firstPacket = nil
		return packet, nil
	}
	mb, err := c.reader.ReadMultiBuffer()
	if len(mb) > 0 {
		r, err := DecodeUDPPacket(mb[0])
		if err != nil {
			buf.ReleaseMulti(mb)
			return nil, fmt.Errorf("unable to parse UDP request, %w", err)
		}
		return &udp.Packet{
			Payload: mb[0],
			Target:  r.Destination(),
		}, nil
	}
	return nil, fmt.Errorf("failed to read UDP packet, %w", err)
}

func (c *UDPConn) WritePacket(p *udp.Packet) error {
	defer p.Payload.Release()
	udpMessage, err := EncodeUDPPacketFromAddress(p.Source, p.Payload.Bytes())
	if err != nil {
		return fmt.Errorf("failed to encode UDP packet, %w", err)
	}
	if err = c.writer.WriteMultiBuffer(buf.MultiBuffer{udpMessage}); err != nil {
		return fmt.Errorf("failed to write UDP packet, %w", err)
	} //send the returned udp packet(from server) back to the client
	return nil
}

func (c *UDPConn) CloseWrite() error {
	return nil
}

func (c *UDPConn) Close() error {
	return nil
}
