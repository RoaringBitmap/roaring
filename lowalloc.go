package roaring

import (
	"fmt"
	"math/bits"
	"sync"
	"unsafe"
)

const (
	containerStorageBytes     = 8192
	bitmapLongCount           = containerStorageBytes >> 3
	containerCountOffset      = uint32(4)
	containerDescriptionStart = uint32(8)
	bytesPerContainer         = uint32(4)

	maxContainers          = uint32(256)
	fatWritableSliceLength = int(maxContainers * containerStorageBytes)

	// 0: 2 bytes for key,
	// 2: 2 bytes for cardinality - 1
	cardinalityIncrement = uint32(2)
)

var writableArrayPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 0, fatWritableSliceLength)
	},
}

type RoaringBitmap struct {
	data         []byte
	header       []byte
	containerMax uint32
	containers   uint32
	writable     bool
}

func (bitmap *RoaringBitmap) FromBuffer(bytes []byte) error {
	cookie := ReadSingleInt(bytes, 0)
	if cookie != serialCookieNoRunContainer {
		return fmt.Errorf("unexpected signature %d", cookie)
	}
	size := ReadSingleInt(bytes, 4)
	if size > 0 {
		pointer := 8 + (4 * (size - 1))
		finalContainerCardinality := ReadSingleShort(bytes, pointer+2)
		pointer += 4 * size
		finalOffset := ReadSingleInt(bytes, pointer) + uint32(lengthFromCardinalityShort(finalContainerCardinality))
		if finalOffset != uint32(len(bytes)) {
			return fmt.Errorf("expect the bitmap to be %d bytes, not %d bytes",
				finalOffset,
				len(bytes))
		}
	}
	bitmap.data = bytes
	bitmap.writable = false
	return nil
}

// Contains returns true if the integer is contained in the bitmap
// this isn't meant to be performant, just here for tests.
func (bitmap *RoaringBitmap) Contains(x uint32) bool {
	key := uint16(x >> 16)
	containers := bitmap.getContainerCount()
	for i := uint32(0); i < containers; i++ {
		currentKey := bitmap.getKeyAtContainerIndex(i)
		if currentKey == key {
			target := uint16(x)
			cardinality := uint32(bitmap.getCardinalityMinusOneFromContainerIndex(i)) + 1
			offset := bitmap.getOffsetForKeyAtPosition(uint32(key), i)
			if cardinality <= arrayDefaultMaxSize {
				for j := uint32(0); j < cardinality; j++ {
					val := ReadSingleShort(bitmap.data, offset+2*j)
					if val == target {
						return true
					} else if target < val {
						return false
					}
				}
			} else {
				byt := bitmap.data[offset+uint32(target>>3)]
				mask := byte(1 << (target & 7))
				return byt&mask != 0
			}
		} else if currentKey > key {
			return false
		}
	}
	return false
}

// This is intended for reusing the underlying slices.
func (bitmap *RoaringBitmap) Clear() {
	if bitmap.writable {
		bitmap.containers = 0
	} else {
		bitmap.data = nil
	}
}

func (bitmap *RoaringBitmap) Clone() *RoaringBitmap {
	if bitmap == nil {
		// this is probably a programming error, I think
		return &RoaringBitmap{}
	}
	if bitmap.writable {
		newData := writableArrayPool.Get().([]byte)
		clone := &RoaringBitmap{
			data:         newData,
			header:       make([]byte, len(bitmap.header), 4*bitmap.containerMax),
			containers:   bitmap.containers,
			containerMax: 1,
			writable:     true,
		}
		copy(clone.header, bitmap.header)
		clone.increaseMaxContainers(bitmap.containerMax)
		//TODO: optionally copy just the live data
		copy(clone.data, bitmap.data)

		return clone
	}
	// if it isn't writable,
	// then a separate pointer to the same data is safe.
	return &RoaringBitmap{data: bitmap.data}
}

func (bitmap *RoaringBitmap) MakeWritable() {
	bitmap.MakeWritableWithConfiguredContainerMax(maxContainers)
}

func (bitmap *RoaringBitmap) MakeWritableWithConfiguredContainerMax(containerMax uint32) {
	if bitmap.writable {
		return
	}
	if len(bitmap.data) == 0 {
		//bootstrap with empty container bytes. Hopefully this shouldn't be common.
		bitmap.data = []byte{58, 58, 0, 0,
			0, 0, 0, 0}
	}
	bitmap.containers = ReadSingleInt(bitmap.data, containerCountOffset)
	bitmap.containerMax = containerMax
	if bitmap.containers > 0 {
		bitmapMax := uint32(bitmap.getKeyAtContainerIndex(bitmap.containers-1)) + 1
		if bitmapMax >= containerMax {
			bitmap.containerMax = bitmapMax
		}
	}

	newData := writableArrayPool.Get().([]byte)
	if cap(newData) < int(bitmap.containerMax*containerStorageBytes) {
		newData = make([]byte, bitmap.containerMax*containerStorageBytes)
	} else {
		newData = newData[:bitmap.containerMax*containerStorageBytes]
	}

	bitmap.header = make([]byte, bitmap.containers*bytesPerContainer, bitmap.containerMax*bytesPerContainer)
	copy(bitmap.header, bitmap.data[containerDescriptionStart:containerDescriptionStart+4*bitmap.containers])

	offsetReadPointer := 8 + 4*bitmap.containers
	// need to set this so we can just read off the header.
	bitmap.writable = true
	for i := uint32(0); i < bitmap.containers; i++ {
		key := bitmap.getKeyAtContainerIndex(i)
		cardinalityMinusOne := bitmap.getCardinalityMinusOneFromContainerIndex(i)
		oldOffset := ReadSingleInt(bitmap.data, offsetReadPointer)

		copy(newData[uint32(key)*containerStorageBytes:],
			bitmap.data[oldOffset:oldOffset+
				uint32(lengthFromCardinalityShort(cardinalityMinusOne))])
		offsetReadPointer += 4
	}
	bitmap.data = newData
}

func (bitmap *RoaringBitmap) getContainerCount() uint32 {
	if bitmap == nil || len(bitmap.data) == 0 {
		return 0
	} else if bitmap.writable {
		return bitmap.containers
	}
	return ReadSingleInt(bitmap.data, containerCountOffset)
}

func (bitmap *RoaringBitmap) IsEmpty() bool {
	return bitmap == nil || len(bitmap.data) == 0 || (bitmap.writable && bitmap.containers == 0)
}

func (bitmap *RoaringBitmap) getOffsetForKeyAtPosition(key uint32, pos uint32) uint32 {
	if bitmap.writable {
		return key * containerStorageBytes
	}
	return ReadSingleInt(bitmap.data, 8+4*bitmap.getContainerCount()+4*pos)
}

func (bitmap *RoaringBitmap) getCardinalityMinusOneFromContainerIndex(pos uint32) uint16 {
	if bitmap.writable {
		return ReadSingleShort(bitmap.header, bytesPerContainer*pos+cardinalityIncrement)
	}
	return ReadSingleShort(bitmap.data, 8+4*pos+2)
}

func (bitmap *RoaringBitmap) getKeyAtContainerIndex(index uint32) uint16 {
	if bitmap.writable {
		return ReadSingleShort(bitmap.header, bytesPerContainer*index)
	} else {
		return ReadSingleShort(bitmap.data, 8+4*index)
	}
}

