package roaring

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bits-and-blooms/bitset"
)

// BENCHMARKS, to run them type "go test -bench Benchmark -run -"

// go test -bench BenchmarkIteratorAlloc -benchmem -run -
func BenchmarkIteratorAlloc(b *testing.B) {
	bm := NewBitmap()
	domain := 100000000
	count := 10000
	for j := 0; j < count; j++ {
		v := uint32(rand.Intn(domain))
		bm.Add(v)
	}
	i := IntIterator{}
	expectedCardinality := bm.GetCardinality()
	counter := uint64(0)
	b.Run("simple iteration with alloc", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			counter = 0
			i := bm.Iterator()
			for i.HasNext() {
				i.Next()
				counter++
			}
		}
		b.StopTimer()
	})
	if counter != expectedCardinality {
		b.Fatalf("Cardinalities don't match: %d, %d", counter, expectedCardinality)
	}
	b.Run("simple iteration", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			counter = 0
			i.Initialize(bm)
			for i.HasNext() {
				i.Next()
				counter++
			}
		}
		b.StopTimer()
	})
	if counter != expectedCardinality {
		b.Fatalf("Cardinalities don't match: %d, %d", counter, expectedCardinality)
	}
	b.Run("reverse iteration with alloc", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			counter = 0
			ir := bm.ReverseIterator()
			for ir.HasNext() {
				ir.Next()
				counter++
			}
		}
		b.StopTimer()
	})
	if counter != expectedCardinality {
		b.Fatalf("Cardinalities don't match: %d, %d", counter, expectedCardinality)
	}
	ir := IntReverseIterator{}

	b.Run("reverse iteration", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			counter = 0
			ir.Initialize(bm)
			for ir.HasNext() {
				ir.Next()
				counter++
			}
		}
		b.StopTimer()
	})
	if counter != expectedCardinality {
		b.Fatalf("Cardinalities don't match: %d, %d", counter, expectedCardinality)
	}

	b.Run("many iteration with alloc", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			counter = 0
			buf := make([]uint32, 1024)
			im := bm.ManyIterator()
			for n := im.NextMany(buf); n != 0; n = im.NextMany(buf) {
				counter += uint64(n)
			}
		}
		b.StopTimer()
	})
	if counter != expectedCardinality {
		b.Fatalf("Cardinalities don't match: %d, %d", counter, expectedCardinality)
	}
	im := ManyIntIterator{}
	buf := make([]uint32, 1024)

	b.Run("many iteration", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			counter = 0
			im.Initialize(bm)
			for n := im.NextMany(buf); n != 0; n = im.NextMany(buf) {
				counter += uint64(n)
			}
		}
		b.StopTimer()
	})
	if counter != expectedCardinality {
		b.Fatalf("Cardinalities don't match: %d, %d", counter, expectedCardinality)
	}
}

// go test -bench BenchmarkOrs -benchmem -run -
func BenchmarkOrs(b *testing.B) {

	bms := []*Bitmap{}
	maxCount := 50
	domain := 100000000
	bitmapCount := 100
	for i := 0; i < bitmapCount; i++ {
		newBm := NewBitmap()
		count := rand.Intn(maxCount) + 5
		for j := 0; j < count; j++ {
			v := uint32(rand.Intn(domain))
			newBm.Add(v)
		}
		bms = append(bms, newBm)
	}
	var twotwocard uint64
	var fastcard uint64
	var nextcard uint64

	b.Run("two-by-two", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			newBm := NewBitmap()
			for _, bm := range bms {
				newBm.Or(bm)
			}
			twotwocard = newBm.GetCardinality()
		}
		b.StopTimer()
	})

	b.Run("fast", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			newBm := FastOr(bms...)
			fastcard = newBm.GetCardinality()
		}
		b.StopTimer()
	})

	b.Run("next/add", func(b *testing.B) {
		buf := make([]uint32, 100)
		for n := 0; n < b.N; n++ {
			newBm := NewBitmap()
			for _, bm := range bms {
				iter := bm.ManyIterator()
				for vs := iter.NextMany(buf); vs != 0; vs = iter.NextMany(buf) {
					newBm.AddMany(buf[:vs])
				}
			}
			nextcard = newBm.GetCardinality()
		}
		b.StopTimer()
	})
	if fastcard != nextcard {
		b.Fatalf("Cardinalities don't match: %d, %d", fastcard, nextcard)
	}
	if fastcard != twotwocard {
		b.Fatalf("Cardinalities don't match: %d, %d", fastcard, twotwocard)
	}
}

