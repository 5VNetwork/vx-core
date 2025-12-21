package dns

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"sync"
	"time"

	"github.com/5vnetwork/vx-core/app/inbound"
	"github.com/5vnetwork/vx-core/app/util"
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/session"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/transport/security/tls"
	"github.com/miekg/dns"
	"github.com/quic-go/quic-go"
	"github.com/rs/zerolog/log"
)

// NextProtoDQ - During connection establishment, DNS/QUIC support is indicated
// by selecting the ALPN token "doq" in the crypto handshake.
const NextProtoDQ = "doq"

const handshakeIdleTimeout = time.Second * 8

// QUICNameServer implemented DNS over QUIC
type QUICNameServer struct {
	sync.RWMutex
	cache       *rrCache
	name        string
	destination net.Destination
	connection  *quic.Conn
	clientIp    net.IP

	packetHandler i.PacketHandler
	ipToDomain    *IPToDomain
	ipResolver    i.IPResolver
}

type QuicNameServerOption struct {
	Name        string
	Destination net.Destination
	ClientIp    net.IP
	Handler     i.PacketHandler
	IpToDomain  *IPToDomain
	IPResolver  i.IPResolver
	RrCache     *rrCache
}

// NewQUICNameServer creates DNS-over-QUIC client object for local resolving
func NewQUICNameServer(option QuicNameServerOption) (*QUICNameServer, error) {
	rrCache := option.RrCache
	if rrCache == nil {
		rrCache = NewRrCache(RrCacheSetting{})
	}
	s := &QUICNameServer{
		cache:         rrCache,
		name:          option.Name,
		destination:   option.Destination,
		packetHandler: option.Handler,
		ipToDomain:    option.IpToDomain,
		ipResolver:    option.IPResolver,
		clientIp:      option.ClientIp,
	}

	return s, nil
}

func (s *QUICNameServer) Start() error {
	return s.cache.Start()
}

func (s *QUICNameServer) Close() error {
	return s.cache.Close()
}

func (s *QUICNameServer) Name() string {
	return s.name
}

// func (d *QUICNameServer) GetResolver(domain string, ip net.Address) (string, bool) {
// 	q := dns.Question{
// 		Name:   dns.Fqdn(domain),
// 		Qclass: dns.ClassINET,
// 	}
// 	if ip.Family().IsIPv4() {
// 		q.Qtype = dns.TypeA
// 	} else {
// 		q.Qtype = dns.TypeAAAA
// 	}
// 	entry, ok := d.cache.cache[q]
// 	if ok {
// 		return entry.resolver, true
// 	}
// 	return "", false
// }

// QueryIP is called from dns.Server->queryIPTimeout
func (s *QUICNameServer) HandleQuery(ctx context.Context, msg *dns.Msg, tcp bool) (*dns.Msg, error) {
	if len(msg.Question) == 0 {
		return nil, fmt.Errorf("no question in dns message")
	}

	question := msg.Question[0]
	cachedMsg, ok := s.cache.Get(&question)
	if ok {
		log.Ctx(ctx).Debug().Str("domain", question.Name).
			Str("type", dns.TypeToString[question.Qtype]).
			Any("reply", cachedMsg).Msg("dns quic cache hit")
		return makeReply(msg, cachedMsg), nil
	}

	// if there is clientIp, set it in EDNS0 Client Subnet option
	if len(s.clientIp) > 0 {
		addClientIP(msg, s.clientIp)
	}

	b := buf.New()
	defer b.Release()
	dnsMsgBytes, err := msg.PackBuffer(b.BytesRange(2, b.Cap()))
	if err != nil {
		return nil, err
	}
	binary.Write(b, binary.BigEndian, uint16(len(dnsMsgBytes)))
	b.Extend(int32(len(dnsMsgBytes)))

	startTime := time.Now()

	stream, err := s.openStream(ctx)
	if err != nil {
		return nil, err
	}

	_, err = stream.Write(b.Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed to write DNS message: %w", err)
	}

	_ = stream.Close()

	b.Clear()
	respBuf := b
	n, err := respBuf.ReadFullFrom(stream, 2)
	if err != nil && n == 0 {
		return nil, fmt.Errorf("failed to read response length: %w", err)
	}
	var length int16
	err = binary.Read(bytes.NewReader(respBuf.Bytes()), binary.BigEndian, &length)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response length: %w", err)
	}
	respBuf.Clear()
	n, err = respBuf.ReadFullFrom(stream, int32(length))
	if err != nil && n == 0 {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	rply := new(dns.Msg)
	if err := rply.Unpack(respBuf.Bytes()); err != nil {
		return nil, fmt.Errorf("failed to unpack dns quic response: %w", err)
	}

	log.Ctx(ctx).Debug().Str("domain", msg.Question[0].Name).
		Dur("t", time.Since(startTime)).
		Str("type", dns.TypeToString[msg.Question[0].Qtype]).
		Any("reply", rply).Msg("dns quic reply")
	if s.cache != nil {
		s.cache.Set(rply)
	}
	if s.ipToDomain != nil {
		s.ipToDomain.SetDomain(msg, s.destination.Address)
	}
	return rply, nil
}

