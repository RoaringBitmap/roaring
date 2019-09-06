package roaring

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
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
			assert.EqualValues(t,  2097162, p) // 32 << 16 | 10
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
