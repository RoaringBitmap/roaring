package roaring

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/RoaringBitmap/roaring/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetAndGet(t *testing.T) {

	bsi := NewBSI(999, 0)
	require.NotNil(t, bsi.bA)
	assert.Equal(t, 10, len(bsi.bA))

	bsi.SetValue(1, 8)
	gv, ok := bsi.GetValue(1)
	assert.True(t, ok)
	assert.Equal(t, int64(8), gv)
}

func TestSetMany(t *testing.T) {
	bsi := setup()
	// update with mix of existing and new columns
	upd := roaring.BitmapOf(30, 31, 32, 33, 34, 35, 101, 102, 103)
	bsi.SetMany(upd, 35)

	matches := bsi.CompareValue(0, EQ, 35, 0, nil)

	assert.True(t, upd.Equals(matches))
}

func setup() *BSI {

	bsi := NewBSI(100, 0)
	// Setup values
	for i := 0; i < int(bsi.MaxValue); i++ {
		bsi.SetValue(uint64(i), int64(i))
	}
	return bsi
}

func setupLargeBSI(t testing.TB) *BSI {
	t.Helper()

	datEBM, err := os.ReadFile("./testdata/age/EBM")
	if err != nil {
		return nil
	}
	b := make([][]byte, 9)
	b[0] = datEBM
	for i := 1; i <= 8; i++ {
		b[i], err = os.ReadFile(fmt.Sprintf("./testdata/age/%d", i))
		if err != nil {
			return nil
		}
	}
	bsi := NewDefaultBSI()
	err = bsi.UnmarshalBinary(b)
	require.NoError(t, err)
	return bsi
}

func setupNegativeBoundary() *BSI {

	bsi := NewBSI(5, -5)
	// Setup values
	for i := int(bsi.MinValue); i <= int(bsi.MaxValue); i++ {
		bsi.SetValue(uint64(i), int64(i))
	}
	return bsi
}

func setupAllNegative() *BSI {
	bsi := NewBSI(-1, -100)
	// Setup values
	for i := int(bsi.MinValue); i <= int(bsi.MaxValue); i++ {
		bsi.SetValue(uint64(i), int64(i))
	}
	return bsi
}

func setupAutoSizeNegativeBoundary() *BSI {
	bsi := NewDefaultBSI()
	// Setup values
	for i := int(-5); i <= int(5); i++ {
		bsi.SetValue(uint64(i), int64(i))
	}
	return bsi
}

func setupRandom() *BSI {
	bsi := NewBSI(99, -1)
	rg := rand.New(rand.NewSource(time.Now().UnixNano()))
	// Setup values
	for i := 0; bsi.GetExistenceBitmap().GetCardinality() < 100; {
		rv := rg.Int63n(bsi.MaxValue) - 50
		_, ok := bsi.GetValue(uint64(i))
		if ok {
			continue
		}
		bsi.SetValue(uint64(i), rv)
		i++
	}
	batch := make([]uint32, 100)
	iter := bsi.GetExistenceBitmap().ManyIterator()
	iter.NextMany(batch)
	var min, max int64
	min = Max64BitSigned
	max = Min64BitSigned
	for i := 0; i < len(batch); i++ {
		v, _ := bsi.GetValue(uint64(batch[i]))
		if v > max {
			max = v
		}
		if v < min {
			min = v
		}
	}
	bsi.MinValue = min
	bsi.MaxValue = max
	return bsi
}

func TestEQ(t *testing.T) {
	bsi := setup()
	eq := bsi.CompareValue(0, EQ, 50, 0, nil)
	assert.Equal(t, uint64(1), eq.GetCardinality())

	assert.True(t, eq.ContainsInt(50))
}

func TestLT(t *testing.T) {

	bsi := setup()
	lt := bsi.CompareValue(0, LT, 50, 0, nil)
	assert.Equal(t, uint64(50), lt.GetCardinality())

	i := lt.Iterator()
	for i.HasNext() {
		v := i.Next()
		assert.Less(t, uint64(v), uint64(50))
	}
}

func TestGT(t *testing.T) {

	bsi := setup()
	gt := bsi.CompareValue(0, GT, 50, 0, nil)
	assert.Equal(t, uint64(49), gt.GetCardinality())

	i := gt.Iterator()
	for i.HasNext() {
		v := i.Next()
		assert.Greater(t, uint64(v), uint64(50))
	}
}

