package roaring

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRle16RandomIntersectAgainstOtherContainers010(t *testing.T) {

	Convey("runContainer16 `and` operation against other container types should correctly do the intersection", t, func() {
		seed := int64(42)
		p("seed is %v", seed)
		rand.Seed(seed)

		trials := []trial{
			trial{n: 100, percentFill: .95, ntrial: 1},
		}

		tester := func(tr trial) {
			for j := 0; j < tr.ntrial; j++ {
				p("TestRleRandomAndAgainstOtherContainers on check# j=%v", j)
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

				//showArray16(a, "a")
				//showArray16(b, "b")

				// hash version of intersect:
				hashi := make(map[int]bool)
				for k := range ma {
					if mb[k] {
						hashi[k] = true
					}
				}

				// RunContainer's Intersect
				rc := newRunContainer16FromVals(false, a...)

				p("rc from a is %v", rc)

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

				rc_vs_bc_isect := rc.and(bc)
				rc_vs_ac_isect := rc.and(ac)
				rc_vs_rcb_isect := rc.and(rcb)

				p("rc_vs_bc_isect is %v", rc_vs_bc_isect)
				p("rc_vs_ac_isect is %v", rc_vs_ac_isect)
				p("rc_vs_rcb_isect is %v", rc_vs_rcb_isect)

				//showHash("hashi", hashi)

				for k := range hashi {
					p("hashi has %v, checking in rc_vs_bc_isect", k)
					So(rc_vs_bc_isect.contains(uint16(k)), ShouldBeTrue)

					p("hashi has %v, checking in rc_vs_ac_isect", k)
					So(rc_vs_ac_isect.contains(uint16(k)), ShouldBeTrue)

					p("hashi has %v, checking in rc_vs_rcb_isect", k)
					So(rc_vs_rcb_isect.contains(uint16(k)), ShouldBeTrue)
				}

				p("checking for cardinality agreement: rc_vs_bc_isect is %v, len(hashi) is %v", rc_vs_bc_isect.getCardinality(), len(hashi))
				p("checking for cardinality agreement: rc_vs_ac_isect is %v, len(hashi) is %v", rc_vs_ac_isect.getCardinality(), len(hashi))
				p("checking for cardinality agreement: rc_vs_rcb_isect is %v, len(hashi) is %v", rc_vs_rcb_isect.getCardinality(), len(hashi))
				So(rc_vs_bc_isect.getCardinality(), ShouldEqual, len(hashi))
				So(rc_vs_ac_isect.getCardinality(), ShouldEqual, len(hashi))
				So(rc_vs_rcb_isect.getCardinality(), ShouldEqual, len(hashi))
			}
			p("done with randomized and() vs bitmapContainer and arrayContainer checks for trial %#v", tr)
		}

		for i := range trials {
			tester(trials[i])
		}

	})
}

func TestRle16RandomUnionAgainstOtherContainers011(t *testing.T) {

	Convey("runContainer16 `or` operation against other container types should correctly do the intersection", t, func() {
		seed := int64(42)
		p("seed is %v", seed)
		rand.Seed(seed)

		trials := []trial{
			trial{n: 100, percentFill: .95, ntrial: 1},
			/*			trial{n: 100, percentFill: .01, ntrial: 10},
						trial{n: 100, percentFill: .99, ntrial: 10},
						trial{n: 100, percentFill: .50, ntrial: 10},
						trial{n: 10, percentFill: 1.0, ntrial: 10},
			*/
		}

		tester := func(tr trial) {
			for j := 0; j < tr.ntrial; j++ {
				p("TestRleRandomAndAgainstOtherContainers on check# j=%v", j)
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

				//showArray16(a, "a")
				//showArray16(b, "b")

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

				p("rc from a is %v", rc)

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

				rc_vs_bc_union := rc.or(bc)
				rc_vs_ac_union := rc.or(ac)
				rc_vs_rcb_union := rc.or(rcb)

				p("rc_vs_bc_union is %v", rc_vs_bc_union)
				p("rc_vs_ac_union is %v", rc_vs_ac_union)
				p("rc_vs_rcb_union is %v", rc_vs_rcb_union)

				//showHash("hashi", hashi)

				for k := range hashi {
					p("hashi has %v, checking in rc_vs_bc_union", k)
					So(rc_vs_bc_union.contains(uint16(k)), ShouldBeTrue)

					p("hashi has %v, checking in rc_vs_ac_union", k)
					So(rc_vs_ac_union.contains(uint16(k)), ShouldBeTrue)

					p("hashi has %v, checking in rc_vs_rcb_union", k)
					So(rc_vs_rcb_union.contains(uint16(k)), ShouldBeTrue)
				}

				p("checking for cardinality agreement: rc_vs_bc_union is %v, len(hashi) is %v", rc_vs_bc_union.getCardinality(), len(hashi))
				p("checking for cardinality agreement: rc_vs_ac_union is %v, len(hashi) is %v", rc_vs_ac_union.getCardinality(), len(hashi))
				p("checking for cardinality agreement: rc_vs_rcb_union is %v, len(hashi) is %v", rc_vs_rcb_union.getCardinality(), len(hashi))
				So(rc_vs_bc_union.getCardinality(), ShouldEqual, len(hashi))
				So(rc_vs_ac_union.getCardinality(), ShouldEqual, len(hashi))
				So(rc_vs_rcb_union.getCardinality(), ShouldEqual, len(hashi))
			}
			p("done with randomized or() vs bitmapContainer and arrayContainer checks for trial %#v", tr)
		}

		for i := range trials {
			tester(trials[i])
		}

	})
}

