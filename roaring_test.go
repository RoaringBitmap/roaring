package goroaring

import (
	"log"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/willf/bitset"
)

/*
interface{
	Add(i int)
	Contains(i int) bool
	AndNot(r Bitmap) Bitmap
	And(r Bitmap) Bitmap
	Xor(r Bitmap) Bitmap
	Or(r Bitmap) Bitmap
	GetCardinality() int
        ToArray()[]int
	Equals(b Bitmap) bool
}
*/
func TestRoaringBitmap(t *testing.T) {
	Convey("Test Contains", t, func() {
		rbm1 := NewRoaringBitmap()
		for k := 0; k < 1000; k++ {
			rbm1.Add(17 * k)
		}
		for k := 0; k < 17*1000; k++ {
			So(rbm1.Contains(k), ShouldEqual, (k/17*17 == k))
		}
	})

	Convey("Test ANDNOT", t, func() {
		rr := NewRoaringBitmap()
		for k := 4000; k < 4256; k++ {
			rr.Add(k)
		}
		for k := 65536; k < 65536+4000; k++ {
			rr.Add(k)
		}
		for k := 3 * 65536; k < 3*65536+9000; k++ {
			rr.Add(k)
		}
		for k := 4 * 65535; k < 4*65535+7000; k++ {
			rr.Add(k)
		}
		for k := 6 * 65535; k < 6*65535+10000; k++ {
			rr.Add(k)
		}
		for k := 8 * 65535; k < 8*65535+1000; k++ {
			rr.Add(k)
		}
		for k := 9 * 65535; k < 9*65535+30000; k++ {
			rr.Add(k)
		}

		rr2 := NewRoaringBitmap()
		for k := 4000; k < 4256; k++ {
			rr2.Add(k)
		}
		for k := 65536; k < 65536+4000; k++ {
			rr2.Add(k)
		}
		for k := 3*65536 + 2000; k < 3*65536+6000; k++ {
			rr2.Add(k)
		}
		for k := 6 * 65535; k < 6*65535+1000; k++ {
			rr2.Add(k)
		}
		for k := 7 * 65535; k < 7*65535+1000; k++ {
			rr2.Add(k)
		}
		for k := 10 * 65535; k < 10*65535+5000; k++ {
			rr2.Add(k)
		}
		correct := AndNot(rr, rr2)
		rr.AndNot(rr2)
		So(correct.Equals(rr), ShouldEqual, true)
	})

	Convey("Test ANDNOT4", t, func() {
		rb := NewRoaringBitmap()
		rb2 := NewRoaringBitmap()

		for i := 0; i < 200000; i += 4 {
			rb2.Add(i)
		}
		for i := 200000; i < 400000; i += 14 {
			rb2.Add(i)
		}

		rb2.GetCardinality() //TODO:why is this present?

		andNotresult := AndNot(rb, rb2)
		off := AndNot(rb2, rb)

		So(rb.Equals(andNotresult), ShouldEqual, andNotresult)
		So(rb2.Equals(off), ShouldEqual, true)
		rb2.AndNot(rb)
		So(rb2.Equals(off), ShouldEqual, true)

	})

	Convey("Test AND", t, func() {
		rr := NewRoaringBitmap()
		for k := 0; k < 4000; k++ {
			rr.Add(k)
		}
		rr.Add(100000)
		rr.Add(110000)
		rr2 := NewRoaringBitmap()
		rr2.Add(13)
		rrand := RoaringBitmap.And(rr, rr2)
		array := rrand.ToArray()

		So(len(array), ShouldEqual, 1)
		So(array[0], ShouldEqual, 13)
		rr.And(rr2)
		array = rr.ToArray()

		So(len(array), ShouldEqual, 1)
		So(array[0], ShouldEqual, 13)
	})

	Convey("Test AND 2", t, func() {
		rr := NewRoaringBitmap()
		for k := 4000; k < 4256; k++ {
			rr.Add(k)
		}
		for k := 65536; k < 65536+4000; k++ {
			rr.Add(k)
		}
		for k := 3 * 65536; k < 3*65536+9000; k++ {
			rr.Add(k)
		}
		for k := 4 * 65535; k < 4*65535+7000; k++ {
			rr.Add(k)
		}
		for k := 6 * 65535; k < 6*65535+10000; k++ {
			rr.Add(k)
		}
		for k := 8 * 65535; k < 8*65535+1000; k++ {
			rr.Add(k)
		}
		for k := 9 * 65535; k < 9*65535+30000; k++ {
			rr.Add(k)
		}

		rr2 := NewRoaringBitmap()
		for k := 4000; k < 4256; k++ {
			rr2.Add(k)
		}
		for k := 65536; k < 65536+4000; k++ {
			rr2.Add(k)
		}
		for k := 3*65536 + 2000; k < 3*65536+6000; k++ {
			rr2.Add(k)
		}
		for k := 6 * 65535; k < 6*65535+1000; k++ {
			rr2.Add(k)
		}
		for k := 7 * 65535; k < 7*65535+1000; k++ {
			rr2.Add(k)
		}
		for k := 10 * 65535; k < 10*65535+5000; k++ {
			rr2.Add(k)
		}
		correct := RoaringBitmap.And(rr, rr2)
		rr.And(rr2)
		So(correct.Equals(rr), ShouldEqual, true)
	})

	Convey("Test AND 2", t, func() {
		rr := NewRoaringBitmap()
		for k := 0; k < 4000; k++ {
			rr.Add(k)
		}
		rr.Add(100000)
		rr.Add(110000)
		rr2 := NewRoaringBitmap()
		rr2.Add(13)

		rrand := RoaringBitmap.And(rr, rr2)
		array := rrand.ToArray()
		So(len(array), ShouldEqual, 1)
		So(array[0], ShouldEqual, 13)
	})
	Convey("Test AND 3", t, func() {
		var arrayand [11256]int
		rr := NewRoaringBitmap()
		for k := 4000; k < 4256; k++ {
			rr.Add(k)
		}
		for k := 65536; k < 65536+4000; k++ {
			rr.Add(k)
		}
		for k := 3 * 65536; k < 3*65536+1000; k++ {
			rr.Add(k)
		}
		for k := 3*65536 + 1000; k < 3*65536+7000; k++ {
			rr.Add(k)
		}
		for k := 3*65536 + 7000; k < 3*65536+9000; k++ {
			rr.Add(k)
		}
		for k := 4 * 65536; k < 4*65536+7000; k++ {
			rr.Add(k)
		}
		for k := 6 * 65536; k < 6*65536+10000; k++ {
			rr.Add(k)
		}
		for k := 8 * 65536; k < 8*65536+1000; k++ {
			rr.Add(k)
		}
		for k := 9 * 65536; k < 9*65536+30000; k++ {
			rr.Add(k)
		}

		pos := 0
		rr2 := NewRoaringBitmap()
		for k := 4000; k < 4256; k++ {
			rr2.Add(k)
			arrayand[pos] = k
			pos++
		}
		for k := 65536; k < 65536+4000; k++ {
			rr2.Add(k)
			arrayand[pos] = k
			pos++
		}
		for k := 3*65536 + 1000; k < 3*65536+7000; k++ {
			rr2.Add(k)
			arrayand[pos] = k
			pos++
		}
		for k := 6 * 65536; k < 6*65536+1000; k++ {
			rr2.Add(k)
			arrayand[pos] = k
			pos++
		}
		for k := 7 * 65536; k < 7*65536+1000; k++ {
			rr2.Add(k)
		}
		for k := 10 * 65536; k < 10*65536+5000; k++ {
			rr2.Add(k)
		}

		rrand := RoaringBitmap.And(rr, rr2)
		arrayres := rrand.ToArray()

		for i, _ := range arrayres {
			if arrayres[i] != arrayand[i] {
				log.Println(arrayres[i])
			}
		}

		So(arrayand, ShouldEqual, arrayres)

	})

	Convey("Test AND 4", t, func() {
		rb = NewRoaringBitmap()
		rb2 = NewRoaringBitmap()

		for i := 0; i < 200000; i += 4 {
			rb2.Add(i)
		}
		for i := 200000; i < 400000; i += 14 {
			rb2.Add(i)
		}
		//TODO: RoaringBitmap.And(bm,bm2)
		andresult := RoaringBitmap.And(rb, rb2)
		off := RoaringBitmap.And(rb2, rb)
		So(andresult.Equals(off), ShouldEqual, true)
		So(andresult.GetCardinality(), ShouldEqual, 0)

		for i := 500000; i < 600000; i += 14 {
			rb.Add(i)
		}
		for i := 200000; i < 400000; i += 3 {
			rb2.Add(i)
		}
		andresult2 := RoaringBitmap.And(rb, rb2)
		So(andresult.GetCardinality(), ShouldEqual, 0)
		So(andresult2.GetCardinality(), ShouldEqual, 0)

		for i := 0; i < 200000; i += 4 {
			rb.Add(i)
		}
		for i := 200000; i < 400000; i += 14 {
			rb.Add(i)
		}
		So(andresult.GetCardinality(), ShouldEqual, 0)
		rc := RoaringBitmap.And(rb, rb2)
		rb.And(rb2)
		So(rc.GetCardinality(), ShouldEqual, rb.GetCardinality())

	})

	Convey("ArrayContainerCardinalityTest", t, func() {
		ac := NewArrayContainer()
		for k := int16(0); k < 100; k++ {
			ac.Add(k)
			So(ac.GetCardinality(), ShouldEqual, k+1)
		}
		for k := int16(0); k < 100; k++ {
			ac.Add(k)
			So(ac.GetCardinality(), ShouldEqual, 100)
		}
	})
	/*
		Convey("ArrayTest", t, func() {
			rr := NewArrayContainer()
			rr.Add(int16(110))
			rr.Add(int16(114))
			rr.Add(int16(115))
			var array [3]int16
			for pos:=0; pos<rr.Size();pos++ {
				array[pos++] = rr.Get(pos)?
			}
			So(array[0], ShouldEqual, int16(110))
			So(array[1], ShouldEqual, int16(114))
			So(array[2], ShouldEqual, int16(115))
		})
	*/
	Convey("basic test", t, func() {
		rr := NewRoaringBitmap()
		var a [4002]int
		pos := 0
		for k := 0; k < 4000; k++ {
			rr.Add(k)
			a[pos] = k
			pos++
		}
		rr.Add(100000)
		a[pos] = 100000
		pos++
		rr.Add(110000)
		a[pos] = 110000
		pos++
		array := rr.ToArray()
		for i, _ := range a {
			if array[i] != a[i] {
				log.Println("rr : ", array[i], " a : ", a[i])
			}
		}
		So(array, ShouldEqual, a)
	})

	Convey("BitmapContainerCardinalityTest", t, func() {
		ac := NewBitmapContainer()
		for k := int16(0); k < 100; k++ {
			ac.Add(k)
			So(ac.GetCardinality(), ShouldEqual, k+1)
		}
		for k := int16(0); k < 100; k++ {
			ac.Add(k)
			So(ac.GetCardinality(), ShouldEqual, 100)
		}
	})

	Convey("BitmapContainerTest", t, func() {
		rr := NewBitmapContainer()
		rr.Add(int16(110))
		rr.Add(int16(114))
		rr.Add(int16(115))
		var array [3]int16
		for pos := 0; pos < rr.Size(); pos++ {
			array[pos] = rr.Get(pos)
			pos++
		}
		So(array[0], ShouldEqual, int16(110))
		So(array[1], ShouldEqual, int16(114))
		So(array[2], ShouldEqual, int16(115))
	})
	Convey("cardinality test", t, func() {
		N := 1024
		for gap := 7; gap < 100000; gap *= 10 {
			for offset := 2; offset <= 1024; offset *= 2 {
				rb := NewRoaringBitmap()
				for k := 0; k < N; k++ {
					rb.Add(k * gap)
					So(rb.GetCardinality(), ShouldEqual, k+1)
				}
				So(rb.GetCardinality(), ShouldEqual, N)
				// check the add of existing values
				for k := 0; k < N; k++ {
					rb.Add(k * gap)
					So(rb.GetCardinality(), ShouldEqual, N)
				}

				rb2 := NewRoaringBitmap()

				for k := 0; k < N; k++ {
					rb2.Add(k * gap * offset)
					So(rb2.GetCardinality(), ShouldEqual, k+1)
				}

				So(rb2.GetCardinality(), ShouldEqual, N)

				for k := 0; k < N; k++ {
					rb2.Add(k * gap * offset)
					So(rb2.GetCardinality(), ShouldEqual, N)
				}
				So(RoaringBitmap.And(rb, rb2).GetCardinality(), ShouldEqual, N/offset)
				So(RoaringBitmap.Or(rb, rb2).GetCardinality(), ShouldEqual, 2*N-N/offset)
				So(RoaringBitmap.Xor(rb, rb2).GetCardinality(), ShouldEqual, 2*N-2*N/offset)
			}
		}
	})

	Convey("clear test", t, func() {
		rb := NewRoaringBitmap()
		for i := 0; i < 200000; i += 7 {
			// dense
			rb.Add(i)
		}
		for i := 200000; i < 400000; i += 177 {
			// sparse
			rb.Add(i)
		}

		rb2 := NewRoaringBitmap()
		rb3 := NewRoaringBitmap()
		for i := 0; i < 200000; i += 4 {
			rb2.Add(i)
		}
		for i := 200000; i < 400000; i += 14 {
			rb2.Add(i)
		}

		rb.Clear()
		So(rb.GetCardinality(), ShouldEqual, 0)
		So(rb2.GetCardinality(), ShouldNotEqual, 0)

		rb.Add(4)
		rb3.Add(4)
		andresult := RoaringBitmap.And(rb, rb2)
		orresult := RoaringBitmap.Or(rb, rb2)

		So(andresult.GetCardinality(), ShouldEqual, 1)
		So(oresult.GetCardinality(), ShouldEqual, rb2.GetCardinality())

		for i := 0; i < 200000; i += 4 {
			rb.Add(i)
			rb3.Add(i)
		}
		for i := 200000; i < 400000; i += 114 {
			rb.Add(i)
			rb3.Add(i)
		}

		arrayrr := rb.ToArray()
		arrayrr3 := rb3.ToArray()

		So(arrayrr, ShouldEqual, arrayrr3)
	})

	Convey("constainer factory ", t, func() {

		bc1 := NewBitmapContainer()
		bc2 := NewBitmapContainer()
		bc3 := NewBitmapContainer()
		ac1 := NewArrayContainer()
		ac2 := NewArrayContainer()
		ac3 := NewArrayContainer()

		for i := 0; i < 5000; i++ {
			bc1.Add(int16(i * 70))
		}
		for i := 0; i < 5000; i++ {
			bc2.Add(int16(i * 70))
		}
		for i := 0; i < 5000; i++ {
			bc3.Add(int16(i * 70))
		}
		for i := 0; i < 4000; i++ {
			ac1.add(uint16(i * 50))
		}
		for i := 0; i < 4000; i++ {
			ac2.add(uint16(i * 50))
		}
		for i := 0; i < 4000; i++ {
			ac3.Add(uint16(i * 50))
		}

		rbc := ac1.clone().ToBitmapContainer()
		So(validate(rbc, ac1), ShouldEqual, true)
		rbc = ac2.clone().ToBitmapContainer()
		So(validate(rbc, ac2), ShouldEqual, true)
		rbc = ac3.clone().ToBitmapContainer()
		So(validate(rbc, ac3), ShouldEqual, true)
	})
	Convey("flipTest1 ", t, func() {
		rb := NewRoaringBitmap()
		rb.Flip(100000, 200000) // in-place on empty bitmap
		rbcard := rb.GetCardinality()
		So(100000, ShouldEqual, rbcard)

		bs := bitset.New(20000 - 10000)
		for i := 100000; i < 200000; i++ {
			bs.set(i)
		}
		So(equals(bs, rb), ShouldEqual, true)
	})

	Convey("flipTest1A", t, func() {
		rb := NewRoaringBitmap()
		rb1 := RoaringBitmap.Flip(rb, 100000, 200000)
		rbcard := rb1.GetCardinality()
		So(100000, ShouldEqual, rbcard)
		So(0, ShouldEqual, rb.GetCardinality())

		bs := bitset.New()
		So(equals(bs, rb), ShouldEqual, true)

		for i := 100000; i < 200000; i++ {
			bs.Set(i)
		}
		So(equals(bs, rb1), ShouldEqual, true)
	})
	Convey("flipTest2", t, func() {
		rb := NewRoaringBitmap()
		rb.Flip(100000, 100000)
		rbcard := rb.GetCardinality()
		So(0, ShouldEqual, rbcard)

		bs := bitset.New()
		So(equals(bs, rb), ShouldEqual, true)
	})

	Convey("flipTest2A", t, func() {
		rb := NewRoaringBitmap()
		rb1 := RoaringBitmap.Flip(rb, 100000, 100000)

		rb.Add(1)
		rbcard := rb1.GetCardinality()

		So(0, ShouldEqual, rbcard)
		So(1, ShouldEqual, rb.GetCardinality())

		bs := bitset.New()
		So(equals(bs, rb1), ShouldEqual, true)
		bs.Set(1)
		So(equals(bs, rb), ShouldEqual, true)
	})

	Convey("flipTest3A", t, func() {
		rb := NewRoaringBitmap()
		rb.Flip(100000, 200000) // got 100k-199999
		rb.Flip(100000, 199991) // give back 100k-199990
		rbcard := rb.GetCardinality()
		So(9, ShouldEqual, rbcard)

		bs := bitset.New()
		for i := 199991; i < 200000; i++ {
			bs.Set(i)
		}

		So(equals(bs, rb), ShouldEqual, true)
	})

	Convey("flipTest4A", t, func() {
		// fits evenly on both ends
		rb := NewRoaringBitmap()
		rb.Flip(100000, 200000) // got 100k-199999
		rb.Flip(65536, 4*65536)
		rbcard := rb.GetCardinality()

		// 65536 to 99999 are 1s
		// 200000 to 262143 are 1s: total card

		So(96608, ShouldEqual, rbcard)

		bs := bitset.New()
		for i := 65536; i < 100000; i++ {
			bs.Set(i)
		}
		for i := 200000; i < 262144; i++ {
			bs.Set(i)
		}

		So(equals(bs, rb), ShouldEqual, true)
	})

	Convey("flipTest5", t, func() {
		// fits evenly on small end, multiple
		// containers
		rb := NewRoaringBitmap()
		rb.Flip(100000, 132000)
		rb.Flip(65536, 120000)
		rbcard := rb.GetCardinality()

		// 65536 to 99999 are 1s
		// 120000 to 131999

		Assert.assertEquals(46464, rbcard)
		So(46464, ShouldEqual, rbcard)

		bs := NewBitSet()
		for i := 65536; i < 100000; i++ {
			bs.Set(i)
		}
		for i := 120000; i < 132000; i++ {
			bs.Set(i)
		}
		So(equals(bs, rb), ShouldEqual, true)
	})

	Convey("flipTest6", t, func() {
		rb := NewRoaringBitmap()
		rb1 := RoaringBitmap.Flip(rb, 100000, 132000)
		rb2 := RoaringBitmap.Flip(rb1, 65536, 120000)
		rbcard := rb2.GetCardinality()

		bs := NewBitSet()
		for i := 65536; i < 100000; i++ {
			bs.Set(i)
		}
		for i := 120000; i < 132000; i++ {
			bs.Set(i)
		}
		So(equals(bs, rb2), ShouldEqual, true)
	})

	Convey("flipTest6A", t, func() {
		rb := NewRoaringBitmap()
		rb1 := RoaringBitmap.Flip(rb, 100000, 132000)
		rb2 := RoaringBitmap.Flip(rb1, 99000, 2*65536)
		rbcard := rb2.GetCardinality()

		So(1928, ShouldEqual, rbcard)

		bs := NewBitSet()
		for i := 99000; i < 100000; i++ {
			bs.Set(i)
		}
		for i := 2 * 65536; i < 132000; i++ {
			bs.Set(i)
		}
		So(equals(bs, rb2), ShouldEqual, true)
	})

	Convey("flipTest7", t, func() {
		// within 1 word, first container
		rb := NewRoaringBitmap()
		rb.Flip(650, 132000)
		rb.Flip(648, 651)
		rbcard := rb.GetCardinality()

		// 648, 649, 651-131999

		So(132000-651+2, ShouldEqual, rbcard)
		bs := NewBitSet()
		bs.Set(648)
		bs.Set(649)
		for i := 651; i < 132000; i++ {
			bs.Set(i)
		}
		So(equals(bs, rb), ShouldEqual, true)
	})
	Convey("flipTestBig", t, func() {
		numCases := 1000
		rb := NewRoaringBitmap()
		bs := NewBitSet()
		//Random r = new Random(3333);
		checkTime := 2

		for i := 0; i < numCases; i++ {
			start := rand.Intn(65536 * 20)
			end := rand.Intn(65536 * 20)
			if rand.Float64() < float64(0.1) {
				end = start + rand.Intn(100)
			}
			rb.Flip(start, end)
			if start < end {
				bs.Flip(start, end) // throws exception
			}
			// otherwise
			// insert some more ANDs to keep things sparser
			if rand.Float64() < 0.2 {
				mask := NewRoaringBitmap()
				mask1 := NewBitSet()
				startM := rand.Intn(65536 * 20)
				endM := startM + 100000
				mask.Flip(startM, endM)
				mask1.Flip(startM, endM)
				mask.Flip(0, 65536*20+100000)
				mask1.Flip(0, 65536*20+100000)
				rb.And(mask)
				bs.And(mask1)
			}
			// see if we can detect incorrectly shared containers
			if rand.Float64() < 0.1 {
				irrelevant := RoaringBitmap.Flip(rb, 10, 100000)
				irrelevant.Flip(5, 200000)
				irrelevant.Flip(190000, 260000)
			}
			if i > checkTime {
				So(equals(bs, rb), ShouldEqual, true)
				checkTime *= 1.5
			}
		}
	})

	Convey("flipTestBigA", t, func() {
		numCases := 1000
		bs := NewBitSet()
		checkTime := 2
		rb1 := NewRoaringBitmap()

		for i := 0; i < numCases; i++ {
			start := rand.Intn(65536 * 20)
			end := rand.Intn(65536 * 20)
			if rand.Floag64() < 0.1 {
				end = start + rand.Intn(100)
			}

			if (i & 1) == 0 {
				rb2 = RoaringBitmap.Flip(rb1, start, end)
				// tweak the other, catch bad sharing
				rb1.Flip(rand.Intn(65536*20), rand.Intn(65536*20))
			} else {
				rb1 = RoaringBitmap.Flip(rb2, start, end)
				rb2.Flip(rand.Intn(65536*20), rand.Intn(65536*20))
			}

			if start < end {
				bs.Flip(start, end) // throws exception
			}
			// otherwise
			// insert some more ANDs to keep things sparser
			if (rand.Float64() < 0.2) && (i&1) == 0 {
				mask := NewRoaringBitmap()
				mask1 := NewBitSet()
				startM := rand.Intn(65536 * 20)
				endM := startM + 100000
				mask.Flip(startM, endM)
				mask1.Flip(startM, endM)
				mask.Flip(0, 65536*20+100000)
				mask1.Flip(0, 65536*20+100000)
				rb2.And(mask)
				bs.And(mask1)
			}

			if i > checkTime {
				var rb RoaringBitmap

				if (i & 1) == 0 {
					rb = rb2
				} else {
					rb = rb1
				}
				So(equals(bs, rb), ShouldEqual, true)
				checkTime *= 1.5
			}
		}
	})

	Convey("ortest", t, func() {
		rr := NewRoaringBitmap()
		for k := 0; k < 4000; k++ {
			rr.Add(k)
		}
		rr.Add(100000)
		rr.Add(110000)
		rr2 := NewRoaringBitmap()
		for k := 0; k < 4000; k++ {
			rr2.Add(k)
		}

		rror := RoaringBitmap.Or(rr, rr2)

		array := rror.ToArray()
		arrayrr := rr.ToArray()

		//Assert.assertTrue(Arrays.equals(array, arrayrr));
		//So(equals(bs, rb), ShouldEqual, true)

		rr.Or(rr2)
		arrayirr := rr.ToArray()
		So(equals(array, arrayirr), ShouldEqual, true)
	})

	Convey("ORtest", t, func() {
		rr := NewRoaringBitmap()
		for k := 4000; k < 4256; k++ {
			rr.Add(k)
		}
		for k := 65536; k < 65536+4000; k++ {
			rr.Add(k)
		}
		for k := 3 * 65536; k < 3*65536+9000; k++ {
			rr.Add(k)
		}
		for k := 4 * 65535; k < 4*65535+7000; k++ {
			rr.Add(k)
		}
		for k := 6 * 65535; k < 6*65535+10000; k++ {
			rr.Add(k)
		}
		for k := 8 * 65535; k < 8*65535+1000; k++ {
			rr.Add(k)
		}
		for k := 9 * 65535; k < 9*65535+30000; k++ {
			rr.Add(k)
		}

		rr2 := NewRoaringBitmap()
		for k := 4000; k < 4256; k++ {
			rr2.Add(k)
		}
		for k := 65536; k < 65536+4000; k++ {
			rr2.Add(k)
		}
		for k := 3*65536 + 2000; k < 3*65536+6000; k++ {
			rr2.Add(k)
		}
		for k := 6 * 65535; k < 6*65535+1000; k++ {
			rr2.Add(k)
		}
		for k := 7 * 65535; k < 7*65535+1000; k++ {
			rr2.Add(k)
		}
		for k := 10 * 65535; k < 10*65535+5000; k++ {
			rr2.Add(k)
		}
		correct := RoaringBitmap.Or(rr, rr2)
		rr.Or(rr2)
		So(equals(correct, rr), ShouldEqual, true)
	})

	Convey("ortest2", t, func() {
		var arrayrr [4000 + 4000 + 2]int
		pos := 0
		rr := NewRoaringBitmap()
		for k := 0; k < 4000; k++ {
			rr.Add(k)
			arrayrr[pos] = k
			pos++
		}
		rr.Add(100000)
		rr.Add(110000)
		rr2 := NewRoaringBitmap()
		for k := 4000; k < 8000; k++ {
			rr2.Add(k)
			arrayrr[pos] = k
			pos++
		}

		arrayrr[pos] = 100000
		pos++
		arrayrr[pos] = 110000
		pos++

		rror := RoaringBitmap.Or(rr, rr2)

		arrayor := rror.ToArray()

		So(equals(arrayor, arrayrr), ShouldEqual, true)
	})

	Convey("ortest3", t, func() {
		var V1 map[int]bool
		var V2 map[int]bool

		rr := NewRoaringBitmap()
		rr2 := NewRoaringBitmap()
		for k := 0; k < 4000; k++ {
			rr2.Add(k)
			V1[k] = true
		}
		for k := 3500; k < 4500; k++ {
			rr.Add(k)
			V1[k] = true
		}
		for k := 4000; k < 65000; k++ {
			rr2.Add(k)
			V1[k] = true
		}

		// In the second node of each roaring bitmap, we have two bitmap
		// containers.
		// So, we will check the union between two BitmapContainers
		for k := 65536; k < 65536+10000; k++ {
			rr.Add(k)
			V1[k] = true
		}

		for k := 65536; k < 65536+14000; k++ {
			rr2.Add(k)
			V1[k] = true
		}

		// In the 3rd node of each Roaring Bitmap, we have an
		// ArrayContainer, so, we will try the union between two
		// ArrayContainers.
		for k := 4 * 65535; k < 4*65535+1000; k++ {
			rr.Add(k)
			V1[k] = true
		}

		for k := 4 * 65535; k < 4*65535+800; k++ {
			rr2.Add(k)
			V1[k] = true
		}

		// For the rest, we will check if the union will take them in
		// the result
		for k := 6 * 65535; k < 6*65535+1000; k++ {
			rr.Add(k)
			V1[k] = true
		}

		for k := 7 * 65535; k < 7*65535+2000; k++ {
			rr2.Add(k)
			V1[k] = true
		}

		rror := RoaringBitmap.or(rr, rr2)
		valide := true

		// Si tous les elements de rror sont dans V1 et que tous les
		// elements de
		// V1 sont dans rror(V2)
		// alors V1 == rror

		for i, k := range rror.ToArray() {
			_, found := V1[k]
			if !found {
				valide = false
			}
			V2[k] = true
		}

		for k, v := range V1 {
			_, found := V2[k]
			if !found {
				valide = false
			}
		}

		So(valide, ShouldEqual, true)
	})

	Convey("ortest4", t, func() {
		rb := NewRoaringBitmap()
		rb2 := NewRoaringBitmap()

		for i := 0; i < 200000; i += 4 {
			rb2.Add(i)
		}
		for i := 200000; i < 400000; i += 14 {
			rb2.Add(i)
		}
		rb2card := rb2.GetCardinality()

		// check or against an empty bitmap
		orresult := RoaringBitmap.Or(rb, rb2)
		off := RoaringBitmap.Or(rb2, rb)
		So(equals(orresult, off), ShouldEqual, true)

		So(rb2card, ShouldEqual, orresult.GetCardinality())

		for i := 500000; i < 600000; i += 14 {
			rb.Add(i)
		}
		for i := 200000; i < 400000; i += 3 {
			rb2.Add(i)
		}
		// check or against an empty bitmap
		orresult2 := RoaringBitmap.Or(rb, rb2)
		So(rb2card, ShouldEqual, orresult.GetCardinality())
		So(rb2.GetCardinality()+rb.GetCardinality(), ShouldEqual,
			orresult2.GetCardinality())
		rb.Or(rb2)
		So(equals(rb, orresult2), ShouldEqual, true)

	})

	Convey("randomTest", t, func() {
		rTest(15)
		rTest(1024)
		rTest(4096)
		rTest(65536)
		rTest(65536 * 16)
	})

	Convey("SimpleCardinality", t, func() {
		N := 512
		gap := 70

		rb := NewRoaringBitmap()
		for k := 0; k < N; k++ {
			rb.Add(k * gap)
			So(rb.GetCardinality(), ShouldEqual, k+1)
		}
		So(rb.GetCardinality(), ShouldEqual, N)
		for k := 0; k < N; k++ {
			rb.Add(k * gap)
			So(rb.GetCardinality(), ShouldEqual, N)
		}

	})

	Convey("XORtest", t, func() {
		rr := NewRoaringBitmap()
		for k := 4000; k < 4256; k++ {
			rr.Add(k)
		}
		for k := 65536; k < 65536+4000; k++ {
			rr.Add(k)
		}
		for k := 3 * 65536; k < 3*65536+9000; k++ {
			rr.Add(k)
		}
		for k := 4 * 65535; k < 4*65535+7000; k++ {
			rr.Add(k)
		}
		for k := 6 * 65535; k < 6*65535+10000; k++ {
			rr.Add(k)
		}
		for k := 8 * 65535; k < 8*65535+1000; k++ {
			rr.Add(k)
		}
		for k := 9 * 65535; k < 9*65535+30000; k++ {
			rr.Add(k)
		}

		rr2 := NewRoaringBitmap()
		for k := 4000; k < 4256; k++ {
			rr2.Add(k)
		}
		for k := 65536; k < 65536+4000; k++ {
			rr2.Add(k)
		}
		for k := 3*65536 + 2000; k < 3*65536+6000; k++ {
			rr2.Add(k)
		}
		for k := 6 * 65535; k < 6*65535+1000; k++ {
			rr2.Add(k)
		}
		for k := 7 * 65535; k < 7*65535+1000; k++ {
			rr2.Add(k)
		}
		for k := 10 * 65535; k < 10*65535+5000; k++ {
			rr2.Add(k)
		}
		correct := RoaringBitmap.Xor(rr, rr2)
		rr.Xor(rr2)
		So(equals(correct, rr), ShouldEquals, true)
	})

	Convey("xortest1", t, func() {
		var V1 map[int]bool
		var V2 map[int]bool

		rr := NewRoaringBitmap()
		rr2 := NewRoaringBitmap()
		// For the first 65536: rr2 has a bitmap container, and rr has
		// an array container.
		// We will check the union between a BitmapCintainer and an
		// arrayContainer
		for k := 0; k < 4000; k++ {
			rr2.Add(k)
			if k < 3500 {
				V1[k] = true
			}
		}
		for k := 3500; k < 4500; k++ {
			rr.Add(k)
		}
		for k := 4000; k < 65000; k++ {
			rr2.Add(k)
			if k >= 4500 {
				V1[k] = true
			}
		}

		for k := 65536; k < 65536+30000; k++ {
			rr.Add(k)
		}

		for k := 65536; k < 65536+50000; k++ {
			rr2.Add(k)
			if k >= 65536+30000 {
				V1[k] = true
			}
		}

		// In the 3rd node of each Roaring Bitmap, we have an
		// ArrayContainer. So, we will try the union between two
		// ArrayContainers.
		for k := 4 * 65535; k < 4*65535+1000; k++ {
			rr.Add(k)
			if k >= (4*65535 + 800) {
				V1[k] = true
			}
		}

		for k := 4 * 65535; k < 4*65535+800; k++ {
			rr2.Add(k)
		}

		for k := 6 * 65535; k < 6*65535+1000; k++ {
			rr.Add(k)
			V1[k] = true
		}

		for k := 7 * 65535; k < 7*65535+2000; k++ {
			rr2.Add(k)
			V1[k] = true
		}

		rrxor := RoaringBitmap.Xor(rr, rr2)
		valide := true

		// Si tous les elements de rror sont dans V1 et que tous les
		// elements de
		// V1 sont dans rror(V2)
		// alors V1 == rror

		for _, i := range rrxor.ToArray() {
			_, found := V1[i]
			if !found {
				valide = false
			}
			V2[k] = true
		}
		for k, v := range V1 {
			_, found := V2[v]
			if !found {
				valide = false
			}
		}

		So(valide, ShouldEqual, true)
	})

	Convey("XORtest 4", t, func() {
		rb := NewRoaringBitmap()
		rb2 := NewRoaringBitmap()

		for i := 0; i < 200000; i += 4 {
			rb2.Add(i)
		}
		for i := 200000; i < 400000; i += 14 {
			rb2.Add(i)
		}
		rb2card := rb2.GetCardinality()

		// check or against an empty bitmap
		xorresult := RoaringBitmap.Xor(rb, rb2)
		off := RoaringBitmap.Or(rb2, rb)
		So(equals(xorresult, off), ShouldEqual, true)

		So(rb2card, ShouldEqual, xorresult.GetCardinality())

		for i := 500000; i < 600000; i += 14 {
			rb.Add(i)
		}
		for i := 200000; i < 400000; i += 3 {
			rb2.Add(i)
		}
		// check or against an empty bitmap
		xorresult2 := RoaringBitmap.Xor(rb, rb2)
		So(rb2card, ShouldEqual, xorresult.GetCardinality())
		So(rb2.GetCardinality()+rb.getCardinality(), ShouldEqual, xorresult2.GetCardinality())

		rb.Xor(rb2)
		So(equals(xorresult2, rb), ShouldEqual, true)

	})
	//need to add the massives
}

