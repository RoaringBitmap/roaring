//go:build arm64 && !gccgo && !appengine
// +build arm64,!gccgo,!appengine

package roaring

import (
	"fmt"
	"math/rand"
	"testing"
)

func benchPairSeed(shape string, n int, seed int64) (a, b []uint16) {
	r := rand.New(rand.NewSource(int64(42+n) + seed*7919))
	switch shape {
	case "dense50":
		return genSortedUnique(r, n, 2*n), genSortedUnique(r, n, 2*n)
	case "sparse6":
		vr := n * 16
		if vr > 65536 {
			vr = 65536
		}
		return genSortedUnique(r, n, vr), genSortedUnique(r, n, vr)
	case "runs16":
		a = make([]uint16, n)
		b = make([]uint16, n)
		for i := 0; i < n; i++ {
			blk, off := i/16, i%16
			a[i] = uint16(blk*32 + off)
			b[i] = uint16(blk*32 + 16 + off)
		}
		return a, b
	case "spread": // density mismatch: a confined to 1/8 of b's range
		vr := 2 * n
		if vr > 8192 {
			vr = 8192
		}
		return genSortedUnique(r, n, vr), genSortedUnique(r, n, 65536)
	}
	panic("unknown shape")
}

// Rotating fixed-seed variants keeps the branch predictor from memorizing
// one pair's decision sequence, which flatters the scalar path.
const benchVariants = 16

func BenchmarkUnion2By2(b *testing.B) {
	for _, shape := range []string{"dense50", "sparse6", "runs16", "spread"} {
		for _, n := range []int{256, 384, 512, 1024, 4096} {
			as := make([][]uint16, benchVariants)
			bs := make([][]uint16, benchVariants)
			for v := 0; v < benchVariants; v++ {
				as[v], bs[v] = benchPairSeed(shape, n, int64(v))
			}
			buffer := make([]uint16, 2*n)
			for _, impl := range []struct {
				name string
				fn   func([]uint16, []uint16, []uint16) int
			}{
				{"dispatch", union2by2},
				{"scalar", union2by2scalar},
			} {
				b.Run(fmt.Sprintf("%s/%d/%s", shape, n, impl.name), func(b *testing.B) {
					sink := 0
					for i := 0; i < b.N; i++ {
						v := i % benchVariants
						sink += impl.fn(as[v], bs[v], buffer)
					}
					_ = sink
				})
			}
		}
	}
}
