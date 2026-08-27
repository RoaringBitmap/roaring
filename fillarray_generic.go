//go:build !amd64 || appengine
// +build !amd64 appengine

package roaring

// On these targets useVectorFill is a compile-time constant false, so the
// branches that would reach these functions are eliminated and none of them is
// ever executed. They exist so that the shared call sites in bitmapcontainer.go
// and util.go compile, and they delegate back to the scalar paths so that they
// remain correct if anything ever does call them. There is no recursion: the
// guard in each caller is false at compile time on these targets.

func fillArrayVector(bitmap []uint64, container []uint16) {
	fillArrayScalar(bitmap, container)
}

func fillArraySkipVector(bitmap []uint64, container []uint16) {
	fillArrayScalar(bitmap, container)
}

func fillArrayANDVector(container []uint16, bitmap1, bitmap2 []uint64) {
	fillArrayAND(container, bitmap1, bitmap2)
}

func fillArrayANDNOTVector(container []uint16, bitmap1, bitmap2 []uint64) {
	fillArrayANDNOT(container, bitmap1, bitmap2)
}

func fillArrayXORVector(container []uint16, bitmap1, bitmap2 []uint64) {
	fillArrayXOR(container, bitmap1, bitmap2)
}