func (bitmap *RoaringBitmap) increaseMaxContainers(newValue uint32) {
	if bitmap.containerMax == 0 {
		panic("containerMax really shouldn't be zero")
	}
	if newValue < bitmap.containerMax*2 {
		bitmap.containerMax *= 2
	} else {
		bitmap.containerMax = newValue
	}
	if cap(bitmap.header) < int(bitmap.containerMax*bytesPerContainer) {
		newHeader := make([]byte, len(bitmap.header), bitmap.containerMax*bytesPerContainer)
		copy(newHeader, bitmap.header)
		bitmap.header = newHeader
	}

	if cap(bitmap.data) < int(containerStorageBytes*bitmap.containerMax) {
		newData := make([]byte, containerStorageBytes*bitmap.containerMax, containerStorageBytes*bitmap.containerMax)
		copy(newData, bitmap.data)
		bitmap.data = newData
	} else {
		bitmap.data = bitmap.data[:containerStorageBytes*bitmap.containerMax]
	}
}

func (bitmap *RoaringBitmap) Free() {
	if bitmap != nil && bitmap.writable && len(bitmap.data) > 0 {
		writableArrayPool.Put(bitmap.data)
	}
}

func RoaringAnd(left, right *RoaringBitmap) *RoaringBitmap {
	if left.IsEmpty() || right.IsEmpty() {
		return &RoaringBitmap{}
	}
	clone := left.Clone()
	clone.MakeWritable()
	clone.And(right)
	return clone
}

func RoaringOr(left, right *RoaringBitmap) *RoaringBitmap {
	// if left is already writable, MakeRoomy does nothing.
	// if left is not writable, Clone() is a no-op.
	clone := left.Clone()
	clone.MakeWritable()
	clone.Or(right)
	return clone
}

func RoaringAndNot(left *RoaringBitmap, right *RoaringBitmap) *RoaringBitmap {
	clone := left.Clone()
	clone.AndNot(right)
	return clone
}

func (bitmap *RoaringBitmap) GetCardinality() uint64 {
	if bitmap == nil || len(bitmap.data) == 0 {
		return 0
	}
	containers := bitmap.getContainerCount()
	result := uint64(0)
	increment := bytesPerContainer
	pointer := containerDescriptionStart
	headerSlice := bitmap.data
	if bitmap.writable {
		headerSlice = bitmap.header
		pointer = 0
	}
	for i := uint32(0); i < containers; i++ {
		result += uint64(ReadSingleShort(headerSlice, pointer+cardinalityIncrement)) + 1
		pointer += increment
	}
	return result
}

func (bitmap *RoaringBitmap) AndCardinality(other *RoaringBitmap) uint64 {
	pos1 := uint32(0)
	pos2 := uint32(0)
	answer := uint64(0)
	length1 := bitmap.getContainerCount()
	length2 := other.getContainerCount()

main:
	for {
		if pos1 < length1 && pos2 < length2 {
			s1 := bitmap.getKeyAtContainerIndex(pos1)
			s2 := other.getKeyAtContainerIndex(pos2)
			for {
				if s1 == s2 {
					dataOffset1 := bitmap.getOffsetForKeyAtPosition(uint32(s1), pos1)
					card1 := bitmap.getCardinalityMinusOneFromContainerIndex(pos1)
					dataOffset2 := other.getOffsetForKeyAtPosition(uint32(s2), pos2)
					card2 := other.getCardinalityMinusOneFromContainerIndex(pos2)
					answer += uint64(offsetAndCardinality(bitmap.data, dataOffset1, card1,
						other.data, dataOffset2, card2))
					pos1++
					pos2++
					if (pos1 == length1) || (pos2 == length2) {
						break main
					}
					s1 = bitmap.getKeyAtContainerIndex(pos1)
					s2 = other.getKeyAtContainerIndex(pos2)
				} else if s1 < s2 {
					for s1 < s2 {
						pos1++
						if pos1 == length1 {
							break main
						}
						s1 = bitmap.getKeyAtContainerIndex(pos1)
					}
				} else { //s1 > s2
					for s1 > s2 {
						pos2++
						if pos2 == length2 {
							break main
						}
						s2 = other.getKeyAtContainerIndex(pos2)
					}
				}
			}
		} else {
			break
		}
	}
	return answer
}

func (bitmap *RoaringBitmap) Or(other *RoaringBitmap) {
	if other == nil || other.data == nil {
		return
	}
	if bitmap == nil {
		panic("can't call Or on nil, since we can't assign back to a nil pointer.")
	}
	if bitmap.data == nil {
		if other.writable {
			clone := other.Clone()
			bitmap.data = clone.data
			//TODO: figure out the semantics here.
			bitmap.writable = clone.writable
			bitmap.header = clone.header
			bitmap.containerMax = clone.containerMax
			bitmap.containers = clone.containers
			return
		}
		bitmap.data = other.data
		bitmap.writable = false
		return
	}
	// there's a thing here that isn't currently being done.
	// That is when an empty writable bitmap
	//  is ORed against a non-writable bitmap, you could convert to
	// a non-writable bitmap backed by the same slice.
	// I've currently not done that as it'll result in slice churn
	// and make it harder to explicitly manage memory.
	// Most of our OR operations happen in a sequence,
	// so even if the first one would be faster as a rereference
	// you'll have to convert to writable eventually.
	// The code would look like:
	/*
		if bitmap.writable && bitmap.containers == 0 {
		    bitmap.Free()
			bitmap.data = other.data
			bitmap.writable = false
			return
		}

	*/
	if !bitmap.writable {
		bitmap.MakeWritable()
	}
	bitmap.computeOrAgainst(other)
}

func (bitmap *RoaringBitmap) And(other *RoaringBitmap) {
	if bitmap == nil {
		panic("can't call And() on nil, since we can't assign back to a nil pointer.")
	}
	// already empty
	if len(bitmap.data) == 0 || bitmap.getContainerCount() == 0 {
		return
	}
	// the result is empty, a couple ways to clear it.
	if other == nil || len(other.data) == 0 {
		if bitmap.writable {
			bitmap.containers = 0
		} else {
			bitmap.data = nil
			bitmap.writable = true
		}
		return
	}
	if !bitmap.writable {
		bitmap.MakeWritable()
	}
	bitmap.computeAndAgainst(other)
}

func (bitmap *RoaringBitmap) AndNot(other *RoaringBitmap) {
	if bitmap.IsEmpty() {
		return
	}
	if !bitmap.writable {
		bitmap.MakeWritable()
	}
	bitmap.computeAndNotAgainst(other)
}

func (bitmap *RoaringBitmap) Xor(right *RoaringBitmap) {
	if !bitmap.writable {
		bitmap.MakeWritable()
	}
	bitmap.computeXor(right)
}

// this is just localintersect2by2, might want to do galloping intersections in the future.
func arrayAndCardinality(data1 []byte, offset1 uint32, shorts1 uint32, data2 []byte, offset2 uint32, shorts2 uint32) uint32 {
	if 0 == shorts1 || 0 == shorts2 {
		return 0
	}
	k1 := uint32(0)
	k2 := uint32(0)
	pos := uint32(0)
	s1 := ReadSingleShort(data1, offset1+2*k1)
	s2 := ReadSingleShort(data2, offset2+2*k2)
mainwhile:
	for {
		if s2 < s1 {
			for {
				k2++
				if k2 == shorts2 {
					break mainwhile
				}
				s2 = ReadSingleShort(data2, offset2+2*k2)
				if s2 >= s1 {
					break
				}
			}
		}
		if s1 < s2 {
			for {
				k1++
				if k1 == shorts1 {
					break mainwhile
				}
				s1 = ReadSingleShort(data1, offset1+2*k1)
				if s1 >= s2 {
					break
				}
			}
		} else {
			// (set2[k2] == set1[k1])
			pos++
			k1++
			if k1 == shorts1 {
				break
			}
			s1 = ReadSingleShort(data1, offset1+2*k1)
			k2++
			if k2 == shorts2 {
				break
			}
			s2 = ReadSingleShort(data2, offset2+2*k2)
		}
	}
	return pos
}

