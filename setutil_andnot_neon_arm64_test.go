//go:build arm64 && !gccgo && !appengine
// +build arm64,!gccgo,!appengine

package roaring

import (
	"math/rand"
	"testing"
)

const andnotCanary = 0xABAB

// checkDifference verifies the exact-capacity, in-place and spare-capacity
// buffer contracts.
func checkDifference(t *testing.T, label string, a, b []uint16) {
	t.Helper()
	want := make([]uint16, len(a))
	wn := localdifference(a, b, want)
	want = want[:wn]

	m := len(a)
	backing := make([]uint16, m+16)
	for i := range backing {
		backing[i] = andnotCanary
	}
	gn := difference(a, b, backing[0:0:m])
	if gn != wn {
		t.Fatalf("%s: la=%d lb=%d want len %d got %d", label, len(a), len(b), wn, gn)
	}
	for i := 0; i < wn; i++ {
		if backing[i] != want[i] {
			t.Fatalf("%s: la=%d lb=%d idx %d want %d got %d", label, len(a), len(b), i, want[i], backing[i])
		}
	}
	for i := m; i < m+16; i++ {
		if backing[i] != andnotCanary {
			t.Fatalf("%s: overstore past cap at %d (cap %d)", label, i, m)
		}
	}

	inplace := make([]uint16, len(a)+16)
	for i := range inplace {
		inplace[i] = andnotCanary
	}
	copy(inplace, a)
	gn = difference(inplace[:len(a)], b, inplace[0:len(a):len(a)])
	if gn != wn {
		t.Fatalf("%s inplace: la=%d lb=%d want len %d got %d", label, len(a), len(b), wn, gn)
	}
	for i := 0; i < wn; i++ {
		if inplace[i] != want[i] {
			t.Fatalf("%s inplace: idx %d want %d got %d", label, i, want[i], inplace[i])
		}
	}
	for i := len(a); i < len(a)+16; i++ {
		if inplace[i] != andnotCanary {
			t.Fatalf("%s inplace: overstore past cap at %d (cap %d)", label, i, len(a))
		}
	}

	spare := make([]uint16, len(a)+24)
	for i := range spare {
		spare[i] = andnotCanary
	}
	copy(spare, a)
	gn = difference(spare[:len(a)], b, spare[0:len(a):len(a)+8])
	if gn != wn {
		t.Fatalf("%s sparecap: la=%d lb=%d want len %d got %d", label, len(a), len(b), wn, gn)
	}
	for i := 0; i < wn; i++ {
		if spare[i] != want[i] {
			t.Fatalf("%s sparecap: idx %d want %d got %d", label, i, want[i], spare[i])
		}
	}
	for i := len(a) + 8; i < len(a)+24; i++ {
		if spare[i] != andnotCanary {
			t.Fatalf("%s sparecap: overstore past cap at %d (cap %d)", label, i, len(a)+8)
		}
	}
}

func checkDifferenceBoth(t *testing.T, label string, a, b []uint16) {
	t.Helper()
	checkDifference(t, label, a, b)
	checkDifference(t, label+" swapped", b, a)
}

func TestAndnotNEONSmallSizes(t *testing.T) {
	r := rand.New(rand.NewSource(41))
	for la := 0; la <= 40; la++ {
		for lb := 12; lb <= 44; lb++ {
			a := andnotSortedUnique(r, la, 160)
			b := andnotSortedUnique(r, lb, 160)
			checkDifferenceBoth(t, "small", a, b)
		}
	}
}

func TestAndnotNEONRangeDisjoint(t *testing.T) {
	for _, n := range []int{64, 512} {
		a := make([]uint16, n)
		b := make([]uint16, n)
		for i := 0; i < n; i++ {
			a[i] = uint16(2 * i)
			b[i] = uint16(2*i + 1)
		}
		checkDifferenceBoth(t, "interleaved", a, b)

		for i := 0; i < n; i++ {
			a[i] = uint16(i)
			b[i] = uint16(65535 - n + 1 + i)
		}
		checkDifferenceBoth(t, "extremes", a, b)

		for i := 0; i < n; i++ {
			b[i] = uint16(n - 1 + i)
		}
		checkDifferenceBoth(t, "endpoint", a, b)
	}
}

