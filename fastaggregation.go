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

// The walk covers every input, up to fastAndWide, when the driver has one
// key, and fastAndNarrow of them otherwise.
const (
	fastAndWide   = 8
	fastAndNarrow = 4
)

// FastAnd computes the intersection between many bitmaps quickly.
// Compared to the And function, it can take many bitmaps as input.
//
// FastAnd walks the keys of the input with the fewest keys and looks each
// one up in the other inputs before allocating anything for it; inputs past
// the walk are intersected into the result one at a time.
func FastAnd(bitmaps ...*Bitmap) *Bitmap {
	switch len(bitmaps) {
	case 0:
		return NewBitmap()
	case 1:
		return bitmaps[0].Clone()
	}
	driver, fewest := 0, bitmaps[0].highlowcontainer.size()
	for i := 1; i < len(bitmaps); i++ {
		if n := bitmaps[i].highlowcontainer.size(); n < fewest {
			driver, fewest = i, n
		}
	}
	dra := &bitmaps[driver].highlowcontainer
	if len(bitmaps) == 2 {
		// A pair over many keys is And's job; a pair on one key is walked.
		if dra.size() > 1 {
			return And(bitmaps[0], bitmaps[1])
		}
		if dra.size() == 0 || bitmaps[1-driver].highlowcontainer.getIndex(dra.getKeyAtIndex(0)) < 0 {
			return NewBitmap()
		}
	}
	n := len(bitmaps)
	if n > fastAndWide || dra.size() > 1 {
		n = min(n, fastAndNarrow)
	}
	var pos [fastAndWide]int
	var cs [fastAndWide]container
	answer := NewBitmap()
	survived := false // the last key passed every input, so checks are skipped
keys:
	for j := 0; j < dra.size(); j++ {
		key := dra.getKeyAtIndex(j)
		dc := dra.getContainerAtIndex(j)
		dx, isArray := dc.(*arrayContainer)
		// An array driver is checked for a value it shares with each input.
		check := isArray && !survived
		// Full containers change nothing; the smallest array goes first.
		m, small, smallLen := 0, -1, 0
		if isArray {
			cs[0], m, small, smallLen = dc, 1, 0, len(dx.content)
		} else if !dc.isFull() {
			cs[0], m = dc, 1
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
			if x, ok := c.(*arrayContainer); ok {
				if small < 0 || len(x.content) < smallLen {
					small, smallLen = m, len(x.content)
				}
			} else if c.isFull() {
				continue
			}
			if check && !dc.intersects(c) {
				survived = false
				continue keys
			}
			cs[m] = c
			m++
		}
		c := andK(cs[:m], small)
		survived = c != nil
		if c != nil {
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
// nil when the result is empty. small is the index of the smallest array in
// cs, or -1 if there is none.
func andK(cs []container, small int) container {
	switch len(cs) {
	case 0:
		return newRunContainer16Range(0, maxCapacity-1)
	case 1:
		return cs[0].clone()
	}
	if small >= 0 {
		a := cs[small].(*arrayContainer)
		copy(cs[1:small+1], cs[:small]) // the rest keep their order after it
		return andKArray(a, cs[1:])
	}
	n := 0
	for i, c := range cs {
		if _, ok := c.(*runContainer16); ok {
			cs[n], cs[i] = cs[i], cs[n]
			n++
		}
	}
	if n == len(cs) {
		return andKRuns(cs)
	}
	if n == 1 && len(cs) == 2 {
		// A single interval against one bitmap is counted in place.
		if rc := cs[0].(*runContainer16); len(rc.iv) == 1 {
			return rc.andBitmapCounted(cs[1].(*bitmapContainer))
		}
	}
	return andKBitmaps(cs[n:], cs[:n])
}

// andKArray narrows the array through rest in a buffer on the stack, sized
// to the array, and allocates only once a value survives every container.
func andKArray(ac *arrayContainer, rest []container) container {
	if len(rest) == 1 {
		return andArrayLast(ac.content, rest[0])
	}
	switch n := len(ac.content); {
	case n <= 64:
		var buf [64]uint16
		return andKArrayInto(ac.content, rest, buf[:])
	case n <= 512:
		var buf [512]uint16
		return andKArrayInto(ac.content, rest, buf[:])
	default:
		var buf [arrayDefaultMaxSize]uint16
		return andKArrayInto(ac.content, rest, buf[:])
	}
}

func andKArrayInto(vals []uint16, rest []container, buf []uint16) container {
	last := len(rest) - 1
	for _, c := range rest[:last] {
		n := 0
		switch c := c.(type) {
		case *arrayContainer:
			n = intersection2by2(vals, c.content, buf)
		case *bitmapContainer:
			for _, v := range vals {
				buf[n] = v
				n += int(c.bitValue(v))
			}
		case *runContainer16:
			n = andArrayInto(c.iv, vals, buf)
		}
		if n == 0 {
			return nil
		}
		vals = buf[:n]
	}
	return andArrayLast(vals, rest[last])
}

// andArrayLast intersects vals with c from the first value they share, so
// nothing is allocated when there is none.
func andArrayLast(vals []uint16, c container) container {
	switch c := c.(type) {
	case *arrayContainer:
		i, j := firstShared2by2(vals, c.content)
		if i < 0 {
			return nil
		}
		ac := newArrayContainerCapacity(min(len(vals)-i, len(c.content)-j))
		ac.content = ac.content[:intersection2by2(vals[i:], c.content[j:], ac.content[:cap(ac.content)])]
		return ac
	case *bitmapContainer:
		for i, v := range vals {
			if c.bitValue(v) != 0 {
				return c.andArray(&arrayContainer{content: vals[i:]})
			}
		}
	case *runContainer16:
		if i := c.firstInArray(vals); i >= 0 {
			return c.andArray(&arrayContainer{content: vals[i:]})
		}
	}
	return nil
}

// andKRuns intersects runs interval by interval, allocating only where they
// all overlap.
func andKRuns(cs []container) container {
	a := cs[0].(*runContainer16)
	if !a.intersects(cs[1]) {
		return nil
	}
	if len(cs) == 2 {
		return a.intersect(cs[1].(*runContainer16)).toEfficientContainer()
	}
	var pos [fastAndWide]int
	var iv []interval16
	for {
		lo, hi, end := 0, maxCapacity-1, maxCapacity-1
		for i, c := range cs {
			rc := c.(*runContainer16)
			if pos[i] == len(rc.iv) {
				if len(iv) == 0 {
					return nil
				}
				return (&runContainer16{iv: iv}).toEfficientContainer()
			}
			r := rc.iv[pos[i]]
			lo, hi, end = max(lo, int(r.start)), min(hi, int(r.last())), min(end, int(r.last()))
		}
		if lo <= hi {
			iv = append(iv, newInterval16Range(uint16(lo), uint16(hi)))
		}
		for i, c := range cs { // step past every interval that ends first
			if int(c.(*runContainer16).iv[pos[i]].last()) == end {
				pos[i]++
			}
		}
	}
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

// andBitmapsMeetWords is how many leading words are scanned for a shared bit
// before dense bitmaps are ANDed straight into the result, skipping scratch.
const andBitmapsMeetWords = 8

// andKBitmaps ANDs the bitmaps, masked by the runs, allocating only for a
// non-empty result.
func andKBitmaps(bms, runs []container) container {
	if len(runs) == 0 && andWordsMeet(bms, andBitmapsMeetWords) {
		answer := newBitmapContainer()
		src := bms[0].(*bitmapContainer).bitmap
		card := uint64(0)
		for _, c := range bms[1:] {
			card = andCardSlice(answer.bitmap, src, c.(*bitmapContainer).bitmap)
			src = answer.bitmap
		}
		answer.cardinality = int(card)
		if card <= arrayDefaultMaxSize {
			return answer.toArrayContainer()
		}
		return answer
	}
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

// andWordsMeet reports whether the bitmaps share a bit in their first n words.
func andWordsMeet(bms []container, n int) bool {
	for i := 0; i < n; i++ {
		w := bms[0].(*bitmapContainer).bitmap[i]
		for _, c := range bms[1:] {
			w &= c.(*bitmapContainer).bitmap[i]
		}
		if w != 0 {
			return true
		}
	}
	return false
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