var Rb *Bitmap

func BenchmarkNewBitmap(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		Rb = New()
	}
}

var emptyArray []byte

func BenchmarkEmptyArray(b *testing.B) {
	for i := 0; i < b.N; i++ {
		emptyArray = make([]byte, 0)
	}
}

var c9 uint

// go test -bench BenchmarkIntersection -run -
func BenchmarkIntersectionBitset(b *testing.B) {
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
	b.StartTimer()
	card := uint(0)
	for j := 0; j < b.N; j++ {
		s3 := s1.Intersection(s2)
		card = card + s3.Count()
	}
}

// go test -bench BenchmarkIntersection -run -
func BenchmarkIntersectionRoaring(b *testing.B) {
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
	b.StartTimer()
	card := uint64(0)
	for j := 0; j < b.N; j++ {
		s3 := And(s1, s2)
		card = card + s3.GetCardinality()
	}
}

// go test -bench BenchmarkIntersectionCardinalityRoaring -run -
func BenchmarkIntersectionCardinalityRoaring(b *testing.B) {
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
	b.StartTimer()
	card := uint64(0)
	for j := 0; j < b.N; j++ {
		card += s1.AndCardinality(s2)
	}
}

// go test -bench BenchmarkUnion -run -
func BenchmarkUnionBitset(b *testing.B) {
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
	b.StartTimer()
	card := uint(0)
	for j := 0; j < b.N; j++ {
		s3 := s1.Union(s2)
		card = card + s3.Count()
	}
}

// go test -bench BenchmarkUnion -run -
func BenchmarkUnionRoaring(b *testing.B) {
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
	b.StartTimer()
	card := uint64(0)
	for j := 0; j < b.N; j++ {
		s3 := Or(s1, s2)
		card = card + s3.GetCardinality()
	}
}

// BenchmarkUnionInPlaceCopyOnWrite tests the performance of bitmap.Or()
// when the bitmap was generated via FromBuffer.
// In this case all left containers need to be copied in order to be updated.
// The nested for-loops test a number of different scenarios
// with respect to the ranges and densities of bitmaps.
func BenchmarkUnionInPlaceCopyOnWrite(b *testing.B) {
	//uint32s to maintain 1.12 compatibility, which requires unsigned shifts.
	startingContainerPower := uint32(4)
	finalContainerPower := uint32(10)
	containerIncrement := uint32(3)
	startingItemsPower := uint32(3)
	finalItemsPower := uint32(10)
	itemsIncrement := uint32(7)
	for leftContainerPower := startingContainerPower; leftContainerPower <= finalContainerPower; leftContainerPower += containerIncrement {
		for rightContainerPower := startingContainerPower; rightContainerPower <= finalContainerPower; rightContainerPower += containerIncrement {
			for leftItemsPerContainerPower := startingItemsPower; leftItemsPerContainerPower <= finalItemsPower; leftItemsPerContainerPower += itemsIncrement {
				for rightItemsPerContainerPower := startingItemsPower; rightItemsPerContainerPower <= finalItemsPower; rightItemsPerContainerPower += itemsIncrement {
					b.Run(fmt.Sprintf("%d-%d-%d-%d", leftContainerPower, rightContainerPower, leftItemsPerContainerPower, rightItemsPerContainerPower),
						func(b *testing.B) {
							leftMax := (1 << 16) << leftContainerPower
							rightMax := (1 << 16) << rightContainerPower
							leftItems := 1 << (leftContainerPower + leftItemsPerContainerPower)
							rightItems := 1 << (rightContainerPower + rightItemsPerContainerPower)
							left := make([][]byte, 10)
							right := make([]*Bitmap, 10)
							for i := 0; i < 10; i++ {
								right[i] = NewBitmap()
								left[i] = generateRandomBitmap(b, leftMax, leftItems)
								_, err := right[i].FromBuffer(generateRandomBitmap(b, rightMax, rightItems))
								require.NoError(b, err)
							}
							// This tests a destructive operation, Or() so have to have a fresh bitmap per test.
							targetLefts := make([]*Bitmap, b.N)
							for i := 0; i < b.N; i++ {
								targetLefts[i] = NewBitmap()
								_, err := targetLefts[i].FromBuffer(left[i%10])
								require.NoError(b, err)
							}
							runActualBenchmark(b, targetLefts, right)
						})
				}
			}
		}
	}
}

