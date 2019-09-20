package roaring

// to run just these tests: go test -run TestSetUtil*

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
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

func BenchmarkIntersectAlgorithms(b *testing.B) {
	sz1 := 1000
	s1 := make([]uint16, sz1)

	sz2 := MaxUint16
	s2 := make([]uint16, sz2)

	for i := 0; i < sz2; i++ {
		s2[i] = uint16(i)
	}

	r := rand.New(rand.NewSource(0))
	k := 0

	for i := 0; i < sz1 && k < sz2; i++ {
		n := r.Intn(100)
		k += n

		// prevent adding duplicates
		if n == 0 && i > 0 {
			k++
		}

		s1[i] = uint16(s2[k])
	}

	buf := make([]uint16, sz1+sz2)

	b.Run("onesidedgallopingintersect2by2", func(b *testing.B) {
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			onesidedgallopingintersect2by2(s1, s2, buf)
		}
	})

	b.Run("shotgun4", func(b *testing.B) {
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			shotgun4Intersect(s1, s2, buf)
		}
	})
}
