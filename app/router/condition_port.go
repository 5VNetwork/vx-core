package router

import (
	"context"

	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/session"
)

type PortMatcher struct {
	portRanges []*net.PortRange
	onSource   bool
}

func NewPortMatcher(portRanges []*net.PortRange, onSource bool) *PortMatcher {
	return &PortMatcher{
		portRanges: portRanges,
		onSource:   onSource,
	}
}

func (m *PortMatcher) Apply(c context.Context, info *session.Info, rw interface{}) (interface{}, bool) {
	var portToMatch net.Port
	if m.onSource {
		portToMatch = info.GetSourcePort()
	} else {
		portToMatch = info.GetTargetPort()
	}
	for _, pr := range m.portRanges {
		if pr.Contains(portToMatch) {
			return rw, true
		}
	}
	return rw, false
}
