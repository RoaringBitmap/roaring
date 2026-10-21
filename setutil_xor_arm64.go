//go:build arm64 && !gccgo && !appengine
// +build arm64,!gccgo,!appengine

package roaring

func exclusiveUnion2by2(set1 []uint16, set2 []uint16, buffer []uint16) int {
	return localexclusiveUnion2by2(set1, set2, buffer)
}
