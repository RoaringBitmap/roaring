package roaring

// to run just these tests: go test -run TestSetUtil*

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetUtilDifference(t *testing.T) {
	data1 := []uint16{0, 1, 2, 3, 4, 9}
	data2 := []uint16{2, 3, 4, 5, 8, 9, 11}
	result := make([]uint16, 0, len(data1)+len(data2))
	expectedresult := []uint16{0, 1}
	nl := difference(data1, data2, result)
	result = result[:nl]

	assert.Equal(t, expectedresult, result)

	expectedresult = []uint16{5, 8, 11}
	nl = difference(data2, data1, result)
	result = result[:nl]

	assert.Equal(t, expectedresult, result)

	// empty set2
	data2 = []uint16{}
	expectedresult = []uint16{0, 1, 2, 3, 4, 9}
	nl = difference(data1, data2, result)
	result = result[:nl]

	assert.Equal(t, expectedresult, result)

	// empty set 1
	data1 = []uint16{}
	data2 = []uint16{2, 3, 4, 5, 8, 9, 11}
	expectedresult = []uint16{}
	nl = difference(data1, data2, result)
	result = result[:nl]

	assert.Equal(t, expectedresult, result)
}

func TestSetUtilUnion(t *testing.T) {
	data1 := []uint16{0, 1, 2, 3, 4, 9}
	data2 := []uint16{2, 3, 4, 5, 8, 9, 11}
	result := make([]uint16, 0, len(data1)+len(data2))
	expectedresult := []uint16{0, 1, 2, 3, 4, 5, 8, 9, 11}
	nl := union2by2(data1, data2, result)
	result = result[:nl]

	assert.Equal(t, expectedresult, result)

	nl = union2by2(data2, data1, result)
	result = result[:nl]

	assert.Equal(t, expectedresult, result)
}

func TestSetUtilExclusiveUnion(t *testing.T) {
	data1 := []uint16{0, 1, 2, 3, 4, 9}
	data2 := []uint16{2, 3, 4, 5, 8, 9, 11}
	result := make([]uint16, 0, len(data1)+len(data2))
	expectedresult := []uint16{0, 1, 5, 8, 11}
	nl := exclusiveUnion2by2(data1, data2, result)
	result = result[:nl]

	assert.Equal(t, expectedresult, result)

	nl = exclusiveUnion2by2(data2, data1, result)
	result = result[:nl]

	assert.Equal(t, expectedresult, result)
}

func TestSetUtilIntersection(t *testing.T) {
	data1 := []uint16{0, 1, 2, 3, 4, 9}
	data2 := []uint16{2, 3, 4, 5, 8, 9, 11}
	result := make([]uint16, 0, len(data1)+len(data2))
	expectedresult := []uint16{2, 3, 4, 9}
	nl := intersection2by2(data1, data2, result)
	result = result[:nl]
	result = result[:len(expectedresult)]

	assert.Equal(t, expectedresult, result)

	nl = intersection2by2(data2, data1, result)
	result = result[:nl]

	assert.Equal(t, expectedresult, result)

	data1 = []uint16{4}
	data2 = make([]uint16, 10000)

	for i := range data2 {
		data2[i] = uint16(i)
	}

	result = make([]uint16, 0, len(data1)+len(data2))
	expectedresult = data1
	nl = intersection2by2(data1, data2, result)
	result = result[:nl]

	assert.Equal(t, expectedresult, result)

	nl = intersection2by2(data2, data1, result)
	result = result[:nl]

	assert.Equal(t, expectedresult, result)
}

func TestSetUtilIntersection2(t *testing.T) {
	data1 := []uint16{0, 2, 4, 6, 8, 10, 12, 14, 16, 18}
	data2 := []uint16{0, 3, 6, 9, 12, 15, 18}
	result := make([]uint16, 0, len(data1)+len(data2))
	expectedresult := []uint16{0, 6, 12, 18}
	nl := intersection2by2(data1, data2, result)
	result = result[:nl]
	result = result[:len(expectedresult)]

	assert.Equal(t, expectedresult, result)
}

