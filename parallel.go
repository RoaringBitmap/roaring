package roaring

import (
	"container/heap"
	"fmt"
	"runtime"
)

type parOp int

const (
	parOpAnd parOp = iota
	parOpOr
	parOpXor
	parOpAndNot
)

var defaultWorkerCount int = runtime.NumCPU()

const defaultTaskQueueLength = 4096

type ParAggregator struct {
	taskQueue chan parTask
}

func NewParAggregator(taskQueueLength, workerCount int) ParAggregator {
	agg := ParAggregator{
		make(chan parTask, taskQueueLength),
	}

	for i := 0; i < workerCount; i++ {
		go agg.worker()
	}

	return agg
}

func NewParAggregatorWithDefaults() ParAggregator {
	return NewParAggregator(defaultTaskQueueLength, defaultWorkerCount)
}

func (aggregator ParAggregator) worker() {
	for task := range aggregator.taskQueue {
		var resultContainer container
		switch task.op {
		case parOpAnd:
			resultContainer = task.left.and(task.right)
		}

		result := parResult{
			key: task.key,
			pos: task.pos,
		}

		if resultContainer.getCardinality() > 0 {
			result.container = resultContainer
		} else {
			result.empty = true
		}

		task.result <- result
	}
}

func (aggregator ParAggregator) Shutdown() {
	close(aggregator.taskQueue)
}

type parTask struct {
	op          parOp
	key         uint16
	pos         int
	left, right container
	result      chan<- parResult
}

type parResult struct {
	key       uint16
	pos       int
	container container
	empty     bool
}

func (aggregator ParAggregator) And(x1, x2 *Bitmap) *Bitmap {
	answer := NewBitmap()
	pos1 := 0
	pos2 := 0
	length1 := x1.highlowcontainer.size()
	length2 := x2.highlowcontainer.size()

	var chanLength int
	// take smaller of two input bitmap lengths
	// this makes the buffer large enough not to block the workers
	if length1 > length2 {
		chanLength = length2
	} else {
		chanLength = length1
	}

	resultChan := make(chan parResult, chanLength)
	resultCount := 0

main:
	for pos1 < length1 && pos2 < length2 {
		s1 := x1.highlowcontainer.getKeyAtIndex(pos1)
		s2 := x2.highlowcontainer.getKeyAtIndex(pos2)
		for {
			if s1 == s2 {
				left := x1.highlowcontainer.getContainerAtIndex(pos1)
				right := x2.highlowcontainer.getContainerAtIndex(pos2)

				aggregator.taskQueue <- parTask{
					op:     parOpAnd,
					pos:    resultCount,
					left:   left,
					right:  right,
					result: resultChan,
				}
				resultCount++

				pos1++
				pos2++
				if (pos1 == length1) || (pos2 == length2) {
					break main
				}
				s1 = x1.highlowcontainer.getKeyAtIndex(pos1)
				s2 = x2.highlowcontainer.getKeyAtIndex(pos2)
			} else if s1 < s2 {
				pos1 = x1.highlowcontainer.advanceUntil(s2, pos1)
				if pos1 == length1 {
					break main
				}
				s1 = x1.highlowcontainer.getKeyAtIndex(pos1)
			} else { // s1 > s2
				pos2 = x2.highlowcontainer.advanceUntil(s1, pos2)
				if pos2 == length2 {
					break main
				}
				s2 = x2.highlowcontainer.getKeyAtIndex(pos2)
			}
		}
	}
	// main loop end

	results := make([]parResult, resultCount)

	for result := range resultChan {
		results[result.pos] = result
		resultCount--
		if resultCount == 0 {
			close(resultChan)
			break
		}
	}

	for _, result := range results {
		if !result.empty {
			answer.highlowcontainer.appendContainer(result.key, result.container, false)
		}
	}

	return answer
}

// Wide-or code

type bitmapContainerKey struct {
	bitmap    *Bitmap
	container container
	key       uint16
	idx       int
}

type multipleContainers struct {
	key        uint16
	containers []container
	idx        int
}

type keyedContainer struct {
	key       uint16
	container container
	idx       int
}

type bitmapContainerHeap []bitmapContainerKey

func (h bitmapContainerHeap) Len() int { return len(h) }
func (h bitmapContainerHeap) Less(i, j int) bool {
	// TODO consider container type to comparison
	// we'd prefer bitmap containers to be considered less
	// that will avoid conversions
	return h[i].key < h[j].key
}
func (h bitmapContainerHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *bitmapContainerHeap) Push(x interface{}) {
	// Push and Pop use pointer receivers because they modify the slice's length,
	// not just its contents.
	*h = append(*h, x.(bitmapContainerKey))
}

