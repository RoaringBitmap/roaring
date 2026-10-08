package roaring

import (
	"container/heap"
)

// Or function that requires repairAfterLazy
func lazyOR(x1, x2 *Bitmap) *Bitmap {
	answer := NewBitmap()
	pos1 := 0
	pos2 := 0
	length1 := x1.highlowcontainer.size()
	length2 := x2.highlowcontainer.size()
main:
	for (pos1 < length1) && (pos2 < length2) {
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
				c1 := x1.highlowcontainer.getContainerAtIndex(pos1)
				answer.highlowcontainer.appendContainer(s1, c1.lazyOR(x2.highlowcontainer.getContainerAtIndex(pos2)), false)
				pos1++
				pos2++
				if (pos1 == length1) || (pos2 == length2) {
					break main
				}
				s1 = x1.highlowcontainer.getKeyAtIndex(pos1)
				s2 = x2.highlowcontainer.getKeyAtIndex(pos2)
			}
		}
	}
	if pos1 == length1 {
		answer.highlowcontainer.appendCopyMany(x2.highlowcontainer, pos2, length2)
	} else if pos2 == length2 {
		answer.highlowcontainer.appendCopyMany(x1.highlowcontainer, pos1, length1)
	}
	return answer
}

// In-place Or function that requires repairAfterLazy
func (x1 *Bitmap) lazyOR(x2 *Bitmap) *Bitmap {
	pos1 := 0
	pos2 := 0
	length1 := x1.highlowcontainer.size()
	length2 := x2.highlowcontainer.size()
main:
	for (pos1 < length1) && (pos2 < length2) {
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
				x1.highlowcontainer.insertNewKeyValueAt(pos1, s2, x2.highlowcontainer.getContainerAtIndex(pos2).clone())
				pos2++
				pos1++
				length1++
				if pos2 == length2 {
					break main
				}
				s2 = x2.highlowcontainer.getKeyAtIndex(pos2)
			} else {
				c1 := x1.highlowcontainer.getWritableContainerAtIndex(pos1)
				// runContainer16.lazyIOR falls back to a slow ior path
				// (O(N log R) per merged element); promote to bitmapContainer
				// first, whose lazy union is O(1024) regardless of cardinality.
				// See https://github.com/RoaringBitmap/roaring/issues/81.
				if rc, ok := c1.(*runContainer16); ok && !rc.isFull() {
					c1 = rc.toBitmapContainer()
				}
				x1.highlowcontainer.containers[pos1] = c1.lazyIOR(x2.highlowcontainer.getContainerAtIndex(pos2))
				x1.highlowcontainer.needCopyOnWrite[pos1] = false
				pos1++
				pos2++
				if (pos1 == length1) || (pos2 == length2) {
					break main
				}
				s1 = x1.highlowcontainer.getKeyAtIndex(pos1)
				s2 = x2.highlowcontainer.getKeyAtIndex(pos2)
			}
		}
	}
	if pos1 == length1 {
		x1.highlowcontainer.appendCopyMany(x2.highlowcontainer, pos2, length2)
	}
	return x1
}

// to be called after lazy aggregates
func (x1 *Bitmap) repairAfterLazy() {
	for pos := 0; pos < x1.highlowcontainer.size(); pos++ {
		c := x1.highlowcontainer.getContainerAtIndex(pos)
		switch c.(type) {
		case *bitmapContainer:
			if c.(*bitmapContainer).cardinality == invalidCardinality {
				c = x1.highlowcontainer.getWritableContainerAtIndex(pos)
				c.(*bitmapContainer).computeCardinality()
				if c.(*bitmapContainer).getCardinality() <= arrayDefaultMaxSize {
					x1.highlowcontainer.setContainerAtIndex(pos, c.(*bitmapContainer).toArrayContainer())
				} else if c.(*bitmapContainer).isFull() {
					x1.highlowcontainer.setContainerAtIndex(pos, newRunContainer16Range(0, MaxUint16))
				}
			}
		}
	}
}

