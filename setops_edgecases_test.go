package roaring

// Edge-case tests for set operations between bitmaps, ported from the Java
// (RoaringBitmap) and C (CRoaring) test suites so that the three libraries
// cover the same scenarios.

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Java: TestFastAggregation.horizontal_xorDropsBlocksThatCancelOut.
func TestHeapXorDropsContainersThatCancelOut(t *testing.T) {
	a := BitmapOf(1, 3, 5, 70000)
	b := BitmapOf(1, 3, 5)
	x := HeapXor(a, b)
	assert.True(t, x.Equals(BitmapOf(70000)))
	require.NoError(t, x.Validate())
	assert.Equal(t, uint64(1), x.Stats().Containers)

	x = HeapXor(a, a.Clone())
	assert.True(t, x.IsEmpty())
	require.NoError(t, x.Validate())
	assert.Equal(t, uint64(0), x.Stats().Containers)

	c := a.Clone()
	c.Xor(b)
	assert.Equal(t, uint64(1), c.Stats().Containers)
	assert.Equal(t, uint64(1), Xor(a, b).Stats().Containers)
}

// Java: TestRoaringBitmap.testLazyOr, testFirstLast_AfterLazyMutation1/2.
func TestLazyOrThenRepair(t *testing.T) {
	expected := BitmapOf(1, 2, 3, 4, 5, 7, 8, 9, 65536, 131072, 262144, 524288, 1048576)
	rb1 := BitmapOf(65536, 262144, 524288)
	rb1.lazyOR(BitmapOf(4, 7, 8, 9))
	rb1.lazyOR(BitmapOf(1, 2, 3, 4, 5, 65536, 131072, 1048576))
	rb1.repairAfterLazy()
	assert.True(t, rb1.Equals(expected))
	require.NoError(t, rb1.Validate())

	x := lazyOR(BitmapOf(65536, 262144, 524288), BitmapOf(4, 7, 8, 9))
	x = lazyOR(x, BitmapOf(1, 2, 3, 4, 5, 65536, 131072, 1048576))
	x.repairAfterLazy()
	assert.True(t, x.Equals(expected))
	require.NoError(t, x.Validate())

	// dense case: a bitmap container, lazily or-ed with a few values
	dense := New()
	for i := uint32(3); i <= 65536; i += 3 {
		dense.Add(i)
	}
	small := BitmapOf(0, 2, 4, 6, 8)
	eager := Or(dense, small)
	lazy := dense.Clone()
	lazy.lazyOR(small)
	lazy.repairAfterLazy()
	assert.Equal(t, uint32(0), lazy.Minimum())
	assert.Equal(t, uint32(65535), lazy.Maximum())
	assert.Equal(t, eager.GetCardinality(), lazy.GetCardinality())
	assert.True(t, lazy.Equals(eager))
	require.NoError(t, lazy.Validate())

	lazy = lazyOR(dense, small)
	lazy.repairAfterLazy()
	assert.True(t, lazy.Equals(eager))
	require.NoError(t, lazy.Validate())
}

