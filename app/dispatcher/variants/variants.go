// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package variants

import (
	"context"
	"errors"
	"io"
	"sync/atomic"
	"time"

	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/common/session"
	"github.com/5vnetwork/vx-core/common/signal"
	"github.com/5vnetwork/vx-core/i"
	"github.com/rs/zerolog/log"

	"github.com/5vnetwork/vx-core/common/buf"
)

type TimeoutReaderWriter struct {
	buf.ReaderWriter
	Timeout i.TimeoutSetting
	Idle    *signal.ActivityChecker
	UpOnly  bool
	Ctx     context.Context
}

func (w *TimeoutReaderWriter) CloseWrite() error {
	if w.Timeout.UpLinkOnlyTimeout() != 0 {
		log.Ctx(w.Ctx).Debug().Msg("setting uplink only timeout")
		w.Idle.SetTimeout(w.Timeout.UpLinkOnlyTimeout())
	}
	return w.ReaderWriter.CloseWrite()
}

func (w *TimeoutReaderWriter) WriteMultiBuffer(mb buf.MultiBuffer) error {
	if !w.UpOnly {
		w.Idle.Update()
	}
	return w.ReaderWriter.WriteMultiBuffer(mb)
}

func (w *TimeoutReaderWriter) ReadMultiBuffer() (buf.MultiBuffer, error) {
	w.Idle.Update()
	m, err := w.ReaderWriter.ReadMultiBuffer()
	if err != nil {
		if errors.Is(err, io.EOF) {
			if w.Timeout.DownLinkOnlyTimeout() != 0 {
				log.Ctx(w.Ctx).Debug().Msg("setting downlink only timeout")
				w.Idle.SetTimeout(w.Timeout.DownLinkOnlyTimeout())
			}
		}
	}
	return m, err
}

type TimeoutDeadlineRW struct {
	i.DeadlineRW
	Timeout i.TimeoutSetting
	Idle    *signal.ActivityChecker
	UpOnly  bool
	Ctx     context.Context
}

func (w *TimeoutDeadlineRW) CloseWrite() error {
	if w.Timeout.UpLinkOnlyTimeout() != 0 {
		log.Ctx(w.Ctx).Debug().Msg("setting uplink only timeout")
		w.Idle.SetTimeout(w.Timeout.UpLinkOnlyTimeout())
	}
	return w.DeadlineRW.CloseWrite()
}

func (w *TimeoutDeadlineRW) WriteMultiBuffer(mb buf.MultiBuffer) error {
	if !w.UpOnly {
		w.Idle.Update()
	}
	return w.DeadlineRW.WriteMultiBuffer(mb)
}

func (w *TimeoutDeadlineRW) ReadMultiBuffer() (buf.MultiBuffer, error) {
	w.Idle.Update()
	m, err := w.DeadlineRW.ReadMultiBuffer()
	if err != nil {
		if errors.Is(err, io.EOF) {
			if w.Timeout.DownLinkOnlyTimeout() != 0 {
				log.Ctx(w.Ctx).Debug().Msg("setting downlink only timeout")
				w.Idle.SetTimeout(w.Timeout.DownLinkOnlyTimeout())
			}
		}
	}
	return m, err
}

type TimeoutPacketConn struct {
	Idle *signal.ActivityChecker
	udp.PacketReaderWriter
}

func (p *TimeoutPacketConn) ReadPacket() (*udp.Packet, error) {
	p.Idle.Update()
	return p.PacketReaderWriter.ReadPacket()
}

func (p *TimeoutPacketConn) WritePacket(packet *udp.Packet) error {
	return p.PacketReaderWriter.WritePacket(packet)
}

type StatsReaderWriter struct {
	buf.ReaderWriter
	// might be nil
	upCounter session.UpCounter
	// might be nil
	downCounter   session.DownCounter
	activeChecker *atomic.Value
}

func NewStatsReaderWriter(rw buf.ReaderWriter, upCounter session.UpCounter,
	downCounter session.DownCounter, activeChecker *atomic.Value) *StatsReaderWriter {
	return &StatsReaderWriter{
		ReaderWriter:  rw,
		upCounter:     upCounter,
		downCounter:   downCounter,
		activeChecker: activeChecker,
	}
}

func (w *StatsReaderWriter) WriteMultiBuffer(mb buf.MultiBuffer) error {
	w.downCounter.DownTraffic(uint64(mb.Len()))
	if w.activeChecker != nil {
		w.activeChecker.Store(time.Now())
	}
	return w.ReaderWriter.WriteMultiBuffer(mb)
}

func (w *StatsReaderWriter) ReadMultiBuffer() (buf.MultiBuffer, error) {
	mb, err := w.ReaderWriter.ReadMultiBuffer()
	w.upCounter.UpTraffic(uint64(mb.Len()))
	return mb, err
}

type StatsDeadlineRW struct {
	i.DeadlineRW
	// might be nil
	upCounter session.UpCounter
	// might be nil
	downCounter   session.DownCounter
	activeChecker *atomic.Value
}

func NewStatsDeadlineRW(rw i.DeadlineRW, upCounter session.UpCounter,
	downCounter session.DownCounter, activeChecker *atomic.Value) *StatsDeadlineRW {
	return &StatsDeadlineRW{
		DeadlineRW:    rw,
		upCounter:     upCounter,
		downCounter:   downCounter,
		activeChecker: activeChecker,
	}
}

func (w *StatsDeadlineRW) WriteMultiBuffer(mb buf.MultiBuffer) error {
	w.downCounter.DownTraffic(uint64(mb.Len()))
	if w.activeChecker != nil {
		w.activeChecker.Store(time.Now())
	}
	return w.DeadlineRW.WriteMultiBuffer(mb)
}

func (w *StatsDeadlineRW) ReadMultiBuffer() (buf.MultiBuffer, error) {
	mb, err := w.DeadlineRW.ReadMultiBuffer()
	w.upCounter.UpTraffic(uint64(mb.Len()))
	return mb, err
}

type StatsPacketConn struct {
	udp.PacketReaderWriter
	// might be nil
	upCounter session.UpCounter
	// might be nil
	downCounter   session.DownCounter
	activeChecker *atomic.Value
}

func NewStatsPacketConn(prw udp.PacketReaderWriter, upCounter session.UpCounter,
	downCounter session.DownCounter, activeChecker *atomic.Value) *StatsPacketConn {
	return &StatsPacketConn{
		PacketReaderWriter: prw,
		upCounter:          upCounter,
		downCounter:        downCounter,
		activeChecker:      activeChecker,
	}
}

func (p *StatsPacketConn) ReadPacket() (*udp.Packet, error) {
	packet, err := p.PacketReaderWriter.ReadPacket()
	if err != nil {
		return nil, err
	}
	p.upCounter.UpTraffic(uint64(packet.Payload.Len()))
	return packet, nil
}

func (p *StatsPacketConn) WritePacket(packet *udp.Packet) error {
	p.downCounter.DownTraffic(uint64(packet.Payload.Len()))
	if p.activeChecker != nil {
		p.activeChecker.Store(time.Now())
	}
	return p.PacketReaderWriter.WritePacket(packet)
}