func rTest(N int) {
	log.Println("rtest N=" + N)
	for gap := 1; gap <= 65536; gap *= 2 {
		bs1 := NewBitSet()
		rb1 := NewRoaringBitmap()
		for x := 0; x <= N; x += gap {
			bs1.Set(x)
			rb1.Add(x)
		}
		So(bs1.Cardinality(), ShouldEqual, rb1.GetCardinality())
		So(equals(bs1, rb1), ShouldEqual, true)
		for offset := 1; offset <= gap; offset *= 2 {
			bs2 := NewBitSet()
			rb2 := NewRoaringBitmap()
			for x := 0; x <= N; x += gap {
				bs2.Set(x + offset)
				rb2.Add(x + offset)
			}
			So(bs2.Cardinality(), ShouldEqual, rb2.GetCardinality())
			So(equals(bs2, rb2), ShouldEqual, true)

			clonebs1 := bs1.Clone()
			clonebs1.And(bs2)
			if !equals(clonebs1, RoaringBitmap.And(rb1, rb2)) {
				t := rb1.Clone()
				t.And(rb2)
				So(equals(clonebs1, t), ShouldEqual, true)
			}

			// testing OR
			clonebs1 = bs1.Clone()
			clonebs1.Or(bs2)

			So(equals(clonebs1, RoaringBitmap.Or(rb1, rb2)), ShouldEqual, true)
			// testing XOR
			clonebs1 = bs1.Clone()
			clonebs1.Xor(bs2)
			So(equals(clonebs1, RoaringBitmap.Xor(rb1, rb2)), ShouldEqual, true)

			// testing NOTAND
			clonebs1 = bs1.Clone()
			clonebs1.AndNot(bs2)
			So(equals(clonebs1, RoaringBitmap.AndNot(rb1, rb2)), ShouldEqual, true)

		}
	}
}
func validate(bc BitmapContainer, ac ArrayContainer) bool {
	// Checking the cardinalities of each container

	if bc.GetCardinality() != ac.GetCardinality() {
		log.Println("cardinality differs")
		return false
	}
	// Checking that the two containers contain the same values
	counter := 0

	for i := bc.NextSetBit(0); i >= 0; i = bc.NextSetBit(i + 1) {
		counter++
		if !ac.contains(int16(i)) {
			log.Println("content differs")
			log.Println(bc)
			log.Println(ac)
			return false
		}

	}

	// checking the cardinality of the BitmapContainer
	return counter == bc.GetCardinality()
}

