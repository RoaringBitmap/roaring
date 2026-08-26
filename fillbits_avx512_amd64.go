//go:build amd64 && !appengine
// +build amd64,!appengine

package roaring

//go:noescape
func fillLeastSignificant16bitsAVX512(bitmap []uint64, x []uint32, pos int, mask uint32) int

//go:noescape
func _hasAVX512() bool

// AVX-512 is used only when the operating system saves the ZMM state and the
// CPU exposes AVX-512 Foundation. The scalar implementation remains available
// for CPUs without that feature and for appengine builds.
var useAVX512 = _hasAVX512()
