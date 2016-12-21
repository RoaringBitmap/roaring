// +build amd64,!appengine

package roaring

const deBruijn32 = 0x077CB531

var deBruijn32Lookup = []byte{
	0, 1, 28, 2, 29, 14, 24, 3, 30, 22, 20, 15, 25, 17, 4, 8,
	31, 27, 13, 23, 21, 19, 16, 7, 26, 12, 18, 6, 11, 5, 10, 9,
}

const deBruijn64 = 0x03f79d71b4ca8b09

var deBruijn64Lookup = []byte{
	0, 1, 56, 2, 57, 49, 28, 3, 61, 58, 42, 50, 38, 29, 17, 4,
	62, 47, 59, 36, 45, 43, 51, 22, 53, 39, 33, 30, 24, 18, 12, 5,
	63, 55, 48, 27, 60, 41, 37, 16, 46, 35, 44, 21, 52, 32, 23, 11,
	54, 26, 40, 15, 34, 20, 31, 10, 25, 14, 19, 9, 13, 8, 7, 6,
}

// countTrailingZeros counts the number of zeros
// from the least-significant bit up to the
// first set (1) bit. if x is 0, 64 is returned.
//
// references:
// a. https://en.wikipedia.org/wiki/Find_first_set
// b. TZCNTQ on amd64, page 364 of http://support.amd.com/TechDocs/24594.pdf
//
// *** the following function is defined in ctz_amd64.s
//
// TODO: possibly use "github.com/klauspost/cpuid"
// to check if cpuid.CPU.BMI1() is true before using the assembly version.
//
// The Go version is in ctz_generic.go.
//
func countTrailingZerosAsm(x uint64) int
