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

func TestBatchEqualParallelBSIScanHelperAssertion(t *testing.T) {
	unsortedCols := []uint32{10, 5, 20}
	sortedCols := []uint32{5, 10, 20}
	emptyCols := []uint32{}

	t.Run("ParallelBSIScanHelper_Unsorted", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.Errorf("Expected ParallelBSIScanHelper to panic on unsorted cols")
			}
			msg, ok := r.(string)
			if !ok || msg != "ParallelBSIScanHelper: input cols must be sorted in ascending order" {
				t.Errorf("Expected specific panic message, got: %v", r)
			}
		}()
		_ = roaring.ParallelBSIScanHelper(unsortedCols, nil, 0, nil)
	})

	t.Run("ParallelBSIScanHelper_SortedAndEmpty", func(t *testing.T) {
		dummyBA := []*roaring.Bitmap{roaring.NewBitmap()}
		vals := []uint64{0, 1}
		_ = roaring.ParallelBSIScanHelper(sortedCols, dummyBA, 1, vals)
		_ = roaring.ParallelBSIScanHelper(emptyCols, dummyBA, 1, vals)
	})
}

func TestBatchEqualParallelBSIScanHelperValsAssertion(t *testing.T) {
	unsortedVals := []uint64{10, 5, 20}
	sortedCols := []uint32{5, 10, 20}
	dummyBA := []*roaring.Bitmap{roaring.NewBitmap()}

	t.Run("ParallelBSIScanHelper_UnsortedVals", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.Errorf("Expected ParallelBSIScanHelper to panic on unsorted vals")
			}
			msg, ok := r.(string)
			if !ok || msg != "ParallelBSIScanHelper: input vals must be sorted in ascending order" {
				t.Errorf("Expected specific panic message, got: %v", r)
			}
		}()
		_ = roaring.ParallelBSIScanHelper(sortedCols, dummyBA, 1, unsortedVals)
	})
}

func TestBatchEqualManyBitplanes(t *testing.T) {
	// Create a BSI with 70 bitplanes (more than 64!)
	bsi := NewDefaultBSI()

	bsi.eBM.Add(1)
	bsi.eBM.Add(2)
	bsi.eBM.Add(3)
	bsi.eBM.Add(4)

	bsi.bA = make([]*roaring.Bitmap, 70)
	for i := range bsi.bA {
		bsi.bA[i] = roaring.NewBitmap()
	}

	// Column 1: value is 1<<65 (so only plane 65 has it)
	bsi.bA[65].Add(1)
	// Column 2: value is 1<<3 (so only plane 3 has it)
	bsi.bA[3].Add(2)
	// Column 3: value is (1<<65) | (1<<3)
	bsi.bA[65].Add(3)
	bsi.bA[3].Add(3)
	// Column 4: value is 1<<3
	bsi.bA[3].Add(4)

	query := []int64{8}

	// Test Trie Path
	resTrie := bsi.BatchEqual(0, query)
	assert.True(t, resTrie.Contains(2))
	assert.True(t, resTrie.Contains(4))
	assert.False(t, resTrie.Contains(1))
	assert.False(t, resTrie.Contains(3))

	// Test Parallel Scan Path
	vals := []uint64{8}
	resScan := bsi.parallelBatchEqualScan(1, vals)
	assert.True(t, resScan.Contains(2))
	assert.True(t, resScan.Contains(4))
	assert.False(t, resScan.Contains(1))
	assert.False(t, resScan.Contains(3))
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

func TestBatchEqualManyBitplanesPanicSafety(t *testing.T) {
	// Create a BSI with 130 bitplanes (exceeding 128)
	bsi := NewDefaultBSI()

	bsi.eBM.Add(1)
	bsi.eBM.Add(2)

	bsi.bA = make([]*roaring.Bitmap, 130)
	for i := range bsi.bA {
		bsi.bA[i] = roaring.NewBitmap()
	}

	// Set value on plane 129
	bsi.bA[129].Add(1)

	// query with 130 scattered values (size >= 128)
	query := make([]int64, 130)
	for i := range query {
		query[i] = int64(i) * 3
	}

	// This must not panic (it should safely stay on the trie walk and complete)
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("BatchEqual panicked with > 128 bitplanes: %v", r)
		}
	}()

	res := bsi.BatchEqual(0, query)
	_ = res
}
