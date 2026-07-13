package roaring

import (
	"encoding/binary"
	"math/rand"
	"testing"
)

func getRandomUint64Set(n int) []uint64 {
	seed := int64(42)
	rand.Seed(seed)

	var buf [8]byte
	var o []uint64
	for i := 0; i < n; i++ {
		rand.Read(buf[:])
		o = append(o, binary.LittleEndian.Uint64(buf[:]))
	}
	return o
}

func BenchmarkPopcount(b *testing.B) {
	b.StopTimer()

	r := getRandomUint64Set(64)

	b.ResetTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		popcntSlice(r)
	}
}
