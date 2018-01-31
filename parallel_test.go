package roaring

// to run just these tests: go test -run TestParAggregations

import (
	"testing"
)

func TestParAggregations(t *testing.T) {
	t.Run("simple case", func(t *testing.T) {
		rb1 := NewBitmap()
		rb2 := NewBitmap()
		rb1.Add(1)
		rb2.Add(2)

		if ParAnd(0, rb1, rb2).GetCardinality() != 0 {
			t.Fail()
		}
		if ParOr(0, rb1, rb2).GetCardinality() != 2 {
			t.Fail()
		}
	})

	t.Run("aggregate nothing", func(t *testing.T) {
		if ParAnd(0).GetCardinality() != 0 {
			t.Fail()
		}

		if ParOr(0).GetCardinality() != 0 {
			t.Fail()
		}
	})

	t.Run("single bitmap", func(t *testing.T) {
		rb := BitmapOf(1, 2, 3)

		if ParAnd(0, rb).GetCardinality() != 3 {
			t.Fail()
		}

		if ParOr(0, rb).GetCardinality() != 3 {
			t.Fail()
		}
	})

	t.Run("empty and single elem bitmaps", func(t *testing.T) {
		rb1 := NewBitmap()
		rb2 := BitmapOf(1)

		if ParAnd(0, rb1, rb2).GetCardinality() != 0 {
			t.Fail()
		}
		if ParOr(0, rb1, rb2).GetCardinality() != 1 {
			t.Fail()
		}
	})

	t.Run("two single elem disjoint sets", func(t *testing.T) {
		rb1 := BitmapOf(1)
		rb2 := BitmapOf(2)

		if ParAnd(0, rb1, rb2).Stats().Containers != 0 {
			t.Fail()
		}

		if ParOr(0, rb1, rb2).GetCardinality() != 2 {
			t.Fail()
		}
	})

	t.Run("3 bitmaps with CoW set (not in order of definition)", func(t *testing.T) {
		rb1 := NewBitmap()
		rb2 := NewBitmap()
		rb3 := NewBitmap()
		rb1.SetCopyOnWrite(true)
		rb2.SetCopyOnWrite(true)
		rb3.SetCopyOnWrite(true)
		rb1.Add(1)
		rb1.Add(100000)
		rb2.Add(200000)
		rb3.Add(1)
		rb3.Add(300000)

		if ParAnd(0, rb2, rb1, rb3).GetCardinality() != 0 {
			t.Fail()
		}

		if ParOr(0, rb2, rb1, rb3).GetCardinality() != 4 {
			t.Fail()
		}
	})

	t.Run("3 bitmaps (not in order of definition)", func(t *testing.T) {
		rb1 := NewBitmap()
		rb2 := NewBitmap()
		rb3 := NewBitmap()
		rb1.Add(1)
		rb1.Add(100000)
		rb2.Add(200000)
		rb3.Add(1)
		rb3.Add(300000)

		if ParAnd(0, rb2, rb1, rb3).GetCardinality() != 0 {
			t.Fail()
		}

		if ParOr(0, rb2, rb1, rb3).GetCardinality() != 4 {
			t.Fail()
		}
	})

	t.Run("3 bitmaps", func(t *testing.T) {
		rb1 := NewBitmap()
		rb2 := NewBitmap()
		rb3 := NewBitmap()
		rb1.Add(1)
		rb1.Add(100000)
		rb2.Add(200000)
		rb3.Add(1)
		rb3.Add(300000)

		if ParAnd(0, rb1, rb2, rb3).GetCardinality() != 0 {
			t.Fail()
		}

		if ParOr(0, rb1, rb2, rb3).GetCardinality() != 4 {
			t.Fail()
		}
	})

	t.Run("3 bitmaps with CoW set", func(t *testing.T) {
		rb1 := NewBitmap()
		rb2 := NewBitmap()
		rb3 := NewBitmap()
		rb1.SetCopyOnWrite(true)
		rb2.SetCopyOnWrite(true)
		rb3.SetCopyOnWrite(true)
		rb1.Add(1)
		rb1.Add(100000)
		rb2.Add(200000)
		rb3.Add(1)
		rb3.Add(300000)

		if ParAnd(0, rb1, rb2, rb3).GetCardinality() != 0 {
			t.Fail()
		}

		if ParOr(0, rb1, rb2, rb3).GetCardinality() != 4 {
			t.Fail()
		}
	})

	t.Run("advanced case", func(t *testing.T) {
		rb1 := NewBitmap()
		rb2 := NewBitmap()
		rb3 := NewBitmap()
		for i := uint32(0); i < 1000000; i += 3 {
			rb1.Add(i)
		}
		for i := uint32(0); i < 1000000; i += 7 {
			rb2.Add(i)
		}
		for i := uint32(0); i < 1000000; i += 1001 {
			rb3.Add(i)
		}
		for i := uint32(1000000); i < 2000000; i += 1001 {
			rb1.Add(i)
		}
		for i := uint32(1000000); i < 2000000; i += 3 {
			rb2.Add(i)
		}
		for i := uint32(1000000); i < 2000000; i += 7 {
			rb3.Add(i)
		}

		rb1.Or(rb2)
		rb1.Or(rb3)
		bigand := And(And(rb1, rb2), rb3)

		if !ParOr(0, rb1, rb2, rb3).Equals(rb1) {
			t.Fail()
		}

		if !ParAnd(0, rb1, rb2, rb3).Equals(bigand) {
			t.Fail()
		}
	})

	t.Run("advanced case with runs", func(t *testing.T) {
		rb1 := NewBitmap()
		rb2 := NewBitmap()
		rb3 := NewBitmap()
		for i := uint32(500); i < 75000; i++ {
			rb1.Add(i)
		}
		for i := uint32(0); i < 1000000; i += 7 {
			rb2.Add(i)
		}
		for i := uint32(0); i < 1000000; i += 1001 {
			rb3.Add(i)
		}
		for i := uint32(1000000); i < 2000000; i += 1001 {
			rb1.Add(i)
		}
		for i := uint32(1000000); i < 2000000; i += 3 {
			rb2.Add(i)
		}
		for i := uint32(1000000); i < 2000000; i += 7 {
			rb3.Add(i)
		}
		rb1.RunOptimize()
		rb1.Or(rb2)
		rb1.Or(rb3)
		bigand := And(And(rb1, rb2), rb3)
		if !ParOr(0, rb1, rb2, rb3).Equals(rb1) {
			t.Fail()
		}

		if !ParAnd(0, rb1, rb2, rb3).Equals(bigand) {
			t.Fail()
		}
	})
}
