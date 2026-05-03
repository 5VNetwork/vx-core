package linkstats

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/5vnetwork/vx-core/common"
	"github.com/rs/zerolog/log"
)

type linkStatsAdder interface {
	AddThroughput(uint64)
	AddPing(uint64)
}
type LinkStats struct {
	ctx                    context.Context
	ohStats                linkStatsAdder
	initialWriteTime       time.Time
	prevWriteTime          time.Time
	writeCounter           atomic.Uint64
	hasDoneCalculatingRate bool
	initialReadTime        time.Time
	hadAddedPing           bool
}

func NewLinkStats(ctx context.Context, ohStats linkStatsAdder) *LinkStats {
	return &LinkStats{
		ctx:     ctx,
		ohStats: ohStats,
	}
}

func (w *LinkStats) UpTraffic(n uint64) {
	if w.initialReadTime.IsZero() {
		w.initialReadTime = time.Now()
	}
}

func (w *LinkStats) DownTraffic(n uint64) {
	if !w.hadAddedPing {
		w.ohStats.AddPing(uint64(time.Since(w.initialReadTime).Milliseconds()))
		w.hadAddedPing = true
	}
	if !w.hasDoneCalculatingRate {
		if w.initialWriteTime.IsZero() {
			w.initialWriteTime = time.Now()
		}
		if !w.prevWriteTime.IsZero() && time.Since(w.prevWriteTime).Seconds() > 1 {
			w.hasDoneCalculatingRate = true
		} else {
			w.prevWriteTime = time.Now()
			w.writeCounter.Add(n)
			if w.writeCounter.Load() >= common.OneKB*10 {
				elapsed := time.Since(w.initialWriteTime).Seconds()
				rate := float64(w.writeCounter.Swap(0)) / elapsed
				if rate > 1024*1024*100 {
					log.Ctx(w.ctx).Warn().Float64("elapsed", elapsed).Uint64("rate", uint64(rate)).Msg("throughput is too high")
					w.hasDoneCalculatingRate = true
				} else {
					log.Ctx(w.ctx).Debug().Float64("rate(MBps)", rate/1000/1000).Msg("throughput")
					w.ohStats.AddThroughput(uint64(rate))
				}
				w.initialWriteTime = time.Now()
			}
		}
	}
}
