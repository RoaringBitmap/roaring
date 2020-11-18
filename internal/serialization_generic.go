// +build !amd64,!386,!arm,!arm64,!ppc64le,!mipsle,!mips64le,!mips64p32le,!wasm appengine

package internal

import (
	"encoding/binary"
)

// Uint64SliceAsByteSlice internal func
func Uint64SliceAsByteSlice(slice []uint64) []byte {
	by := make([]byte, len(slice)*8)

	for i, v := range slice {
		binary.LittleEndian.PutUint64(by[i*8:], v)
	}

	return by
}

// Uint16SliceAsByteSlice internal func
func Uint16SliceAsByteSlice(slice []uint16) []byte {
	by := make([]byte, len(slice)*2)

	for i, v := range slice {
		binary.LittleEndian.PutUint16(by[i*2:], v)
	}

	return by
}

// ByteSliceAsUint16Slice internal func
func ByteSliceAsUint16Slice(slice []byte) []uint16 {
	if len(slice)%2 != 0 {
		panic("Slice size should be divisible by 2")
	}

	b := make([]uint16, len(slice)/2)

	for i := range b {
		b[i] = binary.LittleEndian.Uint16(slice[2*i:])
	}

	return b
}

// ByteSliceAsUint64Slice internal func
func ByteSliceAsUint64Slice(slice []byte) []uint64 {
	if len(slice)%8 != 0 {
		panic("Slice size should be divisible by 8")
	}

	b := make([]uint64, len(slice)/8)

	for i := range b {
		b[i] = binary.LittleEndian.Uint64(slice[8*i:])
	}

	return b
}
