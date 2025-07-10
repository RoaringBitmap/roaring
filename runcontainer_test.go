package roaring

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// trial is used in the randomized testing of runContainers
type trial struct {
	n           int
	percentFill float64
	ntrial      int

	// only in the union test
	// only subtract test
	percentDelete float64

	// only in 067 randomized operations
	// we do this + 1 passes
	numRandomOpsPass int

	// allow sampling range control
	// only recent tests respect this.
	srang *interval16
}

// canMerge, and mergeInterval16s should do what they say
func TestRleInterval16s(t *testing.T) {
	a := newInterval16Range(0, 9)
	b := newInterval16Range(0, 1)
	report := sliceToString16([]interval16{a, b})
	_ = report
	c := newInterval16Range(2, 4)
	d := newInterval16Range(2, 5)
	e := newInterval16Range(0, 4)
	f := newInterval16Range(9, 9)
	g := newInterval16Range(8, 9)
	h := newInterval16Range(5, 6)
	i := newInterval16Range(6, 6)

	aIb, empty := intersectInterval16s(a, b)
	assert.False(t, empty)
	assert.EqualValues(t, b, aIb)

	assert.True(t, canMerge16(b, c))
	assert.True(t, canMerge16(c, b))
	assert.True(t, canMerge16(a, h))

	assert.True(t, canMerge16(d, e))
	assert.True(t, canMerge16(f, g))
	assert.True(t, canMerge16(c, h))

	assert.False(t, canMerge16(b, h))
	assert.False(t, canMerge16(h, b))
	assert.False(t, canMerge16(c, i))

	assert.EqualValues(t, e, mergeInterval16s(b, c))
	assert.EqualValues(t, e, mergeInterval16s(c, b))

	assert.EqualValues(t, h, mergeInterval16s(h, i))
	assert.EqualValues(t, h, mergeInterval16s(i, h))

	////// start
	assert.EqualValues(t, newInterval16Range(0, 1), mergeInterval16s(newInterval16Range(0, 0), newInterval16Range(1, 1)))
	assert.EqualValues(t, newInterval16Range(0, 1), mergeInterval16s(newInterval16Range(1, 1), newInterval16Range(0, 0)))
	assert.EqualValues(t, newInterval16Range(0, 5), mergeInterval16s(newInterval16Range(0, 4), newInterval16Range(3, 5)))
	assert.EqualValues(t, newInterval16Range(0, 4), mergeInterval16s(newInterval16Range(0, 4), newInterval16Range(3, 4)))

	assert.EqualValues(t, newInterval16Range(0, 8), mergeInterval16s(newInterval16Range(1, 7), newInterval16Range(0, 8)))
	assert.EqualValues(t, newInterval16Range(0, 8), mergeInterval16s(newInterval16Range(1, 7), newInterval16Range(0, 8)))

	assert.Panics(t, func() { _ = mergeInterval16s(newInterval16Range(0, 0), newInterval16Range(2, 3)) })
}

func TestRunOffset(t *testing.T) {
	v := newRunContainer16TakeOwnership([]interval16{newInterval16Range(34, 39)})
	offtest := uint16(65500)
	l, h := v.addOffset(offtest)

	expected := []int{65534, 65535, 65536, 65537, 65538, 65539}
	wout := make([]int, len(expected))

	var w0card, w1card int

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

	assert.Equal(t, v.getCardinality(), w0card+w1card)
	for i, x := range wout {
		assert.Equal(t, expected[i], x)
	}
}

func TestRunArrayUnionToRuns(t *testing.T) {
	arrayArg := newArrayContainerRange(0, 10)
	runArg := newRunContainer16Range(11, 65535)
	intervals, cardMinusOne := runArrayUnionToRuns(runArg, arrayArg)
	assert.Equal(t, uint16(65535), cardMinusOne)
	assert.Equal(t, []interval16{{start: 0, length: 65535}}, intervals)
}

func TestRleRunIterator16(t *testing.T) {
	t.Run("RunIterator16 unit tests for next, hasNext, and peekNext should pass", func(t *testing.T) {
		{
			rc := newRunContainer16()
			msg := rc.String()
			_ = msg

			assert.EqualValues(t, 0, rc.getCardinality())

			it := rc.newRunIterator16()

			assert.False(t, it.hasNext())
			assert.Panics(t, func() { it.peekNext() })
			assert.Panics(t, func() { it.next() })
		}
		{
			rc := newRunContainer16TakeOwnership([]interval16{newInterval16Range(4, 4)})
			assert.EqualValues(t, 1, rc.getCardinality())

			it := rc.newRunIterator16()

			assert.True(t, it.hasNext())
			assert.EqualValues(t, uint16(4), it.peekNext())
			assert.EqualValues(t, uint16(4), it.next())
		}
		{
			rc := newRunContainer16CopyIv([]interval16{newInterval16Range(4, 9)})
			assert.EqualValues(t, 6, rc.getCardinality())

			it := rc.newRunIterator16()
			assert.True(t, it.hasNext())

			for i := 4; i < 10; i++ {
				assert.Equal(t, uint16(i), it.next())
			}

			assert.False(t, it.hasNext())
		}

		{
			// basic nextMany test
			rc := newRunContainer16CopyIv([]interval16{newInterval16Range(4, 9)})
			assert.EqualValues(t, 6, rc.getCardinality())

			it := rc.newManyRunIterator16()
			buf := make([]uint32, 10)
			n := it.nextMany(0, buf)

			assert.Equal(t, 6, n)

			expected := []uint32{4, 5, 6, 7, 8, 9, 0, 0, 0, 0}
			for i, e := range expected {
				assert.Equal(t, e, buf[i])
			}
		}

		{
			// nextMany with len(buf) == 0
			rc := newRunContainer16CopyIv([]interval16{newInterval16Range(4, 9)})
			assert.EqualValues(t, 6, rc.getCardinality())

			it := rc.newManyRunIterator16()
			var buf []uint32
			n := it.nextMany(0, buf)

			assert.Equal(t, 0, n)
		}

		{
			// basic nextMany test across ranges
			rc := newRunContainer16CopyIv([]interval16{
				newInterval16Range(4, 7),
				newInterval16Range(11, 13),
				newInterval16Range(18, 21),
			})

			assert.EqualValues(t, 11, rc.getCardinality())

			it := rc.newManyRunIterator16()
			buf := make([]uint32, 15)
			n := it.nextMany(0, buf)

			assert.Equal(t, 11, n)

			expected := []uint32{4, 5, 6, 7, 11, 12, 13, 18, 19, 20, 21, 0, 0, 0, 0}
			for i, e := range expected {
				assert.Equal(t, e, buf[i])
			}
		}
		{
			// basic nextMany test across ranges with different buffer sizes
			rc := newRunContainer16CopyIv([]interval16{
				newInterval16Range(4, 7),
				newInterval16Range(11, 13),
				newInterval16Range(18, 21),
			})
			expectedCard := 11
			expectedVals := []uint32{4, 5, 6, 7, 11, 12, 13, 18, 19, 20, 21}
			hs := uint32(1 << 16)

			assert.EqualValues(t, expectedCard, rc.getCardinality())

			for bufSize := 2; bufSize < 15; bufSize++ {
				buf := make([]uint32, bufSize)
				seen := 0
				it := rc.newManyRunIterator16()
				for n := it.nextMany(hs, buf); n != 0; n = it.nextMany(hs, buf) {
					// catch runaway iteration
					assert.LessOrEqual(t, seen+n, expectedCard)

					for i, e := range expectedVals[seen : seen+n] {
						assert.Equal(t, e+hs, buf[i])
					}
					seen += n
					// if we have more values to return then we shouldn't leave empty slots in the buffer
					if seen < expectedCard {
						assert.Equal(t, bufSize, n)
					}
				}
				assert.Equal(t, expectedCard, seen)
			}
		}

		{
			// basic nextMany interaction with hasNext
			rc := newRunContainer16CopyIv([]interval16{newInterval16Range(4, 4)})
			assert.EqualValues(t, 1, rc.getCardinality())

			it := rc.newManyRunIterator16()
			assert.True(t, it.hasNext())

			buf := make([]uint32, 4)
			n := it.nextMany(0, buf)

			assert.Equal(t, 1, n)

			expected := []uint32{4, 0, 0, 0}

			for i, e := range expected {
				assert.Equal(t, e, buf[i])
			}

			assert.False(t, it.hasNext())

			buf = make([]uint32, 4)
			n = it.nextMany(0, buf)

			assert.Equal(t, 0, n)

			expected = []uint32{0, 0, 0, 0}
			for i, e := range expected {
				assert.Equal(t, e, buf[i])
			}
		}
		{
			rc := newRunContainer16TakeOwnership([]interval16{
				newInterval16Range(0, 0),
				newInterval16Range(2, 2),
				newInterval16Range(4, 4),
			})
			rc1 := newRunContainer16TakeOwnership([]interval16{
				newInterval16Range(6, 7),
				newInterval16Range(10, 11),
				newInterval16Range(MaxUint16, MaxUint16),
			})

			rc = rc.union(rc1)

			assert.EqualValues(t, 8, rc.getCardinality())

			it := rc.newRunIterator16()

			assert.EqualValues(t, 0, it.next())
			assert.EqualValues(t, 2, it.next())
			assert.EqualValues(t, 4, it.next())
			assert.EqualValues(t, 6, it.next())
			assert.EqualValues(t, 7, it.next())
			assert.EqualValues(t, 10, it.next())
			assert.EqualValues(t, 11, it.next())
			assert.EqualValues(t, MaxUint16, it.next())
			assert.False(t, it.hasNext())

			newInterval16Range(0, MaxUint16)
			rc2 := newRunContainer16TakeOwnership([]interval16{newInterval16Range(0, MaxUint16)})

			rc2 = rc2.union(rc)
			assert.Equal(t, 1, rc2.numIntervals())
		}
	})
}

