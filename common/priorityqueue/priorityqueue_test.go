package priorityqueue_test

import (
	"fmt"
	"testing"

	. "github.com/5vnetwork/vx-core/common/priorityqueue"
)

// Tests start here

func TestPriorityQueueEmpty(t *testing.T) {
	pq := NewPriorityQueue[string]()

	if pq.Len() != 0 {
		t.Errorf("Expected empty queue to have length 0, got %d", pq.Len())
	}

	_, ok := pq.Peek()
	if ok {
		t.Error("Expected Peek() on empty queue to return ok=false")
	}

	_, _, ok = pq.Dequeue()
	if ok {
		t.Error("Expected Dequeue() on empty queue to return ok=false")
	}
}

func TestPriorityQueueEnqueueDequeue(t *testing.T) {
	pq := NewPriorityQueue[string]()

	// Add items with different priorities
	pq.Enqueue("Low", 1)
	pq.Enqueue("Medium", 5)
	pq.Enqueue("High", 10)

	if pq.Len() != 3 {
		t.Errorf("Expected queue length to be 3, got %d", pq.Len())
	}

	// Check that items come out in priority order (highest first)
	expectedValues := []string{"High", "Medium", "Low"}
	expectedPriorities := []int{10, 5, 1}

	for i := 0; i < 3; i++ {
		value, priority, ok := pq.Dequeue()
		if !ok {
			t.Errorf("Expected Dequeue() to return ok=true for item %d", i)
			continue
		}

		if value != expectedValues[i] {
			t.Errorf("Expected value %s, got %s", expectedValues[i], value)
		}

		if priority != expectedPriorities[i] {
			t.Errorf("Expected priority %d, got %d", expectedPriorities[i], priority)
		}
	}

	// Queue should be empty now
	if pq.Len() != 0 {
		t.Errorf("Expected queue to be empty, got length %d", pq.Len())
	}
}

func TestPriorityQueuePeek(t *testing.T) {
	pq := NewPriorityQueue[string]()

	pq.Enqueue("Low", 1)
	pq.Enqueue("High", 10)
	pq.Enqueue("Medium", 5)

	// Peek should return the highest priority item without removing it
	item, ok := pq.Peek()
	if !ok {
		t.Error("Expected Peek() to return ok=true")
	}

	if item.Value != "High" {
		t.Errorf("Expected Peek() to return value 'High', got '%s'", item.Value)
	}

	if item.Priority != 10 {
		t.Errorf("Expected Peek() to return priority 10, got %d", item.Priority)
	}

	// Queue length should still be 3
	if pq.Len() != 3 {
		t.Errorf("Expected queue length to still be 3 after Peek(), got %d", pq.Len())
	}

	// Peek again should return the same item
	item2, ok := pq.Peek()
	if !ok {
		t.Error("Expected Peek() to return ok=true")
	}

	if item2.Value != item.Value || item2.Priority != item.Priority {
		t.Errorf("Expected second Peek() to return same item, got value '%s' priority %d", item2.Value, item2.Priority)
	}
}

func TestPriorityQueueUpdate(t *testing.T) {
	pq := NewPriorityQueue[string]()

	// Add some items
	pq.Enqueue("Task 1", 3)
	item := pq.Enqueue("Task 2", 2)
	pq.Enqueue("Task 3", 1)

	// Update the middle item to have highest priority
	pq.Update(item, "Task 2 (urgent)", 10)

	// Now Task 2 should come out first
	value, priority, _ := pq.Dequeue()
	if value != "Task 2 (urgent)" || priority != 10 {
		t.Errorf("Expected updated item to have value 'Task 2 (urgent)' and priority 10, got value '%s' priority %d", value, priority)
	}

	// Update the middle item to have lowest priority
	pq = NewPriorityQueue[string]()
	pq.Enqueue("Task 1", 3)
	item = pq.Enqueue("Task 2", 2)
	pq.Enqueue("Task 3", 1)

	pq.Update(item, "Task 2 (low)", 0)

	// Now Task 2 should come out last
	value1, _, _ := pq.Dequeue()
	value2, _, _ := pq.Dequeue()
	value3, _, _ := pq.Dequeue()

	if value1 != "Task 1" || value2 != "Task 3" || value3 != "Task 2 (low)" {
		t.Errorf("Expected order: Task 1, Task 3, Task 2 (low); got: %s, %s, %s", value1, value2, value3)
	}
}

func TestPriorityQueueWithIntegers(t *testing.T) {
	pq := NewPriorityQueue[int]()

	// Add some integers with priorities
	pq.Enqueue(42, 3)
	pq.Enqueue(100, 5)
	pq.Enqueue(7, 1)

	// Check that items come out in priority order
	value, priority, _ := pq.Dequeue()
	if value != 100 || priority != 5 {
		t.Errorf("Expected value 100 with priority 5, got value %d with priority %d", value, priority)
	}

	value, priority, _ = pq.Dequeue()
	if value != 42 || priority != 3 {
		t.Errorf("Expected value 42 with priority 3, got value %d with priority %d", value, priority)
	}

	value, priority, _ = pq.Dequeue()
	if value != 7 || priority != 1 {
		t.Errorf("Expected value 7 with priority 1, got value %d with priority %d", value, priority)
	}
}