func TestGE(t *testing.T) {

	bsi := setup()
	ge := bsi.CompareValue(0, GE, 50, 0, nil)
	assert.Equal(t, uint64(50), ge.GetCardinality())

	i := ge.Iterator()
	for i.HasNext() {
		v := i.Next()
		assert.GreaterOrEqual(t, uint64(v), uint64(50))
	}
}

func TestLE(t *testing.T) {

	bsi := setup()
	le := bsi.CompareValue(0, LE, 50, 0, nil)
	assert.Equal(t, uint64(51), le.GetCardinality())

	i := le.Iterator()
	for i.HasNext() {
		v := i.Next()
		assert.LessOrEqual(t, uint64(v), uint64(50))
	}
}

func TestRange(t *testing.T) {

	bsi := setup()
	set := bsi.CompareValue(0, RANGE, 45, 55, nil)
	assert.Equal(t, uint64(11), set.GetCardinality())

	i := set.Iterator()
	for i.HasNext() {
		v := i.Next()
		assert.GreaterOrEqual(t, uint64(v), uint64(45))
		assert.LessOrEqual(t, uint64(v), uint64(55))
	}
}

func TestCompareValueMatchesGetValue(t *testing.T) {
	type query struct {
		op    Operation
		start int64
		end   int64
	}
	tests := []struct {
		name    string
		newBSI  func() *BSI
		values  []int64
		queries []query
	}{
		{
			name:   "unsigned",
			newBSI: func() *BSI { return NewBSI(127, 0) },
			values: []int64{0, 1, 2, 3, 7, 31, 63, 126, 127},
			queries: []query{
				{EQ, -1, 0}, {EQ, 0, 0}, {EQ, 127, 0}, {EQ, 128, 0},
				{LT, 0, 0}, {LE, 0, 0}, {GT, 127, 0}, {GE, 128, 0},
				{RANGE, -1, 1}, {RANGE, 31, 126}, {RANGE, 127, 126},
			},
		},
		{
			name:   "signed",
			newBSI: NewDefaultBSI,
			values: []int64{Min64BitSigned, -100, -2, -1, 0, 1, 2, 100, Max64BitSigned},
			queries: []query{
				{EQ, Min64BitSigned, 0}, {EQ, -1, 0}, {EQ, 1, 0}, {EQ, Max64BitSigned, 0},
				{LT, Min64BitSigned, 0}, {LT, 1, 0}, {LE, -100, 0}, {LE, 1, 0},
				{GT, -1, 0}, {GT, Max64BitSigned, 0}, {GE, -1, 0}, {GE, 0, 0},
				{RANGE, Min64BitSigned, -1}, {RANGE, -2, 2}, {RANGE, 1, -1}, {RANGE, Max64BitSigned, Max64BitSigned},
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		for _, optimized := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/run_optimized=%t", tc.name, optimized), func(t *testing.T) {
				bsi := tc.newBSI()
				for columnID, value := range tc.values {
					bsi.SetValue(uint64(columnID), value)
				}
				if optimized {
					bsi.RunOptimize()
				}

				foundSet := roaring.NewBitmap()
				for columnID := range tc.values {
					if columnID%2 == 0 {
						foundSet.Add(uint32(columnID))
					}
				}
				for _, selection := range []struct {
					name  string
					input *roaring.Bitmap
				}{
					{name: "existence", input: nil},
					{name: "subset", input: foundSet},
				} {
					for _, parallelism := range []int{0, 1, 3} {
						for _, q := range tc.queries {
							got := bsi.CompareValue(parallelism, q.op, q.start, q.end, selection.input)
							want := compareValueGroundTruth(bsi, q.op, q.start, q.end, selection.input)
							if !got.Equals(want) {
								t.Errorf("%s parallelism=%d op=%d start=%d end=%d: got %v, want %v", selection.name, parallelism, q.op, q.start, q.end, got.ToArray(), want.ToArray())
							}
						}
					}
				}

				before := foundSet.Clone()
				result := bsi.CompareValue(0, EQ, tc.values[0], 0, foundSet)
				result.Remove(0)
				result.Add(1000)
				assert.True(t, foundSet.Equals(before), "CompareValue must not mutate or share foundSet")

				existenceBefore := bsi.GetExistenceBitmap().Clone()
				result = bsi.CompareValue(0, EQ, tc.values[0], 0, nil)
				result.Remove(0)
				result.Add(1000)
				assert.True(t, bsi.GetExistenceBitmap().Equals(existenceBefore), "CompareValue must not mutate or share the existence bitmap")
			})
		}
	}
}

