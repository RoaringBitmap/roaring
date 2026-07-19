package roaring64

import (
	"math/big"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func expectedBSI64CompareValue(bsi *BSI, op Operation, valueOrStart, end int64, foundSet *Bitmap) *Bitmap {
	expected := NewBitmap()
	source := bsi.GetExistenceBitmap()
	if foundSet != nil {
		source = And(source, foundSet)
	}
	iter := source.Iterator()
	for iter.HasNext() {
		col := iter.Next()
		val, ok := bsi.GetValue(col)
		if !ok {
			continue
		}
		switch op {
		case LT:
			if val < valueOrStart {
				expected.Add(col)
			}
		case LE:
			if val <= valueOrStart {
				expected.Add(col)
			}
		case EQ:
			if val == valueOrStart {
				expected.Add(col)
			}
		case GE:
			if val >= valueOrStart {
				expected.Add(col)
			}
		case GT:
			if val > valueOrStart {
				expected.Add(col)
			}
		case RANGE:
			if val >= valueOrStart && val <= end {
				expected.Add(col)
			}
		default:
			panic("unsupported test operation")
		}
	}
	return expected
}

func TestBSI64CompareValueConsistentWithGetValue(t *testing.T) {
	rg := rand.New(rand.NewSource(84))
	for run := 0; run < 15; run++ {
		bsi := NewDefaultBSI()
		numCols := rg.Intn(1000) + 10
		for col := 0; col < numCols; col++ {
			if rg.Float64() < 0.8 {
				bsi.SetValue(uint64(col), rg.Int63n(500)-250)
			}
		}

		foundSet := NewBitmap()
		iter := bsi.GetExistenceBitmap().Iterator()
		for iter.HasNext() {
			col := iter.Next()
			if col%3 != 0 {
				foundSet.Add(col)
			}
		}

		cases := []struct {
			op    Operation
			start int64
			end   int64
		}{
			{LT, -17, 0},
			{LE, -17, 0},
			{EQ, -17, 0},
			{GE, -17, 0},
			{GT, -17, 0},
			{RANGE, -25, 25},
		}
		for _, tc := range cases {
			for _, fs := range []*Bitmap{nil, foundSet} {
				expected := expectedBSI64CompareValue(bsi, tc.op, tc.start, tc.end, fs)
				actual := bsi.CompareValue(0, tc.op, tc.start, tc.end, fs)
				assert.True(t, actual.Equals(expected), "run=%d op=%d foundSet=%v expected=%v actual=%v",
					run, tc.op, fs != nil, expected.ToArray(), actual.ToArray())
			}
		}
	}
}

func BenchmarkBSI64CompareValueEQLargeAgeFixture(b *testing.B) {
	bsi := setupLargeBSI(b)
	if bsi == nil {
		b.Skip("skipping, large BSI setup failed")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res := bsi.CompareValue(0, EQ, 55, 0, nil)
		_ = res
	}
}

func BenchmarkBSI64CompareBigValueEQLargeAgeFixture(b *testing.B) {
	bsi := setupLargeBSI(b)
	if bsi == nil {
		b.Skip("skipping, large BSI setup failed")
	}
	value := big.NewInt(55)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res := bsi.CompareBigValue(0, EQ, value, nil, nil)
		_ = res
	}
}

func BenchmarkBSI64CompareValueEQFoundSetLargeAgeFixture(b *testing.B) {
	bsi := setupLargeBSI(b)
	if bsi == nil {
		b.Skip("skipping, large BSI setup failed")
	}
	foundSet := bsi.CompareValue(0, RANGE, 40, 70, nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res := bsi.CompareValue(0, EQ, 55, 0, foundSet)
		_ = res
	}
}

func BenchmarkBSI64CompareBigValueEQFoundSetLargeAgeFixture(b *testing.B) {
	bsi := setupLargeBSI(b)
	if bsi == nil {
		b.Skip("skipping, large BSI setup failed")
	}
	foundSet := bsi.CompareBigValue(0, RANGE, big.NewInt(40), big.NewInt(70), nil)
	value := big.NewInt(55)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res := bsi.CompareBigValue(0, EQ, value, nil, foundSet)
		_ = res
	}
}

func BenchmarkBSI64CompareValueRangeLargeAgeFixture(b *testing.B) {
	bsi := setupLargeBSI(b)
	if bsi == nil {
		b.Skip("skipping, large BSI setup failed")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res := bsi.CompareValue(0, RANGE, 40, 70, nil)
		_ = res
	}
}

func BenchmarkBSI64CompareBigValueRangeLargeAgeFixture(b *testing.B) {
	bsi := setupLargeBSI(b)
	if bsi == nil {
		b.Skip("skipping, large BSI setup failed")
	}
	start := big.NewInt(40)
	end := big.NewInt(70)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res := bsi.CompareBigValue(0, RANGE, start, end, nil)
		_ = res
	}
}

func BenchmarkBSI64CompareValueGELargeAgeFixture(b *testing.B) {
	bsi := setupLargeBSI(b)
	if bsi == nil {
		b.Skip("skipping, large BSI setup failed")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res := bsi.CompareValue(0, GE, 55, 0, nil)
		_ = res
	}
}

func BenchmarkBSI64CompareBigValueGELargeAgeFixture(b *testing.B) {
	bsi := setupLargeBSI(b)
	if bsi == nil {
		b.Skip("skipping, large BSI setup failed")
	}
	value := big.NewInt(55)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res := bsi.CompareBigValue(0, GE, value, nil, nil)
		_ = res
	}
}
