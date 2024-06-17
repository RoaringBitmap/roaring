package roaring64

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoaringArray64AdvanceUntil(t *testing.T) {
	bitmap := New()
	bitmap.AddRange(0, 128)
	assert.Equal(t, 0, bitmap.highlowcontainer.advanceUntil(0, -1))
	assert.Equal(t, 1, bitmap.highlowcontainer.advanceUntil(0, 0))
	assert.Equal(t, 65, bitmap.highlowcontainer.advanceUntil(64, 0))
}

func TestRoaringArray64AdvanceUntilJavaRegression(t *testing.T) {
	bitmap := New()
	bitmap.AddMany([]uint64{0, 3, 16, 18, 21, 29, 30})
	assert.Equal(t, 1, bitmap.highlowcontainer.advanceUntil(3, -1))
	assert.Equal(t, 5, bitmap.highlowcontainer.advanceUntil(28, -1))
	assert.Equal(t, 5, bitmap.highlowcontainer.advanceUntil(29, -1))
}
