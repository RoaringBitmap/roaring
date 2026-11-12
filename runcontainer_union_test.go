package roaring

import (
	"fmt"
	"math/rand"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

// unionValues is the reference: the sorted union of the values in a and b.
func unionValues(a, b []uint16) []uint16 {
	out := slices.Concat(a, b)
	slices.Sort(out)
	return slices.Compact(out)
}

func containerValues(c container) []uint16 {
	var out []uint16
	for it := c.getShortIterator(); it.hasNext(); {
		out = append(out, it.next())
	}
	return out
}

// randomRuns returns a run container of up to maxRuns runs starting at or
// after lo, each 1..maxLen long, separated by gaps of 1..maxGap.
func randomRuns(r *rand.Rand, lo, maxRuns, maxLen, maxGap int) *runContainer16 {
	var vals []uint16
	v := lo
	for k := 0; k < maxRuns && v < 1<<16; k++ {
		for end := min(v+1+r.Intn(maxLen), 1<<16); v < end; v++ {
			vals = append(vals, uint16(v))
		}
		v += 1 + r.Intn(maxGap)
	}
	return newRunContainer16FromVals(true, vals...)
}

func TestRunContainerUnionExhaustive(t *testing.T) {
	// every pair of subsets of an 8-value window, at both ends of the range
	for _, base := range []int{0, 1<<16 - 8} {
		for x := 0; x < 256; x++ {
			for y := 0; y < 256; y++ {
				var av, bv []uint16
				for i := 0; i < 8; i++ {
					if x&(1<<i) != 0 {
						av = append(av, uint16(base+i))
					}
					if y&(1<<i) != 0 {
						bv = append(bv, uint16(base+i))
					}
				}
				a, b := newRunContainer16FromVals(true, av...), newRunContainer16FromVals(true, bv...)
				want := unionValues(av, bv)
				require.Equal(t, want, containerValues(a.union(b)))
				var both []uint16
				for _, v := range av {
					if slices.Contains(bv, v) {
						both = append(both, v)
					}
				}
				require.Equal(t, both, containerValues(a.intersect(b)))
				got := a.Clone().inplaceUnion(b)
				require.Equal(t, want, containerValues(got))
				if len(want) > 0 {
					require.NoError(t, got.validate())
				}
			}
		}
	}
}

func TestRunContainerInplaceUnionRandom(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	for trial := 0; trial < 500; trial++ {
		maxLen, maxGap := 1+r.Intn(100), 1+r.Intn(100)
		a := randomRuns(r, r.Intn(64), 1+r.Intn(3000), maxLen, maxGap)
		// small operands straddle runInsertMaxRuns; every other trial uses a large one
		b := randomRuns(r, r.Intn(1<<16), 1+r.Intn(2*runInsertMaxRuns), maxLen, maxGap)
		if trial%2 == 1 {
			b = randomRuns(r, r.Intn(64), 1+r.Intn(3000), maxLen, maxGap)
		}
		av, bv := containerValues(a), containerValues(b)

		got := a.inplaceUnion(b)
		require.Equal(t, unionValues(av, bv), containerValues(got))
		require.NoError(t, got.validate())
		require.Equal(t, bv, containerValues(b), "operand modified")
	}
	// self-union
	a := randomRuns(r, 0, 10, 50, 50)
	want := containerValues(a)
	require.Equal(t, want, containerValues(a.inplaceUnion(a)))
}

func TestArrayContainerIorRun16Random(t *testing.T) {
	r := rand.New(rand.NewSource(2))
	for trial := 0; trial < 2000; trial++ {
		ac := newArrayContainer()
		for n := 1 + r.Intn(arrayDefaultMaxSize); n > 0; n-- {
			ac.iadd(uint16(r.Intn(1 << 16)))
		}
		rc := randomRuns(r, r.Intn(1<<16), 1+r.Intn(4*arrayRunInsertMaxRuns), 1+r.Intn(20), 1+r.Intn(2000))
		acv, rcv := containerValues(ac), containerValues(rc)

		got := ac.iorRun16(rc)
		require.Equal(t, unionValues(acv, rcv), containerValues(got))
		require.Equal(t, rcv, containerValues(rc), "operand modified")
		if len(rcv) < len(acv) && len(rcv)+len(acv) < arrayDefaultMaxSize {
			require.Same(t, ac, got)
		}
	}
}

func TestRunContainerToArrayContainer(t *testing.T) {
	r := rand.New(rand.NewSource(3))
	for trial := 0; trial < 500; trial++ {
		rc := randomRuns(r, r.Intn(1000), 1+r.Intn(200), 1+r.Intn(40), 1+r.Intn(300))
		require.Equal(t, containerValues(rc), rc.toArrayContainer().content)
	}
}

func BenchmarkRunContainerInplaceUnion(b *testing.B) {
	receiver := newRunContainer16()
	for v := 0; v < 1<<16; v += 64 {
		receiver.iv = append(receiver.iv, newInterval16Range(uint16(v), uint16(v+31)))
	}
	for _, runs := range []int{4, 1024} {
		operand := newRunContainer16()
		for i := 0; i < runs; i++ {
			v := i * (1 << 16) / runs
			operand.iv = append(operand.iv, newInterval16Range(uint16(v+40), uint16(v+47)))
		}
		b.Run(fmt.Sprintf("runs=%d", runs), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				receiver.Clone().inplaceUnion(operand)
			}
		})
	}
}

func BenchmarkArrayContainerIorRun16(b *testing.B) {
	ac := newArrayContainer()
	for v := 0; v < 1<<16; v += 32 {
		ac.content = append(ac.content, uint16(v))
	}
	for _, runs := range []int{4, 100} {
		rc := newRunContainer16()
		for i := 0; i < runs; i++ {
			v := i*(1<<16)/runs + 7
			rc.iv = append(rc.iv, newInterval16Range(uint16(v), uint16(v+2)))
		}
		b.Run(fmt.Sprintf("runs=%d", runs), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				ac.clone().(*arrayContainer).iorRun16(rc)
			}
		})
	}
}
