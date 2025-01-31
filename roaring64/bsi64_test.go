package roaring64

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"math/rand"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// Min64BitSigned - Minimum 64 bit value
	Min64BitSigned = -9223372036854775808
	// Max64BitSigned - Maximum 64 bit value
	Max64BitSigned = 9223372036854775807
)

func TestSetAndGetSimple(t *testing.T) {

	bsi := NewBSI(999, 0)
	require.NotNil(t, bsi.bA)
	assert.Equal(t, 10, bsi.BitCount())

	bsi.SetValue(1, 8)
	gv, ok := bsi.GetValue(1)
	assert.True(t, ok)
	assert.Equal(t, int64(8), gv)
}

func TestSetAndGetBigValue(t *testing.T) {

	// Set a large UUID value---
	bsi := NewDefaultBSI()
	bigUUID := big.NewInt(-578664753978847603) // Upper bits
	bigUUID.Lsh(bigUUID, 64)
	lowBits := big.NewInt(-5190910309365112881) // Lower bits
	bigUUID.Add(bigUUID, lowBits)               // Lower bits

	bsi.SetBigValue(1, bigUUID)
	assert.Equal(t, bigUUID.BitLen(), bsi.BitCount())
	bv, _ := bsi.GetBigValue(1)
	assert.Equal(t, bigUUID, bv)

	// Any code past this point will expect a panic error.  This will happen if a large value was set
	// with SetBigValue() followed by a call to GetValue() where the set value exceeds 64 bits.
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	bsi.GetValue(1) // this should panic.  If so the test will pass.
}

func TestSetAndGetUUIDValue(t *testing.T) {
	uuidVal, _ := uuid.NewRandom()
	b, errx := uuidVal.MarshalBinary()
	assert.Nil(t, errx)
	bigUUID := new(big.Int)
	bigUUID.SetBytes(b)
	bsi := NewDefaultBSI()
	bsi.SetBigValue(1, bigUUID)
	assert.Equal(t, bigUUID.BitLen(), bsi.BitCount())
	bv, _ := bsi.GetBigValue(1)
	assert.Equal(t, bigUUID, bv)

	newUUID, err := uuid.FromBytes(bv.Bytes())
	assert.Nil(t, err)

	assert.Equal(t, uuidVal.String(), newUUID.String())
}

func secondsAndNanosToBigInt(seconds int64, nanos int32) *big.Int {
	b := make([]byte, 12)
	binary.BigEndian.PutUint64(b[:8], uint64(seconds))
	binary.BigEndian.PutUint32(b[8:], uint32(nanos))
	bigTime := new(big.Int)
	bigTime.SetBytes(b)
	return bigTime
}

func bigIntToSecondsAndNanos(big *big.Int) (seconds int64, nanos int32) {
	buf := make([]byte, 12)
	big.FillBytes(buf)
	seconds = int64(binary.BigEndian.Uint64(buf[:8]))
	nanos = int32(binary.BigEndian.Uint32(buf[8:]))
	return
}

func TestSetAndGetBigTimestamp(t *testing.T) {

	// Store a timestamp in a BSI as 2 values, seconds as int64 and nanosecond interval as int32 (96 bits)
	bigTime := secondsAndNanosToBigInt(int64(33286611346), int32(763295273))
	bsi := NewDefaultBSI()
	bsi.SetBigValue(1, bigTime)

	// Recover and check the known timestamp
	bv, _ := bsi.GetBigValue(1)
	seconds, nanoseconds := bigIntToSecondsAndNanos(bv)
	ts := time.Unix(seconds, int64(nanoseconds))
	assert.Equal(t, "3024-10-23T16:55:46.763295273Z", ts.UTC().Format(time.RFC3339Nano))
	assert.Equal(t, 67, bsi.BitCount())
}

// This tests a corner case where a zero value is set on an empty BSI.  The bit count should never be zero.
func TestSetInitialValueZero(t *testing.T) {
	bsi := NewDefaultBSI()
	bsi.SetBigValue(1, big.NewInt(0))
	assert.Equal(t, 1, bsi.BitCount())
}

func TestRangeBig(t *testing.T) {

	bsi := NewDefaultBSI()

	// Populate large timestamp values
	for i := 0; i <= 100; i++ {
		t := time.Now()
		newTime := t.AddDate(1000, 0, 0) // Add 1000 years
		secs := int64(newTime.UnixMilli() / 1000)
		nano := int32(newTime.Nanosecond())
		bigTime := secondsAndNanosToBigInt(secs, nano)
		bsi.SetBigValue(uint64(i), bigTime)
	}

	start, _ := bsi.GetBigValue(uint64(45)) // starting value at columnID 45
	end, _ := bsi.GetBigValue(uint64(55))   // ending value at columnID 55
	set := bsi.CompareBigValue(0, RANGE, start, end, nil)
	assert.Equal(t, uint64(11), set.GetCardinality())

	i := set.Iterator()
	for i.HasNext() {
		v := i.Next()
		assert.GreaterOrEqual(t, uint64(v), uint64(45))
		assert.LessOrEqual(t, uint64(v), uint64(55))
	}
	assert.Equal(t, 67, bsi.BitCount())
}

