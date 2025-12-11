package gvisor

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/5vnetwork/vx-core/app/inbound/channel"
	"github.com/5vnetwork/vx-core/app/inbound/inboundcommon"
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/5vnetwork/vx-core/common/net/gtcpip"
	"github.com/5vnetwork/vx-core/tun"

	"github.com/rs/zerolog/log"
	"gvisor.dev/gvisor/pkg/buffer"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

type FilterLinkEndpointToTunDevice struct {
	*FilterLinkEndpoint
	name string
}

func NewFilterLinkEndpointToRunnable(linkEndpoint *FilterLinkEndpoint, name string) *FilterLinkEndpointToTunDevice {
	return &FilterLinkEndpointToTunDevice{
		FilterLinkEndpoint: linkEndpoint,
		name:               name,
	}
}

func (l *FilterLinkEndpointToTunDevice) Start() error {
	return nil
}

func (l *FilterLinkEndpointToTunDevice) Close() error {
	return nil
}

func (l *FilterLinkEndpointToTunDevice) Name() string {
	return l.name
}

type FilterLinkEndpoint struct {
	stack.LinkEndpoint
	dispatcher stack.NetworkDispatcher
	reject     inboundcommon.Rejector
	// packetwriter tun.Tun
	retainUdp  bool
	udpChannel chan *buf.Buffer
	once       sync.Once
}

func NewFilterLinkEndpoint(linkEndpoint stack.LinkEndpoint, reject inboundcommon.Rejector, retainUdp bool) *FilterLinkEndpoint {
	f := &FilterLinkEndpoint{
		LinkEndpoint: linkEndpoint,
		reject:       reject,
		retainUdp:    retainUdp,
	}
	if retainUdp {
		f.udpChannel = make(chan *buf.Buffer, 100)
	}
	return f
}

func (l *FilterLinkEndpoint) Attach(dispatcher stack.NetworkDispatcher) {
	log.Info().Interface("dispatcher", dispatcher).Msg("attach")
	l.dispatcher = dispatcher
	l.LinkEndpoint.Attach(l)
}

func (f *FilterLinkEndpoint) DeliverLinkPacket(protocol tcpip.NetworkProtocolNumber, pkt *stack.PacketBuffer) {
	f.dispatcher.DeliverLinkPacket(protocol, pkt)
}

// TODO: use dns conn
func (d *FilterLinkEndpoint) DeliverNetworkPacket(protocol tcpip.NetworkProtocolNumber, pkt *stack.PacketBuffer) {
	if pkt.TransportProtocolNumber == header.ICMPv4ProtocolNumber || pkt.TransportProtocolNumber == header.ICMPv6ProtocolNumber {
		log.Info().Msg("drop icmp packet")
		return
	}

	// TODO: only pull up headers
	packet, ok := pkt.Data().PullUp(pkt.Size())
	if !ok {
		log.Error().Msg("failed to pull up packet")
	} else {
		reject := d.reject.Reject(packet)
		if reject != nil {
			defer reject.Release()
			rejectPkt := stack.NewPacketBuffer(stack.PacketBufferOptions{
				Payload: buffer.MakeWithData(reject.Bytes()),
			})
			defer rejectPkt.DecRef()
			// write returnPkt into LinkEndpoint
			pkts := stack.PacketBufferList{}
			pkts.PushBack(rejectPkt)
			// WritePkackets will increase the ref count f pkts
			n, err := d.LinkEndpoint.WritePackets(pkts)
			if err != nil {
				log.Error().Str("err", err.String()).Msg("failed to write packet")
			}
			if n != 1 {
				log.Error().Int("n", n).Msg("failed to write packet")
			}
			return
		}

		if d.retainUdp {
			var isUdp bool
			if pkt.TransportProtocolNumber != 0 && pkt.TransportProtocolNumber == header.UDPProtocolNumber {
				isUdp = true
			} else {
				ipPacket := gtcpip.NewIPPacket(packet)
				if ipPacket != nil && ipPacket.TransportProtocol() == header.UDPProtocolNumber {
					isUdp = true
				}
			}
			if isUdp {
				b := buf.New()
				b.Write(packet)
				select {
				case d.udpChannel <- b:
				default:
					b.Release()
					log.Warn().Msg("channel is full")
				}
				return
			}
		}
	}
	if d.dispatcher == nil {
		log.Warn().Msg("dispatcher is nil")
		return
	}
	d.dispatcher.DeliverNetworkPacket(protocol, pkt)
}

