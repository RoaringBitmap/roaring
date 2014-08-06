package roaring

type short uint16

// should be replaced with optimized assembly instructions
func BitCount(x uint64) int {
	// bit population count, see
	// http://graphics.stanford.edu/~seander/bithacks.html#CountBitsSetParallel
	x -= (x >> 1) & 0x5555555555555555
	x = (x>>2)&0x3333333333333333 + x&0x3333333333333333
	x += x >> 4
	x &= 0x0f0f0f0f0f0f0f0f
	x *= 0x0101010101010101
	return int(x >> 56)
}

// should be replaced with optimized assembly instructions
func NumberOfTrailingZeros(i uint64) int {
	if i == 0 {
		return 64
	}
	x := i
	n := int64(63)
	y := x << 32
	if y != 0 {
		n -= 32
		x = y
	}
	y = x << 16
	if y != 0 {
		n -= 16
		x = y
	}
	y = x << 8
	if y != 0 {
		n -= 8
		x = y
	}
	y = x << 4
	if y != 0 {
		n -= 4
		x = y
	}
	y = x << 2
	if y != 0 {
		n -= 2
		x = y
	}
	return int(n - int64(uint64(x<<1)>>63))
}

func fill(arr []uint64, val uint64) {
	for i := range arr {
		arr[i] = val
	}
}
func fillRange(arr []uint64, start, end int, val uint64) {
	for i := start; i < end; i++ {
		arr[i] = val
	}
}

func FillArrayAND(container []short, bitmap1, bitmap2 []uint64) {
	pos := 0
	if len(bitmap1) != len(bitmap2) {
		panic("array lengths don't match")
	}
	for k, _ := range bitmap1 {
		bitset := bitmap1[k] & bitmap2[k]
		for bitset != 0 {
			t := bitset & -bitset
			container[pos] = short((k*64 + BitCount(t-1)))
			pos++
			bitset ^= t
		}
	}
}

func FillArrayANDNOT(container []short, bitmap1, bitmap2 []uint64) {
	pos := 0
	if len(bitmap1) != len(bitmap2) {
		panic("array lengths don't match")
	}
	for k, _ := range bitmap1 {
		bitset := bitmap1[k] &^ bitmap2[k]
		for bitset != 0 {
			t := bitset & -bitset
			container[pos] = short((k*64 + BitCount(t-1)))
			pos++
			bitset ^= t
		}
	}
}

func FillArrayXOR(container []short, bitmap1, bitmap2 []uint64) {
	pos := 0
	if len(bitmap1) != len(bitmap2) {
		panic("array lengths don't match")
	}
	for k := 0; k < len(bitmap1); k++ {
		bitset := bitmap1[k] ^ bitmap2[k]
		for bitset != 0 {
			t := bitset & -bitset
			container[pos] = short((k*64 + BitCount(t-1)))
			pos++
			bitset ^= t
		}
	}
}

func Highbits(x int) short {
	u := uint(x)
	return short(u >> 16)
}
func Lowbits(x int) short {
	return short(x & 0xFFFF)
}

func MaxLowBit() short {
	return short(0xFFFF)
}

func ToIntUnsigned(x short) int {
	return int(x & 0xFFFF)
}
