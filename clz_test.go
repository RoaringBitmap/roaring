package roaring

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func numberOfLeadingZeros(i uint64) int {
	if i == 0 {
		return 64
	}
	n := 1
	x := uint32(i >> 32)
	if x == 0 {
		n += 32
		x = uint32(i)
	}
	if (x >> 16) == 0 {
		n += 16
		x <<= 16
	}
	if (x >> 24) == 0 {
		n += 8
		x <<= 8
	}
	if x>>28 == 0 {
		n += 4
		x <<= 4
	}
	if x>>30 == 0 {
		n += 2
		x <<= 2

	}
	n -= int(x >> 31)
	return n
}

func TestCountLeadingZeros072(t *testing.T) {
	assert.Equal(t, 64, numberOfLeadingZeros(0))
	assert.Equal(t, 60, numberOfLeadingZeros(8))
	assert.Equal(t, 64-17-1, numberOfLeadingZeros(1<<17))
	assert.Equal(t, 0, numberOfLeadingZeros(0xFFFFFFFFFFFFFFFF))
	assert.Equal(t, 64, countLeadingZeros(0))
	assert.Equal(t, 60, countLeadingZeros(8))
	assert.Equal(t, 64-17-1, countLeadingZeros(1<<17))
	assert.Equal(t, 0, countLeadingZeros(0xFFFFFFFFFFFFFFFF))
}
