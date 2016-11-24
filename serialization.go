package roaring

import (
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
	"unsafe"

	"github.com/tinylib/msgp/msgp"
)

func (b *arrayContainer) writeTo(stream io.Writer) (int, error) {
	//	buf := uint16SliceAsByteSlice(b.content)
	buf := make([]byte, len(b.content)*2)
	for i := range b.content {
		//p("arrayContainer is writing b.content[i=%v] = %v", i, b.content[i])
		binary.LittleEndian.PutUint16(buf[i*2:], b.content[i])
	}
	return stream.Write(buf)
}

func (b *bitmapContainer) writeTo(stream io.Writer) (int, error) {
	buf := uint64SliceAsByteSlice(b.bitmap)
	return stream.Write(buf)
}

// writeTo for runContainer16 follows this
// spec: https://github.com/RoaringBitmap/RoaringFormatSpec
//
func (b *runContainer16) writeTo(stream io.Writer) (int, error) {
	buf := make([]byte, 2+4*len(b.iv))
	binary.LittleEndian.PutUint16(buf[0:], uint16(len(b.iv)))
	for i, v := range b.iv {
		binary.LittleEndian.PutUint16(buf[2+i*4:], v.start)
		binary.LittleEndian.PutUint16(buf[2+2+i*4:], v.last-v.start)
	}
	return stream.Write(buf)
}

func (b *runContainer32) writeToMsgpack(stream io.Writer) (int, error) {
	bts, err := b.MarshalMsg(nil)
	if err != nil {
		return 0, err
	}
	return stream.Write(bts)
}

func (b *runContainer16) writeToMsgpack(stream io.Writer) (int, error) {
	bts, err := b.MarshalMsg(nil)
	if err != nil {
		return 0, err
	}
	return stream.Write(bts)
}

// readFrom reads from stream.
// PRE-REQUISITE: you must size the arrayContainer correctly (allocate b.content)
// *before* you call readFrom. We can't guess the size in the stream
// by this point.
func (b *arrayContainer) readFrom(stream io.Reader) (int, error) {
	buf := uint16SliceAsByteSlice(b.content)
	n, err := io.ReadFull(stream, buf)
	return n, err
}

func (b *bitmapContainer) readFrom(stream io.Reader) (int, error) {
	buf := uint64SliceAsByteSlice(b.bitmap)
	n, err := io.ReadFull(stream, buf)
	b.computeCardinality()
	return n, err
}

func (b *runContainer32) readFromMsgpack(stream io.Reader) (int, error) {
	err := msgp.Decode(stream, b)
	return 0, err
}

func (b *runContainer16) readFromMsgpack(stream io.Reader) (int, error) {
	err := msgp.Decode(stream, b)
	return 0, err
}

func (b *runContainer16) readFrom(stream io.Reader) (int, error) {
	b.iv = b.iv[:0]
	b.card = 0
	var numRuns uint16
	err := binary.Read(stream, binary.LittleEndian, &numRuns)
	if err != nil {
		return 0, err
	}
	encRun := make([]uint16, 2*numRuns)
	bEncRun := uint16SliceAsByteSlice(encRun)
	err = binary.Read(stream, binary.LittleEndian, &bEncRun)
	if err != nil {
		return 0, err
	}
	//p("encRun = %#v", encRun)
	nr := int(numRuns)
	//p("runContainer16.readFrom: I see %v runs; len(encRun)=%v", numRuns, len(encRun))
	for i := 0; i < nr; i++ {
		//p("i = %v. b.iv = %#v", i, b.iv)
		if i > 0 && b.iv[i-1].last >= encRun[i*2] {
			panic(fmt.Errorf("error: stored runContainer had runs that were not in sorted order!! (b.iv[i-1=%v].last = %v >= encRun[i=%v] = %v)", i-1, b.iv[i-1].last, i, encRun[i*2]))
		}
		b.iv = append(b.iv, interval16{start: encRun[i*2], last: encRun[i*2] + encRun[i*2+1]})

		//p("i=%v run: start:%v len:%v", i, encRun[i*2], encRun[i*2+1])
		//p("i=%v after append, b.iv = %#v", i, b.iv)

		b.card += int64(encRun[i*2+1]) + 1
	}
	//p("exiting runContainer16 readFrom")
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
