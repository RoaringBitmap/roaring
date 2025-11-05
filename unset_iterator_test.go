package roaring

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnsetIterator(t *testing.T) {
	t.Run("empty bitmap", func(t *testing.T) {
		bm := New()
		iter := bm.UnsetIterator()

		// First few unset values should be 0, 1, 2...
		assert.True(t, iter.HasNext())
		assert.Equal(t, uint32(0), iter.Next())
		assert.Equal(t, uint32(1), iter.Next())
		assert.Equal(t, uint32(2), iter.Next())
	})

	t.Run("bitmap with some values", func(t *testing.T) {
		bm := New()
		bm.Add(1)
		bm.Add(3)
		bm.Add(5)

		iter := bm.UnsetIterator()
		expected := []uint32{0, 2, 4, 6, 7, 8, 9, 10}
		for _, exp := range expected {
			assert.True(t, iter.HasNext())
			assert.Equal(t, exp, iter.Next())
		}
	})

	t.Run("bitmap with range", func(t *testing.T) {
		bm := New()
		bm.AddRange(10, 20)

		iter := bm.UnsetIterator()
		// First 10 should be unset
		for i := uint32(0); i < 10; i++ {
			assert.True(t, iter.HasNext())
			assert.Equal(t, i, iter.Next())
		}
		// 20-29 should be unset
		for i := uint32(20); i < 30; i++ {
			assert.True(t, iter.HasNext())
			assert.Equal(t, i, iter.Next())
		}
	})

	t.Run("bitmap with multiple containers", func(t *testing.T) {
		bm := New()
		// Add some values in first container (0-65535)
		bm.Add(100)
		bm.Add(200)
		// Add some values in second container (65536-131071)
		bm.Add(65636)
		bm.Add(65736)

		iter := bm.UnsetIterator()
		// Check first few unset values
		for i := uint32(0); i < 100; i++ {
			assert.True(t, iter.HasNext())
			assert.Equal(t, i, iter.Next())
		}
		// 100 is set, so skip it
		assert.Equal(t, uint32(101), iter.Next())
		// Continue to 199
		for i := uint32(102); i < 200; i++ {
			assert.Equal(t, i, iter.Next())
		}
		// 200 is set, skip it
		assert.Equal(t, uint32(201), iter.Next())
	})

	t.Run("bitmap with gap containers", func(t *testing.T) {
		bm := New()
		// Add value in container 0
		bm.Add(100)
		// Add value in container 2 (skip container 1)
		bm.Add(131072)

		iter := bm.UnsetIterator()
		// First 100 values are unset
		for i := uint32(0); i < 100; i++ {
			assert.True(t, iter.HasNext())
			assert.Equal(t, i, iter.Next())
		}
		// 100 is set, so next unset is 101
		assert.Equal(t, uint32(101), iter.Next())

		// Skip ahead to check gap container
		iter.AdvanceIfNeeded(65536)
		// All of container 1 (65536-131071) should be unset
		assert.True(t, iter.HasNext())
		assert.Equal(t, uint32(65536), iter.Next())
		assert.Equal(t, uint32(65537), iter.Next())
	})
}

func TestUnsetIteratorPeekNext(t *testing.T) {
	bm := New()
	bm.Add(1)
	bm.Add(3)
	bm.Add(5)

	iter := bm.UnsetIterator()
	assert.True(t, iter.HasNext())

	for iter.HasNext() {
		peek := iter.PeekNext()
		next := iter.Next()
		assert.Equal(t, peek, next)
		if next > 100 {
			break
		}
	}
}

func TestUnsetIteratorAdvanceIfNeeded(t *testing.T) {
	bm := New()
	bm.Add(10)
	bm.Add(100)
	bm.Add(1000)

	cases := []struct {
		minval   uint32
		expected uint32
	}{
		{0, 0},
		{5, 5},
		{10, 11}, // 10 is set
		{50, 50},
		{100, 101}, // 100 is set
		{500, 500},
		{1000, 1001}, // 1000 is set
	}

	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			iter := bm.UnsetIterator()
			iter.AdvanceIfNeeded(c.minval)
			assert.True(t, iter.HasNext())
			assert.Equal(t, c.expected, iter.Next())
		})
	}
}

func TestUnsetIteratorComplement(t *testing.T) {
	// Test that Iterator and UnsetIterator are complementary
	bm := New()
	for i := uint32(0); i < 1000; i += 3 {
		bm.Add(i)
	}

	// Collect all set values
	setValues := make(map[uint32]bool)
	iter := bm.Iterator()
	for iter.HasNext() {
		setValues[iter.Next()] = true
	}

	// Verify unset iterator returns complement up to 1000
	unsetIter := bm.UnsetIterator()
	count := 0
	for count < 1000 && unsetIter.HasNext() {
		val := unsetIter.Next()
		if val >= 1000 {
			break
		}
		assert.False(t, setValues[val])
		count++
	}
}

func TestUnsetIteratorLargeRange(t *testing.T) {
	// This test demonstrates the bug where UnsetIterator returns set bits as unset
	bm := New()
	bm.AddRange(0, 0x10000) // All bits from 0 to 65535 are set

	// Debug: check what type of container was created
	if len(bm.highlowcontainer.containers) != 1 {
		t.Fatalf("expected 1 container, got %d", len(bm.highlowcontainer.containers))
	}
	container := bm.highlowcontainer.containers[0]
	t.Logf("Container type: %T", container)
	t.Logf("Container cardinality: %d", container.getCardinality())

	unsetIter := bm.UnsetIterator()
	// Check initial state
	if unsetIter.HasNext() {
		t.Logf("Initial PeekNext: %d", unsetIter.PeekNext())
	}

	// Advance to position 100, which is in the middle of the set range
	unsetIter.AdvanceIfNeeded(100)

	if !unsetIter.HasNext() {
		t.Fatal("iterator should have next")
	}
	// Check state after advance
	t.Logf("After AdvanceIfNeeded(100), PeekNext: %d", unsetIter.PeekNext())

	// The next unset bit should be 65536 (0x10000), not 100
	val := unsetIter.Next()
	assert.Equal(t, uint32(0x10000), val, "expected first unset bit after range, got bit within set range")
}
