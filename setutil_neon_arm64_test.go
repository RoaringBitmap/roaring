//go:build arm64 && !gccgo && !appengine
// +build arm64,!gccgo,!appengine

package roaring

import (
	"math/rand"
	"reflect"
	"testing"
)

func refUnion(a, b []uint16) []uint16 {
	out := make([]uint16, len(a)+len(b))
	n := scalarMergeUnion(a, b, out)
	return out[:n]
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

func checkUnion(t *testing.T, a, b []uint16, label string) {
	t.Helper()
	want := refUnion(a, b)
	buffer := make([]uint16, len(a)+len(b))
	got := buffer[:union2by2(a, b, buffer)]
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("%s: len(a)=%d len(b)=%d: want %d elems, got %d\nwant %v\ngot  %v",
			label, len(a), len(b), len(want), len(got), want, got)
	}
}

func TestUnion2By2NEONAdversarial(t *testing.T) {
	for n := 8; n <= 2048; n *= 2 {
		identical := make([]uint16, n)
		evens := make([]uint16, n)
		odds := make([]uint16, n)
		low := make([]uint16, n)
		high := make([]uint16, n)
		for i := 0; i < n; i++ {
			identical[i] = uint16(3 * i)
			evens[i] = uint16(2 * i)
			odds[i] = uint16(2*i + 1)
			low[i] = uint16(i)
			high[i] = uint16(65535 - n + 1 + i)
		}
		checkUnion(t, identical, identical, "identical")
		checkUnion(t, evens, odds, "interleaved")
		checkUnion(t, odds, evens, "interleaved-swap")
		checkUnion(t, low, high, "disjoint-extremes")
		checkUnion(t, high, low, "disjoint-extremes-swap")
	}
}

func TestUnion2By2NEONRandom(t *testing.T) {
	r := rand.New(rand.NewSource(12345))
	for iter := 0; iter < 2000; iter++ {
		valRange := 256 + r.Intn(65280)
		a := genSortedUnique(r, r.Intn(5000), valRange)
		b := genSortedUnique(r, r.Intn(5000), valRange)
		checkUnion(t, a, b, "random")
	}
}

func checkKernel(t *testing.T, a, b []uint16, label string) {
	t.Helper()
	if len(a) < 8 || len(b) < 8 {
		checkUnion(t, a, b, label)
		return
	}
	want := refUnion(a, b)
	buffer := make([]uint16, 0, len(a)+len(b))
	n := unionNEON(a, b, buffer)
	got := buffer[:n]
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("%s: la=%d lb=%d want %v got %v", label, len(a), len(b), want, got)
	}
}

func TestUnion2By2NEONBoundaryMatrix(t *testing.T) {
	for _, sz := range [][2]int{{1, 1}, {2, 8}, {7, 8}, {8, 8}, {8, 9}, {8, 16}, {15, 16}, {16, 17}} {
		a := make([]uint16, sz[0])
		b := make([]uint16, sz[1])
		for i := range a {
			a[i] = uint16(2 * i)
		}
		for i := range b {
			b[i] = uint16(3*i + 1)
		}
		checkKernel(t, a, b, "boundary")
		checkKernel(t, b, a, "boundary-swap")
	}
}

// A duplicate straddling lanes 7 and 0, and blocks touching 0xFFFF.
func TestUnion2By2NEONLaneStraddleAndHighEnd(t *testing.T) {
	a := []uint16{0, 2, 4, 6, 7, 20, 22, 24}
	b := []uint16{1, 3, 5, 7, 21, 23, 25, 27}
	checkKernel(t, a, b, "straddle")
	checkKernel(t, b, a, "straddle-swap")

	high := make([]uint16, 8)
	for i := range high {
		high[i] = uint16(65528 + i)
	}
	checkKernel(t, high, high, "identical-high")
	low := make([]uint16, 8)
	for i := range low {
		low[i] = uint16(65520 + i)
	}
	checkKernel(t, low, high, "adjacent-high")
}

// Every block seam shares its boundary value, pinning the fast path's
// choice of >= over >: the duplicate must fall to the dedup chain.
func TestUnion2By2NEONEqualityStreak(t *testing.T) {
	for blocks := 2; blocks <= 64; blocks *= 2 {
		var a, b []uint16
		next := uint16(0)
		for i := 0; i < blocks; i++ {
			for l := 0; l < 8; l++ {
				a = append(a, next+uint16(l))
			}
			next += 7 // b's block starts at a's block's last value
			for l := 0; l < 8; l++ {
				b = append(b, next+uint16(l))
			}
			next += 7 // and a's next block starts at b's last value
		}
		checkKernel(t, a, b, "equality-streak")
		checkKernel(t, b, a, "equality-streak-swap")

		want := refUnion(a, b)
		shared := make([]uint16, len(a)+len(b))
		copy(shared[len(b):], a)
		n := unionNEON(shared[len(b):], b, shared)
		if !reflect.DeepEqual(want, shared[:n]) {
			t.Fatalf("equality-streak aliased, blocks=%d: want %d got %d elems",
				blocks, len(want), n)
		}
	}
}