func setup() *BSI {
	bsi := NewBSI(100, 0)
	// Setup values
	for i := 0; i <= int(bsi.MaxValue); i++ {
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

func setupRandom() (bsi *BSI, min, max int64) {
	bsi = NewBSI(99, -1)
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
	return bsi, min, max
}

func TestTwosComplement(t *testing.T) {
	assert.Equal(t, "1001110", twosComplement(big.NewInt(-50), 7).Text(2))
	assert.Equal(t, "110010", twosComplement(big.NewInt(50), 7).Text(2))
	assert.Equal(t, "0", twosComplement(big.NewInt(0), 7).Text(2))
	assert.Equal(t, "111001110", twosComplement(big.NewInt(-50), 9).Text(2))
	assert.Equal(t, "1111101", twosComplement(big.NewInt(-3), 7).Text(2))
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
	assert.Equal(t, uint64(50), gt.GetCardinality())

	i := gt.Iterator()
	for i.HasNext() {
		v := i.Next()
		assert.Greater(t, uint64(v), uint64(50))
	}
}

func TestNewBSI(t *testing.T) {
	bsi := NewBSI(100, 0)
	assert.Equal(t, 7, bsi.BitCount())
	bsi = NewBSI(5, -5)
	negBits := big.NewInt(-5)
	assert.Equal(t, negBits.BitLen(), bsi.BitCount())
	posBits := big.NewInt(5)
	assert.Equal(t, posBits.BitLen(), bsi.BitCount())

	bsi = NewDefaultBSI()
	assert.Equal(t, 0, bsi.BitCount())
	bsi.SetValue(1, int64(0))
	assert.Equal(t, 1, bsi.BitCount())
	bsi.SetValue(1, int64(-1))
	assert.Equal(t, 1, bsi.BitCount())
}

func TestGE(t *testing.T) {

	bsi := setup()
	ge := bsi.CompareValue(0, GE, 50, 0, nil)
	assert.Equal(t, uint64(51), ge.GetCardinality())

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

func TestRangeSimple(t *testing.T) {

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

func TestSumSimple(t *testing.T) {

	bsi := setup()
	set := bsi.CompareValue(0, RANGE, 45, 55, nil)

	sum, count := bsi.Sum(set)
	assert.Equal(t, uint64(11), count)
	assert.Equal(t, int64(550), sum)
}

func TestTransposeSimple(t *testing.T) {

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
	bsi := NewDefaultBSI()
	// Setup values
	for i := 1; i <= 10; i++ {
		bsi.SetValue(uint64(i), int64(i))
	}
	clone := bsi.Clone()
	for i := 0; i < 10; i++ {
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
	assert.Equal(t, uint64(10), clone.GetCardinality())
	bsi.Add(clone)
	assert.Equal(t, uint64(10), bsi.GetCardinality())
	for i := 1; i <= 10; i++ {
		a, _ := bsi.GetValue(uint64(i))
		b, _ := clone.GetValue(uint64(i))
		assert.Equal(t, b*2, a)
	}

}

func TestBatchValueBig(t *testing.T) {
	bsi := NewDefaultBSI()

	// create a big value
	bv := big.NewInt(Max64BitSigned)
	bv.Mul(bv, big.NewInt(100))

	// Populate large timestamp values
	for i := 0; i <= 100; i++ {
		bsi.SetBigValue(uint64(i), bv)
	}
	result := bsi.BatchEqualBig(0, []*big.Int{bv})
	assert.Equal(t, uint64(101), result.GetCardinality())
}

func TestIncrementSimple(t *testing.T) {
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

func TestIncrementFromZero(t *testing.T) {
	bsi := NewDefaultBSI()
	for i := 0; i < 10; i++ {
		bsi.SetValue(uint64(i), 0)
	}
	bsi.IncrementAll()

	assert.Equal(t, uint64(10), bsi.GetCardinality())
	sum, cnt := bsi.Sum(bsi.GetExistenceBitmap())
	assert.Equal(t, uint64(10), cnt)
	assert.Equal(t, int64(10), sum)
}

func TestTransposeWithCounts(t *testing.T) {
	bsi := setup()
	bsi.SetValue(101, 50)
	transposed := bsi.TransposeWithCounts(0, bsi.GetExistenceBitmap(), bsi.GetExistenceBitmap())
	a, ok := transposed.GetValue(uint64(50))
	assert.True(t, ok)
	assert.Equal(t, int64(2), a)
	a, ok = transposed.GetValue(uint64(49))
	assert.True(t, ok)
	assert.Equal(t, int64(1), a)
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
	//set := bsi.CompareValue(0, GE, 3, 0, nil)
	//assert.Equal(t, uint64(3), set.GetCardinality())
	set := bsi.CompareValue(0, GE, -3, 0, nil)
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
	set := bsi.CompareValue(0, RANGE, -3, 3, nil)

	i := set.Iterator()
	for i.HasNext() {
		val, _ := bsi.GetValue(uint64(i.Next()))
		assert.GreaterOrEqual(t, val, int64(-3))
		assert.LessOrEqual(t, val, int64(3))
	}
}

func TestMinMaxSimple(t *testing.T) {
	bsi := setup()
	assert.Equal(t, int64(0), bsi.MinMax(0, MIN, bsi.GetExistenceBitmap()))
	assert.Equal(t, int64(100), bsi.MinMax(0, MAX, bsi.GetExistenceBitmap()))
}

func TestMinMaxAllNegative(t *testing.T) {
	bsi := setupAllNegative()
	assert.Equal(t, int64(-100), bsi.MinMax(0, MIN, bsi.GetExistenceBitmap()))
	assert.Equal(t, int64(-1), bsi.MinMax(0, MAX, bsi.GetExistenceBitmap()))
}

func TestMinMaxWithNegative(t *testing.T) {
	bsi := setupAutoSizeNegativeBoundary()
	assert.Equal(t, int64(-5), bsi.MinMax(0, MIN, bsi.GetExistenceBitmap()))
	assert.Equal(t, int64(5), bsi.MinMax(0, MAX, bsi.GetExistenceBitmap()))
}

func TestMinMaxWithRandom(t *testing.T) {
	bsi, min, max := setupRandom()
	assert.Equal(t, min, bsi.MinMax(0, MIN, bsi.GetExistenceBitmap()))
	assert.Equal(t, max, bsi.MinMax(0, MAX, bsi.GetExistenceBitmap()))
}

func TestMinMaxWithNilFoundSet(t *testing.T) {
	bsi, min, max := setupRandom()
	assert.Equal(t, min, bsi.MinMax(0, MIN, nil))
	assert.Equal(t, max, bsi.MinMax(0, MAX, nil))
}

func TestBSIWriteToReadFrom(t *testing.T) {
	file, err := ioutil.TempFile("./testdata", "bsi-test")
	if err != nil {
		t.Fatal(err)
	}
	defer t.Cleanup(func() { os.Remove(file.Name()) })
	defer file.Close()
	bsi, min, max := setupRandom()
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
	assert.Equal(t, min, bsi2.MinMax(0, MIN, bsi2.GetExistenceBitmap()))
	assert.Equal(t, max, bsi2.MinMax(0, MAX, bsi2.GetExistenceBitmap()))
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

func TestSumWithNil(t *testing.T) {
	bsi := setupNegativeBoundary()
	assert.Equal(t, uint64(11), bsi.GetCardinality())
	sum, cnt := bsi.Sum(nil)
	assert.Equal(t, uint64(11), cnt)
	assert.Equal(t, int64(0), sum)
}

func TestTransposeWithCountsNil(t *testing.T) {
	bsi := setup()
	bsi.SetValue(101, 50)
	transposed := bsi.TransposeWithCounts(0, nil, nil)
	a, ok := transposed.GetValue(uint64(50))
	assert.True(t, ok)
	assert.Equal(t, int64(2), a)
	a, ok = transposed.GetValue(uint64(49))
	assert.True(t, ok)
	assert.Equal(t, int64(1), a)
}

func TestRangeNilBig(t *testing.T) {

	bsi := NewDefaultBSI()

	// Populate large timestamp values
	for i := 0; i <= 100; i++ {
		t := time.Now()
		newTime := t.AddDate(1000, 0, 0) // Add 1000 years
		secs := int64(newTime.UnixMilli() / 1000)
		nano := int32(newTime.Nanosecond())
		bigTime := secondsAndNanosToBigInt(secs, nano)
		bsi.SetBigValue(uint64(i), bigTime)
	}

	start, _ := bsi.GetBigValue(uint64(45)) // starting value at columnID 45
	end, _ := bsi.GetBigValue(uint64(55))   // ending value at columnID 55
	setStart := bsi.CompareBigValue(0, RANGE, nil, end, nil)
	tmpStart := bsi.CompareBigValue(0, RANGE, bsi.MinMaxBig(0, MIN, nil), end, nil)
	assert.Equal(t, tmpStart.GetCardinality(), setStart.GetCardinality())

	setEnd := bsi.CompareBigValue(0, RANGE, start, nil, nil)
	tmpEnd := bsi.CompareBigValue(0, RANGE, start, bsi.MinMaxBig(0, MAX, nil), nil)
	assert.Equal(t, tmpEnd.GetCardinality(), setEnd.GetCardinality())

	setAll := bsi.CompareBigValue(0, RANGE, nil, nil, nil)
	tmpAll := bsi.CompareBigValue(0, RANGE, bsi.MinMaxBig(0, MIN, nil), bsi.MinMaxBig(0, MAX, nil), nil)
	assert.Equal(t, tmpAll.GetCardinality(), setAll.GetCardinality())
}
