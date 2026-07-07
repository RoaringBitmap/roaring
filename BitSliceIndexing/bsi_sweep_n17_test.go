package roaring

import "testing"

func BenchmarkBatchEqualScatteredN17Sweep(b *testing.B) {
	bsi := setupLargeBSI(b)
	if bsi == nil {
		b.Skip("skipping, large BSI setup failed")
		return
	}
	vals := make([]int64, 17)
	for i := range vals {
		vals[i] = int64(i) * 5
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bsi.BatchEqual(0, vals)
	}
}
