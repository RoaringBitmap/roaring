package roaring

import (
	"runtime"
	"sort"
	"testing"
)

type parOp int

const (
	opAnd parOp = iota
	opOr
)

// TODO check if using pointer types for containers is possible
// As I understand Golang this would result in one less copy
type parTask struct {
	op          parOp
	key         uint16
	left, right container
	passReady   bool
	result      chan<- parResult
}

type parResult struct {
	key       uint16
	container container
	ready     bool
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

var workerCount int = runtime.NumCPU()

const taskQueueLength = 4096

var taskQueue chan parTask = make(chan parTask, taskQueueLength)

func init() {
	for i := 0; i < workerCount; i++ {
		go worker()
	}
}

func worker() {
	for task := range taskQueue {
		var resultContainer container
		if task.passReady {
			task.result <- parResult{ready: true,}
		} else {
			switch task.op {
			case opAnd:
				resultContainer = task.left.and(task.right)
			}
			if resultContainer.getCardinality() > 0 {
				task.result <- parResult{
					key:       task.key,
					container: resultContainer,
				}
			}
		}
	}
}

func parAnd(x1, x2 *Bitmap) *Bitmap {
	answer := NewBitmap()
	pos1 := 0
	pos2 := 0
	length1 := x1.highlowcontainer.size()
	length2 := x2.highlowcontainer.size()

	resultChan := make(chan parResult, 64)
	maxExpectedResults := 0

main:
	for pos1 < length1 && pos2 < length2 {
		s1 := x1.highlowcontainer.getKeyAtIndex(pos1)
		s2 := x2.highlowcontainer.getKeyAtIndex(pos2)
		for {
			if s1 == s2 {
				C := x1.highlowcontainer.getContainerAtIndex(pos1)

				maxExpectedResults++
				taskQueue <- parTask{
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

	taskQueue <- parTask{passReady: true}
	results := make([]parResult, 0, maxExpectedResults)

	for result := range resultChan {
		if result.ready {
			break
		} else {
			results = append(results, result)
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
	result := parAnd(left, right)

	if !result.Equals(right) {
		t.Errorf("Result bitmap differs from expected")
	}
}