// runActualBenchmark is broken out primarily so you can profile the tests,
// as otherwise the generation overwhelms the actual test.
func runActualBenchmark(b *testing.B, targetLefts []*Bitmap, right []*Bitmap) uint64 {
	b.ResetTimer()
	b.ReportAllocs()
	total := uint64(0)
	for i := 0; i < b.N; i++ {
		targetLefts[i].Or(right[i%10])
		total += targetLefts[i].GetCardinality()
	}
	return total
}

func generateRandomBitmap(b *testing.B, max, terms int) []byte {
	bitmap := NewBitmap()
	for i := 0; i < terms; i++ {
		bitmap.Add(uint32(rand.Intn(max)))
	}
	result, err := bitmap.ToBytes()
	require.NoError(b, err)
	return result
}

func BenchmarkChecksum(b *testing.B) {
	for terms := 1; terms <= (1 << 20); terms *= 2 {
		rb1, rb2 := NewBitmap(), NewBitmap()
		set := map[uint32]struct{}{}
		for len(set) < terms {
			v := uint32(rand.Intn(1500000))
			set[v] = struct{}{}
			rb1.Add(v)
			rb2.AddRange(uint64(v), uint64(v)+100)
		}
		rb2.RunOptimize()

		b.Run(fmt.Sprintf("checksum-%d", terms), func(b *testing.B) {
			b.ReportAllocs()
			rb1.Checksum()
		})
		b.Run(fmt.Sprintf("checksum-compressed-%d", terms), func(b *testing.B) {
			b.ReportAllocs()
			rb2.Checksum()
		})
	}
}

// go test -bench BenchmarkSet -run -
func BenchmarkSetRoaring(b *testing.B) {
	b.StopTimer()
	r := rand.New(rand.NewSource(0))
	sz := 1000000
	s := NewBitmap()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		s.Add(uint32(r.Int31n(int32(sz))))
	}
}

func BenchmarkSetBitset(b *testing.B) {
	b.StopTimer()
	r := rand.New(rand.NewSource(0))
	sz := 1000000
	s := bitset.New(0)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		s.Set(uint(r.Int31n(int32(sz))))
	}
}

// go test -bench BenchmarkGetTest -run -
func BenchmarkGetTestRoaring(b *testing.B) {
	b.StopTimer()
	r := rand.New(rand.NewSource(0))
	sz := 1000000
	initsize := 50000
	s := NewBitmap()
	for i := 0; i < initsize; i++ {
		s.Add(uint32(r.Int31n(int32(sz))))
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		s.Contains(uint32(r.Int31n(int32(sz))))
	}
}

func BenchmarkGetTestBitSet(b *testing.B) {
	b.StopTimer()
	r := rand.New(rand.NewSource(0))
	sz := 1000000
	initsize := 50000
	s := bitset.New(0)
	for i := 0; i < initsize; i++ {
		s.Set(uint(r.Int31n(int32(sz))))
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		s.Test(uint(r.Int31n(int32(sz))))
	}
}

// go test -bench BenchmarkCount -run -
func BenchmarkCountRoaring(b *testing.B) {
	b.StopTimer()
	r := rand.New(rand.NewSource(0))
	s := NewBitmap()
	sz := 1000000
	initsize := 50000
	for i := 0; i < initsize; i++ {
		s.Add(uint32(r.Int31n(int32(sz))))
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		s.GetCardinality()
	}
}

