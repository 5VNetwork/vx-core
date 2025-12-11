package udp

import (
	"context"
	"errors"
	"fmt"

	gonet "net"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/transport/dlhelper"

	"github.com/rs/zerolog/log"
)

type HubOption func(h *Hub)

func HubCapacity(capacity int) HubOption {
	return func(h *Hub) {
		h.capacity = capacity
	}
}

func HubReceiveOriginalDestination(r bool) HubOption {
	return func(h *Hub) {
		h.recvOrigDest = r
	}
}

type Hub struct {
	conn         *net.UDPConn
	cache        chan *udp.Packet
	capacity     int
	recvOrigDest bool
	closed       bool
}

func ListenUDP(ctx context.Context, address net.IP, port net.Port, so *dlhelper.SocketSetting, options ...HubOption) (*Hub, error) {
	hub := &Hub{
		capacity:     256,
		recvOrigDest: false,
	}
	for _, opt := range options {
		opt(hub)
	}

	if so != nil && so.ReceiveOriginalDestAddress {
		hub.recvOrigDest = true
	}

	udpConn, err := dlhelper.ListenSystemPacket(ctx, "udp", fmt.Sprintf("%s:%d", address, port), so)
	if err != nil {
		return nil, err
	}
	// errors.New("listening UDP on ", address, ":", port).WriteToLog()
	hub.conn = udpConn.(*net.UDPConn)
	hub.cache = make(chan *udp.Packet, hub.capacity)

	go hub.start()
	return hub, nil
}

// Close implements net.Listener.
func (h *Hub) Close() error {
	h.closed = true
	h.conn.Close()
	return nil
}

func (h *Hub) WriteTo(payload []byte, dest net.Destination) (int, error) {
	return h.conn.WriteToUDP(payload, &net.UDPAddr{
		IP:   dest.Address.IP(),
		Port: int(dest.Port),
	})
}

func (h *Hub) start() {
	c := h.cache
	defer close(c)

	oobBytes := make([]byte, 256)

	for {
		buffer := buf.New()
		var noob int
		var addr *net.UDPAddr
		rawBytes := buffer.Extend(buf.Size)

		n, noob, _, addr, err := ReadUDPMsg(h.conn, rawBytes, oobBytes)
		if err != nil {
			if !h.closed || !errors.Is(err, gonet.ErrClosed) {
				log.Err(err).Msg("failed to read UDP msg")
			}
			buffer.Release()
			break
		}
		buffer.Resize(0, int32(n))

		if buffer.IsEmpty() {
			buffer.Release()
			continue
		}

		payload := &udp.Packet{
			Payload: buffer,
			Source:  net.UDPDestination(net.IPAddress(addr.IP), net.Port(addr.Port)),
		}
		if h.recvOrigDest && noob > 0 {
			payload.Target = RetrieveOriginalDest(oobBytes[:noob])
			if payload.Target.IsValid() {
				log.Debug().Str("dst", payload.Target.String()).Msg("UDP original destination")
			} else {
				log.Warn().Msg("failed to read UDP original destination")
			}
		}

		select {
		case c <- payload:
		default:
			buffer.Release()
			payload.Payload = nil
		}
	}
}

// Addr implements net.Listener.
func (h *Hub) Addr() net.Addr {
	return h.conn.LocalAddr()
}

func (h *Hub) Receive() <-chan *udp.Packet {
	return h.cache
}
