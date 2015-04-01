// Package roaring is an implementation of Roaring Bitmaps in Go.
// They provide fast compressed bitmap data structures (also called bitset).
// They are ideally suited to represent sets of integers over
// relatively small ranges.
// See http://roaringbitmap.org for details.
package roaring

import (
	"bytes"
	"io"
	"strconv"
)

// RoaringBitmap represents a compressed bitmap where you can add integers.
type RoaringBitmap struct {
	highlowcontainer roaringArray
}

// Write out a serialized version of this bitmap to stream
func (b *RoaringBitmap) WriteTo(stream io.Writer) (int, error) {
	return b.highlowcontainer.writeTo(stream)
}

// Read a serialized version of this bitmap from stream
func (b *RoaringBitmap) ReadFrom(stream io.Reader) (int, error) {
	return b.highlowcontainer.readFrom(stream)
}

// NewRoaringBitmap creates a new empty RoaringBitmap
func NewRoaringBitmap() *RoaringBitmap {
	return &RoaringBitmap{*newRoaringArray()}
}

// Clear removes all content from the RoaringBitmap and frees the memory
func (rb *RoaringBitmap) Clear() {
	rb.highlowcontainer = *newRoaringArray()
}

// ToArray creates a new slice containing all of the integers stored in the RoaringBitmap in sorted order
func (rb *RoaringBitmap) ToArray() []int {
	array := make([]int, rb.GetCardinality())
	pos := 0
	pos2 := 0

	for pos < rb.highlowcontainer.size() {
		hs := toIntUnsigned(rb.highlowcontainer.getKeyAtIndex(pos)) << 16
		c := rb.highlowcontainer.getContainerAtIndex(pos)
		pos++
		c.fillLeastSignificant16bits(array, pos2, hs)
		pos2 += c.getCardinality()
	}
	return array
}

// GetSizeInBytes estimates the memory usage of the RoaringBitmap. Note that this
// might differ slightly from the amount of bytes required for persistent storage
func (rb *RoaringBitmap) GetSizeInBytes() int {
	size := 8
	for i := 0; i < rb.highlowcontainer.size(); i++ {
		c := rb.highlowcontainer.getContainerAtIndex(i)
		size += 2 + c.getSizeInBytes()
	}
	return size
}

// GetSerializedSizeInBytes computes the serialized size in bytes  the RoaringBitmap. It should correspond to the
// number of bytes written when invoking WriteTo
func (rb *RoaringBitmap) GetSerializedSizeInBytes() int {
	return rb.highlowcontainer.serializedSizeInBytes()
}

// IntIterable allows you to iterate over the values in a RoaringBitmap
type IntIterable interface {
	HasNext() bool
	Next() int
}

type intIterator struct {
	pos              int
	hs               int
	iter             shortIterable
	highlowcontainer *roaringArray
}

// HasNext returns true if there are more integers to iterate over
func (ii *intIterator) HasNext() bool {
	return ii.pos < ii.highlowcontainer.size()
}

func (ii *intIterator) init() {
	if ii.highlowcontainer.size() > ii.pos {
		ii.iter = ii.highlowcontainer.getContainerAtIndex(ii.pos).getShortIterator()
		ii.hs = toIntUnsigned(ii.highlowcontainer.getKeyAtIndex(ii.pos)) << 16
	}
}

// Next returns the next integer
func (ii *intIterator) Next() int {
	x := toIntUnsigned(ii.iter.next()) | ii.hs
	if !ii.iter.hasNext() {
		ii.pos = ii.pos + 1
		ii.init()
	}
	return x
}

func newIntIterator(a *RoaringBitmap) *intIterator {
	p := new(intIterator)
	p.pos = 0
	p.highlowcontainer = &a.highlowcontainer
	p.init()
	return p
}

// String creates a string representation of the RoaringBitmap
func (rb *RoaringBitmap) String() string {
	// inspired by https://github.com/fzandona/goroar/blob/master/roaringbitmap.go
	var buffer bytes.Buffer
	start := []byte("{")
	buffer.Write(start)
	i := rb.Iterator()
	for i.HasNext() {
		buffer.WriteString(strconv.Itoa(int(i.Next())))
		if i.HasNext() { // todo: optimize
			buffer.WriteString(",")
		}
	}
	buffer.WriteString("}")
	return buffer.String()
}

// Iterator creates a new IntIterable to iterate over the integers contained in the bitmap, in sorted order
func (rb *RoaringBitmap) Iterator() IntIterable {
	return newIntIterator(rb)
}

