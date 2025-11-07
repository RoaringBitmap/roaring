package roaring

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBackwardCount(t *testing.T) {
	array := []int{2, 63, 64, 65, 4095, 4096, 4097, 4159, 4160, 4161, 5000, 20000, 66666}
	for _, testSize := range array {
		b := New()
		for i := uint32(0); i < uint32(testSize); i++ {
			b.Add(i)
		}
		it := Values(b)

		count := 0
		it(func(_ uint32) bool {
			count++
			return true
		})

		assert.Equal(t, testSize, count)
	}
}

func TestBackward(t *testing.T) {
	t.Run("#1", func(t *testing.T) {
		values := []uint32{0, 2, 15, 16, 31, 32, 33, 9999, MaxUint16}
		b := New()
		for n := 0; n < len(values); n++ {
			b.Add(values[n])
		}
		it := Backward(b)
		n := len(values) - 1

		it(func(val uint32) bool {
			assert.EqualValues(t, val, values[n])
			n--
			return true
		})

		it = Backward(b)
		n = len(values) - 1
		it(func(val uint32) bool {
			assert.EqualValues(t, val, values[n])
			assert.True(t, n >= 0)
			n--
			return true
		})
	})

	t.Run("#2", func(t *testing.T) {
		b := New()
		it := Backward(b)

		count := 0
		it(func(_ uint32) bool {
			count++
			return true
		})

		assert.Equal(t, 0, count)
	})

	t.Run("#3", func(t *testing.T) {
		b := New()
		b.AddInt(0)
		it := Backward(b)

		// only one value zero
		it(func(val uint32) bool {
			assert.EqualValues(t, 0, val)
			return true
		})
	})

	t.Run("#4", func(t *testing.T) {
		b := New()
		b.AddInt(9999)
		it := Backward(b)

		// only one value 9999
		it(func(val uint32) bool {
			assert.EqualValues(t, 9999, val)
			return true
		})
	})

	t.Run("#5", func(t *testing.T) {
		b := New()
		b.AddInt(MaxUint16)
		it := Values(b)

		// only one value MaxUint16
		it(func(val uint32) bool {
			assert.EqualValues(t, MaxUint16, val)
			return true
		})
	})
}

func TestValues(t *testing.T) {
	b := New()

	testSize := 5000
	for i := 0; i < testSize; i++ {
		b.AddInt(i)
	}

	it := Values(b)
	n := 0
	it(func(val uint32) bool {
		assert.Equal(t, uint32(n), val)
		n++
		return true
	})

	assert.Equal(t, testSize, n)
}

