//go:build arm64 && !gccgo && !appengine
// +build arm64,!gccgo,!appengine

package roaring

// uniqshuf[m] compacts the lanes not set in m. Initialize it before any
// package-level bitmap unions, which can run before init functions.
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
	return t
}

//go:noescape
func union2by2scalar(set1 []uint16, set2 []uint16, buffer []uint16) (size int)

//go:noescape
func unionPartKernelNEON(set1, set2, buffer []uint16, shuf *byte) (outLen, pos1, pos2 int)

func union2by2(set1 []uint16, set2 []uint16, buffer []uint16) int {
	// Small balanced inputs favor scalar; unequal inputs need enough total work.
	if len(set1) < 32 || len(set2) < 32 || len(set1)+len(set2) < 128 {
		return union2by2scalar(set1, set2, buffer)
	}
	// lazyorArray passes a zero-length buffer with capacity.
	buffer = buffer[:cap(buffer)]
	// In-place self-union aliases set2 at the output base, with set1 an
	// identical copy. Scalar reads stay ahead of writes in this geometry.
	if &buffer[0] == &set2[0] {
		return union2by2scalar(set1, set2, buffer)
	}
	outLen, pos1, pos2 := unionPartKernelNEON(set1, set2, buffer, &uniqshuf[0])
	// The remaining values are strictly greater than the last emitted value.
	return outLen + union2by2scalar(set1[pos1:], set2[pos2:], buffer[outLen:])
}