func TestPriorityQueueWithCustomStruct(t *testing.T) {
	type Task struct {
		ID          int
		Description string
	}

	pq := NewPriorityQueue[Task]()

	// Add some tasks with priorities
	pq.Enqueue(Task{1, "Fix bug"}, 3)
	pq.Enqueue(Task{2, "Implement feature"}, 5)
	pq.Enqueue(Task{3, "Write tests"}, 1)

	// Check that items come out in priority order
	task, priority, _ := pq.Dequeue()
	if task.ID != 2 || priority != 5 {
		t.Errorf("Expected task ID 2 with priority 5, got ID %d with priority %d", task.ID, priority)
	}

	task, priority, _ = pq.Dequeue()
	if task.ID != 1 || priority != 3 {
		t.Errorf("Expected task ID 1 with priority 3, got ID %d with priority %d", task.ID, priority)
	}

	task, priority, _ = pq.Dequeue()
	if task.ID != 3 || priority != 1 {
		t.Errorf("Expected task ID 3 with priority 1, got ID %d with priority %d", task.ID, priority)
	}
}

func TestPriorityQueueWithEqualPriorities(t *testing.T) {
	pq := NewPriorityQueue[string]()

	// Add items with the same priority
	pq.Enqueue("First", 5)
	pq.Enqueue("Second", 5)
	pq.Enqueue("Third", 5)

	// Items with equal priorities should maintain stable ordering
	// (though heap doesn't guarantee this, we can at least check they all come out)
	values := make([]string, 0, 3)
	for pq.Len() > 0 {
		value, _, _ := pq.Dequeue()
		values = append(values, value)
	}

	// Check that all items were dequeued
	if len(values) != 3 {
		t.Errorf("Expected 3 values, got %d", len(values))
	}

	// Check that all expected values are present
	expectedValues := []string{"First", "Second", "Third"}
	for _, expected := range expectedValues {
		found := false
		for _, actual := range values {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to find value '%s' in dequeued items, but it was missing", expected)
		}
	}
}

func TestPriorityQueueStress(t *testing.T) {
	pq := NewPriorityQueue[int]()

	// Add a large number of items
	const numItems = 10000
	for i := 0; i < numItems; i++ {
		pq.Enqueue(i, i)
	}

	// Check that they come out in the right order
	for i := numItems - 1; i >= 0; i-- {
		value, priority, _ := pq.Dequeue()
		if value != i || priority != i {
			t.Errorf("Expected value and priority %d, got value %d with priority %d", i, value, priority)
			break // Don't flood the output with errors
		}
	}
}

func TestPriorityQueueEnqueueDequeueMixed(t *testing.T) {
	pq := NewPriorityQueue[string]()

	// Mix enqueue and dequeue operations
	pq.Enqueue("A", 1)
	pq.Enqueue("B", 3)

	value, _, _ := pq.Dequeue() // Should get B
	if value != "B" {
		t.Errorf("Expected 'B', got '%s'", value)
	}

	pq.Enqueue("C", 2)
	pq.Enqueue("D", 4)

	value, _, _ = pq.Dequeue() // Should get D
	if value != "D" {
		t.Errorf("Expected 'D', got '%s'", value)
	}

	value, _, _ = pq.Dequeue() // Should get C
	if value != "C" {
		t.Errorf("Expected 'C', got '%s'", value)
	}

	value, _, _ = pq.Dequeue() // Should get A
	if value != "A" {
		t.Errorf("Expected 'A', got '%s'", value)
	}
}

func TestPriorityQueueClear(t *testing.T) {
	pq := NewPriorityQueue[string]()

	// Add some items
	pq.Enqueue("A", 1)
	pq.Enqueue("B", 2)
	pq.Enqueue("C", 3)

	// Clear the queue by dequeueing all items
	for pq.Len() > 0 {
		pq.Dequeue()
	}

	// Queue should be empty
	if pq.Len() != 0 {
		t.Errorf("Expected empty queue, got length %d", pq.Len())
	}

	// Should be able to add new items
	pq.Enqueue("D", 4)
	if pq.Len() != 1 {
		t.Errorf("Expected queue length 1, got %d", pq.Len())
	}

	value, priority, _ := pq.Dequeue()
	if value != "D" || priority != 4 {
		t.Errorf("Expected value 'D' with priority 4, got value '%s' with priority %d", value, priority)
	}
}

func ExamplePriorityQueue() {
	// Create a priority queue
	pq := NewPriorityQueue[string]()

	// Add some items
	pq.Enqueue("Low priority", 1)
	pq.Enqueue("High priority", 10)
	pq.Enqueue("Medium priority", 5)

	// Process the queue
	for pq.Len() > 0 {
		value, priority, _ := pq.Dequeue()
		fmt.Printf("%s (priority: %d)\n", value, priority)
	}

	// Output:
	// High priority (priority: 10)
	// Medium priority (priority: 5)
	// Low priority (priority: 1)
}

func BenchmarkPriorityQueueEnqueue(b *testing.B) {
	pq := NewPriorityQueue[int]()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pq.Enqueue(i, i)
	}
}

func BenchmarkPriorityQueueDequeue(b *testing.B) {
	pq := NewPriorityQueue[int]()

	// Pre-fill the queue
	for i := 0; i < b.N; i++ {
		pq.Enqueue(i, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if pq.Len() == 0 {
			b.Fatal("Queue is empty")
		}
		pq.Dequeue()
	}
}

func BenchmarkPriorityQueueUpdate(b *testing.B) {
	pq := NewPriorityQueue[int]()

	// Pre-fill the queue and keep track of items
	items := make([]*Item[int], b.N)
	for i := 0; i < b.N; i++ {
		items[i] = pq.Enqueue(i, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pq.Update(items[i], i, b.N-i)
	}
}

func BenchmarkPriorityQueueEnqueueDequeue(b *testing.B) {
	pq := NewPriorityQueue[int]()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pq.Enqueue(i, i)
		if pq.Len() > 1000 { // Keep the queue size bounded
			pq.Dequeue()
		}
	}
}
