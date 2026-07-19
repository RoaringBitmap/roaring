package roaring64

import (
	"math/big"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func expectedBSI64BatchEqual(bsi *BSI, query []int64) *Bitmap {
	expected := NewBitmap()
	want := make(map[int64]struct{}, len(query))
	for _, q := range query {
		want[q] = struct{}{}
	}
	iter := bsi.GetExistenceBitmap().Iterator()
	for iter.HasNext() {
		col := iter.Next()
		val, ok := bsi.GetValue(col)
		if ok {
			if _, hit := want[val]; hit {
				expected.Add(col)
			}
		}
	}
	return expected
}

func TestBSI64BatchEqualEdgeCases(t *testing.T) {
	bsi := NewDefaultBSI()
	res := bsi.BatchEqual(0, nil)
	assert.True(t, res.IsEmpty())

	res = bsi.BatchEqual(0, []int64{})
	assert.True(t, res.IsEmpty())

	bsi.SetValue(10, 42)
	bsi.SetValue(20, 100)
	bsi.SetValue(30, 42)
	bsi.SetValue(40, -5)
	bsi.SetValue(50, 5)

	res = bsi.BatchEqual(0, []int64{42})
	assert.Equal(t, uint64(2), res.GetCardinality())
	assert.True(t, res.Contains(10))
	assert.True(t, res.Contains(30))

	res = bsi.BatchEqual(0, []int64{42, 100, 42, 999})
	assert.Equal(t, uint64(3), res.GetCardinality())
	assert.True(t, res.Contains(10))
	assert.True(t, res.Contains(20))
	assert.True(t, res.Contains(30))

	res = bsi.BatchEqual(0, []int64{-5})
	assert.Equal(t, uint64(1), res.GetCardinality())
	assert.True(t, res.Contains(40))
	assert.False(t, res.Contains(50), "negative and positive values with the same magnitude must not collide")

	res = bsi.BatchEqual(0, []int64{5})
	assert.Equal(t, uint64(1), res.GetCardinality())
	assert.True(t, res.Contains(50))
	assert.False(t, res.Contains(40), "positive and negative values with the same magnitude must not collide")

	bsi62 := NewBSI(1<<62, 0)
	bsi62.SetValue(10, 5)
	res = bsi62.BatchEqual(0, []int64{5})
	assert.Equal(t, uint64(1), res.GetCardinality())
	assert.True(t, res.Contains(10))
}

func TestBSI64BatchEqualSubBitWidthMatchesGetValue(t *testing.T) {
	bsi := NewBSI(100, 0)
	assert.Equal(t, 7, bsi.BitCount())

	bsi.SetValue(10, 42)
	bsi.SetValue(20, 99)

	for _, query := range [][]int64{{-5}, {200}, {-5, 42, 200}} {
		expected := expectedBSI64BatchEqual(bsi, query)
		actual := bsi.BatchEqual(0, query)
		assert.True(t, actual.Equals(expected), "query %v expected %v got %v", query, expected.ToArray(), actual.ToArray())
	}
}

func TestBSI64BatchEqualResultIsolation(t *testing.T) {
	bsi := NewDefaultBSI()
	bsi.SetValue(10, 42)
	bsi.SetValue(20, 100)

	res := bsi.BatchEqual(0, []int64{42})
	assert.True(t, res.Contains(10))

	res.Add(999)
	res.Remove(10)

	assert.False(t, bsi.GetExistenceBitmap().Contains(999))
	assert.True(t, bsi.GetExistenceBitmap().Contains(10))

	val, ok := bsi.GetValue(10)
	assert.True(t, ok)
	assert.Equal(t, int64(42), val)

	_, ok = bsi.GetValue(999)
	assert.False(t, ok)
}

func TestBSI64BatchEqualConsistentWithGetValue(t *testing.T) {
	rg := rand.New(rand.NewSource(42))
	for run := 0; run < 15; run++ {
		bsi := NewDefaultBSI()
		numCols := rg.Intn(1000) + 10
		for col := 0; col < numCols; col++ {
			if rg.Float64() < 0.8 {
				val := rg.Int63n(500) - 250
				bsi.SetValue(uint64(col), val)
			}
		}

		querySizes := []int{rg.Intn(10) + 1, rg.Intn(50) + 50, rg.Intn(200) + 100}
		for _, querySize := range querySizes {
			query := make([]int64, querySize)
			for i := range query {
				query[i] = rg.Int63n(600) - 300
			}
			expected := expectedBSI64BatchEqual(bsi, query)

			for _, parallelism := range []int{0, 1, 2, 4} {
				actual := bsi.BatchEqual(parallelism, query)
				if !actual.Equals(expected) {
					t.Fatalf("run=%d querySize=%d parallelism=%d query=%v expected=%v actual=%v",
						run, querySize, parallelism, query, expected.ToArray(), actual.ToArray())
				}
			}
		}
	}
}

func TestBSI64BatchEqualBitCubePattern(t *testing.T) {
	bsi := NewDefaultBSI()
	for col := uint64(0); col < 512; col++ {
		bsi.SetValue(col, int64(col%256))
	}

	odds := make([]int64, 0, 128)
	for v := int64(1); v < 256; v += 2 {
		odds = append(odds, v)
	}

	expected := expectedBSI64BatchEqual(bsi, odds)
	actual := bsi.BatchEqual(0, odds)
	assert.True(t, actual.Equals(expected), "expected %v got %v", expected.ToArray(), actual.ToArray())
}

func TestBSI64BatchEqualExistenceAuthority(t *testing.T) {
	ebm := BitmapOf(1)
	plane := BitmapOf(1, 2)
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

	large := setupLargeBSI(t)
	if large == nil {
		t.Skip("skipping, large BSI setup failed")
	}
	for _, vals := range [][]int64{{16}, {55, 57}, {0, 1, 2, 3}} {
		res := large.BatchEqual(0, vals)
		outside := AndNot(res, large.GetExistenceBitmap())
		assert.True(t, outside.IsEmpty(), "BatchEqual(%v) returned %d columns outside eBM", vals, outside.GetCardinality())
	}
}

func BenchmarkBSI64BatchEqualLargeAgeFixture(b *testing.B) {
	bsi := setupLargeBSI(b)
	if bsi == nil {
		b.Skip("skipping, large BSI setup failed")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res := bsi.BatchEqual(0, []int64{55, 57})
		_ = res
	}
}

func BenchmarkBSI64BatchEqualBigLargeAgeFixture(b *testing.B) {
	bsi := setupLargeBSI(b)
	if bsi == nil {
		b.Skip("skipping, large BSI setup failed")
	}
	values := []*big.Int{big.NewInt(55), big.NewInt(57)}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res := bsi.BatchEqualBig(0, values)
		_ = res
	}
}

func BenchmarkBSI64BatchEqualM128(b *testing.B)          { benchmarkBSI64BatchEqualM(b, 128, 1) }
func BenchmarkBSI64BatchEqualM128Scattered(b *testing.B) { benchmarkBSI64BatchEqualM(b, 128, 2) }
func BenchmarkBSI64BatchEqualM200(b *testing.B)          { benchmarkBSI64BatchEqualM(b, 200, 1) }

func benchmarkBSI64BatchEqualM(b *testing.B, m int, stride int64) {
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
