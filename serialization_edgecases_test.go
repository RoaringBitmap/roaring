package roaring

// Edge cases of the portable serialization format, ported from the Java
// (RoaringBitmap) and C (CRoaring) test suites so that the three libraries
// test the same scenarios.

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mustDecodeHex decodes a hex string, ignoring spaces.
func mustDecodeHex(t testing.TB, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(strings.ReplaceAll(s, " ", ""))
	require.NoError(t, err)
	return b
}

// deserializers lists every way of reading a portable bitmap into an existing receiver.
var deserializers = []struct {
	name string
	read func(rb *Bitmap, data []byte) error
}{
	{"ReadFrom", func(rb *Bitmap, data []byte) error {
		_, err := rb.ReadFrom(bytes.NewReader(data))
		return err
	}},
	{"ReadFromPlainReader", func(rb *Bitmap, data []byte) error {
		// a reader that is not an internal.ByteInput goes through the adapter
		_, err := rb.ReadFrom(io.MultiReader(bytes.NewReader(data)))
		return err
	}},
	{"UnmarshalBinary", func(rb *Bitmap, data []byte) error {
		return rb.UnmarshalBinary(data)
	}},
	{"FromBuffer", func(rb *Bitmap, data []byte) error {
		_, err := rb.FromBuffer(data)
		return err
	}},
	{"FromUnsafeBytes", func(rb *Bitmap, data []byte) error {
		_, err := rb.FromUnsafeBytes(data)
		return err
	}},
}

// assertUsableAfterFailure checks that a bitmap whose deserialization failed
// is still usable: it is valid and either empty or unchanged.
func assertUsableAfterFailure(t *testing.T, rb *Bitmap, before []uint32) {
	t.Helper()
	require.NotPanics(t, func() {
		require.NoError(t, rb.Validate())
		card := rb.GetCardinality()
		if card != 0 {
			assert.Equal(t, before, rb.ToArray())
		}
		rb.Contains(7)
		rb.Add(12345)
		assert.True(t, rb.Contains(12345))
	})
}

// MustReadFrom must surface a read error. A truncated body is left empty, and
// an empty bitmap validates, so dropping the error looks like a successful load.
func TestMustReadFromReportsReadError(t *testing.T) {
	src, err := BitmapOf(7, 70000, 140000).ToBytes()
	require.NoError(t, err)

	rb := BitmapOf(1, 2, 3)
	_, err = rb.MustReadFrom(bytes.NewReader(src[:len(src)-1]))
	assert.Error(t, err)
	assert.EqualValues(t, 0, rb.GetCardinality())
	assert.NoError(t, rb.Validate())

	n, err := rb.MustReadFrom(bytes.NewReader(src))
	require.NoError(t, err)
	assert.EqualValues(t, len(src), n)
	assert.Equal(t, []uint32{7, 70000, 140000}, rb.ToArray())

	// A bad cookie fails before the receiver is modified.
	rb = BitmapOf(1, 2, 3)
	_, err = rb.MustReadFrom(bytes.NewReader([]byte{0, 0, 0, 0}))
	assert.Error(t, err)
	assert.Equal(t, []uint32{1, 2, 3}, rb.ToArray())
}

// Gap 1 (serialization): a failed deserialization must not leave the receiver
// with nil containers. Gap 10: every strict prefix of a valid buffer is rejected.
func TestDeserializationFailureLeavesUsableBitmap(t *testing.T) {
	src, err := BitmapOf(7, 70000, 140000).ToBytes()
	require.NoError(t, err)
	require.Equal(t, mustDecodeHex(t, "3a300000 03000000 00000000 01000000 02000000 20000000 22000000 24000000 0700 7011 e022"), src)

	for _, d := range deserializers {
		t.Run(d.name, func(t *testing.T) {
			for k := 0; k < len(src); k++ {
				rb := BitmapOf(1, 2, 3)
				require.Errorf(t, d.read(rb, src[:k]), "prefix of length %d accepted", k)
				assertUsableAfterFailure(t, rb, []uint32{1, 2, 3})
			}
			rb := BitmapOf(1, 2, 3)
			require.NoError(t, d.read(rb, src))
			assert.Equal(t, []uint32{7, 70000, 140000}, rb.ToArray())
		})
	}
}

