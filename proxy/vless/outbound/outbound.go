package outbound

import (
	"bytes"
	"context"
	gotls "crypto/tls"
	"fmt"
	"io"
	"reflect"
	"time"
	"unsafe"

	"github.com/5vnetwork/vx-core/app/inbound"
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/dispatcher"
	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/5vnetwork/vx-core/common/mux"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/common/protocol"
	"github.com/5vnetwork/vx-core/common/session"
	"github.com/5vnetwork/vx-core/common/signal"
	"github.com/5vnetwork/vx-core/common/task"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/proxy/vless"
	"github.com/5vnetwork/vx-core/proxy/vless/encoding"
	"github.com/5vnetwork/vx-core/proxy/vless/xudp"

	"github.com/5vnetwork/vx-core/transport/security/reality"
	"github.com/5vnetwork/vx-core/transport/security/tls"

	utls "github.com/refraction-networking/utls"
	"github.com/rs/zerolog/log"
)

// Handler is an outbound connection handler for VLess protocol.
type Handler struct {
	ClientSettings
}

type ClientSettings struct {
	Address        net.Address
	ServerPicker   i.PortSelector
	Account        *vless.MemoryAccount
	TimeoutSetting i.TimeoutSetting
	Dialer         i.Dialer
}

// New creates a new VLess outbound handler.
func New(settings ClientSettings) *Handler {
	handler := &Handler{
		ClientSettings: settings,
	}
	return handler
}

func (h *Handler) HandleFlow(ctx context.Context, dst net.Destination, rw buf.ReaderWriter) error {
	return h.handle(ctx, dst, rw)
}