func lengthFromCardinalityShort(cardinalityMinusOne uint16) uint16 {
	if cardinalityMinusOne < arrayDefaultMaxSize {
		return 2 + 2*cardinalityMinusOne
	}
	return containerStorageBytes
}

func (bitmap *RoaringBitmap) computeOrAgainst(x2 *RoaringBitmap) {
	if !bitmap.writable {
		panic("can't call in place method on non-writable bitmap.")
	}
	pos1 := uint32(0)
	pos2 := uint32(0)
	length1 := bitmap.getContainerCount()
	length2 := x2.getContainerCount()
main:
	for (pos1 < length1) && (pos2 < length2) {
		s1 := bitmap.getKeyAtContainerIndex(pos1)
		s2 := x2.getKeyAtContainerIndex(pos2)

		for {
			if s1 < s2 {
				pos1++
				if pos1 == length1 {
					break main
				}
				s1 = bitmap.getKeyAtContainerIndex(pos1)
			} else if s1 > s2 {
				cardShort := x2.getCardinalityMinusOneFromContainerIndex(pos2)
				length := lengthFromCardinalityShort(cardShort)
				offset := x2.getOffsetForKeyAtPosition(uint32(s2), pos2)
				bitmap.insertNewContainerAtIndex(pos1, s2, cardShort, x2.data, offset, length)
				pos1++
				length1++
				pos2++
				if pos2 == length2 {
					break main
				}
				s2 = x2.getKeyAtContainerIndex(pos2)
			} else {
				cardShort := x2.getCardinalityMinusOneFromContainerIndex(pos2)
				offset := x2.getOffsetForKeyAtPosition(uint32(s2), pos2)
				bitmap.orContainerAtIndex(pos1, x2.data, offset, cardShort)
				pos1++
				pos2++
				if (pos1 == length1) || (pos2 == length2) {
					break main
				}
				s1 = bitmap.getKeyAtContainerIndex(pos1)
				s2 = x2.getKeyAtContainerIndex(pos2)
			}
		}
	}
	if pos1 == length1 {
		for pos2 < length2 {
			s2 := x2.getKeyAtContainerIndex(pos2)
			cardShort := x2.getCardinalityMinusOneFromContainerIndex(pos2)
			length := lengthFromCardinalityShort(cardShort)
			offset := x2.getOffsetForKeyAtPosition(uint32(s2), pos2)
			bitmap.insertNewContainerAtIndex(pos1, s2, cardShort, x2.data, offset, length)
			pos1++
			length1++
			pos2++
		}
	}
}

func (bitmap *RoaringBitmap) insertNewContainerAtIndex(containerIndex uint32, key uint16, cardinalityMinusOne uint16, sourceData []byte, offset uint32, length uint16) {
	if !bitmap.writable {
		panic("can't write to non-writable")
	}
	// check if you have room in the header.
	containerCount := bitmap.getContainerCount()
	if uint32(key) >= bitmap.containerMax {
		bitmap.increaseMaxContainers(uint32(key) + 1)
	}
	// shift all later container datas by 4 bytes
	bitmap.containers++
	bitmap.header = bitmap.header[:bytesPerContainer*bitmap.containers]

	copy(bitmap.header[bytesPerContainer*(containerIndex+1):bytesPerContainer*(containerCount+1)],
		bitmap.header[bytesPerContainer*containerIndex:bytesPerContainer*containerCount])

	//update header values, namely container count and new container's data
	WriteShort(bitmap.header, bytesPerContainer*containerIndex, key)
	WriteShort(bitmap.header, bytesPerContainer*containerIndex+cardinalityIncrement, cardinalityMinusOne)
	copy(bitmap.data[uint32(key)*containerStorageBytes:], sourceData[offset:offset+uint32(length)])
}

func (bitmap *RoaringBitmap) orContainerAtIndex(pos1 uint32, data []byte, offset uint32, cardinalityMinusOne uint16) {
	if cardinalityMinusOne < arrayDefaultMaxSize {
		bitmap.orContainerAgainstArrayAtIndex(pos1, data, offset, cardinalityMinusOne+1)
	} else {
		bitmap.orContainerAgainstBitmapAtIndex(pos1, data, offset)
	}
}

func (bitmap *RoaringBitmap) orContainerAgainstArrayAtIndex(pos1 uint32, data []byte, offset uint32, shorts uint16) {
	containerPointer := bytesPerContainer * pos1
	card := bitmap.getCardinalityMinusOneFromContainerIndex(pos1)
	//TODO: this stuff could be passed into the function, as it is always known.
	key := bitmap.getKeyAtContainerIndex(pos1)
	totalOffset := uint32(key) * containerStorageBytes
	if card < arrayDefaultMaxSize {
		maxSize := 2 * (card + shorts + 1)
		readOffset := totalOffset
		// two array containers,
		// First need to figure out if we know it will still be array without taking the union.
		if maxSize <= containerStorageBytes {
			// great, we know the result will still be an array container
			// and our 8192 bytes will be enough
			// and since we've fixed the sizes at 8192 for writable arrays,
			// no need to expand
			readOffset += 2 * uint32(shorts)
			copy(bitmap.data[totalOffset+2*uint32(shorts):totalOffset+uint32(maxSize)], bitmap.data[totalOffset:totalOffset+2*uint32(card)+2])

			// there'll likely be dirty bytes after the union, but we'll know we can use them.
			card = uint16(union2By2(bitmap.data, totalOffset, bitmap.data, readOffset, data, offset, uint32(shorts), uint32(card+1)))
			WriteShort(bitmap.header, containerPointer+cardinalityIncrement, card-1)
		} else {
			// it could be larger than fits in an array container,
			// and definitely needs more than 8192 working bytes.
			// First do a count check so we know the right final form.
			unionSize := byteBackedUnion2by2Cardinality(bitmap.data, data, totalOffset, offset, uint32(card)+1, uint32(shorts))
			WriteShort(bitmap.header, containerPointer+cardinalityIncrement, uint16(unionSize-1))
			if unionSize > arrayDefaultMaxSize {
				// this should be written as a bitmap.
				// use an array as working space.
				// there should be 8192 capacity.
				tmp := [containerStorageBytes]byte{}
				for i := uint32(0); i <= uint32(card); i++ {
					s := ReadSingleShort(bitmap.data, totalOffset+2*i)
					tmp[uint32(s>>3)] |= 1 << (s % 8)
				}
				for i := uint32(0); i < uint32(shorts); i++ {
					s := ReadSingleShort(data, offset+2*i)
					tmp[uint32(s>>3)] |= 1 << (s % 8)
				}
				copy(bitmap.data[totalOffset:], tmp[:containerStorageBytes])
			} else {
				// will be an array.
				// will fit 8192 bytes, but can't use capacity for working space.
				// Copy out current data into a tmp slice
				// this is the only temporary make call,
				// Could probably be done with an [8192]byte,
				// but don't want to reimplement union2By2 for that.
				// you could also imagine an optimistic union2by2,
				// which tracked if the zipper union overran the data
				// at the back of the working space.
				tmpSlice := make([]byte, 2*card+2)
				copy(tmpSlice, bitmap.data[totalOffset:])
				union2By2(bitmap.data, totalOffset, tmpSlice, 0, data, offset, uint32(shorts), uint32(card+1))
			}
		}
	} else {
		// bitmaps are easy.
		additions := uint16(0)
		for i := uint32(0); i < uint32(shorts); i++ {
			s := ReadSingleShort(data, offset+2*i)
			currentByte := bitmap.data[totalOffset+uint32(s)/8]
			if (currentByte>>(s%8))&1 == 0 {
				bitmap.data[totalOffset+uint32(s)/8] |= 1 << (s % 8)
				additions++
			}
		}
		if additions > 0 {
			WriteShort(bitmap.header, containerPointer+cardinalityIncrement, card+additions)
		}
	}
}