func compareValueGroundTruth(bsi *BSI, op Operation, start, end int64, foundSet *roaring.Bitmap) *roaring.Bitmap {
	if foundSet == nil {
		foundSet = bsi.GetExistenceBitmap()
	}

	result := roaring.NewBitmap()
	iter := foundSet.Iterator()
	for iter.HasNext() {
		columnID := iter.Next()
		value, exists := bsi.GetValue(uint64(columnID))
		if !exists {
			continue
		}

		matches := false
		switch op {
		case LT:
			matches = value < start
		case LE:
			matches = value <= start
		case EQ:
			matches = value == start
		case GE:
			matches = value >= start
		case GT:
			matches = value > start
		case RANGE:
			matches = value >= start && value <= end
		default:
			panic(fmt.Sprintf("Unknown operation [%v]", op))
		}
		if matches {
			result.Add(columnID)
		}
	}
	return result
}

func TestExists(t *testing.T) {

	bsi := NewBSI(10, 0)
	// Setup values
	for i := 1; i < int(bsi.MaxValue); i++ {
		bsi.SetValue(uint64(i), int64(i))
	}

	assert.Equal(t, uint64(9), bsi.GetCardinality())
	assert.False(t, bsi.ValueExists(uint64(0)))
	bsi.SetValue(uint64(0), int64(0))
	assert.Equal(t, uint64(10), bsi.GetCardinality())
	assert.True(t, bsi.ValueExists(uint64(0)))
}

func TestSum(t *testing.T) {

	bsi := setup()
	set := bsi.CompareValue(0, RANGE, 45, 55, nil)

	sum, count := bsi.Sum(set)
	assert.Equal(t, uint64(11), count)
	assert.Equal(t, int64(550), sum)
}

func TestTranspose(t *testing.T) {

	bsi := NewBSI(100, 0)
	// Setup values
	for i := 0; i < int(bsi.MaxValue); i++ {
		bsi.SetValue(uint64(i+100), int64(i))
	}

	set := bsi.Transpose()
	assert.Equal(t, uint64(100), set.GetCardinality())

	i := set.Iterator()
	j := 0
	for i.HasNext() {
		v := i.Next()
		assert.Equal(t, uint64(v), uint64(j))
		j++
	}
}

func TestAutoSize(t *testing.T) {

	bsi := NewDefaultBSI()
	for i := 0; i < 100; i++ {
		bsi.SetValue(uint64(i), int64(i))
	}

	require.NotNil(t, bsi.bA)
	assert.Equal(t, 7, bsi.BitCount())

	for i := 0; i < 100; i++ {
		gv, ok := bsi.GetValue(uint64(i))
		assert.True(t, ok)
		assert.Equal(t, int64(i), gv)
	}
}

func TestParOr(t *testing.T) {

	bsi1 := NewDefaultBSI()
	for i := 0; i < 100; i++ {
		bsi1.SetValue(uint64(i), int64(i))
	}
	bsi2 := NewDefaultBSI()
	for i := 0; i < 100; i++ {
		bsi2.SetValue(uint64(i+100), int64(i+100))
	}
	bsi1.ParOr(0, bsi2)
	for i := 0; i < 200; i++ {
		gv, ok := bsi1.GetValue(uint64(i))
		assert.True(t, ok)
		assert.Equal(t, int64(i), gv)
	}
	assert.Equal(t, uint64(200), bsi1.eBM.GetCardinality())
}

func TestNewBSIRetainSet(t *testing.T) {

	bsi := setup()
	foundSet := roaring.BitmapOf(50)
	newBSI := bsi.NewBSIRetainSet(foundSet)
	assert.Equal(t, uint64(1), newBSI.GetCardinality())
	val, ok := newBSI.GetValue(50)
	assert.True(t, ok)
	assert.Equal(t, val, int64(50))
}

func TestLargeFile(t *testing.T) {

	bsi := setupLargeBSI(t)
	if bsi == nil {
		fmt.Fprintf(os.Stderr, "\n\nIMPORTANT: For testing file IO, the roaring library requires disk access.\nWe omit some tests for now.\n\n")
		return
	}

	resultA := bsi.CompareValue(0, EQ, 55, 0, nil)
	assert.Equal(t, uint64(520157), resultA.GetCardinality())

	resultB := bsi.BatchEqual(0, []int64{55, 57})
	assert.Equal(t, uint64(520157+486001), resultB.GetCardinality())

	bsi.ClearValues(resultA)
	resultC := bsi.BatchEqual(0, []int64{55, 57})
	assert.Equal(t, uint64(486001), resultC.GetCardinality())

}