func TestRoaringArray(t *testing.T) {

	a := NewRoaringArray()
	Convey("Test Init", t, func() {
		So(a.Size(), ShouldEqual, 0)
	})

	Convey("Test Insert", t, func() {
		a.Append(0, NewArrayContainer())

		So(a.Size(), ShouldEqual, 1)
	})

	Convey("Test Remove", t, func() {
		a.Remove(0)
		So(a.Size(), ShouldEqual, 0)
	})

	Convey("Test Bitcount Full", t, func() {
		res := BitCount(-1)
		So(res, ShouldEqual, 64)
	})

	Convey("Test Bitcount Empty", t, func() {
		res := BitCount(0)
		So(res, ShouldEqual, 0)
	})
	Convey("Test ArrayContainer Add", t, func() {
		ar := NewArrayContainer()
		ar.Add(1)
		So(ar.cardinality, ShouldEqual, 1)
	})

	Convey("Test ArrayContainer Add wacky", t, func() {
		ar := NewArrayContainer()
		ar.Add(0)
		ar.Add(5000)
		So(ar.cardinality, ShouldEqual, 2)
	})

	Convey("Test ArrayContainer Add Reverse", t, func() {
		ar := NewArrayContainer()
		ar.Add(5000)
		ar.Add(2048)
		ar.Add(0)
		So(ar.cardinality, ShouldEqual, 3)
	})

	Convey("Test BitmapContainer Add ", t, func() {
		bm := NewBitmapContainer()
		bm.Add(0)
		So(bm.cardinality, ShouldEqual, 1)
	})

}
