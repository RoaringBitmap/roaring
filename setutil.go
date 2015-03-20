package roaring

func equal(a, b []uint16) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func difference(set1 []uint16, set2 []uint16, buffer []uint16) int {
	if 0 == len(set2) {
		for k := 0; k < len(set1); k++ {
			buffer[k] = set1[k]
		}
		return len(set1)
	}
	if 0 == len(set1) {
		return 0
	}
	pos := 0
	k1 := 0
	k2 := 0
	buffer = buffer[:cap(buffer)]
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

func exclusiveUnion2by2(set1 []uint16, set2 []uint16, buffer []uint16) int {
	if 0 == len(set2) {
		buffer = buffer[:len(set1)]
		copy(buffer, set1[:len(set1)])
		return len(set1)
	}
	if 0 == len(set1) {
		buffer = buffer[:len(set2)]
		copy(buffer, set2[:len(set2)])
		return len(set2)
	}
	pos := 0
	k1 := 0
	k2 := 0
	buffer = buffer[:cap(buffer)]
	for {
		if set1[k1] < set2[k2] {
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
		} else if set1[k1] == set2[k2] {
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

func union2by2(set1 []uint16, set2 []uint16, buffer []uint16) int {
	pos := 0
	k1 := 0
	k2 := 0
	if 0 == len(set2) {
		buffer = buffer[:len(set1)]
		copy(buffer, set1[:len(set1)])
		return len(set1)
	}
	if 0 == len(set1) {
		buffer = buffer[:len(set2)]
		copy(buffer, set2[:len(set2)])
		return len(set2)
	}
	buffer = buffer[:cap(buffer)]
	for {
		if set1[k1] < set2[k2] {
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
		} else if set1[k1] == set2[k2] {
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

func intersection2by2(
	set1 []uint16,
	set2 []uint16,
	buffer []uint16) int {

	if len(set1)*64 < len(set2) {
		return onesidedgallopingintersect2by2(set1, set2, buffer)
	} else if len(set2)*64 < len(set1) {
		return onesidedgallopingintersect2by2(set2, set1, buffer)
	} else {
		return localintersect2by2(set1, set2, buffer)
	}
}

func localintersect2by2(
	set1 []uint16,
	set2 []uint16,
	buffer []uint16) int {

	if (0 == len(set1)) || (0 == len(set2)) {
		return 0
	}
	k1 := 0
	k2 := 0
	pos := 0
	buffer = buffer[:cap(buffer)]
mainwhile:
	for {

		if set2[k2] < set1[k1] {
			for {
				k2++
				if k2 == len(set2) {
					break mainwhile
				}
				if set2[k2] >= set1[k1] {
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
				if set1[k1] >= set2[k2] {
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

func advanceUntil(
	array []uint16,
	pos int,
	length int,
	min uint16) int {
	lower := pos + 1

	if lower >= length || array[lower] >= min {
		return lower
	}

	spansize := 1

	for lower+spansize < length && array[lower+spansize] < min {
		spansize *= 2
	}
	var upper int
	if lower+spansize < length {
		upper = lower + spansize
	} else {
		upper = length - 1
	}

	if array[upper] == min {
		return upper
	}

	if array[upper] < min {
		// means
		// array
		// has no
		// item
		// >= min
		// pos = array.length;
		return length
	}

	// we know that the next-smallest span was too small
	lower += (spansize / 2)

	mid := 0
	for lower+1 != upper {
		mid = (lower + upper) / 2
		if array[mid] == min {
			return mid
		} else if array[mid] < min {
			lower = mid
		} else {
			upper = mid
		}
	}
	return upper

}

func onesidedgallopingintersect2by2(
	smallset []uint16,
	largeset []uint16,
	buffer []uint16) int {

	if 0 == len(smallset) {
		return 0
	}
	buffer = buffer[:cap(buffer)]
	k1 := 0
	k2 := 0
	pos := 0
mainwhile:

	for {
		if largeset[k1] < smallset[k2] {
			k1 = advanceUntil(largeset, k1, len(largeset), smallset[k2])
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

			k1 = advanceUntil(largeset, k1, len(largeset), smallset[k2])
			if k1 == len(largeset) {
				break mainwhile
			}
		}

	}
	return pos
}

// probably useless
func binarySearchOverRange(array []uint16, begin, end int, k uint16) int {
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

func binarySearch(array []uint16, k uint16) int {
	low := 0
	high := len(array) - 1
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
