package roaring

// decodeWithTrailingZeros selects the bitmapContainerManyIterator decode loop
// built on bits.TrailingZeros64 and whole-word draining. arm64 has no scalar
// popcount, so bits.OnesCount64 goes through the vector unit for every set
// bit. The other targets keep the original loop: on amd64 the result depends
// on the microarchitecture and on GOAMD64, and did not consistently improve.
const decodeWithTrailingZeros = true
