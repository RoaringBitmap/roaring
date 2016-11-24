package roaring

import (
	"fmt"
	"unsafe"
)

//go:generate msgp -unexported

type arrayContainer struct {
	Content []uint16
}

func (c *arrayContainer) String() string {
	var s string = "{"
	for it := c.getShortIterator(); it.hasNext(); {
		s += fmt.Sprintf("%v, ", it.next())
	}
	return s + "}"
}

func (ac *arrayContainer) fillLeastSignificant16bits(x []uint32, i int, mask uint32) {
	for k := 0; k < len(ac.Content); k++ {
		x[k+i] = uint32(ac.Content[k]) | mask
	}
}

func (ac *arrayContainer) getShortIterator() shortIterable {
	return &shortIterator{ac.Content, 0}
}

// unsafe.Sizeof calculates the memory used by the top level of the slice
// descriptor - not including the size of the memory referenced by the slice.
// http://golang.org/pkg/unsafe/#Sizeof
const arrayBaseSize = int(unsafe.Sizeof([]uint16{}))

func (ac *arrayContainer) getSizeInBytes() int {
	return ac.getCardinality() * 2 // + arrayBaseSize
}

func (ac *arrayContainer) serializedSizeInBytes() int {
	return ac.Msgsize()
	//return ac.getCardinality() * 2 //+ arrayBaseSize
}

func arrayContainerSizeInBytes(card int) int {
	return card * 2 //+ arrayBaseSize
}

// add the values in the range [firstOfRange,lastofRange)
func (ac *arrayContainer) iaddRange(firstOfRange, lastOfRange int) container {
	if firstOfRange >= lastOfRange {
		return ac
	}
	indexstart := binarySearch(ac.Content, uint16(firstOfRange))
	if indexstart < 0 {
		indexstart = -indexstart - 1
	}
	indexend := binarySearch(ac.Content, uint16(lastOfRange-1))
	if indexend < 0 {
		indexend = -indexend - 1
	} else {
		indexend++
	}
	rangelength := lastOfRange - firstOfRange
	newcardinality := indexstart + (ac.getCardinality() - indexend) + rangelength
	if newcardinality > arrayDefaultMaxSize {
		a := ac.toBitmapContainer()
		return a.iaddRange(firstOfRange, lastOfRange)
	}
	if cap(ac.Content) < newcardinality {
		tmp := make([]uint16, newcardinality, newcardinality)
		copy(tmp[:indexstart], ac.Content[:indexstart])
		copy(tmp[indexstart+rangelength:], ac.Content[indexend:])

		ac.Content = tmp
	} else {
		ac.Content = ac.Content[:newcardinality]
		copy(ac.Content[indexstart+rangelength:], ac.Content[indexend:])

	}
	for k := 0; k < rangelength; k++ {
		ac.Content[k+indexstart] = uint16(firstOfRange + k)
	}
	return ac
}

// remove the values in the range [firstOfRange,lastOfRange)
func (ac *arrayContainer) iremoveRange(firstOfRange, lastOfRange int) container {
	if firstOfRange >= lastOfRange {
		return ac
	}
	indexstart := binarySearch(ac.Content, uint16(firstOfRange))
	if indexstart < 0 {
		indexstart = -indexstart - 1
	}
	indexend := binarySearch(ac.Content, uint16(lastOfRange-1))
	if indexend < 0 {
		indexend = -indexend - 1
	} else {
		indexend++
	}
	rangelength := indexend - indexstart
	answer := ac
	copy(answer.Content[indexstart:], ac.Content[indexstart+rangelength:])
	answer.Content = answer.Content[:ac.getCardinality()-rangelength]
	return answer
}

// flip the values in the range [firstOfRange,lastOfRange)
func (ac *arrayContainer) not(firstOfRange, lastOfRange int) container {
	if firstOfRange >= lastOfRange {
		return ac.clone()
	}
	return ac.notClose(firstOfRange, lastOfRange-1) // remove everything in [firstOfRange,lastOfRange-1]
}

