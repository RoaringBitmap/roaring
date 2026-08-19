package roaring

import "testing"

func TestAndStoreSlice(t *testing.T) {
	for _, n := range []int{0, 1, 3, 4, 5, bitmapContainerSize} {
		left := make([]uint64, n)
		right := make([]uint64, n)
		for i := range left {
			left[i] = uint64(i+1) * 0x5555555555555555
			right[i] = ^uint64(i * 3)
		}

		got := make([]uint64, n)
		andStoreSlice(got, left, right)
		for i := range got {
			want := left[i] & right[i]
			if got[i] != want {
				t.Fatalf("separate destination len=%d index=%d: got %#x, want %#x", n, i, got[i], want)
			}
		}

		inPlace := append([]uint64(nil), left...)
		andStoreSlice(inPlace, inPlace, right)
		for i := range inPlace {
			want := left[i] & right[i]
			if inPlace[i] != want {
				t.Fatalf("in-place destination len=%d index=%d: got %#x, want %#x", n, i, inPlace[i], want)
			}
		}
	}
}
