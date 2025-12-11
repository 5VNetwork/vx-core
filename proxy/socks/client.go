package socks

import (
	"context"
	"fmt"
	"time"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/common/protocol"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/proxy"
	"github.com/5vnetwork/vx-core/proxy/helper"

	"github.com/rs/zerolog/log"
)

type User struct {
	Name   string
	Secret string
}

type Client struct {
	ClientSettings
}

type ClientSettings struct {
	ServerDest     net.Destination
	User, Secret   string
	Policy         i.TimeoutSetting
	DelayAuthWrite bool
	DNS            i.IPResolver
	Dialer         i.Dialer
}

// config won'e be nil, but may be zero-value
func NewClient(settings *ClientSettings) *Client {
	c := &Client{
		ClientSettings: *settings,
	}
	return c
}

func (c *Client) handshake(ctx context.Context, request *protocol.RequestHeader, conn net.Conn) (*net.Destination, error) {
	if c.User != "" || c.Secret != "" {
		request.Account = &User{Name: c.User, Secret: c.Secret}
	}

	if err := conn.SetDeadline(time.Now().Add(c.Policy.HandshakeTimeout())); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("cannot set deadline on tcp conn")
	}

	var udpDst *net.Destination
	var err error
	udpDst, err = ClientHandshake(request, conn, conn, c.DelayAuthWrite)
	if err != nil {
		return nil, errors.New("failed to establish connection to server").Base(err)
	}
	if udpDst != nil {
		if udpDst.Address == net.AnyIP || udpDst.Address == net.AnyIPv6 {
			udpDst.Address = c.ServerDest.Address
		}
	}

	if err := conn.SetDeadline(time.Time{}); err != nil {
		return nil, errors.New("failed to clear deadline after handshake").Base(err)
	}
	return udpDst, nil
}

func (c *Client) HandleFlow(ctx context.Context, dst net.Destination, rw buf.ReaderWriter) error {
	conn, err := c.Dialer.Dial(ctx, c.ServerDest)
	if err != nil {
		return fmt.Errorf("cannot connected to a socks server, %w", err)
	}

	log.Ctx(ctx).Debug().Str("laddr", conn.LocalAddr().String()).Msg("socks client dial ok")
	defer conn.Close()

	request := &protocol.RequestHeader{
		Version: socks5Version,
		Command: protocol.RequestCommandTCP,
		Address: dst.Address,
		Port:    dst.Port,
	}
	if dst.Network == net.Network_UDP {
		request.Command = protocol.RequestCommandUDP
	}

	udpDst, err := c.handshake(ctx, request, conn)
	if err != nil {
		return errors.New("failed to establish connection to server").Base(err)
	}

	if dst.Network == net.Network_TCP {
		return helper.Relay(ctx, rw, rw, buf.NewReader(conn), buf.NewWriter(conn))
	} else if dst.Network == net.Network_UDP {
		// udpDst is not nil
		udpConn, err := c.Dialer.Dial(ctx, *udpDst)
		if err != nil {
			return err
		}
		defer udpConn.Close()
		log.Ctx(ctx).Debug().Str("laddr", udpConn.LocalAddr().String()).Msg("socks client dial ok")
		return helper.Relay(ctx, rw, rw, &UDPReader{reader: udpConn},
			&buf.SequentialWriter{Writer: NewUDPWriter(request, udpConn)},
		)
	}

	return nil
}

func (c *Client) HandlePacketConn(ctx context.Context, dst net.Destination, pc udp.PacketReaderWriter) error {
	target := dst
	if !target.IsValid() {
		return errors.New("socks.Client cannot process because there no outbound info in ctx")
	}

	conn, err := c.Dialer.Dial(ctx, c.ServerDest)
	if err != nil {
		return fmt.Errorf("cannot connected to a socks server, %w", err)
	}

	defer func() {
		if err := conn.Close(); err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("cannot close outbound conn")
		}
	}()

	request := &protocol.RequestHeader{
		Version: socks5Version,
		Command: protocol.RequestCommandUDP,
		Address: c.ServerDest.Address,
		Port:    c.ServerDest.Port,
	}

	udpDest, err := c.handshake(ctx, request, conn)
	if err != nil {
		return fmt.Errorf("failed to handshke. %w", err)
	}

	udpConn, err := c.Dialer.Dial(ctx, *udpDest)
	if err != nil {
		return err
	}
	defer udpConn.Close()

	log.Ctx(ctx).Debug().Str("laddr", udpConn.LocalAddr().String()).Msg("socks client dial ok")

	return helper.RelayUDPPacketConn(ctx, pc, &udp.PacketRW{
		PacketReader: &UDPReader{reader: udpConn},
		PacketWriter: &udp.WriteToerToPacketWriter{
			WriteToer: NewUDPWriter(request, udpConn),
		},
	})
}