func (h *Handler) HandlePacketConn(ctx context.Context, dst net.Destination, p udp.PacketReaderWriter) error {
	dialer := h.Dialer
	port := h.ServerPicker.SelectPort()
	dest := net.TCPDestination(h.Address, net.Port(port))
	account := h.Account
	requestAddons := &encoding.Addons{
		Flow: account.Flow,
	}

	if requestAddons.Flow == (vless.XRV + "-udp443") {
		if dst.Port == 443 {
			return ErrRejectQuic
		}
		requestAddons.Flow = requestAddons.Flow[:16]
	}

	// xudp case, full cone NAT
	if requestAddons.Flow == vless.XRV ||
		(dst.Port != 53 && dst.Port != 443) {
		var conn net.Conn
		conn, err := dialer.Dial(ctx, dest)
		if err != nil {
			return fmt.Errorf("failed to dial, %w", err)
		}
		defer conn.Close()

		if conn.LocalAddr() != nil {
			log.Ctx(ctx).Debug().Str("laddr", conn.LocalAddr().String()).Msg("vless dial ok")
		} else {
			log.Ctx(ctx).Warn().Msg("vless dial ok, but local addr is nil")
		}

		request := &protocol.RequestHeader{
			Version: encoding.Version,
			Account: account,
			Command: protocol.RequestCommandMux,
			Address: net.DomainAddress("v1.mux.cool"),
			Port:    net.Port(666),
		}
		trafficState := vless.NewTrafficState(account.ID.Bytes())
		postRequest := func() error {
			bufferWriter := buf.NewBufferedWriter(buf.NewWriter(conn))
			if err := encoding.EncodeRequestHeader(bufferWriter, request, requestAddons); err != nil {
				return errors.New("failed to encode request header").Base(err)
			}
			// default: serverWriter := bufferWriter
			serverWriter := encoding.EncodeBodyAddons(bufferWriter, request, requestAddons, trafficState, ctx)
			serverWriter = xudp.NewPacketWriter(serverWriter, dst, xudp.GetGlobalID(ctx))
			if requestAddons.Flow == vless.XRV {
				mb := make(buf.MultiBuffer, 1)
				log.Ctx(ctx).Debug().Msg("Insert padding with empty content to camouflage VLESS header")
				if err := serverWriter.WriteMultiBuffer(mb); err != nil {
					return err // ...
				}
			}
			// Flush; bufferWriter.WriteMultiBuffer now is bufferWriter.writer.WriteMultiBuffer
			if err := bufferWriter.SetBuffered(false); err != nil {
				return errors.New("failed to write A request payload").Base(err)
			}
			if requestAddons.Flow == vless.XRV {
				if tlsConn, ok := conn.(*tls.Conn); ok {
					if tlsConn.ConnectionState().Version != gotls.VersionTLS13 {
						return errors.New(`failed to use `+requestAddons.Flow+`, found outer tls version `, tlsConn.ConnectionState().Version)
					}
				} else if utlsConn, ok := conn.(*tls.UConn); ok {
					if utlsConn.ConnectionState().Version != utls.VersionTLS13 {
						return errors.New(`failed to use `+requestAddons.Flow+`, found outer tls version `, utlsConn.ConnectionState().Version)
					}
				}
			}
			// from clientReader.ReadMultiBuffer to serverWriter.WriteMultiBuffer
			xudpWriter := serverWriter.(*xudp.PacketWriter)
			for {
				p, err := p.ReadPacket()
				if err != nil {
					if err == io.EOF {
						return nil
					}
					return fmt.Errorf("failed to read packet from packetConn: %w", err)
				}
				if err := xudpWriter.WritePacket(p); err != nil {
					return fmt.Errorf("failed to write packet to server: %w", err)
				}
			}
		}
		getResponse := func() error {
			responseAddons, err := encoding.DecodeResponseHeader(conn, request)
			if err != nil {
				return errors.New("failed to decode response header").Base(err)
			}

			// default: serverReader := buf.NewReader(conn)
			serverReader := encoding.DecodeBodyAddons(conn, request, responseAddons)
			if requestAddons.Flow == vless.XRV {
				serverReader = vless.NewVisionReader(serverReader, trafficState, ctx)
			}
			if requestAddons.Flow == vless.XRV {
				serverReader = xudp.NewPacketReader(&buf.BufferedReader{Reader: serverReader})
			} else {
				serverReader = xudp.NewPacketReader(conn)
			}
			xudpReader := serverReader.(*xudp.PacketReader)
			for {
				pk, err := xudpReader.ReadPacket()
				if err != nil {
					if err == io.EOF {
						return nil
					}
					return fmt.Errorf("failed to read packet from server: %w", err)
				}
				if err := p.WritePacket(pk); err != nil {
					return fmt.Errorf("failed to write packet to packetConn: %w", err)
				}
			}
		}
		if err := task.Run(ctx, postRequest, getResponse); err != nil {
			return fmt.Errorf("connection ends: %w", err)
		}
		return nil
	} else {
		d := dispatcher.NewPacketDispatcher(ctx, h, dispatcher.WithRequestTimeout(time.Second*5),
			dispatcher.WithResponseCallback(func(packet *udp.Packet) {
				p.WritePacket(packet)
			}))
		defer d.Close()
		for {
			packet, err := p.ReadPacket()
			if err != nil {
				return err
			}
			d.DispatchPacket(packet.Target, packet.Payload)
		}
	}
}

var ErrRejectQuic = errors.New("XTLS rejected QUIC traffic")

