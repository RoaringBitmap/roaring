package goroaring

import (
	"bytes"
	"encoding/gob"
	"io"
)

const (
	ARRAY_DEFAULT_MAX_SIZE = 4096
	MAX_CAPACITY           = 1 << 16
)

type Container interface {
	Clone() Container
	And(Container) Container
	AndNot(Container) Container
	Inot(firstOfRange, lastOfRange int) Container
	GetCardinality() int
	Add(short) Container
	Not(start, final int) Container
	Xor(r Container) Container
	//	Or(r Container) Container
	//	Contains(i int) bool
	//	ToArray() []int
}

type ArrayContainer struct {
	cardinality int
	content     []short
}

func (self *ArrayContainer) Not(firstOfRange, lastOfRange int) Container {
	if firstOfRange > lastOfRange {
		return self.Clone()
	}

	// determine the span of array indices to be affected^M
	startIndex := Unsigned_binarySearch(self.content, 0, self.cardinality, short(firstOfRange))
	if startIndex < 0 {
		startIndex = -startIndex - 1
	}
	lastIndex := Unsigned_binarySearch(self.content, 0, self.cardinality, short(lastOfRange))
	if lastIndex < 0 {
		lastIndex = -lastIndex - 2
	}
	currentValuesInRange := lastIndex - startIndex + 1
	spanToBeFlipped := lastOfRange - firstOfRange + 1
	newValuesInRange := spanToBeFlipped - currentValuesInRange
	cardinalityChange := newValuesInRange - currentValuesInRange
	newCardinality := self.cardinality + cardinalityChange

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

	for i := lastIndex + 1; i < self.cardinality; i++ {
		answer.content[outPos] = self.content[i]
		outPos++
	}
	answer.cardinality = newCardinality
	return answer

}

