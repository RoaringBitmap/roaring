package roaring64

import (
	"math/rand"
	"testing"
)

var containsSink int

func BenchmarkContains64(b *testing.B) {
	const nq = 4096
	type cfg struct {
		name  string
		build func(r *rand.Rand) *Bitmap
	}
	cfgs := []cfg{
		{"hi1_array", func(r *rand.Rand) *Bitmap {
			rb := New()
			for i := 0; i < 64; i++ {
				for j := 0; j < 1000; j++ {
					rb.Add(uint64(i)<<16 | uint64(r.Intn(1<<16)))
				}
			}
			return rb
		}},
		{"hi64_sparse", func(r *rand.Rand) *Bitmap {
			rb := New()
			for h := 0; h < 64; h++ {
				for j := 0; j < 256; j++ {
					rb.Add(uint64(h)<<32 | uint64(r.Intn(16))<<16 | uint64(r.Intn(1<<16)))
				}
			}
			return rb
		}},
		{"hi1024_sparse", func(r *rand.Rand) *Bitmap {
			rb := New()
			for h := 0; h < 1024; h++ {
				for j := 0; j < 8; j++ {
					rb.Add(uint64(h)<<32 | uint64(r.Intn(4))<<16 | uint64(r.Intn(1<<16)))
				}
			}
			return rb
		}},
	}
	for _, c := range cfgs {
		b.Run(c.name, func(b *testing.B) {
			r := rand.New(rand.NewSource(42))
			rb := c.build(r)
			vals := rb.ToArray()
			maxv := vals[len(vals)-1]
			q := make([]uint64, nq)
			for i := range q {
				if i%2 == 0 {
					q[i] = vals[r.Intn(len(vals))]
				} else {
					// random value with a plausible high part and low part
					hi := uint64(r.Int63n(int64(maxv>>32) + 1))
					q[i] = hi<<32 | uint64(r.Intn(16))<<16 | uint64(r.Intn(1<<16))
				}
			}
			b.ResetTimer()
			for j := 0; j < b.N; j++ {
				n := 0
				for _, x := range q {
					if rb.Contains(x) {
						n++
					}
				}
				containsSink += n
			}
			b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*len(q)), "ns/query")
		})
	}
}
