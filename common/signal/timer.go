package signal

import (
	"sync"
	"time"

	"github.com/5vnetwork/vx-core/common/task"
)

// periodically check if there is any activity in the past t(a time) by receiving from the channel, if there is no,
// run the onInactive(); if there is, go on
// some goroutines will ReportActivity using the channel
type ActivityChecker struct {
	sync.RWMutex
	activityReportingChan chan struct{}
	periodicTask          *task.Periodic
	onInactive            func()
}

func (t *ActivityChecker) Update() {
	select {
	case t.activityReportingChan <- struct{}{}:
	default:
	}
}

func (t *ActivityChecker) check() error {
	select {
	case <-t.activityReportingChan:
	default:
		t.Finish()
	}
	return nil
}

func (t *ActivityChecker) Finish() {
	t.Lock()
	defer t.Unlock()

	if t.onInactive != nil {
		t.onInactive()
		t.onInactive = nil
	}
	if t.periodicTask != nil {
		t.periodicTask.Close()
		t.periodicTask = nil
	}
}

// does not call Finish()
func (t *ActivityChecker) Cancel() {
	t.Lock()
	defer t.Unlock()
	if t.onInactive != nil {
		t.onInactive = nil
	}
	if t.periodicTask != nil {
		t.periodicTask.Close()
		t.periodicTask = nil
	}
}

func (t *ActivityChecker) SetTimeout(timeout time.Duration) {
	if timeout == 0 {
		t.Finish()
		return
	}

	periodicTask := &task.Periodic{
		Interval: timeout,
		Execute:  t.check,
		// Done:     make(chan struct{}),
	}

	t.Lock()
	if t.periodicTask != nil {
		t.periodicTask.Close()
	}
	t.periodicTask = periodicTask
	t.Unlock()

	t.Update()
	periodicTask.Start()
}

func NewActivityChecker(onInactive func(), timeout time.Duration) *ActivityChecker {
	checker := &ActivityChecker{
		activityReportingChan: make(chan struct{}, 1),
		onInactive:            onInactive,
	}
	checker.SetTimeout(timeout)
	return checker
}
