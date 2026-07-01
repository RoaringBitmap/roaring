//go:build arm64 && !appengine
// +build arm64,!appengine

package roaring

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

// edge lengths exercise the NEON main loop (multiples of 2) and the scalar-width
// tail (len % 2 != 0), including the empty and sub-block cases.
var neonTestLengths = []int{0, 1, 2, 3, 4, 5, 7, 8, 15, 16, 17, 31, 63, 64, 65, 1023, 1024, 1025}

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

func benchPopcntPair(b *testing.B, neon bool, fn func(s, m []uint64) uint64) {
	if neon && !useNEON {
		b.Skip("NEON not available")
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

func BenchmarkPopcntAndSlice1024NEON(b *testing.B) {
	benchPopcntPair(b, true, _popcntAndSliceNEON)
}

func BenchmarkPopcntAndSlice1024Go(b *testing.B) {
	benchPopcntPair(b, false, popcntAndSliceGo)
}

func BenchmarkPopcntSlice1024NEON(b *testing.B) {
	if !useNEON {
		b.Skip("NEON not available")
	}
	r := rand.New(rand.NewSource(1))
	s := randomUint64Slice(r, 1024)
	var sink uint64
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink += _popcntSliceNEON(s)
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

func TestNEONPopcntDispatch(t *testing.T) {
	// Verify the runtime dispatch wrappers agree with the Go reference both when
	// NEON is selected and when the scalar fallback is forced.
	saved := useNEON
	defer func() { useNEON = saved }()

	r := rand.New(rand.NewSource(7))
	for _, on := range []bool{false, true} {
		useNEON = on
		for _, n := range neonTestLengths {
			s := randomUint64Slice(r, n)
			m := randomUint64Slice(r, n)
			assert.Equalf(t, popcntSliceGo(s), popcntSlice(s), "popcntSlice neon=%v len=%d", on, n)
			assert.Equalf(t, popcntAndSliceGo(s, m), popcntAndSlice(s, m), "popcntAndSlice neon=%v len=%d", on, n)
			assert.Equalf(t, popcntOrSliceGo(s, m), popcntOrSlice(s, m), "popcntOrSlice neon=%v len=%d", on, n)
			assert.Equalf(t, popcntXorSliceGo(s, m), popcntXorSlice(s, m), "popcntXorSlice neon=%v len=%d", on, n)
			assert.Equalf(t, popcntMaskSliceGo(s, m), popcntMaskSlice(s, m), "popcntMaskSlice neon=%v len=%d", on, n)
		}
	}
}

func TestNEONPopcntDifferential(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	for _, n := range neonTestLengths {
		for iter := 0; iter < 64; iter++ {
			s := randomUint64Slice(r, n)
			m := randomUint64Slice(r, n)

			assert.Equalf(t, popcntSliceGo(s), _popcntSliceNEON(s),
				"popcntSlice len=%d", n)
			assert.Equalf(t, popcntAndSliceGo(s, m), _popcntAndSliceNEON(s, m),
				"popcntAndSlice len=%d", n)
			assert.Equalf(t, popcntOrSliceGo(s, m), _popcntOrSliceNEON(s, m),
				"popcntOrSlice len=%d", n)
			assert.Equalf(t, popcntXorSliceGo(s, m), _popcntXorSliceNEON(s, m),
				"popcntXorSlice len=%d", n)
			assert.Equalf(t, popcntMaskSliceGo(s, m), _popcntMaskSliceNEON(s, m),
				"popcntMaskSlice len=%d", n)
		}
	}
}
