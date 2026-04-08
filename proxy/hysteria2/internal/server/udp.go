package server

import (
	"errors"
	"math/rand"
	"sync"
	"time"

	"github.com/apernet/quic-go"
	"github.com/rs/zerolog/log"

	"github.com/5vnetwork/vx-core/app/inbound"
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/proxy"
	"github.com/5vnetwork/vx-core/proxy/hysteria2/internal/internal/frag"
	"github.com/5vnetwork/vx-core/proxy/hysteria2/internal/internal/protocol"
	"github.com/5vnetwork/vx-core/proxy/hysteria2/internal/internal/utils"
)

const (
	idleCleanupInterval = 1 * time.Second
)

type udpIO interface {
	ReceiveMessage() (*protocol.UDPMessage, error)
	SendMessage([]byte, *protocol.UDPMessage) error
	i.PacketHandler
}

type udpSessionEntry struct {
	ID   uint32
	D    *frag.Defragger
	Last *utils.AtomicTime
	// used for sending messages
	IO udpIO

	DialFunc func(dest net.Destination) (conn *udp.PacketLink, err error)
	ExitFunc func(err error)

	conn     *udp.PacketLink
	connLock sync.Mutex
	closed   bool
}

func newUDPSessionEntry(
	id uint32, io udpIO,
	dialFunc func(dest net.Destination) (conn *udp.PacketLink, err error),
	exitFunc func(error),
) (e *udpSessionEntry) {
	e = &udpSessionEntry{
		ID:   id,
		D:    &frag.Defragger{},
		Last: utils.NewAtomicTime(time.Now()),
		IO:   io,

		DialFunc: dialFunc,
		ExitFunc: exitFunc,
	}

	return e
}

// CloseWithErr closes the session and calls ExitFunc with the given error.
// A nil error indicates the session is cleaned up due to timeout.
func (e *udpSessionEntry) CloseWithErr(err error) {
	// We need this lock to ensure not to create conn after session exit
	e.connLock.Lock()

	if e.closed {
		// Already closed
		e.connLock.Unlock()
		return
	}

	e.closed = true
	if e.conn != nil {
		_ = e.conn.Close()
	}
	e.connLock.Unlock()

	e.ExitFunc(err)
}

// Feed feeds a UDP message to the session.
// If the message itself is a complete message, or it completes a fragmented message,
// the message is written to the session's UDP connection, and the number of bytes
// written is returned.
// Otherwise, 0 and nil are returned.
func (e *udpSessionEntry) Feed(msg *protocol.UDPMessage) (int, error) {
	e.Last.Set(time.Now())
	dfMsg := e.D.Feed(msg)
	if dfMsg == nil {
		return 0, nil
	}

	if e.conn == nil {
		err := e.initConn(dfMsg)
		if err != nil {
			return 0, err
		}
	}

	addr := dfMsg.Addr
	netAddr, err := net.ParseDestination(addr)
	if err != nil {
		return 0, err
	}

	b := buf.New()
	b.Write(dfMsg.Data)
	err = e.conn.WritePacket(&udp.Packet{
		Payload: b,
		Target: net.Destination{
			Address: netAddr.Address,
			Port:    netAddr.Port,
			Network: net.Network_UDP,
		},
	})
	if err != nil {
		return 0, err
	}
	return len(dfMsg.Data), nil
}

// initConn initializes the UDP connection of the session.
// If no error is returned, the e.conn is set to the new connection.
func (e *udpSessionEntry) initConn(firstMsg *protocol.UDPMessage) error {
	// We need this lock to ensure not to create conn after session exit
	e.connLock.Lock()

	if e.closed {
		e.connLock.Unlock()
		return errors.New("session is closed")
	}

	dest, err := net.ParseDestination(firstMsg.Addr)
	if err != nil {
		e.connLock.Unlock()
		e.CloseWithErr(err)
		return err
	}

	conn, err := e.DialFunc(dest)
	if err != nil {
		e.connLock.Unlock()
		e.CloseWithErr(err)
		return err
	}
	e.conn = conn

	go e.receiveLoop()

	e.connLock.Unlock()
	return nil
}

// receiveLoop receives incoming UDP packets, packs them into UDP messages,
// and sends using the IO.
// Exit when either the underlying UDP connection returns error (e.g. closed),
// or the IO returns error when sending.
func (e *udpSessionEntry) receiveLoop() {
	msgBuf := make([]byte, protocol.MaxUDPSize)
	for {
		packet, err := e.conn.ReadPacket()
		if err != nil {
			e.CloseWithErr(err)
			packet.Release()
			return
		}
		e.Last.Set(time.Now())

		msg := &protocol.UDPMessage{
			SessionID: e.ID,
			PacketID:  0,
			FragID:    0,
			FragCount: 1,
			Addr:      packet.Source.NetAddr(),
			Data:      packet.Payload.Bytes(),
		}
		err = sendMessageAutoFrag(e.IO, msgBuf, msg)
		if err != nil {
			e.CloseWithErr(err)
		}
		packet.Release()
	}
}

