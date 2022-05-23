package roaring64

import (
	"github.com/RoaringBitmap/roaring"
)

func highbits(x uint64) uint32 {
	return uint32(x >> 32)
}

func lowbits(x uint64) uint32 {
	return uint32(x & maxLowBit)
}

const maxLowBit = roaring.MaxUint32
const maxUint32 = roaring.MaxUint32

func minOfInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func minOfInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxOfInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func maxOfUint32(a, b uint32) uint32 {
	if a > b {
		return a
	}
	return b
}

func minOfUint32(a, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}

func CreateR64FromR32Slice(r32 ...*roaring.Bitmap) *Bitmap {
	if len(r32) == 0 {
		return NewBitmap()
	}
	size := len(r32)
	rb := NewBitmap()
	rb.highlowcontainer = roaringArray64{}
	rb.highlowcontainer.keys = make([]uint32, size)
	rb.highlowcontainer.containers = make([]*roaring.Bitmap, size)
	rb.highlowcontainer.needCopyOnWrite = make([]bool, size)
	for k, v := range r32 {
		rb.highlowcontainer.keys[k] = uint32(k)
		rb.highlowcontainer.containers[k] = v
		rb.highlowcontainer.needCopyOnWrite[k] = false
	}
	return rb
}
