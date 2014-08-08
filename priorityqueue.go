package roaring

import "container/heap"

type item struct {
	value *RoaringBitmap
	index int
}

type priorityQueue []*item

func (pq priorityQueue) len() int { return len(pq) }

func (pq priorityQueue) less(i, j int) bool {
	return pq[i].value.getSizeInBytes() > pq[j].value.getSizeInBytes()
}

func (pq priorityQueue) swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *priorityQueue) push(x interface{}) {
	n := len(*pq)
	item := x.(*item)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *priorityQueue) pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

func (pq *PriorityQueue) update(item *Item, value *RoaringBitmap) {
	item.value = value
	heap.Fix(pq, item.index)
}
