package roaring

import (
	"fmt"
	"math/rand"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// iteratorReferenceData mirrors CRoaring's test_iterator_generate_data
// (tests/toplevel_unit.c): values at structural boundaries, one pattern per
// container, in ascending order.
func iteratorReferenceData() []uint32 {
	values := make([]uint32, 0, 500000)
	base := uint32(1234)
	values = append(values, 0)
	// only the first value of a container
	values = append(values, base*65536)
	base += 2
	// only the last value of a container
	values = append(values, base*65536+65535)
	base += 2
	// full container
	for i := uint32(0); i < 65536; i++ {
		values = append(values, base*65536+i)
	}
	base += 2
	// even values
	for i := uint32(0); i < 65536; i += 2 {
		values = append(values, base*65536+i)
	}
	base += 2
	// odd values
	for i := uint32(1); i < 65536; i += 2 {
		values = append(values, base*65536+i)
	}
	base += 2
	// each 64-bit word rotated by one
	for i := uint32(0); i < 65536; i += 65 {
		values = append(values, base*65536+i)
	}
	base += 2
	// runs of increasing length
	for i, runIndex := uint32(0), uint32(0); i < 65536; i++ {
		if i != (runIndex+1)*(runIndex+2)/2-1 {
			values = append(values, base*65536+i)
		} else {
			runIndex++
		}
	}
	base += 2
	// a run over three adjacent containers
	for i := uint32(65536 - 100); i < 65536; i++ {
		values = append(values, base*65536+i)
	}
	base++
	for i := uint32(0); i < 65536; i++ {
		values = append(values, base*65536+i)
	}
	base++
	for i := uint32(0); i < 100; i++ {
		values = append(values, base*65536+i)
	}
	base += 2
	// random
	r := rand.New(rand.NewSource(1234))
	for i := 0; i < 65536; i += r.Intn(10) + 1 {
		values = append(values, base*65536+uint32(i))
	}
	values = append(values, MaxUint32)
	return values
}

// forceContainerType converts every container of rb to the given type
// ("array", "bitmap" or "run"). The result may be non-canonical, which is
// fine for iteration.
func iterForceContainerType(rb *Bitmap, kind string) {
	ra := &rb.highlowcontainer
	for i, c := range ra.containers {
		// build the bitmap value by value: runContainer16.toBitmapContainer
		// does not maintain the cardinality field.
		bc := newBitmapContainer()
		for it := c.getShortIterator(); it.hasNext(); {
			bc.iadd(it.next())
		}
		switch kind {
		case "array":
			ra.containers[i] = bc.toArrayContainer()
		case "bitmap":
			ra.containers[i] = bc
		case "run":
			ra.containers[i] = newRunContainer16FromBitmapContainer(bc)
		}
		ra.needCopyOnWrite[i] = 0
	}
}

func iterEqualUint32s(t *testing.T, expected, actual []uint32, msg string) {
	t.Helper()
	if !slices.Equal(expected, actual) {
		require.Equal(t, len(expected), len(actual), msg)
		for i := range expected {
			if expected[i] != actual[i] {
				t.Fatalf("%s: mismatch at %d: %d != %d", msg, i, expected[i], actual[i])
			}
		}
	}
}

func TestIteratorsOnReferenceDataset(t *testing.T) {
	ref := iteratorReferenceData()
	reversed := slices.Clone(ref)
	slices.Reverse(reversed)
	for _, mode := range []string{"native", "runoptimized", "array", "bitmap", "run"} {
		t.Run(mode, func(t *testing.T) {
			rb := BitmapOf(ref...)
			switch mode {
			case "native":
			case "runoptimized":
				rb.RunOptimize()
			default:
				iterForceContainerType(rb, mode)
			}
			require.EqualValues(t, len(ref), rb.GetCardinality())

			got := make([]uint32, 0, len(ref))
			for it := rb.Iterator(); it.HasNext(); {
				v := it.PeekNext()
				if it.Next() != v {
					t.Fatal("PeekNext and Next disagree")
				}
				got = append(got, v)
			}
			iterEqualUint32s(t, ref, got, "Iterator")

			got = got[:0]
			for it := rb.ReverseIterator(); it.HasNext(); {
				got = append(got, it.Next())
			}
			iterEqualUint32s(t, reversed, got, "ReverseIterator")

			got = got[:0]
			for v := range Values(rb) {
				got = append(got, v)
			}
			iterEqualUint32s(t, ref, got, "Values")

			got = got[:0]
			for v := range Backward(rb) {
				got = append(got, v)
			}
			iterEqualUint32s(t, reversed, got, "Backward")

			got = got[:0]
			rb.Iterate(func(x uint32) bool {
				got = append(got, x)
				return true
			})
			iterEqualUint32s(t, ref, got, "Iterate")

			iterEqualUint32s(t, ref, rb.ToArray(), "ToArray")

			// batch reads; 0 means "one container at a time"
			for _, step := range []int{1, 2, 7, len(ref) - 1, len(ref), 0} {
				it := rb.ManyIterator()
				got = got[:0]
				rest := ref
				full := make([]uint32, max(step, 65536))
				for len(rest) > 0 {
					ask := step
					if step == 0 {
						ask = 0
						for _, v := range rest {
							if v>>16 != rest[0]>>16 {
								break
							}
							ask++
						}
					}
					n := it.NextMany(full[:ask])
					if n != min(ask, len(rest)) {
						t.Fatalf("step %d: NextMany returned %d, expected %d", step, n, min(ask, len(rest)))
					}
					got = append(got, full[:n]...)
					rest = rest[n:]
				}
				iterEqualUint32s(t, ref, got, fmt.Sprintf("NextMany step %d", step))
				buf := make([]uint32, 8)
				assert.Equal(t, 0, it.NextMany(buf))
				assert.Equal(t, 0, it.NextMany(buf))
			}
		})
	}
}

func TestIteratorAdvanceToMaxUint32(t *testing.T) {
	present := BitmapOf(0, 65535, 65536, MaxUint32)
	inRun := BitmapOf(0, 65535, 65536)
	inRun.AddRange(0xFFFFFF00, 1<<32)
	inRun.RunOptimize()
	require.True(t, inRun.HasRunCompression())
	for name, rb := range map[string]*Bitmap{"array": present, "run": inRun} {
		t.Run(name, func(t *testing.T) {
			it := rb.Iterator()
			it.AdvanceIfNeeded(MaxUint32)
			require.True(t, it.HasNext())
			assert.Equal(t, uint32(MaxUint32), it.PeekNext())
			assert.Equal(t, uint32(MaxUint32), it.Next())
			assert.False(t, it.HasNext())
		})
	}
	it := BitmapOf(1, 2).Iterator()
	it.AdvanceIfNeeded(MaxUint32)
	assert.False(t, it.HasNext())
	it.AdvanceIfNeeded(MaxUint32)
	assert.False(t, it.HasNext())
}

func TestIteratorAdvanceIntoGaps(t *testing.T) {
	rb := New()
	rb.AddRange(2000000, 2200000)
	rb.AddRange(4000000, 4300000)
	it := rb.Iterator()
	assert.EqualValues(t, 2000000, it.PeekNext())
	it.AdvanceIfNeeded(2100000)
	assert.EqualValues(t, 2100000, it.PeekNext())
	it.AdvanceIfNeeded(2300000)
	assert.EqualValues(t, 4000000, it.PeekNext())
	it.AdvanceIfNeeded(4000000)
	assert.EqualValues(t, 4000000, it.PeekNext())

	rb.AddRange(6000000, 6400000)
	for _, optimize := range []bool{false, true} {
		if optimize {
			rb.RunOptimize()
		}
		it = rb.Iterator()
		it.AdvanceIfNeeded(2100000)
		// target is in the container that holds the end of range 2, past its last value
		it.AdvanceIfNeeded(4325376 - 5)
		assert.EqualValues(t, 6000000, it.PeekNext())
		it = rb.Iterator()
		// target is in a key with no container
		it.AdvanceIfNeeded(4325376 + 5)
		assert.EqualValues(t, 6000000, it.PeekNext())
		assert.EqualValues(t, 6000000, it.Next())
	}

	// Java report128
	rb = BitmapOf(59798854, 91274955, 97569495, 101993170)
	it = rb.Iterator()
	it.AdvanceIfNeeded(100620278)
	require.True(t, it.HasNext())
	assert.EqualValues(t, 101993170, it.Next())
	assert.False(t, it.HasNext())
}

func TestIteratorAdvanceStepwise(t *testing.T) {
	t.Run("bitmap containers", func(t *testing.T) {
		rb := New()
		for i := uint32(0); i < 100000; i++ {
			rb.Add(2 * i)
		}
		require.EqualValues(t, 4, rb.Stats().Containers)
		require.EqualValues(t, 3, rb.Stats().BitmapContainers)
		for i := uint32(0); i < 100000; i += 7 {
			it := rb.Iterator()
			it.AdvanceIfNeeded(2 * i)
			if it.PeekNext() != 2*i || it.Next() != 2*i {
				t.Fatalf("advance to %d", 2*i)
			}
		}
	})
	t.Run("run container", func(t *testing.T) {
		rb := New()
		rb.AddRange(4, 100000)
		rb.RunOptimize()
		for i := uint32(4); i < 100000; i += 5 {
			it := rb.Iterator()
			it.AdvanceIfNeeded(i)
			if it.PeekNext() != i || it.Next() != i {
				t.Fatalf("advance to %d", i)
			}
		}
	})
	t.Run("mixed containers", func(t *testing.T) {
		rb := New()
		for i := uint32(0); i < 66000; i++ {
			rb.Add(i % 3)
		}
		rb.AddRange(100000, 200000)
		for i := uint32(300000); i < 500000; i += 100 {
			rb.Add(i)
		}
		rb.AddRange(600000, 700000)
		for i := uint32(800000); i < 900000; i += 7 {
			rb.Add(i)
		}
		for _, optimize := range []bool{false, true} {
			if optimize {
				rb.RunOptimize()
			}
			values := rb.ToArray()
			it := rb.Iterator()
			for k, v := range values {
				if k%3 != 0 {
					continue
				}
				it.AdvanceIfNeeded(v)
				if !it.HasNext() || it.PeekNext() != v {
					t.Fatalf("advance to present %d", v)
				}
				it.AdvanceIfNeeded(v + 1)
				if k+1 < len(values) {
					if !it.HasNext() || it.PeekNext() != values[k+1] {
						t.Fatalf("advance past %d", v)
					}
				} else if it.HasNext() {
					t.Fatal("expected exhausted iterator")
				}
			}
		}
	})
}

func TestNextManyZeroLengthBuffer(t *testing.T) {
	rb := New()
	rb.AddRange(0, 10000)
	it := rb.ManyIterator()
	assert.Equal(t, 0, it.NextMany(make([]uint32, 0)))
	assert.Equal(t, 0, it.NextMany(nil))
	buf := make([]uint32, 10000)
	require.Equal(t, 10000, it.NextMany(buf))
	for i, v := range buf {
		if v != uint32(i) {
			t.Fatalf("buf[%d] = %d", i, v)
		}
	}
	assert.Equal(t, 0, it.NextMany(buf))
}

func TestIteratorsOnEmptyBitmap(t *testing.T) {
	rb := New()
	buf := make([]uint32, 1024)
	assert.Equal(t, 0, rb.ManyIterator().NextMany(buf))
	assert.False(t, rb.Iterator().HasNext())
	assert.False(t, rb.ReverseIterator().HasNext())
	it := rb.Iterator()
	it.AdvanceIfNeeded(0)
	assert.False(t, it.HasNext())
	it.AdvanceIfNeeded(MaxUint32)
	assert.False(t, it.HasNext())
	rb.Add(1)
	m := rb.ManyIterator()
	assert.Equal(t, 1, m.NextMany(buf))
	assert.EqualValues(t, 1, buf[0])
	assert.Equal(t, 0, m.NextMany(buf))
}

func TestNextManyThroughRunContainers(t *testing.T) {
	rb := New()
	rb.AddRange(65530, 65540)
	rb.AddRange(100000, 300000)
	rb.Add(MaxUint32)
	rb.RunOptimize()
	require.True(t, rb.HasRunCompression())
	expected := rb.ToArray()
	card := len(expected)
	for _, size := range []int{1, 2, 7, card - 1, card} {
		it := rb.ManyIterator()
		buf := make([]uint32, size)
		got := make([]uint32, 0, card)
		for n := it.NextMany(buf); n != 0; n = it.NextMany(buf) {
			got = append(got, buf[:n]...)
		}
		iterEqualUint32s(t, expected, got, fmt.Sprintf("size %d", size))
		assert.Equal(t, 0, it.NextMany(buf))
		assert.Equal(t, 0, it.NextMany(buf))
	}
}

func TestReverseIteratorManyContainers(t *testing.T) {
	const n = 32769
	rb := New()
	for i := uint32(0); i < n; i++ {
		rb.Add(i << 16)
	}
	it := rb.ReverseIterator()
	require.True(t, it.HasNext())
	assert.EqualValues(t, uint32(n-1)<<16, it.Next())
	count := 1
	for it.HasNext() {
		it.Next()
		count++
	}
	assert.Equal(t, n, count)
	count = 0
	for range Backward(rb) {
		count++
	}
	assert.Equal(t, n, count)
	buf := make([]uint32, 1000)
	count = 0
	m := rb.ManyIterator()
	for k := m.NextMany(buf); k != 0; k = m.NextMany(buf) {
		count += k
	}
	assert.Equal(t, n, count)
}

// Java issue 475: iteration over a run ending at 0x7FFFFFFF or 0xFFFFFFFF.
func TestIterateRunEndingAtSignedAndUnsignedMax(t *testing.T) {
	for _, last := range []uint32{0x7FFFFFFF, 0xFFFFFFFF} {
		rb := New()
		rb.AddRange(uint64(last)-3, uint64(last)+1)
		rb.RunOptimize()
		require.True(t, rb.HasRunCompression())
		expected := []uint32{last - 3, last - 2, last - 1, last}
		var got []uint32
		rb.Iterate(func(x uint32) bool {
			assert.True(t, rb.Contains(x))
			got = append(got, x)
			return true
		})
		assert.Equal(t, expected, got)
		got = got[:0]
		for v := range Values(rb) {
			got = append(got, v)
		}
		assert.Equal(t, expected, got)
		got = got[:0]
		for it := rb.Iterator(); it.HasNext(); {
			got = append(got, it.Next())
		}
		assert.Equal(t, expected, got)
	}
}

func TestIteratorInitializeOnAnotherBitmap(t *testing.T) {
	a := New()
	a.AddRange(1, 1001)
	b := BitmapOf(5, 70000)
	empty := New()

	it := a.Iterator().(*intIterator)
	for i := 0; i < 500; i++ {
		it.Next()
	}
	it.Initialize(b)
	var got []uint32
	for it.HasNext() {
		got = append(got, it.Next())
	}
	assert.Equal(t, []uint32{5, 70000}, got)
	it.Initialize(empty)
	assert.False(t, it.HasNext())

	rit := a.ReverseIterator().(*intReverseIterator)
	for i := 0; i < 500; i++ {
		rit.Next()
	}
	rit.Initialize(b)
	got = got[:0]
	for rit.HasNext() {
		got = append(got, rit.Next())
	}
	assert.Equal(t, []uint32{70000, 5}, got)
	rit.Initialize(empty)
	assert.False(t, rit.HasNext())

	mit := a.ManyIterator().(*manyIntIterator)
	buf := make([]uint32, 500)
	require.Equal(t, 500, mit.NextMany(buf))
	mit.Initialize(b)
	require.Equal(t, 2, mit.NextMany(buf))
	assert.Equal(t, []uint32{5, 70000}, buf[:2])
	mit.Initialize(empty)
	assert.Equal(t, 0, mit.NextMany(buf))
}

func TestStatsRunAndMixedContainers(t *testing.T) {
	rb := New()
	rb.AddRange(0, 1000)
	st := rb.Stats()
	assert.EqualValues(t, 1, st.Containers)
	assert.EqualValues(t, 1, st.RunContainers)
	assert.EqualValues(t, 1000, st.RunContainerValues)
	assert.EqualValues(t, perIntervalRc16Size*1+baseRc16Size, st.RunContainerBytes)
	assert.EqualValues(t, 1000, st.Cardinality)

	// mixed bitmap from Java's BitmapAnalyserTest
	rb = BitmapOf(1, 3, 6, 26, 110, 1024)
	rb.AddRange(70000, 80000)
	for i := uint32(5 << 16); i < 6<<16; i++ {
		rb.Add(i)
	}
	// Go converts a bitmap container that becomes full into a run container,
	// so the full key 5 is reported as a run (Java reports a bitmap).
	for i := uint32(7 << 16); i < 8<<16; i += 2 {
		rb.Add(i)
	}
	st = rb.Stats()
	assert.EqualValues(t, 4, st.Containers)
	assert.EqualValues(t, 75542+32768, st.Cardinality)
	assert.EqualValues(t, 1, st.ArrayContainers)
	assert.EqualValues(t, 6, st.ArrayContainerValues)
	assert.EqualValues(t, 12, st.ArrayContainerBytes)
	assert.EqualValues(t, 2, st.RunContainers)
	assert.EqualValues(t, 10000+65536, st.RunContainerValues)
	assert.EqualValues(t, 2*(perIntervalRc16Size+baseRc16Size), st.RunContainerBytes)
	assert.EqualValues(t, 1, st.BitmapContainers)
	assert.EqualValues(t, 32768, st.BitmapContainerValues)
	assert.EqualValues(t, 8192, st.BitmapContainerBytes)
}

func TestStringEdgeCases(t *testing.T) {
	assert.Equal(t, "{}", New().String())
	assert.Equal(t, "{1,2,3,4294967295}", BitmapOf(1, 2, 3, MaxUint32).String())

	rb := New()
	rb.AddRange(0, 1<<20)
	s := rb.String()
	assert.True(t, strings.HasPrefix(s, "{0,1,2,"))
	assert.True(t, strings.HasSuffix(s, ",...}"))
	assert.Less(t, len(s), 4*1024*1024)
}

func TestSizeInBytesEstimates(t *testing.T) {
	// Java issue 566: a few sparse values must not cost more than a bitset
	rb := BitmapOf(17, 1023, 2048, 3999, 4999)
	assert.Less(t, rb.GetSizeInBytes()*8, uint64(5000))

	rb = New()
	for i := uint32(0); i < 100000; i++ {
		rb.Add(i)
	}
	before := rb.GetSizeInBytes()
	rb.RunOptimize()
	assert.Less(t, rb.GetSizeInBytes(), before)
}
