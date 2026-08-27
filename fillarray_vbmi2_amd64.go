//go:build amd64 && !appengine
// +build amd64,!appengine

package roaring

// Implemented in fillarray_vbmi2_amd64.s. Only called when useVectorFill (see
// fillbits_vbmi2_amd64.go) reports that the CPU can run them.

//go:noescape
func fillArrayVector(bitmap []uint64, container []uint16)

//go:noescape
func fillArraySkipVector(bitmap []uint64, container []uint16)

//go:noescape
func fillArrayANDVector(container []uint16, bitmap1, bitmap2 []uint64)

//go:noescape
func fillArrayANDNOTVector(container []uint16, bitmap1, bitmap2 []uint64)

//go:noescape
func fillArrayXORVector(container []uint16, bitmap1, bitmap2 []uint64)