func TestRleRunReverseIterator16(t *testing.T) {
	t.Run("RunReverseIterator16 unit tests for next, hasNext, and peekNext should pass", func(t *testing.T) {
		{
			rc := newRunContainer16()
			it := rc.newRunReverseIterator16()
			assert.False(t, it.hasNext())
			assert.Panics(t, func() { it.next() })
		}
		{
			rc := newRunContainer16TakeOwnership([]interval16{newInterval16Range(0, 0)})
			it := rc.newRunReverseIterator16()
			assert.True(t, it.hasNext())
			assert.EqualValues(t, uint16(0), it.next())
			assert.Panics(t, func() { it.next() })
			assert.False(t, it.hasNext())
			assert.Panics(t, func() { it.next() })
		}
		{
			rc := newRunContainer16TakeOwnership([]interval16{newInterval16Range(4, 4)})
			it := rc.newRunReverseIterator16()
			assert.True(t, it.hasNext())
			assert.EqualValues(t, uint16(4), it.next())
			assert.False(t, it.hasNext())
		}
		{
			rc := newRunContainer16TakeOwnership([]interval16{newInterval16Range(MaxUint16, MaxUint16)})
			it := rc.newRunReverseIterator16()
			assert.True(t, it.hasNext())
			assert.EqualValues(t, uint16(MaxUint16), it.next())
			assert.False(t, it.hasNext())
		}
		{
			rc := newRunContainer16TakeOwnership([]interval16{newInterval16Range(4, 9)})
			it := rc.newRunReverseIterator16()
			assert.True(t, it.hasNext())
			for i := 9; i >= 4; i-- {
				assert.Equal(t, uint16(i), it.next())
				if i > 4 {
					assert.True(t, it.hasNext())
				} else if i == 4 {
					assert.False(t, it.hasNext())
				}
			}
			assert.False(t, it.hasNext())
			assert.Panics(t, func() { it.next() })
		}
		{
			rc := newRunContainer16TakeOwnership([]interval16{
				newInterval16Range(0, 0),
				newInterval16Range(2, 2),
				newInterval16Range(4, 4),
				newInterval16Range(6, 7),
				newInterval16Range(10, 12),
				newInterval16Range(MaxUint16, MaxUint16),
			})

			it := rc.newRunReverseIterator16()
			assert.Equal(t, uint16(MaxUint16), it.next())
			assert.Equal(t, uint16(12), it.next())
			assert.Equal(t, uint16(11), it.next())
			assert.Equal(t, uint16(10), it.next())
			assert.Equal(t, uint16(7), it.next())
			assert.Equal(t, uint16(6), it.next())
			assert.Equal(t, uint16(4), it.next())
			assert.Equal(t, uint16(2), it.next())
			assert.Equal(t, uint16(0), it.next())
			assert.Equal(t, false, it.hasNext())
			assert.Panics(t, func() { it.next() })
		}
	})
}

func TestRleIntersection16(t *testing.T) {
	t.Run("RunContainer16.intersect of two RunContainer16(s) should return their intersection", func(t *testing.T) {
		{
			vals := []uint16{0, 2, 4, 6, 8, 10, 12, 14, 16, 18, MaxUint16 - 3, MaxUint16 - 1}

			a := newRunContainer16FromVals(true, vals[:5]...)
			b := newRunContainer16FromVals(true, vals[2:]...)

			assert.True(t, haveOverlap16(newInterval16Range(0, 2), newInterval16Range(2, 2)))
			assert.False(t, haveOverlap16(newInterval16Range(0, 2), newInterval16Range(3, 3)))

			isect := a.intersect(b)

			assert.EqualValues(t, 3, isect.getCardinality())
			assert.True(t, isect.contains(4))
			assert.True(t, isect.contains(6))
			assert.True(t, isect.contains(8))

			newInterval16Range(0, MaxUint16)
			d := newRunContainer16TakeOwnership([]interval16{newInterval16Range(0, MaxUint16)})
			isect = isect.intersect(d)

			assert.EqualValues(t, 3, isect.getCardinality())
			assert.True(t, isect.contains(4))
			assert.True(t, isect.contains(6))
			assert.True(t, isect.contains(8))

			e := newRunContainer16TakeOwnership(
				[]interval16{
					newInterval16Range(2, 4),
					newInterval16Range(8, 9),
					newInterval16Range(14, 16),
					newInterval16Range(20, 22),
				},
			)
			f := newRunContainer16TakeOwnership(
				[]interval16{
					newInterval16Range(3, 18),
					newInterval16Range(22, 23),
				},
			)

			{
				isect = e.intersect(f)

				assert.EqualValues(t, 8, isect.getCardinality())
				assert.True(t, isect.contains(3))
				assert.True(t, isect.contains(4))
				assert.True(t, isect.contains(8))
				assert.True(t, isect.contains(9))
				assert.True(t, isect.contains(14))
				assert.True(t, isect.contains(15))
				assert.True(t, isect.contains(16))
				assert.True(t, isect.contains(22))
			}

			{
				// check for symmetry
				isect = f.intersect(e)

				assert.EqualValues(t, 8, isect.getCardinality())
				assert.True(t, isect.contains(3))
				assert.True(t, isect.contains(4))
				assert.True(t, isect.contains(8))
				assert.True(t, isect.contains(9))
				assert.True(t, isect.contains(14))
				assert.True(t, isect.contains(15))
				assert.True(t, isect.contains(16))
				assert.True(t, isect.contains(22))
			}
		}
	})
}

func TestRleRandomIntersection16(t *testing.T) {
	t.Run("RunContainer.intersect of two RunContainers should return their intersection, and this should hold over randomized container content when compared to intersection done with hash maps", func(t *testing.T) {
		seed := int64(42)
		rand.Seed(seed)

		trials := []trial{
			{n: 100, percentFill: .80, ntrial: 10},
			{n: 1000, percentFill: .20, ntrial: 20},
			{n: 10000, percentFill: .01, ntrial: 10},
			{n: 1000, percentFill: .99, ntrial: 10},
		}

		tester := func(tr trial) {
			for j := 0; j < tr.ntrial; j++ {
				ma := make(map[int]bool)
				mb := make(map[int]bool)

				n := tr.n
				a := []uint16{}
				b := []uint16{}

				var first, second int

				draw := int(float64(n) * tr.percentFill)
				for i := 0; i < draw; i++ {
					r0 := rand.Intn(n)
					a = append(a, uint16(r0))
					ma[r0] = true
					if i == 0 {
						first = r0
						second = r0 + 1
						a = append(a, uint16(second))
						ma[second] = true
					}

					r1 := rand.Intn(n)
					b = append(b, uint16(r1))
					mb[r1] = true
				}

				// print a; very likely it has dups
				sort.Sort(uint16Slice(a))
				stringA := ""
				for i := range a {
					stringA += fmt.Sprintf("%v, ", a[i])
				}

				// hash version of intersect:
				hashi := make(map[int]bool)
				for k := range ma {
					if mb[k] {
						hashi[k] = true
					}
				}

				// RunContainer's Intersect
				brle := newRunContainer16FromVals(false, b...)

				// arle := newRunContainer16FromVals(false, a...)
				// instead of the above line, create from array
				// get better test coverage:
				arr := newArrayContainerRange(int(first), int(second))
				arle := newRunContainer16FromArray(arr)
				arle.set(false, a...)

				isect := arle.intersect(brle)

				// showHash("hashi", hashi)

				for k := range hashi {
					assert.True(t, isect.contains(uint16(k)))
				}

				assert.EqualValues(t, len(hashi), isect.getCardinality())
			}
		}

		for i := range trials {
			tester(trials[i])
		}
	})
}

func TestRleRandomUnion16(t *testing.T) {
	t.Run("RunContainer.union of two RunContainers should return their union, and this should hold over randomized container content when compared to union done with hash maps", func(t *testing.T) {
		seed := int64(42)
		rand.Seed(seed)

		trials := []trial{
			{n: 100, percentFill: .80, ntrial: 10},
			{n: 1000, percentFill: .20, ntrial: 20},
			{n: 10000, percentFill: .01, ntrial: 10},
			{n: 1000, percentFill: .99, ntrial: 10, percentDelete: .04},
		}

		tester := func(tr trial) {
			for j := 0; j < tr.ntrial; j++ {
				ma := make(map[int]bool)
				mb := make(map[int]bool)

				n := tr.n
				a := []uint16{}
				b := []uint16{}

				draw := int(float64(n) * tr.percentFill)
				numDel := int(float64(n) * tr.percentDelete)
				for i := 0; i < draw; i++ {
					r0 := rand.Intn(n)
					a = append(a, uint16(r0))
					ma[r0] = true

					r1 := rand.Intn(n)
					b = append(b, uint16(r1))
					mb[r1] = true
				}

				// hash version of union:
				hashu := make(map[int]bool)
				for k := range ma {
					hashu[k] = true
				}
				for k := range mb {
					hashu[k] = true
				}

				// showHash("hashu", hashu)

				// RunContainer's Union
				arle := newRunContainer16()
				for i := range a {
					arle.Add(a[i])
				}
				brle := newRunContainer16()
				brle.set(false, b...)

				union := arle.union(brle)
				un := union.AsSlice()
				sort.Sort(uint16Slice(un))

				for kk, v := range un {
					_ = kk
					assert.True(t, hashu[int(v)])
				}

				for k := range hashu {
					assert.True(t, union.contains(uint16(k)))
				}

				assert.EqualValues(t, len(hashu), union.getCardinality())

				// do the deletes, exercising the remove functionality
				for i := 0; i < numDel; i++ {
					r1 := rand.Intn(len(a))
					goner := a[r1]
					union.removeKey(goner)
					delete(hashu, int(goner))
				}

				// verify the same as in the hashu
				assert.EqualValues(t, len(hashu), union.getCardinality())

				for k := range hashu {
					assert.True(t, union.contains(uint16(k)))
				}
			}
		}

		for i := range trials {
			tester(trials[i])
		}
	})
}

func TestRleAndOrXor16(t *testing.T) {
	t.Run("RunContainer And, Or, Xor tests", func(t *testing.T) {
		{
			rc := newRunContainer16TakeOwnership([]interval16{
				newInterval16Range(0, 0),
				newInterval16Range(2, 2),
				newInterval16Range(4, 4),
			})
			b0 := NewBitmap()
			b0.Add(2)
			b0.Add(6)
			b0.Add(8)

			and := rc.And(b0)
			or := rc.Or(b0)
			xor := rc.Xor(b0)

			assert.EqualValues(t, 1, and.GetCardinality())
			assert.EqualValues(t, 5, or.GetCardinality())
			assert.EqualValues(t, 4, xor.GetCardinality())

			// test creating size 0 and 1 from array
			arr := newArrayContainerCapacity(0)
			empty := newRunContainer16FromArray(arr)
			onceler := newArrayContainerCapacity(1)
			onceler.content = append(onceler.content, uint16(0))
			oneZero := newRunContainer16FromArray(onceler)

			assert.EqualValues(t, 0, empty.getCardinality())
			assert.EqualValues(t, 1, oneZero.getCardinality())
			assert.EqualValues(t, 0, empty.And(b0).GetCardinality())
			assert.EqualValues(t, 3, empty.Or(b0).GetCardinality())

			// exercise newRunContainer16FromVals() with 0 and 1 inputs.
			empty2 := newRunContainer16FromVals(false, []uint16{}...)
			assert.EqualValues(t, 0, empty2.getCardinality())

			one2 := newRunContainer16FromVals(false, []uint16{1}...)
			assert.EqualValues(t, 1, one2.getCardinality())
		}
	})
}