// FastAnd computes the intersection between many bitmaps quickly.
// Compared to the And function, it can take many bitmaps as input.
//
// With three or more inputs, FastAnd walks the keys of the input with the
// fewest keys and looks each one up in the first four inputs, also checking
// an array container for a shared value, before allocating anything for it.
// The remaining inputs are then intersected into the result one at a time.
func FastAnd(bitmaps ...*Bitmap) *Bitmap {
	switch len(bitmaps) {
	case 0:
		return NewBitmap()
	case 1:
		return bitmaps[0].Clone()
	case 2:
		return And(bitmaps[0], bitmaps[1])
	}
	driver := 0
	for i := 1; i < len(bitmaps); i++ {
		if bitmaps[i].highlowcontainer.size() < bitmaps[driver].highlowcontainer.size() {
			driver = i
		}
	}
	dra := &bitmaps[driver].highlowcontainer
	// The walk covers the first width inputs, each with a cursor that resumes
	// where the previous key stopped. Past those, intersecting an input into
	// the result costs less per key than walking it.
	const width = 4
	n := min(len(bitmaps), width)
	var pos [width]int
	var cs [width + 1]container // the window and a driver past it
	answer := NewBitmap()
keys:
	for j := 0; j < dra.size(); j++ {
		key := dra.getKeyAtIndex(j)
		dc := dra.getContainerAtIndex(j)
		// Only an array is checked for a value it shares with each input: a
		// lookup per value that usually stops at the first. For a run or a
		// bitmap the check costs about as much as the intersection.
		_, check := dc.(*arrayContainer)
		m := 0
		if !dc.isFull() { // a full container changes nothing
			cs[0] = dc
			m = 1
		}
		for i, bm := range bitmaps[:n] {
			if i == driver {
				continue
			}
			ra := &bm.highlowcontainer
			p := pos[i]
			if p < ra.size() && ra.getKeyAtIndex(p) < key {
				p = ra.advanceUntil(key, p)
			}
			pos[i] = p
			if p == ra.size() {
				break keys
			}
			if ra.getKeyAtIndex(p) != key {
				continue keys
			}
			pos[i]++
			c := ra.getContainerAtIndex(p)
			if c.isFull() {
				continue
			}
			if check && !dc.intersects(c) {
				continue keys
			}
			cs[m] = c
			m++
		}
		if c := andK(cs[:m]); c != nil {
			if answer.IsEmpty() {
				// Room for up to 64 of the driver's remaining keys at once.
				k := min(dra.size()-j, 64)
				answer.highlowcontainer = roaringArray{
					keys:            make([]uint16, 0, k),
					containers:      make([]container, 0, k),
					needCopyOnWrite: make([]bool, 0, k),
				}
			}
			answer.highlowcontainer.appendContainer(key, c, false)
		}
	}
	for i := n; i < len(bitmaps) && !answer.IsEmpty(); i++ {
		if i != driver {
			answer.And(bitmaps[i])
		}
	}
	return answer
}

// andK intersects the containers of one key, none of them full, and returns
// nil when the intersection is empty. It may reorder cs.
func andK(cs []container) container {
	switch len(cs) {
	case 0:
		return newRunContainer16Range(0, maxCapacity-1)
	case 1:
		return cs[0].clone()
	}
	// The smallest array goes first; the rest keep their order.
	smallest := -1
	for i, c := range cs {
		if x, ok := c.(*arrayContainer); ok && (smallest < 0 || x.getCardinality() < cs[smallest].getCardinality()) {
			smallest = i
		}
	}
	if smallest >= 0 {
		a := cs[smallest]
		copy(cs[1:smallest+1], cs[:smallest])
		cs[0] = a
		return andKChain(cs)
	}
	n := 0
	for i, c := range cs {
		if _, ok := c.(*runContainer16); ok {
			cs[n], cs[i] = cs[i], cs[n]
			n++
		}
	}
	if n == len(cs) {
		return andKChain(cs)
	}
	return andKBitmaps(cs[n:], cs[:n])
}

