package task

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

func TestPeriodicTask_Start(t *testing.T) {
	// Create a counter to track how many times our task runs
	var counter int32

	// Create a task that increments the counter
	task := NewPeriodicTask(50*time.Millisecond, func() error {
		atomic.AddInt32(&counter, 1)
		return nil
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

	task := NewPeriodicTask(50*time.Millisecond, func() error {
		atomic.AddInt32(&counter, 1)
		return nil
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
	task := NewPeriodicTask(50*time.Millisecond, func() error {
		return nil
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

func TestPeriodicTask_TaskError(t *testing.T) {
	var counter int32
	var errorCount int32

	// Create a test error
	testError := fmt.Errorf("test error")

	task := NewPeriodicTask(50*time.Millisecond, func() error {
		current := atomic.AddInt32(&counter, 1)
		// Return an error on the second run
		if current == 2 {
			atomic.AddInt32(&errorCount, 1)
			return testError
		}
		return nil
	})

	// Start the task
	task.Start()

	// Sleep to allow multiple runs
	time.Sleep(150 * time.Millisecond)

	// Close the task
	task.Close()

	// The task should have continued running even after an error
	if counter < 2 {
		t.Errorf("Task should have run at least 2 times, but ran %d times", counter)
	}

	// Should have encountered exactly one error
	if errorCount != 1 {
		t.Errorf("Expected 1 error, but got %d", errorCount)
	}
}

func TestPeriodicTask_ImmediateRun(t *testing.T) {
	var counter int32

	task := NewPeriodicTask(1*time.Hour, func() error {
		atomic.AddInt32(&counter, 1)
		return nil
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
