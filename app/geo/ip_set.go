package geo

import (
	"net"

	"github.com/5vnetwork/vx-core/common/geo"
	"github.com/5vnetwork/vx-core/i"
)

type GreatIPSet struct {
	InMatchers ipSets
	ExMatcher  ipSets
}

func NewGreatIPSet(im []i.IPSet, es []i.IPSet) *GreatIPSet {
	return &GreatIPSet{
		InMatchers: im,
		ExMatcher:  es,
	}
}

type ipSets []i.IPSet

func (l ipSets) Match(ip net.IP) bool {
	for _, m := range l {
		if m.Match(ip) {
			return true
		}
	}
	return false
}

func (l *GreatIPSet) Match(ip net.IP) bool {
	if l.ExMatcher != nil && l.ExMatcher.Match(ip) {
		return false
	}
	return l.InMatchers.Match(ip)
}

type IpSet struct {
	tags []string
	h    i.GeoHelper
	m    *geo.IPMatcher
}

func NewIPSet(tags []string, h i.GeoHelper, cidrs ...*geo.CIDR) (*IpSet, error) {
	i := &IpSet{
		tags: tags,
		h:    h,
	}
	if len(cidrs) > 0 {
		m, err := geo.NewIPMatcherFromGeoCidrs(cidrs, false)
		if err != nil {
			return nil, err
		}
		i.m = m
	}
	return i, nil
}

func (i *IpSet) Match(ip net.IP) bool {
	if i.m != nil {
		if i.m.Match(ip) {
			return true
		}
	}
	if i.h != nil {
		for _, tag := range i.tags {
			if i.h.MatchIP(ip, tag) {
				return true
			}
		}
	}
	return false
}