// Gap 10: truncation sweep on the shared files.
func TestDeserializationTruncationSweep(t *testing.T) {
	for _, fn := range []string{
		"testfrozendata/runs_only.portable",
		"testfrozendata/arrays_only.portable",
		"testfrozendata/mixed.portable",
		"testdata/bitmapwithruns.bin",
		"testdata/bitmapwithoutruns.bin",
	} {
		data, err := os.ReadFile(fn)
		require.NoError(t, err)
		t.Run(fn, func(t *testing.T) {
			want := NewBitmap()
			_, err := want.FromBuffer(data)
			require.NoError(t, err)
			require.NoError(t, want.Validate())
			// FromBuffer does not copy, so every prefix can be tried.
			rb := NewBitmap()
			for k := 0; k < len(data); k++ {
				_, err := rb.FromBuffer(data[:k])
				require.Errorf(t, err, "prefix of length %d accepted", k)
				require.NoError(t, rb.Validate())
			}
			// The copying readers are sampled: every byte of the header, then a stride.
			stride := 1 + len(data)/200
			for k := 0; k < len(data); k++ {
				if k > 64 && k%stride != 0 && k != len(data)-1 {
					continue
				}
				for _, d := range deserializers[:3] {
					rb := BitmapOf(1, 2, 3)
					require.Errorf(t, d.read(rb, data[:k]), "%s: prefix of length %d accepted", d.name, k)
					assertUsableAfterFailure(t, rb, []uint32{1, 2, 3})
				}
			}
			for _, d := range deserializers {
				rb := NewBitmap()
				require.NoError(t, d.read(rb, data))
				assert.True(t, rb.Equals(want), d.name)
			}
		})
	}
}

// Gap 2: negative, huge and maximal container counts.
func TestDeserializationContainerCount(t *testing.T) {
	for _, vec := range []string{
		"3A300000 00000080", // negative as a signed 32-bit integer
		"3A300000 01000100", // 65537 containers
		"3A300000 FFFFFFFF",
	} {
		data := mustDecodeHex(t, vec)
		for _, d := range deserializers {
			rb := BitmapOf(1)
			assert.Errorf(t, d.read(rb, data), "%s %s", d.name, vec)
			assertUsableAfterFailure(t, rb, []uint32{1})
		}
	}

	// 65536 containers is the maximum. All keys are 0 here, so the result is
	// accepted but invalid. One more container is rejected.
	build := func(n int) []byte {
		buf := make([]byte, 8+n*(4+4+2))
		binary.LittleEndian.PutUint32(buf, serialCookieNoRunContainer)
		binary.LittleEndian.PutUint32(buf[4:], uint32(n))
		return buf
	}
	data := build(1 << 16)
	rb := NewBitmap()
	n, err := rb.FromBuffer(data)
	require.NoError(t, err)
	assert.EqualValues(t, len(data), n)
	assert.Equal(t, 1<<16, len(rb.highlowcontainer.keys))
	assert.ErrorIs(t, rb.Validate(), ErrKeySortOrder)

	data = build(1<<16 + 1)
	rb = NewBitmap()
	_, err = rb.FromBuffer(data)
	assert.Error(t, err)
	assert.NoError(t, rb.Validate())
}

// readPortable deserializes data with every API and returns the FromBuffer result,
// checking that all APIs agree and consume the same number of bytes.
func readPortable(t *testing.T, data []byte) (*Bitmap, int64) {
	t.Helper()
	rb := NewBitmap()
	n, err := rb.FromBuffer(data)
	require.NoError(t, err)
	other := NewBitmap()
	m, err := other.ReadFrom(bytes.NewReader(data))
	require.NoError(t, err)
	assert.Equal(t, n, m)
	assert.Equal(t, rb.ToArray(), other.ToArray())
	other = NewBitmap()
	m, err = other.FromUnsafeBytes(data)
	require.NoError(t, err)
	assert.Equal(t, n, m)
	assert.Equal(t, rb.ToArray(), other.ToArray())
	return rb, n
}

