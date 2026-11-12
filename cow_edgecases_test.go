package roaring

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// cowEdgeSource mirrors the source bitmap of CRoaring's
// test_inplace_mutators_copy_on_write: [0, 2^17) plus {5<<16}.
func cowEdgeSource(cow, runOptimize bool) *Bitmap {
	rb := New()
	rb.SetCopyOnWrite(cow)
	rb.AddRange(0, 1<<17)
	rb.Add(5 << 16)
	if runOptimize {
		rb.RunOptimize()
	}
	return rb
}

func requireCOWSourceIntact(t *testing.T, rb *Bitmap, msg string) {
	t.Helper()
	require.NoError(t, rb.Validate(), msg)
	require.EqualValues(t, 131073, rb.GetCardinality(), msg)
	require.True(t, rb.Contains(0), msg)
	require.True(t, rb.Contains(131071), msg)
	require.True(t, rb.Contains(327680), msg)
}

func cowRangeBitmap(start, end uint64) *Bitmap {
	rb := New()
	rb.AddRange(start, end)
	return rb
}

type cowMutator struct {
	name string
	f    func(*Bitmap)
}

func cowMutators() []cowMutator {
	removeMany := func(rb *Bitmap) {
		for _, v := range []uint32{1, 3, 5, 7, 65537, 65539, 327680} {
			rb.Remove(v)
		}
	}
	return []cowMutator{
		{"Add", func(rb *Bitmap) { rb.Add(2); rb.Add(131072) }},
		{"CheckedAdd", func(rb *Bitmap) { rb.CheckedAdd(2); rb.CheckedAdd(131072) }},
		{"AddMany", func(rb *Bitmap) { rb.AddMany([]uint32{131073, 131075, 2, 4}) }},
		{"Remove", func(rb *Bitmap) { rb.Remove(2) }},
		{"CheckedRemove", func(rb *Bitmap) { rb.CheckedRemove(2) }},
		{"AddRange", func(rb *Bitmap) { rb.AddRange(131073, 131172) }},
		{"RemoveRange", func(rb *Bitmap) { rb.RemoveRange(10, 100) }},
		{"RemoveRangeLastContainer", func(rb *Bitmap) { rb.RemoveRange(327680, 327681) }},
		{"Flip", func(rb *Bitmap) { rb.Flip(10, 100) }},
		{"FlipAcrossContainers", func(rb *Bitmap) { rb.Flip(65530, 327690) }},
		{"RunOptimize", func(rb *Bitmap) { rb.RunOptimize() }},
		{"And", func(rb *Bitmap) { rb.And(cowRangeBitmap(0, 100)) }},
		{"Or", func(rb *Bitmap) { rb.Or(cowRangeBitmap(131073, 131172)) }},
		{"OrSameKeys", func(rb *Bitmap) { rb.Or(BitmapOf(70000, 327681)) }},
		{"Xor", func(rb *Bitmap) { rb.Xor(cowRangeBitmap(0, 100)) }},
		{"AndNot", func(rb *Bitmap) { rb.AndNot(cowRangeBitmap(0, 100)) }},
		{"RemoveMany", removeMany},
		{"Clear", func(rb *Bitmap) { rb.Clear() }},
	}
}

// Every in-place mutator must unshare the containers it writes, in both
// directions (CRoaring test_inplace_mutators_copy_on_write, Java
// TestCopyOnWriteEdgeCases).
func TestCOWEveryMutatorUnshares(t *testing.T) {
	for _, runOptimize := range []bool{false, true} {
		for _, m := range cowMutators() {
			t.Run(fmt.Sprintf("%s/runopt=%v", m.name, runOptimize), func(t *testing.T) {
				expected := cowEdgeSource(false, runOptimize)
				m.f(expected)

				src := cowEdgeSource(true, runOptimize)
				cp := src.Clone()
				m.f(cp)
				requireCOWSourceIntact(t, src, "source after mutating the copy")
				require.NoError(t, cp.Validate())
				require.True(t, cp.Equals(expected))

				src = cowEdgeSource(true, runOptimize)
				cp = src.Clone()
				m.f(src)
				requireCOWSourceIntact(t, cp, "copy after mutating the source")
				require.NoError(t, src.Validate())
				require.True(t, src.Equals(expected))
			})
		}
	}
}

// CRoaring test_remove_many_copy_on_write: many clones, each mutated in turn.
func TestCOWManyCopies(t *testing.T) {
	src := cowEdgeSource(true, true)
	copies := make([]*Bitmap, 4)
	for i := range copies {
		copies[i] = src.Clone()
	}
	for i := range copies {
		for _, v := range []uint32{1, 3, 5, 7, 65537, 65539, 327680} {
			copies[i].Remove(v)
		}
		require.NoError(t, copies[i].Validate())
		require.EqualValues(t, 131073-7, copies[i].GetCardinality())
		requireCOWSourceIntact(t, src, "source")
		for j := i + 1; j < len(copies); j++ {
			requireCOWSourceIntact(t, copies[j], fmt.Sprintf("copy %d after mutating copy %d", j, i))
		}
	}
}

