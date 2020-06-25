package roaring

// to run just these tests: go test -run TestFastAggregations*

import (
	"container/heap"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFastAggregationsSize(t *testing.T) {
	rb1 := NewBitmap()
	rb2 := NewBitmap()
	rb3 := NewBitmap()
	for i := uint32(0); i < 1000000; i += 3 {
		rb1.Add(i)
	}
	for i := uint32(0); i < 1000000; i += 7 {
		rb2.Add(i)
	}
	for i := uint32(0); i < 1000000; i += 1001 {
		rb3.Add(i)
	}
	pq := make(priorityQueue, 3)
	pq[0] = &item{rb1, 0}
	pq[1] = &item{rb2, 1}
	pq[2] = &item{rb3, 2}
	heap.Init(&pq)

	assert.Equal(t, rb3.GetSizeInBytes(), heap.Pop(&pq).(*item).value.GetSizeInBytes())
	assert.Equal(t, rb2.GetSizeInBytes(), heap.Pop(&pq).(*item).value.GetSizeInBytes())
	assert.Equal(t, rb1.GetSizeInBytes(), heap.Pop(&pq).(*item).value.GetSizeInBytes())
}

func TestFastAggregationsCont(t *testing.T) {
	rb1 := NewBitmap()
	rb2 := NewBitmap()
	rb3 := NewBitmap()
	for i := uint32(0); i < 10; i += 3 {
		rb1.Add(i)
	}
	for i := uint32(0); i < 10; i += 7 {
		rb2.Add(i)
	}
	for i := uint32(0); i < 10; i += 1001 {
		rb3.Add(i)
	}
	for i := uint32(1000000); i < 1000000+10; i += 1001 {
		rb1.Add(i)
	}
	for i := uint32(1000000); i < 1000000+10; i += 7 {
		rb2.Add(i)
	}
	for i := uint32(1000000); i < 1000000+10; i += 3 {
		rb3.Add(i)
	}
	rb1.Add(500000)
	pq := make(containerPriorityQueue, 3)
	pq[0] = &containeritem{rb1, 0, 0}
	pq[1] = &containeritem{rb2, 0, 1}
	pq[2] = &containeritem{rb3, 0, 2}
	heap.Init(&pq)
	expected := []int{6, 4, 5, 6, 5, 4, 6}
	counter := 0
	for pq.Len() > 0 {
		x1 := heap.Pop(&pq).(*containeritem)
		assert.EqualValues(t, expected[counter], x1.value.GetCardinality())

		counter++
		x1.keyindex++
		if x1.keyindex < x1.value.highlowcontainer.size() {
			heap.Push(&pq, x1)
		}
	}
}

func TestFastAggregationsAdvanced_run(t *testing.T) {
	rb1 := NewBitmap()
	rb2 := NewBitmap()
	rb3 := NewBitmap()
	for i := uint32(500); i < 75000; i++ {
		rb1.Add(i)
	}
	for i := uint32(0); i < 1000000; i += 7 {
		rb2.Add(i)
	}
	for i := uint32(0); i < 1000000; i += 1001 {
		rb3.Add(i)
	}
	for i := uint32(1000000); i < 2000000; i += 1001 {
		rb1.Add(i)
	}
	for i := uint32(1000000); i < 2000000; i += 3 {
		rb2.Add(i)
	}
	for i := uint32(1000000); i < 2000000; i += 7 {
		rb3.Add(i)
	}
	rb1.RunOptimize()
	rb1.Or(rb2)
	rb1.Or(rb3)
	bigand := And(And(rb1, rb2), rb3)
	bigxor := Xor(Xor(rb1, rb2), rb3)

	assert.True(t, FastOr(rb1, rb2, rb3).Equals(rb1))
	assert.True(t, HeapOr(rb1, rb2, rb3).Equals(rb1))
	assert.Equal(t, rb1.GetCardinality(), HeapOr(rb1, rb2, rb3).GetCardinality())
	assert.True(t, HeapXor(rb1, rb2, rb3).Equals(bigxor))
	assert.True(t, FastAnd(rb1, rb2, rb3).Equals(bigand))
}

func TestFastAggregationsXOR(t *testing.T) {
	rb1 := NewBitmap()
	rb2 := NewBitmap()
	rb3 := NewBitmap()

	for i := uint32(0); i < 40000; i++ {
		rb1.Add(i)
	}
	for i := uint32(0); i < 40000; i += 4000 {
		rb2.Add(i)
	}
	for i := uint32(0); i < 40000; i += 5000 {
		rb3.Add(i)
	}

	assert.EqualValues(t, 40000, rb1.GetCardinality())

	xor1 := Xor(rb1, rb2)
	xor1alt := Xor(rb2, rb1)
	assert.True(t, xor1alt.Equals(xor1))
	assert.True(t, HeapXor(rb1, rb2).Equals(xor1))

	xor2 := Xor(rb2, rb3)
	xor2alt := Xor(rb3, rb2)
	assert.True(t, xor2alt.Equals(xor2))
	assert.True(t, HeapXor(rb2, rb3).Equals(xor2))

	bigxor := Xor(Xor(rb1, rb2), rb3)
	bigxoralt1 := Xor(rb1, Xor(rb2, rb3))
	bigxoralt2 := Xor(rb1, Xor(rb3, rb2))
	bigxoralt3 := Xor(rb3, Xor(rb1, rb2))
	bigxoralt4 := Xor(Xor(rb1, rb2), rb3)

	assert.True(t, bigxoralt2.Equals(bigxor))
	assert.True(t, bigxoralt1.Equals(bigxor))
	assert.True(t, bigxoralt3.Equals(bigxor))
	assert.True(t, bigxoralt4.Equals(bigxor))

	assert.True(t, HeapXor(rb1, rb2, rb3).Equals(bigxor))
}

func TestFastAggregationsXOR_run(t *testing.T) {
	rb1 := NewBitmap()
	rb2 := NewBitmap()
	rb3 := NewBitmap()

	for i := uint32(0); i < 40000; i++ {
		rb1.Add(i)
	}
	rb1.RunOptimize()
	for i := uint32(0); i < 40000; i += 4000 {
		rb2.Add(i)
	}
	for i := uint32(0); i < 40000; i += 5000 {
		rb3.Add(i)
	}

	assert.EqualValues(t, 40000, rb1.GetCardinality())

	xor1 := Xor(rb1, rb2)
	xor1alt := Xor(rb2, rb1)
	assert.True(t, xor1alt.Equals(xor1))
	assert.True(t, HeapXor(rb1, rb2).Equals(xor1))

	xor2 := Xor(rb2, rb3)
	xor2alt := Xor(rb3, rb2)
	assert.True(t, xor2alt.Equals(xor2))
	assert.True(t, HeapXor(rb2, rb3).Equals(xor2))

	bigxor := Xor(Xor(rb1, rb2), rb3)
	bigxoralt1 := Xor(rb1, Xor(rb2, rb3))
	bigxoralt2 := Xor(rb1, Xor(rb3, rb2))
	bigxoralt3 := Xor(rb3, Xor(rb1, rb2))
	bigxoralt4 := Xor(Xor(rb1, rb2), rb3)

	assert.True(t, bigxoralt2.Equals(bigxor))
	assert.True(t, bigxoralt1.Equals(bigxor))
	assert.True(t, bigxoralt3.Equals(bigxor))
	assert.True(t, bigxoralt4.Equals(bigxor))

	assert.True(t, HeapXor(rb1, rb2, rb3).Equals(bigxor))
}

func TestFastAggregationsAndAny(t *testing.T) {
	base := NewBitmap()
	rb1 := NewBitmap()
	rb2 := NewBitmap()
	rb3 := NewBitmap()
	// only one filter has some values
	from := uint32(maxCapacity * 4)
	for i := uint32(from); i < from+100; i += 2 {
		rb1.Add(i)
	}
	// only base has values
	from = maxCapacity * 7
	for i := uint32(from); i < from+100; i += 2 {
		base.Add(i)
	}
	// base and one of filters have same values
	from = maxCapacity * 8
	for i := uint32(from); i < from+100; i += 2 {
		base.Add(i)
		rb1.Add(i)
	}
	// small union
	from = maxCapacity * 10
	for i := uint32(from); i < from+1000; i += 10 {
		base.Add(i)
		base.Add(i + i%3)

		rb1.Add(i)
		rb1.Add(i + 1)

		rb2.Add(i + 2)
		rb2.Add(i + i%7)

		rb3.Add(200 + i)
	}
	// run filters
	from = maxCapacity * 10
	for i := uint32(from); i < from+1000; i += 3 {
		base.Add(i)
	}
	for i := uint32(from); i < from+100; i++ {
		rb1.Add(i)
		rb2.Add(i + 333)
		rb3.Add(i + 433)
	}
	// large union
	from = maxCapacity * 16
	for i := uint32(from); i < from+arrayDefaultMaxSize*10; i += 3 {
		base.Add(i)
		base.Add(i + i%2 + 1)
		rb2.Add(i)
		rb3.Add(i + 1)
	}

	// some extra base values
	from = maxCapacity * 17
	for i := uint32(from); i < from+1000; i++ {
		base.Add(i)
	}

	base.RunOptimize()
	rb1.RunOptimize()
	rb2.RunOptimize()
	rb3.RunOptimize()

	orFirst := base.Clone()
	orFirst.And(FastOr(rb1, rb2, rb3))

	fast := base.Clone()
	fast.AndAny(rb1, rb2, rb3)

	assert.True(t, fast.Equals(orFirst))
}
