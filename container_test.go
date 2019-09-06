package roaring

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func makeContainer(ss []uint16) container {
	c := newArrayContainer()
	for _, s := range ss {
		c.iadd(s)
	}
	return c
}

func checkContent(c container, s []uint16) bool {
	si := c.getShortIterator()
	ctr := 0
	fail := false
	for si.hasNext() {
		if ctr == len(s) {
			log.Println("HERE")
			fail = true
			break
		}
		i := si.next()
		if i != s[ctr] {

			log.Println("THERE", i, s[ctr])
			fail = true
			break
		}
		ctr++
	}
	if ctr != len(s) {
		log.Println("LAST")
		fail = true
	}
	if fail {
		log.Println("fail, found ")
		si = c.getShortIterator()
		z := 0
		for si.hasNext() {
			si.next()
			z++
		}
		log.Println(z, len(s))
	}

	return !fail
}

func testContainerIteratorPeekNext(t *testing.T, c container) {
	testSize := 5000
	for i := 0; i < testSize; i++ {
		c.iadd(uint16(i))
	}

	i := c.getShortIterator()
	assert.True(t, i.hasNext())

	for i.hasNext() {
		assert.Equal(t, i.peekNext(), i.next())
		testSize--
	}

	assert.Equal(t, 0, testSize)
}

func testContainerIteratorAdvance(t *testing.T, con container) {
	values := []uint16{1, 2, 15, 16, 31, 32, 33, 9999}
	for _, v := range values {
		con.iadd(v)
	}

	cases := []struct {
		minval   uint16
		expected uint16
	}{
		{0, 1},
		{1, 1},
		{2, 2},
		{3, 15},
		{15, 15},
		{30, 31},
		{31, 31},
		{33, 33},
		{34, 9999},
		{9998, 9999},
		{9999, 9999},
	}

	t.Run("advance by using a new short iterator", func(t *testing.T) {
		for _, c := range cases {
			i := con.getShortIterator()
			i.advanceIfNeeded(c.minval)

			assert.True(t, i.hasNext())
			assert.Equal(t, c.expected, i.peekNext())
		}
	})

	t.Run("advance by using the same short iterator", func(t *testing.T) {
		i := con.getShortIterator()

		for _, c := range cases {
			i.advanceIfNeeded(c.minval)

			assert.True(t, i.hasNext())
			assert.Equal(t, c.expected, i.peekNext())
		}
	})

	t.Run("advance out of a container value", func(t *testing.T) {
		i := con.getShortIterator()

		i.advanceIfNeeded(33)
		assert.True(t, i.hasNext())
		assert.EqualValues(t, 33, i.peekNext())

		i.advanceIfNeeded(MaxUint16 - 1)
		assert.False(t, i.hasNext())

		i.advanceIfNeeded(MaxUint16)
		assert.False(t, i.hasNext())
	})

	t.Run("advance on a value that is less than the pointed value", func(t *testing.T) {
		i := con.getShortIterator()
		i.advanceIfNeeded(29)
		assert.True(t, i.hasNext())
		assert.EqualValues(t, 31, i.peekNext())

		i.advanceIfNeeded(13)
		assert.True(t, i.hasNext())
		assert.EqualValues(t, 31, i.peekNext())
	})
}

func benchmarkContainerIteratorAdvance(b *testing.B, con container) {
	for _, initsize := range []int{1, 650, 6500, MaxUint16} {
		for i := 0; i < initsize; i++ {
			con.iadd(uint16(i))
		}

		b.Run(fmt.Sprintf("init size %d shortIterator advance", initsize), func(b *testing.B) {
			b.StartTimer()
			diff := uint16(0)

			for n := 0; n < b.N; n++ {
				val := uint16(n % initsize)

				i := con.getShortIterator()
				i.advanceIfNeeded(val)

				diff += i.peekNext() - val
			}

			b.StopTimer()

			if diff != 0 {
				b.Fatalf("Expected diff 0, got %d", diff)
			}
		})
	}
}

func benchmarkContainerIteratorNext(b *testing.B, con container) {
	for _, initsize := range []int{1, 650, 6500, MaxUint16} {
		for i := 0; i < initsize; i++ {
			con.iadd(uint16(i))
		}

		b.Run(fmt.Sprintf("init size %d shortIterator next", initsize), func(b *testing.B) {
			b.StartTimer()
			diff := 0

			for n := 0; n < b.N; n++ {
				i := con.getShortIterator()
				j := 0

				for i.hasNext() {
					i.next()
					j++
				}

				diff += j - initsize
			}

			b.StopTimer()

			if diff != 0 {
				b.Fatalf("Expected diff 0, got %d", diff)
			}
		})
	}
}

func TestContainerReverseIterator(t *testing.T) {
	content := []uint16{1, 3, 5, 7, 9}
	c := makeContainer(content)
	si := c.getReverseIterator()
	i := 4

	for si.hasNext() {
		assert.Equal(t, content[i], si.next())
		i--
	}

	assert.Equal(t, -1, i)
}

func TestRoaringContainer(t *testing.T) {
	t.Run("countTrailingZeros", func(t *testing.T) {
		x := uint64(0)
		o := countTrailingZeros(x)
		assert.Equal(t, 64, o)

		x = 1 << 3
		o = countTrailingZeros(x)
		assert.Equal(t, 3, o)
	})

	t.Run("ArrayShortIterator", func(t *testing.T) {
		content := []uint16{1, 3, 5, 7, 9}
		c := makeContainer(content)
		si := c.getShortIterator()
		i := 0
		for si.hasNext() {
			si.next()
			i++
		}

		assert.Equal(t, 5, i)
	})

	t.Run("BinarySearch", func(t *testing.T) {
		content := []uint16{1, 3, 5, 7, 9}
		res := binarySearch(content, 5)
		assert.Equal(t, 2, res)

		res = binarySearch(content, 4)
		assert.Less(t, res, 0)
	})

	t.Run("bitmapcontainer", func(t *testing.T) {
		content := []uint16{1, 3, 5, 7, 9}
		a := newArrayContainer()
		b := newBitmapContainer()
		for _, v := range content {
			a.iadd(v)
			b.iadd(v)
		}
		c := a.toBitmapContainer()

		assert.Equal(t, b.getCardinality(), a.getCardinality())
		assert.Equal(t, b.getCardinality(), c.getCardinality())
	})

	t.Run("inottest0", func(t *testing.T) {
		content := []uint16{9}
		c := makeContainer(content)
		c = c.inot(0, 11)
		si := c.getShortIterator()
		i := 0
		for si.hasNext() {
			si.next()
			i++
		}

		assert.Equal(t, 10, i)
	})

	t.Run("inotTest1", func(t *testing.T) {
		// Array container, range is complete
		content := []uint16{1, 3, 5, 7, 9}
		//content := []uint16{1}
		edge := 1 << 13
		c := makeContainer(content)
		c = c.inot(0, edge+1)
		size := edge - len(content)
		s := make([]uint16, size+1)
		pos := 0
		for i := uint16(0); i < uint16(edge+1); i++ {
			if binarySearch(content, i) < 0 {
				s[pos] = i
				pos++
			}
		}

		assert.True(t, checkContent(c, s))
	})
}
