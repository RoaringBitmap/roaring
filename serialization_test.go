package roaring

// to run just these tests: go test -run TestSerialization*

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
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

func TestSerializationReadRunsFromFile039(t *testing.T) {
	fn := "testdata/bitmapwithruns.bin"

	by, err := ioutil.ReadFile(fn)
	assert.NoError(t, err)

	newrb := NewBitmap()
	_, err = newrb.ReadFrom(bytes.NewBuffer(by))

	assert.NoError(t, err)
}

func TestSerializationBasic4WriteAndReadFile040(t *testing.T) {
	fname := "testdata/all3.classic"

	rb := NewBitmap()
	for k := uint32(0); k < 100000; k += 1000 {
		rb.Add(k)
	}
	for k := uint32(100000); k < 200000; k++ {
		rb.Add(3 * k)
	}
	for k := uint32(700000); k < 800000; k++ {
		rb.Add(k)
	}

	rb.highlowcontainer.runOptimize()
	fout, err := os.Create(fname)

	assert.NoError(t, err)

	var l int64

	l, err = rb.WriteTo(fout)

	assert.NoError(t, err)
	assert.EqualValues(t, l, rb.GetSerializedSizeInBytes())

	fout.Close()
	fin, err := os.Open(fname)

	assert.NoError(t, err)

	defer fin.Close()

	newrb := NewBitmap()
	_, err = newrb.ReadFrom(fin)

	assert.NoError(t, err)
	assert.True(t, rb.Equals(newrb))
}

func TestSerializationFromJava051(t *testing.T) {
	fname := "testdata/bitmapwithoutruns.bin"
	newrb := NewBitmap()
	fin, err := os.Open(fname)

	assert.NoError(t, err)

	defer func() {
		fin.Close()
	}()

	_, _ = newrb.ReadFrom(fin)
	fmt.Println(newrb.GetCardinality())
	rb := NewBitmap()
	for k := uint32(0); k < 100000; k += 1000 {
		rb.Add(k)
	}
	for k := uint32(100000); k < 200000; k++ {
		rb.Add(3 * k)
	}
	for k := uint32(700000); k < 800000; k++ {
		rb.Add(k)
	}

	assert.True(t, rb.Equals(newrb))
}

func TestSerializationFromJavaWithRuns052(t *testing.T) {
	fname := "testdata/bitmapwithruns.bin"

	newrb := NewBitmap()
	fin, err := os.Open(fname)

	assert.NoError(t, err)

	defer func() {
		fin.Close()
	}()
	_, _ = newrb.ReadFrom(fin)
	rb := NewBitmap()

	for k := uint32(0); k < 100000; k += 1000 {
		rb.Add(k)
	}
	for k := uint32(100000); k < 200000; k++ {
		rb.Add(3 * k)
	}
	for k := uint32(700000); k < 800000; k++ {
		rb.Add(k)
	}

	assert.True(t, rb.Equals(newrb))
}

func TestSerializationBasic2_041(t *testing.T) {
	rb := BitmapOf(1, 2, 3, 4, 5, 100, 1000, 10000, 100000, 1000000)
	buf := &bytes.Buffer{}
	sz := rb.GetSerializedSizeInBytes()
	ub := BoundSerializedSizeInBytes(rb.GetCardinality(), 1000001)

	assert.False(t, sz > ub+10)

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
	rb := BitmapOf(1, 2, 3, 4, 5, 100, 1000, 10000, 100000, 1000000)
	for i := 5000000; i < 5000000+2*(1<<16); i++ {
		rb.AddInt(i)
	}

	// confirm all three types present
	var bc, ac, rc bool
	for _, v := range rb.highlowcontainer.containers {
		switch cn := v.(type) {
		case *bitmapContainer:
			bc = true
		case *arrayContainer:
			ac = true
		case *runContainer16:
			rc = true
		default:
			panic(fmt.Errorf("Unrecognized container implementation: %T", cn))
		}
	}

	assert.True(t, bc, "no bitmapContainer found, change your test input so we test all three!")
	assert.True(t, ac, "no arrayContainer found, change your test input so we test all three!")
	assert.True(t, rc, "no runContainer16 found, change your test input so we test all three!")

	var buf bytes.Buffer
	_, err := rb.WriteTo(&buf)

	assert.NoError(t, err)
	assert.EqualValues(t, buf.Len(), rb.GetSerializedSizeInBytes())

	newrb := NewBitmap()
	_, err = newrb.ReadFrom(&buf)

	assert.NoError(t, err)
	assert.Equal(t, rb.GetCardinality(), newrb.GetCardinality())
	assert.True(t, newrb.Equals(rb))
}

