package roaring

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

// kwayKey picks the container key for the k-th key of an input: two inputs
// built with the same k may land on different keys, and some keys are left
// out, so key sets interleave and go missing the way real inputs do.
func kwayKey(rng *rand.Rand, k int) (uint64, bool) {
	return uint64(2*k+rng.Intn(2)) << 16, rng.Intn(4) != 0
}

// kwayBitmap builds a bitmap over up to keys keys holding random values at
// the given density: arrays when sparse, bitmaps when dense.
func kwayBitmap(rng *rand.Rand, keys int, density float64) *Bitmap {
	bm := New()
	for k := 0; k < keys; k++ {
		base, present := kwayKey(rng, k)
		if !present {
			continue
		}
		for i := 0; i < int(density*65536); i++ {
			bm.Add(uint32(base) + uint32(rng.Intn(65536)))
		}
	}
	return bm
}

// kwayRuns builds a run-optimized bitmap of random ranges over up to keys keys.
func kwayRuns(rng *rand.Rand, keys int) *Bitmap {
	bm := New()
	for k := 0; k < keys; k++ {
		base, present := kwayKey(rng, k)
		if !present {
			continue
		}
		for start := 0; start < 65536; start += 1000 + rng.Intn(3000) {
			bm.AddRange(base+uint64(start), base+uint64(min(start+1+rng.Intn(800), 65536)))
		}
	}
	bm.RunOptimize()
	return bm
}

// intervalBitmap holds one run container at key 0 made of the given
// half-open intervals.
func intervalBitmap(t *testing.T, intervals [][2]uint64) *Bitmap {
	t.Helper()
	bm := New()
	for _, iv := range intervals {
		bm.AddRange(iv[0], iv[1])
	}
	bm.RunOptimize()
	require.Equal(t, uint64(1), bm.Stats().RunContainers, "%v", intervals)
	return bm
}

func randomIntervals(rng *rand.Rand) [][2]uint64 {
	var out [][2]uint64
	pos := uint64(rng.Intn(64))
	for n := 1 + rng.Intn(40); n > 0 && pos < 65536; n-- {
		length := min(uint64(4+rng.Intn(3000)), 65536-pos)
		out = append(out, [2]uint64{pos, pos + length})
		pos += length + uint64(rng.Intn(200))
	}
	return out
}

// requireFastAnd checks FastAnd against the pairwise definition, that the
// result is valid, and that it shares no storage with its inputs.
func requireFastAnd(t *testing.T, inputs ...*Bitmap) {
	t.Helper()
	before := make([]*Bitmap, len(inputs))
	for i, in := range inputs {
		before[i] = in.Clone()
	}
	want := And(inputs[0], inputs[1])
	for _, bm := range inputs[2:] {
		want.And(bm)
	}
	got := FastAnd(inputs...)
	require.NoError(t, got.Validate())
	require.True(t, got.Equals(want), "%d values, want %d", got.GetCardinality(), want.GetCardinality())
	got.Flip(0, 1<<21) // rewrites every container the inputs could share
	for i, in := range inputs {
		require.True(t, in.Equals(before[i]), "input %d was modified through the result", i)
	}
}

func TestFastAndMatchesPairwise(t *testing.T) {
	rng := rand.New(rand.NewSource(3))
	densities := []float64{0.0005, 0.01, 0.05, 0.3, 0.8, 0.98}
	for trial := 0; trial < 400; trial++ {
		inputs := make([]*Bitmap, 3+rng.Intn(4))
		for i := range inputs {
			switch rng.Intn(5) {
			case 0:
				inputs[i] = kwayRuns(rng, 1+rng.Intn(4))
			case 1:
				inputs[i] = intervalBitmap(t, randomIntervals(rng))
			default:
				inputs[i] = kwayBitmap(rng, 1+rng.Intn(4), densities[rng.Intn(len(densities))])
			}
		}
		requireFastAnd(t, inputs...)
	}
}

