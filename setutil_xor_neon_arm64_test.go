//go:build arm64 && !gccgo && !appengine
// +build arm64,!gccgo,!appengine

package roaring

import (
	"fmt"
	"math/bits"
	"math/rand"
	"sort"
	"testing"
)

const xrCanary = 0xA5A5

// xrForced skips the dispatch threshold; a direct call keeps the wrapper allocation-free.
func xrForced(set1, set2, buffer []uint16) int {
	if len(set1) < 16 || len(set2) < 16 || cap(buffer) < len(set1)+len(set2) {
		return localexclusiveUnion2by2(set1, set2, buffer)
	}
	buffer = buffer[:cap(buffer)]
	outLen, pos1, pos2 := xorKernelNEON(set1, set2, buffer, &uniqshuf[0])
	return outLen + localexclusiveUnion2by2(set1[pos1:], set2[pos2:], buffer[outLen:])
}

// xrOracle is a toggle bitset that shares no code with either merge, valid for duplicate-free inputs.
func xrOracle(a, b []uint16) []uint16 {
	var bs [1024]uint64
	for _, v := range a {
		bs[v>>6] ^= 1 << (v & 63)
	}
	for _, v := range b {
		bs[v>>6] ^= 1 << (v & 63)
	}
	out := make([]uint16, 0, len(a)+len(b))
	for w := 0; w < 1024; w++ {
		x := bs[w]
		for x != 0 {
			out = append(out, uint16(w*64+bits.TrailingZeros64(x)))
			x &= x - 1
		}
	}
	return out
}

func xrScalar(a, b []uint16) []uint16 {
	buf := make([]uint16, 0, len(a)+len(b))
	n := localexclusiveUnion2by2(a, b, buf)
	return buf[:cap(buf)][:n]
}

func xrEqual(x, y []uint16) bool {
	if len(x) != len(y) {
		return false
	}
	for i := range x {
		if x[i] != y[i] {
			return false
		}
	}
	return true
}

// xrRun checks both arms in both orientations on an exact-capacity buffer with canaries.
func xrRun(t *testing.T, name string, a, b []uint16, dupFree bool) {
	t.Helper()
	arms := []struct {
		n string
		f func(x, y, buf []uint16) int
	}{
		{"dispatch", exclusiveUnion2by2},
		{"forced", xrForced},
	}
	for _, or := range []struct {
		n    string
		x, y []uint16
	}{{"ab", a, b}, {"ba", b, a}} {
		want := xrScalar(or.x, or.y)
		if dupFree {
			if got := xrOracle(or.x, or.y); !xrEqual(want, got) {
				t.Fatalf("%s/%s: scalar and oracle disagree: scalar %v oracle %v", name, or.n, want, got)
			}
		}
		for _, arm := range arms {
			cx := append([]uint16(nil), or.x...)
			cy := append([]uint16(nil), or.y...)
			n := len(or.x) + len(or.y)
			back := make([]uint16, n+16)
			for i := range back {
				back[i] = xrCanary
			}
			got := arm.f(or.x, or.y, back[0:0:n])
			if got != len(want) {
				t.Fatalf("%s/%s/%s: length %d, want %d", name, or.n, arm.n, got, len(want))
			}
			if !xrEqual(back[:got], want) {
				t.Fatalf("%s/%s/%s: got %v, want %v", name, or.n, arm.n, back[:got], want)
			}
			for i := n; i < n+16; i++ {
				if back[i] != xrCanary {
					t.Fatalf("%s/%s/%s: canary %d clobbered with %d", name, or.n, arm.n, i-n, back[i])
				}
			}
			if !xrEqual(cx, or.x) || !xrEqual(cy, or.y) {
				t.Fatalf("%s/%s/%s: input mutated", name, or.n, arm.n)
			}
		}
	}
}

func xrSeq(start, step, n int) []uint16 {
	out := make([]uint16, n)
	for i := range out {
		out[i] = uint16(start + step*i)
	}
	return out
}

