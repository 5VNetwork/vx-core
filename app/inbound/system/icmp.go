// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package system

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"sync"
	"time"

	"github.com/5vnetwork/vx-core/app/router"
	"github.com/5vnetwork/vx-core/common/buf"
	mynet "github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/gtcpip"
	"github.com/5vnetwork/vx-core/common/session"
	"github.com/5vnetwork/vx-core/i"

	"github.com/rs/zerolog/log"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/header"
)

const (
	icmpIdleTimeout   = 30 * time.Second
	maxICMPPacketSize = 2048
)

// ICMPHandler handles ICMP IP packets demuxed by TunSystemInbound.
// Handle takes ownership of p.
type ICMPHandler interface {
	Handle(p *buf.Buffer, ipPacket gtcpip.IPPacket)
	// ClearRoutes drops cached outbound ICMP sockets so the next echo re-picks a handler.
	ClearRoutes()
	Close() error
}

type packetWriter interface {
	WritePacket(*buf.Buffer) error
}

// ICMPEchoHandler answers local echo and forwards others via router-selected ICMP outbounds.
type ICMPEchoHandler struct {
	tun      packetWriter
	router   i.Router
	tag      string
	listen4  net.IP
	listen6  net.IP
	ctx      context.Context
	isClosed func() bool
	routes   *icmpRouteTable
}

type ICMPEchoHandlerOption struct {
	Tun      packetWriter
	Router   i.Router
	Tag      string
	Listen4  net.IP
	Listen6  net.IP
	Ctx      context.Context
	IsClosed func() bool
}

func NewICMPEchoHandler(opt ICMPEchoHandlerOption) *ICMPEchoHandler {
	ctx := opt.Ctx
	if ctx == nil {
		ctx = context.Background()
	}
	isClosed := opt.IsClosed
	if isClosed == nil {
		isClosed = func() bool { return false }
	}
	return &ICMPEchoHandler{
		tun:      opt.Tun,
		router:   opt.Router,
		tag:      opt.Tag,
		listen4:  opt.Listen4,
		listen6:  opt.Listen6,
		ctx:      ctx,
		isClosed: isClosed,
		routes:   newICMPRouteTable(ctx, opt.Tun),
	}
}

// ClearRoutes closes all cached ICMP outbound routes without shutting down the handler.
func (h *ICMPEchoHandler) ClearRoutes() {
	if h.routes != nil {
		h.routes.ClearRoutes()
	}
}

func (h *ICMPEchoHandler) Close() error {
	h.ClearRoutes()
	return nil
}

func (h *ICMPEchoHandler) Handle(p *buf.Buffer, ipPacket gtcpip.IPPacket) {
	switch ipPacket.TransportProtocol() {
	case header.ICMPv4ProtocolNumber:
		h.handleIPv4(p, ipPacket)
	case header.ICMPv6ProtocolNumber:
		h.handleIPv6(p, ipPacket)
	default:
		p.Release()
	}
}

func (h *ICMPEchoHandler) handleIPv4(p *buf.Buffer, ipPacket gtcpip.IPPacket) {
	if isDisallowed(ipPacket.DestinationAddress()) {
		p.Release()
		return
	}
	icmpHdr := header.ICMPv4(ipPacket.Payload())
	if len(icmpHdr) < header.ICMPv4MinimumSize {
		p.Release()
		return
	}
	if icmpHdr.Type() != header.ICMPv4Echo || icmpHdr.Code() != 0 {
		p.Release()
		return
	}

	srcIP := mynet.NetIPFromGvisorTcpipAddress(ipPacket.SourceAddress())
	dstIP := mynet.NetIPFromGvisorTcpipAddress(ipPacket.DestinationAddress())

	if h.listen4 != nil && dstIP.Equal(h.listen4) {
		writeLocalICMPv4EchoReply(ipPacket, icmpHdr)
		if err := h.tun.WritePacket(p); err != nil && !h.isClosed() {
			log.Error().Err(err).Msg("failed to write icmpv4 echo reply")
		}
		return
	}

	defer p.Release()
	h.forward(mynet.IPToNetipAddr(srcIP), mynet.IPToNetipAddr(dstIP), append([]byte(nil), icmpHdr...))
}

