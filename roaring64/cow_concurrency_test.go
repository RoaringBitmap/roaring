package roaring64

import (
	"bytes"
	"sync"
	"testing"
)

// newSharedCOWBitmap returns a copy-on-write bitmap with several high keys
// and array, bitmap and run containers, meant to be shared read-only between
// goroutines.
func newSharedCOWBitmap() *Bitmap {
	rb := New()
	for hi := uint64(0); hi < 6; hi++ {
		base := hi << 32
		for i := uint64(0); i < 1<<20; i += 977 { // array containers
			rb.Add(base + i)
		}
		for i := uint64(1 << 20); i < 1<<20+40000; i += 3 { // bitmap container
			rb.Add(base + i)
		}
		rb.AddRange(base+1<<22, base+1<<22+100000) // run containers
	}
	rb.RunOptimize()
	rb.SetCopyOnWrite(true)
	return rb
}

// newOtherCOWBitmap returns a fresh copy-on-write bitmap that overlaps src
// only on its first high key, so that in-place operations with src as
// argument share src's trailing containers.
func newOtherCOWBitmap() *Bitmap {
	x := New()
	x.SetCopyOnWrite(true)
	x.Add(5)
	x.Add(70000)
	return x
}

// runConcurrently runs every function in fns from several goroutines at the
// same time, each one repeatedly.
func runConcurrently(fns []func()) {
	const repeat = 4
	start := make(chan struct{})
	var wg sync.WaitGroup
	for r := 0; r < repeat; r++ {
		for _, f := range fns {
			wg.Add(1)
			go func(f func()) {
				defer wg.Done()
				<-start
				for i := 0; i < 5; i++ {
					f()
				}
			}(f)
		}
	}
	close(start)
	wg.Wait()
}

// deepCopy returns a copy of rb that shares no containers with it.
func deepCopy(rb *Bitmap) *Bitmap {
	c := New()
	c.AddMany(rb.ToArray())
	return c
}

// TestConcurrentCloneOfCOWBitmap is the scenario of issue #528 for 64-bit
// bitmaps: goroutines that only call Clone on one copy-on-write bitmap must
// not race.
func TestConcurrentCloneOfCOWBitmap(t *testing.T) {
	src := newSharedCOWBitmap()
	want := deepCopy(src)

	runConcurrently([]func(){
		func() { _ = src.Clone() },
		func() { _ = src.Clone() },
	})
	if !src.Equals(want) {
		t.Fatal("concurrent clones changed the source")
	}
}

// TestConcurrentReadOnlyUseOfCOWBitmap checks that operations which leave a
// copy-on-write bitmap unchanged can run concurrently on it, including the
// ones that update its copy-on-write flags, and that clones can be modified
// independently while the source is shared.
func TestConcurrentReadOnlyUseOfCOWBitmap(t *testing.T) {
	src := newSharedCOWBitmap()
	other := newOtherCOWBitmap()
	want := deepCopy(src)
	card := want.GetCardinality()

	check := func(ok bool, msg string) {
		if !ok {
			t.Error(msg)
		}
	}
	runConcurrently([]func(){
		func() { _ = src.Clone() },
		func() {
			c := src.Clone()
			c.RemoveRange(0, 1<<19)
			c.Add(3)
			c.Remove(1 << 20)
			c.AddRange(1<<32+1<<22-10, 1<<32+1<<22+10)
			check(c.Contains(3) && !src.Contains(3), "clone write leaked into the source")
		},
		func() { x := newOtherCOWBitmap(); x.Or(src) },
		func() { x := newOtherCOWBitmap(); x.Xor(src) },
		func() { x := newOtherCOWBitmap(); x.And(src) },
		func() { x := newOtherCOWBitmap(); x.AndNot(src) },
		func() { x := src.Clone(); x.Or(src) },
		func() { _ = Or(other, src) },
		func() { _ = Xor(src, other) },
		func() { _ = And(src, other) },
		func() { _ = AndNot(src, other) },
		func() { _ = Flip(src, 10, 1<<33) },
		func() { _ = FastOr(other, src) },
		func() { _ = FastAnd(src, other) },
		func() { _ = ParOr(2, other, src) },
		func() { check(src.GetCardinality() == card, "cardinality changed") },
		func() { check(src.Contains(1<<32+1<<20+3), "missing value") },
		func() { _ = src.Rank(1 << 33) },
		func() { _, _ = src.Select(100) },
		func() { _ = src.ToArray() },
		func() {
			it := src.Iterator()
			for it.HasNext() {
				it.Next()
			}
		},
		func() { _ = src.Intersects(other) },
		func() { _ = src.AndCardinality(other) },
		func() { _ = src.OrCardinality(other) },
		func() { _ = src.Equals(other) },
		func() {
			var buf bytes.Buffer
			_, err := src.WriteTo(&buf)
			check(err == nil, "WriteTo failed")
		},
	})
	if !src.Equals(want) {
		t.Fatal("concurrent read-only operations changed the source")
	}
	// Every container of src is now shared: modifying src must not affect
	// a clone taken before.
	clone := src.Clone()
	src.Add(7)
	src.RemoveRange(1<<32+1<<22, 1<<32+1<<22+10)
	if !clone.Equals(want) {
		t.Fatal("modifying the source changed a clone")
	}
}
