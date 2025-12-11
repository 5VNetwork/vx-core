//go:build server

package inbound

import (
	"bytes"
	"context"
	gotls "crypto/tls"
	"io"
	"reflect"
	"strconv"
	sync "sync"
	"time"
	"unsafe"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/pipe"
	"github.com/5vnetwork/vx-core/common/protocol"
	"github.com/5vnetwork/vx-core/common/signal"
	"github.com/5vnetwork/vx-core/common/task"
	"github.com/5vnetwork/vx-core/common/uuid"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/proxy"
	"github.com/5vnetwork/vx-core/proxy/vless"
	"github.com/5vnetwork/vx-core/proxy/vless/encoding"
	"github.com/5vnetwork/vx-core/transport/security/reality"
	"github.com/5vnetwork/vx-core/transport/security/tls"

	"github.com/rs/zerolog/log"
)

// Handler is an inbound connection handler that handles messages in VLess protocol.
type Handler struct {
	policyManager i.TimeoutSetting
	users         sync.Map //key: uuid.UUID secret
	d             i.Handler
}

type HandlerSettings struct {
	PolicyManager i.TimeoutSetting
	Handler       i.Handler
}

// New creates a new VLess inbound handler.
func New(settings HandlerSettings) (*Handler, error) {
	handler := &Handler{
		policyManager: settings.PolicyManager,
		d:             settings.Handler,
	}

	return handler, nil
}

func isMuxAndNotXUDP(request *protocol.RequestHeader, first *buf.Buffer) bool {
	if request.Command != protocol.RequestCommandMux {
		return false
	}
	if first.Len() < 7 {
		return true
	}
	firstBytes := first.Bytes()
	return !(firstBytes[2] == 0 && // ID high
		firstBytes[3] == 0 && // ID low
		firstBytes[6] == 2) // Network type: UDP
}

func (h *Handler) AddUser(user i.User) {
	account := &vless.MemoryAccount{
		Uid:        user.Uid(),
		ID:         protocol.NewID(uuid.StringToUUID(user.Secret())),
		Encryption: "none",
		Flow:       "xtls-rprx-vision",
	}
	h.users.Store(account.ID.UUID(), account)
}

func (h *Handler) RemoveUser(uid, secret string) {
	h.users.Delete(uuid.StringToUUID(uid))
}

// Network implements proxy.Inbound.Network().
func (*Handler) Network() []net.Network {
	return []net.Network{net.Network_TCP, net.Network_UNIX}
}

