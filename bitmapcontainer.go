package roaring

type bitmapContainer struct {
	cardinality int
	bitmap      []uint64
}

func newBitmapContainer() *bitmapContainer {
	p := new(bitmapContainer)
	size := (1 << 16) / 64
	p.bitmap = make([]uint64, size, size)
	return p
}

func newBitmapContainerwithRange(firstOfRun, lastOfRun int) *bitmapContainer {
	this := newBitmapContainer()
	this.cardinality = lastOfRun - firstOfRun + 1
	if this.cardinality == maxCapacity {
		fill(this.bitmap, uint64(0xffffffffffffffff))
	} else {
		firstWord := firstOfRun / 64
		lastWord := lastOfRun / 64
		zeroPrefixLength := uint64(firstOfRun & 63)
		zeroSuffixLength := uint64(63 - (lastOfRun & 63))

		fillRange(this.bitmap, firstWord, lastWord+1, uint64(0xffffffffffffffff))
		this.bitmap[firstWord] ^= ((1 << zeroPrefixLength) - 1)
		blockOfOnes := (uint64(1) << zeroSuffixLength) - 1
		maskOnLeft := blockOfOnes << (uint64(64) - zeroSuffixLength)
		this.bitmap[lastWord] ^= maskOnLeft
	}
	return this
}

type bitmapContainerShortIterator struct {
	ptr *bitmapContainer
	i   int
}

func (bcsi *bitmapContainerShortIterator) next() uint16 {
	j := bcsi.i
	bcsi.i = bcsi.ptr.NextSetBit(bcsi.i + 1)
	return uint16(j)
}
func (bcsi *bitmapContainerShortIterator) hasNext() bool {
	return bcsi.i >= 0
}
func newBitmapContainerShortIterator(a *bitmapContainer) *bitmapContainerShortIterator {
	return &bitmapContainerShortIterator{a, a.NextSetBit(0)}
}
func (bc *bitmapContainer) getShortIterator() shortIterable {
	return newBitmapContainerShortIterator(bc)
}

func (bc *bitmapContainer) getSizeInBytes() int {
	return len(bc.bitmap) * 8
}

func bitmapEquals(a, b []uint64) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func (bc *bitmapContainer) fillLeastSignificant16bits(x []int, i, mask int) {
	pos := i
	for k := 0; k < len(bc.bitmap); k++ {
		bitset := bc.bitmap[k]
		for bitset != 0 {
			t := bitset & -bitset
			x[pos] = (k*64 + bitCount(t-1)) | mask
			pos++
			bitset ^= t
		}
	}
}

func (bc *bitmapContainer) equals(o interface{}) bool {
	srb := o.(*bitmapContainer)
	if srb != nil {
		if srb.cardinality != bc.cardinality {
			return false
		}
		return bitmapEquals(bc.bitmap, srb.bitmap)
	}
	return false
}

func (bc *bitmapContainer) add(i uint16) container {
	x := int(i)
	previous := bc.bitmap[x/64]
	bc.bitmap[x/64] |= (1 << (uint(x) % 64))
	bc.cardinality += int(uint(previous^bc.bitmap[x/64]) >> (uint(x) % 64))
	return bc
}

func (bc *bitmapContainer) getCardinality() int {
	return bc.cardinality
}

func (bc *bitmapContainer) clone() container {
	ptr := bitmapContainer{bc.cardinality, make([]uint64, len(bc.bitmap))}
	copy(ptr.bitmap, bc.bitmap[:])
	return &ptr
}

func (bc *bitmapContainer) inot(firstOfRange, lastOfRange int) container {
	return bc.NotBitmap(bc, firstOfRange, lastOfRange)
}

