package mux

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/net"
	nethelper "github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/pipe"
	"github.com/5vnetwork/vx-core/common/signal/done"

	"github.com/rs/zerolog/log"
)

type client struct {
	sync.RWMutex
	sessions             map[uint16]*clientSession
	count                atomic.Uint32
	lastEmptySessionTime time.Time

	ctx  context.Context
	link *pipe.Link // the other end of the PP is at the proxyC
	ClientStrategy
}

var (
	muxCoolAddressSrc = nethelper.DomainAddress("cool.mux.cool")
	muxCoolPortSrc    = nethelper.Port(7259)
	MuxCoolAddressDst = nethelper.DomainAddress("v1.mux.cool")
	MuxCoolPortDst    = nethelper.Port(9527)
)

// NewClient creates a new mux.Client.
func NewClient(ctx context.Context, link *pipe.Link, s ClientStrategy) (*client, error) {
	c := &client{
		ctx:            ctx,
		sessions:       make(map[uint16]*clientSession),
		link:           link,
		ClientStrategy: s,
	}

	go c.split()
	return c, nil
}

func (m *client) Close() {
	m.link.Interrupt(nil)
}

func (m *client) interrupt() {
	m.Lock()
	defer m.Unlock()
	for _, s := range m.sessions {
		select {
		case s.errChan <- errors.New("session error"):
		default:
		}
	}
}

func (m *client) AddSession(s *clientSession) {
	m.Lock()
	defer m.Unlock()
	m.sessions[s.ID] = s
}

func (m *client) RemoveSession(s *clientSession) {
	m.Lock()
	defer m.Unlock()
	delete(m.sessions, s.ID)
	if len(m.sessions) == 0 {
		m.lastEmptySessionTime = time.Now()
	}
}

func (m *client) IsIdle() bool {
	m.RLock()
	defer m.RUnlock()
	return len(m.sessions) == 0 &&
		time.Since(m.lastEmptySessionTime) > time.Second*5
}

func (m *client) IsEmpty() bool {
	m.RLock()
	defer m.RUnlock()
	return len(m.sessions) == 0
}

// has reached maxConnection
func (m *client) IsClosing() bool {
	m.RLock()
	defer m.RUnlock()

	if m.MaxConnection > 0 && m.count.Load() >= m.MaxConnection {
		return true
	}
	return false
}

func (m *client) IsFull() bool {
	m.RLock()
	defer m.RUnlock()
	if m.MaxConcurrency > 0 && len(m.sessions) >= int(m.MaxConcurrency) || m.IsClosing() {
		return true
	}
	return false
}

func (m *client) merge(ctx context.Context, dest net.Destination, s *clientSession) {
	if !dest.IsValid() {
		s.onError(errors.New("invalid target"))
	}

	transferType := TransferTypeStream
	if dest.Network == nethelper.Network_UDP {
		transferType = TransferTypePacket
	}

	s.writer = NewMuxWriter(s.ID, dest, m.link, transferType) //every session has a corresponding muxwriter
	if err := writeFirstPayload(s.rw, s.writer); err != nil {
		s.onError(fmt.Errorf("failed to write first payload: %w", err))
		return
	}

	err := buf.Copy(s.rw, s.writer)
	if err != nil {
		s.onError(fmt.Errorf("failed to copy request data: %w", err))
		s.notifyPeerSessionError()
		return
	}

	s.notifyPeerEOF()
	s.leftToRightDone.Close()
}

func writeFirstPayload(reader buf.Reader, writer *MuxWriter) error {
	err := buf.CopyOnceTimeout(reader, writer, time.Millisecond*100)
	if err == buf.ErrNotTimeoutReader || err == buf.ErrReadTimeout ||
		errors.Is(err, os.ErrDeadlineExceeded) {
		return writer.WriteMultiBuffer(buf.MultiBuffer{})
	}
	return err
}

func (m *client) split() {
	reader := &buf.BufferedReader{Reader: m.link}
	var meta FrameMetadata
	for {
		err := meta.Unmarshal(reader)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				log.Ctx(m.ctx).Error().Err(err).Msg("failed to read metadata")
			}
			break
		}

		switch meta.SessionStatus {
		case SessionStatusKeepAlive:
			err = m.handleStatueKeepAlive(&meta, reader)
		case SessionStatusEnd:
			err = m.handleStatusEnd(&meta, reader)
		case SessionStatusNew:
			err = m.handleStatusNew(&meta, reader)
		case SessionStatusKeep:
			err = m.handleStatusKeep(&meta, reader)
		default:
			status := meta.SessionStatus
			log.Ctx(m.ctx).Error().Msgf("unknown status: %b", status)
			return
		}
		if err != nil {
			log.Ctx(m.ctx).Error().Err(err).Uint8("sessionStatus", uint8(meta.SessionStatus)).Msg("failed to process data")
			return
		}
	}
}

