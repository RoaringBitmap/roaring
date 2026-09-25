package roaring64

import "testing"

func TestOrCardinalityContainers(t *testing.T) {
	const base uint64 = 1 << 40
	array := BitmapOf(base+1, base+3, base+63, base+64, base+65535)
	bitmap := New()
	for i := uint64(0); i < 65536; i += 2 {
		bitmap.Add(base + i)
	}
	runs := New()
	runs.AddRange(base+100, base+60000)
	otherKey := BitmapOf(base + (1 << 48) + 1)
	cases := []struct {
		name   string
		bitmap *Bitmap
	}{
		{"empty", New()}, {"array", array}, {"bitmap", bitmap},
		{"run", runs}, {"other-key", otherKey},
	}
	for _, left := range cases {
		for _, right := range cases {
			t.Run(left.name+"/"+right.name, func(t *testing.T) {
				x, y := left.bitmap.Clone(), right.bitmap.Clone()
				want := Or(x, y).GetCardinality()
				if got := x.OrCardinality(y); got != want {
					t.Fatalf("got %d, want %d", got, want)
				}
				allocs := testing.AllocsPerRun(10, func() { x.OrCardinality(y) })
				if allocs != 0 {
					t.Errorf("OrCardinality allocated %g times", allocs)
				}
				if !x.Equals(left.bitmap) || !y.Equals(right.bitmap) {
					t.Fatal("OrCardinality modified an input")
				}
			})
		}
	}
}

func BenchmarkOrCardinalityArray(b *testing.B) {
	x, y := New(), New()
	for high := uint64(1); high <= 4; high++ {
		for low := uint64(0); low < 8192; low += 4 {
			x.Add(high<<32 | low)
			x.Add(high<<32 | (low + 2))
			y.Add(high<<32 | low)
			y.Add(high<<32 | (low + 1))
		}
	}
	var got uint64
	b.ReportAllocs()
	for b.Loop() {
		got = x.OrCardinality(y)
	}
	if want := uint64(4 * 6144); got != want {
		b.Fatalf("got %d, want %d", got, want)
	}
}
