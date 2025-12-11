// Package mph implements a minimal perfect hash table over strings.
package strmatcher

// import (
// 	"math/bits"

// 	"github.com/fxamacker/circlehash"
// 	"golang.org/x/exp/slices"
// )

// // A Table is an immutable hash table that provides constant-time lookups of key
// // indices using a minimal perfect hash.
// type Table struct {
// 	keys       []string
// 	level0     []uint32 // power of 2 size
// 	level0Mask int      // len(Level0) - 1
// 	level1     []uint32 // power of 2 size >= len(keys)
// 	level1Mask int      // len(Level1) - 1
// }

// func Build(keys []string) *Table {
// 	var (
// 		level0        = make([]uint32, nextPow21(len(keys)/2)) // Larger level0
// 		level0Mask    = len(level0) - 1
// 		level1        = make([]uint32, nextPow21(2*len(keys))) // Larger level1
// 		level1Mask    = len(level1) - 1
// 		sparseBuckets = make([][]int, len(level0))
// 	)
// 	for i, s := range keys {
// 		n := int(circlehash.Hash64String(s, 0)) & level0Mask
// 		sparseBuckets[n] = append(sparseBuckets[n], i)
// 	}
// 	var buckets []indexBucket
// 	for n, vals := range sparseBuckets {
// 		if len(vals) > 0 {
// 			buckets = append(buckets, indexBucket{n, vals})
// 		}
// 	}
// 	slices.SortFunc(buckets, func(a, b indexBucket) int {
// 		return len(b.vals) - len(a.vals)
// 	})

// 	occ := make([]bool, len(level1))
// 	var tmpOcc []int
// 	for _, bucket := range buckets {
// 		var seed uint32
// 		maxAttempts := 1000 // Limit seed attempts
// 		attempts := 0
// 	trySeed:
// 		if attempts >= maxAttempts {
// 			// Fallback: Increase level1 size or split bucket (simplified here)
// 			panic("failed to find seed; consider increasing level1 size")
// 		}
// 		tmpOcc = tmpOcc[:0]
// 		for _, i := range bucket.vals {
// 			n := int(circlehash.Hash64String(keys[i], uint64(seed))) & level1Mask
// 			if occ[n] {
// 				for _, n := range tmpOcc {
// 					occ[n] = false
// 				}
// 				seed++
// 				attempts++
// 				goto trySeed
// 			}
// 			occ[n] = true
// 			tmpOcc = append(tmpOcc, n)
// 			level1[n] = uint32(i)
// 		}
// 		level0[int(bucket.n)] = uint32(seed)
// 	}

// 	return &Table{
// 		keys:       keys,
// 		level0:     level0,
// 		level0Mask: level0Mask,
// 		level1:     level1,
// 		level1Mask: level1Mask,
// 	}
// }

// func nextPow21(n int) int {
// 	switch n {
// 	case 0:
// 		return 1
// 	case 1:
// 		return 2
// 	default:
// 		return (1 << bits.Len(uint(n-1)))
// 	}
// }

// // Lookup searches for s in t and returns its index and whether it was found.
// func (t *Table) Lookup(s string) (n uint32, ok bool) {
// 	i0 := int(circlehash.Hash64String(s, 0)) & t.level0Mask
// 	seed := t.level0[i0]
// 	i1 := int(circlehash.Hash64String(s, uint64(seed))) & t.level1Mask
// 	n = t.level1[i1]
// 	return n, s == t.keys[int(n)]
// }

// type indexBucket struct {
// 	n    int
// 	vals []int
// }