func TestUnset(t *testing.T) {
	t.Run("empty bitmap", func(t *testing.T) {
		b := New()
		it := Unset(b, 5, 10)

		expected := []uint32{5, 6, 7, 8, 9, 10}
		actual := make([]uint32, 0)

		it(func(val uint32) bool {
			actual = append(actual, val)
			return true
		})

		assert.Equal(t, expected, actual)
	})

	t.Run("bitmap with some values set", func(t *testing.T) {
		b := New()
		b.AddInt(3)
		b.AddInt(7)
		b.AddInt(12)

		it := Unset(b, 5, 10)

		expected := []uint32{5, 6, 8, 9, 10}
		actual := make([]uint32, 0)

		it(func(val uint32) bool {
			actual = append(actual, val)
			return true
		})

		assert.Equal(t, expected, actual)
	})

	t.Run("range completely outside bitmap", func(t *testing.T) {
		b := New()
		b.AddInt(1)
		b.AddInt(2)
		b.AddInt(3)

		it := Unset(b, 10, 15)

		expected := []uint32{10, 11, 12, 13, 14, 15}
		actual := make([]uint32, 0)

		it(func(val uint32) bool {
			actual = append(actual, val)
			return true
		})

		assert.Equal(t, expected, actual)
	})

	t.Run("range includes set and unset values", func(t *testing.T) {
		b := New()
		b.AddInt(5)
		b.AddInt(8)
		b.AddInt(9)

		it := Unset(b, 3, 12)

		expected := []uint32{3, 4, 6, 7, 10, 11, 12}
		actual := make([]uint32, 0)

		it(func(val uint32) bool {
			actual = append(actual, val)
			return true
		})

		assert.Equal(t, expected, actual)
	})

	t.Run("min greater than max", func(t *testing.T) {
		b := New()
		it := Unset(b, 10, 5)

		count := 0
		it(func(val uint32) bool {
			count++
			return true
		})

		assert.Equal(t, 0, count)
	})

	t.Run("single value range - unset", func(t *testing.T) {
		b := New()
		b.AddInt(5)

		it := Unset(b, 3, 3)

		expected := []uint32{3}
		actual := make([]uint32, 0)

		it(func(val uint32) bool {
			actual = append(actual, val)
			return true
		})

		assert.Equal(t, expected, actual)
	})

	t.Run("single value range - set", func(t *testing.T) {
		b := New()
		b.AddInt(5)

		it := Unset(b, 5, 5)

		count := 0
		it(func(val uint32) bool {
			count++
			return true
		})

		assert.Equal(t, 0, count)
	})

	t.Run("early termination", func(t *testing.T) {
		b := New()

		it := Unset(b, 1, 10)

		actual := make([]uint32, 0)
		it(func(val uint32) bool {
			actual = append(actual, val)
			return len(actual) < 3 // Stop after 3 values
		})

		expected := []uint32{1, 2, 3}
		assert.Equal(t, expected, actual)
	})

	t.Run("large range with sparse bitmap", func(t *testing.T) {
		b := New()
		b.AddInt(100)
		b.AddInt(500)
		b.AddInt(1000)

		it := Unset(b, 50, 150)

		actual := make([]uint32, 0)
		it(func(val uint32) bool {
			actual = append(actual, val)
			return true
		})

		// Should include all values from 50-150 except 100
		assert.Equal(t, 100, len(actual)) // 150-50+1-1 = 100
		assert.Contains(t, actual, uint32(50))
		assert.Contains(t, actual, uint32(99))
		assert.NotContains(t, actual, uint32(100))
		assert.Contains(t, actual, uint32(101))
		assert.Contains(t, actual, uint32(150))
	})

	t.Run("min is in the bitmap", func(t *testing.T) {
		b := New()
		b.AddInt(100)

		it := Unset(b, 100, 105)

		actual := make([]uint32, 0)
		it(func(val uint32) bool {
			actual = append(actual, val)
			return true
		})
		expected := []uint32{101, 102, 103, 104, 105}

		assert.Equal(t, expected, actual)
	})

	t.Run("extreme max", func(t *testing.T) {
		b := New()
		b.Add(4294967295)

		it := Unset(b, 4294967294, 4294967295)

		actual := make([]uint32, 0)
		it(func(val uint32) bool {
			actual = append(actual, val)
			return true
		})
		expected := []uint32{4294967294}

		assert.Equal(t, expected, actual)
	})
}

