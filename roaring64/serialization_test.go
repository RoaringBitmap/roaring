package roaring64

// to run just these tests: go test -run TestSerialization*

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSerializationOfEmptyBitmap(t *testing.T) {
	rb := NewBitmap()

	buf := &bytes.Buffer{}
	_, err := rb.WriteTo(buf)

	assert.NoError(t, err)
	assert.EqualValues(t, buf.Len(), rb.GetSerializedSizeInBytes())

	newrb := NewBitmap()
	_, err = newrb.ReadFrom(buf)

	assert.NoError(t, err)
	assert.True(t, rb.Equals(newrb))
}

func TestBase64_036(t *testing.T) {
	rb := BitmapOf(1, 2, 3, 4, 5, 100, 1000)

	bstr, _ := rb.ToBase64()
	assert.NotEmpty(t, bstr)

	newrb := NewBitmap()

	_, err := newrb.FromBase64(bstr)

	assert.NoError(t, err)
	assert.True(t, rb.Equals(newrb))
}

func TestSerializationBasic037(t *testing.T) {
	rb := BitmapOf(1, 2, 3, 4, 5, 100, 1000)

	buf := &bytes.Buffer{}
	_, err := rb.WriteTo(buf)

	assert.NoError(t, err)
	assert.EqualValues(t, buf.Len(), rb.GetSerializedSizeInBytes())

	newrb := NewBitmap()
	_, err = newrb.ReadFrom(buf)

	assert.NoError(t, err)
	assert.True(t, rb.Equals(newrb))
}

func TestSerializationToFile038(t *testing.T) {
	rb := BitmapOf(1, 2, 3, 4, 5, 100, 1000)
	fname := "myfile.bin"
	fout, err := os.OpenFile(fname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0660)

	assert.NoError(t, err)

	var l int64
	l, err = rb.WriteTo(fout)

	assert.NoError(t, err)
	assert.EqualValues(t, l, rb.GetSerializedSizeInBytes())

	fout.Close()

	newrb := NewBitmap()
	fin, err := os.Open(fname)

	assert.NoError(t, err)

	defer func() {
		fin.Close()
		assert.NoError(t, os.Remove(fname))
	}()

	_, _ = newrb.ReadFrom(fin)
	assert.True(t, rb.Equals(newrb))
}

func TestSerializationBasic2_041(t *testing.T) {
	rb := BitmapOf(1, 2, 3, 4, 5, 100, 1000, 10000, 100000, 1000000)
	buf := &bytes.Buffer{}

	l := int(rb.GetSerializedSizeInBytes())
	_, err := rb.WriteTo(buf)

	assert.NoError(t, err)
	assert.Equal(t, l, buf.Len())

	newrb := NewBitmap()
	_, err = newrb.ReadFrom(buf)

	assert.NoError(t, err)
	assert.True(t, rb.Equals(newrb))
}

// roaringarray.writeTo and .readFrom should serialize and unserialize when containing all 3 container types
func TestSerializationBasic3_042(t *testing.T) {
	rb := BitmapOf(1, 2, 3, 4, 5, 100, 1000, 10000, 100000, 1000000, maxUint32+10, maxUint32<<10)
	for i := uint64(maxUint32); i < maxUint32+2*(1<<16); i++ {
		rb.Add(i)
	}

	var buf bytes.Buffer
	_, err := rb.WriteTo(&buf)

	assert.NoError(t, err)
	assert.EqualValues(t, buf.Len(), int(rb.GetSerializedSizeInBytes()))

	newrb := NewBitmap()
	_, err = newrb.ReadFrom(&buf)

	assert.NoError(t, err)
	assert.True(t, newrb.Equals(rb))
}

func TestHoldReference(t *testing.T) {
	t.Run("Test Hold Reference", func(t *testing.T) {
		rb := New()
		buf := &bytes.Buffer{}

		for i := uint64(0); i < 650; i++ {
			rb.Add(i)
		}

		_, err := rb.WriteTo(buf)
		assert.NoError(t, err)

		nb := New()
		data := buf.Bytes()
		_, err = nb.ReadFrom(bytes.NewReader(data))

		assert.NoError(t, err)

		buf = nil
		rb = nil
		data = nil

		runtime.GC()

		iterator := nb.Iterator()
		i := uint64(0)

		for iterator.HasNext() {
			v := iterator.Next()

			if v != i {
				return
			}

			assert.Equal(t, i, v)
			i++
		}
	})
}

func BenchmarkUnserializeReadFrom(b *testing.B) {
	for _, size := range []uint64{650, 6500, 65000, 650000, 6500000} {
		rb := New()
		buf := &bytes.Buffer{}

		for i := uint64(0); i < size; i++ {
			rb.Add(i)
		}

		_, err := rb.WriteTo(buf)

		if err != nil {
			b.Fatalf("Unexpected error occurs: %v", err)
		}

		b.Run(fmt.Sprintf("ReadFrom-%d", size), func(b *testing.B) {
			b.ReportAllocs()
			b.StartTimer()

			for n := 0; n < b.N; n++ {
				reader := bytes.NewReader(buf.Bytes())
				nb := New()

				if _, err := nb.ReadFrom(reader); err != nil {
					b.Fatalf("Unexpected error occurs: %v", err)
				}
			}

			b.StopTimer()
		})
	}
}
