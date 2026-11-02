package roaring64

// Edge cases at the 64-bit boundaries (2^16, 2^32, 2^63, 2^64-1), ported from
// CRoaring (cpp_roaring64_unit.cpp, roaring64_unit.cpp, cpp_unit.cpp) and Java
// (TestRoaring64Bitmap.java).
// To run just these tests: go test -run TestEdge64

import (
	"math"
	"math/rand"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	edgeK = uint64(1) << 32
	edgeH = uint64(1) << 63
	edgeM = uint64(math.MaxUint64)
)

func requireValidWithValues(t *testing.T, rb *Bitmap, values ...uint64) {
	t.Helper()
	require.NoError(t, rb.Validate())
	if len(values) == 0 {
		require.True(t, rb.IsEmpty())
		require.Equal(t, 0, rb.highlowcontainer.size())
		return
	}
	require.Equal(t, values, rb.ToArray())
}

// R1
func TestEdge64RangesAcrossHighKeyBoundary(t *testing.T) {
	K := edgeK
	rb := NewBitmap()
	rb.AddRange(K-2, K)
	requireValidWithValues(t, rb, K-2, K-1)
	assert.Equal(t, uint64(1), rb.Stats().Containers)
	assert.Equal(t, 1, rb.highlowcontainer.size())

	rb = NewBitmap()
	rb.AddRange(K-10, K+10)
	require.NoError(t, rb.Validate())
	assert.Equal(t, uint64(20), rb.GetCardinality())
	assert.Equal(t, K-10, rb.Minimum())
	assert.Equal(t, K+9, rb.Maximum())

	rb.RemoveRange(K-5, K+5)
	require.NoError(t, rb.Validate())
	assert.Equal(t, uint64(10), rb.GetCardinality())
	assert.True(t, rb.Contains(K-6))
	assert.False(t, rb.Contains(K-5))
	assert.False(t, rb.Contains(K+4))
	assert.True(t, rb.Contains(K+5))

	rb.RemoveRange(K-10, K)
	require.NoError(t, rb.Validate())
	assert.Equal(t, K+5, rb.Minimum())
	assert.Equal(t, uint64(5), rb.GetCardinality())

	// Several high keys at once, with exact boundaries at both ends.
	rb = NewBitmap()
	rb.AddRange(2*K, 5*K)
	assert.Equal(t, 3*K, rb.GetCardinality())
	rb.RemoveRange(3*K, 4*K)
	require.NoError(t, rb.Validate())
	assert.Equal(t, 2*K, rb.GetCardinality())
	assert.True(t, rb.Contains(3*K-1))
	assert.False(t, rb.Contains(3*K))
	assert.False(t, rb.Contains(4*K-1))
	assert.True(t, rb.Contains(4*K))
	assert.Equal(t, 2, rb.highlowcontainer.size())
}

// R2
func TestEdge64RangesTouchingMaxUint64(t *testing.T) {
	M := edgeM
	rb := NewBitmap()
	rb.AddRange(M-3, M)
	requireValidWithValues(t, rb, M-3, M-2, M-1)
	assert.False(t, rb.Contains(M))

	rb.Add(M)
	rb.RemoveRange(M-3, M)
	requireValidWithValues(t, rb, M)

	rb = NewBitmap()
	rb.AddRange(M-0x10000-10, M-0x10000+10)
	require.NoError(t, rb.Validate())
	assert.Equal(t, uint64(20), rb.GetCardinality())
	assert.Equal(t, M-0x10000-10, rb.Minimum())
	assert.Equal(t, M-0x10000+9, rb.Maximum())

	rb = BitmapOf(0, M-1, M)
	rb.RemoveRange(1, M)
	requireValidWithValues(t, rb, 0, M)
}

// R3
func TestEdge64AddRangeAcrossSignBit(t *testing.T) {
	H := edgeH
	rb := NewBitmap()
	rb.AddRange(H-2, H+2)
	requireValidWithValues(t, rb, H-2, H-1, H, H+1)
	assert.Equal(t, uint64(2), rb.Rank(H-1))
	v, err := rb.Select(2)
	require.NoError(t, err)
	assert.Equal(t, H, v)
}