func TestUnsetIteratorPeekable(t *testing.T) {
	t.Run("peek next", func(t *testing.T) {
		b := New()
		b.AddInt(5)
		b.AddInt(8)

		it := b.UnsetIterator(3, 11)

		// First value should be 3
		assert.True(t, it.HasNext())
		assert.Equal(t, uint32(3), it.PeekNext())
		assert.Equal(t, uint32(3), it.Next())

		// Next should be 4
		assert.True(t, it.HasNext())
		assert.Equal(t, uint32(4), it.PeekNext())
		assert.Equal(t, uint32(4), it.Next())

		// Next should be 6 (skipping 5 which is set)
		assert.True(t, it.HasNext())
		assert.Equal(t, uint32(6), it.PeekNext())
		assert.Equal(t, uint32(6), it.Next())

		// Next should be 7
		assert.True(t, it.HasNext())
		assert.Equal(t, uint32(7), it.PeekNext())
		assert.Equal(t, uint32(7), it.Next())

		// Next should be 9 (skipping 8 which is set)
		assert.True(t, it.HasNext())
		assert.Equal(t, uint32(9), it.PeekNext())
		assert.Equal(t, uint32(9), it.Next())

		// Next should be 10
		assert.True(t, it.HasNext())
		assert.Equal(t, uint32(10), it.PeekNext())
		assert.Equal(t, uint32(10), it.Next())

		// No more values
		assert.False(t, it.HasNext())
	})

	t.Run("advance if needed", func(t *testing.T) {
		b := New()
		b.AddInt(5)
		b.AddInt(8)
		b.AddInt(12)

		it := b.UnsetIterator(1, 16)

		// Skip to values >= 7
		it.AdvanceIfNeeded(7)

		// Should now be at 7 (skipping 5 which is set)
		assert.True(t, it.HasNext())
		assert.Equal(t, uint32(7), it.PeekNext())
		assert.Equal(t, uint32(7), it.Next())

		// Next should be 9 (skipping 8 which is set)
		assert.True(t, it.HasNext())
		assert.Equal(t, uint32(9), it.PeekNext())
		assert.Equal(t, uint32(9), it.Next())

		// Skip to values >= 11
		it.AdvanceIfNeeded(11)

		// Should now be at 11 (skipping 12 which is set)
		assert.True(t, it.HasNext())
		assert.Equal(t, uint32(11), it.PeekNext())
		assert.Equal(t, uint32(11), it.Next())

		// Next should be 13
		assert.True(t, it.HasNext())
		assert.Equal(t, uint32(13), it.PeekNext())
		assert.Equal(t, uint32(13), it.Next())

		// Skip beyond range
		it.AdvanceIfNeeded(20)
		assert.False(t, it.HasNext())
	})

	t.Run("advance if needed before range", func(t *testing.T) {
		b := New()
		b.AddInt(5)

		it := b.UnsetIterator(10, 16)

		// Try to advance to a value before our range start
		it.AdvanceIfNeeded(5)

		// Should still start from 10
		assert.True(t, it.HasNext())
		assert.Equal(t, uint32(10), it.PeekNext())
	})

	t.Run("advance if needed beyond range", func(t *testing.T) {
		b := New()
		b.AddInt(5)

		it := b.UnsetIterator(10, 16)

		// Advance beyond our range
		it.AdvanceIfNeeded(20)

		// Should have no more values
		assert.False(t, it.HasNext())
	})

	t.Run("advance if needed on current value", func(t *testing.T) {
		b := New()
		b.AddRange(0, 0x10000)
		iter := b.UnsetIterator(0, 0x10003)
		var got []uint32
		prev := uint32(0)
		for len(got) < 10 {
			iter.AdvanceIfNeeded(prev)
			if !iter.HasNext() {
				break
			}
			x := iter.Next()
			got = append(got, x)
			prev = x
		}
		assert.Equal(t, []uint32{0x10000, 0x10001, 0x10002}, got)
	})

	t.Run("peek next on empty iterator", func(t *testing.T) {
		b := New()
		b.AddInt(5) // Set bit in middle of range

		it := b.UnsetIterator(5, 6) // Range contains only the set bit

		// Should have no values
		assert.False(t, it.HasNext())

		// PeekNext should panic when HasNext is false
		assert.Panics(t, func() {
			it.PeekNext()
		})
	})

	t.Run("range including max uint32 unset", func(t *testing.T) {
		b := New()
		b.Add(4294967294) // Set the value before max

		it := b.UnsetIterator(4294967294, 4294967296)

		// Should have 4294967295 (max uint32) as it's unset
		assert.True(t, it.HasNext())
		assert.Equal(t, uint32(4294967295), it.PeekNext())
		assert.Equal(t, uint32(4294967295), it.Next())

		// No more values
		assert.False(t, it.HasNext())
	})

	t.Run("max uint32 set", func(t *testing.T) {
		b := New()
		b.Add(4294967295) // Set max uint32

		it := b.UnsetIterator(4294967294, 4294967296)

		// Should have 4294967294 as it's unset, but not 4294967295
		assert.True(t, it.HasNext())
		assert.Equal(t, uint32(4294967294), it.PeekNext())
		assert.Equal(t, uint32(4294967294), it.Next())

		// No more values
		assert.False(t, it.HasNext())
	})
}
