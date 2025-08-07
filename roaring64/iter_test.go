package roaring64

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBackwardCount(t *testing.T) {
	array := []int{2, 63, 64, 65, 4095, 4096, 4097, 4159, 4160, 4161, 5000, 20000, 66666}
	for _, testSize := range array {
		b := New()
		for i := uint64(0); i < uint64(testSize); i++ {
			b.Add(i)
		}
		it := Values(b)

		count := 0
		it(func(_ uint64) bool {
			count++
			return true
		})

		assert.Equal(t, testSize, count)
	}
}

func TestBackward(t *testing.T) {
	t.Run("#1", func(t *testing.T) {
		values := []uint64{0, 2, 15, 16, 31, 32, 33, 9999, math.MaxUint16}
		b := New()
		for n := 0; n < len(values); n++ {
			b.Add(values[n])
		}
		it := Backward(b)
		n := len(values) - 1

		it(func(val uint64) bool {
			assert.EqualValues(t, val, values[n])
			n--
			return true
		})

		it = Backward(b)
		n = len(values) - 1
		it(func(val uint64) bool {
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
		it(func(_ uint64) bool {
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
		it(func(val uint64) bool {
			assert.EqualValues(t, 0, val)
			return true
		})
	})

	t.Run("#4", func(t *testing.T) {
		b := New()
		b.AddInt(9999)
		it := Backward(b)

		// only one value 9999
		it(func(val uint64) bool {
			assert.EqualValues(t, 9999, val)
			return true
		})
	})

	t.Run("#5", func(t *testing.T) {
		b := New()
		b.AddInt(math.MaxUint16)
		it := Values(b)

		// only one value MaxUint16
		it(func(val uint64) bool {
			assert.EqualValues(t, math.MaxUint16, val)
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
	it(func(val uint64) bool {
		assert.Equal(t, uint64(n), val)
		n++
		return true
	})

	assert.Equal(t, testSize, n)
}
