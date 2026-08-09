//go:build arm64 && !gccgo && !appengine
// +build arm64,!gccgo,!appengine

package roaring

//go:noescape
func intersectCardKernelNEON(set1, set2 []uint16) (card, pos1, pos2 int)

//go:noescape
func intersectKernelNEON(set1, set2, buffer []uint16, shuf *byte, spill *[8]uint16) (outLen, pos1, pos2, spilled int)

// Below this the kernel setup usually loses to the scalar path.
// Duplicate values: sorted in-cap result, never a panic; sizes can over-report.
const (
	neonIntersectCardThreshold = 16
	neonIntersectThreshold     = 16
)

func intersection2by2(
	set1 []uint16,
	set2 []uint16,
	buffer []uint16,
) int {
	if len(set1)*64 < len(set2) {
		return onesidedgallopingintersect2by2(set1, set2, buffer)
	} else if len(set2)*64 < len(set1) {
		return onesidedgallopingintersect2by2(set2, set1, buffer)
	}
	if len(set1) < neonIntersectThreshold || len(set2) < neonIntersectThreshold {
		return localintersect2by2(set1, set2, buffer)
	}
	if set1[len(set1)-1] < set2[0] || set2[len(set2)-1] < set1[0] {
		return 0
	}
	// andArray passes len 0; COW cloning keeps iandArray's spare cap private.
	// buffer may alias set1 (iandArray) or both inputs (self-And), not set2 alone.
	buffer = buffer[:cap(buffer)]
	var spill [8]uint16
	outLen, pos1, pos2, spilled := intersectKernelNEON(set1, set2, buffer, &uniqshuf[0], &spill)
	if spilled != 0 {
		// set1 may alias buffer: drain from this copy, never reread.
		i := 0
		for i < 8 && pos2 < len(set2) && outLen < len(buffer) {
			switch {
			case spill[i] < set2[pos2]:
				i++
			case set2[pos2] < spill[i]:
				pos2++
			default:
				buffer[outLen] = spill[i]
				outLen++
				i++
				pos2++
			}
		}
	}
	// Every kernel exit leaves min(len(tail1), len(tail2)) <= 7.
	if len(buffer)-outLen >= 7 {
		return outLen + localintersect2by2(set1[pos1:], set2[pos2:], buffer[outLen:])
	}
	return intersectTailBounded(set1[pos1:], set2[pos2:], buffer, outLen)
}

// Tail merge for a near-full buffer; only duplicate input can overfill it.
func intersectTailBounded(tail1, tail2, buffer []uint16, outLen int) int {
	i, j := 0, 0
	for i < len(tail1) && j < len(tail2) {
		s1, s2 := tail1[i], tail2[j]
		switch {
		case s1 < s2:
			i++
		case s2 < s1:
			j++
		default:
			if outLen == len(buffer) {
				return outLen
			}
			buffer[outLen] = s1
			outLen++
			i++
			j++
		}
	}
	return outLen
}

func intersection2by2Cardinality(
	set1 []uint16,
	set2 []uint16,
) int {
	if len(set1)*64 < len(set2) {
		return onesidedgallopingintersect2by2Cardinality(set1, set2)
	} else if len(set2)*64 < len(set1) {
		return onesidedgallopingintersect2by2Cardinality(set2, set1)
	}
	if len(set1) < neonIntersectCardThreshold || len(set2) < neonIntersectCardThreshold {
		return localintersect2by2Cardinality(set1, set2)
	}
	if set1[len(set1)-1] < set2[0] || set2[len(set2)-1] < set1[0] {
		return 0
	}
	card, pos1, pos2 := intersectCardKernelNEON(set1, set2)
	return card + localintersect2by2Cardinality(set1[pos1:], set2[pos2:])
}

var uniqshuf = buildUniqshuf()

func buildUniqshuf() (t [256 * 16]byte) {
	for m := 0; m < 256; m++ {
		pos := 0
		for lane := 0; lane < 8; lane++ {
			if m&(1<<lane) == 0 {
				t[m*16+pos*2] = byte(2 * lane)
				t[m*16+pos*2+1] = byte(2*lane + 1)
				pos++
			}
		}
		for ; pos < 8; pos++ {
			t[m*16+pos*2] = 0xFF
			t[m*16+pos*2+1] = 0xFF
		}
	}
	return
}
