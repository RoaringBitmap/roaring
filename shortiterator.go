package goroaring

type ShortIterable interface {
	HasNext() bool
	Next() short
}

type ShortIterator struct {
	slice []short
	loc   int
}

func (self *ShortIterator) HasNext() bool {
	return self.loc < len(self.slice)
}

func (self *ShortIterator) Next() short {
	a := self.slice[self.loc]
	self.loc++
	return a
}
