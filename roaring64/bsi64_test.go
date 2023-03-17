package roaring64

import (
	"bytes"
	"encoding/binary"
	"fmt"
	_ "fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"sort"
	"testing"
	"time"

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

func setup() *BSI {

	bsi := NewBSI(100, 0)
	// Setup values
	for i := 0; i < int(bsi.MaxValue); i++ {
		bsi.SetValue(uint64(i), int64(i))
	}
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
	batch := make([]uint64, 100)
	iter := bsi.GetExistenceBitmap().ManyIterator()
	iter.NextMany(batch)
	var min, max int64
	min = Max64BitSigned
	max = Min64BitSigned
	for i := 0; i < len(batch); i++ {
		v, _ := bsi.GetValue(batch[i])
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
	foundSet := BitmapOf(50)
	newBSI := bsi.NewBSIRetainSet(foundSet)
	assert.Equal(t, uint64(1), newBSI.GetCardinality())
	val, ok := newBSI.GetValue(50)
	assert.True(t, ok)
	assert.Equal(t, val, int64(50))
}

func TestLargeFile(t *testing.T) {

	datEBM, err := ioutil.ReadFile("./testdata/age/EBM")
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n\nIMPORTANT: For testing file IO, the roaring library requires disk access.\nWe omit some tests for now.\n\n")
		return
	}
	dat1, err := ioutil.ReadFile("./testdata/age/1")
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n\nIMPORTANT: For testing file IO, the roaring library requires disk access.\nWe omit some tests for now.\n\n")
		return
	}
	dat2, err := ioutil.ReadFile("./testdata/age/2")
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n\nIMPORTANT: For testing file IO, the roaring library requires disk access.\nWe omit some tests for now.\n\n")
		return
	}
	dat3, err := ioutil.ReadFile("./testdata/age/3")
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n\nIMPORTANT: For testing file IO, the roaring library requires disk access.\nWe omit some tests for now.\n\n")
		return
	}
	dat4, err := ioutil.ReadFile("./testdata/age/4")
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n\nIMPORTANT: For testing file IO, the roaring library requires disk access.\nWe omit some tests for now.\n\n")
		return
	}
	dat5, err := ioutil.ReadFile("./testdata/age/5")
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n\nIMPORTANT: For testing file IO, the roaring library requires disk access.\nWe omit some tests for now.\n\n")
		return
	}
	dat6, err := ioutil.ReadFile("./testdata/age/6")
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n\nIMPORTANT: For testing file IO, the roaring library requires disk access.\nWe omit some tests for now.\n\n")
		return
	}
	dat7, err := ioutil.ReadFile("./testdata/age/7")
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n\nIMPORTANT: For testing file IO, the roaring library requires disk access.\nWe omit some tests for now.\n\n")
		return
	}
	dat8, err := ioutil.ReadFile("./testdata/age/8")
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n\nIMPORTANT: For testing file IO, the roaring library requires disk access.\nWe omit some tests for now.\n\n")
		return
	}

	b := [][]byte{datEBM, dat1, dat2, dat3, dat4, dat5, dat6, dat7, dat8}

	bsi := NewDefaultBSI()
	//bsi.RunOptimize()
	err = bsi.UnmarshalBinary(b)
	require.Nil(t, err)

	resultA := bsi.CompareValue(0, EQ, 55, 0, nil)
	assert.Equal(t, uint64(574600), resultA.GetCardinality())

	resultB := bsi.BatchEqual(0, []int64{55, 57})
	assert.Equal(t, uint64(574600+515233), resultB.GetCardinality())

	bsi.ClearValues(resultA)
	resultC := bsi.BatchEqual(0, []int64{55, 57})
	assert.Equal(t, uint64(515233), resultC.GetCardinality())

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
	bsi.Increment(BitmapOf(0))
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
	transposed := bsi.TransposeWithCounts(0, bsi.GetExistenceBitmap(), bsi.GetExistenceBitmap())
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

