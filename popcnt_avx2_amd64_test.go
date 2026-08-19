//go:build amd64 && !appengine
// +build amd64,!appengine

package roaring

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

// edge lengths exercise the AVX2 main loop (multiples of 4) and the scalar
// POPCNTQ tail (len % 4 != 0), including the empty and sub-block cases.
var avx2TestLengths = []int{0, 1, 2, 3, 4, 5, 7, 8, 15, 16, 17, 31, 63, 64, 65, 1023, 1024, 1025}

func randomUint64Slice(r *rand.Rand, n int) []uint64 {
	s := make([]uint64, n)
	for i := range s {
		// mix fully-random words with sparse and dense ones to vary popcounts.
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

func benchPopcntPair(b *testing.B, avx2 bool, fn func(s, m []uint64) uint64) {
	if avx2 && !useAVX2 {
		b.Skip("AVX2 not available")
	}
	r := rand.New(rand.NewSource(1))
	s := randomUint64Slice(r, 1024)
	m := randomUint64Slice(r, 1024)
	var sink uint64
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink += fn(s, m)
	}
	_ = sink
}

func BenchmarkPopcntAndSlice1024AVX2(b *testing.B) {
	benchPopcntPair(b, true, _popcntAndSliceAVX2)
}

func BenchmarkPopcntAndSlice1024Go(b *testing.B) {
	benchPopcntPair(b, false, popcntAndSliceGo)
}

func BenchmarkPopcntSlice1024AVX2(b *testing.B) {
	if !useAVX2 {
		b.Skip("AVX2 not available")
	}
	r := rand.New(rand.NewSource(1))
	s := randomUint64Slice(r, 1024)
	var sink uint64
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink += _popcntSliceAVX2(s)
	}
	_ = sink
}

func BenchmarkPopcntSlice1024Go(b *testing.B) {
	r := rand.New(rand.NewSource(1))
	s := randomUint64Slice(r, 1024)
	var sink uint64
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink += popcntSliceGo(s)
	}
	_ = sink
}

func TestAVX2PopcntDispatch(t *testing.T) {
	// Verify the runtime dispatch wrappers agree with the Go reference both
	// when AVX2 is selected and when the scalar fallback is forced.
	saved := useAVX2
	defer func() { useAVX2 = saved }()

	r := rand.New(rand.NewSource(7))
	for _, on := range []bool{false, true} {
		if on && !saved {
			continue // CPU has no AVX2; only the fallback exists
		}
		useAVX2 = on
		for _, n := range avx2TestLengths {
			s := randomUint64Slice(r, n)
			m := randomUint64Slice(r, n)
			assert.Equalf(t, popcntSliceGo(s), popcntSlice(s), "popcntSlice avx2=%v len=%d", on, n)
			assert.Equalf(t, popcntAndSliceGo(s, m), popcntAndSlice(s, m), "popcntAndSlice avx2=%v len=%d", on, n)
			assert.Equalf(t, popcntOrSliceGo(s, m), popcntOrSlice(s, m), "popcntOrSlice avx2=%v len=%d", on, n)
			assert.Equalf(t, popcntXorSliceGo(s, m), popcntXorSlice(s, m), "popcntXorSlice avx2=%v len=%d", on, n)
			assert.Equalf(t, popcntMaskSliceGo(s, m), popcntMaskSlice(s, m), "popcntMaskSlice avx2=%v len=%d", on, n)
		}
	}
}

func TestAVX2PopcntDifferential(t *testing.T) {
	if !useAVX2 {
		t.Skip("AVX2 not available on this CPU")
	}
	r := rand.New(rand.NewSource(42))
	for _, n := range avx2TestLengths {
		for iter := 0; iter < 64; iter++ {
			s := randomUint64Slice(r, n)
			m := randomUint64Slice(r, n)

			assert.Equalf(t, popcntSliceGo(s), _popcntSliceAVX2(s),
				"popcntSlice len=%d", n)
			assert.Equalf(t, popcntAndSliceGo(s, m), _popcntAndSliceAVX2(s, m),
				"popcntAndSlice len=%d", n)
			assert.Equalf(t, popcntOrSliceGo(s, m), _popcntOrSliceAVX2(s, m),
				"popcntOrSlice len=%d", n)
			assert.Equalf(t, popcntXorSliceGo(s, m), _popcntXorSliceAVX2(s, m),
				"popcntXorSlice len=%d", n)
			assert.Equalf(t, popcntMaskSliceGo(s, m), _popcntMaskSliceAVX2(s, m),
				"popcntMaskSlice len=%d", n)
		}
	}
}

