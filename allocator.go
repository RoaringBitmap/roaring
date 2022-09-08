package roaring

// Allocator is the interface for allocating various datastructures used
// in this library. Its primary purpose it provides users with the ability
// to control individual allocations in a relatively non-invasive way.
type Allocator interface {
	AllocateBytes(size, capacity int) []byte
	AllocateUInt16s(size, capacity int) []uint16
}
