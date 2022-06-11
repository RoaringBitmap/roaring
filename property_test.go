package roaring

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPropertyOr(t *testing.T) {
	// Make test deterministic.
	rand := rand.New(rand.NewSource(0))

	testFn := func(t *testing.T) {
		roaring1, roaring2, reference1, reference2 := genPropTestInputs(rand)

		roaring1.Or(roaring2)
		reference1.Or(reference2)

		assertRoaringEqualsReference(t, roaring1, reference1)
	}

	for i := 0; i < 1000; i++ {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			testFn(t)
		})
	}
}

func TestPropertyAnd(t *testing.T) {
	// Make test deterministic.
	rand := rand.New(rand.NewSource(0))

	testFn := func(t *testing.T) {
		roaring1, roaring2, reference1, reference2 := genPropTestInputs(rand)

		roaring1.And(roaring2)
		reference1.And(reference2)

		assertRoaringEqualsReference(t, roaring1, reference1)
	}

	for i := 0; i < 100; i++ {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			testFn(t)
		})
	}
}

func genPropTestInputs(rand *rand.Rand) (*Bitmap, *Bitmap, *reference, *reference) {
	var (
		aSize   = rand.Intn(1_000)
		bSize   = rand.Intn(1_000)
		aValues = make([]uint32, 0, aSize)
		bValues = make([]uint32, 0, bSize)
	)
	for j := 0; j < aSize; j++ {
		aValues = append(aValues, rand.Uint32())
	}
	for j := 0; j < bSize; j++ {
		bValues = append(bValues, rand.Uint32())
	}

	var (
		roaring1 = New()
		roaring2 = New()

		reference1 = newReference()
		reference2 = newReference()
	)
	for _, v := range aValues {
		if rand.Intn(20) == 0 {
			rangeStart := rand.Uint32()
			roaring1.AddRange(uint64(rangeStart), uint64(rangeStart+100))
			reference1.AddRange(uint64(rangeStart), uint64(rangeStart+100))
			continue
		}

		roaring1.Add(v)
		reference1.Add(v)
	}

	for _, v := range bValues {
		if rand.Intn(20) == 0 {
			rangeStart := rand.Uint32()
			roaring2.AddRange(uint64(rangeStart), uint64(rangeStart+100))
			reference2.AddRange(uint64(rangeStart), uint64(rangeStart+100))
			continue
		}

		roaring2.Add(v)
		reference2.Add(v)
	}

	return roaring1, roaring2, reference1, reference2
}

// reference is a reference implementation that can be used in property tests
// to assert the correctness of the actual roaring implementation.
type reference struct {
	m map[uint32]struct{}
}

func newReference() *reference {
	return &reference{
		m: make(map[uint32]struct{}),
	}
}

func (r *reference) Add(x uint32) {
	r.m[x] = struct{}{}
}

func (r *reference) AddRange(start, end uint64) {
	for i := start; i < end; i++ {
		r.m[uint32(i)] = struct{}{}
	}
}

func (r *reference) Contains(x uint32) bool {
	_, ok := r.m[x]
	return ok
}

func (r *reference) Cardinality() uint64 {
	return uint64(len(r.m))
}

func (r *reference) Or(other *reference) {
	for v := range other.m {
		r.m[v] = struct{}{}
	}
}

func (r *reference) And(other *reference) {
	newM := map[uint32]struct{}{}
	for v := range other.m {
		if _, ok := r.m[v]; ok {
			newM[v] = struct{}{}
		}
	}
	r.m = newM
}

func assertRoaringEqualsReference(
	t *testing.T,
	roaring *Bitmap,
	reference *reference,
) {
	// round-trip the roaring bitmap to ensure our property still holds
	// true after a round of ser/der.
	rounedTrippedRoaring := roundTripRoaring(t, roaring)
	require.Equal(t, reference.Cardinality(), rounedTrippedRoaring.Stats().Cardinality)
	roaring.Iterate(func(x uint32) bool {
		require.True(t, reference.Contains(x))
		return true
	})
}