// Tightest alias geometry: set1 at output offset len(set2) with a
// one-block set2, forcing the lookahead tail.
func TestUnion2By2NEONMinimumGapAlias(t *testing.T) {
	set2 := []uint16{0, 1, 2, 3, 4, 5, 6, 7}
	set1 := make([]uint16, 64)
	for i := range set1 {
		set1[i] = uint16(100 + i)
	}
	want := refUnion(set1, set2)
	shared := make([]uint16, len(set1)+len(set2))
	copy(shared[len(set2):], set1)
	n := unionNEON(shared[len(set2):], set2, shared)
	if !reflect.DeepEqual(want, shared[:n]) {
		t.Fatalf("minimum-gap alias: want %d elems got %d", len(want), n)
	}
}

// The kernel's 16-byte stores must never touch beyond len(set1)+len(set2).
func TestUnion2By2NEONBufferCanaries(t *testing.T) {
	r := rand.New(rand.NewSource(31337))
	for iter := 0; iter < 300; iter++ {
		a := genSortedUnique(r, 8+r.Intn(1000), 8192)
		b := genSortedUnique(r, 8+r.Intn(1000), 8192)
		need := len(a) + len(b)
		full := make([]uint16, need+16)
		for i := need; i < len(full); i++ {
			full[i] = 0xDEAD
		}
		n := unionNEON(a, b, full[0:0:need])
		if !reflect.DeepEqual(refUnion(a, b), full[:n]) {
			t.Fatalf("canary iter %d: wrong result", iter)
		}
		for i := need; i < len(full); i++ {
			if full[i] != 0xDEAD {
				t.Fatalf("canary iter %d: guard word %d clobbered (0x%04X)", iter, i, full[i])
			}
		}
	}
}

// lazyorArray passes a zero-length buffer with spare capacity.
func TestUnion2By2NEONZeroLenBuffer(t *testing.T) {
	r := rand.New(rand.NewSource(99))
	for iter := 0; iter < 200; iter++ {
		a := genSortedUnique(r, 8+r.Intn(500), 4096)
		b := genSortedUnique(r, 8+r.Intn(500), 4096)
		want := refUnion(a, b)
		buffer := make([]uint16, 0, len(a)+len(b))
		n := union2by2(a, b, buffer)
		got := buffer[:n]
		if !reflect.DeepEqual(want, got) {
			t.Fatalf("zero-len buffer: mismatch (want %d got %d elems)", len(want), n)
		}
	}
}

// iorArray's geometry: set1 lives in the upper region of the output array.
func TestUnion2By2NEONAliasedBuffer(t *testing.T) {
	r := rand.New(rand.NewSource(7))
	for iter := 0; iter < 500; iter++ {
		valRange := 256 + r.Intn(65280)
		set1 := genSortedUnique(r, 8+r.Intn(2000), valRange)
		set2 := genSortedUnique(r, 8+r.Intn(2000), valRange)
		want := refUnion(set1, set2)

		max := len(set1) + len(set2)
		shared := make([]uint16, max)
		copy(shared[len(set2):max], set1)
		got := shared[:union2by2(shared[len(set2):max], set2, shared)]
		if !reflect.DeepEqual(want, got) {
			t.Fatalf("aliased: len1=%d len2=%d: mismatch (want %d got %d elems)",
				len(set1), len(set2), len(want), len(got))
		}
	}
}

// iorArray's self-union geometry: set2 and the output share a backing array.
func TestUnion2By2NEONSelfAliasedBuffer(t *testing.T) {
	r := rand.New(rand.NewSource(11))
	for iter := 0; iter < 200; iter++ {
		valRange := 256 + r.Intn(65280)
		set := genSortedUnique(r, 8+r.Intn(2000), valRange)

		n := len(set)
		shared := make([]uint16, 2*n)
		copy(shared, set)
		copy(shared[n:], set)
		got := shared[:union2by2(shared[n:], shared[:n], shared)]
		if !reflect.DeepEqual(set, got) {
			t.Fatalf("self-aliased: n=%d: mismatch (got %d elems)", n, len(got))
		}
	}
}
