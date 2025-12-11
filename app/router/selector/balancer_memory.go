package selector

import (
	"math/rand"
	"sync"

	"github.com/5vnetwork/vx-core/common/cache"
	"github.com/5vnetwork/vx-core/common/dns"
	"github.com/5vnetwork/vx-core/common/session"
	"github.com/5vnetwork/vx-core/i"
	"github.com/rs/zerolog/log"
)

type MemoryBalancer struct {
	lock               sync.RWMutex
	appDomainToHandler cache.Lru // key is app or domain, value is handler name
	handlers           map[string]i.HandlerWith6Info
	allHandlers        []i.HandlerWith6Info
	handlerSupport6    []i.HandlerWith6Info
}

func NewMemoryBalancer() *MemoryBalancer {
	return &MemoryBalancer{
		appDomainToHandler: cache.NewLru(1000),
		allHandlers:        make([]i.HandlerWith6Info, 0),
		handlerSupport6:    make([]i.HandlerWith6Info, 0),
		handlers:           make(map[string]i.HandlerWith6Info),
	}
}

func (b *MemoryBalancer) Support6() bool {
	b.lock.RLock()
	defer b.lock.RUnlock()
	return len(b.handlerSupport6) > 0
}

func (b *MemoryBalancer) UpdateHandlers(handlers []i.HandlerWith6Info) {
	b.lock.Lock()
	defer b.lock.Unlock()

	newAppToHandler := cache.NewLru(1000)
	newAllHandlers := make([]i.HandlerWith6Info, 0)
	newHandler6 := make([]i.HandlerWith6Info, 0)
	newHandlers := make(map[string]i.HandlerWith6Info)
	for _, h := range handlers {
		newHandlers[h.Tag()] = h
		if h.Support6() {
			newHandler6 = append(newHandler6, h)
		}
		newAllHandlers = append(newAllHandlers, h)
	}
	for _, h := range handlers {
		appId, ok := b.appDomainToHandler.GetKeyFromValue(h.Tag())
		if ok {
			newAppToHandler.Put(appId, h.Tag())
		}
	}
	b.appDomainToHandler = newAppToHandler
	b.handlers = newHandlers
	b.allHandlers = newAllHandlers
	b.handlerSupport6 = newHandler6
	log.Debug().Int("handler4", len(newAllHandlers)).Int("handler46", len(newHandler6)).Msg("memory balancer")
}

func (b *MemoryBalancer) GetHandler(info *session.Info) i.Outbound {
	b.lock.RLock()
	defer b.lock.RUnlock()

	if len(b.allHandlers) == 0 {
		return nil
	}

	domain := b.getDomain(info)
	if domain != "" {
		rootDomain := dns.RootDomain(domain)
		if h, ok := b.appDomainToHandler.Get(rootDomain); ok {
			return b.handlers[h.(string)]
		}
		if (info.Target.Address != nil && info.Target.Address.Family().IsIPv6()) || (info.FakeIP != nil && info.FakeIP.To4() == nil) {
			if len(b.handlerSupport6) > 0 {
				h := b.handlerSupport6[rand.Intn(len(b.handlerSupport6))]
				b.appDomainToHandler.Put(rootDomain, h.Tag())
				return h
			}
		}
		h := b.allHandlers[rand.Intn(len(b.allHandlers))]
		b.appDomainToHandler.Put(rootDomain, h.Tag())
		return h
	}
	if info.GetAppId() != "" {
		app := getApp(info.GetAppId())
		if h, ok := b.appDomainToHandler.Get(app); ok {
			return b.handlers[h.(string)]
		}
		if (info.Target.Address != nil && info.Target.Address.Family().IsIPv6()) || (info.FakeIP != nil && info.FakeIP.To4() == nil) {
			if len(b.handlerSupport6) > 0 {
				h := b.handlerSupport6[rand.Intn(len(b.handlerSupport6))]
				b.appDomainToHandler.Put(app, h.Tag())
				return h
			}
		}
		h := b.allHandlers[rand.Intn(len(b.allHandlers))]
		b.appDomainToHandler.Put(app, h.Tag())
		return h
	}
	if (info.Target.Address != nil && info.Target.Address.Family().IsIPv6()) || (info.FakeIP != nil && info.FakeIP.To4() == nil) {
		if len(b.handlerSupport6) > 0 {
			return b.handlerSupport6[rand.Intn(len(b.handlerSupport6))]
		}
	}
	return b.allHandlers[rand.Intn(len(b.allHandlers))]
}

func (b *MemoryBalancer) getDomain(info *session.Info) string {
	return info.GetTargetDomain()
}