func (h *ICMPEchoHandler) handleIPv6(p *buf.Buffer, ipPacket gtcpip.IPPacket) {
	if isDisallowed(ipPacket.DestinationAddress()) {
		p.Release()
		return
	}
	icmpHdr := header.ICMPv6(ipPacket.Payload())
	if len(icmpHdr) < header.ICMPv6MinimumSize {
		p.Release()
		return
	}
	if icmpHdr.Type() != header.ICMPv6EchoRequest || icmpHdr.Code() != 0 {
		p.Release()
		return
	}

	srcIP := mynet.NetIPFromGvisorTcpipAddress(ipPacket.SourceAddress())
	dstIP := mynet.NetIPFromGvisorTcpipAddress(ipPacket.DestinationAddress())

	if h.listen6 != nil && dstIP.Equal(h.listen6) {
		writeLocalICMPv6EchoReply(ipPacket, icmpHdr)
		if err := h.tun.WritePacket(p); err != nil && !h.isClosed() {
			log.Error().Err(err).Msg("failed to write icmpv6 echo reply")
		}
		return
	}

	defer p.Release()
	h.forward(mynet.IPToNetipAddr(srcIP), mynet.IPToNetipAddr(dstIP), append([]byte(nil), icmpHdr...))
}

func (h *ICMPEchoHandler) forward(src, dst netip.Addr, icmpPayload []byte) {
	if h.routes == nil || h.router == nil {
		return
	}
	key := icmpRouteKey{Source: src, Destination: dst}
	if route := h.routes.get(key); route != nil {
		if err := h.routes.write(route, icmpPayload); err != nil {
			log.Debug().Err(err).Str("src", src.String()).Str("dst", dst.String()).Msg("failed to write icmp")
		}
		return
	}

	srcDest := mynet.ICMPDestination(mynet.IPAddress(src.AsSlice()))
	dstDest := mynet.ICMPDestination(mynet.IPAddress(dst.AsSlice()))
	info := &session.Info{
		Source:     srcDest,
		Target:     dstDest,
		InboundTag: h.tag,
	}
	handler, err := h.router.PickHandler(h.ctx, info)
	if err != nil {
		if !errors.Is(err, router.ErrBlocked) && !errors.Is(err, router.ErrNoRule) && !errors.Is(err, router.ErrNoHandler) {
			log.Debug().Err(err).Str("dst", dst.String()).Msg("icmp pick handler failed")
		}
		return
	}
	icmpOut, ok := i.AsICMPOutbound(handler)
	if !ok || !icmpOut.SupportsICMP() {
		log.Debug().Str("handler", handler.Tag()).Str("dst", dst.String()).Msg("outbound does not support icmp")
		return
	}

	conn, err := icmpOut.OpenICMP(h.ctx, dstDest)
	if err != nil {
		log.Debug().Err(err).Str("handler", handler.Tag()).Str("dst", dst.String()).Msg("failed to open icmp")
		return
	}
	route := h.routes.put(key, conn)
	if err := h.routes.write(route, icmpPayload); err != nil {
		log.Debug().Err(err).Str("src", src.String()).Str("dst", dst.String()).Msg("failed to write icmp")
	}
}

type icmpRouteKey struct {
	Source      netip.Addr
	Destination netip.Addr
}

// icmpRouteTable caches outbound ICMP sockets keyed by (src, dst).
type icmpRouteTable struct {
	mu      sync.Mutex
	routes  map[icmpRouteKey]*icmpRoute
	timeout time.Duration
	tun     packetWriter
	parent  context.Context
}

type icmpRoute struct {
	key    icmpRouteKey
	conn   net.PacketConn
	cancel context.CancelFunc
	closed bool
}

func newICMPRouteTable(ctx context.Context, tun packetWriter) *icmpRouteTable {
	return &icmpRouteTable{
		routes:  make(map[icmpRouteKey]*icmpRoute),
		timeout: icmpIdleTimeout,
		tun:     tun,
		parent:  ctx,
	}
}

