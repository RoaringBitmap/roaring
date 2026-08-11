package roaring64

import (
	"math/big"
	"testing"
)

var benchmarkMinMaxBigResult *big.Int

func BenchmarkBSI64MinMaxBig(b *testing.B) {
	const rows = 1_000_000
	bsi := NewDefaultBSI()
	foundSet := NewBitmap()
	for row := uint64(0); row < rows; row++ {
		bsi.SetValue(row, int64(row%50))
		foundSet.Add(row)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchmarkMinMaxBigResult = bsi.MinMaxBig(0, MIN, foundSet)
		benchmarkMinMaxBigResult = bsi.MinMaxBig(0, MAX, foundSet)
	}
}
