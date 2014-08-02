package goroaring

type ArrayContainer struct {
	content []short
}

func (self *ArrayContainer) FillLeastSignificant16bits(x []int, i, mask int) {
	for k := 0; k < len(self.content); k++ {
		x[k+i] = ToIntUnsigned(self.content[k]) | mask
	}
}

func (self *ArrayContainer) GetShortIterator() ShortIterable {
	return &ShortIterator{self.content, 0}
}

func (self *ArrayContainer) Not(firstOfRange, lastOfRange int) Container {
	if firstOfRange > lastOfRange {
		return self.Clone()
	}

	// determine the span of array indices to be affected^M
	startIndex := binarySearch(self.content, short(firstOfRange))
	if startIndex < 0 {
		startIndex = -startIndex - 1
	}
	lastIndex := binarySearch(self.content, short(lastOfRange))
	if lastIndex < 0 {
		lastIndex = -lastIndex - 2
	}
	currentValuesInRange := lastIndex - startIndex + 1
	spanToBeFlipped := lastOfRange - firstOfRange + 1
	newValuesInRange := spanToBeFlipped - currentValuesInRange
	cardinalityChange := newValuesInRange - currentValuesInRange
	newCardinality := len(self.content) + cardinalityChange

	if newCardinality >= ARRAY_DEFAULT_MAX_SIZE {
		return self.ToBitmapContainer().Not(firstOfRange, lastOfRange)
	}
	answer := NewArrayContainerCapacity(newCardinality)

	copy(answer.content, self.content[:startIndex])
	outPos := startIndex
	inPos := startIndex
	valInRange := firstOfRange
	for ; valInRange <= lastOfRange && inPos <= lastIndex; valInRange++ {
		if short(valInRange) != self.content[inPos] {
			answer.content[outPos] = short(valInRange)
			outPos++
		} else {
			inPos++
		}
	}

	for ; valInRange <= lastOfRange; valInRange++ {
		answer.content[outPos] = short(valInRange)
		outPos++
	}

	for i := lastIndex + 1; i < len(self.content); i++ {
		answer.content[outPos] = self.content[i]
		outPos++
	}
	answer.content = answer.content[:newCardinality]
	return answer

}

