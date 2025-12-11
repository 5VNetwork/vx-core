package udpmux

import (
	"context"
	"errors"
	"fmt"
	"io"
	gonet "net"
	"net/netip"
	"sync"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/task"
	"github.com/5vnetwork/vx-core/common/uuid"
	"github.com/5vnetwork/vx-core/i"

	"github.com/rs/zerolog/log"
)

type Muxer struct {
	sync.RWMutex
	sessions  map[uuid.UUID]*udpSession
	dnsGetter i.IPResolver
}

func NewMuxer(d i.IPResolver) *Muxer {
	return &Muxer{sessions: make(map[uuid.UUID]*udpSession), dnsGetter: d}
}

type MyPacketConn interface {
	ReadFrom(p []byte) (n int, addr net.Addr, err error)
	WriteTo(p []byte, addr net.Addr) (n int, err error)
	Close() error
}

// Returns when an error occurs or ctx is done.
func (m *Muxer) Handle(ctx context.Context, udpUuid uuid.UUID, dest net.AddressPort,
	rw buf.ReaderWriter, dialPacket func(ctx context.Context) (MyPacketConn, error)) error {
	m.Lock()
	s, ok := m.sessions[udpUuid]
	// if session not found, create a new session
	if !ok {
		pc, err := dialPacket(ctx)
		if err != nil {
			return fmt.Errorf("failed to dial packet conn: %w", err)
		}
		log.Ctx(ctx).Debug().Msg("muxer dial success")
		s = &udpSession{
			ctx:        ctx,
			uuid:       udpUuid,
			packetConn: pc,
			dstToRw:    make(map[netip.AddrPort]buf.ReaderWriter),
			parent:     m,
		}
		m.sessions[udpUuid] = s
		go s.handleResponse()
	}
	m.Unlock()
	// change dest to net.Addr
	var dst net.Addr
	if dest.Address.Family().IsDomain() {
		ip, err := m.dnsGetter.LookupIP(ctx, dest.Address.Domain())
		if err != nil {
			return fmt.Errorf("failed to resolve domain %s, %w", dest.Address.Domain(), err)
		}
		dst = &net.UDPAddr{IP: ip[0], Port: int(dest.Port)}
	} else {
		dst = &net.UDPAddr{IP: dest.Address.IP(), Port: int(dest.Port)}
	}
	dstAddrPort, err := netip.ParseAddrPort(dst.String())
	if err != nil {
		return fmt.Errorf("failed to parse destination address %s, %w", dst.String(), err)
	}
	s.addRW(dstAddrPort, rw)
	defer s.removeRW(dstAddrPort)
	// send to dest

	return task.Run(ctx, func() error {
		for {
			mb, err := rw.ReadMultiBuffer()
			if err != nil {
				if !errors.Is(err, io.EOF) {
					return fmt.Errorf("failed to read multi buffer from link, %w", err)
				}
				return nil
			}
			for _, b := range mb {
				_, err := s.packetConn.WriteTo(b.Bytes(), dst)
				if err != nil {
					return fmt.Errorf("failed to write to packet conn, %w", err)
				}
				b.Release()
			}
		}
	})
}
func (m *Muxer) deleteSession(s *udpSession) {
	m.Lock()
	defer m.Unlock()
	delete(m.sessions, s.uuid)
	s.close()
}

type udpSession struct {
	sync.RWMutex
	closed     bool
	ctx        context.Context
	uuid       uuid.UUID
	packetConn MyPacketConn
	dstToRw    map[netip.AddrPort]buf.ReaderWriter
	parent     *Muxer
	sync.Once
}

func (s *udpSession) close() {
	s.Once.Do(func() {
		s.closed = true
		s.packetConn.Close()
	})
}

func (s *udpSession) addRW(addr netip.AddrPort, rw buf.ReaderWriter) {
	s.Lock()
	defer s.Unlock()
	s.dstToRw[addr] = rw
}

func (s *udpSession) removeRW(addr netip.AddrPort) {
	s.Lock()
	defer s.Unlock()
	delete(s.dstToRw, addr)
	if len(s.dstToRw) == 0 {
		s.parent.deleteSession(s)
	}
}

func (s *udpSession) getRW(addr netip.AddrPort) (buf.ReaderWriter, bool) {
	s.RLock()
	defer s.RUnlock()
	rw, ok := s.dstToRw[addr]
	return rw, ok
}

func (s *udpSession) handleResponse() {
	for {
		buffer := buf.New()
		n, addr, err := s.packetConn.ReadFrom(buffer.BytesTo(buffer.Cap()))
		if err != nil {
			buffer.Release()
			if s.closed && errors.Is(err, gonet.ErrClosed) {
				return
			}
			log.Ctx(s.ctx).Error().Err(err).Msg("failed to read from packet conn")
			return
		}
		buffer.Extend(int32(n))
		addr.(*net.UDPAddr).Zone = ""
		key := netip.MustParseAddrPort(addr.String())
		rw, ok := s.getRW(key)
		if !ok {
			buffer.Release()
			continue
		}
		if err = rw.WriteMultiBuffer(buf.MultiBuffer{buffer}); err != nil {
			log.Ctx(s.ctx).Error().Err(err).Msg("failed to write to rw")
			s.removeRW(key)
		}
	}
}