// R4
func TestEdge64RangeWithMoreThan2To32Values(t *testing.T) {
	K := edgeK
	rb := NewBitmap()
	rb.AddRange(0, K+3)
	assert.Equal(t, K+3, rb.GetCardinality())
	v, err := rb.Select(0)
	require.NoError(t, err)
	assert.Equal(t, uint64(0), v)
	v, err = rb.Select(K + 2)
	require.NoError(t, err)
	assert.Equal(t, K+2, v)
	_, err = rb.Select(K + 3)
	assert.Error(t, err)
	assert.Equal(t, K+3, rb.Rank(edgeM))
	assert.Equal(t, K, rb.Rank(K-1))
}

// R5
func TestEdge64EmptyAndInvertedRangesAreNoOps(t *testing.T) {
	K := edgeK
	rb := BitmapOf(K + 19)
	rb.AddRange(5, 5)
	rb.AddRange(6, 5)
	rb.RemoveRange(K+19, K+19)
	rb.RemoveRange(K+20, K+19)
	rb.Flip(9, 9)
	rb.Flip(10, 9)
	rb.FlipInt(10, 9)
	requireValidWithValues(t, rb, K+19)
	requireValidWithValues(t, Flip(rb, 10, 9), K+19)
	requireValidWithValues(t, Flip(rb, 9, 9), K+19)
}

// flipBoth applies the in-place and the package-level Flip and checks that they
// agree; it returns the in-place result.
func flipBoth(t *testing.T, rb *Bitmap, start, end uint64) *Bitmap {
	t.Helper()
	before := rb.Clone()
	var copied *Bitmap
	require.NotPanics(t, func() { copied = Flip(rb, start, end) }, "Flip(%d, %d)", start, end)
	require.True(t, rb.Equals(before), "package Flip changed its input")
	require.NotPanics(t, func() { rb.Flip(start, end) }, "rb.Flip(%d, %d)", start, end)
	require.NoError(t, rb.Validate())
	require.NoError(t, copied.Validate())
	require.True(t, rb.Equals(copied), "Flip(%d, %d)", start, end)
	require.Equal(t, rb.GetCardinality(), copied.GetCardinality())
	return rb
}

// F2: flips that cross 2^32, or end exactly on a multiple of 2^32.
func TestEdge64FlipAcrossHighKeyBoundary(t *testing.T) {
	K := edgeK
	rb := flipBoth(t, BitmapOf(K-2, K, K+2, 1<<50), K-2, K+2)
	requireValidWithValues(t, rb, K-1, K+1, K+2, 1<<50)

	rb = flipBoth(t, NewBitmap(), K-2, K)
	requireValidWithValues(t, rb, K-2, K-1)
	assert.Equal(t, uint64(1), rb.Stats().Containers)
	assert.Equal(t, 1, rb.highlowcontainer.size())

	rb = flipBoth(t, BitmapOf(K-1), K-1, K)
	requireValidWithValues(t, rb)

	// The bucket in front of the range is emptied by the flip, the next one
	// does not exist yet.
	rb = flipBoth(t, BitmapOf(K-1), K-1, K+1)
	requireValidWithValues(t, rb, K)

	rb = flipBoth(t, BitmapOf(K-1, 3*K), K-1, K+1)
	requireValidWithValues(t, rb, K, 3*K)
	rb = flipBoth(t, BitmapOf(K-1, K+5, 3*K), K-1, 2*K+2)
	assert.Equal(t, K+2, rb.GetCardinality())
	assert.False(t, rb.Contains(K+5))
	assert.True(t, rb.Contains(2*K+1))
	it := rb.Iterator()
	var first []uint64
	for len(first) < 6 {
		first = append(first, it.Next())
	}
	assert.Equal(t, []uint64{K, K + 1, K + 2, K + 3, K + 4, K + 6}, first)

	// End exactly at a bucket boundary that holds values: they must stay.
	rb = flipBoth(t, BitmapOf(5, 2*K, 2*K+1), 2*K-3, 2*K)
	requireValidWithValues(t, rb, 5, 2*K-3, 2*K-2, 2*K-1, 2*K, 2*K+1)
	rb = flipBoth(t, BitmapOf(2*K-1, 2*K), 2*K-1, 2*K)
	requireValidWithValues(t, rb, 2*K)
}

