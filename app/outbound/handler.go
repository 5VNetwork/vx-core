package outbound

import (
	"errors"

	"github.com/5vnetwork/vx-core/app/util"
	"github.com/5vnetwork/vx-core/i"
)

type HandlerWithStats struct {
	i.Outbound
	Stats *OutboundHandlerStats
}

var ErrIpv6NotSupported = errors.New("ipv6 not supported")

func (i *HandlerWithStats) GetHandlerStats() *OutboundHandlerStats {
	return i.Stats
}

type HandlerWithSupport6Info struct {
	i.Outbound
	util.IPv6SupportChangeNotifier
	support6 bool
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
