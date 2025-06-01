package roaring

func Values(b *Bitmap) func(func(uint32) bool) {
	it := b.Iterator()
	return func(yield func(uint32) bool) {
		for it.HasNext() {
			if !yield(it.Next()) {
				return
			}
		}
	}
}

func Backward(b *Bitmap) func(func(uint32) bool) {
	it := b.ReverseIterator()
	return func(yield func(uint32) bool) {
		for it.HasNext() {
			if !yield(it.Next()) {
				return
			}
		}
	}
}
