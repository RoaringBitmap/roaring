package roaring

import (
	"log"
	"math/rand"
	"strconv"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/willf/bitset"
)

func TestRoaringBitmapRank(t *testing.T) {
	for N := 1; N <= 1048576; N *= 2 {
		Convey("rank tests"+strconv.Itoa(N), t, func() {
			for gap := 1; gap <= 65536; gap *= 2 {
				rb1 := NewRoaringBitmap()
				for x := 0; x <= N; x += gap {
					rb1.Add(x)
				}
				for y := 0; y <= N; y += 1 {
					if rb1.Rank(y) != (y+1+gap-1)/gap {
						So(rb1.Rank(y), ShouldEqual, (y+1+gap-1)/gap)
					}
				}
			}
		})
	}
}

func TestRoaringBitmapSelect(t *testing.T) {
	for N := 1; N <= 1048576; N *= 2 {
		Convey("rank tests"+strconv.Itoa(N), t, func() {
			for gap := 1; gap <= 65536; gap *= 2 {
				rb1 := NewRoaringBitmap()
				for x := 0; x <= N; x += gap {
					rb1.Add(x)
				}
				for y := 0; y <= N/gap; y += 1 {
					expectedInt := y * gap
					i, err := rb1.Select(y)
					if err != nil {
						t.Fatal(err)
					}

					if i != expectedInt {
						So(i, ShouldEqual, expectedInt)
					}
				}
			}
		})
	}
}

// some extra tests
func TestRoaringBitmapExtra(t *testing.T) {
	for N := 1; N <= 65536; N *= 2 {
		Convey("extra tests"+strconv.Itoa(N), t, func() {
			for gap := 1; gap <= 65536; gap *= 2 {
				bs1 := bitset.New(0)
				rb1 := NewRoaringBitmap()
				for x := 0; x <= N; x += gap {
					bs1.Set(uint(x))
					rb1.Add(x)
				}
				So(bs1.Count(), ShouldEqual, rb1.GetCardinality())
				So(equalsBitSet(bs1, rb1), ShouldEqual, true)
				for offset := 1; offset <= gap; offset *= 2 {
					bs2 := bitset.New(0)
					rb2 := NewRoaringBitmap()
					for x := 0; x <= N; x += gap {
						bs2.Set(uint(x + offset))
						rb2.Add(x + offset)
					}
					So(bs2.Count(), ShouldEqual, rb2.GetCardinality())
					So(equalsBitSet(bs2, rb2), ShouldEqual, true)

					clonebs1 := bs1.Clone()
					clonebs1.InPlaceIntersection(bs2)
					if !equalsBitSet(clonebs1, And(rb1, rb2)) {
						t := rb1.Clone()
						t.And(rb2)
						So(equalsBitSet(clonebs1, t), ShouldEqual, true)
					}

					// testing OR
					clonebs1 = bs1.Clone()
					clonebs1.InPlaceUnion(bs2)

					So(equalsBitSet(clonebs1, Or(rb1, rb2)), ShouldEqual, true)
					// testing XOR
					clonebs1 = bs1.Clone()
					clonebs1.InPlaceSymmetricDifference(bs2)
					So(equalsBitSet(clonebs1, Xor(rb1, rb2)), ShouldEqual, true)

					//testing NOTAND
					clonebs1 = bs1.Clone()
					clonebs1.InPlaceDifference(bs2)
					So(equalsBitSet(clonebs1, AndNot(rb1, rb2)), ShouldEqual, true)
				}
			}
		})
	}
}

