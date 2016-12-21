// +build 386 amd64,!appengine

package roaring

import (
	"io"
)

func (b *arrayContainer) writeTo(stream io.Writer) (int, error) {
	buf := uint16SliceAsByteSlice(b.content)
	return stream.Write(buf)
}

func (b *bitmapContainer) writeTo(stream io.Writer) (int, error) {
	buf := uint64SliceAsByteSlice(b.bitmap)
	return stream.Write(buf)
}

// readFrom reads an arrayContainer from stream.
// PRE-REQUISITE: you must size the arrayContainer correctly (allocate b.content)
// *before* you call readFrom. We can't guess the size in the stream
// by this point.
func (b *arrayContainer) readFrom(stream io.Reader) (int, error) {
	buf := uint16SliceAsByteSlice(b.content)
	return io.ReadFull(stream, buf)
}

func (b *bitmapContainer) readFrom(stream io.Reader) (int, error) {
	buf := uint64SliceAsByteSlice(b.bitmap)
	n, err := io.ReadFull(stream, buf)
	b.computeCardinality()
	return n, err
}

