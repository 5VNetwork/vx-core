package outbound

import (
	"sync"
	"sync/atomic"
	"time"
)

type OutStats struct {
	sync.Mutex
	Map map[string]*OutboundHandlerStats
}

func NewOutStats() *OutStats {
	return &OutStats{
		Map: make(map[string]*OutboundHandlerStats),
	}
}

func (o *OutStats) Get(tag string) *OutboundHandlerStats {
	o.Lock()
	defer o.Unlock()

	stats, ok := o.Map[tag]
	if !ok {
		stats = NewHandlerStats(0, 0)
		o.Map[tag] = stats
	}
	stats.Time.Store(time.Now())
	return stats
}

func (o *OutStats) CleanOldStats() {
	o.Lock()
	defer o.Unlock()
	for tag, stats := range o.Map {
		if time.Since(stats.Time.Load().(time.Time)) > 60*time.Second {
			delete(o.Map, tag)
		}
	}
}

type OutboundHandlerStats struct {
	UpCounter   atomic.Uint64
	DownCounter atomic.Uint64
	Throughput  atomic.Uint64
	Ping        atomic.Uint64
	Time        atomic.Value
}

func NewHandlerStats(throughput uint64, ping uint64) *OutboundHandlerStats {
	s := &OutboundHandlerStats{}
	s.Throughput.Store(throughput)
	s.Ping.Store(ping)
	s.Time.Store(time.Now())
	return s
}

func (s *OutboundHandlerStats) AddThroughput(v uint64) {
	if s.Throughput.Load() == 0 {
		s.Throughput.Store(v)
	} else {
		s.Throughput.Swap(uint64(float64(s.Throughput.Load())*0.875 + 0.125*float64(v)))
	}
}

func (s *OutboundHandlerStats) AddPing(v uint64) {
	if s.Ping.Load() == 0 {
		s.Ping.Store(v)
	} else {
		s.Ping.Swap(uint64(float64(s.Ping.Load())*0.875 + 0.125*float64(v)))
	}
}