func (self *ArrayContainer) ToBitmapContainer() *BitmapContainer {
	bc := NewBitmapContainer()
	bc.loadData(self)
	return bc

}
func (self *ArrayContainer) Add(x short) Container {
	if self.cardinality >= ARRAY_DEFAULT_MAX_SIZE {
		a := self.ToBitmapContainer()
		a.Add(x)
		return a
	}
	if (self.cardinality == 0) || (x > self.content[self.cardinality-1]) {
		self.content = append(self.content, x)
		self.cardinality++
		return self
	}
	loc := Unsigned_binarySearch(self.content, 0, self.cardinality, x)

	if loc < 0 {
		s := self.content
		i := -loc - 1
		s = append(s, 0)
		copy(s[i+1:], s[i:])
		s[i] = x
		self.content = s
		self.cardinality++
	}
	return self
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
		for k := 0; k < value2.cardinality; k++ {
			i := uint(ToIntUnsigned(value2.content[k])) >> 6
			bc.bitmap[i] ^= (1 << value2.content[k])
		}
		for k := 0; k < self.cardinality; k++ {
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
	answer.cardinality = Unsigned_ExclusiveUnionb2by2(value1.content, value1.GetCardinality(), value2.content, value2.GetCardinality(), answer.content)
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
	answer.cardinality = Unsigned_difference(value1.content, value1.GetCardinality(), value2.content, value2.GetCardinality(), answer.content)
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
	startIndex := Unsigned_binarySearch(self.content, 0, self.cardinality, short(firstOfRange))
	if startIndex < 0 {
		startIndex = -startIndex - 1
	}
	lastIndex := Unsigned_binarySearch(self.content, 0, self.cardinality, short(lastOfRange))
	if lastIndex < 0 {
		lastIndex = -lastIndex - 1 - 1
	}
	currentValuesInRange := lastIndex - startIndex + 1
	spanToBeFlipped := lastOfRange - firstOfRange + 1
	newValuesInRange := spanToBeFlipped - currentValuesInRange
	buffer := make([]short, newValuesInRange)
	cardinalityChange := newValuesInRange - currentValuesInRange
	newCardinality := self.cardinality + cardinalityChange

	if cardinalityChange > 0 {
		if newCardinality > len(self.content) {
			if newCardinality >= ARRAY_DEFAULT_MAX_SIZE {
				return self.ToBitmapContainer().Inot(firstOfRange, lastOfRange)
			}
			self.content = CopyOf(self.content, newCardinality)
		}
		for pos := self.cardinality - 1; pos > lastIndex; pos-- {
			self.content[pos+cardinalityChange] = self.content[pos]
		}
		self.negateRange(buffer, startIndex, lastIndex, firstOfRange, lastOfRange)
	} else { // no expansion needed
		self.negateRange(buffer, startIndex, lastIndex, firstOfRange, lastOfRange)
		if cardinalityChange < 0 {

			for i := startIndex + newValuesInRange; i < newCardinality; i++ {
				self.content[i] = self.content[i-cardinalityChange]
			}
		}
	}
	self.cardinality = newCardinality
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
	answer.cardinality = Unsigned_intersect2by2(
		self.content, self.GetCardinality(),
		value2.content, value2.GetCardinality(),
		answer.content)
	return answer

}

func (self *ArrayContainer) GetCardinality() int {
	return self.cardinality
}
func (self *ArrayContainer) Clone() Container {
	ptr := new(ArrayContainer)
	return ptr
}

func (self *ArrayContainer) loadData(bitmapContainer *BitmapContainer) {
	self.cardinality = bitmapContainer.cardinality
	bitmapContainer.fillArray(self.content)
}
func NewArrayContainer() *ArrayContainer {
	p := new(ArrayContainer)
	p.cardinality = 0
	return p
}

func NewArrayContainerCapacity(size int) *ArrayContainer {
	p := new(ArrayContainer)
	p.cardinality = 0
	p.content = make([]short, size, size)
	return p
}
func NewArrayContainerRange(firstOfRun, lastOfRun int) *ArrayContainer {
	valuesInRange := lastOfRun - firstOfRun + 1
	this := NewArrayContainerCapacity(valuesInRange)
	for i := 0; i < valuesInRange; i++ {
		this.content[i] = short(firstOfRun + i)
	}
	this.cardinality = valuesInRange
	return this
}

type BitmapContainer struct {
	cardinality int
	bitmap      []int64
}

func NewBitmapContainer() *BitmapContainer {
	p := new(BitmapContainer)
	size := (1 << 16) / 64
	p.bitmap = make([]int64, size, size)
	return p
}

func fill(arr []int64, val int64) {
	for i := range arr {
		arr[i] = val
	}
}
func fillRange(arr []int64, start, end int, val int64) {
	for i := start; i < end; i++ {
		arr[i] = val
	}
}

func NewBitmapContainerwithRange(firstOfRun, lastOfRun int) *BitmapContainer {
	this := NewBitmapContainer()
	this.cardinality = lastOfRun - firstOfRun + 1
	if this.cardinality == MAX_CAPACITY {
		fill(this.bitmap, -1)
	} else {
		firstWord := firstOfRun / 64
		lastWord := lastOfRun / 64
		zeroPrefixLength := uint64(firstOfRun & 63)
		zeroSuffixLength := uint64(63 - (lastOfRun & 63))

		fillRange(this.bitmap, firstWord, lastWord+1, -1)
		this.bitmap[firstWord] ^= ((1 << zeroPrefixLength) - 1)
		blockOfOnes := (int64(1) << zeroSuffixLength) - 1
		maskOnLeft := blockOfOnes << (uint64(64) - zeroSuffixLength)
		this.bitmap[lastWord] ^= maskOnLeft
	}
	return this
}

func (self *BitmapContainer) Add(i short) Container {
	x := int(i)
	previous := self.bitmap[x/64]
	self.bitmap[x/64] |= (1 << uint(x))
	self.cardinality += int(uint(previous^self.bitmap[x/64]) >> uint(x))
	return self
}

func (self *BitmapContainer) GetCardinality() int {
	return self.cardinality
}

func (self *BitmapContainer) Clone() Container {
	ptr := new(BitmapContainer)
	//need to copy the data over
	return ptr
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
		copy(answer.bitmap[rangeLastWord+1:], self.bitmap[rangeLastWord+1:len(self.bitmap)-(rangeLastWord+1)])

	}

	// unfortunately, the simple expression gives the wrong mask for
	// rangeLastBitPos==63
	// no branchless way comes to mind
	maskOnLeft := int64(-1)
	if rangeLastBitPos != 63 {
		maskOnLeft = (1 << uint((rangeLastBitPos + 1))) - 1
	}
	mask := int64(-1) // now zero out stuff in the prefix

	mask ^= (1 << uint(rangeFirstBitPos)) - 1

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
		answer.cardinality += 1 - 2*int(uint(answer.bitmap[index]&(1<<value2.content[k]))>>value2.content[k])

		answer.bitmap[index] = answer.bitmap[index] ^ (1 << value2.content[k])
	}
	if answer.cardinality <= ARRAY_DEFAULT_MAX_SIZE {
		return answer.ToArrayContainer()
	}
	return answer
}

