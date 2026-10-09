package roaring

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

// kwayKey picks the container key for the k-th key of an input: inputs share
// their keys, but each leaves some out, the way real inputs do.
func kwayKey(rng *rand.Rand, k int) (uint64, bool) {
	return uint64(k) << 16, rng.Intn(8) != 0
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
// result is valid, and that it shares no storage with its inputs, and
// reports whether the result is empty.
func requireFastAnd(t *testing.T, inputs ...*Bitmap) bool {
	t.Helper()
	before := make([]*Bitmap, len(inputs))
	for i, in := range inputs {
		before[i] = in.Clone()
	}
	want := pairwiseAnd(inputs)
	got := FastAnd(inputs...)
	require.NoError(t, got.Validate())
	require.True(t, got.Equals(want), "%d values, want %d", got.GetCardinality(), want.GetCardinality())
	empty := got.IsEmpty()
	got.Flip(0, 1<<21) // rewrites every container the inputs could share
	for i, in := range inputs {
		require.True(t, in.Equals(before[i]), "input %d was modified through the result", i)
	}
	return empty
}

func TestFastAndMatchesPairwise(t *testing.T) {
	rng := rand.New(rand.NewSource(3))
	densities := []float64{0.0005, 0.01, 0.05, 0.3, 0.8, 0.98}
	whole := New()
	whole.AddRange(0, 4<<16)
	nonEmpty := 0
	for trial := 0; trial < 400; trial++ {
		inputs := make([]*Bitmap, 3+rng.Intn(10))
		for i := range inputs {
			switch rng.Intn(6) {
			case 0:
				inputs[i] = kwayRuns(rng, 1+rng.Intn(4))
			case 1:
				inputs[i] = intervalBitmap(t, randomIntervals(rng))
			case 2:
				inputs[i] = whole
			default:
				inputs[i] = kwayBitmap(rng, 1+rng.Intn(4), densities[rng.Intn(len(densities))])
			}
		}
		if !requireFastAnd(t, inputs...) {
			nonEmpty++
		}
	}
	require.Greater(t, nonEmpty, 100, "most trials must reach the intersection, not stop at a missing key")
}

func TestFastAndEdges(t *testing.T) {
	rng := rand.New(rand.NewSource(12))
	dense := func() *Bitmap {
		bm := New()
		for range 32768 {
			bm.Add(uint32(rng.Intn(65536)))
		}
		bm.AddRange(0, 256) // so a mask's edges always have values to drop
		return bm
	}
	full := intervalBitmap(t, [][2]uint64{{0, 65536}})
	// A range ending on a word boundary on one side only is masked on the other.
	requireFastAnd(t, dense(), intervalBitmap(t, [][2]uint64{{64, 100}}), dense())
	requireFastAnd(t, dense(), intervalBitmap(t, [][2]uint64{{70, 128}}), dense())
	// Two ranges spanning whole words still mask the gap between them.
	requireFastAnd(t, dense(), intervalBitmap(t, [][2]uint64{{0, 64}, {128, 192}}), dense())
	// A range narrower than a many-interval run's span clips it.
	var manyShort [][2]uint64
	for p := uint64(0); p < 65536; p += 64 {
		manyShort = append(manyShort, [2]uint64{p, p + 8})
	}
	requireFastAnd(t, intervalBitmap(t, [][2]uint64{{100, 40000}}), intervalBitmap(t, manyShort), dense(), dense())
	// Two ranges that each meet the bitmap but overlap only where it is
	// empty, so the masked result is empty.
	gap := New()
	gap.Add(0)
	gap.AddRange(128, 5300)
	requireFastAnd(t, gap, intervalBitmap(t, [][2]uint64{{0, 128}}), intervalBitmap(t, [][2]uint64{{64, 192}}))
	// Ranges that each meet the first one but have no common value end the
	// key before a bitmap is read, as do disjoint runs.
	requireFastAnd(t, intervalBitmap(t, [][2]uint64{{0, 40000}}), intervalBitmap(t, [][2]uint64{{30000, 65536}}), intervalBitmap(t, [][2]uint64{{0, 20000}}), dense())
	requireFastAnd(t, intervalBitmap(t, [][2]uint64{{0, 1000}}), intervalBitmap(t, [][2]uint64{{2000, 3000}}), dense())
	// Whole-key runs change nothing.
	y := New()
	y.AddMany([]uint32{1, 5, 9, 70000})
	require.True(t, FastAnd(y, full, y).Equals(And(y, full)))
	require.True(t, FastAnd(full, full, full).Equals(full))
}

// singletons returns n inputs holding the values 0 to n-1, one each: the
// second input already fails the check.
func singletons(n int) []*Bitmap {
	inputs := make([]*Bitmap, n)
	for i := range inputs {
		inputs[i] = New()
		inputs[i].Add(uint32(i))
	}
	return inputs
}

// kwayTriple returns three dense inputs whose pairs all overlap but whose
// intersection is empty.
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

	// Nothing is allocated for the inputs past an empty result.
	inputs := singletons(1000)
	require.True(t, FastAnd(inputs...).IsEmpty())
	require.LessOrEqual(t, testing.AllocsPerRun(50, func() { FastAnd(inputs...) }), 1.0)
}