// Clone creates a copy of the RoaringBitmap
func (rb *RoaringBitmap) Clone() *RoaringBitmap {
	ptr := new(RoaringBitmap)
	ptr.highlowcontainer = *rb.highlowcontainer.clone()
	return ptr
}

// Contains returns true if the integer is contained in the bitmap
func (rb *RoaringBitmap) Contains(x int) bool {
	hb := highbits(x)
	c := rb.highlowcontainer.getContainer(hb)
	return c != nil && c.contains(lowbits(x))

}

// Equals returns true if the two bitmaps contain the same integers
func (rb *RoaringBitmap) Equals(o interface{}) bool {
	srb, ok := o.(*RoaringBitmap)
	if ok {
		return srb.highlowcontainer.equals(rb.highlowcontainer)
	}
	return false
}

// Add the integer x to the bitmap
func (rb *RoaringBitmap) Add(x int) {
	hb := highbits(x)
	i := rb.highlowcontainer.getIndex(hb)
	if i >= 0 {
		rb.highlowcontainer.setContainerAtIndex(i, rb.highlowcontainer.getContainerAtIndex(i).add(lowbits(x)))
	} else {
		newac := newArrayContainer()
		rb.highlowcontainer.insertNewKeyValueAt(-i-1, hb, newac.add(lowbits(x)))
	}
}

// Remove the integer x from the bitmap
func (rb *RoaringBitmap) Remove(x int) {
	hb := highbits(x)
	i := rb.highlowcontainer.getIndex(hb)
	if i >= 0 {
		rb.highlowcontainer.setContainerAtIndex(i, rb.highlowcontainer.getContainerAtIndex(i).remove(lowbits(x)))
		if rb.highlowcontainer.getContainerAtIndex(i).getCardinality() == 0 {
			rb.highlowcontainer.removeAtIndex(i)
		}
	}
}

// IsEmpty returns true if the RoaringBitmap is empty (it is faster than doing (GetCardinality() == 0))
func (rb *RoaringBitmap) IsEmpty() bool {
	return rb.highlowcontainer.size() == 0
}

// GetCardinality returns the number of integers contained in the bitmap
func (rb *RoaringBitmap) GetCardinality() int {
	size := 0
	for i := 0; i < rb.highlowcontainer.size(); i++ {
		size += rb.highlowcontainer.getContainerAtIndex(i).getCardinality()
	}
	return size
}

// Rank returns the number of integers that are smaller or equal to x (Rank(infinity) would be GetCardinality())
func (rb *RoaringBitmap) Rank(x int) int {
	size := 0
	for i := 0; i < rb.highlowcontainer.size(); i++ {
		key := rb.highlowcontainer.getKeyAtIndex(i)
		if key > highbits(x) {
			return size
		}
		if key < highbits(x) {
			size += rb.highlowcontainer.getContainerAtIndex(i).getCardinality()
		} else {
			return size + rb.highlowcontainer.getContainerAtIndex(i).rank(lowbits(x))
		}
	}
	return size
}

// And computes the intersection between two bitmaps and store the result in the current bitmap
func (rb *RoaringBitmap) And(x2 *RoaringBitmap) *RoaringBitmap {
	results := And(rb, x2) // Todo: could be computed in-place for reduced memory usage
	rb.highlowcontainer = results.highlowcontainer
	return rb
}

// Xor computes the symmetric difference between two bitmaps and store the result in the current bitmap
func (rb *RoaringBitmap) Xor(x2 *RoaringBitmap) *RoaringBitmap {
	results := Xor(rb, x2) // Todo: could be computed in-place for reduced memory usage
	rb.highlowcontainer = results.highlowcontainer
	return rb
}

// Or computes the union between two bitmaps and store the result in the current bitmap
func (rb *RoaringBitmap) Or(x2 *RoaringBitmap) *RoaringBitmap {
	results := Or(rb, x2) // Todo: could be computed in-place for reduced memory usage
	rb.highlowcontainer = results.highlowcontainer
	return rb
}

// AndNot computes the difference between two bitmaps and store the result in the current bitmap
func (rb *RoaringBitmap) AndNot(x2 *RoaringBitmap) *RoaringBitmap {
	results := AndNot(rb, x2) // Todo: could be computed in-place for reduced memory usage
	rb.highlowcontainer = results.highlowcontainer
	return rb
}

