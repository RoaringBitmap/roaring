package roaring

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)

// bitmapContainer's numberOfRuns() function should be correct against the runContainer equivalent
func TestBitmapContainerNumberOfRuns024(t *testing.T) {
	seed := int64(42)
	rand.Seed(seed)

	trials := []trial{
		{n: 1000, percentFill: .1, ntrial: 10},
	}

	for _, tr := range trials {
		for j := 0; j < tr.ntrial; j++ {
			ma := make(map[int]bool)

			n := tr.n
			a := []uint16{}

			draw := int(float64(n) * tr.percentFill)
			for i := 0; i < draw; i++ {
				r0 := rand.Intn(n)
				a = append(a, uint16(r0))
				ma[r0] = true
			}

			// RunContainer compute this automatically
			rc := newRunContainer16FromVals(false, a...)
			rcNr := rc.numberOfRuns()

			// vs bitmapContainer
			bc := newBitmapContainer()
			for k := range ma {
				bc.iadd(uint16(k))
			}

			bcNr := bc.numberOfRuns()
			assert.Equal(t, rcNr, bcNr)
		}
	}
}

// bitmap containers get cardinality in range, miss the last index, issue #183
func TestBitmapcontainerAndCardinality(t *testing.T) {
	for r := 0; r <= 65535; r++ {
		c1 := newRunContainer16Range(0, uint16(r))
		c2 := newBitmapContainerwithRange(0, int(r))

		assert.Equal(t, r+1, c1.andCardinality(c2))
	}
}

func TestIssue181(t *testing.T) {
	t.Run("Initial issue 181", func(t *testing.T) {
		a := New()
		var x uint32

		// adding 1M integers
		for i := 1; i <= 1000000; i++ {
			x += uint32(rand.Intn(10) + 1)
			a.Add(x)
		}
		b := New()
		for i := 1; i <= int(x); i++ {
			b.Add(uint32(i))
		}

		assert.Equal(t, b.AndCardinality(a), a.AndCardinality(b))
		assert.Equal(t, b.AndCardinality(a), And(a, b).GetCardinality())
	})

	t.Run("Second version of issue 181", func(t *testing.T) {
		a := New()
		var x uint32

		// adding 1M integers
		for i := 1; i <= 1000000; i++ {
			x += uint32(rand.Intn(10) + 1)
			a.Add(x)
		}
		b := New()
		b.AddRange(1, uint64(x))

		assert.Equal(t, b.AndCardinality(a), a.AndCardinality(b))
		assert.Equal(t, b.AndCardinality(a), And(a, b).GetCardinality())
	})
}

// RunReverseIterator16 unit tests for cur, next, hasNext, and remove should pass
func TestBitmapContainerReverseIterator(t *testing.T) {
	t.Run("reverse iterator on the empty container", func(t *testing.T) {
		bc := newBitmapContainer()
		it := bc.getReverseIterator()

		assert.False(t, it.hasNext())
		assert.Panics(t, func() { it.next() })
	})

	t.Run("reverse iterator on the container with range(0,0)", func(t *testing.T) {
		bc := newBitmapContainerwithRange(0, 0)
		it := bc.getReverseIterator()

		assert.True(t, it.hasNext())
		assert.EqualValues(t, 0, it.next())
	})

	t.Run("reverse iterator on the container with range(4,4)", func(t *testing.T) {
		bc := newBitmapContainerwithRange(4, 4)
		it := bc.getReverseIterator()

		assert.True(t, it.hasNext())
		assert.EqualValues(t, 4, it.next())
	})

	t.Run("reverse iterator on the container with range(4,9)", func(t *testing.T) {
		bc := newBitmapContainerwithRange(4, 9)
		it := bc.getReverseIterator()

		assert.True(t, it.hasNext())

		for i := 9; i >= 4; i-- {
			assert.EqualValues(t, i, it.next())

			if i > 4 {
				assert.True(t, it.hasNext())
			} else if i == 4 {
				assert.False(t, it.hasNext())
			}
		}

		assert.False(t, it.hasNext())
		assert.Panics(t, func() { it.next() })
	})

	t.Run("reverse iterator on the container with values", func(t *testing.T) {
		values := []uint16{0, 2, 15, 16, 31, 32, 33, 9999, MaxUint16}
		bc := newBitmapContainer()

		for n := 0; n < len(values); n++ {
			bc.iadd(values[n])
		}

		it := bc.getReverseIterator()
		n := len(values)

		assert.True(t, it.hasNext())

		for it.hasNext() {
			n--
			assert.Equal(t, values[n], it.next())
		}

		assert.Equal(t, 0, n)
	})
}