func (m *client) handleStatueKeepAlive(meta *FrameMetadata, reader *buf.BufferedReader) error {
	if meta.Option.Has(OptionData) {
		return buf.Copy(buf.NewSizedReader(reader), buf.Discard)
	}
	return nil
}

func (m *client) handleStatusNew(meta *FrameMetadata, reader *buf.BufferedReader) error {
	if meta.Option.Has(OptionData) {
		return buf.Copy(buf.NewSizedReader(reader), buf.Discard)
	}
	return nil
}

func (w *client) handleStatusKeep(meta *FrameMetadata, reader *buf.BufferedReader) error {
	if !meta.Option.Has(OptionData) {
		return nil
	}

	w.RLock()
	s, found := w.sessions[meta.SessionID]
	w.RUnlock()
	if !found {
		// Notify remote peer to close this session.
		closingWriter := NewResponseMuxWriter(meta.SessionID, w.link, TransferTypeStream)
		closingWriter.hasError = true
		closingWriter.SendSessionStatusEnd()
		return buf.Copy(buf.NewSizedReader(reader), buf.Discard)
	}

	rr := buf.NewSizedReader(reader)
	err := buf.Copy(rr, s.rw)
	if err != nil && buf.IsWriteError(err) {
		log.Ctx(s.ctx).Error().Err(err).Msg("client failed to copy response data")
		s.notifyPeerSessionError()
		s.onError(err)
		drainErr := buf.Copy(rr, buf.Discard)
		return drainErr
	}

	return err
}

func (m *client) handleStatusEnd(meta *FrameMetadata, reader *buf.BufferedReader) error {
	m.RLock()
	s, found := m.sessions[meta.SessionID]
	m.RUnlock()
	if found {
		// When received a sessionStatusEnd with no error, it means that one direction's all data has been seen: EOF;
		// When received a sessionStatusEnd with error, it means that something went wrong and terminate this session.
		if meta.Option.Has(OptionError) {
			s.receivedSessionEndError.Store(true)
			log.Ctx(s.ctx).Error().Msg("Received sessionStatusEnd with error")
			s.onError(errors.New("session ended by peer"))
		} else {
			log.Ctx(s.ctx).Debug().Msg("Received sessionStatusEnd without error")
			s.rw.CloseWrite()
			s.rightToLeftDone.Close()
		}
	}
	if meta.Option.Has(OptionData) {
		return buf.Copy(buf.NewSizedReader(reader), buf.Discard)
	}
	return nil
}

type clientSession struct {
	sync.Mutex
	ctx    context.Context
	writer *MuxWriter
	rw     buf.ReaderWriter
	ID     uint16

	sendSessionEndError     atomic.Bool
	receivedSessionEndError atomic.Bool

	errChan         chan error
	leftToRightDone *done.Instance
	rightToLeftDone *done.Instance
}

func (s *clientSession) onError(err error) {
	select {
	case s.errChan <- err:
	default:
	}
}

// When s.rw.Read ruturns EOF, all data in that direction has been read, so informs the
// peer that this direction is done by sending a sessionStatusEnd. If an EOF has been received
// from the peer, close the session.
func (s *clientSession) notifyPeerEOF() {
	if s.sendSessionEndError.Load() {
		return
	}
	log.Ctx(s.ctx).Debug().Msg("notifyPeerEOF")
	err := s.writer.SendSessionStatusEnd()
	if err != nil {
		log.Ctx(s.ctx).Debug().Err(err).Msg("failed to notifyPeerEOF")
	}
}

func (s *clientSession) notifyPeerSessionError() {
	// if has received sessionStatusEnd, peer has closed the session, do nothing
	if !s.receivedSessionEndError.Load() && !s.sendSessionEndError.Load() {
		s.writer.hasError = true
		s.sendSessionEndError.Store(true)
		err := s.writer.SendSessionStatusEnd()
		if err != nil {
			log.Ctx(s.ctx).Debug().Err(err).Msg("failed to notifyPeerSessionError")
		}
	}
}
