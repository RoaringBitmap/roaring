//go:build arm64 && !gccgo && !appengine
// +build arm64,!gccgo,!appengine

package roaring

func difference(set1 []uint16, set2 []uint16, buffer []uint16) int {
	return localdifference(set1, set2, buffer)
}
