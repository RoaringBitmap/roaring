package roaring

// to run just these tests: go test -run TestArrayContainer*

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArrayContainerTransition(t *testing.T) {
	v := container(newArrayContainer())

	for i := 0; i < arrayDefaultMaxSize; i++ {
		v = v.iaddReturnMinimized(uint16(i))
	}

	assert.Equal(t, arrayDefaultMaxSize, v.getCardinality())
	assert.IsType(t, newArrayContainer(), v)

	for i := 0; i < arrayDefaultMaxSize; i++ {
		v = v.iaddReturnMinimized(uint16(i))
	}

	assert.Equal(t, arrayDefaultMaxSize, v.getCardinality())
	assert.IsType(t, newArrayContainer(), v)

	v = v.iaddReturnMinimized(uint16(arrayDefaultMaxSize))

	assert.Equal(t, arrayDefaultMaxSize+1, v.getCardinality())
	assert.IsType(t, newBitmapContainer(), v)

	v = v.iremoveReturnMinimized(uint16(arrayDefaultMaxSize))

	assert.Equal(t, arrayDefaultMaxSize, v.getCardinality())
	assert.IsType(t, newArrayContainer(), v)
}

func TestArrayContainerRank(t *testing.T) {
	v := container(newArrayContainer())
	v = v.iaddReturnMinimized(10)
	v = v.iaddReturnMinimized(100)
	v = v.iaddReturnMinimized(1000)

	assert.Equal(t, 3, v.getCardinality())

	for i := 0; i <= arrayDefaultMaxSize; i++ {
		thisrank := v.rank(uint16(i))

		if i < 10 {
			assert.Equalf(t, 0, thisrank, "At %d should be zero but is %d", i, thisrank)
		} else if i < 100 {
			assert.Equalf(t, 1, thisrank, "At %d should be one but is %d", i, thisrank)
		} else if i < 1000 {
			assert.Equalf(t, 2, thisrank, "At %d should be two but is %d", i, thisrank)
		} else {
			assert.Equalf(t, 3, thisrank, "At %d should be three but is %d", i, thisrank)
		}
	}
}

func TestArrayOffset(t *testing.T) {
	nums := []uint16{10, 100, 1000}
	expected := make([]int, len(nums))
	offtest := uint16(65000)
	v := container(newArrayContainer())
	for i, n := range nums {
		v = v.iaddReturnMinimized(n)
		expected[i] = int(n) + int(offtest)
	}
	l, h := v.addOffset(offtest)

	var w0card, w1card int
	wout := make([]int, len(nums))

	if l != nil {
		w0card = l.getCardinality()

		for i := 0; i < w0card; i++ {
			wout[i] = l.selectInt(uint16(i))
		}
	}

	if h != nil {
		w1card = h.getCardinality()

		for i := 0; i < w1card; i++ {
			wout[i+w0card] = h.selectInt(uint16(i)) + 65536
		}
	}

	assert.Equal(t, 3, w0card+w1card)
	for i, x := range wout {
		assert.Equal(t, expected[i], x)
	}
}

func TestArrayContainerMassiveSetAndGet(t *testing.T) {
	v := container(newArrayContainer())

	for j := 0; j <= arrayDefaultMaxSize; j++ {
		v = v.iaddReturnMinimized(uint16(j))
		assert.Equal(t, 1+j, v.getCardinality())

		success := true
		i := 0

		for ; i <= arrayDefaultMaxSize && success; i++ {
			if i <= j {
				success = v.contains(uint16(i))
			} else {
				success = !v.contains(uint16(i))
			}
		}

		assert.Truef(t, success, "failed at %d iteration", i)
	}
}

func TestArrayContainerUnsupportedType(t *testing.T) {
	a := container(newArrayContainer())
	testContainerPanics(t, a)

	b := container(newBitmapContainer())
	testContainerPanics(t, b)
}

func testContainerPanics(t *testing.T, c container) {
	f := &struct {
		arrayContainer
	}{}

	assert.Panics(t, func() { c.or(f) })
	assert.Panics(t, func() { c.ior(f) })
	assert.Panics(t, func() { c.lazyIOR(f) })
	assert.Panics(t, func() { c.lazyOR(f) })
	assert.Panics(t, func() { c.and(f) })
	assert.Panics(t, func() { c.intersects(f) })
	assert.Panics(t, func() { c.iand(f) })
	assert.Panics(t, func() { c.xor(f) })
	assert.Panics(t, func() { c.andNot(f) })
	assert.Panics(t, func() { c.iandNot(f) })
}

