//go:build arm64 && !gccgo && !appengine
// +build arm64,!gccgo,!appengine

package roaring

//go:noescape
func union2by2scalar(set1 []uint16, set2 []uint16, buffer []uint16) (size int)

//go:noescape
func unionKernelNEON(set1, set2, buffer []uint16, shuf *byte, leftover *[16]uint16) (outLen, pos1, pos2, leftoverLen int)

// uniqshuf[m] is the TBL index vector that compacts the lanes not set in
// mask m to the front. A variable initializer, not init(): package-level
// initializers in other files run first and would read a zero table.
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

// Below this the kernel's setup cost usually loses to the scalar merge.
const neonUnionThreshold = 256

func union2by2(set1 []uint16, set2 []uint16, buffer []uint16) int {
	if len(set1) < neonUnionThreshold || len(set2) < neonUnionThreshold {
		return union2by2scalar(set1, set2, buffer)
	}
	return unionNEON(set1, set2, buffer)
}

// unionNEON requires at least 8 elements in each input.
func unionNEON(set1 []uint16, set2 []uint16, buffer []uint16) int {
	// Callers such as lazyorArray pass a zero-length buffer with capacity.
	buffer = buffer[:cap(buffer)]
	// iorArray's in-place self-union passes set2 and buffer sharing a
	// backing array from offset 0, with set1 a copy of set2. The kernel's
	// block stores would clobber unread set2; on identical inputs the
	// scalar merge never writes a position it has not already read.
	if &buffer[0] == &set2[0] {
		return union2by2scalar(set1, set2, buffer)
	}
	var leftover [16]uint16
	outLen, pos1, pos2, ll := unionKernelNEON(set1, set2, buffer, &uniqshuf[0], &leftover)
	// The leftovers and the exhausted input's tail are two sorted runs.
	var tmp [16]uint16
	if pos1 == len(set1)/8 {
		m := scalarMergeUnion(leftover[:ll], set1[8*pos1:], tmp[:])
		outLen += mergeUnionLookahead(tmp[:m], set2[8*pos2:], buffer[outLen:])
	} else {
		m := scalarMergeUnion(leftover[:ll], set2[8*pos2:], tmp[:])
		outLen += mergeUnionLookahead(tmp[:m], set1[8*pos1:], buffer[outLen:])
	}
	return outLen
}

// mergeUnionLookahead reads b in 8-element chunks ahead of the writes:
// iorArray aliases set1 above the write region, and writes can lead the
// b reads by up to 7 elements there.
func mergeUnionLookahead(a, b, out []uint16) int {
	i, k := 0, 0
	for j := 0; j < len(b); {
		// The copy cannot pass the read frontier in the aliased case.
		if i == len(a) {
			return k + copy(out[k:], b[j:])
		}
		var w [8]uint16
		c := copy(w[:], b[j:])
		j += c
		for x := 0; x < c; x++ {
			for i < len(a) && a[i] < w[x] {
				out[k] = a[i]
				i++
				k++
			}
			if i < len(a) && a[i] == w[x] {
				i++
			}
			out[k] = w[x]
			k++
		}
	}
	for ; i < len(a); i++ {
		out[k] = a[i]
		k++
	}
	return k
}

// scalarMergeUnion merges two sorted duplicate-free slices into out.
func scalarMergeUnion(a, b, out []uint16) int {
	i, j, k := 0, 0, 0
	for i < len(a) && j < len(b) {
		switch {
		case a[i] < b[j]:
			out[k] = a[i]
			i++
		case b[j] < a[i]:
			out[k] = b[j]
			j++
		default:
			out[k] = a[i]
			i++
			j++
		}
		k++
	}
	k += copy(out[k:], a[i:])
	k += copy(out[k:], b[j:])
	return k
}
