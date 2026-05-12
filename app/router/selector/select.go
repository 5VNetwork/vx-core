// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package selector

import (
	"context"

	"github.com/5vnetwork/vx-core/app/xsqlite"
	"github.com/5vnetwork/vx-core/common/session"
	"github.com/5vnetwork/vx-core/i"
	"github.com/rs/zerolog/log"
)

type Db interface {
	GetAllHandlers() ([]*xsqlite.OutboundHandler, error)
	GetHandlersByGroup(group string) ([]*xsqlite.OutboundHandler, error)
	GetBatchedHandlers(batchSize int, offset int) ([]*xsqlite.OutboundHandler, error)
	GetHandler(id int) *xsqlite.OutboundHandler
}

type HandlerErrorChangeSubject interface {
	AddHandlerErrorObserver(observer i.HandlerErrorObserver)
	RemoveHandlerErrorObserver(observer i.HandlerErrorObserver)
}

type Tester interface {
	// test speed with given size and retry
	TestSpeed(ctx context.Context, outbound i.Outbound, size uint32, retry bool) int64
	TestUsable(ctx context.Context, outbound i.Outbound, retry bool) bool
	TestIPv6(context.Context, i.Outbound) bool
	TestPing(context.Context, i.Outbound) int
}

type OutHandler interface {
	GetHandler() (i.Outbound, error)
	Name() string
	GetOk() int
	SetOk(ok int)
	GetPing() int
	SetPing(ping int)
	GetSpeed() int
	SetSpeed(speed int)
	GetSupport6() int
	SetSupport6(support6 int)
}

type handler struct {
	i.Outbound
	OutHandler
}

func (h *handler) GetHandler() (i.Outbound, error) {
	return h.Outbound, nil
}
func (h *handler) Support6() bool {
	return h.GetSupport6() > 0
}

type oStats struct {
	ok       int
	ping     int
	speed    int
	support6 int
}

func (o *oStats) GetSupport6() int {
	return o.support6
}

func (o *oStats) SetSupport6(support6 int) {
	o.support6 = support6
}

func (o *oStats) GetSpeed() int {
	return o.speed
}

func (o *oStats) SetSpeed(speed int) {
	o.speed = speed
}

func (o *oStats) GetPing() int {
	return o.ping
}

func (o *oStats) SetPing(ping int) {
	o.ping = ping
}

func (o *oStats) GetOk() int {
	return o.ok
}

func (o *oStats) SetOk(ok int) {
	o.ok = ok
}

func TestHandlerPing(ctx context.Context, s Tester, item OutHandler) {
	logger := log.With().Uint32("sid", uint32(session.NewID())).
		Str("handler", item.Name()).Logger()
	logger.Debug().Msg("test ping start")
	ctx = logger.WithContext(ctx)

	oh, err := item.GetHandler()
	if err != nil {
		logger.Error().Err(err).Msg("failed to get handler")
		return
	}
	ping := s.TestPing(ctx, oh)
	item.SetPing(ping)
	item.SetOk(ping)
	logger.Debug().Int("ping", ping).Msg("test handler ping result")
}

func TestHandler6(ctx context.Context, s Tester, item OutHandler) {
	logger := log.With().Uint32("sid", uint32(session.NewID())).
		Str("handler", item.Name()).Logger()
	logger.Debug().Msg("test 6 start")
	ctx = logger.WithContext(ctx)

	h, err := item.GetHandler()
	if err != nil {
		logger.Error().Err(err).Msg("failed to get handler")
		return
	}
	if h == nil {
		logger.Fatal().Any("item", item).Msg("handler is nil")
		return
	}
	ok := s.TestIPv6(ctx, h)
	if ok {
		item.SetSupport6(1)
		item.SetOk(1)
	} else {
		item.SetSupport6(-1)
	}
	logger.Debug().Bool("yes", ok).Msg("ipv6 test done")
}

// used to test unusable handlers
func TestHandlerUsable(ctx context.Context, s Tester, item OutHandler) {
	logger := log.With().Uint32("sid", uint32(session.NewID())).
		Str("handler", item.Name()).Logger()
	logger.Debug().Msg("usable test start")
	ctx = logger.WithContext(ctx)

	oh, err := item.GetHandler()
	if err != nil {
		logger.Error().Err(err).Msg("failed to get handler")
		return
	}
	usable := s.TestUsable(ctx, oh, false)
	if usable {
		item.SetOk(1)
		item.SetPing(0)
		item.SetSpeed(0)
	} else {
		item.SetOk(-1)
		item.SetSpeed(-1)
		item.SetPing(-1)
	}
	logger.Debug().Bool("usable", usable).Msg("handler usable result")
}

func TestHandlerSpeed(ctx context.Context, s Tester, item OutHandler, size uint32) {
	logger := log.With().Uint32("sid", uint32(session.NewID())).
		Str("handler", item.Name()).Logger()
	logger.Debug().Msg("speed test start")
	ctx = logger.WithContext(ctx)

	oh, err := item.GetHandler()
	if err != nil {
		logger.Error().Err(err).Msg("failed to get handler")
		return
	}
	speed := s.TestSpeed(ctx, oh, size, false)
	if speed <= 0 {
		item.SetSpeed(-1)
		item.SetOk(-1)
	} else {
		item.SetSpeed(int(speed))
		item.SetOk(int(speed))
	}
	logger.Debug().Int64("speed", speed).Msg("handler speed result")
}