func BenchmarkCountBitset(b *testing.B) {
	b.StopTimer()
	r := rand.New(rand.NewSource(0))
	s := bitset.New(0)
	sz := 1000000
	initsize := 50000
	for i := 0; i < initsize; i++ {

		s.Set(uint(r.Int31n(int32(sz))))
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		s.Count()
	}
}

// go test -bench BenchmarkIterate -run -
func BenchmarkIterateRoaring(b *testing.B) {
	newBitmap := func() *Bitmap {
		r := rand.New(rand.NewSource(0))
		s := NewBitmap()
		sz := 150000
		initsize := 65000
		for i := 0; i < initsize; i++ {
			s.Add(uint32(r.Int31n(int32(sz))))
		}
		return s
	}

	b.Run("iterator-compressed", func(b *testing.B) {
		b.ReportAllocs()

		s := newBitmap()
		s.RunOptimize()

		b.ResetTimer()

		for j := 0; j < b.N; j++ {
			c9 = uint(0)
			i := s.Iterator()
			for i.HasNext() {
				i.Next()
				c9++
			}
		}
	})

	b.Run("iterator", func(b *testing.B) {
		b.ReportAllocs()

		s := newBitmap()

		b.ResetTimer()

		for j := 0; j < b.N; j++ {
			c9 = uint(0)
			i := s.Iterator()
			for i.HasNext() {
				i.Next()
				c9++
			}
		}
	})

	b.Run("iterate-compressed", func(b *testing.B) {
		b.ReportAllocs()

		s := newBitmap()
		s.RunOptimize()

		b.ResetTimer()

		for j := 0; j < b.N; j++ {
			c9 = uint(0)
			s.Iterate(func(x uint32) bool {
				c9++
				return true
			})
		}
	})

	b.Run("iterate", func(b *testing.B) {
		b.ReportAllocs()

		s := newBitmap()

		b.ResetTimer()

		for j := 0; j < b.N; j++ {
			c9 = uint(0)
			s.Iterate(func(x uint32) bool {
				c9++
				return true
			})
		}
	})
}

// go test -bench BenchmarkSparseIterate -run -
func BenchmarkSparseIterateRoaring(b *testing.B) {
	b.StopTimer()
	r := rand.New(rand.NewSource(0))
	s := NewBitmap()
	sz := 100000000
	initsize := 65000
	for i := 0; i < initsize; i++ {
		s.Add(uint32(r.Int31n(int32(sz))))
	}
	b.StartTimer()
	for j := 0; j < b.N; j++ {
		c9 = uint(0)
		i := s.Iterator()
		for i.HasNext() {
			i.Next()
			c9++
		}
	}

}

// go test -bench BenchmarkSparseAdvance -run -
func BenchmarkSparseAdvanceRoaring(b *testing.B) {
	b.StopTimer()

	s := NewBitmap()
	initsize := 65000

	for i := 0; i < initsize; i++ {
		s.Add(uint32(i))
	}

	for _, gap := range []int{1, 2, 65, 650} {
		b.Run(fmt.Sprintf("advance from %d", gap), func(b *testing.B) {
			b.ReportAllocs()
			b.StartTimer()

			diff := uint32(0)

			for n := 0; n < b.N; n++ {
				val := uint32((gap * n) % initsize)

				i := s.Iterator()
				i.AdvanceIfNeeded(val)

				diff += i.PeekNext() - val
			}

			b.StopTimer()

			if diff != 0 {
				b.Fatalf("Expected diff 0, got %d", diff)
			}
		})
	}
}

// go test -bench BenchmarkSparseAdvance -run -
func BenchmarkSparseAdvanceOnHugeData(b *testing.B) {
	b.ReportAllocs()

	s := NewBitmap()
	initsize := 6500000
	sz := 100000000
	r := rand.New(rand.NewSource(0))

	for i := 0; i < initsize; i++ {
		s.Add(uint32(r.Int31n(int32(sz))))
	}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		val := uint32(n)

		i := s.Iterator()
		i.AdvanceIfNeeded(val)
	}
}

