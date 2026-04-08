package task

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestPeriodicTask_Start(t *testing.T) {
	// Create a counter to track how many times our task runs
	var counter int32

	// Create a task that increments the counter
	task := NewPeriodicTask(50*time.Millisecond, func() {
		atomic.AddInt32(&counter, 1)
	})

	// Start the task
	err := task.Start()
	if err != nil {
		t.Fatalf("Failed to start task: %v", err)
	}

	// Sleep to allow the task to run multiple times
	time.Sleep(200 * time.Millisecond)

	// Close the task
	err = task.Close()
	if err != nil {
		t.Fatalf("Failed to close task: %v", err)
	}

	// Check that the task ran at least 3 times (allowing for some timing variability)
	count := atomic.LoadInt32(&counter)
	if count < 3 {
		t.Errorf("Task should have run at least 3 times, but ran %d times", count)
	}
}

func TestPeriodicTask_MultipleStarts(t *testing.T) {
	var counter int32

	task := NewPeriodicTask(50*time.Millisecond, func() {
		atomic.AddInt32(&counter, 1)
	})

	// Start the task
	err := task.Start()
	if err != nil {
		t.Fatalf("Failed to start task: %v", err)
	}

	// Try to start again - should be a no-op
	err = task.Start()
	if err != nil {
		t.Fatalf("Second start should succeed: %v", err)
	}

	// Sleep briefly
	time.Sleep(75 * time.Millisecond)

	// Close the task
	err = task.Close()
	if err != nil {
		t.Fatalf("Failed to close task: %v", err)
	}

	// Record the count after the first run
	count1 := atomic.LoadInt32(&counter)

	// Sleep a bit to ensure no more runs happen
	time.Sleep(100 * time.Millisecond)

	// The counter should not have increased
	count2 := atomic.LoadInt32(&counter)
	if count1 != count2 {
		t.Errorf("Task continued to run after close: count1=%d, count2=%d", count1, count2)
	}
}

func TestPeriodicTask_MultipleCloses(t *testing.T) {
	task := NewPeriodicTask(50*time.Millisecond, func() {
	})

	// Start the task
	err := task.Start()
	if err != nil {
		t.Fatalf("Failed to start task: %v", err)
	}

	// Close the task
	err = task.Close()
	if err != nil {
		t.Fatalf("Failed to close task: %v", err)
	}

	// Try to close again - should be a no-op
	err = task.Close()
	if err != nil {
		t.Fatalf("Second close should succeed: %v", err)
	}
}

func TestPeriodicTask_ImmediateRun(t *testing.T) {
	var counter int32

	task := NewPeriodicTask(1*time.Hour, func() {
		atomic.AddInt32(&counter, 1)
	}, WithStartImmediately())

	// Start the task - it should run immediately despite the long interval
	task.Start()

	// Sleep just a tiny bit to allow the immediate execution
	time.Sleep(10 * time.Millisecond)

	// Close the task
	task.Close()

	// The counter should be 1 from the immediate run
	count := atomic.LoadInt32(&counter)
	if count != 1 {
		t.Errorf("Task should have run exactly once immediately, but ran %d times", count)
	}
}

func TestPeriodicTask_InitialDelay(t *testing.T) {
	var counter int32

	task := NewPeriodicTask(200*time.Millisecond, func() {
		atomic.AddInt32(&counter, 1)
	}, WithInitialDelay(40*time.Millisecond))

	task.Start()
	defer task.Close()

	time.Sleep(20 * time.Millisecond)
	if count := atomic.LoadInt32(&counter); count != 0 {
		t.Fatalf("Task should not have run before initial delay, but ran %d times", count)
	}

	time.Sleep(35 * time.Millisecond)
	if count := atomic.LoadInt32(&counter); count != 1 {
		t.Fatalf("Task should have run once after initial delay, but ran %d times", count)
	}

	time.Sleep(80 * time.Millisecond)
	if count := atomic.LoadInt32(&counter); count != 1 {
		t.Fatalf("Task should not run again before interval elapses, but ran %d times", count)
	}
}
