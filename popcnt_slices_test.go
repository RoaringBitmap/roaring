//go:build amd64 && !appengine && !go1.9
// +build amd64,!appengine,!go1.9

// This file tests the popcnt functions

package roaring

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPopcntSlice(t *testing.T) {
	s := []uint64{2, 3, 5, 7, 11, 13, 17, 19, 23, 29}
	resGo := popcntSliceGo(s)
	resAsm := popcntSliceAsm(s)
	res := popcntSlice(s)

	assert.Equal(t, resGo, resAsm)
	assert.Equal(t, resGo, res)
}

func TestPopcntMaskSlice(t *testing.T) {
	s := []uint64{2, 3, 5, 7, 11, 13, 17, 19, 23, 29}
	m := []uint64{31, 37, 41, 43, 47, 53, 59, 61, 67, 71}
	resGo := popcntMaskSliceGo(s, m)
	resAsm := popcntMaskSliceAsm(s, m)
	res := popcntMaskSlice(s, m)

	assert.Equal(t, resGo, resAsm)
	assert.Equal(t, resGo, res)
}

func TestPopcntAndSlice(t *testing.T) {
	s := []uint64{2, 3, 5, 7, 11, 13, 17, 19, 23, 29}
	m := []uint64{31, 37, 41, 43, 47, 53, 59, 61, 67, 71}
	resGo := popcntAndSliceGo(s, m)
	resAsm := popcntAndSliceAsm(s, m)
	res := popcntAndSlice(s, m)

	assert.Equal(t, resGo, resAsm)
	assert.Equal(t, resGo, res)
}

func TestPopcntOrSlice(t *testing.T) {
	s := []uint64{2, 3, 5, 7, 11, 13, 17, 19, 23, 29}
	m := []uint64{31, 37, 41, 43, 47, 53, 59, 61, 67, 71}
	resGo := popcntOrSliceGo(s, m)
	resAsm := popcntOrSliceAsm(s, m)
	res := popcntOrSlice(s, m)

	assert.Equal(t, resGo, resAsm)
	assert.Equal(t, resGo, res)
}

func TestPopcntXorSlice(t *testing.T) {
	s := []uint64{2, 3, 5, 7, 11, 13, 17, 19, 23, 29}
	m := []uint64{31, 37, 41, 43, 47, 53, 59, 61, 67, 71}
	resGo := popcntXorSliceGo(s, m)
	resAsm := popcntXorSliceAsm(s, m)
	res := popcntXorSlice(s, m)

	assert.Equal(t, resGo, resAsm)
	assert.Equal(t, resGo, res)
}
