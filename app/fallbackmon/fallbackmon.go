package fallbackmon

import (
	"context"
	"sync"

	"github.com/5vnetwork/vx-core/app/geo"
	cgeo "github.com/5vnetwork/vx-core/common/geo"
	"github.com/5vnetwork/vx-core/common/session"
	"github.com/rs/zerolog/log"
)

type FallbackMon struct {
	sync.Mutex
	sessions map[session.ID]*session.Info
	setting  *FallbackMonSetting
}

type FallbackMonSetting struct {
	Db            Db
	DomainSetName string
	Geo           *geo.Geo
}

type Db interface {
	AddGeoDomain(domain string, domainSetName string) error
}

func NewFallbackMon(setting *FallbackMonSetting) *FallbackMon {
	return &FallbackMon{
		sessions: make(map[session.ID]*session.Info),
		setting:  setting,
	}
}

func (r *FallbackMon) Start() error {
	return nil
}

func (r *FallbackMon) Close() error {
	return nil
}

func (r *FallbackMon) OnFallback(info *session.Info, previousTag, tag string) {
	if previousTag == "direct" {
		r.Lock()
		r.sessions[info.ID] = info
		r.Unlock()
	}
}

func (r *FallbackMon) FlowSessionEnd(ctx context.Context, info *session.Info, err error) {
	var s *session.Info
	r.Lock()
	if info, ok := r.sessions[info.ID]; ok {
		delete(r.sessions, info.ID)
		s = info
	}
	r.Unlock()

	if s == nil {
		return
	}
	if s.SessionDownCounter.Load() > 0 && s.GetTargetDomain() != "" {
		err := r.setting.Db.AddGeoDomain(s.GetTargetDomain(), r.setting.DomainSetName)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("failed to add geo domain to fallback set")
		} else {
			log.Ctx(ctx).Info().Str("domain", s.GetTargetDomain()).Msg("added geo domain to fallback set")
		}
		err = r.setting.Geo.AddDomain(r.setting.DomainSetName, &cgeo.Domain{
			Value: s.GetTargetDomain(),
			Type:  cgeo.Domain_Full,
		})
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("failed to add geo domain to fallback set")
		}
	}
}

func (r *FallbackMon) PacketConnSessionEnd(ctx context.Context, info *session.Info, err error) {
}