// Gaps 3, 5, 6, 7, 25, 34 and bad-bitmap.bin (gap 14): byte vectors that are
// accepted by the deserializer and then either pass or fail Validate.
// Accepting malformed input and reporting it in Validate is the documented contract.
func TestDeserializationMalformedThenValidate(t *testing.T) {
	ones := strings.Repeat("FF", 8192)
	tests := []struct {
		name     string
		hex      string
		card     uint64
		consumed int64
		err      error // nil means valid; errAny means any error
	}{
		// gap 3
		{"duplicate keys", "3B30 0100 00 0000 0000 0000 0000 0000 0000", 2, 17, ErrKeySortOrder},
		{"unsorted keys", "3B30 0100 00 0100 0000 0000 0000 0000 0000", 2, 17, ErrKeySortOrder},
		// gap 4 (also covered by mutation in TestBitMapValidationFromDeserialization)
		{"descending array", "3B30 0000 00 0000 0100 0100 0000", 2, 13, ErrArrayIncorrectSort},
		{"duplicate array", "3B30 0000 00 0000 0100 0000 0000", 2, 13, ErrArrayIncorrectSort},
		// gap 5
		{"full bitset", "3B30 0000 00 0000 FFFF" + ones, 65536, 8201, nil},
		{"bitset with wrong cardinality", "3B30 0000 00 0000 FFFF" + ones[:len(ones)-2] + "FE", 65536, 8201, errAny},
		// gap 6
		{"zero runs", "3B30 0000 01 0000 0000 0000", 0, 11, ErrRunIntervalsEmpty},
		{"contiguous runs", "3B30 0000 01 0000 0100 0200 0000 0000 0100 0000", 2, 19, ErrRunIntervalOverlap},
		{"overlapping runs", "3B30 0000 01 0000 0400 0200 0000 0400 0100 0000", 6, 19, ErrRunIntervalOverlap},
		{"run overflow", "3B30 0000 01 0000 0400 0100 FEFF 0400", 5, 15, ErrRunIntervalOutOfRange},
		// gap 7
		{"run ending at 65535", "3B30 0000 01 0000 1702 0100 E8FD 1702", 536, 15, nil},
		{"run ending at 65536", "3B30 0000 01 0000 1702 0100 E8FD 1802", 537, 15, ErrRunIntervalOutOfRange},
		// gap 8: the cardinality in the header of a run container is ignored
		{"run header cardinality ignored", "3B30 0000 01 0000 0000 0100 0000 0800", 9, 15, nil},
		// gap 25: valid but not the most compact encoding
		{"non-minimal run", "3B30 0000 01 0000 0100 0100 0000 0100", 2, 15, nil},
		// gap 34: CRoaring robust_deserialization test1 (trailing byte not consumed)
		{"run crossing the container end", "3b300000 01 0000 fa2e 0100 0002 ffff 00", 0x10000, 15, ErrRunIntervalOutOfRange},
		// gap 33: PyRoaringBitMap issue 124 vector
		{"pyroaring 124", "3a300000 01000000 0000 0900 10000000 3200 3300 3400 3500 3600 3700 3800 3a00 3b00 3c00", 10, 36, nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data := mustDecodeHex(t, tc.hex)
			rb, n := readPortable(t, data)
			assert.Equal(t, tc.consumed, n)
			assert.Equal(t, tc.card, rb.GetCardinality())
			err := rb.Validate()
			switch {
			case tc.err == nil:
				assert.NoError(t, err)
			case errors.Is(tc.err, errAny):
				assert.Error(t, err)
			default:
				assert.ErrorIs(t, err, tc.err)
			}
			// a truncated copy is always rejected
			_, err = NewBitmap().FromBuffer(data[:n-1])
			assert.Error(t, err)
		})
	}

	t.Run("3000 runs of length 1", func(t *testing.T) {
		data := []byte{0x3B, 0x30, 0, 0, 1, 0, 0}
		data = binary.LittleEndian.AppendUint16(data, 2999)
		data = binary.LittleEndian.AppendUint16(data, 3000)
		for i := 0; i < 3000; i++ {
			data = binary.LittleEndian.AppendUint16(data, uint16(4*i))
			data = binary.LittleEndian.AppendUint16(data, 0)
		}
		require.Equal(t, 12011, len(data))
		rb, n := readPortable(t, data)
		assert.EqualValues(t, len(data), n)
		assert.EqualValues(t, 3000, rb.GetCardinality())
		assert.NoError(t, rb.Validate())
		assert.True(t, rb.Contains(4*2999))
	})

	t.Run("bad-bitmap.bin", func(t *testing.T) {
		data, err := os.ReadFile("testdata/bad-bitmap.bin")
		require.NoError(t, err)
		rb, n := readPortable(t, data)
		assert.EqualValues(t, len(data), n)
		assert.EqualValues(t, 1518, rb.GetCardinality())
		assert.ErrorIs(t, rb.Validate(), ErrArrayIncorrectSort)
		out, err := rb.ToBytes()
		require.NoError(t, err)
		assert.Equal(t, data, out)
	})
}

