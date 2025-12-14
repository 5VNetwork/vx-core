package trojan

import (
	"context"
	"fmt"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/common/vision"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/proxy"
	"github.com/5vnetwork/vx-core/proxy/helper"
	"github.com/rs/zerolog/log"
)

// Client is an inbound handler for trojan protocol
type Client struct {
	ClientSettings
}

type ClientSettings struct {
	Address    net.Address
	PortPicker i.PortSelector
	Account    *MemoryAccount
	Dialer     i.Dialer
	Vision     bool
}

// NewClient create a new trojan client.
func NewClient(settings ClientSettings) *Client {
	client := &Client{
		ClientSettings: settings,
	}
	return client
}

func (c *Client) HandleFlow(ctx context.Context, dst net.Destination, rw buf.ReaderWriter) error {
	target := dst
	if !target.IsValid() {
		return errors.New("target not specified")
	}
	network := target.Network

	port := c.PortPicker.SelectPort()

	dest := net.TCPDestination(c.Address, net.Port(port))
	account := c.Account
	conn, err := c.Dialer.Dial(ctx, dest)
	if err != nil {
		return fmt.Errorf("failed to find an available destination, %w", err)
	}
	if c.Vision {
		conn = vision.NewVisionConn(ctx, conn, true, c.headerLen(dst))
	}
	defer conn.Close()

	if conn.LocalAddr() != nil && conn.RemoteAddr() != nil {
		log.Ctx(ctx).Debug().Str("laddr", conn.LocalAddr().String()).Str("raddr", conn.RemoteAddr().String()).Msg("trojan dialed ok")
	} else {
		log.Ctx(ctx).Warn().Msg("trojan dialed ok, but addr is nil")
	}

	var bodyWriter buf.Writer
	bufferWriter := buf.NewBufferedWriter(buf.NewWriter(conn))
	connWriter := &ConnWriter{Writer: bufferWriter, Target: target, Account: account}

	if target.Network == net.Network_UDP {
		bodyWriter = &PacketWriter{writer: connWriter, client: true, Dest: target}
	} else {
		bodyWriter = connWriter
	}

	err = buf.CopyOnceTimeout(rw, bodyWriter, proxy.FirstPayloadTimeout)
	if err != nil {
		if err == buf.ErrNotTimeoutReader || err == buf.ErrReadTimeout {
			if err := connWriter.WriteHeader(); err != nil {
				return fmt.Errorf("failed to write request header: %w", err)
			}
		} else {
			return fmt.Errorf("copy once timeout: %w", err)
		}
	}
	if err = bufferWriter.SetBuffered(false); err != nil {
		return errors.New("failed to flush payload").Base(err)
	}
	var reader buf.Reader
	if network == net.Network_UDP {
		reader = &PacketReader{
			reader: conn,
			client: true,
		}
	} else {
		reader = buf.NewReader(conn)
	}

	err = helper.Relay(ctx, rw, rw, reader, bodyWriter)
	if err != nil {
		return fmt.Errorf("failed to relay: %w", err)
	}
	return nil
}

func (c Client) HandlePacketConn(ctx context.Context, dst net.Destination, pc udp.PacketReaderWriter) error {
	port := c.PortPicker.SelectPort()
	dest := net.TCPDestination(c.Address, net.Port(port))
	account := c.Account
	conn, err := c.Dialer.Dial(ctx, dest)
	if err != nil {
		return fmt.Errorf("failed to find an available destination, %w", err)
	}
	if c.Vision {
		conn = vision.NewVisionConn(ctx, conn, true, c.headerLen(dst))
	}

	defer conn.Close()
	// bufferWriter := buf.NewBufferedWriter(buf.NewWriter(conn))
	connWriter := &ConnWriter{
		Writer:  conn,
		Account: account,
		Target:  dst,
	}
	return helper.RelayUDPPacketConn(ctx, pc, &udp.PacketRW{
		PacketWriter: &PacketWriter{writer: connWriter, client: true},
		PacketReader: &PacketReader{reader: conn, client: true},
	})
}

func (c *Client) ProxyDial(ctx context.Context, dst net.Destination,
	initialData buf.MultiBuffer) (i.FlowConn, error) {
	target := dst
	if !target.IsValid() {
		return nil, errors.New("target not specified")
	}
	network := target.Network

	port := c.PortPicker.SelectPort()
	dest := net.TCPDestination(c.Address, net.Port(port))
	account := c.Account
	conn, err := c.Dialer.Dial(ctx, dest)
	if err != nil {
		return nil, fmt.Errorf("failed to find an available destination, %w", err)
	}
	if conn.LocalAddr() != nil && conn.RemoteAddr() != nil {
		log.Ctx(ctx).Debug().Str("laddr", conn.LocalAddr().String()).Str("raddr", conn.RemoteAddr().String()).Msg("trojan dialed ok")
	} else {
		log.Ctx(ctx).Warn().Msg("trojan dialed ok, but addr is nil")
	}

	var bodyWriter buf.Writer
	bufferWriter := buf.NewBufferedWriter(buf.NewWriter(conn))
	connWriter := &ConnWriter{Writer: bufferWriter, Target: target, Account: account}

	if target.Network == net.Network_UDP {
		bodyWriter = &PacketWriter{writer: connWriter, client: true, Dest: target}
	} else {
		bodyWriter = connWriter
	}

	if initialData.Len() > 0 {
		err = bodyWriter.WriteMultiBuffer(initialData)
	} else {
		err = connWriter.WriteHeader()
	}
	if err != nil {
		conn.Close()
		return nil, err
	}

	if err := bufferWriter.SetBuffered(false); err != nil {
		conn.Close()
		return nil, err
	}

	var reader buf.Reader
	if network == net.Network_UDP {
		reader = &PacketReader{
			reader: conn,
			client: true,
		}
	} else {
		reader = buf.NewReader(conn)
	}

	return proxy.NewFlowConn(proxy.FlowConnOption{
		Reader:      reader,
		Writer:      bodyWriter,
		Close:       conn.Close,
		SetDeadline: conn,
	}), nil
}

func (c *Client) ListenPacket(ctx context.Context, dst net.Destination) (udp.UdpConn, error) {
	port := c.PortPicker.SelectPort()
	dest := net.TCPDestination(c.Address, net.Port(port))
	account := c.Account
	conn, err := c.Dialer.Dial(ctx, dest)
	if err != nil {
		return nil, fmt.Errorf("failed to find an available destination, %w", err)
	}
	// bufferWriter := buf.NewBufferedWriter(buf.NewWriter(conn))
	connWriter := &ConnWriter{
		Writer:  conn,
		Account: account,
		Target:  dst,
	}
	return &udp.PacketRW{
		PacketWriter: &PacketWriter{writer: connWriter, client: true},
		PacketReader: &PacketReader{reader: conn, client: true},
		OnClose:      conn.Close,
	}, nil
}

func (c *Client) headerLen(target net.Destination) int {
	buffer := buf.StackNew()
	defer buffer.Release()

	command := commandTCP
	if target.Network == net.Network_UDP {
		command = commandUDP
	}

	if _, err := buffer.Write(c.Account.Key); err != nil {
		return 0
	}
	if _, err := buffer.Write(crlf); err != nil {
		return 0
	}
	if err := buffer.WriteByte(command); err != nil {
		return 0
	}
	if err := addrParser.WriteAddressPort(&buffer, target.Address, target.Port); err != nil {
		return 0
	}
	if _, err := buffer.Write(crlf); err != nil {
		return 0
	}
	return int(buffer.Len())
}
