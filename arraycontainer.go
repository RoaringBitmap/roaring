package roaring

type arrayContainer struct {
	content []uint16
}

func (self *arrayContainer) fillLeastSignificant16bits(x []int, i, mask int) {
	for k := 0; k < len(self.content); k++ {
		x[k+i] = toIntUnsigned(self.content[k]) | mask
	}
}

func (self *arrayContainer) getShortIterator() shortIterable {
	return &shortIterator{self.content, 0}
}

func (self *arrayContainer) not(firstOfRange, lastOfRange int) container {
	if firstOfRange > lastOfRange {
		return self.clone()
	}

	// determine the span of array indices to be affected^M
	startIndex := binarySearch(self.content, uint16(firstOfRange))
	if startIndex < 0 {
		startIndex = -startIndex - 1
	}
	lastIndex := binarySearch(self.content, uint16(lastOfRange))
	if lastIndex < 0 {
		lastIndex = -lastIndex - 2
	}
	currentValuesInRange := lastIndex - startIndex + 1
	spanToBeFlipped := lastOfRange - firstOfRange + 1
	newValuesInRange := spanToBeFlipped - currentValuesInRange
	cardinalityChange := newValuesInRange - currentValuesInRange
	newCardinality := len(self.content) + cardinalityChange

	if newCardinality >= array_default_max_size {
		return self.toBitmapContainer().not(firstOfRange, lastOfRange)
	}
	answer := newArrayContainer()
	answer.content = make([]uint16, newCardinality, newCardinality) //a hack for sure

	copy(answer.content, self.content[:startIndex])
	outPos := startIndex
	inPos := startIndex
	valInRange := firstOfRange
	for ; valInRange <= lastOfRange && inPos <= lastIndex; valInRange++ {
		if uint16(valInRange) != self.content[inPos] {
			answer.content[outPos] = uint16(valInRange)
			outPos++
		} else {
			inPos++
		}
	}

	for ; valInRange <= lastOfRange; valInRange++ {
		answer.content[outPos] = uint16(valInRange)
		outPos++
	}

	for i := lastIndex + 1; i < len(self.content); i++ {
		answer.content[outPos] = self.content[i]
		outPos++
	}
	answer.content = answer.content[:newCardinality]
	return answer

}

func (self *arrayContainer) equals(o interface{}) bool {
	srb := o.(*arrayContainer)
	if srb != nil {
		if len(srb.content) != len(self.content) {
			return false
		}
		for i := 0; i < len(self.content); i++ {
			if self.content[i] != srb.content[i] {
				return false
			}
		}
		return true
	}
	return false
}

func (self *arrayContainer) toBitmapContainer() *bitmapContainer {
	bc := newBitmapContainer()
	bc.loadData(self)
	return bc

}
func (self *arrayContainer) add(x uint16) container {
	if len(self.content) >= array_default_max_size {
		a := self.toBitmapContainer()
		a.add(x)
		return a
	}
	if (len(self.content) == 0) || (x > self.content[len(self.content)-1]) {
		self.content = append(self.content, x)
		return self
	}
	loc := binarySearch(self.content, x)
	if loc < 0 {
		s := self.content
		i := -loc - 1
		s = append(s, 0)
		copy(s[i+1:], s[i:])
		s[i] = x
		self.content = s
	}
	return self
}

func (self *arrayContainer) or(a container) container {
	switch a.(type) {
	case *arrayContainer:
		return self.orArray(a.(*arrayContainer))
	case *bitmapContainer:
		return a.or(self)
	}
	return nil
}