// Skewed and sparse inputs exercise retained-block spill and drain handoffs.
func TestAndnotNEONSkewSpill(t *testing.T) {
	r := rand.New(rand.NewSource(43))
	small := andnotSortedUnique(r, 32, 60000)
	large := andnotSortedUnique(r, 2048, 60000)
	checkDifferenceBoth(t, "skew", small, large)

	a := make([]uint16, 512)
	for i := range a {
		a[i] = uint16(i * 3)
	}
	var b []uint16
	for i := 5; i < 512; i += 9 {
		b = append(b, uint16(i*3))
	}
	checkDifferenceBoth(t, "pickoff", a, b)
}

// Alternating disjoint runs exercise the fast-forward store/skip loop.
func TestAndnotNEONBlockRuns(t *testing.T) {
	for _, run := range []int{1, 2, 4, 8} {
		n := 1024
		a := make([]uint16, n)
		b := make([]uint16, n)
		for i := 0; i < n; i++ {
			blk, off := i/(8*run), i%(8*run)
			a[i] = uint16(blk*16*run + off)
			b[i] = uint16(blk*16*run + 8*run + off)
		}
		checkDifferenceBoth(t, "blockruns", a, b)
	}
}

func TestAndnotNEONRandom(t *testing.T) {
	r := rand.New(rand.NewSource(47))
	for i := 0; i < 100; i++ {
		la := r.Intn(3000)
		lb := r.Intn(3000)
		valRange := 512 + r.Intn(65000)
		a := andnotSortedUnique(r, la, valRange)
		b := andnotSortedUnique(r, lb, valRange)
		checkDifferenceBoth(t, "random", a, b)
	}
}

// Three-way aliasing must return empty without writing past the capacity.
func TestAndnotNEONSelfAlias(t *testing.T) {
	r := rand.New(rand.NewSource(53))
	for _, n := range []int{16, 65, 512} {
		a := andnotSortedUnique(r, n, 4*n)
		self := make([]uint16, n+16)
		for i := range self {
			self[i] = andnotCanary
		}
		copy(self, a)
		gn := difference(self[:n], self[:n], self[0:n:n])
		if gn != 0 {
			t.Fatalf("self n=%d: want empty, got len %d", n, gn)
		}
		for i := n; i < n+16; i++ {
			if self[i] != andnotCanary {
				t.Fatalf("self n=%d: overstore past cap at %d", n, i)
			}
		}
		checkDifference(t, "equal-distinct", a, append([]uint16(nil), a...))
	}

	rb := BitmapOf()
	for i := 0; i < 500; i++ {
		rb.Add(uint32(i * 5))
	}
	rb.AndNot(rb)
	if !rb.IsEmpty() {
		t.Fatalf("bitmap self AndNot: want empty, got %d", rb.GetCardinality())
	}
}

// Duplicate-bearing inputs get unspecified values but bounded stores and no panic.
func TestAndnotNEONDuplicateInputBounded(t *testing.T) {
	dup := func(vals ...uint16) []uint16 { return vals }
	cases := [][2][]uint16{
		{dup(1, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15),
			dup(1, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34)},
		{dup(100, 100, 100, 100, 101, 101, 101, 101, 101, 101, 102, 102, 102, 102, 103, 103),
			dup(100, 101, 101, 101, 101, 101, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102)},
	}
	for ci, c := range cases {
		for o := 0; o < 2; o++ {
			a, b := c[0], c[1]
			if o == 1 {
				a, b = b, a
			}
			m := len(a)
			backing := make([]uint16, m+16)
			for i := range backing {
				backing[i] = andnotCanary
			}
			gn := difference(a, b, backing[0:0:m])
			if gn < 0 || gn > m {
				t.Fatalf("dup case %d/%d: out of range length %d", ci, o, gn)
			}
			for i := m; i < m+16; i++ {
				if backing[i] != andnotCanary {
					t.Fatalf("dup case %d/%d: overstore past cap at %d", ci, o, i)
				}
			}
			inplace := make([]uint16, m+16)
			for i := range inplace {
				inplace[i] = andnotCanary
			}
			copy(inplace, a)
			gn = difference(inplace[:m], b, inplace[0:m:m])
			if gn < 0 || gn > m {
				t.Fatalf("dup case %d/%d inplace: out of range length %d", ci, o, gn)
			}
			for i := m; i < m+16; i++ {
				if inplace[i] != andnotCanary {
					t.Fatalf("dup case %d/%d inplace: overstore past cap at %d", ci, o, i)
				}
			}
		}
	}
}

