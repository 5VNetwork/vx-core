package tunsorter

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/5vnetwork/vx-core/app/inbound"
	"github.com/5vnetwork/vx-core/app/inbound/gvisor/packetparse"
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/i"

	"github.com/rs/zerolog/log"
)

func NewTunSorter(tunWriter io.Writer, h i.Handler, ctx context.Context) *TunSorter {
	return &TunSorter{
		tunWriter:  tunWriter,
		dispatcher: h,
		ctx:        ctx,
	}
}

type TunSorter struct {
	tunWriter  io.Writer
	dispatcher i.Handler

	trackedConnections sync.Map
	ctx                context.Context
}

func (t *TunSorter) OnPacketReceived(b []byte) (n int, err error) {
	src, dst, data, err := packetparse.TryParseAsUDPPacket(b)
	if err != nil {
		return 0, err
	}
	conn := newTrackedUDPConnection(src, t)
	trackedConnection, loaded := t.trackedConnections.LoadOrStore(src.String(), conn)
	conn = trackedConnection.(*trackedUDPConnection)
	if !loaded {
		t.onNewConnection(conn)
	}
	conn.onNewPacket(dst, data)
	return len(b), nil
}

func (t *TunSorter) onNewConnection(connection *trackedUDPConnection) {
	connection.c = make(chan *udp.Packet, 100)
	ctx, cancel := inbound.GetCtx(context.Background(), connection.src, connection.src, "gvisor")
	go func() {
		err := t.dispatcher.HandlePacketConn(ctx, net.Destination{}, connection)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("failed to handle udp connection")
		}
		connection.Close()
		cancel(err)
	}()
}

func (t *TunSorter) onWritePacket(src net.Destination, dest net.Destination, data []byte) error {
	data, err := packetparse.TryConstructUDPPacket(src, dest, data)
	if err != nil {
		return fmt.Errorf("failed to construct udp packet: %w", err)
	}
	_, err = t.tunWriter.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write udp packet: %w", err)
	}
	return nil
}

func newTrackedUDPConnection(src net.Destination, tunSorter *TunSorter) *trackedUDPConnection {
	return &trackedUDPConnection{
		tunSorter: tunSorter,
		src:       src,
	}
}

type trackedUDPConnection struct {
	tunSorter *TunSorter
	src       net.Destination
	ctx       context.Context
	c         chan *udp.Packet
}

func (t *trackedUDPConnection) ReadPacket() (*udp.Packet, error) {
	p, ok := <-t.c
	if ok {
		return p, nil
	}
	return nil, io.EOF
}

func (t *trackedUDPConnection) WritePacket(p *udp.Packet) error {
	defer p.Release()
	return t.onWritePacket(p.Source, p.Payload.Bytes())
}

func (t *trackedUDPConnection) onNewPacket(dst net.Destination, data []byte) {
	p := &udp.Packet{
		Target:  dst,
		Source:  t.src,
		Payload: buf.FromBytes(data),
	}
	select {
	case t.c <- p:
	default:
		log.Ctx(t.ctx).Warn().Msg("udp packet queue full")
	}
}

func (t *trackedUDPConnection) onWritePacket(src net.Destination, data []byte) error {
	return t.tunSorter.onWritePacket(src, t.src, data)
}

func (t *trackedUDPConnection) Close() error {
	t.tunSorter.trackedConnections.Delete(t.src.String())
	return nil
}
