//go:build arm64 && !gccgo && !appengine
// +build arm64,!gccgo,!appengine

package roaring

import (
	"math/rand"
	"testing"
)

func checkPair(t *testing.T, label string, a, b []uint16) {
	t.Helper()
	want := localintersect2by2Cardinality(a, b)
	got := intersection2by2Cardinality(a, b)
	if want != got {
		t.Fatalf("%s: la=%d lb=%d want %d got %d", label, len(a), len(b), want, got)
	}
	got = intersection2by2Cardinality(b, a)
	if want != got {
		t.Fatalf("%s swapped: la=%d lb=%d want %d got %d", label, len(b), len(a), want, got)
	}
	checkMaterialize(t, label, a, b)
	checkMaterialize(t, label+" swapped", b, a)
}

const canary = 0xABAB

// checkMaterialize checks both shipping buffer contracts: exact capacity
// with canaries (andArray) and a buffer aliasing set1 in place (iandArray).
func checkMaterialize(t *testing.T, label string, a, b []uint16) {
	t.Helper()
	m := len(a)
	if len(b) < m {
		m = len(b)
	}
	want := make([]uint16, m)
	wn := localintersect2by2(a, b, want)
	want = want[:wn]

	backing := make([]uint16, m+16)
	for i := range backing {
		backing[i] = canary
	}
	gn := intersection2by2(a, b, backing[0:0:m])
	if gn != wn {
		t.Fatalf("%s: la=%d lb=%d want len %d got %d", label, len(a), len(b), wn, gn)
	}
	for i := 0; i < wn; i++ {
		if backing[i] != want[i] {
			t.Fatalf("%s: la=%d lb=%d idx %d want %d got %d", label, len(a), len(b), i, want[i], backing[i])
		}
	}
	for i := m; i < m+16; i++ {
		if backing[i] != canary {
			t.Fatalf("%s: overstore past cap at %d (cap %d)", label, i, m)
		}
	}

	inplace := make([]uint16, len(a)+16)
	for i := range inplace {
		inplace[i] = canary
	}
	copy(inplace, a)
	gn = intersection2by2(inplace[:len(a)], b, inplace[0:len(a):len(a)])
	if gn != wn {
		t.Fatalf("%s inplace: la=%d lb=%d want len %d got %d", label, len(a), len(b), wn, gn)
	}
	for i := 0; i < wn; i++ {
		if inplace[i] != want[i] {
			t.Fatalf("%s inplace: idx %d want %d got %d", label, i, want[i], inplace[i])
		}
	}
	for i := len(a); i < len(a)+16; i++ {
		if inplace[i] != canary {
			t.Fatalf("%s inplace: overstore past cap at %d (cap %d)", label, i, len(a))
		}
	}
}

func TestIntersectNEONSmallSizes(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	for la := 0; la <= 40; la++ {
		for lb := 12; lb <= 44; lb++ {
			a := genSortedUnique(rng, la, 120)
			b := genSortedUnique(rng, lb, 120)
			checkPair(t, "small", a, b)
		}
	}
}

func TestIntersectNEONRangeDisjoint(t *testing.T) {
	// Interleaved 0xFFFF fast-forward, and the boundary match at endpoints.
	for _, n := range []int{64, 512} {
		a := make([]uint16, n)
		b := make([]uint16, n)
		for i := 0; i < n; i++ {
			a[i] = uint16(65534 - 2*(n-1-i))
			b[i] = uint16(65535 - 2*(n-1-i))
		}
		checkPair(t, "extremes-interleaved", a, b)
		for i := 0; i < n; i++ {
			a[i] = uint16(i)
			b[i] = uint16(n - 1 + i)
		}
		checkPair(t, "endpoint", a, b)
	}
}

func TestIntersectNEONGallopingBoundary(t *testing.T) {
	// 32x2048 stays on the kernel; 32x2049 crosses into galloping.
	rng := rand.New(rand.NewSource(5))
	small := genSortedUnique(rng, 32, 65536)
	for _, large := range []int{2048, 2049} {
		big := genSortedUnique(rng, large, 65536)
		checkPair(t, "gallop-boundary", small, big)
	}
}

