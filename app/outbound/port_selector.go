package outbound

import (
	"sync"

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
	ports := portRange.ToPort() - portRange.FromPort()
	if ports == 0 {
		return uint16(portRange.FromPort())
	}

	port := uint16(portRange.FromPort()) + uint16(dice.Roll(int(ports+1)))
	return port
}

// func (s *RandomPortSelector) DisablePort(port uint16, duration time.Duration) {
// 	s.Lock()
// 	defer s.Unlock()
// 	s.enabled = slices.DeleteFunc(s.enabled, func(p uint16) bool {
// 		return p == port
// 	})
// 	s.disabled = append(s.disabled, port)
// 	time.AfterFunc(duration, func() {
// 		s.Lock()
// 		defer s.Unlock()
// 		s.disabled = slices.DeleteFunc(s.disabled, func(p uint16) bool {
// 			return p == port
// 		})
// 		s.enabled = append(s.enabled, port)
// 	})
// }

// type TimeoutPortSelector struct {
// 	timeout               time.Duration
// 	randomPortSelector    *RandomPortSelector
// 	previousPortStartTime time.Time
// 	currentPort           uint16
// }

// func NewTimeoutPortSelector(ranges []*net.PortRange, timeout time.Duration) *TimeoutPortSelector {
// 	r := NewRandomPortSelector(ranges)
// 	return &TimeoutPortSelector{
// 		timeout:               timeout,
// 		randomPortSelector:    r,
// 		previousPortStartTime: time.Now(),
// 		currentPort:           r.SelectPort(),
// 	}
// }

// // return the current port, and update the port if timeout
// func (s *TimeoutPortSelector) SelectPort() uint16 {
// 	now := time.Now()
// 	if now.Sub(s.previousPortStartTime) > s.timeout {
// 		s.previousPortStartTime = now
// 		s.currentPort = s.randomPortSelector.SelectPort()
// 	}
// 	return s.currentPort
// }

// func (s *TimeoutPortSelector) DisablePort(port uint16, duration time.Duration) {
// 	s.randomPortSelector.DisablePort(port, duration)
// 	if s.currentPort == port {
// 		s.previousPortStartTime = time.Now()
// 		s.currentPort = s.randomPortSelector.SelectPort()
// 	}
// }