func TestClone(t *testing.T) {
	bsi := setup()
	clone := bsi.Clone()
	for i := 0; i < int(bsi.MaxValue); i++ {
		a, _ := bsi.GetValue(uint64(i))
		b, _ := clone.GetValue(uint64(i))
		assert.Equal(t, a, b)
	}
}

func TestAdd(t *testing.T) {
	bsi := NewDefaultBSI()
	// Setup values
	for i := 1; i <= 10; i++ {
		bsi.SetValue(uint64(i), int64(i))
	}
	clone := bsi.Clone()
	bsi.Add(clone)
	assert.Equal(t, uint64(10), bsi.GetCardinality())
	for i := 1; i <= 10; i++ {
		a, _ := bsi.GetValue(uint64(i))
		b, _ := clone.GetValue(uint64(i))
		assert.Equal(t, b*2, a)
	}

}

func TestIncrement(t *testing.T) {
	bsi := setup()
	bsi.IncrementAll()
	for i := 0; i < int(bsi.MaxValue); i++ {
		a, _ := bsi.GetValue(uint64(i))
		assert.Equal(t, int64(i+1), a)
	}
	bsi.Increment(roaring.BitmapOf(0))
	x, _ := bsi.GetValue(uint64(0))
	assert.Equal(t, int64(2), x)
	for i := 1; i < int(bsi.MaxValue); i++ {
		a, _ := bsi.GetValue(uint64(i))
		assert.Equal(t, int64(i+1), a)
	}
}

func TestTransposeWithCounts(t *testing.T) {
	bsi := setup()
	bsi.SetValue(101, 50)
	transposed := bsi.TransposeWithCounts(0, bsi.GetExistenceBitmap())
	a, ok := transposed.GetValue(uint64(50))
	assert.True(t, ok)
	assert.Equal(t, int64(2), a)
}

func TestRangeAllNegative(t *testing.T) {
	bsi := setupAllNegative()
	assert.Equal(t, uint64(100), bsi.GetCardinality())
	set := bsi.CompareValue(0, RANGE, -55, -45, nil)
	assert.Equal(t, uint64(11), set.GetCardinality())

	i := set.Iterator()
	for i.HasNext() {
		val, _ := bsi.GetValue(uint64(i.Next()))
		assert.GreaterOrEqual(t, val, int64(-55))
		assert.LessOrEqual(t, val, int64(-45))
	}
}

func TestSumWithNegative(t *testing.T) {
	bsi := setupNegativeBoundary()
	assert.Equal(t, uint64(11), bsi.GetCardinality())
	sum, cnt := bsi.Sum(bsi.GetExistenceBitmap())
	assert.Equal(t, uint64(11), cnt)
	assert.Equal(t, int64(0), sum)
}

func TestGEWithNegative(t *testing.T) {
	bsi := setupNegativeBoundary()
	assert.Equal(t, uint64(11), bsi.GetCardinality())
	set := bsi.CompareValue(0, GE, 3, 0, nil)
	assert.Equal(t, uint64(3), set.GetCardinality())
	set = bsi.CompareValue(0, GE, -3, 0, nil)
	assert.Equal(t, uint64(9), set.GetCardinality())
}

func TestLEWithNegative(t *testing.T) {
	bsi := setupNegativeBoundary()
	assert.Equal(t, uint64(11), bsi.GetCardinality())
	set := bsi.CompareValue(0, LE, -3, 0, nil)
	assert.Equal(t, uint64(3), set.GetCardinality())
	set = bsi.CompareValue(0, LE, 3, 0, nil)
	assert.Equal(t, uint64(9), set.GetCardinality())
}

func TestRangeWithNegative(t *testing.T) {
	bsi := setupNegativeBoundary()
	assert.Equal(t, uint64(11), bsi.GetCardinality())
	set := bsi.CompareValue(0, RANGE, -3, 3, nil)
	assert.Equal(t, uint64(7), set.GetCardinality())

	i := set.Iterator()
	for i.HasNext() {
		val, _ := bsi.GetValue(uint64(i.Next()))
		assert.GreaterOrEqual(t, val, int64(-3))
		assert.LessOrEqual(t, val, int64(3))
	}
}

