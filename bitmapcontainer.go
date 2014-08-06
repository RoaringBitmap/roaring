package roaring

type BitmapContainer struct {
	cardinality int
	bitmap      []uint64
}

func NewBitmapContainer() *BitmapContainer {
	p := new(BitmapContainer)
	size := (1 << 16) / 64
	p.bitmap = make([]uint64, size, size)
	return p
}

func NewBitmapContainerwithRange(firstOfRun, lastOfRun int) *BitmapContainer {
	this := NewBitmapContainer()
	this.cardinality = lastOfRun - firstOfRun + 1
	if this.cardinality == MAX_CAPACITY {
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

type BitmapContainerShortIterator struct {
	ptr *BitmapContainer
	i   int
}

func (self *BitmapContainerShortIterator) Next() short {
	j := self.i
	self.i = self.ptr.NextSetBit(self.i + 1)
	return short(j)
}
func (self *BitmapContainerShortIterator) HasNext() bool {
	return self.i >= 0
}
func NewBitmapContainerShortIterator(a *BitmapContainer) *BitmapContainerShortIterator {
	return &BitmapContainerShortIterator{a, a.NextSetBit(0)}
}
func (self *BitmapContainer) GetShortIterator() ShortIterable {
	return NewBitmapContainerShortIterator(self)
}

func BitmapEquals(a, b []uint64) bool {
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

func (self *BitmapContainer) FillLeastSignificant16bits(x []int, i, mask int) {
	pos := i
	for k := 0; k < len(self.bitmap); k++ {
		bitset := self.bitmap[k]
		for bitset != 0 {
			t := bitset & -bitset
			x[pos] = (k*64 + BitCount(t-1)) | mask
			pos++
			bitset ^= t
		}
	}
}

func (self *BitmapContainer) Equals(o interface{}) bool {
	srb := o.(*BitmapContainer)
	if srb != nil {
		if srb.cardinality != self.cardinality {
			return false
		}
		return BitmapEquals(self.bitmap, srb.bitmap)
	}
	return false
}

func (self *BitmapContainer) Add(i short) Container {
	x := int(i)
	previous := self.bitmap[x/64]
	self.bitmap[x/64] |= (1 << (uint(x) % 64))
	self.cardinality += int(uint(previous^self.bitmap[x/64]) >> (uint(x) % 64))
	return self
}

func (self *BitmapContainer) GetCardinality() int {
	return self.cardinality
}

func (self *BitmapContainer) Clone() Container {
	ptr := BitmapContainer{self.cardinality, make([]uint64, len(self.bitmap))}
	copy(ptr.bitmap, self.bitmap[:])
	return &ptr
}

func (self *BitmapContainer) Inot(firstOfRange, lastOfRange int) Container {
	return self.NotBitmap(self, firstOfRange, lastOfRange)
}

func (self *BitmapContainer) Not(firstOfRange, lastOfRange int) Container {
	return self.NotBitmap(NewBitmapContainer(), firstOfRange, lastOfRange)

}
func (self *BitmapContainer) NotBitmap(answer *BitmapContainer, firstOfRange, lastOfRange int) Container {
	if (lastOfRange - firstOfRange + 1) == MAX_CAPACITY {
		newCardinality := MAX_CAPACITY - self.cardinality
		for k := 0; k < len(self.bitmap); k++ {
			answer.bitmap[k] = ^self.bitmap[k]
		}
		answer.cardinality = newCardinality
		if newCardinality <= ARRAY_DEFAULT_MAX_SIZE {
			return answer.ToArrayContainer()
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
		cardinalityChange = -BitCount(self.bitmap[rangeFirstWord])
		answer.bitmap[rangeFirstWord] = self.bitmap[rangeFirstWord] ^ mask
		cardinalityChange += BitCount(answer.bitmap[rangeFirstWord])
		answer.cardinality = self.cardinality + cardinalityChange

		if answer.cardinality <= ARRAY_DEFAULT_MAX_SIZE {
			return answer.ToArrayContainer()
		}
		return answer
	}

	// range spans words
	cardinalityChange += -BitCount(self.bitmap[rangeFirstWord])
	answer.bitmap[rangeFirstWord] = self.bitmap[rangeFirstWord] ^ mask
	cardinalityChange += BitCount(answer.bitmap[rangeFirstWord])

	cardinalityChange += -BitCount(self.bitmap[rangeLastWord])
	answer.bitmap[rangeLastWord] = self.bitmap[rangeLastWord] ^ maskOnLeft
	cardinalityChange += BitCount(answer.bitmap[rangeLastWord])

	// negate the words, if any, strictly between first and last
	for i := rangeFirstWord + 1; i < rangeLastWord; i++ {
		cardinalityChange += (64 - 2*BitCount(self.bitmap[i]))
		answer.bitmap[i] = ^self.bitmap[i]
	}
	answer.cardinality = self.cardinality + cardinalityChange

	if answer.cardinality <= ARRAY_DEFAULT_MAX_SIZE {
		return answer.ToArrayContainer()
	}
	return answer
}

func (self *BitmapContainer) Or(a Container) Container {
	switch a.(type) {
	case *ArrayContainer:
		return self.OrArray(a.(*ArrayContainer))
	case *BitmapContainer:
		return self.OrBitmap(a.(*BitmapContainer))
	}
	return nil
}

func (self *BitmapContainer) OrArray(value2 *ArrayContainer) Container {
	answer := self.Clone().(*BitmapContainer)
	for k := 0; k < value2.GetCardinality(); k++ {
		i := uint(ToIntUnsigned(value2.content[k])) >> 6
		//java	answer.cardinality += ((~answer.bitmap[i]) & (1 << (value2.content[k] %64))) >>> (value2.content[k]%64);
		answer.cardinality += int(uint(^answer.bitmap[i]&(1<<(value2.content[k]%64))) >> (value2.content[k] % 64))
		answer.bitmap[i] = answer.bitmap[i] | (1 << (value2.content[k] % 64))
	}
	return answer
}
func (self *BitmapContainer) OrBitmap(value2 *BitmapContainer) Container {
	answer := NewBitmapContainer()
	for k := 0; k < len(answer.bitmap); k++ {
		answer.bitmap[k] = self.bitmap[k] | value2.bitmap[k]
		answer.cardinality += BitCount(answer.bitmap[k])
	}
	return answer
}

func (self *BitmapContainer) Xor(a Container) Container {
	switch a.(type) {
	case *ArrayContainer:
		return self.XorArray(a.(*ArrayContainer))
	case *BitmapContainer:
		return self.XorBitmap(a.(*BitmapContainer))
	}
	return nil
}

func (self *BitmapContainer) XorArray(value2 *ArrayContainer) Container {
	answer := self.Clone().(*BitmapContainer)
	for k := 0; k < value2.GetCardinality(); k++ {
		index := uint(ToIntUnsigned(value2.content[k])) >> 6
		answer.cardinality += 1 - 2*int(uint(answer.bitmap[index]&(1<<(value2.content[k]%64)))>>(value2.content[k]%64))

		answer.bitmap[index] = answer.bitmap[index] ^ (1 << (value2.content[k] % 64))
	}
	if answer.cardinality <= ARRAY_DEFAULT_MAX_SIZE {
		return answer.ToArrayContainer()
	}
	return answer
}

func (self *BitmapContainer) XorBitmap(value2 *BitmapContainer) Container {
	/*
		for k := 0; k < len(self.bitmap); k++ {
			newCardinality += BitCount(self.bitmap[k] ^ value2.bitmap[k])
		}
	*/

	newCardinality := int(popcntXorSlice(self.bitmap, value2.bitmap))

	if newCardinality > ARRAY_DEFAULT_MAX_SIZE {
		answer := NewBitmapContainer()
		for k := 0; k < len(answer.bitmap); k++ {
			answer.bitmap[k] = self.bitmap[k] ^ value2.bitmap[k]
		}
		answer.cardinality = newCardinality
		return answer
	}
	ac := NewArrayContainerCapacity(newCardinality)
	ac.content = make([]short, newCardinality)
	FillArrayXOR(ac.content, self.bitmap, value2.bitmap)
	ac.content = ac.content[:newCardinality]
	return ac
}

func (self *BitmapContainer) And(a Container) Container {
	switch a.(type) {
	case *ArrayContainer:
		return self.AndArray(a.(*ArrayContainer))
	case *BitmapContainer:
		return self.AndBitmap(a.(*BitmapContainer))
	}
	return nil
}
func (self *BitmapContainer) AndArray(value2 *ArrayContainer) *ArrayContainer {
	answer := NewArrayContainerCapacity(len(value2.content))
	for k := 0; k < value2.GetCardinality(); k++ {
		if self.Contains(value2.content[k]) {
			answer.content = append(answer.content, value2.content[k])
		}
	}
	return answer

}

func (self *BitmapContainer) AndBitmap(value2 *BitmapContainer) Container {
	newcardinality := int(popcntAndSlice(self.bitmap, value2.bitmap))
	if newcardinality > ARRAY_DEFAULT_MAX_SIZE {
		answer := NewBitmapContainer()
		for k := 0; k < len(answer.bitmap); k++ {
			answer.bitmap[k] = self.bitmap[k] & value2.bitmap[k]
		}
		answer.cardinality = newcardinality
		return answer
	}
	ac := NewArrayContainerCapacity(newcardinality)
	ac.content = make([]short, newcardinality)
	FillArrayAND(ac.content, self.bitmap, value2.bitmap)
	ac.content = ac.content[:newcardinality] //not sure why i need this
	return ac

}

func (self *BitmapContainer) AndNot(a Container) Container {
	switch a.(type) {
	case *ArrayContainer:
		return self.AndNotArray(a.(*ArrayContainer))
	case *BitmapContainer:
		return self.AndNotBitmap(a.(*BitmapContainer))
	}
	return nil
}
func (self *BitmapContainer) AndNotArray(value2 *ArrayContainer) Container {
	answer := self.Clone().(*BitmapContainer)
	for k := 0; k < value2.GetCardinality(); k++ {
		i := uint(ToIntUnsigned(value2.content[k])) >> 6
		answer.bitmap[i] = answer.bitmap[i] &^ (1 << (value2.content[k] % 64))
		answer.cardinality -= int(uint(answer.bitmap[i]^self.bitmap[i]) >> (value2.content[k] % 64))
	}
	if answer.cardinality <= ARRAY_DEFAULT_MAX_SIZE {
		return answer.ToArrayContainer()
	}
	return answer
}

func (self *BitmapContainer) AndNotBitmap(value2 *BitmapContainer) Container {
	/*
		newCardinality := 0
		for k := 0; k < len(self.bitmap); k++ {
			newCardinality += BitCount(self.bitmap[k] &^ value2.bitmap[k])
		}
	*/
	newCardinality := int(popcntMaskSlice(self.bitmap, value2.bitmap))
	if newCardinality > ARRAY_DEFAULT_MAX_SIZE {
		answer := NewBitmapContainer()
		for k := 0; k < len(answer.bitmap); k++ {
			answer.bitmap[k] = self.bitmap[k] &^ value2.bitmap[k]
		}
		answer.cardinality = newCardinality
		return answer
	}
	ac := NewArrayContainerCapacity(newCardinality)
	FillArrayANDNOT(ac.content, self.bitmap, value2.bitmap)
	ac.content = ac.content[:len(self.bitmap)]

	return ac
}

func (self *BitmapContainer) Contains(i short) bool { //testbit
	x := int(i)
	return (self.bitmap[x/64] & (1 << uint(x%64))) != 0
}
func (self *BitmapContainer) loadData(arrayContainer *ArrayContainer) {

	self.cardinality = arrayContainer.GetCardinality()
	for k := 0; k < arrayContainer.GetCardinality(); k++ {
		x := arrayContainer.content[k]
		i := int(x) / 64
		self.bitmap[i] |= (uint64(1) << uint(x%64))
	}
}

func (self *BitmapContainer) ToArrayContainer() *ArrayContainer {
	ac := NewArrayContainerCapacity(self.cardinality)
	ac.loadData(self)
	return ac
}
func (self *BitmapContainer) fillArray(array []short) {
	pos := 0
	for k := 0; k < len(self.bitmap); k++ {
		bitset := self.bitmap[k]
		for bitset != 0 {
			t := bitset & -bitset
			array[pos] = short((k * 64) + BitCount(t-1))
			pos++
			bitset ^= t
		}
	}
}

func (self *BitmapContainer) NextSetBit(i int) int {
	x := i / 64
	if x >= len(self.bitmap) {
		return -1
	}
	w := self.bitmap[x]
	//w = int64(uint64(w) >> uint(i))
	w = w >> uint(i%64)
	if w != 0 {
		return i + NumberOfTrailingZeros(w)
	}
	x++
	for ; x < len(self.bitmap); x++ {
		if self.bitmap[x] != 0 {
			return (x * 64) + NumberOfTrailingZeros(self.bitmap[x])
		}
	}
	return -1
}