func (self *BitmapContainer) XorBitmap(value2 *BitmapContainer) Container {
	newCardinality := 0
	for k := 0; k < len(self.bitmap); k++ {
		newCardinality += BitCount(self.bitmap[k] ^ value2.bitmap[k])
	}
	if newCardinality > ARRAY_DEFAULT_MAX_SIZE {
		answer := NewBitmapContainer()
		for k := 0; k < len(answer.bitmap); k++ {
			answer.bitmap[k] = self.bitmap[k] ^ value2.bitmap[k]
		}
		answer.cardinality = newCardinality
		return answer
	}
	ac := NewArrayContainerCapacity(newCardinality)
	FillArrayXOR(ac.content, self.bitmap, value2.bitmap)
	ac.cardinality = newCardinality
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
			answer.content[answer.cardinality] = value2.content[k]
			answer.cardinality++
		}
	}
	return answer

}

func (self *BitmapContainer) AndBitmap(value2 *BitmapContainer) Container {
	newcardinality := 0
	for k := 0; k < len(self.bitmap); k++ {
		newcardinality += BitCount(self.bitmap[k] & value2.bitmap[k])
	}
	if newcardinality > ARRAY_DEFAULT_MAX_SIZE {
		answer := NewBitmapContainer()
		for k := 0; k < len(answer.bitmap); k++ {
			answer.bitmap[k] = self.bitmap[k] & value2.bitmap[k]
		}
		answer.cardinality = newcardinality
		return answer
	}
	ac := NewArrayContainerCapacity(newcardinality)
	FillArrayAND(ac.content, self.bitmap, value2.bitmap)
	ac.cardinality = newcardinality
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

	for k := 0; k < value2.cardinality; k++ {
		i := uint(ToIntUnsigned(value2.content[k])) >> 6
		answer.bitmap[i] = answer.bitmap[i] &^ (1 << value2.content[k])
		answer.cardinality -= int(uint(answer.bitmap[i]^self.bitmap[i]) >> value2.content[k])
	}
	if answer.cardinality <= ARRAY_DEFAULT_MAX_SIZE {
		return answer.ToArrayContainer()
	}
	return answer
}

func (self *BitmapContainer) AndNotBitmap(value2 *BitmapContainer) Container {
	newCardinality := 0
	for k := 0; k < len(self.bitmap); k++ {
		newCardinality += BitCount(self.bitmap[k] &^ value2.bitmap[k])
	}
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
	ac.cardinality = newCardinality
	return ac
}

func (self *BitmapContainer) Contains(i short) bool { //testbit
	x := int(i)
	return (self.bitmap[x/64] & (1 << uint(x))) != 0
}
func (self *BitmapContainer) loadData(arrayContainer *ArrayContainer) {
	self.cardinality = arrayContainer.cardinality
	for k := 0; k < arrayContainer.cardinality; k++ {
		x := arrayContainer.content[k]
		self.bitmap[int(x)/64] |= (1 << uint(x))
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
			array[pos] = short(k*64 + BitCount(t-1))
			pos++
			bitset ^= t
		}
	}
}

type Element struct {
	key   short
	value Container
}

func (self *Element) Clone() Element {
	var c Element
	c.key = self.key
	c.value = self.value.Clone()
	return c
}

func (self *Element) GobEncode() (buf []byte, err error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	//gob.Register(self.Container)
	err = encoder.Encode(self.key)
	if err != nil {
		return nil, err
	}

	err = encoder.Encode(self.value)
	if err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (self *Element) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)
	err := decoder.Decode(self.key)
	if err != nil {
		return err
	}
	err = decoder.Decode(self.value)
	if err != nil {
		return err
	}
	return nil
}

func NewElement(key short, value Container) *Element {
	ptr := new(Element)
	ptr.key = key
	ptr.value = value
	return ptr
}

type RoaringArray struct {
	array []*Element
}

func NewRoaringArray() *RoaringArray {
	return &RoaringArray{make([]*Element, 0, 0)}
}

func (self *RoaringArray) Append(key short, value Container) {
	self.array = append(self.array, NewElement(key, value))
}

