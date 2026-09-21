// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package filetransfer

import (
	"sync"
	"time"

	"github.com/5vnetwork/vx-core/common/buf"
)

const (
	progressInterval = 100 * time.Millisecond
	progressMinDelta = 256 * 1024
)

// ProgressFunc is called with bytes copied so far and the known total (0 if unknown).
type ProgressFunc func(done, total uint64)

type throttledProgress struct {
	mu              sync.Mutex
	lastReportTime  time.Time
	lastReportBytes uint64
	done            uint64
	total           uint64
	fn              ProgressFunc
}

func newThrottledProgress(total uint64, fn ProgressFunc) *throttledProgress {
	return &throttledProgress{total: total, fn: fn}
}

func (p *throttledProgress) Add(n int64) {
	if n <= 0 || p == nil || p.fn == nil {
		return
	}
	p.mu.Lock()
	p.done += uint64(n)
	done := p.done
	report := time.Since(p.lastReportTime) >= progressInterval ||
		done-p.lastReportBytes >= progressMinDelta ||
		(p.total > 0 && done >= p.total)
	if report {
		p.lastReportTime = time.Now()
		p.lastReportBytes = done
	}
	fn := p.fn
	total := p.total
	p.mu.Unlock()
	if report {
		fn(done, total)
	}
}

func (p *throttledProgress) Force(done uint64) {
	if p == nil || p.fn == nil {
		return
	}
	p.mu.Lock()
	p.done = done
	p.lastReportBytes = done
	p.lastReportTime = time.Now()
	fn := p.fn
	total := p.total
	p.mu.Unlock()
	fn(done, total)
}

type progressDataHandler struct {
	p *throttledProgress
}

func (h progressDataHandler) HandleData(mb buf.MultiBuffer) {
	h.p.Add(int64(mb.Len()))
}
