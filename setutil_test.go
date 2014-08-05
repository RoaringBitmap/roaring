package roaring

// to run just these tests: go test -run TestSetUtil*

import (
	"testing"
)

func TestSetUtilDifference(t *testing.T) {
	data1 := []short{0, 1, 2, 3, 4, 9}
	data2 := []short{2, 3, 4, 5, 8, 9, 11}
	result := make([]short, 0, len(data1)+len(data2))
	expectedresult := []short{0, 1}
	nl := Difference(data1, data2, result)
	result = result[:nl]
	if !Equal(result, expectedresult) {
		t.Errorf("Difference is broken")
	}
	expectedresult = []short{5, 8, 11}
	nl = Difference(data2, data1, result)
	result = result[:nl]
	if !Equal(result, expectedresult) {
		t.Errorf("Difference is broken")
	}
}

func TestSetUtilUnion(t *testing.T) {
	data1 := []short{0, 1, 2, 3, 4, 9}
	data2 := []short{2, 3, 4, 5, 8, 9, 11}
	result := make([]short, 0, len(data1)+len(data2))
	expectedresult := []short{0, 1, 2, 3, 4, 5, 8, 9, 11}
	nl := Union2by2(data1, data2, result)
	result = result[:nl]
	if !Equal(result, expectedresult) {
		t.Errorf("Union is broken")
	}
	nl = Union2by2(data2, data1, result)
	result = result[:nl]
	if !Equal(result, expectedresult) {
		t.Errorf("Union is broken")
	}
}

func TestSetUtilExclusiveUnion(t *testing.T) {
	data1 := []short{0, 1, 2, 3, 4, 9}
	data2 := []short{2, 3, 4, 5, 8, 9, 11}
	result := make([]short, 0, len(data1)+len(data2))
	expectedresult := []short{0, 1, 5, 8, 11}
	nl := ExclusiveUnion2by2(data1, data2, result)
	result = result[:nl]
	if !Equal(result, expectedresult) {
		t.Errorf("Exclusive Union is broken")
	}
	nl = ExclusiveUnion2by2(data2, data1, result)
	result = result[:nl]
	if !Equal(result, expectedresult) {
		t.Errorf("Exclusive Union is broken")
	}
}

func TestSetUtilIntersection(t *testing.T) {
	data1 := []short{0, 1, 2, 3, 4, 9}
	data2 := []short{2, 3, 4, 5, 8, 9, 11}
	result := make([]short, 0, len(data1)+len(data2))
	expectedresult := []short{2, 3, 4, 9}
	nl := Intersection2by2(data1, data2, result)
	result = result[:nl]
	result = result[:len(expectedresult)]
	if !Equal(result, expectedresult) {
		t.Errorf("Intersection is broken")
	}
	nl = Intersection2by2(data2, data1, result)
	result = result[:nl]
	if !Equal(result, expectedresult) {
		t.Errorf("Intersection is broken")
	}
	data1 = []short{4}

	data2 = make([]short, 10000)
	for i := range data2 {
		data2[i] = short(i)
	}
	result = make([]short, 0, len(data1)+len(data2))
	expectedresult = data1
	nl = Intersection2by2(data1, data2, result)
	result = result[:nl]
	result = result[:len(expectedresult)]

	if !Equal(result, expectedresult) {
		t.Errorf("Long intersection is broken")
	}
	nl = Intersection2by2(data2, data1, result)
	result = result[:nl]
	if !Equal(result, expectedresult) {
		t.Errorf("Long intersection is broken")
	}

}

func TestSetUtilBinarySearch(t *testing.T) {
	data := make([]short, 256)
	for i := range data {
		data[i] = short(2 * i)
	}
	for i := 0; i < 2*len(data); i += 1 {
		key := short(i)
		loc := binarySearch(data, key)
		if (key & 1) == 0 {
			if loc != int(key)/2 {
				t.Errorf("binary search is broken")
			}
		} else {
			if loc != -int(key)/2-2 {
				t.Errorf("neg binary search is broken")
			}
		}
	}
}
