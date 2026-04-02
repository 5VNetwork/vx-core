package task

import (
	"context"
	"sync"
	"time"
)

// PeriodicTask is a struct that periodically runs a specified task
type PeriodicTask struct {
	Interval         time.Duration      // How often to run the task
	task             func() error       // The task to run
	ctx              context.Context    // Context to control cancellation
	cancel           context.CancelFunc // Function to cancel the context
	wg               sync.WaitGroup     // WaitGroup to wait for the task to finish
	mu               sync.Mutex         // Mutex to protect the isRunning state
	isRunning        bool               // Flag to track if the task is running
	startImmediately bool
	initialDelay     time.Duration
	ticker           *time.Ticker
}

type PeriodicTaskOption func(*PeriodicTask)

func WithStartImmediately() PeriodicTaskOption {
	return func(pt *PeriodicTask) {
		pt.startImmediately = true
	}
}

func WithInitialDelay(delay time.Duration) PeriodicTaskOption {
	return func(pt *PeriodicTask) {
		pt.initialDelay = delay
	}
}

// NewPeriodicTask creates a new PeriodicTask with the given interval and task
func NewPeriodicTask(interval time.Duration, task func() error, opts ...PeriodicTaskOption) *PeriodicTask {
	ctx, cancel := context.WithCancel(context.Background())
	pt := &PeriodicTask{
		Interval:  interval,
		task:      task,
		ctx:       ctx,
		cancel:    cancel,
		isRunning: false,
	}
	for _, opt := range opts {
		opt(pt)
	}
	return pt
}

func (pt *PeriodicTask) ResetInterval(interval time.Duration) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	pt.Interval = interval
	if pt.ticker != nil {
		pt.ticker.Reset(interval)
	}
}

// Start begins running the task periodically.
// does not block
func (pt *PeriodicTask) Start() error {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	if pt.isRunning {
		return nil // Already running
	}

	pt.isRunning = true
	pt.wg.Add(1)

	go func() {
		defer pt.wg.Done()

		runTask := pt.task

		if pt.startImmediately {
			runTask()
		} else {
			firstDelay := pt.Interval
			if pt.initialDelay > 0 {
				firstDelay = pt.initialDelay
			}

			timer := time.NewTimer(firstDelay)
			defer timer.Stop()

			select {
			case <-timer.C:
				runTask()
			case <-pt.ctx.Done():
				return
			}
		}

		pt.mu.Lock()
		if !pt.isRunning {
			pt.mu.Unlock()
			return
		}
		pt.ticker = time.NewTicker(pt.Interval)
		pt.mu.Unlock()
		defer pt.ticker.Stop()

		for {
			select {
			case <-pt.ticker.C:
				runTask()
			case <-pt.ctx.Done():
				return
			}
		}
	}()

	return nil
}

// Close stops the periodic task and waits for it to finish
func (pt *PeriodicTask) Close() error {
	pt.mu.Lock()
	if !pt.isRunning {
		pt.mu.Unlock()
		return nil // Not running
	}
	pt.isRunning = false
	pt.cancel()
	pt.mu.Unlock()

	// Wait for the goroutine to finish
	pt.wg.Wait()
	return nil
}

// Close stops the periodic task and does not wait for any goroutine to finish
func (pt *PeriodicTask) CloseNoWait() error {
	pt.mu.Lock()
	if !pt.isRunning {
		pt.mu.Unlock()
		return nil // Not running
	}
	pt.isRunning = false
	pt.cancel()
	pt.mu.Unlock()

	return nil
}
