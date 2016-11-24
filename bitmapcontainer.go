package roaring

import (
	"fmt"
	"unsafe"
)

//go:generate msgp -unexported

type bitmapContainer struct {
	Cardinality int
	Bitmap      []uint64
}

func (c bitmapContainer) String() string {
	var s string
	for it := c.getShortIterator(); it.hasNext(); {
		s += fmt.Sprintf("%v, ", it.next())
	}
	return s
}

func newBitmapContainer() *bitmapContainer {
	p := new(bitmapContainer)
	size := (1 << 16) / 64
	p.Bitmap = make([]uint64, size, size)
	return p
}

func newBitmapContainerwithRange(firstOfRun, lastOfRun int) *bitmapContainer {
	this := newBitmapContainer()
	this.Cardinality = lastOfRun - firstOfRun + 1
	if this.Cardinality == maxCapacity {
		fill(this.Bitmap, uint64(0xffffffffffffffff))
	} else {
		firstWord := firstOfRun / 64
		lastWord := lastOfRun / 64
		zeroPrefixLength := uint64(firstOfRun & 63)
		zeroSuffixLength := uint64(63 - (lastOfRun & 63))

		fillRange(this.Bitmap, firstWord, lastWord+1, uint64(0xffffffffffffffff))
		this.Bitmap[firstWord] ^= ((uint64(1) << zeroPrefixLength) - 1)
		blockOfOnes := (uint64(1) << zeroSuffixLength) - 1
		maskOnLeft := blockOfOnes << (uint64(64) - zeroSuffixLength)
		this.Bitmap[lastWord] ^= maskOnLeft
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
	return len(bc.Bitmap) * 8 // + bcBaseBytes
}

func (bc *bitmapContainer) serializedSizeInBytes() int {
	return bc.Msgsize()
	//return len(bc.Bitmap) * 8 // +  bcBaseBytes
}

const bcBaseBytes = int(unsafe.Sizeof(bitmapContainer{}))

// bitmapContainer doesn't depend on card, always fully allocated
func bitmapContainerSizeInBytes() int {
	return bcBaseBytes + (1<<16)/8
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

func (bc *bitmapContainer) fillLeastSignificant16bits(x []uint32, i int, mask uint32) {
	// TODO: should be written as optimized assembly
	pos := i
	base := mask
	for k := 0; k < len(bc.Bitmap); k++ {
		bitset := bc.Bitmap[k]
		for bitset != 0 {
			t := bitset & -bitset
			x[pos] = base + uint32(popcount(t-1))
			pos++
			bitset ^= t
		}
		base += 64
	}
}

func (bc *bitmapContainer) equals(o interface{}) bool {
	srb, ok := o.(*bitmapContainer)
	if ok {
		if srb.Cardinality != bc.Cardinality {
			return false
		}
		return bitmapEquals(bc.Bitmap, srb.Bitmap)
	}

	ac, ok := o.(container)
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

func (bc *bitmapContainer) iaddReturnMinimized(i uint16) container {
	bc.iadd(i)
	if bc.isFull() {
		return newRunContainer16Range(0, MaxUint16)
	}
	return bc
}

func (bc *bitmapContainer) iadd(i uint16) bool {
	x := int(i)
	previous := bc.Bitmap[x/64]
	mask := uint64(1) << (uint(x) % 64)
	newb := previous | mask
	bc.Bitmap[x/64] = newb
	bc.Cardinality += int(uint64(previous^newb) >> (uint(x) % 64))
	return newb != previous
}

func (bc *bitmapContainer) iremoveReturnMinimized(i uint16) container {
	if bc.iremove(i) {
		if bc.Cardinality == arrayDefaultMaxSize {
			return bc.toArrayContainer()
		}
	}
	return bc
}

// iremove returns true if i was found.
func (bc *bitmapContainer) iremove(i uint16) bool {
	if bc.contains(i) {
		bc.Cardinality--
		bc.Bitmap[i/64] &^= (uint64(1) << (i % 64))
		return true
	}
	return false
}

func (bc *bitmapContainer) isFull() bool {
	return bc.Cardinality == int(MaxUint16)+1
}

func (bc *bitmapContainer) getCardinality() int {
	return bc.Cardinality
}

func (bc *bitmapContainer) clone() container {
	ptr := bitmapContainer{bc.Cardinality, make([]uint64, len(bc.Bitmap))}
	copy(ptr.Bitmap, bc.Bitmap[:])
	return &ptr
}

// add all values in range [firstOfRange,lastOfRange)
func (bc *bitmapContainer) iaddRange(firstOfRange, lastOfRange int) container {
	bc.Cardinality += setBitmapRangeAndCardinalityChange(bc.Bitmap, firstOfRange, lastOfRange)
	return bc
}

// add all values in range [firstOfRange,lastOfRange)
// unused code
/*func (bc *bitmapContainer) addRange(firstOfRange, lastOfRange int) container {
	answer := &bitmapContainer{bc.Cardinality, make([]uint64, len(bc.Bitmap))}
	copy(answer.Bitmap, bc.Bitmap[:])
	answer.Cardinality += setBitmapRangeAndCardinalityChange(answer.Bitmap, firstOfRange, lastOfRange)
	return answer
}*/

// remove all values in range [firstOfRange,lastOfRange)
// unused code
/*func (bc *bitmapContainer) removeRange(firstOfRange, lastOfRange int) container {
	answer := &bitmapContainer{bc.Cardinality, make([]uint64, len(bc.Bitmap))}
	copy(answer.Bitmap, bc.Bitmap[:])
	answer.Cardinality += resetBitmapRangeAndCardinalityChange(answer.Bitmap, firstOfRange, lastOfRange)
	if answer.getCardinality() <= arrayDefaultMaxSize {
		return answer.toArrayContainer()
	}
	return answer
}*/

// remove all values in range [firstOfRange,lastOfRange)
func (bc *bitmapContainer) iremoveRange(firstOfRange, lastOfRange int) container {
	bc.Cardinality += resetBitmapRangeAndCardinalityChange(bc.Bitmap, firstOfRange, lastOfRange)
	if bc.getCardinality() <= arrayDefaultMaxSize {
		return bc.toArrayContainer()
	}
	return bc
}

// flip all values in range [firstOfRange,lastOfRange)
func (bc *bitmapContainer) inot(firstOfRange, lastOfRange int) container {
	if lastOfRange-firstOfRange == maxCapacity {
		flipBitmapRange(bc.Bitmap, firstOfRange, lastOfRange)
		bc.Cardinality = maxCapacity - bc.Cardinality
	} else if lastOfRange-firstOfRange > maxCapacity/2 {
		flipBitmapRange(bc.Bitmap, firstOfRange, lastOfRange)
		bc.computeCardinality()
	} else {
		bc.Cardinality += flipBitmapRangeAndCardinalityChange(bc.Bitmap, firstOfRange, lastOfRange)
	}
	if bc.getCardinality() <= arrayDefaultMaxSize {
		return bc.toArrayContainer()
	}
	return bc
}

// flip all values in range [firstOfRange,lastOfRange)
func (bc *bitmapContainer) not(firstOfRange, lastOfRange int) container {
	answer := bc.clone()
	return answer.inot(firstOfRange, lastOfRange)
}

func (bc *bitmapContainer) or(a container) container {
	switch a.(type) {
	case *arrayContainer:
		return bc.orArray(a.(*arrayContainer))
	case *bitmapContainer:
		return bc.orBitmap(a.(*bitmapContainer))
	}
	panic("unsupported container type")
}

func (bc *bitmapContainer) ior(a container) container {
	switch a.(type) {
	case *arrayContainer:
		return bc.iorArray(a.(*arrayContainer))
	case *bitmapContainer:
		return bc.iorBitmap(a.(*bitmapContainer))
	}
	panic("unsupported container type")
}

func (bc *bitmapContainer) lazyIOR(a container) container {
	switch a.(type) {
	case *arrayContainer:
		return bc.lazyIORArray(a.(*arrayContainer))
	case *bitmapContainer:
		return bc.lazyIORBitmap(a.(*bitmapContainer))
	}
	panic("unsupported container type")
}

func (bc *bitmapContainer) lazyOR(a container) container {
	switch a.(type) {
	case *arrayContainer:
		return bc.lazyORArray(a.(*arrayContainer))
	case *bitmapContainer:
		return bc.lazyORBitmap(a.(*bitmapContainer))
	}
	panic("unsupported container type")
}

func (bc *bitmapContainer) orArray(value2 *arrayContainer) container {
	answer := bc.clone().(*bitmapContainer)
	c := value2.getCardinality()
	for k := 0; k < c; k++ {
		v := value2.Content[k]
		i := uint(v) >> 6
		bef := answer.Bitmap[i]
		aft := bef | (uint64(1) << (v % 64))
		answer.Bitmap[i] = aft
		answer.Cardinality += int((bef - aft) >> 63)
	}
	return answer
}

func (bc *bitmapContainer) orBitmap(value2 *bitmapContainer) container {
	answer := newBitmapContainer()
	for k := 0; k < len(answer.Bitmap); k++ {
		answer.Bitmap[k] = bc.Bitmap[k] | value2.Bitmap[k]
	}
	answer.computeCardinality()
	return answer
}

func (bc *bitmapContainer) computeCardinality() {
	bc.Cardinality = int(popcntSlice(bc.Bitmap))
}

func (bc *bitmapContainer) iorArray(value2 *arrayContainer) container {
	answer := bc
	c := value2.getCardinality()
	for k := 0; k < c; k++ {
		vc := value2.Content[k]
		i := uint(vc) >> 6
		bef := answer.Bitmap[i]
		aft := bef | (uint64(1) << (vc % 64))
		answer.Bitmap[i] = aft
		answer.Cardinality += int((bef - aft) >> 63)
	}
	return answer
}

func (bc *bitmapContainer) iorBitmap(value2 *bitmapContainer) container {
	answer := bc
	answer.Cardinality = 0
	for k := 0; k < len(answer.Bitmap); k++ {
		answer.Bitmap[k] = bc.Bitmap[k] | value2.Bitmap[k]
	}
	answer.computeCardinality()
	return answer
}

func (bc *bitmapContainer) lazyIORArray(value2 *arrayContainer) container {
	answer := bc
	c := value2.getCardinality()
	for k := 0; k < c; k++ {
		vc := value2.Content[k]
		i := uint(vc) >> 6
		answer.Bitmap[i] = answer.Bitmap[i] | (uint64(1) << (vc % 64))
	}
	answer.Cardinality = invalidCardinality
	return answer
}

func (bc *bitmapContainer) lazyORArray(value2 *arrayContainer) container {
	answer := bc.clone().(*bitmapContainer)
	return answer.lazyIORArray(value2)
}

func (bc *bitmapContainer) lazyIORBitmap(value2 *bitmapContainer) container {
	answer := bc
	for k := 0; k < len(answer.Bitmap); k++ {
		answer.Bitmap[k] = bc.Bitmap[k] | value2.Bitmap[k]
	}
	bc.Cardinality = invalidCardinality
	return answer
}

func (bc *bitmapContainer) lazyORBitmap(value2 *bitmapContainer) container {
	answer := bc.clone().(*bitmapContainer)
	return answer.lazyIORBitmap(value2)
}

func (bc *bitmapContainer) xor(a container) container {
	switch a.(type) {
	case *arrayContainer:
		return bc.xorArray(a.(*arrayContainer))
	case *bitmapContainer:
		return bc.xorBitmap(a.(*bitmapContainer))
	}
	panic("unsupported container type")
}

func (bc *bitmapContainer) xorArray(value2 *arrayContainer) container {
	answer := bc.clone().(*bitmapContainer)
	c := value2.getCardinality()
	for k := 0; k < c; k++ {
		vc := value2.Content[k]
		index := uint(vc) >> 6
		abi := answer.Bitmap[index]
		mask := uint64(1) << (vc % 64)
		answer.Cardinality += 1 - 2*int((abi&mask)>>(vc%64))
		answer.Bitmap[index] = abi ^ mask
	}
	if answer.Cardinality <= arrayDefaultMaxSize {
		return answer.toArrayContainer()
	}
	return answer
}

func (bc *bitmapContainer) rank(x uint16) int {
	// TODO: rewrite in assembly
	leftover := (uint(x) + 1) & 63
	if leftover == 0 {
		return int(popcntSlice(bc.Bitmap[:(uint(x)+1)/64]))
	}
	return int(popcntSlice(bc.Bitmap[:(uint(x)+1)/64]) + popcount(bc.Bitmap[(uint(x)+1)/64]<<(64-leftover)))
}

func (bc *bitmapContainer) selectInt(x uint16) int {
	remaining := x
	for k := 0; k < len(bc.Bitmap); k++ {
		w := popcount(bc.Bitmap[k])
		if uint16(w) > remaining {
			return int(k*64 + selectBitPosition(bc.Bitmap[k], int(remaining)))
		}
		remaining -= uint16(w)
	}
	return -1
}

func (bc *bitmapContainer) xorBitmap(value2 *bitmapContainer) container {
	newCardinality := int(popcntXorSlice(bc.Bitmap, value2.Bitmap))

	if newCardinality > arrayDefaultMaxSize {
		answer := newBitmapContainer()
		for k := 0; k < len(answer.Bitmap); k++ {
			answer.Bitmap[k] = bc.Bitmap[k] ^ value2.Bitmap[k]
		}
		answer.Cardinality = newCardinality
		return answer
	}
	ac := newArrayContainerSize(newCardinality)
	fillArrayXOR(ac.Content, bc.Bitmap, value2.Bitmap)
	ac.Content = ac.Content[:newCardinality]
	return ac
}

func (bc *bitmapContainer) and(a container) container {
	switch x := a.(type) {
	case *arrayContainer:
		return bc.andArray(x)
	case *bitmapContainer:
		return bc.andBitmap(x)
	case *runContainer16:
		return x.andBitmapContainer(bc)
	}
	panic("unsupported container type")
}

func (bc *bitmapContainer) intersects(a container) bool {
	switch x := a.(type) {
	case *arrayContainer:
		return bc.intersectsArray(x)
	case *bitmapContainer:
		return bc.intersectsBitmap(x)
	case *runContainer16:
		return x.intersects(bc)

	}
	panic("unsupported container type")
}

func (bc *bitmapContainer) iand(a container) container {
	switch a.(type) {
	case *arrayContainer:
		return bc.andArray(a.(*arrayContainer))
	case *bitmapContainer:
		return bc.iandBitmap(a.(*bitmapContainer))
	}
	panic("unsupported container type")
}

func (bc *bitmapContainer) andArray(value2 *arrayContainer) *arrayContainer {
	answer := newArrayContainerCapacity(len(value2.Content))
	c := value2.getCardinality()
	for k := 0; k < c; k++ {
		v := value2.Content[k]
		if bc.contains(v) {
			answer.Content = append(answer.Content, v)
		}
	}
	return answer
}

func (bc *bitmapContainer) andBitmap(value2 *bitmapContainer) container {
	newcardinality := int(popcntAndSlice(bc.Bitmap, value2.Bitmap))
	if newcardinality > arrayDefaultMaxSize {
		answer := newBitmapContainer()
		for k := 0; k < len(answer.Bitmap); k++ {
			answer.Bitmap[k] = bc.Bitmap[k] & value2.Bitmap[k]
		}
		answer.Cardinality = newcardinality
		return answer
	}
	ac := newArrayContainerSize(newcardinality)
	fillArrayAND(ac.Content, bc.Bitmap, value2.Bitmap)
	ac.Content = ac.Content[:newcardinality] //not sure why i need this
	return ac

}

func (bc *bitmapContainer) intersectsArray(value2 *arrayContainer) bool {
	c := value2.getCardinality()
	for k := 0; k < c; k++ {
		v := value2.Content[k]
		if bc.contains(v) {
			return true
		}
	}
	return false
}

func (bc *bitmapContainer) intersectsBitmap(value2 *bitmapContainer) bool {
	for k := 0; k < len(bc.Bitmap); k++ {
		if (bc.Bitmap[k] & value2.Bitmap[k]) != 0 {
			return true
		}
	}
	return false

}

func (bc *bitmapContainer) iandBitmap(value2 *bitmapContainer) container {
	newcardinality := int(popcntAndSlice(bc.Bitmap, value2.Bitmap))
	if newcardinality > arrayDefaultMaxSize {
		for k := 0; k < len(bc.Bitmap); k++ {
			bc.Bitmap[k] = bc.Bitmap[k] & value2.Bitmap[k]
		}
		bc.Cardinality = newcardinality
		return bc
	}
	ac := newArrayContainerSize(newcardinality)
	fillArrayAND(ac.Content, bc.Bitmap, value2.Bitmap)
	ac.Content = ac.Content[:newcardinality] //not sure why i need this
	return ac

}

func (bc *bitmapContainer) andNot(a container) container {
	switch a.(type) {
	case *arrayContainer:
		return bc.andNotArray(a.(*arrayContainer))
	case *bitmapContainer:
		return bc.andNotBitmap(a.(*bitmapContainer))
	}
	panic("unsupported container type")
}

func (bc *bitmapContainer) iandNot(a container) container {
	switch a.(type) {
	case *arrayContainer:
		// FIXME: this is not iandNotArray, so it won't
		// have the side-effect specified by the inplace 'i' prefix.
		return bc.andNotArray(a.(*arrayContainer))
	case *bitmapContainer:
		return bc.iandNotBitmap(a.(*bitmapContainer))
	}
	panic("unsupported container type")
}

func (bc *bitmapContainer) andNotArray(value2 *arrayContainer) container {
	answer := bc.clone().(*bitmapContainer)
	c := value2.getCardinality()
	for k := 0; k < c; k++ {
		vc := value2.Content[k]
		i := uint(vc) >> 6
		oldv := answer.Bitmap[i]
		newv := oldv &^ (uint64(1) << (vc % 64))
		answer.Bitmap[i] = newv
		answer.Cardinality -= int(uint64(oldv^newv) >> (vc % 64))
	}
	if answer.Cardinality <= arrayDefaultMaxSize {
		return answer.toArrayContainer()
	}
	return answer
}

func (bc *bitmapContainer) andNotBitmap(value2 *bitmapContainer) container {
	newCardinality := int(popcntMaskSlice(bc.Bitmap, value2.Bitmap))
	if newCardinality > arrayDefaultMaxSize {
		answer := newBitmapContainer()
		for k := 0; k < len(answer.Bitmap); k++ {
			answer.Bitmap[k] = bc.Bitmap[k] &^ value2.Bitmap[k]
		}
		answer.Cardinality = newCardinality
		return answer
	}
	ac := newArrayContainerSize(newCardinality)
	fillArrayANDNOT(ac.Content, bc.Bitmap, value2.Bitmap)
	return ac
}

func (bc *bitmapContainer) iandNotBitmapSurely(value2 *bitmapContainer) *bitmapContainer {
	newCardinality := int(popcntMaskSlice(bc.Bitmap, value2.Bitmap))
	for k := 0; k < len(bc.Bitmap); k++ {
		bc.Bitmap[k] = bc.Bitmap[k] &^ value2.Bitmap[k]
	}
	bc.Cardinality = newCardinality
	return bc
}

func (bc *bitmapContainer) iandNotBitmap(value2 *bitmapContainer) container {
	newCardinality := int(popcntMaskSlice(bc.Bitmap, value2.Bitmap))
	if newCardinality > arrayDefaultMaxSize {
		for k := 0; k < len(bc.Bitmap); k++ {
			bc.Bitmap[k] = bc.Bitmap[k] &^ value2.Bitmap[k]
		}
		bc.Cardinality = newCardinality
		return bc
	}
	ac := newArrayContainerSize(newCardinality)
	fillArrayANDNOT(ac.Content, bc.Bitmap, value2.Bitmap)
	return ac
}

func (bc *bitmapContainer) contains(i uint16) bool { //testbit
	x := int(i)
	mask := uint64(1) << uint(x%64)
	return (bc.Bitmap[x/64] & mask) != 0
}

func (bc *bitmapContainer) loadData(arrayContainer *arrayContainer) {
	bc.Cardinality = arrayContainer.getCardinality()
	c := arrayContainer.getCardinality()
	for k := 0; k < c; k++ {
		x := arrayContainer.Content[k]
		i := int(x) / 64
		bc.Bitmap[i] |= (uint64(1) << uint(x%64))
	}
}

func (bc *bitmapContainer) toArrayContainer() *arrayContainer {
	ac := newArrayContainerCapacity(bc.Cardinality)
	ac.loadData(bc)
	return ac
}

func (bc *bitmapContainer) fillArray(container []uint16) {
	//TODO: rewrite in assembly
	pos := 0
	base := 0
	for k := 0; k < len(bc.Bitmap); k++ {
		bitset := bc.Bitmap[k]
		for bitset != 0 {
			t := bitset & -bitset
			container[pos] = uint16((base + int(popcount(t-1))))
			pos = pos + 1
			bitset ^= t
		}
		base += 64
	}
}

func (bc *bitmapContainer) NextSetBit(i int) int {
	x := i / 64
	if x >= len(bc.Bitmap) {
		return -1
	}
	w := bc.Bitmap[x]
	w = w >> uint(i%64)
	if w != 0 {
		return i + numberOfTrailingZeros(w)
	}
	x++
	for ; x < len(bc.Bitmap); x++ {
		if bc.Bitmap[x] != 0 {
			return (x * 64) + numberOfTrailingZeros(bc.Bitmap[x])
		}
	}
	return -1
}

// reference the java implementation
// https://github.com/RoaringBitmap/RoaringBitmap/blob/master/src/main/java/org/roaringbitmap/BitmapContainer.java#L875-L892
//
func (bc *bitmapContainer) numberOfRuns() int {
	if bc.Cardinality == 0 {
		return 0
	}

	var numRuns uint64
	nextWord := bc.Bitmap[0]

	for i := 0; i < len(bc.Bitmap)-1; i++ {
		word := nextWord
		nextWord = bc.Bitmap[i+1]
		numRuns += popcount((^word)&(word<<1)) + ((word >> 63) &^ nextWord)
	}

	word := nextWord
	numRuns += popcount((^word) & (word << 1))
	if (word & 0x8000000000000000) != 0 {
		numRuns++
	}

	return int(numRuns)
}

// convert to run or array *if needed*
func (bc *bitmapContainer) toEfficientContainer() container {

	numRuns := bc.numberOfRuns()

	sizeAsRunContainer := runContainer16SerializedSizeInBytes(numRuns)
	sizeAsBitmapContainer := bitmapContainerSizeInBytes()
	card := int(bc.getCardinality())
	sizeAsArrayContainer := arrayContainerSizeInBytes(card)

	if sizeAsRunContainer <= min(sizeAsBitmapContainer, sizeAsArrayContainer) {
		return newRunContainer16FromBitmapContainer(bc)
	}
	if card <= arrayDefaultMaxSize {
		return bc.toArrayContainer()
	}
	return bc
}

func newBitmapContainerFromRun(rc *runContainer16) *bitmapContainer {
	bc := newBitmapContainer()
	for i := range rc.Iv {
		bc.iaddRange(int(rc.Iv[i].Start), int(rc.Iv[i].Last)+1)
	}
	return bc
}

func (bc *bitmapContainer) containerType() contype {
	return bitmapContype
}