func (bc *bitmapContainer) not(firstOfRange, lastOfRange int) container {
	return bc.NotBitmap(newBitmapContainer(), firstOfRange, lastOfRange)

}
func (bc *bitmapContainer) NotBitmap(answer *bitmapContainer, firstOfRange, lastOfRange int) container {
	if (lastOfRange - firstOfRange + 1) == maxCapacity {
		newCardinality := maxCapacity - bc.cardinality
		for k := 0; k < len(bc.bitmap); k++ {
			answer.bitmap[k] = ^bc.bitmap[k]
		}
		answer.cardinality = newCardinality
		if newCardinality <= arrayDefaultMaxSize {
			return answer.toArrayContainer()
		}
		return answer
	}
	// could be optimized to first determine the answer cardinality,
	// rather than update/create bitmap and then possibly convert

	cardinalityChange := 0
	rangeFirstWord := firstOfRange / 64
	rangeFirstBitPos := firstOfRange & 63
	rangeLastWord := lastOfRange / 64
	rangeLastBitPos := lastOfRange & 63

	// if not in place, we need to duplicate stuff before
	// rangeFirstWord and after rangeLastWord
	if answer != bc {
		//                src            dest
		//System.arraycopy(self.bitmap, 0, answer.bitmap, 0, rangeFirstWord);
		copy(answer.bitmap, bc.bitmap[:rangeFirstWord])
		//System.arraycopy(self.bitmap, rangeLastWord + 1, answer.bitmap, rangeLastWord + 1, len(self.bitmap) - (rangeLastWord + 1))
		base := rangeLastWord + 1
		sz := len(bc.bitmap) - base
		if sz > 0 {
			copy(answer.bitmap[base:], bc.bitmap[base:base+sz])
		}

		//	copy(answer.bitmap[rangeLastWord+1:], self.bitmap[rangeLastWord+1:len(self.bitmap)-(rangeLastWord+1)])

	}

	// unfortunately, the simple expression gives the wrong mask for
	// rangeLastBitPos==63
	// no branchless way comes to mind
	maskOnLeft := uint64(0xffffffffffffffff)
	if rangeLastBitPos != 63 {
		maskOnLeft = (1 << uint((rangeLastBitPos+1)%64)) - 1
	}
	mask := uint64(0xffffffffffffffff) // now zero out stuff in the prefix

	mask ^= (uint64(1) << uint(rangeFirstBitPos%64)) - 1

	if rangeFirstWord == rangeLastWord {
		// range starts and ends in same word (may have
		// unchanged bits on both left and right)
		mask &= maskOnLeft
		cardinalityChange = -bitCount(bc.bitmap[rangeFirstWord])
		answer.bitmap[rangeFirstWord] = bc.bitmap[rangeFirstWord] ^ mask
		cardinalityChange += bitCount(answer.bitmap[rangeFirstWord])
		answer.cardinality = bc.cardinality + cardinalityChange

		if answer.cardinality <= arrayDefaultMaxSize {
			return answer.toArrayContainer()
		}
		return answer
	}

	// range spans words
	cardinalityChange += -bitCount(bc.bitmap[rangeFirstWord])
	answer.bitmap[rangeFirstWord] = bc.bitmap[rangeFirstWord] ^ mask
	cardinalityChange += bitCount(answer.bitmap[rangeFirstWord])

	cardinalityChange += -bitCount(bc.bitmap[rangeLastWord])
	answer.bitmap[rangeLastWord] = bc.bitmap[rangeLastWord] ^ maskOnLeft
	cardinalityChange += bitCount(answer.bitmap[rangeLastWord])

	// negate the words, if any, strictly between first and last
	for i := rangeFirstWord + 1; i < rangeLastWord; i++ {
		cardinalityChange += (64 - 2*bitCount(bc.bitmap[i]))
		answer.bitmap[i] = ^bc.bitmap[i]
	}
	answer.cardinality = bc.cardinality + cardinalityChange

	if answer.cardinality <= arrayDefaultMaxSize {
		return answer.toArrayContainer()
	}
	return answer
}

