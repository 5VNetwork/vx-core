package anytls

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"time"

	"anytls/proxy/session"
	as "anytls/proxy/session"

	"anytls/proxy/padding"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/common/serial/address_parser"
	"github.com/5vnetwork/vx-core/common/uot"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/proxy/helper"
)

type Client struct {
	ClientConfig
	secret        [32]byte
	sessionClient *as.Client
}

type ClientConfig struct {
	Address    net.Address
	PortPicker i.PortSelector
	Password   string
	Dialer     i.Dialer
}

func NewClient(config *ClientConfig) *Client {
	c := &Client{
		ClientConfig: *config,
	}
	c.secret = sha256.Sum256([]byte(config.Password))
	c.sessionClient = session.NewClient(context.Background(),
		c.createOutboundConnection, &padding.DefaultPaddingFactory,
		time.Second*30, time.Second*30, 5)
	return c
}

func (c *Client) HandleFlow(ctx context.Context, dst net.Destination, rw buf.ReaderWriter) error {
	conn, err := c.sessionClient.CreateStream(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	if dst.Network == net.Network_UDP {
		rw = uot.NewUotReaderWriter(rw, dst)
		dst = net.Destination{
			Address: uot.Addr,
			Port:    dst.Port,
		}
	}

	err = address_parser.SocksAddressSerializer.WriteAddressPort(conn, dst.Address, dst.Port)
	if err != nil {
		return err
	}

	return helper.Relay(ctx, rw, rw, buf.NewReader(conn), buf.NewWriter(conn))
}

func (c *Client) HandlePacketConn(ctx context.Context, dst net.Destination, rw udp.PacketReaderWriter) error {
	r := uot.NewUotPacketReaderWriter(rw, dst)
	dst = net.Destination{
		Address: uot.Addr,
		Network: net.Network_TCP,
	}
	return c.HandleFlow(ctx, dst, r)
}

func (c *Client) createOutboundConnection(ctx context.Context) (net.Conn, error) {
	conn, err := c.Dialer.Dial(ctx,
		net.TCPDestination(c.Address, net.Port(c.PortPicker.SelectPort())))
	if err != nil {
		return nil, err
	}

	b := buf.New()
	defer b.Release()

	b.Write(c.secret[:])
	var paddingLen int
	if pad := padding.DefaultPaddingFactory.Load().GenerateRecordPayloadSizes(0); len(pad) > 0 {
		paddingLen = pad[0]
	}
	binary.BigEndian.PutUint16(b.Extend(2), uint16(paddingLen))
	if paddingLen > 0 {
		clear(b.Extend(int32(paddingLen)))
	}

	_, err = conn.Write(b.Bytes())
	if err != nil {
		conn.Close()
		return nil, err
	}

	return conn, nil
}
