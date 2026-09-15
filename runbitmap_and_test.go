package roaring

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

// randomRun builds a canonical run container: up to n sorted intervals of
// at most maxLen values separated by gaps of at most maxGap, optionally
// touching 0 and 65535.
func randomRun(rng *rand.Rand, n, maxLen, maxGap int, touchEdges bool) *runContainer16 {
	rc := &runContainer16{}
	pos := 0
	if !touchEdges {
		pos = rng.Intn(100)
	}
	for i := 0; i < n && pos < 65536; i++ {
		length := min(1+rng.Intn(1+rng.Intn(maxLen)), 65536-pos)
		rc.iv = append(rc.iv, interval16{start: uint16(pos), length: uint16(length - 1)})
		pos += length + 1 + rng.Intn(maxGap)
	}
	if last := int(rc.maximum()); touchEdges && last != 65535 {
		if start := 65535 - rng.Intn(500); last+2 <= start {
			rc.iv = append(rc.iv, interval16{start: uint16(start), length: uint16(65535 - start)})
		}
	}
	return rc
}

// randomBitmapContainer ANDs ands random words per position, so the density
// is 1/2^ands.
func randomBitmapContainer(rng *rand.Rand, ands int) *bitmapContainer {
	bc := newBitmapContainer()
	for i := range bc.bitmap {
		w := rng.Uint64()
		for k := 1; k < ands; k++ {
			w &= rng.Uint64()
		}
		bc.bitmap[i] = w
	}
	bc.computeCardinality()
	return bc
}

func requireAndResult(t *testing.T, want, got container) {
	t.Helper()
	require.True(t, want.equals(got), "want %d values, got %d", want.getCardinality(), got.getCardinality())
	_, isArray := got.(*arrayContainer)
	require.Equal(t, got.getCardinality() <= arrayDefaultMaxSize, isArray, "container kind must follow its cardinality")
}

// The reference is the old path: materialize the run, then the unchanged
// bitmap-bitmap intersection.
func TestRunAndBitmapContainer(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	for trial := 0; trial < 300; trial++ {
		rc := randomRun(rng, 1+rng.Intn(40), 3000, 2000, trial%5 == 0)
		bc := randomBitmapContainer(rng, 1+trial%4)
		want := newBitmapContainerFromRun(rc).andBitmap(bc)
		requireAndResult(t, want, rc.and(bc))
		requireAndResult(t, want, bc.and(rc))
		requireAndResult(t, want, bc.clone().iand(rc))
	}
}

// Runs with more than runAndScratchIntervals intervals take the scratch
// path; check it on regular and random runs against every density.
func TestRunAndBitmapContainerManyIntervals(t *testing.T) {
	rng := rand.New(rand.NewSource(21))
	regular := func(stride, length int) *runContainer16 {
		rc := &runContainer16{}
		for start := 0; start+length <= 65536; start += stride {
			rc.iv = append(rc.iv, interval16{start: uint16(start), length: uint16(length - 1)})
		}
		return rc
	}
	runs := []*runContainer16{regular(32, 8), regular(64, 1), regular(3, 2), regular(200, 100)}
	for i := 0; i < 60; i++ {
		runs = append(runs, randomRun(rng, 65+rng.Intn(400), 60, 120, false))
	}
	for i, rc := range runs {
		require.Greater(t, len(rc.iv), runAndScratchIntervals, "run %d", i)
		for _, bc := range []*bitmapContainer{randomBitmapContainer(rng, 6), randomBitmapContainer(rng, 2), randomBitmapContainer(rng, 1), newBitmapContainerwithRange(0, 65535)} {
			want := newBitmapContainerFromRun(rc).andBitmap(bc)
			requireAndResult(t, want, rc.and(bc))
			requireAndResult(t, want, bc.and(rc))
			requireAndResult(t, want, bc.clone().iand(rc))
			require.Equal(t, !want.isEmpty(), rc.intersects(bc), "run %d", i)
		}
	}
}

func TestRunAndBitmapContainerEdges(t *testing.T) {
	empty := newBitmapContainer()
	full := newBitmapContainerwithRange(0, 65535)
	single := &runContainer16{iv: []interval16{{start: 100, length: 9}}}
	edges := &runContainer16{iv: []interval16{{start: 0, length: 0}, {start: 65535, length: 0}}}
	atThreshold := &runContainer16{iv: []interval16{{start: 0, length: arrayDefaultMaxSize - 1}}}
	pastThreshold := &runContainer16{iv: []interval16{{start: 0, length: arrayDefaultMaxSize}}}
	for _, rc := range []*runContainer16{single, edges, atThreshold, pastThreshold} {
		for _, bc := range []*bitmapContainer{empty, full} {
			want := newBitmapContainerFromRun(rc).andBitmap(bc)
			requireAndResult(t, want, rc.and(bc))
			requireAndResult(t, want, bc.clone().iand(rc))
		}
	}
}

// An empty intersection allocates only its empty array container; the old
// path allocated an 8 KiB bitmap to hold the run first.
func TestRunAndBitmapContainerEmptyAllocation(t *testing.T) {
	rc := &runContainer16{iv: []interval16{{start: 0, length: 999}, {start: 30000, length: 999}}}
	bc := newBitmapContainerwithRange(2000, 20000)
	require.LessOrEqual(t, testing.AllocsPerRun(100, func() {
		if !rc.and(bc).isEmpty() {
			t.Fatal("expected empty")
		}
	}), 1.0)
	// The clone costs two allocations, the empty result one.
	require.LessOrEqual(t, testing.AllocsPerRun(100, func() {
		_ = bc.clone().iand(rc)
	}), 3.0)
}

