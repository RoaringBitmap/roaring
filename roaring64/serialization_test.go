package roaring64

// to run just these tests: go test -run TestSerialization*

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSerializationOfEmptyBitmap(t *testing.T) {
	rb := NewBitmap()

	buf := &bytes.Buffer{}
	_, err := rb.WriteTo(buf)

	require.NoError(t, err)
	assert.EqualValues(t, buf.Len(), rb.GetSerializedSizeInBytes())
	data := buf.Bytes()

	newrb := NewBitmap()
	_, err = newrb.ReadFrom(buf)

	require.NoError(t, err)
	assert.True(t, rb.Equals(newrb))

	newrb2 := NewBitmap()
	_, err = newrb2.FromUnsafeBytes(data)
	require.NoError(t, err)
	assert.True(t, rb.Equals(newrb2))
}

func TestBase64_036(t *testing.T) {
	rb := BitmapOf(1, 2, 3, 4, 5, 100, 1000)

	bstr, _ := rb.ToBase64()
	assert.NotEmpty(t, bstr)

	newrb := NewBitmap()

	_, err := newrb.FromBase64(bstr)

	require.NoError(t, err)
	assert.True(t, rb.Equals(newrb))
}

func TestSerializationBasic037(t *testing.T) {
	rb := BitmapOf(1, 2, 3, 4, 5, 100, 1000)

	buf := &bytes.Buffer{}
	_, err := rb.WriteTo(buf)

	require.NoError(t, err)
	assert.EqualValues(t, buf.Len(), rb.GetSerializedSizeInBytes())
	data := buf.Bytes()

	newrb := NewBitmap()
	_, err = newrb.ReadFrom(buf)

	require.NoError(t, err)
	assert.True(t, rb.Equals(newrb))

	newrb2 := NewBitmap()
	_, err = newrb2.FromUnsafeBytes(data)
	require.NoError(t, err)
	assert.True(t, rb.Equals(newrb2))
}

func TestSerializationToFile038(t *testing.T) {
	rb := BitmapOf(1, 2, 3, 4, 5, 100, 1000)
	fname := "myfile.bin"
	fout, err := os.OpenFile(fname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0660)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n\nIMPORTANT: For testing file IO, the roaring library requires disk access.\nWe omit some tests for now.\n\n")
		return
	}

	var l int64
	l, err = rb.WriteTo(fout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n\nIMPORTANT: For testing file IO, the roaring library requires disk access.\nWe omit some tests for now.\n\n")
		return
	}

	assert.EqualValues(t, l, rb.GetSerializedSizeInBytes())

	fout.Close()

	newrb := NewBitmap()
	fin, err := os.Open(fname)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n\nIMPORTANT: For testing file IO, the roaring library requires disk access.\nWe omit some tests for now.\n\n")
		return
	}
	buf := bytes.NewBuffer(nil)
	teer := io.TeeReader(fin, buf)

	defer func() {
		fin.Close()
		_ = os.Remove(fname)
	}()

	_, _ = newrb.ReadFrom(teer)
	assert.True(t, rb.Equals(newrb))

	newrb2 := NewBitmap()
	_, err = newrb2.FromUnsafeBytes(buf.Bytes())
	require.NoError(t, err)
	assert.True(t, rb.Equals(newrb2))
}

func TestSerializationBasic2_041(t *testing.T) {
	rb := BitmapOf(1, 2, 3, 4, 5, 100, 1000, 10000, 100000, 1000000)
	buf := &bytes.Buffer{}

	l := int(rb.GetSerializedSizeInBytes())
	_, err := rb.WriteTo(buf)

	require.NoError(t, err)
	assert.Equal(t, l, buf.Len())
	data := buf.Bytes()

	newrb := NewBitmap()
	_, err = newrb.ReadFrom(buf)

	require.NoError(t, err)
	assert.True(t, rb.Equals(newrb))

	newrb2 := NewBitmap()
	_, err = newrb2.FromUnsafeBytes(data)
	require.NoError(t, err)
	assert.True(t, rb.Equals(newrb2))
}

