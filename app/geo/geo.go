// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package geo

import (
	"errors"
	"fmt"
	"net"
	"runtime"
	"sync"

	configs "github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/common"

	commongeo "buf.build/gen/go/vvvvv/vx/protocolbuffers/go/vx/common/geo"
	cgeo "github.com/5vnetwork/vx-core/common/geo"
	"github.com/5vnetwork/vx-core/i"

	"github.com/rs/zerolog/log"
)

func (g *Geo) UpdateGeo(geoConfig *configs.GeoConfig) error {
	common.Log()
	runtime.GC()
	common.Log()
	geo, err := NewGeo(geoConfig)
	if err != nil {
		return err
	}
	g.Lock()
	defer g.Unlock()
	g.AppSets = geo.AppSets
	g.DomainSets = geo.DomainSets
	g.IpSets = geo.IpSets
	g.OppositeDomainTags = geo.OppositeDomainTags
	g.OppositeIpTags = geo.OppositeIpTags
	return nil
}

func (g *Geo) AddDomainSet(name string, set i.DomainSet) {
	g.Lock()
	defer g.Unlock()
	g.DomainSets[name] = set
}

func (g *Geo) AddIPSet(name string, set i.IPSet) {
	g.Lock()
	defer g.Unlock()
	g.IpSets[name] = set
}

type Geo struct {
	sync.RWMutex
	// domain
	OppositeDomainTags map[string]string
	DomainSets         map[string]i.DomainSet
	// ip
	OppositeIpTags map[string]string
	IpSets         map[string]i.IPSet
	// app
	AppSets map[string]i.AppSet
}

func (g *Geo) Start() error {
	return nil
}

func (g *Geo) Close() error {
	g.Lock()
	defer g.Unlock()
	g.DomainSets = nil
	g.IpSets = nil
	g.AppSets = nil
	return nil
}

// if the domain set is not found, do nothing
func (g *Geo) AddDomain(name string, domain *commongeo.Domain) error {
	g.RLock()
	defer g.RUnlock()
	matcher, err := cgeo.ToStrMatcher(domain)
	if err != nil {
		return err
	}
	set, ok := g.DomainSets[name]
	if !ok {
		return nil
	}
	if set, ok := set.(*IndexMatcherToDomainSet); ok {
		set.addMatcher(matcher)
	} else {
		return fmt.Errorf("domain set %s is not an IndexMatcherToDomainSet", name)
	}
	return nil
}

func (g *Geo) RemoveDomain(name string, domain *commongeo.Domain) error {
	g.RLock()
	defer g.RUnlock()
	matcher, err := cgeo.ToStrMatcher(domain)
	if err != nil {
		return err
	}
	set, ok := g.DomainSets[name]
	if !ok {
		return fmt.Errorf("domain set %s not found", name)
	}
	if set, ok := set.(*IndexMatcherToDomainSet); ok {
		set.removeMatcher(matcher)
	} else {
		return fmt.Errorf("domain set %s is not an IndexMatcherToDomainSet", name)
	}
	return nil
}

func (g *Geo) MatchDomain(domain string, tag string) (bool, error) {
	g.RLock()
	defer g.RUnlock()
	if m, found := g.DomainSets[tag]; found {
		ret := m.Match(domain)
		log.Debug().Str("domain", domain).Str("tag", tag).Bool("matched", ret).Msg("geo match domain")
		return ret, nil
	}
	// if its opposite is known
	if opposite, found := g.OppositeDomainTags[tag]; found {
		if m, found := g.DomainSets[opposite]; found {
			ret := !m.Match(domain)
			log.Debug().Str("domain", domain).Str("tag", tag).Bool("matched", ret).Msg("geo match domain")
			return ret, nil
		}
	}
	log.Debug().Str("domain", domain).Str("tag", tag).Msg("geo match domain set not found")
	return false, ErrSetNotFound
}

func (g *Geo) MatchIP(ip net.IP, tag string) (bool, error) {
	g.RLock()
	defer g.RUnlock()
	if m, found := g.IpSets[tag]; found {
		ret := m.Match(ip)
		log.Debug().IPAddr("ip", ip).Str("tag", tag).Bool("matched", ret).Msg("geo match ip")
		return ret, nil
	}
	// if its opposite is known
	if oTag, found := g.OppositeIpTags[tag]; found {
		if m, found := g.IpSets[oTag]; found {
			ret := !m.Match(ip)
			log.Debug().IPAddr("ip", ip).Str("tag", tag).Bool("matched", ret).Msg("geo match ip")
			return ret, nil
		}
	}
	log.Debug().IPAddr("ip", ip).Str("tag", tag).Msg("geo match ip set not found")
	return false, ErrSetNotFound
}

func (g *Geo) MatchAppId(appId string, tag string) (bool, error) {
	g.RLock()
	defer g.RUnlock()
	if m, found := g.AppSets[tag]; found {
		ret := m.Match(appId)
		log.Debug().Str("appId", appId).Str("tag", tag).Bool("matched", ret).Msg("geo match appId")
		return ret, nil
	}
	log.Debug().Str("appId", appId).Str("tag", tag).Msg("geo match appId set not found")
	return false, ErrSetNotFound
}

var ErrSetNotFound = errors.New("set not found")