// F3
func TestEdge64FlipAcrossSignBit(t *testing.T) {
	H := edgeH
	rb := flipBoth(t, BitmapOf(0), H-1, H+1)
	requireValidWithValues(t, rb, 0, H-1, H)
	v, err := rb.Select(1)
	require.NoError(t, err)
	assert.Equal(t, H-1, v)
	v, err = rb.Select(2)
	require.NoError(t, err)
	assert.Equal(t, H, v)
}

// F4
func TestEdge64FlipNearMaxUint64(t *testing.T) {
	M := edgeM
	rb := flipBoth(t, NewBitmap(), M-9, M)
	assert.Equal(t, uint64(9), rb.GetCardinality())
	assert.Equal(t, M-1, rb.Maximum())
	assert.False(t, rb.Contains(M))
	rb = flipBoth(t, rb, M-9, M)
	requireValidWithValues(t, rb)

	rb = flipBoth(t, BitmapOf(M), M-9, M)
	assert.Equal(t, uint64(10), rb.GetCardinality())
	assert.True(t, rb.Contains(M))

	rb = flipBoth(t, BitmapOf(M-1), M-1, M)
	requireValidWithValues(t, rb)
}

// F5
func TestEdge64XorAndFlipEmptyFirstOrLastContainer(t *testing.T) {
	for _, second := range []uint64{65536, edgeK, edgeM} {
		base := BitmapOf(0, second)

		x := Xor(base, BitmapOf(0))
		requireValidWithValues(t, x, second)
		assert.Equal(t, second, x.Minimum())
		assert.Equal(t, second, x.Maximum())
		x = base.Clone()
		x.Xor(BitmapOf(0))
		requireValidWithValues(t, x, second)

		x = Xor(base, BitmapOf(second))
		requireValidWithValues(t, x, 0)
		assert.Equal(t, uint64(0), x.Maximum())
		x = base.Clone()
		x.Xor(BitmapOf(second))
		requireValidWithValues(t, x, 0)

		x = flipBoth(t, base.Clone(), 0, 1)
		requireValidWithValues(t, x, second)
		assert.Equal(t, second, x.Minimum())
		v, err := x.Select(0)
		require.NoError(t, err)
		assert.Equal(t, second, v)

		x = flipBoth(t, base.Clone(), second, second+1)
		if second == edgeM {
			// [M, M+1) cannot be expressed: the end wraps to 0.
			requireValidWithValues(t, x, 0, second)
		} else {
			requireValidWithValues(t, x, 0)
		}
	}
}

// V1
func TestEdge64BoundaryValues(t *testing.T) {
	values := []uint64{0, 65535, 65536, edgeK - 1, edgeK, 1<<48 - 1, 1 << 48, edgeH - 1, edgeH, edgeM - 1, edgeM}
	shuffled := slices.Clone(values)
	rand.New(rand.NewSource(7)).Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })
	rb := NewBitmap()
	for _, v := range shuffled {
		assert.True(t, rb.CheckedAdd(v))
	}
	require.NoError(t, rb.Validate())
	assert.Equal(t, uint64(len(values)), rb.GetCardinality())
	assert.Equal(t, values, rb.ToArray())
	for _, v := range values {
		assert.True(t, rb.Contains(v))
		for _, n := range []uint64{v - 1, v + 1} {
			if !slices.Contains(values, n) {
				assert.False(t, rb.Contains(n), "%d", n)
			}
		}
	}
	for i, v := range shuffled {
		assert.True(t, rb.CheckedRemove(v))
		assert.False(t, rb.CheckedRemove(v))
		require.NoError(t, rb.Validate())
		assert.Equal(t, uint64(len(values)-i-1), rb.GetCardinality())
	}
	requireValidWithValues(t, rb)
}

