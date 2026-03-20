package fallbackmon

import (
	"context"
	"os"
	"sync"

	"github.com/5vnetwork/vx-core/app/geo"
	cgeo "github.com/5vnetwork/vx-core/common/geo"
	"github.com/5vnetwork/vx-core/common/session"
	"github.com/rs/zerolog/log"
)

type FallbackMon struct {
	sync.Mutex
	sessions map[session.ID]*session.Info
	*FallbackMonSetting
	localFile *os.File
}

type FallbackMonSetting struct {
	Db            Db
	DomainSetName string
	Geo           *geo.GeoWrapper
	// local file to store the fallback domains
	LocalFile string
}

type Db interface {
	AddGeoDomain(domain string, domainSetName string) error
}

func NewFallbackMon(setting *FallbackMonSetting) *FallbackMon {
	return &FallbackMon{
		sessions:           make(map[session.ID]*session.Info),
		FallbackMonSetting: setting,
	}
}

func (r *FallbackMon) Start() error {
	var err error
	r.localFile, err = os.OpenFile(r.LocalFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (r *FallbackMon) Close() error {
	if r.localFile != nil {
		return r.localFile.Close()
	}
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
		// save to db
		if r.Db != nil {
			err := r.Db.AddGeoDomain(s.GetTargetDomain(), r.DomainSetName)
			if err != nil {
				log.Ctx(ctx).Err(err).Msg("failed to add geo domain to fallback set")
			} else {
				log.Ctx(ctx).Info().Str("domain", s.GetTargetDomain()).Msg("added geo domain to fallback set")
			}
		}

		// save to local file
		if r.localFile != nil {
			if r.Geo != nil {
				existed, _ := r.Geo.GetGeo().MatchDomain(s.GetTargetDomain(), r.DomainSetName)
				if existed {
					// this means the domain has been added to the fallback list
					return
				}
			}
			_, err = r.localFile.WriteString("DOMAIN-SUFFIX," + s.GetTargetDomain() + "\n")
			if err != nil {
				log.Ctx(ctx).Err(err).Msg("failed to write geo domain to fallback set")
			}
		}

		// add to set
		if r.Geo != nil {
			err = r.Geo.GetGeo().AddDomain(r.DomainSetName, &cgeo.Domain{
				Value: s.GetTargetDomain(),
				Type:  cgeo.Domain_RootDomain,
			})
			if err != nil {
				log.Ctx(ctx).Err(err).Msg("failed to add geo domain to fallback set")
			}
		}

	}
}

func (r *FallbackMon) PacketConnSessionEnd(ctx context.Context, info *session.Info, err error) {
}