// Process implements proxy.Outbound.Process().
func (h *Handler) handle(ctx context.Context, dst net.Destination, rw buf.ReaderWriter) error {
	dialer := h.Dialer
	info := session.InfoFromContext(ctx)
	rawConn, ok := inbound.RawConnFromContext(ctx)
	inboundCanSpliceCopy, outboundCanSpliceCopy := 3, 3
	if ok {
		inboundCanSpliceCopy = 1
		outboundCanSpliceCopy = 2
	}
	ob := &vless.OutboundInfo{
		Target:        dst,
		CanSpliceCopy: outboundCanSpliceCopy,
	}
	ctx = vless.WithOutbounds(ctx, []*vless.OutboundInfo{ob})

	ib := &vless.InboundInfo{
		CanSpliceCopy: inboundCanSpliceCopy,
		Conn:          rawConn,
	}
	if info != nil {
		ib.UpCounter = info.UpCounter
		ib.DownCounter = info.DownCounter
		ib.Timer = info.ActivityChecker
	}
	ctx = vless.WithInbound(ctx, ib)

	var conn net.Conn
	var account *vless.MemoryAccount
	port := h.ServerPicker.SelectPort()

	dest := net.TCPDestination(h.Address, net.Port(port))
	account = h.Account
	conn, err := dialer.Dial(ctx, dest)
	if err != nil {
		return fmt.Errorf("failed to find an available destination, %w", err)
	}
	defer conn.Close()

	if conn.LocalAddr() != nil {
		log.Ctx(ctx).Debug().Str("laddr", conn.LocalAddr().String()).Msg("vless dial ok")
	} else {
		log.Ctx(ctx).Warn().Msg("vless dial ok, but local addr is nil")
	}

	target := ob.Target

	command := protocol.RequestCommandTCP
	if target.Network == net.Network_UDP {
		command = protocol.RequestCommandUDP
	}
	if target.Address.Family().IsDomain() && target.Address.Domain() == mux.MuxCoolAddressDst.String() {
		command = protocol.RequestCommandMux
	}

	request := &protocol.RequestHeader{
		Version: encoding.Version,
		Account: account,
		Command: command,
		Address: target.Address,
		Port:    target.Port,
	}

	requestAddons := &encoding.Addons{
		Flow: account.Flow,
	}

	var input *bytes.Reader
	var rawInput *bytes.Buffer
	allowUDP443 := false
	switch requestAddons.Flow {
	case vless.XRV + "-udp443":
		allowUDP443 = true
		requestAddons.Flow = requestAddons.Flow[:16]
		fallthrough
	case vless.XRV:
		ob.CanSpliceCopy = 2
		switch request.Command {
		case protocol.RequestCommandUDP:
			if !allowUDP443 && request.Port == 443 {
				return ErrRejectQuic
			}
		case protocol.RequestCommandMux:
			fallthrough // let server break Mux connections that contain TCP requests
		case protocol.RequestCommandTCP:
			var t reflect.Type
			var p uintptr
			if tlsConn, ok := conn.(*tls.Conn); ok {
				t = reflect.TypeOf(tlsConn.Conn).Elem()
				p = uintptr(unsafe.Pointer(tlsConn.Conn))
			} else if utlsConn, ok := conn.(*tls.UConn); ok {
				t = reflect.TypeOf(utlsConn.Conn).Elem()
				p = uintptr(unsafe.Pointer(utlsConn.Conn))
			} else if realityConn, ok := conn.(*reality.UConn); ok {
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
	default:
		ob.CanSpliceCopy = 3
	}

	// var newCtx context.Context
	// var newCancel context.CancelFunc

	ctx, cancelCause := context.WithCancelCause(ctx)
	timer := signal.NewActivityChecker(func() {
		cancelCause(errors.ErrIdle)
		// if newCancel != nil {
		// 	newCancel()
		// }
	}, h.TimeoutSetting.TcpIdleTimeout())

	clientReader := rw // .(*pipe.Reader)
	clientWriter := rw // .(*pipe.Writer)
	trafficState := vless.NewTrafficState(account.ID.Bytes())
	if request.Command == protocol.RequestCommandUDP && (requestAddons.Flow == vless.XRV ||
		(request.Port != 53 && request.Port != 443)) {
		request.Command = protocol.RequestCommandMux
		request.Address = net.DomainAddress("v1.mux.cool")
		request.Port = net.Port(666)
	}

	postRequest := func() error {
		defer timer.SetTimeout(h.TimeoutSetting.DownLinkOnlyTimeout())

		bufferWriter := buf.NewBufferedWriter(buf.NewWriter(conn))
		if err := encoding.EncodeRequestHeader(bufferWriter, request, requestAddons); err != nil {
			return errors.New("failed to encode request header").Base(err)
		}

		// default: serverWriter := bufferWriter
		serverWriter := encoding.EncodeBodyAddons(bufferWriter, request, requestAddons, trafficState, ctx)
		if request.Command == protocol.RequestCommandMux && request.Port == 666 {
			serverWriter = xudp.NewPacketWriter(serverWriter, target, [8]byte{})
		}
		err1 := buf.CopyOnceTimeout(clientReader, serverWriter, time.Millisecond*500)
		if err1 != nil {
			if err1 != buf.ErrReadTimeout && err1 != buf.ErrNotTimeoutReader {
				return err1
			} else if requestAddons.Flow == vless.XRV {
				mb := make(buf.MultiBuffer, 1)
				log.Ctx(ctx).Debug().Msg("Insert padding with empty content to camouflage VLESS header")
				if err := serverWriter.WriteMultiBuffer(mb); err != nil {
					return err // ...
				}
			}
		}

		// Flush; bufferWriter.WriteMultiBuffer now is bufferWriter.writer.WriteMultiBuffer
		if err := bufferWriter.SetBuffered(false); err != nil {
			return errors.New("failed to write A request payload").Base(err)
		}

		var err error
		if requestAddons.Flow == vless.XRV {
			if tlsConn, ok := conn.(*tls.Conn); ok {
				if tlsConn.ConnectionState().Version != gotls.VersionTLS13 {
					return errors.New(`failed to use `+requestAddons.Flow+`, found outer tls version `, tlsConn.ConnectionState().Version)
				}
			} else if utlsConn, ok := conn.(*tls.UConn); ok {
				if utlsConn.ConnectionState().Version != utls.VersionTLS13 {
					return errors.New(`failed to use `+requestAddons.Flow+`, found outer tls version `, utlsConn.ConnectionState().Version)
				}
			}
			ctx1 := vless.WithInbound(ctx, nil) // TODO enable splice
			err = encoding.XtlsWrite(clientReader, serverWriter, timer, conn, trafficState, ob, ctx1)
		} else {
			// from clientReader.ReadMultiBuffer to serverWriter.WriteMultiBuffer
			err = buf.Copy(clientReader, serverWriter, buf.UpdateActivityCopyOption(timer))
		}
		if err != nil {
			return errors.New("failed to transfer request payload").Base(err)
		}
		// log.Ctx(ctx).Debug().Msg("vless request done")

		// Indicates the end of request payload.
		switch requestAddons.Flow {
		default:
		}
		return nil
	}

	getResponse := func() error {
		defer timer.SetTimeout(h.TimeoutSetting.UpLinkOnlyTimeout())

		responseAddons, err := encoding.DecodeResponseHeader(conn, request)
		if err != nil {
			return errors.New("failed to decode response header").Base(err)
		}

		// default: serverReader := buf.NewReader(conn)
		serverReader := encoding.DecodeBodyAddons(conn, request, responseAddons)
		if requestAddons.Flow == vless.XRV {
			serverReader = vless.NewVisionReader(serverReader, trafficState, ctx)
		}
		if request.Command == protocol.RequestCommandMux && request.Port == 666 {
			if requestAddons.Flow == vless.XRV {
				serverReader = xudp.NewPacketReader(&buf.BufferedReader{Reader: serverReader})
			} else {
				serverReader = xudp.NewPacketReader(conn)
			}
		}

		if requestAddons.Flow == vless.XRV {
			err = encoding.XtlsRead(serverReader, clientWriter, timer, conn, input, rawInput, trafficState, ob, ctx)
		} else {
			// from serverReader.ReadMultiBuffer to clientWriter.WriteMultiBuffer
			err = buf.Copy(serverReader, clientWriter, buf.UpdateActivityCopyOption(timer))
		}
		if err != nil {
			return fmt.Errorf("failed to transfer response payload: %w", err)
		}
		// log.Ctx(ctx).Debug().Msg("vless response done")
		clientWriter.CloseWrite()
		return nil
	}

	// if newCtx != nil {
	// 	ctx = newCtx
	// }

	if err := task.Run(ctx, postRequest, getResponse); err != nil {
		return fmt.Errorf("connection ends: %w", err)
	}
	// log.Ctx(ctx).Debug().Msg("vless connection ends")
	return nil
}
