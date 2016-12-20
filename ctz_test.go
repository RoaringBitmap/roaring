package roaring

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCountTrailingZeros072(t *testing.T) {
	Convey("countTrailingZeros", t, func() {
		So(countTrailingZeros(0), ShouldEqual, 64)
		So(countTrailingZeros(8), ShouldEqual, 3)
		So(countTrailingZeros(7), ShouldEqual, 0)
		So(countTrailingZeros(1<<17), ShouldEqual, 17)
		So(countTrailingZeros(7<<17), ShouldEqual, 17)
		So(countTrailingZeros(255<<33), ShouldEqual, 33)

	})
}
