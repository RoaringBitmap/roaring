//go:build amd64 && !appengine
// +build amd64,!appengine

package roaring

//go:noescape
func xorSliceAVX2(dst, a, b []uint64)