// Process implements proxy.Inbound.Process().
func (h *Handler) Process(ctx context.Context, conn net.Conn) error {
	if err := conn.SetReadDeadline(time.Now().Add(h.policyManager.HandshakeTimeout())); err != nil {
		return errors.New("unable to set read deadline").Base(err)
	}

	first := buf.New()
	firstLen, _ := first.ReadOnce(conn)
	log.Ctx(ctx).Info().Msg("firstLen = " + strconv.Itoa(int(firstLen)))

	reader := &buf.BufferedReader{
		Reader: buf.NewReader(conn),
		Buffer: buf.MultiBuffer{first},
	}

	var request *protocol.RequestHeader
	var requestAddons *encoding.Addons
	var err error

	request, requestAddons, _, err = encoding.DecodeRequestHeader(false, first, reader, &h.users)

	if err != nil {
		if !errors.Is(err, io.EOF) {
			err = errors.New("invalid request from ", conn.RemoteAddr()).Base(err)
		}
		return err
	}

	if err := conn.SetReadDeadline(time.Time{}); err != nil {
		// errors.LogWarningInner(ctx, err, "unable to set back read deadline")
	}
	// errors.LogInfo(ctx, "received request for ", request.Destination())

	// inbound.Name = "vless"
	ctx = proxy.ContextWithUser(ctx, request.User)
	account := request.Account.(*vless.MemoryAccount)

	responseAddons := &encoding.Addons{
		// Flow: requestAddons.Flow,
	}

	var input *bytes.Reader
	var rawInput *bytes.Buffer
	switch requestAddons.Flow {
	case vless.XRV:
		if account.Flow == requestAddons.Flow {
			// info.SpliceCopy.Whether = session.SpliceCopyNotSureYet
			switch request.Command {
			case protocol.RequestCommandUDP:
				return errors.New(requestAddons.Flow + " doesn't support UDP")
			case protocol.RequestCommandMux:
				fallthrough // we will break Mux connections that contain TCP requests
			case protocol.RequestCommandTCP:
				var t reflect.Type
				var p uintptr
				if tlsConn, ok := conn.(*gotls.Conn); ok {
					if tlsConn.ConnectionState().Version != gotls.VersionTLS13 {
						return errors.New(`failed to use `+requestAddons.Flow+`, found outer tls version `, tlsConn.ConnectionState().Version)
					}
					t = reflect.TypeOf(tlsConn).Elem()
					p = uintptr(unsafe.Pointer(tlsConn))
				} else if tlsConn, ok := conn.(*tls.Conn); ok {
					if tlsConn.ConnectionState().Version != gotls.VersionTLS13 {
						return errors.New(`failed to use `+requestAddons.Flow+`, found outer tls version `, tlsConn.ConnectionState().Version)
					}
					t = reflect.TypeOf(tlsConn.Conn).Elem()
					p = uintptr(unsafe.Pointer(tlsConn.Conn))
				} else if realityConn, ok := conn.(*reality.Conn); ok {
					t = reflect.TypeOf(realityConn.Conn).Elem()
					p = uintptr(unsafe.Pointer(realityConn.Conn))
				} else {
					return errors.New("XTLS only supports TLS and REALITY directly for now.")
				}

				i, _ := t.FieldByName("input")
				r, _ := t.FieldByName("rawInput")
				input = (*bytes.Reader)(unsafe.Pointer(p + i.Offset))
				rawInput = (*bytes.Buffer)(unsafe.Pointer(p + r.Offset))
			}
		} else {
			return errors.New(account.ID.String() + " is not able to use " + requestAddons.Flow)
		}
	case "":
		// info.SpliceCopy.Whether = session.SpliceCopyNo
		if account.Flow == vless.XRV && (request.Command == protocol.RequestCommandTCP || isMuxAndNotXUDP(request, first)) {
			return errors.New("account " + account.ID.String() + " is rejected since the client flow is empty. Note that the pure TLS proxy has certain TLS in TLS characters.")
		}
	default:
		return errors.New("unknown request flow " + requestAddons.Flow)
	}

	if request.Command != protocol.RequestCommandMux {

	} else if account.Flow == vless.XRV {
		// ctx = session.ContextWithAllowedNetwork(ctx, net.Network_UDP)
	}

	in := &vless.InboundInfo{
		// Conn:          info.RawConn,
		// CanSpliceCopy: info.SpliceCopy.ToVlessNum(),
	}
	ctx = vless.WithInbound(ctx, in)

	ctx, cancel := context.WithCancel(ctx)
	timer := signal.NewActivityChecker(cancel, time.Second*120)
	in.Timer = timer
	// ctx = policy.ContextWithBufferPolicy(ctx, sessionPolicy.Buffer)
	iLink, oLink := pipe.NewLinks(buf.Size, false)
	defer iLink.Interrupt(nil)

	go func() {
		if err = h.d.HandleFlow(ctx, request.Destination(), oLink); err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("failed to handle")
		}
		cancel()
	}()

	serverReader := iLink // .(*pipe.Reader)
	serverWriter := iLink // .(*pipe.Writer)
	trafficState := vless.NewTrafficState(account.ID.Bytes())
	postRequest := func() error {
		defer timer.SetTimeout(h.policyManager.DownLinkOnlyTimeout())

		// default: clientReader := reader
		clientReader := encoding.DecodeBodyAddons(reader, request, requestAddons)

		var err error

		if requestAddons.Flow == vless.XRV {
			ctx1 := vless.WithInbound(ctx, nil) // TODO enable splice
			clientReader = vless.NewVisionReader(clientReader, trafficState, ctx1)
			err = encoding.XtlsRead(clientReader, serverWriter, timer, conn, input, rawInput, trafficState, nil, ctx1)
		} else {
			// from clientReader.ReadMultiBuffer to serverWriter.WriteMultiBuffer
			err = buf.Copy(clientReader, serverWriter, buf.UpdateActivityCopyOption(timer))
		}

		if err != nil {
			return errors.New("failed to transfer request payload").Base(err)
		}

		return nil
	}

	getResponse := func() error {
		defer timer.SetTimeout(h.policyManager.UpLinkOnlyTimeout())

		bufferWriter := buf.NewBufferedWriter(buf.NewWriter(conn))
		if err := encoding.EncodeResponseHeader(bufferWriter, request, responseAddons); err != nil {
			return errors.New("failed to encode response header").Base(err)
		}

		// default: clientWriter := bufferWriter
		clientWriter := encoding.EncodeBodyAddons(bufferWriter, request, requestAddons, trafficState, ctx)
		multiBuffer, err1 := serverReader.ReadMultiBuffer()
		if err1 != nil {
			return err1 // ...
		}
		if err := clientWriter.WriteMultiBuffer(multiBuffer); err != nil {
			return err // ...
		}
		// Flush; bufferWriter.WriteMultiBuffer now is bufferWriter.writer.WriteMultiBuffer
		if err := bufferWriter.SetBuffered(false); err != nil {
			return errors.New("failed to write A response payload").Base(err)
		}

		var err error
		if requestAddons.Flow == vless.XRV {
			err = encoding.XtlsWrite(serverReader, clientWriter, timer, conn, trafficState, nil, ctx)
		} else {
			// from serverReader.ReadMultiBuffer to clientWriter.WriteMultiBuffer
			err = buf.Copy(serverReader, clientWriter, buf.UpdateActivityCopyOption(timer))
		}
		if err != nil {
			return errors.New("failed to transfer response payload").Base(err)
		}
		// Indicates the end of response payload.
		switch responseAddons.Flow {
		default:
		}

		return nil
	}

	if err := task.Run(ctx, task.OnSuccess(postRequest, serverWriter.CloseWrite), getResponse); err != nil {
		common.Interrupt(serverReader)
		common.Interrupt(serverWriter)
		return errors.New("connection ends").Base(err)
	}

	return nil
}