// Or computes the union between two bitmaps and returns the result
func Or(x1, x2 *RoaringBitmap) *RoaringBitmap {
	answer := NewRoaringBitmap()
	pos1 := 0
	pos2 := 0
	length1 := x1.highlowcontainer.size()
	length2 := x2.highlowcontainer.size()
main:
	for {
		if (pos1 < length1) && (pos2 < length2) {
			s1 := x1.highlowcontainer.getKeyAtIndex(pos1)
			s2 := x2.highlowcontainer.getKeyAtIndex(pos2)

			for {
				if s1 < s2 {
					answer.highlowcontainer.appendCopy(x1.highlowcontainer, pos1)
					pos1++
					if pos1 == length1 {
						break main
					}
					s1 = x1.highlowcontainer.getKeyAtIndex(pos1)
				} else if s1 > s2 {
					answer.highlowcontainer.appendCopy(x2.highlowcontainer, pos2)
					pos2++
					if pos2 == length2 {
						break main
					}
					s2 = x2.highlowcontainer.getKeyAtIndex(pos2)
				} else {

					answer.highlowcontainer.append(s1, x1.highlowcontainer.getContainerAtIndex(pos1).or(x2.highlowcontainer.getContainerAtIndex(pos2)))
					pos1++
					pos2++
					if (pos1 == length1) || (pos2 == length2) {
						break main
					}
					s1 = x1.highlowcontainer.getKeyAtIndex(pos1)
					s2 = x2.highlowcontainer.getKeyAtIndex(pos2)
				}
			}
		} else {
			break
		}
	}
	if pos1 == length1 {
		answer.highlowcontainer.appendCopyMany(x2.highlowcontainer, pos2, length2)
	} else if pos2 == length2 {
		answer.highlowcontainer.appendCopyMany(x1.highlowcontainer, pos1, length1)
	}
	return answer
}

// And computes the intersection between two bitmaps and returns the result
func And(x1, x2 *RoaringBitmap) *RoaringBitmap {
	answer := NewRoaringBitmap()
	pos1 := 0
	pos2 := 0
	length1 := x1.highlowcontainer.size()
	length2 := x2.highlowcontainer.size()
main:
	for {
		if pos1 < length1 && pos2 < length2 {
			s1 := x1.highlowcontainer.getKeyAtIndex(pos1)
			s2 := x2.highlowcontainer.getKeyAtIndex(pos2)
			for {
				if s1 == s2 {
					C := x1.highlowcontainer.getContainerAtIndex(pos1)
					C = C.and(x2.highlowcontainer.getContainerAtIndex(pos2))

					if C.getCardinality() > 0 {
						answer.highlowcontainer.append(s1, C)
					}
					pos1++
					pos2++
					if (pos1 == length1) || (pos2 == length2) {
						break main
					}
					s1 = x1.highlowcontainer.getKeyAtIndex(pos1)
					s2 = x2.highlowcontainer.getKeyAtIndex(pos2)
				} else if s1 < s2 {
					pos1 = x1.highlowcontainer.advanceUntil(s2, pos1)
					if pos1 == length1 {
						break main
					}
					s1 = x1.highlowcontainer.getKeyAtIndex(pos1)
				} else { // s1 > s2
					pos2 = x2.highlowcontainer.advanceUntil(s1, pos2)
					if pos2 == length2 {
						break main
					}
					s2 = x2.highlowcontainer.getKeyAtIndex(pos2)
				}
			}
		} else {
			break
		}
	}
	return answer
}

// Xor computes the symmetric difference between two bitmaps and returns the result
func Xor(x1, x2 *RoaringBitmap) *RoaringBitmap {
	answer := NewRoaringBitmap()
	pos1 := 0
	pos2 := 0
	length1 := x1.highlowcontainer.size()
	length2 := x2.highlowcontainer.size()

main:
	for {
		if (pos1 < length1) && (pos2 < length2) {
			s1 := x1.highlowcontainer.getKeyAtIndex(pos1)
			s2 := x2.highlowcontainer.getKeyAtIndex(pos2)
			if s1 < s2 {
				answer.highlowcontainer.appendCopy(x1.highlowcontainer, pos1)
				pos1++
				if pos1 == length1 {
					break main
				}
				s1 = x1.highlowcontainer.getKeyAtIndex(pos1)
			} else if s1 > s2 {
				answer.highlowcontainer.appendCopy(x2.highlowcontainer, pos2)
				pos2++
				if pos2 == length2 {
					break main
				}
				s2 = x2.highlowcontainer.getKeyAtIndex(pos2)
			} else {
				c := x1.highlowcontainer.getContainerAtIndex(pos1).xor(x2.highlowcontainer.getContainerAtIndex(pos2))
				if c.getCardinality() > 0 {
					answer.highlowcontainer.append(s1, c)
				}
				pos1++
				pos2++
				if (pos1 == length1) || (pos2 == length2) {
					break main
				}
				s1 = x1.highlowcontainer.getKeyAtIndex(pos1)
				s2 = x2.highlowcontainer.getKeyAtIndex(pos2)
			}
		} else {
			break
		}
	}
	if pos1 == length1 {
		answer.highlowcontainer.appendCopyMany(x2.highlowcontainer, pos2, length2)
	} else if pos2 == length2 {
		answer.highlowcontainer.appendCopyMany(x1.highlowcontainer, pos1, length1)
	}
	return answer
}

