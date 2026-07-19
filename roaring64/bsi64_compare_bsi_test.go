package roaring64

import (
	"math/big"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func expectedBSI64CompareBSI(left *BSI, op Operation, right *BSI, foundSet *Bitmap) *Bitmap {
	expected := NewBitmap()
	source := And(left.GetExistenceBitmap(), right.GetExistenceBitmap())
	if foundSet != nil {
		source.And(foundSet)
	}
	iter := source.Iterator()
	for iter.HasNext() {
		col := iter.Next()
		leftValue, leftOK := left.GetBigValue(col)
		rightValue, rightOK := right.GetBigValue(col)
		if !leftOK || !rightOK {
			continue
		}
		compare := leftValue.Cmp(rightValue)
		switch op {
		case LT:
			if compare < 0 {
				expected.Add(col)
			}
		case LE:
			if compare <= 0 {
				expected.Add(col)
			}
		case EQ:
			if compare == 0 {
				expected.Add(col)
			}
		case GE:
			if compare >= 0 {
				expected.Add(col)
			}
		case GT:
			if compare > 0 {
				expected.Add(col)
			}
		default:
			panic("unsupported test operation")
		}
	}
	return expected
}

func TestBSI64CompareBSIConsistentWithGetBigValue(t *testing.T) {
	rg := rand.New(rand.NewSource(122))
	for run := 0; run < 25; run++ {
		left := NewDefaultBSI()
		right := NewDefaultBSI()
		numCols := rg.Intn(1000) + 50
		for col := 0; col < numCols; col++ {
			if rg.Float64() < 0.90 {
				left.SetValue(uint64(col), rg.Int63n(2000)-1000)
			}
			if rg.Float64() < 0.85 {
				right.SetValue(uint64(col), rg.Int63n(2000)-1000)
			}
		}
		// Force different bit widths and signs across the two BSIs.
		left.SetValue(uint64(numCols+1), 1<<40)
		right.SetValue(uint64(numCols+1), -1)
		left.SetValue(uint64(numCols+2), -1)
		right.SetValue(uint64(numCols+2), 1<<35)

		foundSet := NewBitmap()
		source := And(left.GetExistenceBitmap(), right.GetExistenceBitmap())
		iter := source.Iterator()
		for iter.HasNext() {
			col := iter.Next()
			if col%3 != 0 {
				foundSet.Add(col)
			}
		}

		for _, op := range []Operation{LT, LE, EQ, GE, GT} {
			for _, fs := range []*Bitmap{nil, foundSet} {
				expected := expectedBSI64CompareBSI(left, op, right, fs)
				actual := left.CompareBSI(op, right, fs)
				assert.True(t, actual.Equals(expected), "run=%d op=%d foundSet=%v expected=%v actual=%v",
					run, op, fs != nil, expected.ToArray(), actual.ToArray())
			}
		}
	}
}

func TestBSI64CompareBSIExistenceAndResultIsolation(t *testing.T) {
	left := NewDefaultBSI()
	right := NewDefaultBSI()
	left.SetValue(1, 10)
	left.SetValue(2, 20)
	right.SetValue(2, 15)
	right.SetValue(3, 5)

	actual := left.CompareBSI(GT, right, nil)
	assert.True(t, actual.Equals(BitmapOf(2)))

	actual.Add(99)
	actual.Remove(2)
	assert.True(t, left.GetExistenceBitmap().Contains(2))
	assert.True(t, right.GetExistenceBitmap().Contains(2))
	assert.False(t, left.GetExistenceBitmap().Contains(99))
}

func TestBSI64CompareBSIBigWidthConsistentWithGetBigValue(t *testing.T) {
	left := NewDefaultBSI()
	right := NewDefaultBSI()
	huge := new(big.Int).Lsh(big.NewInt(1), 90)
	hugePlusOne := new(big.Int).Add(huge, big.NewInt(1))
	negativeHuge := new(big.Int).Neg(huge)
	negativeHugeMinusOne := new(big.Int).Sub(negativeHuge, big.NewInt(1))

	left.SetBigValue(1, huge)
	right.SetBigValue(1, hugePlusOne)
	left.SetBigValue(2, hugePlusOne)
	right.SetBigValue(2, huge)
	left.SetBigValue(3, negativeHuge)
	right.SetBigValue(3, huge)
	left.SetBigValue(4, negativeHugeMinusOne)
	right.SetBigValue(4, negativeHuge)
	left.SetBigValue(5, negativeHuge)
	right.SetBigValue(5, negativeHuge)

	assert.True(t, left.CompareBSI(LT, right, nil).Equals(BitmapOf(1, 3, 4)))
	assert.True(t, left.CompareBSI(GT, right, nil).Equals(BitmapOf(2)))
	assert.True(t, left.CompareBSI(EQ, right, nil).Equals(BitmapOf(5)))
	assert.True(t, left.CompareBSI(LE, right, nil).Equals(BitmapOf(1, 3, 4, 5)))
	assert.True(t, left.CompareBSI(GE, right, nil).Equals(BitmapOf(2, 5)))
}

func BenchmarkBSI64CompareBSISameRowBitwise(b *testing.B) {
	left, right := setupBSI64CompareBSIFixture(b, 100000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res := left.CompareBSI(GT, right, nil)
		_ = res
	}
}

func BenchmarkBSI64CompareBSISameRowGetBigValue(b *testing.B) {
	left, right := setupBSI64CompareBSIFixture(b, 100000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res := expectedBSI64CompareBSI(left, GT, right, nil)
		_ = res
	}
}

func setupBSI64CompareBSIFixture(tb testing.TB, rows int) (*BSI, *BSI) {
	tb.Helper()
	left := NewDefaultBSI()
	right := NewDefaultBSI()
	for row := 0; row < rows; row++ {
		left.SetValue(uint64(row), int64(row%1000)-500)
		right.SetValue(uint64(row), int64((row*7)%1000)-500)
	}
	return left, right
}