// One shared value at every merged lane 0..30, across the 8-lane boundaries.
func TestXorNEONEqualPairEveryLane(t *testing.T) {
	for p := 0; p <= 30; p++ {
		a := []uint16{uint16(p)}
		b := []uint16{uint16(p)}
		for v := 0; v <= 30; v++ {
			if v == p {
				continue
			}
			if len(a) <= len(b) {
				a = append(a, uint16(v))
			} else {
				b = append(b, uint16(v))
			}
		}
		sort.Slice(a, func(i, j int) bool { return a[i] < a[j] })
		sort.Slice(b, func(i, j int) bool { return b[i] < b[j] })
		xrRun(t, fmt.Sprintf("pair@%d", p), a, b, true)
	}
}

// xrCut builds a pair whose first partition consumes T merged lanes; dupCut puts the cutoff a[15] in set2 too.
func xrCut(T int, dupCut bool) ([]uint16, []uint16) {
	a := xrSeq(0, 2, 16) // 0..30, maximum 30
	nb := T - 16
	b := make([]uint16, 0, 16)
	if dupCut {
		for i := 0; i < nb-1; i++ {
			b = append(b, uint16(2*i+1))
		}
		b = append(b, 30)
	} else {
		for i := 0; i < nb; i++ {
			b = append(b, uint16(2*i+1))
		}
	}
	for v := 31; len(b) < 16; v += 2 {
		b = append(b, uint16(v))
	}
	return a, b
}

// T from 16 (a whole-window copy) to 32; a unique cutoff leaves a singleton right before the invalid lanes.
// Equal window maxima give T = 32 with both cursors advancing a full window.
func TestXorNEONCutoff(t *testing.T) {
	for T := 16; T <= 32; T++ {
		if T <= 31 {
			a, b := xrCut(T, false)
			xrRun(t, fmt.Sprintf("cut%d", T), a, b, true)
			// Same cutoff followed by further partitions rather than the tail.
			a2 := append(append([]uint16(nil), a...), xrSeq(200, 2, 40)...)
			b2 := append(append([]uint16(nil), b...), xrSeq(201, 2, 40)...)
			xrRun(t, fmt.Sprintf("cut%d+tail", T), a2, b2, true)
		}
		if T >= 18 {
			a, b := xrCut(T, true)
			xrRun(t, fmt.Sprintf("cut%ddup", T), a, b, true)
			a2 := append(append([]uint16(nil), a...), xrSeq(200, 2, 40)...)
			b2 := append(append([]uint16(nil), b...), xrSeq(201, 2, 40)...)
			xrRun(t, fmt.Sprintf("cut%ddup+tail", T), a2, b2, true)
		}
	}
	for k := 1; k <= 16; k++ {
		a := make([]uint16, 0, 16)
		b := make([]uint16, 0, 16)
		for i := 0; i < 16; i++ {
			a = append(a, uint16(4*i))
			b = append(b, uint16(4*i+2))
		}
		a[15] = 100
		b[15] = 100
		for i := 0; i < k-1 && i < 15; i++ {
			b[i] = a[i]
		}
		sort.Slice(b, func(i, j int) bool { return b[i] < b[j] })
		xrRun(t, fmt.Sprintf("maxima%d", k), a, b, true)
	}
}

// xrExit pins the kernel's own accounting so a fixture cannot silently stop reaching its path.
func xrExit(t *testing.T, name string, a, b []uint16, outLen, pos1, pos2 int) {
	t.Helper()
	buf := make([]uint16, len(a)+len(b))
	gotOut, gotP1, gotP2 := xorKernelNEON(a, b, buf, &uniqshuf[0])
	if gotOut != outLen || gotP1 != pos1 || gotP2 != pos2 {
		t.Fatalf("%s: kernel returned (%d,%d,%d), want (%d,%d,%d)",
			name, gotOut, gotP1, gotP2, outLen, pos1, pos2)
	}
}

