package roaring64

import "testing"

const benchmarkRoaring64BatchCardinality = 1 << 15

func benchmarkRoaring64BatchBitmap() *Bitmap {
	values := make([]uint64, benchmarkRoaring64BatchCardinality)
	for i := range values {
		values[i] = uint64(i * 2)
	}

	bitmap := New()
	bitmap.AddMany(values)
	return bitmap
}

func BenchmarkRoaring64ToArray(b *testing.B) {
	bitmap := benchmarkRoaring64BatchBitmap()
	last := uint64((benchmarkRoaring64BatchCardinality - 1) * 2)

	for b.Loop() {
		values := bitmap.ToArray()
		if len(values) != benchmarkRoaring64BatchCardinality || values[0] != 0 || values[len(values)-1] != last {
			b.Fatal("unexpected bitmap contents")
		}
	}
}

func BenchmarkRoaring64ManyIterator(b *testing.B) {
	bitmap := benchmarkRoaring64BatchBitmap()
	values := make([]uint64, benchmarkRoaring64BatchCardinality)
	last := uint64((benchmarkRoaring64BatchCardinality - 1) * 2)

	for b.Loop() {
		iterator := bitmap.ManyIterator()
		n := 0
		for n < len(values) {
			count := iterator.NextMany(values[n:])
			if count == 0 {
				break
			}
			n += count
		}
		if n != len(values) || values[0] != 0 || values[n-1] != last {
			b.Fatal("unexpected iterator contents")
		}
	}
}