// roaringarray.writeTo and .readFrom should serialize and unserialize when containing all 3 container types
func TestSerializationBasic3_042(t *testing.T) {
	rb := BitmapOf(1, 2, 3, 4, 5, 100, 1000, 10000, 100000, 1000000, maxUint32+10, maxUint32<<10)
	for i := uint64(maxUint32); i < maxUint32+2*(1<<16); i++ {
		rb.Add(i)
	}

	var buf bytes.Buffer
	_, err := rb.WriteTo(&buf)

	require.NoError(t, err)
	assert.EqualValues(t, buf.Len(), int(rb.GetSerializedSizeInBytes()))
	data := buf.Bytes()

	newrb := NewBitmap()
	_, err = newrb.ReadFrom(&buf)

	require.NoError(t, err)
	assert.True(t, newrb.Equals(rb))

	newrb2 := NewBitmap()
	_, err = newrb2.FromUnsafeBytes(data)
	require.NoError(t, err)
	assert.True(t, rb.Equals(newrb2))
}

func TestHoldReference(t *testing.T) {
	t.Run("Test Hold Reference", func(t *testing.T) {
		rb := New()
		buf := &bytes.Buffer{}

		for i := uint64(0); i < 650; i++ {
			rb.Add(i)
		}

		_, err := rb.WriteTo(buf)
		require.NoError(t, err)

		nb := New()
		data := buf.Bytes()
		_, err = nb.ReadFrom(bytes.NewReader(data))

		require.NoError(t, err)

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

func BenchmarkUnserializeFromUnsafeBytes(b *testing.B) {
	benchmarkUnserializeFunc(b, "FromUnsafeBytes", func(bitmap *Bitmap, data []byte) (int64, error) {
		copied := make([]byte, len(data))
		copy(copied, data)
		return bitmap.FromUnsafeBytes(copied)
	})
}

func BenchmarkUnserializeReadFrom(b *testing.B) {
	benchmarkUnserializeFunc(b, "ReadFrom", func(bitmap *Bitmap, data []byte) (int64, error) {
		return bitmap.ReadFrom(bytes.NewReader(data))
	})
}

func benchmarkUnserializeFunc(b *testing.B, name string, f func(*Bitmap, []byte) (int64, error)) {
	b.Helper()

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

		b.Run(fmt.Sprintf("%s-%d", name, size), func(b *testing.B) {
			b.ReportAllocs()
			b.StartTimer()

			for n := 0; n < b.N; n++ {
				nb := New()

				if _, err := f(nb, buf.Bytes()); err != nil {
					b.Fatalf("Unexpected error occurs: %v", err)
				}
			}

			b.StopTimer()
		})
	}
}

func Test_tryReadFromRoaring32WithRoaring64(t *testing.T) {
	r64 := BitmapOf(1, 65535, math.MaxUint32, math.MaxUint64)
	bs, err := r64.ToBytes()
	if err != nil {
		t.Fatal(err)
	}
	nr64 := NewBitmap()
	assert.True(t, nr64.UnmarshalBinary(bs) == nil)
	assert.True(t, nr64.Contains(1))
	assert.True(t, nr64.Contains(65535))
	assert.True(t, nr64.Contains(math.MaxUint32))
	assert.True(t, nr64.Contains(math.MaxUint64))
}

func Test_tryReadFromRoaring32WithRoaring64_File(t *testing.T) {
	tempDir, err := ioutil.TempDir("./", "testdata")
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n\nIMPORTANT: For testing file IO, the roaring library requires disk access.\nWe omit some tests for now.\n\n")
		return
	}
	defer os.RemoveAll(tempDir)

	r64 := BitmapOf(1, 65535, math.MaxUint32, math.MaxUint64)
	bs, err := r64.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	name := filepath.Join(tempDir, "r32")
	if err := ioutil.WriteFile(name, bs, 0600); err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n\nIMPORTANT: For testing file IO, the roaring library requires disk access.\nWe omit some tests for now.\n\n")
		return
	}
	defer file.Close()

	nr64 := NewBitmap()
	nr64.ReadFrom(file)
	assert.True(t, nr64.Contains(1))
	assert.True(t, nr64.Contains(65535))
	assert.True(t, nr64.Contains(math.MaxUint32))
	assert.True(t, nr64.Contains(math.MaxUint64))
}
