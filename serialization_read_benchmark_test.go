package roaring

import (
	"fmt"
	"io"
	"os"
	"testing"
)

type benchmarkCountingReader struct {
	reader io.Reader
	calls  int
}

func (r *benchmarkCountingReader) Read(p []byte) (int, error) {
	r.calls++
	return r.reader.Read(p)
}

type benchmarkShortReader struct {
	reader io.Reader
	max    int
}

func (r *benchmarkShortReader) Read(p []byte) (int, error) {
	if len(p) > r.max {
		p = p[:r.max]
	}
	return r.reader.Read(p)
}

func benchmarkNoRunSparseData(containers int) []byte {
	bitmap := New()
	for i := uint32(0); i < uint32(containers); i++ {
		bitmap.Add(i<<16 | 1)
	}
	data, err := bitmap.ToBytes()
	if err != nil {
		panic(err)
	}
	return data
}

func BenchmarkReadFromNoRunSparse(b *testing.B) {
	for _, containers := range []int{1024, 4096, 16384} {
		data := benchmarkNoRunSparseData(containers)
		b.Run(fmt.Sprintf("file-%d", containers), func(b *testing.B) {
			benchmarkReadFromNoRunSparse(b, data, containers, 0)
		})
		b.Run(fmt.Sprintf("short-%d", containers), func(b *testing.B) {
			benchmarkReadFromNoRunSparse(b, data, containers, 257)
		})
	}
}

func benchmarkReadFromNoRunSparse(b *testing.B, data []byte, containers, maxRead int) {
	b.Helper()

	file, err := os.CreateTemp(b.TempDir(), "roaring-benchmark-")
	if err != nil {
		b.Fatal(err)
	}
	if _, err := file.Write(data); err != nil {
		file.Close()
		b.Fatal(err)
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		file.Close()
		b.Fatal(err)
	}
	defer file.Close()

	countingReader := &benchmarkCountingReader{reader: file}
	var reader io.Reader = countingReader
	if maxRead > 0 {
		reader = &benchmarkShortReader{reader: countingReader, max: maxRead}
	}

	b.ResetTimer()
	var totalReads int
	for b.Loop() {
		b.StopTimer()
		if _, err := file.Seek(0, io.SeekStart); err != nil {
			b.StartTimer()
			b.Fatal(err)
		}
		countingReader.calls = 0
		b.StartTimer()

		bitmap := New()
		if n, err := bitmap.ReadFrom(reader); err != nil {
			b.Fatal(err)
		} else if n != int64(len(data)) || bitmap.GetCardinality() != uint64(containers) {
			b.Fatalf("unexpected decode: bytes=%d cardinality=%d", n, bitmap.GetCardinality())
		}
		totalReads += countingReader.calls
	}
	b.ReportMetric(float64(totalReads)/float64(b.N), "read-calls/op")
}
