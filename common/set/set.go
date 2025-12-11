package set

import (
	"sync"
)

// Set is a generic set implementation with thread safety
type Set[T comparable] struct {
	mu    sync.RWMutex
	items map[T]struct{}
}

// NewSet creates a new set
func NewSet[T comparable]() Set[T] {
	return Set[T]{
		items: make(map[T]struct{}),
	}
}

// Add adds a value to the set
func (s *Set[T]) Add(value T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[value] = struct{}{}
}

// Remove removes a value from the set
func (s *Set[T]) Remove(value T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items, value)
}

// Contains checks if a value is in the set
func (s *Set[T]) Contains(value T) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.items[value]
	return exists
}

// Size returns the number of elements in the set
func (s *Set[T]) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.items)
}

// ToSlice converts the set to a slice
func (s *Set[T]) ToSlice() []T {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]T, 0, len(s.items))
	for value := range s.items {
		result = append(result, value)
	}
	return result
}

// Clear removes all elements from the set
func (s *Set[T]) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for value := range s.items {
		delete(s.items, value)
	}
}