// flip the values in the range [firstOfRange,lastOfRange]
func (ac *arrayContainer) notClose(firstOfRange, lastOfRange int) container {
	if firstOfRange > lastOfRange { // unlike add and remove, not uses an inclusive range [firstOfRange,lastOfRange]
		return ac.clone()
	}

	// determine the span of array indices to be affected^M
	startIndex := binarySearch(ac.Content, uint16(firstOfRange))
	if startIndex < 0 {
		startIndex = -startIndex - 1
	}
	lastIndex := binarySearch(ac.Content, uint16(lastOfRange))
	if lastIndex < 0 {
		lastIndex = -lastIndex - 2
	}
	currentValuesInRange := lastIndex - startIndex + 1
	spanToBeFlipped := lastOfRange - firstOfRange + 1
	newValuesInRange := spanToBeFlipped - currentValuesInRange
	cardinalityChange := newValuesInRange - currentValuesInRange
	newCardinality := len(ac.Content) + cardinalityChange

	if newCardinality > arrayDefaultMaxSize {
		return ac.toBitmapContainer().not(firstOfRange, lastOfRange+1)
	}
	answer := newArrayContainer()
	answer.Content = make([]uint16, newCardinality, newCardinality) //a hack for sure

	copy(answer.Content, ac.Content[:startIndex])
	outPos := startIndex
	inPos := startIndex
	valInRange := firstOfRange
	for ; valInRange <= lastOfRange && inPos <= lastIndex; valInRange++ {
		if uint16(valInRange) != ac.Content[inPos] {
			answer.Content[outPos] = uint16(valInRange)
			outPos++
		} else {
			inPos++
		}
	}

	for ; valInRange <= lastOfRange; valInRange++ {
		answer.Content[outPos] = uint16(valInRange)
		outPos++
	}

	for i := lastIndex + 1; i < len(ac.Content); i++ {
		answer.Content[outPos] = ac.Content[i]
		outPos++
	}
	answer.Content = answer.Content[:newCardinality]
	return answer

}

