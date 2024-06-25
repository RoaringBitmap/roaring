package roaring

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoaringArrayAdvanceUntil(t *testing.T) {
	bitmap := New()
	low := 1 << 16
	mid := 2 << 16
	high := 3 << 16
	bitmap.AddRange(uint64(low)-1, uint64(low)+2)
	bitmap.AddRange(uint64(mid)-1, uint64(mid)+2)
	bitmap.AddRange(uint64(high)-1, uint64(high)+2)

	assert.Equal(t, 0, bitmap.highlowcontainer.advanceUntil(0, -1))
	assert.Equal(t, 1, bitmap.highlowcontainer.advanceUntil(1, -1))
	assert.Equal(t, 2, bitmap.highlowcontainer.advanceUntil(2, -1))
	assert.Equal(t, 3, bitmap.highlowcontainer.advanceUntil(3, -1))
	assert.Equal(t, 4, bitmap.highlowcontainer.advanceUntil(4, -1))

	assert.Equal(t, 1, bitmap.highlowcontainer.advanceUntil(0, 0))
	assert.Equal(t, 2, bitmap.highlowcontainer.advanceUntil(1, 1))
	assert.Equal(t, 3, bitmap.highlowcontainer.advanceUntil(2, 2))
	assert.Equal(t, 4, bitmap.highlowcontainer.advanceUntil(3, 3))
	assert.Equal(t, 5, bitmap.highlowcontainer.advanceUntil(4, 4))
}