// andKChain intersects cs in order: the first two into a new container, then
// the rest into it in place, so nothing bigger than the first result is built.
func andKChain(cs []container) container {
	c := cs[0].and(cs[1])
	for i := 2; i < len(cs) && !c.isEmpty(); i++ {
		c = c.iand(cs[i])
	}
	if c.isEmpty() {
		return nil
	}
	return c
}

// andKBitmaps ANDs the bitmaps in a scratch bitmap on the stack, stopping as
// soon as the result is empty. Runs act as masks: only the words between
// their common start and end are ANDed, and the gaps inside a run are cleared
// afterwards.
func andKBitmaps(bms, runs []container) container {
	lo, hi := 0, maxCapacity-1
	gaps := false
	for _, c := range runs {
		rc := c.(*runContainer16)
		lo, hi = max(lo, int(rc.minimum())), min(hi, int(rc.maximum()))
		gaps = gaps || len(rc.iv) > 1
	}
	if lo > hi {
		return nil
	}
	first, last := lo/64, hi/64
	var scratch [bitmapContainerSize]uint64
	w := scratch[first : last+1]
	src := bms[0].(*bitmapContainer).bitmap[first : last+1]
	card := uint64(0)
	if len(bms) == 1 {
		copy(w, src)
	}
	for _, c := range bms[1:] {
		if card = andCardSlice(w, src, c.(*bitmapContainer).bitmap[first:last+1]); card == 0 {
			return nil
		}
		src = w
	}
	if len(bms) == 1 || gaps || lo%64 != 0 || (hi+1)%64 != 0 {
		resetBitmapRange(scratch[:], first*64, lo)
		resetBitmapRange(scratch[:], hi+1, (last+1)*64)
		for _, c := range runs {
			clearBitmapGaps(scratch[:], c.(*runContainer16).iv, lo, hi+1)
		}
		if card = popcntSlice(w); card == 0 {
			return nil
		}
	}
	return containerFromWords(scratch[:], first, last, int(card))
}

// FastOr computes the union between many bitmaps quickly, as opposed to having to call Or repeatedly.
// It might also be faster than calling Or repeatedly.
func FastOr(bitmaps ...*Bitmap) *Bitmap {
	if len(bitmaps) == 0 {
		return NewBitmap()
	} else if len(bitmaps) == 1 {
		return bitmaps[0].Clone()
	}
	answer := lazyOR(bitmaps[0], bitmaps[1])
	for _, bm := range bitmaps[2:] {
		answer = answer.lazyOR(bm)
	}
	// here is where repairAfterLazy is called.
	answer.repairAfterLazy()
	return answer
}

// HeapOr computes the union between many bitmaps quickly using a heap.
// It might be faster than calling Or repeatedly.
func HeapOr(bitmaps ...*Bitmap) *Bitmap {
	if len(bitmaps) == 0 {
		return NewBitmap()
	}
	// TODO:  for better speed, we could do the operation lazily, see Java implementation
	pq := make(priorityQueue, len(bitmaps))
	for i, bm := range bitmaps {
		pq[i] = &item{bm, i}
	}
	heap.Init(&pq)

	for pq.Len() > 1 {
		x1 := heap.Pop(&pq).(*item)
		x2 := heap.Pop(&pq).(*item)
		heap.Push(&pq, &item{Or(x1.value, x2.value), 0})
	}
	return heap.Pop(&pq).(*item).value
}