// go test -bench BenchmarkSparseAdvance -run -
func BenchmarkSparseAdvanceSequentially(b *testing.B) {
	b.StopTimer()

	s := NewBitmap()
	initsize := 65000

	for i := 0; i < initsize; i++ {
		s.Add(uint32(i))
	}

	for _, gap := range []int{1, 2, 65, 650} {
		b.Run(fmt.Sprintf("advance from %d", gap), func(b *testing.B) {
			b.ReportAllocs()
			b.StartTimer()

			diff := uint32(0)

			for n := 0; n < b.N; n++ {
				val := uint32((gap * n) % initsize)

				i := s.Iterator()

				for i.HasNext() && i.PeekNext() < val {
					i.Next()
				}

				diff += i.PeekNext() - val
			}

			b.StopTimer()

			if diff != 0 {
				b.Fatalf("Expected diff 0, got %d", diff)
			}
		})
	}
}

// go test -bench BenchmarkIterate -run -
func BenchmarkIterateBitset(b *testing.B) {
	b.StopTimer()
	r := rand.New(rand.NewSource(0))
	s := bitset.New(0)
	sz := 150000
	initsize := 65000
	for i := 0; i < initsize; i++ {
		s.Set(uint(r.Int31n(int32(sz))))
	}
	b.StartTimer()
	for j := 0; j < b.N; j++ {
		c9 = uint(0)
		for i, e := s.NextSet(0); e; i, e = s.NextSet(i + 1) {
			c9++
		}
	}
}

// go test -bench BenchmarkSparseContains -run -
func BenchmarkSparseContains(b *testing.B) {
	b.StopTimer()
	r := rand.New(rand.NewSource(0))
	s := NewBitmap()
	sz := 10000000
	initsize := 65000
	for i := 0; i < initsize; i++ {
		s.Add(uint32(r.Int31n(int32(sz))))
	}
	var a [1024]uint32
	for i := 0; i < 1024; i++ {
		a[i] = uint32(r.Int31n(int32(sz)))
	}
	b.StartTimer()
	for j := 0; j < b.N; j++ {
		c9 = uint(0)
		for i := 0; i < 1024; i++ {
			if s.Contains(a[i]) {
				c9++
			}

		}
	}
}

// go test -bench BenchmarkSparseIterate -run -
func BenchmarkSparseIterateBitset(b *testing.B) {
	b.StopTimer()
	r := rand.New(rand.NewSource(0))
	s := bitset.New(0)
	sz := 100000000
	initsize := 65000
	for i := 0; i < initsize; i++ {
		s.Set(uint(r.Int31n(int32(sz))))
	}
	b.StartTimer()
	for j := 0; j < b.N; j++ {
		c9 = uint(0)
		for i, e := s.NextSet(0); e; i, e = s.NextSet(i + 1) {
			c9++
		}
	}
}

func BenchmarkSerializationSparse(b *testing.B) {
	b.ReportAllocs()
	b.StopTimer()
	r := rand.New(rand.NewSource(0))
	s := NewBitmap()
	sz := 100000000
	initsize := 65000
	for i := 0; i < initsize; i++ {
		s.Add(uint32(r.Int31n(int32(sz))))
	}
	buf := make([]byte, 0, s.GetSerializedSizeInBytes())
	b.StartTimer()

	for j := 0; j < b.N; j++ {
		w := bytes.NewBuffer(buf[:0])
		s.WriteTo(w)
	}
}

func BenchmarkSerializationMid(b *testing.B) {
	b.ReportAllocs()
	b.StopTimer()
	r := rand.New(rand.NewSource(0))
	s := NewBitmap()
	sz := 10000000
	initsize := 65000
	for i := 0; i < initsize; i++ {
		s.Add(uint32(r.Int31n(int32(sz))))
	}
	buf := make([]byte, 0, s.GetSerializedSizeInBytes())
	b.StartTimer()

	for j := 0; j < b.N; j++ {
		w := bytes.NewBuffer(buf[:0])
		s.WriteTo(w)
	}
}

