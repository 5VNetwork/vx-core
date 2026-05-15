// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package selector

import (
	"context"
	"math/rand/v2"
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
	filteredHandlers  []OutHandler

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

	handlerInfo handlerInfo

	speedTestSize        uint32
	speedTestSizeMin     uint32
	speedTestSizeMax     uint32
	speedTestUseRange    bool
	speedTestInterval    time.Duration
	pingTestInterval     time.Duration
	unusableTestInterval time.Duration

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
	HandlerInfo              handlerInfo
	SpeedTestSize            uint32
	SpeedTestSizeMin         uint32
	SpeedTestSizeMax         uint32
	SpeedTestUseRange        bool
	SpeedTestInterval        time.Duration
	PingTestInterval         time.Duration
	UnusableTestInterval     time.Duration
}

func newSelector(config selectorConfig) *Selector {
	ctx, cancel := context.WithCancel(context.Background())
	s := &Selector{
		tag:                  config.Tag,
		strategy:             config.Strategy,
		balancer:             config.Balancer,
		tester:               config.Tester,
		onUpdate:             config.OnHandlerBeingUsedChange,
		dispatcher:           config.Dispatcher,
		ctx:                  ctx,
		cancel:               cancel,
		handlerInfo:          config.HandlerInfo,
		handlerBeingTested:   make(map[string]struct{}),
		speedTestSize:        config.SpeedTestSize,
		speedTestSizeMin:     config.SpeedTestSizeMin,
		speedTestSizeMax:     config.SpeedTestSizeMax,
		speedTestUseRange:    config.SpeedTestUseRange,
		speedTestInterval:    config.SpeedTestInterval,
		pingTestInterval:     config.PingTestInterval,
		unusableTestInterval: config.UnusableTestInterval,
	}
	if !s.speedTestUseRange && s.speedTestSize == 0 {
		s.speedTestSize = 1024 * 1024 // 1MB
	}
	if s.speedTestInterval == 0 {
		s.speedTestInterval = time.Minute * 60
	}
	if s.pingTestInterval == 0 {
		s.pingTestInterval = time.Minute * 10
	}
	if s.unusableTestInterval == 0 {
		s.unusableTestInterval = time.Minute * 10
	}
	s.filter.Store(config.Filter)
	return s
}

