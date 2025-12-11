package system

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/netip"
	"sync"
	"time"

	"github.com/5vnetwork/vx-core/app/dns"
	"github.com/5vnetwork/vx-core/app/inbound"
	"github.com/5vnetwork/vx-core/app/inbound/inboundcommon"
	"github.com/5vnetwork/vx-core/app/inbound/udp_session"
	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/buf"
	mynet "github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/gtcpip"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/common/pipe"
	"github.com/5vnetwork/vx-core/common/signal/done"
	"github.com/5vnetwork/vx-core/common/strings"
	"github.com/5vnetwork/vx-core/i"

	"github.com/rs/zerolog/log"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/checksum"
	"gvisor.dev/gvisor/pkg/tcpip/header"
)

type tunDevice interface {
	common.Runnable
	Name() string
	ReadPacket() (*buf.Buffer, error)
	// WritePacket takes ownership of the buffer.
	// Buffer is released no matter success or not
	WritePacket(*buf.Buffer) error
}

type TunSystemInbound struct {
	dispatcher i.Handler
	tag        string
	tun        tunDevice
	// ipv4
	listenIp4   net.IP
	listenPort4 uint16
	natIp4      net.IP
	listener4   net.Listener
	// ipv6
	sixEnabled  bool
	listenIp6   net.IP
	listenPort6 uint16
	natIp6      net.IP
	listener6   net.Listener
	// tcp
	nat *nat
	// udp
	udpSessionManager *udp_session.SessionManager
	udpPackets        chan *buf.Buffer
	// dns dispatcher
	dnsConn    dnsConn
	dnsAddress []mynet.Destination

	rejector    inboundcommon.Rejector
	udpRejector inboundcommon.Rejector

	startOnce sync.Once
	closeOnce sync.Once
	// closed is set to true when Close is called.
	done *done.Instance
}

type dnsConn interface {
	// packet contains a dns message.
	// WritePacket takes ownership of the packet. Caller should not
	// use Packet after WritePacket.
	WritePacket(*udp.Packet) error
	// packet contains a dns message
	ReadPacket() (*udp.Packet, error)
}

type Option func(*TunSystemInbound)

func New(opts ...Option) *TunSystemInbound {
	t := &TunSystemInbound{
		tag:               "tun",
		nat:               NewNat(),
		done:              done.New(),
		udpSessionManager: udp_session.NewManager(),
		udpPackets:        make(chan *buf.Buffer, 1000),
	}

	for _, opt := range opts {
		opt(t)
	}
	return t
}

func WithHandler(h i.Handler) Option {
	return func(t *TunSystemInbound) {
		t.dispatcher = h
	}
}

func WithTun(tun tunDevice) Option {
	return func(t *TunSystemInbound) {
		t.tun = tun
	}
}

func WithTag(tag string) Option {
	return func(t *TunSystemInbound) {
		t.tag = tag
	}
}

func With4(natIP net.IP, listenIP net.IP, listenPort uint16) Option {
	return func(t *TunSystemInbound) {
		t.natIp4 = natIP
		t.listenIp4 = listenIP
		t.listenPort4 = listenPort
	}
}

func With6(natIP net.IP, listenIP net.IP, listenPort uint16) Option {
	return func(t *TunSystemInbound) {
		t.sixEnabled = true
		t.natIp6 = natIP
		t.listenIp6 = listenIP
		t.listenPort6 = listenPort
	}
}

func WithDns(dnsDispatcher dns.DnsConn, dnsAddress []mynet.Destination) Option {
	return func(t *TunSystemInbound) {
		t.dnsConn = dnsDispatcher
		t.dnsAddress = dnsAddress
	}
}

func WithRejector(rejector inboundcommon.Rejector) Option {
	return func(t *TunSystemInbound) {
		t.rejector = rejector
	}
}

func WithUdpRejector(rejector inboundcommon.Rejector) Option {
	return func(t *TunSystemInbound) {
		t.udpRejector = rejector
	}
}

func (ti *TunSystemInbound) Start() error {
	var err error
	ti.startOnce.Do(func() {
		ti.tun.Start()
		if ti.listenIp4 != nil {
			err = ti.listenTcp(net.JoinHostPort(ti.listenIp4.String(), strings.ToString(ti.listenPort4)), true)
			if err != nil {
				ti.Close()
				return
			}
			log.Printf("listenIp4: %v, listenPort4: %v", ti.listenIp4, ti.listenPort4)
		}
		if ti.listenIp6 != nil {
			err = ti.listenTcp(net.JoinHostPort(ti.listenIp6.String(), strings.ToString(ti.listenPort6)), false)
			if err != nil {
				ti.Close()
				return
			}
			log.Printf("listenIp6: %v, listenPort6: %v", ti.listenIp6, ti.listenPort6)
		}
		go ti.readPackets()
		go ti.handleUdp()
		if ti.dnsConn != nil {
			go ti.dnsResponse()
		}
	})
	return err
}