func BenchmarkSerializationDense(b *testing.B) {
	b.ReportAllocs()
	b.StopTimer()
	r := rand.New(rand.NewSource(0))
	s := NewBitmap()
	sz := 150000
	initsize := 65000
	for i := 0; i < initsize; i++ {
		s.Add(uint32(r.Int31n(int32(sz))))
	}
	buf := make([]byte, 0, s.GetSerializedSizeInBytes())
	b.StartTimer()

	for j := 0; j < b.N; j++ {
		w := bytes.NewBuffer(buf[:0])
		s.WriteTo(w)
	}
}

func BenchmarkMarshalBinary(b *testing.B) {
	r := rand.New(rand.NewSource(0))
	s := NewBitmap()
	sz := 10000000
	initsize := 65000
	for i := 0; i < initsize; i++ {
		s.Add(uint32(r.Int31n(int32(sz))))
	}
	b.ReportAllocs()
	b.ResetTimer()
	for j := 0; j < b.N; j++ {
		s.MarshalBinary()
	}
}

func BenchmarkUnmarshalBinary(b *testing.B) {
	r := rand.New(rand.NewSource(0))
	s := NewBitmap()
	sz := 10000000
	initsize := 65000
	for i := 0; i < initsize; i++ {
		s.Add(uint32(r.Int31n(int32(sz))))
	}
	data, _ := s.MarshalBinary()
	b.ReportAllocs()
	b.ResetTimer()
	for j := 0; j < b.N; j++ {
		ub := NewBitmap()
		ub.UnmarshalBinary(data)
	}
}

func BenchmarkEqualsSparse(b *testing.B) {
	b.StopTimer()
	r := rand.New(rand.NewSource(0))
	s := NewBitmap()
	t := NewBitmap()
	sz := 100000000
	initsize := 65000
	for i := 0; i < initsize; i++ {
		n := uint32(r.Int31n(int32(sz)))
		s.Add(n)
		t.Add(n)
	}
	b.StartTimer()

	for j := 0; j < b.N; j++ {
		s.Equals(t)
	}
}

func BenchmarkEqualsClone(b *testing.B) {
	b.StopTimer()
	r := rand.New(rand.NewSource(0))
	s := NewBitmap()
	sz := 100000000
	initsize := 65000
	for i := 0; i < initsize; i++ {
		s.Add(uint32(r.Int31n(int32(sz))))
	}
	t := s.Clone()
	b.StartTimer()

	for j := 0; j < b.N; j++ {
		s.Equals(t)
	}
}

// go test -bench BenchmarkNexts -benchmem -run -
func BenchmarkNexts(b *testing.B) {

	for _, gap := range []uint32{1, 2, 4, 8, 16, 32, 64, 256, 1024, 8096} {

		rrs := make([]uint32, 500000)
		v := uint32(0)
		for i := range rrs {
			rrs[i] = v
			v += gap
		}

		bm := NewBitmap()
		bm.AddMany(rrs)

		var totnext uint64
		var totnextmany uint64

		density := float32(100) / float32(gap)

		densityStr := fmt.Sprintf("__%f%%", density)

		b.Run("next"+densityStr, func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				totnext = 0
				iter := bm.Iterator()
				for iter.HasNext() {
					v := iter.Next()
					totnext += uint64(v)
				}
			}
			b.StopTimer()
		})

		b.Run("nextmany"+densityStr, func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				totnextmany = 0
				iter := bm.ManyIterator()
				// worst case, in practice will reuse buffers across many roars
				buf := make([]uint32, 4096)
				for j := iter.NextMany(buf); j != 0; j = iter.NextMany(buf) {
					for i := 0; i < j; i++ {
						totnextmany += uint64(buf[i])
					}
				}
			}
			b.StopTimer()
		})

		if totnext != totnextmany {
			b.Fatalf("Cardinalities don't match: %d, %d", totnext, totnextmany)
		}
	}
}

