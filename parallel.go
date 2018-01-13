package roaring

import (
	"container/heap"
	"runtime"
	"sync"
)

var defaultWorkerCount = runtime.NumCPU()

type bitmapContainerKey struct {
	key    uint16
	idx    int
	bitmap *Bitmap
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

func (h *bitmapContainerHeap) popIncrementing() (key uint16, container container) {
	k := h.Peek()
	key = k.key
	container = k.bitmap.highlowcontainer.containers[k.idx]

	newIdx := k.idx + 1
	if newIdx < k.bitmap.highlowcontainer.size() {
		k = bitmapContainerKey{
			k.bitmap.highlowcontainer.keys[newIdx],
			newIdx,
			k.bitmap,
		}
		(*h)[0] = k
		heap.Fix(h, 0)
	} else {
		heap.Pop(h)
	}

	return
}

func (h *bitmapContainerHeap) Next(containers []container) multipleContainers {
	if h.Len() == 0 {
		return multipleContainers{}
	}

	key, container := h.popIncrementing()
	containers = append(containers, container)

	for h.Len() > 0 && key == h.Peek().key {
		_, container = h.popIncrementing()
		containers = append(containers, container)
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
				bitmap.highlowcontainer.keys[0],
				0,
				bitmap,
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

	pool := sync.Pool{
		New: func() interface{} {
			return make([]container, 0, len(bitmaps))
		},
	}

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
			pool.Put(input.containers[:0])
		}
	}

	go appenderRoutine(bitmapChan, resultChan, expectedKeysChan)

	for i := 0; i < parallelism; i++ {
		go orFunc()
	}

	idx := 0
	for h.Len() > 0 {
		ck := h.Next(pool.Get().([]container))
		if len(ck.containers) == 1 {
			resultChan <- keyedContainer{
				ck.key,
				ck.containers[0],
				idx,
			}
			pool.Put(ck.containers[:0])
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
		ck := h.Next(make([]container, 0, 4))
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
