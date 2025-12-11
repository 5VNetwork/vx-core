package router

import (
	"context"

	"github.com/5vnetwork/vx-core/common/session"
)

type AllMatcher struct {
	domainMatcher *DomainMatcher
	ipMatcher     *IpMatcher
	appIdMatcher  *AppIdMatcher
}

func (m *AllMatcher) Apply(c context.Context, sInfo *session.Info, rw interface{}) (interface{}, bool) {
	rw0, ok := m.domainMatcher.Apply(c, sInfo, rw)
	rw = rw0
	if ok {
		return rw, true
	}
	rw0, ok = m.ipMatcher.Apply(c, sInfo, rw)
	rw = rw0
	if ok {
		return rw, true
	}
	rw0, ok = m.appIdMatcher.Apply(c, sInfo, rw)
	rw = rw0
	if ok {
		return rw, true
	}
	return rw, false
}