// C: mixed_container_unit.c run_array_andnot_bug_test.
func TestRunAndNotArrayRemovesValueInsideRun(t *testing.T) {
	runContents := []uint32{
		196608, 196611, 196612, 196613, 196616, 196619, 196621, 196623, 196628,
		196629, 196630, 196631, 196632, 196633, 196634, 196635, 196636, 196638,
		196639, 196640, 196641, 196642, 196644, 196645, 196646, 196647, 196648,
		196649, 196650, 196652, 196653, 196654, 196656, 196658, 196659, 196660,
		196662, 196663, 196664, 196665, 196666, 196667, 196669, 196670, 196671,
		196672, 196673, 196674, 196675, 196677, 196678, 196679, 196680, 196682,
		196684, 196685, 196686, 196688, 196689, 196690, 196691, 196692, 196693,
		196694, 196695, 196697, 196698, 196699, 196700, 196701, 196702, 196703,
		196704, 196705, 196706, 196707, 196708, 196709, 196710, 196711, 196712,
		196713, 196714, 196715, 196717, 196719, 196720, 196722, 196723, 196725,
		196726, 196727, 196728, 196729,
	}
	arrayContents := []uint32{196722, 196824, 196989}
	newRun := func() *runContainer16 {
		rc := newRunContainer16()
		for _, v := range runContents {
			rc.iadd(uint16(v % 65536))
		}
		return rc
	}
	ac := newArrayContainer()
	for _, v := range arrayContents {
		ac.iadd(uint16(v % 65536))
	}
	expected := newArrayContainer()
	for _, v := range runContents {
		if !ac.contains(uint16(v % 65536)) {
			expected.iadd(uint16(v % 65536))
		}
	}
	for name, result := range map[string]container{
		"andNot":      newRun().andNot(ac),
		"andNotArray": newRun().andNotArray(ac),
		"iandNot":     newRun().iandNot(ac),
	} {
		assert.False(t, result.contains(196722%65536), name)
		assert.True(t, result.equals(expected), name)
		assert.Equal(t, len(runContents)-1, result.getCardinality(), name)
		assert.NoError(t, result.validate(), name)
	}
	// bitmap level, with a run-optimized left operand
	left := BitmapOf(runContents...)
	left.RunOptimize()
	diff := AndNot(left, BitmapOf(arrayContents...))
	assert.False(t, diff.Contains(196722))
	assert.Equal(t, uint64(len(runContents)-1), diff.GetCardinality())
	require.NoError(t, diff.Validate())
}

