package roaring

func Values(b *Bitmap) func(func(uint32) bool) {
	return func(yield func(uint32) bool) {
		it := b.Iterator()
		for it.HasNext() {
			if !yield(it.Next()) {
				return
			}
		}
	}
}

func Backward(b *Bitmap) func(func(uint32) bool) {
	return func(yield func(uint32) bool) {
		it := b.ReverseIterator()
		for it.HasNext() {
			if !yield(it.Next()) {
				return
			}
		}
	}
}
