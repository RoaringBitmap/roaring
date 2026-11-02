package roaring64

// to run just these tests: go test -run TestPortable

import (
	"bytes"
	"encoding/binary"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/RoaringBitmap/roaring/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// portableFixtures are reference streams shared with CRoaring and the Java
// implementation. All but 64maprunbitsetvals.bin were produced by CRoaring;
// that one was written by the Java library.
var portableFixtures = []string{
	"64mapempty.bin",
	"64map32bitvals.bin",
	"64mapspreadvals.bin",
	"64maphighvals.bin",
	"64maprunbitsetvals.bin",
}

// testdataPath resolves a fixture relative to this source file: some CI jobs run
// the compiled test binary from the repository root rather than from this
// package's directory.
func testdataPath(t *testing.T, name string) string {
	t.Helper()
	_, self, _, ok := runtime.Caller(0)
	require.True(t, ok, "could not locate the test source file")
	return filepath.Join(filepath.Dir(self), "testdata", name)
}

func TestPortableSerializationMatchesCRoaring(t *testing.T) {
	for _, name := range portableFixtures {
		t.Run(name, func(t *testing.T) {
			reference, err := os.ReadFile(testdataPath(t, name))
			require.NoError(t, err)

			bm := NewBitmap()
			n, err := bm.ReadPortableFrom(bytes.NewReader(reference))
			require.NoError(t, err)
			assert.Equal(t, int64(len(reference)), n)
			require.NoError(t, bm.Validate())

			assert.Equal(t, uint64(len(reference)), bm.GetPortableSerializedSizeInBytes())
			written, err := bm.ToPortableBytes()
			require.NoError(t, err)
			assert.Equal(t, reference, written)

			// The portable format is what ReadFrom/WriteTo already speak.
			legacy := NewBitmap()
			_, err = legacy.ReadFrom(bytes.NewReader(reference))
			require.NoError(t, err)
			assert.True(t, legacy.Equals(bm))
		})
	}
}

// 64maprunbitsetvals.bin holds run, bitset and array containers, and a bucket
// with more than four containers, so its 32-bit bitmap carries the offset
// header.
func TestPortableFixtureWithRunAndBitsetContainers(t *testing.T) {
	reference, err := os.ReadFile(testdataPath(t, "64maprunbitsetvals.bin"))
	require.NoError(t, err)
	bm := NewBitmap()
	_, err = bm.ReadPortableFrom(bytes.NewReader(reference))
	require.NoError(t, err)

	const k = uint64(1) << 32
	const h = uint64(1) << 63
	const m = uint64(math.MaxUint64)
	expected := NewBitmap()
	expected.Add(5)
	expected.AddRange(k, k+1000) // a run
	expected.Add(k + 65536)
	for i := uint64(0); i < 5000; i++ {
		expected.Add(2*k + 2*i) // a bitset
	}
	for i := uint64(1); i <= 4; i++ {
		expected.Add(2*k + i*65536)
	}
	expected.Add(h - 1)
	expected.Add(h)
	expected.AddRange(m-99, m)
	expected.Add(m)

	assert.Equal(t, uint64(6108), bm.GetCardinality())
	assert.True(t, bm.Contains(m))
	assert.True(t, expected.Equals(bm))

	// Run-optimized, the expected bitmap serializes to the same bytes.
	expected.RunOptimize()
	written, err := expected.ToPortableBytes()
	require.NoError(t, err)
	assert.Equal(t, reference, written)
}

func TestPortableDeserializationRejectsInvalidBucketCount(t *testing.T) {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, maxPortableBucketCount+1)

	_, err := NewBitmap().ReadPortableFrom(bytes.NewReader(buf))
	require.Error(t, err)
}

// portableStream assembles a stream from raw (key, 32-bit bitmap) pairs,
// bypassing WritePortableTo so that invalid streams can be built.
func portableStream(t *testing.T, count uint64, keys []uint32, buckets []*roaring.Bitmap) []byte {
	t.Helper()
	var out bytes.Buffer
	header := make([]byte, 8)
	binary.LittleEndian.PutUint64(header, count)
	out.Write(header)
	for i, key := range keys {
		keyBuf := make([]byte, 4)
		binary.LittleEndian.PutUint32(keyBuf, key)
		out.Write(keyBuf)
		_, err := buckets[i].WriteTo(&out)
		require.NoError(t, err)
	}
	return out.Bytes()
}

func TestPortableDeserializationRejectsUnsortedBucketKeys(t *testing.T) {
	for _, tc := range []struct {
		name string
		keys []uint32
	}{
		{"duplicate", []uint32{1, 1}},
		{"decreasing", []uint32{2, 1}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			buckets := []*roaring.Bitmap{roaring.BitmapOf(3), roaring.BitmapOf(4)}
			data := portableStream(t, 2, tc.keys, buckets)

			_, err := NewBitmap().ReadPortableFrom(bytes.NewReader(data))
			require.Error(t, err)
		})
	}
}

func TestPortableDeserializationDropsEmptyBucket(t *testing.T) {
	buckets := []*roaring.Bitmap{roaring.NewBitmap(), roaring.BitmapOf(7)}
	data := portableStream(t, 2, []uint32{0, 1}, buckets)

	bm := NewBitmap()
	n, err := bm.ReadPortableFrom(bytes.NewReader(data))
	require.NoError(t, err)
	assert.Equal(t, int64(len(data)), n)
	require.NoError(t, bm.Validate())
	assert.Equal(t, []uint64{uint64(1)<<32 | 7}, bm.ToArray())
}

func TestPortableDeserializationIsAtomicWhenTruncated(t *testing.T) {
	source := BitmapOf(1, 1<<32, math.MaxUint64)
	complete, err := source.ToPortableBytes()
	require.NoError(t, err)

	for _, cut := range []int{1, 5, 20} {
		bm := BitmapOf(42)
		_, err := bm.ReadPortableFrom(bytes.NewReader(complete[:len(complete)-cut]))
		require.Error(t, err)
		assert.True(t, bm.Equals(BitmapOf(42)), "bitmap was modified by a failed read")
	}
}

func TestPortableDeserializationKeepsCopyOnWrite(t *testing.T) {
	data, err := BitmapOf(1, 1<<32).ToPortableBytes()
	require.NoError(t, err)

	for _, copyOnWrite := range []bool{false, true} {
		bm := NewBitmap()
		bm.SetCopyOnWrite(copyOnWrite)
		_, err := bm.ReadPortableFrom(bytes.NewReader(data))
		require.NoError(t, err)
		assert.Equal(t, copyOnWrite, bm.GetCopyOnWrite())
	}
}

func TestPortableRoundTrip(t *testing.T) {
	source := BitmapOf(1, 2, 3, 1<<16, 1<<32, 1<<48, math.MaxUint64)
	source.AddRange(1<<40, 1<<40+5000)
	source.RunOptimize()

	data, err := source.ToPortableBytes()
	require.NoError(t, err)
	assert.Equal(t, uint64(len(data)), source.GetPortableSerializedSizeInBytes())

	bm := NewBitmap()
	n, err := bm.ReadPortableFrom(bytes.NewReader(data))
	require.NoError(t, err)
	assert.Equal(t, int64(len(data)), n)
	assert.True(t, bm.Equals(source))
}
