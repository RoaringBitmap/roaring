package roaring64

import (
	"math/big"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBSI64GetBigValuesConsistentWithGetBigValue(t *testing.T) {
	rg := rand.New(rand.NewSource(864))
	for run := 0; run < 25; run++ {
		bsi := NewDefaultBSI()
		numCols := rg.Intn(1000) + 50
		for col := 0; col < numCols; col++ {
			if rg.Float64() < 0.85 {
				bsi.SetValue(uint64(col), rg.Int63n(4000)-2000)
			}
		}

		columnIDs := make([]uint64, 0, numCols+8)
		for col := numCols - 1; col >= 0; col-- {
			if col%3 != 0 {
				columnIDs = append(columnIDs, uint64(col))
			}
		}
		columnIDs = append(columnIDs, uint64(numCols+10), 7, 7, 11)

		actual := bsi.GetBigValues(columnIDs)
		if len(actual) != len(columnIDs) {
			t.Fatalf("run=%d values length = %d, want %d", run, len(actual), len(columnIDs))
		}
		for i, columnID := range columnIDs {
			expectedValue, expectedOK := bsi.GetBigValue(columnID)
			actualValue := actual[i]
			if !expectedOK {
				assert.Nil(t, actualValue, "run=%d column=%d", run, columnID)
				continue
			}
			if assert.NotNil(t, actualValue, "run=%d column=%d", run, columnID) {
				assert.Equal(t, 0, actualValue.Cmp(expectedValue), "run=%d column=%d", run, columnID)
			}
		}
	}
}

func TestBSI64GetBigValuesHandlesBigWidthAndDuplicates(t *testing.T) {
	bsi := NewDefaultBSI()
	huge := new(big.Int).Lsh(big.NewInt(1), 90)
	hugePlusSeven := new(big.Int).Add(huge, big.NewInt(7))
	negativeHuge := new(big.Int).Neg(hugePlusSeven)

	bsi.SetBigValue(1, hugePlusSeven)
	bsi.SetBigValue(2, negativeHuge)
	bsi.SetValue(4, 0)

	values := bsi.GetBigValues([]uint64{2, 3, 1, 2, 4})
	assert.Equal(t, 5, len(values))
	assert.Equal(t, 0, values[0].Cmp(negativeHuge))
	assert.Nil(t, values[1])
	assert.Equal(t, 0, values[2].Cmp(hugePlusSeven))
	assert.Equal(t, 0, values[3].Cmp(negativeHuge))
	assert.Equal(t, 0, values[4].Cmp(big.NewInt(0)))

	values[0].SetInt64(12)
	assert.Equal(t, 0, values[3].Cmp(negativeHuge), "duplicate result values should be independent")
	stored, ok := bsi.GetBigValue(2)
	assert.True(t, ok)
	assert.Equal(t, 0, stored.Cmp(negativeHuge), "mutating returned values must not alter the BSI")
}

func BenchmarkBSI64GetBigValuesLargeFixture(b *testing.B) {
	bsi, _ := setupBSI64CompareBSIFixture(b, 100000)
	columnIDs := bsi64SequentialColumns(100000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		values := bsi.GetBigValues(columnIDs)
		_ = values
	}
}

func BenchmarkBSI64GetBigValueLoopLargeFixture(b *testing.B) {
	bsi, _ := setupBSI64CompareBSIFixture(b, 100000)
	columnIDs := bsi64SequentialColumns(100000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		values := make([]*big.Int, len(columnIDs))
		for j, columnID := range columnIDs {
			value, ok := bsi.GetBigValue(columnID)
			if ok {
				values[j] = value
			}
		}
		_ = values
	}
}

func bsi64SequentialColumns(n int) []uint64 {
	columnIDs := make([]uint64, n)
	for i := range columnIDs {
		columnIDs[i] = uint64(i)
	}
	return columnIDs
}
