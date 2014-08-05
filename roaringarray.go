package goroaring

import (
	"bytes"
	"encoding/gob"
	"io"
	"log"
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
	GetShortIterator() ShortIterable
	Contains(i short) bool
	Equals(i interface{}) bool
	FillLeastSignificant16bits(array []int, i, mask int)
	Or(r Container) Container
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

func (self *RoaringArray) AppendCopy(sa RoaringArray, startingindex int) {
	self.array = append(self.array, NewElement(sa.array[startingindex].key, sa.array[startingindex].value.Clone()))
}

func (self *RoaringArray) AppendCopyMany(sa RoaringArray, startingindex, end int) {
	for i := startingindex; i < end; i++ {
		self.AppendCopy(sa, i)
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
	sa := new(RoaringArray)
	sa.array = make([]*Element, len(self.array))
	copy(sa.array, self.array[:])
	return sa
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
func (self *RoaringArray) Equals(o interface{}) bool {
	//srb := o.(*RoaringArray)
	srb, ok := o.(RoaringArray)
	if ok {

		if srb.Size() != self.Size() {
			log.Println("NOT SAME SIZE", srb.Size(), self.Size())
			return false
		}
		for i := 0; i < srb.Size(); i++ {
			oself := self.array[i]
			other := srb.array[i]
			if oself.key != other.key || !oself.value.Equals(other.value) {
				return false
			}
		}
		return true
	}
	log.Println("NOPE")
	return false
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