// An identical window at entry, after each copy path, after a general partition, before the tail.
func TestXorNEONAnnihilation(t *testing.T) {
	block := func(base int) []uint16 { return xrSeq(base, 1, 16) }

	// At entry.
	a := append(block(0), xrSeq(100, 2, 20)...)
	b := append(block(0), xrSeq(101, 2, 20)...)
	xrRun(t, "annih/entry", a, b, true)

	// After a 16-lane copy.
	a = append(xrSeq(0, 1, 16), block(1000)...)
	a = append(a, xrSeq(2000, 2, 5)...)
	b = append(block(1000), xrSeq(2001, 2, 5)...)
	xrRun(t, "annih/aftercopy16", a, b, true)
	xrExit(t, "annih/aftercopy16", a, b, 16, 32, 16)

	// After an 8-lane copy.
	a = append(xrSeq(0, 1, 8), block(1000)...)
	a = append(a, xrSeq(2000, 2, 5)...)
	b = append(block(1000), xrSeq(2001, 2, 5)...)
	xrRun(t, "annih/aftercopy8", a, b, true)
	xrExit(t, "annih/aftercopy8", a, b, 8, 24, 16)

	// After a general partition; the shared maximum keeps the next block aligned.
	head1 := append(xrSeq(0, 2, 15), 31)
	head2 := append(xrSeq(1, 2, 15), 31)
	a = append(append([]uint16(nil), head1...), block(1000)...)
	a = append(a, xrSeq(2000, 2, 20)...)
	b = append(append([]uint16(nil), head2...), block(1000)...)
	b = append(b, xrSeq(2001, 2, 20)...)
	xrRun(t, "annih/aftergeneral", a, b, true)

	// Immediately before the scalar tail.
	a = append(append([]uint16(nil), head1...), block(1000)...)
	a = append(a, xrSeq(2000, 2, 5)...)
	b = append(append([]uint16(nil), head2...), block(1000)...)
	b = append(b, xrSeq(2001, 2, 5)...)
	xrRun(t, "annih/beforetail", a, b, true)
	xrExit(t, "annih/beforetail", a, b, 30, 32, 32)

	// Back to back annihilations over a long identical run.
	long := xrSeq(0, 1, 200)
	xrRun(t, "annih/long", long, append([]uint16(nil), long...), true)
	xrExit(t, "annih/long", long, long, 0, 192, 192)
}

// One edit in a 200-element identical run sends a lone survivor through the whole machinery.
func TestXorNEONIdenticalRunEdits(t *testing.T) {
	base := xrSeq(0, 2, 200)
	for _, p := range []int{0, 1, 7, 8, 15, 16, 17, 23, 24, 31, 32, 63, 64, 99, 100, 150, 190, 198, 199} {
		ins := append([]uint16(nil), base[:p+1]...)
		ins = append(ins, base[p]+1)
		ins = append(ins, base[p+1:]...)
		xrRun(t, fmt.Sprintf("insert@%d", p), ins, base, true)

		del := append([]uint16(nil), base[:p]...)
		del = append(del, base[p+1:]...)
		xrRun(t, fmt.Sprintf("delete@%d", p), del, base, true)

		sub := append([]uint16(nil), base...)
		sub[p] = sub[p] + 1
		xrRun(t, fmt.Sprintf("substitute@%d", p), sub, base, true)
	}
}

// Copy gates at boundary equality, where a non-strict compare would emit a value that must cancel.
func TestXorNEONStrictGates(t *testing.T) {
	for _, n := range []int{16, 24, 32, 48} {
		a := xrSeq(0, 1, n)
		b := xrSeq(n-1, 1, n)
		xrRun(t, fmt.Sprintf("gate16/eq/%d", n), a, b, true)
		b = xrSeq(n+100, 1, n)
		xrRun(t, fmt.Sprintf("gate16/lt/%d", n), a, b, true)
		a = xrSeq(0, 1, n)
		b = xrSeq(7, 1, n)
		xrRun(t, fmt.Sprintf("gate8/eq/%d", n), a, b, true)
		b = xrSeq(8, 1, n)
		xrRun(t, fmt.Sprintf("gate8/lt/%d", n), a, b, true)
	}
}

