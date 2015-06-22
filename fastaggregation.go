package roaring

import (
	"container/heap"
	"sort"
)

type rblist []*RoaringBitmap

func (p rblist) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p rblist) Len() int           { return len(p) }
func (p rblist) Less(i, j int) bool { return p[i].GetSizeInBytes() > p[j].GetSizeInBytes() }

// FastAnd computes the intersection between many bitmaps quickly
func FastAnd(bitmaps ...*RoaringBitmap) *RoaringBitmap {
	if len(bitmaps) == 0 {
		return NewRoaringBitmap()
	} else if len(bitmaps) == 1 {
		return bitmaps[0].Clone()
	}
	array := make(rblist, len(bitmaps), len(bitmaps))
	copy(array, bitmaps)
	sort.Sort(array)
	answer := And(array[0], array[1])

	for _, bm := range array[2:] {
		answer.And(bm)
	}
	return answer
}

//FastHorizontalOr computes the union between many bitmaps quickly, it can be expected to be faster and use less memory than FastOr
func FastHorizontalOr(bitmaps ...*RoaringBitmap) *RoaringBitmap {
	answer := NewRoaringBitmap()
	if len(bitmaps) == 0 {
		return answer
	}
	pq := make(containerPriorityQueue, 0, len(bitmaps))
	for _, bm := range bitmaps {
		if bm.GetCardinality() > 0 {
			pq = append(pq, &containeritem{bm, 0, len(pq)})
		}
	}
	heap.Init(&pq)
	for pq.Len() > 0 {
		x1 := heap.Pop(&pq).(*containeritem)
		thiscontainer := x1.value.highlowcontainer.getContainerAtIndex(x1.keyindex)
		thiskey := x1.value.highlowcontainer.getKeyAtIndex(x1.keyindex)
		x1.keyindex++
		if x1.keyindex < x1.value.highlowcontainer.size() {
			heap.Push(&pq, x1)
		}
		for pq.Len() > 0 && pq[0].value.highlowcontainer.getKeyAtIndex(pq[0].keyindex) == thiskey {
			x2 := heap.Pop(&pq).(*containeritem)
			thisothercontainer := x2.value.highlowcontainer.getContainerAtIndex(x2.keyindex)
			thiscontainer = thiscontainer.lazyIOR(thisothercontainer) // todo: should be an inplace-or
			x2.keyindex++
			if x2.keyindex < x2.value.highlowcontainer.size() {
				heap.Push(&pq, x2)
			}
		}
		switch thiscontainer.(type) {
		case *bitmapContainer:
			thiscontainer.(*bitmapContainer).computeCardinality()
		}
		answer.highlowcontainer.appendContainer(thiskey, thiscontainer)
	}
	return answer
}

// FastOr computes the union between many bitmaps quickly (see also FastHorizontalOr)
func FastOr(bitmaps ...*RoaringBitmap) *RoaringBitmap {
	// Todo: we really want a port of horizontal_or (see https://github.com/lemire/RoaringBitmap/blob/master/src/main/java/org/roaringbitmap/FastAggregation.java#L84-L126 ) for better speed
	if len(bitmaps) == 0 {
		return NewRoaringBitmap()
	}

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

// FastXor computes the intersection between many bitmaps quickly
func FastXor(bitmaps ...*RoaringBitmap) *RoaringBitmap {
	if len(bitmaps) == 0 {
		return NewRoaringBitmap()
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
