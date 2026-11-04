package roaring

import (
	"fmt"
	"testing"

	"github.com/RoaringBitmap/roaring/v2"
	"github.com/stretchr/testify/require"
)

func TestBatchEqualScatteredDifferential(t *testing.T) {
	for _, shape := range []string{"AgeRandom", "AgeNearCube", "Narrow14", "Wide200", "Wide1000", "WideStride2", "WideSharedPrefix", "Composite22", "CompositeWide"} {
		t.Run(shape, func(t *testing.T) {
			if testing.Short() && (shape == "AgeRandom" || shape == "AgeNearCube") {
				t.Skip("age fixture differential")
			}
			bsi, values := setupBatchEqualScattered(t, shape)
			checkBatchEqualScan(t, bsi, values)
		})
	}
}

func TestBatchEqualScatteredBoundaries(t *testing.T) {
	for _, sparse := range []bool{false, true} {
		t.Run(fmt.Sprint(sparse), func(t *testing.T) {
			bsi, values := setupBatchEqualScattered(t, "Wide200")
			// Matching keys must still compare the rest of a wide value.
			bsi.SetValue(300000, values[0]^(1<<60))
			if sparse {
				bsi.SetValue(uint64(^uint32(0)), values[0])
			}
			// GetValue ignores planes above bit 63. Stray bits, including ones
			// past every existing column, must not leak into the result.
			bsi.bA = append(bsi.bA, bsi.bA[0].Clone(), roaring.BitmapOf(^uint32(0)-1))
			bsi.bA[0].Add(^uint32(0) - 1)
			checkBatchEqualScan(t, bsi, values)
		})
	}

	// Sparse column containers must not force a full decode for each column.
	bsi := NewDefaultBSI()
	values := make([]int64, 128)
	for i := range values {
		values[i] = int64(2*i + 1)
	}
	for col := uint64(0); col < 100001; col++ {
		bsi.SetValue(col*65537, int64(col%257))
	}
	values[0] = 256 // Break the cube to exercise sparse direct checks.
	checkBatchEqualScan(t, bsi, values)

	// A cube may have fixed low bits and a fixed sign bit.
	values[0] = 1
	bsi = NewDefaultBSI()
	for i := range values {
		values[i] += Min64BitSigned
	}
	for col := uint64(0); col < 100001; col++ {
		bsi.SetValue(col, Min64BitSigned+int64(col%1024))
	}
	checkBatchEqualScan(t, bsi, values)
}

func checkBatchEqualScan(t *testing.T, bsi *BSI, values []int64) {
	t.Helper()
	values = append(append([]int64(nil), values...), values[0])
	if bsi.BitCount() < 64 {
		values = append(values, Min64BitSigned, Max64BitSigned)
	}
	want := make(map[int64]bool)
	for _, v := range values {
		want[v] = true
	}
	expected := roaring.NewBitmap()
	iter := bsi.eBM.Iterator()
	for iter.HasNext() {
		col := iter.Next()
		value, ok := bsi.GetValue(uint64(col))
		if ok && want[value] {
			expected.Add(col)
		}
	}
	workers := []int{-2, -1, 0, 1, 2, 4, 9, 1000000}
	if bsi.eBM.GetCardinality() > 1000000 {
		workers = []int{0, 2}
	}
	for _, optimize := range []bool{false, true} {
		if optimize {
			bsi.RunOptimize()
		}
		for _, n := range workers {
			actual := bsi.BatchEqual(n, values)
			require.True(t, actual.Equals(expected), "optimized=%t workers=%d", optimize, n)
			actual.Clear()
		}
		// Clearing the returned bitmap must not change subsequent queries.
		require.True(t, bsi.BatchEqual(0, values).Equals(expected))
	}
}