var errAny = errors.New("any error")

// Gap 9: the offset header is ignored by the Go reader: containers are read sequentially.
func TestDeserializationIgnoresOffsets(t *testing.T) {
	// two containers, the second offset is wrong (24 instead of 26)
	rb, n := readPortable(t, mustDecodeHex(t, "3A300000 02000000 0000 0000 0100 0000 18000000 18000000 0700 0900"))
	assert.EqualValues(t, 28, n)
	assert.Equal(t, []uint32{7, 65545}, rb.ToArray())
	assert.NoError(t, rb.Validate())

	// one container, the offset points past the end
	rb, n = readPortable(t, mustDecodeHex(t, "3A300000 01000000 0500 0000 FF000000 0700"))
	assert.EqualValues(t, 18, n)
	assert.Equal(t, []uint32{327687}, rb.ToArray())
	assert.NoError(t, rb.Validate())
}

// Gap 21: trailing bytes after a valid bitmap are not consumed.
func TestDeserializationTrailingBytes(t *testing.T) {
	rb, n := readPortable(t, mustDecodeHex(t, "3A300000 01000000 0500 0000 10000000 0700 AABB"))
	assert.EqualValues(t, 18, n)
	assert.Equal(t, []uint32{327687}, rb.ToArray())
}

// Gap 11: deserializing random bytes returns an error or a bitmap; it never panics.
// When the result is valid, it must behave like a bitmap.
func TestDeserializationRandomInput(t *testing.T) {
	r := rand.New(rand.NewSource(1234))
	valid := 0
	for i := 0; i < 3000; i++ {
		data := make([]byte, r.Intn(1024))
		r.Read(data)
		if len(data) >= 4 && i%2 == 0 {
			// start with a valid cookie so that we get past the header
			if i%4 == 0 {
				binary.LittleEndian.PutUint32(data, serialCookieNoRunContainer)
				if len(data) >= 8 {
					binary.LittleEndian.PutUint32(data[4:], uint32(r.Intn(8)))
				}
			} else {
				binary.LittleEndian.PutUint16(data, serialCookie)
				binary.LittleEndian.PutUint16(data[2:], uint16(r.Intn(8)))
			}
		}
		checkRandomInput(t, data)
		if rb := NewBitmap(); rb.UnmarshalBinary(data) == nil && rb.Validate() == nil {
			valid++
		}
	}
	assert.Greater(t, valid, 0)
}

func checkRandomInput(t testing.TB, data []byte) {
	orig := bytes.Clone(data)
	for _, d := range deserializers {
		rb := NewBitmap()
		err := d.read(rb, data)
		if verr := rb.Validate(); err != nil {
			if verr != nil {
				t.Fatalf("%s: invalid bitmap left after error %v: %v", d.name, err, verr)
			}
			continue
		} else if verr != nil {
			continue
		}
		card := rb.GetCardinality()
		added := uint64(0)
		for v := uint32(100); v < 1000; v++ {
			if rb.CheckedAdd(v) {
				added++
			}
		}
		if rb.GetCardinality() != card+added {
			t.Fatalf("%s: cardinality %d, expected %d", d.name, rb.GetCardinality(), card+added)
		}
		if err := rb.Validate(); err != nil {
			t.Fatalf("%s: %v", d.name, err)
		}
		if !bytes.Equal(data, orig) {
			t.Fatalf("%s: the input buffer was modified", d.name)
		}
	}
}

// FuzzDeserializePortable is the native-fuzzing version of TestDeserializationRandomInput.
func FuzzDeserializePortable(f *testing.F) {
	for _, pattern := range []string{"testdata/crash*", "testdata/bad-bitmap.bin", "testfrozendata/*.portable"} {
		files, _ := filepath.Glob(pattern)
		for _, fn := range files {
			if data, err := os.ReadFile(fn); err == nil {
				f.Add(data)
			}
		}
	}
	f.Add(mustDecodeHex(f, "3B30 0000 01 0000 1702 0100 E8FD 1802"))
	f.Fuzz(func(t *testing.T, data []byte) {
		checkRandomInput(t, data)
	})
}