// V2
func TestEdge64RankSelectMinMaxAtExtremes(t *testing.T) {
	H, M := edgeH, edgeM
	rb := BitmapOf(0, 1, H-1, H)
	for i, want := range []uint64{0, 1, H - 1, H} {
		v, err := rb.Select(uint64(i))
		require.NoError(t, err)
		assert.Equal(t, want, v)
	}
	for _, c := range []struct{ x, rank uint64 }{
		{0, 1}, {1, 2}, {2, 2}, {H - 2, 2}, {H - 1, 3}, {H, 4}, {M, 4},
	} {
		assert.Equal(t, c.rank, rb.Rank(c.x), "rank(%d)", c.x)
	}
	assert.Equal(t, uint64(0), rb.Minimum())
	assert.Equal(t, H, rb.Maximum())

	rb = BitmapOf(M)
	assert.Equal(t, uint64(0), rb.Rank(M-1))
	assert.Equal(t, uint64(1), rb.Rank(M))
	v, err := rb.Select(0)
	require.NoError(t, err)
	assert.Equal(t, M, v)
	assert.Equal(t, M, rb.Minimum())
	assert.Equal(t, M, rb.Maximum())

	rb = BitmapOf(1, M)
	v, err = rb.Select(1)
	require.NoError(t, err)
	assert.Equal(t, M, v)
	assert.Equal(t, uint64(2), rb.Rank(M))
}

// V3
func TestEdge64SelectOutOfRangeAndRankOnEmpty(t *testing.T) {
	empty := NewBitmap()
	for _, p := range []uint64{0, 1, edgeH - 1, edgeH, edgeM} {
		assert.Equal(t, uint64(0), empty.Rank(p))
	}
	_, err := empty.Select(0)
	assert.Error(t, err)

	rb := BitmapOf(123)
	_, err = rb.Select(1)
	assert.Error(t, err)
	_, err = rb.Select(2)
	assert.Error(t, err)
	_, err = BitmapOf(1, edgeM).Select(edgeM)
	assert.Error(t, err)
}

// V5
func TestEdge64RemoveLastValueOfHighBucketThenSelect(t *testing.T) {
	H := edgeH
	rb := BitmapOf(123, H-1)
	rb.Remove(H - 1)
	require.NoError(t, rb.Validate())
	assert.Equal(t, uint64(1), rb.GetCardinality())
	v, err := rb.Select(0)
	require.NoError(t, err)
	assert.Equal(t, uint64(123), v)
	_, err = rb.Select(1)
	assert.Error(t, err)
	rb.Add(H - 1)
	v, err = rb.Select(1)
	require.NoError(t, err)
	assert.Equal(t, H-1, v)

	rb.Remove(H - 1)
	rb.Add((H - 1) / 2)
	v, err = rb.Select(1)
	require.NoError(t, err)
	assert.Equal(t, (H-1)/2, v)

	rb = BitmapOf(123, (H-1)/2)
	rb.Remove((H - 1) / 2)
	rb.Add(H - 1)
	v, err = rb.Select(1)
	require.NoError(t, err)
	assert.Equal(t, H-1, v)
	assert.Equal(t, uint64(1), rb.Rank((H-1)/2))
	assert.Equal(t, uint64(2), rb.Rank(H-1))
}

// O1
func TestEdge64SelfAliasingSetOps(t *testing.T) {
	makeB := func() *Bitmap {
		b := NewBitmap()
		b.AddRange(1, 11)
		b.RunOptimize()
		b.Add(edgeK + 3)
		b.Add(edgeM)
		return b
	}
	expected := makeB().ToArray()
	card := uint64(len(expected))

	b := makeB()
	assert.Equal(t, card, b.AndCardinality(b))
	assert.Equal(t, card, b.OrCardinality(b))
	assert.True(t, b.Intersects(b))
	requireValidWithValues(t, Or(b, b), expected...)
	requireValidWithValues(t, And(b, b), expected...)
	requireValidWithValues(t, Xor(b, b))
	requireValidWithValues(t, AndNot(b, b))
	requireValidWithValues(t, b, expected...)

	b.Or(b)
	requireValidWithValues(t, b, expected...)
	b.And(b)
	requireValidWithValues(t, b, expected...)
	b.Xor(b)
	requireValidWithValues(t, b)
	b = makeB()
	b.AndNot(b)
	requireValidWithValues(t, b)
}

