package roaring64

import (
	"math/big"
	"math/rand"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBSI64BatchEqualValuesConsistentWithBatchEqual(t *testing.T) {
	rg := rand.New(rand.NewSource(909))
	for run := 0; run < 25; run++ {
		bsi := NewDefaultBSI()
		numCols := rg.Intn(1000) + 50
		for col := 0; col < numCols; col++ {
			if rg.Float64() < 0.90 {
				bsi.SetValue(uint64(col), rg.Int63n(400)-200)
			}
		}

		values := []int64{-200, -99, -5, 0, 7, 42, 42, 199}
		foundSet := NewBitmap()
		for col := 0; col < numCols; col++ {
			if col%3 != 0 {
				foundSet.Add(uint64(col))
			}
		}

		for _, fs := range []*Bitmap{nil, foundSet} {
			expected := expectedBSI64BatchEqualValues(bsi, values, fs)
			actual := bsi.BatchEqualValues(0, values, fs)
			assert.Equal(t, expected, sortedBSI64ValuePairs(actual), "run=%d foundSet=%v", run, fs != nil)
		}
	}
}

func TestBSI64BatchEqualValuesHandlesBigWidthFallback(t *testing.T) {
	bsi := NewDefaultBSI()
	huge := new(big.Int).Lsh(big.NewInt(1), 90)
	bsi.SetBigValue(1, huge)
	bsi.SetValue(2, -7)
	bsi.SetValue(3, 11)
	bsi.SetValue(4, -7)

	foundSet := BitmapOf(1, 2, 3)
	actual := sortedBSI64ValuePairs(bsi.BatchEqualValues(0, []int64{-7, 11}, foundSet))
	assert.Equal(t, []BSIValuePair{
		{ColumnID: 2, Value: -7},
		{ColumnID: 3, Value: 11},
	}, actual)
}

func BenchmarkBSI64BatchEqualValuesLargeFixture(b *testing.B) {
	bsi, values, foundSet := setupBSI64BatchEqualValuesFixture(b, 100000, 100, 27)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pairs := bsi.BatchEqualValues(0, values, foundSet)
		_ = pairs
	}
}

func BenchmarkBSI64BatchEqualGetBigValuesLargeFixture(b *testing.B) {
	bsi, values, foundSet := setupBSI64BatchEqualValuesFixture(b, 100000, 100, 27)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		matched := bsi.BatchEqual(0, values)
		matched.And(foundSet)
		columnIDs := matched.ToArray()
		bigValues := bsi.GetBigValues(columnIDs)
		pairs := make([]BSIValuePair, 0, len(columnIDs))
		for j, columnID := range columnIDs {
			if bigValues[j] != nil {
				pairs = append(pairs, BSIValuePair{ColumnID: columnID, Value: bigValues[j].Int64()})
			}
		}
		_ = pairs
	}
}

func BenchmarkBSI64BatchEqualGetValueLoopLargeFixture(b *testing.B) {
	bsi, values, foundSet := setupBSI64BatchEqualValuesFixture(b, 100000, 100, 27)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		matched := bsi.BatchEqual(0, values)
		matched.And(foundSet)
		pairs := make([]BSIValuePair, 0, int(matched.GetCardinality()))
		iter := matched.Iterator()
		for iter.HasNext() {
			columnID := iter.Next()
			value, ok := bsi.GetValue(columnID)
			if ok {
				pairs = append(pairs, BSIValuePair{ColumnID: columnID, Value: value})
			}
		}
		_ = pairs
	}
}

func expectedBSI64BatchEqualValues(bsi *BSI, values []int64, foundSet *Bitmap) []BSIValuePair {
	matched := bsi.BatchEqual(0, values)
	if foundSet != nil {
		matched.And(foundSet)
	}
	pairs := make([]BSIValuePair, 0, int(matched.GetCardinality()))
	iter := matched.Iterator()
	for iter.HasNext() {
		columnID := iter.Next()
		value, ok := bsi.GetValue(columnID)
		if ok {
			pairs = append(pairs, BSIValuePair{ColumnID: columnID, Value: value})
		}
	}
	return sortedBSI64ValuePairs(pairs)
}

func sortedBSI64ValuePairs(pairs []BSIValuePair) []BSIValuePair {
	sorted := append([]BSIValuePair(nil), pairs...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].ColumnID != sorted[j].ColumnID {
			return sorted[i].ColumnID < sorted[j].ColumnID
		}
		return sorted[i].Value < sorted[j].Value
	})
	return sorted
}

func setupBSI64BatchEqualValuesFixture(tb testing.TB, rows, valueDomain, valueCount int) (*BSI, []int64, *Bitmap) {
	tb.Helper()
	bsi := NewDefaultBSI()
	for row := 0; row < rows; row++ {
		value := int64(row%valueDomain) - int64(valueDomain/2)
		bsi.SetValue(uint64(row), value)
	}

	values := make([]int64, 0, valueCount)
	for i := 0; i < valueCount; i++ {
		values = append(values, int64((i*7)%valueDomain)-int64(valueDomain/2))
	}

	foundSet := NewBitmap()
	for row := 0; row < rows; row++ {
		if row%5 != 0 {
			foundSet.Add(uint64(row))
		}
	}
	return bsi, values, foundSet
}