func (d *FilterLinkEndpoint) ReadPacket() (*buf.Buffer, error) {
	b, ok := <-d.udpChannel
	if !ok {
		return nil, errors.ErrClosed
	}
	return b, nil
}

func (d *FilterLinkEndpoint) WritePacket(p *buf.Buffer) error {
	defer p.Release()
	pkt := stack.NewPacketBuffer(stack.PacketBufferOptions{
		Payload: buffer.MakeWithData(p.Bytes()),
	})
	defer pkt.DecRef()
	// write returnPkt into LinkEndpoint
	pkts := stack.PacketBufferList{}
	pkts.PushBack(pkt)
	// WritePkackets will increase the ref count f pkts
	n, err := d.LinkEndpoint.WritePackets(pkts)
	if err != nil {
		return fmt.Errorf("failed to write packet: %s", err.String())
	}
	if n != 1 {
		return fmt.Errorf("failed to write packet, got %d", n)
	}
	return nil
}

// Read packets from ReadWriter, send the packets to gvisor stack; meanwhile,
// read packets from gvisor stack, write the packets to ReadWriter
type TunLinkEndpoint struct {
	*channel.Endpoint
	prw    tun.TunDeviceWithInfo
	mtu    uint32
	closed bool
	reject inboundcommon.Rejector
}

type TunLinkEndpointOption func(*TunLinkEndpoint)

func TunLinkEndpointWithRejector(reject inboundcommon.Rejector) TunLinkEndpointOption {
	return func(e *TunLinkEndpoint) {
		e.reject = reject
	}
}

