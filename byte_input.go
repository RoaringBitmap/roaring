package roaring

import (
	"encoding/binary"
	"io"
	"io/ioutil"
)

type byteInput struct {
	r         io.Reader
	readBytes int
}

type byteBuffer interface {
	// Next returns a slice containing the next n bytes from the buffer,
	// advancing the buffer as if the bytes had been returned by Read.
	Next(n int) []byte
}

// Read reads up to len(p) bytes into p. It returns the number of bytes
// read (0 <= n <= len(p)) and any error encountered.
func (b *byteInput) Read(p []byte) (n int, err error) {
	return b.r.Read(p)
}

// Next returns a slice containing the next n bytes from the reader
// If there are fewer bytes than the given n, io.ErrUnexpectedEOF will be returned
func (b *byteInput) next(n int) ([]byte, error) {
	if buf, ok := b.r.(byteBuffer); ok {
		data := buf.Next(n)
		b.readBytes += len(data)

		if len(data) != n {
			return nil, io.ErrUnexpectedEOF
		}

		return data, nil
	}

	buf := make([]byte, n)
	readBytes, err := b.r.Read(buf)
	b.readBytes += readBytes

	if err != nil {
		return nil, err
	}

	if readBytes != n {
		return nil, io.ErrUnexpectedEOF
	}

	return buf[:n], nil
}

// ReadUInt64 reads uint64 with LittleEndian order
func (b *byteInput) readUInt64() (uint64, error) {
	buf, err := b.next(8)

	if err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint64(buf), nil
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
	if buf, ok := b.r.(byteBuffer); ok {
		data := buf.Next(n)
		b.readBytes += len(data)

		if len(data) != n {
			err = io.ErrUnexpectedEOF
		}

		return
	}

	readBytes, err := io.CopyN(ioutil.Discard, b.r, int64(n))
	b.readBytes += int(readBytes)

	if int(readBytes) != n {
		err = io.ErrUnexpectedEOF
	}

	return
}