func TestRlePanics16(t *testing.T) {
	t.Run("Some RunContainer calls/methods should panic if misused", func(t *testing.T) {
		// newRunContainer16FromVals
		assert.Panics(t, func() { newRunContainer16FromVals(true, 1, 0) })

		arr := newArrayContainerRange(1, 3)
		arr.content = []uint16{2, 3, 3, 2, 1}
		assert.Panics(t, func() { newRunContainer16FromArray(arr) })
	})
}

func TestRleCoverageOddsAndEnds16(t *testing.T) {
	t.Run("Some RunContainer code paths that don't otherwise get coverage -- these should be tested to increase percentage of code coverage in testing", func(t *testing.T) {
		rc := &runContainer16{}
		assert.Equal(t, "runContainer16{}", rc.String())

		rc.iv = make([]interval16, 1)
		rc.iv[0] = newInterval16Range(3, 4)

		assert.Equal(t, "runContainer16{0:[3, 4], }", rc.String())

		a := newInterval16Range(5, 9)
		b := newInterval16Range(0, 1)
		c := newInterval16Range(1, 2)

		// intersectInterval16s(a, b interval16)
		isect, isEmpty := intersectInterval16s(a, b)
		assert.True(t, isEmpty)

		// [0,0] can't be trusted: assert.Equal(t, 0, isect.runlen())
		isect, isEmpty = intersectInterval16s(b, c)

		assert.False(t, isEmpty)
		assert.EqualValues(t, 1, isect.runlen())

		// runContainer16.union
		{
			ra := newRunContainer16FromVals(false, 4, 5)
			rb := newRunContainer16FromVals(false, 4, 6, 8, 9, 10)
			ra.union(rb)

			assert.EqualValues(t, 2, rb.indexOfIntervalAtOrAfter(4, 2))
			assert.EqualValues(t, 2, rb.indexOfIntervalAtOrAfter(3, 2))
		}

		// runContainer.intersect
		{
			ra := newRunContainer16()
			rb := newRunContainer16()

			assert.EqualValues(t, 0, ra.intersect(rb).getCardinality())
		}
		{
			ra := newRunContainer16FromVals(false, 1)
			rb := newRunContainer16FromVals(false, 4)

			assert.EqualValues(t, 0, ra.intersect(rb).getCardinality())
		}

		// runContainer.Add
		{
			ra := newRunContainer16FromVals(false, 1)
			rb := newRunContainer16FromVals(false, 4)

			assert.EqualValues(t, 1, ra.getCardinality())
			assert.EqualValues(t, 1, rb.getCardinality())

			ra.Add(5)

			assert.EqualValues(t, 2, ra.getCardinality())

			// newRunIterator16()
			empty := newRunContainer16()
			it := empty.newRunIterator16()

			assert.Panics(t, func() { it.next() })

			it2 := ra.newRunIterator16()
			it2.curIndex = len(it2.rc.iv)

			assert.Panics(t, func() { it2.next() })

			// runIterator16.peekNext()
			emptyIt := empty.newRunIterator16()

			assert.Panics(t, func() { emptyIt.peekNext() })

			// newRunContainer16FromArray
			arr := newArrayContainerRange(1, 6)
			arr.content = []uint16{5, 5, 5, 6, 9}
			rc3 := newRunContainer16FromArray(arr)

			assert.EqualValues(t, 3, rc3.getCardinality())

			// runContainer16SerializedSizeInBytes
			// runContainer16.SerializedSizeInBytes
			_ = runContainer16SerializedSizeInBytes(3)
			_ = rc3.serializedSizeInBytes()

			// findNextIntervalThatIntersectsStartingFrom
			idx, _ := rc3.findNextIntervalThatIntersectsStartingFrom(0, 100)

			assert.EqualValues(t, 1, idx)

			// deleteAt / remove
			rc3.Add(10)
			rc3.removeKey(10)
			rc3.removeKey(9)

			assert.EqualValues(t, 2, rc3.getCardinality())

			rc3.Add(9)
			rc3.Add(10)
			rc3.Add(12)

			assert.EqualValues(t, 5, rc3.getCardinality())

			it3 := rc3.newRunIterator16()
			it3.next()
			it3.next()
			it3.next()
			it3.next()

			assert.EqualValues(t, 12, it3.peekNext())
			assert.EqualValues(t, 12, it3.next())
		}

		// runContainer16.equals
		{
			rc16 := newRunContainer16()
			assert.True(t, rc16.equals16(rc16))
			rc16b := newRunContainer16()

			assert.True(t, rc16.equals16(rc16b))

			rc16.Add(1)
			rc16b.Add(2)

			assert.False(t, rc16.equals16(rc16b))
		}
	})
}

func TestRleStoringMax16(t *testing.T) {
	t.Run("Storing the MaxUint16 should be possible, because it may be necessary to do so--users will assume that any valid uint16 should be storable. In particular the smaller 16-bit version will definitely expect full access to all bits.", func(t *testing.T) {
		rc := newRunContainer16()
		rc.Add(MaxUint16)

		assert.True(t, rc.contains(MaxUint16))
		assert.EqualValues(t, 1, rc.getCardinality())

		rc.removeKey(MaxUint16)

		assert.False(t, rc.contains(MaxUint16))
		assert.EqualValues(t, 0, rc.getCardinality())

		rc.set(false, MaxUint16-1, MaxUint16)

		assert.EqualValues(t, 2, rc.getCardinality())
		assert.True(t, rc.contains(MaxUint16-1))
		assert.True(t, rc.contains(MaxUint16))

		rc.removeKey(MaxUint16 - 1)

		assert.EqualValues(t, 1, rc.getCardinality())

		rc.removeKey(MaxUint16)

		assert.EqualValues(t, 0, rc.getCardinality())

		rc.set(false, MaxUint16-2, MaxUint16-1, MaxUint16)

		assert.EqualValues(t, 3, rc.getCardinality())
		assert.EqualValues(t, 1, rc.numIntervals())

		rc.removeKey(MaxUint16 - 1)

		assert.EqualValues(t, 2, rc.numIntervals())
		assert.EqualValues(t, 2, rc.getCardinality())
	})
}

// go test -bench BenchmarkFromBitmap -run -
func BenchmarkFromBitmap16(b *testing.B) {
	b.StopTimer()
	seed := int64(42)
	rand.Seed(seed)

	tr := trial{n: 10000, percentFill: .95, ntrial: 1, numRandomOpsPass: 100}
	_, _, bc := getRandomSameThreeContainers(tr)

	b.StartTimer()

	for j := 0; j < b.N; j++ {
		newRunContainer16FromBitmapContainer(bc)
	}
}

func TestRle16RandomIntersectAgainstOtherContainers010(t *testing.T) {
	t.Run("runContainer16 `and` operation against other container types should correctly do the intersection", func(t *testing.T) {
		seed := int64(42)
		rand.Seed(seed)

		trials := []trial{
			{n: 100, percentFill: .95, ntrial: 1},
		}

		tester := func(tr trial) {
			for j := 0; j < tr.ntrial; j++ {
				ma := make(map[int]bool)
				mb := make(map[int]bool)

				n := tr.n
				a := []uint16{}
				b := []uint16{}

				draw := int(float64(n) * tr.percentFill)
				for i := 0; i < draw; i++ {
					r0 := rand.Intn(n)
					a = append(a, uint16(r0))
					ma[r0] = true

					r1 := rand.Intn(n)
					b = append(b, uint16(r1))
					mb[r1] = true
				}

				// showArray16(a, "a")
				// showArray16(b, "b")

				// hash version of intersect:
				hashi := make(map[int]bool)
				for k := range ma {
					if mb[k] {
						hashi[k] = true
					}
				}

				// RunContainer's Intersect
				rc := newRunContainer16FromVals(false, a...)

				// vs bitmapContainer
				bc := newBitmapContainer()
				for _, bv := range b {
					bc.iadd(bv)
				}

				// vs arrayContainer
				ac := newArrayContainer()
				for _, bv := range b {
					ac.iadd(bv)
				}

				// vs runContainer
				rcb := newRunContainer16FromVals(false, b...)

				rcVsBcIsect := rc.and(bc)
				rcVsAcIsect := rc.and(ac)
				rcVsRcbIsect := rc.and(rcb)

				for k := range hashi {
					assert.True(t, rcVsBcIsect.contains(uint16(k)))

					assert.True(t, rcVsAcIsect.contains(uint16(k)))

					assert.True(t, rcVsRcbIsect.contains(uint16(k)))
				}

				assert.Equal(t, len(hashi), rcVsBcIsect.getCardinality())
				assert.Equal(t, len(hashi), rcVsAcIsect.getCardinality())
				assert.Equal(t, len(hashi), rcVsRcbIsect.getCardinality())
			}
		}

		for i := range trials {
			tester(trials[i])
		}
	})
}

func TestRle16RandomUnionAgainstOtherContainers011(t *testing.T) {
	t.Run("runContainer16 `or` operation against other container types should correctly do the intersection", func(t *testing.T) {
		seed := int64(42)
		rand.Seed(seed)

		trials := []trial{
			{n: 100, percentFill: .95, ntrial: 1},
		}

		tester := func(tr trial) {
			for j := 0; j < tr.ntrial; j++ {
				ma := make(map[int]bool)
				mb := make(map[int]bool)

				n := tr.n
				a := []uint16{}
				b := []uint16{}

				draw := int(float64(n) * tr.percentFill)
				for i := 0; i < draw; i++ {
					r0 := rand.Intn(n)
					a = append(a, uint16(r0))
					ma[r0] = true

					r1 := rand.Intn(n)
					b = append(b, uint16(r1))
					mb[r1] = true
				}

				// showArray16(a, "a")
				// showArray16(b, "b")

				// hash version of union
				hashi := make(map[int]bool)
				for k := range ma {
					hashi[k] = true
				}
				for k := range mb {
					hashi[k] = true
				}

				// RunContainer's 'or'
				rc := newRunContainer16FromVals(false, a...)

				// vs bitmapContainer
				bc := newBitmapContainer()
				for _, bv := range b {
					bc.iadd(bv)
				}

				// vs arrayContainer
				ac := newArrayContainer()
				for _, bv := range b {
					ac.iadd(bv)
				}

				// vs runContainer
				rcb := newRunContainer16FromVals(false, b...)

				rcVsBcUnion := rc.or(bc)
				rcVsAcUnion := rc.or(ac)
				rcVsRcbUnion := rc.or(rcb)

				for k := range hashi {
					assert.True(t, rcVsBcUnion.contains(uint16(k)))
					assert.True(t, rcVsAcUnion.contains(uint16(k)))
					assert.True(t, rcVsRcbUnion.contains(uint16(k)))
				}
				assert.Equal(t, len(hashi), rcVsBcUnion.getCardinality())
				assert.Equal(t, len(hashi), rcVsAcUnion.getCardinality())
				assert.Equal(t, len(hashi), rcVsRcbUnion.getCardinality())
			}
		}

		for i := range trials {
			tester(trials[i])
		}
	})
}

