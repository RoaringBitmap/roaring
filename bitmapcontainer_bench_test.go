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

const bitmapXorDenseWord = uint64(0xaaaaaaaaaaaaaaaa)

// newBitmapContainerXorFixture creates two bitmap containers with a chosen
// symmetric-difference cardinality. Both inputs remain dense bitmap containers.
func newBitmapContainerXorFixture(xorCardinality int) (*bitmapContainer, *bitmapContainer) {
	left := newBitmapContainer()
	right := newBitmapContainer()
	for i := range left.bitmap {
		left.bitmap[i] = bitmapXorDenseWord
		right.bitmap[i] = bitmapXorDenseWord
	}
	for i := 0; i < xorCardinality; i++ {
		right.bitmap[i/64] ^= uint64(1) << (i % 64)
	}
	left.computeCardinality()
	right.computeCardinality()
	return left, right
}

func newDenseXorBitmapFixture(containers int) (*Bitmap, *Bitmap) {
	leftWords := make([]uint64, containers*bitmapContainerSize)
	rightWords := make([]uint64, len(leftWords))
	for i := range leftWords {
		leftWords[i] = bitmapXorDenseWord
		rightWords[i] = ^bitmapXorDenseWord
	}
	return FromDense(leftWords, true), FromDense(rightWords, true)
}

// BenchmarkBitmapXorDenseContainers measures Bitmap.Xor with matching dense
// keys. Each operation alternates between a half-full and full bitmap result,
// so every matching pair takes ixorBitmap's bitmap-result branch.
func BenchmarkBitmapXorDenseContainers(b *testing.B) {
	for _, benchmark := range []struct {
		name       string
		containers int
	}{
		{name: "one", containers: 1},
		{name: "fifty", containers: 50},
		{name: "two-fifty-six", containers: 256},
	} {
		b.Run(benchmark.name, func(b *testing.B) {
			left, right := newDenseXorBitmapFixture(benchmark.containers)

			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				left.Xor(right)
			}
			b.StopTimer()

			want := uint64(benchmark.containers * maxCapacity / 2)
			if b.N%2 != 0 {
				want *= 2
			}
			if got := left.GetCardinality(); got != want {
				b.Fatalf("unexpected cardinality: got %d, want %d", got, want)
			}
		})
	}
}

func newDenseXorThresholdFixture() (*Bitmap, *Bitmap) {
	leftWords := make([]uint64, bitmapContainerSize)
	rightWords := make([]uint64, bitmapContainerSize)
	for i := range leftWords {
		leftWords[i] = bitmapXorDenseWord
		rightWords[i] = bitmapXorDenseWord
	}
	for i := 0; i < arrayDefaultMaxSize; i++ {
		rightWords[i/64] ^= uint64(1) << (i % 64)
	}
	return FromDense(leftWords, true), FromDense(rightWords, true)
}

// BenchmarkBitmapXorDenseThresholdBatch measures independent ordinary Xor
// calls at the array-result threshold. It resets receivers outside the timed
// region and batches calls to reduce timer noise without changing the path.
func BenchmarkBitmapXorDenseThresholdBatch(b *testing.B) {
	const batchSize = 16

	lefts := make([]*Bitmap, batchSize)
	originals := make([]container, batchSize)
	_, right := newDenseXorThresholdFixture()
	for i := range lefts {
		lefts[i], _ = newDenseXorThresholdFixture()
		originals[i] = lefts[i].highlowcontainer.getContainerAtIndex(0)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		for _, left := range lefts {
			left.Xor(right)
		}
		b.StopTimer()
		for i, left := range lefts {
			left.highlowcontainer.setContainerAtIndex(0, originals[i])
		}
		b.StartTimer()
	}
	b.StopTimer()

	lefts[0].Xor(right)
	if got := lefts[0].GetCardinality(); got != arrayDefaultMaxSize {
		b.Fatalf("unexpected cardinality: got %d, want %d", got, arrayDefaultMaxSize)
	}
	if _, ok := lefts[0].highlowcontainer.getContainerAtIndex(0).(*arrayContainer); !ok {
		b.Fatal("expected an array result")
	}
}