// AndNot computes the difference between two bitmaps and returns the result
func AndNot(x1, x2 *RoaringBitmap) *RoaringBitmap {
	answer := NewRoaringBitmap()
	pos1 := 0
	pos2 := 0
	length1 := x1.highlowcontainer.size()
	length2 := x2.highlowcontainer.size()

main:
	for {
		if pos1 < length1 && pos2 < length2 {
			s1 := x1.highlowcontainer.getKeyAtIndex(pos1)
			s2 := x2.highlowcontainer.getKeyAtIndex(pos2)
			for {
				if s1 < s2 {
					answer.highlowcontainer.appendCopy(x1.highlowcontainer, pos1)
					pos1++
					if pos1 == length1 {
						break main
					}
					s1 = x1.highlowcontainer.getKeyAtIndex(pos1)
				} else if s1 > s2 {
					pos2++
					if pos2 == length2 {
						break main
					}
					s2 = x2.highlowcontainer.getKeyAtIndex(pos2)
				} else {
					C := x1.highlowcontainer.getContainerAtIndex(pos1)
					C.andNot(x2.highlowcontainer.getContainerAtIndex(pos2))
					if C.getCardinality() > 0 {
						answer.highlowcontainer.append(s1, C)
					}
					pos1++
					pos2++
					if (pos1 == length1) || (pos2 == length2) {
						break main
					}
					s1 = x1.highlowcontainer.getKeyAtIndex(pos1)
					s2 = x2.highlowcontainer.getKeyAtIndex(pos2)
				}
			}
		} else {
			break
		}
	}
	if pos2 == length2 {
		answer.highlowcontainer.appendCopyMany(x1.highlowcontainer, pos1, length1)
	}
	return answer
}

// BitmapOf generates a new bitmap filled with the specified integer
func BitmapOf(dat ...int) *RoaringBitmap {
	ans := NewRoaringBitmap()
	for _, i := range dat {
		ans.Add(i)
	}
	return ans
}

// Flip negates the bits in the given range, any integer present in this range and in the bitmap is removed,
// and any integer present in the range and not in the bitmap is added
func (rb *RoaringBitmap) Flip(rangeStart, rangeEnd int) *RoaringBitmap {
	results := Flip(rb, rangeStart, rangeEnd) //Todo: the computation could be done in-place to reduce memory usage
	rb.highlowcontainer = results.highlowcontainer
	return rb
}

// Flip negates the bits in the given range, any integer present in this range and in the bitmap is removed,
// and any integer present in the range and not in the bitmap is added, a new bitmap is returned leaving
// the current bitmap unchanged
func Flip(bm *RoaringBitmap, rangeStart, rangeEnd int) *RoaringBitmap {
	if rangeStart >= rangeEnd {
		return bm.Clone()
	}

	answer := NewRoaringBitmap()
	hbStart := highbits(rangeStart)
	lbStart := lowbits(rangeStart)
	hbLast := highbits(rangeEnd - 1)
	lbLast := lowbits(rangeEnd - 1)

	// copy the containers before the active area
	answer.highlowcontainer.appendCopiesUntil(bm.highlowcontainer, hbStart)

	max := toIntUnsigned(maxLowBit())
	for hb := hbStart; hb <= hbLast; hb++ {
		containerStart := 0
		if hb == hbStart {
			containerStart = toIntUnsigned(lbStart)
		}
		containerLast := max
		if hb == hbLast {
			containerLast = toIntUnsigned(lbLast)
		}

		i := bm.highlowcontainer.getIndex(hb)
		j := answer.highlowcontainer.getIndex(hb)

		if i >= 0 {
			c := bm.highlowcontainer.getContainerAtIndex(i).not(containerStart, containerLast)
			if c.getCardinality() > 0 {
				answer.highlowcontainer.insertNewKeyValueAt(-j-1, hb, c)
			}

		} else { // *think* the range of ones must never be
			// empty.
			answer.highlowcontainer.insertNewKeyValueAt(-j-1, hb,
				rangeOfOnes(containerStart, containerLast))
		}
	}
	// copy the containers after the active area.
	answer.highlowcontainer.appendCopiesAfter(bm.highlowcontainer, hbLast)

	return answer
}
