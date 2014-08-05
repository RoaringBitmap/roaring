package roaring

// to run just these tests: go test -run TestArrayContainer*

import (
	"testing"
)

func TestArrayContainerSetAndGet(t *testing.T) {
	v := Container(NewArrayContainer())
	v = v.Add(100)
	if v.GetCardinality() != 1 {
		t.Errorf("Bogus cardinality.")
	}
	for i := 0; i <= ARRAY_DEFAULT_MAX_SIZE; i++ {
		if i == 100 {
			if v.Contains(short(i)) != true {
				t.Errorf("I added a number in vain.")
			}
		} else {
			if v.Contains(short(i)) != false {
				t.Errorf("Ghost content")
				break
			}
		}
	}
}

func TestArrayContainerMassiveSetAndGet(t *testing.T) {
	v := Container(NewArrayContainer())
	for j := 0; j <= ARRAY_DEFAULT_MAX_SIZE; j++ {

		v = v.Add(short(j))
		if v.GetCardinality() != 1+j {
			t.Errorf("Bogus cardinality. ", v.GetCardinality(), j)
		}
		for i := 0; i <= ARRAY_DEFAULT_MAX_SIZE; i++ {
			if i <= j {
				if v.Contains(short(i)) != true {
					t.Errorf("I added a number in vain.")
				}
			} else {
				if v.Contains(short(i)) != false {
					t.Errorf("Ghost content")
					break
				}
			}
		}
	}
}
