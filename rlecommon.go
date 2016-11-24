package roaring

import (
	"fmt"
	"sort"
)

// common to rle32.go and rle16.go

// rleVerbose controls whether p() prints show up.
// The testing package sets this based on
// testing.Verbose().
var rleVerbose bool

// p is a shorthand for fmt.Printf with beginning and
// trailing newlines. p() makes it easy
// to add diagnostic print statements.
func p(format string, args ...interface{}) {
	if rleVerbose {
		fmt.Printf("\n"+format+"\n", args...)
	}
}

// MaxUint32 is the largest uint32 value.
const MaxUint32 = 4294967295

// MaxUint16 is the largest 16 bit unsigned int.
// This is the largest value an interval16 can store.
const MaxUint16 = 65535

// searchOptions allows us to accelerate runContainer32.search with
// prior knowledge of (mostly lower) bounds. This is used by Union
// and Intersect.
type searchOptions struct {

	// start here instead of at 0
	StartIndex int64

	// upper bound instead of len(rc.iv);
	// EndxIndex == 0 means ignore the bound and use
	// EndxIndex == n ==len(rc.iv) which is also
	// naturally the default for search()
	// when opt = nil.
	EndxIndex int64
}

// And finds the intersection of rc and b.
func (rc *runContainer32) And(b *Bitmap) *Bitmap {
	out := NewBitmap()
	for _, p := range rc.Iv {
		for i := p.Start; i <= p.Last; i++ {
			if b.Contains(i) {
				out.Add(i)
			}
		}
	}
	return out
}

// Xor returns the exclusive-or of rc and b.
func (rc *runContainer32) Xor(b *Bitmap) *Bitmap {
	out := b.Clone()
	for _, p := range rc.Iv {
		for v := p.Start; v <= p.Last; v++ {
			if out.Contains(v) {
				out.RemoveRange(uint64(v), uint64(v+1))
			} else {
				out.Add(v)
			}
		}
	}
	return out
}

// Or returns the union of rc and b.
func (rc *runContainer32) Or(b *Bitmap) *Bitmap {
	out := b.Clone()
	for _, p := range rc.Iv {
		for v := p.Start; v <= p.Last; v++ {
			out.Add(v)
		}
	}
	return out
}

func showHash(name string, h map[int]bool) {
	hv := []int{}
	for k := range h {
		hv = append(hv, k)
	}
	sort.Sort(sort.IntSlice(hv))
	stringH := ""
	for i := range hv {
		stringH += fmt.Sprintf("%v, ", hv[i])
	}

	fmt.Printf("%s is (len %v): %s", name, len(h), stringH)
}

// trial is used in the randomized testing of runContainers
type trial struct {
	n           int
	percentFill float64
	ntrial      int

	// only in the union test
	// only subtract test
	percentDelete float64
}

// And finds the intersection of rc and b.
func (rc *runContainer16) And(b *Bitmap) *Bitmap {
	out := NewBitmap()
	for _, p := range rc.Iv {
		for i := p.Start; i <= p.Last; i++ {
			if b.Contains(uint32(i)) {
				out.Add(uint32(i))
			}
		}
	}
	return out
}

// Xor returns the exclusive-or of rc and b.
func (rc *runContainer16) Xor(b *Bitmap) *Bitmap {
	out := b.Clone()
	for _, p := range rc.Iv {
		for v := p.Start; v <= p.Last; v++ {
			w := uint32(v)
			if out.Contains(w) {
				out.RemoveRange(uint64(w), uint64(w+1))
			} else {
				out.Add(w)
			}
		}
	}
	return out
}

// Or returns the union of rc and b.
func (rc *runContainer16) Or(b *Bitmap) *Bitmap {
	out := b.Clone()
	for _, p := range rc.Iv {
		for v := p.Start; v <= p.Last; v++ {
			out.Add(uint32(v))
		}
	}
	return out
}

func (rc *runContainer32) and(container) container {
	panic("TODO. not yet implemented")
}
