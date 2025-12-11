package router

import (
	"context"

	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/session"
)

type NetworkMatcher struct {
	list [8]bool
}

func NewNetworkMatcher(network []net.Network) NetworkMatcher {
	var matcher NetworkMatcher
	for _, n := range network {
		matcher.list[int(n)] = true
	}
	return matcher
}

// Apply implements Condition.
func (v NetworkMatcher) Apply(c context.Context, info *session.Info, rw interface{}) (interface{}, bool) {
	return rw, v.list[int(info.GetNetwork())]
}