// O2: regression vectors from the Java ART implementation (TestRoaring64Bitmap).
func TestEdge64JavaARTRegressions(t *testing.T) {
	t.Run("issue428", func(t *testing.T) {
		input := BitmapOf(1353768194141061120)
		compare := []uint64{5192650370358181888, 5193776270265024512, 5194532734264934400,
			5194544828892839936, 5194545653526560768, 5194545688960040960, 5194545692181266432,
			5194545705066168320, 5194545722246037504, 5194545928404467712, 5194550326450978816,
			5194620695195156480, 5206161169240293376}
		a := BitmapOf(compare...)
		a.And(input)
		requireValidWithValues(t, a)
		o := BitmapOf(compare...)
		o.Or(input)
		assert.Equal(t, uint64(14), o.GetCardinality())
		n := BitmapOf(compare...)
		n.AndNot(input)
		assert.Equal(t, uint64(13), n.GetCardinality())
	})
	t.Run("crossers", func(t *testing.T) {
		input := BitmapOf(5183829215128059904)
		for _, crossers := range [][]uint64{
			{4413527634823086080, 4418031234450456576, 4421408934170984448, 4421690409147695104,
				4421479302915162112, 4421426526357028864, 4421413332217495552, 4421416630752378880,
				4421416905630285824, 4421417111788716032, 4421417128968585216, 4421417133263552512,
				4421417134337294336},
			{4413527634823086080, 4418031234450456576, 4421408934170984448, 4421127459194273792,
				4420916352961740800, 4420863576403607552, 4420850382264074240, 4420847083729190912,
				4420847358607097856, 4420847564765528064, 4420847616305135616, 4420847620600102912,
				4420847623821328384},
			{4421416447812311717, 4420658333523655893, 4420658332008999025},
		} {
			c := BitmapOf(crossers...)
			c.And(input)
			requireValidWithValues(t, c)
		}
	})
	t.Run("issue537", func(t *testing.T) {
		a := BitmapOf(275846320)
		b := BitmapOf(275846320)
		c := BitmapOf(275845652, 275845746, 275846148, 275847372, 275847380, 275847388, 275847459, 275847528, 275847586, 275847588, 275847600, 275847607, 275847610, 275847613, 275847631, 275847664, 275847672, 275847677, 275847680, 275847742, 275847808, 275847811, 275847824, 275847830, 275847856, 275847861, 275847863, 275847872, 275847896, 275847923, 275847924, 275847975, 275847990, 275847995, 275848003, 275848080, 275848081, 275848084, 275848095, 275848100, 275848120, 275848129, 275848134, 275848163, 275848174, 275848206, 275848218, 275848231, 275848272, 275848281, 275848308, 275848344, 275848376, 275848382, 275848395, 275848400, 275848411, 275848426, 275848445, 275848449, 275848451, 275848454, 275848469)
		c.And(b)
		assert.False(t, c.Contains(275846320))
		c.And(a)
		assert.False(t, c.Contains(275846320))
		require.NoError(t, c.Validate())
	})
	t.Run("issue580", func(t *testing.T) {
		rb := BitmapOf(3242766498713841665, 3492544636360507394, 3418218112527884289,
			3220956490660966402, 3495344165583036418, 3495023214002368514, 3485108231289675778)
		assert.Len(t, iterate64(rb.Iterator()), 7)
	})
	t.Run("issue619", func(t *testing.T) {
		const cleanerValue, addressValue = 140664568792144, 140662937752432
		addressSpace, cleaner := NewBitmap(), BitmapOf(cleanerValue)
		for iteration := 0; ; iteration++ {
			addressSpace.Add(addressValue)
			addressSpace.Add(cleanerValue)
			if iteration == 33 {
				break
			}
			addressSpace.AndNot(cleaner)
		}
		requireValidWithValues(t, addressSpace, addressValue, cleanerValue)
	})
	t.Run("toArrayAfterAndHasEmptyContainer", func(t *testing.T) {
		a, b := BitmapOf(0), BitmapOf(1)
		requireValidWithValues(t, And(a, b))
		a.And(b)
		requireValidWithValues(t, a)
	})
	t.Run("andDisjoint", func(t *testing.T) {
		leftData := []uint64{1076595327100, 1074755534972, 5060192403580, 5060308664444}
		rightData := []uint64{3470563844}
		requireValidWithValues(t, And(BitmapOf(leftData...), BitmapOf(rightData...)))
		requireValidWithValues(t, And(BitmapOf(rightData...), BitmapOf(leftData...)))
		left := BitmapOf(leftData...)
		left.And(BitmapOf(rightData...))
		requireValidWithValues(t, left)
		_, err := left.Select(0)
		assert.Error(t, err)
		swapLeft := BitmapOf(rightData...)
		swapLeft.And(BitmapOf(leftData...))
		requireValidWithValues(t, swapLeft)
	})
}

