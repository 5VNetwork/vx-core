// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package system

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/5vnetwork/vx-core/app/router"
	"github.com/5vnetwork/vx-core/common/buf"
	mynet "github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/gtcpip"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/common/session"
	"github.com/5vnetwork/vx-core/common/signal/done"
	"github.com/5vnetwork/vx-core/i"

	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/header"
)

type captureTun struct {
	written []*buf.Buffer
}

func (c *captureTun) Start() error                     { return nil }
func (c *captureTun) Close() error                     { return nil }
func (c *captureTun) Name() string                     { return "test" }
func (c *captureTun) ReadPacket() (*buf.Buffer, error) { return nil, nil }
func (c *captureTun) WritePacket(p *buf.Buffer) error {
	c.written = append(c.written, p)
	return nil
}

type fakePacketConn struct {
	mu      sync.Mutex
	written [][]byte
	closed  chan struct{}
	once    sync.Once
}

func newFakePacketConn() *fakePacketConn {
	return &fakePacketConn{closed: make(chan struct{})}
}

func (f *fakePacketConn) ReadFrom(p []byte) (int, net.Addr, error) {
	<-f.closed
	return 0, nil, net.ErrClosed
}
func (f *fakePacketConn) WriteTo(p []byte, addr net.Addr) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.written = append(f.written, append([]byte(nil), p...))
	return len(p), nil
}
func (f *fakePacketConn) Close() error {
	f.once.Do(func() { close(f.closed) })
	return nil
}
func (f *fakePacketConn) LocalAddr() net.Addr                { return &net.IPAddr{} }
func (f *fakePacketConn) SetDeadline(t time.Time) error      { return nil }
func (f *fakePacketConn) SetReadDeadline(t time.Time) error  { return nil }
func (f *fakePacketConn) SetWriteDeadline(t time.Time) error { return nil }

type mockRouter struct {
	handler i.Outbound
	err     error
	calls   int
}

func (m *mockRouter) PickHandler(ctx context.Context, si *session.Info) (i.Outbound, error) {
	m.calls++
	return m.handler, m.err
}
func (m *mockRouter) PickHandlerWithData(ctx context.Context, si *session.Info, rw interface{}) (interface{}, i.Outbound, []i.Fallback, error) {
	m.calls++
	return rw, m.handler, nil, m.err
}

type icmpOutbound struct {
	tag     string
	support bool
	opened  int
	pc      *fakePacketConn
}

func (o *icmpOutbound) Tag() string { return o.tag }
func (o *icmpOutbound) HandleFlow(context.Context, mynet.Destination, buf.ReaderWriter) error {
	return nil
}
func (o *icmpOutbound) HandlePacketConn(context.Context, mynet.Destination, udp.PacketReaderWriter) error {
	return nil
}
func (o *icmpOutbound) SupportsICMP() bool { return o.support }
func (o *icmpOutbound) OpenICMP(ctx context.Context, dst mynet.Destination) (net.PacketConn, error) {
	o.opened++
	if o.pc == nil {
		o.pc = newFakePacketConn()
	}
	return o.pc, nil
}

type proxyOutbound struct{ tag string }

func (o *proxyOutbound) Tag() string { return o.tag }
func (o *proxyOutbound) HandleFlow(context.Context, mynet.Destination, buf.ReaderWriter) error {
	return nil
}
func (o *proxyOutbound) HandlePacketConn(context.Context, mynet.Destination, udp.PacketReaderWriter) error {
	return nil
}

type unwrapOutbound struct {
	inner i.Outbound
}

func (u *unwrapOutbound) Tag() string { return u.inner.Tag() }
func (u *unwrapOutbound) HandleFlow(ctx context.Context, dst mynet.Destination, rw buf.ReaderWriter) error {
	return u.inner.HandleFlow(ctx, dst, rw)
}
func (u *unwrapOutbound) HandlePacketConn(ctx context.Context, dst mynet.Destination, p udp.PacketReaderWriter) error {
	return u.inner.HandlePacketConn(ctx, dst, p)
}
func (u *unwrapOutbound) UnwrapOutbound() i.Outbound { return u.inner }

func buildIPv4EchoRequest(src, dst net.IP, payload []byte) *buf.Buffer {
	total := header.IPv4MinimumSize + header.ICMPv4MinimumSize + len(payload)
	b := buf.New()
	packet := b.Extend(int32(total))
	ipHdr := header.IPv4(packet[:header.IPv4MinimumSize])
	ipHdr.Encode(&header.IPv4Fields{
		TotalLength: uint16(total),
		Protocol:    uint8(header.ICMPv4ProtocolNumber),
		SrcAddr:     tcpip.AddrFrom4Slice(src.To4()),
		DstAddr:     tcpip.AddrFrom4Slice(dst.To4()),
		TTL:         64,
	})
	icmpHdr := header.ICMPv4(packet[header.IPv4MinimumSize:])
	icmpHdr.SetType(header.ICMPv4Echo)
	icmpHdr.SetCode(0)
	icmpHdr.SetIdent(0x1234)
	icmpHdr.SetSequence(1)
	copy(icmpHdr[header.ICMPv4MinimumSize:], payload)
	icmpHdr.SetChecksum(header.ICMPv4Checksum(icmpHdr, 0))
	ipHdr.SetChecksum(0)
	ipHdr.SetChecksum(^ipHdr.CalculateChecksum())
	return b
}

