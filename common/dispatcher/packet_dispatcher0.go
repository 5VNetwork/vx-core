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
	"github.com/5vnetwork/vx-core/common/session"
	"github.com/5vnetwork/vx-core/common/signal/done"
	"github.com/5vnetwork/vx-core/i"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type PacketDispatcher0 struct {
	sync.RWMutex
	tLinks       map[net.Destination][]*tLink0
	dispatcher   i.FlowHandler
	ctx          context.Context
	done         *done.Instance
	callback     atomic.Value // ResponseCallback
	linkLifetime time.Duration
	bufferSize   int
}

func NewPacketDispatcher0(ctx context.Context, dispatcher i.FlowHandler) *PacketDispatcher0 {
	p := &PacketDispatcher0{
		ctx:          ctx,
		dispatcher:   dispatcher,
		tLinks:       make(map[net.Destination][]*tLink0),
		done:         done.New(),
		bufferSize:   buf.BufferSize * 2,
		linkLifetime: 5 * time.Minute,
	}

	return p
}

func (p *PacketDispatcher0) SetResponseCallback(callback func(packet *udp.Packet)) {
	p.callback.Store(callback)
}

// payload's ownership is transferred to the dispatcher
func (s *PacketDispatcher0) DispatchPacket(destination net.Destination, payload *buf.Buffer) error {
	tLink, err := s.getTimeoutLink(destination)
	if err != nil {
		payload.Release()
		return fmt.Errorf("failed to get timeout link for %v: %w", destination, err)
	}
	var success bool
	for i := len(tLink) - 1; i >= 0; i-- {
		l := tLink[i]
		p := payload
		if i != 0 {
			p = payload.Clone()
		}
		if err1 := l.WriteMultiBuffer(buf.MultiBuffer{p}); err1 == nil {
			success = true
		} else {
			err = err1
		}
	}
	if !success {
		return err
	}
	return nil
}

func (s *PacketDispatcher0) Close() error {
	s.Lock()
	defer s.Unlock()
	if !s.done.Done() {
		s.done.Close()
		for _, l := range s.tLinks {
			for _, l := range l {
				l.Interrupt(nil)
			}
		}
	}
	return nil
}

func (s *PacketDispatcher0) getTimeoutLink(dest net.Destination) ([]*tLink0, error) {
	s.Lock()
	tlinks, found := s.tLinks[dest]
	if found {
		for i, l := range tlinks {
			if !l.IsOld() && l.isActive() {
				s.Unlock()
				return tlinks[i : i+1], nil
			}
			if l.IsOld() {
				defer s.removeTLink(dest, l)
			}
		}
	}

	if len(s.tLinks) > 1000 {
		s.Unlock()
		return nil, errors.New("too many links")
	}

	ctx := s.ctx
	if zerolog.GlobalLevel() == zerolog.DebugLevel {
		ctx = session.GetCtx(ctx)
		log.Ctx(ctx).Debug().Any("dst", dest).Msg("new udp sub session")
	}

	ctx, cancel := context.WithCancel(ctx)
	iLink, oLink := pipe.NewLinks(int32(s.bufferSize), true)
	tLink := &tLink0{
		ctx:  ctx,
		Link: iLink,
	}
	tLink.lastResponseTime.Store(time.Now().Unix())
	if s.linkLifetime > 0 {
		expireTime := time.Now().Add(s.linkLifetime)
		tLink.obseleteTime = &expireTime
	}

	// remove the first inactive tlink
	if len(tlinks) >= 8 {
		for i, l := range tlinks {
			if !l.isActive() {
				tlinks = append(tlinks[:i], tlinks[i+1:]...)
				break
			}
		}
	}
	// append the new tlink to the end of the list
	s.tLinks[dest] = append(tlinks, tLink)

	log.Ctx(ctx).Debug().Str("dst", dest.String()).Int("count", len(s.tLinks[dest])).Msg("new tlink0")

	go func() {
		if err := s.dispatcher.HandleFlow(ctx, dest, oLink); err != nil {
			if !s.done.Done() {
				log.Ctx(ctx).Debug().Err(err).Msg("failed to handle flow")
			}
		}
		cancel()
		s.removeTLink(dest, tLink)
		tLink.Interrupt(nil)
	}()
	go s.handleResponsePakcets(ctx, tLink, dest)

	result := s.tLinks[dest]
	s.Unlock()
	return result, nil
}

// each ppEnd is associated with a ctx
func (s *PacketDispatcher0) handleResponsePakcets(ctx context.Context, link *tLink0, addr net.Destination) {
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

func (s *PacketDispatcher0) removeTLink(dest net.Destination, tLink *tLink0) {
	s.Lock()
	defer s.Unlock()
	if tlinks, found := s.tLinks[dest]; found {
		for i, l := range tlinks {
			if l == tLink {
				s.tLinks[dest] = append(tlinks[:i], tlinks[i+1:]...)
				break
			}
		}
		log.Ctx(s.ctx).Debug().Str("dst", dest.String()).Int("count", len(s.tLinks[dest])).Msg("removeTLink")
	}
}

type tLink0 struct {
	*pipe.Link
	lastResponseTime atomic.Int64
	obseleteTime     *time.Time
	ctx              context.Context
}

func (t *tLink0) IsOld() bool {
	return t.obseleteTime != nil && t.obseleteTime.Before(time.Now())
}

func (t *tLink0) isActive() bool {
	return time.Now().Unix()-t.lastResponseTime.Load() < 4
}

func (t *tLink0) ReadMultiBuffer() (buf.MultiBuffer, error) {
	mb, err := t.Link.ReadMultiBuffer()
	if mb.Len() > 0 {
		t.lastResponseTime.Store(time.Now().Unix())
		// if old > 0 && time.Now().Unix()-old > 4 {
		// 	log.Ctx(t.ctx).Debug().Int64("since", time.Now().Unix()-old).Msg("an inactive tlink is still alive")
		// }
	}
	return mb, err
}