// Gap 15: golden bytes around the no-offset threshold (4 containers) with run containers.
func TestSerializationGoldenRunContainers(t *testing.T) {
	golden := map[int]string{
		1: "3b3000000100000900010000000900",
		2: "3b300100030000090001000900010000000900010000000900",
		3: "3b30020007000009000100090002000900010000000900010000000900010000000900",
		4: "3b3003000f00000900010009000200090003000900250000002b0000003100000037000000010000000900010000000900010000000900010000000900",
	}
	sizes := map[int]int{1: 15, 2: 25, 3: 35, 4: 61, 5: 75}
	for k := 1; k <= 5; k++ {
		rb := NewBitmap()
		plain := NewBitmap()
		for j := 0; j < k; j++ {
			rb.AddRange(uint64(j)<<16, uint64(j)<<16+10)
			plain.Add(uint32(j)<<16 + 7)
		}
		rb.RunOptimize()
		data, err := rb.ToBytes()
		require.NoError(t, err)
		assert.Equal(t, sizes[k], len(data), k)
		assert.EqualValues(t, len(data), rb.GetSerializedSizeInBytes())
		if g, ok := golden[k]; ok {
			assert.Equal(t, g, hex.EncodeToString(data), k)
		}
		back, n := readPortable(t, data)
		assert.EqualValues(t, len(data), n)
		assert.True(t, back.Equals(rb))
		assert.NoError(t, back.Validate())

		data, err = plain.ToBytes()
		require.NoError(t, err)
		assert.Equal(t, 8+10*k, len(data), k)
		back, _ = readPortable(t, data)
		assert.True(t, back.Equals(plain))
	}
}

// Gap 17: the empty bitmap.
func TestSerializationEmptyGolden(t *testing.T) {
	data, err := NewBitmap().ToBytes()
	require.NoError(t, err)
	assert.Equal(t, mustDecodeHex(t, "3a30000000000000"), data)
	back, n := readPortable(t, data)
	assert.EqualValues(t, 8, n)
	assert.True(t, back.IsEmpty())
	assert.NoError(t, back.Validate())

	for _, bad := range [][]byte{{}, data[:4], data[:7]} {
		for _, d := range deserializers {
			assert.Error(t, d.read(NewBitmap(), bad), "%s %x", d.name, bad)
		}
	}
}

// Gap 18: an array with 4096 values and a bitset with 4097 values.
func TestSerializationArrayBitsetBoundary(t *testing.T) {
	for _, tc := range []struct {
		n      int
		header string
	}{
		{4096, "3A300000 01000000 0000 FF0F 10000000"},
		{4097, "3A300000 01000000 0000 0010 10000000"},
	} {
		rb := NewBitmap()
		for i := 0; i < tc.n; i++ {
			rb.Add(uint32(2 * i))
		}
		data, err := rb.ToBytes()
		require.NoError(t, err)
		assert.Equal(t, 8208, len(data))
		assert.Equal(t, mustDecodeHex(t, tc.header), data[:16])
		back, _ := readPortable(t, data)
		assert.EqualValues(t, tc.n, back.GetCardinality())
		assert.NoError(t, back.Validate())
		again, err := back.ToBytes()
		require.NoError(t, err)
		assert.Equal(t, data, again)
	}
}

// Gap 19: the shared files hold a known set and re-serialize to the same bytes.
func TestSerializationSharedFilesExactContent(t *testing.T) {
	expected := NewBitmap()
	for k := uint32(0); k < 100000; k += 1000 {
		expected.Add(k)
	}
	for k := uint32(100000); k < 200000; k++ {
		expected.Add(3 * k)
	}
	expected.AddRange(700000, 800000)

	for _, tc := range []struct {
		file string
		runs int
	}{
		{"testdata/bitmapwithruns.bin", 3},
		{"testdata/bitmapwithoutruns.bin", 0},
	} {
		data, err := os.ReadFile(tc.file)
		require.NoError(t, err)
		rb, n := readPortable(t, data)
		assert.EqualValues(t, len(data), n)
		assert.True(t, rb.Equals(expected), tc.file)
		assert.EqualValues(t, 200100, rb.GetCardinality())
		stats := rb.Stats()
		assert.EqualValues(t, 11, stats.Containers)
		assert.EqualValues(t, tc.runs, stats.RunContainers)
		assert.NoError(t, rb.Validate())
		out, err := rb.ToBytes()
		require.NoError(t, err)
		assert.Equal(t, data, out, tc.file)
	}
}

