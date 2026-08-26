//go:build !amd64 || appengine
// +build !amd64 appengine

package roaring

const useAVX512 = false

func fillLeastSignificant16bitsAVX512(bitmap []uint64, x []uint32, pos int, mask uint32) int {
	return fillLeastSignificant16bitsScalar(bitmap, x, pos, mask)
}