func newTestInbound(tun *captureTun, icmp ICMPHandler) *TunSystemInbound {
	inbound := New(
		WithTun(tun),
		With4(net.ParseIP("172.19.0.2").To4(), net.ParseIP("172.19.0.1").To4(), 0),
		WithICMPHandler(icmp),
	)
	inbound.done = done.New()
	return inbound
}

func newTestICMP(tun *captureTun, r i.Router) *ICMPEchoHandler {
	return NewICMPEchoHandler(ICMPEchoHandlerOption{
		Tun:     tun,
		Router:  r,
		Tag:     "tun",
		Listen4: net.ParseIP("172.19.0.1").To4(),
		Ctx:     context.Background(),
	})
}

func TestLocalICMPv4EchoReply(t *testing.T) {
	tun := &captureTun{}
	icmp := newTestICMP(tun, nil)
	inbound := newTestInbound(tun, icmp)

	payload := []byte("ping-payload")
	req := buildIPv4EchoRequest(net.ParseIP("10.0.0.5").To4(), net.ParseIP("172.19.0.1").To4(), payload)
	inbound.processIPPakcet(req)

	if len(tun.written) != 1 {
		t.Fatalf("expected 1 written packet, got %d", len(tun.written))
	}
	reply := tun.written[0].Bytes()
	ipPacket := gtcpip.NewIPPacket(reply)
	if ipPacket == nil {
		t.Fatal("invalid reply ip packet")
	}
	src := mynetIP(ipPacket.SourceAddress())
	dst := mynetIP(ipPacket.DestinationAddress())
	if !src.Equal(net.ParseIP("172.19.0.1").To4()) {
		t.Fatalf("reply src = %v, want 172.19.0.1", src)
	}
	if !dst.Equal(net.ParseIP("10.0.0.5").To4()) {
		t.Fatalf("reply dst = %v, want 10.0.0.5", dst)
	}
	icmpHdr := header.ICMPv4(ipPacket.Payload())
	if icmpHdr.Type() != header.ICMPv4EchoReply {
		t.Fatalf("icmp type = %d, want echo reply", icmpHdr.Type())
	}
}

func TestICMPv4ForwardViaRouterDirect(t *testing.T) {
	tun := &captureTun{}
	out := &icmpOutbound{tag: "direct", support: true}
	r := &mockRouter{handler: out}
	icmp := newTestICMP(tun, r)
	inbound := newTestInbound(tun, icmp)

	req := buildIPv4EchoRequest(net.ParseIP("10.0.0.5").To4(), net.ParseIP("8.8.8.8").To4(), []byte("x"))
	inbound.processIPPakcet(req)

	if r.calls != 1 {
		t.Fatalf("expected 1 PickHandler call, got %d", r.calls)
	}
	if out.opened != 1 {
		t.Fatalf("expected OpenICMP once, got %d", out.opened)
	}
	out.pc.mu.Lock()
	n := len(out.pc.written)
	out.pc.mu.Unlock()
	if n != 1 {
		t.Fatalf("expected 1 icmp write, got %d", n)
	}
	_ = icmp.Close()
}

func TestICMPv4DropWhenProxySelected(t *testing.T) {
	tun := &captureTun{}
	r := &mockRouter{handler: &proxyOutbound{tag: "proxy"}}
	inbound := newTestInbound(tun, newTestICMP(tun, r))

	req := buildIPv4EchoRequest(net.ParseIP("10.0.0.5").To4(), net.ParseIP("8.8.8.8").To4(), []byte("x"))
	inbound.processIPPakcet(req)

	if r.calls != 1 {
		t.Fatalf("expected PickHandler, got %d", r.calls)
	}
	if len(tun.written) != 0 {
		t.Fatalf("proxy path should not write to tun, wrote %d", len(tun.written))
	}
}

func TestICMPv4ForwardThroughWrapper(t *testing.T) {
	tun := &captureTun{}
	inner := &icmpOutbound{tag: "direct", support: true}
	r := &mockRouter{handler: &unwrapOutbound{inner: inner}}
	icmp := newTestICMP(tun, r)
	inbound := newTestInbound(tun, icmp)

	req := buildIPv4EchoRequest(net.ParseIP("10.0.0.5").To4(), net.ParseIP("8.8.8.8").To4(), []byte("x"))
	inbound.processIPPakcet(req)

	if inner.opened != 1 {
		t.Fatalf("expected OpenICMP on inner outbound, got %d", inner.opened)
	}
	_ = icmp.Close()
}

func TestICMPv4DropWhenBlocked(t *testing.T) {
	tun := &captureTun{}
	r := &mockRouter{err: router.ErrBlocked}
	inbound := newTestInbound(tun, newTestICMP(tun, r))

	req := buildIPv4EchoRequest(net.ParseIP("10.0.0.5").To4(), net.ParseIP("8.8.8.8").To4(), []byte("x"))
	inbound.processIPPakcet(req)
	if len(tun.written) != 0 {
		t.Fatalf("blocked icmp should not write, wrote %d", len(tun.written))
	}
}

func TestICMPv4DroppedWithoutHandler(t *testing.T) {
	tun := &captureTun{}
	inbound := newTestInbound(tun, nil)

	req := buildIPv4EchoRequest(net.ParseIP("10.0.0.5").To4(), net.ParseIP("8.8.8.8").To4(), []byte("x"))
	inbound.processIPPakcet(req)
	if len(tun.written) != 0 {
		t.Fatalf("expected drop without icmp handler, wrote %d", len(tun.written))
	}
}

func mynetIP(addr tcpip.Address) net.IP {
	return net.IP(addr.AsSlice())
}
