package roaring

type shortIterable interface {
	hasNext() bool
	next() uint16
}

type shortIterator struct {
	slice []uint16
	loc   int
}

func (self *shortIterator) hasNext() bool {
	return self.loc < len(self.slice)
}

func (self *shortIterator) next() uint16 {
	a := self.slice[self.loc]
	self.loc++
	return a
}