func TestRle16RandomInplaceUnionAgainstOtherContainers012(t *testing.T) {
	t.Run("runContainer16 `ior` inplace union operation against other container types should correctly do the intersection", func(t *testing.T) {
		seed := int64(42)
		rand.Seed(seed)

		trials := []trial{
			{n: 10, percentFill: .95, ntrial: 1},
		}

		tester := func(tr trial) {
			for j := 0; j < tr.ntrial; j++ {
				ma := make(map[int]bool)
				mb := make(map[int]bool)

				n := tr.n
				a := []uint16{}
				b := []uint16{}

				draw := int(float64(n) * tr.percentFill)
				for i := 0; i < draw; i++ {
					r0 := rand.Intn(n)
					a = append(a, uint16(r0))
					ma[r0] = true

					r1 := rand.Intn(n)
					b = append(b, uint16(r1))
					mb[r1] = true
				}

				// showArray16(a, "a")
				// showArray16(b, "b")

				// hash version of union
				hashi := make(map[int]bool)
				for k := range ma {
					hashi[k] = true
				}
				for k := range mb {
					hashi[k] = true
				}

				// RunContainer's 'or'
				rc := newRunContainer16FromVals(false, a...)
				rcVsBcUnion := rc.Clone()
				rcVsAcUnion := rc.Clone()
				rcVsRcbUnion := rc.Clone()

				// vs bitmapContainer
				bc := newBitmapContainer()
				for _, bv := range b {
					bc.iadd(bv)
				}

				// vs arrayContainer
				ac := newArrayContainer()
				for _, bv := range b {
					ac.iadd(bv)
				}

				// vs runContainer
				rcb := newRunContainer16FromVals(false, b...)

				rcVsBcUnion.ior(bc)
				rcVsAcUnion.ior(ac)
				rcVsRcbUnion.ior(rcb)

				for k := range hashi {
					assert.True(t, rcVsBcUnion.contains(uint16(k)))

					assert.True(t, rcVsAcUnion.contains(uint16(k)))

					assert.True(t, rcVsRcbUnion.contains(uint16(k)))
				}

				assert.Equal(t, len(hashi), rcVsBcUnion.getCardinality())
				assert.Equal(t, len(hashi), rcVsAcUnion.getCardinality())
				assert.Equal(t, len(hashi), rcVsRcbUnion.getCardinality())
			}
		}

		for i := range trials {
			tester(trials[i])
		}
	})
}

func TestRle16RandomInplaceIntersectAgainstOtherContainers014(t *testing.T) {
	t.Run("runContainer16 `iand` inplace-and operation against other container types should correctly do the intersection", func(t *testing.T) {
		seed := int64(42)
		rand.Seed(seed)

		trials := []trial{
			{n: 100, percentFill: .95, ntrial: 1},
		}

		tester := func(tr trial) {
			for j := 0; j < tr.ntrial; j++ {
				ma := make(map[int]bool)
				mb := make(map[int]bool)

				n := tr.n
				a := []uint16{}
				b := []uint16{}

				draw := int(float64(n) * tr.percentFill)
				for i := 0; i < draw; i++ {
					r0 := rand.Intn(n)
					a = append(a, uint16(r0))
					ma[r0] = true

					r1 := rand.Intn(n)
					b = append(b, uint16(r1))
					mb[r1] = true
				}

				// showArray16(a, "a")
				// showArray16(b, "b")

				// hash version of intersect:
				hashi := make(map[int]bool)
				for k := range ma {
					if mb[k] {
						hashi[k] = true
					}
				}

				// RunContainer's Intersect
				rc := newRunContainer16FromVals(false, a...)

				// vs bitmapContainer
				bc := newBitmapContainer()
				for _, bv := range b {
					bc.iadd(bv)
				}

				// vs arrayContainer
				ac := newArrayContainer()
				for _, bv := range b {
					ac.iadd(bv)
				}

				// vs runContainer
				rcb := newRunContainer16FromVals(false, b...)

				var rcVsBcIsect container = rc.Clone()
				var rcVsAcIsect container = rc.Clone()
				var rcVsRcbIsect container = rc.Clone()

				rcVsBcIsect = rcVsBcIsect.iand(bc)
				rcVsAcIsect = rcVsAcIsect.iand(ac)
				rcVsRcbIsect = rcVsRcbIsect.iand(rcb)

				for k := range hashi {
					assert.True(t, rcVsBcIsect.contains(uint16(k)))

					assert.True(t, rcVsAcIsect.contains(uint16(k)))

					assert.True(t, rcVsRcbIsect.contains(uint16(k)))
				}

				assert.Equal(t, len(hashi), rcVsBcIsect.getCardinality())
				assert.Equal(t, len(hashi), rcVsAcIsect.getCardinality())
				assert.Equal(t, len(hashi), rcVsRcbIsect.getCardinality())
			}
		}

		for i := range trials {
			tester(trials[i])
		}
	})
}

func TestRle16RemoveApi015(t *testing.T) {
	t.Run("runContainer16 `remove` (a minus b) should work", func(t *testing.T) {
		seed := int64(42)
		rand.Seed(seed)

		trials := []trial{
			{n: 100, percentFill: .95, ntrial: 1},
		}

		tester := func(tr trial) {
			for j := 0; j < tr.ntrial; j++ {
				ma := make(map[int]bool)
				mb := make(map[int]bool)

				n := tr.n
				a := []uint16{}
				b := []uint16{}

				draw := int(float64(n) * tr.percentFill)
				for i := 0; i < draw; i++ {
					r0 := rand.Intn(n)
					a = append(a, uint16(r0))
					ma[r0] = true

					r1 := rand.Intn(n)
					b = append(b, uint16(r1))
					mb[r1] = true
				}

				// showArray16(a, "a")
				// showArray16(b, "b")

				// hash version of remove:
				hashrm := make(map[int]bool)
				for k := range ma {
					hashrm[k] = true
				}
				for k := range mb {
					delete(hashrm, k)
				}

				// RunContainer's remove
				rc := newRunContainer16FromVals(false, a...)

				for k := range mb {
					rc.iremove(uint16(k))
				}

				for k := range hashrm {
					assert.True(t, rc.contains(uint16(k)))
				}

				assert.Equal(t, len(hashrm), rc.getCardinality())
			}
		}

		for i := range trials {
			tester(trials[i])
		}
	})
}

func showArray16(a []uint16, name string) {
	sort.Sort(uint16Slice(a))
	stringA := ""
	for i := range a {
		stringA += fmt.Sprintf("%v, ", a[i])
	}
}

func TestRle16RandomAndNot016(t *testing.T) {
	t.Run("runContainer16 `andNot` operation against other container types should correctly do the and-not operation", func(t *testing.T) {
		seed := int64(42)
		rand.Seed(seed)

		trials := []trial{
			{n: 1000, percentFill: .95, ntrial: 2},
		}

		tester := func(tr trial) {
			for j := 0; j < tr.ntrial; j++ {
				ma := make(map[int]bool)
				mb := make(map[int]bool)

				n := tr.n
				a := []uint16{}
				b := []uint16{}

				draw := int(float64(n) * tr.percentFill)
				for i := 0; i < draw; i++ {
					r0 := rand.Intn(n)
					a = append(a, uint16(r0))
					ma[r0] = true

					r1 := rand.Intn(n)
					b = append(b, uint16(r1))
					mb[r1] = true
				}

				// showArray16(a, "a")
				// showArray16(b, "b")

				// hash version of and-not
				hashi := make(map[int]bool)
				for k := range ma {
					hashi[k] = true
				}
				for k := range mb {
					delete(hashi, k)
				}

				// RunContainer's and-not
				rc := newRunContainer16FromVals(false, a...)

				// vs bitmapContainer
				bc := newBitmapContainer()
				for _, bv := range b {
					bc.iadd(bv)
				}

				// vs arrayContainer
				ac := newArrayContainer()
				for _, bv := range b {
					ac.iadd(bv)
				}

				// vs runContainer
				rcb := newRunContainer16FromVals(false, b...)

				rcVsBcAndnot := rc.andNot(bc)
				rcVsAcAndnot := rc.andNot(ac)
				rcVsRcbAndnot := rc.andNot(rcb)

				for k := range hashi {
					assert.True(t, rcVsBcAndnot.contains(uint16(k)))
					assert.True(t, rcVsAcAndnot.contains(uint16(k)))
					assert.True(t, rcVsRcbAndnot.contains(uint16(k)))
				}

				assert.Equal(t, len(hashi), rcVsBcAndnot.getCardinality())
				assert.Equal(t, len(hashi), rcVsAcAndnot.getCardinality())
				assert.Equal(t, len(hashi), rcVsRcbAndnot.getCardinality())
			}
		}

		for i := range trials {
			tester(trials[i])
		}
	})
}

func TestRle16RandomInplaceAndNot017(t *testing.T) {
	t.Run("runContainer16 `iandNot` operation against other container types should correctly do the inplace-and-not operation", func(t *testing.T) {
		seed := int64(42)
		rand.Seed(seed)

		trials := []trial{
			{n: 1000, percentFill: .95, ntrial: 2},
		}

		tester := func(tr trial) {
			for j := 0; j < tr.ntrial; j++ {
				ma := make(map[int]bool)
				mb := make(map[int]bool)

				n := tr.n
				a := []uint16{}
				b := []uint16{}

				draw := int(float64(n) * tr.percentFill)
				for i := 0; i < draw; i++ {
					r0 := rand.Intn(n)
					a = append(a, uint16(r0))
					ma[r0] = true

					r1 := rand.Intn(n)
					b = append(b, uint16(r1))
					mb[r1] = true
				}

				// showArray16(a, "a")
				// showArray16(b, "b")

				// hash version of and-not
				hashi := make(map[int]bool)
				for k := range ma {
					hashi[k] = true
				}
				for k := range mb {
					delete(hashi, k)
				}

				// RunContainer's and-not
				rc := newRunContainer16FromVals(false, a...)

				// vs bitmapContainer
				bc := newBitmapContainer()
				for _, bv := range b {
					bc.iadd(bv)
				}

				// vs arrayContainer
				ac := newArrayContainer()
				for _, bv := range b {
					ac.iadd(bv)
				}

				// vs runContainer
				rcb := newRunContainer16FromVals(false, b...)

				rcVsBcIandnot := rc.Clone()
				rcVsAcIandnot := rc.Clone()
				rcVsRcbIandnot := rc.Clone()

				rcVsBcIandnot.iandNot(bc)
				rcVsAcIandnot.iandNot(ac)
				rcVsRcbIandnot.iandNot(rcb)

				for k := range hashi {
					assert.True(t, rcVsBcIandnot.contains(uint16(k)))
					assert.True(t, rcVsAcIandnot.contains(uint16(k)))
					assert.True(t, rcVsRcbIandnot.contains(uint16(k)))
				}
				assert.Equal(t, len(hashi), rcVsBcIandnot.getCardinality())
				assert.Equal(t, len(hashi), rcVsAcIandnot.getCardinality())
				assert.Equal(t, len(hashi), rcVsRcbIandnot.getCardinality())
			}
		}

		for i := range trials {
			tester(trials[i])
		}
	})
}