func TestArrayContainerNumberOfRuns025(t *testing.T) {
	seed := int64(42)
	rand.Seed(seed)

	trials := []trial{
		{n: 1000, percentFill: .1, ntrial: 10},
		/*
			trial{n: 100, percentFill: .5, ntrial: 10},
			trial{n: 100, percentFill: .01, ntrial: 10},
			trial{n: 100, percentFill: .99, ntrial: 10},
		*/
	}

	tester := func(tr trial) {
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

			// RunContainer computes this automatically
			rc := newRunContainer16FromVals(false, a...)
			rcNr := rc.numberOfRuns()

			// vs arrayContainer
			ac := newArrayContainer()
			for k := range ma {
				ac.iadd(uint16(k))
			}

			acNr := ac.numberOfRuns()
			assert.Equal(t, acNr, rcNr)

			// get coverage of arrayContainer coners...
			assert.Equal(t, 2*len(ma), ac.serializedSizeInBytes())
			assert.NotPanics(t, func() { ac.iaddRange(2, 1) })
			assert.NotPanics(t, func() { ac.iremoveRange(2, 1) })

			ac.iremoveRange(0, 2)
			ac.iremoveRange(0, 2)
			delete(ma, 0)
			delete(ma, 1)

			assert.Equal(t, len(ma), ac.getCardinality())

			ac.iadd(0)
			ac.iadd(1)
			ac.iadd(2)
			ma[0] = true
			ma[1] = true
			ma[2] = true
			newguy := ac.not(0, 3).(*arrayContainer)

			assert.False(t, newguy.contains(0))
			assert.False(t, newguy.contains(1))
			assert.False(t, newguy.contains(2))

			newguy.notClose(0, 2)
			newguy.remove(2)
			newguy.remove(2)
			newguy.ior(ac)

			messedUp := newArrayContainer()

			assert.Equal(t, 0, messedUp.numberOfRuns())

			// messed up
			messedUp.content = []uint16{1, 1}
			assert.Panics(t, func() { messedUp.numberOfRuns() })

			messedUp.content = []uint16{2, 1}
			assert.Panics(t, func() { messedUp.numberOfRuns() })

			shouldBeBit := newArrayContainer()
			for i := 0; i < arrayDefaultMaxSize+1; i++ {
				shouldBeBit.iadd(uint16(i * 2))
			}
			bit := shouldBeBit.toEfficientContainer()
			_, isBit := bit.(*bitmapContainer)

			assert.True(t, isBit)
		}
	}

	for i := range trials {
		tester(trials[i])
	}
}

func TestArrayContainerIaddRangeNearMax068(t *testing.T) {
	iv := []interval16{
		newInterval16Range(65525, 65527),
		newInterval16Range(65530, 65530),
		newInterval16Range(65534, 65535),
	}
	rc := newRunContainer16TakeOwnership(iv)

	ac2 := rc.toArrayContainer()

	assert.True(t, ac2.equals(rc))
	assert.True(t, rc.equals(ac2))

	ac := newArrayContainer()
	endx := int(MaxUint16) + 1
	first := endx - 3
	ac.iaddRange(first-20, endx-20)
	ac.iaddRange(first-6, endx-6)
	ac.iaddRange(first, endx)

	assert.Equal(t, 9, ac.getCardinality())
}

func TestArrayContainerEtc070(t *testing.T) {
	iv := []interval16{
		newInterval16Range(65525, 65527),
		newInterval16Range(65530, 65530),
		newInterval16Range(65534, 65535),
	}
	rc := newRunContainer16TakeOwnership(iv)
	ac := rc.toArrayContainer()

	// not when nothing to do just returns a clone
	assert.True(t, ac.equals(ac.not(0, 0)))
	assert.True(t, ac.equals(ac.notClose(1, 0)))

	// not will promote to bitmapContainer if card is big enough
	ac = newArrayContainer()
	ac.inot(0, MaxUint16+1)
	rc = newRunContainer16Range(0, MaxUint16)

	assert.True(t, rc.equals(ac))

	// comparing two array containers with different card
	ac2 := newArrayContainer()
	assert.False(t, ac2.equals(ac))

	// comparing two arrays with same card but different content
	ac3 := newArrayContainer()
	ac4 := newArrayContainer()
	ac3.iadd(1)
	ac3.iadd(2)
	ac4.iadd(1)

	assert.False(t, ac3.equals(ac4))

	// compare array vs other with different card
	assert.False(t, ac3.equals(rc))

	// compare array vs other, same card, different content
	rc = newRunContainer16Range(0, 0)
	assert.False(t, ac4.equals(rc))

	// remove from middle of array
	ac5 := newArrayContainer()
	ac5.iaddRange(0, 10)

	assert.Equal(t, 10, ac5.getCardinality())

	ac6 := ac5.remove(5)
	assert.Equal(t, 9, ac6.getCardinality())

	// lazyorArray that converts to bitmap
	ac5.iaddRange(0, arrayLazyLowerBound-1)
	ac6.iaddRange(arrayLazyLowerBound, 2*arrayLazyLowerBound-2)
	ac6a := ac6.(*arrayContainer)
	bc := ac5.lazyorArray(ac6a)
	_, isBitmap := bc.(*bitmapContainer)

	assert.True(t, isBitmap)

	// andBitmap
	ac = newArrayContainer()
	ac.iaddRange(0, 10)
	bc9 := newBitmapContainer()
	bc9.iaddRange(0, 5)
	and := ac.andBitmap(bc9)

	assert.Equal(t, 5, and.getCardinality())

	// numberOfRuns with 1 member
	ac10 := newArrayContainer()
	ac10.iadd(1)

	assert.Equal(t, 1, ac10.numberOfRuns())
}