func (self *RoaringArray) AppendCopy(sa RoaringArray, startingindex, end int) {
	for i := startingindex; i < end; i++ {
		self.array = append(self.array, NewElement(sa.array[i].key, sa.array[i].value.Clone()))
	}
}

func (self *RoaringArray) AppendCopiesUntil(sa RoaringArray, stoppingKey short) {
	for i := 0; i < sa.Size(); i++ {
		if sa.array[i].key >= stoppingKey {
			break
		}
		self.array = append(self.array, NewElement(sa.array[i].key, sa.array[i].value.Clone()))
	}
}

func (self *RoaringArray) AppendCopiesAfter(sa RoaringArray, beforeStart short) {
	startLocation := sa.GetIndex(beforeStart)
	if startLocation >= 0 {
		startLocation++
	} else {
		startLocation = -startLocation - 1
	}

	for i := startLocation; i < sa.Size(); i++ {
		self.array = append(self.array, NewElement(sa.array[i].key, sa.array[i].value.Clone()))
	}
}

func (self *RoaringArray) Clear() {
	self.array = make([]*Element, 0, 0)
}

func (self *RoaringArray) Clone() *RoaringArray {
	var sa RoaringArray
	copy(sa.array, self.array)
	return &sa
}

func (self *RoaringArray) ContainsKey(x short) bool {
	return (self.BinarySearch(0, len(self.array), x) >= 0)
}

func (self *RoaringArray) GetContainer(x short) Container {
	i := self.BinarySearch(0, len(self.array), x)
	if i < 0 {
		return nil
	}
	return self.array[i].value
}

func (self *RoaringArray) GetContainerAtIndex(i int) Container {
	return self.array[i].value
}

func (self *RoaringArray) GetIndex(x short) int {
	// before the binary search, we optimize for frequent cases
	size := len(self.array)
	if (size == 0) || (self.array[size-1].key == x) {
		return size - 1
	}
	return self.BinarySearch(0, size, x)
}

func (self *RoaringArray) GetKeyAtIndex(i int) short {
	return self.array[i].key
}

func (self *RoaringArray) insertNewKeyValueAt(i int, key short, value Container) {
	s := self.array
	s = append(s, nil)
	copy(s[i+1:], s[i:])
	s[i] = NewElement(key, value)
	self.array = s
}

func (self *RoaringArray) Remove(key short) bool {
	i := self.BinarySearch(0, len(self.array), key)
	if i >= 0 { // if a new key
		self.RemoveAtIndex(i)
		return true
	}
	return false
}

func (self *RoaringArray) RemoveAtIndex(i int) {
	a := self.array
	copy(a[i:], a[i+1:])
	a[len(a)-1] = nil // or the zero value of T
	a = a[:len(a)-1]
	self.array = a //should be the same reference i think
}

func (self *RoaringArray) setContainerAtIndex(i int, c Container) {
	self.array[i].value = c
}

func (self *RoaringArray) Size() int {
	return len(self.array)
}

func (self *RoaringArray) BinarySearch(begin, end int, key short) int {
	low := begin
	high := end - 1
	ikey := int(key)

	for low <= high {
		middleIndex := int(uint((low + high)) >> 1)
		middleValue := int(self.array[middleIndex].key)

		if middleValue < ikey {
			low = middleIndex + 1
		} else if middleValue > ikey {
			high = middleIndex - 1
		} else {
			return middleIndex
		}
	}
	return -(low + 1)
}

func (self *RoaringArray) Serialize(out io.Writer) error {
	enc := gob.NewEncoder(out)
	err := enc.Encode(len(self.array))
	if err != nil {
		return err
	}
	for _, item := range self.array {
		err = enc.Encode(item)
		if err != nil {
			return err
		}
	}
	return nil
}

func (self *RoaringArray) Deserialize(in io.Reader) error {
	dec := gob.NewDecoder(in)
	var size int
	err := dec.Decode(&size)
	if err != nil {
		return err
	}
	self.array = make([]*Element, size, size)
	for i := 0; i < size; i++ {
		element := new(Element)
		err = dec.Decode(&element)
		if err != nil {
			return err
		}
		self.array[i] = element
	}
	return nil
}

type RoaringBitmap struct {
	highlowcontainer RoaringArray
}

func NewRoaringBitmap() *RoaringBitmap {
	a := new(RoaringBitmap)
	p := NewRoaringArray()
	a.highlowcontainer = *p
	return a
}
func (self *RoaringBitmap) Clone() *RoaringBitmap {

	return &RoaringBitmap{self.highlowcontainer}
}

