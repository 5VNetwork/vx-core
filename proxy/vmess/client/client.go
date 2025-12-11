package client

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"hash/crc64"
	"os"
	"time"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/dispatcher"
	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/5vnetwork/vx-core/common/mux"
	"github.com/5vnetwork/vx-core/common/net"
	nethelper "github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/common/platform"
	"github.com/5vnetwork/vx-core/common/protocol"
	"github.com/5vnetwork/vx-core/common/task"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/proxy/vmess"
	"github.com/5vnetwork/vx-core/proxy/vmess/encoding"

	"github.com/5vnetwork/vx-core/proxy"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

// Client is an outbound connection handler for VMess protocol.
type Client struct {
	ClientSettings
}

type ClientSettings struct {
	ServerPicker protocol.ServerPicker
	Dialer       i.Dialer
}

// New creates a new VMess outbound handler.
func NewClient(settings ClientSettings) *Client {
	handler := &Client{
		ClientSettings: settings,
	}
	return handler
}

// Process implements proxy.Outbound.Process().
func (h *Client) HandleFlow(ctx context.Context, dst net.Destination, rw buf.ReaderWriter) error {
	rec := h.ServerPicker.PickServer()
	dialer := h.Dialer

	log.Ctx(ctx).Debug().Str("dst", rec.Destination().String()).Msg("vmess client dial")

	conn, err := dialer.Dial(ctx, rec.Destination())
	if err != nil {
		return fmt.Errorf("failed to find an available destination, %w", err)
	}
	defer conn.Close()

	log.Ctx(ctx).Debug().Str("local addr", conn.LocalAddr().String()).Msg("vmess client dial ok")

	target := dst
	if !target.IsValid() {
		return errors.New("target not specified")
	}

	command := protocol.RequestCommandTCP
	if target.Network == nethelper.Network_UDP {
		command = protocol.RequestCommandUDP
	}
	if target.Address.Family().IsDomain() && target.Address.Domain() == mux.MuxCoolAddressDst.Domain() {
		command = protocol.RequestCommandMux
	}

	request := &protocol.RequestHeader{
		Version: encoding.Version,
		Account: rec.GetProtocolSetting().(*vmess.MemoryAccount),
		Command: command,
		Address: target.Address,
		Port:    target.Port,
		Option:  protocol.RequestOptionChunkStream,
	}
	account := request.Account.(*vmess.MemoryAccount)
	request.Security = request.Account.(*vmess.MemoryAccount).Security

	if request.Security == protocol.SecurityType_AES128_GCM || request.Security == protocol.SecurityType_NONE || request.Security == protocol.SecurityType_CHACHA20_POLY1305 {
		request.Option.Set(protocol.RequestOptionChunkMasking)
	}

	if shouldEnablePadding(request.Security) && request.Option.Has(protocol.RequestOptionChunkMasking) {
		request.Option.Set(protocol.RequestOptionGlobalPadding)
	}

	if request.Security == protocol.SecurityType_ZERO {
		request.Security = protocol.SecurityType_NONE
		request.Option.Clear(protocol.RequestOptionChunkStream)
		request.Option.Clear(protocol.RequestOptionChunkMasking)
	}

	if account.AuthenticatedLengthExperiment {
		request.Option.Set(protocol.RequestOptionAuthenticatedLength)
	}

	input := rw
	output := rw

	isAEAD := false
	if !aeadDisabled && len(account.AlterIDs) == 0 {
		isAEAD = true
	}

	hashkdf := hmac.New(sha256.New, []byte("VMessBF"))
	hashkdf.Write(account.ID.Bytes())

	behaviorSeed := crc64.Checksum(hashkdf.Sum(nil), crc64.MakeTable(crc64.ISO))

	session := encoding.NewClientSession(ctx, isAEAD, protocol.DefaultIDHash, int64(behaviorSeed))

	//TODO: add timer
	requestDone := func() error {
		writer := buf.NewBufferedWriter(buf.NewWriter(conn))
		if err := session.EncodeRequestHeader(request, writer); err != nil {
			return fmt.Errorf("failed to encode request, %w", err)
		}

		bodyWriter, err := session.EncodeRequestBody(request, writer)
		if err != nil {
			return fmt.Errorf("failed to start encoding, %w", err)
		}
		if err := buf.CopyOnceTimeout(input, bodyWriter, proxy.FirstPayloadTimeout); err != nil &&
			err != buf.ErrNotTimeoutReader && err != buf.ErrReadTimeout &&
			!errors.Is(err, os.ErrDeadlineExceeded) {
			return fmt.Errorf("failed to write first payload, %w", err)
		}

		if err := writer.SetBuffered(false); err != nil {
			return err
		}

		if err := buf.Copy(input, bodyWriter); err != nil {
			return fmt.Errorf("leftToRight failed, %w", err)
		}

		if request.Option.Has(protocol.RequestOptionChunkStream) {
			if err := bodyWriter.WriteMultiBuffer(buf.MultiBuffer{}); err != nil {
				return err
			}
		}
		bodyWriter.CloseWrite()
		log.Ctx(ctx).Debug().Msg("vmess client request done")
		return nil
	}

	responseDone := func() error {
		reader := &buf.BufferedReader{Reader: buf.NewReader(conn)}
		header, err := session.DecodeResponseHeader(reader)
		if err != nil {
			return fmt.Errorf("failed to read header, %w", err)
		}
		h.handleCommand(rec.Destination(), header.Command)

		bodyReader, err := session.DecodeResponseBody(ctx, request, reader)
		if err != nil {
			return fmt.Errorf("failed to start encoding response, %w", err)
		}

		if err = buf.Copy(bodyReader, output); err != nil {
			return fmt.Errorf("rightToLeft failed, %w", err)
		}
		output.CloseWrite()
		log.Ctx(ctx).Debug().Msg("vmess client response done")
		return nil
	}

	err = task.Run(ctx, requestDone, responseDone)
	if err != nil {
		// when the server does not send vmess end packet, and the underlying is websocket
		var closeError *websocket.CloseError
		if errors.As(err, &closeError) && closeError.Code == websocket.CloseNormalClosure {
			return nil
		}
	}
	return err
}

