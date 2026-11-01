package roaring

import (
	"math"
	"testing"

	"github.com/RoaringBitmap/roaring/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// identityBSI sets column x to value x for x in [1, 100), like Java's RBBsiTest.
func identityBSI() *BSI {
	bsi := NewDefaultBSI()
	for x := uint64(1); x < 100; x++ {
		bsi.SetValue(x, int64(x))
	}
	return bsi
}

func rangeArray(start, end uint32) []uint32 {
	out := make([]uint32, 0, end-start)
	for v := start; v < end; v++ {
		out = append(out, v)
	}
	return out
}

// Java RBBsiTest.testValueZero: value 0 is distinct from "absent".
func TestBSIValueZero(t *testing.T) {
	bsi := NewDefaultBSI()
	bsi.SetValue(0, 0)
	bsi.SetValue(1, 0)
	bsi.SetValue(2, 1)
	assert.Equal(t, []uint32{0, 1}, bsi.CompareValue(0, EQ, 0, 0, nil).ToArray())
	assert.Equal(t, []uint32{2}, bsi.CompareValue(0, EQ, 1, 0, nil).ToArray())
	v, ok := bsi.GetValue(0)
	assert.True(t, ok)
	assert.EqualValues(t, 0, v)
	v, ok = bsi.GetValue(3)
	assert.False(t, ok)
	assert.EqualValues(t, 0, v)
}

// Java RBBsiTest.testIssue753: RANGE bounds outside [min, max].
func TestBSIRangeOutsideBounds(t *testing.T) {
	bsi := identityBSI()
	cases := []struct {
		start, end int64
		card       uint64
	}{
		{-4, 56, 56}, {-4, 129, 99}, {-4, 200, 99}, {-4, 20000, 99}, {-4, -129, 0}, {-4, -2, 0},
		{4, 56, 53}, {4, 129, 96}, {4, 200, 96}, {4, 20000, 96}, {4, -129, 0}, {4, 2, 0},
		{-129, -14, 0}, {129, 2000, 0},
		{1, 200, 99}, {1000, 2000, 0}, {10, 20, 11},
	}
	for _, c := range cases {
		got := bsi.CompareValue(0, RANGE, c.start, c.end, nil)
		assert.Equal(t, c.card, got.GetCardinality(), "RANGE(%d, %d)", c.start, c.end)
	}
	// other comparisons with a negative value on a BSI without negative values
	assert.True(t, bsi.CompareValue(0, LT, -4, 0, nil).IsEmpty())
	assert.True(t, bsi.CompareValue(0, LE, -4, 0, nil).IsEmpty())
	assert.True(t, bsi.CompareValue(0, EQ, -4, 0, nil).IsEmpty())
	assert.EqualValues(t, 99, bsi.CompareValue(0, GE, -4, 0, nil).GetCardinality())
	assert.EqualValues(t, 99, bsi.CompareValue(0, GT, -4, 0, nil).GetCardinality())
	found := roaring.BitmapOf(5, 6, 50)
	assert.Equal(t, []uint32{5, 6, 50}, bsi.CompareValue(0, GT, -1, 0, found).ToArray())
	assert.Equal(t, []uint32{5, 6}, bsi.CompareValue(0, RANGE, -1, 6, found).ToArray())
}

// Java RBBsiTest.testAdd and testAddAndEvaluate: Add of BSIs with different column sets.
func TestBSIAddDifferentColumnSets(t *testing.T) {
	a := identityBSI()
	b := NewDefaultBSI()
	for x := uint64(1); x < 120; x++ {
		b.SetValue(x, int64(x))
	}
	a.Add(b)
	for x := uint64(1); x < 120; x++ {
		v, ok := a.GetValue(x)
		require.True(t, ok, "column %d", x)
		if x < 100 {
			require.EqualValues(t, 2*x, v, "column %d", x)
		} else {
			require.EqualValues(t, x, v, "column %d", x)
		}
	}

	a = identityBSI()
	b = NewDefaultBSI()
	for x := uint64(1); x < 120; x++ {
		b.SetValue(120-x, int64(x))
	}
	a.Add(b)
	assert.Equal(t, rangeArray(1, 100), a.CompareValue(0, EQ, 120, 0, nil).ToArray())
	assert.Equal(t, rangeArray(100, 120), a.CompareValue(0, RANGE, 1, 20, nil).ToArray())
}

// Java RBBsiTest.testIssue743, TestIO4Stream: serialization round trip.
func TestBSIMarshalRoundTrip(t *testing.T) {
	bsi := NewDefaultBSI()
	bsi.SetValue(100, 3)
	bsi.SetValue(1, 392)
	data, err := bsi.MarshalBinary()
	require.NoError(t, err)
	got := NewDefaultBSI()
	require.NoError(t, got.UnmarshalBinary(data))
	v, ok := got.GetValue(100)
	assert.True(t, ok)
	assert.EqualValues(t, 3, v)
	v, ok = got.GetValue(1)
	assert.True(t, ok)
	assert.EqualValues(t, 392, v)
	assert.EqualValues(t, 2, got.GetCardinality())

	bsi = identityBSI()
	data, err = bsi.MarshalBinary()
	require.NoError(t, err)
	got = NewDefaultBSI()
	require.NoError(t, got.UnmarshalBinary(data))
	assert.EqualValues(t, 99, got.GetExistenceBitmap().GetCardinality())
	for x := uint64(1); x < 100; x++ {
		v, ok := got.GetValue(x)
		require.True(t, ok)
		require.EqualValues(t, x, v)
	}
	assert.Equal(t, rangeArray(10, 21), got.CompareValue(0, RANGE, 10, 20, nil).ToArray())
}

// Java RBBsiTest.testGT/GE/LT/LE: comparisons at the extremes, with and without foundSet.
func TestBSICompareExtremes(t *testing.T) {
	bsi := identityBSI()
	all := rangeArray(1, 100)
	check := func(op Operation, v int64, found *roaring.Bitmap, expected []uint32) {
		t.Helper()
		got := bsi.CompareValue(0, op, v, 0, found).ToArray()
		if len(expected) == 0 {
			assert.Empty(t, got, "op %d value %d", op, v)
		} else {
			assert.Equal(t, expected, got, "op %d value %d", op, v)
		}
	}
	check(GT, 50, nil, rangeArray(51, 100))
	check(GT, 0, nil, all)
	check(GT, 99, nil, nil)
	check(GE, 50, nil, rangeArray(50, 100))
	check(GE, 1, nil, all)
	check(GE, 100, nil, nil)
	check(LT, 50, nil, rangeArray(1, 50))
	check(LT, 1, nil, nil)
	check(LT, math.MaxInt32, nil, all)
	check(LE, 50, nil, rangeArray(1, 51))
	check(LE, 0, nil, nil)
	check(LE, math.MaxInt32, nil, all)

	found := roaring.BitmapOf(51, 52, 53)
	check(GE, 50, found, []uint32{51, 52, 53})
	check(GE, 0, found, []uint32{51, 52, 53})
	check(GE, math.MaxInt32, found, nil)
	found = roaring.BitmapOf(1, 2, 3)
	check(LE, 50, found, []uint32{1, 2, 3})
	check(LE, math.MaxInt32, found, []uint32{1, 2, 3})
	check(LE, 0, found, nil)
}

// Sum: Go's Sum(nil) sums over every column (Java's sum(null) returns (0, 0);
// this is a known divergence). The count is the cardinality of the foundSet,
// even for columns that are absent from the BSI, as in Java.
func TestBSISumSemantics(t *testing.T) {
	bsi := identityBSI()
	sum, count := bsi.Sum(nil)
	assert.EqualValues(t, 99*100/2, sum)
	assert.EqualValues(t, 99, count)
	sum, count = bsi.Sum(roaring.New())
	assert.EqualValues(t, 0, sum)
	assert.EqualValues(t, 0, count)

	bsi = NewDefaultBSI()
	bsi.SetValue(1, 7)
	sum, count = bsi.Sum(roaring.BitmapOf(1, 1000))
	assert.EqualValues(t, 7, sum)
	assert.EqualValues(t, 2, count)
}