func xrRuns(r, n int) ([]uint16, []uint16) {
	a := make([]uint16, 0, n)
	b := make([]uint16, 0, n)
	v := 0
	for len(a) < n || len(b) < n {
		for i := 0; i < r && len(a) < n; i++ {
			a = append(a, uint16(v))
			v++
		}
		for i := 0; i < r && len(b) < n; i++ {
			b = append(b, uint16(v))
			v++
		}
	}
	return a, b
}

// Input shapes through both arms in both orientations: ownership runs, every length pair
// up to 40, extreme values, and random picks from four universes.
func TestXorNEONShapes(t *testing.T) {
	t.Run("runs", func(t *testing.T) {
		for _, r := range []int{4, 7, 8, 9, 15, 16, 17, 32} {
			for _, n := range []int{16, 33, 64, 129, 256} {
				a, b := xrRuns(r, n)
				xrRun(t, fmt.Sprintf("runs%d/%d", r, n), a, b, true)
			}
		}
	})
	t.Run("lengths", func(t *testing.T) {
		for la := 0; la <= 40; la++ {
			for lb := 0; lb <= 40; lb++ {
				a := xrSeq(0, 2, la)
				b := xrSeq(0, 3, lb)
				xrRun(t, fmt.Sprintf("len%dx%d", la, lb), a, b, true)
			}
		}
	})
	t.Run("extremes", func(t *testing.T) {
		edges := []uint16{0, 1, 0xFFFE, 0xFFFF}
		mid := xrSeq(1000, 2, 40)
		for _, e := range edges {
			a := append([]uint16{e}, mid...)
			sort.Slice(a, func(i, j int) bool { return a[i] < a[j] })
			b := append([]uint16(nil), mid...)
			for i := range b {
				b[i]++
			}
			xrRun(t, fmt.Sprintf("edge%d/one", e), a, b, true)

			b2 := append([]uint16{e}, b...)
			sort.Slice(b2, func(i, j int) bool { return b2[i] < b2[j] })
			xrRun(t, fmt.Sprintf("edge%d/both", e), a, b2, true)
		}
		a := append([]uint16{0, 1}, xrSeq(0xFFC0, 1, 64)...)
		b := append([]uint16{0}, xrSeq(0xFFC1, 2, 32)...)
		xrRun(t, "edge/all", a, b, true)

		// Singleton cutoff 0xFFFE with 0xFFFF as the invalid lane behind it.
		xrRun(t, "edge/cutoffFFFE", xrSeq(0xFFE0, 2, 16), xrSeq(0xFFE1, 2, 16), true)

		// 0xFFFF cancels through the general path, not annihilation.
		xrRun(t, "edge/sharedFFFF",
			append(xrSeq(0xFFC0, 2, 15), 0xFFFF),
			append(xrSeq(0xFFC1, 2, 15), 0xFFFF), true)
	})
	t.Run("random", func(t *testing.T) {
		rng := rand.New(rand.NewSource(20260905))
		universes := []int{64, 512, 4096, 0x10000}
		for i := 0; i < 5000; i++ {
			u := universes[i%len(universes)]
			la := rng.Intn(300)
			lb := rng.Intn(300)
			if la > u {
				la = u
			}
			if lb > u {
				lb = u
			}
			a := xrPick(rng, u, la)
			b := xrPick(rng, u, lb)
			xrRun(t, fmt.Sprintf("rand%d", i), a, b, true)
		}
	})
}