func TestBSIWriteToReadFrom(t *testing.T) {
	file, err := ioutil.TempFile("./testdata", "bsi-test")
	if err != nil {
		t.Fatal(err)
	}
	defer t.Cleanup(func() { os.Remove(file.Name()) })
	defer file.Close()
	bsi := setupRandom()
	_, err = bsi.WriteTo(file)
	if err != nil {
		t.Fatal(err)
	}

	file.Seek(io.SeekStart, 0)

	bsi2 := NewDefaultBSI()
	_, err3 := bsi2.ReadFrom(file)
	if err3 != nil {
		t.Fatal(err3)
	}
	assert.True(t, bsi.Equals(bsi2))
	assert.Equal(t, bsi.MinValue, bsi2.MinMax(0, MIN, bsi2.GetExistenceBitmap()))
	assert.Equal(t, bsi.MaxValue, bsi2.MinMax(0, MAX, bsi2.GetExistenceBitmap()))
}

type bsiColValPair struct {
	col uint64
	val int64
}

func bytesToBsiColValPairs(b []byte) (slice []bsiColValPair, err error) {
	r := bytes.NewReader(b)
	for {
		var pair bsiColValPair
		pair.col, err = binary.ReadUvarint(r)
		if err == io.EOF {
			err = nil
			return
		}
		if err != nil {
			return
		}
		pair.val, err = binary.ReadVarint(r)
		if err != nil {
			return
		}
		slice = append(slice, pair)
	}
}

// Checks that the given column values write out and read back in to a BSI without changing. Slice
// should not have duplicate column indexes, as iterator will not render duplicates to match,
// and the BSI will contain the last value set.
func testBsiRoundTrip(t *testing.T, pairs []bsiColValPair) {
	bsi := NewDefaultBSI()
	for _, pair := range pairs {
		bsi.SetValue(pair.col, pair.val)
	}
	var buf bytes.Buffer
	_, err := bsi.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	_, err = bsi.ReadFrom(&buf)
	if err != nil {
		t.Fatal(err)
	}
	it := bsi.GetExistenceBitmap().Iterator()
	// The column ordering needs to match the one given by the iterator. This reorders the caller's
	// slice.
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].col < pairs[j].col
	})
	for _, pair := range pairs {
		if !it.HasNext() {
			t.Fatalf("expected more columns: %v", pair.col)
		}
		bsiCol := it.Next()
		if pair.col != bsiCol {
			t.Fatalf("expected col %d, got %d", pair.col, bsiCol)
		}
		bsiVal, ok := bsi.GetValue(bsiCol)
		if !ok {
			t.Fatalf("expected col %d to exist", bsiCol)
		}
		if pair.val != bsiVal {
			t.Fatalf("expected col %d to have value %d, got %d", bsiCol, pair.val, bsiVal)
		}
	}
	if it.HasNext() {
		t.Fatal("expected no more columns")
	}

}

func TestBsiStreaming(t *testing.T) {
	testBsiRoundTrip(t, []bsiColValPair{})
	testBsiRoundTrip(t, []bsiColValPair{{0, 0}})
	testBsiRoundTrip(t, []bsiColValPair{{48, 0}})
}

// Test that the BSI can be mutated and still be equal to a fresh BSI with the same values.
func TestMutatedBsiEquality(t *testing.T) {
	mutated := NewDefaultBSI()
	mutated.SetValue(0, 2)
	mutated.SetValue(0, 1)
	fresh := NewDefaultBSI()
	fresh.SetValue(0, 1)
	assert.True(t, fresh.Equals(mutated))
	fresh.SetValue(0, 2)
	assert.False(t, fresh.Equals(mutated))
	// Now fresh has been mutated in the same pattern as mutated.
	fresh.SetValue(0, 1)
	assert.True(t, fresh.Equals(mutated))
}
