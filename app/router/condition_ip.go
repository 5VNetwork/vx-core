// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package router

import (
	"context"

	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/session"
	"github.com/5vnetwork/vx-core/i"
	"github.com/rs/zerolog/log"
)

// a IpMatcher consists of a list of geoIpMatcher, each geoIpMatcher is created from
// a geo.GeoIP which corresponds to ips of a specific country.
type IpMatcher struct {
	MatchSourceIp bool
	IpSet         i.IPSet
	IpResolver    i.IPResolver
	ResolveHard   bool
	ResolveSoft   bool
}

func (m *IpMatcher) Apply(c context.Context, info *session.Info, rw interface{}) (interface{}, bool) {
	var ip net.IP
	if m.MatchSourceIp {
		ip = info.GetSourceIPs()
	} else {
		ip = info.GetTargetIP()
		if ip == nil && (m.ResolveHard || m.ResolveSoft) && info.GetTargetDomain() != "" {
			ips, _ := m.IpResolver.LookupIP(c, info.GetTargetDomain())
			if len(ips) > 0 {
				for i, ip := range ips {
					if !m.IpSet.Match(ip) {
						log.Ctx(c).Error().Int("index", i).Any("ip", ip).Msg("ip not matched")
						if m.ResolveHard {
							return rw, false
						}
						continue
					} else {
						log.Ctx(c).Info().Int("index", i).Any("ip", ip).Msg("ip matched")
						if m.ResolveSoft {
							info.Target.Address = net.IPAddress(ips[i])
							return rw, true
						}
					}
				}
				// this means no ip matches
				if m.ResolveSoft {
					return rw, false
				}
				// this means all ips match
				return rw, true
			} else {
				// no ip resolved
				return rw, false
			}
		}
	}
	if len(ip) > 0 && m.IpSet.Match(ip) {
		return rw, true
	}
	return rw, false
}