func NewTunLinkEndpoint(prw tun.TunDeviceWithInfo, mtu uint32, opts ...TunLinkEndpointOption) *TunLinkEndpoint {
	e := &TunLinkEndpoint{
		Endpoint: channel.New(1024, mtu, "", 0),
		prw:      prw,
		mtu:      mtu,
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

func (e *TunLinkEndpoint) Start() error {
	err := e.prw.Start()
	if err != nil {
		return err
	}
	go e.outbounds()
	go e.inbounds()
	return nil
}

func (e *TunLinkEndpoint) Close() {
	e.closed = true
	e.Endpoint.Close()
	e.prw.Close()
}

// Array copying is unavoidable.
// in terms of original app, the packets are outbound. For gvisor stack, these packets are inbound.
func (e *TunLinkEndpoint) outbounds() {
	for {
		p, err := e.prw.ReadPacket()
		if err != nil {
			if e.closed {
				return
			}
			log.Err(err).Send()
			return
		}

		if e.reject != nil {
			rejectPacket := e.reject.Reject(p.Bytes())
			if rejectPacket != nil {
				p.Release()
				err := e.prw.WritePacket(rejectPacket)
				if err != nil {
					log.Err(err).Msg("failed to write packet")
					return
				}
				continue
			}
		}

		if header.IPVersion(p.Bytes()) == header.IPv4Version {
			pkt := stack.NewPacketBuffer(stack.PacketBufferOptions{
				Payload:           buffer.MakeWithData(p.Bytes()),
				IsForwardedPacket: true,
			})
			e.InjectInbound(header.IPv4ProtocolNumber, pkt)
			pkt.DecRef()
		} else if header.IPVersion(p.Bytes()) == header.IPv6Version {
			pkt := stack.NewPacketBuffer(stack.PacketBufferOptions{
				Payload:           buffer.MakeWithData(p.Bytes()),
				IsForwardedPacket: true,
			})
			e.InjectInbound(header.IPv6ProtocolNumber, pkt)
			pkt.DecRef()
		}
		p.Release()
	}
}

func (e *TunLinkEndpoint) inbounds() {
	ctx := context.Background()
	for {
		pkt := e.ReadContext(ctx)
		if pkt == nil {
			log.Print("pkt is nil")
			return
		}
		e.writeResponse(pkt)
	}
}

func (e *TunLinkEndpoint) writeResponse(pkt *stack.PacketBuffer) {
	defer pkt.DecRef()

	v := pkt.ToView()
	p := buf.NewWithRelease(v.AsSlice(), v.Release)
	defer p.Release()

	if p.Len() == 0 {
		return
	}

	err := e.prw.WritePacket(p)
	if err != nil {
		log.Err(err).Send()
	}
}

// Read packets from ReadWriter, send the packets to gvisor stack; meanwhile,
// read packets from gvisor stack, write the packets to ReadWriter
// rw is owned by IOLinkEndpoint
type IOLinkEndpoint struct {
	*channel.Endpoint
	reject inboundcommon.Rejector
	rw     io.ReadWriteCloser
	mtu    uint32
	offset uint32
}

type IOLinkEndpointOption func(*IOLinkEndpoint)

func IOLinkEndpointWithOffset(offset uint32) IOLinkEndpointOption {
	return func(e *IOLinkEndpoint) {
		e.offset = offset
	}
}

func IOLinkEndpointWithMtu(mtu uint32) IOLinkEndpointOption {
	return func(e *IOLinkEndpoint) {
		e.mtu = mtu
	}
}

func IOLinkEndpointWithRejector(reject inboundcommon.Rejector) IOLinkEndpointOption {
	return func(e *IOLinkEndpoint) {
		e.reject = reject
	}
}

// rw's ownership is transfered
func NewIOLinkEndpoint(rw io.ReadWriteCloser, opts ...IOLinkEndpointOption) *IOLinkEndpoint {
	e := &IOLinkEndpoint{
		rw:  rw,
		mtu: 1500,
	}
	for _, opt := range opts {
		opt(e)
	}
	e.Endpoint = channel.New(1024, e.mtu, "", uint16(e.offset))
	return e
}

func (e *IOLinkEndpoint) Start() error {
	go e.outbounds()
	go e.inbounds()
	return nil
}

func (e *IOLinkEndpoint) Close() {
	e.Endpoint.Close()
	e.rw.Close()
}

// Hand outbound packets to gvisor stack, for the stack, these packets
// are inbound.
func (e *IOLinkEndpoint) outbounds() {
	b := make([]byte, e.mtu+e.offset)
	for {
		n, err := e.rw.Read(b)
		// TODO
		if err != nil && err != io.EOF && err != os.ErrClosed {
			log.Err(err).Send()
			return
		}
		if n == 0 {
			continue
		}

		if e.reject != nil {
			rejectPacket := e.reject.Reject(b[e.offset:n])
			if rejectPacket != nil {
				if e.offset > 0 {
					rejectPacket.RetreatStart(int32(e.offset))
					// TODO ipv4 case
					copy(rejectPacket.BytesRange(0, int32(e.offset)), ipv6FourBytes)
				}
				_, err := e.rw.Write(rejectPacket.Bytes())
				rejectPacket.Release()
				if err != nil {
					log.Err(err).Msg("failed to write packet")
					return
				}
				continue
			}
		}

		if header.IPVersion(b[e.offset:n]) == header.IPv4Version {
			pkt := stack.NewPacketBuffer(stack.PacketBufferOptions{
				Payload:           buffer.MakeWithData(b[e.offset:n]), //will copy data into a buffer managed by stack
				IsForwardedPacket: true,
			})
			e.InjectInbound(header.IPv4ProtocolNumber, pkt)
			pkt.DecRef()
		} else if header.IPVersion(b[e.offset:n]) == header.IPv6Version {
			pkt := stack.NewPacketBuffer(stack.PacketBufferOptions{
				Payload:           buffer.MakeWithData(b[e.offset:n]), //will copy data into a buffer managed by stack
				IsForwardedPacket: true,
			})
			e.InjectInbound(header.IPv6ProtocolNumber, pkt)
			pkt.DecRef()
		}
	}
}

var zeroFourBytes = []byte{0, 0, 0, 0}
var ipv4FourBytes = []byte{0, 0, 0, 2}
var ipv6FourBytes = []byte{0, 0, 0, 30}

// TODO efficient vector
func (e *IOLinkEndpoint) inbounds() {
	ctx := context.Background()
	for {
		pkt := e.ReadContext(ctx)
		if pkt == nil {
			log.Print("pkt is nil")
			return
		}
		if e.offset > 0 {
			s := pkt.LinkHeader().Push(int(e.offset))
			if pkt.NetworkProtocolNumber == header.IPv4ProtocolNumber {
				copy(s, ipv4FourBytes)
			} else {
				copy(s, ipv6FourBytes)
			}
		}
		v := pkt.ToView()
		_, err := e.rw.Write(v.AsSlice())
		v.Release()
		pkt.DecRef()
		if err != nil {
			log.Err(err).Msg("failed to write packet")
			return
		}
	}
}
