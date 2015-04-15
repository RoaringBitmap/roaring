package roaring

// to run just these tests: go test -run TestFastAggregations*

import (
	"container/heap"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestFastAggregations(t *testing.T) {
	Convey("Fast", t, func() {
		rb1 := NewRoaringBitmap()
		rb2 := NewRoaringBitmap()
		rb1.Add(1)
		rb2.Add(2)

		So(FastAnd(rb1, rb2).GetCardinality(), ShouldEqual, 0)
		So(FastOr(rb1, rb2).GetCardinality(), ShouldEqual, 2)
		So(FastXor(rb1, rb2).GetCardinality(), ShouldEqual, 2)
	})
}

func TestFastAggregationsSize(t *testing.T) {
	Convey("Fast", t, func() {
		rb1 := NewRoaringBitmap()
		rb2 := NewRoaringBitmap()
		rb3 := NewRoaringBitmap()
		for i := 0; i < 1000000; i += 3 {
			rb1.Add(i)
		}
		for i := 0; i < 1000000; i += 7 {
			rb2.Add(i)
		}
		for i := 0; i < 1000000; i += 1001 {
			rb3.Add(i)
		}
		pq := make(priorityQueue, 3)
		pq[0] = &item{rb1, 0}
		pq[1] = &item{rb2, 1}
		pq[2] = &item{rb3, 2}
		heap.Init(&pq)
		So(heap.Pop(&pq).(*item).value.GetSizeInBytes(), ShouldEqual, rb3.GetSizeInBytes())
		So(heap.Pop(&pq).(*item).value.GetSizeInBytes(), ShouldEqual, rb2.GetSizeInBytes())
		So(heap.Pop(&pq).(*item).value.GetSizeInBytes(), ShouldEqual, rb1.GetSizeInBytes())
	})
}

func TestFastAggregationsCont(t *testing.T) {
	Convey("Fast", t, func() {
		rb1 := NewRoaringBitmap()
		rb2 := NewRoaringBitmap()
		rb3 := NewRoaringBitmap()
		for i := 0; i < 10; i += 3 {
			rb1.Add(i)
		}
		for i := 0; i < 10; i += 7 {
			rb2.Add(i)
		}
		for i := 0; i < 10; i += 1001 {
			rb3.Add(i)
		}
		for i := 1000000; i < 1000000+10; i += 1001 {
			rb1.Add(i)
		}
		for i := 1000000; i < 1000000+10; i += 7 {
			rb2.Add(i)
		}
		for i := 1000000; i < 1000000+10; i += 3 {
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
			So(x1.value.GetCardinality(), ShouldEqual, expected[counter])
			counter++
			x1.keyindex++
			if x1.keyindex < x1.value.highlowcontainer.size() {
				heap.Push(&pq, x1)
			}
		}
	})
}
func TestFastAggregationsAdvanced(t *testing.T) {
	Convey("Fast", t, func() {
		rb1 := NewRoaringBitmap()
		rb2 := NewRoaringBitmap()
		rb3 := NewRoaringBitmap()
		for i := 0; i < 1000000; i += 3 {
			rb1.Add(i)
		}
		for i := 0; i < 1000000; i += 7 {
			rb2.Add(i)
		}
		for i := 0; i < 1000000; i += 1001 {
			rb3.Add(i)
		}
		for i := 1000000; i < 2000000; i += 1001 {
			rb1.Add(i)
		}
		for i := 1000000; i < 2000000; i += 3 {
			rb2.Add(i)
		}
		for i := 1000000; i < 2000000; i += 7 {
			rb3.Add(i)
		}
		rb1.Or(rb2)
		rb1.Or(rb3)
		bigand := And(And(rb1, rb2), rb3)
		bigxor := Xor(Xor(rb1, rb2), rb3)
		So(FastOr(rb1, rb2, rb3).Equals(rb1), ShouldEqual, true)
		So(FastHorizontalOr(rb1, rb2, rb3).Equals(rb1), ShouldEqual, true)
		So(FastHorizontalOr(rb1, rb2, rb3).GetCardinality(), ShouldEqual, rb1.GetCardinality())
		So(FastXor(rb1, rb2, rb3).Equals(bigxor), ShouldEqual, true)
		So(FastAnd(rb1, rb2, rb3).Equals(bigand), ShouldEqual, true)
	})
}