func TestRle16InversionOfIntervals018(t *testing.T) {
	t.Run("runContainer `invert` operation should do a NOT on the set of intervals, in-place", func(t *testing.T) {
		seed := int64(42)
		rand.Seed(seed)

		trials := []trial{
			{n: 1000, percentFill: .90, ntrial: 1},
		}

		tester := func(tr trial) {
			for j := 0; j < tr.ntrial; j++ {
				ma := make(map[int]bool)
				hashNotA := make(map[int]bool)

				n := tr.n
				a := []uint16{}

				// hashNotA will be NOT ma
				// for i := 0; i < n; i++ {
				for i := 0; i < MaxUint16+1; i++ {
					hashNotA[i] = true
				}

				draw := int(float64(n) * tr.percentFill)
				for i := 0; i < draw; i++ {
					r0 := rand.Intn(n)
					a = append(a, uint16(r0))
					ma[r0] = true
					delete(hashNotA, r0)
				}

				// RunContainer's invert
				rc := newRunContainer16FromVals(false, a...)

				inv := rc.invert()

				assert.Equal(t, 1+MaxUint16-rc.getCardinality(), inv.getCardinality())

				for k := 0; k < n; k++ {
					if hashNotA[k] {
						assert.True(t, inv.contains(uint16(k)))
					}
				}

				// skip for now, too big to do 2^16-1
				assert.Equal(t, len(hashNotA), inv.getCardinality())
			}
		}

		for i := range trials {
			tester(trials[i])
		}
	})
}

func TestRle16SubtractionOfIntervals019(t *testing.T) {
	t.Run("runContainer `subtract` operation removes an interval in-place", func(t *testing.T) {
		// basics

		i22 := newInterval16Range(2, 2)
		left, _ := i22.subtractInterval(i22)
		assert.EqualValues(t, 0, len(left))

		v := newInterval16Range(1, 6)
		left, _ = v.subtractInterval(newInterval16Range(3, 4))

		assert.EqualValues(t, 2, len(left))
		assert.EqualValues(t, 1, left[0].start)
		assert.EqualValues(t, 2, left[0].last())
		assert.EqualValues(t, 5, left[1].start)
		assert.EqualValues(t, 6, left[1].last())

		v = newInterval16Range(1, 6)
		left, _ = v.subtractInterval(newInterval16Range(4, 10))

		assert.EqualValues(t, 1, len(left))
		assert.EqualValues(t, 1, left[0].start)
		assert.EqualValues(t, 3, left[0].last())

		v = newInterval16Range(5, 10)
		left, _ = v.subtractInterval(newInterval16Range(0, 7))

		assert.EqualValues(t, 1, len(left))
		assert.EqualValues(t, 8, left[0].start)
		assert.EqualValues(t, 10, left[0].last())

		seed := int64(42)
		rand.Seed(seed)

		trials := []trial{
			{n: 1000, percentFill: .90, ntrial: 1},
		}

		tester := func(tr trial) {
			for j := 0; j < tr.ntrial; j++ {
				ma := make(map[int]bool)
				mb := make(map[int]bool)

				n := tr.n
				a := []uint16{}
				b := []uint16{}

				// hashAminusB will be  ma - mb
				hashAminusB := make(map[int]bool)

				draw := int(float64(n) * tr.percentFill)
				for i := 0; i < draw; i++ {
					r0 := rand.Intn(n)
					a = append(a, uint16(r0))
					ma[r0] = true
					hashAminusB[r0] = true

					r1 := rand.Intn(n)
					b = append(b, uint16(r1))
					mb[r1] = true
				}

				for k := range mb {
					delete(hashAminusB, k)
				}

				// RunContainer's subtract A - B
				rc := newRunContainer16FromVals(false, a...)
				rcb := newRunContainer16FromVals(false, b...)

				abkup := rc.Clone()

				it := rcb.newRunIterator16()
				for it.hasNext() {
					nx := it.next()
					rc.isubtract(newInterval16Range(nx, nx))
				}

				// also check full interval subtraction
				for _, p := range rcb.iv {
					abkup.isubtract(p)
				}

				for k := range hashAminusB {
					assert.True(t, rc.contains(uint16(k)))
					assert.True(t, abkup.contains(uint16(k)))
				}

				assert.EqualValues(t, len(hashAminusB), rc.getCardinality())
				assert.EqualValues(t, len(hashAminusB), abkup.getCardinality())
			}
		}

		for i := range trials {
			tester(trials[i])
		}
	})
}

func TestRle16Rank020(t *testing.T) {
	v := container(newRunContainer16())
	v = v.iaddReturnMinimized(10)
	v = v.iaddReturnMinimized(100)
	v = v.iaddReturnMinimized(1000)
	if v.getCardinality() != 3 {
		t.Errorf("Bogus cardinality.")
	}
	for i := 0; i <= arrayDefaultMaxSize; i++ {
		thisrank := v.rank(uint16(i))
		if i < 10 {
			if thisrank != 0 {
				t.Errorf("At %d should be zero but is %d ", i, thisrank)
			}
		} else if i < 100 {
			if thisrank != 1 {
				t.Errorf("At %d should be zero but is %d ", i, thisrank)
			}
		} else if i < 1000 {
			if thisrank != 2 {
				t.Errorf("At %d should be zero but is %d ", i, thisrank)
			}
		} else {
			if thisrank != 3 {
				t.Errorf("At %d should be zero but is %d ", i, thisrank)
			}
		}
	}
}

func TestRle16NotAlsoKnownAsFlipRange021(t *testing.T) {
	t.Run("runContainer `Not` operation should flip the bits of a range on the new returned container", func(t *testing.T) {
		seed := int64(42)
		rand.Seed(seed)

		trials := []trial{
			{n: 100, percentFill: .8, ntrial: 2},
		}

		tester := func(tr trial) {
			for j := 0; j < tr.ntrial; j++ {

				// what is the interval we are going to flip?

				ma := make(map[int]bool)
				flipped := make(map[int]bool)

				n := tr.n
				a := []uint16{}

				draw := int(float64(n) * tr.percentFill)
				for i := 0; i < draw; i++ {
					r0 := rand.Intn(n)
					a = append(a, uint16(r0))
					ma[r0] = true
					flipped[r0] = true
				}

				// pick an interval to flip
				begin := rand.Intn(n)
				last := rand.Intn(n)
				if last < begin {
					begin, last = last, begin
				}

				// do the flip on the hash `flipped`
				for i := begin; i <= last; i++ {
					if flipped[i] {
						delete(flipped, i)
					} else {
						flipped[i] = true
					}
				}

				// RunContainer's Not
				rc := newRunContainer16FromVals(false, a...)
				flp := rc.Not(begin, last+1)

				assert.EqualValues(t, len(flipped), flp.getCardinality())

				for k := 0; k < n; k++ {
					if flipped[k] {
						assert.True(t, flp.contains(uint16(k)))
					} else {
						assert.False(t, flp.contains(uint16(k)))
					}
				}

				assert.EqualValues(t, len(flipped), flp.getCardinality())
			}
		}

		for i := range trials {
			tester(trials[i])
		}
	})
}

func TestRleEquals022(t *testing.T) {
	t.Run("runContainer `equals` should accurately compare contents against other container types", func(t *testing.T) {
		seed := int64(42)
		rand.Seed(seed)

		trials := []trial{
			{n: 100, percentFill: .2, ntrial: 10},
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

				rc := newRunContainer16FromVals(false, a...)

				// make bitmap and array versions:
				bc := newBitmapContainer()
				ac := newArrayContainer()
				for k := range ma {
					ac.iadd(uint16(k))
					bc.iadd(uint16(k))
				}

				// compare equals() across all three
				assert.True(t, rc.equals(ac))
				assert.True(t, rc.equals(bc))

				assert.True(t, ac.equals(rc))
				assert.True(t, ac.equals(bc))

				assert.True(t, bc.equals(ac))
				assert.True(t, bc.equals(rc))

				// and for good measure, check against the hash
				assert.EqualValues(t, len(ma), rc.getCardinality())
				assert.EqualValues(t, len(ma), ac.getCardinality())
				assert.EqualValues(t, len(ma), bc.getCardinality())

				for k := range ma {
					assert.True(t, rc.contains(uint16(k)))
					assert.True(t, ac.contains(uint16(k)))
					assert.True(t, bc.contains(uint16(k)))
				}
			}
		}

		for i := range trials {
			tester(trials[i])
		}
	})
}

func TestRleIntersects023(t *testing.T) {
	t.Run("runContainer `intersects` query should work against any mix of container types", func(t *testing.T) {
		seed := int64(42)
		rand.Seed(seed)

		trials := []trial{
			{n: 10, percentFill: .293, ntrial: 1000},
		}

		tester := func(tr trial) {
			for j := 0; j < tr.ntrial; j++ {

				ma := make(map[int]bool)
				mb := make(map[int]bool)

				n := tr.n
				a := []uint16{}
				b := []uint16{}

				draw := int(float64(n) * tr.percentFill)
				for i := 0; i < draw; i++ {
					r0 := rand.Intn(n)
					a = append(a, uint16(r0))
					ma[r0] = true

					r1 := rand.Intn(n)
					b = append(b, uint16(r1))
					mb[r1] = true
				}

				// determine if they intersect from the maps
				isect := false
				for k := range ma {
					if mb[k] {
						isect = true
						break
					}
				}

				rcA := newRunContainer16FromVals(false, a...)
				rcB := newRunContainer16FromVals(false, b...)

				// make bitmap and array versions:
				bcA := newBitmapContainer()
				bcB := newBitmapContainer()

				acA := newArrayContainer()
				acB := newArrayContainer()
				for k := range ma {
					acA.iadd(uint16(k))
					bcA.iadd(uint16(k))
				}
				for k := range mb {
					acB.iadd(uint16(k))
					bcB.iadd(uint16(k))
				}

				// compare intersects() across all three

				// same type
				assert.Equal(t, isect, rcA.intersects(rcB))
				assert.Equal(t, isect, acA.intersects(acB))
				assert.Equal(t, isect, bcA.intersects(bcB))

				// across types
				assert.Equal(t, isect, rcA.intersects(acB))
				assert.Equal(t, isect, rcA.intersects(bcB))

				assert.Equal(t, isect, acA.intersects(rcB))
				assert.Equal(t, isect, acA.intersects(bcB))

				assert.Equal(t, isect, bcA.intersects(acB))
				assert.Equal(t, isect, bcA.intersects(rcB))

				// and swap the call pattern, so we test B intersects A as well.

				// same type
				assert.Equal(t, isect, rcB.intersects(rcA))
				assert.Equal(t, isect, acB.intersects(acA))
				assert.Equal(t, isect, bcB.intersects(bcA))

				// across types
				assert.Equal(t, isect, rcB.intersects(acA))
				assert.Equal(t, isect, rcB.intersects(bcA))

				assert.Equal(t, isect, acB.intersects(rcA))
				assert.Equal(t, isect, acB.intersects(bcA))

				assert.Equal(t, isect, bcB.intersects(acA))
				assert.Equal(t, isect, bcB.intersects(rcA))
			}
		}

		for i := range trials {
			tester(trials[i])
		}
	})
}

