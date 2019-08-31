package roaring

import (
	"bytes"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestByteInputFlow(t *testing.T) {
	Convey("Test should be an error on empty data", t, func() {
		buf := bytes.NewBuffer([]byte{})

		instances := []byteInput{
			newByteInput(buf.Bytes()),
			newByteInputFromReader(buf),
		}

		for _, input := range instances {
			n, err := input.readUInt16()
			So(n, ShouldEqual, 0)
			So(err, ShouldBeError)

			p, err := input.readUInt32()
			So(p, ShouldEqual, 0)
			So(err, ShouldBeError)

			b, err := input.next(10)
			So(b, ShouldEqual, nil)
			So(err, ShouldBeError)

			err = input.skipBytes(10)
			So(err, ShouldBeError)
		}
	})

	Convey("Test not empty data", t, func() {
		buf := bytes.NewBuffer(uint16SliceAsByteSlice([]uint16{1, 10, 32, 66, 23}))

		instances := []byteInput{
			newByteInput(buf.Bytes()),
			newByteInputFromReader(buf),
		}

		for _, input := range instances {
			n, err := input.readUInt16()
			So(n, ShouldEqual, 1)
			So(err, ShouldBeNil)

			p, err := input.readUInt32()
			So(p, ShouldEqual, 2097162) // 32 << 16 | 10
			So(err, ShouldBeNil)

			b, err := input.next(2)
			So([]byte{66, 0}, ShouldResemble, b)
			So(err, ShouldBeNil)

			err = input.skipBytes(2)
			So(err, ShouldBeNil)

			b, err = input.next(1)
			So(b, ShouldEqual, nil)
			So(err, ShouldBeError)
		}
	})
}