// O3
func TestEdge64SetOpsAcrossHighKeys(t *testing.T) {
	K1, K2, M := uint64(1)<<16, edgeK, edgeM
	aVals := []uint64{1, 2, 7, K1 + 1, K1 + 5, K2 + 1, M}
	bVals := []uint64{3, 4, 7, K1 + 5, K1 + 9, K2 + 2, M}
	for _, c := range []struct {
		name     string
		pkg      func(x, y *Bitmap) *Bitmap
		inplace  func(x, y *Bitmap)
		expected []uint64
	}{
		{"and", And, (*Bitmap).And, []uint64{7, K1 + 5, M}},
		{"or", Or, (*Bitmap).Or, []uint64{1, 2, 3, 4, 7, K1 + 1, K1 + 5, K1 + 9, K2 + 1, K2 + 2, M}},
		{"xor", Xor, (*Bitmap).Xor, []uint64{1, 2, 3, 4, K1 + 1, K1 + 9, K2 + 1, K2 + 2}},
		{"andnot", AndNot, (*Bitmap).AndNot, []uint64{1, 2, K1 + 1, K2 + 1}},
	} {
		t.Run(c.name, func(t *testing.T) {
			a, b := BitmapOf(aVals...), BitmapOf(bVals...)
			requireValidWithValues(t, c.pkg(a, b), c.expected...)
			requireValidWithValues(t, a, aVals...)
			requireValidWithValues(t, b, bVals...)
			c.inplace(a, b)
			requireValidWithValues(t, a, c.expected...)
			requireValidWithValues(t, b, bVals...)
		})
	}
}

// O4
func TestEdge64CardinalitiesAndIntersects(t *testing.T) {
	K := edgeK
	r1, r2 := BitmapOf(1, 2*K+5), BitmapOf(7*K+1, 9*K+5)
	assert.Equal(t, uint64(0), r1.AndCardinality(r2))
	assert.False(t, r1.Intersects(r2))
	assert.Equal(t, uint64(4), r1.OrCardinality(r2))

	r1, r2 = NewBitmap(), NewBitmap()
	r1.AddRange(0, 100)
	r1.AddRange(2*K, 2*K+100)
	r2.AddRange(50, 150)
	r2.AddRange(8*K+50, 8*K+150)
	assert.Equal(t, uint64(350), r1.OrCardinality(r2))
	assert.Equal(t, Or(r1, r2).GetCardinality(), r1.OrCardinality(r2))
	assert.Equal(t, uint64(50), r1.AndCardinality(r2))
	assert.Equal(t, And(r1, r2).GetCardinality(), r1.AndCardinality(r2))
	assert.True(t, r1.Intersects(r2))
	r1.Add(9*K + 1)
	assert.Equal(t, uint64(151), AndNot(r1, r2).GetCardinality())
	assert.Equal(t, uint64(150), AndNot(r2, r1).GetCardinality())
	assert.Equal(t, uint64(301), Xor(r1, r2).GetCardinality())

	empty := NewBitmap()
	card := r1.GetCardinality()
	assert.Equal(t, uint64(0), r1.AndCardinality(empty))
	assert.Equal(t, uint64(0), empty.AndCardinality(r1))
	assert.Equal(t, card, r1.OrCardinality(empty))
	assert.Equal(t, card, empty.OrCardinality(r1))
	assert.Equal(t, card, Xor(r1, empty).GetCardinality())
	assert.Equal(t, card, AndNot(r1, empty).GetCardinality())
	assert.Equal(t, uint64(0), AndNot(empty, r1).GetCardinality())
	assert.False(t, r1.Intersects(empty))
	assert.False(t, empty.Intersects(r1))
	assert.False(t, empty.Intersects(NewBitmap()))
}

