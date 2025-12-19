package gvisor

import (
	"context"
	"io"
	"sync"

	"github.com/5vnetwork/vx-core/app/inbound"
	tun_net "github.com/5vnetwork/vx-core/app/inbound/gvisor/net"
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/common/task"

	"github.com/rs/zerolog/log"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	gvisor_udp "gvisor.dev/gvisor/pkg/tcpip/transport/udp"
	"gvisor.dev/gvisor/pkg/waiter"
)

type udpConn struct {
	*gonet.UDPConn
	id stack.TransportEndpointID
}

func (c *udpConn) ID() *stack.TransportEndpointID {
	return &c.id
}

func SetUDPHandler(in *TunGvisorInbound) StackOption {
	return func(s *stack.Stack) error {
		udpForwarder := gvisor_udp.NewForwarder(s,
			// this function is called on every udp packet that does not belongs to an exsiting flow
			func(r *gvisor_udp.ForwarderRequest) {
				wg := new(waiter.Queue)
				// this registers a flow, packets belong to the flow will not be given to this callback anymore
				linkedEndpoint, err := r.CreateEndpoint(wg)
				if err != nil {
					// errors.New("failed to create endpoint: ", err).WriteToLog(session.ExportIDToError(ctx))
					log.Error().Str("err", err.String()).Msg("failed to create endpoint")
					return
				}
				conn := &udpConn{
					UDPConn: gonet.NewUDPConn(wg, linkedEndpoint),
					id:      r.ID(),
				}

				go in.HandleUdp(conn)
			})
		s.SetTransportProtocolHandler(gvisor_udp.ProtocolNumber, udpForwarder.HandlePacket)
		return nil
	}
}

func (h *TunGvisorInbound) HandleUdp(conn tun_net.UDPConn) {
	defer conn.Close()
	id := conn.ID()

	dest := net.UDPDestination(tun_net.AddressFromTCPIPAddr(id.LocalAddress), net.Port(id.LocalPort))
	src := net.UDPDestination(tun_net.AddressFromTCPIPAddr(id.RemoteAddress), net.Port(id.RemotePort))

	if dest.Port == 53 || dest.Port == 443 {
		h.HandleUdpFlow(conn, src, dest)
	} else {
		h.HandleUdpPacketConn(conn, src, dest)
	}
}

func (h *TunGvisorInbound) HandleUdpFlow(conn tun_net.UDPConn, src, dest net.Destination) {
	ctx, cancel := inbound.GetCtx(src, dest, h.Tag())
	err := h.option.Handler.HandleFlow(ctx, dest, buf.NewRWD(buf.NewReader(conn), buf.NewWriter(conn), conn))
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to handle UDP connection")
	}
	cancel(err)
}

func (h *TunGvisorInbound) HandleUdpPacketConn(conn tun_net.UDPConn, src, dest net.Destination) {
	log.Debug().Str("src", src.String()).Str("dest", dest.String()).Msg("udp packet conn")

	var ctx context.Context
	var cancel context.CancelCauseFunc

	h.Lock()
	s, ok := h.udpSession[src]
	if !ok {
		ctx, cancel = inbound.GetCtx(src, dest, h.Tag())
		l1, l2 := udp.NewLink(10)
		s = &udpSession0{
			ctx:     ctx,
			src:     src,
			link:    l1,
			dstToRw: make(map[net.Destination]buf.ReaderWriter),
		}
		h.udpSession[src] = s
		go func() {
			err := h.option.Handler.HandlePacketConn(ctx, dest, l2)
			if err != nil {
				log.Ctx(ctx).Error().Err(err).Msg("failed to handle packet conn")
			}
			h.deleteSession(s)
		}()
		go s.handleResponse()
	}
	h.Unlock()

	rw := buf.NewRW(buf.NewReader(conn), buf.NewWriter(conn))
	s.addRW(dest, rw)
	defer s.removeRW(dest)

	err := task.Run(s.ctx, func() error {
		for {
			mb, err := rw.ReadMultiBuffer()
			log.Ctx(s.ctx).Debug().Int("len", int(mb.Len())).Msg("udp session0: read packets")
			for mb.Len() > 0 {
				b := mb[0]
				mb = mb[1:]
				err := s.link.WritePacket(&udp.Packet{
					Target:  dest,
					Source:  src,
					Payload: b,
				})
				if err != nil {
					buf.ReleaseMulti(mb)
					log.Ctx(s.ctx).Error().Err(err).Msg("failed to write packet to l1")
					return err
				}
			}
			if err != nil {
				if !errors.Is(err, io.EOF) {
					log.Ctx(s.ctx).Error().Err(err).Msg("failed to read packets from rw")
				}
				return err
			}
		}
	})
	cancel(err)
}

func (m *TunGvisorInbound) deleteSession(s *udpSession0) {
	m.Lock()
	defer m.Unlock()
	delete(m.udpSession, s.src)
	s.close()
}

type udpSession0 struct {
	ctx context.Context

	sync.RWMutex
	dstToRw map[net.Destination]buf.ReaderWriter

	src  net.Destination
	link *udp.PacketLink

	closed bool
}

func (s *udpSession0) close() {
	s.Lock()
	defer s.Unlock()
	if s.closed {
		return
	}
	s.closed = true
	s.link.Close()
}

func (s *udpSession0) addRW(addr net.Destination, rw buf.ReaderWriter) {
	s.Lock()
	defer s.Unlock()
	s.dstToRw[addr] = rw
}

func (s *udpSession0) removeRW(addr net.Destination) {
	s.Lock()
	defer s.Unlock()
	delete(s.dstToRw, addr)
}

func (s *udpSession0) handleResponse() {
	for {
		if s.closed {
			return
		}
		p, err := s.link.ReadPacket()
		if err != nil {
			log.Ctx(s.ctx).Error().Err(err).Msg("failed to read packet from link")
			return
		}
		s.RLock()
		rw, ok := s.dstToRw[p.Source]
		s.RUnlock()

		if ok {
			err := rw.WriteMultiBuffer(buf.MultiBuffer{p.Payload})
			if err != nil {
				s.removeRW(p.Source)
				log.Ctx(s.ctx).Error().Err(err).Msg("failed to write to rw")
			} else {
				log.Ctx(s.ctx).Debug().Str("source", p.Source.String()).Msg("udp session0: write to rw")
			}
		} else {
			log.Ctx(s.ctx).Debug().Str("source", p.Source.String()).Msg("udp session0: no rw found")
			p.Release()
		}
	}
}
