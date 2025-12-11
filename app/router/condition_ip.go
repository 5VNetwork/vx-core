package router

import (
	"context"
	"net"

	"github.com/5vnetwork/vx-core/common/session"
	"github.com/5vnetwork/vx-core/i"
)

// a IpMatcher consists of a list of geoIpMatcher, each geoIpMatcher is created from
// a geo.GeoIP which corresponds to ips of a specific country.
type IpMatcher struct {
	MatchSourceIp bool
	IpSet         i.IPSet
	IpResolver    i.IPResolver
	Resolve       bool
}

func (m *IpMatcher) Apply(c context.Context, info *session.Info, rw interface{}) (interface{}, bool) {
	var ip net.IP
	if m.MatchSourceIp {
		ip = info.GetSourceIPs()
	} else {
		ip = info.GetTargetIP()
		if ip == nil && m.Resolve && info.GetTargetDomain() != "" {
			ips, _ := m.IpResolver.LookupIP(c, info.GetTargetDomain())
			if len(ips) > 0 {
				for _, ip := range ips {
					if !m.IpSet.Match(ip) {
						return rw, false
					}
				}
				return rw, true
			}
		}
	}
	if len(ip) > 0 && m.IpSet.Match(ip) {
		return rw, true
	}
	return rw, false
}
