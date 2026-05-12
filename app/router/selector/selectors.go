// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package selector

import (
	"slices"
	"sync"
	"time"

	router "buf.build/gen/go/vvvvv/vx/protocolbuffers/go/vx/router"
	"github.com/5vnetwork/vx-core/app/util"
	"github.com/5vnetwork/vx-core/i"
	"github.com/rs/zerolog/log"
)

type Selectors struct {
	lock      sync.RWMutex
	started   bool
	selectors map[string]*Selector
	// notify listeners when any selector's ipv6 support changed or any selector is added or removed
	i.IPv6SupportChangeSubject
	SelectedHandlersChangeNotifier
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
	selector.onUpdate = func(handlers []string) {
		s.NotifySelectedHandlersChanged(selector.tag, handlers)
	}
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
	*router.SelectorConfig
	CreateHandler             i.HandlerFactory
	HandlerErrorChangeSubject HandlerErrorChangeSubject
	Tester                    Tester
	Filter                    Filter
	OnHandlerBeingUsedChange  HandlersBeingUsedUpdate
	HandlerInfo               handlerInfo
	SelectStrategy            selectStrategy
}

func NewSelector(config SelectorConfig) *Selector {
	log.Debug().Str("tag", config.SelectorConfig.GetTag()).Msg("NewSelector")
	sc := config.SelectorConfig

	var balancer Balancer
	switch sc.BalanceStrategy {
	case router.SelectorConfig_RANDOM:
		balancer = NewRandomBanlancer()
	case router.SelectorConfig_MEMORY:
		balancer = NewMemoryBalancer()
	}

	var se selectStrategy
	if config.SelectStrategy != nil {
		se = config.SelectStrategy
	} else {
		switch sc.Strategy {
		case router.SelectorConfig_ALL:
			se = &allStrategy{}
		case router.SelectorConfig_ALL_OK:
			se = &allOkStrategy{}
		case router.SelectorConfig_MOST_THROUGHPUT:
			se = &highestThroughputStrategy{}
		case router.SelectorConfig_LEAST_PING:
			se = &leastPingStrategy{}
		case router.SelectorConfig_TOP_PING:
			se = &topPingStrategy{}
		case router.SelectorConfig_TOP_THROUGHPUT:
			se = &TopThroughputStrategy{}
		}
	}

	selector0 := newSelector(selectorConfig{
		Tag:                      sc.Tag,
		Strategy:                 se,
		Filter:                   config.Filter,
		Balancer:                 balancer,
		Tester:                   config.Tester,
		OnHandlerBeingUsedChange: config.OnHandlerBeingUsedChange,
		Dispatcher:               config.HandlerErrorChangeSubject,
		HandlerInfo:              config.HandlerInfo,
		SpeedTestSize:            sc.GetSpeedTestSize(),
		SpeedTestInterval:        time.Duration(sc.GetSpeedTestInterval()) * time.Minute,
		PingTestInterval:         time.Duration(sc.GetPingTestInterval()) * time.Minute,
		UnusableTestInterval:     time.Duration(sc.GetUnusableTestInterval()) * time.Minute,
	})
	return selector0
}

type SelectedHandlersChangeNotifier struct {
	lock      sync.RWMutex
	observers []SelectedHandlersChangeObserver
}

type SelectedHandlersChangeObserver interface {
	OnSelectedHandlersChanged(tag string, handlers []string)
}

type OnSelectedHandlersChangedFunc func()

func (f OnSelectedHandlersChangedFunc) OnSelectedHandlersChanged() {
	f()
}

func (n *SelectedHandlersChangeNotifier) RegisterSelectedHandlersChangeObserver(observer SelectedHandlersChangeObserver) {
	n.lock.Lock()
	n.observers = append(n.observers, observer)
	n.lock.Unlock()
}

func (n *SelectedHandlersChangeNotifier) UnregisterSelectedHandlersChangeObserver(observer SelectedHandlersChangeObserver) {
	n.lock.Lock()
	defer n.lock.Unlock()
	for i, o := range n.observers {
		if o == observer {
			n.observers = slices.Delete(n.observers, i, i+1)
			break
		}
	}
}

func (n *SelectedHandlersChangeNotifier) NotifySelectedHandlersChanged(tag string, handlers []string) {
	n.lock.RLock()
	defer n.lock.RUnlock()
	for _, o := range n.observers {
		go o.OnSelectedHandlersChanged(tag, handlers)
	}
}