// threeContainers mirrors Java's TestCopyOnWriteEdgeCases.threeContainers:
// an array at key 0, a run at key 2 and a bitmap at key 4.
func cowThreeContainers(cow bool) *Bitmap {
	rb := New()
	rb.SetCopyOnWrite(cow)
	rb.AddMany([]uint32{1, 2, 3})
	rb.AddRange(131072, 132072)
	for v := uint32(262144); v < 272144; v += 2 {
		rb.Add(v)
	}
	rb.RunOptimize()
	return rb
}

// Writes after an insertion or a removal of a container shift the index of
// shared containers; the per-container copy-on-write flags must follow.
func TestCOWWritesAfterContainersShift(t *testing.T) {
	st := cowThreeContainers(false).Stats()
	require.EqualValues(t, 1, st.ArrayContainers)
	require.EqualValues(t, 1, st.RunContainers)
	require.EqualValues(t, 1, st.BitmapContainers)

	cases := []cowMutator{
		{"insert key 1 then remove from bitmap", func(rb *Bitmap) {
			rb.Add(65536)
			rb.Remove(262146)
		}},
		{"insert key 1 then or into bitmap", func(rb *Bitmap) {
			rb.Add(65536)
			rb.Or(BitmapOf(262145))
		}},
		{"drop key 0 then remove from run", func(rb *Bitmap) {
			rb.Add(5)
			for _, v := range []uint32{1, 2, 3, 5} {
				rb.Remove(v)
			}
			rb.Remove(131072)
		}},
		{"insert key 3 then flip in run and bitmap", func(rb *Bitmap) {
			rb.Add(3 << 16)
			rb.Flip(131070, 262150)
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			expected := cowThreeContainers(false)
			c.f(expected)
			src := cowThreeContainers(true)
			cp := src.Clone()
			c.f(cp)
			require.True(t, src.Equals(cowThreeContainers(false)))
			require.NoError(t, src.Validate())
			require.True(t, cp.Equals(expected))
			require.NoError(t, cp.Validate())
		})
	}
}

// A bitmap loaded with FromBuffer must never write into the buffer, whatever
// the mutator (Java TestCopyOnWriteEdgeCases.assertCopyOnWrite).
func TestCOWFromBufferNeverWritesTheBuffer(t *testing.T) {
	const key2, key4 = 131072, 262144
	buf, err := cowThreeContainers(false).ToBytes()
	require.NoError(t, err)
	orig := bytes.Clone(buf)
	writes := []cowMutator{
		{"Add", func(rb *Bitmap) { rb.Add(key4 + 1); rb.Add(5); rb.Add(key2 + 2000) }},
		{"Remove", func(rb *Bitmap) { rb.Remove(key4 + 2); rb.Remove(1); rb.Remove(key2 + 5) }},
		{"CheckedAdd", func(rb *Bitmap) { rb.CheckedAdd(key4 + 1); rb.CheckedAdd(4) }},
		{"CheckedRemove", func(rb *Bitmap) { rb.CheckedRemove(key4 + 2); rb.CheckedRemove(2) }},
		{"FlipOne", func(rb *Bitmap) { rb.Flip(key4+2, key4+3) }},
		{"FlipRange", func(rb *Bitmap) { rb.Flip(key4, key4+100) }},
		{"FlipAll", func(rb *Bitmap) { rb.Flip(0, key4+20000) }},
		{"AddRange", func(rb *Bitmap) { rb.AddRange(key4+1, key4+100) }},
		{"RemoveRangeBitmap", func(rb *Bitmap) { rb.RemoveRange(key4+200, key4+400) }},
		{"RemoveRangeRun", func(rb *Bitmap) { rb.RemoveRange(key2, key2+10) }},
		{"RemoveRangeArray", func(rb *Bitmap) { rb.RemoveRange(2, 3) }},
		{"AddMany", func(rb *Bitmap) { rb.AddMany([]uint32{key4 + 1, key4 + 3, 7, key2 + 1001}) }},
		{"Or", func(rb *Bitmap) { rb.Or(BitmapOf(key4+1, 9, key2+1500)) }},
		{"Xor", func(rb *Bitmap) { rb.Xor(BitmapOf(key4, key4+1, 2, 9, key2+3)) }},
		{"AndNot", func(rb *Bitmap) { rb.AndNot(BitmapOf(key4, 2, key2+3)) }},
		{"And", func(rb *Bitmap) { rb.And(BitmapOf(key4, key4+1, 2, key2+3)) }},
		{"AndRange", func(rb *Bitmap) { rb.And(cowRangeBitmap(0, key4+10)) }},
		{"RunOptimize", func(rb *Bitmap) { rb.RunOptimize() }},
		{"Clear", func(rb *Bitmap) { rb.Clear(); rb.Add(key4) }},
	}
	for _, w := range writes {
		t.Run(w.name, func(t *testing.T) {
			expected := cowThreeContainers(false)
			w.f(expected)
			rb := New()
			_, err := rb.FromBuffer(buf)
			require.NoError(t, err)
			w.f(rb)
			require.True(t, bytes.Equal(orig, buf), "buffer modified")
			require.NoError(t, rb.Validate())
			require.True(t, rb.Equals(expected))

			// same through a copy-on-write clone of the buffer-backed bitmap
			rb = New()
			_, err = rb.FromBuffer(buf)
			require.NoError(t, err)
			cp := rb.Clone()
			w.f(cp)
			require.True(t, bytes.Equal(orig, buf), "buffer modified through clone")
			require.True(t, cp.Equals(expected))
			require.True(t, rb.Equals(cowThreeContainers(false)))
		})
	}
}