func (bitmap *RoaringBitmap) orContainerAgainstBitmapAtIndex(pos1 uint32, data []byte, offset uint32) {
	containerPointer := bytesPerContainer * pos1
	card := bitmap.getCardinalityMinusOneFromContainerIndex(pos1)
	//TODO: this stuff could be passed into the function, as it is always known.
	key := bitmap.getKeyAtContainerIndex(pos1)
	totalOffset := uint32(key) * containerStorageBytes
	if card < arrayDefaultMaxSize {
		// the result will still be a bitmap
		// this could maybe be an array instead of a slice.
		tmp := make([]byte, 2*card+2)
		copy(tmp, bitmap.data[totalOffset:])
		copy(bitmap.data[totalOffset:totalOffset+containerStorageBytes], data[offset:offset+containerStorageBytes])
		for i := uint32(0); i <= uint32(card); i++ {
			s := ReadSingleShort(tmp, 2*i)
			bitmap.data[totalOffset+uint32(s)/8] |= 1 << (s % 8)
		}

		cardinality := 0
		for i := uint32(0); i < 1024; i++ {
			cardinality += bits.OnesCount64(*(*uint64)(unsafe.Pointer(&bitmap.data[totalOffset+8*i])))
		}
		WriteShort(bitmap.header, containerPointer+cardinalityIncrement, uint16(cardinality-1))
	} else {
		cardinality := 0
		for i := uint32(0); i < 1024; i++ {
			*(*uint64)(unsafe.Pointer(&bitmap.data[totalOffset+8*i])) |=
				*(*uint64)(unsafe.Pointer(&data[offset+8*i]))
			cardinality += bits.OnesCount64(*(*uint64)(unsafe.Pointer(&bitmap.data[totalOffset+8*i])))
		}
		WriteShort(bitmap.header, containerPointer+cardinalityIncrement, uint16(cardinality-1))
	}
}

// returns the cardinality of the intersection.
// this allows for the caller to remove it from the header if necessary.
func (bitmap *RoaringBitmap) andContainerAtIndex(key uint32, pos1 uint32, data []byte, offset uint32, cardinalityMinusOne uint16) uint32 {
	if cardinalityMinusOne < arrayDefaultMaxSize {
		return bitmap.andContainerAgainstArrayAtIndex(key, pos1, data, offset, cardinalityMinusOne+1)
	} else {
		return bitmap.andContainerAgainstBitmapAtIndex(pos1, data, offset)
	}
}

func (bitmap *RoaringBitmap) andContainerAgainstArrayAtIndex(key uint32, pos1 uint32, data []byte, offset uint32, shorts uint16) uint32 {
	card := bitmap.getCardinalityMinusOneFromContainerIndex(pos1)
	totalOffset := key * containerStorageBytes
	if card < arrayDefaultMaxSize {
		// two array containers, do a union2by2 into this bitmap.
		// intersection of two arrays stays an array.
		cardinality := intersection2By2(bitmap.data, totalOffset, uint32(card+1), data, offset, uint32(shorts), bitmap.data, totalOffset)
		return cardinality
	} else {
		// in order to intersect bitmap's bitmap container
		// with an array container need a bit of working space.
		// do this by first creating a temporary bitmap in a 512 byte array,
		// which shouldn't escape from the heap.
		matches := [512]byte{}
		for i := uint32(0); i < uint32(shorts); i++ {
			s := ReadSingleShort(data, offset+2*i)
			currentByte := bitmap.data[totalOffset+uint32(s)>>3]
			if (currentByte & (1 << (s & 7))) > 0 {
				matches[i>>3] |= 1 << (i & 7)
			}
		}
		// TODO: it would be better to iterate over the bitmap, probably.
		written := uint32(0)
		for i := uint32(0); i < uint32(shorts); i++ {
			if matches[i>>3]&(1<<(i&7)) > 0 {
				s := ReadSingleShort(data, offset+2*i)
				WriteShort(bitmap.data, totalOffset+2*written, s)
				written++
			}
		}
		return written
	}
}

func (bitmap *RoaringBitmap) andContainerAgainstBitmapAtIndex(containerIndex uint32, data []byte, offset uint32) uint32 {
	card := bitmap.getCardinalityMinusOneFromContainerIndex(containerIndex)
	//TODO: this stuff could be passed into the function, as it is always known.
	key := bitmap.getKeyAtContainerIndex(containerIndex)
	totalOffset := uint32(key) * containerStorageBytes
	if card < arrayDefaultMaxSize {
		// sweet, array container against a bitmap, can just walk up it.
		written := uint32(0)
		for i := uint32(0); i <= uint32(card); i++ {
			s := ReadSingleShort(bitmap.data, totalOffset+2*i)
			if data[offset+uint32(s>>3)]&(1<<(s&7)) > 0 {
				WriteShort(bitmap.data, totalOffset+2*written, s)
				written++
			}
		}
		return written
	} else {
		cardinality := 0
		for i := uint32(0); i < 1024; i++ {
			*(*uint64)(unsafe.Pointer(&bitmap.data[totalOffset+8*i])) &=
				*(*uint64)(unsafe.Pointer(&data[offset+8*i]))
			cardinality += bits.OnesCount64(*(*uint64)(unsafe.Pointer(&bitmap.data[totalOffset+8*i])))
		}
		// gotta switch it to an array
		// write to a tmp array.
		if cardinality <= arrayDefaultMaxSize {
			container := [containerStorageBytes]byte{}
			pos := uint32(0)
			base := 0
			for k := uint32(0); k < bitmapLongCount && pos < uint32(cardinality); k++ {
				bitset := *(*uint64)(unsafe.Pointer(&bitmap.data[totalOffset+8*k]))
				for bitset != 0 {
					t := bitset & -bitset
					s := uint16(base + bits.OnesCount64(t-1))
					*(*uint16)(unsafe.Pointer(&container[2*pos])) = s
					pos++
					bitset ^= t
				}
				base += 64
			}
			copy(bitmap.data[totalOffset:totalOffset+containerStorageBytes], container[:containerStorageBytes])
		}
		return uint32(cardinality)
	}
}

// returns the cardinality of the xor.
// this allows for the caller to remove it from the header if necessary.

func (bitmap *RoaringBitmap) andNotContainerAtIndex(pos1 uint32, data []byte, offset uint32, cardinalityMinusOne uint16) uint32 {
	if cardinalityMinusOne < arrayDefaultMaxSize {
		return bitmap.andNotContainerAgainstArrayAtIndex(pos1, data, offset, cardinalityMinusOne+1)
	} else {
		return bitmap.andNotContainerAgainstBitmapAtIndex(pos1, data, offset)
	}
}