func TestRleToEfficientContainer027(t *testing.T) {
	t.Run("runContainer toEfficientContainer should return equivalent containers", func(t *testing.T) {
		seed := int64(42)
		rand.Seed(seed)

		// 4096 or fewer integers -> array typically

		trials := []trial{
			{n: 8000, percentFill: .01, ntrial: 10},
			{n: 8000, percentFill: .99, ntrial: 10},
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

				rc := newRunContainer16FromVals(false, a...)
				c := rc.toEfficientContainer()

				assert.True(t, rc.equals(c))
			}
		}

		for i := range trials {
			tester(trials[i])
		}
	})

	t.Run("runContainer toEfficientContainer should return an equivalent bitmap when that is efficient", func(t *testing.T) {
		a := []uint16{}

		// odd intergers should be smallest as a bitmap
		for i := 0; i < MaxUint16; i++ {
			if i%2 == 1 {
				a = append(a, uint16(i))
			}
		}

		rc := newRunContainer16FromVals(false, a...)

		c := rc.toEfficientContainer()
		assert.True(t, rc.equals(c))

		_, isBitmapContainer := c.(*bitmapContainer)
		assert.True(t, isBitmapContainer)
	})
}

func TestRle16RandomFillLeastSignificant16bits029(t *testing.T) {
	t.Run("runContainer16.fillLeastSignificant16bits() should fill contents as expected, matching the same function on bitmap and array containers", func(t *testing.T) {
		seed := int64(42)
		rand.Seed(seed)

		trials := []trial{
			{n: 100, percentFill: .95, ntrial: 1},
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

				// showArray16(a, "a")

				// RunContainer
				rc := newRunContainer16FromVals(false, a...)

				// vs bitmapContainer
				bc := newBitmapContainer()
				for _, av := range a {
					bc.iadd(av)
				}

				// vs arrayContainer
				ac := newArrayContainer()
				for _, av := range a {
					ac.iadd(av)
				}

				acOut := make([]uint32, n+10)
				bcOut := make([]uint32, n+10)
				rcOut := make([]uint32, n+10)

				pos2 := 0

				// see Bitmap.ToArray() for principal use
				hs := uint32(43) << 16
				ac.fillLeastSignificant16bits(acOut, pos2, hs)
				bc.fillLeastSignificant16bits(bcOut, pos2, hs)
				rc.fillLeastSignificant16bits(rcOut, pos2, hs)

				assert.EqualValues(t, acOut, rcOut)
				assert.EqualValues(t, bcOut, rcOut)
			}
		}

		for i := range trials {
			tester(trials[i])
		}
	})
}

func TestRle16RandomGetShortIterator030(t *testing.T) {
	t.Run("runContainer16.getShortIterator should traverse the contents expected, matching the traversal of the bitmap and array containers", func(t *testing.T) {
		seed := int64(42)
		rand.Seed(seed)

		trials := []trial{
			{n: 100, percentFill: .95, ntrial: 1},
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

				// showArray16(a, "a")

				// RunContainer
				rc := newRunContainer16FromVals(false, a...)

				// vs bitmapContainer
				bc := newBitmapContainer()
				for _, av := range a {
					bc.iadd(av)
				}

				// vs arrayContainer
				ac := newArrayContainer()
				for _, av := range a {
					ac.iadd(av)
				}

				rit := rc.getShortIterator()
				ait := ac.getShortIterator()
				bit := bc.getShortIterator()

				for ait.hasNext() {
					rn := rit.next()
					an := ait.next()
					bn := bit.next()

					assert.Equal(t, an, rn)
					assert.Equal(t, bn, rn)
				}
			}
		}

		for i := range trials {
			tester(trials[i])
		}
	})
}

func TestRle16RandomIaddRangeIremoveRange031(t *testing.T) {
	t.Run("runContainer16.iaddRange and iremoveRange should add/remove contents as expected, matching the same operations on the bitmap and array containers and the hashmap pos control", func(t *testing.T) {
		seed := int64(42)
		rand.Seed(seed)

		trials := []trial{
			{n: 101, percentFill: .9, ntrial: 10},
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

				// showArray16(a, "a")

				// RunContainer
				rc := newRunContainer16FromVals(false, a...)

				// vs bitmapContainer
				bc := newBitmapContainer()
				for _, av := range a {
					bc.iadd(av)
				}

				// vs arrayContainer
				ac := newArrayContainer()
				for _, av := range a {
					ac.iadd(av)
				}

				// iaddRange and iRemoveRange : pick some distinct random endpoints
				a0 := rand.Intn(n)
				a1 := a0
				for a1 == a0 {
					a1 = rand.Intn(n)
				}
				if a0 > a1 {
					a0, a1 = a1, a0
				}

				r0 := rand.Intn(n)
				r1 := r0
				for r1 == r0 {
					r1 = rand.Intn(n)
				}
				if r0 > r1 {
					r0, r1 = r1, r0
				}

				// do the add
				for i := a0; i <= a1; i++ {
					ma[i] = true
				}
				// then the remove
				for i := r0; i <= r1; i++ {
					delete(ma, i)
				}

				rc.iaddRange(a0, a1+1)
				rc.iremoveRange(r0, r1+1)

				bc.iaddRange(a0, a1+1)
				bc.iremoveRange(r0, r1+1)

				ac.iaddRange(a0, a1+1)
				ac.iremoveRange(r0, r1+1)

				assert.EqualValues(t, len(ma), rc.getCardinality())
				assert.Equal(t, ac.getCardinality(), rc.getCardinality())
				assert.Equal(t, bc.getCardinality(), rc.getCardinality())

				rit := rc.getShortIterator()
				ait := ac.getShortIterator()
				bit := bc.getShortIterator()

				for ait.hasNext() {
					rn := rit.next()
					an := ait.next()
					bn := bit.next()

					assert.Equal(t, an, rn)
					assert.Equal(t, bn, rn)
				}
				// verify againt the map
				for k := range ma {
					assert.True(t, rc.contains(uint16(k)))
				}

				// coverage for run16 method
				assert.Equal(t, 2+4*rc.numberOfRuns(), rc.serializedSizeInBytes())
			}
		}

		for i := range trials {
			tester(trials[i])
		}
	})
}

func TestAllContainerMethodsAllContainerTypes065(t *testing.T) {
	t.Run("each of the container methods that takes two containers should handle all 3x3==9 possible ways of being called -- without panic", func(t *testing.T) {
		a := newArrayContainer()
		r := newRunContainer16()
		b := newBitmapContainer()

		arr := []container{a, r, b}
		for _, i := range arr {
			for _, j := range arr {
				i.and(j)
				i.iand(j)
				i.andNot(j)

				i.iandNot(j)
				i.xor(j)
				i.equals(j)

				i.or(j)
				i.ior(j)
				i.intersects(j)

				i.lazyOR(j)
				i.lazyIOR(j)
			}
		}
	})
}

type twoCall func(r container) container

type twofer struct {
	name string
	call twoCall
	cn   container
}

func TestAllContainerMethodsAllContainerTypesWithData067(t *testing.T) {
	t.Run("each of the container methods that takes two containers should handle all 3x3==9 possible ways of being called -- and return results that agree with each other", func(t *testing.T) {
		seed := int64(42)
		rand.Seed(seed)

		srang := newInterval16Range(MaxUint16-100, MaxUint16)
		trials := []trial{
			{n: 100, percentFill: .7, ntrial: 1, numRandomOpsPass: 100},
			{n: 100, percentFill: .7, ntrial: 1, numRandomOpsPass: 100, srang: &srang},
		}

		tester := func(tr trial) {
			for j := 0; j < tr.ntrial; j++ {

				a, r, b := getRandomSameThreeContainers(tr)
				a2, r2, b2 := getRandomSameThreeContainers(tr)

				receiver := []container{a, r, b}
				arg := []container{a2, r2, b2}
				callme := []twofer{}

				nCalls := 0
				for k, c := range receiver {
					callme = append(callme, twofer{"and", c.and, c})
					callme = append(callme, twofer{"iand", c.iand, c})
					callme = append(callme, twofer{"ior", c.ior, c})
					callme = append(callme, twofer{"lazyOR", c.lazyOR, c})
					callme = append(callme, twofer{"lazyIOR", c.lazyIOR, c})
					callme = append(callme, twofer{"or", c.or, c})
					callme = append(callme, twofer{"xor", c.xor, c})
					callme = append(callme, twofer{"andNot", c.andNot, c})
					callme = append(callme, twofer{"iandNot", c.iandNot, c})
					if k == 0 {
						nCalls = len(callme)
					}
				}

				for pass := 0; pass < tr.numRandomOpsPass+1; pass++ {
					for k := 0; k < nCalls; k++ {
						perm := getRandomPermutation(nCalls)
						kk := perm[k]
						c1 := callme[kk]          // array receiver
						c2 := callme[kk+nCalls]   // run receiver
						c3 := callme[kk+2*nCalls] // bitmap receiver

						if c1.name != c2.name {
							panic("internal logic error")
						}
						if c3.name != c2.name {
							panic("internal logic error")
						}

						for k2, a := range arg {

							if !c1.cn.equals(c2.cn) {
								panic("c1 not equal to c2")
							}
							if !c1.cn.equals(c3.cn) {
								panic("c1 not equal to c3")
							}

							res1 := c1.call(a) // array
							res2 := c2.call(a) // run
							res3 := c3.call(a) // bitmap

							z := c1.name

							// In-place operation are best effort
							// User should not assume the receiver is modified, returned container has to be used
							if strings.HasPrefix(z, "i") || z == "lazyIOR" {
								c1.cn = res1
								c2.cn = res2
								c3.cn = res3
							}

							if strings.HasPrefix(z, "lazy") {
								// on purpose, the lazy functions
								// do not scan to update their cardinality
								if asBc, isBc := res1.(*bitmapContainer); isBc {
									asBc.computeCardinality()
								}
								if asBc, isBc := res2.(*bitmapContainer); isBc {
									asBc.computeCardinality()
								}
								if asBc, isBc := res3.(*bitmapContainer); isBc {
									asBc.computeCardinality()
								}
							}

							// check for equality all ways...
							// excercising equals() calls too.

							if !res1.equals(res2) {
								panic(fmt.Sprintf("k:%v, k2:%v, res1 != res2,"+
									" call is '%s'", k, k2, c1.name))
							}
							if !res2.equals(res1) {
								panic(fmt.Sprintf("k:%v, k2:%v, res2 != res1,"+
									" call is '%s'", k, k2, c1.name))
							}
							if !res1.equals(res3) {
								panic(fmt.Sprintf("k:%v, k2:%v, res1 != res3,"+
									" call is '%s'", k, k2, c1.name))
							}
							if !res3.equals(res1) {
								panic(fmt.Sprintf("k:%v, k2:%v, res3 != res1,"+
									" call is '%s'", k, k2, c1.name))
							}
							if !res2.equals(res3) {
								panic(fmt.Sprintf("k:%v, k2:%v, res2 != res3,"+
									" call is '%s'", k, k2, c1.name))
							}
							if !res3.equals(res2) {
								panic(fmt.Sprintf("k:%v, k2:%v, res3 != res2,"+
									" call is '%s'", k, k2, c1.name))
							}
						}
					} // end k
				} // end pass

			} // end j
		} // end tester

		for i := range trials {
			tester(trials[i])
		}
	})
}

