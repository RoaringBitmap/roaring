package roaring

// to run just these tests: go test -run TestSerialization*

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"math/rand"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBase64(t *testing.T) {
	rb := BitmapOf(1, 2, 3, 4, 5, 100, 1000)

	bstr, _ := rb.ToBase64()

	if bstr == "" {
		t.Errorf("ToBase64 failed returned empty string")
	}

	newrb := NewBitmap()

	_, err := newrb.FromBase64(bstr)

	if err != nil {
		t.Errorf("Failed reading from base64 string")
	}

	if !rb.Equals(newrb) {
		t.Errorf("comparing the base64 to and from failed cannot retrieve serialized version")
	}
}

func TestSerializationBasic(t *testing.T) {
	rb := BitmapOf(1, 2, 3, 4, 5, 100, 1000)

	buf := &bytes.Buffer{}
	_, err := rb.WriteTo(buf)
	if err != nil {
		t.Errorf("Failed writing")
	}

	newrb := NewBitmap()
	_, err = newrb.ReadFrom(buf)
	if err != nil {
		t.Errorf("Failed reading")
	}
	if !rb.Equals(newrb) {
		t.Errorf("Cannot retrieve serialized version")
	}
}

func TestSerializationToFile(t *testing.T) {
	rb := BitmapOf(1, 2, 3, 4, 5, 100, 1000)
	fname := "myfile.bin"
	fout, err := os.OpenFile(fname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0660)
	if err != nil {
		t.Errorf("Can't open a file for writing")
	}
	defer fout.Close()
	_, err = rb.WriteTo(fout)
	if err != nil {
		t.Errorf("Failed writing")
	}
	newrb := NewBitmap()
	fin, err := os.Open(fname)

	if err != nil {
		t.Errorf("Failed reading")
	}
	defer func() {
		fin.Close()
		err := os.Remove(fname)
		if err != nil {
			t.Errorf("could not delete %s ", fname)
		}
	}()
	_, _ = newrb.ReadFrom(fin)
	if !rb.Equals(newrb) {
		t.Errorf("Cannot retrieve serialized version")
	}
}

func TestSerializationBasic4WriteAndReadFile(t *testing.T) {
	fname := "testdata/all3.msgp.snappy"

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
	if err != nil {
		t.Errorf("Failed creating '%s'", fname)
	}
	_, err = rb.WriteTo(fout)
	if err != nil {
		t.Errorf("Failed writing to '%s'", fname)
	}
	fout.Close()

	fin, err := os.Open(fname)
	if err != nil {
		t.Errorf("Failed to Open '%s'", fname)
	}
	defer func() {
		fin.Close()
	}()

	newrb := NewBitmap()
	_, err = newrb.ReadFrom(fin)
	if err != nil {
		t.Errorf("Failed reading from '%s'", fname)
	}
	if !rb.Equals(newrb) {
		t.Errorf("Bad serialization")
	}
}

func TestSerializationBasic2(t *testing.T) {
	rb := BitmapOf(1, 2, 3, 4, 5, 100, 1000, 10000, 100000, 1000000)
	buf := &bytes.Buffer{}
	sz := rb.GetSerializedSizeInBytes()
	ub := BoundSerializedSizeInBytes(rb.GetCardinality(), 1000001)
	if sz > ub+10 {
		t.Errorf("Bad GetSerializedSizeInBytes; sz=%v, upper-bound=%v", sz, ub)
	}
	l := int(rb.GetSerializedSizeInBytes())
	_, err := rb.WriteTo(buf)
	if err != nil {
		t.Errorf("Failed writing")
	}
	if l != buf.Len() {
		t.Errorf("Bad GetSerializedSizeInBytes")
	}
	newrb := NewBitmap()
	_, err = newrb.ReadFrom(buf)
	if err != nil {
		t.Errorf("Failed reading")
	}
	if !rb.Equals(newrb) {
		t.Errorf("Cannot retrieve serialized version")
	}
}

func TestSerializationBasic3(t *testing.T) {

	Convey("roaringarray.writeTo and .readFrom should serialize and unserialize when containing all 3 container types", t, func() {
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
				fmt.Errorf("Unrecognized container implementation: %T", cn)
			}
		}
		if !bc {
			t.Errorf("no bitmapContainer found, change your test input so we test all three!")
		}
		if !ac {
			t.Errorf("no arrayContainer found, change your test input so we test all three!")
		}
		if !rc {
			t.Errorf("no runContainer16 found, change your test input so we test all three!")
		}

		var buf bytes.Buffer
		_, err := rb.WriteTo(&buf)
		if err != nil {
			t.Errorf("Failed writing")
		}

		newrb := NewBitmap()
		_, err = newrb.ReadFrom(&buf)
		if err != nil {
			t.Errorf("Failed reading")
		}
		c1, c2 := rb.GetCardinality(), newrb.GetCardinality()
		So(c2, ShouldEqual, c1)
		So(newrb.Equals(rb), ShouldBeTrue)
		//fmt.Printf("\n Basic3: good: match on card = %v", c1)
	})
}

func TestGobcoding(t *testing.T) {
	rb := BitmapOf(1, 2, 3, 4, 5, 100, 1000)

	buf := new(bytes.Buffer)
	encoder := gob.NewEncoder(buf)
	err := encoder.Encode(rb)
	if err != nil {
		t.Errorf("Gob encoding failed")
	}

	var b Bitmap
	decoder := gob.NewDecoder(buf)
	err = decoder.Decode(&b)
	if err != nil {
		t.Errorf("Gob decoding failed")
	}

	if !b.Equals(rb) {
		t.Errorf("Decoded bitmap does not equal input bitmap")
	}
}

func TestSerializationRunContainerMsgpack028(t *testing.T) {

	Convey("runContainer writeTo and readFrom should return logically equivalent containers", t, func() {
		seed := int64(42)
		p("seed is %v", seed)
		rand.Seed(seed)

		trials := []trial{
			trial{n: 10, percentFill: .2, ntrial: 10},
			trial{n: 10, percentFill: .8, ntrial: 10},
			trial{n: 10, percentFill: .50, ntrial: 10},
			/*
				trial{n: 10, percentFill: .01, ntrial: 10},
				trial{n: 1000, percentFill: .50, ntrial: 10},
				trial{n: 1000, percentFill: .99, ntrial: 10},
			*/
		}

		tester := func(tr trial) {
			for j := 0; j < tr.ntrial; j++ {
				p("TestSerializationRunContainerMsgpack028 on check# j=%v", j)

				ma := make(map[int]bool)

				n := tr.n
				a := []uint16{}

				draw := int(float64(n) * tr.percentFill)
				//p("draw is %v", draw)
				for i := 0; i < draw; i++ {
					r0 := rand.Intn(n)
					a = append(a, uint16(r0))
					ma[r0] = true
				}

				orig := newRunContainer16FromVals(false, a...)

				// serialize
				var buf bytes.Buffer
				_, err := orig.writeTo(&buf)
				if err != nil {
					panic(err)
				}

				// deserialize
				restored := &runContainer16{}
				_, err = restored.readFrom(&buf)
				if err != nil {
					panic(err)
				}

				// and compare
				So(restored.equals(orig), ShouldBeTrue)

			}
			p("done with serialization of runContainer16 check for trial %#v", tr)
		}

		for i := range trials {
			tester(trials[i])
		}

	})
}
