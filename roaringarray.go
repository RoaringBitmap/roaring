package roaring

import (
	"bytes"
	"encoding/gob"
	"io"
	"log"
)

type container interface {
	clone() container
	and(container) container
	andNot(container) container
	inot(firstOfRange, lastOfRange int) container
	getCardinality() int
	add(uint16) container
	not(start, final int) container
	xor(r container) container
	getShortIterator() shortIterable
	contains(i uint16) bool
	equals(i interface{}) bool
	fillLeastSignificant16bits(array []int, i, mask int)
	or(r container) container
}

func rangeOfOnes(start, last int) container {
	if (last-start+1) > array_default_max_size {
		return newBitmapContainerwithRange(start, last)
	}

	return newArrayContainerRange(start, last)
}



type element struct {
	key   uint16
	value container
}

func (self *element) clone() element {
	var c element
	c.key = self.key
	c.value = self.value.clone()
	return c
}

func (self *element) gobEncode() (buf []byte, err error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	//gob.Register(self.container)
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

func (self *element) gobDecode(buf []byte) error {
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

func newelement(key uint16, value container) *element {
	ptr := new(element)
	ptr.key = key
	ptr.value = value
	return ptr
}

type roaringArray struct {
	array []*element
}

func newRoaringArray() *roaringArray {
	return &roaringArray{make([]*element, 0, 0)}
}

func (self *roaringArray) append(key uint16, value container) {
	self.array = append(self.array, newelement(key, value))
}

func (self *roaringArray) appendCopy(sa roaringArray, startingindex int) {
	self.array = append(self.array, newelement(sa.array[startingindex].key, sa.array[startingindex].value.clone()))
}

func (self *roaringArray) appendCopyMany(sa roaringArray, startingindex, end int) {
	for i := startingindex; i < end; i++ {
		self.appendCopy(sa, i)
	}
}

func (self *roaringArray) appendCopiesUntil(sa roaringArray, stoppingKey uint16) {
	for i := 0; i < sa.size(); i++ {
		if sa.array[i].key >= stoppingKey {
			break
		}
		self.array = append(self.array, newelement(sa.array[i].key, sa.array[i].value.clone()))
	}
}

func (self *roaringArray) appendCopiesAfter(sa roaringArray, beforeStart uint16) {
	startLocation := sa.getIndex(beforeStart)
	if startLocation >= 0 {
		startLocation++
	} else {
		startLocation = -startLocation - 1
	}

	for i := startLocation; i < sa.size(); i++ {
		self.array = append(self.array, newelement(sa.array[i].key, sa.array[i].value.clone()))
	}
}

func (self *roaringArray) clear() {
	self.array = make([]*element, 0, 0)
}

func (self *roaringArray) clone() *roaringArray {
	sa := new(roaringArray)
	sa.array = make([]*element, len(self.array))
	copy(sa.array, self.array[:])
	return sa
}

func (self *roaringArray) containsKey(x uint16) bool {
	return (self.binarySearch(0, len(self.array), x) >= 0)
}

func (self *roaringArray) getContainer(x uint16) container {
	i := self.binarySearch(0, len(self.array), x)
	if i < 0 {
		return nil
	}
	return self.array[i].value
}

func (self *roaringArray) getContainerAtIndex(i int) container {
	return self.array[i].value
}

func (self *roaringArray) getIndex(x uint16) int {
	// before the binary search, we optimize for frequent cases
	size := len(self.array)
	if (size == 0) || (self.array[size-1].key == x) {
		return size - 1
	}
	return self.binarySearch(0, size, x)
}

func (self *roaringArray) getKeyAtIndex(i int) uint16 {
	return self.array[i].key
}

func (self *roaringArray) insertNewKeyValueAt(i int, key uint16, value container) {
	s := self.array
	s = append(s, nil)
	copy(s[i+1:], s[i:])
	s[i] = newelement(key, value)
	self.array = s
}

func (self *roaringArray) remove(key uint16) bool {
	i := self.binarySearch(0, len(self.array), key)
	if i >= 0 { // if a new key
		self.removeAtIndex(i)
		return true
	}
	return false
}

func (self *roaringArray) removeAtIndex(i int) {
	a := self.array
	copy(a[i:], a[i+1:])
	a[len(a)-1] = nil // or the zero value of T
	a = a[:len(a)-1]
	self.array = a //should be the same reference i think
}

func (self *roaringArray) setContainerAtIndex(i int, c container) {
	self.array[i].value = c
}

func (self *roaringArray) size() int {
	return len(self.array)
}

func (self *roaringArray) binarySearch(begin, end int, key uint16) int {
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
func (self *roaringArray) equals(o interface{}) bool {
	srb, ok := o.(roaringArray)
	if ok {

		if srb.size() != self.size() {
			log.Println("NOT SAME SIZE", srb.size(), self.size())
			return false
		}
		for i := 0; i < srb.size(); i++ {
			oself := self.array[i]
			other := srb.array[i]
			if oself.key != other.key || !oself.value.equals(other.value) {
				return false
			}
		}
		return true
	}
	log.Println("NOPE")
	return false
}

func (self *roaringArray) serialize(out io.Writer) error {
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

func (self *roaringArray) deserialize(in io.Reader) error {
	dec := gob.NewDecoder(in)
	var size int
	err := dec.Decode(&size)
	if err != nil {
		return err
	}
	self.array = make([]*element, size, size)
	for i := 0; i < size; i++ {
		element := new(element)
		err = dec.Decode(&element)
		if err != nil {
			return err
		}
		self.array[i] = element
	}
	return nil
}