// HeapXor computes the symmetric difference between many bitmaps quickly (as opposed to calling Xor repeated).
// Internally, this function uses a heap.
// It might be faster than calling Xor repeatedly.
func HeapXor(bitmaps ...*Bitmap) *Bitmap {
	if len(bitmaps) == 0 {
		return NewBitmap()
	}

	pq := make(priorityQueue, len(bitmaps))
	for i, bm := range bitmaps {
		pq[i] = &item{bm, i}
	}
	heap.Init(&pq)

	for pq.Len() > 1 {
		x1 := heap.Pop(&pq).(*item)
		x2 := heap.Pop(&pq).(*item)
		heap.Push(&pq, &item{Xor(x1.value, x2.value), 0})
	}
	return heap.Pop(&pq).(*item).value
}

// AndAny provides a result equivalent to x1.And(FastOr(bitmaps)).
// It's optimized to minimize allocations. It also might be faster than separate calls.
func (x1 *Bitmap) AndAny(bitmaps ...*Bitmap) {
	if len(bitmaps) == 0 {
		return
	} else if len(bitmaps) == 1 {
		x1.And(bitmaps[0])
		return
	}

	type withPos struct {
		bitmap *roaringArray
		pos    int
		key    uint16
	}
	filters := make([]withPos, 0, len(bitmaps))

	for _, b := range bitmaps {
		if b.highlowcontainer.size() > 0 {
			filters = append(filters, withPos{
				bitmap: &b.highlowcontainer,
				pos:    0,
				key:    b.highlowcontainer.getKeyAtIndex(0),
			})
		}
	}

	basePos := 0
	intersections := 0
	keyContainers := make([]container, 0, len(filters))
	var (
		tmpArray   *arrayContainer
		tmpBitmap  *bitmapContainer
		minNextKey uint16
	)

	for basePos < x1.highlowcontainer.size() && len(filters) > 0 {
		baseKey := x1.highlowcontainer.getKeyAtIndex(basePos)

		// accumulate containers for current key, find next minimal key in filters
		// and exclude filters that do not have related values anymore
		i := 0
		maxPossibleOr := 0
		minNextKey = MaxUint16
		for _, f := range filters {
			if f.key < baseKey {
				f.pos = f.bitmap.advanceUntil(baseKey, f.pos)
				if f.pos == f.bitmap.size() {
					continue
				}
				f.key = f.bitmap.getKeyAtIndex(f.pos)
			}

			if f.key == baseKey {
				cont := f.bitmap.getContainerAtIndex(f.pos)
				keyContainers = append(keyContainers, cont)
				maxPossibleOr += cont.getCardinality()

				f.pos++
				if f.pos == f.bitmap.size() {
					continue
				}
				f.key = f.bitmap.getKeyAtIndex(f.pos)
			}

			minNextKey = minOfUint16(minNextKey, f.key)
			filters[i] = f
			i++
		}
		filters = filters[:i]

		if len(keyContainers) == 0 {
			basePos = x1.highlowcontainer.advanceUntil(minNextKey, basePos)
			continue
		}

		var ored container

		if len(keyContainers) == 1 {
			ored = keyContainers[0]
		} else {
			//TODO: special case for run containers?
			if maxPossibleOr > arrayDefaultMaxSize {
				if tmpBitmap == nil {
					tmpBitmap = newBitmapContainer()
				}
				tmpBitmap.resetTo(keyContainers[0])
				ored = tmpBitmap
			} else {
				if tmpArray == nil {
					tmpArray = newArrayContainerCapacity(maxPossibleOr)
				}
				tmpArray.realloc(maxPossibleOr)
				tmpArray.resetTo(keyContainers[0])
				ored = tmpArray
			}
			for _, c := range keyContainers[1:] {
				ored = ored.ior(c)
			}
		}

		result := x1.highlowcontainer.getWritableContainerAtIndex(basePos).iand(ored)
		if !result.isEmpty() {
			x1.highlowcontainer.replaceKeyAndContainerAtIndex(intersections, baseKey, result, false)
			intersections++
		}

		keyContainers = keyContainers[:0]
		basePos = x1.highlowcontainer.advanceUntil(minNextKey, basePos)
	}

	x1.highlowcontainer.resize(intersections)
}
