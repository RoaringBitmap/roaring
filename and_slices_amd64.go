//go:build amd64 && !appengine
// +build amd64,!appengine

package roaring

//go:noescape
func _andStoreSliceAVX2(dst, s, m []uint64)

func andStoreSlice(dst, s, m []uint64) {
	if useAVX2 {
		_andStoreSliceAVX2(dst, s, m)
		return
	}
	andStoreSliceGo(dst, s, m)
}