func TestAutoSizeWithNegative(t *testing.T) {
	bsi := setupAutoSizeNegativeBoundary()
	assert.Equal(t, uint64(11), bsi.GetCardinality())
	assert.Equal(t, 64, bsi.BitCount())
	set := bsi.CompareValue(0, RANGE, -3, 3, nil)
	assert.Equal(t, uint64(7), set.GetCardinality())

	i := set.Iterator()
	for i.HasNext() {
		val, _ := bsi.GetValue(uint64(i.Next()))
		assert.GreaterOrEqual(t, val, int64(-3))
		assert.LessOrEqual(t, val, int64(3))
	}
}

func TestMinMaxWithRandom(t *testing.T) {
	bsi := setupRandom()
	assert.Equal(t, bsi.MinValue, bsi.MinMax(0, MIN, bsi.GetExistenceBitmap()))
	assert.Equal(t, bsi.MaxValue, bsi.MinMax(0, MAX, bsi.GetExistenceBitmap()))
}

func BenchmarkSetRoaring(b *testing.B) {
	b.StopTimer()
	r := rand.New(rand.NewSource(0))
	sz := 100_000_000
	s := NewDefaultBSI()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < 100; j++ {
			s.SetValue(uint64(r.Int31n(int32(sz))), int64(r.Int31n(int32(sz))))
		}
	}
}

func BenchmarkClearValues(b *testing.B) {
	bsi := setupLargeBSI(b)
	if bsi == nil {
		b.Skip("\n\nIMPORTANT: For testing file IO, the roaring library requires disk access.\nWe omit some tests for now.\n\n")
		return
	}
	resultA := bsi.CompareValue(0, EQ, 55, 0, nil)
	assert.Equal(b, uint64(520157), resultA.GetCardinality())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		b2 := bsi.Clone()
		b.StartTimer()
		b2.ClearValues(resultA)
	}
}

func TestIssue426(t *testing.T) {
	bsi := NewBSI(101, 0)
	bsi.SetValue(3, 5)
	bitmap := bsi.CompareValue(0, EQ, 101, 0, nil)
	fmt.Println(bitmap.ToArray())
	assert.Equal(t, uint64(0), bitmap.GetCardinality())
}

func TestMinMaxWithNil(t *testing.T) {
	bsi := setupRandom()
	assert.Equal(t, bsi.MinValue, bsi.MinMax(0, MIN, nil))
	assert.Equal(t, bsi.MaxValue, bsi.MinMax(0, MAX, nil))
}

func TestSumWithNil(t *testing.T) {

	bsi := setup()

	sum, count := bsi.Sum(bsi.GetExistenceBitmap())
	sumNil, countNil := bsi.Sum(nil)
	assert.Equal(t, countNil, count)
	assert.Equal(t, sumNil, sum)
}

func TestTransposeWithCountsNil(t *testing.T) {
	bsi := setup()
	bsi.SetValue(101, 50)
	transposed := bsi.TransposeWithCounts(0, nil)
	a, ok := transposed.GetValue(uint64(50))
	assert.True(t, ok)
	assert.Equal(t, int64(2), a)
}

// TestBatchEqualLargeQueryValues drives the scattered-query scan path: a large
// existence bitmap and a 128+ value query push BatchEqual past the crossover, and
// the result is pinned to the GetValue ground truth across parallelism levels, so
// the linear scan and its partitioning stay authoritative (subset of eBM) and exact.
func TestBatchEqualLargeQueryValues(t *testing.T) {
	rg := rand.New(rand.NewSource(12345))
	for run := 0; run < 10; run++ {
		// Large values (>= 2^20) and a big column set push past the scan crossover.
		bsi := NewDefaultBSI()
		numCols := rg.Intn(50000) + 120000
		for col := 0; col < numCols; col++ {
			if rg.Float64() < 0.8 {
				val := rg.Int63n(100000) + 1048500
				bsi.SetValue(uint64(col), val)
			}
		}

		querySize := rg.Intn(100) + 128
		query := make([]int64, querySize)
		for i := range query {
			query[i] = rg.Int63n(100100) + 1048500
		}

		// Ground truth: GetValue per existing column.
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

		for _, parallelism := range []int{0, 1, 2, 4} {
			actual := bsi.BatchEqual(parallelism, query)
			if !actual.Equals(expected) {
				t.Fatalf("mismatch in run %d parallelism %d: expected %v, got %v", run, parallelism, expected.ToArray(), actual.ToArray())
			}
		}
	}
}
