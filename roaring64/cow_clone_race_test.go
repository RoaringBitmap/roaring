package roaring64

import (
	"sync"
	"testing"
)

func TestConcurrentCloneOfCOWBitmap(t *testing.T) {
	src := New()
	for i := uint64(0); i < 8; i++ {
		src.Add(i << 32)
	}
	src.SetCopyOnWrite(true)
	card := src.GetCardinality()

	start := make(chan struct{})
	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for i := 0; i < 100; i++ {
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
	clone.Add(1)
	if src.Contains(1) {
		t.Fatal("writing a clone changed the source")
	}
	if !clone.Contains(1) {
		t.Fatal("clone did not keep its own write")
	}
}