func (self *ArrayContainer) Equals(o interface{}) bool {
	srb := o.(*ArrayContainer)
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

func (self *ArrayContainer) ToBitmapContainer() *BitmapContainer {
	bc := NewBitmapContainer()
	bc.loadData(self)
	return bc

}
func (self *ArrayContainer) Add(x short) Container {
	if len(self.content) >= ARRAY_DEFAULT_MAX_SIZE {
		a := self.ToBitmapContainer()
		a.Add(x)
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

func (self *ArrayContainer) Or(a Container) Container {
	switch a.(type) {
	case *ArrayContainer:
		return self.OrArray(a.(*ArrayContainer))
	case *BitmapContainer:
		return a.Or(self)
	}
	return nil
}

func (self *ArrayContainer) OrArray(value2 *ArrayContainer) Container {
	value1 := self
	totalCardinality := value1.GetCardinality() + value2.GetCardinality()
	if totalCardinality > ARRAY_DEFAULT_MAX_SIZE { // it could be a bitmap!^M
		bc := NewBitmapContainer()
		for k := 0; k < len(value2.content); k++ {
			i := uint(ToIntUnsigned(value2.content[k])) >> 6
			bc.bitmap[i] |= (1 << (value2.content[k] % 64))
		}
		for k := 0; k < len(self.content); k++ {
			i := uint(ToIntUnsigned(self.content[k])) >> 6
			bc.bitmap[i] |= (1 << (self.content[k] % 64))
		}
		bc.cardinality = 0
		for _, k := range bc.bitmap {
			bc.cardinality += BitCount(k)
		}
		if bc.cardinality <= ARRAY_DEFAULT_MAX_SIZE {
			return bc.ToArrayContainer()
		}
		return bc
	}
	desiredCapacity := totalCardinality
	answer := NewArrayContainerCapacity(desiredCapacity)
	nl := Union2by2(value1.content, value2.content, answer.content)
	answer.content = answer.content[:nl] //what is this voodo?
	return answer
}

func (self *ArrayContainer) And(a Container) Container {
	switch a.(type) {
	case *ArrayContainer:
		return self.AndArray(a.(*ArrayContainer))
	case *BitmapContainer:
		return a.And(self)
	}
	return nil
}

func (self *ArrayContainer) Xor(a Container) Container {
	switch a.(type) {
	case *ArrayContainer:
		return self.XorArray(a.(*ArrayContainer))
	case *BitmapContainer:
		return a.Xor(self)
	}
	return nil
}

func (self *ArrayContainer) XorArray(value2 *ArrayContainer) Container {
	value1 := self
	totalCardinality := value1.GetCardinality() + value2.GetCardinality()
	if totalCardinality > ARRAY_DEFAULT_MAX_SIZE { // it could be a bitmap!^M
		bc := NewBitmapContainer()
		for k := 0; k < len(value2.content); k++ {
			i := uint(ToIntUnsigned(value2.content[k])) >> 6
			bc.bitmap[i] ^= (1 << value2.content[k])
		}
		for k := 0; k < len(self.content); k++ {
			i := uint(ToIntUnsigned(self.content[k])) >> 6
			bc.bitmap[i] ^= (1 << self.content[k])
		}
		bc.cardinality = 0
		for _, k := range bc.bitmap {
			bc.cardinality += BitCount(k)
		}
		if bc.cardinality <= ARRAY_DEFAULT_MAX_SIZE {
			return bc.ToArrayContainer()
		}
		return bc
	}
	desiredCapacity := totalCardinality
	answer := NewArrayContainerCapacity(desiredCapacity)
	length := ExclusiveUnion2by2(value1.content, value2.content, answer.content)
	answer.content = answer.content[:length]
	return answer

}

func (self *ArrayContainer) AndNot(a Container) Container {
	switch a.(type) {
	case *ArrayContainer:
		return self.AndNotArray(a.(*ArrayContainer))
	case *BitmapContainer:
		return a.AndNot(self)
	}
	return nil
}

func (self *ArrayContainer) AndNotArray(value2 *ArrayContainer) Container {
	value1 := self
	desiredcapacity := value1.GetCardinality()
	answer := NewArrayContainerCapacity(desiredcapacity)
	length := Difference(value1.content, value2.content, answer.content)
	answer.content = answer.content[:length]
	return answer
}

func CopyOf(array []short, size int) []short {
	result := make([]short, size)
	for i, x := range array {
		if i == size {
			break
		}
		result[i] = x
	}
	return result
}

func (self *ArrayContainer) Inot(firstOfRange, lastOfRange int) Container {
	// determine the span of array indices to be affected
	startIndex := binarySearch(self.content, short(firstOfRange))
	if startIndex < 0 {
		startIndex = -startIndex - 1
	}
	lastIndex := binarySearch(self.content, short(lastOfRange))
	if lastIndex < 0 {
		lastIndex = -lastIndex - 1 - 1
	}
	currentValuesInRange := lastIndex - startIndex + 1
	spanToBeFlipped := lastOfRange - firstOfRange + 1

	newValuesInRange := spanToBeFlipped - currentValuesInRange
	buffer := make([]short, newValuesInRange)
	cardinalityChange := newValuesInRange - currentValuesInRange
	newCardinality := len(self.content) + cardinalityChange
	if cardinalityChange > 0 {
		if newCardinality > len(self.content) {
			if newCardinality >= ARRAY_DEFAULT_MAX_SIZE {
				return self.ToBitmapContainer().Inot(firstOfRange, lastOfRange)
			}
			self.content = CopyOf(self.content, newCardinality)
		}
		copy(self.content[lastIndex+1+cardinalityChange:], self.content[lastIndex+1:len(self.content)-1-lastIndex])

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

func (self *ArrayContainer) negateRange(buffer []short, startIndex, lastIndex, startRange, lastRange int) {
	// compute the negation into buffer

	outPos := 0
	inPos := startIndex // value here always >= valInRange,
	// until it is exhausted
	// n.b., we can start initially exhausted.

	valInRange := startRange
	for ; valInRange <= lastRange && inPos <= lastIndex; valInRange++ {
		if short(valInRange) != self.content[inPos] {
			buffer[outPos] = short(valInRange)
			outPos++
		} else {
			inPos++
		}
	}

	// if there are extra items (greater than the biggest
	// pre-existing one in range), buffer them
	for ; valInRange <= lastRange; valInRange++ {
		buffer[outPos] = short(valInRange)
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

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func (self *ArrayContainer) AndArray(value2 *ArrayContainer) *ArrayContainer {

	desiredcapacity := Min(self.GetCardinality(), value2.GetCardinality())
	answer := NewArrayContainerCapacity(desiredcapacity)
	length := Intersection2by2(
		self.content,
		value2.content,
		answer.content)
	answer.content = answer.content[:length]
	return answer

}

func (self *ArrayContainer) GetCardinality() int {
	return len(self.content)
}
func (self *ArrayContainer) Clone() Container {
	ptr := ArrayContainer{make([]short, len(self.content))}
	copy(ptr.content, self.content[:])
	return &ptr
}
func (self *ArrayContainer) Contains(x short) bool {
	return binarySearch(self.content, x) >= 0
}

func (self *ArrayContainer) loadData(bitmapContainer *BitmapContainer) {
	self.content = make([]short, bitmapContainer.cardinality)
	bitmapContainer.fillArray(self.content)
}
func NewArrayContainer() *ArrayContainer {
	p := new(ArrayContainer)
	return p
}

func NewArrayContainerCapacity(size int) *ArrayContainer {
	p := new(ArrayContainer)
	p.content = make([]short, 0, size)
	return p
}
func NewArrayContainerRange(firstOfRun, lastOfRun int) *ArrayContainer {
	valuesInRange := lastOfRun - firstOfRun + 1
	this := NewArrayContainerCapacity(valuesInRange)
	for i := 0; i < valuesInRange; i++ {
		this.content[i] = short(firstOfRun + i)
	}
	return this
}