// go test -bench BenchmarkRLENexts -benchmem -run -
func BenchmarkNextsRLE(b *testing.B) {

	var totadd uint64
	var totaddmany uint64

	bm := NewBitmap()
	bm.AddRange(0, 1000000)

	b.Run("next", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			totadd = 0
			iter := bm.Iterator()
			for iter.HasNext() {
				v := iter.Next()
				totadd += uint64(v)
			}
		}
		b.StopTimer()
	})

	b.Run("nextmany", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			totaddmany = 0
			iter := bm.ManyIterator()
			// worst case, in practice will reuse buffers across many roars
			buf := make([]uint32, 2048)
			for j := iter.NextMany(buf); j != 0; j = iter.NextMany(buf) {
				for i := 0; i < j; i++ {
					totaddmany += uint64(buf[i])
				}
			}
		}
		b.StopTimer()
	})
	if totadd != totaddmany {
		b.Fatalf("Cardinalities don't match: %d, %d", totadd, totaddmany)
	}

}

func BenchmarkXor(b *testing.B) {
	b.StopTimer()
	r := rand.New(rand.NewSource(0))
	s := NewBitmap()
	sz := 100000000
	initsize := 65000
	for i := 0; i < initsize; i++ {
		n := uint32(r.Int31n(int32(sz)))
		s.Add(n)
	}
	x2 := NewBitmap()
	for i := 0; i < initsize; i++ {
		n := uint32(r.Int31n(int32(sz)))
		x2.Add(n)
	}
	b.StartTimer()

	for j := 0; j < b.N; j++ {
		s.Clone().Xor(x2)
	}
}

func BenchmarkXorLopsided(b *testing.B) {
	b.StopTimer()
	r := rand.New(rand.NewSource(0))
	s := NewBitmap()
	sz := 100000000
	initsize := 65000
	for i := 0; i < initsize; i++ {
		n := uint32(r.Int31n(int32(sz)))
		s.Add(n)
	}
	x2 := NewBitmap()
	for i := 0; i < 32; i++ {
		n := uint32(r.Int31n(int32(sz)))
		x2.Add(n)
	}
	b.StartTimer()

	for j := 0; j < b.N; j++ {
		s.Clone().Xor(x2)
	}
}

func BenchmarkBitmapReuseWithoutClear(b *testing.B) {
	for j := 0; j < b.N; j++ {
		s := NewBitmap()
		for i := 0; i < 100000; i++ {
			s.Add(uint32(i * 4096))
		}
	}
}

func BenchmarkBitmapReuseWithClear(b *testing.B) {
	s := NewBitmap()
	for i := 0; i < 100000; i++ {
		s.Add(uint32(i * 4096))
	}
	b.ResetTimer()

	for j := 0; j < b.N; j++ {
		s.Clear() // reuse the same bitmap
		for i := 0; i < 100000; i++ {
			s.Add(uint32(i * 4096))
		}
	}
}

func BenchmarkAndAny(b *testing.B) {
	runSet := func(name string, base *Bitmap, filters []*Bitmap) {
		var (
			andFirstCard, orFirstCard, andAnyCard uint64
			andFirstRan, orFirstRan, andAnyRan    bool
		)

		b.Run(name+"_or-first", func(b *testing.B) {
			andAnyRan = true
			for n := 0; n < b.N; n++ {
				clone := base.Clone()

				b.StartTimer()
				clone.And(FastOr(filters...))
				orFirstCard = clone.GetCardinality()
				b.StopTimer()
			}
		})

		b.Run(name+"_and-first", func(b *testing.B) {
			orFirstRan = true
			for n := 0; n < b.N; n++ {
				anded := make([]*Bitmap, 0, len(filters))

				b.StartTimer()
				for _, f := range filters {
					anded = append(anded, And(base, f))
				}
				andFirstCard = FastOr(anded...).GetCardinality()
				b.StopTimer()
			}
		})

		b.Run(name+"_AndAny", func(b *testing.B) {
			andAnyRan = true
			for n := 0; n < b.N; n++ {
				clone := base.Clone()

				b.StartTimer()
				clone.AndAny(filters...)
				andAnyCard = clone.GetCardinality()
				b.StopTimer()
			}
		})

		if andFirstRan && andAnyRan && andFirstCard != andAnyCard {
			b.Fatalf("Cardinalities don't match: %d, %d", andFirstCard, andAnyCard)
		}
		if orFirstRan && andAnyRan && orFirstCard != andAnyCard {
			b.Fatalf("Cardinalities don't match: %d, %d", orFirstCard, andAnyCard)
		}
	}

	r := rand.New(rand.NewSource(0))

	genOne := func(r *rand.Rand, maxCount, domain int) *Bitmap {
		ret := NewBitmap()
		count := r.Intn(maxCount) + 5
		for j := 0; j < count; j++ {
			v := uint32(r.Intn(domain))
			ret.Add(v)
		}
		return ret
	}

	genMulti := func(r *rand.Rand, num, maxCount, domain int) []*Bitmap {
		var ret []*Bitmap
		for i := 0; i < num; i++ {
			ret = append(ret, genOne(r, maxCount, domain))
		}
		return ret
	}

	filtersNum := 10
	domain := 100000000

	smallSize := 100
	defaultSize := 100000
	largeSize := 1000000

	runSet("small-base", genOne(r, smallSize, domain), genMulti(r, filtersNum, largeSize, domain))
	runSet("small-filters", genOne(r, largeSize, domain), genMulti(r, filtersNum, smallSize, domain))
	runSet("equal", genOne(r, defaultSize, domain), genMulti(r, filtersNum, defaultSize, domain))
}

