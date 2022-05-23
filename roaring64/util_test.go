package roaring64

import (
	"math"
	"testing"

	"github.com/RoaringBitmap/roaring"
	"github.com/stretchr/testify/assert"
)

func Test_CreateR64FromR32Slice(t *testing.T) {
	nums := []uint64{1, math.MaxUint32, math.MaxUint32 + 1, math.MaxUint32 * 2, math.MaxUint32*2 + 1, math.MaxUint32 * 3}
	bms := map[uint32]*roaring.Bitmap{
		0: roaring.NewBitmap(),
		1: roaring.NewBitmap(),
		2: roaring.NewBitmap(),
	}
	r64a := NewBitmap()
	for _, v := range nums {
		h := highbits(v)
		l := lowbits(v)
		if _, ok := bms[h]; !ok {
			bms[h] = roaring.NewBitmap()
		}
		bms[h].Add(l)
		r64a.Add(v)
	}
	r32 := make([]*roaring.Bitmap, len(bms))
	for i, l := 0, len(bms); i < l; i++ {
		r32[i] = bms[uint32(i)]
	}
	r64b := CreateR64FromR32Slice(r32...)
	assert.Equal(t, r64a.GetCardinality(), r64b.GetCardinality())
	for _, v := range nums {
		assert.Equal(t, r64a.Contains(v), r64b.Contains(v))
	}
}
