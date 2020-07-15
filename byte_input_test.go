package roaring

import (
	"bytes"
	"encoding/binary"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"unsafe"
)

func TestByteInputFlow(t *testing.T) {
	t.Run("Test should be an error on empty data", func(t *testing.T) {
		buf := bytes.NewBuffer([]byte{})

		instances := []byteInput{
			newByteInput(buf.Bytes()),
			newByteInputFromReader(buf),
		}

		for _, input := range instances {
			n, err := input.readUInt16()

			assert.EqualValues(t, 0, n)
			assert.Error(t, err)

			p, err := input.readUInt32()
			assert.EqualValues(t, 0, p)
			assert.Error(t, err)

			b, err := input.next(10)
			assert.Nil(t, b)
			assert.Error(t, err)

			err = input.skipBytes(10)
			assert.Error(t, err)
		}
	})

	t.Run("Test on nonempty data", func(t *testing.T) {
		buf := bytes.NewBuffer(uint16SliceAsByteSlice([]uint16{1, 10, 32, 66, 23}))

		instances := []byteInput{
			newByteInput(buf.Bytes()),
			newByteInputFromReader(buf),
		}

		for _, input := range instances {
			n, err := input.readUInt16()
			assert.EqualValues(t, 1, n)
			assert.NoError(t, err)

			p, err := input.readUInt32()
			assert.EqualValues(t, 2097162, p) // 32 << 16 | 10
			assert.NoError(t, err)

			b, err := input.next(2)
			assert.EqualValues(t, []byte{66, 0}, b)
			assert.NoError(t, err)

			err = input.skipBytes(2)
			assert.NoError(t, err)

			b, err = input.next(1)
			assert.Nil(t, b)
			assert.Error(t, err)
		}
	})
}

func BenchmarkUint32SafeReading(b *testing.B) {
	n := 1 << 20
	arr := make([]byte, n*4)
	nums := make([]uint32, n)
	for j := 0; j < n; j++ {
		rnd := rand.Int31()
		nums[j] = uint32(rnd)
		binary.LittleEndian.PutUint32(arr[4*j:4*(j+1)], uint32(rnd))
	}
	b.ResetTimer()
	val := uint32(0)
	pointer := uint64(0)
	for pointer < uint64(len(arr)) {
		var newVal uint32
		newVal = binary.LittleEndian.Uint32(arr[pointer:])
		pointer += 4
		val += newVal
	}
}

func BenchmarkUint32UnsafeReading(b *testing.B) {
	n := 1 << 20
	arr := make([]byte, n*4)
	nums := make([]uint32, n)
	for j := 0; j < n; j++ {
		rnd := rand.Int31()
		nums[j] = uint32(rnd)
		binary.LittleEndian.PutUint32(arr[4*j:4*(j+1)], uint32(rnd))
	}
	b.ResetTimer()
	val := uint32(0)
	pointer := uint64(0)
	for pointer < uint64(len(arr)) {
		var newVal uint32
		newVal = *(*uint32)(unsafe.Pointer(&arr[pointer]))
		pointer += 4
		val += newVal
	}
}