func TestIntersectNEONRandom(t *testing.T) {
	rng := rand.New(rand.NewSource(99))
	for iter := 0; iter < 100; iter++ {
		la := 64 + rng.Intn(4000)
		lb := 64 + rng.Intn(4000)
		if iter%5 == 0 {
			lb = 64 + rng.Intn(120) // skewed pairs
		}
		max := 256 + rng.Intn(65280)
		if la > max {
			la = max
		}
		if lb > max {
			lb = max
		}
		a := genSortedUnique(rng, la, max)
		b := genSortedUnique(rng, lb, max)
		checkPair(t, "random", a, b)
	}
}

// And of a bitmap with itself: set1, set2, and the output are one array.
func TestIntersectNEONSelfAlias(t *testing.T) {
	rng := rand.New(rand.NewSource(13))
	for iter := 0; iter < 25; iter++ {
		n := 8 + rng.Intn(3000)
		set := genSortedUnique(rng, n, 256+rng.Intn(65280))
		n = len(set)

		shared := make([]uint16, n, n+16)
		copy(shared, set)
		got := shared[:intersection2by2(shared, shared, shared)]
		if len(got) != len(set) {
			t.Fatalf("self-alias: n=%d got len %d", n, len(got))
		}
		for i := range set {
			if got[i] != set[i] {
				t.Fatalf("self-alias: n=%d idx %d want %d got %d", n, i, set[i], got[i])
			}
		}

		// set1 == set2 with an independent buffer.
		out := make([]uint16, n)
		if g := intersection2by2(set, set, out[:0:n]); g != len(set) {
			t.Fatalf("same-inputs: n=%d want %d got %d", n, len(set), g)
		}
	}

	// Spare capacity (cap>len) selects a different store path.
	a := genSortedUnique(rng, 64, 256)
	b := genSortedUnique(rng, 64, 256)
	want := make([]uint16, 64)
	wn := localintersect2by2(a, b, want)
	backing := make([]uint16, 64+64)
	for i := range backing {
		backing[i] = canary
	}
	copy(backing, a)
	gn := intersection2by2(backing[:64], b, backing[:0:64+40])
	if gn != wn {
		t.Fatalf("sparecap: want len %d got %d", wn, gn)
	}
	for i := 0; i < wn; i++ {
		if backing[i] != want[i] {
			t.Fatalf("sparecap: idx %d want %d got %d", i, want[i], backing[i])
		}
	}
	for i := 64 + 40; i < 64+64; i++ {
		if backing[i] != canary {
			t.Fatalf("sparecap: overstore past cap at %d", i)
		}
	}
}

