package roaring

func andStoreSliceGo(dst, s, m []uint64) {
	for i := range dst {
		dst[i] = s[i] & m[i]
	}
}
