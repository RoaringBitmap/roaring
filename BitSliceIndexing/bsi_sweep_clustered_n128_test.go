package roaring

import "testing"

func BenchmarkBatchEqualClusteredLargeN128(b *testing.B) {
	bsi := setupLargeBSI(b)
	if bsi == nil {
		b.Skip("skipping, large BSI setup failed")
		return
	}
	vals := make([]int64, 0, 128)
	for cluster := 0; cluster < 16; cluster++ {
		start := int64(cluster) * 100
		for i := int64(0); i < 8; i++ {
			vals = append(vals, start+i)
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bsi.BatchEqual(0, vals)
	}
}
