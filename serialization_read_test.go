package roaring

import (
	"bytes"
	"io"
	"os"
	"testing"
	"unsafe"
)

type readCountingReader struct {
	reader io.Reader
	calls  int
	max    int
}

func (r *readCountingReader) Read(p []byte) (int, error) {
	r.calls++
	if r.max > 0 && len(p) > r.max {
		p = p[:r.max]
	}
	return r.reader.Read(p)
}

func noRunSparseBitmap(containers int) *Bitmap {
	bitmap := New()
	for i := uint32(0); i < uint32(containers); i++ {
		bitmap.Add(i<<16 | 1)
	}
	return bitmap
}

func mixedNoRunBitmap() *Bitmap {
	bitmap := New()
	bitmap.Add(1<<16 | 7)
	for low := uint32(0); low <= 8192; low += 2 {
		bitmap.Add(2<<16 | low)
	}
	bitmap.Add(3<<16 | 9)
	return bitmap
}

func serializedBitmap(t *testing.T, bitmap *Bitmap) []byte {
	t.Helper()
	data, err := bitmap.ToBytes()
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestReadFromNoRunBatchesPayloads(t *testing.T) {
	want := noRunSparseBitmap(16384)
	data := serializedBitmap(t, want)
	withSentinel := append(append([]byte(nil), data...), []byte("sentinel")...)
	file, err := os.CreateTemp(t.TempDir(), "bitmap-")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if _, err := file.Write(withSentinel); err != nil {
		t.Fatal(err)
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		t.Fatal(err)
	}

	for name, source := range map[string]io.Reader{"memory": bytes.NewReader(withSentinel), "file": file} {
		t.Run(name, func(t *testing.T) {
			reader := &readCountingReader{reader: source}
			got := New()
			read, err := got.ReadFrom(reader)
			if err != nil {
				t.Fatal(err)
			}
			if read != int64(len(data)) {
				t.Fatalf("ReadFrom consumed %d bytes, want %d", read, len(data))
			}
			if !got.Equals(want) {
				t.Fatal("decoded bitmap differs from source")
			}
			if reader.calls != 5 {
				t.Fatalf("underlying reader calls = %d, want 5", reader.calls)
			}
			remaining, err := io.ReadAll(source)
			if err != nil {
				t.Fatal(err)
			}
			if string(remaining) != "sentinel" {
				t.Fatalf("remaining stream = %q, want sentinel", remaining)
			}
		})
	}
}

func TestReadFromNoRunSplitsOnlyBetweenContainers(t *testing.T) {
	for _, tc := range []struct{ containers, calls int }{{32768, 5}, {32769, 6}, {65536, 6}} {
		want := noRunSparseBitmap(tc.containers)
		data := serializedBitmap(t, want)
		reader := &readCountingReader{reader: bytes.NewReader(data)}
		got := New()

		read, err := got.ReadFrom(reader)
		if err != nil {
			t.Fatal(err)
		}
		if read != int64(len(data)) {
			t.Fatalf("ReadFrom consumed %d bytes, want %d", read, len(data))
		}
		if !got.Equals(want) {
			t.Fatal("decoded bitmap differs from source")
		}
		if reader.calls != tc.calls {
			t.Fatalf("%d containers: underlying reader calls = %d, want %d", tc.containers, reader.calls, tc.calls)
		}
	}
}

func TestReadFromNoRunHandlesShortReads(t *testing.T) {
	want := noRunSparseBitmap(16384)
	data := serializedBitmap(t, want)
	withSentinel := append(append([]byte(nil), data...), []byte("sentinel")...)
	source := bytes.NewReader(withSentinel)
	reader := &readCountingReader{reader: source, max: 3}
	got := New()

	read, err := got.ReadFrom(reader)
	if err != nil {
		t.Fatal(err)
	}
	if read != int64(len(data)) {
		t.Fatalf("ReadFrom consumed %d bytes, want %d", read, len(data))
	}
	if !got.Equals(want) {
		t.Fatal("decoded bitmap differs from source")
	}
	remaining, err := io.ReadAll(source)
	if err != nil {
		t.Fatal(err)
	}
	if string(remaining) != "sentinel" {
		t.Fatalf("remaining stream = %q, want sentinel", remaining)
	}
}

func TestReadFromNoRunAlignsBitmapContainers(t *testing.T) {
	want := New()
	want.Add(1)
	want.Add(2)
	want.Add(3)
	for low := uint32(0); low <= arrayDefaultMaxSize; low++ {
		want.Add(1<<16 | low)
	}

	data := serializedBitmap(t, want)
	got := New()
	if _, err := got.ReadFrom(bytes.NewReader(data)); err != nil {
		t.Fatal(err)
	}

	bitmap, ok := got.highlowcontainer.containers[1].(*bitmapContainer)
	if !ok {
		t.Fatalf("container 1 has type %T, want *bitmapContainer", got.highlowcontainer.containers[1])
	}
	if address := uintptr(unsafe.Pointer(&bitmap.bitmap[0])); address%8 != 0 {
		t.Fatalf("bitmap backing address = %#x, want 8-byte alignment", address)
	}
	if !got.Equals(want) {
		t.Fatal("decoded bitmap differs from source")
	}
}

func TestReadFromNoRunPreservesContainerBoundaries(t *testing.T) {
	want := mixedNoRunBitmap()
	data := serializedBitmap(t, want)
	got := New()
	if _, err := got.ReadFrom(bytes.NewReader(data)); err != nil {
		t.Fatal(err)
	}

	want.Add(1<<16 | 8)
	want.Add(2<<16 | 1)
	want.Remove(3<<16 | 9)
	got.Add(1<<16 | 8)
	got.Add(2<<16 | 1)
	got.Remove(3<<16 | 9)
	if !got.Equals(want) {
		t.Fatal("mutating one decoded container changed another container")
	}
}

func TestReadFromNoRunTruncationReportsConsumedBytes(t *testing.T) {
	for _, containers := range []int{16384, 32769} {
		data := serializedBitmap(t, noRunSparseBitmap(containers))
		truncated := data[:len(data)-1]
		got := BitmapOf(1, 2, 3)

		read, err := got.ReadFrom(bytes.NewReader(truncated))
		if err == nil {
			t.Fatal("truncated bitmap unexpectedly decoded")
		}
		if read != int64(len(truncated)) {
			t.Fatalf("ReadFrom consumed %d bytes, want %d", read, len(truncated))
		}
		if !got.IsEmpty() {
			t.Fatal("failed decode left a nonempty bitmap")
		}
		if err := got.Validate(); err != nil {
			t.Fatal(err)
		}
		if _, err := got.ReadFrom(bytes.NewReader(data)); err != nil {
			t.Fatal(err)
		}
		if !got.Equals(noRunSparseBitmap(containers)) {
			t.Fatal("reusing the receiver after a failed decode changed its contents")
		}
	}
}

func TestFromBufferNoRunRetainsCopyOnWrite(t *testing.T) {
	data := serializedBitmap(t, mixedNoRunBitmap())
	got := New()
	if _, err := got.FromBuffer(data); err != nil {
		t.Fatal(err)
	}
	for i, needsCopy := range got.highlowcontainer.needCopyOnWrite {
		if !needsCopy {
			t.Fatalf("container %d is not marked copy-on-write", i)
		}
	}
}
