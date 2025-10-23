//go:build (386 && !appengine) || (amd64 && !appengine) || (arm && !appengine) || (arm64 && !appengine) || (ppc64le && !appengine) || (mipsle && !appengine) || (mips64le && !appengine) || (mips64p32le && !appengine) || (wasm && !appengine)
// +build 386,!appengine amd64,!appengine arm,!appengine arm64,!appengine ppc64le,!appengine mipsle,!appengine mips64le,!appengine mips64p32le,!appengine wasm,!appengine

package roaring

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFrozenFormat(t *testing.T) {
	tests := [...]struct {
		name, frozenPath, portablePath string
	}{
		{
			name:         "bitmaps only",
			frozenPath:   "testfrozendata/bitmaps_only.frozen",
			portablePath: "testfrozendata/bitmaps_only.portable",
		},
		{
			name:         "arrays only",
			frozenPath:   "testfrozendata/arrays_only.frozen",
			portablePath: "testfrozendata/arrays_only.portable",
		},
		{
			name:         "runs only",
			frozenPath:   "testfrozendata/runs_only.frozen",
			portablePath: "testfrozendata/runs_only.portable",
		},
		{
			name:         "mixed",
			frozenPath:   "testfrozendata/mixed.frozen",
			portablePath: "testfrozendata/mixed.portable",
		},
	}

	for _, test := range tests {
		// NOTE: opted for loading files twice rather than optimizing it because:
		// 1. It's still cheap enough, it's small files; and
		// 2. In a buggy scenario one of the tests may write into the buffer and cause
		//    a race condition, making it harder to figure out why the tests fail.
		name, fpath, ppath := test.name, test.frozenPath, test.portablePath
		t.Run("view "+name, func(t *testing.T) {
			t.Parallel()

			frozenBuf, err := ioutil.ReadFile(fpath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "\n\nIMPORTANT: For testing file IO, the roaring library requires disk access.\nWe omit some tests for now.\n\n")
				return
			}
			portableBuf, err := ioutil.ReadFile(ppath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "\n\nIMPORTANT: For testing file IO, the roaring library requires disk access.\nWe omit some tests for now.\n\n")
				return
			}

			frozen, portable := New(), New()
			if err := frozen.FrozenView(frozenBuf); err != nil {
				t.Fatalf("failed to load bitmap from %s: %s", fpath, err)
			}
			if _, err := portable.FromBuffer(portableBuf); err != nil {
				t.Fatalf("failed to load bitmap from %s: %s", ppath, err)
			}

			if !frozen.Equals(portable) {
				t.Fatalf("bitmaps for %s and %s differ", fpath, ppath)
			}
		})
		t.Run("freeze "+name, func(t *testing.T) {
			t.Parallel()

			frozenBuf, err := ioutil.ReadFile(fpath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "\n\nIMPORTANT: For testing file IO, the roaring library requires disk access.\nWe omit some tests for now.\n\n")
				return
			}
			portableBuf, err := ioutil.ReadFile(ppath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "\n\nIMPORTANT: For testing file IO, the roaring library requires disk access.\nWe omit some tests for now.\n\n")
				return
			}

			portable := New()
			if _, err := portable.FromBuffer(portableBuf); err != nil {
				t.Fatalf("failed to load bitmap from %s: %s", ppath, err)
			}

			frozenSize := portable.GetFrozenSizeInBytes()
			if int(frozenSize) != len(frozenBuf) {
				t.Errorf("size for serializing %s differs from %s's size", ppath, fpath)
			}
			frozen, err := portable.Freeze()
			if err != nil {
				t.Fatalf("can't freeze %s: %s", ppath, err)
			}
			if !reflect.DeepEqual(frozen, frozenBuf) {
				t.Fatalf("frozen file for %s and %s differ", fpath, ppath)
			}
		})
		t.Run("freeze with writer"+name, func(t *testing.T) {
			t.Parallel()

			frozenBuf, err := ioutil.ReadFile(fpath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "\n\nIMPORTANT: For testing file IO, the roaring library requires disk access.\nWe omit some tests for now.\n\n")
				return
			}
			portableBuf, err := ioutil.ReadFile(ppath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "\n\nIMPORTANT: For testing file IO, the roaring library requires disk access.\nWe omit some tests for now.\n\n")
				return
			}

			portable := New()
			if _, err := portable.FromBuffer(portableBuf); err != nil {
				t.Fatalf("failed to load bitmap from %s: %s", ppath, err)
			}

			wr := &bytes.Buffer{}
			frozenSize, err := portable.WriteFrozenTo(wr)
			if err != nil {
				t.Fatalf("can't freeze %s: %s", ppath, err)
			}
			if int(frozenSize) != len(frozenBuf) {
				t.Errorf("size for serializing %s differs from %s's size", ppath, fpath)
			}
			if !reflect.DeepEqual(wr.Bytes(), frozenBuf) {
				t.Fatalf("frozen file for %s and %s differ", fpath, ppath)
			}
		})
	}
}