func (s *Selector) Start() error {
	if s.strategy.Speed() {
		s.periodicTestSpeed = task.NewPeriodicTask(s.speedTestInterval, s.TestSpeedAll)
	}
	if s.strategy.Ping() {
		s.periodicTestPing = task.NewPeriodicTask(s.pingTestInterval, s.TestPingAll)
	}
	if s.strategy.Usable() || s.strategy.Speed() || s.strategy.Ping() {
		if (s.periodicTestPing != nil && s.pingTestInterval <= s.unusableTestInterval) ||
			(s.periodicTestSpeed != nil && s.speedTestInterval <= s.unusableTestInterval) {
			// if ping/speed test is enabled and its interval is less than unusable test interval, there
			// is no need to test unusable handlers
		} else {
			s.periodicTestUnusableHandlers = task.NewPeriodicTask(s.unusableTestInterval, s.TestAllUnusable)
		}
	}
	if (s.strategy.Speed() || s.strategy.Usable() || s.strategy.Ping()) && s.dispatcher != nil {
		s.dispatcher.AddHandlerErrorObserver(s)
	}
	if s.periodicTestSpeed != nil {
		s.periodicTestSpeed.Start()
	}
	if s.periodicTestPing != nil {
		s.periodicTestPing.Start()
	}
	if s.periodicTestUnusableHandlers != nil {
		s.periodicTestUnusableHandlers.Start()
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

type handlerInfo interface {
	IsHandlerActive(tag string) bool
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
	handlersToBeTestedForSpeed := make([]OutHandler, 0, len(handlers))
	handlersToBeTestedForIpv6 := make([]OutHandler, 0, len(handlers))
	handlersToBeTestedForPing := make([]OutHandler, 0, len(handlers))
	for _, os := range handlers {
		index := slices.IndexFunc(s.filteredHandlers, func(h OutHandler) bool {
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
		if s.strategy.Speed() && os.GetSpeed() == 0 {
			handlersToBeTestedForSpeed = append(handlersToBeTestedForSpeed, os)
		}
		if s.strategy.Ping() && os.GetPing() == 0 {
			handlersToBeTestedForPing = append(handlersToBeTestedForPing, os)
		}
		if os.GetOk() == 0 {
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
			s.testItems(handlersToBeTestedForSpeed, s.testSpeed)
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

func (s *Selector) pickSpeedTestSize() uint32 {
	if !s.speedTestUseRange {
		return s.speedTestSize
	}
	minB, maxB := s.speedTestSizeMin, s.speedTestSizeMax
	if minB >= maxB {
		return minB
	}
	span := uint64(maxB) - uint64(minB) + 1
	return minB + uint32(rand.Uint64N(span))
}

func (s *Selector) testSpeed(ctx context.Context, t Tester, item OutHandler) {
	TestHandlerSpeed(ctx, t, item, s.pickSpeedTestSize())
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
	log.Debug().Int("len", len(filteredHandlers)).Msg("filtered handlers")

	selectedHandlers := s.strategy.Select(filteredHandlers)
	log.Debug().Int("len", len(selectedHandlers)).Msg("selected handlers")

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
				OutHandler: selectedHandler,
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

	if s.handlerInfo != nil && s.handlerInfo.IsHandlerActive(tag) {
		log.Debug().Str("tag", tag).Msg("handler is active, skip testing")
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
		usable := handler.OutHandler.GetOk() > 0
		if !usable {
			s.setHandlers()
			// since handler unusable might be temporary, test it again after 10 seconds
			go func() {
				select {
				case <-time.After(time.Second * 10):
					TestHandlerUsable(s.ctx, s.tester, handler)
					usable = handler.OutHandler.GetOk() > 0
					log.Debug().Bool("usable", usable).Msg("retry usable test done")
					if usable {
						s.setHandlers()
					}
				case <-s.ctx.Done():
					return
				}
			}()
		}

		s.handlerBeingTestedLock.Lock()
		delete(s.handlerBeingTested, tag)
		s.handlerBeingTestedLock.Unlock()
	}
}

func (s *Selector) testItems(items []OutHandler,
	testFunc func(ctx context.Context, s Tester, item OutHandler)) {
	// Process in batches of 50
	batchSize := 50
	for i := 0; i < len(items); i += batchSize {
		var wg sync.WaitGroup
		end := i + batchSize
		if end > len(items) {
			end = len(items)
		}
		// Process current batch
		for _, item := range items[i:end] {
			wg.Add(1)
			go func(item OutHandler) {
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

func (s *Selector) getOutHandlers() []OutHandler {
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
	s.testItems(s.getOutHandlers(), s.testSpeed)
	s.setHandlers()
}

func (s *Selector) TestPingAll() {
	s.testItems(s.getOutHandlers(), TestHandlerPing)
	s.setHandlers()
}

func (s *Selector) TestAllUnusable() {
	handlers := s.getOutHandlers()
	unusableHandlers := make([]OutHandler, 0, len(handlers))
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
	log.Debug().Str("tag", tag).Int32("speed", speed).Msg("on handler speed changed")
	s.handlersLock.RLock()
	index := slices.IndexFunc(s.filteredHandlers, func(h OutHandler) bool {
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
	if _, ok := s.strategy.(*TopThroughputStrategy); ok {
		s.setHandlers()
	}
}

type selectStrategy interface {
	Select(handlers []OutHandler) []OutHandler
	Usable() bool
	Speed() bool
	Ping() bool
}

type leastPingStrategy struct{}

func (s *leastPingStrategy) Usable() bool {
	return false
}
func (s *leastPingStrategy) Speed() bool {
	return false
}
func (s *leastPingStrategy) Ping() bool {
	return true
}

func (s *leastPingStrategy) Select(handlers []OutHandler) []OutHandler {
	if len(handlers) == 0 {
		return nil
	} else {
		var best OutHandler
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
		return []OutHandler{best}
	}
}

type highestThroughputStrategy struct{}

func (s *highestThroughputStrategy) Usable() bool {
	return false
}
func (s *highestThroughputStrategy) Speed() bool {
	return true
}
func (s *highestThroughputStrategy) Ping() bool {
	return false
}

func (s *highestThroughputStrategy) Select(handlers []OutHandler) []OutHandler {
	if len(handlers) == 0 {
		return nil
	} else {
		var largest OutHandler
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
		return []OutHandler{largest}
	}
}

type allStrategy struct{}

func (s *allStrategy) Select(handlers []OutHandler) []OutHandler {
	return handlers
}

func (s *allStrategy) Usable() bool {
	return false
}
func (s *allStrategy) Speed() bool {
	return false
}
func (s *allStrategy) Ping() bool {
	return false
}

type allOkStrategy struct{}

func (s *allOkStrategy) Usable() bool {
	return true
}
func (s *allOkStrategy) Speed() bool {
	return false
}
func (s *allOkStrategy) Ping() bool {
	return false
}

func (s *allOkStrategy) Select(handlers []OutHandler) []OutHandler {
	okHandlers := make([]OutHandler, 0, len(handlers))
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

// TopPingStrategy selects all nodes within 30% of the least ping (e.g. min 100ms -> select all with ping <= 130ms).
type TopPingStrategy struct{}

func (s *TopPingStrategy) Usable() bool {
	return false
}
func (s *TopPingStrategy) Speed() bool {
	return false
}
func (s *TopPingStrategy) Ping() bool {
	return true
}

func (s *TopPingStrategy) Select(handlers []OutHandler) []OutHandler {
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
		okHandlers := make([]OutHandler, 0, len(handlers))
		for _, h := range handlers {
			if h.GetOk() >= 0 {
				okHandlers = append(okHandlers, h)
			}
		}
		return okHandlers
	}
	threshold := minPing + (minPing * 30 / 100) // 30% above least
	selected := make([]OutHandler, 0, len(handlers))
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

// TopThroughputStrategy selects all nodes with throughput >= 70% of the highest (e.g. max 100 -> select all with speed >= 70).
type TopThroughputStrategy struct{}

func (s *TopThroughputStrategy) Usable() bool {
	return false
}
func (s *TopThroughputStrategy) Speed() bool {
	return true
}
func (s *TopThroughputStrategy) Ping() bool {
	return false
}

func (s *TopThroughputStrategy) Select(handlers []OutHandler) []OutHandler {
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
		okHandlers := make([]OutHandler, 0, len(handlers))
		for _, h := range handlers {
			if h.GetOk() >= 0 {
				okHandlers = append(okHandlers, h)
			}
		}
		return okHandlers
	}
	threshold := maxSpeed * 70 / 100 // 70% of highest
	selected := make([]OutHandler, 0, len(handlers))
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