func TestFastAndEdges(t *testing.T) {
	rng := rand.New(rand.NewSource(12))
	dense := func() *Bitmap { return kwayBitmap(rng, 1, 0.5) }
	full := intervalBitmap(t, [][2]uint64{{0, 65536}})
	var manyShort [][2]uint64
	for p := uint64(0); p < 65536; p += 64 {
		manyShort = append(manyShort, [2]uint64{p, p + 8})
	}
	for _, ivs := range [][][2]uint64{
		{{0, 65536}},
		{{777, 800}},
		{{64, 128}, {1024, 4096}},
		{{65472, 65536}},
		{{60, 70}, {130, 140}, {65530, 65536}},
		{{100, 200}, {201, 300}},
		manyShort,
	} {
		mask := intervalBitmap(t, ivs)
		requireFastAnd(t, dense(), mask, dense())             // one run masks the bitmaps
		requireFastAnd(t, dense(), mask, full, dense())       // two runs fold into one mask
		requireFastAnd(t, dense(), mask, full)                // one bitmap under two runs
		requireFastAnd(t, mask, full, intervalBitmap(t, ivs)) // runs only
	}
	// Two runs whose fold, narrower than the first run, misses the one bitmap.
	spot := New()
	spot.Add(10)
	spot.AddRange(5000, 20000)
	requireFastAnd(t, intervalBitmap(t, [][2]uint64{{0, 100}}), intervalBitmap(t, [][2]uint64{{50, 100}}), spot)
	// Ranges that each meet the first one but have no common value end the
	// key before a bitmap is read; a range narrower than the mask's span
	// clips a many-interval run.
	requireFastAnd(t, intervalBitmap(t, [][2]uint64{{0, 40000}}), intervalBitmap(t, [][2]uint64{{30000, 65536}}), intervalBitmap(t, [][2]uint64{{0, 20000}}), dense())
	requireFastAnd(t, intervalBitmap(t, [][2]uint64{{100, 40000}}), intervalBitmap(t, manyShort), dense(), dense())
	requireFastAnd(t, intervalBitmap(t, [][2]uint64{{100, 40000}}), intervalBitmap(t, [][2]uint64{{50, 60}, {30000, 50000}}), dense())
	// Two ranges that each meet the bitmap but overlap only where it is
	// empty, on word boundaries: the one-bitmap span counts to zero.
	gap := New()
	gap.Add(0)
	gap.AddRange(128, 5300)
	requireFastAnd(t, gap, intervalBitmap(t, [][2]uint64{{0, 128}}), intervalBitmap(t, [][2]uint64{{64, 192}}))
	// Disjoint runs end the call before a bitmap is read.
	requireFastAnd(t, intervalBitmap(t, [][2]uint64{{0, 1000}}), intervalBitmap(t, [][2]uint64{{2000, 3000}}), dense())
	// Keys present in some inputs only, and none shared at all.
	x := kwayBitmap(rng, 6, 0.3)
	requireFastAnd(t, x, AddOffset64(x, 1<<16), x)
	requireFastAnd(t, x, AddOffset64(x, 40<<16), dense())
	y := New()
	y.AddMany([]uint32{1, 5, 9, 70000})
	require.True(t, FastAnd(y, full, y).Equals(And(y, full)))
	require.True(t, FastAnd(y, New(), full).IsEmpty())
	require.True(t, FastAnd(full, full, full).Equals(full))
	require.True(t, FastAnd(y, y, y).Equals(y))
}

// kwayTriple returns three dense inputs, every pair overlapping, all three
// disjoint, so only the counted k-way pass can prove the result empty.
func kwayTriple() (a, b, c *Bitmap) {
	a, b, c = New(), New(), New()
	for id := uint32(0); id < 4<<16; id++ {
		switch id % 3 {
		case 0:
			a.Add(id)
			b.Add(id)
		case 1:
			b.Add(id)
			c.Add(id)
		case 2:
			a.Add(id)
			c.Add(id)
		}
	}
	return a, b, c
}