// Duplicate input: membership and size unspecified, but no panics, no
// out-of-cap stores, and sorted output for sorted input.
func TestIntersectNEONDuplicateInputBounded(t *testing.T) {
	sorted := func(s []uint16) bool {
		for i := 1; i < len(s); i++ {
			if s[i-1] > s[i] {
				return false
			}
		}
		return true
	}
	check := func(label string, a, b []uint16) {
		t.Helper()
		m := len(a)
		if len(b) < m {
			m = len(b)
		}
		// Exact cap and spare cap select different store gates in the kernel.
		for _, spare := range []int{0, 16} {
			backing := make([]uint16, m+spare+16)
			for i := range backing {
				backing[i] = canary
			}
			n := intersection2by2(a, b, backing[0:0:m+spare])
			if n < 0 || n > m+spare {
				t.Fatalf("%s cap+%d: size %d outside [0,%d]", label, spare, n, m+spare)
			}
			if sorted(a) && sorted(b) && !sorted(backing[:n]) {
				t.Fatalf("%s cap+%d: unsorted result %v", label, spare, backing[:n])
			}
			for i := m + spare; i < len(backing); i++ {
				if backing[i] != canary {
					t.Fatalf("%s cap+%d: store past cap at index %d", label, spare, i)
				}
			}
		}
		if c := intersection2by2Cardinality(a, b); c < 0 || c > len(a)+len(b) {
			t.Fatalf("%s: cardinality %d outside [0,%d]", label, c, len(a)+len(b))
		}

		// The iandArray geometry: output aliases set1, exact cap and spare cap.
		for _, spare := range []int{0, 16} {
			inplace := make([]uint16, len(a)+spare+16)
			for i := range inplace {
				inplace[i] = canary
			}
			copy(inplace, a)
			n := intersection2by2(inplace[:len(a)], b, inplace[0:len(a):len(a)+spare])
			if n < 0 || n > len(a)+spare {
				t.Fatalf("%s inplace+%d: size %d outside [0,%d]", label, spare, n, len(a)+spare)
			}
			if sorted(a) && sorted(b) && !sorted(inplace[:n]) {
				t.Fatalf("%s inplace+%d: unsorted result %v", label, spare, inplace[:n])
			}
			for i := len(a) + spare; i < len(inplace); i++ {
				if inplace[i] != canary {
					t.Fatalf("%s inplace+%d: store past cap at index %d", label, spare, i)
				}
			}
		}
	}

	flat := make([]uint16, 64)
	for i := range flat {
		flat[i] = 100
	}
	ramp := make([]uint16, 32)
	for i := range ramp {
		ramp[i] = uint16(100 + i)
	}
	seam := make([]uint16, 33)
	for i := 0; i < 32; i++ {
		seam[i] = uint16(69 + i)
	}
	seam[32] = 100
	check("flat", flat, ramp)
	check("flat-swapped", ramp, flat)
	check("dup-seam", ramp, seam)
	check("dup-seam-swapped", seam, ramp)

	// Without the kernel's store clamp these corrupt the canaries.
	wa := []uint16{100, 100, 100, 100, 101, 101, 101, 101, 101, 101, 102, 102, 102, 102, 103, 103}
	wb := []uint16{100, 101, 101, 101, 101, 101, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102}
	check("clamp-witness", wa, wb)
	check("clamp-witness-swapped", wb, wa)
	deep := make([]uint16, 25)
	for i := range deep {
		deep[i] = 100
	}
	deep[24] = 200
	tall := []uint16{100, 100, 100, 100, 100, 100, 100, 200, 201, 202, 203, 204, 205, 206, 207, 208}
	check("clamp-spill", tall, deep)
	check("clamp-spill-swapped", deep, tall)

	// Fills the buffer inside the kernel with matches still pending.
	pa := []uint16{0, 1, 1, 1, 2, 3, 4, 5, 6, 7, 7, 7, 9, 9, 10, 10, 12}
	pb := []uint16{0, 1, 3, 4, 4, 4, 5, 7, 7, 9, 10, 10, 12, 13, 13, 15}
	check("tail-overfill", pa, pb)
	check("tail-overfill-swapped", pb, pa)

	// Fills the buffer inside the bounded tail merge with matches pending.
	ma := []uint16{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 2, 2, 2, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 4, 4, 5, 5, 5, 6, 6, 6, 6, 6, 6, 6, 6, 6, 7, 7, 8, 8, 8, 9, 9, 9, 9, 9, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 11, 11, 11, 12, 12, 12, 12, 12, 12, 12, 12}
	mb := []uint16{0, 1, 1, 1, 1, 1, 2, 3, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 5, 5, 6, 6, 6, 6, 6, 6, 6, 6, 7, 7, 7, 7, 7, 8, 8, 8, 9, 9, 9, 9, 9, 9, 9, 9, 9, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 11}
	check("merge-midfill", ma, mb)
	check("merge-midfill-swapped", mb, ma)

	// Fills the buffer inside the spill drain with matches pending.
	da := []uint16{0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 2, 2, 2, 3, 4, 4, 4, 4, 4, 4, 5, 5, 5, 5, 5, 5, 5, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 8, 8, 8, 8, 8, 8, 8, 8, 9, 9, 9, 9, 9, 10, 10, 10}
	db := []uint16{0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 2, 2, 2, 2}
	check("drain-midfill", da, db)
	check("drain-midfill-swapped", db, da)

	// Needs all seven free slots at the tail dispatch: pins the >= 7 bound.
	ga := []uint16{0, 0, 0, 1, 1, 1, 2, 2, 2, 2, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 4, 4, 4, 4, 4, 4, 4, 4, 4, 5, 5, 6, 6, 7, 7, 7, 7, 7, 7, 8, 8, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 10, 10, 10, 11, 12, 12, 12, 12, 12, 12, 13, 13, 14, 14, 14, 14, 15, 15, 15, 15, 15, 15, 15, 16, 16, 16, 16, 16, 16}
	gb := []uint16{0, 0, 0, 1, 1, 1, 1, 1, 2, 3, 3, 3, 4, 4, 5, 5, 5, 5, 5, 5, 5, 6, 7, 7, 7, 7, 8, 8, 8, 8, 8, 9, 10, 10, 10, 10, 10, 10, 10, 10, 10, 11, 12, 13, 13, 13, 14, 14, 15, 16, 16, 16, 16, 16, 16}
	check("dispatch-boundary", ga, gb)
	check("dispatch-boundary-swapped", gb, ga)

	// Below the kernel threshold the scalar dispatch must hold the same bound.
	small := genSortedDup(rand.New(rand.NewSource(7)), 15)
	check("sub-threshold", small, small[:15])

	// A 64x size ratio sends duplicates down the galloping dispatch.
	flat16 := make([]uint16, 16)
	flat1025 := make([]uint16, 1025)
	for i := range flat16 {
		flat16[i] = 7
	}
	for i := range flat1025 {
		flat1025[i] = 7
	}
	check("gallop-dup", flat16, flat1025)
	check("gallop-dup-swapped", flat1025, flat16)

	// Buffer capacity beyond min(len1, len2): the kernel's other store gate.
	sp1 := []uint16{100, 100, 100, 100, 100, 100, 100, 200, 201, 202, 203, 204, 205, 206, 207, 208}
	sp2 := make([]uint16, 25)
	for i := range sp2 {
		sp2[i] = 100
	}
	sp2[24] = 200
	spare := make([]uint16, 25+16)
	for i := range spare {
		spare[i] = canary
	}
	if n := intersection2by2(sp1, sp2, spare[0:0:25]); n < 0 || n > 25 {
		t.Fatalf("spare-cap: size %d outside [0,25]", n)
	}
	for i := 25; i < len(spare); i++ {
		if spare[i] != canary {
			t.Fatalf("spare-cap: store past cap at index %d", i)
		}
	}

	// Sorted-with-duplicates sweep over sizes and geometries.
	rng := rand.New(rand.NewSource(4242))
	for iter := 0; iter < 5000; iter++ {
		a := genSortedDup(rng, 16+rng.Intn(200))
		b := genSortedDup(rng, 16+rng.Intn(200))
		check("sweep", a, b)
	}

	// Unsorted values also arrive through UnmarshalBinary; same bounds.
	for iter := 0; iter < 1000; iter++ {
		a := make([]uint16, 16+rng.Intn(200))
		b := make([]uint16, 16+rng.Intn(200))
		for i := range a {
			a[i] = uint16(rng.Intn(1 << 16))
		}
		for i := range b {
			b[i] = uint16(rng.Intn(1 << 16))
		}
		check("garbage", a, b)
	}
}

