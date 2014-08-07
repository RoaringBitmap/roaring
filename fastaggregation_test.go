package roaring

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFastAggregations(t *testing.T) {
	Convey("Fast", t, func() {
		rb1 := NewRoaringBitmap()
		rb2 := NewRoaringBitmap()
		rb1.Add(1)
		rb2.Add(2)

		So(FastAnd(rb1, rb2).GetCardinality(), ShouldEqual, 0)
		So(FastOr(rb1, rb2).GetCardinality(), ShouldEqual, 2)
		So(FastXor(rb1, rb2).GetCardinality(), ShouldEqual, 2)
	})
}
