//go:build arm64 && !gccgo && !appengine
// +build arm64,!gccgo,!appengine

package roaring

import (
	"fmt"
	"math/rand"
	"slices"
	"testing"
)

func unionSortedUnique(values []uint16) []uint16 {
	values = slices.Clone(values)
	slices.Sort(values)
	return slices.Compact(values)
}

func unionReference(a, b []uint16) []uint16 {
	return unionSortedUnique(append(slices.Clone(a), b...))
}

func unionTestPair(shape string, na, nb int, seed int64) (a, b []uint16) {
	r := rand.New(rand.NewSource(seed))
	random := func(n, limit int) []uint16 {
		values := make(map[uint16]struct{}, n)
		for len(values) < n {
			values[uint16(r.Intn(limit))] = struct{}{}
		}
		out := make([]uint16, 0, n)
		for value := range values {
			out = append(out, value)
		}
		slices.Sort(out)
		return out
	}
	switch shape {
	case "sparse":
		return random(na, 65536), random(nb, 65536)
	case "dense":
		return random(na, 3*(na+nb)/2+1), random(nb, 3*(na+nb)/2+1)
	case "identical":
		pool := random(max(na, nb), 65536)
		return slices.Clone(pool[:na]), slices.Clone(pool[:nb])
	case "sharedprefix":
		pool := random(max(na, nb)+min(na, nb)/8+1, 65536)
		return slices.Clone(pool[:na]), slices.Clone(pool[len(pool)-nb:])
	case "interleaved", "runs":
		run := 1
		if shape == "runs" {
			run = 4 + r.Intn(29)
		}
		value := r.Intn(1024)
		for len(a) < na || len(b) < nb {
			for i := 0; i < run && len(a) < na; i++ {
				a = append(a, uint16(value))
				value++
			}
			for i := 0; i < run && len(b) < nb; i++ {
				b = append(b, uint16(value))
				value++
			}
		}
		return a, b
	default:
		panic("unknown union test shape")
	}
}

func checkUnionBuffers(t *testing.T, a, b []uint16) {
	t.Helper()
	want := unionReference(a, b)
	originalA, originalB := slices.Clone(a), slices.Clone(b)
	for _, relocated := range []bool{false, true} {
		const guard = 8
		need := len(a) + len(b)
		backing := make([]uint16, need+2*guard)
		for i := range backing {
			backing[i] = 0xdead
		}
		out := backing[guard : guard+need : guard+need]
		input := a
		if relocated {
			copy(out[len(b):], a)
			input = out[len(b):]
		}
		n := union2by2(input, b, out[:0])
		if n < 0 || n > need || !slices.Equal(out[:n], want) {
			t.Fatalf("%dx%d relocated=%t: got cardinality %d, want %d or wrong elements", len(a), len(b), relocated, n, len(want))
		}
		for i := 0; i < guard; i++ {
			if backing[i] != 0xdead || backing[guard+need+i] != 0xdead {
				t.Fatalf("%dx%d relocated=%t: output overwrote canary", len(a), len(b), relocated)
			}
		}
		if !slices.Equal(a, originalA) || !slices.Equal(b, originalB) {
			t.Fatal("union modified a non-output input")
		}
	}
}

func TestUnion2By2Boundaries(t *testing.T) {
	pairs := [][2]int{
		{0, 0}, {0, 64}, {64, 0}, {1, 128}, {15, 128}, {16, 128},
		{31, 4096}, {32, 4096}, {4096, 32}, {32, 95}, {32, 96}, {32, 97},
		{95, 32}, {96, 32}, {97, 32}, {48, 79}, {48, 80}, {48, 81},
		{63, 63}, {63, 64}, {64, 63}, {64, 64}, {64, 65}, {65, 64},
		{79, 80}, {95, 96}, {127, 128}, {255, 256}, {4096, 4096},
	}
	for _, shape := range []string{"sparse", "dense", "identical", "sharedprefix", "interleaved", "runs"} {
		for _, pair := range pairs {
			t.Run(fmt.Sprintf("%s/%dx%d", shape, pair[0], pair[1]), func(t *testing.T) {
				a, b := unionTestPair(shape, pair[0], pair[1], 42)
				checkUnionBuffers(t, a, b)
				if len(a) > 0 && len(b) > 0 {
					a[len(a)-1], b[len(b)-1] = 0xffff, 0xffff
					checkUnionBuffers(t, a, b)
				}
			})
		}
	}
}

