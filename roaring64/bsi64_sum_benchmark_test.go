package roaring64

import (
	"fmt"
	"math/big"
	"math/rand"
	"testing"
)

func BenchmarkBSI64SumBigValues(b *testing.B) {
	for _, test := range []struct {
		rows, width, selected int
		scattered             bool
	}{
		{128, 63, 0, false},
		{100_000, 16, 0, false},
		{100_000, 63, 0, false},
		{100_000, 64, 0, false},
		{2_000_000, 16, 0, false},
		{2_000_000, 63, 0, false},
		{2_000_000, 64, 0, false},
		{2_000_000, 63, 100_000, false},
		{2_000_000, 63, 100_000, true},
		{2_000_000, 63, 262_143, true},
		{8_000_000, 63, 250_000, true},
	} {
		filter := "all"
		if test.selected != 0 {
			filter = fmt.Sprintf("range%d", test.selected)
			if test.scattered {
				filter = fmt.Sprintf("scattered%d", test.selected)
			}
		}
		b.Run(fmt.Sprintf("rows%d/bits%d/%s", test.rows, test.width, filter), func(b *testing.B) {
			maxValue, minValue := int64(Max64BitSigned), int64(0)
			if test.width == 16 {
				maxValue = 1<<16 - 1
			} else if test.width == 64 {
				minValue = Min64BitSigned
			}
			bsi := NewBSI(maxValue, minValue)
			rng := rand.New(rand.NewSource(552))
			for row := 0; row < test.rows; row++ {
				value := int64(rng.Uint64())
				if test.width < 64 {
					value &= maxValue
				}
				if test.rows == 128 {
					value = int64(row)
				}
				bsi.SetValue(uint64(row), value)
			}
			foundSet := bsi.GetExistenceBitmap()
			if test.selected != 0 {
				foundSet = NewBitmap()
				if test.scattered {
					for foundSet.GetCardinality() < uint64(test.selected) {
						foundSet.Add(uint64(rng.Intn(test.rows)))
					}
				} else {
					foundSet.AddRange(0, uint64(test.selected))
				}
			}
			wantCount := foundSet.GetCardinality()
			var sum *big.Int
			var count uint64
			b.ReportAllocs()
			for b.Loop() {
				sum, count = bsi.SumBigValues(foundSet)
			}
			if sum == nil || count != wantCount {
				b.Fatalf("SumBigValues() = (%v, %d), want count %d", sum, count, wantCount)
			}
		})
	}
}
