package roaring

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoaringArrayAdvanceUntil(t *testing.T) {
	bitmap := New()
	bitmap.AddRange(0, 128)
	assert.Equal(t, 0, bitmap.highlowcontainer.advanceUntil(0, -1))
	assert.Equal(t, 1, bitmap.highlowcontainer.advanceUntil(0, 0))
	assert.Equal(t, 65, bitmap.highlowcontainer.advanceUntil(64, 0))
}

func TestRoaringArrayAdvanceUntilJavaRegression(t *testing.T) {
	bitmap := New()
	bitmap.AddMany([]uint32{0, 3, 16, 18, 21, 29, 30})
	assert.Equal(t, 1, bitmap.highlowcontainer.advanceUntil(3, -1))
	assert.Equal(t, 5, bitmap.highlowcontainer.advanceUntil(28, -1))
	assert.Equal(t, 5, bitmap.highlowcontainer.advanceUntil(29, -1))
}
