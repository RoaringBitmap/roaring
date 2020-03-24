package roaring64

import "github.com/RoaringBitmap/roaring"

func highbits(x uint64) uint32 {
	return uint32(x >> 32)
}

func lowbits(x uint64) uint32 {
	return uint32(x & maxLowBit)
}

const maxLowBit = roaring.MaxUint32