func iterate64(it IntIterable64) []uint64 {
	var out []uint64
	for it.HasNext() {
		out = append(out, it.Next())
	}
	return out
}

// I1
func TestEdge64AdvanceIntoGapsAcrossHighKeys(t *testing.T) {
	const mi4 = uint64(4) << 20
	p1, p2, p3 := uint64(2e9), uint64(4e9), uint64(6e9)
	rb := NewBitmap()
	rb.AddRange(p1, p1+mi4)
	rb.AddRange(p2, p2+mi4)
	rb.AddRange(p3, p3+mi4)

	it := rb.Iterator()
	assert.Equal(t, p1, it.PeekNext())
	it.AdvanceIfNeeded(p1 + mi4/2)
	assert.Equal(t, p1+mi4/2, it.PeekNext())
	it.AdvanceIfNeeded(p1 + mi4 + 5)
	assert.Equal(t, p2, it.PeekNext())
	it.AdvanceIfNeeded(p2 + mi4/2 + mi4)
	assert.Equal(t, p3, it.PeekNext())

	it = rb.Iterator()
	it.AdvanceIfNeeded(p2)
	assert.Equal(t, p2, it.PeekNext())
	it.AdvanceIfNeeded(p2 + mi4/2 + 5*mi4)
	assert.Equal(t, p3, it.PeekNext())
	it.AdvanceIfNeeded(p3)
	assert.Equal(t, p3, it.PeekNext())
	assert.Equal(t, p3, it.Next())

	// target in a high key with no bucket at all
	rb = BitmapOf(5, 3*edgeK+7)
	it = rb.Iterator()
	it.AdvanceIfNeeded(edgeK + 1)
	assert.True(t, it.HasNext())
	assert.Equal(t, 3*edgeK+7, it.PeekNext())
	it.AdvanceIfNeeded(3*edgeK + 8)
	assert.False(t, it.HasNext())
}

// I2
func TestEdge64AdvanceAcrossSignBit(t *testing.T) {
	H := edgeH
	rb := NewBitmap()
	rb.AddRange(H-1, H+3)
	it := rb.Iterator()
	it.AdvanceIfNeeded(H - 1)
	assert.Equal(t, H-1, it.PeekNext())
	it.AdvanceIfNeeded(H + 1)
	assert.Equal(t, H+1, it.PeekNext())
	assert.Equal(t, []uint64{H + 1, H + 2}, iterate64(it))
}

// I3
func TestEdge64AdvanceToMaxUint64(t *testing.T) {
	M := edgeM
	it := BitmapOf(5, M-1, M).Iterator()
	it.AdvanceIfNeeded(M)
	require.True(t, it.HasNext())
	assert.Equal(t, M, it.PeekNext())
	assert.Equal(t, M, it.Next())
	assert.False(t, it.HasNext())

	it = BitmapOf(5, M-1).Iterator()
	it.AdvanceIfNeeded(M)
	assert.False(t, it.HasNext())
	it.AdvanceIfNeeded(M)
	assert.False(t, it.HasNext())
}

// I4
func TestEdge64UnsignedIterationOrder(t *testing.T) {
	H, M := edgeH, edgeM
	rb := BitmapOf(H, 0, M, 1, H-1)
	forward := []uint64{0, 1, H - 1, H, M}
	reverse := slices.Clone(forward)
	slices.Reverse(reverse)
	assert.Equal(t, forward, rb.ToArray())
	assert.Equal(t, forward, iterate64(rb.Iterator()))
	assert.Equal(t, reverse, iterate64(rb.ReverseIterator()))
	assert.Equal(t, forward, slices.Collect(Values(rb)))
	assert.Equal(t, reverse, slices.Collect(Backward(rb)))
	buf := make([]uint64, 2)
	var many []uint64
	mit := rb.ManyIterator()
	for n := mit.NextMany(buf); n > 0; n = mit.NextMany(buf) {
		many = append(many, buf[:n]...)
	}
	assert.Equal(t, forward, many)
}