func TestBitMapValidationFromFrozen(t *testing.T) {
	// To understand what is going on here, read https://github.com/RoaringBitmap/RoaringFormatSpec
	// Maintainers: The loader and corruptor are dependent on one another
	// The tests expect a certain size, with values at certain location.
	// The tests are geared toward single byte corruption.

	// There is no way to test Bitmap container corruption once the bitmap is deserialized

	deserializationTests := []struct {
		name      string
		loader    func(bm *Bitmap)
		corruptor func(s []byte)
		err       error
	}{
		{
			name: "Corrupts Run Length vs Num Runs",
			loader: func(bm *Bitmap) {
				bm.AddRange(0, 2)
				bm.AddRange(4, 6)
				bm.AddRange(8, 100)
			},
			corruptor: func(s []byte) {
				// 21 is the length of the run of the last run/range
				// Shortening causes interval sum to be to short
				s[10] = 1
			},
			err: ErrRunIntervalSize,
		},
		{
			name: "Corrupts Run Length",
			loader: func(bm *Bitmap) {
				bm.AddRange(100, 110)
			},
			corruptor: func(s []byte) {
				s[2] = 0
			},
			err: ErrRunIntervalSize,
		},
		{
			name: "Creates Interval Overlap",
			loader: func(bm *Bitmap) {
				bm.AddRange(100, 110)
				bm.AddRange(115, 125)
			},
			corruptor: func(s []byte) {
				// sets the start of the second run
				// Creates overlapping intervals
				s[4] = 108
			},
			err: ErrRunIntervalOverlap,
		},
		{
			name: "Break Array Sort Order",
			loader: func(bm *Bitmap) {
				arrayEntries := make([]uint32, 0, 10)
				for i := 0; i < 10; i++ {
					arrayEntries = append(arrayEntries, uint32(i))
				}
				bm.AddMany(arrayEntries)
			},
			corruptor: func(s []byte) {
				// breaks the sort order
				s[4] = 0
			},
			err: ErrArrayIncorrectSort,
		},
	}

	for _, tt := range deserializationTests {
		t.Run(tt.name, func(t *testing.T) {
			bm := NewBitmap()
			tt.loader(bm)
			assert.NoError(t, bm.Validate())
			serialized, err := bm.Freeze()
			assert.NoError(t, err)
			tt.corruptor(serialized)
			corruptedDeserializedBitMap := NewBitmap()

			// Check that MustFrozenView returns nil if and only if tt.err is nil
			frozenViewErr := corruptedDeserializedBitMap.MustFrozenView(serialized)
			if tt.err == nil {
				assert.NoError(t, frozenViewErr, "expected MustFrozenView to succeed when tt.err is nil")
			} else {
				assert.Error(t, frozenViewErr, "expected MustFrozenView to fail when tt.err is not nil")
			}
		})
	}
}
