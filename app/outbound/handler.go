// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package outbound

import (
	"errors"

	"github.com/5vnetwork/vx-core/app/util"
	"github.com/5vnetwork/vx-core/i"
)

var ErrIpv6NotSupported = errors.New("ipv6 not supported")

type HandlerWithSupport6Info struct {
	i.Outbound
	util.IPv6SupportChangeNotifier
	support6 bool
}

func NewHandlerWithSupport6Info(outbound i.Outbound, support6 bool) *HandlerWithSupport6Info {
	return &HandlerWithSupport6Info{
		Outbound: outbound,
		support6: support6,
	}
}

func (h *HandlerWithSupport6Info) Support6() bool {
	return h.support6
}

func (h *HandlerWithSupport6Info) SetSupport6(support6 bool) {
	if h.support6 == support6 {
		return
	}
	h.support6 = support6
	h.Notify()
}
