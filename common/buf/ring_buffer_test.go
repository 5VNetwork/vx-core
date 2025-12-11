package buf

import (
	"reflect"
	"sync"
	"testing"
)

func TestRingBuffer_Add(t *testing.T) {
	tests := []struct {
		name     string
		size     int
		inputs   []int
		expected []int
	}{
		{
			name:     "add to empty buffer",
			size:     3,
			inputs:   []int{1},
			expected: []int{1},
		},
		{
			name:     "add until full",
			size:     3,
			inputs:   []int{1, 2, 3},
			expected: []int{1, 2, 3},
		},
		{
			name:     "add with wraparound",
			size:     3,
			inputs:   []int{1, 2, 3, 4, 5},
			expected: []int{3, 4, 5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rb := NewRingBuffer[int](tt.size)

			for _, v := range tt.inputs {
				rb.Add(v)
			}

			result := rb.GetAll()
			if len(result) != len(tt.expected) {
				t.Errorf("got length %v, want %v", len(result), len(tt.expected))
			}

			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("at index %d got %v, want %v", i, v, tt.expected[i])
				}
			}
		})
	}
}

func TestRingBuffer_ReadOldestValue(t *testing.T) {
	tests := []struct {
		name          string
		size          int
		inputs        []int
		expectedValue int
		expectedOk    bool
	}{
		{
			name:          "read from empty buffer",
			size:          3,
			inputs:        []int{},
			expectedValue: 0,
			expectedOk:    false,
		},
		{
			name:          "read single value",
			size:          3,
			inputs:        []int{1},
			expectedValue: 1,
			expectedOk:    true,
		},
		{
			name:          "read oldest after wraparound",
			size:          3,
			inputs:        []int{1, 2, 3, 4},
			expectedValue: 2,
			expectedOk:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rb := NewRingBuffer[int](tt.size)

			for _, v := range tt.inputs {
				rb.Add(v)
			}

			value, ok := rb.ReadOldestValue()
			if ok != tt.expectedOk {
				t.Errorf("got ok=%v, want %v", ok, tt.expectedOk)
			}
			if ok && value != tt.expectedValue {
				t.Errorf("got value=%v, want %v", value, tt.expectedValue)
			}
		})
	}
}

func TestRingBuffer_Peek(t *testing.T) {
	rb := NewRingBuffer[int](3)

	// Peek empty buffer
	value, ok := rb.Peek()
	if ok {
		t.Errorf("expected false for empty buffer, got true with value %v", value)
	}

	// Add value and peek
	rb.Add(1)
	value, ok = rb.Peek()
	if !ok || value != 1 {
		t.Errorf("expected (1, true), got (%v, %v)", value, ok)
	}

	// Verify peek doesn't remove value
	value, ok = rb.Peek()
	if !ok || value != 1 {
		t.Errorf("expected (1, true) on second peek, got (%v, %v)", value, ok)
	}
}

func TestRingBuffer_IsFull(t *testing.T) {
	rb := NewRingBuffer[int](2)

	if rb.IsFull() {
		t.Error("new buffer should not be full")
	}

	rb.Add(1)
	if rb.IsFull() {
		t.Error("buffer with one item should not be full")
	}

	rb.Add(2)
	if !rb.IsFull() {
		t.Error("buffer should be full")
	}
}

func TestRingBuffer_IsEmpty(t *testing.T) {
	rb := NewRingBuffer[int](2)

	if !rb.IsEmpty() {
		t.Error("new buffer should be empty")
	}

	rb.Add(1)
	if rb.IsEmpty() {
		t.Error("buffer should not be empty after adding value")
	}

	rb.ReadOldestValue()
	if !rb.IsEmpty() {
		t.Error("buffer should be empty after reading only value")
	}
}

func TestRingBuffer_Count(t *testing.T) {
	rb := NewRingBuffer[int](3)

	if rb.Count() != 0 {
		t.Errorf("new buffer count = %d, want 0", rb.Count())
	}

	rb.Add(1)
	if rb.Count() != 1 {
		t.Errorf("buffer count = %d, want 1", rb.Count())
	}

	rb.Add(2)
	rb.Add(3)
	rb.Add(4) // Overwrites 1
	if rb.Count() != 3 {
		t.Errorf("buffer count = %d, want 3", rb.Count())
	}

	rb.ReadOldestValue()
	if rb.Count() != 2 {
		t.Errorf("buffer count = %d, want 2", rb.Count())
	}
}

