// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package outbound

import (
	"sync"
	"time"

	"github.com/5vnetwork/vx-core/common/dice"
	"github.com/5vnetwork/vx-core/common/net"
)

type RandomPortSelector struct {
	sync.RWMutex
	ranges []*net.PortRange
	// disabled []uint16
}

func NewRandomPortSelector(ranges []*net.PortRange) *RandomPortSelector {
	return &RandomPortSelector{
		ranges: ranges,
	}
}

// randomly select one enabled port from the list
func (s *RandomPortSelector) SelectPort() uint16 {
	s.RLock()
	defer s.RUnlock()
	if len(s.ranges) == 0 {
		return 0
	}
	portRangeIndex := dice.Roll(len(s.ranges))
	portRange := s.ranges[portRangeIndex]
	ports := portRange.To - portRange.From
	if ports == 0 {
		return uint16(portRange.From)
	}

	port := uint16(portRange.From) + uint16(dice.Roll(int(ports+1)))
	return port
}

type OnePortSelector struct {
	sync.Mutex
	randomSelector *RandomPortSelector
	interval       time.Duration
	minInterval    time.Duration
	maxInterval    time.Duration
	currentPort    uint16
	nextSwitchAt   time.Time
}

func NewOnePortSelector(ranges []*net.PortRange, interval, minInterval, maxInterval time.Duration) *OnePortSelector {
	if minInterval > maxInterval {
		minInterval, maxInterval = maxInterval, minInterval
	}
	return &OnePortSelector{
		randomSelector: NewRandomPortSelector(ranges),
		interval:       interval,
		minInterval:    minInterval,
		maxInterval:    maxInterval,
	}
}

func (s *OnePortSelector) SelectPort() uint16 {
	s.Lock()
	defer s.Unlock()

	now := time.Now()
	if s.currentPort == 0 {
		s.currentPort = s.randomSelector.SelectPort()
		s.scheduleNextSwitch(now)
		return s.currentPort
	}
	if !s.nextSwitchAt.IsZero() && now.After(s.nextSwitchAt) {
		s.currentPort = s.randomSelector.SelectPort()
		s.scheduleNextSwitch(now)
	}
	return s.currentPort
}

func (s *OnePortSelector) scheduleNextSwitch(now time.Time) {
	d := s.switchDuration()
	if d <= 0 {
		s.nextSwitchAt = time.Time{}
		return
	}
	s.nextSwitchAt = now.Add(d)
}

func (s *OnePortSelector) switchDuration() time.Duration {
	if s.interval > 0 {
		return s.interval
	}
	if s.maxInterval <= 0 {
		return 0
	}
	if s.minInterval <= 0 {
		return s.maxInterval
	}
	if s.minInterval == s.maxInterval {
		return s.minInterval
	}
	span := int((s.maxInterval - s.minInterval) / time.Second)
	return s.minInterval + time.Second*time.Duration(dice.Roll(span+1))
}