func (bitmap *RoaringBitmap) andNotContainerAgainstArrayAtIndex(pos1 uint32, data []byte, offset uint32, shorts uint16) uint32 {
	card := bitmap.getCardinalityMinusOneFromContainerIndex(pos1)
	//TODO: this stuff could be passed into the function, as it is always known.
	key := bitmap.getKeyAtContainerIndex(pos1)
	totalOffset := uint32(key) * containerStorageBytes
	if card < arrayDefaultMaxSize {
		// two array containers, difference into the bitmap. Difference always shrinks, so you'll have room.
		cardinality := byteBackedDifference(bitmap.data, totalOffset, uint32(card+1), data, offset, uint32(shorts), bitmap.data, totalOffset)
		return cardinality
	} else {
		removed := uint32(0)
		for i := uint32(0); i < uint32(shorts); i++ {
			s := ReadSingleShort(data, offset+2*i)
			currentByte := bitmap.data[totalOffset+(uint32(s)>>3)]
			if (currentByte & (1 << (s & 7))) > 0 {
				bitmap.data[totalOffset+(uint32(s)>>3)] &^= 1 << (s & 7)
				removed++
			}
		}
		newSize := uint32(card) + 1 - removed
		// need to switch to an array
		if newSize <= arrayDefaultMaxSize {
			// use an array.
			container := [8192]byte{}
			pos := uint32(0)
			base := 0
			for k := uint32(0); k < 1024 && pos < newSize; k++ {
				bitset := *(*uint64)(unsafe.Pointer(&bitmap.data[totalOffset+8*k]))
				for bitset != 0 {
					t := bitset & -bitset
					s := uint16(base + bits.OnesCount64(t-1))
					*(*uint16)(unsafe.Pointer(&container[2*pos])) = s
					pos++
					bitset ^= t
				}
				base += 64
			}
		}
		return newSize
	}
}

func (bitmap *RoaringBitmap) andNotContainerAgainstBitmapAtIndex(pos1 uint32, data []byte, offset uint32) uint32 {
	card := bitmap.getCardinalityMinusOneFromContainerIndex(pos1)
	//TODO: this stuff could be passed into the function, as it is always known.
	key := bitmap.getKeyAtContainerIndex(pos1)
	totalOffset := uint32(key) * containerStorageBytes
	if card < arrayDefaultMaxSize {
		// sweet, array container against a bitmap, can just walk up it.
		written := uint32(0)
		for i := uint32(0); i <= uint32(card); i++ {
			s := ReadSingleShort(bitmap.data, totalOffset+2*i)
			if data[offset+uint32(s>>3)]&(1<<(s&7)) == 0 {
				WriteShort(bitmap.data, totalOffset+2*written, s)
				written++
			}
		}
		return written
	} else {
		cardinality := 0
		// intersect in place, computing cardinality
		for i := uint32(0); i < 1024; i++ {
			*(*uint64)(unsafe.Pointer(&bitmap.data[totalOffset+8*i])) &^=
				*(*uint64)(unsafe.Pointer(&data[offset+8*i]))
			cardinality += bits.OnesCount64(*(*uint64)(unsafe.Pointer(&bitmap.data[totalOffset+8*i])))
		}
		// gotta switch it to an array container
		// write to a tmp array.
		if cardinality <= arrayDefaultMaxSize {
			container := [8192]byte{}
			pos := uint32(0)
			base := 0
			for k := uint32(0); k < 1024 && pos < uint32(cardinality); k++ {
				bitset := *(*uint64)(unsafe.Pointer(&bitmap.data[totalOffset+8*k]))
				for bitset != 0 {
					t := bitset & -bitset
					s := uint16(base + bits.OnesCount64(t-1))
					*(*uint16)(unsafe.Pointer(&container[2*pos])) = s
					pos++
					bitset ^= t
				}
				base += 64
			}
			copy(bitmap.data[totalOffset:], container[:containerStorageBytes])
		}
		return uint32(cardinality)
	}
}
func (bitmap *RoaringBitmap) xOrContainerAtIndex(key uint32, pos1 uint32, data []byte, offset uint32, cardinalityMinusOne uint16) uint32 {
	if cardinalityMinusOne < arrayDefaultMaxSize {
		return bitmap.xOrContainerAgainstArrayAtIndex(key, pos1, data, offset, cardinalityMinusOne+1)
	} else {
		return bitmap.xOrContainerAgainstBitmapAtIndex(pos1, data, offset, uint32(cardinalityMinusOne)+1)
	}
}

func (bitmap *RoaringBitmap) xOrContainerAgainstArrayAtIndex(key uint32, pos1 uint32, data []byte, offset uint32, shorts uint16) uint32 {
	card := bitmap.getCardinalityMinusOneFromContainerIndex(pos1)
	totalOffset := key * containerStorageBytes
	if card < arrayDefaultMaxSize {
		if shorts+card+1 <= arrayDefaultMaxSize {
			readOffset := totalOffset + 2*uint32(shorts)
			copy(bitmap.data[totalOffset+2*uint32(shorts):], bitmap.data[totalOffset:totalOffset+2*uint32(card)+2])

			// there'll likely be dirty bytes after the union, but we'll know we can use them.
			return exclusiveUnion2By2(bitmap.data, totalOffset, bitmap.data, readOffset, data, offset, uint32(shorts), uint32(card+1))
		}
		// gonna preemptively start a bitmap.
		tmp := [8192]byte{}
		for i := uint32(0); i < uint32(shorts); i++ {
			s := ReadSingleShort(data, offset+2*i)
			tmp[s/8] ^= 1 << (s % 8)
		}
		cardinality := shorts
		for i := uint32(0); i < uint32(card)+1; i++ {
			s := ReadSingleShort(bitmap.data, totalOffset+2*i)
			tmp[s/8] ^= 1 << (s % 8)
			if (tmp[s/8]>>(s%8))&1 == 1 {
				cardinality++
			} else {
				cardinality--
			}
		}
		// gotta switch it to an array
		// write to a tmp array.
		// TODO: factor out this common code, maybe?
		if cardinality <= arrayDefaultMaxSize {
			pos := uint32(0)
			base := 0
			for k := uint32(0); k < bitmapLongCount && pos < uint32(cardinality); k++ {
				bitset := *(*uint64)(unsafe.Pointer(&tmp[8*k]))
				for bitset != 0 {
					t := bitset & -bitset
					s := uint16(base + bits.OnesCount64(t-1))
					*(*uint16)(unsafe.Pointer(&bitmap.data[totalOffset+2*pos])) = s
					pos++
					bitset ^= t
				}
				base += 64
			}
		} else {
			copy(bitmap.data[totalOffset:], tmp[0:])
			calculatedCard := bitmapAndCardinality(bitmap.data, totalOffset, bitmap.data, totalOffset)
			if calculatedCard != uint32(cardinality) {
				panic("oops")
			}
		}
		return uint32(cardinality)
	} else {
		// bitmap's container is a bitmap container.
		// xor the bits, then check for conversion.
		cardinality := uint32(card) + 1
		for i := uint32(0); i < uint32(shorts); i++ {
			s := ReadSingleShort(data, offset+2*i)
			bitmap.data[totalOffset+uint32(s)/8] ^= 1 << (s % 8)
			currentByte := bitmap.data[totalOffset+uint32(s)/8]
			if (currentByte>>(s%8))&1 == 0 {
				cardinality--
			} else {
				cardinality++
			}
		}
		// if we dropped below 4096
		// then have to switch to an array container
		// write to a tmp byte array.
		if cardinality <= arrayDefaultMaxSize {
			container := [containerStorageBytes]byte{}
			pos := uint32(0)
			base := 0
			for k := uint32(0); k < bitmapLongCount && pos < uint32(cardinality); k++ {
				bitset := *(*uint64)(unsafe.Pointer(&bitmap.data[totalOffset+8*k]))
				for bitset != 0 {
					t := bitset & -bitset
					s := uint16(base + bits.OnesCount64(t-1))
					*(*uint16)(unsafe.Pointer(&container[2*pos])) = s
					pos++
					bitset ^= t
				}
				base += 64
			}
			copy(bitmap.data[totalOffset:totalOffset+containerStorageBytes], container[0:])
		}
		return cardinality
	}
}

