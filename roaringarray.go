package roaring

import (
	"bytes"
	"fmt"
	"io"

	snappy "github.com/glycerine/go-unsnap-stream"
	"github.com/tinylib/msgp/msgp"
)

//go:generate msgp -unexported

type container interface {
	clone() container
	and(container) container
	iand(container) container // i stands for inplace
	andNot(container) container
	iandNot(container) container // i stands for inplace
	getCardinality() int
	// rank returns the number of integers that are
	// smaller or equal to x. rank(infinity) would be getCardinality().
	rank(uint16) int

	iadd(x uint16) bool                   // inplace, returns true if x was new.
	iaddReturnMinimized(uint16) container // may change return type to minimize storage.

	//addRange(start, final int) container  // range is [firstOfRange,lastOfRange) (unused)
	iaddRange(start, final int) container // i stands for inplace, range is [firstOfRange,lastOfRange)

	iremove(x uint16) bool                   // inplace, returns true if x was present.
	iremoveReturnMinimized(uint16) container // may change return type to minimize storage.

	not(start, final int) container               // range is [firstOfRange,lastOfRange)
	inot(firstOfRange, lastOfRange int) container // i stands for inplace, range is [firstOfRange,lastOfRange)
	xor(r container) container
	getShortIterator() shortIterable
	contains(i uint16) bool

	// equals is now logical equals; it does not require the
	// same underlying container types, but compares across
	// any of the implementations.
	equals(i interface{}) bool

	fillLeastSignificant16bits(array []uint32, i int, mask uint32)
	or(r container) container
	ior(r container) container   // i stands for inplace
	intersects(r container) bool // whether the two containers intersect
	lazyOR(r container) container
	lazyIOR(r container) container
	getSizeInBytes() int
	//removeRange(start, final int) container  // range is [firstOfRange,lastOfRange) (unused)
	iremoveRange(start, final int) container // i stands for inplace, range is [firstOfRange,lastOfRange)
	selectInt(x uint16) int                  // selectInt returns the xth integer in the container
	serializedSizeInBytes() int
	readFrom(io.Reader) (int, error)
	writeTo(io.Writer) (int, error)

	numberOfRuns() int
	toEfficientContainer() container
	String() string
	containerType() contype
}

type contype uint8

const (
	bitmapContype contype = iota
	arrayContype
	run16Contype
	run32Contype
)

// careful: range is [firstOfRange,lastOfRange]
func rangeOfOnes(start, last int) container {
	if start > MaxUint16 {
		panic("rangeOfOnes called with start > MaxUint16")
	}
	if last > MaxUint16 {
		panic("rangeOfOnes called with last > MaxUint16")
	}
	if start < 0 {
		panic("rangeOfOnes called with start < 0")
	}
	if last < 0 {
		panic("rangeOfOnes called with last < 0")
	}
	return newRunContainer16Range(uint16(start), uint16(last))
}

type roaringArray struct {
	keys            []uint16
	containers      []container `msg:"-"` // don't try to serialize directly.
	needCopyOnWrite []bool
	copyOnWrite     bool

	// conserz is used at serialization time
	// to serialize containers. Otherwise empty.
	conserz []containerSerz
}

// containerSerz facilitates serializing container (tricky to
// serialize because it is an interface) by providing a
// light wrapper with a type identifier.
type containerSerz struct {
	t contype  `msg:"t"` // type
	r msgp.Raw `msg:"r"` // Raw msgpack of the actual container type
}

func newRoaringArray() *roaringArray {
	return &roaringArray{}
}

// runOptimize compresses the element containers to minimize space consumed.
// Q: how does this interact with copyOnWrite and needCopyOnWrite?
// A: since we aren't changing the logical content, just the representation,
//    we don't both to check the needCopyOnWrite bits. We replace
//    possible all elements of ra.containers in-place with space
//    optimized versions.
func (ra *roaringArray) runOptimize() {
	for i := range ra.containers {
		ra.containers[i] = ra.containers[i].toEfficientContainer()
	}
}

