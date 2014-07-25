package goroaring

import (
	"testing"
)

func testEq(a, b []short) bool {
    if len(a) != len(b) {
        return false
    }

    for i := range a {
        if a[i] != b[i] {
            return false
        }
    }

    return true
}

func Testdifference(t *testing.T) {
    data1 := []short{0, 1, 2, 3, 4, 9}
    data2 := []short{2, 3, 4, 5, 8, 9, 11}
    result :=  make([]short, 0, len(data1) + len(data2))
    expectedresult := []short{0, 1, 5, 8, 11}  
    Unsigned_difference(data1,data2,result)
    if ! testEq( result, expectedresult) {
        t.Errorf("symmetric difference is broken")
    }
    Unsigned_difference(data2,data1,result)
    if ! testEq( result, expectedresult) {
        t.Errorf("symmetric difference is broken")
    }
}