func (bitmap *RoaringBitmap) xOrContainerAgainstBitmapAtIndex(containerIndex uint32, data []byte, offset uint32, bitmapCard uint32) uint32 {
	card := bitmap.getCardinalityMinusOneFromContainerIndex(containerIndex)
	//TODO: this stuff could be passed into the function, as it is always known.
	key := bitmap.getKeyAtContainerIndex(containerIndex)
	totalOffset := uint32(key) * containerStorageBytes
	if card < arrayDefaultMaxSize {
		tmp := [containerStorageBytes]byte{}
		copy(tmp[0:], data[offset:offset+containerStorageBytes])
		for i := uint32(0); i <= uint32(card); i++ {
			s := ReadSingleShort(bitmap.data, totalOffset+2*i)
			tmp[uint32(s>>3)] ^= 1 << (s & 7)
			if tmp[uint32(s>>3)]&(1<<(s&7)) > 0 {
				bitmapCard++
			} else {
				bitmapCard--
			}
		}
		if bitmapCard > arrayDefaultMaxSize {
			copy(bitmap.data[totalOffset:], tmp[0:])
			calculatedCard := bitmapAndCardinality(bitmap.data, totalOffset, bitmap.data, totalOffset)
			if calculatedCard != uint32(bitmapCard) {
				panic("oops")
			}
			return bitmapCard
		}
		pos := uint32(0)
		base := 0
		for k := uint32(0); k < 1024 && pos < bitmapCard; k++ {
			bitset := *(*uint64)(unsafe.Pointer(&tmp[8*k]))
			for bitset != 0 {
				t := bitset & -bitset
				s := uint16(base + bits.OnesCount64(t-1))
				*(*uint16)(unsafe.Pointer(&bitmap.data[totalOffset+2*pos])) = s
				pos++
				bitset ^= t
			}
			base += 64
		}
		return bitmapCard
	} else {
		cardinality := 0
		for i := uint32(0); i < 1024; i++ {
			*(*uint64)(unsafe.Pointer(&bitmap.data[totalOffset+8*i])) ^=
				*(*uint64)(unsafe.Pointer(&data[offset+8*i]))
			cardinality += bits.OnesCount64(*(*uint64)(unsafe.Pointer(&bitmap.data[totalOffset+8*i])))
		}
		// gotta switch it to an array
		// write to a tmp array.
		// TODO: factor out this common code, maybe?
		if cardinality <= arrayDefaultMaxSize {
			container := [containerStorageBytes]byte{}
			pos := uint32(0)
			base := 0
			for k := uint32(0); k < bitmapLongCount && pos < uint32(cardinality); k++ {
				bitset := *(*uint64)(unsafe.Pointer(&bitmap.data[totalOffset+8*k]))
				for bitset != 0 {
					t := bitset & -bitset
					s := uint16(base + bits.OnesCount64(t-1))
					*(*uint16)(unsafe.Pointer(&container[2*pos])) = s
					pos++
					bitset ^= t
				}
				base += 64
			}
			copy(bitmap.data[totalOffset:totalOffset+containerStorageBytes], container[:containerStorageBytes])
		}
		return uint32(cardinality)
	}
}

// returns the cardinality of the xor.
// this allows for the caller to remove it from the header if necessary.
func (bitmap *RoaringBitmap) computeAndAgainst(other *RoaringBitmap) {
	pos1 := uint32(0)
	pos2 := uint32(0)
	intersectionsize := uint32(0)
	length1 := bitmap.getContainerCount()
	length2 := other.getContainerCount()

main:
	for {
		if pos1 < length1 && pos2 < length2 {
			s1 := bitmap.getKeyAtContainerIndex(pos1)
			s2 := other.getKeyAtContainerIndex(pos2)
			for {
				if s1 == s2 {
					cardShort := other.getCardinalityMinusOneFromContainerIndex(pos2)
					offset := other.getOffsetForKeyAtPosition(uint32(s2), pos2)
					intersectionCard := bitmap.andContainerAtIndex(uint32(s1), pos1, other.data, offset, cardShort)
					if intersectionCard > 0 {
						// the  offset never changes, just the cardinality.
						if intersectionsize < pos1 {
							// the headers moved, write the new key
							WriteShort(bitmap.header, bytesPerContainer*intersectionsize, s1)
						}
						WriteShort(bitmap.header, bytesPerContainer*intersectionsize+cardinalityIncrement, uint16(intersectionCard-1))
						intersectionsize++
					}
					pos1++
					pos2++
					if pos1 == length1 || pos2 == length2 {
						break main
					}
					s1 = bitmap.getKeyAtContainerIndex(pos1)
					s2 = other.getKeyAtContainerIndex(pos2)
				} else if s1 < s2 {
					// TODO:  this isn't as fast as highlowcontainer.advanceUntil()
					//        which does a fancy binary search. Port that over.
					for s1 < s2 {
						pos1++
						if pos1 == length1 {
							break main
						}
						s1 = bitmap.getKeyAtContainerIndex(pos1)
					}
				} else {
					for s1 > s2 {
						pos2++
						if pos2 == length2 {
							break main
						}
						s2 = other.getKeyAtContainerIndex(pos2)
					}
				}
			}
		} else {
			break
		}
	}
	bitmap.containers = intersectionsize
}

func (bitmap *RoaringBitmap) computeAndNotAgainst(other *RoaringBitmap) {
	pos1 := uint32(0)
	pos2 := uint32(0)
	intersectionsize := uint32(0)
	length1 := bitmap.getContainerCount()
	length2 := other.getContainerCount()
main:
	for {
		if pos1 < length1 && pos2 < length2 {
			s1 := bitmap.getKeyAtContainerIndex(pos1)
			s2 := other.getKeyAtContainerIndex(pos2)
			for {
				if s1 == s2 {
					cardShort := other.getCardinalityMinusOneFromContainerIndex(pos2)
					offset := other.getOffsetForKeyAtPosition(uint32(s2), pos2)
					intersectionCard := bitmap.andNotContainerAtIndex(pos1, other.data, offset, cardShort)
					if intersectionCard > 0 {
						// the  offset never changes, just the cardinality.
						if intersectionsize != pos1 {
							// the headers moved, write the new key
							WriteShort(bitmap.header, bytesPerContainer*intersectionsize, s1)
						}
						WriteShort(bitmap.header, bytesPerContainer*intersectionsize+cardinalityIncrement, uint16(intersectionCard-1))
						intersectionsize++
					}
					pos1++
					pos2++
					if pos1 == length1 || pos2 == length2 {
						break main
					}
					s1 = bitmap.getKeyAtContainerIndex(pos1)
					s2 = other.getKeyAtContainerIndex(pos2)
				} else if s1 < s2 {
					if pos1 != intersectionsize {
						// need to copy 4 bytes. Is this faster than copy?
						WriteInt(bitmap.header, bytesPerContainer*intersectionsize,
							ReadSingleInt(bitmap.header, bytesPerContainer*pos1))
					}
					intersectionsize++
					pos1++
					if pos1 == length1 {
						break main
					}
					s1 = bitmap.getKeyAtContainerIndex(pos1)
				} else {
					for s1 > s2 {
						pos2++
						if pos2 == length2 {
							break main
						}
						s2 = other.getKeyAtContainerIndex(pos2)
					}
				}
			}
		} else {
			break
		}
	}
	if pos1 < length1 {
		if intersectionsize != pos1 {
			copy(bitmap.header[bytesPerContainer*intersectionsize:bytesPerContainer*(intersectionsize+length1-pos1)],
				bitmap.header[bytesPerContainer*pos1:bytesPerContainer*length1])
		}
		intersectionsize += length1 - pos1
	}
	bitmap.containers = intersectionsize
}