func TestNextValueRun(t *testing.T) {
	t.Run("Java Regression1", func(t *testing.T) {
		// [Java1] https://github.com/RoaringBitmap/RoaringBitmap/blob/5235aa62c32fa3bf7fae40a562e3edc75f61be4e/RoaringBitmap/src/test/java/org/roaringbitmap/TestRunContainer.java#L3645
		runContainer := newRunContainer16()
		runContainer.iaddRange(64, 129)
		assert.Equal(t, 64, runContainer.nextValue(0))
		assert.Equal(t, 64, runContainer.nextValue(64))
		assert.Equal(t, 65, runContainer.nextValue(65))
		assert.Equal(t, 128, runContainer.nextValue(128))
		assert.Equal(t, -1, runContainer.nextValue(129))
	})

	t.Run("Java Regression2", func(t *testing.T) {
		// [Java2] https://github.com/RoaringBitmap/RoaringBitmap/blob/5235aa62c32fa3bf7fae40a562e3edc75f61be4e/RoaringBitmap/src/test/java/org/roaringbitmap/TestRunContainer.java#L3655

		runContainer := newRunContainer16()
		runContainer.iaddRange(64, 129)
		runContainer.iaddRange(256, 256+64+1)
		assert.Equal(t, 64, runContainer.nextValue(0))
		assert.Equal(t, 64, runContainer.nextValue(64))
		assert.Equal(t, 65, runContainer.nextValue(65))
		assert.Equal(t, 128, runContainer.nextValue(128))
		assert.Equal(t, 256, runContainer.nextValue(129))
		assert.Equal(t, -1, runContainer.nextValue(512))
	})

	t.Run("Java Regression3", func(t *testing.T) {
		// [Java3] https://github.com/RoaringBitmap/RoaringBitmap/blob/5235aa62c32fa3bf7fae40a562e3edc75f61be4e/RoaringBitmap/src/test/java/org/roaringbitmap/TestRunContainer.java#L3666

		runContainer := newRunContainer16()
		runContainer.iaddRange(64, 129)
		runContainer.iaddRange(256, 200+300+1)
		runContainer.iaddRange(200, 200+300+1)
		runContainer.iaddRange(5000, 5000+200+1)
		assert.Equal(t, 64, runContainer.nextValue(0))
		assert.Equal(t, 64, runContainer.nextValue(64))
		assert.Equal(t, 64, runContainer.nextValue(64))
		assert.Equal(t, 65, runContainer.nextValue(65))
		assert.Equal(t, 128, runContainer.nextValue(128))
		assert.Equal(t, 200, runContainer.nextValue(129))
		assert.Equal(t, 200, runContainer.nextValue(199))
		assert.Equal(t, 200, runContainer.nextValue(200))
		assert.Equal(t, 250, runContainer.nextValue(250))
		assert.Equal(t, 5000, runContainer.nextValue(2500))
		assert.Equal(t, 5000, runContainer.nextValue(5000))
		assert.Equal(t, 5200, runContainer.nextValue(5200))
		assert.Equal(t, -1, runContainer.nextValue(5201))
	})
}

func TestPreviousValueRun(t *testing.T) {
	t.Run("Java Regression1", func(t *testing.T) {
		// [Java 1] https://github.com/RoaringBitmap/RoaringBitmap/blob/5235aa62c32fa3bf7fae40a562e3edc75f61be4e/RoaringBitmap/src/test/java/org/roaringbitmap/TestRunContainer.java#L3684
		runContainer := newRunContainer16()
		runContainer.iaddRange(64, 129)
		assert.Equal(t, -1, runContainer.previousValue(0))
		assert.Equal(t, -1, runContainer.previousValue(63))
		assert.Equal(t, 64, runContainer.previousValue(64))
		assert.Equal(t, 65, runContainer.previousValue(65))
		assert.Equal(t, 128, runContainer.previousValue(128))
		assert.Equal(t, 128, runContainer.previousValue(129))
	})

	t.Run("Java Regression2", func(t *testing.T) {
		// [Java 2]  https://github.com/RoaringBitmap/RoaringBitmap/blob/5235aa62c32fa3bf7fae40a562e3edc75f61be4e/RoaringBitmap/src/test/java/org/roaringbitmap/TestRunContainer.java#L3695
		runContainer := newRunContainer16()
		runContainer.iaddRange(64, 129)
		runContainer.iaddRange(200, 200+300+1)
		runContainer.iaddRange(5000, 5000+200+1)
		assert.Equal(t, -1, runContainer.previousValue(0))
		assert.Equal(t, -1, runContainer.previousValue(63))
		assert.Equal(t, 64, runContainer.previousValue(64))
		assert.Equal(t, 65, runContainer.previousValue(65))
		assert.Equal(t, 128, runContainer.previousValue(128))
		assert.Equal(t, 128, runContainer.previousValue(129))
		assert.Equal(t, 128, runContainer.previousValue(199))
		assert.Equal(t, 200, runContainer.previousValue(200))
		assert.Equal(t, 250, runContainer.previousValue(250))
		assert.Equal(t, 500, runContainer.previousValue(2500))
		assert.Equal(t, 5000, runContainer.previousValue(5000))
		assert.Equal(t, 5200, runContainer.previousValue(5200))
		// TODO Question
		assert.Equal(t, 5200, runContainer.previousValue(5201))
	})
}

func TestNextAbsentValueRun(t *testing.T) {
	t.Run("Java Regression1", func(t *testing.T) {
		// [Java 1] https://github.com/RoaringBitmap/RoaringBitmap/blob/5235aa62c32fa3bf7fae40a562e3edc75f61be4e/RoaringBitmap/src/test/java/org/roaringbitmap/TestRunContainer.java#L3760
		runContainer := newRunContainer16()
		runContainer.iaddRange(64, 129)
		assert.Equal(t, 0, runContainer.nextAbsentValue(0))
		assert.Equal(t, 63, runContainer.nextAbsentValue(63))
		assert.Equal(t, 129, runContainer.nextAbsentValue(64))
		assert.Equal(t, 129, runContainer.nextAbsentValue(65))
		assert.Equal(t, 129, runContainer.nextAbsentValue(128))
		assert.Equal(t, 129, runContainer.nextAbsentValue(129))
	})

	t.Run("Java Regression2", func(t *testing.T) {
		// [Java 2] https://github.com/RoaringBitmap/RoaringBitmap/blob/5235aa62c32fa3bf7fae40a562e3edc75f61be4e/RoaringBitmap/src/test/java/org/roaringbitmap/TestRunContainer.java#L3815
		runContainer := newRunContainer16()
		runContainer.iaddRange(64, 129)
		runContainer.iaddRange(200, 501)
		runContainer.iaddRange(5000, 5201)

		assert.Equal(t, 0, runContainer.nextAbsentValue(0))
		assert.Equal(t, 63, runContainer.nextAbsentValue(63))
		assert.Equal(t, 129, runContainer.nextAbsentValue(64))
		assert.Equal(t, 129, runContainer.nextAbsentValue(65))
		assert.Equal(t, 129, runContainer.nextAbsentValue(128))
		assert.Equal(t, 129, runContainer.nextAbsentValue(129))
		assert.Equal(t, 199, runContainer.nextAbsentValue(199))
		assert.Equal(t, 501, runContainer.nextAbsentValue(200))
		assert.Equal(t, 501, runContainer.nextAbsentValue(250))
		assert.Equal(t, 2500, runContainer.nextAbsentValue(2500))
		assert.Equal(t, 5201, runContainer.nextAbsentValue(5000))
		assert.Equal(t, 5201, runContainer.nextAbsentValue(5200))
	})

	t.Run("Java Regression3", func(t *testing.T) {
		// [Java 3] https://github.com/RoaringBitmap/RoaringBitmap/blob/5235aa62c32fa3bf7fae40a562e3edc75f61be4e/RoaringBitmap/src/test/java/org/roaringbitmap/TestRunContainer.java#L3832
		runContainer := newRunContainer16()
		for i := 0; i < 1000; i++ {
			assert.Equal(t, i, runContainer.nextAbsentValue(uint16(i)))
		}
	})
}

func TestPreviousAbsentValueRun(t *testing.T) {
	t.Run("Java Regression 1", func(t *testing.T) {
		// [Java 1] https://github.com/RoaringBitmap/RoaringBitmap/blob/5235aa62c32fa3bf7fae40a562e3edc75f61be4e/RoaringBitmap/src/test/java/org/roaringbitmap/TestRunContainer.java#L3732
		runContainer := newRunContainer16()
		runContainer.iaddRange(64, 129)

		assert.Equal(t, 0, runContainer.previousAbsentValue(0))
		assert.Equal(t, 63, runContainer.previousAbsentValue(63))
		assert.Equal(t, 63, runContainer.previousAbsentValue(64))
		assert.Equal(t, 63, runContainer.previousAbsentValue(65))
		assert.Equal(t, 63, runContainer.previousAbsentValue(128))
		assert.Equal(t, 129, runContainer.previousAbsentValue(129))
	})

	t.Run("Java Regression2", func(t *testing.T) {
		// [Java 2] https://github.com/RoaringBitmap/RoaringBitmap/blob/5235aa62c32fa3bf7fae40a562e3edc75f61be4e/RoaringBitmap/src/test/java/org/roaringbitmap/TestRunContainer.java#L3743

		runContainer := newRunContainer16()
		runContainer.iaddRange(64, 129)
		runContainer.iaddRange(200, 501)
		runContainer.iaddRange(5000, 5201)

		assert.Equal(t, 0, runContainer.previousAbsentValue(0))
		assert.Equal(t, 63, runContainer.previousAbsentValue(63))
		assert.Equal(t, 63, runContainer.previousAbsentValue(64))
		assert.Equal(t, 63, runContainer.previousAbsentValue(65))
		assert.Equal(t, 63, runContainer.previousAbsentValue(128))
		assert.Equal(t, 129, runContainer.previousAbsentValue(129))
		assert.Equal(t, 199, runContainer.previousAbsentValue(199))
		assert.Equal(t, 199, runContainer.previousAbsentValue(200))
		assert.Equal(t, 199, runContainer.previousAbsentValue(250))
		assert.Equal(t, 2500, runContainer.previousAbsentValue(2500))
		assert.Equal(t, 4999, runContainer.previousAbsentValue(5000))
		assert.Equal(t, 4999, runContainer.previousAbsentValue(5200))
	})

	t.Run("Java Regression3", func(t *testing.T) {
		// [Java 3] https://github.com/RoaringBitmap/RoaringBitmap/blob/5235aa62c32fa3bf7fae40a562e3edc75f61be4e/RoaringBitmap/src/test/java/org/roaringbitmap/TestRunContainer.java#L3760
		runContainer := newRunContainer16()
		for i := 0; i < 1000; i++ {
			assert.Equal(t, i, runContainer.previousAbsentValue(uint16(i)))
		}
	})
}