func (h *Client) handleSwitchAccount(cmd *protocol.CommandSwitchAccount) {
	rawAccount := &vmess.MemoryAccount{
		ID:       protocol.NewID(cmd.ID),
		Security: protocol.SecurityType_LEGACY,
	}
	rawAccount.AlterIDs = protocol.NewAlterIDs(rawAccount.ID, uint16(cmd.AlterIds))
	user := &protocol.MemoryUser{
		Email:   "",
		Level:   cmd.Level,
		Account: rawAccount,
	}
	dest := nethelper.TCPDestination(cmd.Host, cmd.Port)
	until := time.Now().Add(time.Duration(cmd.ValidMin) * time.Minute)
	h.ServerPicker.AddServer(protocol.NewServerSpec(dest, protocol.BeforeTime(until), user))
}

func (h *Client) handleCommand(dest nethelper.Destination, cmd protocol.ResponseCommand) {
	switch typedCommand := cmd.(type) {
	case *protocol.CommandSwitchAccount:
		if typedCommand.Host == nil {
			typedCommand.Host = dest.Address
		}
		h.handleSwitchAccount(typedCommand)
	default:
	}
}

func (h *Client) HandlePacketConn(ctx context.Context, dst net.Destination, pc udp.PacketReaderWriter) error {
	d := dispatcher.NewPacketDispatcher(ctx, h, dispatcher.WithResponseCallback(func(packet *udp.Packet) {
		pc.WritePacket(packet)
	}))
	defer d.Close()

	for {
		packet, err := pc.ReadPacket()
		if err != nil {
			return err
		}

		d.DispatchPacket(packet.Target, packet.Payload)
	}
}

