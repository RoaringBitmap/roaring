package goroaring

// TODO: need to resize buffer, not return an int
func Unsigned_difference(set1 []short, set2 []short,  buffer []short) int {
	pos := 0
	k1 := 0
	k2 := 0
	if 0 == len(set2) {
		for k := 0; k < len(set1); k++ {
			buffer[k] = set1[k]
		}
		return len(set1)
	}
	if 0 == len(set1) {
		return 0
	}
	for {
		if set1[k1] < set2[k2] {
			buffer[pos] = set1[k1]
			pos++
			k1++
			if k1 >= len(set1) {
				break
			}
		} else if set1[k1] == set2[k2] {
			k1++
			k2++
			if k1 >= len(set1) {
				break
			}
			if k2 >= len(set2) {
				for ; k1 < len(set1); k1++ {
					buffer[pos] = set1[k1]
					pos++
				}
				break
			}
		} else { // if (val1>val2)
			k2++
			if k2 >= len(set2) {
				for ; k1 < len(set1); k1++ {
					buffer[pos] = set1[k1]
					pos++
				}
				break
			}
		}
	}
	return pos
}

func Unsigned_ExclusiveUnionb2by2(set1 []short, set2 []short, buffer []short) int {
	pos := 0
	k1 := 0
	k2 := 0
	if 0 == len(set2) {
		//	System.arraycopy(set1, 0, buffer, 0, len(set1));
		copy(buffer, set1[:len(set1)])
		return len(set1)
	}
	if 0 == len(set1) {
		//	System.arraycopy(set2, 0, buffer, 0, len(set2));
		copy(buffer, set2[:len(set2)])
		return len(set2)
	}
	for {
		if ToIntUnsigned(set1[k1]) < ToIntUnsigned(set2[k2]) {
			buffer[pos] = set1[k1]
			pos++
			k1++
			if k1 >= len(set1) {
				for ; k2 < len(set2); k2++ {
					buffer[pos] = set2[k2]
					pos++
				}
				break
			}
		} else if ToIntUnsigned(set1[k1]) == ToIntUnsigned(set2[k2]) {
			k1++
			k2++
			if k1 >= len(set1) {
				for ; k2 < len(set2); k2++ {
					buffer[pos] = set2[k2]
					pos++
				}
				break
			}
			if k2 >= len(set2) {
				for ; k1 < len(set1); k1++ {
					buffer[pos] = set1[k1]
					pos++
				}
				break
			}
		} else { // if (val1>val2)
			buffer[pos] = set2[k2]
			pos++
			k2++
			if k2 >= len(set2) {
				for ; k1 < len(set1); k1++ {
					buffer[pos] = set1[k1]
					pos++
				}
				break
			}
		}
	}
	return pos
}

func UnsignedUnion2by2(set1 []short, set2 []short, buffer []short) int {
	pos := 0
	k1 := 0
	k2 := 0
	if 0 == len(set2) {
		copy(buffer, set1[:len(set1)])
		return len(set1)
	}
	if 0 == len(set1) {
		copy(buffer, set2[:len(set2)])
		return len(set2)
	}
	for {
		if ToIntUnsigned(set1[k1]) < ToIntUnsigned(set2[k2]) {
			buffer[pos] = set1[k1]
			pos++
			k1++
			if k1 >= len(set1) {
				for ; k2 < len(set2); k2++ {
					buffer[pos] = set2[k2]
					pos++
				}
				break
			}
		} else if ToIntUnsigned(set1[k1]) == ToIntUnsigned(set2[k2]) {
			buffer[pos] = set1[k1]
			pos++
			k1++
			k2++
			if k1 >= len(set1) {
				for ; k2 < len(set2); k2++ {
					buffer[pos] = set2[k2]
					pos++
				}
				break
			}
			if k2 >= len(set2) {
				for ; k1 < len(set1); k1++ {
					buffer[pos] = set1[k1]
					pos++
				}
				break
			}
		} else { // if (set1[k1]>set2[k2])
			buffer[pos] = set2[k2]
			pos++
			k2++
			if k2 >= len(set2) {
				for ; k1 < len(set1); k1++ {
					buffer[pos] = set1[k1]
					pos++
				}
				break
			}
		}
	}
	return pos
}


func Unsigned_intersect2by2(
	set1 []short,
	set2 []short,
	buffer []short) int {
	if len(set1)*64 < len(set2) {
		return Unsigned_onesidedgallopingintersect2by2(set1, set2, buffer)
	} else if len(set2)*64 < len(set1) {
		return Unsigned_onesidedgallopingintersect2by2(set2, set1, buffer)
	}

	return Unsigned_localintersect2by2(set1, set2, buffer)
}

func Unsigned_localintersect2by2(
	set1 []short,
	set2 []short,
	buffer []short) int {

	if (0 == len(set1)) || (0 == len(set2)) {
		return 0
	}
	k1 := 0
	k2 := 0
	pos := 0

mainwhile:
	for {
		if set2[k2] < set1[k1] {
			for {
				k2++
				if k2 == len(set2) {
					break mainwhile
				}
				if set2[k2] < set1[k1] {
					break
				}
			}
		}
		if set1[k1] < set2[k2] {
			for {
				k1++
				if k1 == len(set1) {
					break mainwhile
				}
				if set1[k1] < set2[k2] {
					break
				}
			}

		} else {
			// (set2[k2] == set1[k1])
			buffer[pos] = set1[k1]
			pos++
			k1++
			if k1 == len(set1) {
				break
			}
			k2++
			if k2 == len(set2) {
				break
			}
		}
	}
	return pos
}

func Unsigned_onesidedgallopingintersect2by2(
	smallset []short,
	largeset []short,
	buffer []short) int {

	if 0 == len(smallset) {
		return 0
	}
	k1 := 0
	k2 := 0
	pos := 0
mainwhile:
	for {
		if largeset[k1] < smallset[k2] {
			k1 = AdvanceUntil(largeset, k1, len(largeset), smallset[k2])
			if k1 == len(largeset) {
				break mainwhile
			}
		}
		if smallset[k2] < largeset[k1] {
			k2++
			if k2 == len(smallset) {
				break mainwhile
			}
		} else {

			buffer[pos] = smallset[k2]
			pos++
			k2++
			if k2 == len(smallset) {
				break
			}
			k1 = AdvanceUntil(largeset, k1, len(largeset), smallset[k2])
			if k1 == len(largeset) {
				break mainwhile
			}
		}

	}
	return pos
}

func Unsigned_binarySearch(array []short, begin, end int, k short) int {
	low := begin
	high := end - 1
	ikey := int(k)

	for low <= high {
		middleIndex := int(uint(low+high) >> 1)
		middleValue := int(array[middleIndex])

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