// Gap 20: several bitmaps back to back in one stream or buffer.
func TestSerializationBackToBack(t *testing.T) {
	a := BitmapOf(1, 2, 70000)
	a.AddRange(200000, 300000)
	var buf bytes.Buffer
	_, err := NewBitmap().WriteTo(&buf)
	require.NoError(t, err)
	_, err = a.WriteTo(&buf)
	require.NoError(t, err)
	_, err = a.WriteTo(&buf)
	require.NoError(t, err)
	data := buf.Bytes()

	r := bytes.NewReader(data)
	first := BitmapOf(5)
	n, err := first.ReadFrom(r)
	require.NoError(t, err)
	assert.EqualValues(t, 8, n)
	assert.True(t, first.IsEmpty())
	for i := 0; i < 2; i++ {
		next := NewBitmap()
		_, err = next.ReadFrom(r)
		require.NoError(t, err)
		assert.True(t, next.Equals(a))
	}
	assert.Equal(t, 0, r.Len())

	rest := data
	for i := 0; i < 3; i++ {
		rb := NewBitmap()
		n, err := rb.FromBuffer(rest)
		require.NoError(t, err)
		if i == 0 {
			assert.EqualValues(t, 8, n)
			assert.True(t, rb.IsEmpty())
		} else {
			assert.True(t, rb.Equals(a))
		}
		rest = rest[n:]
	}
	assert.Empty(t, rest)
}

// Gap 22: streams that return short reads (the offset header of 200 containers is skipped).
func TestSerializationShortReads(t *testing.T) {
	rb := NewBitmap()
	for i := uint32(0); i < 200; i++ {
		rb.Add(i << 16)
	}
	rb.AddRange(5<<16, 5<<16+10000)
	data, err := rb.ToBytes()
	require.NoError(t, err)
	for name, r := range map[string]func() io.Reader{
		"OneByteReader": func() io.Reader { return iotest.OneByteReader(bytes.NewReader(data)) },
		"HalfReader":    func() io.Reader { return iotest.HalfReader(bytes.NewReader(data)) },
		"DataErrReader": func() io.Reader { return iotest.DataErrReader(bytes.NewReader(data)) },
	} {
		back := NewBitmap()
		n, err := back.ReadFrom(r())
		require.NoError(t, err, name)
		assert.EqualValues(t, len(data), n, name)
		assert.True(t, back.Equals(rb), name)
	}

	back := BitmapOf(1)
	_, err = back.ReadFrom(iotest.ErrReader(errors.New("boom")))
	assert.Error(t, err)
	assertUsableAfterFailure(t, back, []uint32{1})

	back = BitmapOf(1)
	_, err = back.ReadFrom(iotest.TimeoutReader(iotest.OneByteReader(bytes.NewReader(data))))
	assert.Error(t, err)
	assertUsableAfterFailure(t, back, []uint32{1})
}

// Gap 23: deserializing from an unaligned position.
func TestSerializationUnalignedBuffer(t *testing.T) {
	data, err := os.ReadFile("testfrozendata/mixed.portable")
	require.NoError(t, err)
	want, _ := readPortable(t, data)
	storage := make([]byte, len(data)+8)
	for off := 0; off < 8; off++ {
		copy(storage[off:], data)
		rb := NewBitmap()
		_, err := rb.FromBuffer(storage[off : off+len(data)])
		require.NoError(t, err)
		assert.True(t, rb.Equals(want), off)
		assert.NoError(t, rb.Validate())
	}
}

// Gap 16: small bitmaps with 1 to 15 containers round-trip through every API.
func TestSerializationSmallContainerCounts(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	for keys := 1; keys <= 15; keys++ {
		for _, optimize := range []bool{false, true} {
			rb := NewBitmap()
			for k := 0; k < keys; k++ {
				base := uint32(k) * 3 << 16
				switch r.Intn(3) {
				case 0:
					for i := 0; i < 1+r.Intn(100); i++ {
						rb.Add(base + uint32(r.Intn(65536)))
					}
				case 1:
					for i := 0; i < 5000; i++ {
						rb.Add(base + uint32(r.Intn(65536)))
					}
				default:
					start := uint64(base) + uint64(r.Intn(1000))
					rb.AddRange(start, start+uint64(1+r.Intn(20000)))
				}
			}
			if optimize {
				rb.RunOptimize()
			}
			data, err := rb.ToBytes()
			require.NoError(t, err)
			assert.EqualValues(t, len(data), rb.GetSerializedSizeInBytes())
			for _, d := range deserializers {
				back := NewBitmap()
				require.NoError(t, d.read(back, data))
				assert.True(t, back.Equals(rb), "%s keys=%d", d.name, keys)
				assert.NoError(t, back.Validate())
			}
		}
	}
}

