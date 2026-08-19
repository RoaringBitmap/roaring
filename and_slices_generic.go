//go:build !amd64 || appengine
// +build !amd64 appengine

package roaring

func andStoreSlice(dst, s, m []uint64) {
	andStoreSliceGo(dst, s, m)
}
