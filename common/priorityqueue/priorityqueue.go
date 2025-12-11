package priorityqueue

import (
	"container/heap"
	"fmt"
	"sync"
)

// Item represents an item in the priority queue
type Item[T any] struct {
	Value    T   // The value of the item
	Priority int // The priority of the item
	Index    int // The index of the item in the heap
}

// PriorityQueue implements heap.Interface and holds Items
type PriorityQueue[T any] []*Item[T]

func (pq PriorityQueue[T]) Len() int { return len(pq) }

func (pq PriorityQueue[T]) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here
	return pq[i].Priority > pq[j].Priority
}

func (pq PriorityQueue[T]) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].Index = i
	pq[j].Index = j
}

func (pq *PriorityQueue[T]) Push(x interface{}) {
	n := len(*pq)
	item := x.(*Item[T])
	item.Index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue[T]) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.Index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

// Update modifies the priority and value of an Item in the queue
func (pq *PriorityQueue[T]) Update(item *Item[T], value T, priority int) error {
	// Validate that the item is in the queue
	if item.Index < 0 || item.Index >= pq.Len() || (*pq)[item.Index] != item {
		return fmt.Errorf("item not in queue")
	}
	item.Value = value
	item.Priority = priority
	heap.Fix(pq, item.Index)
	return nil
}

// NewPriorityQueue creates a new priority queue
func NewPriorityQueue[T any]() *PriorityQueue[T] {
	pq := make(PriorityQueue[T], 0)
	heap.Init(&pq)
	return &pq
}

// Enqueue adds an item to the priority queue
func (pq *PriorityQueue[T]) Enqueue(value T, priority int) *Item[T] {
	item := &Item[T]{
		Value:    value,
		Priority: priority,
	}
	heap.Push(pq, item)
	return item
}

// Dequeue removes and returns the highest priority item
func (pq *PriorityQueue[T]) Dequeue() (T, int, bool) {
	if pq.Len() == 0 {
		var zero T
		return zero, 0, false
	}
	item := heap.Pop(pq).(*Item[T])
	return item.Value, item.Priority, true
}

// Peek returns the highest priority item without removing it
func (pq *PriorityQueue[T]) Peek() (*Item[T], bool) {
	if pq.Len() == 0 {
		return nil, false
	}
	return (*pq)[0], true
}

// ThreadSafePriorityQueue wraps a PriorityQueue with a mutex for thread safety
type ThreadSafePriorityQueue[T any] struct {
	pq  PriorityQueue[T]
	mux sync.RWMutex
}

// NewThreadSafePriorityQueue creates a new thread-safe priority queue
func NewThreadSafePriorityQueue[T any]() *ThreadSafePriorityQueue[T] {
	return &ThreadSafePriorityQueue[T]{
		pq: make(PriorityQueue[T], 0),
	}
}

// Read-only operations use RLock/RUnlock
func (tspq *ThreadSafePriorityQueue[T]) Len() int {
	tspq.mux.RLock()
	defer tspq.mux.RUnlock()
	return tspq.pq.Len()
}

// Enqueue adds an item to the priority queue
func (tspq *ThreadSafePriorityQueue[T]) Enqueue(value T, priority int) *Item[T] {
	tspq.mux.Lock()
	defer tspq.mux.Unlock()

	item := &Item[T]{
		Value:    value,
		Priority: priority,
	}
	heap.Push(&tspq.pq, item)
	return item
}

// Dequeue removes and returns the highest priority item
func (tspq *ThreadSafePriorityQueue[T]) Dequeue() (T, int, bool) {
	tspq.mux.Lock()
	defer tspq.mux.Unlock()

	if tspq.pq.Len() == 0 {
		var zero T
		return zero, 0, false
	}

	item := heap.Pop(&tspq.pq).(*Item[T])
	return item.Value, item.Priority, true
}

// Peek returns the highest priority item without removing it
func (tspq *ThreadSafePriorityQueue[T]) Peek() (*Item[T], bool) {
	tspq.mux.RLock()
	defer tspq.mux.RUnlock()

	if tspq.pq.Len() == 0 {
		return nil, false
	}

	return tspq.pq[0], true
}

// Update modifies the priority and value of an Item in the queue
func (tspq *ThreadSafePriorityQueue[T]) Update(item *Item[T], value T, priority int) error {
	tspq.mux.Lock()
	defer tspq.mux.Unlock()

	// Validate that the item is in the queue
	if item.Index < 0 || item.Index >= tspq.pq.Len() || tspq.pq[item.Index] != item {
		return fmt.Errorf("item not in queue")
	}

	item.Value = value
	item.Priority = priority
	heap.Fix(&tspq.pq, item.Index)
	return nil
}

// returns all items in the queue
func (tspq *ThreadSafePriorityQueue[T]) Items() []*Item[T] {
	tspq.mux.RLock()
	defer tspq.mux.RUnlock()
	return tspq.pq
}
