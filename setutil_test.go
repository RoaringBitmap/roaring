package roaring

// to run just these tests: go test -run TestSetUtil*

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"sort"
	"testing"
)

func TestSetUtilDifference(t *testing.T) {
	data1 := []uint16{0, 1, 2, 3, 4, 9}
	data2 := []uint16{2, 3, 4, 5, 8, 9, 11}
	result := make([]uint16, 0, len(data1)+len(data2))
	expectedresult := []uint16{0, 1}
	nl := difference(data1, data2, result)
	result = result[:nl]

	assert.Equal(t, expectedresult, result)

	expectedresult = []uint16{5, 8, 11}
	nl = difference(data2, data1, result)
	result = result[:nl]

	assert.Equal(t, expectedresult, result)
}

func TestSetUtilUnion(t *testing.T) {
	data1 := []uint16{0, 1, 2, 3, 4, 9}
	data2 := []uint16{2, 3, 4, 5, 8, 9, 11}
	result := make([]uint16, 0, len(data1)+len(data2))
	expectedresult := []uint16{0, 1, 2, 3, 4, 5, 8, 9, 11}
	nl := union2by2(data1, data2, result)
	result = result[:nl]

	assert.Equal(t, expectedresult, result)

	nl = union2by2(data2, data1, result)
	result = result[:nl]

	assert.Equal(t, expectedresult, result)
}

func TestSetUtilExclusiveUnion(t *testing.T) {
	data1 := []uint16{0, 1, 2, 3, 4, 9}
	data2 := []uint16{2, 3, 4, 5, 8, 9, 11}
	result := make([]uint16, 0, len(data1)+len(data2))
	expectedresult := []uint16{0, 1, 5, 8, 11}
	nl := exclusiveUnion2by2(data1, data2, result)
	result = result[:nl]

	assert.Equal(t, expectedresult, result)

	nl = exclusiveUnion2by2(data2, data1, result)
	result = result[:nl]

	assert.Equal(t, expectedresult, result)
}

func TestSetUtilIntersection(t *testing.T) {
	data1 := []uint16{0, 1, 2, 3, 4, 9}
	data2 := []uint16{2, 3, 4, 5, 8, 9, 11}
	result := make([]uint16, 0, len(data1)+len(data2))
	expectedresult := []uint16{2, 3, 4, 9}
	nl := intersection2by2(data1, data2, result)
	result = result[:nl]
	result = result[:len(expectedresult)]

	assert.Equal(t, expectedresult, result)

	nl = intersection2by2(data2, data1, result)
	result = result[:nl]

	assert.Equal(t, expectedresult, result)

	data1 = []uint16{4}
	data2 = make([]uint16, 10000)

	for i := range data2 {
		data2[i] = uint16(i)
	}

	result = make([]uint16, 0, len(data1)+len(data2))
	expectedresult = data1
	nl = intersection2by2(data1, data2, result)
	result = result[:nl]

	assert.Equal(t, expectedresult, result)

	nl = intersection2by2(data2, data1, result)
	result = result[:nl]

	assert.Equal(t, expectedresult, result)
}

// go test -run TestSetUtilIntersectionCases
func TestSetUtilIntersectionCases(t *testing.T) {
	cases := []struct {
		name string
		algo func(a, b, buf []uint16) int
	}{
		{
			name: "onesidedgallopingintersect2by2",
			algo: onesidedgallopingintersect2by2,
		},
		{
			name: "shotgun4Intersect",
			algo: shotgun4Intersect,
		},
	}

	data1 := []uint16{0, 3, 6, 9, 12, 15, 18}
	data2 := []uint16{0, 2, 4, 6, 8, 10, 12, 14, 16, 18}
	expected := []uint16{0, 6, 12, 18}

	for _, c := range cases {
		result := make([]uint16, 0, len(data1)+len(data2))
		n := c.algo(data1, data2, result)

		assert.Equalf(t, expected, result[:n], "failed algorithm: %s", c.name)
	}
}

func TestSetUtilBinarySearch(t *testing.T) {
	data := make([]uint16, 256)
	for i := range data {
		data[i] = uint16(2 * i)
	}
	for i := 0; i < 2*len(data); i++ {
		key := uint16(i)
		loc := binarySearch(data, key)
		if (key & 1) == 0 {
			assert.Equal(t, int(key)/2, loc)
		} else {
			assert.Equal(t, -int(key)/2-2, loc)
		}
	}
}

// go test  -bench BenchmarkIntersectAlgorithms -run -
func BenchmarkIntersectAlgorithms(b *testing.B) {
	// sz1 is the small array
	sz1 := 64 // this should not be *too* large
	s1 := make([]uint16, sz1)

	// to get more realistic results, we try different
	// large array sizes. Our benchmarks is going to be
	// an average of those...

	sz2 := 3000
	s2 := make([]uint16, sz2)

	sz3 := 2040
	s3 := make([]uint16, sz3)

	sz4 := 1200
	s4 := make([]uint16, sz4)

	r := rand.New(rand.NewSource(1234))

	// We are going to populate our large arrays with
	// random data. Importantly, we need to sort.
	// There might be a few duplicates, by random chance,
	// but it should not affect results too much.

	for i := 0; i < sz2; i++ {
		s2[i] = uint16(r.Intn(MaxUint16))
	}
	sort.Sort(uint16Slice(s2))

	for i := 0; i < sz3; i++ {
		s3[i] = uint16(r.Intn(MaxUint16))
	}
	sort.Sort(uint16Slice(s3))

	for i := 0; i < sz4; i++ {
		s4[i] = uint16(r.Intn(MaxUint16))
	}
	sort.Sort(uint16Slice(s4))

	buf := make([]uint16, sz1+sz2+sz3+sz4)
	commonseed := 123456
	r = rand.New(rand.NewSource(commonseed)) // we set the same seed in both instances

	b.Run("onesidedgallopingintersect2by2", func(b *testing.B) {

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			// this is important: you want to start with a new
			// small array each time otherwise onesidedgallopingintersect2by2
			// might benefit from nearly perfect branch prediction, making
			// the benchmark unrealistic.
			// This needs to be super fast, which it should be if sz1 is
			// small enough.
			for i := 0; i < sz1; i++ {
				// This needs to be super fast
				s1[i] = uint16(r.Intn(MaxUint16))
			}
			sort.Sort(uint16Slice(s1)) // There might be duplicates, ignore them

			onesidedgallopingintersect2by2(s1, s2, buf)
			onesidedgallopingintersect2by2(s1, s3, buf)
			onesidedgallopingintersect2by2(s1, s4, buf)

		}
	})
	r = rand.New(rand.NewSource(commonseed)) // we set the same seed in both instances

	b.Run("shotgun4", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// this is important: you want to start with a new
			// small array each time otherwise onesidedgallopingintersect2by2
			// might benefit from nearly perfect branch prediction, making
			// the benchmark unrealistic.
			// This needs to be super fast, which it should be if sz1 is
			// small enough.
			for i := 0; i < sz1; i++ {
				s1[i] = uint16(r.Intn(MaxUint16))
			}
			sort.Sort(uint16Slice(s1)) // There might be duplicates, ignore them

			shotgun4Intersect(s1, s2, buf)
			shotgun4Intersect(s1, s3, buf)
			shotgun4Intersect(s1, s4, buf)

		}
	})
}
