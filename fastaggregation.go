package roaring

import (
	"container/heap"
	"sort"
)

type rblist []*RoaringBitmap

func (p rblist) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p rblist) Len() int           { return len(p) }
func (p rblist) Less(i, j int) bool { return p[i].getSizeInBytes() > p[j].getSizeInBytes() }

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
func FastOr(bitmaps ...*RoaringBitmap) *RoaringBitmap {
	if len(bitmaps) == 0 {
		return NewRoaringBitmap()
	}

	pq := make(PriorityQueue, len(bitmaps))
	for i, bm := range bitmaps {
		pq[i] = &Item{bm, i}
	}
	heap.Init(&pq)

	for pq.Len() > 1 {
		x1 := heap.Pop(&pq).(*Item)
		x2 := heap.Pop(&pq).(*Item)
		heap.Push(&pq, &Item{Or(x1.value, x2.value), 0})
	}
	return heap.Pop(&pq).(*Item).value
}

func FastXor(bitmaps ...*RoaringBitmap) *RoaringBitmap {
	if len(bitmaps) == 0 {
		return NewRoaringBitmap()
	}

	pq := make(PriorityQueue, len(bitmaps))
	for i, bm := range bitmaps {
		pq[i] = &Item{bm, i}
	}
	heap.Init(&pq)

	for pq.Len() > 1 {
		x1 := heap.Pop(&pq).(*Item)
		x2 := heap.Pop(&pq).(*Item)
		heap.Push(&pq, &Item{Xor(x1.value, x2.value), 0})
	}
	return heap.Pop(&pq).(*Item).value
}
