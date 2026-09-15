package roaring

import (
	"fmt"
	"testing"
)

var benchmarkMinMaxResult int64

// BenchmarkMinMax covers both widths a BSI can take: a narrow one holding
// only non-negative values, and the 64 bit wide one that a negative
// declared minimum produces.
func BenchmarkMinMax(b *testing.B) {
	for _, shape := range []struct {
		name               string
		maxValue, minValue int64
	}{
		{"non-negative", 99, 0},
		{"signed", 99, -1},
	} {
		for _, rows := range []int{100, 10000} {
			bsi := NewBSI(shape.maxValue, shape.minValue)
			for row := 0; row < rows; row++ {
				value := int64(row % 99)
				if shape.minValue < 0 {
					value -= 50
				}
				bsi.SetValue(uint64(row), value)
			}
			b.Run(fmt.Sprintf("%s/planes%d/rows%d", shape.name, bsi.BitCount(), rows), func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					benchmarkMinMaxResult = bsi.MinMax(0, MAX, nil)
				}
			})
		}
	}
}
