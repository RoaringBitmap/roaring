package roaring64

import (
	"math/rand"
	"testing"
)

// BENCHMARKS, to run them type "go test -bench Benchmark -run -"

// go test -bench BenchmarkIteratorAlloc -benchmem -run -
func BenchmarkIteratorAlloc(b *testing.B) {
	bm := New()
	domain := uint64(100000000)
	count := 10000
	for j := 0; j < count; j++ {
		v := uint64(rand.Int63n(int64(domain)))
		bm.Add(v)
	}
	i := IntIterator64{}
	expectedCardinality := bm.GetCardinality()
	counter := uint64(0)
	b.Run("simple iteration with alloc", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			counter = 0
			i := bm.Iterator()
			for i.HasNext() {
				i.Next()
				counter++
			}
		}
		b.StopTimer()
	})
	if counter != expectedCardinality {
		b.Fatalf("Cardinalities don't match: %d, %d", counter, expectedCardinality)
	}
	b.Run("simple iteration", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			counter = 0
			i.Initialize(bm)
			for i.HasNext() {
				i.Next()
				counter++
			}
		}
		b.StopTimer()
	})
	if counter != expectedCardinality {
		b.Fatalf("Cardinalities don't match: %d, %d", counter, expectedCardinality)
	}
	b.Run("values iteration", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			counter = 0
			Values(bm)(func(_ uint64) bool {
				counter++
				return true
			})
		}
		b.StopTimer()
	})
	if counter != expectedCardinality {
		b.Fatalf("Cardinalities don't match: %d, %d", counter, expectedCardinality)
	}
	b.Run("reverse iteration with alloc", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			counter = 0
			ir := bm.ReverseIterator()
			for ir.HasNext() {
				ir.Next()
				counter++
			}
		}
		b.StopTimer()
	})
	if counter != expectedCardinality {
		b.Fatalf("Cardinalities don't match: %d, %d", counter, expectedCardinality)
	}
	ir := IntReverseIterator64{}

	b.Run("reverse iteration", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			counter = 0
			ir.Initialize(bm)
			for ir.HasNext() {
				ir.Next()
				counter++
			}
		}
		b.StopTimer()
	})
	if counter != expectedCardinality {
		b.Fatalf("Cardinalities don't match: %d, %d", counter, expectedCardinality)
	}
	b.Run("backward iteration", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			counter = 0
			Backward(bm)(func(_ uint64) bool {
				counter++
				return true
			})
		}
		b.StopTimer()
	})
	if counter != expectedCardinality {
		b.Fatalf("Cardinalities don't match: %d, %d", counter, expectedCardinality)
	}

	b.Run("many iteration with alloc", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			counter = 0
			buf := make([]uint64, 1024)
			im := bm.ManyIterator()
			for n := im.NextMany(buf); n != 0; n = im.NextMany(buf) {
				counter += uint64(n)
			}
		}
		b.StopTimer()
	})
	if counter != expectedCardinality {
		b.Fatalf("Cardinalities don't match: %d, %d", counter, expectedCardinality)
	}
	im := ManyIntIterator64{}
	buf := make([]uint64, 1024)

	b.Run("many iteration", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			counter = 0
			im.Initialize(bm)
			for n := im.NextMany(buf); n != 0; n = im.NextMany(buf) {
				counter += uint64(n)
			}
		}
		b.StopTimer()
	})
	if counter != expectedCardinality {
		b.Fatalf("Cardinalities don't match: %d, %d", counter, expectedCardinality)
	}
}

func BenchmarkAndNotArrayRun(b *testing.B) {
	left, right := andNotArrayRunInputs()
	if left.Stats().ArrayContainers != 1 || right.Stats().RunContainers != 1 {
		b.Fatal("benchmark inputs do not use array/run containers")
	}
	if got := AndNot(left, right).GetCardinality(); got != 1024 {
		b.Fatalf("unexpected result cardinality: got %d, want 1024", got)
	}

	b.ResetTimer()
	for b.Loop() {
		AndNot(left, right)
	}
}

func BenchmarkAndNotArrayArray(b *testing.B) {
	left, right := andNotArrayArrayInputs()
	if left.Stats().ArrayContainers != 1 || right.Stats().ArrayContainers != 1 {
		b.Fatal("benchmark inputs do not use array/array containers")
	}
	if got := AndNot(left, right).GetCardinality(); got != 1920 {
		b.Fatalf("unexpected result cardinality: got %d, want 1920", got)
	}

	b.ResetTimer()
	for b.Loop() {
		AndNot(left, right)
	}
}

// BenchmarkOrAscendingParts unions parts of one container each, in ascending
// key order, one Or per part.
func BenchmarkOrAscendingParts(b *testing.B) {
	parts := make([]*Bitmap, 32768)
	for i := range parts {
		parts[i] = BitmapOf(uint64(i)<<32, uint64(i)<<32|7, uint64(i)<<32|99, uint64(i)<<32|1000, uint64(i)<<32|40000)
	}
	b.ReportAllocs()
	for b.Loop() {
		union := NewBitmap()
		for _, part := range parts {
			union.Or(part)
		}
	}
}