func BenchmarkAndNot(b *testing.B) {
	type generator struct {
		name string
		f    func() *Bitmap
	}
	makeRunContainer := generator{
		name: "run",
		f: func() *Bitmap {
			rb := NewBitmap()
			for i := 0; i < 100; i++ {
				start := rand.Intn(math.MaxUint16)
				limit := start + rand.Intn(math.MaxUint16-start)
				rb.AddRange(uint64(start), uint64(limit))
			}
			rb.RunOptimize()
			return rb
		},
	}

	makeArrayContainer := generator{
		name: "array",
		f: func() *Bitmap {
			rb := NewBitmap()
			for i := 0; i < arrayDefaultMaxSize/2; i++ {
				rb.Add(uint32(rand.Intn(math.MaxUint16)))
			}
			return rb
		},
	}
	makeBitmapContainer := generator{
		name: "bitmap",
		f: func() *Bitmap {
			buf := make([]uint64, 1024)
			for i := range buf {
				buf[i] = rand.Uint64()
			}

			return FromDense(buf, false)
		},
	}

	for _, inPlace := range []bool{true, false} {
		for _, leftGen := range []generator{makeRunContainer, makeArrayContainer, makeBitmapContainer} {
			for _, rightGen := range []generator{makeRunContainer, makeArrayContainer, makeBitmapContainer} {
				b.Run(fmt.Sprintf("inPlace=%v/left=%s/right=%s", inPlace, leftGen.name, rightGen.name), func(b *testing.B) {
					b.StopTimer()
					serializedLefts := make([][]byte, 1000)
					for i := range serializedLefts {
						var err error
						serializedLefts[i], err = leftGen.f().ToBytes()
						if err != nil {
							b.Fatal(err)
						}
					}
					serializedRights := make([][]byte, 1000)
					for i := range serializedRights {
						var err error
						serializedRights[i], err = rightGen.f().ToBytes()
						if err != nil {
							b.Fatal(err)
						}
					}

					lefts := make([]*Bitmap, b.N)
					for i := range lefts {
						buf := serializedLefts[i%len(serializedLefts)]
						lefts[i] = NewBitmap()
						if _, err := lefts[i].FromBuffer(buf); err != nil {
							b.Fatal(err)
						}
						lefts[i] = lefts[i].Clone()
					}
					rights := make([]*Bitmap, b.N)
					for i := range rights {
						buf := serializedRights[i%len(serializedRights)]
						rights[i] = NewBitmap()
						if _, err := rights[i].FromBuffer(buf); err != nil {
							b.Fatal(err)
						}
						rights[i] = rights[i].Clone()
					}
					b.StartTimer()
					for i := 0; i < b.N; i++ {
						if inPlace {
							lefts[i].AndNot(rights[i])
						} else {
							_ = AndNot(lefts[i], rights[i])
						}
					}
				})
			}
		}
	}
}
