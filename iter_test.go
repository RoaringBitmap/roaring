package roaring

import (
	"testing"
	"time"

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

func collectRanges(b *Bitmap) ([][2]uint64, bool) {
	done := make(chan [][2]uint64, 1)
	go func() {
		var r [][2]uint64
		for s, e := range b.Ranges() {
			r = append(r, [2]uint64{uint64(s), e})
		}
		done <- r
	}()
	select {
	case r := <-done:
		return r, true
	case <-time.After(5 * time.Second):
		return nil, false
	}
}

func referenceRanges(b *Bitmap) [][2]uint64 {
	var result [][2]uint64
	it := b.Iterator()
	if !it.HasNext() {
		return nil
	}
	s := uint64(it.Next())
	e := s + 1
	for it.HasNext() {
		v := uint64(it.Next())
		if v == e {
			e++
		} else {
			result = append(result, [2]uint64{s, e})
			s = v
			e = v + 1
		}
	}
	return append(result, [2]uint64{s, e})
}

func checkRanges(t *testing.T, b *Bitmap) {
	t.Helper()
	got, ok := collectRanges(b)
	if !ok {
		t.Fatal("Ranges() hung")
	}
	assert.Equal(t, referenceRanges(b), got)
}

// 4097 elements at even offsets in [32768,65534] → bitmap container,
// lower half free for test-specific bits.
func bitmapContainerBitmap() *Bitmap {
	b := New()
	for i := uint32(0); i <= arrayDefaultMaxSize; i++ {
		b.Add(32768 + i*2)
	}
	return b
}

func addBits(b *Bitmap, bits ...uint32) {
	for _, v := range bits {
		b.Add(v)
	}
}

func addRange(b *Bitmap, lo, hi uint32) {
	for i := lo; i < hi; i++ {
		b.Add(i)
	}
}

func TestRanges(t *testing.T) {
	t.Run("array ranges", func(t *testing.T) {
		b := New()
		b.AddRange(5, 10)
		b.AddRange(20, 25)
		b.AddRange(100, 105)
		got, ok := collectRanges(b)
		assert.True(t, ok)
		assert.Equal(t, [][2]uint64{{5, 10}, {20, 25}, {100, 105}}, got)
	})

	t.Run("cross-container merge", func(t *testing.T) {
		b := New()
		b.AddRange(0xFFF0, 0x10010)
		got, _ := collectRanges(b)
		assert.Equal(t, [][2]uint64{{0xFFF0, 0x10010}}, got)
	})

	t.Run("scattered", func(t *testing.T) {
		b := New()
		addBits(b, 1, 3, 5)
		got, _ := collectRanges(b)
		assert.Equal(t, [][2]uint64{{1, 2}, {3, 4}, {5, 6}}, got)
	})

	t.Run("empty", func(t *testing.T) {
		got, ok := collectRanges(New())
		assert.True(t, ok)
		assert.Nil(t, got)
	})

	t.Run("break after 2", func(t *testing.T) {
		b := New()
		b.AddRange(0, 10)
		b.AddRange(20, 30)
		b.AddRange(40, 50)
		n := 0
		for range b.Ranges() {
			n++
			if n == 2 {
				break
			}
		}
		assert.Equal(t, 2, n)
	})
}

func TestRangesBitmapContainer(t *testing.T) {
	bc := bitmapContainerBitmap

	t.Run("baseline", func(t *testing.T) { checkRanges(t, bc()) })

	t.Run("bit 0", func(t *testing.T) {
		b := bc()
		b.Add(0)
		checkRanges(t, b)
	})
	t.Run("bit 63", func(t *testing.T) {
		b := bc()
		b.Add(63)
		checkRanges(t, b)
	})
	t.Run("bit 64", func(t *testing.T) {
		b := bc()
		b.Add(64)
		checkRanges(t, b)
	})
	t.Run("bits 0+63", func(t *testing.T) {
		b := bc()
		addBits(b, 0, 63)
		checkRanges(t, b)
	})
	t.Run("bits 0,1,63", func(t *testing.T) {
		b := bc()
		addBits(b, 0, 1, 63)
		checkRanges(t, b)
	})
	t.Run("non-adjacent same word", func(t *testing.T) {
		b := bc()
		addBits(b, 2, 5)
		checkRanges(t, b)
	})
	t.Run("adjacent same word", func(t *testing.T) {
		b := bc()
		addBits(b, 10, 11)
		checkRanges(t, b)
	})
	t.Run("short run in word", func(t *testing.T) {
		b := bc()
		addRange(b, 3, 7)
		checkRanges(t, b)
	})
	t.Run("full word", func(t *testing.T) {
		b := bc()
		addRange(b, 0, 64)
		checkRanges(t, b)
	})
	t.Run("span 2 words", func(t *testing.T) {
		b := bc()
		addRange(b, 60, 70)
		checkRanges(t, b)
	})
	t.Run("span 3 words", func(t *testing.T) {
		b := bc()
		addRange(b, 60, 130)
		checkRanges(t, b)
	})
	t.Run("3 runs in one word", func(t *testing.T) {
		b := bc()
		addBits(b, 0, 1, 10, 50, 51, 52)
		checkRanges(t, b)
	})
	t.Run("disjoint runs same word", func(t *testing.T) {
		b := bc()
		addBits(b, 0, 1, 10, 11, 62, 63)
		checkRanges(t, b)
	})
	t.Run("gap of 1 in word", func(t *testing.T) {
		b := bc()
		addBits(b, 0, 1, 2, 4, 5, 6) // gap at 3
		checkRanges(t, b)
	})
	t.Run("every 2nd bit word0", func(t *testing.T) {
		b := bc()
		for i := uint32(0); i < 64; i += 2 {
			b.Add(i)
		}
		checkRanges(t, b)
	})
	t.Run("every 3rd 4 words", func(t *testing.T) {
		b := bc()
		for i := uint32(0); i < 256; i += 3 {
			b.Add(i)
		}
		checkRanges(t, b)
	})
	t.Run("every 7th full container", func(t *testing.T) {
		b := New()
		for i := uint32(0); i < 65536; i += 7 {
			b.Add(i)
		}
		checkRanges(t, b)
	})
	t.Run("dense 60k", func(t *testing.T) {
		b := New()
		addRange(b, 0, 60000)
		checkRanges(t, b)
	})
	t.Run("full 64k", func(t *testing.T) {
		b := New()
		b.AddRange(0, 65536)
		checkRanges(t, b)
	})
	t.Run("bit 65535", func(t *testing.T) {
		b := bc()
		b.Add(65535)
		checkRanges(t, b)
	})
	t.Run("bits 63+64", func(t *testing.T) {
		b := bc()
		addBits(b, 63, 64)
		checkRanges(t, b)
	})
	t.Run("end at word boundary", func(t *testing.T) {
		b := bc()
		addRange(b, 56, 64)
		checkRanges(t, b)
	})
	t.Run("start at word boundary", func(t *testing.T) {
		b := bc()
		addRange(b, 64, 72)
		checkRanges(t, b)
	})
	t.Run("1 bit per word", func(t *testing.T) {
		b := bc()
		for w := 0; w < 100; w++ {
			b.Add(uint32(w*64 + 17))
		}
		checkRanges(t, b)
	})
	t.Run("bit 63 each word", func(t *testing.T) {
		b := bc()
		for w := 0; w < 100; w++ {
			b.Add(uint32(w*64 + 63))
		}
		checkRanges(t, b)
	})
	t.Run("bit 0 each word", func(t *testing.T) {
		b := bc()
		for w := 0; w < 100; w++ {
			b.Add(uint32(w * 64))
		}
		checkRanges(t, b)
	})
	t.Run("alternating full words", func(t *testing.T) {
		b := bc()
		for w := 0; w < 50; w += 2 {
			addRange(b, uint32(w*64), uint32(w*64+64))
		}
		checkRanges(t, b)
	})
	t.Run("bits 32-63", func(t *testing.T) {
		b := bc()
		addRange(b, 32, 64)
		checkRanges(t, b)
	})
	t.Run("high 4 bits per word", func(t *testing.T) {
		b := bc()
		for w := 0; w < 50; w++ {
			addRange(b, uint32(w*64+60), uint32(w*64+64))
		}
		checkRanges(t, b)
	})
	t.Run("low 4 bits per word", func(t *testing.T) {
		b := bc()
		for w := 0; w < 50; w++ {
			addRange(b, uint32(w*64), uint32(w*64+4))
		}
		checkRanges(t, b)
	})
	t.Run("4097 even", func(t *testing.T) {
		b := New()
		for i := uint32(0); i <= arrayDefaultMaxSize; i++ {
			b.Add(i * 2)
		}
		checkRanges(t, b)
	})
	t.Run("checkerboard 64k", func(t *testing.T) {
		b := New()
		for i := uint32(0); i < 65536; i += 2 {
			b.Add(i)
		}
		checkRanges(t, b)
	})
	t.Run("every 7th 202k multi-container", func(t *testing.T) {
		b := New()
		for i := uint32(0); i < 202240; i += 7 {
			b.Add(i)
		}
		checkRanges(t, b)
	})

	// cross-word paths
	t.Run("xword landing with remainder", func(t *testing.T) {
		b := bc()
		addBits(b, 62, 63, 64, 65, 70, 71)
		checkRanges(t, b)
	})
	t.Run("xword 3 full then partial", func(t *testing.T) {
		b := bc()
		addRange(b, 60, 260)
		checkRanges(t, b)
	})
	t.Run("two xword back to back", func(t *testing.T) {
		b := bc()
		addBits(b, 62, 63, 64, 65, 126, 127, 128, 129)
		checkRanges(t, b)
	})
	t.Run("xword chain through full word", func(t *testing.T) {
		b := bc()
		addBits(b, 62, 63)
		addRange(b, 64, 128)
		addBits(b, 128, 129)
		checkRanges(t, b)
	})
	t.Run("scattered across words", func(t *testing.T) {
		b := bc()
		addBits(b, 0, 64+33, 5*64+7, 10*64+63)
		checkRanges(t, b)
	})
	t.Run("4 consecutive full words", func(t *testing.T) {
		b := bc()
		addRange(b, 0, 256)
		checkRanges(t, b)
	})
	t.Run("every other word full", func(t *testing.T) {
		b := bc()
		for w := 0; w < 10; w += 2 {
			addRange(b, uint32(w*64), uint32(w*64+64))
		}
		checkRanges(t, b)
	})
	t.Run("exact word 1", func(t *testing.T) {
		b := bc()
		addRange(b, 64, 128)
		checkRanges(t, b)
	})
	t.Run("xword to container end", func(t *testing.T) {
		b := bc()
		addRange(b, 65526, 65536)
		checkRanges(t, b)
	})
	t.Run("last 3 bits", func(t *testing.T) {
		b := bc()
		addBits(b, 65533, 65534, 65535)
		checkRanges(t, b)
	})
	t.Run("xword landing zero remainder", func(t *testing.T) {
		b := bc()
		addBits(b, 5*64+62, 5*64+63, 6*64+0, 6*64+1)
		checkRanges(t, b)
	})
	t.Run("xword landing non-adjacent remainder", func(t *testing.T) {
		b := bc()
		addBits(b, 3*64+63, 4*64+0, 4*64+1, 4*64+2, 4*64+50)
		checkRanges(t, b)
	})
}

func TestRangesRunContainer(t *testing.T) {
	run := func(t *testing.T, setup func(b *Bitmap)) {
		t.Helper()
		b := New()
		setup(b)
		b.RunOptimize()
		checkRanges(t, b)
	}

	t.Run("single", func(t *testing.T) { run(t, func(b *Bitmap) { b.AddRange(100, 200) }) })
	t.Run("disjoint", func(t *testing.T) {
		run(t, func(b *Bitmap) { b.AddRange(10, 20); b.AddRange(50, 60); b.AddRange(100, 110) })
	})
	t.Run("adjacent merge", func(t *testing.T) {
		b := New()
		b.AddRange(10, 20)
		b.AddRange(20, 30)
		b.RunOptimize()
		got, _ := collectRanges(b)
		assert.Equal(t, [][2]uint64{{10, 30}}, got)
	})
	t.Run("full container", func(t *testing.T) { run(t, func(b *Bitmap) { b.AddRange(0, 65536) }) })
	t.Run("single value", func(t *testing.T) {
		b := New()
		b.Add(42)
		b.RunOptimize()
		got, _ := collectRanges(b)
		assert.Equal(t, [][2]uint64{{42, 43}}, got)
	})
	t.Run("container end", func(t *testing.T) { run(t, func(b *Bitmap) { b.AddRange(65530, 65536) }) })
	t.Run("container start", func(t *testing.T) { run(t, func(b *Bitmap) { b.AddRange(0, 10) }) })
	t.Run("many small", func(t *testing.T) {
		run(t, func(b *Bitmap) {
			for i := uint64(0); i < 1000; i += 5 {
				b.AddRange(i, i+2)
			}
		})
	})
}

func TestRangesArrayContainer(t *testing.T) {
	t.Run("single", func(t *testing.T) {
		b := New()
		b.Add(42)
		got, _ := collectRanges(b)
		assert.Equal(t, [][2]uint64{{42, 43}}, got)
	})
	t.Run("adjacent", func(t *testing.T) {
		b := New()
		addBits(b, 10, 11)
		got, _ := collectRanges(b)
		assert.Equal(t, [][2]uint64{{10, 12}}, got)
	})
	t.Run("non-adjacent", func(t *testing.T) {
		b := New()
		addBits(b, 10, 20)
		got, _ := collectRanges(b)
		assert.Equal(t, [][2]uint64{{10, 11}, {20, 21}}, got)
	})
	t.Run("contiguous 100", func(t *testing.T) {
		b := New()
		addRange(b, 100, 200)
		got, _ := collectRanges(b)
		assert.Equal(t, [][2]uint64{{100, 200}}, got)
	})
	t.Run("max array contiguous", func(t *testing.T) {
		b := New()
		addRange(b, 0, arrayDefaultMaxSize)
		got, _ := collectRanges(b)
		assert.Equal(t, [][2]uint64{{0, uint64(arrayDefaultMaxSize)}}, got)
	})
	t.Run("max array sparse", func(t *testing.T) {
		b := New()
		for i := uint32(0); i < arrayDefaultMaxSize; i++ {
			b.Add(i * 3)
		}
		checkRanges(t, b)
	})
	t.Run("zero", func(t *testing.T) {
		b := New()
		b.Add(0)
		got, _ := collectRanges(b)
		assert.Equal(t, [][2]uint64{{0, 1}}, got)
	})
	t.Run("65535 in container 1", func(t *testing.T) {
		b := New()
		b.Add(0x1FFFF)
		got, _ := collectRanges(b)
		assert.Equal(t, [][2]uint64{{0x1FFFF, 0x20000}}, got)
	})
}

func TestRangesMultiContainer(t *testing.T) {
	t.Run("no merge", func(t *testing.T) {
		b := New()
		b.AddRange(10, 20)
		b.AddRange(0x10010, 0x10020)
		checkRanges(t, b)
	})
	t.Run("merge at boundary", func(t *testing.T) {
		b := New()
		b.AddRange(0xFFF0, 0x10000)
		b.AddRange(0x10000, 0x10010)
		got, _ := collectRanges(b)
		assert.Equal(t, [][2]uint64{{0xFFF0, 0x10010}}, got)
	})
	t.Run("mixed types", func(t *testing.T) {
		b := New()
		addBits(b, 5, 10, 15)
		addRange(b, 0x10000, 0x10000+5000)
		b.AddRange(0x20000, 0x20100)
		b.RunOptimize()
		checkRanges(t, b)
	})
	t.Run("bitmap then array gap", func(t *testing.T) {
		b := New()
		addRange(b, 0, 5000)
		b.Add(0x10000)
		checkRanges(t, b)
	})
	t.Run("bitmap merge at boundary", func(t *testing.T) {
		b := New()
		for i := uint32(0); i < 4097; i++ {
			b.Add(i * 3)
		}
		addRange(b, 0xFFF0, 0x10000)
		for i := uint32(0x10000); i < 0x10000+4097; i++ {
			b.Add(i)
		}
		checkRanges(t, b)
	})
	t.Run("3 boundary merge", func(t *testing.T) {
		b := New()
		b.AddRange(0xFF00, 0x10000)
		b.AddRange(0x10000, 0x20000)
		b.AddRange(0x20000, 0x20100)
		got, _ := collectRanges(b)
		assert.Equal(t, [][2]uint64{{0xFF00, 0x20100}}, got)
	})
	t.Run("gap of 1", func(t *testing.T) {
		b := New()
		b.AddRange(0xFFF0, 0xFFFF)
		b.AddRange(0x10000, 0x10010)
		got, _ := collectRanges(b)
		assert.Equal(t, [][2]uint64{{0xFFF0, 0xFFFF}, {0x10000, 0x10010}}, got)
	})
	t.Run("high key bitmap merge", func(t *testing.T) {
		base := uint32(5) * 0x10000
		b := New()
		addRange(b, base, base+5000)
		b.AddRange(uint64(base)+0xFF00, uint64(base)+0x10000)
		b.AddRange(uint64(base)+0x10000, uint64(base)+0x10010)
		checkRanges(t, b)
	})
	t.Run("1 per container x100", func(t *testing.T) {
		b := New()
		for i := 0; i < 100; i++ {
			b.Add(uint32(i) * 0x10000)
		}
		got, _ := collectRanges(b)
		assert.Equal(t, 100, len(got))
		for i, r := range got {
			v := uint64(i) * 0x10000
			assert.Equal(t, [2]uint64{v, v + 1}, r)
		}
	})
	t.Run("sparse 10 containers", func(t *testing.T) {
		b := New()
		for c := 0; c < 10; c++ {
			base := uint32(c) * 0x10000
			addBits(b, base+100, base+200, base+300)
		}
		checkRanges(t, b)
	})
}

func TestRangesMaxUint32(t *testing.T) {
	M := uint64(MaxUint32)

	t.Run("single", func(t *testing.T) {
		b := New()
		b.Add(MaxUint32)
		got, _ := collectRanges(b)
		assert.Equal(t, [][2]uint64{{M, M + 1}}, got)
	})
	t.Run("range to max", func(t *testing.T) {
		b := New()
		b.AddRange(M-5, M+1)
		got, _ := collectRanges(b)
		assert.Equal(t, [][2]uint64{{M - 5, M + 1}}, got)
	})
	t.Run("adjacent at max", func(t *testing.T) {
		b := New()
		addBits(b, MaxUint32-1, MaxUint32)
		got, _ := collectRanges(b)
		assert.Equal(t, [][2]uint64{{M - 1, M + 1}}, got)
	})
	t.Run("0 and max", func(t *testing.T) {
		b := New()
		addBits(b, 0, MaxUint32)
		got, _ := collectRanges(b)
		assert.Equal(t, [][2]uint64{{0, 1}, {M, M + 1}}, got)
	})
	t.Run("full last container", func(t *testing.T) {
		base := uint64(0xFFFF) << 16
		b := New()
		b.AddRange(base, base+65536)
		got, _ := collectRanges(b)
		assert.Equal(t, [][2]uint64{{base, M + 1}}, got)
	})
	t.Run("last container bitmap sparse", func(t *testing.T) {
		base := uint32(0xFFFF) << 16
		b := New()
		for i := uint32(0); i < 5000; i++ {
			b.Add(base + i*3)
		}
		checkRanges(t, b)
	})
	t.Run("last container bitmap run at end", func(t *testing.T) {
		base := uint32(0xFFFF) << 16
		b := New()
		for i := uint32(0); i < 5000; i++ {
			b.Add(base + i*3)
		}
		for i := uint32(0); i < 10; i++ {
			b.Add(MaxUint32 - i)
		}
		checkRanges(t, b)
	})
	t.Run("merge last 2 containers", func(t *testing.T) {
		base := uint64(0xFFFE) << 16
		b := New()
		b.AddRange(base+0xFF00, base+0x10000)
		b.AddRange(base+0x10000, base+0x20000)
		got, _ := collectRanges(b)
		assert.Equal(t, [][2]uint64{{base + 0xFF00, M + 1}}, got)
	})
}

func TestRangesEarlyTermination(t *testing.T) {
	breakAfter := func(t *testing.T, b *Bitmap, n int) {
		t.Helper()
		count := 0
		for range b.Ranges() {
			count++
			if count == n {
				break
			}
		}
		assert.Equal(t, n, count)
	}

	t.Run("bitmap sparse", func(t *testing.T) {
		b := New()
		for i := uint32(0); i < 5000; i++ {
			b.Add(i * 3)
		}
		breakAfter(t, b, 1)
	})
	t.Run("bitmap dense", func(t *testing.T) {
		b := New()
		addRange(b, 0, 5000)
		breakAfter(t, b, 1)
	})
	t.Run("run", func(t *testing.T) {
		b := New()
		b.AddRange(0, 100)
		b.AddRange(200, 300)
		b.AddRange(400, 500)
		b.RunOptimize()
		breakAfter(t, b, 2)
	})
	t.Run("cross container", func(t *testing.T) {
		b := New()
		addBits(b, 5, 0x10005, 0x20005)
		breakAfter(t, b, 2)
	})
}
