//go:build arm64 && !gccgo && !appengine
// +build arm64,!gccgo,!appengine

package roaring

//go:noescape
func xorKernelNEON(set1, set2, buffer []uint16, shuf *byte) (outLen, pos1, pos2 int)

// The kernel needs 16 unread elements on both sides and a fresh buffer with
// cap(buffer) >= len(set1)+len(set2). Below 32 its setup usually loses to the
// scalar path. Duplicate values get unspecified results; stores stay within capacity.
const neonXorThreshold = 32

func exclusiveUnion2by2(set1 []uint16, set2 []uint16, buffer []uint16) int {
	if len(set1) < neonXorThreshold || len(set2) < neonXorThreshold || cap(buffer) < len(set1)+len(set2) {
		return localexclusiveUnion2by2(set1, set2, buffer)
	}
	// Callers such as xorArray pass a zero-length buffer with capacity.
	buffer = buffer[:cap(buffer)]
	outLen, pos1, pos2 := xorKernelNEON(set1, set2, buffer, &uniqshuf[0])
	// Everything unread is above the last emitted value, so the tails need no seam check.
	return outLen + localexclusiveUnion2by2(set1[pos1:], set2[pos2:], buffer[outLen:])
}