func And(x1, x2 *RoaringBitmap) *RoaringBitmap {
	answer := NewRoaringBitmap()
	pos1 := 0
	pos2 := 0
	length1 := x1.highlowcontainer.Size()
	length2 := x2.highlowcontainer.Size()
main:
	for {
		if pos1 < length1 && pos2 < length2 {
			s1 := x1.highlowcontainer.GetKeyAtIndex(pos1)
			s2 := x2.highlowcontainer.GetKeyAtIndex(pos2)
			for {
				if s1 < s2 {
					pos1++
					if pos1 == length1 {
						break main
					}
					s1 = x1.highlowcontainer.GetKeyAtIndex(pos1)
				} else if s1 > s2 {
					pos2++
					if pos2 == length2 {
						break main
					}
					s2 = x2.highlowcontainer.GetKeyAtIndex(pos2)
				} else {
					C := x1.highlowcontainer.GetContainerAtIndex(pos1)
					C = C.And(x2.highlowcontainer.GetContainerAtIndex(pos2))

					if C.GetCardinality() > 0 {
						answer.highlowcontainer.Append(s1, C)
						pos1++
						pos2++
						if (pos1 == length1) || (pos2 == length2) {
							break main
						}
						s1 = x1.highlowcontainer.GetKeyAtIndex(pos1)
						s2 = x2.highlowcontainer.GetKeyAtIndex(pos2)
					}
				}
			}
		} else {
			break
		}
	}
	return answer
}
func Xor(x1, x2 *RoaringBitmap) *RoaringBitmap {
	answer := NewRoaringBitmap()
	pos1 := 0
	pos2 := 0
	length1 := x1.highlowcontainer.Size()
	length2 := x2.highlowcontainer.Size()

main:
	for {
		if (pos1 < length1) && (pos2 < length2) {
			s1 := x1.highlowcontainer.GetKeyAtIndex(pos1)
			s2 := x2.highlowcontainer.GetKeyAtIndex(pos2)
			if s1 < s2 {
				answer.highlowcontainer.AppendCopy(x1.highlowcontainer, pos1, x1.highlowcontainer.Size())

				pos1++
				if pos1 == length1 {
					break main
				}
				s1 = x1.highlowcontainer.GetKeyAtIndex(pos1)
			} else if s1 > s2 {
				answer.highlowcontainer.AppendCopy(x2.highlowcontainer, pos2, x2.highlowcontainer.Size())
				pos2++
				if pos2 == length2 {
					break main
				}
				s2 = x2.highlowcontainer.GetKeyAtIndex(pos2)
			} else {
				c := x1.highlowcontainer.GetContainerAtIndex(pos1).Xor(x2.highlowcontainer.GetContainerAtIndex(pos2))
				if c.GetCardinality() > 0 {
					answer.highlowcontainer.Append(s1, c)
				}
				pos1++
				pos2++
				if (pos1 == length1) || (pos2 == length2) {
					break main
				}
				s1 = x1.highlowcontainer.GetKeyAtIndex(pos1)
				s2 = x2.highlowcontainer.GetKeyAtIndex(pos2)
			}
		} else {
			break
		}
	}
	if pos1 == length1 {
		answer.highlowcontainer.AppendCopy(x2.highlowcontainer, pos2, length2)
	} else if pos2 == length2 {
		answer.highlowcontainer.AppendCopy(x1.highlowcontainer, pos1, length1)
	}
	return answer
}

func AndNot(x1, x2 *RoaringBitmap) *RoaringBitmap {
	answer := NewRoaringBitmap()
	pos1 := 0
	pos2 := 0
	length1 := x1.highlowcontainer.Size()
	length2 := x2.highlowcontainer.Size()

main:
	for {
		if pos1 < length1 && pos2 < length2 {
			s1 := x1.highlowcontainer.GetKeyAtIndex(pos1)
			s2 := x2.highlowcontainer.GetKeyAtIndex(pos2)
			for {
				if s1 < s2 {
					answer.highlowcontainer.AppendCopy(x1.highlowcontainer, pos1, x1.highlowcontainer.Size())
					pos1++
					if pos1 == length1 {
						break main
					}
					s1 = x1.highlowcontainer.GetKeyAtIndex(pos1)
				} else if s1 > s2 {
					pos2++
					if pos2 == length2 {
						break main
					}
					s2 = x2.highlowcontainer.GetKeyAtIndex(pos2)
				} else {
					C := x1.highlowcontainer.GetContainerAtIndex(pos1)
					C.AndNot(x2.highlowcontainer.GetContainerAtIndex(pos2))
					if C.GetCardinality() > 0 {
						answer.highlowcontainer.Append(s1, C)
					}
					pos1++
					pos2++
					if (pos1 == length1) || (pos2 == length2) {
						break main
					}
					s1 = x1.highlowcontainer.GetKeyAtIndex(pos1)
					s2 = x2.highlowcontainer.GetKeyAtIndex(pos2)
				}
			}
		} else {
			break
		}
	}
	if pos2 == length2 {
		answer.highlowcontainer.AppendCopy(x1.highlowcontainer, pos1, length1)
	}
	return answer
}

