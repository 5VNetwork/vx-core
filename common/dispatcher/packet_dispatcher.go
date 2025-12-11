package dispatcher

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/common/pipe"
	"github.com/5vnetwork/vx-core/common/signal"
	"github.com/5vnetwork/vx-core/common/signal/done"
	"github.com/5vnetwork/vx-core/i"

	"github.com/rs/zerolog/log"
)

// TODO: make it as a PacketConn
type PacketDispatcher struct {
	sync.RWMutex
	tLinks          map[net.Destination]*tLink
	dispatcher      i.FlowHandler
	ctx             context.Context
	done            *done.Instance
	callback        atomic.Value // ResponseCallback
	requestTimeout  time.Duration
	responseTimeout time.Duration
	linkLifetime    time.Duration
	bufferSize      int
}

func NewPacketDispatcher(ctx context.Context, dispatcher i.FlowHandler,
	opts ...PacketDispatcherOption) *PacketDispatcher {
	p := &PacketDispatcher{
		ctx:        ctx,
		dispatcher: dispatcher,
		tLinks:     make(map[net.Destination]*tLink),
		done:       done.New(),
		bufferSize: buf.BufferSize,
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

type PacketDispatcherOption func(*PacketDispatcher)

func WithRequestTimeout(timeout time.Duration) PacketDispatcherOption {
	return func(p *PacketDispatcher) {
		p.requestTimeout = timeout
	}
}

func WithResponseTimeout(timeout time.Duration) PacketDispatcherOption {
	return func(p *PacketDispatcher) {
		p.responseTimeout = timeout
	}
}

func WithBufferSize(size int) PacketDispatcherOption {
	return func(p *PacketDispatcher) {
		p.bufferSize = size
	}
}

func WithResponseCallback(callback func(packet *udp.Packet)) PacketDispatcherOption {
	return func(p *PacketDispatcher) {
		p.callback.Store(callback)
	}
}

func WithLinkLifetime(lifetime time.Duration) PacketDispatcherOption {
	return func(p *PacketDispatcher) {
		p.linkLifetime = lifetime
	}
}

func (p *PacketDispatcher) SetResponseCallback(callback func(packet *udp.Packet)) {
	p.callback.Store(callback)
}

// payload's ownership is transferred to the dispatcher. PacketDispatcher releases it even if DispatchPacket fails.
func (s *PacketDispatcher) DispatchPacket(destination net.Destination, payload *buf.Buffer) error {
	tLink, err := s.getTimeoutLink(destination)
	if err != nil {
		return fmt.Errorf("failed to get timeout link for %v: %w", destination, err)
	}
	if err := tLink.WriteMultiBuffer(buf.MultiBuffer{payload}); err != nil {
		return fmt.Errorf("failed to write UDP payload for %v", destination)
	}
	return nil
}

var ErrClosed = errors.New("closed")

func (s *PacketDispatcher) Close() error {
	s.Lock()
	defer s.Unlock()
	if !s.done.Done() {
		s.done.Close()
		for _, l := range s.tLinks {
			l.Interrupt(nil)
		}
	}
	return nil
}

func (s *PacketDispatcher) getTimeoutLink(dest net.Destination) (*tLink, error) {
	s.Lock()
	defer s.Unlock()

	if tlink, found := s.tLinks[dest]; found && !tlink.IsOld() {
		return tlink, nil
	}

	if len(s.tLinks) > 1000 {
		return nil, errors.New("too many links")
	}

	// ctx := session.ContextWithInfo(s.ctx, newInfo)
	// ctx = log.With().Uint32("sid", uint32(newInfo.ID)).Uint32("old_id", uint32(s.info.ID)).Logger().WithContext(ctx)
	// log.Ctx(ctx).Debug().Str("dst", dest.String()).Msg("new udp sub session")

	ctx, cancel := context.WithCancel(s.ctx)
	iLink, oLink := pipe.NewLinks(int32(s.bufferSize), false)
	tLink := &tLink{
		Link: iLink,
	}

	if s.requestTimeout > 0 {
		tLink.requestActivityChecker = signal.NewActivityChecker(func() {
			tLink.Interrupt(errors.ErrIdle)
			log.Ctx(ctx).Debug().Msg("request timeout")
		}, s.requestTimeout)
	}
	if s.responseTimeout > 0 {
		tLink.responseActivityChecker = signal.NewActivityChecker(func() {
			tLink.Interrupt(errors.ErrIdle)
			log.Ctx(ctx).Debug().Msg("response timeout")
		}, s.responseTimeout)
	}
	if s.linkLifetime > 0 {
		expireTime := time.Now().Add(s.linkLifetime)
		tLink.obseleteTime = &expireTime
	}

	s.tLinks[dest] = tLink

	log.Ctx(ctx).Debug().Str("dst", dest.String()).Msg("new tlink")

	go func() {
		if err := s.dispatcher.HandleFlow(ctx, dest, oLink); err != nil {
			if !s.done.Done() {
				log.Ctx(ctx).Debug().Err(err).Msg("failed to handle flow")
			}
		}
		cancel()
		s.removeTLink(dest, tLink)
		tLink.Interrupt(nil)
		if tLink.requestActivityChecker != nil {
			tLink.requestActivityChecker.Cancel()
		}
		if tLink.responseActivityChecker != nil {
			tLink.responseActivityChecker.Cancel()
		}
	}()
	go s.handleResponsePakcets(ctx, tLink, dest)
	return tLink, nil
}

// each ppEnd is associated with a ctx
func (s *PacketDispatcher) handleResponsePakcets(ctx context.Context, link *tLink, addr net.Destination) {
	reader := link

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.done.Wait():
			return
		default:
		}

		mb, err := reader.ReadMultiBuffer()
		for _, b := range mb {
			cb := s.callback.Load()
			if cb != nil {
				cb.(func(packet *udp.Packet))(&udp.Packet{
					Payload: b,
					Source:  addr,
				})
			} else {
				b.Release()
			}
		}
		if err != nil {
			log.Ctx(ctx).Debug().Err(err).Str("dst", addr.String()).Msg("handle Response end")
			return
		}
	}
}

func (s *PacketDispatcher) removeTLink(dest net.Destination, tLink *tLink) {
	s.Lock()
	defer s.Unlock()
	if current, found := s.tLinks[dest]; found && current == tLink {
		delete(s.tLinks, dest)
		log.Ctx(s.ctx).Debug().Str("dst", dest.String()).Int("ramaining_links", len(s.tLinks)).Msg("removeTLink")
	}
}

type tLink struct {
	*pipe.Link
	requestActivityChecker  *signal.ActivityChecker
	responseActivityChecker *signal.ActivityChecker
	obseleteTime            *time.Time
}

func (t *tLink) IsOld() bool {
	return t.obseleteTime != nil && t.obseleteTime.Before(time.Now())
}

func (t *tLink) WriteMultiBuffer(mb buf.MultiBuffer) error {
	if t.requestActivityChecker != nil {
		t.requestActivityChecker.Update()
	}
	return t.Link.WriteMultiBuffer(mb)
}

func (t *tLink) ReadMultiBuffer() (buf.MultiBuffer, error) {
	mb, err := t.Link.ReadMultiBuffer()
	if t.responseActivityChecker != nil && mb.Len() > 0 {
		t.responseActivityChecker.Update()
	}
	return mb, err
}
