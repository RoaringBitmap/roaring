package goroaring

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
	GetShortIterator() ShortIterable
	Contains(i short) bool
	Equals(i interface{}) bool
	FillLeastSignificant16bits(array []int, i, mask int)
	Or(r Container) Container
	//	ToArray() []int
}

type RoaringBitmap struct {
	highlowcontainer RoaringArray
}

func NewRoaringBitmap() *RoaringBitmap {
	return &RoaringBitmap{*NewRoaringArray()}
}

func (self *RoaringBitmap) Clear() {
	self.highlowcontainer = *NewRoaringArray()
}

func (self *RoaringBitmap) ToArray() []int {
	array := make([]int, self.GetCardinality())

	pos := 0
	pos2 := 0
	for pos < self.highlowcontainer.Size() {
		hs := ToIntUnsigned(self.highlowcontainer.GetKeyAtIndex(pos)) << 16
		c := self.highlowcontainer.GetContainerAtIndex(pos)
		pos++
		c.FillLeastSignificant16bits(array, pos2, hs)
		pos2 += c.GetCardinality()
	}
	return array
}
func (self *RoaringBitmap) Clone() *RoaringBitmap {
	ptr := new(RoaringBitmap)
	ptr.highlowcontainer = *self.highlowcontainer.Clone()
	return ptr
}

func (self *RoaringBitmap) Contains(x int) bool {
	hb := Highbits(x)
	c := self.highlowcontainer.GetContainer(hb)
	return c != nil && c.Contains(Lowbits(x))

}

func (self *RoaringBitmap) Equals(o interface{}) bool {
	srb := o.(*RoaringBitmap)
	if srb != nil {
		return srb.highlowcontainer.Equals(self.highlowcontainer)
	}
	return false
}

func (self *RoaringBitmap) Add(x int) {
	hb := Highbits(x)
	i := self.highlowcontainer.GetIndex(hb)
	if i >= 0 {
		self.highlowcontainer.setContainerAtIndex(i, self.highlowcontainer.GetContainerAtIndex(i).Add(Lowbits(x)))
	} else {
		newac := NewArrayContainer()
		self.highlowcontainer.insertNewKeyValueAt(-i-1, hb, newac.Add(Lowbits(x)))
	}
}

func (self *RoaringBitmap) GetCardinality() int {
	size := 0
	for i := 0; i < self.highlowcontainer.Size(); i++ {
		size += self.highlowcontainer.GetContainerAtIndex(i).GetCardinality()
	}
	return size
}

func (self *RoaringBitmap) And(x2 *RoaringBitmap) *RoaringBitmap {
	return And(self, x2)
}

func (self *RoaringBitmap) Or(x2 *RoaringBitmap) *RoaringBitmap {
	return Or(self, x2)
}

func Or(x1, x2 *RoaringBitmap) *RoaringBitmap {
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

			for {
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
					answer.highlowcontainer.Append(s1, x1.highlowcontainer.GetContainerAtIndex(pos1).Or(x2.highlowcontainer.GetContainerAtIndex(pos2)))
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
	if pos1 == length1 {
		answer.highlowcontainer.AppendCopy(x2.highlowcontainer, pos2, length2)
	} else if pos2 == length2 {
		answer.highlowcontainer.AppendCopy(x1.highlowcontainer, pos1, length1)
	}
	return answer
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
func (self *RoaringBitmap) Xor(a *RoaringBitmap) *RoaringBitmap {
	return Xor(self, a)
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

func (self *RoaringBitmap) AndNot(x2 *RoaringBitmap) *RoaringBitmap {
	return AndNot(self, x2)
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

func (self *RoaringBitmap) Flip(rangeStart, rangeEnd int) *RoaringBitmap {
	return Flip(self, rangeStart, rangeEnd)
}
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
