//go:build server

package trojan

import (
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
	h.validator.Add(NewMemoryAccount(user.Uid(), user.Secret()))
}

func (h *Server) RemoveUser(uid, secret string) {
	h.validator.Del(uid)
}

func (h *Server) WithOnUnauthorizedRequest(f i.UnauthorizedReport) {
	h.OnUnauthorizedRequest = f
}

// Network implements proxy.Inbound.Network().
func (s *Server) Network() []net.Network {
	return []net.Network{net.Network_TCP, net.Network_UNIX}
}

func (s *Server) FallbackProcess(ctx context.Context, conn net.Conn) (bool, buf.MultiBuffer, error) {
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
	destination, err := ParseHeader(conn)
	if err != nil {
		return err
	}

	if s.Vision {
		conn = vision.NewVisionConn(ctx, conn, false, 0)
	}

	ctx = proxy.ContextWithUser(ctx, user.Uid)
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
