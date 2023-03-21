//go:build go1.18
// +build go1.18

package roaring64

import "testing"

func FuzzBsiStreaming(f *testing.F) {
	f.Fuzz(func(t *testing.T, b []byte) {
		slice, err := bytesToBsiColValPairs(b)
		if err != nil {
			t.SkipNow()
		}
		cols := make(map[uint64]struct{}, len(slice))
		for _, pair := range slice {
			_, ok := cols[pair.col]
			if ok {
				t.Skip("duplicate column")
			}
			cols[pair.col] = struct{}{}
		}
		testBsiRoundTrip(t, slice)
	})
}
