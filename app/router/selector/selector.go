// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package selector

import (
	"context"
	"slices"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/5vnetwork/vx-core/app/util"
	"github.com/5vnetwork/vx-core/common/session"
	"github.com/5vnetwork/vx-core/common/task"
	"github.com/5vnetwork/vx-core/i"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Selector struct {
	tag      string
	filter   atomic.Value
	strategy selectStrategy

	balancerLock sync.RWMutex
	balancer     Balancer

	handlersLock      sync.RWMutex
	handlersBeingUsed []*handler
	filteredHandlers  []outHandler

	handlerBeingTestedLock sync.RWMutex
	handlerBeingTested     map[string]struct{}

	// When there is no handler usable, enter fast recovery mode:
	// test all unusable handlers every 10 seconds
	isRecovery                                 bool
	taskLock                                   sync.RWMutex
	periodicTestUnusableHandlersInFastRevovery *task.PeriodicTask
	FastRecoveryChangeNotifier

	tester Tester
	util.IPv6SupportChangeNotifier

	periodicTestSpeed            *task.PeriodicTask
	periodicTestPing             *task.PeriodicTask
	periodicTestUnusableHandlers *task.PeriodicTask

	onUpdate HandlersBeingUsedUpdate

	dispatcher HandlerErrorChangeSubject

	ctx    context.Context
	cancel context.CancelFunc
	closed bool
}

type selectorConfig struct {
	Tag                      string
	Strategy                 selectStrategy
	Filter                   Filter
	Balancer                 Balancer
	Tester                   Tester
	OnHandlerBeingUsedChange HandlersBeingUsedUpdate
	Dispatcher               HandlerErrorChangeSubject
}

func newSelector(config selectorConfig) *Selector {
	ctx, cancel := context.WithCancel(context.Background())
	s := &Selector{
		tag:                config.Tag,
		strategy:           config.Strategy,
		balancer:           config.Balancer,
		tester:             config.Tester,
		onUpdate:           config.OnHandlerBeingUsedChange,
		dispatcher:         config.Dispatcher,
		ctx:                ctx,
		cancel:             cancel,
		handlerBeingTested: make(map[string]struct{}),
	}
	s.filter.Store(config.Filter)
	return s
}

func (s *Selector) Start() error {
	if _, ok := s.strategy.(*highestThroughputStrategy); ok {
		s.periodicTestSpeed = task.NewPeriodicTask(time.Minute*60, s.TestSpeedAll)
		s.periodicTestUnusableHandlers = task.NewPeriodicTask(time.Minute*10, s.TestAllUnusable)
	}
	if _, ok := s.strategy.(*topThroughputStrategy); ok {
		s.periodicTestSpeed = task.NewPeriodicTask(time.Minute*60, s.TestSpeedAll)
		s.periodicTestUnusableHandlers = task.NewPeriodicTask(time.Minute*10, s.TestAllUnusable)
	}
	if _, ok := s.strategy.(*leastPingStrategy); ok {
		s.periodicTestPing = task.NewPeriodicTask(time.Minute*10, s.TestPingAll)
	}
	if _, ok := s.strategy.(*topPingStrategy); ok {
		s.periodicTestPing = task.NewPeriodicTask(time.Minute*10, s.TestPingAll)
	}
	if _, ok := s.strategy.(*allOkStrategy); ok {
		s.periodicTestUnusableHandlers = task.NewPeriodicTask(time.Minute*10, s.TestAllUnusable)
	}
	if _, ok := s.strategy.(*allStrategy); !ok && s.dispatcher != nil {
		s.dispatcher.AddHandlerErrorObserver(s)
	}
	s.Load()
	return nil
}

func (s *Selector) Close() error {
	s.closed = true
	s.cancel()
	if s.periodicTestSpeed != nil {
		go s.periodicTestSpeed.Close()
	}
	if s.periodicTestPing != nil {
		go s.periodicTestPing.Close()
	}
	if s.periodicTestUnusableHandlers != nil {
		go s.periodicTestUnusableHandlers.Close()
	}
	if s.periodicTestUnusableHandlersInFastRevovery != nil {
		go s.periodicTestUnusableHandlersInFastRevovery.Close()
	}
	if s.dispatcher != nil {
		s.dispatcher.RemoveHandlerErrorObserver(s)
	}
	return nil
}

func (s *Selector) Tag() string {
	return s.tag
}

func (s *Selector) GetHandler(info *session.Info) i.Outbound {
	return s.getBalancer().GetHandler(info)
}

// reload handlers
func (s *Selector) Load() {
	handlers, err := s.filter.Load().(Filter).GetHandlers()
	if err != nil {
		log.Error().Err(err).Msg("get filtered handlers")
		// if strings.Contains(err.Error(), "no such file or directory") {
		// 	log.Fatal().Err(err).Msg("no such file or directory")
		// }
		return
	}
	log.Debug().Int("len", len(handlers)).Msg("filtered handlers")

	s.handlersLock.Lock()
	handlersToBeTestedForSpeed := make([]outHandler, 0, len(handlers))
	handlersToBeTestedForIpv6 := make([]outHandler, 0, len(handlers))
	handlersToBeTestedForPing := make([]outHandler, 0, len(handlers))
	for _, os := range handlers {
		index := slices.IndexFunc(s.filteredHandlers, func(h outHandler) bool {
			return h.Name() == os.Name()
		})
		if index != -1 {
			existing := s.filteredHandlers[index]
			if os.GetOk() == 0 {
				os.SetOk(existing.GetOk())
			}
			if os.GetSpeed() == 0 {
				os.SetSpeed(existing.GetSpeed())
			}
			if os.GetPing() == 0 {
				os.SetPing(existing.GetPing())
			}
			if os.GetSupport6() == 0 {
				os.SetSupport6(existing.GetSupport6())
			}
		}
		if os.GetSupport6() == 0 {
			handlersToBeTestedForIpv6 = append(handlersToBeTestedForIpv6, os)
		}
		if _, ok := s.strategy.(*highestThroughputStrategy); ok && os.GetSpeed() == 0 {
			handlersToBeTestedForSpeed = append(handlersToBeTestedForSpeed, os)
		}
		if _, ok := s.strategy.(*topThroughputStrategy); ok && os.GetSpeed() == 0 {
			handlersToBeTestedForSpeed = append(handlersToBeTestedForSpeed, os)
		}
		if _, ok := s.strategy.(*leastPingStrategy); ok && os.GetPing() == 0 {
			handlersToBeTestedForPing = append(handlersToBeTestedForPing, os)
		}
		if _, ok := s.strategy.(*topPingStrategy); ok && os.GetPing() == 0 {
			handlersToBeTestedForPing = append(handlersToBeTestedForPing, os)
		}
		if _, ok := s.strategy.(*allOkStrategy); ok && os.GetOk() == 0 {
			handlersToBeTestedForPing = append(handlersToBeTestedForPing, os)
		}
	}
	s.filteredHandlers = handlers
	s.handlersLock.Unlock()

	if len(handlersToBeTestedForIpv6) > 0 {
		go func() {
			s.testItems(handlersToBeTestedForIpv6, TestHandler6)
			s.setHandlers()
		}()
	}
	if len(handlersToBeTestedForSpeed) > 0 {
		go func() {
			s.testItems(handlersToBeTestedForSpeed, TestHandlerSpeed)
			s.setHandlers()
		}()
	}
	if len(handlersToBeTestedForPing) > 0 {
		go func() {
			s.testItems(handlersToBeTestedForPing, TestHandlerPing)
			s.setHandlers()
		}()
	}

	s.setHandlers()
}

func (s *Selector) getBalancer() Balancer {
	s.taskLock.RLock()
	defer s.taskLock.RUnlock()
	return s.balancer
}

func (s *Selector) setBalancer(balancer Balancer) {
	s.taskLock.Lock()
	defer s.taskLock.Unlock()
	s.balancer = balancer
}

func (s *Selector) setHandlers() {
	if s.closed {
		return
	}

	filteredHandlers := s.getOutHandlers()
	if len(filteredHandlers) == 0 {
		log.Warn().Msg("no handlers")
		return
	}

	selectedHandlers := s.strategy.Select(filteredHandlers)

	if len(selectedHandlers) == 0 {
		s.enterRecoveryIfNot()
		// use all handlers
		selectedHandlers = filteredHandlers
	} else if s.isRecovery {
		s.exitRecovery()
	}

	handlerToBeUsed := make([]i.HandlerWith6Info, 0, len(selectedHandlers))
	handlersBeingUsed := make([]*handler, 0, len(selectedHandlers))
	for _, selectedHandler := range selectedHandlers {
		ha, ok := selectedHandler.(*handler)
		if !ok {
			h, err := selectedHandler.GetHandler()
			if err != nil {
				log.Error().Err(err).Msg("get handler")
				selectedHandler.SetOk(-1)
				continue
			}
			ha = &handler{
				Outbound:   h,
				outHandler: selectedHandler,
			}
		}
		handlerToBeUsed = append(handlerToBeUsed, ha)
		handlersBeingUsed = append(handlersBeingUsed, ha)
	}

	s.updateBalancerHandlers(handlerToBeUsed)

	s.handlersLock.Lock()
	s.handlersBeingUsed = handlersBeingUsed
	s.handlersLock.Unlock()

	if s.onUpdate != nil {
		handlerNames := make([]string, 0, len(handlersBeingUsed))
		for _, h := range handlersBeingUsed {
			handlerNames = append(handlerNames, h.Tag())
		}
		go s.onUpdate(handlerNames)
	}
}

func (s *Selector) GetHandlersBeingUsed() []string {
	s.handlersLock.RLock()
	defer s.handlersLock.RUnlock()
	handlerNames := make([]string, 0, len(s.handlersBeingUsed))
	for _, h := range s.handlersBeingUsed {
		handlerNames = append(handlerNames, h.Tag())
	}
	return handlerNames
}

func (s *Selector) GetFilter() Filter {
	return s.filter.Load().(Filter)
}

func (s *Selector) UpdateFilter(filter Filter) {
	s.filter.Store(filter)
	s.Load()
}

func (s *Selector) UpdateBalancer(balancer Balancer) {
	s.setBalancer(balancer)

	s.handlersLock.RLock()
	handlerToBeUsed := make([]i.HandlerWith6Info, 0, len(s.handlersBeingUsed))
	for _, h := range s.handlersBeingUsed {
		handlerToBeUsed = append(handlerToBeUsed, h)
	}
	s.handlersLock.RUnlock()
	s.updateBalancerHandlers(handlerToBeUsed)
}

func (s *Selector) updateBalancerHandlers(handlerToBeUsed []i.HandlerWith6Info) {
	s.balancerLock.RLock()
	defer s.balancerLock.RUnlock()
	oldSupport6 := s.balancer.Support6()
	s.balancer.UpdateHandlers(handlerToBeUsed)
	newSupport6 := s.balancer.Support6()
	if oldSupport6 != newSupport6 {
		go s.IPv6SupportChangeNotifier.Notify()
	}
	log.Debug().Func(func(e *zerolog.Event) {
		for i, h := range handlerToBeUsed {
			e.Str(strconv.Itoa(i), h.Tag())
		}
	}).Bool("support6", newSupport6).Msg("handlers being used")
}

func (s *Selector) OnHandlerError(tag string, err error) {
	log.Debug().Str("tag", tag).Err(err).Msg("on handler error")
	s.handlersLock.RLock()
	var handler *handler
	for _, h := range s.handlersBeingUsed {
		if h.Tag() == tag {
			handler = h
			break
		}
	}
	s.handlersLock.RUnlock()

	if handler == nil {
		return
	}

	if handler.GetOk() > 0 {
		s.handlerBeingTestedLock.Lock()
		_, ok := s.handlerBeingTested[tag]
		if ok {
			s.handlerBeingTestedLock.Unlock()
			return
		} else {
			s.handlerBeingTested[tag] = struct{}{}
			s.handlerBeingTestedLock.Unlock()
		}

		TestHandlerUsable(s.ctx, s.tester, handler)
		usable := handler.outHandler.GetOk() > 0
		if !usable {
			s.setHandlers()
		}

		s.handlerBeingTestedLock.Lock()
		delete(s.handlerBeingTested, tag)
		s.handlerBeingTestedLock.Unlock()
	}
}

func (s *Selector) testItems(items []outHandler,
	testFunc func(ctx context.Context, s Tester, item outHandler)) {
	// Process in batches of 10
	batchSize := 10
	for i := 0; i < len(items); i += batchSize {
		var wg sync.WaitGroup
		end := i + batchSize
		if end > len(items) {
			end = len(items)
		}
		// Process current batch
		for _, item := range items[i:end] {
			wg.Add(1)
			go func(item outHandler) {
				defer wg.Done()
				testFunc(s.ctx, s.tester, item)
				if s.isRecovery && item.GetOk() > 0 {
					s.setHandlers()
				}
			}(item)
		}
		// Wait for current batch to complete before starting next batch
		wg.Wait()
	}
}

func (s *Selector) getOutHandlers() []outHandler {
	s.handlersLock.RLock()
	defer s.handlersLock.RUnlock()
	return s.filteredHandlers
}

func (s *Selector) enterRecoveryIfNot() {
	s.taskLock.Lock()
	defer s.taskLock.Unlock()
	if s.isRecovery {
		return
	}
	log.Info().Msg("enter recovery")
	s.isRecovery = true
	if s.periodicTestUnusableHandlersInFastRevovery == nil {
		s.periodicTestUnusableHandlersInFastRevovery = task.NewPeriodicTask(
			time.Second*10, func() {
				log.Debug().Msg("TestAllUsabe")
				s.testItems(s.getOutHandlers(), TestHandlerUsable)
			}, task.WithStartImmediately())
		s.periodicTestUnusableHandlersInFastRevovery.Start()
	}
	s.FastRecoveryChangeNotifier.Notify(true)
}
func (s *Selector) exitRecovery() {
	log.Info().Msg("exit recovery")
	s.taskLock.Lock()
	s.isRecovery = false
	if s.periodicTestUnusableHandlersInFastRevovery != nil {
		go s.periodicTestUnusableHandlersInFastRevovery.Close()
		s.periodicTestUnusableHandlersInFastRevovery = nil
	}
	s.taskLock.Unlock()
	s.FastRecoveryChangeNotifier.Notify(false)
}

func (s *Selector) TestSpeedAll() {
	s.testItems(s.getOutHandlers(), TestHandlerSpeed)
	s.setHandlers()
}

func (s *Selector) TestPingAll() {
	s.testItems(s.getOutHandlers(), TestHandlerPing)
	s.setHandlers()
}

func (s *Selector) TestAllUnusable() {
	handlers := s.getOutHandlers()
	unusableHandlers := make([]outHandler, 0, len(handlers))
	for _, h := range handlers {
		if h.GetOk() <= 0 {
			unusableHandlers = append(unusableHandlers, h)
		}
	}
	s.testItems(unusableHandlers, TestHandlerUsable)
	s.setHandlers()
}

func (s *Selector) OnHandlerChanged() {
	s.Load()
}

func (s *Selector) ResetTestAllUnusableInterval(interval time.Duration) {
	if a := s.periodicTestUnusableHandlers; a != nil {
		a.ResetInterval(interval)
	}
}

func (s *Selector) OnHandlerSpeedChanged(tag string, speed int32) {
	s.handlersLock.RLock()
	index := slices.IndexFunc(s.filteredHandlers, func(h outHandler) bool {
		return h.Name() == tag
	})
	if index == -1 {
		s.handlersLock.RUnlock()
		return
	}
	handler := s.filteredHandlers[index]
	s.handlersLock.RUnlock()

	handler.SetOk(int(speed))
	handler.SetSpeed(int(speed))
	if _, ok := s.strategy.(*highestThroughputStrategy); ok {
		s.setHandlers()
	}
	if _, ok := s.strategy.(*topThroughputStrategy); ok {
		s.setHandlers()
	}
}

type selectStrategy interface {
	Select(handlers []outHandler) []outHandler
}

type leastPingStrategy struct{}

func (s *leastPingStrategy) Select(handlers []outHandler) []outHandler {
	if len(handlers) == 0 {
		return nil
	} else {
		var best outHandler
		for _, v := range handlers {
			if (v.GetOk() > 0) && best == nil {
				best = v
			} else {
				if v.GetOk() > 0 && v.GetPing() > 0 && v.GetPing() < best.GetPing() {
					best = v
				}
			}
		}
		if best == nil {
			for _, v := range handlers {
				if v.GetOk() == 0 {
					best = v
					break
				}
			}
		}
		if best == nil {
			return nil
		}
		return []outHandler{best}
	}
}

type highestThroughputStrategy struct{}

func (s *highestThroughputStrategy) Select(handlers []outHandler) []outHandler {
	if len(handlers) == 0 {
		return nil
	} else {
		var largest outHandler
		for _, v := range handlers {
			if (v.GetOk() > 0) && largest == nil {
				largest = v
			} else {
				if v.GetOk() > 0 && v.GetSpeed() > largest.GetSpeed() {
					largest = v
				}
			}
		}
		if largest == nil {
			for _, v := range handlers {
				if v.GetOk() == 0 {
					largest = v
					break
				}
			}
		}
		if largest == nil {
			return nil
		}
		return []outHandler{largest}
	}
}

type allStrategy struct{}

func (s *allStrategy) Select(handlers []outHandler) []outHandler {
	return handlers
}

type allOkStrategy struct{}

func (s *allOkStrategy) Select(handlers []outHandler) []outHandler {
	okHandlers := make([]outHandler, 0, len(handlers))
	for _, h := range handlers {
		if h.GetOk() > 0 {
			okHandlers = append(okHandlers, h)
		}
	}
	if len(okHandlers) == 0 {
		for _, h := range handlers {
			if h.GetOk() >= 0 {
				okHandlers = append(okHandlers, h)
			}
		}
	}
	return okHandlers
}

// topPingStrategy selects all nodes within 30% of the least ping (e.g. min 100ms -> select all with ping <= 130ms).
type topPingStrategy struct{}

func (s *topPingStrategy) Select(handlers []outHandler) []outHandler {
	if len(handlers) == 0 {
		return nil
	}
	minPing := -1
	for _, v := range handlers {
		if v.GetOk() > 0 && v.GetPing() > 0 {
			if minPing < 0 || v.GetPing() < minPing {
				minPing = v.GetPing()
			}
		}
	}
	if minPing < 0 {
		// no usable ping data, fall back to all with GetOk >= 0
		okHandlers := make([]outHandler, 0, len(handlers))
		for _, h := range handlers {
			if h.GetOk() >= 0 {
				okHandlers = append(okHandlers, h)
			}
		}
		return okHandlers
	}
	threshold := minPing + (minPing * 30 / 100) // 30% above least
	selected := make([]outHandler, 0, len(handlers))
	for _, h := range handlers {
		if h.GetOk() > 0 && h.GetPing() > 0 && h.GetPing() <= threshold {
			selected = append(selected, h)
		}
	}
	if len(selected) == 0 {
		for _, h := range handlers {
			if h.GetOk() >= 0 {
				selected = append(selected, h)
			}
		}
	}
	return selected
}

// topThroughputStrategy selects all nodes with throughput >= 70% of the highest (e.g. max 100 -> select all with speed >= 70).
type topThroughputStrategy struct{}

func (s *topThroughputStrategy) Select(handlers []outHandler) []outHandler {
	if len(handlers) == 0 {
		return nil
	}
	var maxSpeed int
	for _, v := range handlers {
		if v.GetOk() > 0 && v.GetSpeed() > maxSpeed {
			maxSpeed = v.GetSpeed()
		}
	}
	if maxSpeed == 0 {
		// no usable speed data, fall back to all with GetOk >= 0
		okHandlers := make([]outHandler, 0, len(handlers))
		for _, h := range handlers {
			if h.GetOk() >= 0 {
				okHandlers = append(okHandlers, h)
			}
		}
		return okHandlers
	}
	threshold := maxSpeed * 70 / 100 // 70% of highest
	selected := make([]outHandler, 0, len(handlers))
	for _, h := range handlers {
		if h.GetOk() > 0 && h.GetSpeed() >= threshold {
			selected = append(selected, h)
		}
	}
	if len(selected) == 0 {
		for _, h := range handlers {
			if h.GetOk() >= 0 {
				selected = append(selected, h)
			}
		}
	}
	return selected
}
