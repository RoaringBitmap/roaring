package roaring64

// to run just these tests: go test -run TestFastAggregations*

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFastAggregationsAdvanced_run(t *testing.T) {
	rb1 := NewBitmap()
	rb2 := NewBitmap()
	rb3 := NewBitmap()
	for i := uint64(500); i < 75000; i++ {
		rb1.Add(i)
	}
	for i := uint64(0); i < 1000000; i += 7 {
		rb2.Add(i)
	}
	for i := uint64(0); i < 1000000; i += 1001 {
		rb3.Add(i)
	}
	for i := uint64(1000000); i < 2000000; i += 1001 {
		rb1.Add(i)
	}
	for i := uint64(1000000); i < 2000000; i += 3 {
		rb2.Add(i)
	}
	for i := uint64(1000000); i < 2000000; i += 7 {
		rb3.Add(i)
	}
	rb1.RunOptimize()
	rb1.Or(rb2)
	rb1.Or(rb3)
	bigand := And(And(rb1, rb2), rb3)

	assert.True(t, FastOr(rb1, rb2, rb3).Equals(rb1))
	assert.True(t, FastAnd(rb1, rb2, rb3).Equals(bigand))
}