// Replacing the content of a copy-on-write clone must drop the shared state
// (Java TestCopyOnWriteEdgeCases, CRoaring overwrite).
func TestCOWReplacingContentDropsSharedState(t *testing.T) {
	other := BitmapOf(3, 70000, 1<<20)
	otherBytes, err := other.ToBytes()
	require.NoError(t, err)
	origOther := bytes.Clone(otherBytes)
	replacements := []struct {
		name    string
		f       func(t *testing.T, rb *Bitmap)
		content *Bitmap
	}{
		{"Clear", func(t *testing.T, rb *Bitmap) { rb.Clear() }, New()},
		{"UnmarshalBinary", func(t *testing.T, rb *Bitmap) {
			require.NoError(t, rb.UnmarshalBinary(otherBytes))
		}, other},
		{"ReadFrom", func(t *testing.T, rb *Bitmap) {
			_, err := rb.ReadFrom(bytes.NewReader(otherBytes))
			require.NoError(t, err)
		}, other},
		{"FromBuffer", func(t *testing.T, rb *Bitmap) {
			_, err := rb.FromBuffer(otherBytes)
			require.NoError(t, err)
		}, other},
	}
	for _, runOptimize := range []bool{false, true} {
		for _, r := range replacements {
			t.Run(fmt.Sprintf("%s/runopt=%v", r.name, runOptimize), func(t *testing.T) {
				src := cowEdgeSource(true, runOptimize)
				cp := src.Clone()
				r.f(t, cp)
				cp.AddRange(0, 1<<17)
				cp.Remove(5)
				cp.Remove(70000)
				requireCOWSourceIntact(t, src, "source")
				require.NoError(t, cp.Validate())
				expected := r.content.Clone()
				expected.AddRange(0, 1<<17)
				expected.Remove(5)
				expected.Remove(70000)
				require.True(t, cp.Equals(expected))
				require.True(t, bytes.Equal(origOther, otherBytes))
			})
		}
	}
}

// Java containerSharingWithXor / testAstesana: a container moved into r2 by
// an in-place xor/or must not be mutated through r2 afterwards.
func TestContainerSharedByInPlaceOpNotMutatedLater(t *testing.T) {
	for _, cow := range []bool{false, true} {
		for _, useXor := range []bool{false, true} {
			t.Run(fmt.Sprintf("cow=%v/xor=%v", cow, useXor), func(t *testing.T) {
				r1 := New()
				r1.SetCopyOnWrite(cow)
				r1.Flip(131000, 131001)
				r2 := New()
				r2.SetCopyOnWrite(cow)
				r2.Add(220000)
				r3 := New()
				r3.SetCopyOnWrite(cow)
				r3.Add(66000)
				if useXor {
					r2.Xor(r1)
				} else {
					r2.Or(r1)
				}
				r2.Or(r3)
				assert.True(t, r2.Contains(131000))
				assert.True(t, r2.Contains(66000))
				assert.True(t, r2.Contains(220000))
				assert.False(t, r1.Contains(66000))
				assert.True(t, r1.Equals(BitmapOf(131000)))
				assert.True(t, r3.Equals(BitmapOf(66000)))
				require.NoError(t, r1.Validate())
				require.NoError(t, r2.Validate())
			})
		}
	}
}