func (c *Client) ProxyDial(ctx context.Context,
	dst net.Destination, initialData buf.MultiBuffer) (i.FlowConn, error) {
	command := protocol.RequestCommandTCP
	if dst.Network == net.Network_UDP {
		command = protocol.RequestCommandUDP
	}
	if dst.Address.Family().IsDomain() && dst.Address.Domain() == mux.MuxCoolAddressDst.Domain() {
		command = protocol.RequestCommandMux
	}

	rec := c.ServerPicker.PickServer()
	account := rec.GetProtocolSetting().(*vmess.MemoryAccount)

	conn, err := c.Dialer.Dial(ctx, rec.Destination())
	if err != nil {
		return nil, fmt.Errorf("failed to dial, %w", err)
	}

	request := &protocol.RequestHeader{
		Version:  encoding.Version,
		Account:  account,
		Command:  command,
		Address:  dst.Address,
		Port:     dst.Port,
		Option:   protocol.RequestOptionChunkStream,
		Security: account.Security,
	}
	if request.Security == protocol.SecurityType_AES128_GCM || request.Security == protocol.SecurityType_NONE || request.Security == protocol.SecurityType_CHACHA20_POLY1305 {
		request.Option.Set(protocol.RequestOptionChunkMasking)
	}
	if shouldEnablePadding(request.Security) && request.Option.Has(protocol.RequestOptionChunkMasking) {
		request.Option.Set(protocol.RequestOptionGlobalPadding)
	}
	if request.Security == protocol.SecurityType_ZERO {
		request.Security = protocol.SecurityType_NONE
		request.Option.Clear(protocol.RequestOptionChunkStream)
		request.Option.Clear(protocol.RequestOptionChunkMasking)
	}
	if account.AuthenticatedLengthExperiment {
		request.Option.Set(protocol.RequestOptionAuthenticatedLength)
	}
	isAEAD := false
	if !aeadDisabled && len(account.AlterIDs) == 0 {
		isAEAD = true
	}
	hashkdf := hmac.New(sha256.New, []byte("VMessBF"))
	hashkdf.Write(account.ID.Bytes())

	behaviorSeed := crc64.Checksum(hashkdf.Sum(nil), crc64.MakeTable(crc64.ISO))

	session := encoding.NewClientSession(ctx, isAEAD, protocol.DefaultIDHash, int64(behaviorSeed))

	writer := buf.NewBufferedWriter(buf.NewWriter(conn))
	if err := session.EncodeRequestHeader(request, writer); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to encode request, %w", err)
	}

	bodyWriter, err := session.EncodeRequestBody(request, writer)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to start encoding, %w", err)
	}
	if initialData.Len() > 0 {
		if err := bodyWriter.WriteMultiBuffer(initialData); err != nil {
			conn.Close()
			return nil, err
		}
	}
	if err := writer.SetBuffered(false); err != nil {
		conn.Close()
		return nil, err
	}

	return &clientBufConn{
		conn:              conn,
		Writer:            bodyWriter,
		ctx:               ctx,
		client:            c,
		serverDestination: rec.Destination(),
		session:           session,
		request:           request,
		closeWrite:        request.Option.Has(protocol.RequestOptionChunkStream),
	}, nil
}

func (c *Client) ListenPacket(ctx context.Context, dst net.Destination) (udp.UdpConn, error) {
	return dispatcher.NewDispatcherToPacketConn(ctx, c), nil
}

type clientBufConn struct {
	conn net.Conn
	buf.Reader
	buf.Writer
	ctx               context.Context
	client            *Client
	serverDestination net.Destination
	session           *encoding.ClientSession
	request           *protocol.RequestHeader
	closeWrite        bool
}

func (c *clientBufConn) Close() error {
	return c.conn.Close()
}

func (c *clientBufConn) ReadMultiBuffer() (buf.MultiBuffer, error) {
	if c.Reader == nil {
		reader := &buf.BufferedReader{Reader: buf.NewReader(c.conn)}
		header, err := c.session.DecodeResponseHeader(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to read header, %w", err)
		}
		c.client.handleCommand(c.serverDestination, header.Command)
		bodyReader, err := c.session.DecodeResponseBody(c.ctx, c.request, reader)
		if err != nil {
			return nil, fmt.Errorf("failed to start encoding response, %w", err)
		}
		c.request = nil
		c.ctx = nil
		c.client = nil
		c.session = nil
		c.Reader = bodyReader
	}

	return c.Reader.ReadMultiBuffer()
}

func (c *clientBufConn) CloseWrite() error {
	if c.closeWrite {
		err := c.Writer.WriteMultiBuffer(buf.MultiBuffer{})
		if err != nil {
			return err
		}
	}
	return c.Writer.CloseWrite()
}

var (
	enablePadding = false
	aeadDisabled  = false
)

func shouldEnablePadding(s protocol.SecurityType) bool {
	return enablePadding || s == protocol.SecurityType_AES128_GCM || s == protocol.SecurityType_CHACHA20_POLY1305 || s == protocol.SecurityType_AUTO
}

func init() {
	const defaultFlagValue = "NOT_DEFINED_AT_ALL"

	paddingValue := platform.NewEnvFlag("v2ray.padding").GetValue(func() string { return defaultFlagValue })
	if paddingValue != defaultFlagValue {
		enablePadding = true
	}

	isAeadDisabled := platform.NewEnvFlag("v2ray.aead.disabled").GetValue(func() string { return defaultFlagValue })
	if isAeadDisabled == "true" {
		aeadDisabled = true
	}
}

// func (c *Client) Dial(ctx context.Context, dst net.Destination) (net.Conn, error) {
// 	return c.DialWithInitialData(ctx, dst, nil)
// }

// func (c *Client) DialWithInitialData(ctx context.Context, dst net.Destination, initialData []byte) (net.Conn, error) {
// 	command := protocol.RequestCommandTCP
// 	if dst.Network == net.Network_UDP {
// 		command = protocol.RequestCommandUDP
// 	}
// 	if dst.Address.Family().IsDomain() && dst.Address.Domain() == mux.MuxCoolAddressDst.Domain() {
// 		command = protocol.RequestCommandMux
// 	}