func TestUnionArrayContainerPaths(t *testing.T) {
	for _, pair := range [][2]int{{32, 95}, {32, 96}, {48, 80}, {63, 64}, {64, 64}, {65, 255}} {
		for _, shape := range []string{"sparse", "sharedprefix", "runs"} {
			a, b := unionTestPair(shape, pair[0], pair[1], 73)
			want := unionReference(a, b)
			fresh := func(values []uint16, capacity int) *arrayContainer {
				return &arrayContainer{content: append(make([]uint16, 0, capacity), values...)}
			}
			for _, capacity := range []int{len(a), 2 * (len(a) + len(b))} {
				left, right := fresh(a, capacity), fresh(b, len(b))
				for _, result := range []container{left.orArray(right), left.lazyorArray(right), left.iorArray(right)} {
					if !slices.Equal(result.(*arrayContainer).content, want) {
						t.Fatalf("%s/%dx%d capacity=%d: incorrect container union", shape, len(a), len(b), capacity)
					}
				}
				if !slices.Equal(right.content, b) {
					t.Fatal("union modified the other container")
				}
				self := fresh(a, capacity)
				if !slices.Equal(self.iorArray(self).(*arrayContainer).content, a) {
					t.Fatal("incorrect self union")
				}
			}
		}
	}
}

func TestUnionPartitionPrefixes(t *testing.T) {
	for _, size := range []int{32, 64, 127, 4096} {
		for _, shape := range []string{"sparse", "identical", "interleaved", "sharedprefix"} {
			a, b := unionTestPair(shape, size, size, 53)
			backing := make([]uint16, len(a)+len(b)+2)
			backing[0], backing[len(backing)-1] = 0xdead, 0xdead
			out := backing[1 : len(backing)-1 : len(backing)-1]
			n, posA, posB := unionPartKernelNEON(a, b, out, &uniqshuf[0])
			if n <= 0 || n > len(out) || posA < 0 || posA > len(a) || posB < 0 || posB > len(b) {
				t.Fatalf("%s/%d: invalid kernel accounting: %d, %d, %d", shape, size, n, posA, posB)
			}
			if !slices.Equal(out[:n], unionReference(a[:posA], b[:posB])) {
				t.Fatalf("%s/%d: output differs from consumed prefixes", shape, size)
			}
			if posA < len(a) && a[posA] <= out[n-1] || posB < len(b) && b[posB] <= out[n-1] {
				t.Fatalf("%s/%d: unconsumed input does not follow output", shape, size)
			}
			if backing[0] != 0xdead || backing[len(backing)-1] != 0xdead {
				t.Fatal("kernel overwrote output canary")
			}
		}
	}
}

func TestUnionWindowTies(t *testing.T) {
	for _, run := range []int{8, 16} {
		var a, b []uint16
		value := 0
		for window := 0; window < 16; window++ {
			for i := 0; i < run; i++ {
				a = append(a, uint16(value))
				value++
			}
			value-- // Equality at a window boundary must reach deduplication.
			for i := 0; i < run; i++ {
				b = append(b, uint16(value))
				value++
			}
			value--
		}
		checkUnionBuffers(t, a, b)
		checkUnionBuffers(t, b, a)
		checkUnionBuffers(t, a[:48], b[:80])
	}
}