func TestArrayContainerIAndNot(t *testing.T) {
	var ac container
	ac = newArrayContainer()
	ac.iadd(12)
	ac.iadd(27)
	ac.iadd(32)
	ac.iadd(88)
	ac.iadd(188)
	ac.iadd(289)

	var rc container
	rc = newRunContainer16Range(0, 15)
	rc = rc.iaddRange(1500, 2000)
	rc = rc.iaddRange(55, 100)
	rc = rc.iaddRange(25, 50)
	ac = ac.iandNot(rc)

	require.ElementsMatch(t, []uint16{188, 289}, ac.(*arrayContainer).content)
	require.Equal(t, 2, ac.getCardinality())
}

func TestArrayContainerIand(t *testing.T) {
	a := NewBitmap()
	a.AddRange(0, 200000)
	b := BitmapOf(50, 100000, 150000)
	b.And(a)
	r := b.ToArray()

	assert.Len(t, r, 3)
	assert.EqualValues(t, 50, r[0])
	assert.EqualValues(t, 100000, r[1])
	assert.EqualValues(t, 150000, r[2])
}

func TestArrayIteratorPeekNext(t *testing.T) {
	testContainerIteratorPeekNext(t, newArrayContainer())
}

func TestArrayIteratorAdvance(t *testing.T) {
	testContainerIteratorAdvance(t, newArrayContainer())
}

func TestArrayContainerResetTo(t *testing.T) {
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

	makeDirty := func() *arrayContainer {
		ret := newArrayContainer()
		for i := 0; i < arrayDefaultMaxSize; i += 3 {
			ret.iadd(uint16(i))
		}
		return ret
	}

	t.Run("to array container", func(t *testing.T) {
		clean := newArrayContainer()
		clean.resetTo(array)
		assert.True(t, clean.equals(array))

		dirty := makeDirty()
		dirty.resetTo(array)
		assert.True(t, dirty.equals(array))
	})

	t.Run("to bitmap container", func(t *testing.T) {
		clean := newArrayContainer()
		clean.resetTo(bitmap)
		assert.True(t, clean.equals(bitmap))

		dirty := makeDirty()
		dirty.resetTo(bitmap)
		assert.True(t, dirty.equals(bitmap.toArrayContainer()))
	})

	t.Run("to run container", func(t *testing.T) {
		clean := newArrayContainer()
		clean.resetTo(run)
		assert.True(t, clean.toEfficientContainer().equals(run))

		dirty := makeDirty()
		dirty.resetTo(run)
		assert.True(t, dirty.toEfficientContainer().equals(run))
	})
}

func TestArrayContainerValidation(t *testing.T) {
	array := newArrayContainer()
	upperBound := arrayDefaultMaxSize

	err := array.validate()
	assert.Error(t, err)

	for i := 0; i < upperBound; i++ {
		array.iadd(uint16(i))
	}
	err = array.validate()
	assert.NoError(t, err)

	// Introduce a sort error
	// We know that upperbound is unsorted because we populated up to upperbound
	array.content[500] = uint16(upperBound + upperBound)

	err = array.validate()
	assert.Error(t, err)

	array = newArrayContainer()

	// Technically a run, but make sure the incorrect sort detection handles equal elements
	for i := 0; i < upperBound; i++ {
		array.iadd(uint16(1))
	}
	err = array.validate()
	assert.NoError(t, err)

	array = newArrayContainer()

	for i := 0; i < 2*upperBound; i++ {
		array.iadd(uint16(i))
	}
	err = array.validate()
	assert.Error(t, err)
}

// go test -bench BenchmarkShortIteratorAdvance -run -
func BenchmarkShortIteratorAdvanceArray(b *testing.B) {
	benchmarkContainerIteratorAdvance(b, newArrayContainer())
}

// go test -bench BenchmarkShortIteratorNext -run -
func BenchmarkShortIteratorNextArray(b *testing.B) {
	benchmarkContainerIteratorNext(b, newArrayContainer())
}
