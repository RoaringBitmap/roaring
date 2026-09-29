package roaring64

import "testing"

func TestXorInPlaceDenseAllocations(t *testing.T) {
	x, y := New(), New()
	for i := uint64(0); i < 65536; i++ {
		if i%2 == 0 {
			x.Add(1<<40 | i)
		}
		if i%4 < 2 {
			y.Add(1<<40 | i)
		}
	}
	original := x.Clone()
	allocs := testing.AllocsPerRun(10, func() {
		x.Xor(y)
		x.Xor(y)
	})
	if allocs != 0 {
		t.Fatalf("dense in-place XOR allocated %g times", allocs)
	}
	if !x.Equals(original) {
		t.Fatal("two XORs must restore the receiver")
	}
}

func BenchmarkXorInPlace64(b *testing.B) {
	for _, kind := range []string{"dense", "array", "run", "disjoint-keys"} {
		x, y := New(), New()
		for high := uint64(1); high <= 4; high++ {
			base := high << 40
			switch kind {
			case "dense":
				for i := uint64(0); i < 65536; i++ {
					if i%2 == 0 {
						x.Add(base | i)
					}
					if i%4 < 2 {
						y.Add(base | i)
					}
				}
			case "array":
				for i := uint64(0); i < 128; i++ {
					if i%2 == 0 {
						x.Add(base | i)
					}
					if i%4 < 2 {
						y.Add(base | i)
					}
				}
			case "run":
				x.AddRange(base+100, base+30000)
				y.AddRange(base+20000, base+50000)
			case "disjoint-keys":
				x.Add(base | 1)
				y.Add((base - 1<<32) | 2)
			}
		}
		for _, clone := range []bool{false, true} {
			name := kind + "/reuse"
			if clone {
				name = kind + "/clone-and-xor"
			}
			b.Run(name, func(b *testing.B) {
				x, y := x.Clone(), y.Clone()
				original := x.Clone()
				want := Xor(x, y)
				b.ReportAllocs()
				var result *Bitmap
				for b.Loop() {
					if clone {
						result = x.Clone()
						result.Xor(y)
					} else {
						x.Xor(y)
						x.Xor(y)
					}
				}
				if clone {
					if !result.Equals(want) {
						b.Fatal("wrong XOR result")
					}
				} else {
					if !x.Equals(original) {
						b.Fatal("two XORs did not restore the receiver")
					}
					b.ReportMetric(2, "xors/op")
				}
			})
		}
	}
}
