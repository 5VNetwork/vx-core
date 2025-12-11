package selector

import (
	"math/rand"
	"sync"

	"github.com/5vnetwork/vx-core/common/session"
	"github.com/5vnetwork/vx-core/i"
)

type RandomBanlancer struct {
	lock        sync.RWMutex
	allHandlers []i.HandlerWith6Info
	handler6    []i.HandlerWith6Info
}

func NewRandomBanlancer() *RandomBanlancer {
	return &RandomBanlancer{
		allHandlers: make([]i.HandlerWith6Info, 0),
		handler6:    make([]i.HandlerWith6Info, 0),
	}
}

func (b *RandomBanlancer) Support6() bool {
	b.lock.RLock()
	defer b.lock.RUnlock()
	return len(b.handler6) > 0
}

func (b *RandomBanlancer) GetHandler(info *session.Info) i.Outbound {
	b.lock.RLock()
	defer b.lock.RUnlock()

	if len(b.allHandlers) == 0 {
		return nil
	}
	if len(b.allHandlers) == 1 {
		return b.allHandlers[0]
	}

	if (info.Target.Address != nil && info.Target.Address.Family().IsIPv6()) || (info.FakeIP != nil && info.FakeIP.To4() == nil) {
		if len(b.handler6) > 0 {
			return b.handler6[rand.Intn(len(b.handler6))]
		}
	}

	return b.allHandlers[rand.Intn(len(b.allHandlers))]
}

func (b *RandomBanlancer) UpdateHandlers(handlers []i.HandlerWith6Info) {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.allHandlers = make([]i.HandlerWith6Info, 0)
	b.handler6 = make([]i.HandlerWith6Info, 0)
	for _, h := range handlers {
		if h.Support6() {
			b.handler6 = append(b.handler6, h)
		}
		b.allHandlers = append(b.allHandlers, h)
	}
}