func (bc *bitmapContainer) or(a container) container {
	switch a.(type) {
	case *arrayContainer:
		return bc.orArray(a.(*arrayContainer))
	case *bitmapContainer:
		return bc.orBitmap(a.(*bitmapContainer))
	}
	return nil
}

func (bc *bitmapContainer) orArray(value2 *arrayContainer) container {
	answer := bc.clone().(*bitmapContainer)
	for k := 0; k < value2.getCardinality(); k++ {
		i := uint(toIntUnsigned(value2.content[k])) >> 6
		answer.cardinality += int(uint(^answer.bitmap[i]&(1<<(value2.content[k]%64))) >> (value2.content[k] % 64))
		answer.bitmap[i] = answer.bitmap[i] | (uint64(1) << (value2.content[k] % 64))
	}
	return answer
}
func (bc *bitmapContainer) orBitmap(value2 *bitmapContainer) container {
	answer := newBitmapContainer()
	for k := 0; k < len(answer.bitmap); k++ {
		answer.bitmap[k] = bc.bitmap[k] | value2.bitmap[k]
		answer.cardinality += bitCount(answer.bitmap[k])
	}
	return answer
}

func (bc *bitmapContainer) xor(a container) container {
	switch a.(type) {
	case *arrayContainer:
		return bc.xorArray(a.(*arrayContainer))
	case *bitmapContainer:
		return bc.xorBitmap(a.(*bitmapContainer))
	}
	return nil
}

func (bc *bitmapContainer) xorArray(value2 *arrayContainer) container {
	answer := bc.clone().(*bitmapContainer)
	for k := 0; k < value2.getCardinality(); k++ {
		index := uint(toIntUnsigned(value2.content[k])) >> 6
		answer.cardinality += 1 - 2*int(uint(answer.bitmap[index]&(1<<(value2.content[k]%64)))>>(value2.content[k]%64))

		answer.bitmap[index] = answer.bitmap[index] ^ (uint64(1) << (value2.content[k] % 64))
	}
	if answer.cardinality <= arrayDefaultMaxSize {
		return answer.toArrayContainer()
	}
	return answer
}

func (bc *bitmapContainer) xorBitmap(value2 *bitmapContainer) container {
	/*
		for k := 0; k < len(bc.bitmap); k++ {
			newCardinality += BitCount(bc.bitmap[k] ^ value2.bitmap[k])
		}
	*/

	newCardinality := int(popcntXorSlice(bc.bitmap, value2.bitmap))

	if newCardinality > arrayDefaultMaxSize {
		answer := newBitmapContainer()
		for k := 0; k < len(answer.bitmap); k++ {
			answer.bitmap[k] = bc.bitmap[k] ^ value2.bitmap[k]
		}
		answer.cardinality = newCardinality
		return answer
	}
	ac := newArrayContainerSize(newCardinality)
	fillArrayXOR(ac.content, bc.bitmap, value2.bitmap)
	ac.content = ac.content[:newCardinality]
	return ac
}

func (bc *bitmapContainer) and(a container) container {
	switch a.(type) {
	case *arrayContainer:
		return bc.andArray(a.(*arrayContainer))
	case *bitmapContainer:
		return bc.andBitmap(a.(*bitmapContainer))
	}
	return nil
}
func (bc *bitmapContainer) andArray(value2 *arrayContainer) *arrayContainer {
	answer := newArrayContainerCapacity(len(value2.content))
	for k := 0; k < value2.getCardinality(); k++ {
		if bc.contains(value2.content[k]) {
			answer.content = append(answer.content, value2.content[k])
		}
	}
	return answer

}

