package roaring

type activeContState struct {
	cType int // 0: array, 1: bitmap, 2: run
	shift uint

	// arrayContainer fields
	arrayContent []uint16
	arrayIdx     int

	// bitmapContainer fields
	bitmapContent []uint64

	// runContainer16 fields
	runIv  []interval16
	runIdx int
}

// ParallelBSIScanHelper is an internal optimization helper. It is exported solely because
// the BSI functionality resides in the subpackage "github.com/RoaringBitmap/roaring/v2/BitSliceIndexing"
// which must call into the core package to perform the fast parallel linear scan.
// This function accesses unexported container implementations and their fields (such as keys,
// arrayContainer.content, bitmapContainer.bitmap, and runContainer16.iv) to perform
// highly specialized, zero-allocation container-level membership scans, which would be
// impossible to implement with the same efficiency on the public API.
//
// Normal library users should not call this function directly.
func ParallelBSIScanHelper(cols []uint32, bA []*Bitmap, bitCount int, vals []uint64) *Bitmap {
	// Guard the sorted column ID assumption
	for i := 1; i < len(cols); i++ {
		if cols[i] < cols[i-1] {
			panic("ParallelBSIScanHelper: input cols must be sorted in ascending order")
		}
	}

	// Guard the sorted vals assumption
	for i := 1; i < len(vals); i++ {
		if vals[i] < vals[i-1] {
			panic("ParallelBSIScanHelper: input vals must be sorted in ascending order")
		}
	}

	out := NewBitmap()
	nCols := len(cols)
	if nCols == 0 {
		return out
	}

	if bitCount > 128 {
		panic("ParallelBSIScanHelper: bitCount exceeds 128")
	}
	var curIndexBuf [128]int
	curIndex := curIndexBuf[:bitCount]

	var iCol int
	for iCol < nCols {
		col := cols[iCol]
		hb := uint16(col >> 16)

		var activeBuf [128]activeContState
		active := activeBuf[:0]

		for p := 0; p < bitCount; p++ {
			ra := bA[p].highlowcontainer
			idx := ra.binarySearch(int64(curIndex[p]), int64(len(ra.keys)), hb)
			if idx >= 0 {
				curIndex[p] = idx
				c := ra.containers[idx]

				state := activeContState{
					shift: uint(p),
				}
				switch tc := c.(type) {
				case *arrayContainer:
					state.cType = 0
					state.arrayContent = tc.content
				case *bitmapContainer:
					state.cType = 1
					state.bitmapContent = tc.bitmap
				case *runContainer16:
					state.cType = 2
					state.runIv = tc.iv
				}
				active = append(active, state)
			} else {
				curIndex[p] = -idx - 1
			}
		}

		// Process all columns in the batch that share this hb
		for iCol < nCols {
			currCol := cols[iCol]
			currHb := uint16(currCol >> 16)
			if currHb != hb {
				break
			}

			val := uint64(0)
			overflow := false
			lb := uint16(currCol & 0xffff)
			for p := 0; p < len(active); p++ {
				ac := &active[p]
				found := false
				switch ac.cType {
				case 0: // arrayContainer
					for ac.arrayIdx < len(ac.arrayContent) && ac.arrayContent[ac.arrayIdx] < lb {
						ac.arrayIdx++
					}
					if ac.arrayIdx < len(ac.arrayContent) && ac.arrayContent[ac.arrayIdx] == lb {
						found = true
					}
				case 1: // bitmapContainer
					if (ac.bitmapContent[lb>>6] & (uint64(1) << (lb & 63))) != 0 {
						found = true
					}
				case 2: // runContainer16
					for ac.runIdx < len(ac.runIv) && uint32(ac.runIv[ac.runIdx].start)+uint32(ac.runIv[ac.runIdx].length) < uint32(lb) {
						ac.runIdx++
					}
					if ac.runIdx < len(ac.runIv) && lb >= ac.runIv[ac.runIdx].start {
						found = true
					}
				}

				if found {
					if ac.shift >= 64 {
						overflow = true
						break
					}
					val |= uint64(1) << ac.shift
				}
			}

			if !overflow {
				// Binary search inline on vals
				l, r := 0, len(vals)-1
				foundVal := false
				for l <= r {
					m := (l + r) >> 1
					v := vals[m]
					if v == val {
						foundVal = true
						break
					}
					if v < val {
						l = m + 1
					} else {
						r = m - 1
					}
				}
				if foundVal {
					out.Add(currCol)
				}
			}
			iCol++
		}
	}
	return out
}