/*
func BitmapOf(dat ...int) Bitmap {
	ans := NewRoaringBitmap()
	for _, i := range dat {
		ans.Add(i)
	}
	return ans
}
*/

func Flip(bm *RoaringBitmap, rangeStart, rangeEnd int) *RoaringBitmap {
	if rangeStart >= rangeEnd {
		return bm.Clone()
	}

	answer := NewRoaringBitmap()
	hbStart := Highbits(rangeStart)
	lbStart := Lowbits(rangeStart)
	hbLast := Highbits(rangeEnd - 1)
	lbLast := Lowbits(rangeEnd - 1)

	// copy the containers before the active area
	answer.highlowcontainer.AppendCopiesUntil(bm.highlowcontainer, hbStart)

	max := ToIntUnsigned(MaxLowBit())
	for hb := hbStart; hb <= hbLast; hb++ {
		containerStart := 0
		if hb == hbStart {
			containerStart = ToIntUnsigned(lbStart)
		}
		containerLast := max
		if hb == hbLast {
			containerLast = ToIntUnsigned(lbLast)
		}

		i := bm.highlowcontainer.GetIndex(hb)
		j := answer.highlowcontainer.GetIndex(hb)

		if i >= 0 {
			c := bm.highlowcontainer.GetContainerAtIndex(i).Not(containerStart, containerLast)
			if c.GetCardinality() > 0 {
				answer.highlowcontainer.insertNewKeyValueAt(-j-1, hb, c)
			}

		} else { // *think* the range of ones must never be
			// empty.
			answer.highlowcontainer.insertNewKeyValueAt(-j-1, hb,
				RangeOfOnes(containerStart, containerLast))
		}
	}
	// copy the containers after the active area.
	answer.highlowcontainer.AppendCopiesAfter(bm.highlowcontainer, hbLast)

	return answer
}

func RangeOfOnes(start, last int) Container {
	if (last - start + 1) > ARRAY_DEFAULT_MAX_SIZE {
		return NewBitmapContainerwithRange(start, last)
	}

	return NewArrayContainerRange(start, last)
}

func Unsigned_intersect2by2(
	set1 []short,
	length1 int,
	set2 []short,
	length2 int,
	buffer []short) int {
	if len(set1)*64 < len(set2) {
		return Unsigned_onesidedgallopingintersect2by2(set1,
			length1, set2, length2, buffer)
	} else if len(set2)*64 < len(set1) {
		return Unsigned_onesidedgallopingintersect2by2(set2,
			length2, set1, length1, buffer)
	}

	return Unsigned_localintersect2by2(set1, length1, set2,
		length2, buffer)
}

func Unsigned_localintersect2by2(
	set1 []short,
	length1 int,
	set2 []short,
	length2 int,
	buffer []short) int {

	if (0 == length1) || (0 == length2) {
		return 0
	}
	k1 := 0
	k2 := 0
	pos := 0

mainwhile:
	for {
		if set2[k2] < set1[k1] {
			for {
				k2++
				if k2 == length2 {
					break mainwhile
				}
				if set2[k2] < set1[k1] {
					break
				}
			}
		}
		if set1[k1] < set2[k2] {
			for {
				k1++
				if k1 == length1 {
					break mainwhile
				}
				if set1[k1] < set2[k2] {
					break
				}
			}

		} else {
			// (set2[k2] == set1[k1])
			buffer[pos] = set1[k1]
			pos++
			k1++
			if k1 == length1 {
				break
			}
			k2++
			if k2 == length2 {
				break
			}
		}
	}
	return pos
}

