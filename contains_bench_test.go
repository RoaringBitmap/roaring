package roaring

import (
	"math/rand"
	"testing"
)

// Benchmarks for Contains, isolating each container type and the key search.
// Queries are random (half hits, half misses) so branch prediction cannot
// learn the pattern.

var containsSink int

func containsQueries(r *rand.Rand, rb *Bitmap, universe uint32, n int) []uint32 {
	vals := rb.ToArray()
	q := make([]uint32, n)
	for i := range q {
		if i%2 == 0 {
			q[i] = vals[r.Intn(len(vals))]
		} else {
			q[i] = uint32(r.Int63n(int64(universe)))
		}
	}
	r.Shuffle(len(q), func(i, j int) { q[i], q[j] = q[j], q[i] })
	return q
}

func runContainsBench(b *testing.B, rb *Bitmap, q []uint32) {
	b.ResetTimer()
	for j := 0; j < b.N; j++ {
		c := 0
		for _, x := range q {
			if rb.Contains(x) {
				c++
			}
		}
		containsSink += c
	}
	b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*len(q)), "ns/query")
}

func BenchmarkContains(b *testing.B) {
	const nq = 4096
	cfgs := containsBenchConfigs
	for _, c := range cfgs {
		b.Run(c.name, func(b *testing.B) {
			r := rand.New(rand.NewSource(42))
			rb, u := c.build(r)
			q := containsQueries(r, rb, u, nq)
			runContainsBench(b, rb, q)
		})
	}
}

var containsBenchConfigs = []struct {
	name  string
	build func(r *rand.Rand) (*Bitmap, uint32)
}{
	{"array_1c_4000", func(r *rand.Rand) (*Bitmap, uint32) {
		rb := New()
		for rb.GetCardinality() < 4000 {
			rb.Add(uint32(r.Intn(1 << 16)))
		}
		return rb, 1 << 16
	}},
	{"array_1c_64", func(r *rand.Rand) (*Bitmap, uint32) {
		rb := New()
		for rb.GetCardinality() < 64 {
			rb.Add(uint32(r.Intn(1 << 16)))
		}
		return rb, 1 << 16
	}},
	{"bitmap_64c", func(r *rand.Rand) (*Bitmap, uint32) {
		rb := New()
		for i := 0; i < 64<<16; i++ {
			if r.Intn(2) == 0 {
				rb.Add(uint32(i))
			}
		}
		return rb, 64 << 16
	}},
	{"run_1c_1000runs", func(r *rand.Rand) (*Bitmap, uint32) {
		rb := New()
		for i := 0; i < 1000; i++ {
			s := uint64(i) * 65
			rb.AddRange(s, s+uint64(1+r.Intn(40)))
		}
		rb.RunOptimize()
		return rb, 1 << 16
	}},
	{"run_1c_16runs", func(r *rand.Rand) (*Bitmap, uint32) {
		rb := New()
		for i := 0; i < 16; i++ {
			s := uint64(i) * 4000
			rb.AddRange(s, s+uint64(1+r.Intn(2000)))
		}
		rb.RunOptimize()
		return rb, 1 << 16
	}},
	{"keys_4096c_sparse", func(r *rand.Rand) (*Bitmap, uint32) {
		rb := New()
		for i := 0; i < 4096; i++ {
			k := uint32(i) * 15
			for j := 0; j < 4; j++ {
				rb.Add(k<<16 | uint32(r.Intn(1<<16)))
			}
		}
		return rb, 4096 * 15 << 16
	}},
	{"keys_64c_sparse", func(r *rand.Rand) (*Bitmap, uint32) {
		rb := New()
		for i := 0; i < 64; i++ {
			k := uint32(i) * 2
			for j := 0; j < 4; j++ {
				rb.Add(k<<16 | uint32(r.Intn(1<<16)))
			}
		}
		return rb, 128 << 16
	}},
	{"mixed_1024c", func(r *rand.Rand) (*Bitmap, uint32) {
		rb := New()
		for i := 0; i < 1024; i++ {
			base := uint64(i) << 16
			switch i % 3 {
			case 0:
				for j := 0; j < 1000; j++ {
					rb.Add(uint32(base) | uint32(r.Intn(1<<16)))
				}
			case 1:
				for j := 0; j < 20000; j++ {
					rb.Add(uint32(base) | uint32(r.Intn(1<<16)))
				}
			case 2:
				for j := 0; j < 50; j++ {
					s := base + uint64(j)*1300
					rb.AddRange(s, s+uint64(1+r.Intn(600)))
				}
			}
		}
		rb.RunOptimize()
		return rb, 1024 << 16
	}},
}
