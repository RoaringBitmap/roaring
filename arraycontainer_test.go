package roaring

// to run just these tests: go test -run TestArrayContainer*

import (
	"testing"
)

func TestArrayContainerSetAndGet(t *testing.T) {
	v := container(newArrayContainer())
	v = v.add(100)
	if v.getCardinality() != 1 {
		t.Errorf("Bogus cardinality.")
	}
	for i := 0; i <= array_default_max_size; i++ {
		if i == 100 {
			if v.contains(uint16(i)) != true {
				t.Errorf("I added a number in vain.")
			}
		} else {
			if v.contains(uint16(i)) != false {
				t.Errorf("Ghost content")
				break
			}
		}
	}
}

func TestArrayContainerMassiveSetAndGet(t *testing.T) {
	v := container(newArrayContainer())
	for j := 0; j <= array_default_max_size; j++ {

		v = v.add(uint16(j))
		if v.getCardinality() != 1+j {
			t.Errorf("Bogus cardinality. ", v.getCardinality(), j)
		}
		for i := 0; i <= array_default_max_size; i++ {
			if i <= j {
				if v.contains(uint16(i)) != true {
					t.Errorf("I added a number in vain.")
				}
			} else {
				if v.contains(uint16(i)) != false {
					t.Errorf("Ghost content")
					break
				}
			}
		}
	}
}