func (t *TunSystemInbound) WritePacket(p *udp.Packet) error {
	return t.tun.WritePacket(udp.UdpPacketToIpPacket(p))
}

func (t *TunSystemInbound) Close() error {
	var err error
	t.closeOnce.Do(func() {
		t.done.Close()

		if t.listener4 != nil {
			t.listener4.Close()
		}
		if t.listener6 != nil {
			t.listener6.Close()
		}

		t.udpSessionManager.Close()

		t.nat.Lock()
		for _, s := range t.nat.srcToSession {
			s.cancel()
		}
		t.nat.Unlock()

		err = t.tun.Close()
	})
	return err
}

// listenTcp listens on a tcp address and returns the port number that it is listening on.
func (t *TunSystemInbound) listenTcp(address string, ipv4 bool) error {
	l, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to listen on %s > %v", address, err)
	}
	log.Info().Str("addr", l.Addr().String()).Msg("tunInbound starts to listen tcp")

	if ipv4 {
		t.listener4 = l
		t.listenPort4 = uint16(l.Addr().(*net.TCPAddr).Port)
	} else {
		t.listener6 = l
		t.listenPort6 = uint16(l.Addr().(*net.TCPAddr).Port)
	}

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				if t.done.Done() {
					return
				}
				log.Error().Err(err).Str("addr", l.Addr().String()).Msg("listener failed to accept connection.")
				return
			}
			go t.handleConn(conn)
		}
	}()
	return nil
}

// handle a tcp connection
func (t *TunSystemInbound) handleConn(conn net.Conn) {
	s := t.nat.findNatSessionByNatport(uint16(conn.RemoteAddr().(*net.TCPAddr).Port))
	if s == nil {
		log.Error().Str("src", conn.RemoteAddr().String()).Str("dst", conn.LocalAddr().String()).Msg("failed to find a nat session")
		conn.Close()
		return
	}

	src := mynet.TCPDestination(
		mynet.IPAddress(s.src.Addr().AsSlice()),
		mynet.Port(s.src.Port()),
	)
	dst := mynet.TCPDestination(
		mynet.IPAddress(s.dst.Addr().AsSlice()),
		mynet.Port(s.dst.Port()),
	)
	gateway := mynet.TCPDestination(
		mynet.IPAddress(t.listenIp4),
		mynet.Port(t.listenPort4),
	)

	ctx, cancelCause := inbound.GetCtx(context.Background(), src, gateway, t.tag)
	ctx = inbound.ContextWithRawConn(ctx, conn)

	defer func() {
		s.cancel()
		conn.Close()
		// after conn is closed, there might still be packets in the closing stages of the tcp connection.
		select {
		case <-t.done.Wait():
			return
		case <-time.After(time.Second * 121):
			t.nat.removeNatSession(s)
			log.Ctx(ctx).Debug().Msg("removed nat session")
		}
	}()
	log.Ctx(ctx).Debug().Uint16("nat port", s.natPort).Send()

	bufConn := buf.NewRWD(buf.NewReader(conn), buf.NewWriter(conn), conn)
	err := t.dispatcher.HandleFlow(ctx, dst, bufConn)
	cancelCause(err)
	log.Ctx(ctx).Debug().Err(err).Msg("tcp session end")
}

func (t *TunSystemInbound) Tag() string {
	return t.tag
}

func (t *TunSystemInbound) readPackets() {
	for {
		b, err := t.tun.ReadPacket()
		if err != nil {
			if t.done.Done() {
				return
			}
			// if err == io.EOF {
			// 	log.Error().Msg("error eof")
			// 	continue
			// }
			log.Error().Err(err).Msg("failed to read packet from tun device")
			return
		}

		t.processIPPakcet(b)
	}
}