func (ac *arrayContainer) equals(o interface{}) bool {

	srb, ok := o.(*arrayContainer)
	if ok {
		// Check if the containers are the same object.
		if ac == srb {
			return true
		}

		if len(srb.Content) != len(ac.Content) {
			return false
		}

		for i, v := range ac.Content {
			if v != srb.Content[i] {
				return false
			}
		}
		return true
	}

	bc, ok := o.(container)
	if ok {
		// use generic comparison
		if bc.getCardinality() != ac.getCardinality() {
			return false
		}
		ait := ac.getShortIterator()
		bit := bc.getShortIterator()

		for ait.hasNext() {
			if bit.next() != ait.next() {
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
func (ac *arrayContainer) iadd(x uint16) (wasNew bool) {
	// Special case adding to the end of the container.
	l := len(ac.Content)
	if l > 0 && l < arrayDefaultMaxSize && ac.Content[l-1] < x {
		ac.Content = append(ac.Content, x)
		return true
	}

	loc := binarySearch(ac.Content, x)

	if loc < 0 {
		s := ac.Content
		i := -loc - 1
		s = append(s, 0)
		copy(s[i+1:], s[i:])
		s[i] = x
		ac.Content = s
		return true
	}
	return false
}

func (ac *arrayContainer) iaddReturnMinimized(x uint16) container {
	// Special case adding to the end of the container.
	l := len(ac.Content)
	if l > 0 && l < arrayDefaultMaxSize && ac.Content[l-1] < x {
		ac.Content = append(ac.Content, x)
		return ac
	}

	loc := binarySearch(ac.Content, x)

	if loc < 0 {
		if len(ac.Content) >= arrayDefaultMaxSize {
			a := ac.toBitmapContainer()
			a.iadd(x)
			return a
		}
		s := ac.Content
		i := -loc - 1
		s = append(s, 0)
		copy(s[i+1:], s[i:])
		s[i] = x
		ac.Content = s
	}
	return ac
}

// iremoveReturnMinimized is allowed to change the return type to minimize storage.
func (ac *arrayContainer) iremoveReturnMinimized(x uint16) container {
	ac.iremove(x)
	return ac
}

func (ac *arrayContainer) iremove(x uint16) bool {
	loc := binarySearch(ac.Content, x)
	if loc >= 0 {
		s := ac.Content
		s = append(s[:loc], s[loc+1:]...)
		ac.Content = s
		return true
	}
	return false
}

func (ac *arrayContainer) remove(x uint16) container {
	out := &arrayContainer{make([]uint16, len(ac.Content))}
	copy(out.Content, ac.Content[:])

	loc := binarySearch(out.Content, x)
	if loc >= 0 {
		s := out.Content
		s = append(s[:loc], s[loc+1:]...)
		out.Content = s
	}
	return out
}

func (ac *arrayContainer) or(a container) container {
	switch a.(type) {
	case *arrayContainer:
		return ac.orArray(a.(*arrayContainer))
	case *bitmapContainer:
		return a.or(ac)
	}
	panic("unsupported container type")
}

func (ac *arrayContainer) ior(a container) container {
	switch a.(type) {
	case *arrayContainer:
		return ac.orArray(a.(*arrayContainer))
	case *bitmapContainer:
		return a.ior(ac)
	}
	panic("unsupported container type")
}

func (ac *arrayContainer) lazyIOR(a container) container {
	switch a.(type) {
	case *arrayContainer:
		return ac.lazyorArray(a.(*arrayContainer))
	case *bitmapContainer:
		return a.lazyOR(ac)
	}
	panic("unsupported container type")
}

func (ac *arrayContainer) lazyOR(a container) container {
	switch a.(type) {
	case *arrayContainer:
		return ac.lazyorArray(a.(*arrayContainer))
	case *bitmapContainer:
		return a.lazyOR(ac)
	}
	panic("unsupported container type")
}

func (ac *arrayContainer) orArray(value2 *arrayContainer) container {
	value1 := ac
	maxPossibleCardinality := value1.getCardinality() + value2.getCardinality()
	if maxPossibleCardinality > arrayDefaultMaxSize { // it could be a bitmap!^M
		bc := newBitmapContainer()
		for k := 0; k < len(value2.Content); k++ {
			v := value2.Content[k]
			i := uint(v) >> 6
			mask := uint64(1) << (v % 64)
			bc.Bitmap[i] |= mask
		}
		for k := 0; k < len(ac.Content); k++ {
			v := ac.Content[k]
			i := uint(v) >> 6
			mask := uint64(1) << (v % 64)
			bc.Bitmap[i] |= mask
		}
		bc.Cardinality = int(popcntSlice(bc.Bitmap))
		if bc.Cardinality <= arrayDefaultMaxSize {
			return bc.toArrayContainer()
		}
		return bc
	}
	answer := newArrayContainerCapacity(maxPossibleCardinality)
	nl := union2by2(value1.Content, value2.Content, answer.Content)
	answer.Content = answer.Content[:nl] // reslice to match actual used capacity
	return answer
}

func (ac *arrayContainer) lazyorArray(value2 *arrayContainer) container {
	value1 := ac
	maxPossibleCardinality := value1.getCardinality() + value2.getCardinality()
	if maxPossibleCardinality > arrayLazyLowerBound { // it could be a bitmap!^M
		bc := newBitmapContainer()
		for k := 0; k < len(value2.Content); k++ {
			v := value2.Content[k]
			i := uint(v) >> 6
			mask := uint64(1) << (v % 64)
			bc.Bitmap[i] |= mask
		}
		for k := 0; k < len(ac.Content); k++ {
			v := ac.Content[k]
			i := uint(v) >> 6
			mask := uint64(1) << (v % 64)
			bc.Bitmap[i] |= mask
		}
		bc.Cardinality = invalidCardinality
		return bc
	}
	answer := newArrayContainerCapacity(maxPossibleCardinality)
	nl := union2by2(value1.Content, value2.Content, answer.Content)
	answer.Content = answer.Content[:nl] // reslice to match actual used capacity
	return answer
}

func (ac *arrayContainer) and(a container) container {
	switch x := a.(type) {
	case *arrayContainer:
		return ac.andArray(x)
	case *bitmapContainer:
		return x.and(ac)
	case *runContainer16:
		return x.andArray(ac)
	}
	panic("unsupported container type")
}

func (ac *arrayContainer) intersects(a container) bool {
	switch x := a.(type) {
	case *arrayContainer:
		return ac.intersectsArray(x)
	case *bitmapContainer:
		return x.intersects(ac)
	case *runContainer16:
		return x.intersects(ac)
	}
	panic("unsupported container type")
}

func (ac *arrayContainer) iand(a container) container {
	switch a.(type) {
	case *arrayContainer:
		return ac.iandArray(a.(*arrayContainer))
	case *bitmapContainer:
		return ac.iandBitmap(a.(*bitmapContainer))
	}
	panic("unsupported container type")
}

func (ac *arrayContainer) iandBitmap(bc *bitmapContainer) *arrayContainer {
	pos := 0
	c := ac.getCardinality()
	for k := 0; k < c; k++ {
		if bc.contains(ac.Content[k]) {
			ac.Content[pos] = ac.Content[k]
			pos++
		}
	}
	ac.Content = ac.Content[:pos]
	return ac

}

func (ac *arrayContainer) xor(a container) container {
	switch a.(type) {
	case *arrayContainer:
		return ac.xorArray(a.(*arrayContainer))
	case *bitmapContainer:
		return a.xor(ac)
	}
	panic("unsupported container type")
}

func (ac *arrayContainer) xorArray(value2 *arrayContainer) container {
	value1 := ac
	totalCardinality := value1.getCardinality() + value2.getCardinality()
	if totalCardinality > arrayDefaultMaxSize { // it could be a bitmap!
		bc := newBitmapContainer()
		for k := 0; k < len(value2.Content); k++ {
			v := value2.Content[k]
			i := uint(v) >> 6
			bc.Bitmap[i] ^= (uint64(1) << (v % 64))
		}
		for k := 0; k < len(ac.Content); k++ {
			v := ac.Content[k]
			i := uint(v) >> 6
			bc.Bitmap[i] ^= (uint64(1) << (v % 64))
		}
		bc.computeCardinality()
		if bc.Cardinality <= arrayDefaultMaxSize {
			return bc.toArrayContainer()
		}
		return bc
	}
	desiredCapacity := totalCardinality
	answer := newArrayContainerCapacity(desiredCapacity)
	length := exclusiveUnion2by2(value1.Content, value2.Content, answer.Content)
	answer.Content = answer.Content[:length]
	return answer

}

func (ac *arrayContainer) andNot(a container) container {
	switch a.(type) {
	case *arrayContainer:
		return ac.andNotArray(a.(*arrayContainer))
	case *bitmapContainer:
		return ac.andNotBitmap(a.(*bitmapContainer))
	}
	panic("unsupported container type")
}

func (ac *arrayContainer) iandNot(a container) container {
	switch a.(type) {
	case *arrayContainer:
		return ac.iandNotArray(a.(*arrayContainer))
	case *bitmapContainer:
		return ac.iandNotBitmap(a.(*bitmapContainer))
	}
	panic("unsupported container type")
}

func (ac *arrayContainer) andNotArray(value2 *arrayContainer) container {
	value1 := ac
	desiredcapacity := value1.getCardinality()
	answer := newArrayContainerCapacity(desiredcapacity)
	length := difference(value1.Content, value2.Content, answer.Content)
	answer.Content = answer.Content[:length]
	return answer
}

func (ac *arrayContainer) iandNotArray(value2 *arrayContainer) container {
	length := difference(ac.Content, value2.Content, ac.Content)
	ac.Content = ac.Content[:length]
	return ac
}

func (ac *arrayContainer) andNotBitmap(value2 *bitmapContainer) container {
	desiredcapacity := ac.getCardinality()
	answer := newArrayContainerCapacity(desiredcapacity)
	answer.Content = answer.Content[:desiredcapacity]
	pos := 0
	for _, v := range ac.Content {
		if !value2.contains(v) {
			answer.Content[pos] = v
			pos++
		}
	}
	answer.Content = answer.Content[:pos]
	return answer
}

func (ac *arrayContainer) andBitmap(value2 *bitmapContainer) container {
	desiredcapacity := ac.getCardinality()
	answer := newArrayContainerCapacity(desiredcapacity)
	answer.Content = answer.Content[:desiredcapacity]
	pos := 0
	for _, v := range ac.Content {
		if value2.contains(v) {
			answer.Content[pos] = v
			pos++
		}
	}
	answer.Content = answer.Content[:pos]
	return answer
}

func (ac *arrayContainer) iandNotBitmap(value2 *bitmapContainer) container {
	pos := 0
	for _, v := range ac.Content {
		if !value2.contains(v) {
			ac.Content[pos] = v
			pos++
		}
	}
	ac.Content = ac.Content[:pos]
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

// flip the values in the range [firstOfRange,lastOfRange)
func (ac *arrayContainer) inot(firstOfRange, lastOfRange int) container {
	if firstOfRange >= lastOfRange {
		return ac
	}
	return ac.inotClose(firstOfRange, lastOfRange-1) // remove everything in [firstOfRange,lastOfRange-1]
}

// flip the values in the range [firstOfRange,lastOfRange]
func (ac *arrayContainer) inotClose(firstOfRange, lastOfRange int) container {
	if firstOfRange > lastOfRange { // unlike add and remove, not uses an inclusive range [firstOfRange,lastOfRange]
		return ac
	}
	// determine the span of array indices to be affected
	startIndex := binarySearch(ac.Content, uint16(firstOfRange))
	if startIndex < 0 {
		startIndex = -startIndex - 1
	}
	lastIndex := binarySearch(ac.Content, uint16(lastOfRange))
	if lastIndex < 0 {
		lastIndex = -lastIndex - 1 - 1
	}
	currentValuesInRange := lastIndex - startIndex + 1
	spanToBeFlipped := lastOfRange - firstOfRange + 1

	newValuesInRange := spanToBeFlipped - currentValuesInRange
	buffer := make([]uint16, newValuesInRange)
	cardinalityChange := newValuesInRange - currentValuesInRange
	newCardinality := len(ac.Content) + cardinalityChange
	if cardinalityChange > 0 {
		if newCardinality > len(ac.Content) {
			if newCardinality > arrayDefaultMaxSize {
				return ac.toBitmapContainer().inot(firstOfRange, lastOfRange+1)
			}
			ac.Content = copyOf(ac.Content, newCardinality)
		}
		base := lastIndex + 1
		copy(ac.Content[lastIndex+1+cardinalityChange:], ac.Content[base:base+len(ac.Content)-1-lastIndex])
		ac.negateRange(buffer, startIndex, lastIndex, firstOfRange, lastOfRange+1)
	} else { // no expansion needed
		ac.negateRange(buffer, startIndex, lastIndex, firstOfRange, lastOfRange+1)
		if cardinalityChange < 0 {

			for i := startIndex + newValuesInRange; i < newCardinality; i++ {
				ac.Content[i] = ac.Content[i-cardinalityChange]
			}
		}
	}
	ac.Content = ac.Content[:newCardinality]
	return ac
}

func (ac *arrayContainer) negateRange(buffer []uint16, startIndex, lastIndex, startRange, lastRange int) {
	// compute the negation into buffer
	outPos := 0
	inPos := startIndex // value here always >= valInRange,
	// until it is exhausted
	// n.b., we can start initially exhausted.

	valInRange := startRange
	for ; valInRange < lastRange && inPos <= lastIndex; valInRange++ {
		if uint16(valInRange) != ac.Content[inPos] {
			buffer[outPos] = uint16(valInRange)
			outPos++
		} else {
			inPos++
		}
	}

	// if there are extra items (greater than the biggest
	// pre-existing one in range), buffer them
	for ; valInRange < lastRange; valInRange++ {
		buffer[outPos] = uint16(valInRange)
		outPos++
	}

	if outPos != len(buffer) {
		panic("negateRange: internal bug")
	}

	for i, item := range buffer {
		ac.Content[i+startIndex] = item
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
		ac.Content,
		value2.Content,
		answer.Content)
	answer.Content = answer.Content[:length]
	return answer
}

func (ac *arrayContainer) intersectsArray(value2 *arrayContainer) bool {
	return intersects2by2(
		ac.Content,
		value2.Content)
}

func (ac *arrayContainer) iandArray(value2 *arrayContainer) *arrayContainer {
	length := intersection2by2(
		ac.Content,
		value2.Content,
		ac.Content)
	ac.Content = ac.Content[:length]
	return ac
}

func (ac *arrayContainer) getCardinality() int {
	return len(ac.Content)
}

func (ac *arrayContainer) rank(x uint16) int {
	answer := binarySearch(ac.Content, x)
	if answer >= 0 {
		return answer + 1
	}
	return -answer - 1

}

func (ac *arrayContainer) selectInt(x uint16) int {
	return int(ac.Content[x])
}

func (ac *arrayContainer) clone() container {
	ptr := arrayContainer{make([]uint16, len(ac.Content))}
	copy(ptr.Content, ac.Content[:])
	return &ptr
}

func (ac *arrayContainer) contains(x uint16) bool {
	return binarySearch(ac.Content, x) >= 0
}

func (ac *arrayContainer) loadData(bitmapContainer *bitmapContainer) {
	ac.Content = make([]uint16, bitmapContainer.Cardinality, bitmapContainer.Cardinality)
	bitmapContainer.fillArray(ac.Content)
}
func newArrayContainer() *arrayContainer {
	p := new(arrayContainer)
	return p
}

func newArrayContainerCapacity(size int) *arrayContainer {
	p := new(arrayContainer)
	p.Content = make([]uint16, 0, size)
	return p
}

func newArrayContainerSize(size int) *arrayContainer {
	p := new(arrayContainer)
	p.Content = make([]uint16, size, size)
	return p
}

func newArrayContainerRange(firstOfRun, lastOfRun int) *arrayContainer {
	valuesInRange := lastOfRun - firstOfRun + 1
	this := newArrayContainerCapacity(valuesInRange)
	for i := 0; i < valuesInRange; i++ {
		this.Content = append(this.Content, uint16(firstOfRun+i))
	}
	return this
}

func (ac *arrayContainer) numberOfRuns() (nr int) {
	n := len(ac.Content)
	var runlen uint16
	var cur, prev uint16

	switch n {
	case 0:
		return 0
	case 1:
		return 1
	default:
		for i := 1; i < n; i++ {
			prev = ac.Content[i-1]
			cur = ac.Content[i]

			if cur == prev+1 {
				runlen++
			} else {
				if cur < prev {
					panic("then fundamental arrayContainer assumption of sorted ac.Content was broken")
				}
				if cur == prev {
					panic("then fundamental arrayContainer assumption of deduplicated content was broken")
				} else {
					nr++
					runlen = 0
				}
			}
		}
		nr++
	}
	return
}

// convert to run or array *if needed*
func (ac *arrayContainer) toEfficientContainer() container {

	numRuns := ac.numberOfRuns()

	sizeAsRunContainer := runContainer16SerializedSizeInBytes(numRuns)
	sizeAsBitmapContainer := bitmapContainerSizeInBytes()
	card := int(ac.getCardinality())
	sizeAsArrayContainer := arrayContainerSizeInBytes(card)

	if sizeAsRunContainer <= min(sizeAsBitmapContainer, sizeAsArrayContainer) {
		return newRunContainer16FromArray(ac)
	}
	if card <= arrayDefaultMaxSize {
		return ac
	}
	return ac.toBitmapContainer()
}

func (bc *arrayContainer) containerType() contype {
	return arrayContype
}
