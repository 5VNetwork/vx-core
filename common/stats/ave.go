package stats

import "sync"

// CircularBuffer is a simple circular buffer to hold recent values.
type CircularBuffer struct {
	sync.RWMutex
	values []int
	index  int
	size   int
	full   bool
	sum    int
}

// NewCircularBuffer creates a new CircularBuffer of the given size.
func NewCircularBuffer(size int) *CircularBuffer {
	return &CircularBuffer{
		values: make([]int, size),
		size:   size,
	}
}

// Add adds a new value to the circular buffer and updates the average.
func (cb *CircularBuffer) Add(value int) {
	cb.Lock()
	defer cb.Unlock()
	if cb.full {
		cb.sum -= cb.values[cb.index] // Remove the oldest value from sum
	}
	cb.values[cb.index] = value
	cb.sum += value

	cb.index = (cb.index + 1) % cb.size
	if cb.index == 0 {
		cb.full = true
	}
}

// Average returns the average of the values in the circular buffer.
func (cb *CircularBuffer) Average() int {
	cb.RLock()
	defer cb.RUnlock()
	if cb.full {
		return cb.sum / int(cb.size)
	}
	return cb.sum / int(cb.index)
}