// Every kernel exit hands back exact cursors, and a spilled block with its
// match mask when set2 runs out first.
func TestAndnotNEONKernelExits(t *testing.T) {
	seq := func(from, n int) []uint16 {
		s := make([]uint16, n)
		for i := range s {
			s[i] = uint16(from + i)
		}
		return s
	}
	cat := func(parts ...[]uint16) []uint16 {
		var s []uint16
		for _, p := range parts {
			s = append(s, p...)
		}
		return s
	}
	type exit struct {
		name                           string
		a, b                           []uint16
		out, pos1, pos2, spilled, mask int
	}
	cases := []exit{
		// equal maxima retire both blocks
		{"tie", seq(0, 16), seq(0, 16), 0, 16, 16, 0, 0},
		// mask carried across two set2 advances into the fast-forward retire,
		// then whole-block copies and a fresh block with a clear mask
		{"ffmask", cat(seq(10, 1), seq(20, 1), seq(30, 1), seq(40, 1), seq(50, 1), seq(60, 1), seq(70, 1), seq(80, 1), seq(90, 16)),
			cat(seq(0, 7), seq(10, 1), seq(11, 7), seq(20, 1), seq(100, 1), seq(110, 7)), 21, 24, 17, 0, 0},
		// set1 exhausted while set2 is held: the lane equal to the retired
		// maximum is consumed, the tail value above it survives
		{"exiteq", seq(0, 17), cat(seq(0, 8), seq(15, 1), seq(17, 7)), 7, 16, 9, 0, 0},
		// spill after an earlier block was emitted, same-start alias included
		{"spillafter", cat(seq(0, 8), seq(100, 8), seq(200, 8)), cat(seq(7, 15), seq(103, 1)), 7, 16, 16, 1, 1 << 3},
		// all-matched retire, a skipped set2 block, a later block accepted
		// with two matches, exit skips exactly those two lanes
		{"reloadexit", cat(seq(0, 8), seq(100, 9)), cat(seq(0, 24), seq(103, 1), seq(107, 7)), 6, 16, 26, 0, 0},
		// 65535 as the retired maximum through a copy and an equal retire
		{"top", seq(65520, 16), cat(seq(65496, 8), seq(65528, 8)), 8, 16, 16, 0, 0},
	}
	// both set2-exhaustion exits, through the skip and after a match, for
	// every set2 tail length
	a := cat(seq(100, 8), seq(200, 8))
	for tail := 0; tail <= 7; tail++ {
		cases = append(cases,
			exit{"ffB", a, cat(seq(0, 16), []uint16{103, 107, 200, 201, 202, 203, 204}[:tail]), 0, 8, 16, 1, 0},
			exit{"advB", a, cat(seq(0, 15), seq(103, 1), []uint16{107, 200, 201, 202, 203, 204, 205}[:tail]), 0, 8, 16, 1, 1 << 3})
	}
	// every reachable skip count on the set1-exhaust exit, every tail length
	for c := 1; c <= 7; c++ {
		for tail := 0; tail <= 7; tail++ {
			cases = append(cases, exit{"exitskip", seq(0, 16+tail), cat(seq(0, 8+c), seq(16, 8-c)), 8 - c, 16, 8 + c, 0, 0})
		}
	}
	for _, c := range cases {
		checkDifferenceBoth(t, c.name, c.a, c.b)
		buffer := make([]uint16, len(c.a))
		var spill [8]uint16
		out, pos1, pos2, spilled, mask := andnotKernelNEON(c.a, c.b, buffer, &uniqshuf[0], &spill)
		if out != c.out || pos1 != c.pos1 || pos2 != c.pos2 || spilled != c.spilled || mask != c.mask {
			t.Fatalf("%s: out=%d pos1=%d pos2=%d spilled=%d mask=%08b, want %d %d %d %d %08b",
				c.name, out, pos1, pos2, spilled, mask, c.out, c.pos1, c.pos2, c.spilled, c.mask)
		}
		if spilled != 0 {
			for i := 0; i < 8; i++ {
				if spill[i] != c.a[pos1-8+i] {
					t.Fatalf("%s: spill %v is not the retained block %v", c.name, spill, c.a[pos1-8:pos1])
				}
			}
		}
	}
}

func andnotSortedUnique(r *rand.Rand, n, valRange int) []uint16 {
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
