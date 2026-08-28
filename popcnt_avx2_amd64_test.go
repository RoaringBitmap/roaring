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

func TestAVX2AndStoreDifferential(t *testing.T) {
	if !useAVX2 {
		t.Skip("AVX2 not available on this CPU")
	}
	r := rand.New(rand.NewSource(43))
	for _, n := range avx2TestLengths {
		for iter := 0; iter < 64; iter++ {
			a := randomUint64Slice(r, n)
			b := randomUint64Slice(r, n)
			want := make([]uint64, n)
			andSliceGo(want, a, b)

			got := make([]uint64, n)
			_andStoreSliceAVX2(got, a, b)
			assert.Equalf(t, want, got, "separate destination len=%d", n)

			aliasA := append(make([]uint64, 0, n), a...)
			_andStoreSliceAVX2(aliasA, aliasA, b)
			assert.Equalf(t, want, aliasA, "destination aliases a len=%d", n)

			aliasB := append(make([]uint64, 0, n), b...)
			_andStoreSliceAVX2(aliasB, a, aliasB)
			assert.Equalf(t, want, aliasB, "destination aliases b len=%d", n)

			wantCard := andCardSliceGo(want, a, b)
			gotCard := make([]uint64, n)
			card := _andCardStoreSliceAVX2(gotCard, a, b)
			assert.Equalf(t, want, gotCard, "fused destination len=%d", n)
			assert.Equalf(t, wantCard, card, "fused cardinality len=%d", n)

			aliasA = append(make([]uint64, 0, n), a...)
			card = _andCardStoreSliceAVX2(aliasA, aliasA, b)
			assert.Equalf(t, want, aliasA, "fused destination aliases a len=%d", n)
			assert.Equalf(t, wantCard, card, "fused aliases a cardinality len=%d", n)

			aliasB = append(make([]uint64, 0, n), b...)
			card = _andCardStoreSliceAVX2(aliasB, a, aliasB)
			assert.Equalf(t, want, aliasB, "fused destination aliases b len=%d", n)
			assert.Equalf(t, wantCard, card, "fused aliases b cardinality len=%d", n)
		}
	}
}

func TestAVX2AndStoreDispatch(t *testing.T) {
	if !useAVX2 {
		t.Skip("AVX2 not available on this CPU")
	}
	savedAVX512 := useAVX512BitsetOps
	savedAVX512Popcnt := useAVX512Popcnt
	savedAVX2 := useAVX2
	defer func() {
		useAVX512BitsetOps = savedAVX512
		useAVX512Popcnt = savedAVX512Popcnt
		useAVX2 = savedAVX2
	}()
	useAVX512BitsetOps = false
	useAVX512Popcnt = false
	useAVX2 = true

	r := rand.New(rand.NewSource(44))
	for _, n := range avx2TestLengths {
		a := randomUint64Slice(r, n)
		b := randomUint64Slice(r, n)
		want := make([]uint64, n)
		andSliceGo(want, a, b)
		got := make([]uint64, n)
		andSlice(got, a, b)
		assert.Equalf(t, want, got, "andSlice AVX2 dispatch len=%d", n)

		gotCard := make([]uint64, n)
		wantCard := andCardSliceGo(want, a, b)
		card := andCardSlice(gotCard, a, b)
		assert.Equalf(t, want, gotCard, "andCardSlice AVX2 dispatch len=%d", n)
		assert.Equalf(t, wantCard, card, "andCardSlice AVX2 cardinality len=%d", n)
	}
}
