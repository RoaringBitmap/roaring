package roaring

import "math/bits"

// bit population count, uses Go 1.9 math/bits library
func popcount(x uint64) uint64 {
	return uint64(bits.OnesCount64(x))
}

func popcntSlice(s []uint64) uint64 {
	cnt := uint64(0)
	for _, x := range s {
		cnt += popcount(x)
	}
	return cnt
}

func popcntMaskSlice(s, m []uint64) uint64 {
	cnt := uint64(0)
	for i := range s {
		cnt += popcount(s[i] &^ m[i])
	}
	return cnt
}

func popcntAndSlice(s, m []uint64) uint64 {
	cnt := uint64(0)
	for i := range s {
		cnt += popcount(s[i] & m[i])
	}
	return cnt
}

func popcntOrSlice(s, m []uint64) uint64 {
	cnt := uint64(0)
	for i := range s {
		cnt += popcount(s[i] | m[i])
	}
	return cnt
}

func popcntXorSlice(s, m []uint64) uint64 {
	cnt := uint64(0)
	for i := range s {
		cnt += popcount(s[i] ^ m[i])
	}
	return cnt
}
