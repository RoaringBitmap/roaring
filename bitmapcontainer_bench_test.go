package roaring

import (
	"math/rand"
	"testing"
)

var sink uint32

func BenchmarkBitmapContainerFillLeastSignificant16bits(b *testing.B) {
	r := rand.New(rand.NewSource(42))
	bc := newBitmapContainer()
	for i := 0; i < 32768; i++ {
		val := uint16(r.Intn(65536))
		bc.iadd(val)
	}

	x := make([]uint32, 65536)
	mask := uint32(123) << 16

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pos := bc.fillLeastSignificant16bits(x, 0, mask)
		sink += x[pos-1]
	}
}