func (t *icmpRouteTable) ClearRoutes() {
	t.mu.Lock()
	defer t.mu.Unlock()
	for key, route := range t.routes {
		route.close()
		delete(t.routes, key)
	}
}

func (t *icmpRouteTable) Close() {
	t.ClearRoutes()
}

func (r *icmpRoute) close() {
	if r.closed {
		return
	}
	r.closed = true
	r.cancel()
	if r.conn != nil {
		_ = r.conn.Close()
	}
}

func (t *icmpRouteTable) get(key icmpRouteKey) *icmpRoute {
	t.mu.Lock()
	defer t.mu.Unlock()
	if route, ok := t.routes[key]; ok && !route.closed {
		return route
	}
	return nil
}

func (t *icmpRouteTable) put(key icmpRouteKey, conn net.PacketConn) *icmpRoute {
	t.mu.Lock()
	defer t.mu.Unlock()
	if route, ok := t.routes[key]; ok && !route.closed {
		_ = conn.Close()
		return route
	}
	_, cancel := context.WithCancel(t.parent)
	route := &icmpRoute{
		key:    key,
		conn:   conn,
		cancel: cancel,
	}
	t.routes[key] = route
	go t.readLoop(route)
	return route
}

func (t *icmpRouteTable) remove(route *icmpRoute) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if existing, ok := t.routes[route.key]; ok && existing == route {
		delete(t.routes, route.key)
	}
	route.close()
}

func (t *icmpRouteTable) readLoop(route *icmpRoute) {
	defer t.remove(route)

	replyBuf := make([]byte, maxICMPPacketSize)
	for {
		_ = route.conn.SetReadDeadline(time.Now().Add(t.timeout))
		n, peer, err := route.conn.ReadFrom(replyBuf)
		if err != nil {
			if t.parent.Err() != nil {
				return
			}
			log.Debug().Err(err).
				Str("src", route.key.Source.String()).
				Str("dst", route.key.Destination.String()).
				Msg("icmp route read ended")
			return
		}
		peerAddr, ok := addrToNetip(peer)
		if !ok {
			continue
		}
		icmpPayload := stripIPHeaderIfPresent(replyBuf[:n], &peerAddr)
		if route.key.Destination.IsValid() && peerAddr.IsValid() && peerAddr != route.key.Destination {
			continue
		}
		if len(icmpPayload) < header.ICMPv4MinimumSize {
			continue
		}

		pkt, err := buildICMPReplyPacket(route.key.Source, peerAddr, icmpPayload)
		if err != nil {
			log.Debug().Err(err).Msg("failed to build icmp reply packet")
			continue
		}
		if err := t.tun.WritePacket(pkt); err != nil {
			log.Debug().Err(err).Msg("failed to write icmp reply to tun")
			return
		}
	}
}

func (t *icmpRouteTable) write(route *icmpRoute, icmpPayload []byte) error {
	_, err := route.conn.WriteTo(icmpPayload, &net.IPAddr{IP: route.key.Destination.AsSlice()})
	return err
}

func stripIPHeaderIfPresent(b []byte, peer *netip.Addr) []byte {
	if len(b) == 0 {
		return b
	}
	switch header.IPVersion(b) {
	case header.IPv4Version:
		if len(b) < header.IPv4MinimumSize {
			return b
		}
		ipHdr := header.IPv4(b)
		if !ipHdr.IsValid(len(b)) || ipHdr.TransportProtocol() != header.ICMPv4ProtocolNumber {
			return b
		}
		src := mynet.IPToNetipAddr(mynet.NetIPFromGvisorTcpipAddress(ipHdr.SourceAddress()))
		if src.IsValid() {
			*peer = src
		}
		return ipHdr.Payload()
	case header.IPv6Version:
		if len(b) < header.IPv6MinimumSize {
			return b
		}
		ipHdr := header.IPv6(b)
		if !ipHdr.IsValid(len(b)) || ipHdr.TransportProtocol() != header.ICMPv6ProtocolNumber {
			return b
		}
		src := mynet.IPToNetipAddr(mynet.NetIPFromGvisorTcpipAddress(ipHdr.SourceAddress()))
		if src.IsValid() {
			*peer = src
		}
		return ipHdr.Payload()
	default:
		return b
	}
}

