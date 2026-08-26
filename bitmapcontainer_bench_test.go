package roaring

import (
	"math/bits"
	"math/rand"
	"testing"
)

var sink uint32

func BenchmarkBitmapContainerFillLeastSignificant16bits(b *testing.B) {
	r := rand.New(rand.NewSource(42))
	bc := newBitmapContainer()
	for i := 0; i < 32768; i++ {
		val := uint16(r.Intn(65536))
		bc.iadd(val)
	}

	x := make([]uint32, 65536)
	mask := uint32(123) << 16

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pos := bc.fillLeastSignificant16bits(x, 0, mask)
		sink += x[pos-1]
	}
}

// BenchmarkBitmapToArrayDenseContainers measures ToArray across four dense
// bitmap containers and validates its complete ordered output before timing.
func BenchmarkBitmapToArrayDenseContainers(b *testing.B) {
	const containerCount = 4
	words := make([]uint64, bitmapContainerSize*containerCount)
	state := uint64(1)
	for i := range words {
		state += 0x9e3779b97f4a7c15
		word := state
		word = (word ^ (word >> 30)) * 0xbf58476d1ce4e5b9
		word = (word ^ (word >> 27)) * 0x94d049bb133111eb
		words[i] = word ^ (word >> 31)
	}

	bitmap := FromDense(words, false)
	if len(bitmap.highlowcontainer.containers) != containerCount {
		b.Fatalf("workload has %d containers, want %d", len(bitmap.highlowcontainer.containers), containerCount)
	}
	for _, c := range bitmap.highlowcontainer.containers {
		if _, ok := c.(*bitmapContainer); !ok {
			b.Fatal("workload did not produce bitmap containers")
		}
	}

	expected := make([]uint32, 0, len(words)*32)
	for i, word := range words {
		base := uint32(i/bitmapContainerSize)<<16 + uint32(i%bitmapContainerSize*64)
		for word != 0 {
			expected = append(expected, base+uint32(bits.TrailingZeros64(word)))
			word &= word - 1
		}
	}
	actual := bitmap.ToArray()
	if len(actual) != len(expected) {
		b.Fatalf("validation returned %d values, want %d", len(actual), len(expected))
	}
	for i := range expected {
		if actual[i] != expected[i] {
			b.Fatalf("validation mismatch at %d: got %d, want %d", i, actual[i], expected[i])
		}
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		result := bitmap.ToArray()
		sink += result[len(result)-1]
	}
}

// BenchmarkParOrBitmapContainers measures ParOr across four inputs with dense
// bitmap containers at 64 shared keys using four workers.
func BenchmarkParOrBitmapContainers(b *testing.B) {
	const (
		bitmapCount         = 4
		containersPerBitmap = 64
		parallelism         = 4
	)

	bitmaps := make([]*Bitmap, bitmapCount)
	for i := range bitmaps {
		words := make([]uint64, bitmapContainerSize*containersPerBitmap)
		state := uint64(i + 1)
		for j := range words {
			state += 0x9e3779b97f4a7c15
			word := state
			word = (word ^ (word >> 30)) * 0xbf58476d1ce4e5b9
			word = (word ^ (word >> 27)) * 0x94d049bb133111eb
			words[j] = word ^ (word >> 31)
		}
		bitmaps[i] = FromDense(words, false)
		for _, c := range bitmaps[i].highlowcontainer.containers {
			if _, ok := c.(*bitmapContainer); !ok {
				b.Fatal("workload did not produce bitmap containers")
			}
		}
	}

	expected := bitmaps[0].Clone()
	for _, bitmap := range bitmaps[1:] {
		expected.Or(bitmap)
	}
	expectedCardinality := expected.GetCardinality()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		result := ParOr(parallelism, bitmaps...)
		if result.GetCardinality() != expectedCardinality {
			b.Fatal("unexpected cardinality")
		}
	}
}
