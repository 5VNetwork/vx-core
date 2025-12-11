package mux

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/pipe"
	"github.com/5vnetwork/vx-core/common/session"
	"github.com/5vnetwork/vx-core/i"

	"github.com/rs/zerolog/log"
)

type server struct {
	link       buf.ReaderWriter
	Dispatcher i.FlowHandler

	sessionLock sync.RWMutex
	sessions    map[uint16]*serverSession
}

func Serve(ctx context.Context, rw buf.ReaderWriter, d i.FlowHandler) error {
	log.Ctx(ctx).Debug().Msg("MuxServer started")
	server := &server{
		link:       rw,
		sessions:   make(map[uint16]*serverSession),
		Dispatcher: d,
	}
	return server.run(ctx)
}

func (w *server) run(ctx context.Context) error {
	rw := w.link
	reader := &buf.BufferedReader{Reader: rw}

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			err := w.handleFrame(ctx, reader)
			if err != nil {
				if errors.Is(err, io.EOF) {
					rw.CloseWrite()
				}
				w.sessionLock.RLock()
				defer w.sessionLock.RUnlock()
				for _, s := range w.sessions {
					s.link.Interrupt(nil)
				}
				return err
			}
		}
	}
}

func (w *server) handleFrame(ctx context.Context, reader *buf.BufferedReader) error {
	var meta FrameMetadata
	err := meta.Unmarshal(reader)
	if err != nil {
		return errors.Join(errors.New("failed to read metadata"), err)
	}

	switch meta.SessionStatus {
	case SessionStatusKeepAlive:
		err = w.handleStatusKeepAlive(&meta, reader)
	case SessionStatusEnd:
		err = w.handleStatusEnd(&meta, reader)
	case SessionStatusNew:
		err = w.handleStatusNew(ctx, &meta, reader)
	case SessionStatusKeep:
		err = w.handleStatusKeep(&meta, reader)
	default:
		status := meta.SessionStatus
		return fmt.Errorf("unknown status: %b", status)
	}
	if err != nil {
		return errors.Join(errors.New("failed to read metadata"), err)
	}
	return nil
}

func (w *server) handleResponseData(s *serverSession) {
	if err := buf.Copy(s.link, s.writer); err != nil {
		log.Ctx(s.ctx).Error().Err(err).Msg("failed to copy response data")
		w.onSessionError(s, err)
		return
	}
	s.notifyPeerSessionEOF()
}

func (w *server) handleStatusNew(ctx context.Context, meta *FrameMetadata, reader *buf.BufferedReader) error {
	newCtx := log.With().Uint32("sid", uint32(session.NewID())).Logger().WithContext(ctx)
	log.Ctx(newCtx).Debug().Uint16("mux_sid", meta.SessionID).Str("dst", meta.Target.String()).Msg("new mux session")

	iLink, oLink := pipe.NewLinks(64*1024, meta.Target.Network == net.Network_UDP)

	go func() {
		err := w.Dispatcher.HandleFlow(newCtx, meta.Target, oLink)
		if err != nil {
			log.Ctx(newCtx).Error().Err(err).Msg("HandleFlow failed")
		}

		w.sessionLock.Lock()
		delete(w.sessions, meta.SessionID)
		w.sessionLock.Unlock()
	}()

	s := &serverSession{
		id:   meta.SessionID,
		ctx:  newCtx,
		link: iLink,
	}
	transferType := TransferTypeStream
	if meta.Target.Network == net.Network_UDP {
		transferType = TransferTypePacket
	}
	s.writer = NewResponseMuxWriter(s.id, w.link, transferType)

	w.sessionLock.Lock()
	w.sessions[s.id] = s
	w.sessionLock.Unlock()

	// handle returning data, send them back
	go w.handleResponseData(s)

	if !meta.Option.Has(OptionData) {
		return nil
	}

	// copy the additional data in this frame to the linkA.
	rr := buf.NewSizedReader(reader)
	if err := buf.Copy(rr, s.link); err != nil {
		buf.Copy(rr, buf.Discard)
		go w.onSessionError(s, err)
	}
	return nil
}

func (w *server) handleStatusKeepAlive(meta *FrameMetadata, reader *buf.BufferedReader) error {
	if meta.Option.Has(OptionData) {
		return buf.Copy(buf.NewSizedReader(reader), buf.Discard)
	}
	return nil
}

func (ms *server) handleStatusKeep(meta *FrameMetadata, reader *buf.BufferedReader) error {
	if !meta.Option.Has(OptionData) {
		return nil
	}

	ms.sessionLock.RLock()
	s, found := ms.sessions[meta.SessionID]
	ms.sessionLock.RUnlock()
	if !found {
		// Notify remote peer to close this session.
		closingWriter := NewResponseMuxWriter(meta.SessionID, ms.link, TransferTypeStream)
		closingWriter.hasError = true
		closingWriter.SendSessionStatusEnd()
		return buf.Copy(buf.NewSizedReader(reader), buf.Discard)
	}

	rr := buf.NewSizedReader(reader)
	err := buf.Copy(rr, s.link)
	// if sth went wrong, inform the remote peer and end the session
	if err != nil && buf.IsWriteError(err) {
		log.Ctx(s.ctx).Error().Err(err).Msg("mux server failed to copy request data")
		// Notify remote peer to close this session.
		go ms.onSessionError(s, err)
		drainErr := buf.Copy(rr, buf.Discard)
		return drainErr
	}

	return err
}

func (w *server) onSessionError(s *serverSession, err error) {
	s.link.Interrupt(err)

	if !s.sendSessionEndError.Load() && !s.receivedSessionEndError.Load() {
		s.writer.hasError = true
		s.sendSessionEndError.Store(true)
		s.writer.SendSessionStatusEnd()
	}

	w.sessionLock.Lock()
	delete(w.sessions, s.id)
	w.sessionLock.Unlock()
}

func (w *server) handleStatusEnd(meta *FrameMetadata, reader *buf.BufferedReader) error {
	w.sessionLock.RLock()
	s, found := w.sessions[meta.SessionID]
	w.sessionLock.RUnlock()
	if found {
		if meta.Option.Has(OptionError) {
			s.receivedSessionEndError.Store(true)
			log.Ctx(s.ctx).Error().Msg("Received sessionStatusEnd with error")
			go w.onSessionError(s, errors.New("session ended by peer"))
		} else {
			log.Ctx(s.ctx).Debug().Msg("Received sessionStatusEnd without error")
			s.link.CloseWrite()
		}
	}
	if meta.Option.Has(OptionData) {
		return buf.Copy(buf.NewSizedReader(reader), buf.Discard)
	}
	return nil
}

type serverSession struct {
	id   uint16
	link *pipe.Link
	ctx  context.Context

	writer *MuxWriter

	sendSessionEndError     atomic.Bool
	receivedSessionEndError atomic.Bool
}

func (s *serverSession) notifyPeerSessionEOF() {
	if s.sendSessionEndError.Load() {
		return
	}
	log.Ctx(s.ctx).Debug().Msg("notifyPeerEOF")
	err := s.writer.SendSessionStatusEnd()
	if err != nil {
		log.Ctx(s.ctx).Debug().Err(err).Msg("failed to notifyPeerEOF")
	}
}
