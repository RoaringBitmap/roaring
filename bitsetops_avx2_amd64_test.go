//go:build amd64 && !appengine
// +build amd64,!appengine

package roaring

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

var xorSliceAVX2TestLengths = []int{0, 1, 3, 4, 7, 15, 16, 17, 31, 32, 33, 1023, 1024, 1025}

func TestXorSliceAVX2(t *testing.T) {
	if !useAVX2 {
		t.Skip("AVX2 not available on this CPU")
	}

	savedAVX2, savedAVX512 := useAVX2, useAVX512BitsetOps
	useAVX2 = true
	useAVX512BitsetOps = false
	defer func() {
		useAVX2 = savedAVX2
		useAVX512BitsetOps = savedAVX512
	}()

	r := rand.New(rand.NewSource(17))
	for _, n := range xorSliceAVX2TestLengths {
		a := randomWords(r, n)
		b := randomWords(r, n)
		want := make([]uint64, n)
		xorSliceGo(want, a, b)

		dst := make([]uint64, n)
		xorSliceAVX2(dst, a, b)
		assert.Equalf(t, want, dst, "xorSliceAVX2 len=%d", n)

		dst = append(make([]uint64, 0, n), a...)
		xorSliceAVX2(dst, dst, b)
		assert.Equalf(t, want, dst, "xorSliceAVX2 dst==a len=%d", n)

		dst = append(make([]uint64, 0, n), b...)
		xorSliceAVX2(dst, a, dst)
		assert.Equalf(t, want, dst, "xorSliceAVX2 dst==b len=%d", n)

		dst = append(make([]uint64, 0, n), a...)
		xorSlice(dst, dst, b)
		assert.Equalf(t, want, dst, "xorSlice dispatch len=%d", n)
	}
}
