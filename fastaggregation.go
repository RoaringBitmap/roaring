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
// FastOr computes the union between many bitmaps quickly
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
