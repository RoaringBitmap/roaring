package roaring64

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/RoaringBitmap/roaring/v2"
)

// maxPortableBucketCount bounds the bucket count a portable stream may declare:
// bucket keys are uint32, so there are at most 2^32 distinct ones.
const maxPortableBucketCount = uint64(1) << 32

// WritePortableTo writes this bitmap in the portable 64-bit format, specified at
// https://github.com/RoaringBitmap/RoaringFormatSpec#extention-for-64-bit-implementations
// This is the format WriteTo already produces. Call RunOptimize first for better
// compression.
func (rb *Bitmap) WritePortableTo(stream io.Writer) (int64, error) {
	return rb.WriteTo(stream)
}

// ToPortableBytes returns the bytes WritePortableTo writes.
func (rb *Bitmap) ToPortableBytes() ([]byte, error) {
	var buf bytes.Buffer
	_, err := rb.WritePortableTo(&buf)
	return buf.Bytes(), err
}

// GetPortableSerializedSizeInBytes returns the number of bytes WritePortableTo
// writes, without writing them.
func (rb *Bitmap) GetPortableSerializedSizeInBytes() uint64 {
	return rb.GetSerializedSizeInBytes()
}

// ReadPortableFrom reads a bitmap in the portable 64-bit format. Unlike ReadFrom
// it checks the bucket count, the bucket key order and the absence of an empty
// bucket, and leaves rb untouched when the stream is truncated or malformed. The
// contents of each 32-bit bucket are not checked; call Validate if the source is
// untrusted.
func (rb *Bitmap) ReadPortableFrom(stream io.Reader) (p int64, err error) {
	buf := make([]byte, 8)
	n, err := io.ReadFull(stream, buf)
	p += int64(n)
	if err != nil {
		return p, err
	}
	size := binary.LittleEndian.Uint64(buf)
	if size > maxPortableBucketCount {
		return p, fmt.Errorf("error in bitmap.ReadPortableFrom: invalid bucket count %d", size)
	}

	var hlc roaringArray64
	// Capped so that a header claiming billions of buckets costs nothing until
	// the buckets actually show up.
	capHint := size
	if capHint > 1024 {
		capHint = 1024
	}
	if capHint > 0 {
		hlc.keys = make([]uint32, 0, capHint)
		hlc.containers = make([]*roaring.Bitmap, 0, capHint)
		hlc.needCopyOnWrite = make([]bool, 0, capHint)
	}

	keyBuf := buf[:4]
	previousKey := uint32(0)
	for i := uint64(0); i < size; i++ {
		n, err = io.ReadFull(stream, keyBuf)
		p += int64(n)
		if err != nil {
			return p, fmt.Errorf("error in bitmap.ReadPortableFrom: could not read key #%d: %w", i, err)
		}
		key := binary.LittleEndian.Uint32(keyBuf)
		if i > 0 && key <= previousKey {
			return p, fmt.Errorf("error in bitmap.ReadPortableFrom: bucket keys must be strictly increasing, key #%d is %d after %d", i, key, previousKey)
		}
		previousKey = key

		c := roaring.NewBitmap()
		nc, cerr := c.ReadFrom(stream)
		p += nc
		if cerr != nil {
			return p, fmt.Errorf("error in bitmap.ReadPortableFrom: could not deserialize bitmap for key #%d: %w", i, cerr)
		}
		if nc == 0 {
			return p, fmt.Errorf("error in bitmap.ReadPortableFrom: could not deserialize bitmap for key #%d", i)
		}
		// The format never stores an empty bucket; drop it rather than break
		// the invariant that every container is non-empty.
		if c.IsEmpty() {
			continue
		}
		hlc.appendContainer(key, c, false)
	}
	hlc.copyOnWrite = rb.highlowcontainer.copyOnWrite
	if hlc.copyOnWrite {
		hlc.ensureCowMu()
	}
	rb.highlowcontainer = hlc
	return p, nil
}