func TestUnionMixedWindowOwnership(t *testing.T) {
	r := rand.New(rand.NewSource(8080))
	for trial := 0; trial < 64; trial++ {
		var a, b []uint16
		for value := 0; value < 2048; {
			run, owner := 1+r.Intn(40), r.Intn(2)
			for i := 0; i < run; i++ {
				shared := r.Intn(8) == 0
				if owner == 0 || shared {
					a = append(a, uint16(value))
				}
				if owner == 1 || shared {
					b = append(b, uint16(value))
				}
				value++
			}
		}
		checkUnionBuffers(t, a, b)
	}
}

func TestUnionBitmapCopyOnWrite(t *testing.T) {
	a, b := unionTestPair("sparse", 64, 96, 91)
	left, right, want := NewBitmap(), NewBitmap(), NewBitmap()
	for _, value := range a {
		left.Add(uint32(value))
		want.Add(uint32(value))
	}
	for _, value := range b {
		right.Add(uint32(value))
		want.Add(uint32(value))
	}
	originalLeft, originalRight := left.Clone(), right.Clone()
	for _, leftCOW := range []bool{false, true} {
		for _, rightCOW := range []bool{false, true} {
			left.SetCopyOnWrite(leftCOW)
			right.SetCopyOnWrite(rightCOW)
			copyLeft, copyRight := left.Clone(), right.Clone()
			result := Or(copyLeft, copyRight)
			copyLeft.Or(copyRight)
			if !result.Equals(want) || !copyLeft.Equals(want) || !FastOr(left, right).Equals(want) {
				t.Fatal("incorrect bitmap union")
			}
			if !left.Equals(originalLeft) || !right.Equals(originalRight) || !copyRight.Equals(originalRight) {
				t.Fatal("union modified a shared input")
			}
		}
	}
}

func FuzzUnion2By2(f *testing.F) {
	for _, pair := range [][2]int{{0, 0}, {32, 96}, {64, 64}, {128, 255}, {4096, 32}} {
		a, b := unionTestPair("dense", pair[0], pair[1], 123)
		encode := func(values []uint16) []byte {
			out := make([]byte, 2*len(values))
			for i, value := range values {
				out[2*i], out[2*i+1] = byte(value), byte(value>>8)
			}
			return out
		}
		f.Add(encode(a), encode(b))
	}
	f.Fuzz(func(t *testing.T, aBytes, bBytes []byte) {
		decode := func(data []byte) []uint16 {
			data = data[:min(len(data), 8192)]
			values := make([]uint16, len(data)/2)
			for i := range values {
				values[i] = uint16(data[2*i]) | uint16(data[2*i+1])<<8
			}
			return unionSortedUnique(values)
		}
		checkUnionBuffers(t, decode(aBytes), decode(bBytes))
	})
}

var unionBenchmarkSize int

func BenchmarkUnion2By2(b *testing.B) {
	for _, shape := range []string{"sparse", "dense", "interleaved", "runs", "identical", "sharedprefix"} {
		for _, size := range [][2]int{{32, 32}, {64, 64}, {256, 256}, {4096, 4096}, {32, 96}} {
			b.Run(fmt.Sprintf("%s/%dx%d", shape, size[0], size[1]), func(b *testing.B) {
				var inputs [64][2][]uint16
				for i := range inputs {
					inputs[i][0], inputs[i][1] = unionTestPair(shape, size[0], size[1], int64(i+100))
					for side, values := range inputs[i] {
						if len(values) != size[side] || !slices.Equal(values, unionSortedUnique(values)) {
							b.Fatal("benchmark generator returned wrong length or non-unique sorted input")
						}
					}
				}
				out := make([]uint16, size[0]+size[1])
				b.Run("dispatch", func(b *testing.B) {
					b.ReportAllocs()
					for i := 0; i < b.N; i++ {
						pair := inputs[i&63]
						unionBenchmarkSize = union2by2(pair[0], pair[1], out)
					}
				})
				b.Run("scalar", func(b *testing.B) {
					b.ReportAllocs()
					for i := 0; i < b.N; i++ {
						pair := inputs[i&63]
						unionBenchmarkSize = union2by2scalar(pair[0], pair[1], out)
					}
				})
			})
		}
	}
}