func (self *arrayContainer) orArray(value2 *arrayContainer) container {
	value1 := self
	totalCardinality := value1.getCardinality() + value2.getCardinality()
	if totalCardinality > array_default_max_size { // it could be a bitmap!^M
		bc := newBitmapContainer()
		for k := 0; k < len(value2.content); k++ {
			i := uint(toIntUnsigned(value2.content[k])) >> 6
			bc.bitmap[i] |= (1 << (value2.content[k] % 64))
		}
		for k := 0; k < len(self.content); k++ {
			i := uint(toIntUnsigned(self.content[k])) >> 6
			bc.bitmap[i] |= (1 << (self.content[k] % 64))
		}
		bc.cardinality = 0
		for _, k := range bc.bitmap {
			bc.cardinality += bitCount(k)
		}
		if bc.cardinality <= array_default_max_size {
			return bc.toArrayContainer()
		}
		return bc
	}
	desiredCapacity := totalCardinality
	answer := newArrayContainerCapacity(desiredCapacity)
	nl := union2by2(value1.content, value2.content, answer.content)
	answer.content = answer.content[:nl] //what is this voodo?
	return answer
}

func (self *arrayContainer) and(a container) container {
	switch a.(type) {
	case *arrayContainer:
		return self.andArray(a.(*arrayContainer))
	case *bitmapContainer:
		return a.and(self)
	}
	return nil
}

func (self *arrayContainer) xor(a container) container {
	switch a.(type) {
	case *arrayContainer:
		return self.xorArray(a.(*arrayContainer))
	case *bitmapContainer:
		return a.xor(self)
	}
	return nil
}

func (self *arrayContainer) xorArray(value2 *arrayContainer) container {
	value1 := self
	totalCardinality := value1.getCardinality() + value2.getCardinality()
	if totalCardinality > array_default_max_size { // it could be a bitmap!^M
		bc := newBitmapContainer()
		for k := 0; k < len(value2.content); k++ {
			i := uint(toIntUnsigned(value2.content[k])) >> 6
			bc.bitmap[i] ^= (1 << value2.content[k])
		}
		for k := 0; k < len(self.content); k++ {
			i := uint(toIntUnsigned(self.content[k])) >> 6
			bc.bitmap[i] ^= (1 << self.content[k])
		}
		bc.cardinality = 0
		for _, k := range bc.bitmap {
			bc.cardinality += bitCount(k)
		}
		if bc.cardinality <= array_default_max_size {
			return bc.toArrayContainer()
		}
		return bc
	}
	desiredCapacity := totalCardinality
	answer := newArrayContainerCapacity(desiredCapacity)
	length := exclusiveUnion2by2(value1.content, value2.content, answer.content)
	answer.content = answer.content[:length]
	return answer

}

func (self *arrayContainer) andNot(a container) container {
	switch a.(type) {
	case *arrayContainer:
		return self.andNotArray(a.(*arrayContainer))
	case *bitmapContainer:
		return a.andNot(self)
	}
	return nil
}

func (self *arrayContainer) andNotArray(value2 *arrayContainer) container {
	value1 := self
	desiredcapacity := value1.getCardinality()
	answer := newArrayContainerCapacity(desiredcapacity)
	length := difference(value1.content, value2.content, answer.content)
	answer.content = answer.content[:length]
	return answer
}

func copyOf(array []uint16, size int) []uint16 {
	result := make([]uint16, size)
	for i, x := range array {
		if i == size {
			break
		}
		result[i] = x
	}
	return result
}

