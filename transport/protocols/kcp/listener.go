package kcp

import (
	"context"
	"crypto/cipher"
	"crypto/tls"
	"fmt"
	"sync"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/transport/dlhelper"
	"github.com/5vnetwork/vx-core/transport/headers"
	"github.com/5vnetwork/vx-core/transport/protocols/udp"

	"github.com/rs/zerolog/log"
)

type ConnectionID struct {
	Remote net.Address
	Port   net.Port
	Conv   uint16
}

// Listener defines a server listening for connections
type Listener struct {
	sync.Mutex
	sessions  map[ConnectionID]*Connection
	hub       *udp.Hub
	tlsConfig *tls.Config
	config    *KcpConfig
	reader    PacketReader
	header    headers.PacketHeader
	security  cipher.AEAD
	h         func(net.Conn)
}

func Listen(ctx context.Context, addr net.Destination,
	config *KcpConfig, tlsConfig *tls.Config,
	so *dlhelper.SocketSetting, h func(net.Conn)) (*Listener, error) {
	header, err := config.GetPackerHeader()
	if err != nil {
		return nil, fmt.Errorf("failed to create header: %w", err)
	}
	security, err := config.GetSecurity()
	if err != nil {
		return nil, fmt.Errorf("failed to create security: %w", err)
	}
	l := &Listener{
		header:   header,
		security: security,
		reader: &KCPPacketReader{
			Header:   header,
			Security: security,
		},
		sessions: make(map[ConnectionID]*Connection),
		config:   config,
		h:        h,
	}

	l.tlsConfig = tlsConfig

	hub, err := udp.ListenUDP(ctx, addr.Address.IP(),
		addr.Port, so, udp.HubCapacity(1024))
	if err != nil {
		return nil, err
	}
	l.Lock()
	l.hub = hub
	l.Unlock()
	// errors.New("listening on ", address, ":", port).WriteToLog()

	go l.handlePackets()

	return l, nil
}

func (l *Listener) handlePackets() {
	receive := l.hub.Receive()
	for payload := range receive {
		l.OnReceive(payload.Payload, payload.Source)
	}
}

func (l *Listener) OnReceive(payload *buf.Buffer, src net.Destination) {
	segments := l.reader.Read(payload.Bytes())
	payload.Release()

	if len(segments) == 0 {
		log.Warn().Str("src", src.String()).Msg("discarding empty payload")
		return
	}

	conv := segments[0].Conversation()
	cmd := segments[0].Command()

	id := ConnectionID{
		Remote: src.Address,
		Port:   src.Port,
		Conv:   conv,
	}

	l.Lock()
	defer l.Unlock()

	conn, found := l.sessions[id]

	if !found {
		if cmd == CommandTerminate {
			return
		}
		writer := &Writer{
			id:       id,
			hub:      l.hub,
			dest:     src,
			listener: l,
		}
		remoteAddr := &net.UDPAddr{
			IP:   src.Address.IP(),
			Port: int(src.Port),
		}
		localAddr := l.hub.Addr()
		conn = NewConnection(ConnMetadata{
			LocalAddr:    localAddr,
			RemoteAddr:   remoteAddr,
			Conversation: conv,
		}, &KCPPacketWriter{
			Header:   l.header,
			Security: l.security,
			Writer:   writer,
		}, writer, l.config)
		var netConn net.Conn = conn
		if l.tlsConfig != nil {
			netConn = tls.Server(conn, l.tlsConfig)
		}
		l.h(netConn)
		l.sessions[id] = conn
	}
	conn.Input(segments)
}

func (l *Listener) Remove(id ConnectionID) {
	l.Lock()
	delete(l.sessions, id)
	l.Unlock()
}

// Close stops listening on the UDP address. Already Accepted connections are not closed.
func (l *Listener) Close() error {
	l.hub.Close()

	l.Lock()
	defer l.Unlock()

	for _, conn := range l.sessions {
		go conn.Terminate()
	}

	return nil
}

func (l *Listener) ActiveConnections() int {
	l.Lock()
	defer l.Unlock()

	return len(l.sessions)
}

// Addr returns the listener's network address, The Addr returned is shared by all invocations of Addr, so do not modify it.
func (l *Listener) Addr() net.Addr {
	return l.hub.Addr()
}

type Writer struct {
	id       ConnectionID
	dest     net.Destination
	hub      *udp.Hub
	listener *Listener
}

func (w *Writer) Write(payload []byte) (int, error) {
	return w.hub.WriteTo(payload, w.dest)
}

func (w *Writer) Close() error {
	w.listener.Remove(w.id)
	return nil
}
