package roaring

import "testing"

func bitmapOrBulkMergeTailAdjacentFixture() bitmapOrBulkMergeFixture {
	const containers = 4096

	leftKeys := make([]uint16, 0, containers-1)
	for key := 0; key < containers-2; key++ {
		leftKeys = append(leftKeys, uint16(key))
	}
	leftKeys = append(leftKeys, containers-1)

	return newBitmapOrBulkMergeFixture(leftKeys, []uint16{containers - 2}, false)
}

func BenchmarkBitmapOrBulkMergeTailAdjacent(b *testing.B) {
	b.Run("fresh-single-interior-4096", func(b *testing.B) {
		b.Run("fresh-tail-adjacent-4096", func(b *testing.B) {
			fixture := bitmapOrBulkMergeTailAdjacentFixture()

			b.ReportAllocs()
			var cardinality uint64
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				fixtureIndex := i & 1
				receiver := fixture.lefts[fixtureIndex].Clone()
				receiver.Or(fixture.rights[fixtureIndex])
				cardinality += receiver.GetCardinality()
			}
			b.StopTimer()
			if cardinality != fixture.cardinality*uint64(b.N) {
				b.Fatalf("unexpected total cardinality: got %d, want %d", cardinality, fixture.cardinality*uint64(b.N))
			}
		})
	})
}
