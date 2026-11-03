package roaring

import (
	"sync"
	"testing"
)

func TestConcurrentCloneOfCOWBitmap(t *testing.T) {
	src := New()
	for i := uint32(0); i < 1<<20; i += 977 {
		src.Add(i)
	}
	src.SetCopyOnWrite(true)
	card := src.GetCardinality()

	const (
		goroutines = 8
		iterations = 100
	)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for i := 0; i < iterations; i++ {
				_ = src.Clone()
			}
		}()
	}
	close(start)
	wg.Wait()

	if src.GetCardinality() != card {
		t.Fatalf("cardinality changed: got %d, want %d", src.GetCardinality(), card)
	}
	clone := src.Clone()
	if clone.Contains(1) {
		t.Fatal("1 was already present")
	}
	clone.Add(1)
	if src.Contains(1) {
		t.Fatal("writing a clone changed the source")
	}
	if !clone.Contains(1) {
		t.Fatal("clone did not keep its own write")
	}
}