// Gap 24: GetSerializedSizeInBytes never exceeds BoundSerializedSizeInBytes
// (port of Java TestSerializedSize).
func TestBoundSerializedSizeInBytes(t *testing.T) {
	check := func(rb *Bitmap, universe uint64) {
		t.Helper()
		bound := BoundSerializedSizeInBytes(rb.GetCardinality(), universe)
		assert.LessOrEqual(t, rb.GetSerializedSizeInBytes(), bound)
		rb.RunOptimize()
		assert.LessOrEqual(t, rb.GetSerializedSizeInBytes(), bound)
	}
	check(NewBitmap(), 0)
	for k := uint32(0); k < 100000; k += 100 {
		check(BitmapOf(k), uint64(k)+1)
		rb := NewBitmap()
		rb.AddRange(0, uint64(k)+1)
		check(rb, uint64(k)+1)
	}
	for _, scale := range []uint32{15, 225, 3375} {
		rb := NewBitmap()
		universe := uint64(0)
		for k := uint32(0); k < 100000; k++ {
			rb.Add(scale * k)
			universe = uint64(scale*k) + 1
		}
		check(rb, universe)
	}
	for stepsize := uint64(1); stepsize < 32; stepsize += 5 {
		for step := uint64(1); step < 500; step += 37 {
			rb := NewBitmap()
			universe := uint64(0)
			for i := uint64(0); i < step; i++ {
				rb.AddRange(i<<16, i<<16+stepsize)
				universe = i<<16 + stepsize
			}
			check(rb, universe)
		}
	}
	// primes
	primes := []uint32{}
	for n := uint32(3); len(primes) < 10000; n++ {
		isPrime := true
		for d := uint32(2); d*d <= n; d++ {
			if n%d == 0 {
				isPrime = false
				break
			}
		}
		if isPrime {
			primes = append(primes, n)
		}
	}
	rb := BitmapOf(primes...)
	check(rb, uint64(primes[len(primes)-1])+1)
	check(rb, 1<<31-1)
}

// Gap 36: addoffsetinput.bin (from CRoaring): deserialize, add an offset, round-trip.
func TestSerializationAddOffsetInput(t *testing.T) {
	data, err := os.ReadFile("testdata/addoffsetinput.bin")
	require.NoError(t, err)
	rb, n := readPortable(t, data)
	assert.EqualValues(t, len(data), n)
	require.NoError(t, rb.Validate())
	assert.EqualValues(t, 933468, rb.GetCardinality())

	const offset = 4107040
	shifted := AddOffset64(rb, offset)
	require.NoError(t, shifted.Validate())
	out, err := shifted.ToBytes()
	require.NoError(t, err)
	back, _ := readPortable(t, out)
	require.NoError(t, back.Validate())
	assert.Equal(t, rb.GetCardinality(), back.GetCardinality())
	it := back.Iterator()
	rb.Iterate(func(x uint32) bool {
		if !it.HasNext() {
			t.Fatalf("missing value %d", x+offset)
		}
		if v := it.Next(); v != x+offset {
			t.Fatalf("expected %d, got %d", x+offset, v)
		}
		return true
	})
	assert.False(t, it.HasNext())
}

// Gap 37: CRoaring issue 245: a run of 9 values then 2047 isolated values.
func TestSerializationIssue245(t *testing.T) {
	rb := NewBitmap()
	rb.AddRange(0, 9)
	for i := uint32(0); i < 2047; i++ {
		rb.Add(10 + 2*i)
	}
	rb.RunOptimize()
	roundTrip := func() {
		data, err := rb.ToBytes()
		require.NoError(t, err)
		assert.EqualValues(t, len(data), rb.GetSerializedSizeInBytes())
		back, _ := readPortable(t, data)
		assert.True(t, back.Equals(rb))
		assert.NoError(t, back.Validate())
	}
	roundTrip()
	rb.Add(10 + 2*2047)
	rb.RunOptimize()
	roundTrip()
	for i := uint32(2048); i < 2100; i++ {
		rb.Add(10 + 2*i)
	}
	roundTrip()
}
