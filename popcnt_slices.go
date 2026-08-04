package roaring

import "math/bits"

func popcntSliceGo(s []uint64) uint64 {
	cnt := uint64(0)
	for _, x := range s {
		cnt += uint64(bits.OnesCount64(x))
	}
	return cnt
}

func popcntMaskSliceGo(s, m []uint64) uint64 {
	cnt := uint64(0)
	for i := range s {
		cnt += uint64(bits.OnesCount64(s[i] &^ m[i]))
	}
	return cnt
}

func popcntAndSliceGo(s, m []uint64) uint64 {
	cnt := uint64(0)
	for i := range s {
		cnt += uint64(bits.OnesCount64(s[i] & m[i]))
	}
	return cnt
}

// andPopcntSliceGo writes s & m to dst and returns its cardinality. All slices
// must have the same length.
func andPopcntSliceGo(dst, s, m []uint64) uint64 {
	for i := range dst {
		dst[i] = s[i] & m[i]
	}
	return popcntSlice(dst)
}

func popcntOrSliceGo(s, m []uint64) uint64 {
	cnt := uint64(0)
	for i := range s {
		cnt += uint64(bits.OnesCount64(s[i] | m[i]))
	}
	return cnt
}

func popcntXorSliceGo(s, m []uint64) uint64 {
	cnt := uint64(0)
	for i := range s {
		cnt += uint64(bits.OnesCount64(s[i] ^ m[i]))
	}
	return cnt
}