// The reference is the old definition, !rc.and(c).isEmpty().
func TestRunContainerIntersects(t *testing.T) {
	rng := rand.New(rand.NewSource(11))
	for trial := 0; trial < 400; trial++ {
		rc := randomRun(rng, 1+rng.Intn(30), 3000, 2000, trial%7 == 0)
		ac := newArrayContainer()
		for i := 0; i < rng.Intn(300); i++ {
			ac.iadd(uint16(rng.Intn(65536)))
		}
		others := []container{randomBitmapContainer(rng, 1+trial%3*3), ac, randomRun(rng, 1+rng.Intn(30), 3000, 2000, trial%5 == 0)}
		for _, c := range others {
			want := !rc.and(c).isEmpty()
			require.Equal(t, want, rc.intersects(c), "trial %d run vs %T", trial, c)
			require.Equal(t, want, c.intersects(rc), "trial %d %T vs run", trial, c)
		}
	}
	// A few values against many intervals take the search branch; the
	// reference is the same definition.
	many := randomRun(rng, 300, 60, 120, false)
	require.Greater(t, len(many.iv), 200)
	for trial := 0; trial < 200; trial++ {
		few := newArrayContainer()
		for i := 0; i < 1+rng.Intn(6); i++ {
			few.iadd(uint16(rng.Intn(65536)))
		}
		want := !many.and(few).isEmpty()
		require.Equal(t, want, many.intersects(few), "trial %d few vs many", trial)
		require.Equal(t, want, few.intersects(many), "trial %d many vs few", trial)
	}
	// Edges: empty array, touching intervals, single values at both ends.
	empty := newArrayContainer()
	edges := &runContainer16{iv: []interval16{{start: 0, length: 0}, {start: 65535, length: 0}}}
	require.False(t, edges.intersects(empty))
	one := newArrayContainer()
	one.iadd(65535)
	require.True(t, edges.intersects(one))
	touching := &runContainer16{iv: []interval16{{start: 1, length: 65533}}}
	require.False(t, edges.intersects(touching))
	require.True(t, edges.intersects(newBitmapContainerwithRange(0, 0)))
}

func TestRunContainerIntersectsAllocatesNothing(t *testing.T) {
	rc := &runContainer16{iv: []interval16{{start: 0, length: 999}, {start: 30000, length: 999}}}
	bc := newBitmapContainerwithRange(2000, 20000)
	ac := newArrayContainerRange(2000, 3000)
	other := &runContainer16{iv: []interval16{{start: 5000, length: 100}}}
	for _, c := range []container{bc, ac, other} {
		require.Zero(t, testing.AllocsPerRun(100, func() { _ = rc.intersects(c) }), "%T", c)
	}
}

// runBitmap holds one run container at key 0 covering n intervals of the
// given length, spaced stride apart.
func runBitmap(n, length, stride int) *Bitmap {
	bm := New()
	for i := 0; i < n; i++ {
		bm.AddRange(uint64(i*stride), uint64(i*stride+length))
	}
	bm.RunOptimize()
	return bm
}

// BenchmarkRunAndBitmap intersects a run-optimized bitmap with a dense one
// over a single key through the public API, one case per result kind, plus
// the in-place form and the Intersects test.
func BenchmarkRunAndBitmap(b *testing.B) {
	rng := rand.New(rand.NewSource(7))
	short := runBitmap(20, 300, 3000)  // 6000 values
	long := runBitmap(4, 10000, 16000) // 40000 values
	many := runBitmap(1024, 8, 64)     // 8192 values in 1024 intervals
	gaps, quarter, half, dense := New(), New(), New(), New()
	for v := uint32(0); v < 65536; v++ {
		if !short.Contains(v) && rng.Intn(2) == 0 {
			gaps.Add(v)
		}
		if rng.Intn(4) == 0 {
			quarter.Add(v)
		}
		if rng.Intn(2) == 0 {
			half.Add(v)
		}
		if rng.Intn(10) != 0 {
			dense.Add(v)
		}
	}
	for _, tc := range []struct {
		name string
		x, y *Bitmap
	}{
		{"and/empty", short, gaps},
		{"and/array", short, half},
		{"and/bitmap", long, dense},
		{"and/many-intervals/array", many, quarter},
		{"and/many-intervals/bitmap", many, dense},
	} {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_ = And(tc.x, tc.y)
			}
		})
	}
	for _, tc := range []struct {
		name string
		x    *Bitmap
	}{
		{"iand/bitmap", long},
		{"iand/many-intervals", many},
	} {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			x := dense.Clone()
			for b.Loop() {
				x.And(tc.x)
			}
		})
	}
	few := New()
	few.AddMany([]uint32{100, 30000, 60000})
	for _, tc := range []struct {
		name string
		x, y *Bitmap
	}{
		{"intersects/disjoint", short, gaps},
		{"intersects/overlapping", short, half},
		{"intersects/few-values-many-intervals", many, few},
	} {
		b.Run(tc.name, func(b *testing.B) {
			for b.Loop() {
				_ = tc.x.Intersects(tc.y)
			}
		})
	}
}