func TestBitmapNextSet(t *testing.T) {
	testSize := 5000
	bc := newBitmapContainer()

	for i := 0; i < testSize; i++ {
		bc.iadd(uint16(i))
	}

	m := 0

	for n := 0; m < testSize; n, m = bc.NextSetBit(n+1), m+1 {
		assert.Equal(t, m, n)
	}

	assert.Equal(t, 5000, m)
}

func TestBitmapPrevSet(t *testing.T) {
	testSize := 5000
	bc := newBitmapContainer()

	for i := 0; i < testSize; i++ {
		bc.iadd(uint16(i))
	}

	m := testSize - 1

	for n := testSize - 1; n > 0; n, m = bc.PrevSetBit(n-1), m-1 {
		assert.Equal(t, m, n)
	}

	assert.Equal(t, 0, m)
}

func TestBitmapIteratorPeekNext(t *testing.T) {
	testContainerIteratorPeekNext(t, newBitmapContainer())
}

func TestBitmapIteratorAdvance(t *testing.T) {
	testContainerIteratorAdvance(t, newBitmapContainer())
}

// go test -bench BenchmarkShortIteratorAdvance -run -
func BenchmarkShortIteratorAdvanceBitmap(b *testing.B) {
	benchmarkContainerIteratorAdvance(b, newBitmapContainer())
}

// go test -bench BenchmarkShortIteratorNext -run -
func BenchmarkShortIteratorNextBitmap(b *testing.B) {
	benchmarkContainerIteratorNext(b, newBitmapContainer())
}

func TestBitmapOffset(t *testing.T) {
	nums := []uint16{10, 100, 1000}
	expected := make([]int, len(nums))
	offtest := uint16(65000)
	v := container(newBitmapContainer())
	for i, n := range nums {
		v.iadd(n)
		expected[i] = int(n) + int(offtest)
	}
	w := v.addOffset(offtest)
	w0card := w[0].getCardinality()
	w1card := w[1].getCardinality()

	assert.Equal(t, 3, w0card+w1card)

	wout := make([]int, len(nums))
	for i := 0; i < w0card; i++ {
		wout[i] = w[0].selectInt(uint16(i))
	}
	for i := 0; i < w1card; i++ {
		wout[i+w0card] = w[1].selectInt(uint16(i)) + 65536
	}

	for i, x := range wout {
		assert.Equal(t, expected[i], x)
	}
}

func TestBitmapContainerResetTo(t *testing.T) {
	array := newArrayContainer()
	for i := 0; i < 1000; i++ {
		array.iadd(uint16(i*1000 + i + 50))
	}

	bitmap := newBitmapContainer()
	for i := 0; i < 10000; i++ {
		bitmap.iadd(uint16(i*1000 + i + 50))
	}

	run := newRunContainer16()
	for i := 0; i < 10; i++ {
		start := i*1000 + i + 50
		run.iaddRange(start, start+100+i)
	}

	makeDirty := func() *bitmapContainer {
		ret := newBitmapContainer()
		for i := 0; i < maxCapacity; i += 42 {
			ret.iadd(uint16(i))
		}
		return ret
	}

	t.Run("to array container", func(t *testing.T) {
		clean := newBitmapContainer()
		clean.resetTo(array)
		assert.True(t, clean.toArrayContainer().equals(array))

		dirty := makeDirty()
		dirty.resetTo(array)
		assert.True(t, dirty.toArrayContainer().equals(array))
	})

	t.Run("to bitmap container", func(t *testing.T) {
		clean := newBitmapContainer()
		clean.resetTo(bitmap)
		assert.True(t, clean.equals(bitmap))

		dirty := makeDirty()
		dirty.resetTo(bitmap)
		assert.True(t, dirty.equals(bitmap))
	})

	t.Run("to run container", func(t *testing.T) {
		clean := newBitmapContainer()
		clean.resetTo(run)
		assert.EqualValues(t, clean.cardinality, run.cardinality())
		assert.True(t, clean.toEfficientContainer().equals(run))

		dirty := makeDirty()
		dirty.resetTo(run)
		assert.EqualValues(t, dirty.cardinality, run.cardinality())
		assert.True(t, dirty.toEfficientContainer().equals(run))
	})
}