func TestGobcoding043(t *testing.T) {
	rb := BitmapOf(1, 2, 3, 4, 5, 100, 1000)

	buf := new(bytes.Buffer)
	encoder := gob.NewEncoder(buf)
	err := encoder.Encode(rb)

	assert.NoError(t, err)

	var b Bitmap
	decoder := gob.NewDecoder(buf)
	err = decoder.Decode(&b)

	assert.NoError(t, err)
	assert.True(t, b.Equals(rb))
}

// runContainer writeTo and readFrom should return logically equivalent containers
func TestSerializationRunContainerMsgpack028(t *testing.T) {
	seed := int64(42)
	rand.Seed(seed)

	trials := []trial{
		{n: 10, percentFill: .2, ntrial: 10},
		{n: 10, percentFill: .8, ntrial: 10},
		{n: 10, percentFill: .50, ntrial: 10},
	}

	tester := func(tr trial) {
		for j := 0; j < tr.ntrial; j++ {

			ma := make(map[int]bool)

			n := tr.n
			a := []uint16{}

			draw := int(float64(n) * tr.percentFill)
			for i := 0; i < draw; i++ {
				r0 := rand.Intn(n)
				a = append(a, uint16(r0))
				ma[r0] = true
			}

			orig := newRunContainer16FromVals(false, a...)

			// serialize
			var buf bytes.Buffer
			_, err := orig.writeToMsgpack(&buf)
			if err != nil {
				panic(err)
			}

			// deserialize
			restored := &runContainer16{}
			_, err = restored.readFromMsgpack(&buf)
			if err != nil {
				panic(err)
			}

			// and compare
			assert.True(t, restored.equals(orig))
		}
	}

	for i := range trials {
		tester(trials[i])
	}
}

//roaringarray.writeToMsgpack and .readFromMsgpack should serialize and unserialize when containing all 3 container types
func TestSerializationBasicMsgpack035(t *testing.T) {
	rb := BitmapOf(1, 2, 3, 4, 5, 100, 1000, 10000, 100000, 1000000)
	for i := 5000000; i < 5000000+2*(1<<16); i++ {
		rb.AddInt(i)
	}

	// confirm all three types present
	var bc, ac, rc bool
	for _, v := range rb.highlowcontainer.containers {
		switch cn := v.(type) {
		case *bitmapContainer:
			bc = true
			assert.Equal(t, bitmapContype, cn.containerType())
		case *arrayContainer:
			ac = true
			assert.Equal(t, arrayContype, cn.containerType())
		case *runContainer16:
			rc = true
			assert.Equal(t, run16Contype, cn.containerType())
		default:
			panic(fmt.Errorf("Unrecognized container implementation: %T", cn))
		}
	}

	assert.True(t, bc, "no bitmapContainer found, change your test input so we test all three!")
	assert.True(t, ac, "no arrayContainer found, change your test input so we test all three!")
	assert.True(t, rc, "no runContainer16 found, change your test input so we test all three!")

	var buf bytes.Buffer
	_, err := rb.WriteToMsgpack(&buf)

	assert.NoError(t, err)

	newrb := NewBitmap()
	_, err = newrb.ReadFromMsgpack(&buf)

	assert.NoError(t, err)
	assert.Equal(t, rb.GetCardinality(), newrb.GetCardinality())
	assert.True(t, newrb.Equals(rb))
}

