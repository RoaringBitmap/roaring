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

func (self *bitmapContainerShortIterator) next() uint16 {
	j := self.i
	self.i = self.ptr.NextSetBit(self.i + 1)
	return uint16(j)
}
func (self *bitmapContainerShortIterator) hasNext() bool {
	return self.i >= 0
}
func newBitmapContainerShortIterator(a *bitmapContainer) *bitmapContainerShortIterator {
	return &bitmapContainerShortIterator{a, a.NextSetBit(0)}
}
func (self *bitmapContainer) getShortIterator() shortIterable {
	return newBitmapContainerShortIterator(self)
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

func (self *bitmapContainer) fillLeastSignificant16bits(x []int, i, mask int) {
	pos := i
	for k := 0; k < len(self.bitmap); k++ {
		bitset := self.bitmap[k]
		for bitset != 0 {
			t := bitset & -bitset
			x[pos] = (k*64 + bitCount(t-1)) | mask
			pos++
			bitset ^= t
		}
	}
}

func (self *bitmapContainer) equals(o interface{}) bool {
	srb := o.(*bitmapContainer)
	if srb != nil {
		if srb.cardinality != self.cardinality {
			return false
		}
		return bitmapEquals(self.bitmap, srb.bitmap)
	}
	return false
}

func (self *bitmapContainer) add(i uint16) container {
	x := int(i)
	previous := self.bitmap[x/64]
	self.bitmap[x/64] |= (1 << (uint(x) % 64))
	self.cardinality += int(uint(previous^self.bitmap[x/64]) >> (uint(x) % 64))
	return self
}

func (self *bitmapContainer) getCardinality() int {
	return self.cardinality
}

func (self *bitmapContainer) clone() container {
	ptr := bitmapContainer{self.cardinality, make([]uint64, len(self.bitmap))}
	copy(ptr.bitmap, self.bitmap[:])
	return &ptr
}

func (self *bitmapContainer) inot(firstOfRange, lastOfRange int) container {
	return self.NotBitmap(self, firstOfRange, lastOfRange)
}

func (self *bitmapContainer) not(firstOfRange, lastOfRange int) container {
	return self.NotBitmap(newBitmapContainer(), firstOfRange, lastOfRange)

}
func (self *bitmapContainer) NotBitmap(answer *bitmapContainer, firstOfRange, lastOfRange int) container {
	if (lastOfRange - firstOfRange + 1) == maxCapacity {
		newCardinality := maxCapacity - self.cardinality
		for k := 0; k < len(self.bitmap); k++ {
			answer.bitmap[k] = ^self.bitmap[k]
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
	if answer != self {
		//                src            dest
		//System.arraycopy(self.bitmap, 0, answer.bitmap, 0, rangeFirstWord);
		copy(answer.bitmap, self.bitmap[:rangeFirstWord])
		//System.arraycopy(self.bitmap, rangeLastWord + 1, answer.bitmap, rangeLastWord + 1, len(self.bitmap) - (rangeLastWord + 1))
		base := rangeLastWord + 1
		sz := len(self.bitmap) - base
		if sz > 0 {
			copy(answer.bitmap[base:], self.bitmap[base:base+sz])
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
		cardinalityChange = -bitCount(self.bitmap[rangeFirstWord])
		answer.bitmap[rangeFirstWord] = self.bitmap[rangeFirstWord] ^ mask
		cardinalityChange += bitCount(answer.bitmap[rangeFirstWord])
		answer.cardinality = self.cardinality + cardinalityChange

		if answer.cardinality <= arrayDefaultMaxSize {
			return answer.toArrayContainer()
		}
		return answer
	}

	// range spans words
	cardinalityChange += -bitCount(self.bitmap[rangeFirstWord])
	answer.bitmap[rangeFirstWord] = self.bitmap[rangeFirstWord] ^ mask
	cardinalityChange += bitCount(answer.bitmap[rangeFirstWord])

	cardinalityChange += -bitCount(self.bitmap[rangeLastWord])
	answer.bitmap[rangeLastWord] = self.bitmap[rangeLastWord] ^ maskOnLeft
	cardinalityChange += bitCount(answer.bitmap[rangeLastWord])

	// negate the words, if any, strictly between first and last
	for i := rangeFirstWord + 1; i < rangeLastWord; i++ {
		cardinalityChange += (64 - 2*bitCount(self.bitmap[i]))
		answer.bitmap[i] = ^self.bitmap[i]
	}
	answer.cardinality = self.cardinality + cardinalityChange

	if answer.cardinality <= arrayDefaultMaxSize {
		return answer.toArrayContainer()
	}
	return answer
}

func (self *bitmapContainer) or(a container) container {
	switch a.(type) {
	case *arrayContainer:
		return self.orArray(a.(*arrayContainer))
	case *bitmapContainer:
		return self.orBitmap(a.(*bitmapContainer))
	}
	return nil
}

func (self *bitmapContainer) orArray(value2 *arrayContainer) container {
	answer := self.clone().(*bitmapContainer)
	for k := 0; k < value2.getCardinality(); k++ {
		i := uint(toIntUnsigned(value2.content[k])) >> 6
		answer.cardinality += int(uint(^answer.bitmap[i]&(1<<(value2.content[k]%64))) >> (value2.content[k] % 64))
		answer.bitmap[i] = answer.bitmap[i] | (uint64(1) << (value2.content[k] % 64))
	}
	return answer
}
func (self *bitmapContainer) orBitmap(value2 *bitmapContainer) container {
	answer := newBitmapContainer()
	for k := 0; k < len(answer.bitmap); k++ {
		answer.bitmap[k] = self.bitmap[k] | value2.bitmap[k]
		answer.cardinality += bitCount(answer.bitmap[k])
	}
	return answer
}

func (self *bitmapContainer) xor(a container) container {
	switch a.(type) {
	case *arrayContainer:
		return self.xorArray(a.(*arrayContainer))
	case *bitmapContainer:
		return self.xorBitmap(a.(*bitmapContainer))
	}
	return nil
}

func (self *bitmapContainer) xorArray(value2 *arrayContainer) container {
	answer := self.clone().(*bitmapContainer)
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

func (self *bitmapContainer) xorBitmap(value2 *bitmapContainer) container {
	/*
		for k := 0; k < len(self.bitmap); k++ {
			newCardinality += BitCount(self.bitmap[k] ^ value2.bitmap[k])
		}
	*/

	newCardinality := int(popcntXorSlice(self.bitmap, value2.bitmap))

	if newCardinality > arrayDefaultMaxSize {
		answer := newBitmapContainer()
		for k := 0; k < len(answer.bitmap); k++ {
			answer.bitmap[k] = self.bitmap[k] ^ value2.bitmap[k]
		}
		answer.cardinality = newCardinality
		return answer
	}
	ac := newArrayContainerSize(newCardinality)
	fillArrayXOR(ac.content, self.bitmap, value2.bitmap)
	ac.content = ac.content[:newCardinality]
	return ac
}

func (self *bitmapContainer) and(a container) container {
	switch a.(type) {
	case *arrayContainer:
		return self.andArray(a.(*arrayContainer))
	case *bitmapContainer:
		return self.andBitmap(a.(*bitmapContainer))
	}
	return nil
}
func (self *bitmapContainer) andArray(value2 *arrayContainer) *arrayContainer {
	answer := newArrayContainerCapacity(len(value2.content))
	for k := 0; k < value2.getCardinality(); k++ {
		if self.contains(value2.content[k]) {
			answer.content = append(answer.content, value2.content[k])
		}
	}
	return answer

}

func (self *bitmapContainer) andBitmap(value2 *bitmapContainer) container {
	newcardinality := int(popcntAndSlice(self.bitmap, value2.bitmap))
	if newcardinality > arrayDefaultMaxSize {
		answer := newBitmapContainer()
		for k := 0; k < len(answer.bitmap); k++ {
			answer.bitmap[k] = self.bitmap[k] & value2.bitmap[k]
		}
		answer.cardinality = newcardinality
		return answer
	}
	ac := newArrayContainerSize(newcardinality)
	fillArrayAND(ac.content, self.bitmap, value2.bitmap)
	ac.content = ac.content[:newcardinality] //not sure why i need this
	return ac

}

func (self *bitmapContainer) andNot(a container) container {
	switch a.(type) {
	case *arrayContainer:
		return self.andNotArray(a.(*arrayContainer))
	case *bitmapContainer:
		return self.andNotBitmap(a.(*bitmapContainer))
	}
	return nil
}
func (self *bitmapContainer) andNotArray(value2 *arrayContainer) container {
	answer := self.clone().(*bitmapContainer)
	for k := 0; k < value2.getCardinality(); k++ {
		i := uint(toIntUnsigned(value2.content[k])) >> 6
		answer.bitmap[i] = answer.bitmap[i] &^ (uint64(1) << (value2.content[k] % 64))
		answer.cardinality -= int(uint(answer.bitmap[i]^self.bitmap[i]) >> (value2.content[k] % 64))
	}
	if answer.cardinality <= arrayDefaultMaxSize {
		return answer.toArrayContainer()
	}
	return answer
}

func (self *bitmapContainer) andNotBitmap(value2 *bitmapContainer) container {
	/*
		newCardinality := 0
		for k := 0; k < len(self.bitmap); k++ {
			newCardinality += BitCount(self.bitmap[k] &^ value2.bitmap[k])
		}
	*/
	newCardinality := int(popcntMaskSlice(self.bitmap, value2.bitmap))
	if newCardinality > arrayDefaultMaxSize {
		answer := newBitmapContainer()
		for k := 0; k < len(answer.bitmap); k++ {
			answer.bitmap[k] = self.bitmap[k] &^ value2.bitmap[k]
		}
		answer.cardinality = newCardinality
		return answer
	}
	ac := newArrayContainerSize(newCardinality)
	fillArrayANDNOT(ac.content, self.bitmap, value2.bitmap)
	return ac
}

func (self *bitmapContainer) contains(i uint16) bool { //testbit
	x := int(i)
	return (self.bitmap[x/64] & (1 << uint(x%64))) != 0
}
func (self *bitmapContainer) loadData(arrayContainer *arrayContainer) {

	self.cardinality = arrayContainer.getCardinality()
	for k := 0; k < arrayContainer.getCardinality(); k++ {
		x := arrayContainer.content[k]
		i := int(x) / 64
		self.bitmap[i] |= (uint64(1) << uint(x%64))
	}
}

func (self *bitmapContainer) toArrayContainer() *arrayContainer {
	ac := newArrayContainerCapacity(self.cardinality)
	ac.loadData(self)
	return ac
}
func (self *bitmapContainer) fillArray(container []uint16) {
	pos := 0
	for k := 0; k < len(self.bitmap); k++ {
		bitset := self.bitmap[k]
		for bitset != 0 {
			t := bitset & -bitset
			container[pos] = uint16((k*64 + bitCount(t-1)))
			pos = pos + 1
			bitset ^= t
		}
	}
}

func (self *bitmapContainer) NextSetBit(i int) int {
	x := i / 64
	if x >= len(self.bitmap) {
		return -1
	}
	w := self.bitmap[x]
	//w = int64(uint64(w) >> uint(i))
	w = w >> uint(i%64)
	if w != 0 {
		return i + numberOfTrailingZeros(w)
	}
	x++
	for ; x < len(self.bitmap); x++ {
		if self.bitmap[x] != 0 {
			return (x * 64) + numberOfTrailingZeros(self.bitmap[x])
		}
	}
	return -1
}
