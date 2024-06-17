package roaring

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestManyIterator(t *testing.T) {
	type searchTest struct {
		name          string
		iterator      shortIterator
		high          uint64
		buf           []uint64
		expectedValue int
	}

	tests := []searchTest{
		{
			"no values",
			shortIterator{},
			uint64(1024),
			[]uint64{},
			0,
		},
		{
			"1 value ",
			shortIterator{[]uint16{uint16(1)}, 0},
			uint64(1024),
			make([]uint64, 1),
			1,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			iterator := testCase.iterator
			result := iterator.nextMany64(testCase.high, testCase.buf)
			assert.Equal(t, testCase.expectedValue, result)
		})
	}
}
