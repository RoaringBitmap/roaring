//go:build arm64 && !gccgo && !appengine
// +build arm64,!gccgo,!appengine

package roaring

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"
)

const intersectBenchVariants = 8

func benchIntersectPair(shape string, n int, seed int64) (a, b []uint16) {
	r := rand.New(rand.NewSource(int64(42+n) + seed*7919))
	switch shape {
	case "dense50":
		return genSortedUnique(r, n, 2*n), genSortedUnique(r, n, 2*n)
	case "coinflip": // disjoint 8-blocks, shuffled ownership: zero matches
		nb := (n + 7) / 8
		own := make([]bool, 2*nb)
		for i := 0; i < nb; i++ {
			own[i] = true
		}
		r.Shuffle(len(own), func(i, j int) { own[i], own[j] = own[j], own[i] })
		a = make([]uint16, 0, nb*8)
		b = make([]uint16, 0, nb*8)
		base := 0
		for _, toA := range own {
			for x := 0; x < 8; x++ {
				if toA {
					a = append(a, uint16(base+x))
				} else {
					b = append(b, uint16(base+x))
				}
			}
			base += 8
		}
		return a[:n], b[:n]
	case "skew8": // one side 8x longer, below the 64:1 galloping cutoff
		big := 8 * n
		if big > 32768 {
			big = 32768
		}
		return genSortedUnique(r, n, 65536), genSortedUnique(r, big, 65536)
	case "overlap95": // ~95% shared elements: near-total match density
		a = genSortedUnique(r, n, 4*n)
		b = append([]uint16(nil), a...)
		for i := 10; i < n; i += 20 {
			b[i] ^= 1
		}
		sort.Slice(b, func(i, j int) bool { return b[i] < b[j] })
		out := b[:0]
		for i, v := range b {
			if i == 0 || v != b[i-1] {
				out = append(out, v)
			}
		}
		return a, out
	}
	panic("unknown shape")
}

var benchIntersectShapes = []string{"dense50", "coinflip", "skew8", "overlap95"}
var benchIntersectSizes = []int{8, 16, 24, 64, 256, 4096}

func BenchmarkIntersect2By2(b *testing.B) {
	for _, shape := range benchIntersectShapes {
		for _, n := range benchIntersectSizes {
			as := make([][]uint16, intersectBenchVariants)
			bs := make([][]uint16, intersectBenchVariants)
			for v := 0; v < intersectBenchVariants; v++ {
				as[v], bs[v] = benchIntersectPair(shape, n, int64(v))
			}
			// andArray allocates exactly min(len1,len2); mirror that cap.
			mins := make([]int, intersectBenchVariants)
			for v := 0; v < intersectBenchVariants; v++ {
				mins[v] = len(as[v])
				if len(bs[v]) < mins[v] {
					mins[v] = len(bs[v])
				}
			}
			buffer := make([]uint16, n+8)
			scratch := make([]uint16, n+8)
			for _, impl := range []struct {
				name string
				fn   func([]uint16, []uint16, []uint16) int
			}{
				{"dispatch", intersection2by2},
				{"scalar", localintersect2by2},
			} {
				b.Run(fmt.Sprintf("%s/%d/%s", shape, n, impl.name), func(b *testing.B) {
					sink := 0
					for i := 0; i < b.N; i++ {
						v := i % intersectBenchVariants
						sink += impl.fn(as[v], bs[v], buffer[:0:mins[v]])
					}
					_ = sink
				})
			}
			// iandArray geometry; both rows pay the same restore copy.
			if shape != "dense50" || (n != 16 && n != 4096) {
				continue
			}
			for _, impl := range []struct {
				name string
				fn   func([]uint16, []uint16, []uint16) int
			}{
				{"inplace", intersection2by2},
				{"inplaceScalar", localintersect2by2},
			} {
				b.Run(fmt.Sprintf("%s/%d/%s", shape, n, impl.name), func(b *testing.B) {
					sink := 0
					for i := 0; i < b.N; i++ {
						v := i % intersectBenchVariants
						m := copy(scratch, as[v])
						sink += impl.fn(scratch[:m], bs[v], scratch[:0:m])
					}
					_ = sink
				})
			}
		}
	}
}

func BenchmarkIntersectCard2By2(b *testing.B) {
	for _, shape := range benchIntersectShapes {
		for _, n := range benchIntersectSizes {
			as := make([][]uint16, intersectBenchVariants)
			bs := make([][]uint16, intersectBenchVariants)
			for v := 0; v < intersectBenchVariants; v++ {
				as[v], bs[v] = benchIntersectPair(shape, n, int64(v))
			}
			for _, impl := range []struct {
				name string
				fn   func([]uint16, []uint16) int
			}{
				{"dispatch", intersection2by2Cardinality},
				{"scalar", localintersect2by2Cardinality},
			} {
				b.Run(fmt.Sprintf("%s/%d/%s", shape, n, impl.name), func(b *testing.B) {
					sink := 0
					for i := 0; i < b.N; i++ {
						v := i % intersectBenchVariants
						sink += impl.fn(as[v], bs[v])
					}
					_ = sink
				})
			}
		}
	}
}
