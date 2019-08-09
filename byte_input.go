package roaring

import (
	"encoding/binary"
	"io"
)

type byteBuffer interface {
	// Next returns a slice containing the next n bytes from the buffer,
	// advancing the buffer as if the bytes had been returned by Read.
	next(n int) ([]byte, error)
}

type byteInput struct {
	buf       byteBuffer
	readBytes int
}

func newByteInputFromReader(reader io.Reader) *byteInput {
	return &byteInput{
		buf:       &byteBufferAdapter{reader},
		readBytes: 0,
	}
}

func newByteInputFromBuffer(buf []byte) *byteInput {
	return &byteInput{
		buf:       &byteBufferImpl{buf: buf, off: 0},
		readBytes: 0,
	}
}

// Next returns a slice containing the next n bytes from the reader
// If there are fewer bytes than the given n, io.ErrUnexpectedEOF will be returned
func (b *byteInput) next(n int) ([]byte, error) {
	data, err := b.buf.next(n)
	b.readBytes += len(data)

	if err != nil {
		return nil, err
	}

	if len(data) != n {
		return nil, io.ErrUnexpectedEOF
	}

	return data, nil
}

// ReadUInt64 reads uint32 with LittleEndian order
func (b *byteInput) readUInt32() (uint32, error) {
	buf, err := b.next(4)

	if err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint32(buf), nil
}

// readUInt16 reads uint16 with LittleEndian order
func (b *byteInput) readUInt16() (uint16, error) {
	buf, err := b.next(2)

	if err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint16(buf), nil
}

// getReadBytes returns read bytes
func (b *byteInput) getReadBytes() int64 {
	return int64(b.readBytes)
}

// skipBytes skips exactly n bytes, if
func (b *byteInput) skipBytes(n int) (err error) {
	data, err := b.buf.next(n)
	b.readBytes += len(data)

	if err != nil {
		return
	}

	if len(data) != n {
		err = io.ErrUnexpectedEOF
	}

	return
}

type byteBufferAdapter struct {
	r io.Reader
}

// Next returns a slice containing the next n bytes from the buffer,
// advancing the buffer as if the bytes had been returned by Read.
func (b *byteBufferAdapter) next(n int) ([]byte, error) {
	buf := make([]byte, n)
	readBytes, err := b.r.Read(buf)

	if err != nil {
		return nil, err
	}

	return buf[:readBytes], nil
}

type byteBufferImpl struct {
	buf []byte
	off int
}

// Next returns a slice containing the next n bytes from the buffer,
// advancing the buffer as if the bytes had been returned by Read.
func (b *byteBufferImpl) next(n int) ([]byte, error) {
	m := len(b.buf) - b.off

	if n > m {
		n = m
	}

	data := b.buf[b.off : b.off+n]
	b.off += n

	return data, nil
}
