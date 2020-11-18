// +build 386,!appengine amd64,!appengine arm,!appengine arm64,!appengine ppc64le,!appengine mipsle,!appengine mips64le,!appengine mips64p32le,!appengine wasm,!appengine

package roaring

import (
	"errors"
	"io"
	"reflect"
	"runtime"
	"unsafe"

	"github.com/RoaringBitmap/roaring/internal"
)

func (ac *arrayContainer) writeTo(stream io.Writer) (int, error) {
	buf := internal.Uint16SliceAsByteSlice(ac.content)
	return stream.Write(buf)
}

func (bc *bitmapContainer) writeTo(stream io.Writer) (int, error) {
	if bc.cardinality <= arrayDefaultMaxSize {
		return 0, errors.New("refusing to write bitmap container with cardinality of array container")
	}
	buf := internal.Uint64SliceAsByteSlice(bc.bitmap)
	return stream.Write(buf)
}

func (bc *bitmapContainer) asLittleEndianByteSlice() []byte {
	return internal.Uint64SliceAsByteSlice(bc.bitmap)
}

// Deserialization code follows

func byteSliceAsInterval16Slice(slice []byte) (result []interval16) {
	if len(slice)%4 != 0 {
		panic("Slice size should be divisible by 4")
	}
	// reference: https://go101.org/article/unsafe.html

	// make a new slice header
	bHeader := (*reflect.SliceHeader)(unsafe.Pointer(&slice))
	rHeader := (*reflect.SliceHeader)(unsafe.Pointer(&result))

	// transfer the data from the given slice to a new variable (our result)
	rHeader.Data = bHeader.Data
	rHeader.Len = bHeader.Len / 4
	rHeader.Cap = bHeader.Cap / 4

	// instantiate result and use KeepAlive so data isn't unmapped.
	runtime.KeepAlive(&slice) // it is still crucial, GC can free it)

	// return result
	return
}
