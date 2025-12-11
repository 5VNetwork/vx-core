package monitor

import (
	"sync"
	"sync/atomic"
)

type InboundStats struct {
	lock  sync.RWMutex
	stats map[string]*Stats
}

func NewInboundStats() *InboundStats {
	return &InboundStats{
		stats: make(map[string]*Stats),
	}
}

type Stats struct {
	Traffic atomic.Uint64
}

func NewStats() *Stats {
	return &Stats{
		Traffic: atomic.Uint64{},
	}
}

// if not found, a InboundStats will be created
func (s *InboundStats) Get(tag string) *Stats {
	s.lock.Lock()
	defer s.lock.Unlock()
	st, found := s.stats[tag]
	if !found {
		st = NewStats()
		s.stats[tag] = st
	}
	return st
}

func (s *InboundStats) Remove(tag string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.stats, tag)
}

func (s *InboundStats) All() map[string]*Stats {
	s.lock.RLock()
	defer s.lock.RUnlock()
	stats := make(map[string]*Stats, len(s.stats))
	for k, v := range s.stats {
		stats[k] = v
	}
	return stats
}