func TestSetUtilBinarySearch(t *testing.T) {
	data := make([]uint16, 256)
	for i := range data {
		data[i] = uint16(2 * i)
	}
	for i := 0; i < 2*len(data); i++ {
		key := uint16(i)
		loc := binarySearch(data, key)
		if (key & 1) == 0 {
			assert.Equal(t, int(key)/2, loc)
		} else {
			assert.Equal(t, -int(key)/2-2, loc)
		}
	}
}

func TestBinarySearchUntil(t *testing.T) {
	type searchTest struct {
		name          string
		targetSlice   []uint16
		target        uint16
		expectedValue uint16
		isExactMatch  bool
		expectedIndex int
	}

	tests := []searchTest{
		{
			"matches",
			[]uint16{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 13, 14, 15, 16, 17},
			9, 9, true, 9,
		},
		{
			"missing 12 with gap",
			[]uint16{0, 1, 2, 3, 4, 5, 6, 13, 14, 15, 16, 17},
			12, 6, false, 6,
		},
		{
			"missing 10 but close neighbors",
			[]uint16{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 11, 12},
			10, 9, false, 9,
		},
		{
			"missing close to beginning",
			[]uint16{0, 2, 3, 4, 5, 6, 7, 8, 9, 11, 12},
			1, 0, false, 0,
		},
		{
			"missing gap",
			[]uint16{0, 1, 2, 3, 4, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17},
			6, 4, false, 4,
		},
		{
			"out of bounds at beginning",
			[]uint16{1, 2, 3, 4, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17},
			0, 0, false, -1,
		},
		{
			"out of bounds at the end",
			[]uint16{0, 1, 2, 3, 4, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17},
			100, 0, false, 15,
		},
		{
			"missing alternating",
			[]uint16{0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 22, 24, 26},
			20, 18, false, 9,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := binarySearchUntil(testCase.targetSlice, testCase.target)
			assert.Equal(t, testCase.expectedIndex, result.index)
			assert.Equal(t, testCase.expectedIndex, result.index)
			assert.Equal(t, testCase.isExactMatch, result.exactMatch)
		})
	}
}

func TestBinarySearchPastWithBounds(t *testing.T) {
	type searchTest struct {
		name          string
		targetSlice   []uint16
		target        uint16
		expectedValue uint16
		isExactMatch  bool
		expectedIndex int
		low           int
		high          int
	}

	tests := []searchTest{
		{
			"has match but not in range",
			[]uint16{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 13, 14, 15, 16, 17},
			9, 0, false, 15, 0, 4,
		},
		{
			"matches",
			[]uint16{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 13, 14, 15, 16, 17},
			9, 9, true, 9, 0, 10,
		},
		{
			"missing 10-12 full range",
			[]uint16{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 13, 14, 15, 16, 17},
			12, uint16(13), false, 10, 0, 14,
		},
		{
			"has match but not in range",
			[]uint16{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 13, 14, 15, 16, 17},
			9, 0, false, 15, 0, 4,
		},
		{
			"missing 12 with gap",
			[]uint16{0, 1, 2, 3, 4, 5, 6, 13, 14, 15, 16, 17},
			12, 13, false, 7, 4, 11,
		},
		{
			"missing 10 but close neighbors",
			[]uint16{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 11, 12},
			10, 11, false, 10, 6, 11,
		},
		{
			"missing 10 out of range",
			[]uint16{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 11, 12},
			10, 0, false, 12, 0, 5,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := binarySearchPastWithBounds(testCase.targetSlice, testCase.target, testCase.low, testCase.high)
			assert.Equal(t, testCase.expectedIndex, result.index)
			assert.Equal(t, testCase.expectedValue, result.value)
			assert.Equal(t, testCase.isExactMatch, result.exactMatch)
		})
	}
}
