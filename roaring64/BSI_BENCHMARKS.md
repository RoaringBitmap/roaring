# BSI64 Benchmarks

These notes capture local benchmark results for the BSI64 `BatchEqual` and
comparison paths. They are intended as reproducible PR evidence, not as
contractual performance guarantees.

Environment:

- CPU: 12th Gen Intel(R) Core(TM) i7-1255U
- OS/arch: linux/amd64
- Package: `github.com/RoaringBitmap/roaring/v2/roaring64`

Commands:

```sh
go test ./roaring64 -count=1
go test ./roaring64 -run '^$' -bench 'BenchmarkBSI64BatchEqual' -benchmem -count 3
go test ./roaring64 -run '^$' -bench 'BenchmarkBSI64Compare(Big)?Value|BenchmarkBSI64BatchEqual(Big)?LargeAgeFixture' -benchmem -count 1
```

Representative results:

| Benchmark | Before | After | Notes |
| --- | ---: | ---: | --- |
| `BenchmarkBSI64BatchEqualLargeAgeFixture` | ~13-14s/op, ~12.4GB/op | ~145-205ms/op, ~25.5MB/op | Avoids row-by-row `GetBigValue` for int64-width values. |
| `BenchmarkBSI64BatchEqualM128Scattered` | ~1.25s/op, ~458MB/op | ~11-17ms/op, ~12.5MB/op | Detects complete bit-cube value patterns. |
| `BenchmarkBSI64CompareValueEQLargeAgeFixture` | ~4.44s/op, ~461MB/op | ~100-118ms/op, ~19.7MB/op | `EQ` delegates to optimized `BatchEqual`. |
| `BenchmarkBSI64CompareValueRangeLargeAgeFixture` | ~7.49s/op, ~501MB/op | ~204-224ms/op, ~122.6MB/op | Uses bitmap-native signed int64 comparison. |
| `BenchmarkBSI64CompareValueGELargeAgeFixture` | ~3.45s/op, ~500MB/op | ~168-184ms/op, ~82.3MB/op | Uses bitmap-native signed int64 comparison. |

Compatibility:

- Public method signatures are unchanged.
- `CompareBigValue` and `BatchEqualBig` internally delegate to the optimized
  int64 paths only when the BSI and query values fit in signed 64-bit space.
- True wider-than-64-bit values continue to use the existing generic paths.
- `BatchEqualBig` now keys values by sign and magnitude so positive and negative
  values with the same magnitude do not collide.
