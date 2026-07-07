package roaring

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/RoaringBitmap/roaring/v2"
	"github.com/stretchr/testify/assert"
)

func TestBatchEqualLargeQueryValues(t *testing.T) {
	rg := rand.New(rand.NewSource(12345))
	for run := 0; run < 10; run++ {
		// Create a randomized BSI with large values >= 1,048,576
		bsi := NewDefaultBSI()
		numCols := rg.Intn(50000) + 120000
		for col := 0; col < numCols; col++ {
			if rg.Float64() < 0.8 {
				// Generate some large positive values around 1,048,576
				val := rg.Int63n(100000) + 1048500
				bsi.SetValue(uint64(col), val)
			}
		}

		// Generate query values containing values >= 1,048,576
		querySize := rg.Intn(100) + 128
		query := make([]int64, querySize)
		for i := range query {
			query[i] = rg.Int63n(100100) + 1048500
		}

		// Ground truth
		expected := roaring.NewBitmap()
		valMap := make(map[int64]bool)
		for _, q := range query {
			valMap[q] = true
		}
		iter := bsi.GetExistenceBitmap().Iterator()
		for iter.HasNext() {
			col := iter.Next()
			val, ok := bsi.GetValue(uint64(col))
			if ok && valMap[val] {
				expected.Add(col)
			}
		}

		// Test different parallelism settings
		for _, parallelism := range []int{0, 1, 2, 4} {
			actual := bsi.BatchEqual(parallelism, query)
			if !actual.Equals(expected) {
				t.Fatalf("Mismatch with large query values in run %d parallelism %d. Expected: %v, Got: %v", run, parallelism, expected.ToArray(), actual.ToArray())
			}
		}
	}
}

func TestBatchEqualParallelScanCheckedInFixture(t *testing.T) {
	large := setupLargeBSI(t)
	if large == nil {
		t.Skip("skipping, large BSI setup failed")
	}

	// Generate a query that triggers the parallel scan path (e.g. 130 scattered values)
	rg := rand.New(rand.NewSource(12345))
	vals := make([]int64, 130)
	for i := range vals {
		vals[i] = rg.Int63n(100)
	}

	// Result from the fallback path (either automatically triggered or explicitly run)
	resAuto := large.BatchEqual(0, vals)

	// Since len(vals) >= 128 and estimateBranchCount >= 64,
	// BatchEqual(0, vals) will run the parallel scan path.
	// Let's verify that the results are a subset of eBM and perfectly match the ground truth.
	outside := roaring.AndNot(resAuto, large.GetExistenceBitmap())
	assert.True(t, outside.IsEmpty(), "parallel scan returned columns outside eBM")

	// Let's also verify consistency with GetValue ground truth
	expected := roaring.NewBitmap()
	valMap := make(map[int64]bool)
	for _, q := range vals {
		valMap[q] = true
	}
	iter := large.GetExistenceBitmap().Iterator()
	for iter.HasNext() {
		col := iter.Next()
		val, ok := large.GetValue(uint64(col))
		if ok && valMap[val] {
			expected.Add(col)
		}
	}

	assert.True(t, resAuto.Equals(expected), "Parallel scan results do not match ground truth on checked-in fixture")
}

func BenchmarkBatchEqualM128ScatteredLargeValues(b *testing.B) {
	rg := rand.New(rand.NewSource(12345))
	bsi := NewDefaultBSI()
	numCols := 50000
	for col := 0; col < numCols; col++ {
		if rg.Float64() < 0.8 {
			val := rg.Int63n(100000) + 1048500
			bsi.SetValue(uint64(col), val)
		}
	}

	vals := make([]int64, 128)
	for i := range vals {
		vals[i] = rg.Int63n(100000) + 1048500
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res := bsi.BatchEqual(0, vals)
		_ = res
	}
}

func BenchmarkBatchEqualScatteredLowCardinality(b *testing.B) {
	rg := rand.New(rand.NewSource(12345))
	bsi := NewDefaultBSI()
	numCols := 500 // low cardinality
	for col := 0; col < numCols; col++ {
		if rg.Float64() < 0.8 {
			val := rg.Int63n(100)
			bsi.SetValue(uint64(col), val)
		}
	}
	vals := make([]int64, 128)
	for i := range vals {
		vals[i] = rg.Int63n(100)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bsi.BatchEqual(0, vals)
	}
}

func BenchmarkBatchEqualScatteredFewPlanes(b *testing.B) {
	rg := rand.New(rand.NewSource(12345))
	bsi := NewBSI(3, 0) // only 2 bitplanes
	numCols := 50000
	for col := 0; col < numCols; col++ {
		if rg.Float64() < 0.8 {
			val := rg.Int63n(4)
			bsi.SetValue(uint64(col), val)
		}
	}
	vals := make([]int64, 128)
	for i := range vals {
		vals[i] = rg.Int63n(4)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bsi.BatchEqual(0, vals)
	}
}

func BenchmarkBatchEqualSweepBranchCount(b *testing.B) {
	bsi := setupLargeBSI(b)
	if bsi == nil {
		b.Skip("skipping, large BSI setup failed")
		return
	}

	for _, count := range []int{2, 4, 8, 12, 14, 15, 16, 18, 20, 24, 32} {
		b.Run(fmt.Sprintf("BranchCount_%d", count), func(b *testing.B) {
			vals := make([]int64, count)
			for i := range vals {
				vals[i] = int64(i) * 5
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				res := bsi.BatchEqual(0, vals)
				_ = res
			}
		})
	}
}

func BenchmarkBatchEqualClustered(b *testing.B) {
	rg := rand.New(rand.NewSource(12345))
	bsi := setupLargeBSI(b)
	if bsi == nil {
		b.Skip("skipping, large BSI setup failed")
		return
	}
	// Generate query values that form clustered ranges: e.g. 8 clusters of 4 contiguous values each
	vals := make([]int64, 0, 32)
	for cluster := 0; cluster < 8; cluster++ {
		start := rg.Int63n(100) * 10
		for i := int64(0); i < 4; i++ {
			vals = append(vals, start+i)
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bsi.BatchEqual(0, vals)
	}
}
