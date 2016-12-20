// +build amd64,!appengine

package roaring

//go:noescape

// countTrailingZeros counts the number of zeros
// from the least-significant bit up to the
// first set (1) bit.
//
// references:
// a. https://en.wikipedia.org/wiki/Find_first_set
// b. TZCNTQ on amd64, page 363 of http://support.amd.com/TechDocs/24594.pdf
//
func countTrailingZeros(x uint64) int
