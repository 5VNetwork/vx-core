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

type GeoWrapper struct {
	sync.RWMutex
	geo *Geo // *Geo
}

func (g *GeoWrapper) GetGeo() *Geo {
	g.RLock()
	defer g.RUnlock()
	return g.geo
}

func (g *GeoWrapper) UpdateGeo(geoConfig *configs.GeoConfig) error {
	g.Lock()
	defer g.Unlock()
	common.Log()
	runtime.GC()
	common.Log()
	geo, err := NewGeo(geoConfig)
	if err != nil {
		return err
	}
	g.geo = geo
	return nil
}

func (g *GeoWrapper) AddDomainSet(name string, set i.DomainSet) {
	g.Lock()
	defer g.Unlock()
	g.geo.DomainSets[name] = set
}

func (g *GeoWrapper) AddIPSet(name string, set i.IPSet) {
	g.Lock()
	defer g.Unlock()
	g.geo.IpSets[name] = set
}

func (g *GeoWrapper) MatchDomain(domain string, tag string) (bool, error) {
	g.RLock()
	defer g.RUnlock()

	matched, err := g.geo.MatchDomain(domain, tag)
	log.Debug().Str("domain", domain).Str("tag", tag).Bool("matched", matched).Msg("geo match domain")
	return matched, err
}

func (g *GeoWrapper) MatchAppId(appId string, tag string) (bool, error) {
	g.RLock()
	defer g.RUnlock()
	return g.geo.MatchAppId(appId, tag)
}

func (g *GeoWrapper) MatchIP(ip net.IP, tag string) (bool, error) {
	g.RLock()
	defer g.RUnlock()
	matched, err := g.geo.MatchIP(ip, tag)
	log.Debug().IPAddr("ip", ip).Str("tag", tag).Bool("matched", matched).Msg("geo match ip")
	return matched, err
}

type Geo struct {
	// domain
	OppositeDomainTags map[string]string
	DomainSets         map[string]i.DomainSet
	// ip
	OppositeIpTags map[string]string
	IpSets         map[string]i.IPSet
	// app
	AppSets map[string]i.AppSet
}

// if the domain set is not found, do nothing
func (g *Geo) AddDomain(name string, domain *commongeo.Domain) error {
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
	if m, found := g.DomainSets[tag]; found {
		return m.Match(domain), nil
	}
	// if its opposite is known
	if opposite, found := g.OppositeDomainTags[tag]; found {
		if m, found := g.DomainSets[opposite]; found {
			return !m.Match(domain), nil
		}
	}
	return false, ErrSetNotFound
}

func (g *Geo) MatchIP(ip net.IP, tag string) (bool, error) {
	if m, found := g.IpSets[tag]; found {
		return m.Match(ip), nil
	}
	// if its opposite is known
	if oTag, found := g.OppositeIpTags[tag]; found {
		if m, found := g.IpSets[oTag]; found {
			return !m.Match(ip), nil
		}
	}
	return false, ErrSetNotFound
}

func (g *Geo) MatchAppId(appId string, tag string) (bool, error) {
	if m, found := g.AppSets[tag]; found {
		return m.Match(appId), nil
	}
	return false, ErrSetNotFound
}

var ErrSetNotFound = errors.New("set not found")