// Xor computes the symmetric difference between two bitmaps and stores the result in the current bitmap
func (rb *RoaringBitmap) computeXor(x2 *RoaringBitmap) {
	pos1 := uint32(0)
	pos2 := uint32(0)
	length1 := rb.getContainerCount()
	length2 := x2.getContainerCount()
	for {
		if (pos1 < length1) && (pos2 < length2) {
			s1 := rb.getKeyAtContainerIndex(pos1)
			s2 := x2.getKeyAtContainerIndex(pos2)
			if s1 < s2 {

				pos1++
				if pos1 == length1 {
					break
				}
				// TODO: binary advance
				s1 = rb.getKeyAtContainerIndex(pos1)
			} else if s1 > s2 {
				cardShort := x2.getCardinalityMinusOneFromContainerIndex(pos2)
				length := lengthFromCardinalityShort(cardShort)
				offset := x2.getOffsetForKeyAtPosition(uint32(s2), pos2)
				rb.insertNewContainerAtIndex(pos1, s2, cardShort, x2.data, offset, length)
				pos1++
				length1++
				pos2++
			} else {
				cardShort := x2.getCardinalityMinusOneFromContainerIndex(pos2)
				offset := x2.getOffsetForKeyAtPosition(uint32(s2), pos2)
				intersectionCard := rb.xOrContainerAtIndex(uint32(s1), pos1, x2.data, offset, cardShort)
				if intersectionCard > 0 {
					WriteShort(rb.header, bytesPerContainer*pos1+cardinalityIncrement, uint16(intersectionCard-1))
					pos1++
				} else {
					length1--
				}
				pos2++
			}
		} else {
			break
		}
	}
	if pos1 == length1 {
		for pos2 < length2 {
			s2 := x2.getKeyAtContainerIndex(pos2)
			cardShort := x2.getCardinalityMinusOneFromContainerIndex(pos2)
			length := lengthFromCardinalityShort(cardShort)
			offset := x2.getOffsetForKeyAtPosition(uint32(s2), pos2)
			rb.insertNewContainerAtIndex(pos1, s2, cardShort, x2.data, offset, length)
			pos1++
			length1++
			pos2++
		}
	}
	rb.containers = length1
}

func intersection2By2(data1 []byte, offset1 uint32, shorts1 uint32,
	data2 []byte, offset2 uint32, shorts2 uint32,
	buffer []byte, bufferOffset uint32) uint32 {
	if 0 == shorts1 || 0 == shorts2 {
		return 0
	}

	k1 := uint32(0)
	k2 := uint32(0)
	pos := uint32(0)
	s1 := ReadSingleShort(data1, offset1)
	s2 := ReadSingleShort(data2, offset2)
mainwhile:
	for {
		if s2 < s1 {
			for {
				k2++
				if k2 == shorts2 {
					break mainwhile
				}
				s2 = ReadSingleShort(data2, offset2+2*k2)
				if s2 >= s1 {
					break
				}
			}
		}
		if s1 < s2 {
			for {
				k1++
				if k1 == shorts1 {
					break mainwhile
				}
				s1 = ReadSingleShort(data1, offset1+2*k1)
				if s1 >= s2 {
					break
				}
			}
		} else {
			WriteShort(buffer, bufferOffset+2*pos, s1)
			pos++
			k1++
			if k1 == shorts1 {
				break
			}
			s1 = ReadSingleShort(data1, offset1+2*k1)
			k2++
			if k2 == shorts2 {
				break
			}
			s2 = ReadSingleShort(data2, offset2+2*k2)
		}
	}
	return pos
}

func byteBackedDifference(data1 []byte, offset1 uint32, shorts1 uint32,
	data2 []byte, offset2 uint32, shorts2 uint32,
	buffer []byte, bufferOffset uint32) uint32 {
	if 0 == shorts2 {
		// if your buffer output is data1, no need to copy.
		if &buffer != &data1 || offset1 != bufferOffset {
			copy(buffer[bufferOffset:], data1[offset1:offset1+2*shorts1])
		}
		return shorts1
	}
	if 0 == shorts1 {
		return 0
	}
	pos := uint32(0)
	k1 := uint32(0)
	k2 := uint32(0)

	s1 := ReadSingleShort(data1, offset1)
	s2 := ReadSingleShort(data2, offset2)
	for {
		if s1 < s2 {
			WriteShort(buffer, bufferOffset+2*pos, s1)
			pos++
			k1++
			if k1 >= shorts1 {
				break
			}
			s1 = ReadSingleShort(data1, offset1+2*k1)
		} else if s1 == s2 {
			k1++
			k2++
			if k1 >= shorts1 {
				break
			}
			if k2 >= shorts2 {
				copy(buffer[bufferOffset+2*pos:], data1[offset1+2*k1:offset1+2*shorts1])
				pos += shorts1 - k1
				break
			}
			s1 = ReadSingleShort(data1, offset1+2*k1)
			s2 = ReadSingleShort(data2, offset2+2*k2)
		} else {
			k2++
			if k2 >= shorts2 {
				copy(buffer[bufferOffset+2*pos:], data1[offset1+2*k1:offset1+2*shorts1])
				pos += shorts1 - k1
				break
			}
			s2 = ReadSingleShort(data2, offset2+2*k2)
		}
	}
	return pos
}