func TestRuntimeIteratorPeekNext(t *testing.T) {
	testContainerIteratorPeekNext(t, newRunContainer16())
}

func TestRuntimeIteratorAdvance(t *testing.T) {
	testContainerIteratorAdvance(t, newRunContainer16())
}

func TestIntervalOverlaps(t *testing.T) {
	// contiguous runs
	a := newInterval16Range(0, 9)
	b := newInterval16Range(10, 20)

	// Ensure the function is reflexive
	assert.False(t, a.isNonContiguousDisjoint(a))
	assert.False(t, a.isNonContiguousDisjoint(b))
	// Ensure the function is symmetric
	assert.False(t, b.isNonContiguousDisjoint(a))
	assert.Error(t, isNonContiguousDisjoint(a, b))

	// identical runs
	a = newInterval16Range(0, 9)
	b = newInterval16Range(0, 9)

	assert.False(t, a.isNonContiguousDisjoint(b))
	assert.False(t, b.isNonContiguousDisjoint(a))
	assert.Error(t, isNonContiguousDisjoint(a, b))

	// identical start runs
	a = newInterval16Range(0, 9)
	b = newInterval16Range(0, 20)

	assert.False(t, a.isNonContiguousDisjoint(b))
	assert.False(t, b.isNonContiguousDisjoint(a))
	assert.Error(t, isNonContiguousDisjoint(a, b))

	// overlapping runs
	a = newInterval16Range(0, 12)
	b = newInterval16Range(10, 20)

	assert.False(t, a.isNonContiguousDisjoint(b))
	assert.Error(t, isNonContiguousDisjoint(a, b))

	// subset runs
	a = newInterval16Range(0, 12)
	b = newInterval16Range(5, 9)

	assert.False(t, a.isNonContiguousDisjoint(b))
	assert.Error(t, isNonContiguousDisjoint(a, b))

	// degenerate runs
	a = newInterval16Range(0, 0)
	b = newInterval16Range(5, 5)

	assert.True(t, a.isNonContiguousDisjoint(b))
	assert.NoError(t, isNonContiguousDisjoint(a, b))

	// disjoint non-contiguous runs
	a = newInterval16Range(0, 100)
	b = newInterval16Range(1000, 2000)

	assert.True(t, a.isNonContiguousDisjoint(b))
	assert.NoError(t, isNonContiguousDisjoint(a, b))
}

func TestIntervalValidationFailing(t *testing.T) {
	rc := &runContainer16{}
	assert.Error(t, rc.validate())

	a := newInterval16Range(0, 9)
	b := newInterval16Range(0, 9)
	rc = &runContainer16{}
	rc.iv = append(rc.iv, a, b)
	assert.ErrorIs(t, rc.validate(), ErrRunIntervalEqual)

	a = newInterval16Range(0, 9)
	b = newInterval16Range(10, 20)

	rc = &runContainer16{}
	rc.iv = append(rc.iv, a, b)
	assert.ErrorIs(t, rc.validate(), ErrRunIntervalOverlap)

	a = newInterval16Range(0, 12)
	b = newInterval16Range(10, 20)

	rc = &runContainer16{}
	rc.iv = append(rc.iv, a, b)
	assert.Error(t, rc.validate(), ErrRunIntervalOverlap)

	c := newInterval16Range(100, 150)
	d := newInterval16Range(1000, 10000)

	rc = &runContainer16{}
	rc.iv = append(rc.iv, a, b, c, d)
	assert.ErrorIs(t, rc.validate(), ErrRunIntervalOverlap)

	a = newInterval16Range(0, 10)
	b = newInterval16Range(100, 200)

	// missort
	rc = &runContainer16{}
	rc.iv = append(rc.iv, b, a)
	assert.ErrorIs(t, rc.validate(), ErrRunNonSorted)

	rc = &runContainer16{}
	start := -4
	for i := 0; i < MaxNumIntervals; i++ {
		start += 4
		end := start + 1
		a := newInterval16Range(uint16(start), uint16(end))
		rc.iv = append(rc.iv, a)

	}
	assert.ErrorIs(t, rc.validate(), ErrRunIntervalSize)
	// too many small runs, use array
	rc = &runContainer16{}
	start = -3
	for i := 0; i < 10; i++ {
		start += 3
		end := start + 1
		a := newInterval16Range(uint16(start), uint16(end))
		rc.iv = append(rc.iv, a)

	}
	assert.ErrorIs(t, rc.validate(), ErrRunIntervalSize)
}

func TestIntervalValidationsPassing(t *testing.T) {
	rc := &runContainer16{}
	a := newInterval16Range(0, 10)
	b := newInterval16Range(100, 200)
	rc.iv = append(rc.iv, a, b)
	assert.NoError(t, rc.validate())

	// Large total sum, but enough intervals
	rc = &runContainer16{}
	a = newInterval16Range(0, uint16(MaxIntervalsSum+1))
	rc.iv = append(rc.iv, a)
	assert.NoError(t, rc.validate())

	rc = &runContainer16{}
	a = newInterval16Range(0, uint16(MaxIntervalsSum+1))
	b = newInterval16Range(uint16(MaxIntervalsSum+3), uint16(MaxIntervalsSum*2))
	rc.iv = append(rc.iv, a, b)
	assert.NoError(t, rc.validate())
}

func TestRunContainerUnionCardinality(t *testing.T) {
	t.Run("Two Empty Runs", func(t *testing.T) {
		first := runContainer16{}
		second := runContainer16{}
		result := first.unionCardinality(&second)
		assert.Equal(t, uint(0), result)
	})

	t.Run("First Run Empty", func(t *testing.T) {
		first := runContainer16{}
		second := runContainer16{}
		second.iaddRange(0, 1024)
		result := first.unionCardinality(&second)
		assert.Equal(t, uint(1024), result)
	})

	t.Run("Second Run Empty", func(t *testing.T) {
		first := runContainer16{}
		second := runContainer16{}
		first.iaddRange(0, 1024)
		result := first.unionCardinality(&second)
		assert.Equal(t, uint(1024), result)
	})

	t.Run("Disjoint Ranges", func(t *testing.T) {
		first := runContainer16{}
		first.iaddRange(512, 1024)
		second := runContainer16{}
		second.iaddRange(0, 256)
		result := first.unionCardinality(&second)
		assert.Equal(t, uint(256+512), result)
	})

	t.Run("Complete Overlap", func(t *testing.T) {
		first := runContainer16{}
		first.iaddRange(0, 256)
		second := runContainer16{}
		second.iaddRange(0, 256)
		result := first.unionCardinality(&second)
		assert.Equal(t, uint(256), result)
	})
}

func TestRunContainerIntersectCardinality(t *testing.T) {
	t.Run("Two Empty Runs", func(t *testing.T) {
		first := runContainer16{}
		second := runContainer16{}
		result := first.intersectCardinality(&second)
		assert.Equal(t, 0, result)
	})

	t.Run("First Run Empty", func(t *testing.T) {
		first := runContainer16{}
		second := runContainer16{}
		second.iaddRange(0, 1024)
		result := first.intersectCardinality(&second)
		assert.Equal(t, 0, result)
	})

	t.Run("Second Run Empty", func(t *testing.T) {
		first := runContainer16{}
		second := runContainer16{}
		first.iaddRange(0, 1024)
		result := first.intersectCardinality(&second)
		assert.Equal(t, 0, result)
	})

	t.Run("Disjoint Ranges", func(t *testing.T) {
		first := runContainer16{}
		first.iaddRange(512, 1024)
		second := runContainer16{}
		second.iaddRange(0, 256)
		result := first.intersectCardinality(&second)
		assert.Equal(t, 0, result)
	})

	t.Run("Complete Overlap", func(t *testing.T) {
		first := runContainer16{}
		first.iaddRange(0, 256)
		second := runContainer16{}
		second.iaddRange(0, 256)
		result := first.intersectCardinality(&second)
		assert.Equal(t, 256, result)
	})

	t.Run("Single Element Intersection", func(t *testing.T) {
		first := runContainer16{}
		first.iaddRange(0, 257)
		second := runContainer16{}
		second.iaddRange(256, 512)
		result := first.intersectCardinality(&second)
		assert.Equal(t, 1, result)
	})
}

// go test -bench BenchmarkShortIteratorAdvance -run -
func BenchmarkShortIteratorAdvanceRuntime(b *testing.B) {
	benchmarkContainerIteratorAdvance(b, newRunContainer16())
}

// go test -bench BenchmarkShortIteratorNext -run -
func BenchmarkShortIteratorNextRuntime(b *testing.B) {
	benchmarkContainerIteratorNext(b, newRunContainer16())
}

// generate random contents, then return that same
// logical content in three different container types
func getRandomSameThreeContainers(tr trial) (*arrayContainer, *runContainer16, *bitmapContainer) {
	ma := make(map[int]bool)

	n := tr.n
	a := []uint16{}

	var samp interval16
	if tr.srang != nil {
		samp = *tr.srang
	} else {
		if n-1 > MaxUint16 {
			panic(fmt.Errorf("n out of range: %v", n))
		}
		samp.start = 0
		samp.length = uint16(n - 2)
	}

	draw := int(float64(n) * tr.percentFill)
	for i := 0; i < draw; i++ {
		r0 := int(samp.start) + rand.Intn(int(samp.runlen()))
		a = append(a, uint16(r0))
		ma[r0] = true
	}

	rc := newRunContainer16FromVals(false, a...)

	// vs bitmapContainer
	bc := newBitmapContainerFromRun(rc)
	ac := rc.toArrayContainer()

	return ac, rc, bc
}
