package roaring

import (
	"bytes"
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"runtime"
	"testing"
)

func BenchmarkUnserializeReadFrom(b *testing.B) {
	for _, size := range []uint32{650, 6500, 65000, 650000, 6500000} {
		rb := New()
		buf := &bytes.Buffer{}

		for i := uint32(0); i < size; i++ {
			rb.Add(i)
		}

		_, err := rb.WriteTo(buf)

		if err != nil {
			b.Fatalf("Unexpected error occurs: %v", err)
		}

		b.Run(fmt.Sprintf("ReadFrom-%d", size), func(b *testing.B) {
			b.ReportAllocs()
			b.StartTimer()

			for n := 0; n < b.N; n++ {
				reader := bytes.NewReader(buf.Bytes())
				nb := New()

				if _, err := nb.ReadFrom(reader); err != nil {
					b.Fatalf("Unexpected error occurs: %v", err)
				}
			}

			b.StopTimer()
		})
	}
}

func BenchmarkUnserializeFromBuffer(b *testing.B) {
	for _, size := range []uint32{650, 6500, 65000, 650000, 6500000} {
		rb := New()
		buf := &bytes.Buffer{}

		for i := uint32(0); i < size; i++ {
			rb.Add(i)
		}

		_, err := rb.WriteTo(buf)

		if err != nil {
			b.Fatalf("Unexpected error occurs: %v", err)
		}

		b.Run(fmt.Sprintf("FromBuffer-%d", size), func(b *testing.B) {
			b.ReportAllocs()
			b.StartTimer()

			for n := 0; n < b.N; n++ {
				nb := New()

				if _, err := nb.FromBuffer(buf.Bytes()); err != nil {
					b.Fatalf("Unexpected error occurs: %v", err)
				}
			}

			b.StopTimer()
		})
	}
}

func TestHoldReference(t *testing.T) {
	Convey("Test Hold Reference", t, func() {
		rb := New()
		buf := &bytes.Buffer{}

		for i := uint32(0); i < 650; i++ {
			rb.Add(i)
		}

		_, err := rb.WriteTo(buf)
		So(err, ShouldBeNil)

		nb := New()
		data := buf.Bytes()

		if _, err := nb.ReadFrom(bytes.NewReader(data)); err != nil {
			t.Fatalf("Unexpected error occurs: %v", err)
		}

		buf = nil
		rb = nil
		data = nil

		runtime.GC()

		iterator := nb.Iterator()
		i := uint32(0)

		for iterator.HasNext() {
			v := iterator.Next()

			if v != i {
				return
			}

			So(v, ShouldEqual, i)
			i++
		}
	})
}