func TestByteSliceAsUint16Slice(t *testing.T) {
	t.Run("valid slice", func(t *testing.T) {
		expectedSize := 2
		slice := make([]byte, 4)
		binary.LittleEndian.PutUint16(slice, 42)
		binary.LittleEndian.PutUint16(slice[2:], 43)

		uint16Slice := byteSliceAsUint16Slice(slice)

		assert.Equal(t, expectedSize, len(uint16Slice))
		assert.Equal(t, expectedSize, cap(uint16Slice))
		assert.False(t, uint16Slice[0] != 42 || uint16Slice[1] != 43)
	})

	t.Run("inlined", func(t *testing.T) {
		first, second := singleSliceInArray()
		t.Logf("received %v %v", first, second[0])
		if !first.Equals(second[0]) {
			t.Errorf("inline fail %v %v", first, second[0])
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		slice := make([]byte, 0, 0)
		uint16Slice := byteSliceAsUint16Slice(slice)

		assert.Equal(t, 0, len(uint16Slice))
		assert.Equal(t, 0, cap(uint16Slice))
	})

	t.Run("invalid slice size", func(t *testing.T) {
		slice := make([]byte, 1, 1)

		assert.Panics(t, func() {
			byteSliceAsUint16Slice(slice)
		})
	})
}

func singleSliceInArray() (*Bitmap, []*Bitmap) {
	firstSlice := singleSlice()
	containerSlice := make([]*Bitmap, 0)
	secondContainer := singleSlice()
	containerSlice = append(containerSlice, secondContainer)
	return firstSlice, containerSlice
}

func singleSlice() *Bitmap {
	slice := make([]byte, 2)
	return &Bitmap{highlowcontainer: roaringArray{keys: []uint16{0}, containers: []container{&arrayContainer{byteSliceAsUint16Slice(slice)}}}}
}

func TestByteSliceAsUint64Slice(t *testing.T) {
	t.Run("valid slice", func(t *testing.T) {
		expectedSize := 2
		slice := make([]byte, 16)
		binary.LittleEndian.PutUint64(slice, 42)
		binary.LittleEndian.PutUint64(slice[8:], 43)

		uint64Slice := byteSliceAsUint64Slice(slice)

		assert.Equal(t, expectedSize, len(uint64Slice))
		assert.Equal(t, expectedSize, cap(uint64Slice))
		assert.False(t, uint64Slice[0] != 42 || uint64Slice[1] != 43)
	})

	t.Run("empty slice", func(t *testing.T) {
		slice := make([]byte, 0, 0)
		uint64Slice := byteSliceAsUint64Slice(slice)

		assert.Equal(t, 0, len(uint64Slice))
		assert.Equal(t, 0, cap(uint64Slice))
	})

	t.Run("invalid slice size", func(t *testing.T) {
		slice := make([]byte, 1, 1)

		assert.Panics(t, func() {
			byteSliceAsUint64Slice(slice)
		})
	})
}

func TestByteSliceAsInterval16Slice(t *testing.T) {
	t.Run("valid slice", func(t *testing.T) {
		expectedSize := 2
		slice := make([]byte, 8)
		binary.LittleEndian.PutUint16(slice, 10)
		binary.LittleEndian.PutUint16(slice[2:], 2)
		binary.LittleEndian.PutUint16(slice[4:], 20)
		binary.LittleEndian.PutUint16(slice[6:], 2)

		intervalSlice := byteSliceAsInterval16Slice(slice)

		assert.Equal(t, expectedSize, len(intervalSlice))
		assert.Equal(t, expectedSize, cap(intervalSlice))

		i1 := newInterval16Range(10, 12)
		i2 := newInterval16Range(20, 22)

		assert.False(t, intervalSlice[0] != i1 || intervalSlice[1] != i2)
	})

	t.Run("empty slice", func(t *testing.T) {
		slice := make([]byte, 0, 0)
		intervalSlice := byteSliceAsInterval16Slice(slice)

		assert.Equal(t, 0, len(intervalSlice))
		assert.Equal(t, 0, cap(intervalSlice))
	})

	t.Run("invalid slice length", func(t *testing.T) {
		slice := make([]byte, 1, 1)

		assert.Panics(t, func() {
			byteSliceAsInterval16Slice(slice)
		})
	})
}

func TestBitmap_FromBuffer(t *testing.T) {
	t.Run("empty bitmap", func(t *testing.T) {
		rb := NewBitmap()

		buf := &bytes.Buffer{}
		_, err := rb.WriteTo(buf)

		assert.NoError(t, err)
		assert.EqualValues(t, buf.Len(), rb.GetSerializedSizeInBytes())

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

	t.Run("bitmap with runs", func(t *testing.T) {
		file := "testdata/bitmapwithruns.bin"

		buf, err := ioutil.ReadFile(file)
		assert.NoError(t, err)

		rb := NewBitmap()
		_, err = rb.FromBuffer(buf)

		assert.NoError(t, err)
		assert.EqualValues(t, 3, rb.Stats().RunContainers)
		assert.EqualValues(t, 11, rb.Stats().Containers)
	})

	t.Run("bitmap without runs", func(t *testing.T) {
		fn := "testdata/bitmapwithruns.bin"
		buf, err := ioutil.ReadFile(fn)

		assert.NoError(t, err)

		rb := NewBitmap()
		_, err = rb.FromBuffer(buf)

		assert.NoError(t, err)
	})

	// all3.classic somehow created by other tests.
	t.Run("all3.classic bitmap", func(t *testing.T) {
		file := "testdata/all3.classic"
		buf, err := ioutil.ReadFile(file)

		assert.NoError(t, err)

		rb := NewBitmap()
		_, err = rb.FromBuffer(buf)

		assert.NoError(t, err)
	})

	t.Run("testdata/bitmapwithruns.bin bitmap Ops", func(t *testing.T) {
		file := "testdata/bitmapwithruns.bin"
		buf, err := ioutil.ReadFile(file)

		assert.NoError(t, err)

		empt := NewBitmap()

		rb1 := NewBitmap()
		_, err = rb1.FromBuffer(buf)

		assert.NoError(t, err)

		rb2 := NewBitmap()
		_, err = rb2.FromBuffer(buf)

		assert.NoError(t, err)

		rbor := Or(rb1, rb2)
		rbfastor := FastOr(rb1, rb2)
		rband := And(rb1, rb2)
		rbxor := Xor(rb1, rb2)
		rbandnot := AndNot(rb1, rb2)

		assert.True(t, rbor.Equals(rb1))
		assert.True(t, rbfastor.Equals(rbor))
		assert.True(t, rband.Equals(rb1))
		assert.True(t, rbxor.Equals(empt))
		assert.True(t, rbandnot.Equals(empt))
	})

	t.Run("marking all containers as requiring COW", func(t *testing.T) {
		file := "testdata/bitmapwithruns.bin"
		buf, err := ioutil.ReadFile(file)

		assert.NoError(t, err)

		rb := NewBitmap()
		_, err = rb.FromBuffer(buf)

		assert.NoError(t, err)

		for i, cow := range rb.highlowcontainer.needCopyOnWrite {
			assert.Truef(t, cow, "Container at pos %d was not marked as needs-copy-on-write", i)
		}
	})
}

func TestSerializationCrashers(t *testing.T) {
	crashers, err := filepath.Glob("testdata/crash*")

	assert.NoError(t, err)

	for _, crasher := range crashers {
		data, err := ioutil.ReadFile(crasher)
		assert.NoError(t, err)

		// take a copy in case the stream is modified during unpacking attempt
		orig := make([]byte, len(data))
		copy(orig, data)

		_, err = NewBitmap().FromBuffer(data)
		assert.Error(t, err)

		// reset for next one
		copy(data, orig)
		_, err = NewBitmap().ReadFrom(bytes.NewReader(data))

		assert.Error(t, err)
	}
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

		for i := uint32(0); i < 650; i++ {
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
		i := uint32(0)

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
	for _, size := range []uint32{650, 6500, 65000, 650000, 6500000} {
		rb := New()
		buf := &bytes.Buffer{}

		for i := uint32(0); i < size; i++ {
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
	for _, size := range []uint32{650, 6500, 65000, 650000, 6500000} {
		rb := New()
		buf := &bytes.Buffer{}

		for i := uint32(0); i < size; i++ {
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

func BenchmarkUnserializeFromBufferSparse(b *testing.B) {
	for _, size := range []uint32{650, 6500, 65000} {
		rb := New()
		buf := &bytes.Buffer{}

		for i := uint32(0); i < size; i++ {
			rb.Add(i << 14)
		}

		_, err := rb.WriteTo(buf)

		if err != nil {
			b.Fatalf("Unexpected error occurs: %v", err)
		}
		b.N = 50

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
