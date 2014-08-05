// roaring is an implementation of Roaring Bitmaps in Go. See http://roaringbitmap.org for details.
package roaring

import (
	"bytes"
	"strconv"
)

type RoaringBitmap struct {
	highlowcontainer roaringArray
}

func NewRoaringBitmap() *RoaringBitmap {
	return &RoaringBitmap{*newRoaringArray()}
}

func (self *RoaringBitmap) Clear() {
	self.highlowcontainer = *newRoaringArray()
}

func (self *RoaringBitmap) ToArray() []int {
	array := make([]int, self.GetCardinality())
	pos := 0
	pos2 := 0

	for pos < self.highlowcontainer.size() {
		hs := toIntUnsigned(self.highlowcontainer.getKeyAtIndex(pos)) << 16
		c := self.highlowcontainer.getContainerAtIndex(pos)
		pos++
		c.fillLeastSignificant16bits(array, pos2, hs)
		pos2 += c.getCardinality()
	}
	return array
}

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

func (self *intIterator) HasNext() bool {
	return self.pos < self.highlowcontainer.size()
}

func (self *intIterator) init() {
	if self.highlowcontainer.size() > self.pos {
		self.iter = self.highlowcontainer.getContainerAtIndex(self.pos).getShortIterator()
		self.hs = toIntUnsigned(self.highlowcontainer.getKeyAtIndex(self.pos)) << 16
	}
}

func (self *intIterator) Next() int {
	x := toIntUnsigned(self.iter.next()) | self.hs
	if !self.iter.hasNext() {
		self.pos = self.pos + 1
		self.init()
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

func (self *RoaringBitmap) Iterator() IntIterable {
	return newIntIterator(self)
}

func (self *RoaringBitmap) Clone() *RoaringBitmap {
	ptr := new(RoaringBitmap)
	ptr.highlowcontainer = *self.highlowcontainer.clone()
	return ptr
}

func (self *RoaringBitmap) Contains(x int) bool {
	hb := highbits(x)
	c := self.highlowcontainer.getContainer(hb)
	return c != nil && c.contains(lowbits(x))

}

func (self *RoaringBitmap) Equals(o interface{}) bool {
	srb := o.(*RoaringBitmap)
	if srb != nil {
		return srb.highlowcontainer.equals(self.highlowcontainer)
	}
	return false
}

func (self *RoaringBitmap) Add(x int) {
	hb := highbits(x)
	i := self.highlowcontainer.getIndex(hb)
	if i >= 0 {
		self.highlowcontainer.setContainerAtIndex(i, self.highlowcontainer.getContainerAtIndex(i).add(lowbits(x)))
	} else {
		newac := newArrayContainer()
		self.highlowcontainer.insertNewKeyValueAt(-i-1, hb, newac.add(lowbits(x)))
	}
}

func (self *RoaringBitmap) GetCardinality() int {
	size := 0
	for i := 0; i < self.highlowcontainer.size(); i++ {
		size += self.highlowcontainer.getContainerAtIndex(i).getCardinality()
	}
	return size
}

func (self *RoaringBitmap) And(x2 *RoaringBitmap) *RoaringBitmap {
	results := And(self, x2)
	self.highlowcontainer = results.highlowcontainer
	return self
}

func (self *RoaringBitmap) Xor(x2 *RoaringBitmap) *RoaringBitmap {
	results := Xor(self, x2)
	self.highlowcontainer = results.highlowcontainer
	return self
}

func (self *RoaringBitmap) Or(x2 *RoaringBitmap) *RoaringBitmap {
	results := Or(self, x2)
	self.highlowcontainer = results.highlowcontainer
	return self
}

func (self *RoaringBitmap) AndNot(x2 *RoaringBitmap) *RoaringBitmap {
	results := AndNot(self, x2)
	self.highlowcontainer = results.highlowcontainer
	return self
}

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
				if s1 < s2 {
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

func BitmapOf(dat ...int) *RoaringBitmap {
	ans := NewRoaringBitmap()
	for _, i := range dat {
		ans.Add(i)
	}
	return ans
}

func (self *RoaringBitmap) Flip(rangeStart, rangeEnd int) *RoaringBitmap {
	results := Flip(self, rangeStart, rangeEnd)
	self.highlowcontainer = results.highlowcontainer
	return self
}

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
