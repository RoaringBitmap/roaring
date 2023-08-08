package roaring64

// to run just these tests: go test -run TestParAggregations

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func testAggregations(t *testing.T,
	and func(bitmaps ...*Bitmap) *Bitmap,
	or func(bitmaps ...*Bitmap) *Bitmap,
	xor func(bitmaps ...*Bitmap) *Bitmap) {

	t.Run("simple case", func(t *testing.T) {
		rb1 := NewBitmap()
		rb2 := NewBitmap()
		rb1.Add(uint64(1))
		rb2.Add(uint64(2))

		assertAggregation(t, 0, and, rb1, rb2)
		assertAggregation(t, 2, or, rb1, rb2)
		assertAggregation(t, 2, xor, rb1, rb2)
	})

	t.Run("aggregate nothing", func(t *testing.T) {
		assertAggregation(t, 0, and)
		assertAggregation(t, 0, or)
		assertAggregation(t, 0, xor)
	})

	t.Run("single bitmap", func(t *testing.T) {
		rb := BitmapOf(1, 2, 3)

		assertAggregation(t, 3, and, rb)
		assertAggregation(t, 3, or, rb)
		assertAggregation(t, 3, xor, rb)
	})

	t.Run("empty and single elem bitmaps", func(t *testing.T) {
		rb1 := NewBitmap()
		rb2 := BitmapOf(1)

		assertAggregation(t, 0, and, rb1, rb2)
		assertAggregation(t, 1, or, rb1, rb2)
		assertAggregation(t, 1, xor, rb1, rb2)
	})

	t.Run("two single elem disjoint sets", func(t *testing.T) {
		rb1 := BitmapOf(1)
		rb2 := BitmapOf(2)

		assertAggregation(t, 0, and, rb1, rb2)
		assertAggregation(t, 2, or, rb1, rb2)
	})

	t.Run("3 bitmaps with CoW set (not in order of definition)", func(t *testing.T) {
		rb1 := NewBitmap()
		rb2 := NewBitmap()
		rb3 := NewBitmap()
		rb1.SetCopyOnWrite(true)
		rb2.SetCopyOnWrite(true)
		rb3.SetCopyOnWrite(true)
		rb1.Add(uint64(1))
		rb1.Add(uint64(100000))
		rb2.Add(uint64(200000))
		rb3.Add(uint64(1))
		rb3.Add(uint64(300000))

		assertAggregation(t, 0, and, rb2, rb1, rb3)
		assertAggregation(t, 4, or, rb2, rb1, rb3)
		assertAggregation(t, 3, xor, rb2, rb1, rb3)
	})

	t.Run("3 bitmaps (not in order of definition)", func(t *testing.T) {
		rb1 := NewBitmap()
		rb2 := NewBitmap()
		rb3 := NewBitmap()
		rb1.Add(uint64(1))
		rb1.Add(uint64(100000))
		rb2.Add(uint64(200000))
		rb3.Add(uint64(1))
		rb3.Add(uint64(300000))

		assertAggregation(t, 0, and, rb2, rb1, rb3)
		assertAggregation(t, 4, or, rb2, rb1, rb3)
		assertAggregation(t, 3, xor, rb2, rb1, rb3)
	})

	t.Run("3 bitmaps", func(t *testing.T) {
		rb1 := NewBitmap()
		rb2 := NewBitmap()
		rb3 := NewBitmap()
		rb1.Add(uint64(1))
		rb1.Add(uint64(100000))
		rb2.Add(uint64(200000))
		rb3.Add(uint64(1))
		rb3.Add(uint64(300000))

		assertAggregation(t, 0, and, rb1, rb2, rb3)
		assertAggregation(t, 4, or, rb1, rb2, rb3)
		assertAggregation(t, 3, xor, rb1, rb2, rb3)
	})

	t.Run("3 bitmaps with CoW set", func(t *testing.T) {
		rb1 := NewBitmap()
		rb2 := NewBitmap()
		rb3 := NewBitmap()
		rb1.SetCopyOnWrite(true)
		rb2.SetCopyOnWrite(true)
		rb3.SetCopyOnWrite(true)
		rb1.Add(uint64(1))
		rb1.Add(uint64(100000))
		rb2.Add(uint64(200000))
		rb3.Add(uint64(1))
		rb3.Add(uint64(300000))

		assertAggregation(t, 0, and, rb1, rb2, rb3)
		assertAggregation(t, 4, or, rb1, rb2, rb3)
		assertAggregation(t, 3, xor, rb1, rb2, rb3)
	})

	t.Run("advanced case", func(t *testing.T) {
		rb1 := NewBitmap()
		rb2 := NewBitmap()
		rb3 := NewBitmap()
		for i := uint64(0); i < 1000000; i += 3 {
			rb1.Add(i)
		}
		for i := uint64(0); i < 1000000; i += 7 {
			rb2.Add(i)
		}
		for i := uint64(0); i < 1000000; i += 1001 {
			rb3.Add(i)
		}
		for i := uint64(1000000); i < 2000000; i += 1001 {
			rb1.Add(i)
		}
		for i := uint64(1000000); i < 2000000; i += 3 {
			rb2.Add(i)
		}
		for i := uint64(1000000); i < 2000000; i += 7 {
			rb3.Add(i)
		}
		for i := uint64(1000000000000000); i < 20000000000000000; i += 100000000000 {
			rb3.Add(i)
		}

		rb1.Or(rb2)
		rb1.Or(rb3)
		bigand := And(And(rb1, rb2), rb3)
		bigxor := Xor(Xor(rb1, rb2), rb3)

		if or != nil {
			assert.True(t, or(rb1, rb2, rb3).Equals(rb1))
		}

		if and != nil {
			assert.True(t, and(rb1, rb2, rb3).Equals(bigand))
		}

		if xor != nil {
			assert.True(t, xor(rb1, rb2, rb3).Equals(bigxor))
		}
	})

	t.Run("advanced case with runs", func(t *testing.T) {
		rb1 := NewBitmap()
		rb2 := NewBitmap()
		rb3 := NewBitmap()
		for i := uint64(500); i < 75000; i++ {
			rb1.Add(i)
		}
		for i := uint64(0); i < 1000000; i += 7 {
			rb2.Add(i)
		}
		for i := uint64(0); i < 1000000; i += 1001 {
			rb3.Add(i)
		}
		for i := uint64(1000000); i < 2000000; i += 1001 {
			rb1.Add(i)
		}
		for i := uint64(1000000); i < 2000000; i += 3 {
			rb2.Add(i)
		}
		for i := uint64(1000000); i < 2000000; i += 7 {
			rb3.Add(i)
		}
		rb1.RunOptimize()

		rb1.Or(rb2)
		rb1.Or(rb3)
		bigand := And(And(rb1, rb2), rb3)
		bigxor := Xor(Xor(rb1, rb2), rb3)

		if or != nil {
			assert.True(t, or(rb1, rb2, rb3).Equals(rb1))
		}

		if and != nil {
			assert.True(t, and(rb1, rb2, rb3).Equals(bigand))
		}

		if xor != nil {
			assert.True(t, xor(rb1, rb2, rb3).Equals(bigxor))
		}
	})

	t.Run("issue 178", func(t *testing.T) {
		ba1 := []uint64{3585028, 65901253, 143441994, 211160474, 286511937, 356744840, 434332509, 502812785, 576097614, 646557334, 714794241, 775083485, 833704249, 889329147, 941367043}
		ba2 := []uint64{17883, 54494426, 113908938, 174519827, 235465665, 296685741, 357644666, 420192495, 476104304, 523046142, 577855081, 634889665, 692460635, 751350463, 809989192, 863494316, 919127240}

		r1 := BitmapOf(ba1...)
		r2 := BitmapOf(ba2...)

		assertAggregation(t, 32, or, r1, r2)
	})
}

func assertAggregation(t *testing.T, expected uint64, aggr func(bitmaps ...*Bitmap) *Bitmap, bitmaps ...*Bitmap) {
	if aggr != nil {
		assert.Equal(t, aggr(bitmaps...).GetCardinality(), expected)
	}
}

func TestParAggregations(t *testing.T) {
	for _, p := range [...]int{0, 1, 2, 4} {
		//andFunc := func(bitmaps ...*Bitmap) *Bitmap {
		//	return ParAnd(p, bitmaps...)
		//}
		orFunc := func(bitmaps ...*Bitmap) *Bitmap {
			return ParOr(p, bitmaps...)
		}

		t.Run(fmt.Sprintf("par%d", p), func(t *testing.T) {
			//testAggregations(t, andFunc, orFunc, nil)
			testAggregations(t, nil, orFunc, nil)
		})
	}
}

func TestParAggregations2(t *testing.T) {
	orFunc := func(bitmaps ...*Bitmap) *Bitmap {
		return ParOr(0, bitmaps...)
	}
	t.Run("par0", func(t *testing.T) {
		testAggregations(t, nil, orFunc, nil)
	})
}

func TestFastAggregations(t *testing.T) {
	testAggregations(t, nil, FastOr, nil)
}
