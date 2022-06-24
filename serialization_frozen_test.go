//go:build (386 && !appengine) || (amd64 && !appengine) || (arm && !appengine) || (arm64 && !appengine) || (ppc64le && !appengine) || (mipsle && !appengine) || (mips64le && !appengine) || (mips64p32le && !appengine) || (wasm && !appengine)
// +build 386,!appengine amd64,!appengine arm,!appengine arm64,!appengine ppc64le,!appengine mipsle,!appengine mips64le,!appengine mips64p32le,!appengine wasm,!appengine

package roaring

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"
	"io/ioutil"
	"os"
	"reflect"
	"runtime/debug"
	"testing"
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
				t.Fatalf("failed to open %s: %s", fpath, err)
			}
			portableBuf, err := ioutil.ReadFile(ppath)
			if err != nil {
				t.Fatalf("failed to open %s: %s", ppath, err)
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
				t.Fatalf("failed to open %s: %s", fpath, err)
			}
			portableBuf, err := ioutil.ReadFile(ppath)
			if err != nil {
				t.Fatalf("failed to open %s: %s", ppath, err)
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
				t.Fatalf("failed to open %s: %s", fpath, err)
			}
			portableBuf, err := ioutil.ReadFile(ppath)
			if err != nil {
				t.Fatalf("failed to open %s: %s", ppath, err)
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

type frozenTestCase struct {
	name        string
	useMMapped  bool
	useFrozen   bool
	shouldPanic bool
}

func TestFrozenCases(t *testing.T) {
	cases := []frozenTestCase{
		{name: "in-memory-frozen", useMMapped: false,
			useFrozen:   true,
			shouldPanic: false},
		{name: "mmap-frozen", useMMapped: true,
			useFrozen: true,
			// THIS SHOULD NOT BE PANIC/FAULTING
			shouldPanic: true},
		{name: "in-memory", useMMapped: false,
			useFrozen:   false,
			shouldPanic: false},
		{name: "mmap", useMMapped: true,
			useFrozen:   false,
			shouldPanic: false},
	}
	for _, testCase := range cases {
		testFrozenCase(t, testCase)
	}
}
func testFrozenCase(t *testing.T, testCase frozenTestCase) {

	startingBitmap := NewBitmap()
	startingBitmap.highlowcontainer.appendContainer(0, &runContainer16{iv: []interval16{{0, 0},
		{2, 5}}}, false)

	primary := getBitmap(t, testCase, startingBitmap)
	assert.True(t, startingBitmap.Equals(primary))
	clone := primary.Clone()
	primary.Xor(clone)
	res := primary.GetCardinality()
	assert.Equal(t, res, uint64(0))
	if testCase.shouldPanic {
		assert.Panics(t, func() {
			debug.SetPanicOnFault(true)
			primary.Xor(clone)
		})
	} else {
		primary.Xor(clone)
	}
}

func getBitmap(t *testing.T, testCase frozenTestCase, startingBitmap *Bitmap) *Bitmap {
	receiver := NewBitmap()
	if testCase.useFrozen {
		frozenBytes, err := startingBitmap.Freeze()
		require.NoError(t, err)
		if testCase.useMMapped {
			require.NoError(t, receiver.FrozenView(getBytesAsMMap(t, frozenBytes)))
			return receiver
		}
		require.NoError(t, receiver.FrozenView(frozenBytes))
		return receiver
	}
	nonFrozenBytes, err := startingBitmap.ToBytes()
	require.NoError(t, err)

	if testCase.useMMapped {
		_, err = receiver.FromBuffer(getBytesAsMMap(t, nonFrozenBytes))
		require.NoError(t, err)
		return receiver
	}
	_, err = receiver.FromBuffer(nonFrozenBytes)
	require.NoError(t, err)
	return receiver
}

func getBytesAsMMap(t *testing.T, startingBytes []byte) []byte {
	os.WriteFile("data", startingBytes, 0777)
	file, err := os.OpenFile("data", os.O_RDONLY, 0)
	require.NoError(t, err)
	fileinfo, err := file.Stat()
	require.NoError(t, err)

	bytes, err := unix.Mmap(int(file.Fd()), 0, int(fileinfo.Size()),
		unix.PROT_READ, unix.MAP_PRIVATE)
	require.NoError(t, err)
	return bytes
}
