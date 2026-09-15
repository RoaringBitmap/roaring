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
func union2by2(set1 []uint16, set2 []uint16, buffer []uint16) (size int)