func TestAVX2AndStoreSliceDispatch(t *testing.T) {
	saved := useAVX2
	defer func() { useAVX2 = saved }()

	r := rand.New(rand.NewSource(11))
	for _, on := range []bool{false, true} {
		if on && !saved {
			continue // CPU has no AVX2; only the fallback exists
		}
		useAVX2 = on
		for _, n := range avx2TestLengths {
			s := randomUint64Slice(r, n)
			m := randomUint64Slice(r, n)
			expected := append([]uint64(nil), s...)
			andStoreSliceGo(expected, s, m)

			dst := append([]uint64(nil), s...)
			andStoreSlice(dst, s, m)
			assert.Equalf(t, expected, dst, "separate destination avx2=%v len=%d", on, n)

			inPlaceLeft := append([]uint64(nil), s...)
			andStoreSlice(inPlaceLeft, inPlaceLeft, m)
			assert.Equalf(t, expected, inPlaceLeft, "left alias avx2=%v len=%d", on, n)

			inPlaceRight := append([]uint64(nil), m...)
			andStoreSlice(inPlaceRight, s, inPlaceRight)
			assert.Equalf(t, expected, inPlaceRight, "right alias avx2=%v len=%d", on, n)
		}
	}
}

func TestAVX2AndStoreSliceDifferential(t *testing.T) {
	if !useAVX2 {
		t.Skip("AVX2 not available on this CPU")
	}
	r := rand.New(rand.NewSource(13))
	for _, n := range avx2TestLengths {
		for iter := 0; iter < 64; iter++ {
			s := randomUint64Slice(r, n)
			m := randomUint64Slice(r, n)
			expected := append([]uint64(nil), s...)
			andStoreSliceGo(expected, s, m)

			dst := append([]uint64(nil), s...)
			_andStoreSliceAVX2(dst, s, m)
			assert.Equalf(t, expected, dst, "separate destination len=%d", n)

			inPlaceLeft := append([]uint64(nil), s...)
			_andStoreSliceAVX2(inPlaceLeft, inPlaceLeft, m)
			assert.Equalf(t, expected, inPlaceLeft, "left alias len=%d", n)

			inPlaceRight := append([]uint64(nil), m...)
			_andStoreSliceAVX2(inPlaceRight, s, inPlaceRight)
			assert.Equalf(t, expected, inPlaceRight, "right alias len=%d", n)
		}
	}
}

func TestAVX2BitmapAndStoreDispatch(t *testing.T) {
	saved := useAVX2
	defer func() { useAVX2 = saved }()

	newFull := func() *bitmapContainer {
		bc := newBitmapContainer()
		for i := range bc.bitmap {
			bc.bitmap[i] = ^uint64(0)
		}
		bc.cardinality = maxCapacity
		return bc
	}

	right := newBitmapContainer()
	for i := range right.bitmap {
		right.bitmap[i] = 0xaaaaaaaaaaaaaaaa
	}
	right.cardinality = maxCapacity / 2

	small := newBitmapContainer()
	small.bitmap[0] = 1
	small.cardinality = 1

	for _, on := range []bool{false, true} {
		if on && !saved {
			continue // CPU has no AVX2; only the fallback exists
		}
		useAVX2 = on

		answer, ok := newFull().andBitmap(right).(*bitmapContainer)
		if !ok {
			t.Fatalf("expected bitmap result with avx2=%v", on)
		}
		assert.Equalf(t, right.bitmap, answer.bitmap, "bitmap result avx2=%v", on)
		assert.Equalf(t, right.cardinality, answer.cardinality, "bitmap cardinality avx2=%v", on)

		inPlace := newFull()
		answer, ok = inPlace.iandBitmap(right).(*bitmapContainer)
		if !ok {
			t.Fatalf("expected in-place bitmap result with avx2=%v", on)
		}
		assert.Samef(t, inPlace, answer, "in-place bitmap result avx2=%v", on)
		assert.Equalf(t, right.bitmap, inPlace.bitmap, "in-place bitmap contents avx2=%v", on)

		arrayResult, ok := newFull().andBitmap(small).(*arrayContainer)
		if !ok {
			t.Fatalf("expected array result with avx2=%v", on)
		}
		assert.Equalf(t, []uint16{0}, arrayResult.content, "array result avx2=%v", on)

		inPlace = newFull()
		arrayResult, ok = inPlace.iandBitmap(small).(*arrayContainer)
		if !ok {
			t.Fatalf("expected in-place array result with avx2=%v", on)
		}
		assert.Equalf(t, []uint16{0}, arrayResult.content, "in-place array result avx2=%v", on)
	}
}
