package roaring

import (
	"encoding/binary"
	"io"
	"unsafe"
)

type arrayContainer struct {
	content []uint16
}

// writes the content (omitting the cardinality)
func (b *arrayContainer) writeTo(stream io.Writer) (int, error) {
	// Write set
	err := binary.Write(stream, binary.LittleEndian, b.content)
	if err != nil {
		return 0, err
	}
	return 2 * len(b.content), nil
}

func (b *arrayContainer) readFrom(stream io.Reader) (int, error) {
	err := binary.Read(stream, binary.LittleEndian, b.content)
	if err != nil {
		return 0, err
	}
	return 2 * len(b.content), nil
}

func (ac *arrayContainer) fillLeastSignificant16bits(x []int, i, mask int) {
	for k := 0; k < len(ac.content); k++ {
		x[k+i] = toIntUnsigned(ac.content[k]) | mask
	}
}

func (ac *arrayContainer) getShortIterator() shortIterable {
	return &shortIterator{ac.content, 0}
}

func (ac *arrayContainer) getSizeInBytes() int {
	// unsafe.Sizeof calculates the memory used by the top level of the slice
	// descriptor - not including the size of the memory referenced by the slice.
	// http://golang.org/pkg/unsafe/#Sizeof
	return ac.getCardinality()*2 + int(unsafe.Sizeof(ac.content))
}

func (ac *arrayContainer) serializedSizeInBytes() int {
	// based on https://golang.org/src/pkg/encoding/binary/binary.go#265
	// there is no serialization overhead for writing an array of fixed size vals
	return ac.getCardinality() * 2
}

func (ac *arrayContainer) addRange(firstOfRange, lastOfRange int) container {
	indexstart := binarySearch(ac.content, uint16(firstOfRange))
	if indexstart < 0 {
		indexstart = -indexstart - 1
	}
	indexend := binarySearch(ac.content, uint16(lastOfRange-1))
	if indexend < 0 {
		indexend = -indexend - 1
	} else {
		indexend++
	}
	rangelength := lastOfRange - firstOfRange

	newcardinality := indexstart + (ac.getCardinality() - indexend) + rangelength
	if newcardinality >= arrayDefaultMaxSize {
		a := ac.toBitmapContainer()
		return a.iaddRange(firstOfRange, lastOfRange)
	}
	answer := &arrayContainer{make([]uint16, newcardinality)}
	copy(answer.content[:indexstart], ac.content[:indexstart])
	copy(answer.content[indexstart+rangelength:], ac.content[indexend:])
	for k := 0; k < rangelength; k++ {
		answer.content[k+indexstart] = uint16(firstOfRange + k)
	}
	return answer
}

func (ac *arrayContainer) removeRange(firstOfRange, lastOfRange int) container {
	indexstart := binarySearch(ac.content, uint16(firstOfRange))
	if indexstart < 0 {
		indexstart = -indexstart - 1
	}
	indexend := binarySearch(ac.content, uint16(lastOfRange-1))
	if indexend < 0 {
		indexend = -indexend - 1
	} else {
		indexend++
	}
	rangelength := indexend - indexstart
	answer := &arrayContainer{make([]uint16, ac.getCardinality()-rangelength)}
	copy(answer.content[:indexstart], ac.content[:indexstart])
	copy(answer.content[indexstart:], ac.content[indexstart+rangelength:])
	return answer
}

func (ac *arrayContainer) iaddRange(firstOfRange, lastOfRange int) container {
	indexstart := binarySearch(ac.content, uint16(firstOfRange))
	if indexstart < 0 {
		indexstart = -indexstart - 1
	}
	indexend := binarySearch(ac.content, uint16(lastOfRange-1))
	if indexend < 0 {
		indexend = -indexend - 1
	} else {
		indexend++
	}
	rangelength := lastOfRange - firstOfRange
	newcardinality := indexstart + (ac.getCardinality() - indexend) + rangelength
	if newcardinality >= arrayDefaultMaxSize {
		a := ac.toBitmapContainer()
		return a.iaddRange(firstOfRange, lastOfRange)
	}
	if cap(ac.content) < newcardinality {
		tmp := make([]uint16, newcardinality, newcardinality)
		copy(tmp[:indexstart], ac.content[:indexstart])
		ac.content = tmp
	} else {
		ac.content = ac.content[:newcardinality]
	}
	copy(ac.content[indexstart+rangelength:], ac.content[indexend:])
	for k := 0; k < rangelength; k++ {
		ac.content[k+indexstart] = uint16(firstOfRange + k)
	}
	return ac
}

