//go:build server

package server

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/platform"
	"github.com/5vnetwork/vx-core/common/protocol"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/proxy"
	"github.com/5vnetwork/vx-core/proxy/vmess"
	"github.com/5vnetwork/vx-core/proxy/vmess/encoding"

	"github.com/rs/zerolog/log"
)

// Server is an inbound connection handler that handles messages in VMess protocol.
type Server struct {
	ServerSettings
	users          *vmess.TimedUserValidator
	sessionHistory *encoding.SessionHistory
}

type ServerSettings struct {
	Handler               i.Handler
	OnUnauthorizedRequest i.UnauthorizedReport
	PolicyManager         i.TimeoutSetting
	Secure                bool
}

// New creates a new VMess inbound handler.
func New(settings ServerSettings) *Server {
	handler := &Server{
		ServerSettings: settings,
		users:          vmess.NewTimedUserValidator(protocol.DefaultIDHash),
		sessionHistory: encoding.NewSessionHistory(),
	}

	return handler
}

func (h *Server) Start() error {
	return nil
}

// Close implements common.Closable.
func (h *Server) Close() error {
	return errors.Join(
		h.users.Close(),
		h.sessionHistory.Close(),
	)
}

// Network implements proxy.Inbound.Network().
func (*Server) Network() []net.Network {
	return []net.Network{net.Network_TCP, net.Network_UNIX}
}

func (h *Server) AddUser(user i.User) {
	h.users.Add(vmess.NewMemoryAccount(user.Uid(), user.Secret(),
		0, protocol.SecurityType_AUTO, false, false))
}

func (h *Server) RemoveUser(uid, secret string) {
	if secret == "" {
		h.users.RemoveByUid(uid)
	} else {
		h.users.Remove(secret)
	}
}

func (h *Server) WithOnUnauthorizedRequest(f i.UnauthorizedReport) {
	h.OnUnauthorizedRequest = f
}

func isInsecureEncryption(s protocol.SecurityType) bool {
	return s == protocol.SecurityType_NONE || s == protocol.SecurityType_LEGACY || s == protocol.SecurityType_UNKNOWN
}

