package roaring_test

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/RoaringBitmap/roaring"
)

func TestSelectAfterOptimize(t *testing.T) {
	r := roaring.New()

	// load and check test data
	b, err := ioutil.ReadFile(filepath.Join("testdata", "optimizecorrupt.bin"))
	if err != nil {
		t.Fatal(err)
	}
	err = r.UnmarshalBinary(b)
	if err != nil {
		t.Fatal(err)
	}
	if n := r.GetCardinality(); n != 855 {
		t.Fatal("wrong number of entries", n)
	}

	if max := r.Maximum(); max != 920327 {
		t.Fatal("wrong maximum entry", max)
	}

	if sz := len(b); sz != 1734 {
		t.Fatal("wrong size", sz)
	}

	// save original version as array
	origArray := r.ToArray()

	// comment this out to get a passing test
	r.RunOptimize()

	// get a list of values after optimize
	optimized := r.ToArray()

	// this should be fine in both cases
	if diff := len(optimized) - len(origArray); diff != 0 {
		t.Fatal("element count different - diff:", diff)
	}

	// this is also fine
	for i := range optimized {
		if optimized[i] != origArray[i] {
			t.Errorf("array compare %d", i)
		}
	}

	// this produces errors with the optimized version of the bitmap
	n := r.GetCardinality()
	for i := uint64(0); i < n; i++ {

		v, err := r.Select(uint32(i))
		if err != nil {
			t.Fatal(err)
		}

		if diff := origArray[i] - v; diff != 0 {
			t.Errorf("select %03d failed - %d vs %d (diff:%d)", i, origArray[i], v, diff)
		}
	}
}