func isActive(s *quic.Conn) bool {
	select {
	case <-s.Context().Done():
		return false
	default:
		return true
	}
}

func (s *QUICNameServer) getConnection() (*quic.Conn, error) {
	var conn *quic.Conn
	s.RLock()
	conn = s.connection
	if conn != nil && isActive(conn) {
		s.RUnlock()
		return conn, nil
	}
	if conn != nil {
		// we're recreating the connection, let's create a new one
		_ = conn.CloseWithError(0, "")
	}
	s.RUnlock()

	s.Lock()
	defer s.Unlock()

	logger := log.With().Uint32("sid", uint32(session.NewID())).Logger()
	ctx := logger.WithContext(context.Background())
	logger.Debug().Msg("dns quic")

	var err error
	conn, err = s.openConnection(ctx)
	if err != nil {
		// This does not look too nice, but QUIC (or maybe quic-go)
		// doesn't seem stable enough.
		// Maybe retransmissions aren't fully implemented in quic-go?
		// Anyways, the simple solution is to make a second try when
		// it fails to open the QUIC connection.
		conn, err = s.openConnection(ctx)
		if err != nil {
			return nil, err
		}
	}
	s.connection = conn
	return conn, nil
}

func (s *QUICNameServer) openConnection(ctx context.Context) (*quic.Conn, error) {
	tlsConfig := tls.TlsConfig{
		ServerName: func() string {
			switch s.destination.Address.Family() {
			case net.AddressFamilyIPv4, net.AddressFamilyIPv6:
				return s.destination.Address.IP().String()
			case net.AddressFamilyDomain:
				return s.destination.Address.Domain()
			default:
				panic("unknown address family")
			}
		}(),
	}
	quicConfig := &quic.Config{
		HandshakeIdleTimeout: handshakeIdleTimeout,
	}

	tlsCfg, err := tlsConfig.GetTLSConfig(tls.WithNextProtocol([]string{NextProtoDQ}))
	if err != nil {
		return nil, fmt.Errorf("failed to get TLS config: %w", err)
	}

	pl := util.PacketHandlerToListener{
		PacketHandler: s.packetHandler,
		Target:        &s.destination,
	}
	ctx = inbound.ContextWithInboundTag(ctx, s.name)
	pc, err := pl.ListenPacket(ctx, "udp", "")
	if err != nil {
		return nil, err
	}

	destination := s.destination
	if s.destination.Address.Family().IsDomain() {
		ips, err := s.ipResolver.LookupIP(ctx, s.destination.Address.Domain())
		if err != nil {
			return nil, err
		}
		if len(ips) == 0 {
			return nil, fmt.Errorf("no IP found for domain: %s", s.destination.Address.Domain())
		}
		destination = net.UDPDestination(net.IPAddress(ips[0]), s.destination.Port)
	}
	conn, err := quic.Dial(ctx, pc, destination.Addr(), tlsCfg, quicConfig)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (s *QUICNameServer) openStream(ctx context.Context) (*quic.Stream, error) {
	conn, err := s.getConnection()
	if err != nil {
		return nil, err
	}

	// open a new stream
	return conn.OpenStreamSync(ctx)
}
