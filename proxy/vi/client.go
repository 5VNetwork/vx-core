package vi

import (
	"context"
	"fmt"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/dispatcher"
	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/common/protocol"
	"github.com/5vnetwork/vx-core/common/uuid"
	"github.com/5vnetwork/vx-core/common/vision"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/proxy/helper"

	"github.com/rs/zerolog/log"
)

type Client struct {
	serverPicker protocol.ServerPicker
	// serverTuple3 net.Destination
	// secret uuid.UUID
	dialer i.Dialer
}

// New creates a new VLess outbound handler.
func NewClient() *Client {
	handler := &Client{}
	return handler
}

// func (h *Client) WithDestination(dest net.Destination) *Client {
// 	h.serverTuple3 = dest
// 	return h
// }

func (h *Client) WithServerPicker(p protocol.ServerPicker) *Client {
	h.serverPicker = p
	return h
}

// TODO
func (c *Client) HandlePacketConn(ctx context.Context, dst net.Destination, pc udp.PacketReaderWriter) error {
	d := dispatcher.NewPacketDispatcher(ctx, c, dispatcher.WithResponseCallback(func(packet *udp.Packet) {
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

// Process implements proxy.Outbound.Process().
func (h *Client) HandleFlow(ctx context.Context, dst net.Destination, rw buf.ReaderWriter) error {
	dialer := h.dialer
	sp := h.serverPicker.PickServer()
	serverDest := sp.Destination()
	secret := sp.GetProtocolSetting().(uuid.UUID)
	conn, err := dialer.Dial(ctx, serverDest)
	if err != nil {
		return fmt.Errorf("failed to find an available destination, %w", err)
	}
	defer conn.Close()
	log.Ctx(ctx).Debug().Str("laddr", conn.LocalAddr().String()).Msg("vi dial ok")

	target := dst
	if !target.IsValid() {
		return errors.New("target not specified")
	}

	vConn := vision.NewVisionConn(ctx, conn, true, 0)
	defer vConn.Close()

	// read first payload and write request header and it
	b, err := EncodeHeader(Version, secret[:], target, uuid.New())
	if err != nil {
		conn.Close()
		return errors.New("failed to encode request header").Base(err)
	}
	_, err = vConn.Write(b.Bytes())
	if err != nil {
		conn.Close()
		return errors.New("failed to write request header").Base(err)
	}

	serverWriter := buf.NewWriter(vConn)
	serverReader := buf.NewReader(vConn)

	// tLink := link.(buf.TimeoutReader)
	// mb, err := tLink.ReadMultiBufferTimeout(time.Millisecond * 500)
	// if err != nil && !errors.Is(err, buf.ErrReadTimeout) {
	// 	return errors.New("failed to read first payload").Base(err)
	// }
	/* 	bufferWriter.WriteMultiBuffer(mb)
	   	if err := bufferWriter.SetBuffered(false); err != nil {
	   		return errors.New("failed to write first payload").Base(err)
	   	} */

	if target.Network == net.Network_UDP {
		serverReader = buf.NewLengthPacketReader(vConn)
		serverWriter = buf.NewMultiLengthPacketWriter(serverWriter)
	}
	if err := serverWriter.WriteMultiBuffer(nil); err != nil {
		return errors.New("failed to write first payload").Base(err)
	}

	// account := request.User.(*vless.User)
	err = helper.Relay(ctx, rw, rw, serverReader, serverWriter)
	if err != nil {
		return fmt.Errorf("failed to relay: %w", err)
	}
	return nil
}
