//go:build go1.13
// +build go1.13

package roaring

import (
	"math/rand"
	"runtime"
	"testing"

	"github.com/bits-and-blooms/bitset"
)

// go test -bench BenchmarkMemoryUsage -run -
func BenchmarkMemoryUsage(b *testing.B) {
	b.StopTimer()
	bitmaps := make([]*Bitmap, 0, 10)

	incr := uint32(1 << 16)
	max := uint32(1<<32 - 1)
	for x := 0; x < 10; x++ {
		rb := NewBitmap()

		var i uint32
		for i = 0; i <= max-incr; i += incr {
			rb.Add(i)
		}

		bitmaps = append(bitmaps, rb)
	}

	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	b.ReportMetric(float64(stats.HeapInuse), "HeapInUse")
	b.ReportMetric(float64(stats.HeapObjects), "HeapObjects")
	b.StartTimer()
}

// go test -bench BenchmarkSize -run -
func BenchmarkSizeBitset(b *testing.B) {
	b.StopTimer()
	r := rand.New(rand.NewSource(0))
	s1 := bitset.New(0)
	sz := 150000
	initsize := 65000
	for i := 0; i < initsize; i++ {
		s1.Set(uint(r.Int31n(int32(sz))))
	}
	s2 := bitset.New(0)
	sz = 100000000
	initsize = 65000
	for i := 0; i < initsize; i++ {
		s2.Set(uint(r.Int31n(int32(sz))))
	}
	b.ReportMetric(float64(s1.BinaryStorageSize()+s2.BinaryStorageSize())/(1024.0*1024), "BitsetSizeInMB")
}

// go test -bench BenchmarkSize -run -
func BenchmarkSizeRoaring(b *testing.B) {
	b.StopTimer()
	r := rand.New(rand.NewSource(0))
	s1 := NewBitmap()
	sz := 150000
	initsize := 65000
	for i := 0; i < initsize; i++ {
		s1.Add(uint32(r.Int31n(int32(sz))))
	}
	s2 := NewBitmap()
	sz = 100000000
	initsize = 65000
	for i := 0; i < initsize; i++ {
		s2.Add(uint32(r.Int31n(int32(sz))))
	}
	b.ReportMetric(float64(s1.GetSerializedSizeInBytes()+s2.GetSerializedSizeInBytes())/(1024.0*1024), "RoaringSizeInMB")
}
