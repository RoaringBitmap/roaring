package roaring64

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
	//assert.EqualValues(t, buf.Len(), rb.GetSerializedSizeInBytes())

	newrb := NewBitmap()
	_, err = newrb.ReadFrom(buf)

	assert.NoError(t, err)
	assert.True(t, rb.Equals(newrb))
}

func TestSerializationBasic037(t *testing.T) {
	rb := BitmapOf(1, 2, 3, 4, 5, 100, 1000)

	buf := &bytes.Buffer{}
	_, err := rb.WriteTo(buf)

	assert.NoError(t, err)
	//assert.EqualValues(t, buf.Len(), rb.GetSerializedSizeInBytes())

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
	_ = l
	//assert.EqualValues(t, l, rb.GetSerializedSizeInBytes())

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

func TestBitmap_FromBuffer(t *testing.T) {
	t.Run("empty bitmap", func(t *testing.T) {
		rb := NewBitmap()

		buf := &bytes.Buffer{}
		_, err := rb.WriteTo(buf)

		assert.NoError(t, err)
		//assert.EqualValues(t, buf.Len(), rb.GetSerializedSizeInBytes())

		newRb := NewBitmap()
		newRb.FromBuffer(buf.Bytes())

		assert.NoError(t, err)
		assert.True(t, rb.Equals(newRb))
	})

	t.Run("basic bitmap of 7 elements", func(t *testing.T) {
		rb := BitmapOf(1, 2, 3, 4, 5, 100, 1000)

		buf := &bytes.Buffer{}
		_, err := rb.WriteTo(buf)

		assert.NoError(t, err)

		newRb := NewBitmap()
		_, err = newRb.FromBuffer(buf.Bytes())

		assert.NoError(t, err)
		assert.True(t, rb.Equals(newRb))
	})
}

func TestBitmapFromBufferCOW(t *testing.T) {
	rbbogus := NewBitmap()
	rbbogus.Add(100)
	rbbogus.Add(100000)
	rb1 := NewBitmap()
	rb1.Add(1)
	buf1 := &bytes.Buffer{}
	rb1.WriteTo(buf1)
	rb2 := NewBitmap()
	rb2.Add(1000000)
	buf2 := &bytes.Buffer{}
	rb2.WriteTo(buf2)
	newRb1 := NewBitmap()
	newRb1.FromBuffer(buf1.Bytes())
	newRb2 := NewBitmap()
	newRb2.FromBuffer(buf2.Bytes())
	rbor1 := Or(newRb1, newRb2)
	rbor2 := rbor1
	rbor3 := Or(newRb1, newRb2)
	rbor1.CloneCopyOnWriteContainers()
	rbor2.CloneCopyOnWriteContainers()
	rbor3.CloneCopyOnWriteContainers()
	buf1.Reset()
	buf2.Reset()
	rbbogus.WriteTo(buf1)
	rbbogus.WriteTo(buf2)
	rbexpected := NewBitmap()
	rbexpected.Add(1)
	rbexpected.Add(1000000)

	assert.True(t, rbexpected.Equals(rbor2))
	assert.True(t, rbexpected.Equals(rbor3))
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

func BenchmarkUnserializeFromBuffer(b *testing.B) {
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

		b.Run(fmt.Sprintf("FromBuffer-%d", size), func(b *testing.B) {
			b.ReportAllocs()
			b.StartTimer()

			for n := 0; n < b.N; n++ {
				nb := New()

				if _, err := nb.FromBuffer(buf.Bytes()); err != nil {
					b.Fatalf("Unexpected error occurs: %v", err)
				}
			}

			b.StopTimer()
		})
	}
}
