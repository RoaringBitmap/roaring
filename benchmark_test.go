package roaring

import ("testing"
	"math/rand"
	"fmt"
)




// BENCHMARKS, to run them type "go test -bench Benchmark -run -"

func BenchmarkSet(b *testing.B) {
	b.StopTimer()
	r := rand.New(rand.NewSource(0))
	sz := 1000000
	s := NewRoaringBitmap()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		s.Add(int(r.Int31n(int32(sz))))
	}
}

func BenchmarkGetTest(b *testing.B) {
	b.StopTimer()
	r := rand.New(rand.NewSource(0))
	sz := 1000000
	initsize := 50000
	s := NewRoaringBitmap()
	for i := 0; i < initsize; i++ {
		s.Add(int(r.Int31n(int32(sz))))
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		s.Contains(int(r.Int31n(int32(sz))))
	}
}


// go test -bench=Count
func BenchmarkCount(b *testing.B) {
	b.StopTimer()
	r := rand.New(rand.NewSource(0))
	s := NewRoaringBitmap()
	sz := 1000000
	initsize := 50000
	for i := 0; i < initsize; i++ {
		s.Add(int(r.Int31n(int32(sz))))
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		s.GetCardinality()
	}
}

// go test -bench=Iterate
func BenchmarkIterate(b *testing.B) {
	b.StopTimer()
	r := rand.New(rand.NewSource(0))
	s := NewRoaringBitmap()
	sz := 150000
	initsize := 65000
	for i := 0; i < initsize; i++ {
		s.Add(int(r.Int31n(int32(sz))))
	}
	fmt.Print("---Iterating over ", s.GetCardinality()," integers---");
	b.StartTimer()
	for j := 0; j < b.N; j++ {
		c := uint(0)
		i := s.Iterator()
		for i.HasNext() {
			i.Next()
			c++
		}
	}
}

// go test -bench=SparseIterate
func BenchmarkSparseIterate(b *testing.B) {
	b.StopTimer()
	r := rand.New(rand.NewSource(0))
	s := NewRoaringBitmap()
	sz := 10000000
	initsize := 65000
	for i := 0; i < initsize; i++ {
		s.Add(int(r.Int31n(int32(sz))))
	}
	fmt.Print("---Iterating over ", s.GetCardinality()," integers---");
	b.StartTimer()
	for j := 0; j < b.N; j++ {
		c := uint(0)
		i := s.Iterator()
		for i.HasNext() {
			i.Next()
			c++
		}
	}

}
