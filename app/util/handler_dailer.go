package util

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/5vnetwork/vx-core/common/dispatcher"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/common/pipe"
	"github.com/5vnetwork/vx-core/i"
	"github.com/rs/zerolog/log"
)

func HandlerToHttpClient(h i.FlowHandler) *http.Client {
	httpTransport := http.DefaultTransport.(*http.Transport).Clone()
	httpTransport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		log.Ctx(ctx).Debug().Str("addr", addr).Msg("dialing")
		d, err := net.ParseDestination(addr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse destination: %w", err)
		}
		d.Network = net.Network_TCP
		hd := &FlowHandlerToDialer{FlowHandler: h}
		return hd.Dial(ctx, d)
	}
	return &http.Client{
		Transport: httpTransport,
	}
}

// Adapt FlowHandler to Dialer
type FlowHandlerToDialer struct {
	i.FlowHandler
}

func (h *FlowHandlerToDialer) Dial(ctx context.Context, dest net.Destination) (net.Conn, error) {
	l1, l2 := pipe.NewLinks(8192, false)
	go func() {
		err := h.HandleFlow(ctx, dest, l2)
		if err != nil {
			l2.Interrupt(err)
			log.Ctx(ctx).Debug().Err(err).Msg("HandleFlow")
		} else {
			l2.Close()
		}
	}()

	remoteIP := net.AnyIP.IP()
	if !dest.Address.Family().IsDomain() {
		remoteIP = dest.Address.IP()
	}

	return pipe.NewLinkConn(l1, &net.TCPAddr{IP: net.AnyIP.IP(), Port: 0},
		&net.TCPAddr{IP: remoteIP, Port: int(dest.Port)}, dest.Network == net.Network_UDP), nil
}

type PacketHandlerToListener struct {
	i.PacketHandler
	Target *net.Destination
}

func (h *PacketHandlerToListener) ListenPacket(ctx context.Context, network, address string) (net.PacketConn, error) {
	var target net.Destination
	if h.Target != nil {
		target = *h.Target
	} else {
		target = net.AnyUdpDest
	}

	linkA, linkB := udp.NewLink(10)
	localDestination := net.Destination{
		Network: net.Network_UDP,
	}
	addr, port, err := net.SplitHostPort(address)
	if err != nil {
		localDestination.Address = net.LocalHostIP
		localDestination.Port = net.Port(rand.Intn(65535))
	} else {
		localDestination.Address = net.ParseAddress(addr)
		localDestination.Port, _ = net.PortFromString(port)
	}
	netPacketConn := &udp.LinkToNetPacketConn{
		PacketLink:       linkA,
		LocalDestination: localDestination,
	}
	go func() {
		err := h.PacketHandler.HandlePacketConn(ctx, target, linkB)
		if err != nil {
			linkB.Close()
			log.Ctx(ctx).Error().Err(err).Msg("failed to handle packet conn")
		}
	}()

	return netPacketConn, nil
}

type HandlerToDialerListener struct {
	PacketHandlerToListener
	FlowHandlerToDialer
}

type PacketRwToPacketConn struct {
	udp.UdpConn
	udp.PacketReaderToReadFromer
	udp.PacketWriterToWriteToer
	localAddr net.Addr
}

func (p *PacketRwToPacketConn) LocalAddr() net.Addr {
	return p.localAddr
}

func (l *PacketRwToPacketConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (l *PacketRwToPacketConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func (l *PacketRwToPacketConn) SetDeadline(t time.Time) error {
	return nil
}

type HandlerToProxyClient struct {
	i.Handler
}

func (h *HandlerToProxyClient) Dial(ctx context.Context, dst net.Destination) (net.Conn, error) {
	d := FlowHandlerToDialer{
		FlowHandler: h,
	}
	conn, err := d.Dial(ctx, dst)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (h *HandlerToProxyClient) DialWithInitialData(ctx context.Context, dst net.Destination,
	initialData []byte) (net.Conn, error) {
	conn, err := h.Dial(ctx, dst)
	if err != nil {
		return nil, err
	}
	if len(initialData) > 0 {
		_, err := conn.Write(initialData)
		if err != nil {
			return nil, err
		}
	}
	return conn, nil
}

func (h *HandlerToProxyClient) ListenPacket(ctx context.Context, dst net.Destination) (udp.UdpConn, error) {
	return dispatcher.NewDispatcherToPacketConn(ctx, h), nil
}
