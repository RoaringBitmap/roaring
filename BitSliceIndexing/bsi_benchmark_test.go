package roaring

import (
	"math/rand"
	"testing"

	"github.com/RoaringBitmap/roaring/v2"
	"github.com/stretchr/testify/assert"
)

func BenchmarkBatchEqual(b *testing.B) {
	bsi := setupLargeBSI(b)
	if bsi == nil {
		b.Skip("skipping, large BSI setup failed")
		return
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res := bsi.BatchEqual(0, []int64{55, 57})
		_ = res
	}
}

func TestBatchEqualEdgeCases(t *testing.T) {
	// 1. Empty or Nil inputs
	bsi := NewDefaultBSI()
	res := bsi.BatchEqual(0, nil)
	assert.True(t, res.IsEmpty())

	res = bsi.BatchEqual(0, []int64{})
	assert.True(t, res.IsEmpty())

	// 2. Set some values
	bsi.SetValue(10, 42)
	bsi.SetValue(20, 100)
	bsi.SetValue(30, 42)
	bsi.SetValue(40, -5)

	// Test matching positive values
	res = bsi.BatchEqual(0, []int64{42})
	assert.Equal(t, uint64(2), res.GetCardinality())
	assert.True(t, res.Contains(10))
	assert.True(t, res.Contains(30))

	// Test matching multiple values including non-existent and duplicates
	res = bsi.BatchEqual(0, []int64{42, 100, 42, 999})
	assert.Equal(t, uint64(3), res.GetCardinality())
	assert.True(t, res.Contains(10))
	assert.True(t, res.Contains(20))
	assert.True(t, res.Contains(30))

	// Test negative value
	res = bsi.BatchEqual(0, []int64{-5})
	assert.Equal(t, uint64(1), res.GetCardinality())
	assert.True(t, res.Contains(40))

	// Test 1<<62 edge case explicitly
	bsi62 := NewBSI(1<<62, 0)
	bsi62.SetValue(10, 5)
	res = bsi62.BatchEqual(0, []int64{5})
	assert.Equal(t, uint64(1), res.GetCardinality())
	assert.True(t, res.Contains(10))
}

func TestBatchEqualSub64Bit(t *testing.T) {
	// NewBSI(100,0) (BitCount()==7) queried with {-5} and with {200} (>= 2^7)
	// must both return an empty bitmap and must equal a GetValue-derived ground truth.
	bsi := NewBSI(100, 0)
	assert.Equal(t, 7, bsi.BitCount())

	// Set some values inside range
	bsi.SetValue(10, 42)
	bsi.SetValue(20, 99)

	// Ground truth function
	getGroundTruth := func(query []int64) *roaring.Bitmap {
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
		return expected
	}

	for _, q := range []int64{-5, 200} {
		res := bsi.BatchEqual(0, []int64{q})
		assert.True(t, res.IsEmpty())

		expected := getGroundTruth([]int64{q})
		assert.True(t, expected.IsEmpty())
		assert.True(t, res.Equals(expected))
	}
}

func TestBatchEqualResultIsolation(t *testing.T) {
	bsi := NewDefaultBSI()
	bsi.SetValue(10, 42)
	bsi.SetValue(20, 100)

	// Get batch equal result
	res := bsi.BatchEqual(0, []int64{42})
	assert.True(t, res.Contains(10))

	// Mutate the returned bitmap
	res.Add(999)
	res.Remove(10)

	// Assert that the source BSI's internal state (existence bitmap and bit planes) is completely unaffected
	assert.False(t, bsi.GetExistenceBitmap().Contains(999))
	assert.True(t, bsi.GetExistenceBitmap().Contains(10))

	val, ok := bsi.GetValue(10)
	assert.True(t, ok)
	assert.Equal(t, int64(42), val)

	val, ok = bsi.GetValue(999)
	assert.False(t, ok)
}

func TestBatchEqualConsistentWithGetValue(t *testing.T) {
	rg := rand.New(rand.NewSource(42))
	for run := 0; run < 15; run++ {
		// Create a randomized BSI
		bsi := NewDefaultBSI()
		numCols := rg.Intn(1000) + 10
		for col := 0; col < numCols; col++ {
			if rg.Float64() < 0.8 {
				val := rg.Int63n(500) - 250 // Mix of positive, zero, and negative values
				bsi.SetValue(uint64(col), val)
			}
		}

		// Generate query values (small, medium, and large list sizes to test the hybrid threshold)
		querySizes := []int{rg.Intn(10) + 1, rg.Intn(50) + 50, rg.Intn(200) + 100}
		for _, querySize := range querySizes {
			query := make([]int64, querySize)
			for i := range query {
				query[i] = rg.Int63n(600) - 300
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
					t.Fatalf("Mismatch in run %d querySize %d parallelism %d. Query: %v. Expected: %v, Got: %v", run, querySize, parallelism, query, expected.ToArray(), actual.ToArray())
				}
			}
		}
	}
}

// TestBatchEqualExistenceAuthority pins BatchEqual results to the existence
// bitmap. UnmarshalBinary accepts plane data that is not a subset of eBM (the
// checked-in testdata/age fixture is such data), and every read path treats
// eBM as authoritative, so columns present in a plane but absent from eBM must
// never appear in results.
func TestBatchEqualExistenceAuthority(t *testing.T) {
	// Synthetic state: column 2 has bits in plane 0 but is absent from eBM.
	ebm := roaring.BitmapOf(1)
	plane := roaring.BitmapOf(1, 2)
	ebmData, err := ebm.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	planeData, err := plane.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	bsi := NewDefaultBSI()
	if err := bsi.UnmarshalBinary([][]byte{ebmData, planeData}); err != nil {
		t.Fatal(err)
	}
	res := bsi.BatchEqual(0, []int64{1})
	assert.True(t, res.Contains(1))
	assert.False(t, res.Contains(2), "column 2 is not in eBM and must not match")

	// The age fixture ships with plane cardinalities above the eBM cardinality;
	// results must still be a subset of eBM.
	large := setupLargeBSI(t)
	if large == nil {
		t.Skip("skipping, large BSI setup failed")
	}
	for _, vals := range [][]int64{{16}, {55, 57}, {0, 1, 2, 3}} {
		res := large.BatchEqual(0, vals)
		outside := roaring.AndNot(res, large.GetExistenceBitmap())
		assert.True(t, outside.IsEmpty(), "BatchEqual(%v) returned %d columns outside eBM", vals, outside.GetCardinality())
	}
}

// Benchmarks across query-list shapes: work sharing behaves differently for
// small lists, dense contiguous ranges (which collapse to a few plane
// operations), and scattered values (no range collapse).
func BenchmarkBatchEqualM128(b *testing.B)          { benchmarkBatchEqualM(b, 128, 1) }
func BenchmarkBatchEqualM128Scattered(b *testing.B) { benchmarkBatchEqualM(b, 128, 2) }
func BenchmarkBatchEqualM200(b *testing.B)          { benchmarkBatchEqualM(b, 200, 1) }

func benchmarkBatchEqualM(b *testing.B, m int, stride int64) {
	bsi := setupLargeBSI(b)
	if bsi == nil {
		b.Skip("skipping, large BSI setup failed")
	}
	vals := make([]int64, m)
	for i := range vals {
		vals[i] = int64(i)*stride + stride - 1
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res := bsi.BatchEqual(0, vals)
		_ = res
	}
}
