package server

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/protocol"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/proxy"
	"github.com/5vnetwork/vx-core/proxy/vmess"
	"github.com/5vnetwork/vx-core/proxy/vmess/encoding"

	"github.com/rs/zerolog/log"
)

type ServerIO struct {
	policyManager         i.TimeoutSetting
	users                 *vmess.TimedUserValidator
	sessionHistory        *encoding.SessionHistory
	secure                bool
	d                     i.ConnHandler
	onUnauthorizedRequest i.UnauthorizedReport
}

type ServerIOSetting struct {
	TimeoutSetting        i.TimeoutSetting
	Security              bool
	OnUnauthorizedRequest i.UnauthorizedReport
	Handler               i.ConnHandler
}

// New creates a new VMess inbound handler.
func NewIO(setting *ServerIOSetting) *ServerIO {
	handler := &ServerIO{
		users:                 vmess.NewTimedUserValidator(protocol.DefaultIDHash),
		sessionHistory:        encoding.NewSessionHistory(),
		policyManager:         setting.TimeoutSetting,
		secure:                setting.Security,
		d:                     setting.Handler,
		onUnauthorizedRequest: setting.OnUnauthorizedRequest,
	}

	return handler
}

func (h *ServerIO) Start() error {
	return nil
}

// Close implements common.Closable.
func (h *ServerIO) Close() error {
	return errors.Join(
		h.users.Close(),
		h.sessionHistory.Close(),
	)
}

// Network implements proxy.Inbound.Network().
func (*ServerIO) Network() []net.Network {
	return []net.Network{net.Network_TCP, net.Network_UNIX}
}

func (h *ServerIO) AddUser(user i.User) {
	h.users.Add(vmess.NewMemoryAccount(user.Uid(), user.Secret(), 0, protocol.SecurityType_AUTO, false, false))
}

func (h *ServerIO) RemoveUser(uid, secret string) {
	h.users.Remove(secret)
}

func (s *ServerIO) FallbackProcess(ctx context.Context, conn net.Conn) (bool, buf.MultiBuffer, error) {
	if err := conn.SetReadDeadline(time.Now().Add(time.Duration(s.policyManager.HandshakeTimeout()))); err != nil {
		return false, nil, fmt.Errorf("unable to set read deadline, %w", err)
	}

	// Create a cache reader to capture all data read during header decoding
	cacheReader := buf.NewMemoryReader(conn)

	reader := &buf.BufferedReader{Reader: buf.NewReader(cacheReader)}

	svrSession := encoding.NewServerSession(s.users, s.sessionHistory)
	svrSession.SetAEADForced(aeadForced)
	request, err := svrSession.DecodeRequestHeader(reader, false)
	if err1 := conn.SetReadDeadline(time.Time{}); err1 != nil {
		return false, nil, fmt.Errorf("failed to clear deadline on tcp Conn, %w", err)
	}
	if err != nil {
		// Failed to decode vmess header - this is not valid vmess traffic or auth failed
		// Return true to indicate fallback should be used, along with the cached data
		return true, cacheReader.History(), err
	}
	cacheReader.StopMemorize()

	return false, nil, s.processCommon(ctx, conn, request, svrSession, reader)
}

// Process implements proxy.Inbound.Process().
func (h *ServerIO) Process(ctx context.Context, conn net.Conn) error {
	if err := conn.SetReadDeadline(time.Now().Add(time.Duration(h.policyManager.HandshakeTimeout()))); err != nil {
		return fmt.Errorf("unable to set read deadline, %w", err)
	}

	reader := &buf.BufferedReader{Reader: buf.NewReader(conn)}
	svrSession := encoding.NewServerSession(h.users, h.sessionHistory)
	svrSession.SetAEADForced(aeadForced)
	request, err := svrSession.DecodeRequestHeader(reader, true)
	if err != nil {
		if !errors.Is(err, io.EOF) {
			err = fmt.Errorf("invalid request from %v %w", conn.RemoteAddr(), err)
			if h.onUnauthorizedRequest != nil {
				h.onUnauthorizedRequest.ReportUnauthorized(conn.RemoteAddr().String(), "")
			}
		}
		return err
	}
	if err := conn.SetReadDeadline(time.Time{}); err != nil {
		return fmt.Errorf("unable to clear read deadline, %w", err)
	}

	return h.processCommon(ctx, conn, request, svrSession, reader)
}

func (h *ServerIO) processCommon(ctx context.Context, conn net.Conn, request *protocol.RequestHeader,
	svrSession *encoding.ServerSession, reader io.Reader) error {
	if h.secure && isInsecureEncryption(request.Security) {
		return fmt.Errorf("client is using insecure encryption: %v", request.Security)
	}

	var err error
	ctx = proxy.ContextWithUser(ctx, request.User)

	log.Ctx(ctx).Debug().Str("target", request.Destination().String()).Msg("vmess server header decoded")

	bodyReader, err := svrSession.DecodeRequestBody1(ctx, request, reader)
	if err != nil {
		return fmt.Errorf("failed to start decoding, %w", err)
	}

	response := &protocol.ResponseHeader{
		Command: h.generateCommand(ctx, request),
	}
	svrSession.EncodeResponseHeader(response, conn)
	bodyWriter, err := svrSession.EncodeResponseBody1(request, conn)
	if err != nil {
		return fmt.Errorf("failed to start decoding response, %w", err)
	}
	// if the following is not sent, the client will get an error "failed to read size" caused by
	// AuthenticationReader.readSize(), which reads two bytes size from vmess stream
	// This is acutally an  end signal of the stream.
	var closeWrite func()
	if request.Option.Has(protocol.RequestOptionChunkStream) {
		closeWrite = func() {
			if _, err := bodyWriter.Write([]byte{}); err != nil {
				log.Ctx(ctx).Error().Err(err).Msg("bodyWriter failed to close write")
			}
		}
	}

	return h.d.HandleConn(ctx, request.Destination(), proxy.NewProxyConn(proxy.CustomConnOption{
		Conn:       conn,
		Writer:     bodyWriter,
		Reader:     bodyReader,
		CloseWrite: closeWrite,
	}))
}

func (h *ServerIO) generateCommand(ctx context.Context, request *protocol.RequestHeader) protocol.ResponseCommand {
	//todo
	// if h.detours != nil {
	// 	tag := h.detours.To
	// 	if h.inboundHandlerManager != nil {
	// 		handler, err := h.inboundHandlerManager.GetHandler(ctx, tag)
	// 		if err != nil {
	// 			common.WarningLogger.Printf("{%v} failed to get detour handler: ", session.IDFromContext(ctx),tag, err)
	// 			return nil
	// 		}
	// 		proxyS, port, availableMin := handler.
	// 		// inboundHandler, ok := proxyS.(*Handler) //todo
	// 		if ok && inboundHandler != nil {
	// 			if availableMin > 255 {
	// 				availableMin = 255
	// 			}

	// 			common.DebugLogger.Printf("{%v} pick detour handler for port %d for %v minutes.", session.IDFromContext(ctx), port, availableMin)
	// 			user := inboundHandler.GetUser(request.User.Username())
	// 			if user == nil {
	// 				return nil
	// 			}
	// 			account := user.Account.(*vmess.MemoryAccount)
	// 			return &protocol.CommandSwitchAccount{
	// 				Port:     port,
	// 				ID:       account.ID.UUID(),
	// 				AlterIds: uint16(len(account.AlterIDs)),
	// 				Level:    user.Level,
	// 				ValidMin: byte(availableMin),
	// 			}
	// 		}
	// 	}
	// }

	return nil
}
