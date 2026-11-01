package roaring

// Edge-case tests for single-bitmap operations, ported from the Java
// (RoaringBitmap) and C (CRoaring) test suites so that the three libraries
// cover the same scenarios.

import (
	"math/rand"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Java: TestRoaringBitmap.addOffsetMatchesShiftedValues and
// addOffsetMultipleOf65536DiscardsValuesOutOfRange.
func TestAddOffsetAllContainerTypes(t *testing.T) {
	rb := New()
	for _, key := range []uint32{0, 0x7FFF, 0xFFFF} {
		rb.Add(key<<16 | 7)
		rb.Add(key<<16 | 0xFFFF)
	}
	rb.AddRange(0x10000, 0x20000) // full run
	for i := uint32(0); i < 1<<16; i += 2 {
		rb.Add(0x8000<<16 | i) // bitmap container
	}
	rb.AddRange(0xFFFE0000+100, 0xFFFE0000+200)
	rb.RunOptimize()
	require.NoError(t, rb.Validate())
	st := rb.Stats()
	require.NotZero(t, st.ArrayContainers)
	require.NotZero(t, st.BitmapContainers)
	require.NotZero(t, st.RunContainers)

	values := rb.ToArray()
	for _, keyOffset := range []int64{-0x10000, -0xFFFF, -0x8001, -0x8000, -2, -1, 0, 1, 2, 0x7FFF, 0x8000, 0xFFFF, 0x10000} {
		for _, lowOffset := range []int64{0, 1, 0xFFFF} {
			offset := keyOffset*65536 + lowOffset
			expected := make([]uint32, 0, len(values))
			for _, v := range values {
				s := int64(v) + offset
				if s >= 0 && s <= MaxUint32 {
					expected = append(expected, uint32(s))
				}
			}
			got := AddOffset64(rb, offset)
			require.NoError(t, got.Validate(), "offset %d", offset)
			assert.Equal(t, uint64(len(expected)), got.GetCardinality(), "offset %d", offset)
			if len(expected) == 0 {
				assert.True(t, got.IsEmpty(), "offset %d", offset)
			} else {
				assert.Equal(t, expected, got.ToArray(), "offset %d", offset)
			}
		}
	}

	assert.True(t, AddOffset64(BitmapOf(0xFFFFFFFF), 0x10000).IsEmpty())
	assert.Equal(t, []uint32{0x10005}, AddOffset64(BitmapOf(5, 0xFFFFFFFF), 0x10000).ToArray())
	assert.True(t, AddOffset64(BitmapOf(5), -0x10000).IsEmpty())
	assert.Equal(t, []uint32{70000 - 65536}, AddOffset64(BitmapOf(5, 70000), -0x10000).ToArray())
	assert.Equal(t, []uint32{0xFFFFFFFF}, AddOffset64(BitmapOf(0xFFFFFFFE), 1).ToArray())
}

// Java: TestRoaringBitmap.test*Limit, absentValuesWithKeysFromTwoToThe31,
// previousAbsentValueOfZeroWithTheLargestValuePresent,
// testPreviousValue_LastReturnedAsUnsignedLong.
func TestNextPreviousValuesAtUniverseBoundaries(t *testing.T) {
	assert.Equal(t, int64(-1), BitmapOf(0).PreviousAbsentValue(0))
	assert.Equal(t, int64(-1), BitmapOf(0, 0xFFFFFFFF).PreviousAbsentValue(0))
	assert.Equal(t, int64(-1), BitmapOf(0xFFFFFFFF).NextAbsentValue(0xFFFFFFFF))
	assert.Equal(t, int64(-1), BitmapOf(0xFFFFFFFE).NextValue(0xFFFFFFFF))
	assert.Equal(t, int64(-1), BitmapOf(1).PreviousValue(0))
	assert.Equal(t, int64(0xFFFFFFFF), BitmapOf(0xFFFFFFFF).NextValue(0xFFFFFFFF))
	assert.Equal(t, int64(0xFFFFFFFF), BitmapOf(0xFFFFFFFF).PreviousValue(0xFFFFFFFF))

	empty := New()
	for _, p := range []uint32{0, 1, 65535, 65536, 1 << 31, 0xFFFFFFFF} {
		assert.Equal(t, int64(-1), empty.NextValue(p))
		assert.Equal(t, int64(-1), empty.PreviousValue(p))
		assert.Equal(t, int64(p), empty.NextAbsentValue(p))
		assert.Equal(t, int64(p), empty.PreviousAbsentValue(p))
	}

	top := New()
	top.AddRange(1<<32-650002, 1<<32-650000+1)
	assert.Equal(t, int64(4294317296), top.PreviousValue(0xFFFFFFFF))

	const h = uint64(1) << 31
	bitmaps := []*Bitmap{BitmapOf(1 << 31), BitmapOf(1<<31 + 5), BitmapOf(5, 1<<31+5)}
	r := New()
	r.AddRange(h-65536, h+65536)
	bitmaps = append(bitmaps, r)
	r = New()
	r.AddRange(1<<32-65536, 1<<32)
	bitmaps = append(bitmaps, r)
	probes := []uint32{0, 5, 6, 65535, 65536, 1<<31 - 65537, 1<<31 - 1, 1 << 31, 1<<31 + 5, 1<<31 + 6, 1<<31 + 65536, 1<<32 - 65537, 0xFFFFFFFF}
	for bi, b := range bitmaps {
		for _, opt := range []bool{false, true} {
			if opt {
				b.RunOptimize()
			}
			vals := b.ToArray()
			for _, p := range probes {
				// brute force through the sorted values
				i := sort.Search(len(vals), func(i int) bool { return vals[i] >= p })
				next := int64(-1)
				if i < len(vals) {
					next = int64(vals[i])
				}
				j := sort.Search(len(vals), func(i int) bool { return vals[i] > p })
				prev := int64(-1)
				if j > 0 {
					prev = int64(vals[j-1])
				}
				nextAbsent := int64(p)
				for nextAbsent <= MaxUint32 && b.Contains(uint32(nextAbsent)) {
					nextAbsent++
				}
				if nextAbsent > MaxUint32 {
					nextAbsent = -1
				}
				prevAbsent := int64(p)
				for prevAbsent >= 0 && b.Contains(uint32(prevAbsent)) {
					prevAbsent--
				}
				assert.Equal(t, next, b.NextValue(p), "bitmap %d probe %d", bi, p)
				assert.Equal(t, prev, b.PreviousValue(p), "bitmap %d probe %d", bi, p)
				assert.Equal(t, nextAbsent, b.NextAbsentValue(p), "bitmap %d probe %d", bi, p)
				assert.Equal(t, prevAbsent, b.PreviousAbsentValue(p), "bitmap %d probe %d", bi, p)
			}
		}
	}
}

// Java: TestRoaringBitmap.testCheckedAdd/testCheckedRemove, CheckedRemoveTest;
// C++: test_cpp_add_remove_checked.
func TestCheckedAddRemove(t *testing.T) {
	rb := New()
	for i := uint32(0); i < 131072; i += 2 {
		require.True(t, rb.CheckedAdd(i))
	}
	for i := uint32(1); i < 131072; i += 2 {
		require.True(t, rb.CheckedAdd(i))
	}
	for i := uint32(0); i < 131072; i++ {
		require.False(t, rb.CheckedAdd(i))
	}
	require.Equal(t, uint64(131072), rb.GetCardinality())
	for i := uint32(0); i < 131072; i += 2 {
		require.True(t, rb.CheckedRemove(i))
	}
	for i := uint32(0); i < 131072; i += 2 {
		require.False(t, rb.CheckedRemove(i))
	}
	for i := uint32(1); i < 131072; i += 2 {
		require.True(t, rb.CheckedRemove(i))
	}
	assert.True(t, rb.IsEmpty())
	require.NoError(t, rb.Validate())

	// bitmap container shrinking to an array, validating at each step
	rb = New()
	for i := uint32(0); i < 20000; i += 2 {
		rb.Add(i)
	}
	require.Equal(t, 1, int(rb.Stats().BitmapContainers))
	for i := uint32(0); i < 20000; i += 2 {
		require.True(t, rb.CheckedRemove(i))
		require.NoError(t, rb.Validate())
	}
	assert.True(t, rb.IsEmpty())

	// run container
	rb = New()
	rb.AddRange(100, 200)
	rb.RunOptimize()
	assert.False(t, rb.CheckedAdd(150))
	assert.True(t, rb.CheckedAdd(200))
	assert.True(t, rb.CheckedRemove(150))
	assert.False(t, rb.CheckedRemove(150))
	require.NoError(t, rb.Validate())
	assert.Equal(t, uint64(100), rb.GetCardinality())

	for _, v := range []uint32{0xFFFFFFF7, 0xFFFFFFFF} {
		assert.True(t, rb.CheckedAdd(v))
		assert.False(t, rb.CheckedAdd(v))
		assert.True(t, rb.Contains(v))
	}
	for _, v := range []uint32{0xFFFFFFF7, 0xFFFFFFFF} {
		assert.True(t, rb.CheckedRemove(v))
		assert.False(t, rb.CheckedRemove(v))
		assert.False(t, rb.Contains(v))
	}
	require.NoError(t, rb.Validate())
}

// Java: TestRoaringBitmap.testRangeCardinalityAtBoundary (issue 285).
func TestCardinalityInRangeAbsentStartContainer(t *testing.T) {
	rb := BitmapOf(66236)
	assert.Equal(t, uint64(1), rb.CardinalityInRange(60000, 70000))
	assert.Equal(t, uint64(0), rb.CardinalityInRange(60000, 66236))
	assert.Equal(t, uint64(1), rb.CardinalityInRange(0, 1<<32))
}

// Java: TestRoaringBitmap.fliptest_Karpenske.
func TestFlipKarpenske(t *testing.T) {
	values := []uint32{
		343798, 343799, 343800, 343801, 343803, 343804, 343805, 343807, 343809, 343811, 343812,
		343815, 343816, 343817, 343818, 343819, 343821, 343825, 343827, 343828, 343830, 343831,
		343832, 343833, 343835, 343836, 343837, 343838, 343839, 343840, 343841, 343842, 343843,
		343844, 343845, 343847, 343848, 343849, 343850, 343851, 343853, 343854, 343855, 343856,
		343858, 343859, 343860, 343861, 343862, 343863, 343864, 343865, 343866, 343868, 343869,
		343874, 343875, 343877, 343879, 343880, 343881, 343882, 343883, 343887, 343889, 343890,
		343891, 343894, 343895, 343898, 343899, 343900, 343901, 343902, 343904, 343906, 343907,
		343908, 343909, 343910, 343911, 343912, 343913, 343914, 343915, 343916, 343917, 343918,
		343919, 343921, 343922, 343923, 343924, 343927, 343928, 343929, 343930, 343931, 343932,
		343933, 343934, 343935, 343938, 343939, 343941, 343942, 343943, 343944, 343945, 343946,
		343949, 343951, 343953, 343954, 343955, 343956, 343958, 343959, 343961, 343962, 343964,
		343965, 343966, 343967, 343968, 343969, 343971, 343972, 343973, 343974, 343976, 343978,
		343979, 343981, 343982, 343983, 343985, 343987, 343988, 343989, 343992, 343993, 343994,
		343995, 343996, 343997, 343998, 344000, 344001, 344002, 344003, 344004, 344006, 344008,
		344009, 344011, 344012, 344013, 344015, 344017, 344019, 344020, 344021, 344023, 344025,
		344026, 344027, 344028, 344029, 344030, 344031, 344034, 344035, 344036, 344037, 344038,
		344039, 344040, 344042, 344043, 344046, 344047,
	}
	rb := New()
	start := 0
	flush := func(end int) { // values[start:end] is a run of consecutive values
		if end-start == 1 {
			rb.Add(values[start])
		} else {
			rb.Flip(uint64(values[start]), uint64(values[end-1])+1)
		}
	}
	for i := 1; i < len(values); i++ {
		if values[i-1]+1 != values[i] {
			flush(i)
			start = i
		}
	}
	flush(len(values))
	assert.Equal(t, uint64(182), rb.GetCardinality())
	assert.True(t, rb.Equals(BitmapOf(values...)))
	require.NoError(t, rb.Validate())
}

// Java: TestRoaringBitmap.regressionTestRemove377.
func TestRemoveRangeAfterPointRemovals377(t *testing.T) {
	rb := New()
	rb.AddRange(0, 64)
	for i := uint32(0); i < 64; i++ {
		if i != 30 && i != 32 {
			rb.Remove(i)
		}
	}
	rb.RemoveRange(0, 31)
	assert.False(t, rb.Contains(30))
	assert.True(t, rb.Contains(32))
	assert.Equal(t, uint64(1), rb.GetCardinality())
	require.NoError(t, rb.Validate())
}

// Java: TestRoaringBitmap.regressionTestEquals370.
func TestEqualsWithRunContainers370(t *testing.T) {
	a := []uint32{
		239, 240, 241, 242, 243, 244, 259, 260, 261, 262, 263, 264, 265, 266, 267, 268, 269, 270, 273,
		274, 275, 276, 277, 278, 398, 399, 400, 401, 402, 403, 404, 405, 406, 408, 409, 410, 411, 412,
		413, 420, 421, 422, 509, 510, 511, 512, 513, 514, 539, 540, 541, 542, 543, 544, 547, 548, 549,
		550, 551, 552, 553, 554, 555, 556, 557, 558, 578, 579, 580, 581, 582, 583, 584, 585, 586, 587,
		588, 589, 590, 591, 592, 593, 594, 595, 624, 625, 634, 635, 636, 649, 650, 651, 652, 653, 654,
		714, 715, 716, 718, 719, 720, 721, 722, 723, 724, 725, 726, 728, 729, 730, 731, 732, 733, 734,
		735, 736, 739, 740, 741, 742, 743, 744, 771, 772, 773,
	}
	b := []uint32{
		239, 240, 241, 242, 243, 244, 259, 260, 261, 262, 263, 264, 265, 266, 267, 268, 269, 270, 273,
		274, 275, 276, 277, 278, 398, 399, 400, 401, 402, 403, 404, 405, 406, 408, 409, 410, 411, 412,
		413, 420, 421, 422, 509, 510, 511, 512, 513, 514, 539, 540, 541, 542, 543, 544, 547, 548, 549,
		550, 551, 552, 553, 554, 555, 556, 557, 558, 578, 579, 580, 581, 582, 583, 584, 585, 586, 607,
		608, 634, 635, 636, 649, 650, 651, 652, 653, 654, 714, 715, 716, 718, 719, 720, 721, 722, 723,
		724, 725, 726, 728, 729, 730, 731, 732, 733, 734, 735, 736, 739, 740, 741, 742, 743, 744, 771,
		772, 773,
	}
	rbA := BitmapOf(a...)
	rbB := BitmapOf(b...)
	assert.False(t, rbA.Equals(rbB))
	assert.False(t, rbB.Equals(rbA))
	rbA.RunOptimize()
	assert.False(t, rbA.Equals(rbB))
	assert.False(t, rbB.Equals(rbA))
	rbB.RunOptimize()
	assert.False(t, rbA.Equals(rbB))
	assert.False(t, rbB.Equals(rbA))
	// sanity: equal content in different representations compares equal
	assert.True(t, rbA.Equals(BitmapOf(a...)))
}

// C++: test_cpp_flip ("uint32 max can be flipped", "empty range does nothing");
// C: check_full_flip.
func TestFlipLastValueAndWholeUniverse(t *testing.T) {
	rb := BitmapOf(0xFFFFFFFF)
	rb.Flip(0xFFFFFFFF, 1<<32)
	assert.True(t, rb.IsEmpty())
	require.NoError(t, rb.Validate())

	assert.True(t, Flip(BitmapOf(0xFFFFFFFF), 0xFFFFFFFF, 1<<32).IsEmpty())

	full := Flip(New(), 0, 1<<32)
	assert.Equal(t, uint64(1<<32), full.GetCardinality())
	require.NoError(t, full.Validate())
	// The package-level Flip is used here: the in-place Flip removes the
	// emptied containers one at a time, which is quadratic for 65536 keys.
	none := Flip(full, 0, 1<<32)
	assert.True(t, none.IsEmpty())
	require.NoError(t, none.Validate())
	full.Flip(0, 1<<20)
	assert.Equal(t, uint64(1<<32-1<<20), full.GetCardinality())
	assert.Equal(t, uint32(1<<20), full.Minimum())

	rb = BitmapOf(2, 3)
	rb.Flip(2, 2)
	assert.True(t, rb.Equals(BitmapOf(2, 3)))
	assert.True(t, Flip(rb, 2, 2).Equals(BitmapOf(2, 3)))
}

// Java: TestRoaringBitmap intersectBitmapWithRangeHighBits and the high-key
// bitmap tests at :1986/:1999; Fuzzer intersectsUpperBoundary.
func TestIntersectsWithIntervalEdgeCases(t *testing.T) {
	// range entirely above 2^32
	assert.False(t, BitmapOf(5).IntersectsWithInterval(1<<32+5, 1<<32+10))
	top := BitmapOf(0xFFFFFFFF)
	assert.True(t, top.IntersectsWithInterval(0xFFFFFFFF, 1<<32))
	assert.False(t, top.IntersectsWithInterval(0xFFFFFFFE, 0xFFFFFFFF))
	assert.True(t, top.IntersectsWithInterval(0, 1<<32))
	assert.False(t, top.IntersectsWithInterval(1<<32, 1<<32+1))
	assert.False(t, BitmapOf(7).IntersectsWithInterval(7, 7))

	// bitmap container with only odd values in [0, 32768)
	bc := New()
	for i := uint32(1); i < 32768; i += 2 {
		bc.Add(i)
	}
	require.Equal(t, 1, int(bc.Stats().BitmapContainers))
	assert.False(t, bc.IntersectsWithInterval(32768, 65535))
	assert.True(t, bc.IntersectsWithInterval(32767, 32769))

	// bitmap containers at key 0x8000 filled with a fixed word pattern
	for _, word := range []uint64{1789303257167203747, 2697219678014362575} {
		rb := New()
		for w := uint32(0); w < 1024; w++ {
			for bit := uint32(0); bit < 64; bit++ {
				if word&(1<<bit) != 0 {
					rb.Add(0x80000000 + w*64 + bit)
				}
			}
		}
		require.Equal(t, 1, int(rb.Stats().BitmapContainers))
		first := uint64(rb.Minimum())
		assert.False(t, rb.IntersectsWithInterval(0, first))
		assert.False(t, rb.IntersectsWithInterval(first-1, first))
		assert.True(t, rb.IntersectsWithInterval(first-1, first+1))
		assert.True(t, rb.IntersectsWithInterval(first, first+1))
	}
}

// Java: RoaringBitmapIntervalIntersectionTest; Fuzzer.java:436-480.
func TestIntersectsWithIntervalMatchesMaterializedRange(t *testing.T) {
	arrayAt := func(rb *Bitmap, key uint32) {
		for i := uint32(0); i < 1000; i += 3 {
			rb.Add(key<<16 | i)
		}
	}
	bitmapAt := func(rb *Bitmap, key uint32) {
		for i := uint32(0); i < 1<<16; i += 2 {
			rb.Add(key<<16 | i)
		}
	}
	runAt := func(rb *Bitmap, key uint32) {
		rb.AddRange(uint64(key)<<16+100, uint64(key)<<16+30000)
	}
	type tc struct {
		bm       *Bitmap
		min, sup uint64
	}
	var cases []tc
	cases = append(cases,
		tc{BitmapOf(1, 2, 3), 0, 1 << 16},
		tc{BitmapOf(0xC0000000), 0, 1 << 16},
		tc{BitmapOf(0xC0000000), 0, 256},
		tc{BitmapOf(1, 0xC0000000), 0, 256},
		tc{BitmapOf(1, 1<<16, 0xC0000000), 0, 1 << 32},
	)
	mixed := New()
	arrayAt(mixed, 10)
	bitmapAt(mixed, 20)
	runAt(mixed, 30)
	mixed.AddRange(70000, 150000)
	cases = append(cases, tc{mixed, 70000, 150000}, tc{mixed, 71000, 140000})
	fillers := []func(*Bitmap, uint32){arrayAt, bitmapAt, runAt}
	perms := [][3]int{{0, 1, 2}, {0, 2, 1}, {1, 0, 2}, {1, 2, 0}, {2, 0, 1}, {2, 1, 0}}
	for _, p := range perms {
		rb := New()
		fillers[p[0]](rb, 0)
		fillers[p[1]](rb, 1)
		fillers[p[2]](rb, 20)
		cases = append(cases, tc{rb, 67000, 150000})
	}
	big := New()
	arrayAt(big, 0)
	arrayAt(big, 1)
	arrayAt(big, 2)
	bitmapAt(big, 200)
	runAt(big, 205)
	cases = append(cases, tc{big, 199 << 16, 200<<16 + 1<<14})

	for i, c := range cases {
		c.bm.RunOptimize()
		rangeBm := New()
		rangeBm.AddRange(c.min, c.sup)
		assert.Equal(t, c.bm.Intersects(rangeBm), c.bm.IntersectsWithInterval(c.min, c.sup), "case %d", i)
	}

	// random bitmaps: invariants around the first and last values
	r := rand.New(rand.NewSource(42))
	for iter := 0; iter < 20; iter++ {
		rb := New()
		for j := 0; j < 1000; j++ {
			rb.Add(r.Uint32() >> uint(r.Intn(16)))
		}
		first, last := uint64(rb.Minimum()), uint64(rb.Maximum())
		if first > 0 {
			assert.True(t, rb.IntersectsWithInterval(first-1, first+1))
			assert.False(t, rb.IntersectsWithInterval(0, first))
		}
		assert.True(t, rb.IntersectsWithInterval(last-1, last+1))
		assert.True(t, rb.IntersectsWithInterval(last, last+1))
		assert.False(t, rb.IntersectsWithInterval(last+1, 1<<32))
	}
}

// C: select_test; Java: TestRoaringBitmap_FastRank selectOnEmpty,
// testSelectExceptionOutOfBounds.
func TestSelectOutOfBounds(t *testing.T) {
	_, err := New().Select(0)
	assert.Error(t, err)
	rb := New()
	rb.AddRange(0, 100)
	for _, opt := range []bool{false, true} {
		if opt {
			rb.RunOptimize()
		}
		v, err := rb.Select(99)
		assert.NoError(t, err)
		assert.Equal(t, uint32(99), v)
		for n := uint32(0); n < 10; n++ {
			_, err := rb.Select(100 + n)
			assert.Error(t, err)
		}
		_, err = rb.Select(0xFFFFFFFF)
		assert.Error(t, err)
	}
	rb = BitmapOf(0xFFFFFFFF, 0x80000000)
	v, err := rb.Select(1)
	assert.NoError(t, err)
	assert.Equal(t, uint32(0xFFFFFFFF), v)
}

// C: test_add_range ("invalid range", "max range"); C++ test_cpp_add_range;
// Java TestRange.testRangeRemoval.
func TestEmptyAndInvertedRangesAreNoOps(t *testing.T) {
	newR := func() *Bitmap {
		r := BitmapOf(5)
		r.AddRange(100, 200)
		return r
	}
	ref := newR()
	ops := map[string]func(*Bitmap){
		"AddRange(200,100)":    func(r *Bitmap) { r.AddRange(200, 100) },
		"AddRange(50,50)":      func(r *Bitmap) { r.AddRange(50, 50) },
		"RemoveRange(150,150)": func(r *Bitmap) { r.RemoveRange(150, 150) },
		"RemoveRange(300,10)":  func(r *Bitmap) { r.RemoveRange(300, 10) },
		"Flip(10,5)":           func(r *Bitmap) { r.Flip(10, 5) },
		"Flip(7,7)":            func(r *Bitmap) { r.Flip(7, 7) },
		"RemoveRange(0,1)":     func(r *Bitmap) { r.RemoveRange(0, 1) },
		"Flip(1<<32,1<<32)":    func(r *Bitmap) { r.Flip(1<<32, 1<<32) },
	}
	for name, op := range ops {
		r := newR()
		op(r)
		assert.True(t, r.Equals(ref), name)
		assert.NoError(t, r.Validate(), name)
	}
	assert.True(t, Flip(ref, 10, 5).Equals(ref))

	all := New()
	all.AddRange(0, 1<<32)
	assert.Equal(t, uint64(1<<32), all.GetCardinality())
	require.NoError(t, all.Validate())
}

// C++: test_cpp_add_many, test_cpp_add_bulk.
func TestAddManyUnsortedWithDuplicates(t *testing.T) {
	values := []uint32{9999, 123, 0xFFFFFFFF, 0xFFFFFFF7, 9999, 0}
	bulk := New()
	bulk.AddMany(values)
	single := New()
	for _, v := range values {
		single.Add(v)
	}
	assert.True(t, bulk.Equals(single))
	assert.Equal(t, uint64(5), bulk.GetCardinality())
	assert.Equal(t, []uint32{0, 123, 9999, 0xFFFFFFF7, 0xFFFFFFFF}, bulk.ToArray())
	require.NoError(t, bulk.Validate())
	assert.True(t, BitmapOf(values...).Equals(single))

	r := rand.New(rand.NewSource(7))
	large := make([]uint32, 0, 2*10001)
	for i := uint32(0); i <= 10000; i++ {
		large = append(large, i, 1<<31+i*3)
	}
	large = append(large, large[:5000]...) // duplicates
	r.Shuffle(len(large), func(i, j int) { large[i], large[j] = large[j], large[i] })
	bulk = New()
	bulk.AddMany(large)
	single = New()
	for _, v := range large {
		single.Add(v)
	}
	assert.True(t, bulk.Equals(single))
	assert.Equal(t, uint64(2*10001), bulk.GetCardinality())
	require.NoError(t, bulk.Validate())
}

// Java: TestRoaringBitmap.testRankOverflow, testRankBigInts.
func TestRankAtHighValues(t *testing.T) {
	rb := BitmapOf(65537)
	assert.Equal(t, uint64(0), rb.Rank(0))
	assert.Equal(t, uint64(0), rb.Rank(1))
	assert.Equal(t, uint64(0), rb.Rank(65536))
	assert.Equal(t, uint64(1), rb.Rank(65537))
	assert.Equal(t, uint64(1), rb.Rank(65538))
	assert.Equal(t, uint64(1), rb.Rank(0xFFFFFFFF))

	rb = New()
	for k := uint32(0); k < 100000; k += 7 {
		rb.Add(1<<31 + k)
	}
	for k := uint32(0); k < 100000; k += 7 {
		require.Equal(t, uint64(1+k/7), rb.Rank(1<<31+k))
	}
	assert.Equal(t, uint64(0), rb.Rank(1<<31-1))
	assert.Equal(t, rb.GetCardinality(), rb.Rank(0xFFFFFFFF))
}

// Java: TestRoaringBitmap testNegative_last, testFirstLastSigned.
func TestMinimumMaximumHighValues(t *testing.T) {
	rb := BitmapOf(1<<31, 0xFFFFFFFF)
	assert.Equal(t, uint32(1<<31), rb.Minimum())
	assert.Equal(t, uint32(0xFFFFFFFF), rb.Maximum())
	rb = BitmapOf(777, 0xFFFFFFF9)
	assert.Equal(t, uint32(777), rb.Minimum())
	assert.Equal(t, uint32(0xFFFFFFF9), rb.Maximum())
	rb = BitmapOf(0x80000000-1, 0x80000000)
	assert.Equal(t, uint32(0x7FFFFFFF), rb.Minimum())
	assert.Equal(t, uint32(0x80000000), rb.Maximum())
	rb = New()
	rb.AddRange(0xFFFFFF00, 1<<32)
	rb.RunOptimize()
	assert.Equal(t, uint32(0xFFFFFF00), rb.Minimum())
	assert.Equal(t, uint32(0xFFFFFFFF), rb.Maximum())
}

// C: test_flip_*_container_removal.
func TestFlipThatRemovesEveryValue(t *testing.T) {
	check := func(name string, rb *Bitmap, end uint64) {
		rb.Flip(0, end)
		assert.True(t, rb.IsEmpty(), name)
		assert.Equal(t, 0, rb.highlowcontainer.size(), name)
		assert.NoError(t, rb.Validate(), name)
	}
	rb := New()
	for i := uint32(0); i < 100; i++ {
		rb.Add(i)
	}
	check("array", rb, 100)
	rb = New()
	for i := uint32(0); i < 10000; i++ {
		rb.Add(i)
	}
	require.Equal(t, 1, int(rb.Stats().BitmapContainers))
	check("bitmap", rb, 10000)
	rb = New()
	rb.AddRange(0, 10000)
	check("run", rb, 10000)

	rb = New()
	rb.AddRange(0, 66002)
	rb.Flip(0, 987653576)
	assert.Equal(t, uint64(987653576-66002), rb.GetCardinality())
	assert.False(t, rb.Contains(66001))
	assert.True(t, rb.Contains(66002))
	require.NoError(t, rb.Validate())
}