func (ac *arrayContainer) iremoveRange(firstOfRange, lastOfRange int) container {
	indexstart := binarySearch(ac.content, uint16(firstOfRange))
	if indexstart < 0 {
		indexstart = -indexstart - 1
	}
	indexend := binarySearch(ac.content, uint16(lastOfRange-1))
	if indexend < 0 {
		indexend = -indexend - 1
	} else {
		indexend++
	}
	rangelength := indexend - indexstart
	answer := ac
	copy(answer.content[indexstart:], ac.content[indexstart+rangelength:])
	answer.content = answer.content[:ac.getCardinality()-rangelength]
	return answer
}

func (ac *arrayContainer) not(firstOfRange, lastOfRange int) container {
	if firstOfRange > lastOfRange {
		return ac.clone()
	}

	// determine the span of array indices to be affected^M
	startIndex := binarySearch(ac.content, uint16(firstOfRange))
	if startIndex < 0 {
		startIndex = -startIndex - 1
	}
	lastIndex := binarySearch(ac.content, uint16(lastOfRange))
	if lastIndex < 0 {
		lastIndex = -lastIndex - 2
	}
	currentValuesInRange := lastIndex - startIndex + 1
	spanToBeFlipped := lastOfRange - firstOfRange + 1
	newValuesInRange := spanToBeFlipped - currentValuesInRange
	cardinalityChange := newValuesInRange - currentValuesInRange
	newCardinality := len(ac.content) + cardinalityChange

	if newCardinality >= arrayDefaultMaxSize {
		return ac.toBitmapContainer().not(firstOfRange, lastOfRange)
	}
	answer := newArrayContainer()
	answer.content = make([]uint16, newCardinality, newCardinality) //a hack for sure

	copy(answer.content, ac.content[:startIndex])
	outPos := startIndex
	inPos := startIndex
	valInRange := firstOfRange
	for ; valInRange <= lastOfRange && inPos <= lastIndex; valInRange++ {
		if uint16(valInRange) != ac.content[inPos] {
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

	for i := lastIndex + 1; i < len(ac.content); i++ {
		answer.content[outPos] = ac.content[i]
		outPos++
	}
	answer.content = answer.content[:newCardinality]
	return answer

}

func (ac *arrayContainer) equals(o interface{}) bool {
	srb, ok := o.(*arrayContainer)
	if ok {
		if len(srb.content) != len(ac.content) {
			return false
		}
		for i := 0; i < len(ac.content); i++ {
			if ac.content[i] != srb.content[i] {
				return false
			}
		}
		return true
	}
	return false
}

func (ac *arrayContainer) toBitmapContainer() *bitmapContainer {
	bc := newBitmapContainer()
	bc.loadData(ac)
	return bc

}
func (ac *arrayContainer) add(x uint16) container {
	loc := binarySearch(ac.content, x)
	if loc < 0 {
		if len(ac.content) >= arrayDefaultMaxSize {
			a := ac.toBitmapContainer()
			a.add(x)
			return a
		}
		s := ac.content
		i := -loc - 1
		s = append(s, 0)
		copy(s[i+1:], s[i:])
		s[i] = x
		ac.content = s
	}
	return ac
}

func (ac *arrayContainer) remove(x uint16) container {
	loc := binarySearch(ac.content, x)
	if loc >= 0 {
		s := ac.content
		s = append(s[:loc], s[loc+1:]...)
		ac.content = s
	}
	return ac
}

func (ac *arrayContainer) or(a container) container {
	switch a.(type) {
	case *arrayContainer:
		return ac.orArray(a.(*arrayContainer))
	case *bitmapContainer:
		return a.or(ac)
	}
	return nil
}

func (ac *arrayContainer) ior(a container) container {
	switch a.(type) {
	case *arrayContainer:
		return ac.orArray(a.(*arrayContainer))
	case *bitmapContainer:
		return a.ior(ac)
	}
	return nil
}

func (ac *arrayContainer) lazyIOR(a container) container {
	switch a.(type) {
	case *arrayContainer:
		return ac.orArray(a.(*arrayContainer))
	case *bitmapContainer:
		return a.lazyIOR(ac)
	}
	return nil
}

func (ac *arrayContainer) orArray(value2 *arrayContainer) container {
	value1 := ac
	maxPossibleCardinality := value1.getCardinality() + value2.getCardinality()
	if maxPossibleCardinality > arrayDefaultMaxSize { // it could be a bitmap!^M
		bc := newBitmapContainer()
		for k := 0; k < len(value2.content); k++ {
			i := uint(toIntUnsigned(value2.content[k])) >> 6
			bc.bitmap[i] |= (1 << (value2.content[k] % 64))
		}
		for k := 0; k < len(ac.content); k++ {
			i := uint(toIntUnsigned(ac.content[k])) >> 6
			bc.bitmap[i] |= (1 << (ac.content[k] % 64))
		}
		bc.cardinality = int(popcntSlice(bc.bitmap))
		if bc.cardinality <= arrayDefaultMaxSize {
			return bc.toArrayContainer()
		}
		return bc
	}
	answer := newArrayContainerCapacity(maxPossibleCardinality)
	nl := union2by2(value1.content, value2.content, answer.content)
	answer.content = answer.content[:nl] // reslice to match actual used capacity
	return answer
}

func (ac *arrayContainer) and(a container) container {
	switch a.(type) {
	case *arrayContainer:
		return ac.andArray(a.(*arrayContainer))
	case *bitmapContainer:
		return a.and(ac)
	}
	return nil
}

func (ac *arrayContainer) xor(a container) container {
	switch a.(type) {
	case *arrayContainer:
		return ac.xorArray(a.(*arrayContainer))
	case *bitmapContainer:
		return a.xor(ac)
	}
	return nil
}

func (ac *arrayContainer) xorArray(value2 *arrayContainer) container {
	value1 := ac
	totalCardinality := value1.getCardinality() + value2.getCardinality()
	if totalCardinality > arrayDefaultMaxSize { // it could be a bitmap!
		bc := newBitmapContainer()
		for k := 0; k < len(value2.content); k++ {
			i := uint(toIntUnsigned(value2.content[k])) >> 6
			bc.bitmap[i] ^= (uint64(1) << (value2.content[k] % 64))
		}
		for k := 0; k < len(ac.content); k++ {
			i := uint(toIntUnsigned(ac.content[k])) >> 6
			bc.bitmap[i] ^= (uint64(1) << (ac.content[k] % 64))
		}
		bc.computeCardinality()
		if bc.cardinality <= arrayDefaultMaxSize {
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

func (ac *arrayContainer) andNot(a container) container {
	switch a.(type) {
	case *arrayContainer:
		return ac.andNotArray(a.(*arrayContainer))
	case *bitmapContainer:
		return ac.andNotBitmap(a.(*bitmapContainer))
	}
	return nil
}

func (ac *arrayContainer) andNotArray(value2 *arrayContainer) container {
	value1 := ac
	desiredcapacity := value1.getCardinality()
	answer := newArrayContainerCapacity(desiredcapacity)
	length := difference(value1.content, value2.content, answer.content)
	answer.content = answer.content[:length]
	return answer
}

func (ac *arrayContainer) andNotBitmap(value2 *bitmapContainer) container {
	desiredcapacity := ac.getCardinality()
	answer := newArrayContainerCapacity(desiredcapacity)
	answer.content = answer.content[:desiredcapacity]
	pos := 0
	for _,v := range ac.content {
		if ! value2.contains(v) {
			answer.content[pos] = v
			pos++
		}
	}
	answer.content = answer.content[:pos]
	return answer
}

//  TODO: fully implement inplace andNots for performance reasons (current function unused)
func (ac *arrayContainer) iandNotBitmap(value2 *bitmapContainer) container {
	pos := 0
	for _,v := range ac.content {
		if ! value2.contains(v) {
			ac.content[pos] = v
			pos++
		}
	}
	ac.content = ac.content[:pos]
	return ac
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

func (ac *arrayContainer) inot(firstOfRange, lastOfRange int) container {
	// determine the span of array indices to be affected
	startIndex := binarySearch(ac.content, uint16(firstOfRange))
	if startIndex < 0 {
		startIndex = -startIndex - 1
	}
	lastIndex := binarySearch(ac.content, uint16(lastOfRange))
	if lastIndex < 0 {
		lastIndex = -lastIndex - 1 - 1
	}
	currentValuesInRange := lastIndex - startIndex + 1
	spanToBeFlipped := lastOfRange - firstOfRange + 1

	newValuesInRange := spanToBeFlipped - currentValuesInRange
	buffer := make([]uint16, newValuesInRange)
	cardinalityChange := newValuesInRange - currentValuesInRange
	newCardinality := len(ac.content) + cardinalityChange
	if cardinalityChange > 0 {
		if newCardinality > len(ac.content) {
			if newCardinality >= arrayDefaultMaxSize {
				return ac.toBitmapContainer().inot(firstOfRange, lastOfRange)
			}
			ac.content = copyOf(ac.content, newCardinality)
		}
		base := lastIndex + 1
		copy(ac.content[lastIndex+1+cardinalityChange:], ac.content[base:base+len(ac.content)-1-lastIndex])

		ac.negateRange(buffer, startIndex, lastIndex, firstOfRange, lastOfRange)
	} else { // no expansion needed
		ac.negateRange(buffer, startIndex, lastIndex, firstOfRange, lastOfRange)
		if cardinalityChange < 0 {

			for i := startIndex + newValuesInRange; i < newCardinality; i++ {
				ac.content[i] = ac.content[i-cardinalityChange]
			}
		}
	}
	ac.content = ac.content[:newCardinality]
	return ac
}

func (ac *arrayContainer) negateRange(buffer []uint16, startIndex, lastIndex, startRange, lastRange int) {
	// compute the negation into buffer

	outPos := 0
	inPos := startIndex // value here always >= valInRange,
	// until it is exhausted
	// n.b., we can start initially exhausted.

	valInRange := startRange
	for ; valInRange <= lastRange && inPos <= lastIndex; valInRange++ {
		if uint16(valInRange) != ac.content[inPos] {
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
		ac.content[i] = item
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func (ac *arrayContainer) andArray(value2 *arrayContainer) *arrayContainer {
	desiredcapacity := min(ac.getCardinality(), value2.getCardinality())
	answer := newArrayContainerCapacity(desiredcapacity)
	length := intersection2by2(
		ac.content,
		value2.content,
		answer.content)
	answer.content = answer.content[:length]
	return answer

}

func (ac *arrayContainer) getCardinality() int {
	return len(ac.content)
}

func (ac *arrayContainer) rank(x uint16) int {
	answer := binarySearch(ac.content, x)
	if answer >= 0 {
		return answer + 1
	} else {
		return -answer - 1
	}
}

func (ac *arrayContainer) selectInt(x uint16) int {
	return int(ac.content[x])
}

func (ac *arrayContainer) clone() container {
	ptr := arrayContainer{make([]uint16, len(ac.content))}
	copy(ptr.content, ac.content[:])
	return &ptr
}

func (ac *arrayContainer) contains(x uint16) bool {
	return binarySearch(ac.content, x) >= 0
}

func (ac *arrayContainer) loadData(bitmapContainer *bitmapContainer) {
	ac.content = make([]uint16, bitmapContainer.cardinality, bitmapContainer.cardinality)
	bitmapContainer.fillArray(ac.content)
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