func (c *Client) ProxyDial(ctx context.Context, dst net.Destination,
	initialData buf.MultiBuffer) (i.FlowConn, error) {
	target := dst
	if !target.IsValid() {
		return nil, errors.New("socks.Client cannot process because there no outbound info in ctx")
	}

	conn, err := c.Dialer.Dial(ctx, c.ServerDest)
	if err != nil {
		return nil, fmt.Errorf("cannot connected to a socks server, %w", err)
	}

	log.Ctx(ctx).Debug().Str("laddr", conn.LocalAddr().String()).Msg("socks client dial ok")

	request := &protocol.RequestHeader{
		Version: socks5Version,
		Command: protocol.RequestCommandTCP,
		Address: dst.Address,
		Port:    dst.Port,
	}
	if target.Network == net.Network_UDP {
		request.Command = protocol.RequestCommandUDP
	}

	udpDst, err := c.handshake(ctx, request, conn)
	if err != nil {
		conn.Close()
		return nil, errors.New("failed to establish connection to server").Base(err)
	}

	var ret i.FlowConn
	if dst.Network == net.Network_TCP {
		ret = proxy.NewFlowConn(
			proxy.FlowConnOption{
				Reader:      buf.NewReader(conn),
				Writer:      buf.NewWriter(conn),
				Close:       conn.Close,
				SetDeadline: conn,
			})
	} else if dst.Network == net.Network_UDP {
		// udpDst is not nil
		udpConn, err := c.Dialer.Dial(ctx, *udpDst)
		if err != nil {
			return nil, err
		}
		log.Ctx(ctx).Debug().Str("laddr", udpConn.LocalAddr().String()).Msg("socks client dial ok")
		ret = proxy.NewFlowConn(
			proxy.FlowConnOption{
				Reader:      &UDPReader{reader: udpConn},
				Writer:      &buf.SequentialWriter{Writer: NewUDPWriter(request, udpConn)},
				Close:       udpConn.Close,
				SetDeadline: udpConn,
			})
	} else {
		return nil, errors.New("invalid network type")
	}
	if initialData.Len() > 0 {
		err = ret.WriteMultiBuffer(initialData)
		if err != nil {
			ret.Close()
			return nil, err
		}
	}
	return ret, nil
}

func (c *Client) ListenPacket(ctx context.Context, dst net.Destination) (udp.UdpConn, error) {
	target := dst
	if !target.IsValid() {
		return nil, errors.New("socks.Client cannot process because there no outbound info in ctx")
	}

	conn, err := c.Dialer.Dial(ctx, c.ServerDest)
	if err != nil {
		return nil, fmt.Errorf("cannot connected to a socks server, %w", err)
	}

	request := &protocol.RequestHeader{
		Version: socks5Version,
		Command: protocol.RequestCommandUDP,
		Address: c.ServerDest.Address,
		Port:    c.ServerDest.Port,
	}

	udpDest, err := c.handshake(ctx, request, conn)
	if err != nil {
		return nil, fmt.Errorf("failed to handshke. %w", err)
	}

	udpConn, err := c.Dialer.Dial(ctx, *udpDest)
	if err != nil {
		return nil, err
	}

	log.Ctx(ctx).Debug().Str("laddr", udpConn.LocalAddr().String()).Msg("socks client dial ok")

	return &udp.PacketRW{
		PacketReader: &UDPReader{reader: udpConn},
		PacketWriter: &udp.WriteToerToPacketWriter{
			WriteToer: NewUDPWriter(request, udpConn),
		},
		OnClose: func() error {
			conn.Close()
			return udpConn.Close()
		},
	}, nil
}