// Valid input whose tails exceed the remaining buffer (retained-block exit).
func TestIntersectNEONRetainedBlockTail(t *testing.T) {
	a := make([]uint16, 23)
	b := make([]uint16, 17)
	for i := 0; i < 16; i++ {
		a[i] = uint16(i)
	}
	for i := 16; i < 23; i++ {
		a[i] = uint16(84 + i)
	}
	for i := 0; i < 15; i++ {
		b[i] = uint16(i)
	}
	b[15], b[16] = 16, 17
	checkPair(t, "retained-block", a, b)

	// Near-full buffer on valid input: the bounded tail merge does the store.
	c := make([]uint16, 23)
	d := make([]uint16, 17)
	for i := range c {
		c[i] = uint16(i)
	}
	for i := range d {
		d[i] = uint16(i)
	}
	checkPair(t, "bounded-tail-store", c, d)
}

// Duplicate arrays pass Validate; And after a marshal round trip must not panic.
func TestIntersectNEONDuplicateRoundtrip(t *testing.T) {
	dupBitmap := func(vals []uint16) *Bitmap {
		bm := New()
		bm.highlowcontainer.appendContainer(0, &arrayContainer{content: vals}, false)
		return bm
	}
	a := dupBitmap([]uint16{0, 1, 1, 1, 2, 3, 4, 5, 6, 7, 7, 7, 9, 9, 10, 10, 12})
	b := dupBitmap([]uint16{0, 1, 3, 4, 4, 4, 5, 7, 7, 9, 10, 10, 12, 13, 13, 15})
	for _, bm := range []*Bitmap{a, b} {
		if err := bm.Validate(); err != nil {
			t.Fatalf("duplicate bitmap failed Validate: %v", err)
		}
	}
	for _, pair := range [][2]*Bitmap{{a, b}, {b, a}} {
		data, err := pair[0].MarshalBinary()
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		got := New()
		if err := got.UnmarshalBinary(data); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		out := And(got, pair[1])
		if c := out.GetCardinality(); c > 16 {
			t.Fatalf("And cardinality %d exceeds smaller input size", c)
		}
		if !out.IsEmpty() {
			if err := out.Validate(); err != nil {
				t.Fatalf("And result fails Validate: %v", err)
			}
		}
		got.And(pair[1]) // in-place geometry
		if !got.IsEmpty() {
			if err := got.Validate(); err != nil {
				t.Fatalf("in-place And result fails Validate: %v", err)
			}
		}
	}
}

