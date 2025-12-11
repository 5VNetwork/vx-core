package selector

import (
	"context"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/xsqlite"
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
	TestSpeed(context.Context, i.Outbound, bool) int64
	TestUsable(context.Context, i.Outbound, bool) bool
	TestIPv6(context.Context, i.Outbound) bool
	TestPing(context.Context, i.Outbound) int
}

type CreateHandlerFunc func(h *configs.HandlerConfig, landHandlerIds []*xsqlite.OutboundHandler) (i.Outbound, error)

type outHandler interface {
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
	outHandler
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

func TestHandlerPing(ctx context.Context, s Tester, item outHandler) {
	log.Debug().Str("tag", item.Name()).Msg("test handler ping")
	oh, err := item.GetHandler()
	if err != nil {
		log.Error().Str("tag", item.Name()).Err(err).Msg("failed to get handler")
		return
	}
	ping := s.TestPing(ctx, oh)
	item.SetPing(ping)
	item.SetOk(ping)
	log.Debug().Str("name", item.Name()).Int("ping", ping).Msg("test handler ping result")
}

func TestHandler6(ctx context.Context, s Tester, item outHandler) {
	log.Debug().Str("tag", item.Name()).Msg("test handler ping")
	h, err := item.GetHandler()
	if err != nil {
		log.Error().Str("tag", item.Name()).Err(err).Msg("failed to get handler")
		return
	}
	ok := s.TestIPv6(ctx, h)
	if ok {
		item.SetSupport6(1)
		item.SetOk(1)
	} else {
		item.SetSupport6(-1)
	}
}

// used to test unusable handlers
func TestHandlerUsable(ctx context.Context, s Tester, item outHandler) {
	oh, err := item.GetHandler()
	if err != nil {
		log.Error().Str("tag", item.Name()).Err(err).Msg("failed to get handler")
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
}

func TestHandlerSpeed(ctx context.Context, s Tester, item outHandler) {
	log.Debug().Str("tag", item.Name()).Msg("test handler speed")
	oh, err := item.GetHandler()
	if err != nil {
		log.Error().Str("tag", item.Name()).Err(err).Msg("failed to get handler")
		return
	}
	speed := s.TestSpeed(ctx, oh, false)
	if speed <= 0 {
		item.SetSpeed(-1)
		item.SetOk(-1)
	} else {
		item.SetSpeed(int(speed))
		item.SetOk(int(speed))
	}
	log.Debug().Str("name", item.Name()).Int64("speed", speed).Msg("test handler speed result")
}
