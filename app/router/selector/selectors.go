package selector

import (
	"sync"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/outbound"
	"github.com/5vnetwork/vx-core/app/util"
	"github.com/5vnetwork/vx-core/app/xsqlite"
	"github.com/5vnetwork/vx-core/i"
	"github.com/rs/zerolog/log"
)

type Selectors struct {
	lock      sync.RWMutex
	started   bool
	selectors map[string]*Selector
	// notify listeners when any selector's ipv6 support changed or any selector is added or removed
	i.IPv6SupportChangeSubject
}

func NewSelectors() *Selectors {
	return &Selectors{
		selectors:                make(map[string]*Selector),
		IPv6SupportChangeSubject: &util.IPv6SupportChangeNotifier{},
	}
}

func (s *Selectors) AddSelector(selector *Selector) {
	s.lock.Lock()
	defer s.lock.Unlock()
	existing, ok := s.selectors[selector.tag]
	if ok {
		existing.Close()
	}
	s.selectors[selector.tag] = selector
	selector.Register(s)
	if s.started {
		selector.Start()
	}
	s.OnIPv6SupportChanged()
}

func (s *Selectors) RemoveAllSelectors() {
	s.lock.Lock()
	defer s.lock.Unlock()
	for _, selector := range s.selectors {
		selector.Close()
	}
	s.selectors = make(map[string]*Selector)
}

func (s *Selectors) RemoveSelector(tag string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	selector := s.selectors[tag]
	if selector != nil {
		selector.Close()
		delete(s.selectors, tag)
	}
	s.OnIPv6SupportChanged()
}

func (s *Selectors) OnIPv6SupportChanged() {
	s.Notify()
}

func (s *Selectors) GetAllSelectors() []*Selector {
	s.lock.RLock()
	defer s.lock.RUnlock()
	var ret []*Selector
	for _, selector := range s.selectors {
		ret = append(ret, selector)
	}
	return ret
}

func (s *Selectors) GetSelector(tag string) *Selector {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.selectors[tag]
}

func (s *Selectors) Start() error {
	s.started = true
	for _, selector := range s.selectors {
		selector.Start()
	}
	return nil
}

func (s *Selectors) Close() error {
	for _, selector := range s.selectors {
		selector.Close()
	}
	return nil
}

func (s *Selectors) OnHandlerChanged() {
	log.Debug().Msg("Selectors OnHandlerChanged")
	for _, selector := range s.selectors {
		selector.OnHandlerChanged()
	}
}

func (s *Selectors) OnHandlerSpeedChanged(tag string, speed int32) {
	for _, selector := range s.selectors {
		selector.OnHandlerSpeedChanged(tag, speed)
	}
}

type HandlersBeingUsedUpdate func([]string)

type SelectorConfig struct {
	*configs.SelectorConfig
	CreateHandler             CreateHandlerFunc
	HandlerErrorChangeSubject HandlerErrorChangeSubject
	Tester                    Tester
	Database                  Db
	OutboundManager           *outbound.Manager
	OnHandlerBeingUsedChange  HandlersBeingUsedUpdate
	LandHandlers              []*xsqlite.OutboundHandler
}

func NewSelector(config SelectorConfig) *Selector {
	log.Debug().Str("tag", config.SelectorConfig.GetTag()).Msg("NewSelector")
	// TODO: tmp solution. ONlY proxy selector update
	if config.SelectorConfig.GetTag() != "代理" {
		config.OnHandlerBeingUsedChange = nil
	}
	sc := config.SelectorConfig

	var balancer Balancer
	switch sc.BalanceStrategy {
	case configs.SelectorConfig_RANDOM:
		balancer = NewRandomBanlancer()
	case configs.SelectorConfig_MEMORY:
		balancer = NewMemoryBalancer()
	}

	var filter Filter
	if sc.SelectFromOm {
		filter = NewOmFilter(sc.GetFilter(), config.OutboundManager)
	} else {
		filter = NewDbFilter(config.Database, sc.GetFilter(),
			config.LandHandlers, config.CreateHandler)
	}

	var se selectStrategy
	switch sc.Strategy {
	case configs.SelectorConfig_ALL:
		se = &allStrategy{}
	case configs.SelectorConfig_ALL_OK:
		se = &allOkStrategy{}
	case configs.SelectorConfig_MOST_THROUGHPUT:
		se = &highestThroughputStrategy{}
	case configs.SelectorConfig_LEAST_PING:
		se = &leastPingStrategy{}
	}
	selector0 := newSelector(selectorConfig{
		Tag:                      sc.Tag,
		Strategy:                 se,
		Filter:                   filter,
		Balancer:                 balancer,
		Tester:                   config.Tester,
		OnHandlerBeingUsedChange: config.OnHandlerBeingUsedChange,
		Dispatcher:               config.HandlerErrorChangeSubject,
	})
	return selector0
}
