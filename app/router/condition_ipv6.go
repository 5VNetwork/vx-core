package router

import (
	"context"
	"net"

	"github.com/5vnetwork/vx-core/common/session"
)

type Ipv6Matcher struct{}

func (m *Ipv6Matcher) Apply(c context.Context, info *session.Info, rw interface{}) (interface{}, bool) {
	ip := info.GetTargetIP()
	return rw, ip != nil && len(ip) == net.IPv6len && ip.To4() == nil
}
