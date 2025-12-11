package vi

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/session"
	"github.com/5vnetwork/vx-core/common/uuid"
	"github.com/5vnetwork/vx-core/common/vision"
	"github.com/5vnetwork/vx-core/i"
)

type User struct {
	Uid    string
	Secret string
}

type Server struct {
	policy i.TimeoutSetting
	users  sync.Map //key: uuid.UUID secret, value: string uid
	d      i.Handler
}

// New creates a new VLess inbound handler.
func New() *Server {
	handler := &Server{}

	return handler
}
func (s *Server) WithHandler(h i.Handler) {
	s.d = h
}

func (s *Server) WithTimeoutPolicy(policy i.TimeoutSetting) {
	s.policy = policy
}

// Network implements proxy.Inbound.Network().
func (*Server) Network() []net.Network {
	return []net.Network{net.Network_TCP, net.Network_UNIX}
}

func (h *Server) AddUser(user i.User) {
	h.users.Store(user.Secret(), user.Uid())
}

func (h *Server) RemoveUser(uuid uuid.UUID) {
	h.users.Delete(uuid)
}

// Inspect first bytes to determine if it is a vless request. If it is, read a complete header, decode it and return it.
// If it is not, return with [isFallback] set to true.
// [first] is what has been read from the [conn] during inspecting. It is non-nil when header is nil. It is nil
// when header is not nil
// return an err if
// TODO: change [first] to a slice of bytes
func tryDecodeHeader(conn net.Conn, um *sync.Map) (first []byte, request *Header, err error) {
	first = make([]byte, 0, 2048)

	n, err := conn.Read(first[:2])
	first = first[:n]
	if err != nil {
		return nil, nil, errors.New("failed to read first two bytes").Base(err)
	}
	if n < 2 {
		return first, nil, nil
	}
	len := uint16(net.PortFromBytes(first))

	if len < 18 || len > 1460 {
		return first, nil, nil
	}

	n, err = conn.Read(first[2 : int32(len)+2])
	first = first[:n+2]
	if err != nil {
		return nil, nil, errors.New("failed to read from conn").Base(err)
	}
	if n != int(len) {
		return first, nil, nil
	}

	header, err := DecodeHeader(buf.FromBytes(first[:int32(len)+2]), um)
	if err != nil {
		if errors.Is(err, InvalidUser) || errors.Is(err, InvalidVersion) {
			return first, nil, nil
		} else {
			return nil, nil, err
		}
	} else {
		return nil, header, err
	}
}

// Process implements proxy.Inbound.Process().
func (h *Server) Process(ctx context.Context, conn net.Conn) error {
	if err := conn.SetReadDeadline(time.Now().Add(h.policy.HandshakeTimeout())); err != nil {
		return errors.New("unable to set handshake deadline").Base(err)
	}

	_, header, err := tryDecodeHeader(conn, &h.users)
	if err != nil {
		return err
	}

	if err := conn.SetReadDeadline(time.Time{}); err != nil {
		return fmt.Errorf("unable to clear read deadline: %w", err)
	}
	// log.PrintDebug(sid, "received request for ", header.Dest)

	info := session.InfoFromContext(ctx)
	info.Target = header.Dest
	info.User = header.User
	info.UdpUuid = header.UdpUuid
	// info.SpliceCopy.Whether = session.SpliceCopyNotSureYet
	// if runtime.GOOS != "linux" && runtime.GOOS != "android" {
	// 	info.SpliceCopy.Whether = session.SpliceCopyNo
	// }

	vConn := vision.NewVisionConn(ctx, conn, false, 0)
	clientReader := buf.NewReader(vConn)
	clientWriter := buf.NewWriter(vConn)
	if header.Dest.Network == net.Network_UDP {
		// var cancelcause context.CancelCauseFunc
		// ctx, cancelcause = context.WithCancelCause(ctx)
		// clientReader = buf.NewIdleReader(buf.NewLengthPacketReader(vConn), cancelcause, h.policy.ConnectionIdleTimeout())
		//TODO
		clientReader = buf.NewLengthPacketReader(vConn)
		clientWriter = buf.NewMultiLengthPacketWriter(clientWriter)
	}

	// if err = helper.Relay(ctx, clientReader, clientWriter, ilink, ilink); err != nil {
	// 	ilink.Interrupt()
	// 	return fmt.Errorf("failed to relay: %w", err)
	// }
	if err := h.d.HandleFlow(ctx, header.Dest, buf.NewRWD(clientReader, clientWriter, conn)); err != nil {
		return fmt.Errorf("failed to dispatch reader writer: %w", err)
	}
	return nil
}
