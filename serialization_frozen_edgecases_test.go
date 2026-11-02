//go:build (386 && !appengine) || (amd64 && !appengine) || (arm && !appengine) || (arm64 && !appengine) || (ppc64le && !appengine) || (mipsle && !appengine) || (mips64le && !appengine) || (mips64p32le && !appengine) || (wasm && !appengine)

package roaring

// Edge cases of the frozen format, ported from CRoaring.

import (
	"bytes"
	"encoding/binary"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Gap 26: the frozen form of an empty bitmap (CRoaring #933).
func TestFrozenEmptyBitmap(t *testing.T) {
	rb := New()
	assert.EqualValues(t, 4, rb.GetFrozenSizeInBytes())
	buf, err := rb.Freeze()
	require.NoError(t, err)
	assert.Equal(t, []byte{0xc6, 0x35, 0, 0}, buf)
	var w bytes.Buffer
	n, err := rb.WriteFrozenTo(&w)
	require.NoError(t, err)
	assert.Equal(t, 4, n)
	assert.Equal(t, buf, w.Bytes())
	view := BitmapOf(1, 2, 3)
	require.NoError(t, view.FrozenView(buf))
	assert.True(t, view.IsEmpty())
	assert.NoError(t, view.Validate())
}

// Gap 27: the maximal number of containers (65536) needs 17 bits in the header.
func TestFrozenMaxContainers(t *testing.T) {
	rb := New()
	for i := uint32(0); i < 1<<16; i++ {
		rb.Add(i << 16)
	}
	assert.EqualValues(t, (1<<16)*7+4, rb.GetFrozenSizeInBytes())
	buf, err := rb.Freeze()
	require.NoError(t, err)
	assert.Equal(t, uint32(frozenCookie|(1<<16)<<15), binary.LittleEndian.Uint32(buf[len(buf)-4:]))
	view := New()
	require.NoError(t, view.FrozenView(buf))
	assert.True(t, view.Equals(rb))
	assert.NoError(t, view.Validate())
}

// Gap 28: a frozen buffer that claims 65537 containers is rejected.
func TestFrozenTooManyContainers(t *testing.T) {
	const n = 1<<16 + 1
	buf := make([]byte, 7*n+4)
	keys := buf[2*n : 4*n]
	for i := 0; i < n; i++ {
		binary.LittleEndian.PutUint16(keys[2*i:], uint16(i))
		buf[6*n+i] = 2 // array typecode
	}
	binary.LittleEndian.PutUint32(buf[7*n:], frozenCookie|uint32(n)<<15)
	rb := New()
	assert.ErrorIs(t, rb.FrozenView(buf), ErrFrozenBitmapOverpopulated)
	assert.True(t, rb.IsEmpty())
}

// Gap 29: freezing copy-on-write bitmaps and bitmaps read with FromBuffer (CRoaring #928).
func TestFrozenCopyOnWrite(t *testing.T) {
	rb := BitmapOf(1)
	rb.AddRange(65536, 65636)
	for i := uint32(131072); i < 196608; i += 2 {
		rb.Add(i)
	}
	rb.RunOptimize()
	want, err := rb.Freeze()
	require.NoError(t, err)

	rb.SetCopyOnWrite(true)
	c := rb.Clone()
	assert.Equal(t, rb.GetFrozenSizeInBytes(), c.GetFrozenSizeInBytes())
	got, err := rb.Freeze()
	require.NoError(t, err)
	assert.Equal(t, want, got)
	got, err = c.Freeze()
	require.NoError(t, err)
	assert.Equal(t, want, got)

	portable, err := rb.ToBytes()
	require.NoError(t, err)
	fb := New()
	_, err = fb.FromBuffer(portable)
	require.NoError(t, err)
	got, err = fb.Freeze()
	require.NoError(t, err)
	assert.Equal(t, want, got)

	view := New()
	require.NoError(t, view.FrozenView(got))
	assert.True(t, view.Equals(rb))
	assert.NoError(t, view.Validate())
}

// Gap 30: malformed frozen buffers.
func TestFrozenMalformed(t *testing.T) {
	rb := BitmapOf(1, 2, 3, 100000)
	rb.AddRange(200000, 300000)
	for i := uint32(400000); i < 460000; i += 3 {
		rb.Add(i)
	}
	rb.RunOptimize()
	buf, err := rb.Freeze()
	require.NoError(t, err)

	check := func(name string, data []byte, want error) {
		t.Helper()
		view := New()
		err := view.FrozenView(data)
		if want == nil {
			assert.Error(t, err, name)
		} else {
			assert.ErrorIs(t, err, want, name)
		}
	}
	check("truncated at the end", buf[:len(buf)-1], nil)
	check("truncated at the start", buf[1:], nil)
	check("trailing byte", []byte{1, 0xc6, 0x35, 0, 0}, ErrFrozenBitmapUnexpectedData)
	check("3 bytes", []byte{0xc6, 0x35, 0}, ErrFrozenBitmapIncomplete)
	check("big endian", []byte{0, 0, 0x35, 0xc6}, ErrFrozenBitmapBigEndian)
	check("portable bytes", []byte{0x3a, 0x30, 0, 0, 0, 0, 0, 0}, ErrFrozenBitmapInvalidCookie)
	check("pyroaring 124 portable", mustDecodeHex(t, "3a300000 01000000 0000 0900 10000000 3200 3300 3400 3500 3600 3700 3800 3a00 3b00 3c00"), ErrFrozenBitmapInvalidCookie)
	one, err := BitmapOf(7).Freeze()
	require.NoError(t, err)
	for _, typecode := range []byte{0, 4} {
		bad := bytes.Clone(one)
		bad[len(bad)-5] = typecode
		check("bad typecode", bad, ErrFrozenBitmapInvalidTypecode)
	}

	_, err = rb.FreezeTo(make([]byte, rb.GetFrozenSizeInBytes()-1))
	assert.ErrorIs(t, err, ErrFrozenBitmapBufferTooSmall)
}

// Gap 23: frozen views of a buffer at every alignment.
func TestFrozenViewUnaligned(t *testing.T) {
	data, err := os.ReadFile("testfrozendata/mixed.frozen")
	require.NoError(t, err)
	portable, err := os.ReadFile("testfrozendata/mixed.portable")
	require.NoError(t, err)
	want := New()
	_, err = want.FromBuffer(portable)
	require.NoError(t, err)
	storage := make([]byte, len(data)+8)
	for off := 0; off < 8; off++ {
		copy(storage[off:], data)
		view := New()
		require.NoError(t, view.FrozenView(storage[off:off+len(data)]))
		assert.True(t, view.Equals(want), off)
	}
}
