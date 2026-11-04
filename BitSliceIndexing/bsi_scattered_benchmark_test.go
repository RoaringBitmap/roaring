package roaring

import (
	"math/rand"
	"testing"

	"github.com/RoaringBitmap/roaring/v2"
)

// These shapes distinguish complete bit-pattern sets from general scattered
// queries, and dense column IDs from holes in the column space.
func BenchmarkBatchEqualScattered(b *testing.B) {
	for _, shape := range []string{"AgeRandom", "AgeNearCube", "Narrow14", "Wide200", "Wide1000", "WideStride2", "WideSharedPrefix", "Composite22", "CompositeWide"} {
		b.Run(shape, func(b *testing.B) {
			bsi, values := setupBatchEqualScattered(b, shape)
			for _, workers := range []struct {
				name string
				n    int
			}{{"Workers2", 2}, {"DefaultWorkers", 0}} {
				b.Run(workers.name, func(b *testing.B) {
					b.ReportAllocs()
					var result *roaring.Bitmap
					for b.Loop() {
						result = bsi.BatchEqual(workers.n, values)
					}
					if result == nil {
						b.Fatal("BatchEqual returned nil")
					}
				})
			}
		})
	}
}

func setupBatchEqualScattered(t testing.TB, shape string) (*BSI, []int64) {
	t.Helper()
	rng := rand.New(rand.NewSource(42))
	if shape == "AgeRandom" || shape == "AgeNearCube" {
		bsi := setupLargeBSI(t)
		if bsi == nil {
			t.Fatal("age fixture unavailable")
		}
		values := make([]int64, 128)
		if shape == "AgeRandom" {
			for i, v := range rng.Perm(256)[:len(values)] {
				values[i] = int64(v)
			}
		} else {
			for i := range values {
				values[i] = int64(2*i + 1)
			}
			values[0] = 0
		}
		return bsi, values
	}
	const columns = 300000
	count := 200
	if shape == "Wide1000" || shape == "WideStride2" {
		count = 1000
	}
	values := make([]int64, count)
	bsi := NewDefaultBSI()
	for col := uint64(0); col < columns; col++ {
		value := int64(rng.Uint64())
		if shape == "Narrow14" {
			value &= 1<<14 - 1
		}
		if shape == "WideSharedPrefix" {
			value = int64(uint64(value)&(1<<48-1) | uint64(0x9234)<<48)
		}
		if shape == "Composite22" || shape == "CompositeWide" {
			shift := uint(16)
			if shape == "CompositeWide" {
				shift = 40
			}
			value = int64(col%64)<<shift | int64(uint64(value)&0xffff)
		}
		id := col
		if shape == "WideStride2" {
			id *= 2
		}
		bsi.SetValue(id, value)
		if col < uint64(len(values)) {
			values[col] = value
			if shape == "Composite22" || shape == "CompositeWide" {
				shift := uint(16)
				if shape == "CompositeWide" {
					shift = 40
				}
				values[col] = values[col]&0xffff | int64(17)<<shift
			}
		}
	}
	return bsi, values
}
