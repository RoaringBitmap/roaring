package roaring

import (
	"io"
	"reflect"
	"unsafe"

	"github.com/tinylib/msgp/msgp"
)

func (b *arrayContainer) writeTo(stream io.Writer) (int, error) {
	buf := uint16SliceAsByteSlice(b.content)
	return stream.Write(buf)
}

func (b *bitmapContainer) writeTo(stream io.Writer) (int, error) {
	buf := uint64SliceAsByteSlice(b.bitmap)
	return stream.Write(buf)
}

func (b *runContainer32) writeTo(stream io.Writer) (int, error) {
	bts, err := b.MarshalMsg(nil)
	if err != nil {
		return 0, err
	}
	return stream.Write(bts)
}

func (b *runContainer16) writeTo(stream io.Writer) (int, error) {
	bts, err := b.MarshalMsg(nil)
	if err != nil {
		return 0, err
	}
	return stream.Write(bts)
}

func (b *arrayContainer) readFrom(stream io.Reader) (int, error) {
	buf := uint16SliceAsByteSlice(b.content)
	return io.ReadFull(stream, buf)
}

func (b *bitmapContainer) readFrom(stream io.Reader) (int, error) {
	buf := uint64SliceAsByteSlice(b.bitmap)
	return io.ReadFull(stream, buf)
}

func (b *runContainer32) readFrom(stream io.Reader) (int, error) {
	err := msgp.Decode(stream, b)
	return 0, err
}

func (b *runContainer16) readFrom(stream io.Reader) (int, error) {
	err := msgp.Decode(stream, b)
	return 0, err
}

func uint64SliceAsByteSlice(slice []uint64) []byte {
	// make a new slice header
	header := *(*reflect.SliceHeader)(unsafe.Pointer(&slice))

	// update its capacity and length
	header.Len *= 8
	header.Cap *= 8

	// return it
	return *(*[]byte)(unsafe.Pointer(&header))
}

func uint16SliceAsByteSlice(slice []uint16) []byte {
	// make a new slice header
	header := *(*reflect.SliceHeader)(unsafe.Pointer(&slice))

	// update its capacity and length
	header.Len *= 2
	header.Cap *= 2

	// return it
	return *(*[]byte)(unsafe.Pointer(&header))
}
