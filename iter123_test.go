//go:build go1.23
// +build go1.23

package roaring

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBackwardCount123(t *testing.T) {
	array := []int{2, 63, 64, 65, 4095, 4096, 4097, 4159, 4160, 4161, 5000, 20000, 66666}
	for _, testSize := range array {
		b := New()
		for i := uint32(0); i < uint32(testSize); i++ {
			b.Add(i)
		}

		count := 0
		for range Values(b) {
			count++
		}

		assert.Equal(t, testSize, count)
	}
}

func TestBackward123(t *testing.T) {
	t.Run("#1", func(t *testing.T) {
		values := []uint32{0, 2, 15, 16, 31, 32, 33, 9999, MaxUint16, MaxUint32}
		b := New()
		for n := 0; n < len(values); n++ {
			b.Add(values[n])
		}
		n := len(values) - 1
		for val := range Backward(b) {
			assert.EqualValues(t, val, values[n])
			n--
		}
	})

	t.Run("#2", func(t *testing.T) {
		b := New()

		count := 0
		for range Backward(b) {
			count++
		}

		assert.Equal(t, 0, count)
	})

	t.Run("#3", func(t *testing.T) {
		b := New()
		b.AddInt(0)

		// only one value zero
		for val := range Backward(b) {
			assert.EqualValues(t, 0, val)
		}
	})

	t.Run("#4", func(t *testing.T) {
		b := New()
		b.AddInt(9999)

		// only one value 9999
		for val := range Backward(b) {
			assert.EqualValues(t, 9999, val)
		}
	})

	t.Run("#5", func(t *testing.T) {
		b := New()
		b.AddInt(MaxUint16)

		// only one value MaxUint16
		for val := range Backward(b) {
			assert.EqualValues(t, MaxUint16, val)
		}
	})

	t.Run("#6", func(t *testing.T) {
		b := New()
		b.AddInt(MaxUint32)

		// only one value MaxUint32
		for val := range Backward(b) {
			assert.EqualValues(t, MaxUint32, val)
		}
	})
}

func TestValues123(t *testing.T) {
	b := New()

	testSize := 5000
	for i := 0; i < testSize; i++ {
		b.AddInt(i)
	}

	n := 0
	for val := range Values(b) {
		assert.Equal(t, uint32(n), val)
		n++

	}

	assert.Equal(t, testSize, n)
}
