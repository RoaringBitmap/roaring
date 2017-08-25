// +build go1.9

package roaring

import "math/bits"

func popcount(x uint64) uint64 {
	return uint64(bits.OnesCount64(x))
}