// 	rec := c.ServerPicker.PickServer()
// 	account := rec.GetProtocolSetting().(*vmess.MemoryAccount)

// 	conn, err := c.Dialer.Dial(ctx, rec.Destination())
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to dial, %w", err)
// 	}

// 	request := &protocol.RequestHeader{
// 		Version:  encoding.Version,
// 		Account:  account,
// 		Command:  command,
// 		Address:  dst.Address,
// 		Port:     dst.Port,
// 		Option:   protocol.RequestOptionChunkStream,
// 		Security: account.Security,
// 	}
// 	if request.Security == protocol.SecurityType_AES128_GCM || request.Security == protocol.SecurityType_NONE || request.Security == protocol.SecurityType_CHACHA20_POLY1305 {
// 		request.Option.Set(protocol.RequestOptionChunkMasking)
// 	}
// 	if shouldEnablePadding(request.Security) && request.Option.Has(protocol.RequestOptionChunkMasking) {
// 		request.Option.Set(protocol.RequestOptionGlobalPadding)
// 	}
// 	if request.Security == protocol.SecurityType_ZERO {
// 		request.Security = protocol.SecurityType_NONE
// 		request.Option.Clear(protocol.RequestOptionChunkStream)
// 		request.Option.Clear(protocol.RequestOptionChunkMasking)
// 	}
// 	if account.AuthenticatedLengthExperiment {
// 		request.Option.Set(protocol.RequestOptionAuthenticatedLength)
// 	}
// 	isAEAD := false
// 	if !aeadDisabled && len(account.AlterIDs) == 0 {
// 		isAEAD = true
// 	}
// 	hashkdf := hmac.New(sha256.New, []byte("VMessBF"))
// 	hashkdf.Write(account.ID.Bytes())

// 	behaviorSeed := crc64.Checksum(hashkdf.Sum(nil), crc64.MakeTable(crc64.ISO))

// 	session := encoding.NewClientSession(ctx, isAEAD, protocol.DefaultIDHash, int64(behaviorSeed))

// 	writer := buf.NewBufferedWriter(buf.NewWriter(conn))
// 	if err := session.EncodeRequestHeader(request, writer); err != nil {
// 		conn.Close()
// 		return nil, fmt.Errorf("failed to encode request, %w", err)
// 	}

// 	bodyWriter, err := session.EncodeRequestBody1(request, writer)
// 	if err != nil {
// 		conn.Close()
// 		return nil, fmt.Errorf("failed to start encoding, %w", err)
// 	}
// 	if len(initialData) > 0 {
// 		_, err := bodyWriter.Write(initialData)
// 		if err != nil {
// 			conn.Close()
// 			return nil, err
// 		}
// 	}
// 	if err := writer.SetBuffered(false); err != nil {
// 		conn.Close()
// 		return nil, err
// 	}

// 	return &clientConn{
// 		Conn:              conn,
// 		Writer:            bodyWriter,
// 		session:           session,
// 		request:           request,
// 		client:            c,
// 		serverDestination: rec.Destination(),
// 		closeWrite:        request.Option.Has(protocol.RequestOptionChunkStream),
// 	}, nil
// }

// type clientConn struct {
// 	net.Conn
// 	io.Reader
// 	io.Writer
// 	session           *encoding.ClientSession
// 	serverDestination net.Destination
// 	client            *Client
// 	request           *protocol.RequestHeader
// 	closeWrite        bool
// }

// func (c *clientConn) Write(b []byte) (int, error) {
// 	return c.Writer.Write(b)
// }

// func (c *clientConn) Read(b []byte) (int, error) {
// 	if c.Reader == nil {
// 		header, err := c.session.DecodeResponseHeader(c.Conn)
// 		if err != nil {
// 			return 0, fmt.Errorf("failed to read header, %w", err)
// 		}
// 		c.client.handleCommand(c.serverDestination, header.Command)
// 		bodyReader, err := c.session.DecodeResponseBody1(c.request, c.Conn)
// 		if err != nil {
// 			return 0, fmt.Errorf("failed to start encoding response, %w", err)
// 		}
// 		c.request = nil
// 		c.session = nil
// 		c.client = nil
// 		c.Reader = bodyReader
// 	}

// 	return c.Reader.Read(b)
// }

// func (c *clientConn) CloseWrite() error {
// 	if c.closeWrite {
// 		_, err := c.Writer.Write([]byte{})
// 		return err
// 	}
// 	return nil
// }