func TestFastAndEmptyAllocatesNoContainer(t *testing.T) {
	a, b, c := kwayTriple()
	require.True(t, a.Intersects(b) && b.Intersects(c) && a.Intersects(c))
	require.True(t, FastAnd(a, b, c).IsEmpty())
	// Only the answer bitmap; never a container.
	require.LessOrEqual(t, testing.AllocsPerRun(50, func() { FastAnd(a, b, c) }), 1.0)

	mask := intervalBitmap(t, [][2]uint64{{0, 65536}})
	evens, odds := New(), New()
	for v := uint32(0); v < 65536; v += 2 {
		evens.Add(v)
		odds.Add(v + 1)
	}
	require.True(t, FastAnd(mask, evens, odds).IsEmpty())
	require.LessOrEqual(t, testing.AllocsPerRun(50, func() { FastAnd(mask, evens, odds) }), 1.0)
}

// shapeKey builds one container of the given kind and cardinality on key:
// random values for arrays and bitmaps, card/8 short ranges for runs.
func shapeKey(rng *rand.Rand, kind string, card int, key uint32) *Bitmap {
	bm := New()
	base := uint64(key) << 16
	switch kind {
	case "array", "bitmap":
		for bm.GetCardinality() < uint64(card) {
			bm.Add(uint32(base) + uint32(rng.Intn(65536)))
		}
	case "runs":
		for i := 0; i < card/8; i++ {
			bm.AddRange(base+uint64(i*16), base+uint64(i*16+8))
		}
		bm.RunOptimize()
	}
	return bm
}

func pairwiseAnd(bms []*Bitmap) *Bitmap {
	r := And(bms[0], bms[1])
	for _, bm := range bms[2:] {
		r.And(bm)
	}
	return r
}

// BenchmarkFastAndShapes runs each kernel path once against the pairwise
// chain on the same inputs: the smallest input an array, tiny and large,
// against a bitmap, a run and a many-interval run; bitmaps under a run
// mask; runs only; dense inputs that are pairwise non-empty and jointly
// empty; and a term spanning eight keys.
func BenchmarkFastAndShapes(b *testing.B) {
	rng := rand.New(rand.NewSource(9))
	window := New()
	window.AddRange(0, 65536)
	wide := New()
	wide.AddRange(0, 8<<16)
	spread := func(kind string, card int) *Bitmap {
		bm := New()
		for k := uint32(0); k < 8; k++ {
			bm.Or(shapeKey(rng, kind, card, k))
		}
		return bm
	}
	x, y, z := kwayTriple()
	for _, tc := range []struct {
		name   string
		inputs []*Bitmap
	}{
		{"array30-bitmap", []*Bitmap{window, shapeKey(rng, "array", 30, 0), shapeKey(rng, "bitmap", 33000, 0)}},
		{"array4000-bitmap", []*Bitmap{window, shapeKey(rng, "array", 4000, 0), shapeKey(rng, "bitmap", 33000, 0)}},
		{"array30-runs", []*Bitmap{window, shapeKey(rng, "array", 30, 0), shapeKey(rng, "runs", 4000, 0)}},
		{"array4000-runs", []*Bitmap{window, shapeKey(rng, "array", 4000, 0), shapeKey(rng, "runs", 4000, 0)}},
		{"bitmap-bitmap", []*Bitmap{window, shapeKey(rng, "bitmap", 33000, 0), shapeKey(rng, "bitmap", 33000, 0)}},
		{"runs-bitmap", []*Bitmap{window, shapeKey(rng, "runs", 4000, 0), shapeKey(rng, "bitmap", 33000, 0)}},
		{"runs-runs", []*Bitmap{window, shapeKey(rng, "runs", 4000, 0), shapeKey(rng, "runs", 30000, 0)}},
		{"jointly-empty", []*Bitmap{x, y, z}},
		{"eight-keys", []*Bitmap{wide, spread("array", 30), spread("bitmap", 33000)}},
	} {
		b.Run(tc.name+"/fastand", func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_ = FastAnd(tc.inputs...)
			}
		})
		b.Run(tc.name+"/pairwise", func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_ = pairwiseAnd(tc.inputs)
			}
		})
	}
}
