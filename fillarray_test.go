package roaring

import (
	"math/bits"
	"math/rand"
	"testing"
)

func fillArrayTestBitmap(cardinality int, seed int64) *bitmapContainer {
	r := rand.New(rand.NewSource(seed))
	bc := newBitmapContainer()
	for bc.cardinality < cardinality {
		bc.iadd(uint16(r.Intn(65536)))
	}
	return bc
}

func wordsToValues(words []uint64) []uint16 {
	out := make([]uint16, 0, 65536)
	for k, w := range words {
		for w != 0 {
			out = append(out, uint16(k*64+bits.TrailingZeros64(w)))
			w &= w - 1
		}
	}
	return out
}

func combineWords(a, b []uint64, op func(x, y uint64) uint64) []uint64 {
	out := make([]uint64, len(a))
	for i := range a {
		out[i] = op(a[i], b[i])
	}
	return out
}

// cardinalities straddle the skip/compress crossover and the 32-values-per-
// store and 64-values-per-word boundaries.
var fillArrayTestCardinalities = []int{0, 1, 31, 32, 33, 63, 64, 65, 511, 512, 513, 4096, 30000, 65535, 65536}

func TestFillArrayMatchesScalar(t *testing.T) {
	for _, cardinality := range fillArrayTestCardinalities {
		bc := fillArrayTestBitmap(cardinality, int64(cardinality)+1)

		want := make([]uint16, cardinality)
		fillArrayScalar(bc.bitmap, want)
		got := make([]uint16, cardinality)
		bc.fillArray(got)
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("fillArray card=%d at %d: got %d want %d", cardinality, i, got[i], want[i])
			}
		}

		other := fillArrayTestBitmap(cardinality, int64(cardinality)+99)
		for _, tc := range []struct {
			name string
			op   func(x, y uint64) uint64
			fn   func(container []uint16, bitmap1, bitmap2 []uint64)
		}{
			{"AND", func(x, y uint64) uint64 { return x & y }, fillArrayAND},
			{"ANDNOT", func(x, y uint64) uint64 { return x &^ y }, fillArrayANDNOT},
			{"XOR", func(x, y uint64) uint64 { return x ^ y }, fillArrayXOR},
		} {
			combined := combineWords(bc.bitmap, other.bitmap, tc.op)
			expected := wordsToValues(combined)
			got := make([]uint16, len(expected))
			tc.fn(got, bc.bitmap, other.bitmap)
			for i := range expected {
				if got[i] != expected[i] {
					t.Fatalf("fillArray%s card=%d at %d: got %d want %d",
						tc.name, cardinality, i, got[i], expected[i])
				}
			}
		}
	}
}

func benchFillArray(b *testing.B, cardinality int) {
	bc := fillArrayTestBitmap(cardinality, int64(cardinality)+1)
	out := make([]uint16, cardinality)
	for b.Loop() {
		bc.fillArray(out)
	}
	if cardinality > 0 {
		b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*cardinality), "ns/value")
	}
}

func BenchmarkFillArray(b *testing.B) {
	for _, c := range []int{64, 256, 512, 1024, 2048, 4096} {
		b.Run("card="+itoaFill(c), func(b *testing.B) { benchFillArray(b, c) })
	}
}

func BenchmarkFillArrayAND(b *testing.B) {
	// source containers sized so the AND lands at or below arrayDefaultMaxSize,
	// the only regime in which fillArrayAND is reached
	for _, src := range []int{4096, 8192, 12288, 16384} {
		a := fillArrayTestBitmap(src, 11)
		c := fillArrayTestBitmap(src, 22)
		n := 0
		for i := range a.bitmap {
			n += bits.OnesCount64(a.bitmap[i] & c.bitmap[i])
		}
		out := make([]uint16, n)
		b.Run("src="+itoaFill(src)+"/out="+itoaFill(n), func(b *testing.B) {
			for b.Loop() {
				fillArrayAND(out, a.bitmap, c.bitmap)
			}
			b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*n), "ns/value")
		})
	}
}

func itoaFill(v int) string {
	if v == 0 {
		return "0"
	}
	var buf [12]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	return string(buf[i:])
}

// End to end: bitmapContainer.and / andNot / xor when the result is small
// enough to become an array container, which is the only way the two-input
// fillArray helpers are reached.
func benchBitmapToArrayOp(b *testing.B, srcCardinality int, op func(x, y *bitmapContainer) container) {
	x := fillArrayTestBitmap(srcCardinality, 11)
	y := fillArrayTestBitmap(srcCardinality, 22)
	for b.Loop() {
		c := op(x, y)
		if c == nil {
			b.Fatal("nil container")
		}
	}
}

func BenchmarkBitmapContainerToArrayOps(b *testing.B) {
	for _, src := range []int{4096, 8192, 16384} {
		b.Run("and/src="+itoaFill(src), func(b *testing.B) {
			benchBitmapToArrayOp(b, src, func(x, y *bitmapContainer) container { return x.andBitmap(y) })
		})
		b.Run("andNot/src="+itoaFill(src), func(b *testing.B) {
			benchBitmapToArrayOp(b, src, func(x, y *bitmapContainer) container { return x.andNotBitmap(y) })
		})
	}
}
