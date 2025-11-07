package roaring

import (
	"fmt"
	"math/rand"
	"testing"
)

// TestBitmapProperties runs all invariants against all bitmaps in the corpus
func TestBitmapProperties(t *testing.T) {
	corpus := getBitmapCorpus()
	invariants := getInvariants()

	for _, gen := range corpus {
		for _, inv := range invariants {
			t.Run(fmt.Sprintf("%s/%s", gen.name, inv.name), func(t *testing.T) {
				b := gen.gen()
				inv.test(t, b)
			})
		}
	}
}

// TestBitmapPropertiesWithRunOptimize tests all invariants on RunOptimize'd bitmaps
func TestBitmapPropertiesWithRunOptimize(t *testing.T) {
	corpus := getBitmapCorpus()
	invariants := getInvariants()

	for _, gen := range corpus {
		for _, inv := range invariants {
			t.Run(fmt.Sprintf("%s/%s_optimized", gen.name, inv.name), func(t *testing.T) {
				b := gen.gen()
				b.RunOptimize()
				inv.test(t, b)
			})
		}
	}
}

// bitmapGenerator is a function that creates a test bitmap
type bitmapGenerator struct {
	name string
	gen  func() *Bitmap
}

// invariant is a property that should hold for all bitmaps
type invariant struct {
	name string
	test func(t *testing.T, b *Bitmap)
}

// getInvariants returns all property invariants to test
func getInvariants() []invariant {
	return []invariant{
		{name: "doubleflip", test: doubleFlipInvariant},
		{name: "iteratorbits", test: iteratorBitsInvariant},
		{name: "unsetiteratorbits", test: unsetIteratorBitsInvariant},
	}
}

// doubleFlipInvariant checks that flip(flip(b)) == b
func doubleFlipInvariant(t *testing.T, b *Bitmap) {
	original := b.Clone()

	// Find the range to flip (slightly larger than the bitmap extent)
	var maxVal uint64
	if b.IsEmpty() {
		maxVal = 1000
	} else {
		maxVal = uint64(b.Maximum()) + 1000
	}

	// Flip twice
	b.Flip(0, maxVal)
	b.Flip(0, maxVal)

	// Should be equal to original
	if !original.Equals(b) {
		t.Errorf("double flip should restore original bitmap, original card=%d, result card=%d",
			original.GetCardinality(), b.GetCardinality())
	}
}

// iteratorBitsInvariant checks that creating a bitmap from iterator bits gives the same bitmap
func iteratorBitsInvariant(t *testing.T, b *Bitmap) {
	original := b.Clone()

	// Create new bitmap from iterator
	result := NewBitmap()
	iter := original.Iterator()
	for iter.HasNext() {
		result.Add(iter.Next())
	}

	// Should be equal to original
	if !original.Equals(result) {
		t.Errorf("bitmap reconstructed from iterator should equal original, original card=%d, result card=%d",
			original.GetCardinality(), result.GetCardinality())
	}
}

// unsetIteratorBitsInvariant checks that creating a bitmap from unset iterator, then flipping, gives the same bitmap
func unsetIteratorBitsInvariant(t *testing.T, b *Bitmap) {
	original := b.Clone()

	numUnset := 0x100000000 - b.GetCardinality()
	if numUnset > 1000000 {
		t.Skip("too many iterations")
	}

	// Create bitmap from unset bits
	result := NewBitmap()
	iter := original.UnsetIterator(0, 0x100000000)
	i := 0
	for iter.HasNext() {
		i++
		result.Add(iter.Next())
	}

	// Flip the result in the same range
	result.Flip(0, 0x100000000)

	// Should be equal to original
	if !original.Equals(result) {
		t.Errorf("bitmap reconstructed from unset iterator + flip should equal original, original card=%d, result card=%d",
			original.GetCardinality(), result.GetCardinality())
	}
}

// getBitmapCorpus returns a diverse set of bitmaps for property testing
func getBitmapCorpus() []bitmapGenerator {
	return []bitmapGenerator{
		{
			name: "empty",
			gen: func() *Bitmap {
				return NewBitmap()
			},
		},
		{
			name: "single_bit",
			gen: func() *Bitmap {
				b := NewBitmap()
				b.Add(42)
				return b
			},
		},
		{
			name: "sparse_small",
			gen: func() *Bitmap {
				b := NewBitmap()
				for i := 0; i < 100; i++ {
					b.Add(uint32(i * 1000))
				}
				return b
			},
		},
		{
			name: "sparse_random",
			gen: func() *Bitmap {
				b := NewBitmap()
				r := rand.New(rand.NewSource(12345))
				domain := 100000000
				count := 10000
				for j := 0; j < count; j++ {
					v := uint32(r.Intn(domain))
					b.Add(v)
				}
				return b
			},
		},
		{
			name: "dense_small",
			gen: func() *Bitmap {
				b := NewBitmap()
				for i := 0; i < 10000; i++ {
					b.Add(uint32(i))
				}
				return b
			},
		},
		{
			name: "dense_range",
			gen: func() *Bitmap {
				b := NewBitmap()
				b.AddRange(0, 100000)
				return b
			},
		},
		{
			name: "sequential_ranges",
			gen: func() *Bitmap {
				b := NewBitmap()
				b.AddRange(0, 1000)
				b.AddRange(10000, 11000)
				b.AddRange(100000, 101000)
				return b
			},
		},
		{
			name: "mixed_containers",
			gen: func() *Bitmap {
				b := NewBitmap()
				// Sparse in first container
				for i := 0; i < 100; i++ {
					b.Add(uint32(i * 100))
				}
				// Dense in second container
				for i := 0; i < 60000; i++ {
					b.Add(uint32(65536 + i))
				}
				// Sparse in third container
				for i := 0; i < 50; i++ {
					b.Add(uint32(131072 + i*1000))
				}
				return b
			},
		},
		{
			name: "alternating_bits",
			gen: func() *Bitmap {
				b := NewBitmap()
				for i := 0; i < 100000; i += 2 {
					b.Add(uint32(i))
				}
				return b
			},
		},
		{
			name: "high_values",
			gen: func() *Bitmap {
				b := NewBitmap()
				r := rand.New(rand.NewSource(54321))
				for i := 0; i < 1000; i++ {
					v := uint32(r.Intn(0x70000000) + 0x7fffffff)
					b.Add(v)
				}
				return b
			},
		},
		{
			name: "iterator_benchmark_sparse",
			gen: func() *Bitmap {
				// Based on BenchmarkIteratorAlloc
				b := NewBitmap()
				r := rand.New(rand.NewSource(0))
				sz := 1000000
				initsize := 50000
				for i := 0; i < initsize; i++ {
					b.Add(uint32(r.Intn(sz)))
				}
				return b
			},
		},
		{
			name: "iterator_benchmark_dense",
			gen: func() *Bitmap {
				// Based on BenchmarkNexts
				b := NewBitmap()
				for i := 0; i < 200000; i++ {
					b.Add(uint32(i))
				}
				return b
			},
		},
		{
			name: "iterator_benchmark_rle",
			gen: func() *Bitmap {
				// Based on BenchmarkNextsRLE
				b := NewBitmap()
				b.AddRange(0, 1000000)
				b.RunOptimize()
				return b
			},
		},
		{
			name: "gaps_and_runs",
			gen: func() *Bitmap {
				b := NewBitmap()
				for i := 0; i < 10; i++ {
					start := uint64(i * 100000)
					b.AddRange(start, start+10000)
				}
				return b
			},
		},
	}
}
