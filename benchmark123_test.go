//go:build go1.23
// +build go1.23

package roaring

import (
	"math/rand"
	"testing"
)

func BenchmarkIterator123(b *testing.B) {
	bm := NewBitmap()
	domain := 100000000
	count := 10000
	for j := 0; j < count; j++ {
		v := uint32(rand.Intn(domain))
		bm.Add(v)
	}
	i := IntIterator{}
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
			Values(bm)(func(_ uint32) bool {
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
	ir := IntReverseIterator{}

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
			Backward(bm)(func(_ uint32) bool {
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
			buf := make([]uint32, 1024)
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
	im := ManyIntIterator{}
	buf := make([]uint32, 1024)

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

	b.Run("values iteration 1.23", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			counter = 0
			for range Values(bm) {
				counter++
			}
		}
		b.StopTimer()
	})
	if counter != expectedCardinality {
		b.Fatalf("Cardinalities don't match: %d, %d", counter, expectedCardinality)
	}

	b.Run("backward iteration 1.23", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			counter = 0
			for range Backward(bm) {
				counter++
			}
		}
		b.StopTimer()
	})
	if counter != expectedCardinality {
		b.Fatalf("Cardinalities don't match: %d, %d", counter, expectedCardinality)
	}
}