func (bc *bitmapContainer) andBitmap(value2 *bitmapContainer) container {
	newcardinality := int(popcntAndSlice(bc.bitmap, value2.bitmap))
	if newcardinality > arrayDefaultMaxSize {
		answer := newBitmapContainer()
		for k := 0; k < len(answer.bitmap); k++ {
			answer.bitmap[k] = bc.bitmap[k] & value2.bitmap[k]
		}
		answer.cardinality = newcardinality
		return answer
	}
	ac := newArrayContainerSize(newcardinality)
	fillArrayAND(ac.content, bc.bitmap, value2.bitmap)
	ac.content = ac.content[:newcardinality] //not sure why i need this
	return ac

}

func (bc *bitmapContainer) andNot(a container) container {
	switch a.(type) {
	case *arrayContainer:
		return bc.andNotArray(a.(*arrayContainer))
	case *bitmapContainer:
		return bc.andNotBitmap(a.(*bitmapContainer))
	}
	return nil
}
func (bc *bitmapContainer) andNotArray(value2 *arrayContainer) container {
	answer := bc.clone().(*bitmapContainer)
	for k := 0; k < value2.getCardinality(); k++ {
		i := uint(toIntUnsigned(value2.content[k])) >> 6
		answer.bitmap[i] = answer.bitmap[i] &^ (uint64(1) << (value2.content[k] % 64))
		answer.cardinality -= int(uint(answer.bitmap[i]^bc.bitmap[i]) >> (value2.content[k] % 64))
	}
	if answer.cardinality <= arrayDefaultMaxSize {
		return answer.toArrayContainer()
	}
	return answer
}

func (bc *bitmapContainer) andNotBitmap(value2 *bitmapContainer) container {
	/*
		newCardinality := 0
		for k := 0; k < len(bc.bitmap); k++ {
			newCardinality += BitCount(self.bitmap[k] &^ value2.bitmap[k])
		}
	*/
	newCardinality := int(popcntMaskSlice(bc.bitmap, value2.bitmap))
	if newCardinality > arrayDefaultMaxSize {
		answer := newBitmapContainer()
		for k := 0; k < len(answer.bitmap); k++ {
			answer.bitmap[k] = bc.bitmap[k] &^ value2.bitmap[k]
		}
		answer.cardinality = newCardinality
		return answer
	}
	ac := newArrayContainerSize(newCardinality)
	fillArrayANDNOT(ac.content, bc.bitmap, value2.bitmap)
	return ac
}

func (bc *bitmapContainer) contains(i uint16) bool { //testbit
	x := int(i)
	return (bc.bitmap[x/64] & (1 << uint(x%64))) != 0
}
func (bc *bitmapContainer) loadData(arrayContainer *arrayContainer) {

	bc.cardinality = arrayContainer.getCardinality()
	for k := 0; k < arrayContainer.getCardinality(); k++ {
		x := arrayContainer.content[k]
		i := int(x) / 64
		bc.bitmap[i] |= (uint64(1) << uint(x%64))
	}
}

func (bc *bitmapContainer) toArrayContainer() *arrayContainer {
	ac := newArrayContainerCapacity(bc.cardinality)
	ac.loadData(bc)
	return ac
}
func (bc *bitmapContainer) fillArray(container []uint16) {
	pos := 0
	for k := 0; k < len(bc.bitmap); k++ {
		bitset := bc.bitmap[k]
		for bitset != 0 {
			t := bitset & -bitset
			container[pos] = uint16((k*64 + bitCount(t-1)))
			pos = pos + 1
			bitset ^= t
		}
	}
}

func (bc *bitmapContainer) NextSetBit(i int) int {
	x := i / 64
	if x >= len(bc.bitmap) {
		return -1
	}
	w := bc.bitmap[x]
	//w = int64(uint64(w) >> uint(i))
	w = w >> uint(i%64)
	if w != 0 {
		return i + numberOfTrailingZeros(w)
	}
	x++
	for ; x < len(bc.bitmap); x++ {
		if bc.bitmap[x] != 0 {
			return (x * 64) + numberOfTrailingZeros(bc.bitmap[x])
		}
	}
	return -1
}
