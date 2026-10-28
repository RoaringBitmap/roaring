package roaring

import "testing"

var unionRealDataBitmap *Bitmap
var unionRealDataCardinality uint64

// BENCH_REAL_DATA=1 go test -run '^$' -bench '^BenchmarkUnionRealData$' -benchmem
// GOPATH must contain src/github.com/RoaringBitmap/real-roaring-datasets.
func BenchmarkUnionRealData(b *testing.B) {
	if !benchRealData {
		b.Skip("set BENCH_REAL_DATA=1 and GOPATH to enable the real-data benchmarks")
	}
	for _, dataset := range realDatasets {
		for _, optimize := range []bool{false, true} {
			mode := "raw"
			if optimize {
				mode = "optimized"
			}
			b.Run(dataset+"/"+mode, func(b *testing.B) {
				bitmaps, err := retrieveRealDataBitmaps(dataset, optimize)
				if err != nil {
					b.Fatal(err)
				}
				if len(bitmaps) < 2 {
					b.Fatal("dataset needs at least two bitmaps")
				}
				wantAll := NewBitmap()
				for _, bitmap := range bitmaps {
					iterator := bitmap.Iterator()
					for iterator.HasNext() {
						wantAll.Add(iterator.Next())
					}
				}
				var wantPairs uint64
				for i := 0; i+1 < len(bitmaps); i++ {
					want := bitmaps[i].Clone()
					iterator := bitmaps[i+1].Iterator()
					for iterator.HasNext() {
						want.Add(iterator.Next())
					}
					if !Or(bitmaps[i], bitmaps[i+1]).Equals(want) {
						b.Fatalf("incorrect pairwise union at pair %d", i)
					}
					wantPairs += want.GetCardinality()
				}
				b.Run("pairwise", func(b *testing.B) {
					b.ReportAllocs()
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						var cardinality uint64
						for j := 0; j+1 < len(bitmaps); j++ {
							unionRealDataBitmap = Or(bitmaps[j], bitmaps[j+1])
							cardinality += unionRealDataBitmap.GetCardinality()
						}
						unionRealDataCardinality = cardinality
					}
					b.StopTimer()
					if unionRealDataCardinality != wantPairs {
						b.Fatal("pairwise union cardinality changed during benchmark")
					}
					b.ReportMetric(float64(len(bitmaps)-1), "unions/op")
				})
				b.Run("accumulated", func(b *testing.B) {
					b.ReportAllocs()
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						result := NewBitmap()
						for _, bitmap := range bitmaps {
							result.Or(bitmap)
						}
						unionRealDataBitmap = result
						unionRealDataCardinality = result.GetCardinality()
					}
					b.StopTimer()
					if !unionRealDataBitmap.Equals(wantAll) {
						b.Fatal("incorrect accumulated union")
					}
				})
				b.Run("fastor", func(b *testing.B) {
					b.ReportAllocs()
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						unionRealDataBitmap = FastOr(bitmaps...)
						unionRealDataCardinality = unionRealDataBitmap.GetCardinality()
					}
					b.StopTimer()
					if !unionRealDataBitmap.Equals(wantAll) {
						b.Fatal("incorrect FastOr union")
					}
				})
			})
		}
	}
}