// Results of non-in-place operations and aggregations must not share
// writable containers with their inputs (Java xorDoesNotShareContainersWithItsInputs,
// xorAndAndNotWithAnEmptyArrayReturnANewContainer).
func TestResultsDoNotShareContainersWithInputs(t *testing.T) {
	a := New()
	for v := uint32(1); v < 20; v += 2 {
		a.Add(v)
	}
	runsPlain := cowRangeBitmap(100, 105)
	runsPlain.RunOptimize()
	require.True(t, runsPlain.HasRunCompression())
	runsBytes, err := runsPlain.ToBytes()
	require.NoError(t, err)
	origBytes := bytes.Clone(runsBytes)
	expectedRuns := cowRangeBitmap(100, 105)

	makeRuns := map[string]func() *Bitmap{
		"plain": func() *Bitmap {
			rb := cowRangeBitmap(100, 105)
			rb.RunOptimize()
			return rb
		},
		"cow": func() *Bitmap {
			rb := New()
			rb.SetCopyOnWrite(true)
			rb.AddRange(100, 105)
			rb.RunOptimize()
			return rb
		},
		"frombuffer": func() *Bitmap {
			rb := New()
			_, err := rb.FromBuffer(runsBytes)
			require.NoError(t, err)
			return rb
		},
	}
	ops := map[string]func(runs *Bitmap) *Bitmap{
		"HeapXor":  func(runs *Bitmap) *Bitmap { return HeapXor(a, a.Clone(), runs) },
		"Xor":      func(runs *Bitmap) *Bitmap { return Xor(runs, New()) },
		"XorLeft":  func(runs *Bitmap) *Bitmap { return Xor(New(), runs) },
		"AndNot":   func(runs *Bitmap) *Bitmap { return AndNot(runs, New()) },
		"Or":       func(runs *Bitmap) *Bitmap { return Or(runs, New()) },
		"OrLeft":   func(runs *Bitmap) *Bitmap { return Or(New(), runs) },
		"And":      func(runs *Bitmap) *Bitmap { return And(runs, cowRangeBitmap(0, 1000)) },
		"FastOr":   func(runs *Bitmap) *Bitmap { return FastOr(runs, New()) },
		"FastOr1":  func(runs *Bitmap) *Bitmap { return FastOr(runs) },
		"HeapOr":   func(runs *Bitmap) *Bitmap { return HeapOr(runs, New()) },
		"ParOr":    func(runs *Bitmap) *Bitmap { return ParOr(2, runs, New()) },
		"FastAnd1": func(runs *Bitmap) *Bitmap { return FastAnd(runs) },
	}
	for rname, mk := range makeRuns {
		for oname, op := range ops {
			t.Run(rname+"/"+oname, func(t *testing.T) {
				runs := mk()
				result := op(runs)
				require.True(t, result.Equals(expectedRuns))
				result.Or(runs)
				result.Add(5000)
				result.Remove(102)
				require.True(t, result.Equals(BitmapOf(100, 101, 103, 104, 5000)))
				require.NoError(t, result.Validate())
				require.True(t, runs.Equals(expectedRuns), "input modified")
				require.True(t, bytes.Equal(origBytes, runsBytes), "buffer modified")
			})
		}
	}
}

// CRoaring issue743: an andNot result of copy-on-write bitmaps reused by xor.
func TestCOWAndNotResultReusedByXor(t *testing.T) {
	a := BitmapOf(1, 2, 3, 4)
	a.SetCopyOnWrite(true)
	b := BitmapOf(0x38000000, 0x38000001, 0x38000002, 0x38000003)
	b.SetCopyOnWrite(true)
	c := AndNot(b, a)
	c.SetCopyOnWrite(false)
	a.Clear()
	b.Clear()
	d := BitmapOf(1, 2, 3, 4, 5, 6)
	e := Xor(d, c)
	expected := BitmapOf(1, 2, 3, 4, 5, 6, 939524096, 939524097, 939524098, 939524099)
	require.True(t, e.Equals(expected))
	require.NoError(t, e.Validate())
	e.Add(939524100)
	e.Remove(939524096)
	require.True(t, c.Equals(BitmapOf(939524096, 939524097, 939524098, 939524099)))
	require.NoError(t, c.Validate())
}

// Clear resets the bitmap to a new one, so copy-on-write is turned off.
// Xor and AndNot of a bitmap with itself also empty it, but keep the setting.
func TestClearResetsCopyOnWrite(t *testing.T) {
	rb := BitmapOf(1, 2, 3)
	rb.SetCopyOnWrite(true)
	rb.Clear()
	require.False(t, rb.GetCopyOnWrite())
	rb.Add(4)
	require.NoError(t, rb.Validate())

	for _, op := range []func(*Bitmap){
		func(b *Bitmap) { b.Xor(b) },
		func(b *Bitmap) { b.AndNot(b) },
	} {
		b := BitmapOf(1, 2, 3, 100000)
		b.SetCopyOnWrite(true)
		op(b)
		require.True(t, b.IsEmpty())
		require.True(t, b.GetCopyOnWrite())
		b.Add(5)
		require.NoError(t, b.Validate())
		require.True(t, b.Contains(5))
	}
}