func (s *Server) FallbackProcess(ctx context.Context, conn net.Conn) (bool, buf.MultiBuffer, error) {
	if err := conn.SetReadDeadline(time.Now().Add(time.Duration(s.PolicyManager.HandshakeTimeout()))); err != nil {
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
func (h *Server) Process(ctx context.Context, conn net.Conn) error {
	if err := conn.SetReadDeadline(time.Now().Add(time.Duration(h.PolicyManager.HandshakeTimeout()))); err != nil {
		return fmt.Errorf("unable to set read deadline, %w", err)
	}

	reader := &buf.BufferedReader{Reader: buf.NewReader(conn)}
	svrSession := encoding.NewServerSession(h.users, h.sessionHistory)
	svrSession.SetAEADForced(aeadForced)
	request, err := svrSession.DecodeRequestHeader(reader, true)
	if err != nil {
		if !errors.Is(err, io.EOF) {
			err = fmt.Errorf("invalid request from %v %w", conn.RemoteAddr(), err)
			if h.OnUnauthorizedRequest != nil {
				h.OnUnauthorizedRequest.ReportUnauthorized(conn.RemoteAddr().String(), "")
			}
		}
		return err
	}
	if err := conn.SetReadDeadline(time.Time{}); err != nil {
		return fmt.Errorf("unable to clear read deadline, %w", err)
	}

	return h.processCommon(ctx, conn, request, svrSession, reader)
}

// func (h *Server) processCommon(ctx context.Context, conn net.Conn, request *protocol.RequestHeader,
// 	svrSession *encoding.ServerSession, reader io.Reader) error {
// 	if h.Secure && isInsecureEncryption(request.Security) {
// 		return fmt.Errorf("client is using insecure encryption: %v", request.Security)
// 	}

// 	var err error
// 	ctx = proxy.ContextWithUser(ctx, request.User)
// 	dst := request.Destination()

// 	log.Ctx(ctx).Debug().Str("target", dst.String()).Msg("vmess server header decoded")

// 	iLink, oLink := helper.GetLinks(net.Network_TCP, 0, h.PolicyManager.(i.BufferPolicy))
// 	defer iLink.Interrupt(nil)

// 	go func() {
// 		if err = h.Handler.HandleFlow(ctx, request.Destination(), oLink); err != nil {
// 			log.Ctx(ctx).Err(err).Msg("failed to handle")
// 		}
// 	}()

// 	requestDone := func() error {
// 		var err error
// 		var bodyReader buf.Reader
// 		bodyReader, err = svrSession.DecodeRequestBody(ctx, request, reader)
// 		if err != nil {
// 			return fmt.Errorf("failed to start decoding, %w", err)
// 		}
// 		if err = buf.Copy(bodyReader, iLink); err != nil {
// 			return fmt.Errorf("failed to transfer request, %w", err)
// 		}
// 		iLink.CloseWrite()
// 		return nil
// 	}

// 	responseDone := func() error {
// 		var err error
// 		writer := buf.NewBufferedWriter(buf.NewWriter(conn))
// 		response := &protocol.ResponseHeader{
// 			Command: h.generateCommand(ctx, request),
// 		}
// 		err = transferResponse(svrSession, request, response, iLink, writer)
// 		writer.Flush()
// 		if err != nil {
// 			return fmt.Errorf("failed to transfer response, %w", err)
// 		}
// 		writer.CloseWrite()
// 		return nil
// 	}

// 	err = task.Run(ctx, requestDone, responseDone)
// 	return err
// }
// func transferResponse(session *encoding.ServerSession, request *protocol.RequestHeader, response *protocol.ResponseHeader, input buf.Reader, output *buf.BufferedWriter) error {
// 	session.EncodeResponseHeader(response, output)

// 	bodyWriter, err := session.EncodeResponseBody(request, output)
// 	if err != nil {
// 		return fmt.Errorf("failed to start decoding response, %w", err)
// 	}
// 	{
// 		// Optimize for small response packet
// 		data, err := input.ReadMultiBuffer()
// 		if !data.IsEmpty() {
// 			if err := bodyWriter.WriteMultiBuffer(data); err != nil {
// 				return fmt.Errorf("failed to write first response data, %w", err)
// 			}
// 		}
// 		if err != nil {
// 			if errors.Is(err, io.EOF) {
// 				return nil
// 			}
// 			return fmt.Errorf("failed to read first response data, %w", err)
// 		}
// 	}

// 	if err := output.SetBuffered(false); err != nil {
// 		return err
// 	}

// 	if err := buf.Copy(input, bodyWriter); err != nil {
// 		return err
// 	}
// 	// if the following is not sent, the client will get an error "failed to read size" caused by
// 	// AuthenticationReader.readSize(), which reads two bytes size from vmess stream
// 	// This is acutally an  end signal of the stream.
// 	if request.Option.Has(protocol.RequestOptionChunkStream) {
// 		if err := bodyWriter.WriteMultiBuffer(buf.MultiBuffer{}); err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

func (h *Server) processCommon(ctx context.Context, conn net.Conn, request *protocol.RequestHeader,
	svrSession *encoding.ServerSession, reader io.Reader) error {
	if h.Secure && isInsecureEncryption(request.Security) {
		return fmt.Errorf("client is using insecure encryption: %v", request.Security)
	}

	var err error
	ctx = proxy.ContextWithUser(ctx, request.User)

	log.Ctx(ctx).Debug().Str("target", request.Destination().String()).Msg("vmess server header decoded")

	bodyReader, err := svrSession.DecodeRequestBody(ctx, request, reader)
	if err != nil {
		return fmt.Errorf("failed to start decoding, %w", err)
	}

	// resposne header will be sent together with first response payload
	bufferedWriter := buf.NewBufferedWriter(buf.NewWriter(conn))
	svrSession.EncodeResponseHeader(&protocol.ResponseHeader{
		Command: h.generateCommand(ctx, request),
	}, bufferedWriter)
	bodyWriter, err := svrSession.EncodeResponseBody(request, bufferedWriter)
	if err != nil {
		return fmt.Errorf("failed to start decoding response, %w", err)
	}

	return h.Handler.HandleFlow(ctx, request.Destination(), &rw{
		Reader:         bodyReader,
		Writer:         bodyWriter,
		bufferedWriter: bufferedWriter,
		closeWrite:     request.Option.Has(protocol.RequestOptionChunkStream),
		ReadDeadline:   conn,
	})
}

type rw struct {
	buf.Reader
	writeLock  sync.Mutex
	headerSent bool
	buf.Writer
	bufferedWriter *buf.BufferedWriter
	closeWrite     bool
	i.ReadDeadline
}

func (r *rw) ReadMultiBuffer() (buf.MultiBuffer, error) {
	return r.Reader.ReadMultiBuffer()
}

func (r *rw) WriteMultiBuffer(mb buf.MultiBuffer) error {
	r.writeLock.Lock()
	defer r.writeLock.Unlock()
	if !r.headerSent {
		r.headerSent = true
		err := r.Writer.WriteMultiBuffer(mb)
		if err != nil {
			return err
		}
		if err := r.bufferedWriter.SetBuffered(false); err != nil {
			return err
		}
		return nil
	}
	return r.Writer.WriteMultiBuffer(mb)
}

func (r *rw) CloseWrite() error {
	// in case there is no response payload, flush the buffered header
	r.bufferedWriter.Flush()
	if r.closeWrite {
		err := r.Writer.WriteMultiBuffer(buf.MultiBuffer{})
		if err != nil {
			return err
		}
	}
	return r.Writer.CloseWrite()
}

func (h *Server) generateCommand(ctx context.Context, request *protocol.RequestHeader) protocol.ResponseCommand {
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

var (
	aeadForced     = false
	aeadForced2022 = false
)

func init() {
	defaultFlagValue := "true_by_default_2022"

	isAeadForced := platform.NewEnvFlag("v2ray.aead.forced").GetValue(func() string { return defaultFlagValue })
	if isAeadForced == "true" {
		aeadForced = true
	}

	if isAeadForced == "true_by_default_2022" {
		aeadForced = true
		aeadForced2022 = true
	}
}
