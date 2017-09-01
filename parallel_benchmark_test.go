package roaring

import (
	"runtime"
	"sort"
	"testing"
	"math/rand"
)

type parOp int

const (
	opAnd parOp = iota
	opOr
)

var defaultWorkerCount int = runtime.NumCPU()

var GlobalParAggregator = NewParAggregator()

const defaultTaskQueueLength = 4096

type ParAggregator struct {
	taskQueue chan parTask
}

func NewParAggregator() ParAggregator {
	agg := ParAggregator{
		make(chan parTask, defaultTaskQueueLength),
	}

	for i := 0; i < defaultWorkerCount; i++ {
		go agg.worker()
	}

	return agg
}

func (aggregator ParAggregator) worker() {
	for task := range aggregator.taskQueue {
		var resultContainer container
		switch task.op {
		case opAnd:
			resultContainer = task.left.and(task.right)
		}

		if resultContainer.getCardinality() > 0 {
			task.result <- parResult{
				key:       task.key,
				container: resultContainer,
			}
		} else {
			task.result <- parResult{
				key:   task.key,
				empty: true,
			}
		}
	}
}

func (aggregator ParAggregator) Shutdown() {
	close(aggregator.taskQueue)
}

// TODO check if using pointer types for containers is possible
// As I understand Golang this would result in one less copy
type parTask struct {
	op          parOp
	key         uint16
	left, right container
	result      chan<- parResult
}

type parResult struct {
	key       uint16
	container container
	empty     bool
}

type parResultSlice []parResult

func (s parResultSlice) Len() int {
	return len(s)
}

func (s parResultSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s parResultSlice) Less(i, j int) bool {
	return s[i].key < s[j].key
}

func (aggregator ParAggregator) And(x1, x2 *Bitmap) *Bitmap {
	answer := NewBitmap()
	pos1 := 0
	pos2 := 0
	length1 := x1.highlowcontainer.size()
	length2 := x2.highlowcontainer.size()

	resultChan := make(chan parResult, 64)
	expectedResults := 0

main:
	for pos1 < length1 && pos2 < length2 {
		s1 := x1.highlowcontainer.getKeyAtIndex(pos1)
		s2 := x2.highlowcontainer.getKeyAtIndex(pos2)
		for {
			if s1 == s2 {
				C := x1.highlowcontainer.getContainerAtIndex(pos1)

				expectedResults++
				aggregator.taskQueue <- parTask{
					op:     opAnd,
					left:   C,
					right:  x2.highlowcontainer.getContainerAtIndex(pos2),
					result: resultChan,
				}

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

	results := make([]parResult, 0, expectedResults)

	for result := range resultChan {
		if !result.empty {
			results = append(results, result)
		}
		expectedResults--
		if expectedResults == 0 {
			close(resultChan)
			break
		}
	}
	sort.Sort(parResultSlice(results))

	for _, result := range results {
		answer.highlowcontainer.appendContainer(result.key, result.container, false)
	}

	return answer
}

func TestParAnd(t *testing.T) {
	left := BitmapOf(1, 2)
	right := BitmapOf(1)
	result := GlobalParAggregator.And(left, right)

	if !result.Equals(right) {
		t.Errorf("Result bitmap differs from expected: %v != %v", left, right)
	}
}

func BenchmarkIntersectionSparseParallel(b *testing.B) {
	b.StopTimer()
	initsize := 650000
	r := rand.New(rand.NewSource(0))

	s1 := NewBitmap()
	sz := 150 * 1000 * 1000
	for i := 0; i < initsize; i++ {
		s1.Add(uint32(r.Int31n(int32(sz))))
	}

	s2 := NewBitmap()
	sz = 100 * 1000 * 1000
	for i := 0; i < initsize; i++ {
		s2.Add(uint32(r.Int31n(int32(sz))))
	}

	b.StartTimer()
	card := uint64(0)
	for j := 0; j < b.N; j++ {
		s3 := GlobalParAggregator.And(s1, s2)
		card = card + s3.GetCardinality()
	}
}

func BenchmarkIntersectionSparseRoaring(b *testing.B) {
	b.StopTimer()
	initsize := 650000
	r := rand.New(rand.NewSource(0))

	s1 := NewBitmap()
	sz := 150 * 1000 * 1000
	for i := 0; i < initsize; i++ {
		s1.Add(uint32(r.Int31n(int32(sz))))
	}

	s2 := NewBitmap()
	sz = 100 * 1000 * 1000
	for i := 0; i < initsize; i++ {
		s2.Add(uint32(r.Int31n(int32(sz))))
	}

	b.StartTimer()
	card := uint64(0)
	for j := 0; j < b.N; j++ {
		s3 := And(s1, s2)
		card = card + s3.GetCardinality()
	}
}