// sendMessageAutoFrag tries to send a UDP message as a whole first,
// but if it fails due to quic.ErrMessageTooLarge, it tries again by
// fragmenting the message.
func sendMessageAutoFrag(io udpIO, buf []byte, msg *protocol.UDPMessage) error {
	err := io.SendMessage(buf, msg)
	var errTooLarge *quic.DatagramTooLargeError
	if errors.As(err, &errTooLarge) {
		// Message too large, try fragmentation
		msg.PacketID = uint16(rand.Intn(0xFFFF)) + 1
		fMsgs := frag.FragUDPMessage(msg, int(errTooLarge.MaxDatagramPayloadSize))
		for _, fMsg := range fMsgs {
			err := io.SendMessage(buf, &fMsg)
			if err != nil {
				return err
			}
		}
		return nil
	} else {
		return err
	}
}

// udpSessionManager manages the lifecycle of UDP sessions.
// Each UDP session is identified by a SessionID, and corresponds to a UDP connection.
// A UDP session is created when a UDP message with a new SessionID is received.
// Similar to standard NAT, a UDP session is destroyed when no UDP message is received
// for a certain period of time (specified by idleTimeout).
type udpSessionManager struct {
	io          udpIO
	idleTimeout time.Duration

	gateway net.Addr
	src     net.Addr
	tag     string
	mutex   sync.RWMutex
	m       map[uint32]*udpSessionEntry
}

func newUDPSessionManager(io udpIO, idleTimeout time.Duration, src net.Addr, gateway net.Addr, tag string) *udpSessionManager {
	return &udpSessionManager{
		io:          io,
		idleTimeout: idleTimeout,
		m:           make(map[uint32]*udpSessionEntry),
		src:         src,
		gateway:     gateway,
		tag:         tag,
	}
}

// Run runs the session manager main loop.
// Exit and returns error when the underlying io returns error (e.g. closed).
func (m *udpSessionManager) Run() error {
	stopCh := make(chan struct{})
	go m.idleCleanupLoop(stopCh)
	defer close(stopCh)
	defer m.cleanup(false)

	for {
		msg, err := m.io.ReceiveMessage()
		if err != nil {
			return err
		}
		m.feed(msg)
	}
}

func (m *udpSessionManager) idleCleanupLoop(stopCh <-chan struct{}) {
	ticker := time.NewTicker(idleCleanupInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			m.cleanup(true)
		case <-stopCh:
			return
		}
	}
}

func (m *udpSessionManager) cleanup(idleOnly bool) {
	// We use RLock here as we are only scanning the map, not deleting from it.
	m.mutex.RLock()
	timeoutEntry := make([]*udpSessionEntry, 0, len(m.m))
	now := time.Now()
	for _, entry := range m.m {
		if !idleOnly || now.Sub(entry.Last.Get()) > m.idleTimeout {
			timeoutEntry = append(timeoutEntry, entry)
		}
	}
	m.mutex.RUnlock()

	for _, entry := range timeoutEntry {
		// This eventually calls entry.ExitFunc,
		// where the m.mutex will be locked again to remove the entry from the map.
		entry.CloseWithErr(nil)
	}
}

func (m *udpSessionManager) feed(msg *protocol.UDPMessage) {
	m.mutex.RLock()
	entry := m.m[msg.SessionID]
	m.mutex.RUnlock()

	// Create a new session if not exists
	if entry == nil {
		dialFunc := func(dest net.Destination) (conn *udp.PacketLink, err error) {
			p1, p2 := udp.NewLink(8)
			ctx, cancel := inbound.GetCtx(net.DestinationFromAddr(m.src),
				net.DestinationFromAddr(m.gateway), m.tag)
			ctx = proxy.ContextWithUser(ctx, m.io.(*udpIOImpl).User)
			ctx = proxy.ContextWithInboundProxyProtocol(ctx, "hysteria2")
			log.Ctx(ctx).Debug().Any("dst", dest).Send()

			go func() {
				err := m.io.HandlePacketConn(ctx, dest, p2)
				if err != nil {
					log.Ctx(ctx).Error().Err(err).Any("dest", dest).
						Msg("failed to handle packet conn")
				}
				p2.Close()
				cancel(err)
			}()
			return p1, nil
		}
		exitFunc := func(err error) {
			// Remove the session from the map
			m.mutex.Lock()
			delete(m.m, entry.ID)
			m.mutex.Unlock()
		}

		entry = newUDPSessionEntry(msg.SessionID, m.io, dialFunc, exitFunc)

		// Insert the session into the map
		m.mutex.Lock()
		m.m[msg.SessionID] = entry
		m.mutex.Unlock()
	}

	// Feed the message to the session
	// Feed (send) errors are ignored for now,
	// as some are temporary (e.g. invalid address)
	_, _ = entry.Feed(msg)
}

func (m *udpSessionManager) Count() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.m)
}