func FlipRange(start, end int, bs *bitset.BitSet) {
	for i := start; i < end; i++ {
		bs.Flip(uint(i))
	}
}

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

	Convey("Test Clone", t, func() {
		rb1 := NewRoaringBitmap()
		rb1.Add(10)

		rb2 := rb1.Clone()
		rb2.Remove(10)

		So(rb1.Contains(10), ShouldBeTrue)
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

		off := AndNot(rb2, rb)
		andNotresult := AndNot(rb, rb2)

		So(rb.Equals(andNotresult), ShouldEqual, true)
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
		rrand := And(rr, rr2)
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
		correct := And(rr, rr2)
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

		rrand := And(rr, rr2)
		array := rrand.ToArray()
		So(len(array), ShouldEqual, 1)
		So(array[0], ShouldEqual, 13)
	})
	Convey("Test AND 3a", t, func() {
		rr := NewRoaringBitmap()
		rr2 := NewRoaringBitmap()
		for k := 6 * 65536; k < 6*65536+10000; k++ {
			rr.Add(k)
		}
		for k := 6 * 65536; k < 6*65536+1000; k++ {
			rr2.Add(k)
		}
		result := And(rr, rr2)
		So(result.GetCardinality(), ShouldEqual, 1000)
	})
	Convey("Test AND 3", t, func() {
		var arrayand [11256]int
		//393,216
		pos := 0
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
		for k := 8 * 65536; k < 8*65536+1000; k++ {
			rr.Add(k)
		}
		for k := 9 * 65536; k < 9*65536+30000; k++ {
			rr.Add(k)
		}

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
		for k := 6 * 65536; k < 6*65536+10000; k++ {
			rr.Add(k)
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
		rrand := And(rr, rr2)

		arrayres := rrand.ToArray()
		ok := true
		for i := range arrayres {
			if i < len(arrayand) {
				if arrayres[i] != arrayand[i] {
					log.Println(i, arrayres[i], arrayand[i])
					ok = false
				}
			} else {
				log.Println('x', arrayres[i])
				ok = false
			}
		}

		So(len(arrayand), ShouldEqual, len(arrayres))
		So(ok, ShouldEqual, true)

	})

	Convey("Test AND 4", t, func() {
		rb := NewRoaringBitmap()
		rb2 := NewRoaringBitmap()

		for i := 0; i < 200000; i += 4 {
			rb2.Add(i)
		}
		for i := 200000; i < 400000; i += 14 {
			rb2.Add(i)
		}
		//TODO: RoaringBitmap.And(bm,bm2)
		andresult := And(rb, rb2)
		off := And(rb2, rb)
		So(andresult.Equals(off), ShouldEqual, true)
		So(andresult.GetCardinality(), ShouldEqual, 0)

		for i := 500000; i < 600000; i += 14 {
			rb.Add(i)
		}
		for i := 200000; i < 400000; i += 3 {
			rb2.Add(i)
		}
		andresult2 := And(rb, rb2)
		So(andresult.GetCardinality(), ShouldEqual, 0)
		So(andresult2.GetCardinality(), ShouldEqual, 0)

		for i := 0; i < 200000; i += 4 {
			rb.Add(i)
		}
		for i := 200000; i < 400000; i += 14 {
			rb.Add(i)
		}
		So(andresult.GetCardinality(), ShouldEqual, 0)
		rc := And(rb, rb2)
		rb.And(rb2)
		So(rc.GetCardinality(), ShouldEqual, rb.GetCardinality())

	})

	Convey("ArrayContainerCardinalityTest", t, func() {
		ac := newArrayContainer()
		for k := uint16(0); k < 100; k++ {
			ac.add(k)
			So(ac.getCardinality(), ShouldEqual, k+1)
		}
		for k := uint16(0); k < 100; k++ {
			ac.add(k)
			So(ac.getCardinality(), ShouldEqual, 100)
		}
	})
	/*
		Convey("ArrayTest", t, func() {
			rr := newArrayContainer()
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

	Convey("or test", t, func() {
		rr := NewRoaringBitmap()
		for k := 0; k < 4000; k++ {
			rr.Add(k)
		}
		rr2 := NewRoaringBitmap()
		for k := 4000; k < 8000; k++ {
			rr2.Add(k)
		}
		result := Or(rr, rr2)
		So(result.GetCardinality(), ShouldEqual, rr.GetCardinality()+rr2.GetCardinality())
	})
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
		ok := true
		for i := range a {
			if array[i] != a[i] {
				log.Println("rr : ", array[i], " a : ", a[i])
				ok = false
			}
		}
		So(len(array), ShouldEqual, len(a))
		So(ok, ShouldEqual, true)
	})

	Convey("BitmapContainerCardinalityTest", t, func() {
		ac := newBitmapContainer()
		for k := uint16(0); k < 100; k++ {
			ac.add(k)
			So(ac.getCardinality(), ShouldEqual, k+1)
		}
		for k := uint16(0); k < 100; k++ {
			ac.add(k)
			So(ac.getCardinality(), ShouldEqual, 100)
		}
	})

	Convey("BitmapContainerTest", t, func() {
		rr := newBitmapContainer()
		rr.add(uint16(110))
		rr.add(uint16(114))
		rr.add(uint16(115))
		var array [3]uint16
		pos := 0
		for itr := rr.getShortIterator(); itr.hasNext(); {
			array[pos] = itr.next()
			pos++
		}

		So(array[0], ShouldEqual, uint16(110))
		So(array[1], ShouldEqual, uint16(114))
		So(array[2], ShouldEqual, uint16(115))
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
				So(And(rb, rb2).GetCardinality(), ShouldEqual, N/offset)
				So(Xor(rb, rb2).GetCardinality(), ShouldEqual, 2*N-2*N/offset)
				So(Or(rb, rb2).GetCardinality(), ShouldEqual, 2*N-N/offset)
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
		andresult := And(rb, rb2)
		orresult := Or(rb, rb2)

		So(andresult.GetCardinality(), ShouldEqual, 1)
		So(orresult.GetCardinality(), ShouldEqual, rb2.GetCardinality())

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
		ok := true
		for i := range arrayrr {
			if arrayrr[i] != arrayrr3[i] {
				ok = false
			}
		}
		So(len(arrayrr), ShouldEqual, len(arrayrr3))
		So(ok, ShouldEqual, true)
	})

	Convey("constainer factory ", t, func() {

		bc1 := newBitmapContainer()
		bc2 := newBitmapContainer()
		bc3 := newBitmapContainer()
		ac1 := newArrayContainer()
		ac2 := newArrayContainer()
		ac3 := newArrayContainer()

		for i := 0; i < 5000; i++ {
			bc1.add(uint16(i * 70))
		}
		for i := 0; i < 5000; i++ {
			bc2.add(uint16(i * 70))
		}
		for i := 0; i < 5000; i++ {
			bc3.add(uint16(i * 70))
		}
		for i := 0; i < 4000; i++ {
			ac1.add(uint16(i * 50))
		}
		for i := 0; i < 4000; i++ {
			ac2.add(uint16(i * 50))
		}
		for i := 0; i < 4000; i++ {
			ac3.add(uint16(i * 50))
		}

		rbc := ac1.clone().(*arrayContainer).toBitmapContainer()
		So(validate(rbc, ac1), ShouldEqual, true)
		rbc = ac2.clone().(*arrayContainer).toBitmapContainer()
		So(validate(rbc, ac2), ShouldEqual, true)
		rbc = ac3.clone().(*arrayContainer).toBitmapContainer()
		So(validate(rbc, ac3), ShouldEqual, true)
	})
	Convey("flipTest1 ", t, func() {
		rb := NewRoaringBitmap()
		rb.Flip(100000, 200000) // in-place on empty bitmap
		rbcard := rb.GetCardinality()
		So(100000, ShouldEqual, rbcard)

		bs := bitset.New(20000 - 10000)
		for i := uint(100000); i < 200000; i++ {
			bs.Set(i)
		}
		So(equalsBitSet(bs, rb), ShouldEqual, true)
	})

	Convey("flipTest1A", t, func() {
		rb := NewRoaringBitmap()
		rb1 := Flip(rb, 100000, 200000)
		rbcard := rb1.GetCardinality()
		So(100000, ShouldEqual, rbcard)
		So(0, ShouldEqual, rb.GetCardinality())

		bs := bitset.New(0)
		So(equalsBitSet(bs, rb), ShouldEqual, true)

		for i := uint(100000); i < 200000; i++ {
			bs.Set(i)
		}
		So(equalsBitSet(bs, rb1), ShouldEqual, true)
	})
	Convey("flipTest2", t, func() {
		rb := NewRoaringBitmap()
		rb.Flip(100000, 100000)
		rbcard := rb.GetCardinality()
		So(0, ShouldEqual, rbcard)

		bs := bitset.New(0)
		So(equalsBitSet(bs, rb), ShouldEqual, true)
	})

	Convey("flipTest2A", t, func() {
		rb := NewRoaringBitmap()
		rb1 := Flip(rb, 100000, 100000)

		rb.Add(1)
		rbcard := rb1.GetCardinality()

		So(0, ShouldEqual, rbcard)
		So(1, ShouldEqual, rb.GetCardinality())

		bs := bitset.New(0)
		So(equalsBitSet(bs, rb1), ShouldEqual, true)
		bs.Set(1)
		So(equalsBitSet(bs, rb), ShouldEqual, true)
	})

	Convey("flipTest3A", t, func() {
		rb := NewRoaringBitmap()
		rb.Flip(100000, 200000) // got 100k-199999
		rb.Flip(100000, 199991) // give back 100k-199990
		rbcard := rb.GetCardinality()
		So(9, ShouldEqual, rbcard)

		bs := bitset.New(0)
		for i := uint(199991); i < 200000; i++ {
			bs.Set(i)
		}

		So(equalsBitSet(bs, rb), ShouldEqual, true)
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

		bs := bitset.New(0)
		for i := uint(65536); i < 100000; i++ {
			bs.Set(i)
		}
		for i := uint(200000); i < 262144; i++ {
			bs.Set(i)
		}

		So(equalsBitSet(bs, rb), ShouldEqual, true)
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

		So(46464, ShouldEqual, rbcard)

		bs := bitset.New(0)
		for i := uint(65536); i < 100000; i++ {
			bs.Set(i)
		}
		for i := uint(120000); i < 132000; i++ {
			bs.Set(i)
		}
		So(equalsBitSet(bs, rb), ShouldEqual, true)
	})

	Convey("flipTest6", t, func() {
		rb := NewRoaringBitmap()
		rb1 := Flip(rb, 100000, 132000)
		rb2 := Flip(rb1, 65536, 120000)
		//rbcard := rb2.GetCardinality()

		bs := bitset.New(0)
		for i := uint(65536); i < 100000; i++ {
			bs.Set(i)
		}
		for i := uint(120000); i < 132000; i++ {
			bs.Set(i)
		}
		So(equalsBitSet(bs, rb2), ShouldEqual, true)
	})

	Convey("flipTest6A", t, func() {
		rb := NewRoaringBitmap()
		rb1 := Flip(rb, 100000, 132000)
		rb2 := Flip(rb1, 99000, 2*65536)
		rbcard := rb2.GetCardinality()

		So(1928, ShouldEqual, rbcard)

		bs := bitset.New(0)
		for i := uint(99000); i < 100000; i++ {
			bs.Set(i)
		}
		for i := uint(2 * 65536); i < 132000; i++ {
			bs.Set(i)
		}
		So(equalsBitSet(bs, rb2), ShouldEqual, true)
	})

	Convey("flipTest7", t, func() {
		// within 1 word, first container
		rb := NewRoaringBitmap()
		rb.Flip(650, 132000)
		rb.Flip(648, 651)
		rbcard := rb.GetCardinality()

		// 648, 649, 651-131999

		So(132000-651+2, ShouldEqual, rbcard)
		bs := bitset.New(0)
		bs.Set(648)
		bs.Set(649)
		for i := uint(651); i < 132000; i++ {
			bs.Set(i)
		}
		So(equalsBitSet(bs, rb), ShouldEqual, true)
	})
	Convey("flipTestBig", t, func() {
		numCases := 1000
		rb := NewRoaringBitmap()
		bs := bitset.New(0)
		//Random r = new Random(3333);
		checkTime := 2.0

		for i := 0; i < numCases; i++ {
			start := rand.Intn(65536 * 20)
			end := rand.Intn(65536 * 20)
			if rand.Float64() < float64(0.1) {
				end = start + rand.Intn(100)
			}
			rb.Flip(start, end)
			if start < end {
				FlipRange(start, end, bs) // throws exception
			}
			// otherwise
			// insert some more ANDs to keep things sparser
			if rand.Float64() < 0.2 {
				mask := NewRoaringBitmap()
				mask1 := bitset.New(0)
				startM := rand.Intn(65536 * 20)
				endM := startM + 100000
				mask.Flip(startM, endM)
				FlipRange(startM, endM, mask1)
				mask.Flip(0, 65536*20+100000)
				FlipRange(0, 65536*20+100000, mask1)
				rb.And(mask)
				bs.InPlaceIntersection(mask1)
			}
			// see if we can detect incorrectly shared containers
			if rand.Float64() < 0.1 {
				irrelevant := Flip(rb, 10, 100000)
				irrelevant.Flip(5, 200000)
				irrelevant.Flip(190000, 260000)
			}
			if float64(i) > checkTime {
				So(equalsBitSet(bs, rb), ShouldEqual, true)
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

		rror := Or(rr, rr2)

		array := rror.ToArray()

		rr.Or(rr2)
		arrayirr := rr.ToArray()
		So(IntsEquals(array, arrayirr), ShouldEqual, true)
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
		correct := Or(rr, rr2)
		rr.Or(rr2)
		So(correct.Equals(rr), ShouldEqual, true)
	})

	Convey("ortest2", t, func() {
		arrayrr := make([]int, 4000+4000+2)
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

		rror := Or(rr, rr2)

		arrayor := rror.ToArray()

		So(IntsEquals(arrayor, arrayrr), ShouldEqual, true)
	})

	Convey("ortest3", t, func() {
		V1 := make(map[int]bool)
		V2 := make(map[int]bool)

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

		rror := Or(rr, rr2)
		valide := true

		for _, k := range rror.ToArray() {
			_, found := V1[k]
			if !found {
				valide = false
			}
			V2[k] = true
		}

		for k := range V1 {
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
		orresult := Or(rb, rb2)
		off := Or(rb2, rb)
		So(orresult.Equals(off), ShouldEqual, true)

		So(rb2card, ShouldEqual, orresult.GetCardinality())

		for i := 500000; i < 600000; i += 14 {
			rb.Add(i)
		}
		for i := 200000; i < 400000; i += 3 {
			rb2.Add(i)
		}
		// check or against an empty bitmap
		orresult2 := Or(rb, rb2)
		So(rb2card, ShouldEqual, orresult.GetCardinality())
		So(rb2.GetCardinality()+rb.GetCardinality(), ShouldEqual,
			orresult2.GetCardinality())
		rb.Or(rb2)
		So(rb.Equals(orresult2), ShouldEqual, true)

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
		correct := Xor(rr, rr2)
		rr.Xor(rr2)
		So(correct.Equals(rr), ShouldEqual, true)
	})

	Convey("xortest1", t, func() {
		V1 := make(map[int]bool)
		V2 := make(map[int]bool)

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

		rrxor := Xor(rr, rr2)
		valide := true

		for _, i := range rrxor.ToArray() {
			_, found := V1[i]
			if !found {
				valide = false
			}
			V2[i] = true
		}
		for k := range V1 {
			_, found := V2[k]
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
		xorresult := Xor(rb, rb2)
		off := Or(rb2, rb)
		So(xorresult.Equals(off), ShouldEqual, true)

		So(rb2card, ShouldEqual, xorresult.GetCardinality())

		for i := 500000; i < 600000; i += 14 {
			rb.Add(i)
		}
		for i := 200000; i < 400000; i += 3 {
			rb2.Add(i)
		}
		// check or against an empty bitmap
		xorresult2 := Xor(rb, rb2)
		So(rb2card, ShouldEqual, xorresult.GetCardinality())
		So(rb2.GetCardinality()+rb.GetCardinality(), ShouldEqual, xorresult2.GetCardinality())

		rb.Xor(rb2)
		So(xorresult2.Equals(rb), ShouldEqual, true)

	})
	//need to add the massives
}

func TestBigRandom(t *testing.T) {
	Convey("randomTest", t, func() {
		rTest(15)
		rTest(100)
		rTest(512)
		rTest(1023)
		rTest(1025)
		rTest(4095)
		rTest(4096)
		rTest(4097)
		rTest(65536)
		rTest(65536 * 16)
	})
}

func rTest(N int) {
	log.Println("rtest N=", N)
	for gap := 1; gap <= 65536; gap *= 2 {
		bs1 := bitset.New(0)
		rb1 := NewRoaringBitmap()
		for x := 0; x <= N; x += gap {
			bs1.Set(uint(x))
			rb1.Add(x)
		}
		So(bs1.Count(), ShouldEqual, rb1.GetCardinality())
		So(equalsBitSet(bs1, rb1), ShouldEqual, true)
		for offset := 1; offset <= gap; offset *= 2 {
			bs2 := bitset.New(0)
			rb2 := NewRoaringBitmap()
			for x := 0; x <= N; x += gap {
				bs2.Set(uint(x + offset))
				rb2.Add(x + offset)
			}
			So(bs2.Count(), ShouldEqual, rb2.GetCardinality())
			So(equalsBitSet(bs2, rb2), ShouldEqual, true)

			clonebs1 := bs1.Clone()
			clonebs1.InPlaceIntersection(bs2)
			if !equalsBitSet(clonebs1, And(rb1, rb2)) {
				t := rb1.Clone()
				t.And(rb2)
				So(equalsBitSet(clonebs1, t), ShouldEqual, true)
			}

			// testing OR
			clonebs1 = bs1.Clone()
			clonebs1.InPlaceUnion(bs2)

			So(equalsBitSet(clonebs1, Or(rb1, rb2)), ShouldEqual, true)
			// testing XOR
			clonebs1 = bs1.Clone()
			clonebs1.InPlaceSymmetricDifference(bs2)
			So(equalsBitSet(clonebs1, Xor(rb1, rb2)), ShouldEqual, true)

			//testing NOTAND
			clonebs1 = bs1.Clone()
			clonebs1.InPlaceDifference(bs2)
			So(equalsBitSet(clonebs1, AndNot(rb1, rb2)), ShouldEqual, true)
		}
	}
}

func equalsBitSet(a *bitset.BitSet, b *RoaringBitmap) bool {
	for i, e := a.NextSet(0); e; i, e = a.NextSet(i + 1) {
		if !b.Contains(int(i)) {
			return false
		}
	}
	i := b.Iterator()
	for i.HasNext() {
		if !a.Test(uint(i.Next())) {
			return false
		}
	}
	return true
}

func equalsArray(a []int, b *RoaringBitmap) bool {
	if len(a) != b.GetCardinality() {
		return false
	}
	for _, x := range a {
		if !b.Contains(x) {
			return false
		}
	}
	return true
}

func IntsEquals(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func validate(bc *bitmapContainer, ac *arrayContainer) bool {
	// Checking the cardinalities of each container

	if bc.getCardinality() != ac.getCardinality() {
		log.Println("cardinality differs")
		return false
	}
	// Checking that the two containers contain the same values
	counter := 0

	for i := bc.NextSetBit(0); i >= 0; i = bc.NextSetBit(i + 1) {
		counter++
		if !ac.contains(uint16(i)) {
			log.Println("content differs")
			log.Println(bc)
			log.Println(ac)
			return false
		}

	}

	// checking the cardinality of the BitmapContainer
	return counter == bc.getCardinality()
}

func TestRoaringArray(t *testing.T) {

	a := newRoaringArray()
	Convey("Test Init", t, func() {
		So(a.size(), ShouldEqual, 0)
	})

	Convey("Test Insert", t, func() {
		a.append(0, newArrayContainer())

		So(a.size(), ShouldEqual, 1)
	})

	Convey("Test Remove", t, func() {
		a.remove(0)
		So(a.size(), ShouldEqual, 0)
	})

	Convey("Test popcount Full", t, func() {
		res := popcount(uint64(0xffffffffffffffff))
		So(res, ShouldEqual, 64)
		res32 := popcount32(uint32(0xffffffff))
		So(res32, ShouldEqual, 32)
	})

	Convey("Test popcount Empty", t, func() {
		res := popcount(0)
		So(res, ShouldEqual, 0)
		res32 := popcount32(0)
		So(res32, ShouldEqual, 0)
	})

	Convey("Test popcount 16", t, func() {
		res := popcount(0xff00ff)
		So(res, ShouldEqual, 16)
		res32 := popcount32(0xff00ff)
		So(res32, ShouldEqual, 16)
	})

	Convey("Test ArrayContainer Add", t, func() {
		ar := newArrayContainer()
		ar.add(1)
		So(ar.getCardinality(), ShouldEqual, 1)
	})

	Convey("Test ArrayContainer Add wacky", t, func() {
		ar := newArrayContainer()
		ar.add(0)
		ar.add(5000)
		So(ar.getCardinality(), ShouldEqual, 2)
	})

	Convey("Test ArrayContainer Add Reverse", t, func() {
		ar := newArrayContainer()
		ar.add(5000)
		ar.add(2048)
		ar.add(0)
		So(ar.getCardinality(), ShouldEqual, 3)
	})

	Convey("Test BitmapContainer Add ", t, func() {
		bm := newBitmapContainer()
		bm.add(0)
		So(bm.getCardinality(), ShouldEqual, 1)
	})

}

func TestFlipBigA(t *testing.T) {
	Convey("flipTestBigA ", t, func() {
	numCases := 1000
	bs := bitset.New(0)
	checkTime := 2.0
	rb1 := NewRoaringBitmap()
	rb2 := NewRoaringBitmap()

	for i := 0; i < numCases; i++ {
		start := rand.Intn(65536 * 20)
		end := rand.Intn(65536 * 20)
		if rand.Float64() < 0.1 {
			end = start + rand.Intn(100)
		}

		if (i & 1) == 0 {
			rb2 = Flip(rb1, start, end)
			// tweak the other, catch bad sharing
			rb1.Flip(rand.Intn(65536*20), rand.Intn(65536*20))
		} else {
			rb1 = Flip(rb2, start, end)
			rb2.Flip(rand.Intn(65536*20), rand.Intn(65536*20))
		}

		if start < end {
			FlipRange(start, end, bs) // throws exception
		}
		// otherwise
		// insert some more ANDs to keep things sparser
		if (rand.Float64() < 0.2) && (i&1) == 0 {
			mask := NewRoaringBitmap()
			mask1 := bitset.New(0)
			startM := rand.Intn(65536 * 20)
			endM := startM + 100000
			mask.Flip(startM, endM)
			FlipRange(startM, endM, mask1)
			mask.Flip(0, 65536*20+100000)
			FlipRange(0, 65536*20+100000, mask1)
			rb2.And(mask)
			bs.InPlaceIntersection(mask1)
		}

		if float64(i) > checkTime {
			var rb *RoaringBitmap

			if (i & 1) == 0 {
				rb = rb2
			} else {
				rb = rb1
			}
			So(equalsBitSet(bs, rb), ShouldEqual, true)
			checkTime *= 1.5
		}
	}
})}