// Destination is the bitmap we want the final result to be written to
// Source is to be read from.
// The algorithm is to do a zipper union into destination. You're guaranteed to have enough working space.
func union2By2(destinationWriteData []byte, destinationWriteStart uint32,
	destinationReadData []byte, destinationReadStart uint32,
	sourceData []byte, sourceStart uint32,
	sourceShortCount uint32, destinationShortCount uint32) uint32 {
	if sourceShortCount == 0 {
		return destinationShortCount
	}
	if destinationShortCount == 0 {
		return sourceShortCount
	}
	// we shifted the destination data up.
	k1 := uint32(0)
	k2 := uint32(0)
	sourceReadStart := sourceStart
	s1 := ReadSingleShort(destinationReadData, destinationReadStart+2*k1)
	s2 := ReadSingleShort(sourceData, sourceReadStart+2*k2)
	pos := uint32(0)
	for {
		if s1 < s2 {
			WriteShort(destinationWriteData, destinationWriteStart+2*pos, s1)
			pos++
			k1++
			if k1 >= destinationShortCount {
				copy(destinationWriteData[destinationWriteStart+2*pos:], sourceData[sourceReadStart+2*k2:sourceReadStart+2*sourceShortCount])
				pos += sourceShortCount - k2
				break
			}
			s1 = ReadSingleShort(destinationReadData, destinationReadStart+2*k1)
		} else if s1 == s2 {
			WriteShort(destinationWriteData, destinationWriteStart+2*pos, s1)
			pos++
			k1++
			k2++
			if k1 >= destinationShortCount {
				copy(destinationWriteData[destinationWriteStart+2*pos:], sourceData[sourceReadStart+2*k2:sourceReadStart+2*sourceShortCount])
				pos += sourceShortCount - k2
				break
			}
			if k2 >= sourceShortCount {
				copy(destinationWriteData[destinationWriteStart+2*pos:], destinationReadData[destinationReadStart+2*k1:destinationReadStart+2*destinationShortCount])
				pos += destinationShortCount - k1
				break
			}
			s1 = ReadSingleShort(destinationReadData, destinationReadStart+2*k1)
			s2 = ReadSingleShort(sourceData, sourceReadStart+2*k2)
		} else {
			WriteShort(destinationWriteData, destinationWriteStart+2*pos, s2)
			pos++
			k2++
			if k2 >= sourceShortCount {
				copy(destinationWriteData[destinationWriteStart+2*pos:], destinationReadData[destinationReadStart+2*k1:destinationReadStart+2*destinationShortCount])
				pos += destinationShortCount - k1
				break
			}
			s2 = ReadSingleShort(sourceData, sourceReadStart+2*k2)
		}
	}
	return pos
}

func exclusiveUnion2By2(destinationWriteData []byte, destinationWriteStart uint32,
	destinationReadData []byte, destinationReadStart uint32,
	sourceData []byte, sourceStart uint32,
	sourceShortCount uint32, destinationShortCount uint32) uint32 {
	if sourceShortCount == 0 || destinationShortCount == 0 {
		panic("we don't support empty containers. bad!")
	}
	// we shifted the destination data up.
	k1 := uint32(0)
	k2 := uint32(0)
	sourceReadStart := sourceStart
	s1 := ReadSingleShort(destinationReadData, destinationReadStart+2*k1)
	s2 := ReadSingleShort(sourceData, sourceReadStart+2*k2)
	pos := uint32(0)
	for {
		if s1 < s2 {
			WriteShort(destinationWriteData, destinationWriteStart+2*pos, s1)
			pos++
			k1++
			if k1 >= destinationShortCount {
				copy(destinationWriteData[destinationWriteStart+2*pos:], sourceData[sourceReadStart+2*k2:sourceReadStart+2*sourceShortCount])
				pos += sourceShortCount - k2
				break
			}
			s1 = ReadSingleShort(destinationReadData, destinationReadStart+2*k1)
		} else if s1 == s2 {
			k1++
			k2++
			if k1 >= destinationShortCount {
				copy(destinationWriteData[destinationWriteStart+2*pos:], sourceData[sourceReadStart+2*k2:sourceReadStart+2*sourceShortCount])
				pos += sourceShortCount - k2
				break
			}
			if k2 >= sourceShortCount {
				copy(destinationWriteData[destinationWriteStart+2*pos:], destinationReadData[destinationReadStart+2*k1:destinationReadStart+2*destinationShortCount])
				pos += destinationShortCount - k1
				break
			}
			s1 = ReadSingleShort(destinationReadData, destinationReadStart+2*k1)
			s2 = ReadSingleShort(sourceData, sourceReadStart+2*k2)
		} else {
			WriteShort(destinationWriteData, destinationWriteStart+2*pos, s2)
			pos++
			k2++
			if k2 >= sourceShortCount {
				copy(destinationWriteData[destinationWriteStart+2*pos:], destinationReadData[destinationReadStart+2*k1:destinationReadStart+2*destinationShortCount])
				pos += destinationShortCount - k1
				break
			}
			s2 = ReadSingleShort(sourceData, sourceReadStart+2*k2)
		}
	}
	return pos
}

func byteBackedUnion2by2Cardinality(bytes1 []byte, bytes2 []byte,
	offset1, offset2, size1, size2 uint32) uint32 {
	pos := uint32(0)
	k1 := uint32(0)
	k2 := uint32(0)

	if size1 == 0 {
		return size2
	}
	if size2 == 0 {
		return size1
	}
	s1 := ReadSingleShort(bytes1, offset1+2*k1)
	s2 := ReadSingleShort(bytes2, offset2+2*k2)
	for {
		if s1 < s2 {
			pos++
			k1++
			if k1 >= size1 {
				pos += size2 - k2
				break
			}
			s1 = ReadSingleShort(bytes1, offset1+2*k1)
		} else if s1 == s2 {
			pos++
			k1++
			k2++
			if k1 >= size1 {
				pos += size2 - k2
				break
			}
			if k2 >= size2 {
				pos += size1 - k1
				break
			}
			s1 = ReadSingleShort(bytes1, offset1+2*k1)
			s2 = ReadSingleShort(bytes2, offset2+2*k2)
		} else { // if (set1[k1]>set2[k2])
			pos++
			k2++
			if k2 >= size2 {
				pos += size1 - k1
				break
			}
			s2 = ReadSingleShort(bytes2, offset2+2*k2)
		}
	}
	return pos
}

func offsetAndCardinality(data1 []byte, offset1 uint32, card1 uint16, data2 []byte, offset2 uint32, card2 uint16) uint32 {
	if card1 < arrayDefaultMaxSize && card2 < arrayDefaultMaxSize {
		// both arrays
		return arrayAndCardinality(data1, offset1, uint32(card1)+1, data2, offset2, uint32(card2)+1)

	} else if card1 >= arrayDefaultMaxSize && card2 >= arrayDefaultMaxSize {
		// both bitmaps
		return bitmapAndCardinality(data1, offset1, data2, offset2)
	} else if card1 < arrayDefaultMaxSize {
		return arrayAndBitmapCardinality(data1, offset1, uint32(card1)+1, data2, offset2)
	} else {
		return arrayAndBitmapCardinality(data2, offset2, uint32(card2)+1, data1, offset1)
	}
}

func arrayAndBitmapCardinality(data1 []byte, offset1 uint32, shorts uint32, data2 []byte, offset2 uint32) uint32 {
	count := uint32(0)
	for i := uint32(0); i < shorts; i++ {
		s := uint32(ReadSingleShort(data1, offset1+2*i))
		count += uint32(data2[offset2+(s>>3)]>>(s&7)) & 1
	}
	return count
}

func bitmapAndCardinality(data1 []byte, offset1 uint32, data2 []byte, offset2 uint32) uint32 {
	result := 0
	for i := uint32(0); i < 8192; i += 8 {
		result += bits.OnesCount64(ReadSingleLong(data1, offset1+i) & ReadSingleLong(data2, offset2+i))
	}
	return uint32(result)
}

func ReadSingleShort(data []byte, pointer uint32) uint16 {
	return *(*uint16)(unsafe.Pointer(&data[pointer]))
}

func ReadSingleInt(data []byte, pointer uint32) uint32 {
	return *(*uint32)(unsafe.Pointer(&data[pointer]))
}

func ReadSingleLong(data []byte, pointer uint32) uint64 {
	return *(*uint64)(unsafe.Pointer(&data[pointer]))
}

func WriteShort(data []byte, pointer uint32, val uint16) {
	*(*uint16)(unsafe.Pointer(&data[pointer])) = val
}

func WriteInt(data []byte, pointer uint32, val uint32) {
	pntr := (*uint32)(unsafe.Pointer(&data[pointer]))
	*pntr = val
}