func TestRle16RandomInplaceUnionAgainstOtherContainers012(t *testing.T) {

	Convey("runContainer16 `ior` inplace union operation against other container types should correctly do the intersection", t, func() {
		seed := int64(42)
		p("seed is %v", seed)
		rand.Seed(seed)

		trials := []trial{
			trial{n: 10, percentFill: .95, ntrial: 1},
		}

		tester := func(tr trial) {
			for j := 0; j < tr.ntrial; j++ {
				p("TestRleRandomInplaceUnionAgainstOtherContainers on check# j=%v", j)
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

				//showArray16(a, "a")
				//showArray16(b, "b")

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

				p("rc from a is %v", rc)

				rc_vs_bc_union := rc.Clone()
				rc_vs_ac_union := rc.Clone()
				rc_vs_rcb_union := rc.Clone()

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

				rc_vs_bc_union.ior(bc)
				rc_vs_ac_union.ior(ac)
				rc_vs_rcb_union.ior(rcb)

				p("rc_vs_bc_union is %v", rc_vs_bc_union)
				p("rc_vs_ac_union is %v", rc_vs_ac_union)
				p("rc_vs_rcb_union is %v", rc_vs_rcb_union)

				//showHash("hashi", hashi)

				for k := range hashi {
					p("hashi has %v, checking in rc_vs_bc_union", k)
					So(rc_vs_bc_union.contains(uint16(k)), ShouldBeTrue)

					p("hashi has %v, checking in rc_vs_ac_union", k)
					So(rc_vs_ac_union.contains(uint16(k)), ShouldBeTrue)

					p("hashi has %v, checking in rc_vs_rcb_union", k)
					So(rc_vs_rcb_union.contains(uint16(k)), ShouldBeTrue)
				}

				p("checking for cardinality agreement: rc_vs_bc_union is %v, len(hashi) is %v", rc_vs_bc_union.getCardinality(), len(hashi))
				p("checking for cardinality agreement: rc_vs_ac_union is %v, len(hashi) is %v", rc_vs_ac_union.getCardinality(), len(hashi))
				p("checking for cardinality agreement: rc_vs_rcb_union is %v, len(hashi) is %v", rc_vs_rcb_union.getCardinality(), len(hashi))
				So(rc_vs_bc_union.getCardinality(), ShouldEqual, len(hashi))
				So(rc_vs_ac_union.getCardinality(), ShouldEqual, len(hashi))
				So(rc_vs_rcb_union.getCardinality(), ShouldEqual, len(hashi))
			}
			p("done with randomized or() vs bitmapContainer and arrayContainer checks for trial %#v", tr)
		}

		for i := range trials {
			tester(trials[i])
		}

	})
}

func TestRle16RandomInplaceIntersectAgainstOtherContainers014(t *testing.T) {

	Convey("runContainer16 `iand` inplace-and operation against other container types should correctly do the intersection", t, func() {
		seed := int64(42)
		p("seed is %v", seed)
		rand.Seed(seed)

		trials := []trial{
			trial{n: 100, percentFill: .95, ntrial: 1},
		}

		tester := func(tr trial) {
			for j := 0; j < tr.ntrial; j++ {
				p("TestRleRandomAndAgainstOtherContainers on check# j=%v", j)
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

				//showArray16(a, "a")
				//showArray16(b, "b")

				// hash version of intersect:
				hashi := make(map[int]bool)
				for k := range ma {
					if mb[k] {
						hashi[k] = true
					}
				}

				// RunContainer's Intersect
				rc := newRunContainer16FromVals(false, a...)

				p("rc from a is %v", rc)

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

				rc_vs_bc_isect := rc.Clone()
				rc_vs_ac_isect := rc.Clone()
				rc_vs_rcb_isect := rc.Clone()

				rc_vs_bc_isect.iand(bc)
				rc_vs_ac_isect.iand(ac)
				rc_vs_rcb_isect.iand(rcb)

				p("rc_vs_bc_isect is %v", rc_vs_bc_isect)
				p("rc_vs_ac_isect is %v", rc_vs_ac_isect)
				p("rc_vs_rcb_isect is %v", rc_vs_rcb_isect)

				//showHash("hashi", hashi)

				for k := range hashi {
					p("hashi has %v, checking in rc_vs_bc_isect", k)
					So(rc_vs_bc_isect.contains(uint16(k)), ShouldBeTrue)

					p("hashi has %v, checking in rc_vs_ac_isect", k)
					So(rc_vs_ac_isect.contains(uint16(k)), ShouldBeTrue)

					p("hashi has %v, checking in rc_vs_rcb_isect", k)
					So(rc_vs_rcb_isect.contains(uint16(k)), ShouldBeTrue)
				}

				p("checking for cardinality agreement: rc_vs_bc_isect is %v, len(hashi) is %v", rc_vs_bc_isect.getCardinality(), len(hashi))
				p("checking for cardinality agreement: rc_vs_ac_isect is %v, len(hashi) is %v", rc_vs_ac_isect.getCardinality(), len(hashi))
				p("checking for cardinality agreement: rc_vs_rcb_isect is %v, len(hashi) is %v", rc_vs_rcb_isect.getCardinality(), len(hashi))
				So(rc_vs_bc_isect.getCardinality(), ShouldEqual, len(hashi))
				So(rc_vs_ac_isect.getCardinality(), ShouldEqual, len(hashi))
				So(rc_vs_rcb_isect.getCardinality(), ShouldEqual, len(hashi))
			}
			p("done with randomized and() vs bitmapContainer and arrayContainer checks for trial %#v", tr)
		}

		for i := range trials {
			tester(trials[i])
		}

	})
}