func (ra *roaringArray) appendContainer(key uint16, value container, mustCopyOnWrite bool) {
	ra.keys = append(ra.keys, key)
	ra.containers = append(ra.containers, value)
	ra.needCopyOnWrite = append(ra.needCopyOnWrite, mustCopyOnWrite)
}

func (ra *roaringArray) appendWithoutCopy(sa roaringArray, startingindex int) {
	ra.appendContainer(sa.keys[startingindex], sa.containers[startingindex], false)
}

func (ra *roaringArray) appendCopy(sa roaringArray, startingindex int) {
	ra.appendContainer(sa.keys[startingindex], sa.containers[startingindex], true)
	sa.setNeedsCopyOnWrite(startingindex)
}

func (ra *roaringArray) appendWithoutCopyMany(sa roaringArray, startingindex, end int) {
	for i := startingindex; i < end; i++ {
		ra.appendWithoutCopy(sa, i)
	}
}

func (ra *roaringArray) appendCopyMany(sa roaringArray, startingindex, end int) {
	for i := startingindex; i < end; i++ {
		ra.appendCopy(sa, i)
	}
}

func (ra *roaringArray) appendCopiesUntil(sa roaringArray, stoppingKey uint16) {
	for i := 0; i < sa.size(); i++ {
		if sa.keys[i] >= stoppingKey {
			break
		}
		ra.appendContainer(sa.keys[i], sa.containers[i], true)
		sa.setNeedsCopyOnWrite(i)
	}
}

func (ra *roaringArray) appendCopiesAfter(sa roaringArray, beforeStart uint16) {
	startLocation := sa.getIndex(beforeStart)
	if startLocation >= 0 {
		startLocation++
	} else {
		startLocation = -startLocation - 1
	}

	for i := startLocation; i < sa.size(); i++ {
		ra.appendContainer(sa.keys[i], sa.containers[i], true)
		sa.setNeedsCopyOnWrite(i)
	}
}

func (ra *roaringArray) removeIndexRange(begin, end int) {
	if end <= begin {
		return
	}

	r := end - begin

	copy(ra.keys[begin:], ra.keys[end:])
	copy(ra.containers[begin:], ra.containers[end:])
	copy(ra.needCopyOnWrite[begin:], ra.needCopyOnWrite[end:])

	ra.resize(len(ra.keys) - r)
}

func (ra *roaringArray) resize(newsize int) {
	for k := newsize; k < len(ra.containers); k++ {
		ra.containers[k] = nil
	}

	ra.keys = ra.keys[:newsize]
	ra.containers = ra.containers[:newsize]
	ra.needCopyOnWrite = ra.needCopyOnWrite[:newsize]
}

func (ra *roaringArray) clear() {
	*ra = roaringArray{}
}

func (ra *roaringArray) clone() *roaringArray {

	// shallow copy, slices will have the same backing arrays.
	sa := *ra

	// this is where copyOnWrite is used.
	if ra.copyOnWrite {
		ra.markAllAsNeedingCopyOnWrite()
		// sa.needCopyOnWrite is shared
	} else {
		// make a full copy

		sa.keys = make([]uint16, len(ra.keys))
		copy(sa.keys, ra.keys)

		sa.containers = make([]container, len(ra.containers))
		for i := range sa.containers {
			sa.containers[i] = ra.containers[i].clone()
		}

		sa.needCopyOnWrite = make([]bool, len(ra.needCopyOnWrite))
	}
	return &sa
}

func (ra *roaringArray) containsKey(x uint16) bool {
	return (ra.binarySearch(0, int64(len(ra.keys)), x) >= 0)
}

func (ra *roaringArray) getContainer(x uint16) container {
	i := ra.binarySearch(0, int64(len(ra.keys)), x)
	if i < 0 {
		return nil
	}
	return ra.containers[i]
}

