// +build amd64,!appengine

// This file tests the popcnt functions

package roaring

import (
	"testing"
)

func TestPopcntSlice(t *testing.T) {
	s := []uint64{2, 3, 5, 7, 11, 13, 17, 19, 23, 29}
	res := popcntSlice(s)
	expected := uint64(27)
	if res != expected {
		t.Errorf("Got %d, expected %d", res, expected)
	}
}

func TestPopcntMaskSlice(t *testing.T) {
	s := []uint64{2, 3, 5, 7, 11, 13, 17, 19, 23, 29}
	m := []uint64{31, 37, 41, 43, 47, 53, 59, 61, 67, 71}
	res := popcntMaskSlice(s, m)
	expected := uint64(9)
	if res != expected {
		t.Errorf("Got %d, expected %d", res, expected)
	}
}

func TestPopcntAndSlice(t *testing.T) {
	s := []uint64{2, 3, 5, 7, 11, 13, 17, 19, 23, 29}
	m := []uint64{31, 37, 41, 43, 47, 53, 59, 61, 67, 71}
	res := popcntAndSlice(s, m)
	expected := uint64(18)
	if res != expected {
		t.Errorf("Got %d, expected %d", res, expected)
	}
}

func TestPopcntOrSlice(t *testing.T) {
	s := []uint64{2, 3, 5, 7, 11, 13, 17, 19, 23, 29}
	m := []uint64{31, 37, 41, 43, 47, 53, 59, 61, 67, 71}
	res := popcntOrSlice(s, m)
	expected := uint64(50)
	if res != expected {
		t.Errorf("Got %d, expected %d", res, expected)
	}
}

func TestPopcntXorSlice(t *testing.T) {
	s := []uint64{2, 3, 5, 7, 11, 13, 17, 19, 23, 29}
	m := []uint64{31, 37, 41, 43, 47, 53, 59, 61, 67, 71}
	res := popcntXorSlice(s, m)
	expected := uint64(32)
	if res != expected {
		t.Errorf("Got %d, expected %d", res, expected)
	}
}
