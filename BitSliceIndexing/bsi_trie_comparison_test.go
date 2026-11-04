package roaring

import (
	"github.com/RoaringBitmap/roaring/v2"
	"os"
	"sort"
)

// trieBenchmarkBSI is BatchEqual at 873a1e356cf5dd8688f43c7fcd82dccd159a9b63.
type trieBenchmarkBSI struct{ *BSI }

var compareScan = os.Getenv("BSI_COMPARISON") == "scan"

func (b trieBenchmarkBSI) BatchEqual(parallelism int, values []int64) *roaring.Bitmap {
	if compareScan {
		return b.BSI.BatchEqual(parallelism, values)
	}
	if b.eBM.IsEmpty() || len(values) == 0 {
		return roaring.NewBitmap()
	}

	bitCount := b.BitCount()

	// Deduplicate, and drop values that cannot be represented in bitCount
	// planes: GetValue can never observe such a value, so it matches no column.
	seen := make(map[uint64]struct{}, len(values))
	vals := make([]uint64, 0, len(values))
	for _, v := range values {
		u := uint64(v)
		if bitCount < 64 && (v < 0 || u >= uint64(1)<<uint(bitCount)) {
			continue
		}
		if _, ok := seen[u]; ok {
			continue
		}
		seen[u] = struct{}{}
		vals = append(vals, u)
	}
	if len(vals) == 0 {
		return roaring.NewBitmap()
	}
	sort.Slice(vals, func(i, j int) bool { return vals[i] < vals[j] })

	result := b.matchTrie(vals, bitCount-1, b.eBM, false)
	if b.runOptimized {
		result.RunOptimize()
	}
	return result
}
