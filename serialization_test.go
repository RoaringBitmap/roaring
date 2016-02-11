package roaring

// to run just these tests: go test -run TestSerialization*

import (
	"bytes"
	"testing"
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
	l := int(rb.GetSerializedSizeInBytes())
	buf := new(bytes.Buffer)
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

func TestSerializationBasic2(t *testing.T) {
	rb := BitmapOf(1, 2, 3, 4, 5, 100, 1000, 10000, 100000, 1000000)
	buf := new(bytes.Buffer)
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
	rb := BitmapOf(1, 2, 3, 4, 5, 100, 1000, 10000, 100000, 1000000)
	for i := 5000000; i < 5000000+2*(1<<16); i++ {
		rb.AddInt(i)
	}
	l := int(rb.GetSerializedSizeInBytes())
	buf := new(bytes.Buffer)
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
