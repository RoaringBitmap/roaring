package roaring64

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoaringArray64AdvanceUntil(t *testing.T) {
	bitmap := New()
	low := uint64(1) << 32
	mid := uint64(2) << 32
	high := uint64(3) << 32
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

func TestCopies(t *testing.T) {
	tests := []struct {
		name string
		cow1 bool
		cow2 bool
	}{
		{"AppendCopiesAfterCoW", true, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r1 := uint64(1) << 32
			r2 := uint64(2) << 32
			r3 := uint64(3) << 32
			r4 := uint64(4) << 32

			bitmap1 := New()
			bitmap2 := New()
			bitmap1.SetCopyOnWrite(test.cow1)
			bitmap2.SetCopyOnWrite(test.cow2)

			bitmap2.AddRange(uint64(r1)-1, uint64(r1)+2)
			bitmap2.AddRange(uint64(r2)-1, uint64(r2)+2)
			bitmap2.AddRange(uint64(r3)-1, uint64(r3)+2)
			bitmap2.AddRange(uint64(r4)-1, uint64(r4)+2)

			assert.False(t, bitmap1.Contains(uint64(r2)))
			assert.False(t, bitmap1.Contains(uint64(r3)))
			assert.Equal(t, 0, len(bitmap1.highlowcontainer.keys))
			bitmap1.highlowcontainer.appendCopiesAfter(bitmap2.highlowcontainer, 0)
			assert.Equal(t, 4, len(bitmap1.highlowcontainer.keys))
			assert.True(t, bitmap1.Contains(uint64(r2)))
			assert.True(t, bitmap1.Contains(uint64(r3)))

			for idx1, c1 := range bitmap1.highlowcontainer.containers {
				for idx2, c2 := range bitmap2.highlowcontainer.containers {
					// idx+1 is required because appendCopiesAfter starts at key 1
					if idx1+1 == idx2 {
						if test.cow1 && test.cow2 {
							assert.True(t, c1 == c2)
						} else {
							assert.False(t, c1 == c2)
						}
					}
				}
			}
		})
	}

	tests = []struct {
		name string
		cow1 bool
		cow2 bool
	}{
		{"AppendCopiesUntilCoW", true, true},
		{"AppendCopiesUntilCoW", false, false},
		{"AppendCopiesUntilMixedCoW", true, false},
		{"AppendCopiesUntilMixedCoW", false, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r1 := uint64(1) << 32
			r2 := uint64(2) << 32
			r3 := uint64(3) << 32
			r4 := uint64(4) << 32
			key := highbits(uint64(r4))

			bitmap1 := New()
			bitmap2 := New()
			bitmap1.SetCopyOnWrite(test.cow1)
			bitmap2.SetCopyOnWrite(test.cow2)

			bitmap2.AddRange(uint64(r1)-1, uint64(r1)+2)
			bitmap2.AddRange(uint64(r2)-1, uint64(r2)+2)
			bitmap2.AddRange(uint64(r3)-1, uint64(r3)+2)
			bitmap2.AddRange(uint64(r4)-1, uint64(r4)+2)

			assert.False(t, bitmap1.Contains(uint64(r2)))
			assert.False(t, bitmap1.Contains(uint64(r3)))
			assert.Equal(t, 0, len(bitmap1.highlowcontainer.keys))
			bitmap1.highlowcontainer.appendCopiesUntil(bitmap2.highlowcontainer, key)
			assert.Equal(t, 4, len(bitmap1.highlowcontainer.keys))
			assert.True(t, bitmap1.Contains(uint64(r2)))
			assert.True(t, bitmap1.Contains(uint64(r3)))

			for idx1, c1 := range bitmap1.highlowcontainer.containers {
				for idx2, c2 := range bitmap2.highlowcontainer.containers {
					if idx1 == idx2 {
						if test.cow1 && test.cow2 {
							assert.True(t, c1 == c2)
						} else {
							assert.False(t, c1 == c2)
						}
					}
				}
			}
		})
	}
}
