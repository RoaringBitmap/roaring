package roaring

import "math/bits"

func countTrailingZeros(x uint64) int {
	return bits.TrailingZeros64(x)
}
