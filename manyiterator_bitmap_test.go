package roaring

import (
	"math/bits"
	"math/rand"
	"testing"
)

// bitmapContainerManyIterator decodes whole words in bulk when the buffer has
// room and falls back to the word-at-a-time loop otherwise. Buffer sizes that
// are not multiples of 64 make the two paths interleave.
func TestBitmapContainerNextManyBufferSizes(t *testing.T) {
	r := rand.New(rand.NewSource(11))
	for _, cardinality := range []int{1, 63, 64, 65, 4096, 30000, 65535, 65536} {
		bc := newBitmapContainer()
		for bc.cardinality < cardinality {
			bc.iadd(uint16(r.Intn(65536)))
		}

		want := make([]uint32, 0, cardinality)
		for k, w := range bc.bitmap {
			for w != 0 {
				want = append(want, 0x50000|uint32(k*64+bits.TrailingZeros64(w)))
				w &= w - 1
			}
		}

		for _, bufSize := range []int{1, 2, 63, 64, 65, 100, 127, 128, 129, 1000, 4096, 65536} {
			it := bc.getManyIterator()
			buf := make([]uint32, bufSize)
			got := make([]uint32, 0, cardinality)
			for {
				n := it.nextMany(0x50000, buf)
				if n == 0 {
					break
				}
				if n > bufSize {
					t.Fatalf("card=%d buf=%d: returned %d, more than the buffer holds", cardinality, bufSize, n)
				}
				got = append(got, buf[:n]...)
			}
			if len(got) != len(want) {
				t.Fatalf("card=%d buf=%d: got %d values, want %d", cardinality, bufSize, len(got), len(want))
			}
			for i := range want {
				if got[i] != want[i] {
					t.Fatalf("card=%d buf=%d: at %d got %d want %d", cardinality, bufSize, i, got[i], want[i])
				}
			}
		}
	}
}

func benchBitmapNextMany(b *testing.B, cardinality, bufSize int) {
	r := rand.New(rand.NewSource(12))
	bc := newBitmapContainer()
	for bc.cardinality < cardinality {
		bc.iadd(uint16(r.Intn(65536)))
	}
	buf := make([]uint32, bufSize)
	var sink int
	for b.Loop() {
		it := bc.getManyIterator()
		for n := it.nextMany(0x50000, buf); n != 0; n = it.nextMany(0x50000, buf) {
			sink += n
		}
	}
	_ = sink
	b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*cardinality), "ns/value")
}

func BenchmarkBitmapContainerNextMany(b *testing.B) {
	b.Run("card=4096/buf=4096", func(b *testing.B) { benchBitmapNextMany(b, 4096, 4096) })
	b.Run("card=16384/buf=4096", func(b *testing.B) { benchBitmapNextMany(b, 16384, 4096) })
	b.Run("card=32768/buf=4096", func(b *testing.B) { benchBitmapNextMany(b, 32768, 4096) })
	b.Run("card=65536/buf=4096", func(b *testing.B) { benchBitmapNextMany(b, 65536, 4096) })
	b.Run("card=32768/buf=512", func(b *testing.B) { benchBitmapNextMany(b, 32768, 512) })
	b.Run("card=32768/buf=64", func(b *testing.B) { benchBitmapNextMany(b, 32768, 64) })
}
