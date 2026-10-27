//go:build arm64 && !gccgo && !appengine
// +build arm64,!gccgo,!appengine

package roaring

//go:noescape
func andnotKernelNEON(set1, set2, buffer []uint16, shuf *byte, spill *[8]uint16) (outLen, pos1, pos2, spilled, accMask int)

// The kernel needs whole 8-element blocks on both sides and cap(buffer) >= len(set1).
// Below 16 its setup usually loses to the scalar path.
// Duplicate values get unspecified results; stores stay within capacity.
const neonAndnotThreshold = 16

func difference(set1 []uint16, set2 []uint16, buffer []uint16) int {
	if len(set1) < neonAndnotThreshold || len(set2) < neonAndnotThreshold {
		return localdifference(set1, set2, buffer)
	}
	if set1[len(set1)-1] < set2[0] || set2[len(set2)-1] < set1[0] {
		buffer = buffer[:len(set1)]
		copy(buffer, set1)
		return len(set1)
	}
	buffer = buffer[:cap(buffer)]
	var spill [8]uint16
	outLen, pos1, pos2, spilled, accMask := andnotKernelNEON(set1, set2, buffer, &uniqshuf[0], &spill)
	if spilled != 0 {
		// set1 may alias buffer: drain the retained block from this copy, never reread.
		i := 0
		for i < 8 && pos2 < len(set2) {
			switch {
			case spill[i] < set2[pos2]:
				if accMask&(1<<i) == 0 {
					buffer[outLen] = spill[i]
					outLen++
				}
				i++
			case set2[pos2] < spill[i]:
				pos2++
			default:
				i++
				pos2++
			}
		}
		for ; i < 8; i++ {
			if accMask&(1<<i) == 0 {
				buffer[outLen] = spill[i]
				outLen++
			}
		}
	}
	return outLen + localdifference(set1[pos1:], set2[pos2:], buffer[outLen:])
}
