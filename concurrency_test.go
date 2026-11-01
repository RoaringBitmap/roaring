package roaring

import (
	"runtime"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Port of CRoaring tests/threads_unit.cpp: copy-on-write clones that share
// containers are mutated in parallel. Clone() of a copy-on-write bitmap marks
// the containers of the source as shared, so it writes to the source: every
// clone that a goroutine starts from is made on the main goroutine, and each
// goroutine only clones bitmaps that it owns. Run with -race.
func TestConcurrentMutationOfCOWClones(t *testing.T) {
	r1 := New()
	r1.SetCopyOnWrite(true)
	for i := uint32(0); i < 50000; i++ {
		if i != 300 {
			r1.Add(65536 + i)
		}
	}
	for i := uint32(50000); i < 150000; i += 500 {
		r1.Add(i)
	}
	for i := uint32(150000); i < 200000; i += 2 {
		r1.Add(i)
	}
	r1.RunOptimize()
	r2 := BitmapOf(10010, 10020, 10030, 10040, 10050)
	r2.SetCopyOnWrite(true)
	r3 := r1.Clone()

	card1, card2, card3 := r1.GetCardinality(), r2.GetCardinality(), r3.GetCardinality()
	expected := And(r1, r2)
	expected.AndNot(r3)

	const workers = 2
	var sets [workers][3]*Bitmap
	for w := range sets {
		sets[w] = [3]*Bitmap{r1.Clone(), r2.Clone(), r3.Clone()}
	}
	var wg sync.WaitGroup
	results := make([]bool, workers)
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			ok := true
			for i := 0; i < 100; i++ {
				x1 := sets[w][0].Clone()
				x2 := sets[w][1].Clone()
				x3 := sets[w][2].Clone()
				x1.And(x2)
				x1.AndNot(x3)
				x3.Add(7)
				x3.Remove(65536)
				if !x1.Equals(expected) {
					ok = false
				}
			}
			results[w] = ok
		}(w)
	}
	wg.Wait()
	for w := range results {
		assert.True(t, results[w], "worker %d", w)
	}
	assert.Equal(t, card1, r1.GetCardinality())
	assert.Equal(t, card2, r2.GetCardinality())
	assert.Equal(t, card3, r3.GetCardinality())
	require.NoError(t, r1.Validate())
	require.NoError(t, r3.Validate())
}

// Java TestMemoryMapping: many goroutines read the same buffer-backed bitmap.
func TestConcurrentReadersOnBufferBackedBitmap(t *testing.T) {
	n := runtime.NumCPU()
	if n < 4 {
		n = 4
	}
	src := New()
	for i := 0; i < n; i++ {
		src.Add(uint32(32767 * i))
	}
	// keep the range above every per-goroutine value 32767*g, even with many CPUs
	src.AddRange(1<<31, 1<<31+5000)
	src.RunOptimize()
	buf, err := src.ToBytes()
	require.NoError(t, err)
	rb := New()
	_, err = rb.FromBuffer(buf)
	require.NoError(t, err)
	other := BitmapOf(0, 32767, 1<<31, 1<<31+4999)
	card := rb.GetCardinality()

	var start, wg sync.WaitGroup
	start.Add(1)
	failures := make([]string, n)
	for g := 0; g < n; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			start.Wait()
			v := uint32(32767 * g)
			if !rb.Contains(v) {
				failures[g] = "contains"
				return
			}
			count := uint64(0)
			for it := rb.Iterator(); it.HasNext(); it.Next() {
				count++
			}
			if count != card {
				failures[g] = "iterator count"
				return
			}
			if rb.Rank(v) != uint64(g+1) {
				failures[g] = "rank"
				return
			}
			if And(rb, other).GetCardinality() != 4 {
				failures[g] = "and"
				return
			}
			if rb.GetCardinality() != card || !rb.Equals(src) {
				failures[g] = "content"
			}
		}(g)
	}
	start.Done()
	wg.Wait()
	for g, f := range failures {
		assert.Empty(t, f, "goroutine %d", g)
	}
}
