package roaring

// to run just these tests: go test -run TestParAggregations*

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestParAggregations(t *testing.T) {
	Convey("Par", t, func() {
		rb1 := NewBitmap()
		rb2 := NewBitmap()
		rb1.Add(1)
		rb2.Add(2)

		So(ParAnd(0, rb1, rb2).GetCardinality(), ShouldEqual, 0)
		So(ParOr(0, rb1, rb2).GetCardinality(), ShouldEqual, 2)
	})
}

func TestParAggregationsNothing(t *testing.T) {
	Convey("Par", t, func() {
		So(ParAnd(0).GetCardinality(), ShouldEqual, 0)
		So(ParOr(0).GetCardinality(), ShouldEqual, 0)
	})
}

func TestParAggregationsOneBitmap(t *testing.T) {
	Convey("Par", t, func() {
		rb := BitmapOf(1, 2, 3)

		So(ParAnd(0, rb).GetCardinality(), ShouldEqual, 3)
		So(ParOr(0, rb).GetCardinality(), ShouldEqual, 3)
	})
}

func TestParAggregationsOneEmpty(t *testing.T) {
	Convey("Par", t, func() {
		rb1 := NewBitmap()
		rb2 := NewBitmap()
		rb1.Add(1)

		So(ParAnd(0, rb1, rb2).GetCardinality(), ShouldEqual, 0)
		So(ParOr(0, rb1, rb2).GetCardinality(), ShouldEqual, 1)
	})
}

func TestParAggregationsDisjointSetIntersection(t *testing.T) {
	Convey("Par", t, func() {
		rb1 := BitmapOf(1)
		rb2 := BitmapOf(2)

		So(ParAnd(0, rb1, rb2).Stats().Containers, ShouldEqual, 0)
	})
}

func TestParAggregationsReversed3COW(t *testing.T) {
	Convey("Par", t, func() {
		rb1 := NewBitmap()
		rb2 := NewBitmap()
		rb3 := NewBitmap()
		rb1.SetCopyOnWrite(true)
		rb2.SetCopyOnWrite(true)
		rb3.SetCopyOnWrite(true)
		rb1.Add(1)
		rb1.Add(100000)
		rb2.Add(200000)
		rb3.Add(1)
		rb3.Add(300000)

		So(ParAnd(0, rb2, rb1, rb3).GetCardinality(), ShouldEqual, 0)
		So(ParOr(0, rb2, rb1, rb3).GetCardinality(), ShouldEqual, 4)
	})
}

func TestParAggregationsReversed3(t *testing.T) {
	Convey("Par", t, func() {
		rb1 := NewBitmap()
		rb2 := NewBitmap()
		rb3 := NewBitmap()
		rb1.Add(1)
		rb1.Add(100000)
		rb2.Add(200000)
		rb3.Add(1)
		rb3.Add(300000)

		So(ParAnd(0, rb2, rb1, rb3).GetCardinality(), ShouldEqual, 0)
		So(ParOr(0, rb2, rb1, rb3).GetCardinality(), ShouldEqual, 4)
	})
}

func TestParAggregations3(t *testing.T) {
	Convey("Par", t, func() {
		rb1 := NewBitmap()
		rb2 := NewBitmap()
		rb3 := NewBitmap()
		rb1.Add(1)
		rb1.Add(100000)
		rb2.Add(200000)
		rb3.Add(1)
		rb3.Add(300000)

		So(ParAnd(0, rb1, rb2, rb3).GetCardinality(), ShouldEqual, 0)
		So(ParOr(0, rb1, rb2, rb3).GetCardinality(), ShouldEqual, 4)
	})
}

func TestParAggregations3COW(t *testing.T) {
	Convey("Par", t, func() {
		rb1 := NewBitmap()
		rb2 := NewBitmap()
		rb3 := NewBitmap()
		rb1.SetCopyOnWrite(true)
		rb2.SetCopyOnWrite(true)
		rb3.SetCopyOnWrite(true)
		rb1.Add(1)
		rb1.Add(100000)
		rb2.Add(200000)
		rb3.Add(1)
		rb3.Add(300000)

		So(ParAnd(0, rb1, rb2, rb3).GetCardinality(), ShouldEqual, 0)
		So(ParOr(0, rb1, rb2, rb3).GetCardinality(), ShouldEqual, 4)
	})
}

func TestParAggregationsAdvanced(t *testing.T) {
	Convey("Par", t, func() {
		rb1 := NewBitmap()
		rb2 := NewBitmap()
		rb3 := NewBitmap()
		for i := uint32(0); i < 1000000; i += 3 {
			rb1.Add(i)
		}
		for i := uint32(0); i < 1000000; i += 7 {
			rb2.Add(i)
		}
		for i := uint32(0); i < 1000000; i += 1001 {
			rb3.Add(i)
		}
		for i := uint32(1000000); i < 2000000; i += 1001 {
			rb1.Add(i)
		}
		for i := uint32(1000000); i < 2000000; i += 3 {
			rb2.Add(i)
		}
		for i := uint32(1000000); i < 2000000; i += 7 {
			rb3.Add(i)
		}
		rb1.Or(rb2)
		rb1.Or(rb3)
		bigand := And(And(rb1, rb2), rb3)
		So(ParOr(0, rb1, rb2, rb3).Equals(rb1), ShouldEqual, true)
		So(ParAnd(0, rb1, rb2, rb3).Equals(bigand), ShouldEqual, true)
	})
}

func TestParAggregationsAdvanced_run(t *testing.T) {
	Convey("Par", t, func() {
		rb1 := NewBitmap()
		rb2 := NewBitmap()
		rb3 := NewBitmap()
		for i := uint32(500); i < 75000; i++ {
			rb1.Add(i)
		}
		for i := uint32(0); i < 1000000; i += 7 {
			rb2.Add(i)
		}
		for i := uint32(0); i < 1000000; i += 1001 {
			rb3.Add(i)
		}
		for i := uint32(1000000); i < 2000000; i += 1001 {
			rb1.Add(i)
		}
		for i := uint32(1000000); i < 2000000; i += 3 {
			rb2.Add(i)
		}
		for i := uint32(1000000); i < 2000000; i += 7 {
			rb3.Add(i)
		}
		rb1.RunOptimize()
		rb1.Or(rb2)
		rb1.Or(rb3)
		bigand := And(And(rb1, rb2), rb3)
		So(ParOr(0, rb1, rb2, rb3).Equals(rb1), ShouldEqual, true)
		So(ParAnd(0, rb1, rb2, rb3).Equals(bigand), ShouldEqual, true)
	})
}
