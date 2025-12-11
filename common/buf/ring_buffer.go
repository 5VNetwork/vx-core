package buf

import "sync"

type RingBuffer[T any] struct {
	buffer   []T
	size     int
	writePos int // Position to write next item
	readPos  int // Position to read next item
	count    int // Number of items currently in buffer
	mu       sync.RWMutex
}

func NewRingBuffer[T any](size int) *RingBuffer[T] {
	return &RingBuffer[T]{
		buffer: make([]T, size),
		size:   size,
	}
}

// Add adds a new value to the buffer
func (rb *RingBuffer[T]) Add(value T) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	rb.buffer[rb.writePos] = value
	rb.writePos = (rb.writePos + 1) % rb.size

	// If buffer isn't full yet, increment count
	if rb.count < rb.size {
		rb.count++
	} else {
		// If buffer is full, move read position as we're overwriting oldest value
		rb.readPos = (rb.readPos + 1) % rb.size
	}
}

// ReadOldestValue reads and removes the oldest value from the buffer
// Returns the value and true if successful, zero value and false if buffer is empty
func (rb *RingBuffer[T]) ReadOldestValue() (T, bool) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if rb.count == 0 {
		var zero T
		return zero, false
	}

	value := rb.buffer[rb.readPos]
	var zero T
	rb.buffer[rb.readPos] = zero // Set to zero value
	rb.readPos = (rb.readPos + 1) % rb.size
	rb.count--

	return value, true
}

// Peek returns the oldest value without removing it
func (rb *RingBuffer[T]) Peek() (T, bool) {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	if rb.count == 0 {
		var zero T
		return zero, false
	}

	return rb.buffer[rb.readPos], true
}

// IsFull returns true if the buffer is full
func (rb *RingBuffer[T]) IsFull() bool {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	return rb.count == rb.size
}

// IsEmpty returns true if the buffer is empty
func (rb *RingBuffer[T]) IsEmpty() bool {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	return rb.count == 0
}

// Count returns the current number of items in the buffer
func (rb *RingBuffer[T]) Count() int {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	return rb.count
}

// GetAll returns all values in order from oldest to newest
func (rb *RingBuffer[T]) GetAll() []T {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	if rb.count == 0 {
		return nil
	}

	result := make([]T, rb.count)
	for i := 0; i < rb.count; i++ {
		pos := (rb.readPos + i) % rb.size
		result[i] = rb.buffer[pos]
	}
	return result
}

func (rb *RingBuffer[T]) Clear() {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	rb.buffer = make([]T, rb.size)
	rb.writePos = 0
	rb.readPos = 0
	rb.count = 0
}

func (rb *RingBuffer[T]) Read(slice []T) (int, error) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if rb.count == 0 {
		return 0, nil
	}

	n := 0
	for i := 0; i < len(slice) && rb.count > 0; i++ {
		slice[i] = rb.buffer[rb.readPos]
		rb.readPos = (rb.readPos + 1) % rb.size
		rb.count--
		n++
	}

	return n, nil
}