func TestRingBuffer_Concurrent(t *testing.T) {
	rb := NewRingBuffer[int](100)
	var wg sync.WaitGroup
	writers := 10
	readsPerWriter := 100

	// Start multiple writers
	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func(writerID int) {
			defer wg.Done()
			for j := 0; j < readsPerWriter; j++ {
				rb.Add(writerID*readsPerWriter + j)
			}
		}(i)
	}

	// Start multiple readers
	readErrors := make(chan error, writers)
	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < readsPerWriter; j++ {
				rb.ReadOldestValue()
			}
		}()
	}

	wg.Wait()
	close(readErrors)

	for err := range readErrors {
		if err != nil {
			t.Errorf("concurrent operation error: %v", err)
		}
	}
}

func TestRingBuffer_GetAll(t *testing.T) {
	tests := []struct {
		name     string
		size     int
		inputs   []string
		expected []string
	}{
		{
			name:     "empty buffer",
			size:     3,
			inputs:   []string{},
			expected: nil,
		},
		{
			name:     "partially filled buffer",
			size:     3,
			inputs:   []string{"a", "b"},
			expected: []string{"a", "b"},
		},
		{
			name:     "full buffer",
			size:     3,
			inputs:   []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "wrapped buffer",
			size:     3,
			inputs:   []string{"a", "b", "c", "d", "e"},
			expected: []string{"c", "d", "e"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rb := NewRingBuffer[string](tt.size)

			for _, v := range tt.inputs {
				rb.Add(v)
			}

			result := rb.GetAll()
			if len(result) != len(tt.expected) {
				t.Errorf("got length %v, want %v", len(result), len(tt.expected))
			}

			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("at index %d got %v, want %v", i, v, tt.expected[i])
				}
			}
		})
	}
}
func TestRingBuffer_Read(t *testing.T) {
	t.Run("read from empty buffer", func(t *testing.T) {
		rb := NewRingBuffer[int](5)
		slice := make([]int, 3)

		n, err := rb.Read(slice)

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if n != 0 {
			t.Errorf("expected to read 0 items, got %d", n)
		}
	})

	t.Run("read partial buffer", func(t *testing.T) {
		rb := NewRingBuffer[int](5)
		rb.Add(1)
		rb.Add(2)
		rb.Add(3)

		slice := make([]int, 2)
		n, err := rb.Read(slice)

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if n != 2 {
			t.Errorf("expected to read 2 items, got %d", n)
		}
		if !reflect.DeepEqual(slice, []int{1, 2}) {
			t.Errorf("expected slice to be [1, 2], got %v", slice)
		}
		if rb.Count() != 1 {
			t.Errorf("expected buffer to have 1 item remaining, got %d", rb.Count())
		}
	})

	t.Run("read entire buffer", func(t *testing.T) {
		rb := NewRingBuffer[int](5)
		rb.Add(1)
		rb.Add(2)
		rb.Add(3)

		slice := make([]int, 3)
		n, err := rb.Read(slice)

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if n != 3 {
			t.Errorf("expected to read 3 items, got %d", n)
		}
		if !reflect.DeepEqual(slice, []int{1, 2, 3}) {
			t.Errorf("expected slice to be [1, 2, 3], got %v", slice)
		}
		if !rb.IsEmpty() {
			t.Error("expected buffer to be empty after reading all items")
		}
	})

	t.Run("read with wrapped buffer", func(t *testing.T) {
		rb := NewRingBuffer[int](3)
		// Fill buffer and wrap around
		rb.Add(1)
		rb.Add(2)
		rb.Add(3)
		rb.Add(4) // This will overwrite 1
		rb.Add(5) // This will overwrite 2

		slice := make([]int, 3)
		n, err := rb.Read(slice)

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if n != 3 {
			t.Errorf("expected to read 3 items, got %d", n)
		}
		if !reflect.DeepEqual(slice, []int{3, 4, 5}) {
			t.Errorf("expected slice to be [3, 4, 5], got %v", slice)
		}
		if !rb.IsEmpty() {
			t.Error("expected buffer to be empty after reading all items")
		}
	})

	t.Run("read with larger slice than buffer", func(t *testing.T) {
		rb := NewRingBuffer[int](3)
		rb.Add(1)
		rb.Add(2)

		slice := make([]int, 5)
		n, err := rb.Read(slice)

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if n != 2 {
			t.Errorf("expected to read 2 items, got %d", n)
		}
		if !reflect.DeepEqual(slice[:n], []int{1, 2}) {
			t.Errorf("expected slice[:n] to be [1, 2], got %v", slice[:n])
		}
		if !rb.IsEmpty() {
			t.Error("expected buffer to be empty after reading all items")
		}
	})
}
