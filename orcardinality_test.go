package roaring

import "testing"

func TestOrCardinalityContainers(t *testing.T) {
	array := BitmapOf(1, 3, 63, 64, 65535)
	bitmap := New()
	for i := uint32(0); i < 65536; i += 2 {
		bitmap.Add(i)
	}
	runs := New()
	runs.AddRange(100, 60000)
	otherKey := BitmapOf((1 << 16) + 1)
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
	for _, tc := range []struct {
		name                              string
		leftSize, rightSize, step, offset int
	}{
		{"half-overlap", 4096, 4096, 2, 0},
		{"identical", 4096, 4096, 2, -1},
		{"interleaved", 4096, 4096, 2, 1},
		{"separated", 4096, 4096, 2, 16384},
		{"skew", 4096, 16, 512, -1},
		{"tiny", 16, 16, 2, 0},
	} {
		b.Run(tc.name, func(b *testing.B) {
			x, y := New(), New()
			for high := uint32(0); high < 4; high++ {
				for i := 0; i < tc.leftSize; i++ {
					x.Add(high<<16 | uint32(i*2))
				}
				for i := 0; i < tc.rightSize; i++ {
					v := i * tc.step
					if tc.offset == 0 {
						v = (i/2)*4 + i%2
					} else if tc.offset > 0 {
						v += tc.offset
					}
					y.Add(high<<16 | uint32(v))
				}
			}
			want := Or(x, y).GetCardinality()
			var got uint64
			b.ReportAllocs()
			for b.Loop() {
				got = x.OrCardinality(y)
			}
			if got != want {
				b.Fatalf("got %d, want %d", got, want)
			}
		})
	}
}

func BenchmarkOrCardinalityContainers(b *testing.B) {
	array, bitmap, runs := New(), New(), New()
	for high := uint32(0); high < 4; high++ {
		base := high << 16
		for i := uint32(0); i < 65536; i += 2 {
			bitmap.Add(base + i)
		}
		for i := uint32(0); i < 8192; i += 2 {
			array.Add(base + i)
		}
		runs.AddRange(uint64(base)+100, uint64(base)+60000)
	}
	for _, tc := range []struct {
		name string
		x, y *Bitmap
	}{
		{"bitmap-bitmap", bitmap, bitmap.Clone()},
		{"array-bitmap", array, bitmap},
		{"run-bitmap", runs, bitmap},
		{"run-array", runs, array},
		{"run-run", runs, runs.Clone()},
	} {
		b.Run(tc.name, func(b *testing.B) {
			want := Or(tc.x, tc.y).GetCardinality()
			var got uint64
			b.ReportAllocs()
			for b.Loop() {
				got = tc.x.OrCardinality(tc.y)
			}
			if got != want {
				b.Fatalf("got %d, want %d", got, want)
			}
		})
	}
}
