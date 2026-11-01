package roaring64

// Serialization edge cases ported from CRoaring (roaring64_serialization.cpp,
// roaring64_unit.cpp) and Java (TestRoaring64Bitmap.java).
// To run just these tests: go test -run TestPortableEdge

import (
	"bytes"
	"encoding/binary"
	"math"
	"math/rand"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// deserializers64 lists every 64-bit entry point that reads the portable format.
var deserializers64 = []struct {
	name string
	read func(rb *Bitmap, data []byte) error
}{
	{"ReadPortableFrom", func(rb *Bitmap, data []byte) error {
		_, err := rb.ReadPortableFrom(bytes.NewReader(data))
		return err
	}},
	{"ReadFrom", func(rb *Bitmap, data []byte) error {
		_, err := rb.ReadFrom(bytes.NewReader(data))
		return err
	}},
	{"UnmarshalBinary", func(rb *Bitmap, data []byte) error {
		return rb.UnmarshalBinary(data)
	}},
	{"FromUnsafeBytes", func(rb *Bitmap, data []byte) error {
		_, err := rb.FromUnsafeBytes(data)
		return err
	}},
}

// requireUsable checks that a bitmap can be queried without panicking and that
// it is internally valid.
func requireUsable(t *testing.T, rb *Bitmap) {
	t.Helper()
	require.NotPanics(t, func() {
		_ = rb.GetCardinality()
		_ = rb.ToArray()
		_ = rb.Contains(42)
	})
	require.NoError(t, rb.Validate())
}

// S1: a bitmap whose buckets hold run, bitset and array containers, with one
// bucket large enough to need the offset header, re-serializes byte for byte.
func TestPortableEdgeMixedContainerRoundTrip(t *testing.T) {
	const K = uint64(1) << 32
	const H = uint64(1) << 63
	rb := NewBitmap()
	rb.Add(5)
	rb.AddRange(K, K+1000)
	rb.Add(K + 65536)
	for i := uint64(0); i < 5000; i++ {
		rb.Add(2*K + 2*i)
	}
	for i := uint64(1); i <= 4; i++ {
		rb.Add(2*K + i*65536)
	}
	rb.Add(H - 1)
	rb.Add(H)
	rb.AddRange(math.MaxUint64-99, math.MaxUint64)
	rb.Add(math.MaxUint64)
	rb.RunOptimize()
	expectedCard := uint64(1 + 1000 + 1 + 5000 + 4 + 2 + 100)
	require.Equal(t, expectedCard, rb.GetCardinality())

	stats := rb.Stats()
	assert.Positive(t, stats.RunContainers)
	assert.Positive(t, stats.BitmapContainers)
	assert.Positive(t, stats.ArrayContainers)

	data, err := rb.ToPortableBytes()
	require.NoError(t, err)
	assert.Equal(t, uint64(len(data)), rb.GetPortableSerializedSizeInBytes())

	back := NewBitmap()
	n, err := back.ReadPortableFrom(bytes.NewReader(data))
	require.NoError(t, err)
	assert.Equal(t, int64(len(data)), n)
	require.NoError(t, back.Validate())
	assert.True(t, back.Equals(rb))
	assert.Equal(t, expectedCard, back.GetCardinality())
	assert.True(t, back.Contains(math.MaxUint64))

	again, err := back.ToPortableBytes()
	require.NoError(t, err)
	assert.Equal(t, data, again)
}

// S2: the malformed fixtures shared with CRoaring must be rejected without a
// panic, and the target must stay usable.
func TestPortableEdgeMalformedFixtures(t *testing.T) {
	for _, name := range []string{
		"64mapemptyinput.bin",
		"64mapsizetoosmall.bin",
		"64mapinvalidsize.bin",
		"64mapkeytoosmall.bin",
	} {
		data, err := os.ReadFile(testdataPath(t, name))
		require.NoError(t, err)
		for _, d := range deserializers64 {
			t.Run(name+"/"+d.name, func(t *testing.T) {
				rb := BitmapOf(42)
				var rerr error
				require.NotPanics(t, func() { rerr = d.read(rb, data) })
				require.Error(t, rerr)
				requireUsable(t, rb)
				if d.name == "ReadPortableFrom" {
					assert.True(t, rb.Equals(BitmapOf(42)))
				}
			})
		}
	}
}

// S3: every strict prefix of a valid stream is rejected; ReadPortableFrom
// leaves the target unchanged and the other readers leave it valid.
func TestPortableEdgeTruncationSweep(t *testing.T) {
	source := BitmapOf(1, 1<<32, math.MaxUint64)
	source.AddRange(1<<40, 1<<40+5000)
	source.RunOptimize()
	data, err := source.ToPortableBytes()
	require.NoError(t, err)

	for _, d := range deserializers64 {
		t.Run(d.name, func(t *testing.T) {
			for k := 0; k < len(data); k++ {
				rb := BitmapOf(42)
				var rerr error
				require.NotPanics(t, func() { rerr = d.read(rb, data[:k]) }, "cut %d", k)
				require.Error(t, rerr, "cut %d", k)
				requireUsable(t, rb)
				if d.name == "ReadPortableFrom" {
					require.True(t, rb.Equals(BitmapOf(42)), "cut %d", k)
				}
			}
			rb := BitmapOf(42)
			require.NoError(t, d.read(rb, data))
			require.True(t, rb.Equals(source))
		})
	}
}

// S4: two containers with the same 16-bit key inside one bucket must be
// rejected by the reader or by Validate.
func TestPortableEdgeDuplicateContainerKeyInBucket(t *testing.T) {
	data := []byte{
		1, 0, 0, 0, 0, 0, 0, 0, // one bucket
		0, 0, 0, 0, // bucket key
		0x3B, 0x30, // run cookie
		1, 0, // container count - 1
		0,    // run flags
		0, 0, // key 0
		0, 0, // cardinality - 1
		0, 0, // key 0 again
		0, 0, // cardinality - 1
		0, 0, // value
		0, 0, // value
	}
	for _, d := range deserializers64 {
		t.Run(d.name, func(t *testing.T) {
			rb := NewBitmap()
			if err := d.read(rb, data); err == nil {
				assert.Error(t, rb.Validate())
			}
		})
	}
}

// S5: a run ending at 65535 is valid; one ending at 65536 is not.
func TestPortableEdgeRunPastContainerEnd(t *testing.T) {
	rb := NewBitmap()
	rb.AddRange(65000, 65536)
	rb.RunOptimize()
	data, err := rb.ToPortableBytes()
	require.NoError(t, err)
	require.Equal(t, uint16(535), binary.LittleEndian.Uint16(data[len(data)-2:]))

	ok := NewBitmap()
	_, err = ok.ReadPortableFrom(bytes.NewReader(data))
	require.NoError(t, err)
	require.NoError(t, ok.Validate())
	assert.Equal(t, uint64(536), ok.GetCardinality())

	bad := bytes.Clone(data)
	binary.LittleEndian.PutUint16(bad[len(bad)-2:], 536)
	for _, d := range deserializers64 {
		t.Run(d.name, func(t *testing.T) {
			b := NewBitmap()
			if err := d.read(b, bad); err == nil {
				assert.Error(t, b.Validate())
			}
		})
	}
}

// S8: random bytes never make a reader panic; when the result validates, it
// behaves like a bitmap.
func TestPortableEdgeRandomBytes(t *testing.T) {
	r := rand.New(rand.NewSource(1234))
	buf := make([]byte, 2000)
	for iter := 0; iter < 2000; iter++ {
		n := r.Intn(len(buf))
		r.Read(buf[:n])
		if n >= 8 && r.Intn(2) == 0 {
			// small bucket count so that we reach the bucket parsing code
			binary.LittleEndian.PutUint64(buf, uint64(r.Intn(3)))
			if n >= 14 {
				binary.LittleEndian.PutUint16(buf[12:], 0x303B-uint16(r.Intn(2)))
			}
		}
		data := buf[:n]
		for _, d := range deserializers64 {
			rb := NewBitmap()
			var err error
			require.NotPanics(t, func() { err = d.read(rb, bytes.Clone(data)) }, "%s input %x", d.name, data)
			if err != nil || rb.Validate() != nil {
				continue
			}
			before := rb.GetCardinality()
			added := uint64(0)
			for v := uint64(100); v < 1000; v++ {
				if rb.CheckedAdd(v) {
					added++
				}
			}
			require.Equal(t, before+added, rb.GetCardinality())
		}
	}
}

// S9: removing a value turns a bitset container into an array container; the
// result still round-trips.
func TestPortableEdgeRoundTripAfterRemoveChangesContainerType(t *testing.T) {
	for _, base := range []uint64{0, 1 << 40, 1 << 63} {
		rb := NewBitmap()
		for i := uint64(0); i < 4097; i++ {
			rb.Add(base + 2*i)
		}
		require.Equal(t, uint64(1), rb.Stats().BitmapContainers)
		rb.Remove(base)
		require.Equal(t, uint64(4096), rb.GetCardinality())
		data, err := rb.ToPortableBytes()
		require.NoError(t, err)
		back := NewBitmap()
		_, err = back.ReadPortableFrom(bytes.NewReader(data))
		require.NoError(t, err)
		assert.Equal(t, rb.ToArray(), back.ToArray())
		assert.False(t, back.Contains(base))
		require.NoError(t, back.Validate())
	}
}
