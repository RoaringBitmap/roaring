package roaring

import (
	"container/heap"
	"runtime"
)

var defaultWorkerCount = runtime.NumCPU()

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

func (h bitmapContainerHeap) Len() int           { return len(h) }
func (h bitmapContainerHeap) Less(i, j int) bool { return h[i].key < h[j].key }
func (h bitmapContainerHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

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

func appenderRoutine(bitmapChan chan<- *Bitmap, resultChan <-chan keyedContainer, expectedKeysChan <-chan int) {
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

			appendedKeys++
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
		if containers[i] != nil { // in case a resulting container was empty, see ParAnd function
			answer.highlowcontainer.appendContainer(keys[i], containers[i], false)
		}
	}

	bitmapChan <- answer
}

// ParOr computes the union (OR) of all provided bitmaps in parallel,
// where the parameter "parallelism" determines how many workers are to be used
// (if it is set to 0, a default number of workers is chosen)
func ParOr(parallelism int, bitmaps ...*Bitmap) *Bitmap {
	bitmapCount := len(bitmaps)
	if bitmapCount == 0 {
		return NewBitmap()
	} else if bitmapCount == 1 {
		return bitmaps[0].Clone()
	}

	if parallelism == 0 {
		parallelism = defaultWorkerCount
	}

	h := newBitmapContainerHeap(bitmaps...)

	bitmapChan := make(chan *Bitmap)
	inputChan := make(chan multipleContainers, 128)
	resultChan := make(chan keyedContainer, 32)
	expectedKeysChan := make(chan int)

	orFunc := func() {
		// Assumes only structs with >=2 containers are passed
		for input := range inputChan {
			c := toBitmapContainer(input.containers[0]).lazyOR(input.containers[1])
			for _, next := range input.containers[2:] {
				c = c.lazyIOR(next)
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

	go appenderRoutine(bitmapChan, resultChan, expectedKeysChan)

	for i := 0; i < parallelism; i++ {
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

// ParAnd computes the intersection (AND) of all provided bitmaps in parallel,
// where the parameter "parallelism" determines how many workers are to be used
// (if it is set to 0, a default number of workers is chosen)
func ParAnd(parallelism int, bitmaps ...*Bitmap) *Bitmap {
	bitmapCount := len(bitmaps)
	if bitmapCount == 0 {
		return NewBitmap()
	} else if bitmapCount == 1 {
		return bitmaps[0].Clone()
	}

	if parallelism == 0 {
		parallelism = defaultWorkerCount
	}

	h := newBitmapContainerHeap(bitmaps...)

	bitmapChan := make(chan *Bitmap)
	inputChan := make(chan multipleContainers, 128)
	resultChan := make(chan keyedContainer, 32)
	expectedKeysChan := make(chan int)

	andFunc := func() {
		// Assumes only structs with >=2 containers are passed
		for input := range inputChan {
			c := input.containers[0].and(input.containers[1])
			for _, next := range input.containers[2:] {
				if c.getCardinality() == 0 {
					break
				}
				c = c.iand(next)
			}

			// Send a nil explicitly if the result of the intersection is an empty container
			if c.getCardinality() == 0 {
				c = nil
			}

			kx := keyedContainer{
				input.key,
				c,
				input.idx,
			}
			resultChan <- kx
		}
	}

	go appenderRoutine(bitmapChan, resultChan, expectedKeysChan)

	for i := 0; i < parallelism; i++ {
		go andFunc()
	}

	idx := 0
	for h.Len() > 0 {
		ck := h.PopNextContainers()
		if len(ck.containers) == bitmapCount {
			ck.idx = idx
			inputChan <- ck
			idx++
		}
	}
	expectedKeysChan <- idx

	bitmap := <-bitmapChan

	close(inputChan)
	close(resultChan)
	close(expectedKeysChan)

	return bitmap
}
