package cache

import (
	"fmt"
	"runtime"
	"testing"
)

// ============================================
// Functional tests for Lru1 (same as Lru)
// ============================================

func TestLru1ReplaceValue(t *testing.T) {
	lru := NewLru1(2)
	lru.Put(2, 6)
	lru.Put(1, 5)
	lru.Put(1, 2)
	v, _ := lru.Get(1)
	if v != 2 {
		t.Error("should get 2", v)
	}
	v, _ = lru.Get(2)
	if v != 6 {
		t.Error("should get 6", v)
	}
}

func TestLru1RemoveOld(t *testing.T) {
	lru := NewLru1(2)
	v, ok := lru.Get(2)
	if ok {
		t.Error("should get nil", v)
	}
	lru.Put(1, 1)
	lru.Put(2, 2)
	v, _ = lru.Get(1)
	if v != 1 {
		t.Error("should get 1", v)
	}
	lru.Put(3, 3)
	v, ok = lru.Get(2)
	if ok {
		t.Error("should get nil", v)
	}
	lru.Put(4, 4)
	v, ok = lru.Get(1)
	if ok {
		t.Error("should get nil", v)
	}
	v, _ = lru.Get(3)
	if v != 3 {
		t.Error("should get 3", v)
	}
	v, _ = lru.Get(4)
	if v != 4 {
		t.Error("should get 4", v)
	}
}

func TestLru1GetKeyFromValue(t *testing.T) {
	lru := NewLru1(2)
	lru.Put(3, 3)
	lru.Put(2, 2)
	lru.Put(1, 1)
	v, ok := lru.GetKeyFromValue(3)
	if ok {
		t.Error("should get nil", v)
	}
	v, _ = lru.GetKeyFromValue(2)
	if v != 2 {
		t.Error("should get 2", v)
	}
}

// Test that Lru1 fixes the valueToElem bug on update
func TestLru1ValueToElemUpdateBug(t *testing.T) {
	lru := NewLru1(2)
	lru.Put("key1", "value1")
	lru.Put("key1", "value2") // Update same key with new value

	// Old value should NOT be found
	_, ok := lru.GetKeyFromValue("value1")
	if ok {
		t.Error("old value should not be in reverse map after update")
	}

	// New value should be found
	k, ok := lru.GetKeyFromValue("value2")
	if !ok || k != "key1" {
		t.Error("new value should be in reverse map", k, ok)
	}
}

// ============================================
// Benchmarks: Put operations
// ============================================

// 28474254                44.39 ns/op           45 B/op          2 allocs/op
func BenchmarkLruPut(b *testing.B) {
	lru := NewLru(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lru.Put(i%1000, i)
	}
}

// 13242840                85.85 ns/op           13 B/op          1 allocs/op
func BenchmarkLru1Put(b *testing.B) {
	lru := NewLru1(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lru.Put(i%1000, i)
	}
}

// ============================================
// Benchmarks: Get operations
// ============================================

// 18.10 ns/op            0 B/op          0 allocs/op
func BenchmarkLruGet(b *testing.B) {
	lru := NewLru(1000)
	for i := 0; i < 1000; i++ {
		lru.Put(i, i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lru.Get(i % 1000)
	}
}

// 15.80 ns/op            0 B/op          0 allocs/op
func BenchmarkLru1Get(b *testing.B) {
	lru := NewLru1(1000)
	for i := 0; i < 1000; i++ {
		lru.Put(i, i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lru.Get(i % 1000)
	}
}

// ============================================
// Benchmarks: GetKeyFromValue operations
// ============================================

// 19.29 ns/op            0 B/op          0 allocs/op
func BenchmarkLruGetKeyFromValue(b *testing.B) {
	lru := NewLru(1000)
	for i := 0; i < 1000; i++ {
		lru.Put(i, i+10000) // Different values
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lru.GetKeyFromValue((i % 1000) + 10000)
	}
}

// 16.63 ns/op            0 B/op          0 allocs/op
func BenchmarkLru1GetKeyFromValue(b *testing.B) {
	lru := NewLru1(1000)
	for i := 0; i < 1000; i++ {
		lru.Put(i, i+10000)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lru.GetKeyFromValue((i % 1000) + 10000)
	}
}

// ============================================
// Benchmarks: Mixed Put/Get operations
// ============================================
// 28.61 ns/op           22 B/op          1 allocs/op
func BenchmarkLruMixed(b *testing.B) {
	lru := NewLru(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if i%2 == 0 {
			lru.Put(i%1000, i)
		} else {
			lru.Get(i % 1000)
		}
	}
}

// 48.90 ns/op            6 B/op          0 allocs/op
func BenchmarkLru1Mixed(b *testing.B) {
	lru := NewLru1(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if i%2 == 0 {
			lru.Put(i%1000, i)
		} else {
			lru.Get(i % 1000)
		}
	}
}

// ============================================
// Memory benchmarks
// ============================================

func TestMemoryComparison(t *testing.T) {
	sizes := []int{100, 500, 1000}

	for _, size := range sizes {
		// Measure Lru (sync.Map) - use TotalAlloc for accurate measurement
		runtime.GC()
		var m1 runtime.MemStats
		runtime.ReadMemStats(&m1)

		lruOld := NewLru(size)
		for i := 0; i < size; i++ {
			lruOld.Put(fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i))
		}

		var m2 runtime.MemStats
		runtime.ReadMemStats(&m2)
		memOld := m2.TotalAlloc - m1.TotalAlloc

		// Keep alive to prevent optimization
		lruOld.Get("key0")

		// Measure Lru1 (regular map)
		runtime.GC()
		var m3 runtime.MemStats
		runtime.ReadMemStats(&m3)

		lruNew := NewLru1(size)
		for i := 0; i < size; i++ {
			lruNew.Put(fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i))
		}

		var m4 runtime.MemStats
		runtime.ReadMemStats(&m4)
		memNew := m4.TotalAlloc - m3.TotalAlloc

		// Keep alive
		lruNew.Get("key0")

		savings := float64(memOld-memNew) / float64(memOld) * 100
		t.Logf("Size %4d: Lru=%6d bytes, Lru1=%6d bytes, Savings=%.1f%%, PerEntry: Lru=%d, Lru1=%d",
			size, memOld, memNew, savings, memOld/uint64(size), memNew/uint64(size))
	}
}

// Allocation benchmarks
// 13530 ns/op           28343 B/op        470 allocs/op
func BenchmarkLruAllocations(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		lru := NewLru(100)
		for j := 0; j < 100; j++ {
			lru.Put(j, j)
		}
		for j := 0; j < 100; j++ {
			lru.Get(j)
		}
	}
}

// 7899 ns/op           15184 B/op        210 allocs/op
func BenchmarkLru1Allocations(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		lru := NewLru1(100)
		for j := 0; j < 100; j++ {
			lru.Put(j, j)
		}
		for j := 0; j < 100; j++ {
			lru.Get(j)
		}
	}
}