func (ra *roaringArray) getWritableContainerContainer(x uint16) container {
	i := ra.binarySearch(0, int64(len(ra.keys)), x)
	if i < 0 {
		return nil
	}
	if ra.needCopyOnWrite[i] {
		ra.containers[i] = ra.containers[i].clone()
		ra.needCopyOnWrite[i] = false
	}
	return ra.containers[i]
}

func (ra *roaringArray) getContainerAtIndex(i int) container {
	return ra.containers[i]
}

func (ra *roaringArray) getWritableContainerAtIndex(i int) container {
	if ra.needCopyOnWrite[i] {
		ra.containers[i] = ra.containers[i].clone()
		ra.needCopyOnWrite[i] = false
	}
	return ra.containers[i]
}

func (ra *roaringArray) getIndex(x uint16) int {
	// before the binary search, we optimize for frequent cases
	size := len(ra.keys)
	if (size == 0) || (ra.keys[size-1] == x) {
		return size - 1
	}
	return int(ra.binarySearch(0, int64(size), x))
}

func (ra *roaringArray) getKeyAtIndex(i int) uint16 {
	return ra.keys[i]
}

func (ra *roaringArray) insertNewKeyValueAt(i int, key uint16, value container) {
	ra.keys = append(ra.keys, 0)
	ra.containers = append(ra.containers, nil)

	copy(ra.keys[i+1:], ra.keys[i:])
	copy(ra.containers[i+1:], ra.containers[i:])

	ra.keys[i] = key
	ra.containers[i] = value

	ra.needCopyOnWrite = append(ra.needCopyOnWrite, false)
	copy(ra.needCopyOnWrite[i+1:], ra.needCopyOnWrite[i:])
	ra.needCopyOnWrite[i] = false
}

func (ra *roaringArray) remove(key uint16) bool {
	i := ra.binarySearch(0, int64(len(ra.keys)), key)
	if i >= 0 { // if a new key
		ra.removeAtIndex(i)
		return true
	}
	return false
}

func (ra *roaringArray) removeAtIndex(i int) {
	copy(ra.keys[i:], ra.keys[i+1:])
	copy(ra.containers[i:], ra.containers[i+1:])

	copy(ra.needCopyOnWrite[i:], ra.needCopyOnWrite[i+1:])

	ra.resize(len(ra.keys) - 1)
}

func (ra *roaringArray) setContainerAtIndex(i int, c container) {
	ra.containers[i] = c
}

func (ra *roaringArray) replaceKeyAndContainerAtIndex(i int, key uint16, c container, mustCopyOnWrite bool) {
	ra.keys[i] = key
	ra.containers[i] = c
	ra.needCopyOnWrite[i] = mustCopyOnWrite
}

func (ra *roaringArray) size() int {
	return len(ra.keys)
}

func (ra *roaringArray) binarySearch(begin, end int64, ikey uint16) int {
	low := begin
	high := end - 1
	for low+16 <= high {
		middleIndex := low + (high-low)/2 // avoid overflow
		middleValue := ra.keys[middleIndex]

		if middleValue < ikey {
			low = middleIndex + 1
		} else if middleValue > ikey {
			high = middleIndex - 1
		} else {
			return int(middleIndex)
		}
	}
	for ; low <= high; low++ {
		val := ra.keys[low]
		if val >= ikey {
			if val == ikey {
				return int(low)
			}
			break
		}
	}
	return -int(low + 1)
}

func (ra *roaringArray) equals(o interface{}) bool {
	srb, ok := o.(roaringArray)
	if ok {

		if srb.size() != ra.size() {
			return false
		}
		for i, k := range ra.keys {
			if k != srb.keys[i] {
				return false
			}
		}

		for i, c := range ra.containers {
			if !c.equals(srb.containers[i]) {
				return false
			}
		}
		return true
	}
	return false
}

