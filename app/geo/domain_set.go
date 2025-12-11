package geo

import (
	"fmt"
	"slices"
	"sync"

	"github.com/5vnetwork/vx-core/common/geo"
	"github.com/5vnetwork/vx-core/common/strmatcher"
	"github.com/5vnetwork/vx-core/i"
)

type IndexMatcherToDomainSet struct {
	strmatcher.IndexMatcher
	matchersLock    sync.RWMutex
	matchers        []strmatcher.Matcher
	removedMatchers []strmatcher.Matcher
}

func (i *IndexMatcherToDomainSet) Match(domain string) bool {
	i.matchersLock.RLock()
	defer i.matchersLock.RUnlock()

	for _, m := range i.removedMatchers {
		if m.Match(domain) {
			return false
		}
	}

	if i.MatchAny(domain) {
		return true
	}

	for _, m := range i.matchers {
		if m.Match(domain) {
			return true
		}
	}
	return false
}

func (i *IndexMatcherToDomainSet) addMatcher(matcher strmatcher.Matcher) {
	i.matchersLock.Lock()
	defer i.matchersLock.Unlock()
	for index, m := range i.removedMatchers {
		if m.String() == matcher.String() {
			i.removedMatchers = slices.Delete(i.removedMatchers, index, index+1)
			break
		}
	}

	i.matchers = append(i.matchers, matcher)
}

func (i *IndexMatcherToDomainSet) removeMatcher(matcher strmatcher.Matcher) {
	i.matchersLock.Lock()
	defer i.matchersLock.Unlock()

	for index, m := range i.matchers {
		if m.String() == matcher.String() {
			i.matchers = slices.Delete(i.matchers, index, index+1)
			break
		}
	}

	i.removedMatchers = append(i.removedMatchers, matcher)
}

type domainSets []i.DomainSet

func (l domainSets) Match(domain string) bool {
	for _, m := range l {
		if m.Match(domain) {
			return true
		}
	}
	return false
}

type GreatDomainSet struct {
	inMacthers []string
	exMacthers []string
	geo        i.GeoHelper
}

func NewGreatDomainSet(im []string, es []string, geo i.GeoHelper) *GreatDomainSet {
	return &GreatDomainSet{
		inMacthers: im,
		exMacthers: es,
		geo:        geo,
	}
}

func (d *GreatDomainSet) Match(domain string) bool {
	if d.exMacthers != nil {
		for _, m := range d.exMacthers {
			if d.geo.MatchDomain(domain, m) {
				return false
			}
		}
	}
	if d.inMacthers != nil {
		for _, m := range d.inMacthers {
			if d.geo.MatchDomain(domain, m) {
				return true
			}
		}
	}
	return false
}

type NullDomainSet struct{}

func (d *NullDomainSet) Match(domain string) bool {
	return false
}

type DomainSet struct {
	m    strmatcher.IndexMatcher
	tags []string
	h    i.GeoHelper
}

func NewDomainSet(tags []string, h i.GeoHelper, domains ...*geo.Domain) (*DomainSet, error) {
	d := &DomainSet{
		tags: tags,
		h:    h,
	}
	if len(domains) > 0 {
		d.m = strmatcher.NewMphIndexMatcher()
		for _, domain := range domains {
			matcher, err := geo.ToStrMatcher(domain)
			if err != nil {
				return nil, fmt.Errorf("failed to create domain matcher: %w", err)
			}
			d.m.Add(matcher)
		}
		if err := d.m.Build(); err != nil {
			return nil, fmt.Errorf("failed to build domain matcher: %w", err)
		}
	}
	return d, nil
}

func (d *DomainSet) Match(domain string) bool {
	if d.m != nil && d.m.MatchAny(domain) {
		return true
	}
	if d.h != nil {
		for _, tag := range d.tags {
			if d.h.MatchDomain(domain, tag) {
				return true
			}
		}
	}
	return false
}