func Unsigned_onesidedgallopingintersect2by2(
	smallset []short,
	smalllength int,
	largeset []short,
	largelength int,
	buffer []short) int {

	if 0 == smalllength {
		return 0
	}
	k1 := 0
	k2 := 0
	pos := 0
mainwhile:
	for {
		if largeset[k1] < smallset[k2] {
			k1 = AdvanceUntil(largeset, k1, largelength, smallset[k2])
			if k1 == largelength {
				break mainwhile
			}
		}
		if smallset[k2] < largeset[k1] {
			k2++
			if k2 == smalllength {
				break mainwhile
			}
		} else {

			buffer[pos] = smallset[k2]
			pos++
			k2++
			if k2 == smalllength {
				break
			}
			k1 = AdvanceUntil(largeset, k1, largelength, smallset[k2])
			if k1 == largelength {
				break mainwhile
			}
		}

	}
	return pos
}

func Unsigned_binarySearch(array []short, begin, end int, k short) int {
	low := begin
	high := end - 1
	ikey := int(k)

	for low <= high {
		middleIndex := int(uint(low+high) >> 1)
		middleValue := int(array[middleIndex])

		if middleValue < ikey {
			low = middleIndex + 1
		} else if middleValue > ikey {
			high = middleIndex - 1
		} else {
			return middleIndex
		}
	}
	return -(low + 1)
}

func fillArrayXOR(container []short, bitmap1, bitmap2 []uint64) {
	pos := 0
	if len(bitmap1) != len(bitmap2) {
		panic("fillArrayXOR args not the same")
	}
	for k := 0; k < len(bitmap1); k++ {
		bitset := bitmap1[k] ^ bitmap2[k]
		for bitset != 0 {
			t := bitset & -bitset
			container[pos] = short(k*64 + BitCount(int64(t)-1))
			pos++
			bitset ^= t
		}
	}
}
func Unsigned_difference(set1 []short, length1 int, set2 []short, length2 int, buffer []short) int {
	pos := 0
	k1 := 0
	k2 := 0
	if 0 == length2 {
		for k := 0; k < length1; k++ {
			buffer[k] = set1[k]
		}
		return length1
	}
	if 0 == length1 {
		return 0
	}
	for {
		if set1[k1] < set2[k2] {
			buffer[pos] = set1[k1]
			pos++
			k1++
			if k1 >= length1 {
				break
			}
		} else if set1[k1] == set2[k2] {
			k1++
			k2++
			if k1 >= length1 {
				break
			}
			if k2 >= length2 {
				for ; k1 < length1; k1++ {
					buffer[pos] = set1[k1]
					pos++
				}
				break
			}
		} else { // if (val1>val2)
			k2++
			if k2 >= length2 {
				for ; k1 < length1; k1++ {
					buffer[pos] = set1[k1]
					pos++
				}
				break
			}
		}
	}
	return pos
}

func Unsigned_ExclusiveUnionb2by2(set1 []short, length1 int, set2 []short, length2 int, buffer []short) int {
	pos := 0
	k1 := 0
	k2 := 0
	if 0 == length2 {
		//	System.arraycopy(set1, 0, buffer, 0, length1);
		copy(buffer, set1[:length1])
		return length1
	}
	if 0 == length1 {
		//	System.arraycopy(set2, 0, buffer, 0, length2);
		copy(buffer, set2[:length2])
		return length2
	}
	for {
		if ToIntUnsigned(set1[k1]) < ToIntUnsigned(set2[k2]) {
			buffer[pos] = set1[k1]
			pos++
			k1++
			if k1 >= length1 {
				for ; k2 < length2; k2++ {
					buffer[pos] = set2[k2]
					pos++
				}
				break
			}
		} else if ToIntUnsigned(set1[k1]) == ToIntUnsigned(set2[k2]) {
			k1++
			k2++
			if k1 >= length1 {
				for ; k2 < length2; k2++ {
					buffer[pos] = set2[k2]
					pos++
				}
				break
			}
			if k2 >= length2 {
				for ; k1 < length1; k1++ {
					buffer[pos] = set1[k1]
					pos++
				}
				break
			}
		} else { // if (val1>val2)
			buffer[pos] = set2[k2]
			pos++
			k2++
			if k2 >= length2 {
				for ; k1 < length1; k1++ {
					buffer[pos] = set1[k1]
					pos++
				}
				break
			}
		}
	}
	return pos
}