// Different poison past both inputs must not change the output.
func TestXorNEONNoOverRead(t *testing.T) {
	poisons := []uint16{0, 0xFFFF, 0x5555, 0xAAAA}
	rng := rand.New(rand.NewSource(31337))
	for i := 0; i < 600; i++ {
		la := 16 + rng.Intn(80)
		lb := 16 + rng.Intn(80)
		a := xrPick(rng, 0x8000, la)
		b := xrPick(rng, 0x8000, lb)
		n := la + lb
		var first []uint16
		for _, p := range poisons {
			ba := make([]uint16, la+32)
			bb := make([]uint16, lb+32)
			for j := range ba {
				ba[j] = p
			}
			for j := range bb {
				bb[j] = p
			}
			copy(ba, a)
			copy(bb, b)
			out := make([]uint16, n)
			got := xrForced(ba[:la], bb[:lb], out[:0:n])
			if first == nil {
				first = append([]uint16(nil), out[:got]...)
				continue
			}
			if !xrEqual(out[:got], first) {
				t.Fatalf("overread%d: poison %#x changed the result", i, p)
			}
		}
		if want := xrScalar(a, b); !xrEqual(first, want) {
			t.Fatalf("overread%d: result %v, want %v", i, first, want)
		}
	}
}

func TestXorNEONCombined4096(t *testing.T) {
	rng := rand.New(rand.NewSource(9001))
	for _, split := range [][2]int{{2048, 2048}, {3072, 1024}, {4000, 96}, {4032, 64}, {4096, 0}, {0, 4096}, {64, 4032}} {
		for shape := 0; shape < 3; shape++ {
			var a, b []uint16
			switch shape {
			case 0: // low overlap
				a = xrPick(rng, 0x10000, split[0])
				b = xrPick(rng, 0x10000, split[1])
			case 1: // dense adjacent blocks
				a = xrSeq(0, 1, split[0])
				b = xrSeq(1, 1, split[1])
			case 2: // shared prefix
				n := split[0]
				if split[1] < n {
					n = split[1]
				}
				a = xrSeq(0, 2, split[0])
				b = append(append([]uint16(nil), a[:n]...), xrSeq(20000, 2, split[1]-n)...)
			}
			xrRun(t, fmt.Sprintf("c4096/%d_%d/s%d", split[0], split[1], shape), a, b, true)
		}
	}
}