func (t *TunSystemInbound) processIPPakcet(p *buf.Buffer) {
	ipPacket := gtcpip.NewIPPacket(p.Bytes())
	if ipPacket == nil {
		log.Error().Msg("invalid ip packet")
		p.Release()
		return
	}
	isIpv4 := header.IPVersion(p.Bytes()) == header.IPv4Version
	switch ipPacket.TransportProtocol() {
	case header.ICMPv4ProtocolNumber:
		log.Debug().Msg("drop icmpv4 packet")
		p.Release()
		return
	case header.ICMPv6ProtocolNumber:
		log.Debug().Msg("drop icmpv6 packet")
		p.Release()
		return
	case header.TCPProtocolNumber:
		// check valid?
		tcpPacket := gtcpip.TCP{
			TCP: header.TCP(ipPacket.Payload()),
		}
		srcIP := mynet.NetIPFromGvisorTcpipAddress(ipPacket.SourceAddress())
		dstIP := mynet.NetIPFromGvisorTcpipAddress(ipPacket.DestinationAddress())
		srcPort := tcpPacket.SourcePort()
		dstPort := tcpPacket.DestinationPort()

		// packet is sent from listen address and port
		if srcIP.Equal(t.listenIp4) && srcPort == t.listenPort4 ||
			srcIP.Equal(t.listenIp6) && srcPort == t.listenPort6 {
			s := t.nat.findNatSessionByNatport(dstPort)
			if s == nil {
				log.Debug().Any("srcIP", srcIP).Uint16("src port", srcPort).Any("dstIP", dstIP).
					Uint16("dst port", dstPort).Msg("failed to find nat session")
				p.Release()
				return
			}
			ipPacket.SetSourceAddress(tcpip.AddrFromSlice(s.dst.Addr().AsSlice()))
			ipPacket.SetDestinationAddress(tcpip.AddrFromSlice(s.src.Addr().AsSlice()))
			tcpPacket.SetSourcePort(s.dst.Port())
			tcpPacket.SetDestinationPort(s.src.Port())
		} else {
			// populate natSession
			var s *natSession
			if tcpPacket.Flags().Contains(header.TCPFlagSyn) {
				if t.rejector != nil {
					rejectPacket := t.rejector.Reject(p.Bytes())
					if rejectPacket != nil {
						err := t.tun.WritePacket(rejectPacket)
						if err != nil {
							log.Error().Err(err).Msg("failed to write packet into tun")
						}
						p.Release()
						return
					}
				}
				s = t.nat.createNatSession(netip.AddrPortFrom(mynet.IPToNetipAddr(srcIP), srcPort),
					netip.AddrPortFrom(mynet.IPToNetipAddr(dstIP), dstPort))
			} else {
				var found bool
				if s, found = t.nat.getNatSession(
					netip.AddrPortFrom(mynet.IPToNetipAddr(srcIP), srcPort),
					netip.AddrPortFrom(mynet.IPToNetipAddr(dstIP), dstPort)); !found {
					log.Debug().Any("srcIP", srcIP).Uint16("src port", srcPort).Any("dstIP", dstIP).
						Uint16("dst port", dstPort).Msg("failed to find nat session")
					p.Release()
					return
				}
			}
			if isIpv4 {
				ipPacket.SetSourceAddress(tcpip.AddrFrom4Slice(t.natIp4))
				ipPacket.SetDestinationAddress(tcpip.AddrFrom4Slice(t.listenIp4))
				tcpPacket.SetDestinationPort(t.listenPort4)
			} else {
				if !t.sixEnabled {
					log.Warn().Msg("ipv6 is not enabled but got ipv6 packet")
					p.Release()
					return
				}
				ipPacket.SetSourceAddress(tcpip.AddrFrom16Slice(t.natIp6))
				ipPacket.SetDestinationAddress(tcpip.AddrFrom16Slice(t.listenIp6))
				tcpPacket.SetDestinationPort(t.listenPort6)
			}
			s.t = time.Now()
			tcpPacket.SetSourcePort(s.natPort)
		}
		// reset tcp checksum
		tcpPacket.ResetChecksum(ipPacket.PseudoHeaderChecksum())
		if !tcpPacket.IsChecksumValid(
			ipPacket.SourceAddress(),
			ipPacket.DestinationAddress(),
			checksum.Checksum(tcpPacket.Payload(), 0),
			uint16(len(tcpPacket.Payload()))) {
			log.Error().Msg("invalid tcp checksum")
			p.Release()
			return
		}
		// reset ip checksum
		ipPacket.ResetChecksum()
		if isIpv4 {
			if !ipPacket.(*gtcpip.IPv4).IsChecksumValid() {
				log.Error().Msg("invalid ip checksum")
				p.Release()
				return
			}
		}
		if err := t.tun.WritePacket(p); err != nil {
			if t.done.Done() {
				return
			}
			log.Error().Err(err).Msg("failed to write packet to tun device")
			return
		}
	case header.UDPProtocolNumber:
		if isDisallowed(ipPacket.DestinationAddress()) {
			if log.Debug().Enabled() {
				log.Debug().Str("dst", ipPacket.DestinationAddress().String()).Msg("drop udp packet")
			}
			p.Release()
			return
		}
		udpPacket := header.UDP(ipPacket.Payload())
		if len(udpPacket) < 8 {
			log.Error().Msg("invalid UDP packet")
			p.Release()
			return
		}
		select {
		case t.udpPackets <- p:
		default:
			p.Release()
			log.Error().Msg("udp packets channel is full")
		}
	default:
		p.Release()
	}

}

