package udp

import (
	"context"
	"sync"

	"github.com/5vnetwork/vx-core/app/inbound"
	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/i"

	"github.com/rs/zerolog/log"
)

type UdpPacketHandler struct {
	tag               string
	packetConnHandler i.PacketHandler

	sync.RWMutex
	udpSrcToUdpSession map[net.Destination]*udpSessionFullCone

	writeResponse func(p *udp.Packet) error
}

func NewUdpPacketHandler(
	tag string,
	packetConnHandler i.PacketHandler,
	writeResponse func(p *udp.Packet) error,
) *UdpPacketHandler {
	return &UdpPacketHandler{
		tag:                tag,
		packetConnHandler:  packetConnHandler,
		writeResponse:      writeResponse,
		udpSrcToUdpSession: make(map[net.Destination]*udpSessionFullCone),
	}
}

func (t *UdpPacketHandler) HandleUdpPacket(p *udp.Packet) {
	t.Lock()
	defer t.Unlock()
	s, found := t.udpSrcToUdpSession[p.Source]
	if !found {
		ctx, cancel := inbound.GetCtx(context.Background(), p.Source, p.Target, t.tag)
		s = &udpSessionFullCone{
			cancelCause:   cancel,
			src:           p.Source,
			writeResponse: t.writeResponse,
			requestChan:   make(chan *udp.Packet, 50),
		}
		go func() {
			err := t.packetConnHandler.HandlePacketConn(ctx, p.Target, s)
			t.RemoveUdpSession(p.Source)
			if err != nil {
				log.Ctx(ctx).Error().Err(err).Msg("failed to handle udp session")
			}
		}()
		// }
		t.udpSrcToUdpSession[p.Source] = s
	}
	select {
	case s.requestChan <- p:
	default:
		log.Error().Msg("channel is full")
	}
}

func (u *UdpPacketHandler) RemoveUdpSession(src net.Destination) {
	u.Lock()
	s, found := u.udpSrcToUdpSession[src]
	if found {
		close(s.requestChan)
		for p := range s.requestChan {
			p.Release()
		}
		s.cancelCause(nil)
		delete(u.udpSrcToUdpSession, src)
	}
	u.Unlock()
}

type udpSessionFullCone struct {
	src           net.Destination
	cancelCause   context.CancelCauseFunc
	writeResponse func(p *udp.Packet) error
	requestChan   chan *udp.Packet
}

func (u *udpSessionFullCone) WritePacket(p *udp.Packet) error {
	p.Target = u.src
	return u.writeResponse(p)
}

func (u *udpSessionFullCone) ReadPacket() (*udp.Packet, error) {
	p, open := <-u.requestChan
	if !open {
		return nil, common.ErrClosed
	}
	return p, nil
}
