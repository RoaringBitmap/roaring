package serz

//go:generate msgp

type RunContainer16 struct {
	Iv   []Interval16
	Card int64
}

type Interval16 struct {
	Start uint16
	Last  uint16
}

type RunContainer32 struct {
	Iv   []Interval32
	Card int64
}

type Interval32 struct {
	Start uint32
	Last  uint32
}