func xrPick(rng *rand.Rand, universe, n int) []uint16 {
	if n == 0 {
		return nil
	}
	seen := make(map[uint16]bool, n)
	out := make([]uint16, 0, n)
	for len(out) < n {
		v := uint16(rng.Intn(universe))
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// Duplicate-bearing inputs get unspecified values but stores stay within capacity.
func TestXorNEONDuplicateBearing(t *testing.T) {
	rng := rand.New(rand.NewSource(4242))
	for i := 0; i < 400; i++ {
		la := 16 + rng.Intn(120)
		lb := 16 + rng.Intn(120)
		a := make([]uint16, la)
		b := make([]uint16, lb)
		v := 0
		for j := range a {
			if rng.Intn(3) != 0 {
				v += 1 + rng.Intn(3)
			}
			a[j] = uint16(v)
		}
		v = 0
		for j := range b {
			if rng.Intn(3) != 0 {
				v += 1 + rng.Intn(3)
			}
			b[j] = uint16(v)
		}
		n := la + lb
		back := make([]uint16, n+16)
		for j := range back {
			back[j] = xrCanary
		}
		got := xrForced(a, b, back[0:0:n])
		if got < 0 || got > n {
			t.Fatalf("dup%d: length %d out of [0,%d]", i, got, n)
		}
		for j := n; j < n+16; j++ {
			if back[j] != xrCanary {
				t.Fatalf("dup%d: canary %d clobbered", i, j-n)
			}
		}
	}
}

// The raw kernel stops under 16 unread on a side and emits exactly the xor of the consumed prefixes.
func TestXorNEONKernelExit(t *testing.T) {
	rng := rand.New(rand.NewSource(777))
	for i := 0; i < 2000; i++ {
		la := 16 + rng.Intn(400)
		lb := 16 + rng.Intn(400)
		a := xrPick(rng, 0x10000, la)
		b := xrPick(rng, 0x10000, lb)
		n := len(a) + len(b)
		back := make([]uint16, n+16)
		for j := range back {
			back[j] = xrCanary
		}
		outLen, pos1, pos2 := xorKernelNEON(a, b, back[:n:n], &uniqshuf[0])
		if pos1 < 0 || pos1 > len(a) || pos2 < 0 || pos2 > len(b) {
			t.Fatalf("exit%d: positions %d,%d out of range %d,%d", i, pos1, pos2, len(a), len(b))
		}
		if len(a)-pos1 >= 16 && len(b)-pos2 >= 16 {
			t.Fatalf("exit%d: stopped early with %d,%d unread", i, len(a)-pos1, len(b)-pos2)
		}
		if outLen > pos1+pos2 {
			t.Fatalf("exit%d: emitted %d from %d consumed", i, outLen, pos1+pos2)
		}
		want := xrScalar(a[:pos1], b[:pos2])
		if !xrEqual(back[:outLen], want) {
			t.Fatalf("exit%d: prefix %v, want %v", i, back[:outLen], want)
		}
		for j := n; j < n+16; j++ {
			if back[j] != xrCanary {
				t.Fatalf("exit%d: canary %d clobbered", i, j-n)
			}
		}
	}
}

// The kernel is reached only from the array/array path, from the threshold up.
func TestXorNEONDispatch(t *testing.T) {
	sizes := []int{neonXorThreshold - 1, neonXorThreshold, neonXorThreshold + 1}
	for _, la := range sizes {
		for _, lb := range sizes {
			a := xrSeq(0, 2, la)
			b := xrSeq(1, 2, lb)
			xrRun(t, fmt.Sprintf("bound%dx%d/disjoint", la, lb), a, b, true)
			b = xrSeq(0, 2, lb)
			xrRun(t, fmt.Sprintf("bound%dx%d/prefix", la, lb), a, b, true)
		}
	}

	a := New()
	for i := 0; i < 100; i++ {
		a.Add(uint32(3 * i))
	}
	if got := Xor(a, a); !got.IsEmpty() {
		t.Fatalf("Xor(a,a) cardinality %d, want 0", got.GetCardinality())
	}

	b := New()
	for i := 0; i < 100; i++ {
		b.Add(uint32(2 * i))
	}
	want := xrScalar(xrBitmapValues(a), xrBitmapValues(b))
	if got := xrBitmapValues(Xor(a, b)); !xrEqual(got, want) {
		t.Fatalf("array/array xor mismatch")
	}

	// Run container on one side.
	r := New()
	r.AddRange(0, 300)
	r.RunOptimize()
	want = xrScalar(xrBitmapValues(a), xrBitmapValues(r))
	if got := xrBitmapValues(Xor(a, r)); !xrEqual(got, want) {
		t.Fatalf("array/run xor mismatch")
	}
	if got := xrBitmapValues(Xor(r, a)); !xrEqual(got, want) {
		t.Fatalf("run/array xor mismatch")
	}

	// Bitmap container on one side.
	bc := New()
	for i := 0; i < 9000; i++ {
		bc.Add(uint32(2 * i))
	}
	want = xrScalar(xrBitmapValues(a), xrBitmapValues(bc))
	if got := xrBitmapValues(Xor(a, bc)); !xrEqual(got, want) {
		t.Fatalf("array/bitmap xor mismatch")
	}

	// Above 4096 combined the caller builds a bitmap instead.
	big1 := New()
	big2 := New()
	for i := 0; i < 3000; i++ {
		big1.Add(uint32(2 * i))
		big2.Add(uint32(2*i + 1))
	}
	if got := Xor(big1, big2).GetCardinality(); got != 6000 {
		t.Fatalf("wide xor cardinality %d, want 6000", got)
	}
}

func xrBitmapValues(b *Bitmap) []uint16 {
	out := make([]uint16, 0, b.GetCardinality())
	it := b.Iterator()
	for it.HasNext() {
		out = append(out, uint16(it.Next()))
	}
	return out
}