func setopsArray() *Bitmap {
	return BitmapOf(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
}

func setopsBitset() *Bitmap {
	rb := New()
	for i := uint32(0); i < 20000; i += 3 {
		rb.Add(i)
	}
	return rb
}

func setopsRun() *Bitmap {
	rb := New()
	rb.AddRange(10, 5000)
	rb.RunOptimize()
	return rb
}

// Java: TestRoaringBitmap.testWithYourself.
func TestSetOpsWithItself(t *testing.T) {
	for name, gen := range map[string]func() *Bitmap{"array": setopsArray, "bitset": setopsBitset, "run": setopsRun} {
		ref := gen()
		card := ref.GetCardinality()

		x := gen()
		x.Or(x)
		assert.True(t, x.Equals(ref), name)
		assert.NoError(t, x.Validate(), name)
		x.And(x)
		assert.True(t, x.Equals(ref), name)
		assert.NoError(t, x.Validate(), name)
		x.Xor(x)
		assert.True(t, x.IsEmpty(), name)
		assert.NoError(t, x.Validate(), name)
		x = gen()
		x.AndNot(x)
		assert.True(t, x.IsEmpty(), name)
		assert.NoError(t, x.Validate(), name)

		x = gen()
		assert.True(t, And(x, x).Equals(ref), name)
		assert.True(t, Or(x, x).Equals(ref), name)
		assert.True(t, Xor(x, x).IsEmpty(), name)
		assert.True(t, AndNot(x, x).IsEmpty(), name)
		assert.Equal(t, card, x.AndCardinality(x), name)
		assert.Equal(t, card, x.OrCardinality(x), name)
		assert.True(t, x.Intersects(x), name)
		assert.True(t, x.Equals(ref), name)
	}
	e := New()
	assert.False(t, e.Intersects(e))
	e.Or(e)
	assert.True(t, e.IsEmpty())
}

// C: container_comparison_unit.c generic_equal_test.
func TestCrossTypeContainerEquality(t *testing.T) {
	makers := map[string]func() container{
		"array":  func() container { return newArrayContainer() },
		"bitmap": func() container { return newBitmapContainer() },
		"run":    func() container { return newRunContainer16() },
	}
	fill := func(c container, from, to int) container {
		for i := from; i < to; i++ {
			c.iadd(uint16(i))
		}
		return c
	}
	for n1, m1 := range makers {
		for n2, m2 := range makers {
			name := n1 + "/" + n2
			c1, c2 := m1(), m2()
			assert.True(t, c1.equals(c2), name)
			for i := 0; i < 100; i++ {
				c1.iadd(uint16(i * 10))
				c2.iadd(uint16(i * 10))
			}
			assert.True(t, c1.equals(c2), name)
			assert.True(t, c2.equals(c1), name)
			c1.iadd(273)
			assert.False(t, c1.equals(c2), name)
			assert.False(t, c2.equals(c1), name)
			c2.iadd(854)
			assert.False(t, c1.equals(c2), name)
			assert.False(t, c2.equals(c1), name)
			c1.iadd(854)
			c2.iadd(273)
			assert.True(t, c1.equals(c2), name)
			assert.True(t, c2.equals(c1), name)

			// full containers
			f1, f2 := fill(m1(), 0, 1<<16), fill(m2(), 0, 1<<16)
			assert.True(t, f1.equals(f2), name)

			// first value differs
			d1 := fill(m1(), 0, 1<<16)
			d2 := fill(m2(), 0, 1<<16)
			d1.iremove(0)
			d2.iremove(1)
			assert.False(t, d1.equals(d2), name)
			assert.False(t, d2.equals(d1), name)
			// last value differs
			l1 := fill(m1(), 0, 1<<16)
			l2 := fill(m2(), 0, 1<<16)
			l1.iremove(65534)
			l2.iremove(65535)
			assert.False(t, l1.equals(l2), name)
			assert.False(t, l2.equals(l1), name)
		}
	}
}

// C: test_intersect_small_run_bitset.
func TestIntersectSmallRunWithBitset(t *testing.T) {
	rb1 := New()
	rb1.AddRange(0, 1)
	rb1.RunOptimize()
	rb2 := New()
	for i := uint32(1); i < 8194; i += 2 {
		rb2.Add(i)
	}
	require.Equal(t, uint64(1), rb2.Stats().BitmapContainers)
	assert.False(t, rb1.Intersects(rb2))
	assert.False(t, rb2.Intersects(rb1))
	assert.Equal(t, uint64(0), rb1.AndCardinality(rb2))
	assert.Equal(t, uint64(0), rb2.AndCardinality(rb1))
	assert.True(t, And(rb1, rb2).IsEmpty())
}

// C: test_full_bitset_union_becomes_run.
func TestFullBitsetUnionBecomesRun(t *testing.T) {
	evens, odds := New(), New()
	for i := uint32(0); i < 1<<16; i += 2 {
		evens.Add(i)
		odds.Add(i + 1)
	}
	require.Equal(t, uint64(1), evens.Stats().BitmapContainers)
	check := func(name string, rb *Bitmap) {
		st := rb.Stats()
		assert.Equal(t, uint64(65536), rb.GetCardinality(), name)
		assert.Equal(t, uint64(1), st.Containers, name)
		assert.Equal(t, uint64(1), st.RunContainers, name)
		assert.Equal(t, uint64(0), st.BitmapContainers, name)
		assert.NoError(t, rb.Validate(), name)
	}
	inplace := evens.Clone()
	inplace.Or(odds)
	check("Or", inplace)
	check("package Or", Or(evens, odds))
	check("FastOr", FastOr(evens, odds, odds))
	check("HeapOr", HeapOr(evens, odds, odds))
	check("ParOr", ParOr(0, evens, odds, odds))
	check("ParHeapOr", ParHeapOr(0, evens, odds, odds))
}

// C: andnot_run_containers_tail.
func TestAndNotRunContainersTail(t *testing.T) {
	r1, r2 := New(), New()
	for x := uint64(0); x < 4000; x += 20 {
		r1.AddRange(x, x+5)
		if x < 1000 {
			r2.AddRange(x+1, x+4)
		}
	}
	r1.RunOptimize()
	r2.RunOptimize()
	expected := New()
	for v := uint32(0); v < 4000; v++ {
		if r1.Contains(v) && !r2.Contains(v) {
			expected.Add(v)
		}
	}
	assert.True(t, AndNot(r1, r2).Equals(expected))
	inplace := r1.Clone()
	inplace.AndNot(r2)
	assert.True(t, inplace.Equals(expected))
	require.NoError(t, inplace.Validate())
}

// C: inplaceorwide.
func TestInPlaceOrIntoAlmostWholeUniverse(t *testing.T) {
	r1 := BitmapOf(0)
	r2 := New()
	r2.AddRange(0, 4294901761)
	r1.Or(r2)
	assert.Equal(t, uint64(4294901761), r1.GetCardinality())
	require.NoError(t, r1.Validate())
}

// C: leaks_with_empty, cpp_unit.cpp empty-operand cases.
func TestSetOpsWithEmptyOperand(t *testing.T) {
	for _, cow := range []bool{false, true} {
		r := New()
		for v := uint32(100); v < 70000; v += 3 {
			r.Add(v)
		}
		r.SetCopyOnWrite(cow)
		ref := r.Clone()
		e := New()
		e.SetCopyOnWrite(cow)
		card := r.GetCardinality()

		assert.True(t, Or(r, e).Equals(ref))
		assert.True(t, Xor(r, e).Equals(ref))
		assert.True(t, AndNot(r, e).Equals(ref))
		assert.True(t, Or(e, r).Equals(ref))
		assert.True(t, Xor(e, r).Equals(ref))
		assert.True(t, And(r, e).IsEmpty())
		assert.True(t, And(e, r).IsEmpty())
		assert.True(t, AndNot(e, r).IsEmpty())
		assert.Equal(t, uint64(0), r.AndCardinality(e))
		assert.Equal(t, uint64(0), e.AndCardinality(r))
		assert.Equal(t, card, r.OrCardinality(e))
		assert.Equal(t, card, e.OrCardinality(r))
		assert.False(t, r.Intersects(e))
		assert.False(t, e.Intersects(r))

		// in place on the non-empty receiver
		x := r.Clone()
		x.Or(e)
		assert.True(t, x.Equals(ref))
		x.Xor(e)
		assert.True(t, x.Equals(ref))
		x.AndNot(e)
		assert.True(t, x.Equals(ref))
		x.And(e)
		assert.True(t, x.IsEmpty())

		// in place on an empty receiver: the result must be independent of r
		for name, op := range map[string]func(*Bitmap){
			"Or":  func(b *Bitmap) { b.Or(r) },
			"Xor": func(b *Bitmap) { b.Xor(r) },
		} {
			b := New()
			b.SetCopyOnWrite(cow)
			op(b)
			assert.True(t, b.Equals(ref), name)
			b.Add(101)
			b.Remove(100)
			b.AddRange(1000, 2000)
			assert.True(t, r.Equals(ref), name)
			assert.NoError(t, r.Validate(), name)
			assert.NoError(t, b.Validate(), name)
		}
		b := New()
		b.And(r)
		assert.True(t, b.IsEmpty())
		b.AndNot(r)
		assert.True(t, b.IsEmpty())
		assert.True(t, r.Equals(ref))
	}
}

// Java: ParallelAggregationTest wideOr, hugeOr1-3, disjointBigKeysOR,
// missingMiddleContainerXOR.
func TestWideAggregationsMixedContainers(t *testing.T) {
	bitmapAt := func(rb *Bitmap, key uint32, step uint32) {
		for i := uint32(0); i < 1<<16; i += step {
			rb.Add(key<<16 | i)
		}
	}
	arrayAt := func(rb *Bitmap, key uint32, off uint32) {
		for i := off; i < 1<<16; i += 997 {
			rb.Add(key<<16 | i)
		}
	}
	runAt := func(rb *Bitmap, key uint32, off uint64) {
		rb.AddRange(uint64(key)<<16+off, uint64(key)<<16+off+1000)
	}
	for _, n := range []int{20, 513} {
		inputs := make([]*Bitmap, n)
		for i := range inputs {
			rb := New()
			bitmapAt(rb, 0, uint32(2+i%5))
			arrayAt(rb, 1, uint32(i))
			runAt(rb, 2, uint64(i*7))
			rb.RunOptimize()
			inputs[i] = rb
		}
		expected := FastOr(inputs...)
		require.NoError(t, expected.Validate())
		naive := New()
		for _, in := range inputs {
			naive.Or(in)
		}
		assert.True(t, expected.Equals(naive), "n=%d", n)
		for _, p := range []int{0, 1, 4} {
			assert.True(t, ParOr(p, inputs...).Equals(expected), "ParOr n=%d p=%d", n, p)
		}
		assert.True(t, ParHeapOr(0, inputs...).Equals(expected), "ParHeapOr n=%d", n)
		assert.True(t, HeapOr(inputs...).Equals(expected), "HeapOr n=%d", n)
	}

	// disjoint big keys
	one, two, three := New(), New(), New()
	runAt(one, 0, 0)
	arrayAt(one, 2, 0)
	bitmapAt(one, 0x8001, 2)
	bitmapAt(two, 1, 3)
	runAt(two, 0x8002, 5)
	arrayAt(three, 3, 1)
	runAt(three, 0x8003, 9)
	for _, b := range []*Bitmap{one, two, three} {
		b.RunOptimize()
	}
	expected := FastOr(one, two, three)
	assert.Equal(t, one.GetCardinality()+two.GetCardinality()+three.GetCardinality(), expected.GetCardinality())
	for _, p := range []int{0, 1, 4} {
		assert.True(t, ParOr(p, one, two, three).Equals(expected))
	}
	assert.True(t, HeapOr(one, two, three).Equals(expected))

	// missing middle container xor
	one, two, three = New(), New(), New()
	runAt(one, 0, 0)
	bitmapAt(one, 1, 2)
	arrayAt(one, 2, 0)
	bitmapAt(two, 0, 3)
	arrayAt(two, 2, 1)
	arrayAt(three, 0, 2)
	bitmapAt(three, 1, 5)
	arrayAt(three, 2, 0)
	for _, b := range []*Bitmap{one, two, three} {
		b.RunOptimize()
	}
	chained := Xor(Xor(one, two), three)
	hx := HeapXor(one, two, three)
	assert.True(t, hx.Equals(chained))
	require.NoError(t, hx.Validate())
}

// Java: TestFastAggregation.testWorkShyAnd.
func TestFastAndMixedContainerLists(t *testing.T) {
	bitmapAt := func(rb *Bitmap, key uint32) {
		for i := uint32(0); i < 1<<16; i += 2 {
			rb.Add(key<<16 | i)
		}
	}
	arrayAt := func(rb *Bitmap, key uint32) {
		for i := uint32(0); i < 1<<16; i += 64 {
			rb.Add(key<<16 | i)
		}
	}
	runAt := func(rb *Bitmap, key uint32) {
		rb.AddRange(uint64(key)<<16, uint64(key)<<16+30000)
	}
	fillers := []func(*Bitmap, uint32){bitmapAt, arrayAt, runAt}
	keys := []uint32{0, 1, 2, 3, 4, 20}
	var lists [][]*Bitmap
	for l := 0; l < 10; l++ {
		list := make([]*Bitmap, 3)
		for j := range list {
			rb := New()
			for ki, key := range keys {
				if (l+j+ki)%4 != 3 { // leave some keys out
					fillers[(l+2*j+ki)%3](rb, key)
				}
			}
			rb.RunOptimize()
			list[j] = rb
		}
		lists = append(lists, list)
	}
	for li, list := range lists {
		for k := 0; k <= len(list); k++ {
			name := fmt.Sprintf("list %d prefix %d", li, k)
			prefix := list[:k]
			naive := New()
			if k > 0 {
				naive = prefix[0].Clone()
				for _, b := range prefix[1:] {
					naive.And(b)
				}
			}
			fast := FastAnd(prefix...)
			assert.True(t, fast.Equals(naive), name)
			assert.NoError(t, fast.Validate(), name)
			for _, p := range []int{1, 4} {
				assert.True(t, ParAnd(p, prefix...).Equals(naive), name)
			}
		}
	}
}