func (h *bitmapContainerHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func (h bitmapContainerHeap) Peek() bitmapContainerKey {
	return h[0]
}

func (h *bitmapContainerHeap) PopIncrementing() bitmapContainerKey {
	k := h.Peek()

	newIdx := k.idx + 1
	if newIdx < k.bitmap.highlowcontainer.size() {
		newKey := bitmapContainerKey{
			k.bitmap,
			k.bitmap.highlowcontainer.containers[newIdx],
			k.bitmap.highlowcontainer.keys[newIdx],
			newIdx,
		}
		(*h)[0] = newKey
		heap.Fix(h, 0)
	} else {
		heap.Pop(h)
	}
	return k
}

func (h *bitmapContainerHeap) PopNextContainers() multipleContainers {
	if h.Len() == 0 {
		return multipleContainers{}
	}

	containers := make([]container, 0, 4)
	bk := h.PopIncrementing()
	containers = append(containers, bk.container)
	key := bk.key

	for h.Len() > 0 && key == h.Peek().key {
		bk = h.PopIncrementing()
		containers = append(containers, bk.container)
	}

	return multipleContainers{
		key,
		containers,
		-1,
	}
}

func newBitmapContainerHeap(bitmaps ...*Bitmap) bitmapContainerHeap {
	// Initialize heap
	var h bitmapContainerHeap = make([]bitmapContainerKey, 0, len(bitmaps))
	for _, bitmap := range bitmaps {
		if !bitmap.IsEmpty() {
			key := bitmapContainerKey{
				bitmap,
				bitmap.highlowcontainer.containers[0],
				bitmap.highlowcontainer.keys[0],
				0,
			}
			h = append(h, key)
		}
	}

	heap.Init(&h)

	return h
}

func repairAfterLazy(c container) container {
	switch t := c.(type) {
	case *bitmapContainer:
		if t.cardinality == invalidCardinality {
			t.computeCardinality()
		}

		if t.getCardinality() <= arrayDefaultMaxSize {
			return t.toArrayContainer()
		} else if c.(*bitmapContainer).isFull() {
			return newRunContainer16Range(0, MaxUint16)
		}
	}

	return c
}

func toBitmapContainer(c container) container {
	switch t := c.(type) {
	case *arrayContainer:
		return t.toBitmapContainer()
	case *runContainer16:
		if !t.isFull() {
			return t.toBitmapContainer()
		}
	}
	return c
}

func HorizontalOr(bitmaps ...*Bitmap) *Bitmap {
	h := newBitmapContainerHeap(bitmaps...)
	answer := New()

	for h.Len() > 0 {
		item := h.PopNextContainers()
		if len(item.containers) == 0 {
			answer.highlowcontainer.appendContainer(item.key, item.containers[0], true)
		} else {
			c := toBitmapContainer(item.containers[0])
			for _, cx := range item.containers {
				fmt.Printf("%T ", cx)
			}
			fmt.Printf("\n")
			for _, next := range item.containers[1:] {
				c = c.lazyIOR(next)
			}
			c = repairAfterLazy(c)
			answer.highlowcontainer.appendContainer(item.key, c, false)
		}
	}

	return answer
}

func ParOr(bitmaps ...*Bitmap) *Bitmap {
	h := newBitmapContainerHeap(bitmaps...)

	bitmapChan := make(chan *Bitmap)
	inputChan := make(chan multipleContainers, 128)
	resultChan := make(chan keyedContainer, 32)
	expectedKeysChan := make(chan int)

	orFunc := func() {
		for input := range inputChan {
			c := toBitmapContainer(input.containers[0])
			for _, next := range input.containers[1:] {
				c.lazyIOR(next)
			}
			c = repairAfterLazy(c)
			kx := keyedContainer{
				input.key,
				c,
				input.idx,
			}
			resultChan <- kx
		}
	}

	appenderFun := func() {
		expectedKeys := -1
		appendedKeys := 0
		keys := make([]uint16, 0)
		containers := make([]container, 0)
		for appendedKeys != expectedKeys {
			select {
			case item := <-resultChan:
				if len(keys) <= item.idx {
					keys = append(keys, make([]uint16, item.idx-len(keys)+1)...)
					containers = append(containers, make([]container, item.idx-len(containers)+1)...)
				}
				keys[item.idx] = item.key
				containers[item.idx] = item.container

				appendedKeys += 1
			case msg := <-expectedKeysChan:
				expectedKeys = msg
			}
		}
		answer := &Bitmap{
			roaringArray{
				make([]uint16, 0, expectedKeys),
				make([]container, 0, expectedKeys),
				make([]bool, 0, expectedKeys),
				false,
				nil,
			},
		}
		for i := range keys {
			answer.highlowcontainer.appendContainer(keys[i], containers[i], false)
		}

		bitmapChan <- answer
	}

	go appenderFun()

	for i := 0; i < defaultWorkerCount; i++ {
		go orFunc()
	}

	idx := 0
	for h.Len() > 0 {
		ck := h.PopNextContainers()
		if len(ck.containers) == 1 {
			resultChan <- keyedContainer{
				ck.key,
				ck.containers[0],
				idx,
			}
		} else {
			ck.idx = idx
			inputChan <- ck
		}
		idx++
	}
	expectedKeysChan <- idx

	bitmap := <-bitmapChan

	close(inputChan)
	close(resultChan)
	close(expectedKeysChan)

	return bitmap
}