// warning: this is expensive. We actually do the serialization to compute
// the size. Use only for testing.
func (ra *roaringArray) serializedSizeInBytes() uint64 {
	var buf bytes.Buffer
	ra.writeTo(&buf)
	return uint64(len(buf.Bytes()))
}

func (ra *roaringArray) writeTo(stream io.Writer) error {

	ra.conserz = make([]containerSerz, len(ra.containers))
	for i, v := range ra.containers {
		switch cn := v.(type) {
		case *bitmapContainer:
			bts, err := cn.MarshalMsg(nil)
			if err != nil {
				return err
			}
			ra.conserz[i].t = bitmapContype
			ra.conserz[i].r = bts
		case *arrayContainer:
			bts, err := cn.MarshalMsg(nil)
			if err != nil {
				return err
			}
			ra.conserz[i].t = arrayContype
			ra.conserz[i].r = bts
		case *runContainer16:
			bts, err := cn.MarshalMsg(nil)
			if err != nil {
				return err
			}
			ra.conserz[i].t = run16Contype
			ra.conserz[i].r = bts
		default:
			fmt.Errorf("Unrecognized container implementation: %T", cn)
		}
	}
	w := snappy.NewWriter(stream)
	err := msgp.Encode(w, ra)
	ra.conserz = nil
	return err
}

func (ra *roaringArray) readFrom(stream io.Reader) error {
	r := snappy.NewReader(stream)
	err := msgp.Decode(r, ra)
	if err != nil {
		return err
	}

	if len(ra.containers) != len(ra.keys) {
		ra.containers = make([]container, len(ra.keys))
	}

	for i, v := range ra.conserz {
		switch v.t {
		case bitmapContype:
			c := &bitmapContainer{}
			_, err = c.UnmarshalMsg(v.r)
			if err != nil {
				return err
			}
			ra.containers[i] = c
		case arrayContype:
			c := &arrayContainer{}
			_, err = c.UnmarshalMsg(v.r)
			if err != nil {
				return err
			}
			ra.containers[i] = c
		case run16Contype:
			c := &runContainer16{}
			_, err = c.UnmarshalMsg(v.r)
			if err != nil {
				return err
			}
			ra.containers[i] = c
		default:
			return fmt.Errorf("unrecognized contype serialization code: '%v'", v.t)
		}
	}
	ra.conserz = nil
	return nil
}

func (ra *roaringArray) advanceUntil(min uint16, pos int) int {
	lower := pos + 1

	if lower >= len(ra.keys) || ra.keys[lower] >= min {
		return lower
	}

	spansize := 1

	for lower+spansize < len(ra.keys) && ra.keys[lower+spansize] < min {
		spansize *= 2
	}
	var upper int
	if lower+spansize < len(ra.keys) {
		upper = lower + spansize
	} else {
		upper = len(ra.keys) - 1
	}

	if ra.keys[upper] == min {
		return upper
	}

	if ra.keys[upper] < min {
		// means
		// array
		// has no
		// item
		// >= min
		// pos = array.length;
		return len(ra.keys)
	}

	// we know that the next-smallest span was too small
	lower += (spansize / 2)

	mid := 0
	for lower+1 != upper {
		mid = (lower + upper) / 2
		if ra.keys[mid] == min {
			return mid
		} else if ra.keys[mid] < min {
			lower = mid
		} else {
			upper = mid
		}
	}
	return upper
}

func (ra *roaringArray) markAllAsNeedingCopyOnWrite() {
	for i := range ra.needCopyOnWrite {
		ra.needCopyOnWrite[i] = true
	}
}

func (ra *roaringArray) needsCopyOnWrite(i int) bool {
	return ra.needCopyOnWrite[i]
}

func (ra *roaringArray) setNeedsCopyOnWrite(i int) {
	ra.needCopyOnWrite[i] = true
}