func addrToNetip(addr net.Addr) (netip.Addr, bool) {
	switch a := addr.(type) {
	case *net.IPAddr:
		ip, ok := netip.AddrFromSlice(a.IP)
		if !ok {
			return netip.Addr{}, false
		}
		return ip.Unmap(), true
	case *net.UDPAddr:
		ip, ok := netip.AddrFromSlice(a.IP)
		if !ok {
			return netip.Addr{}, false
		}
		return ip.Unmap(), true
	default:
		ip, err := netip.ParseAddr(addr.String())
		if err != nil {
			return netip.Addr{}, false
		}
		return ip.Unmap(), true
	}
}

func writeLocalICMPv4EchoReply(ipPacket gtcpip.IPPacket, icmpHdr header.ICMPv4) {
	src := ipPacket.SourceAddress()
	ipPacket.SetSourceAddress(ipPacket.DestinationAddress())
	ipPacket.SetDestinationAddress(src)
	icmpHdr.SetType(header.ICMPv4EchoReply)
	icmpHdr.SetChecksum(header.ICMPv4Checksum(icmpHdr, 0))
	ipPacket.ResetChecksum()
}

func writeLocalICMPv6EchoReply(ipPacket gtcpip.IPPacket, icmpHdr header.ICMPv6) {
	src := ipPacket.SourceAddress()
	ipPacket.SetSourceAddress(ipPacket.DestinationAddress())
	ipPacket.SetDestinationAddress(src)
	icmpHdr.SetType(header.ICMPv6EchoReply)
	icmpHdr.SetChecksum(header.ICMPv6Checksum(header.ICMPv6ChecksumParams{
		Header: icmpHdr,
		Src:    ipPacket.SourceAddress(),
		Dst:    ipPacket.DestinationAddress(),
	}))
}

func buildICMPReplyPacket(client, peer netip.Addr, icmpPayload []byte) (*buf.Buffer, error) {
	if peer.Is4() {
		if !client.Is4() {
			return nil, fmt.Errorf("address family mismatch")
		}
		total := header.IPv4MinimumSize + len(icmpPayload)
		b := buf.New()
		packet := b.Extend(int32(total))
		ipHdr := header.IPv4(packet[:header.IPv4MinimumSize])
		ipHdr.Encode(&header.IPv4Fields{
			TotalLength: uint16(total),
			Protocol:    uint8(header.ICMPv4ProtocolNumber),
			SrcAddr:     tcpip.AddrFrom4(peer.As4()),
			DstAddr:     tcpip.AddrFrom4(client.As4()),
			TTL:         64,
		})
		copy(packet[header.IPv4MinimumSize:], icmpPayload)
		icmpHdr := header.ICMPv4(packet[header.IPv4MinimumSize:])
		icmpHdr.SetChecksum(header.ICMPv4Checksum(icmpHdr, 0))
		ipHdr.SetChecksum(0)
		ipHdr.SetChecksum(^ipHdr.CalculateChecksum())
		return b, nil
	}

	if !client.Is6() {
		return nil, fmt.Errorf("address family mismatch")
	}
	total := header.IPv6MinimumSize + len(icmpPayload)
	b := buf.New()
	packet := b.Extend(int32(total))
	ipHdr := header.IPv6(packet[:header.IPv6MinimumSize])
	ipHdr.Encode(&header.IPv6Fields{
		PayloadLength:     uint16(len(icmpPayload)),
		TransportProtocol: header.ICMPv6ProtocolNumber,
		SrcAddr:           tcpip.AddrFrom16(peer.As16()),
		DstAddr:           tcpip.AddrFrom16(client.As16()),
		HopLimit:          64,
	})
	copy(packet[header.IPv6MinimumSize:], icmpPayload)
	icmpHdr := header.ICMPv6(packet[header.IPv6MinimumSize:])
	icmpHdr.SetChecksum(header.ICMPv6Checksum(header.ICMPv6ChecksumParams{
		Header: icmpHdr,
		Src:    ipHdr.SourceAddress(),
		Dst:    ipHdr.DestinationAddress(),
	}))
	return b, nil
}
