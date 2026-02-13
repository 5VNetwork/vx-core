//go:build server

package trojan

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/common/vision"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/proxy"
)

// Server is an inbound connection handler that handles messages in trojan protocol.
type Server struct {
	ServerSettings
	validator Validator
}

type ServerSettings struct {
	PolicyManager         i.TimeoutSetting
	Handler               i.Handler
	OnUnauthorizedRequest i.UnauthorizedReport
	Vision                bool
}

// NewServer creates a new trojan inbound handler.
func NewServer(settings ServerSettings) *Server {
	server := &Server{
		ServerSettings: settings,
	}
	return server
}

func (h *Server) AddUser(user i.User) {
	h.validator.Add(NewMemoryAccount(user))
}

func (h *Server) RemoveUser(user i.User) {
	h.validator.Del(NewMemoryAccount(user))
}

func (h *Server) WithOnUnauthorizedRequest(f i.UnauthorizedReport) {
	h.OnUnauthorizedRequest = f
}

// Network implements proxy.Inbound.Network().
func (s *Server) Network() []net.Network {
	return []net.Network{net.Network_TCP, net.Network_UNIX}
}

func (s *Server) FallbackProcess(ctx context.Context, conn net.Conn) (bool, buf.MultiBuffer, error) {
	ctx = proxy.ContextWithInboundProxyProtocol(ctx, "trojan")

	if err := conn.SetReadDeadline(time.Now().Add(s.PolicyManager.HandshakeTimeout())); err != nil {
		return false, nil, errors.New("unable to set read deadline").Base(err)
	}
	cacheReader := buf.NewMemoryReader(conn)

	user, err := s.auth(ctx, cacheReader)
	if err1 := conn.SetReadDeadline(time.Time{}); err1 != nil {
		return false, nil, errors.New("unable to set read deadline").Base(err1)
	}
	if err != nil {
		return true, cacheReader.History(), err
	}
	cacheReader.StopMemorize()

	return false, nil, s.processCommon(ctx, conn, user)
}

// Process implements proxy.Inbound.Process().
func (s *Server) Process(ctx context.Context, conn net.Conn) error {
	ctx = proxy.ContextWithInboundProxyProtocol(ctx, "trojan")

	if err := conn.SetReadDeadline(time.Now().Add(s.PolicyManager.HandshakeTimeout())); err != nil {
		return errors.New("unable to set read deadline").Base(err)
	}

	user, err := s.auth(ctx, conn)
	if err1 := conn.SetReadDeadline(time.Time{}); err1 != nil {
		return errors.New("unable to set read deadline").Base(err1)
	}
	if err != nil {
		if s.OnUnauthorizedRequest != nil {
			s.OnUnauthorizedRequest.ReportUnauthorized(conn.RemoteAddr().String(), "")
		}
		return err
	}

	return s.processCommon(ctx, conn, user)
}

func (s *Server) auth(ctx context.Context, reader io.Reader) (*MemoryAccount, error) {
	var b [58]byte
	firstLen, err := reader.Read(b[:])
	if err != nil {
		return nil, errors.New("failed to read first request").Base(err)
	}

	var user *MemoryAccount
	if firstLen < 58 || b[56] != '\r' {
		return nil, errors.New("not trojan protocol")
	} else {
		u := s.validator.Get(hexString(b[:56]))
		if u == nil {
			return nil, errors.New("not a valid user")
		}
		user = u
	}

	return user, nil
}

func (s *Server) processCommon(ctx context.Context, conn net.Conn,
	user *MemoryAccount) error {
	destination, useVision, err := s.ParseHeader(conn)
	if err != nil {
		return err
	}

	if useVision {
		conn = vision.NewVisionConn(ctx, conn, false, 0)
	}

	ctx = proxy.ContextWithUser(ctx, user.User)
	if destination.Network == net.Network_UDP { // handle udp request
		return s.handleUDPPayload(ctx,
			&PacketReader{
				reader: &buf.BufferedReader{Reader: buf.NewReader(conn)},
				client: false},
			&PacketWriter{writer: conn}, s.Handler)
	}

	if err := s.Handler.HandleFlow(ctx, destination,
		buf.NewRWD(buf.NewReader(conn), buf.NewWriter(conn), conn)); err != nil {
		return fmt.Errorf("failed to dispatch: %w", err)
	}
	return nil
}

// ParseHeader parses the trojan protocol header
func (s *Server) ParseHeader(reader io.Reader) (dst net.Destination, vision bool, err error) {
	var crlfBuf [2]byte
	var command [1]byte
	// var hash [56]byte
	// if _, err := io.ReadFull(c.Reader, hash[:]); err != nil {
	// 	return errors.New("failed to read user hash").Base(err)
	// }

	// if _, err := io.ReadFull(c.Reader, crlf[:]); err != nil {
	// 	return errors.New("failed to read crlf").Base(err)
	// }

	if _, err := io.ReadFull(reader, command[:]); err != nil {
		return dst, vision, errors.New("failed to read command").Base(err)
	}
	network := net.Network_TCP
	if command[0] == commandUDP {
		network = net.Network_UDP
	}

	addr, port, err := addrParser.ReadAddressPort(nil, reader)
	if err != nil {
		return dst, vision, errors.New("failed to read address and port").Base(err)
	}
	dst = net.Destination{Network: network, Address: addr, Port: port}

	if _, err := io.ReadFull(reader, crlfBuf[:]); err != nil {
		return dst, vision, errors.New("failed to read crlf").Base(err)
	}

	if s.Vision && bytes.Equal(crlfBuf[:], visionCrlf) {
		return dst, true, nil
	} else {
		return dst, false, nil
	}
}

func (s *Server) handleUDPPayload(ctx context.Context, clientReader *PacketReader,
	clientWriter *PacketWriter, d i.Handler) error {

	if err := d.HandlePacketConn(ctx, net.AnyUdpDest,
		&udp.PacketRW{
			PacketReader: clientReader,
			PacketWriter: clientWriter,
		}); err != nil {
		return fmt.Errorf("failed to dispatch UDP, %w", err)
	}
	return nil
}