func (self *arrayContainer) inot(firstOfRange, lastOfRange int) container {
	// determine the span of array indices to be affected
	startIndex := binarySearch(self.content, uint16(firstOfRange))
	if startIndex < 0 {
		startIndex = -startIndex - 1
	}
	lastIndex := binarySearch(self.content, uint16(lastOfRange))
	if lastIndex < 0 {
		lastIndex = -lastIndex - 1 - 1
	}
	currentValuesInRange := lastIndex - startIndex + 1
	spanToBeFlipped := lastOfRange - firstOfRange + 1

	newValuesInRange := spanToBeFlipped - currentValuesInRange
	buffer := make([]uint16, newValuesInRange)
	cardinalityChange := newValuesInRange - currentValuesInRange
	newCardinality := len(self.content) + cardinalityChange
	if cardinalityChange > 0 {
		if newCardinality > len(self.content) {
			if newCardinality >= array_default_max_size {
				return self.toBitmapContainer().inot(firstOfRange, lastOfRange)
			}
			self.content = copyOf(self.content, newCardinality)
		}
		base := lastIndex + 1
		//copy(self.content[lastIndex+1+cardinalityChange:], self.content[lastIndex+1:len(self.content)-1-lastIndex])
		copy(self.content[lastIndex+1+cardinalityChange:], self.content[base:base+len(self.content)-1-lastIndex])

		self.negateRange(buffer, startIndex, lastIndex, firstOfRange, lastOfRange)
	} else { // no expansion needed
		self.negateRange(buffer, startIndex, lastIndex, firstOfRange, lastOfRange)
		if cardinalityChange < 0 {

			for i := startIndex + newValuesInRange; i < newCardinality; i++ {
				self.content[i] = self.content[i-cardinalityChange]
			}
		}
	}
	self.content = self.content[:newCardinality]
	return self
}

func (self *arrayContainer) negateRange(buffer []uint16, startIndex, lastIndex, startRange, lastRange int) {
	// compute the negation into buffer

	outPos := 0
	inPos := startIndex // value here always >= valInRange,
	// until it is exhausted
	// n.b., we can start initially exhausted.

	valInRange := startRange
	for ; valInRange <= lastRange && inPos <= lastIndex; valInRange++ {
		if uint16(valInRange) != self.content[inPos] {
			buffer[outPos] = uint16(valInRange)
			outPos++
		} else {
			inPos++
		}
	}

	// if there are extra items (greater than the biggest
	// pre-existing one in range), buffer them
	for ; valInRange <= lastRange; valInRange++ {
		buffer[outPos] = uint16(valInRange)
		outPos++
	}

	if outPos != len(buffer) {
		//panic("negateRange: outPos " + outPos + " whereas buffer.length=" + len(buffer))
		panic("negateRange: outPos  whereas buffer.length=")
	}

	for i, item := range buffer {
		self.content[i] = item
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func (self *arrayContainer) andArray(value2 *arrayContainer) *arrayContainer {

	desiredcapacity := min(self.getCardinality(), value2.getCardinality())
	answer := newArrayContainerCapacity(desiredcapacity)
	length := intersection2by2(
		self.content,
		value2.content,
		answer.content)
	answer.content = answer.content[:length]
	return answer

}

func (self *arrayContainer) getCardinality() int {
	return len(self.content)
}
func (self *arrayContainer) clone() container {
	ptr := arrayContainer{make([]uint16, len(self.content))}
	copy(ptr.content, self.content[:])
	return &ptr
}
func (self *arrayContainer) contains(x uint16) bool {
	return binarySearch(self.content, x) >= 0
}

func (self *arrayContainer) loadData(bitmapContainer *bitmapContainer) {
	self.content = make([]uint16, bitmapContainer.cardinality, bitmapContainer.cardinality)
	bitmapContainer.fillArray(self.content)
}
func newArrayContainer() *arrayContainer {
	p := new(arrayContainer)
	return p
}

func newArrayContainerCapacity(size int) *arrayContainer {
	p := new(arrayContainer)
	p.content = make([]uint16, 0, size)
	return p
}

func newArrayContainerSize(size int) *arrayContainer {
	p := new(arrayContainer)
	p.content = make([]uint16, size, size)
	return p
}

func newArrayContainerRange(firstOfRun, lastOfRun int) *arrayContainer {
	valuesInRange := lastOfRun - firstOfRun + 1
	this := newArrayContainerCapacity(valuesInRange)
	for i := 0; i < valuesInRange; i++ {
		this.content = append(this.content, uint16(firstOfRun+i))
	}
	return this
}
