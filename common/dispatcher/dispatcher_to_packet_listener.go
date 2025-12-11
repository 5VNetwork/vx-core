package dispatcher

import (
	"context"
	"sync"

	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/i"
	"github.com/rs/zerolog/log"
)

type DispatcherToPacketConn struct {
	dispatcher      *PacketDispatcher
	responsePackets chan *udp.Packet

	closed bool
	lock   sync.Mutex
}

func NewDispatcherToPacketConn(ctx context.Context,
	dispatcher i.FlowHandler, opts ...PacketDispatcherOption) *DispatcherToPacketConn {
	d := &DispatcherToPacketConn{
		responsePackets: make(chan *udp.Packet, 100),
	}
	opts = append(opts, WithResponseCallback(func(packet *udp.Packet) {
		d.lock.Lock()
		defer d.lock.Unlock()
		if d.closed {
			return
		}
		select {
		case d.responsePackets <- packet:
		default:
			log.Ctx(ctx).Error().Msg("response packets channel is full")
		}
	}))
	d.dispatcher = NewPacketDispatcher(ctx, dispatcher, opts...)
	return d
}

func (d *DispatcherToPacketConn) WritePacket(p *udp.Packet) error {
	return d.dispatcher.DispatchPacket(p.Target, p.Payload)
}

func (d *DispatcherToPacketConn) ReadPacket() (*udp.Packet, error) {
	p, ok := <-d.responsePackets
	if !ok {
		return nil, errors.New("response packets channel closed")
	}
	return p, nil
}

func (d *DispatcherToPacketConn) Close() error {
	d.lock.Lock()
	defer d.lock.Unlock()
	if d.closed {
		return nil
	}
	d.closed = true
	close(d.responsePackets)
	for p := range d.responsePackets {
		p.Release()
	}
	d.dispatcher.Close()
	return nil
}
