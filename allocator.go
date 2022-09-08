package roaring

// Allocator is the interface for allocating various datastructures used
// in this library. Its primary purpose it provides users with the ability
// to control individual allocations in a relatively non-invasive way.
type Allocator interface {
	AllocateBytes(size, capacity int) []byte
	AllocateUInt16s(size, capacity int) []uint16
}

// defaultAllocator implements Allocator by just deferring to the default
// Go allocator.
//
// This struct has non-pointer receivers so it does not require an additional
// allocation to be instantiated as part of a larger struct.
type defaultAllocator struct {
}

func (a defaultAllocator) AllocateBytes(size, capacity int) []byte {
	return make([]byte, size, capacity)
}

func (a defaultAllocator) AllocateUInt16s(size, capacity int) []uint16 {
	return make([]uint16, size, capacity)
}
