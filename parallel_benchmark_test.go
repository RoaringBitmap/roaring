package roaring

import (
	"math/rand"
	"testing"
)

func bitmapIntersectionBenchmarkFixture(containerCount int, word uint64) *Bitmap {
	bitmap := make([]uint64, containerCount*(maxCapacity/64))
	for i := range bitmap {
		bitmap[i] = word
	}
	return FromDense(bitmap, false)
}

// BenchmarkIntersectionBitmapParallel exercises ParAnd with bitmap containers
// whose intersection stays above the array-container threshold.
func BenchmarkIntersectionBitmapParallel(b *testing.B) {
	const containerCount = 128
	left := bitmapIntersectionBenchmarkFixture(containerCount, ^uint64(0))
	right := bitmapIntersectionBenchmarkFixture(containerCount, 0x5555555555555555)
	want := left.AndCardinality(right)

	for b.Loop() {
		result := ParAnd(0, left, right)
		if result.GetCardinality() != want {
			b.Fatalf("got %d values, want %d", result.GetCardinality(), want)
		}
	}
}

func BenchmarkIntersectionLargeParallel(b *testing.B) {
	b.StopTimer()

	initsize := 650000
	r := rand.New(rand.NewSource(0))

	s1 := NewBitmap()
	sz := 150 * 1000 * 1000
	for i := 0; i < initsize; i++ {
		s1.Add(uint32(r.Int31n(int32(sz))))
	}

	s2 := NewBitmap()
	sz = 100 * 1000 * 1000
	for i := 0; i < initsize; i++ {
		s2.Add(uint32(r.Int31n(int32(sz))))
	}

	b.StartTimer()
	card := uint64(0)
	for j := 0; j < b.N; j++ {
		s3 := ParAnd(0, s1, s2)
		card = card + s3.GetCardinality()
	}
}

func BenchmarkIntersectionLargeRoaring(b *testing.B) {
	b.StopTimer()
	initsize := 650000
	r := rand.New(rand.NewSource(0))

	s1 := NewBitmap()
	sz := 150 * 1000 * 1000
	for i := 0; i < initsize; i++ {
		s1.Add(uint32(r.Int31n(int32(sz))))
	}

	s2 := NewBitmap()
	sz = 100 * 1000 * 1000
	for i := 0; i < initsize; i++ {
		s2.Add(uint32(r.Int31n(int32(sz))))
	}

	b.StartTimer()
	card := uint64(0)
	for j := 0; j < b.N; j++ {
		s3 := And(s1, s2)
		card = card + s3.GetCardinality()
	}
}