func (t *TunSystemInbound) handleUdp() {
	for {
		select {
		case <-t.done.Wait():
			return
		case p, open := <-t.udpPackets:
			if !open {
				return
			}
			t.handleUdpPacket(p)
		}
	}
}

func (t *TunSystemInbound) dnsResponse() {
	for {
		if t.done.Done() {
			return
		}
		p, err := t.dnsConn.ReadPacket()
		if err != nil && err != io.EOF {
			log.Error().Err(err).Msg("failed to read dns response")
			continue
		}
		if p != nil {
			t.WritePacket(p)
		}
	}
}

func (t *TunSystemInbound) handleUdpPacket(b *buf.Buffer) {
	packet := b.Bytes()
	ipPacket := gtcpip.NewIPPacket(packet)
	udpPacket := header.UDP(ipPacket.Payload())
	dst := mynet.UDPDestination(mynet.IPAddress(
		mynet.NetIPFromGvisorTcpipAddress(ipPacket.DestinationAddress())),
		mynet.Port(udpPacket.DestinationPort()))
	src := mynet.UDPDestination(mynet.IPAddress(
		mynet.NetIPFromGvisorTcpipAddress(ipPacket.SourceAddress())),
		mynet.Port(udpPacket.SourcePort()))
	b.AdvanceStart(int32(ipPacket.HeaderLength()) + 8)

	p := &udp.Packet{
		Source:  src,
		Target:  dst,
		Payload: b,
	}

	if t.dnsConn != nil && p.Target.Port == 53 {
		for _, dns := range t.dnsAddress {
			if p.Target == dns {
				err := t.dnsConn.WritePacket(p)
				if err != nil {
					log.Error().Err(err).Msg("failed to dispatch dns message")
				}
				return
			}
		}
	}

	s, found := t.udpSessionManager.GetUdpSession(p.Source)
	if !found {
		if t.udpRejector != nil {
			rejectPacket := t.udpRejector.Reject(packet)
			if rejectPacket != nil {
				err := t.tun.WritePacket(rejectPacket)
				if err != nil {
					log.Error().Err(err).Msg("failed to write packet into tun")
				}
				b.Release()
				return
			}
		}

		ctx, cancelCause := inbound.GetCtx(context.Background(), p.Source, p.Target, t.tag)

		if p.Target.Port != 443 {
			s = udp_session.NewUdpSessionFullCone(ctx,
				p.Source, cancelCause, t, make(chan *udp.Packet, 64))
			go func() {
				err := t.dispatcher.HandlePacketConn(ctx, p.Target, s.(udp.PacketReaderWriter))
				t.udpSessionManager.CloseAndRemoveUdpSession(p.Source)
				if err != nil && !errors.Is(err, common.ErrClosed) {
					log.Ctx(ctx).Debug().Err(err).Msg("failed to handle udpSessionFullCone")
				}
			}()
		} else {
			s = udp_session.NewUdpSessionSymetric(ctx, p.Source,
				p.Target, cancelCause, t, pipe.NewPipe(buf.BufferSize*32, true))
			go func() {
				err := t.dispatcher.HandleFlow(ctx, p.Target, s.(buf.ReaderWriter))
				t.udpSessionManager.CloseAndRemoveUdpSession(p.Source)
				if err != nil && !errors.Is(err, common.ErrClosed) {
					log.Ctx(ctx).Debug().Err(err).Msg("failed to handle udpSessionSymetric")
				}
			}()
		}
		t.udpSessionManager.AddUdpSession(p.Source, s)
	}

	if err := s.Dispatch(p); err != nil {
		log.Ctx(s.Ctx()).Err(err).Msg("failed to diapatch udp packet")
	}
}

func isDisallowed(ip tcpip.Address) bool {
	ipAddr := mynet.NetIPFromGvisorTcpipAddress(ip)
	return mynet.IsDirectedBoradcast(ipAddr) || !ipAddr.IsGlobalUnicast() || ipAddr.IsMulticast()
}