// randomKey holds card random values on key.
func randomKey(rng *rand.Rand, card int, key uint32) *Bitmap {
	bm := New()
	for bm.GetCardinality() < uint64(card) {
		bm.Add(key<<16 | uint32(rng.Intn(65536)))
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

// BenchmarkFastAndShapes runs FastAnd against pairwise And on the same
// inputs, one shape per container pairing plus many-input shapes; the first
// seven add a whole-key range like a query window.
func BenchmarkFastAndShapes(b *testing.B) {
	rng := rand.New(rand.NewSource(9))
	window := New()
	window.AddRange(0, 65536)
	wide := New()
	wide.AddRange(0, 8<<16)
	spread := func(card int) *Bitmap {
		bm := New()
		for k := uint32(0); k < 8; k++ {
			bm.Or(randomKey(rng, card, k))
		}
		return bm
	}
	x, y, z := kwayTriple()
	same := make([]*Bitmap, 128)
	base := spread(30)
	for i := range same {
		same[i] = base.Clone()
	}
	// 1,024 inputs of 256 keys that all survive: the widest walk.
	keys256 := New()
	for k := uint32(0); k < 256; k++ {
		for v := uint32(0); v < 64; v += 2 {
			keys256.Add(k<<16 | v)
		}
	}
	surviving := make([]*Bitmap, 1024)
	for i := range surviving {
		surviving[i] = keys256.Clone()
	}
	for _, tc := range []struct {
		name   string
		inputs []*Bitmap
	}{
		{"array30-bitmap", []*Bitmap{randomKey(rng, 30, 0), randomKey(rng, 33000, 0), window}},
		{"array4000-bitmap", []*Bitmap{randomKey(rng, 4000, 0), randomKey(rng, 33000, 0), window}},
		{"array30-runs", []*Bitmap{randomKey(rng, 30, 0), runBitmap(250, 16, 32), window}},
		{"array4000-runs", []*Bitmap{randomKey(rng, 4000, 0), runBitmap(250, 16, 32), window}},
		{"bitmap-bitmap", []*Bitmap{randomKey(rng, 33000, 0), randomKey(rng, 33000, 0), window}},
		{"runs-bitmap", []*Bitmap{runBitmap(250, 16, 32), randomKey(rng, 33000, 0), window}},
		{"runs-runs", []*Bitmap{runBitmap(250, 16, 32), runBitmap(1875, 16, 32), window}},
		{"jointly-empty", []*Bitmap{x, y, z}},
		{"eight-keys", []*Bitmap{spread(30), spread(33000), wide}},
		{"singletons-10000", singletons(10000)},
		{"equal-x128", same},
		{"surviving-1024x256", surviving},
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
