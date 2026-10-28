package roaring

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

// lengths exercise the vector main loop (multiples of 32 words), the scalar
// tail (len % 32 != 0), and the empty case.
var bitsetOpsTestLengths = []int{0, 1, 3, 8, 31, 32, 33, 64, 65, 1023, 1024, 1025}

func randomWords(r *rand.Rand, n int) []uint64 {
	s := make([]uint64, n)
	for i := range s {
		switch i % 4 {
		case 0:
			s[i] = r.Uint64()
		case 1:
			s[i] = 0
		case 2:
			s[i] = ^uint64(0)
		default:
			s[i] = r.Uint64() & r.Uint64()
		}
	}
	return s
}

func TestBitsetOpsMatchGo(t *testing.T) {
	r := rand.New(rand.NewSource(3))
	for _, n := range bitsetOpsTestLengths {
		a := randomWords(r, n)
		b := randomWords(r, n)
		want := make([]uint64, n)
		got := make([]uint64, n)

		orSliceGo(want, a, b)
		orSlice(got, a, b)
		assert.Equalf(t, want, got, "orSlice len=%d", n)

		andSliceGo(want, a, b)
		andSlice(got, a, b)
		assert.Equalf(t, want, got, "andSlice len=%d", n)

		xorSliceGo(want, a, b)
		xorSlice(got, a, b)
		assert.Equalf(t, want, got, "xorSlice len=%d", n)

		andNotSliceGo(want, a, b)
		andNotSlice(got, a, b)
		assert.Equalf(t, want, got, "andNotSlice len=%d", n)

		wantCard := orCardSliceGo(want, a, b)
		gotCard := orCardSlice(got, a, b)
		assert.Equalf(t, want, got, "orCardSlice result len=%d", n)
		assert.Equalf(t, wantCard, gotCard, "orCardSlice cardinality len=%d", n)

		wantCard = andCardSliceGo(want, a, b)
		gotCard = andCardSlice(got, a, b)
		assert.Equalf(t, want, got, "andCardSlice result len=%d", n)
		assert.Equalf(t, wantCard, gotCard, "andCardSlice cardinality len=%d", n)
	}
}

// The container operations pass the destination as one of the sources; make
// sure aliasing is handled.
func TestBitsetOpsAliasing(t *testing.T) {
	r := rand.New(rand.NewSource(4))
	for _, n := range bitsetOpsTestLengths {
		a := randomWords(r, n)
		b := randomWords(r, n)

		// copyOf keeps the empty case an empty slice rather than nil, so the
		// comparisons below stay about the contents.
		copyOf := func(src []uint64) []uint64 {
			out := make([]uint64, len(src))
			copy(out, src)
			return out
		}
		want := make([]uint64, n)
		orSliceGo(want, a, b)
		dst := copyOf(a)
		orSlice(dst, dst, b)
		assert.Equalf(t, want, dst, "orSlice dst==a len=%d", n)

		andNotSliceGo(want, a, b)
		dst = copyOf(a)
		andNotSlice(dst, dst, b)
		assert.Equalf(t, want, dst, "andNotSlice dst==a len=%d", n)

		wantCard := andCardSliceGo(want, a, b)
		dst = copyOf(a)
		gotCard := andCardSlice(dst, dst, b)
		assert.Equalf(t, want, dst, "andCardSlice dst==a len=%d", n)
		assert.Equalf(t, wantCard, gotCard, "andCardSlice dst==a cardinality len=%d", n)

		// destination aliasing the second source
		orSliceGo(want, a, b)
		dst = copyOf(b)
		orSlice(dst, a, dst)
		assert.Equalf(t, want, dst, "orSlice dst==b len=%d", n)
	}
}

func benchBitsetOp(b *testing.B, fn func(dst, x, y []uint64)) {
	r := rand.New(rand.NewSource(1))
	x := randomWords(r, bitmapContainerSize)
	y := randomWords(r, bitmapContainerSize)
	dst := make([]uint64, bitmapContainerSize)
	b.SetBytes(int64(bitmapContainerSize * 8))
	for b.Loop() {
		fn(dst, x, y)
	}
}

func BenchmarkOrSlice(b *testing.B)  { benchBitsetOp(b, orSlice) }
func BenchmarkAndSlice(b *testing.B) { benchBitsetOp(b, andSlice) }

func BenchmarkOrCardSlice(b *testing.B) {
	r := rand.New(rand.NewSource(1))
	x := randomWords(r, bitmapContainerSize)
	y := randomWords(r, bitmapContainerSize)
	dst := make([]uint64, bitmapContainerSize)
	var sink uint64
	b.SetBytes(int64(bitmapContainerSize * 8))
	for b.Loop() {
		sink = orCardSlice(dst, x, y)
	}
	_ = sink
}

// The shape the container code used before: write with a scalar loop, then
// make a second pass to count.
func BenchmarkOrThenCountTwoPass(b *testing.B) {
	r := rand.New(rand.NewSource(1))
	x := randomWords(r, bitmapContainerSize)
	y := randomWords(r, bitmapContainerSize)
	dst := make([]uint64, bitmapContainerSize)
	var sink uint64
	b.SetBytes(int64(bitmapContainerSize * 8))
	for b.Loop() {
		orSliceGo(dst, x, y)
		sink = popcntSlice(dst)
	}
	_ = sink
}

// two bitmaps of dense bitmap containers: the case the bitmap-container word
// operations actually govern
func denseBitmapPair(containers int) (*Bitmap, *Bitmap) {
	r := rand.New(rand.NewSource(9))
	mk := func() *Bitmap {
		words := make([]uint64, bitmapContainerSize*containers)
		for i := range words {
			words[i] = r.Uint64()
		}
		return FromDense(words, false)
	}
	return mk(), mk()
}

func BenchmarkDenseBitmapOps(b *testing.B) {
	x, y := denseBitmapPair(64)
	b.Run("Or", func(b *testing.B) {
		for b.Loop() {
			sinkU += Or(x, y).GetCardinality()
		}
	})
	b.Run("And", func(b *testing.B) {
		for b.Loop() {
			sinkU += And(x, y).GetCardinality()
		}
	})
	b.Run("IAnd", func(b *testing.B) {
		for b.Loop() {
			z := x.Clone()
			z.And(y)
			sinkU += z.GetCardinality()
		}
	})
	b.Run("Xor", func(b *testing.B) {
		for b.Loop() {
			sinkU += Xor(x, y).GetCardinality()
		}
	})
	b.Run("AndNot", func(b *testing.B) {
		for b.Loop() {
			sinkU += AndNot(x, y).GetCardinality()
		}
	})
	b.Run("IOr", func(b *testing.B) {
		for b.Loop() {
			z := x.Clone()
			z.Or(y)
			sinkU += z.GetCardinality()
		}
	})
}

var sinkU uint64