// I5
func TestEdge64EmptyIterators(t *testing.T) {
	rb := NewBitmap()
	it := rb.Iterator()
	assert.False(t, it.HasNext())
	assert.NotPanics(t, func() {
		it.AdvanceIfNeeded(0)
		it.AdvanceIfNeeded(edgeM)
	})
	assert.False(t, it.HasNext())
	assert.False(t, rb.ReverseIterator().HasNext())
	assert.Equal(t, 0, rb.ManyIterator().NextMany(make([]uint64, 8)))
	assert.Empty(t, slices.Collect(Values(rb)))
	assert.Empty(t, slices.Collect(Backward(rb)))
}

// I6
func TestEdge64NextManyAcrossHighKeyBoundary(t *testing.T) {
	K := edgeK
	rb := NewBitmap()
	rb.AddRange(K-5, K+5)
	expected := []uint64{K - 5, K - 4, K - 3, K - 2, K - 1, K, K + 1, K + 2, K + 3, K + 4}

	buf := make([]uint64, 11)
	it := rb.ManyIterator()
	n := it.NextMany(buf)
	assert.Equal(t, 10, n)
	assert.Equal(t, expected, buf[:n])
	assert.Equal(t, 0, it.NextMany(buf))

	it = rb.ManyIterator()
	one := make([]uint64, 1)
	for _, e := range expected {
		require.Equal(t, 1, it.NextMany(one))
		assert.Equal(t, e, one[0])
	}
	assert.Equal(t, 0, it.NextMany(one))
	assert.Equal(t, 0, it.NextMany(one))
}

// B1
func TestEdge64AddManyUnsortedWithDuplicates(t *testing.T) {
	M := edgeM
	rb := NewBitmap()
	rb.AddMany([]uint64{1, 1 << 33, 1 << 33, 7, M, 0, 1 << 60, 7})
	requireValidWithValues(t, rb, 0, 1, 7, 1<<33, 1<<60, M)
	requireValidWithValues(t, BitmapOf(M, 1<<60, 1<<33, 7, 1, 0), 0, 1, 7, 1<<33, 1<<60, M)

	// Descending across buckets (Java issue 577).
	rb = NewBitmap()
	for v := uint64(10); v > 0; v-- {
		rb.Add(v * edgeK)
	}
	require.NoError(t, rb.Validate())
	assert.Equal(t, 10*edgeK, rb.ReverseIterator().Next())
	assert.Equal(t, 10*edgeK, rb.Maximum())
}

// B2: random range operations near the boundaries, mirrored in a map.
func TestEdge64RandomRangeOpsVsSet(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	bases := []uint64{0, 1<<16 - 512, edgeK - 512, 1<<48 - 512, edgeM - 1024}
	rb := NewBitmap()
	model := map[uint64]struct{}{}
	for op := 0; op < 1000; op++ {
		lo := bases[r.Intn(len(bases))] + uint64(r.Intn(1024))
		width := uint64(r.Intn(300))
		if r.Intn(10) == 0 {
			width = uint64(r.Intn(3000))
		}
		hi := lo + width
		if hi < lo || hi > edgeM {
			hi = edgeM
		}
		switch r.Intn(5) {
		case 0:
			rb.AddRange(lo, hi)
			for v := lo; v < hi; v++ {
				model[v] = struct{}{}
			}
		case 1:
			rb.RemoveRange(lo, hi)
			for v := lo; v < hi; v++ {
				delete(model, v)
			}
		case 2:
			rb.Flip(lo, hi)
			for v := lo; v < hi; v++ {
				if _, ok := model[v]; ok {
					delete(model, v)
				} else {
					model[v] = struct{}{}
				}
			}
		case 3:
			_, had := model[lo]
			require.Equal(t, !had, rb.CheckedAdd(lo))
			model[lo] = struct{}{}
		case 4:
			_, had := model[lo]
			require.Equal(t, had, rb.CheckedRemove(lo))
			delete(model, lo)
		}
		if r.Intn(10) == 0 {
			rb.RunOptimize()
		}
		require.Equal(t, uint64(len(model)), rb.GetCardinality(), "op %d", op)
		if op%16 == 0 {
			require.NoError(t, rb.Validate())
			expected := make([]uint64, 0, len(model))
			for v := range model {
				expected = append(expected, v)
			}
			slices.Sort(expected)
			require.Equal(t, expected, rb.ToArray(), "op %d", op)
		}
	}
}