// genSortedDup emits a sorted array where runs of equal values are common.
func genSortedDup(r *rand.Rand, n int) []uint16 {
	out := make([]uint16, n)
	v := 0
	for i := range out {
		out[i] = uint16(v)
		if r.Intn(3) > 0 {
			v += 1 + r.Intn(3)
		}
	}
	return out
}

// Independently misaligned slice starts for set1, set2, and the output.
func TestIntersectNEONMisalignment(t *testing.T) {
	rng := rand.New(rand.NewSource(77))
	for _, n := range []int{16, 17, 24, 25, 64, 65} {
		for o1 := 0; o1 < 8; o1++ {
			for o2 := 0; o2 < 8; o2++ {
				for oo := 0; oo < 8; oo++ {
					a := genSortedUnique(rng, n, 3*n)
					b := genSortedUnique(rng, n, 3*n)
					back1 := make([]uint16, o1+len(a)+8)
					back2 := make([]uint16, o2+len(b)+8)
					copy(back1[o1:], a)
					copy(back2[o2:], b)
					s1 := back1[o1 : o1+len(a)]
					s2 := back2[o2 : o2+len(b)]

					m := len(a)
					if len(b) < m {
						m = len(b)
					}
					want := make([]uint16, m)
					wn := localintersect2by2(a, b, want)

					for _, pair := range [][2][]uint16{{s1, s2}, {s2, s1}} {
						backo := make([]uint16, oo+m+16)
						for i := range backo {
							backo[i] = canary
						}
						out := backo[oo : oo : oo+m]
						gn := intersection2by2(pair[0], pair[1], out)
						if gn != wn {
							t.Fatalf("n=%d o1=%d o2=%d oo=%d: len want %d got %d", n, o1, o2, oo, wn, gn)
						}
						for i := 0; i < wn; i++ {
							if backo[oo+i] != want[i] {
								t.Fatalf("n=%d o1=%d o2=%d oo=%d idx %d: want %d got %d", n, o1, o2, oo, i, want[i], backo[oo+i])
							}
						}
						for i := oo + m; i < len(backo); i++ {
							if backo[i] != canary {
								t.Fatalf("n=%d o1=%d o2=%d oo=%d: overstore at %d", n, o1, o2, oo, i)
							}
						}
						if c := intersection2by2Cardinality(pair[0], pair[1]); c != wn {
							t.Fatalf("n=%d o1=%d o2=%d card: want %d got %d", n, o1, o2, wn, c)
						}
					}
				}
			}
		}
	}
}

func genSortedUnique(r *rand.Rand, n, valRange int) []uint16 {
	if n > valRange {
		n = valRange
	}
	seen := make(map[uint16]bool, n)
	for len(seen) < n {
		seen[uint16(r.Intn(valRange))] = true
	}
	out := make([]uint16, 0, n)
	for v := 0; v < valRange; v++ {
		if seen[uint16(v)] {
			out = append(out, uint16(v))
		}
	}
	return out
}
