package roaring

import (
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

func roaringAndBitmapFixture(ranges ...[2]int) *Bitmap {
	words := make([]uint64, bitmapContainerSize)
	for _, r := range ranges {
		for value := r[0]; value < r[1]; value++ {
			words[value/64] |= uint64(1) << uint(value&63)
		}
	}
	return FromDense(words, false)
}

func benchmarkRoaringAndBitmap(b *testing.B, left, right *Bitmap, expected uint64) {
	b.Helper()
	for _, bitmap := range []*Bitmap{left, right} {
		if bitmap.highlowcontainer.size() != 1 {
			b.Fatal("workload did not produce one container")
		}
		if _, ok := bitmap.highlowcontainer.getContainerAtIndex(0).(*bitmapContainer); !ok {
			b.Fatal("workload did not produce bitmap containers")
		}
	}

	if result := And(left, right); result.GetCardinality() != expected {
		b.Fatalf("unexpected cardinality: got %d, want %d", result.GetCardinality(), expected)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		And(left, right)
	}
}

func BenchmarkRoaringAndBitmapDense(b *testing.B) {
	left := roaringAndBitmapFixture([2]int{0, maxCapacity})
	right := roaringAndBitmapFixture([2]int{0, 59000})
	benchmarkRoaringAndBitmap(b, left, right, 59000)
}

func BenchmarkRoaringAndBitmapSparse(b *testing.B) {
	left := roaringAndBitmapFixture([2]int{0, 5000}, [2]int{10000, 15000})
	right := roaringAndBitmapFixture([2]int{0, 5000}, [2]int{20000, 25000})
	benchmarkRoaringAndBitmap(b, left, right, 5000)
}

func BenchmarkRoaringAndBitmapDisjoint(b *testing.B) {
	left := roaringAndBitmapFixture([2]int{0, 5000})
	right := roaringAndBitmapFixture([2]int{10000, 15000})
	benchmarkRoaringAndBitmap(b, left, right, 0)
}