func TestRle16RemoveApi015(t *testing.T) {

	Convey("runContainer16 `remove` (a minus b) should work", t, func() {
		seed := int64(42)
		p("seed is %v", seed)
		rand.Seed(seed)

		trials := []trial{
			trial{n: 100, percentFill: .95, ntrial: 1},
		}

		tester := func(tr trial) {
			for j := 0; j < tr.ntrial; j++ {
				p("TestRle16RemoveApi015 on check# j=%v", j)
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

				//showArray16(a, "a")
				//showArray16(b, "b")

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

				p("rc from a, pre-remove, is %v", rc)

				for k := range mb {
					rc.iremove(uint16(k))
				}

				p("rc from a, post-iremove, is %v", rc)

				//showHash("correct answer is hashrm", hashrm)

				for k := range hashrm {
					p("hashrm has %v, checking in rc", k)
					So(rc.contains(uint16(k)), ShouldBeTrue)
				}

				p("checking for cardinality agreement: rc is %v, len(hashrm) is %v", rc.getCardinality(), len(hashrm))
				So(rc.getCardinality(), ShouldEqual, len(hashrm))
			}
			p("done with randomized remove() checks for trial %#v", tr)
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
	p("%s is '%v'", name, stringA)
}

func TestRle16RandomAndNot016(t *testing.T) {

	Convey("runContainer16 `andNot` operation against other container types should correctly do the and-not operation", t, func() {
		seed := int64(42)
		p("seed is %v", seed)
		rand.Seed(seed)

		trials := []trial{
			trial{n: 1000, percentFill: .95, ntrial: 2},
		}

		tester := func(tr trial) {
			for j := 0; j < tr.ntrial; j++ {
				p("TestRle16RandomAndNot16 on check# j=%v", j)
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

				//showArray16(a, "a")
				//showArray16(b, "b")

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

				p("rc from a is %v", rc)

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

				rc_vs_bc_andnot := rc.andNot(bc)
				rc_vs_ac_andnot := rc.andNot(ac)
				rc_vs_rcb_andnot := rc.andNot(rcb)

				p("rc_vs_bc_andnot is %v", rc_vs_bc_andnot)
				p("rc_vs_ac_andnot is %v", rc_vs_ac_andnot)
				p("rc_vs_rcb_andnot is %v", rc_vs_rcb_andnot)

				//showHash("hashi", hashi)

				for k := range hashi {
					p("hashi has %v, checking in rc_vs_bc_andnot", k)
					So(rc_vs_bc_andnot.contains(uint16(k)), ShouldBeTrue)

					p("hashi has %v, checking in rc_vs_ac_andnot", k)
					So(rc_vs_ac_andnot.contains(uint16(k)), ShouldBeTrue)

					p("hashi has %v, checking in rc_vs_rcb_andnot", k)
					So(rc_vs_rcb_andnot.contains(uint16(k)), ShouldBeTrue)
				}

				p("checking for cardinality agreement: rc_vs_bc_andnot is %v, len(hashi) is %v", rc_vs_bc_andnot.getCardinality(), len(hashi))
				p("checking for cardinality agreement: rc_vs_ac_andnot is %v, len(hashi) is %v", rc_vs_ac_andnot.getCardinality(), len(hashi))
				p("checking for cardinality agreement: rc_vs_rcb_andnot is %v, len(hashi) is %v", rc_vs_rcb_andnot.getCardinality(), len(hashi))
				So(rc_vs_bc_andnot.getCardinality(), ShouldEqual, len(hashi))
				So(rc_vs_ac_andnot.getCardinality(), ShouldEqual, len(hashi))
				So(rc_vs_rcb_andnot.getCardinality(), ShouldEqual, len(hashi))
			}
			p("done with randomized andNot() vs bitmapContainer and arrayContainer checks for trial %#v", tr)
		}

		for i := range trials {
			tester(trials[i])
		}

	})
}

func TestRle16RandomInplaceAndNot017(t *testing.T) {

	Convey("runContainer16 `iandNot` operation against other container types should correctly do the inplace-and-not operation", t, func() {
		seed := int64(42)
		p("seed is %v", seed)
		rand.Seed(seed)

		trials := []trial{
			trial{n: 1000, percentFill: .95, ntrial: 2},
		}

		tester := func(tr trial) {
			for j := 0; j < tr.ntrial; j++ {
				p("TestRle16RandomAndNot16 on check# j=%v", j)
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

				//showArray16(a, "a")
				//showArray16(b, "b")

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

				p("rc from a is %v", rc)

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

				rc_vs_bc_iandnot := rc.Clone()
				rc_vs_ac_iandnot := rc.Clone()
				rc_vs_rcb_iandnot := rc.Clone()

				rc_vs_bc_iandnot.iandNot(bc)
				rc_vs_ac_iandnot.iandNot(ac)
				rc_vs_rcb_iandnot.iandNot(rcb)

				p("rc_vs_bc_iandnot is %v", rc_vs_bc_iandnot)
				p("rc_vs_ac_iandnot is %v", rc_vs_ac_iandnot)
				p("rc_vs_rcb_iandnot is %v", rc_vs_rcb_iandnot)

				//showHash("hashi", hashi)

				for k := range hashi {
					p("hashi has %v, checking in rc_vs_bc_iandnot", k)
					So(rc_vs_bc_iandnot.contains(uint16(k)), ShouldBeTrue)

					p("hashi has %v, checking in rc_vs_ac_iandnot", k)
					So(rc_vs_ac_iandnot.contains(uint16(k)), ShouldBeTrue)

					p("hashi has %v, checking in rc_vs_rcb_iandnot", k)
					So(rc_vs_rcb_iandnot.contains(uint16(k)), ShouldBeTrue)
				}

				p("checking for cardinality agreement: rc_vs_bc_iandnot is %v, len(hashi) is %v", rc_vs_bc_iandnot.getCardinality(), len(hashi))
				p("checking for cardinality agreement: rc_vs_ac_iandnot is %v, len(hashi) is %v", rc_vs_ac_iandnot.getCardinality(), len(hashi))
				p("checking for cardinality agreement: rc_vs_rcb_iandnot is %v, len(hashi) is %v", rc_vs_rcb_iandnot.getCardinality(), len(hashi))
				So(rc_vs_bc_iandnot.getCardinality(), ShouldEqual, len(hashi))
				So(rc_vs_ac_iandnot.getCardinality(), ShouldEqual, len(hashi))
				So(rc_vs_rcb_iandnot.getCardinality(), ShouldEqual, len(hashi))
			}
			p("done with randomized andNot() vs bitmapContainer and arrayContainer checks for trial %#v", tr)
		}

		for i := range trials {
			tester(trials[i])
		}

	})
}

func TestRle16InversionOfIntervals018(t *testing.T) {

	Convey("runContainer `invert` operation should do a NOT on the set of intervals, in-place", t, func() {
		seed := int64(42)
		p("seed is %v", seed)
		rand.Seed(seed)

		trials := []trial{
			trial{n: 1000, percentFill: .90, ntrial: 1},
		}

		tester := func(tr trial) {
			for j := 0; j < tr.ntrial; j++ {
				p("TestRle16InversinoOfIntervals018 on check# j=%v", j)
				ma := make(map[int]bool)
				hashNotA := make(map[int]bool)

				n := tr.n
				a := []uint16{}

				// hashNotA will be NOT ma
				//for i := 0; i < n; i++ {
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

				//showArray16(a, "a")
				// too big to print: showHash("hashNotA is not a:", hashNotA)

				// RunContainer's invert
				rc := newRunContainer16FromVals(false, a...)

				p("rc from a is %v", rc)
				p("rc.cardinality = %v", rc.cardinality())
				inv := rc.invert()

				p("inv of a (card=%v) is %v", inv.cardinality(), inv)

				So(inv.cardinality(), ShouldEqual, 1+MaxUint16-rc.cardinality())

				for k := 0; k < n; k++ {
					if hashNotA[k] {
						//p("hashNotA has %v, checking inv", k)
						So(inv.contains(uint16(k)), ShouldBeTrue)
					}
				}

				// skip for now, too big to do 2^16-1
				p("checking for cardinality agreement: inv is %v, len(hashNotA) is %v", inv.getCardinality(), len(hashNotA))
				So(inv.getCardinality(), ShouldEqual, len(hashNotA))
			}
			p("done with randomized invert() check for trial %#v", tr)
		}

		for i := range trials {
			tester(trials[i])
		}

	})
}

func TestRle16SubtractionOfIntervals019(t *testing.T) {

	Convey("runContainer `subtract` operation removes an interval in-place", t, func() {
		// basics

		i22 := interval16{start: 2, last: 2}
		left, _ := i22.subtractInterval(i22)
		So(len(left), ShouldResemble, 0)

		v := interval16{start: 1, last: 6}
		left, _ = v.subtractInterval(interval16{start: 3, last: 4})
		So(len(left), ShouldResemble, 2)
		So(left[0].start, ShouldEqual, 1)
		So(left[0].last, ShouldEqual, 2)
		So(left[1].start, ShouldEqual, 5)
		So(left[1].last, ShouldEqual, 6)

		v = interval16{start: 1, last: 6}
		left, _ = v.subtractInterval(interval16{start: 4, last: 10})
		So(len(left), ShouldResemble, 1)
		So(left[0].start, ShouldEqual, 1)
		So(left[0].last, ShouldEqual, 3)

		v = interval16{start: 5, last: 10}
		left, _ = v.subtractInterval(interval16{start: 0, last: 7})
		So(len(left), ShouldResemble, 1)
		So(left[0].start, ShouldEqual, 8)
		So(left[0].last, ShouldEqual, 10)

		seed := int64(42)
		p("seed is %v", seed)
		rand.Seed(seed)

		trials := []trial{
			trial{n: 1000, percentFill: .90, ntrial: 1},
		}

		tester := func(tr trial) {
			for j := 0; j < tr.ntrial; j++ {
				p("TestRle16SubtractionOfIntervals019 on check# j=%v", j)
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

				//showHash("hash a is:", ma)
				//showHash("hash b is:", mb)
				//showHash("hashAminusB is:", hashAminusB)

				// RunContainer's subtract A - B
				rc := newRunContainer16FromVals(false, a...)
				rcb := newRunContainer16FromVals(false, b...)

				abkup := rc.Clone()

				p("rc from a is %v", rc)
				p("rc.cardinality = %v", rc.cardinality())
				p("rcb from b is %v", rcb)
				p("rcb.cardinality = %v", rcb.cardinality())
				it := rcb.NewRunIterator16()
				for it.HasNext() {
					nx := it.Next()
					rc.isubtract(interval16{start: nx, last: nx})
				}

				// also check full interval subtraction
				for _, p := range rcb.iv {
					abkup.isubtract(p)
				}

				p("rc = a - b; has (card=%v), is %v", rc.cardinality(), rc)
				p("abkup = a - b; has (card=%v), is %v", abkup.cardinality(), abkup)

				for k := range hashAminusB {
					p("hashAminusB has element %v, checking rc and abkup (which are/should be: A - B)", k)
					So(rc.contains(uint16(k)), ShouldBeTrue)
					So(abkup.contains(uint16(k)), ShouldBeTrue)
				}
				p("checking for cardinality agreement: sub is %v, len(hashAminusB) is %v", rc.getCardinality(), len(hashAminusB))
				So(rc.getCardinality(), ShouldEqual, len(hashAminusB))
				p("checking for cardinality agreement: sub is %v, len(hashAminusB) is %v", abkup.getCardinality(), len(hashAminusB))
				So(abkup.getCardinality(), ShouldEqual, len(hashAminusB))

			}
			p("done with randomized subtract() check for trial %#v", tr)
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

	Convey("runContainer `Not` operation should flip the bits of a range on the new returned container", t, func() {
		seed := int64(42)
		p("seed is %v", seed)
		rand.Seed(seed)

		trials := []trial{
			trial{n: 100, percentFill: .8, ntrial: 2},
			/*			trial{n: 10, percentFill: .01, ntrial: 10},
						trial{n: 10, percentFill: .50, ntrial: 10},
						trial{n: 1000, percentFill: .50, ntrial: 10},
						trial{n: 1000, percentFill: .99, ntrial: 10},
			*/
		}

		tester := func(tr trial) {
			for j := 0; j < tr.ntrial; j++ {
				p("TestRle16NotAlsoKnownAsFlipRange021 on check# j=%v", j)

				// what is the interval we are going to flip?

				ma := make(map[int]bool)
				flipped := make(map[int]bool)

				n := tr.n
				a := []uint16{}

				draw := int(float64(n) * tr.percentFill)
				p("draw is %v", draw)
				for i := 0; i < draw; i++ {
					r0 := rand.Intn(n)
					a = append(a, uint16(r0))
					ma[r0] = true
					flipped[r0] = true
					p("draw r0=%v is being added to a and ma", r0)
				}

				// pick an interval to flip
				begin := rand.Intn(n)
				last := rand.Intn(n)
				if last < begin {
					begin, last = last, begin
				}
				p("our interval to flip is [%v, %v]", begin, last)

				// do the flip on the hash `flipped`
				for i := begin; i <= last; i++ {
					if flipped[i] {
						delete(flipped, i)
					} else {
						flipped[i] = true
					}
				}

				//showArray16(a, "a")
				// can be too big to print:
				//showHash("hash (correct) version of flipped is:", flipped)

				// RunContainer's Not
				rc := newRunContainer16FromVals(false, a...)
				flp := rc.Not(begin, last+1)

				//p("rc from a is %v", rc)
				//p("rc.cardinality = %v", rc.cardinality())

				//p("flp of a (has card=%v) is %v. card of our flipped hash is %v", flp.cardinality(), flp, len(flipped))

				So(flp.cardinality(), ShouldEqual, len(flipped))

				for k := 0; k < n; k++ {
					if flipped[k] {
						//p("flipped has %v, checking flp", k)
						So(flp.contains(uint16(k)), ShouldBeTrue)
					} else {
						//p("flipped lacks %v, checking flp", k)
						So(flp.contains(uint16(k)), ShouldBeFalse)
					}
				}

				//p("checking for cardinality agreement: flp is %v, len(flipped) is %v", flp.getCardinality(), len(flipped))
				So(flp.getCardinality(), ShouldEqual, len(flipped))
			}
			//p("done with randomized Not() check for trial %#v", tr)
		}

		for i := range trials {
			tester(trials[i])
		}

	})
}

func TestRleEquals022(t *testing.T) {

	Convey("runContainer `equals` should accurately compare contents against other container types", t, func() {
		seed := int64(42)
		p("seed is %v", seed)
		rand.Seed(seed)

		trials := []trial{
			trial{n: 100, percentFill: .2, ntrial: 10},
			/*
				trial{n: 10, percentFill: .01, ntrial: 10},
				trial{n: 10, percentFill: .50, ntrial: 10},
				trial{n: 1000, percentFill: .50, ntrial: 10},
				trial{n: 1000, percentFill: .99, ntrial: 10},
			*/
		}

		tester := func(tr trial) {
			for j := 0; j < tr.ntrial; j++ {
				p("TestRleEquals022 on check# j=%v", j)

				ma := make(map[int]bool)

				n := tr.n
				a := []uint16{}

				draw := int(float64(n) * tr.percentFill)
				//p("draw is %v", draw)
				for i := 0; i < draw; i++ {
					r0 := rand.Intn(n)
					a = append(a, uint16(r0))
					ma[r0] = true
				}

				//showArray16(a, "a")
				// can be too big to print:
				//showHash("hash (correct) version of flipped is:", flipped)

				rc := newRunContainer16FromVals(false, a...)

				// make bitmap and array versions:
				bc := newBitmapContainer()
				ac := newArrayContainer()
				for k := range ma {
					ac.iadd(uint16(k))
					bc.iadd(uint16(k))
				}

				// compare equals() across all three
				So(rc.equals(ac), ShouldBeTrue)
				So(rc.equals(bc), ShouldBeTrue)

				So(ac.equals(rc), ShouldBeTrue)
				So(ac.equals(bc), ShouldBeTrue)

				So(bc.equals(ac), ShouldBeTrue)
				So(bc.equals(rc), ShouldBeTrue)

				// and for good measure, check against the hash
				So(rc.getCardinality(), ShouldEqual, len(ma))
				So(ac.getCardinality(), ShouldEqual, len(ma))
				So(bc.getCardinality(), ShouldEqual, len(ma))
				for k := range ma {
					So(rc.contains(uint16(k)), ShouldBeTrue)
					So(ac.contains(uint16(k)), ShouldBeTrue)
					So(bc.contains(uint16(k)), ShouldBeTrue)
				}
			}
			p("done with randomized equals() check for trial %#v", tr)
		}

		for i := range trials {
			tester(trials[i])
		}

	})
}

func TestRleIntersects023(t *testing.T) {

	Convey("runContainer `intersects` query should work against any mix of container types", t, func() {
		seed := int64(42)
		p("seed is %v", seed)
		rand.Seed(seed)

		trials := []trial{
			trial{n: 10, percentFill: .293, ntrial: 1000},
			/*
				trial{n: 10, percentFill: .01, ntrial: 10},
				trial{n: 10, percentFill: .50, ntrial: 10},
				trial{n: 1000, percentFill: .50, ntrial: 10},
				trial{n: 1000, percentFill: .99, ntrial: 10},
			*/
		}

		tester := func(tr trial) {
			for j := 0; j < tr.ntrial; j++ {
				p("TestRleIntersects023 on check# j=%v", j)

				ma := make(map[int]bool)
				mb := make(map[int]bool)

				n := tr.n
				a := []uint16{}
				b := []uint16{}

				draw := int(float64(n) * tr.percentFill)
				//p("draw is %v", draw)
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
				//fmt.Printf("isect was %v\n", isect)

				//showArray16(a, "a")
				// can be too big to print:
				//showHash("hash (correct) version of flipped is:", flipped)

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
				So(rcA.intersects(rcB), ShouldEqual, isect)
				So(acA.intersects(acB), ShouldEqual, isect)
				So(bcA.intersects(bcB), ShouldEqual, isect)

				// across types
				So(rcA.intersects(acB), ShouldEqual, isect)
				So(rcA.intersects(bcB), ShouldEqual, isect)

				So(acA.intersects(rcB), ShouldEqual, isect)
				So(acA.intersects(bcB), ShouldEqual, isect)

				So(bcA.intersects(acB), ShouldEqual, isect)
				So(bcA.intersects(rcB), ShouldEqual, isect)

				// and swap the call pattern, so we test B intersects A as well.

				// same type
				So(rcB.intersects(rcA), ShouldEqual, isect)
				So(acB.intersects(acA), ShouldEqual, isect)
				So(bcB.intersects(bcA), ShouldEqual, isect)

				// across types
				So(rcB.intersects(acA), ShouldEqual, isect)
				So(rcB.intersects(bcA), ShouldEqual, isect)

				So(acB.intersects(rcA), ShouldEqual, isect)
				So(acB.intersects(bcA), ShouldEqual, isect)

				So(bcB.intersects(acA), ShouldEqual, isect)
				So(bcB.intersects(rcA), ShouldEqual, isect)

			}
			p("done with randomized intersects() check for trial %#v", tr)
		}

		for i := range trials {
			tester(trials[i])
		}

	})
}

func TestRleToEfficientContainer027(t *testing.T) {

	Convey("runContainer toEfficientContainer should return equivalent containers", t, func() {
		seed := int64(42)
		p("seed is %v", seed)
		rand.Seed(seed)

		// 4096 or fewer integers -> array typically

		trials := []trial{
			trial{n: 8000, percentFill: .01, ntrial: 10},
			trial{n: 8000, percentFill: .99, ntrial: 10},
			/*
				trial{n: 10, percentFill: .01, ntrial: 10},
				trial{n: 10, percentFill: .50, ntrial: 10},
				trial{n: 1000, percentFill: .50, ntrial: 10},
				trial{n: 1000, percentFill: .99, ntrial: 10},
			*/
		}

		tester := func(tr trial) {
			for j := 0; j < tr.ntrial; j++ {
				p("TestRleToEfficientContainer027 on check# j=%v", j)

				ma := make(map[int]bool)

				n := tr.n
				a := []uint16{}

				draw := int(float64(n) * tr.percentFill)
				//p("draw is %v", draw)
				for i := 0; i < draw; i++ {
					r0 := rand.Intn(n)
					a = append(a, uint16(r0))
					ma[r0] = true
				}

				rc := newRunContainer16FromVals(false, a...)

				c := rc.toEfficientContainer()
				So(rc.equals(c), ShouldBeTrue)

				switch tc := c.(type) {
				case *bitmapContainer:
					p("I see a bitmapContainer with card %v", tc.getCardinality())
				case *arrayContainer:
					p("I see a arrayContainer with card %v", tc.getCardinality())
				case *runContainer16:
					p("I see a runContainer16 with card %v", tc.getCardinality())
				}

			}
			p("done with randomized toEfficientContainer() check for trial %#v", tr)
		}

		for i := range trials {
			tester(trials[i])
		}

	})

	Convey("runContainer toEfficientContainer should return an equivalent bitmap when that is efficient", t, func() {

		a := []uint16{}

		// odd intergers should be smallest as a bitmap
		for i := 0; i < MaxUint16; i++ {
			if i%2 == 1 {
				a = append(a, uint16(i))
			}
		}

		rc := newRunContainer16FromVals(false, a...)

		c := rc.toEfficientContainer()
		So(rc.equals(c), ShouldBeTrue)

		_, isBitmapContainer := c.(*bitmapContainer)
		So(isBitmapContainer, ShouldBeTrue)

		switch tc := c.(type) {
		case *bitmapContainer:
			p("I see a bitmapContainer with card %v", tc.getCardinality())
		case *arrayContainer:
			p("I see a arrayContainer with card %v", tc.getCardinality())
		case *runContainer16:
			p("I see a runContainer16 with card %v", tc.getCardinality())
		}

	})
}

func TestRle16RandomFillLeastSignificant16bits029(t *testing.T) {

	Convey("runContainer16.fillLeastSignificant16bits() should fill contents as expected, matching the same function on bitmap and array containers", t, func() {
		seed := int64(42)
		p("seed is %v", seed)
		rand.Seed(seed)

		trials := []trial{
			trial{n: 100, percentFill: .95, ntrial: 1},
		}

		tester := func(tr trial) {
			for j := 0; j < tr.ntrial; j++ {
				p("TestRle16RandomFillLeastSignificant16bits029 on check# j=%v", j)
				ma := make(map[int]bool)

				n := tr.n
				a := []uint16{}

				draw := int(float64(n) * tr.percentFill)
				for i := 0; i < draw; i++ {
					r0 := rand.Intn(n)
					a = append(a, uint16(r0))
					ma[r0] = true
				}

				//showArray16(a, "a")

				// RunContainer
				rc := newRunContainer16FromVals(false, a...)

				p("rc from a is %v", rc)

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

				So(rcOut, ShouldResemble, acOut)
				So(rcOut, ShouldResemble, bcOut)
			}
			p("done with randomized fillLeastSignificant16bits checks for trial %#v", tr)
		}

		for i := range trials {
			tester(trials[i])
		}

	})
}

func TestRle16RandomGetShortIterator030(t *testing.T) {

	Convey("runContainer16.getShortIterator should traverse the contents expected, matching the traversal of the bitmap and array containers", t, func() {
		seed := int64(42)
		p("seed is %v", seed)
		rand.Seed(seed)

		trials := []trial{
			trial{n: 100, percentFill: .95, ntrial: 1},
		}

		tester := func(tr trial) {
			for j := 0; j < tr.ntrial; j++ {
				p("TestRle16RandomGetShortIterator030 on check# j=%v", j)
				ma := make(map[int]bool)

				n := tr.n
				a := []uint16{}

				draw := int(float64(n) * tr.percentFill)
				for i := 0; i < draw; i++ {
					r0 := rand.Intn(n)
					a = append(a, uint16(r0))
					ma[r0] = true
				}

				//showArray16(a, "a")

				// RunContainer
				rc := newRunContainer16FromVals(false, a...)

				p("rc from a is %v", rc)

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
					So(rn, ShouldEqual, an)
					So(rn, ShouldEqual, bn)
				}
			}
			p("done with randomized TestRle16RandomGetShortIterator030 checks for trial %#v", tr)
		}

		for i := range trials {
			tester(trials[i])
		}

	})
}

func TestRle16RandomIaddRangeIremoveRange031(t *testing.T) {

	Convey("runContainer16.iaddRange and iremoveRange should add/remove contents as expected, matching the same operations on the bitmap and array containers and the hashmap pos control", t, func() {
		seed := int64(42)
		p("seed is %v", seed)
		rand.Seed(seed)

		trials := []trial{
			trial{n: 101, percentFill: .9, ntrial: 10},
		}

		tester := func(tr trial) {
			for j := 0; j < tr.ntrial; j++ {
				p("TestRle16RandomIaddRangeIremoveRange031 on check# j=%v", j)
				ma := make(map[int]bool)

				n := tr.n
				a := []uint16{}

				draw := int(float64(n) * tr.percentFill)
				for i := 0; i < draw; i++ {
					r0 := rand.Intn(n)
					a = append(a, uint16(r0))
					ma[r0] = true
				}

				//showArray16(a, "a")

				// RunContainer
				rc := newRunContainer16FromVals(false, a...)

				p("rc from a is %v", rc)

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
				//showHash(fmt.Sprintf("ma after add [%v,%v] and remove [%v,%v] is:", a0, a1, r0, r1), ma)

				//p("before adding range: %s", rc)
				rc.iaddRange(a0, a1+1)
				//p("after adding [%v,%v] range: %s", a0, a1, rc)
				rc.iremoveRange(r0, r1+1)
				//p("after removing [%v,%v] range: %s", r0, r1, rc)

				bc.iaddRange(a0, a1+1)
				bc.iremoveRange(r0, r1+1)

				ac.iaddRange(a0, a1+1)
				ac.iremoveRange(r0, r1+1)
				//p("for comparison, ac is: %s", ac)

				So(rc.getCardinality(), ShouldEqual, len(ma))
				So(rc.getCardinality(), ShouldEqual, ac.getCardinality())
				So(rc.getCardinality(), ShouldEqual, bc.getCardinality())

				rit := rc.getShortIterator()
				ait := ac.getShortIterator()
				bit := bc.getShortIterator()

				for ait.hasNext() {
					rn := rit.next()
					an := ait.next()
					bn := bit.next()
					So(rn, ShouldEqual, an)
					So(rn, ShouldEqual, bn)
				}
				// verify againt the map
				for k := range ma {
					So(rc.contains(uint16(k)), ShouldBeTrue)
				}

				// coverage for run16 method
				So(rc.serializedSizeInBytes(), ShouldEqual, 2+4*rc.numberOfRuns())
			}
			p("done with randomized TestRle16RandomIaddRangeIremoveRange031 checks for trial %#v", tr)
		}

		for i := range trials {
			tester(trials[i])
		}

	})
}

func TestAllContainerMethodsAllContainerTypes065(t *testing.T) {

	Convey("each of the container methods that takes two containers should handle all 3x3==9 possible ways of being called -- without panic", t, func() {
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

	Convey("each of the container methods that takes two containers should handle all 3x3==9 possible ways of being called -- and return results that agree with each other", t, func() {

		//rleVerbose = true

		seed := int64(42)
		p("seed is %v", seed)
		rand.Seed(seed)

		trials := []trial{
			trial{n: 100, percentFill: .7, ntrial: 1, numRandomOpsPass: 100},
			trial{n: 100, percentFill: .7, ntrial: 1, numRandomOpsPass: 100, srang: &interval16{MaxUint16 - 100, MaxUint16}}}

		tester := func(tr trial) {
			for j := 0; j < tr.ntrial; j++ {
				p("TestAllContainerMethodsAllContainerTypesWithData067 on check# j=%v", j)

				a, r, b := getRandomSameThreeContainers(tr)
				a2, r2, b2 := getRandomSameThreeContainers(tr)

				p("prior to any operations, fresh from getRandom...")
				p("receiver (a) is '%s'", a)
				p("receiver (r) is '%s'", r)
				p("receiver (b) is '%s'", b)

				p("\n argument is '%s'\n", r2)

				m := []string{"array", "run", "bitmap"}

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
						c2 := callme[kk+nCalls]   // run reciever
						c3 := callme[kk+2*nCalls] // bitmap receiver

						if c1.name != c2.name {
							panic("internal logic error")
						}
						if c3.name != c2.name {
							panic("internal logic error")
						}

						p("\n ========== testing calls to '%s' all match\n", c1.name)
						for k2, a := range arg {

							p("\n ------------  on arg as '%s': %s\n", m[k2], a)
							p("\n prior to '%s', array receiver c1 is '%s'\n", c1.name, c1.cn)
							p("\n prior to '%s', run receiver c2 is '%s'\n", c2.name, c2.cn)
							p("\n prior to '%s', bitmap receiver c3 is '%s'\n", c3.name, c3.cn)
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

							p("'%s' receiver, '%s' arg, call='%s', res1=%s",
								m[0], m[k2], z, res1)
							p("'%s' receiver, '%s' arg, call='%s', res2=%s",
								m[1], m[k2], z, res2)
							p("'%s' receiver, '%s' arg, call='%s', res3=%s",
								m[2], m[k2], z, res3)

							if !res1.equals(res2) {
								panic(fmt.Sprintf("k:%v, k2:%v, res1 != res2,"+
									" call is '%s'", k, k2, c1.name))
							}
							if !res2.equals(res1) {
								panic(fmt.Sprintf("k:%v, k2:%v, res2 != res1,"+
									" call is '%s'", k, k2, c1.name))
							}
							if !res1.equals(res3) {
								p("res1 = %s", res1)
								p("res3 = %s", res3)
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
			p("done with randomized TestAllContainerMethodsAllContainerTypesWithData067 checks for trial %#v", tr)
		} // end tester

		for i := range trials {
			tester(trials[i])
		}

	})

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
		samp.start = 0
		if n-1 > MaxUint16 {
			panic(fmt.Errorf("n out of range: %v", n))
		}
		samp.last = uint16(n - 1)
	}

	draw := int(float64(n) * tr.percentFill)
	for i := 0; i < draw; i++ {
		r0 := int(samp.start) + rand.Intn(int(samp.runlen()))
		a = append(a, uint16(r0))
		ma[r0] = true
	}

	//showArray16(a, "a")

	rc := newRunContainer16FromVals(false, a...)

	p("rc from a is %v", rc)

	// vs bitmapContainer
	bc := newBitmapContainerFromRun(rc)
	ac := rc.toArrayContainer()

	return ac, rc, bc
}
