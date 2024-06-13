package roaring64

import (
	"math"
	"testing"

	"github.com/RoaringBitmap/roaring/v2"
	"github.com/stretchr/testify/assert"
)

func TestReverseIteratorCount(t *testing.T) {
	array := []int{2, 63, 64, 65, 4095, 4096, 4097, 4159, 4160, 4161, 5000, 20000, 66666, 140140}
	for _, testSize := range array {
		b := New()
		for i := uint64(0); i < uint64(testSize); i++ {
			b.Add(i)
		}
		it := b.ReverseIterator()
		count := 0
		for it.HasNext() {
			it.Next()
			count++
		}

		assert.Equal(t, testSize, count)
	}
}

func TestReverseIterator(t *testing.T) {
	t.Run("#1", func(t *testing.T) {
		values := []uint64{0, 2, 15, 16, 31, 32, 33, 9999, roaring.MaxUint16, roaring.MaxUint32, roaring.MaxUint32 * 2, math.MaxUint64}
		bm := New()
		for n := 0; n < len(values); n++ {
			bm.Add(values[n])
		}
		i := bm.ReverseIterator()
		n := len(values) - 1

		for i.HasNext() {
			assert.EqualValues(t, i.Next(), values[n])
			n--
		}

		// HasNext() was terminating early - add test
		i = bm.ReverseIterator()
		n = len(values) - 1
		for ; n >= 0; n-- {
			assert.EqualValues(t, i.Next(), values[n])
			assert.False(t, n > 0 && !i.HasNext())
		}
	})

	t.Run("#2", func(t *testing.T) {
		bm := New()
		i := bm.ReverseIterator()

		assert.False(t, i.HasNext())
	})

	t.Run("#3", func(t *testing.T) {
		bm := New()
		bm.AddInt(0)
		i := bm.ReverseIterator()

		assert.True(t, i.HasNext())
		assert.EqualValues(t, 0, i.Next())
		assert.False(t, i.HasNext())
	})

	t.Run("#4", func(t *testing.T) {
		bm := New()
		bm.AddInt(9999)
		i := bm.ReverseIterator()

		assert.True(t, i.HasNext())
		assert.EqualValues(t, 9999, i.Next())
		assert.False(t, i.HasNext())
	})

	t.Run("#5", func(t *testing.T) {
		bm := New()
		bm.AddInt(roaring.MaxUint16)
		i := bm.ReverseIterator()

		assert.True(t, i.HasNext())
		assert.EqualValues(t, roaring.MaxUint16, i.Next())
		assert.False(t, i.HasNext())
	})

	t.Run("#6", func(t *testing.T) {
		bm := New()
		bm.Add(roaring.MaxUint32)
		i := bm.ReverseIterator()

		assert.True(t, i.HasNext())
		assert.EqualValues(t, uint32(roaring.MaxUint32), i.Next())
		assert.False(t, i.HasNext())
	})
}

func TestIteratorPeekNext(t *testing.T) {
	values := []uint64{0, 2, 15, 16, 31, 32, 33, 9999, roaring.MaxUint16, roaring.MaxUint32, roaring.MaxUint32 * 2, math.MaxUint64}
	bm := New()

	for n := 0; n < len(values); n++ {
		bm.Add(values[n])
	}

	i := bm.Iterator()
	assert.True(t, i.HasNext())

	for i.HasNext() {
		assert.Equal(t, i.PeekNext(), i.Next())
	}
}

func TestIteratorAdvance(t *testing.T) {
	values := []uint64{1, 2, 15, 16, 31, 32, 33, 9999, roaring.MaxUint16}
	bm := New()

	for n := 0; n < len(values); n++ {
		bm.Add(values[n])
	}

	cases := []struct {
		minval   uint64
		expected uint64
	}{
		{0, 1},
		{1, 1},
		{2, 2},
		{3, 15},
		{30, 31},
		{33, 33},
		{9998, 9999},
		{roaring.MaxUint16, roaring.MaxUint16},
	}

	t.Run("advance by using a new int iterator", func(t *testing.T) {
		for _, c := range cases {
			i := bm.Iterator()
			i.AdvanceIfNeeded(c.minval)

			assert.True(t, i.HasNext())
			assert.Equal(t, c.expected, i.PeekNext())
		}
	})

	t.Run("advance by using the same int iterator", func(t *testing.T) {
		i := bm.Iterator()

		for _, c := range cases {
			i.AdvanceIfNeeded(c.minval)

			assert.True(t, i.HasNext())
			assert.Equal(t, c.expected, i.PeekNext())
		}
	})

	t.Run("advance out of a container value", func(t *testing.T) {
		i := bm.Iterator()

		i.AdvanceIfNeeded(roaring.MaxUint32)
		assert.False(t, i.HasNext())

		i.AdvanceIfNeeded(roaring.MaxUint32)
		assert.False(t, i.HasNext())
	})

	t.Run("advance on a value that is less than the pointed value", func(t *testing.T) {
		i := bm.Iterator()
		i.AdvanceIfNeeded(29)

		assert.True(t, i.HasNext())
		assert.EqualValues(t, 31, i.PeekNext())

		i.AdvanceIfNeeded(13)

		assert.True(t, i.HasNext())
		assert.EqualValues(t, 31, i.PeekNext())
	})
}
