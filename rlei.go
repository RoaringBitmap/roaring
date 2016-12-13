package roaring

///////////////////////////////////////////////////
//
// container interface methods for runContainer16
//
///////////////////////////////////////////////////

// compile time verify we meet interface requirements
var _ container = &runContainer16{}

func (rc *runContainer16) clone() container {
	return newRunContainer16CopyIv(rc.Iv)
}

func (rc *runContainer16) and(a container) container {
	switch c := a.(type) {
	case *runContainer16:
		return rc.intersect(c)
	case *arrayContainer:
		return rc.andArray(c)
	case *bitmapContainer:
		return rc.andBitmapContainer(c)
	}
	panic("unsupported container type")
}

// andBitmapContainer finds the intersection of rc and b.
func (rc *runContainer16) andBitmapContainer(bc *bitmapContainer) container {
	bc2 := newBitmapContainerFromRun(rc)
	return bc2.andBitmap(bc)
}

func (rc *runContainer16) andArray(ac *arrayContainer) container {
	out := newRunContainer16()
	for _, p := range rc.Iv {
		for i := p.Start; i <= p.Last; i++ {
			if ac.contains(i) {
				out.Add(i)
			}
		}
	}
	return out
}

func (rc *runContainer16) iand(a container) container {
	switch c := a.(type) {
	case *runContainer16:
		return rc.inplaceIntersect(c)
	case *arrayContainer:
		return rc.iandArray(c)
	case *bitmapContainer:
		return rc.iandBitmapContainer(c)
	}
	panic("unsupported container type")
}

func (rc *runContainer16) inplaceIntersect(rc2 *runContainer16) container {
	// TODO: optimize by doing less allocation, possibly?

	// sect will be new
	sect := rc.intersect(rc2)
	*rc = *sect
	return rc
}

func (rc *runContainer16) iandBitmapContainer(bc *bitmapContainer) container {
	isect := rc.andBitmapContainer(bc)
	*rc = *newRunContainer16FromContainer(isect)
	return rc
}

func (rc *runContainer16) iandArray(ac *arrayContainer) container {
	// TODO: optimize by doing less allocation, possibly?
	out := newRunContainer16()
	for _, p := range rc.Iv {
		for i := p.Start; i <= p.Last; i++ {
			if ac.contains(i) {
				out.Add(i)
			}
		}
	}
	*rc = *out
	return rc
}

func (rc *runContainer16) andNot(a container) container {
	switch c := a.(type) {
	case *arrayContainer:
		return rc.andNotArray(c)
	case *bitmapContainer:
		return rc.andNotBitmap(c)
	case *runContainer16:
		return rc.andNotRunContainer16(c)
	}
	panic("unsupported container type")
}

func (rc *runContainer16) fillLeastSignificant16bits(x []uint32, i int, mask uint32) {
	k := 0
	var val int64
	for _, p := range rc.Iv {
		n := p.runlen()
		for j := int64(0); j < n; j++ {
			val = int64(p.Start) + j
			x[k+i] = uint32(val) | mask
			k++
		}
	}
}

func (rc *runContainer16) getShortIterator() shortIterable {
	return rc.NewRunIterator16()
}

// add the values in the range [firstOfRange,lastofRange). lastofRange
// is still abe to express 2^16 because it is an int not an uint16.
func (rc *runContainer16) iaddRange(firstOfRange, lastOfRange int) container {
	addme := newRunContainer16TakeOwnership([]interval16{
		{
			Start: uint16(firstOfRange),
			Last:  uint16(lastOfRange - 1),
		},
	})
	*rc = *rc.union(addme)
	return rc
}

// remove the values in the range [firstOfRange,lastOfRange)
func (rc *runContainer16) iremoveRange(firstOfRange, lastOfRange int) container {
	x := interval16{Start: uint16(firstOfRange), Last: uint16(lastOfRange - 1)}
	rc.isubtract(x)
	return rc
}

// not flip the values in the range [firstOfRange,lastOfRange)
func (rc *runContainer16) not(firstOfRange, lastOfRange int) container {
	return rc.Not(firstOfRange, lastOfRange)
}

// Not flips the values in the range [firstOfRange,endx).
// This is not inplace. Only the returned value has the flipped bits.
//
// Currently implemented as (!A intersect B) union (A minus B),
// where A is rc, and B is the supplied [firstOfRange, endx) interval.
//
// TODO(time optimization): convert this to a single pass
// algorithm by copying AndNotRunContainer16() and modifying it.
// Current routine is correct but
// makes 2 more passes through the arrays than should be
// strictly necessary. Measure both ways though--this may not matter.
//
func (rc *runContainer16) Not(firstOfRange, endx int) *runContainer16 {

	//p("top of Not with interval [%v, %v): rc is %s", firstOfRange, endx, rc.String())
	if firstOfRange >= endx {
		//p("returning early with clone, first >= endx")
		return rc.Clone()
	}

	a := rc
	// algo:
	// (!A intersect B) union (A minus B)

	nota := a.invert()

	bs := []interval16{interval16{Start: uint16(firstOfRange), Last: uint16(endx - 1)}}
	b := newRunContainer16TakeOwnership(bs)
	//p("b is %s", b)

	notAintersectB := nota.intersect(b)
	//p("notAintersectB is %s", notAintersectB)

	aMinusB := a.AndNotRunContainer16(b)
	//p("aMinusB is %s", aMinusB)

	rc2 := notAintersectB.union(aMinusB)
	//p("rc = ((!A intersect B) union (A minus B)) is %s", rc2)
	return rc2
}

// equals is now logical equals; it does not require the
// same underlying container type.
func (rc *runContainer16) equals(o interface{}) bool {
	srb, ok := o.(*runContainer16)

	if !ok {
		// maybe value instead of pointer
		val, valok := o.(runContainer16)
		if valok {
			//p("was runContainer16 value...")
			srb = &val
			ok = true
		}
	}
	if ok {
		//p("both rc16")
		// Check if the containers are the same object.
		if rc == srb {
			//p("same object")
			return true
		}

		if len(srb.Iv) != len(rc.Iv) {
			//p("Iv len differ")
			return false
		}

		for i, v := range rc.Iv {
			if v != srb.Iv[i] {
				//p("differ at Iv i=%v, srb.Iv[i]=%v, rc.Iv[i]=%v", i, srb.Iv[i], rc.Iv[i])
				return false
			}
		}
		//p("all intervals same, returning true")
		return true
	}

	//p("not both rc16; o is %T / val=%#v", o, o)
	bc, ok := o.(container)
	if ok {
		// use generic comparison
		if bc.getCardinality() != rc.getCardinality() {
			//p("card differ bc.card=%v, rc.card=%v", bc.getCardinality(), rc.getCardinality())
			return false
		}
		rit := rc.getShortIterator()
		bit := bc.getShortIterator()

		//k := 0
		for rit.hasNext() {
			if bit.next() != rit.next() {
				//p("differ at pos %k", k)
				return false
			}
			//k++
		}
		return true
	}
	//p("o was not a container!")
	return false
}

func (rc *runContainer16) iaddReturnMinimized(x uint16) container {
	rc.Add(x)
	return rc
}

func (rc *runContainer16) iadd(x uint16) (wasNew bool) {
	return rc.Add(x)
}

func (rc *runContainer16) iremoveReturnMinimized(x uint16) container {
	rc.removeKey(x)
	return rc
}

func (rc *runContainer16) iremove(x uint16) bool {
	return rc.removeKey(x)
}

func (rc *runContainer16) or(a container) container {
	switch c := a.(type) {
	case *runContainer16:
		return rc.union(c)
	case *arrayContainer:
		return rc.orArray(c)
	case *bitmapContainer:
		return rc.orBitmapContainer(c)
	}
	panic("unsupported container type")
}

// orBitmapContainer finds the union of rc and bc.
func (rc *runContainer16) orBitmapContainer(bc *bitmapContainer) container {
	bc2 := newBitmapContainerFromRun(rc)
	return bc.or(bc2)
}

// orArray finds the union of rc and ac.
func (rc *runContainer16) orArray(ac *arrayContainer) container {
	out := ac.clone()
	for _, p := range rc.Iv {
		for i := p.Start; i <= p.Last; i++ {
			out.iadd(i)
		}
	}
	return out
}

func (rc *runContainer16) ior(a container) container {
	switch c := a.(type) {
	case *runContainer16:
		return rc.inplaceUnion(c)
	case *arrayContainer:
		return rc.iorArray(c)
	case *bitmapContainer:
		return rc.iorBitmapContainer(c)
	}
	panic("unsupported container type")
}

func (rc *runContainer16) inplaceUnion(rc2 *runContainer16) container {
	for _, p := range rc2.Iv {
		for i := p.Start; i <= p.Last; i++ {
			rc.Add(i)
		}
	}
	return rc
}

func (rc *runContainer16) iorBitmapContainer(bc *bitmapContainer) container {

	it := bc.getShortIterator()
	for it.hasNext() {
		rc.Add(it.next())
	}
	return rc
}

func (rc *runContainer16) iorArray(ac *arrayContainer) container {
	it := ac.getShortIterator()
	for it.hasNext() {
		rc.Add(it.next())
	}
	return rc
}

// lazyIOR is described (not yet implemented) in
// this nice note from @lemire on
// https://github.com/RoaringBitmap/roaring/pull/70#issuecomment-263613737
//
// Description of lazyOR and lazyIOR from @lemire:
//
// Lazy functions are optional and can be simply
// wrapper around non-lazy functions.
//
// The idea of "laziness" is as follows. It is
// inspired by the concept of lazy evaluation
// you might be familiar with (functional programming
// and all that). So a roaring bitmap is
// such that all its containers are, in some
// sense, chosen to use as little memory as
// possible. This is nice. Also, all bitsets
// are "cardinality aware" so that you can do
// fast rank/select queries, or query the
// cardinality of the whole bitmap... very fast,
// without latency.
//
// However, imagine that you are aggregating 100
// bitmaps together. So you OR the first two, then OR
// that with the third one and so forth. Clearly,
// intermediate bitmaps don't need to be as
// compressed as possible, right? They can be
// in a "dirty state". You only need the end
// result to be in a nice state... which you
// can achieve by calling repairAfterLazy at the end.
//
// The Java/C code does something special for
// the in-place lazy OR runs. The idea is that
// instead of taking two run containers and
// generating a new one, we actually try to
// do the computation in-place through a
// technique invented by @gssiyankai (pinging him!).
// What you do is you check whether the host
// run container has lots of extra capacity.
// If it does, you move its data at the end of
// the backing array, and then you write
// the answer at the beginning. What this
// trick does is minimize memory allocations.
//
func (rc *runContainer16) lazyIOR(a container) container {
	panic("TODO: runContainer16.lazyIOR not yet implemented")

	/*
		switch c := a.(type) {
		case *arrayContainer:
			return rc.lazyIorArray(c)
		case *bitmapContainer:
			return rc.lazyIorBitmap(c)
		case *runContainer16:
			return rc.lazyIorRunContainer16(c)
		}
		panic("unsupported container type")
	*/
}

// lazyOR is described above in lazyIOR.
func (rc *runContainer16) lazyOR(a container) container {
	panic("TODO: runContainer16.lazyOR not yet implemented")

	/*
		switch c := a.(type) {
		case *arrayContainer:
			return rc.lazyOrArray(c)
		case *bitmapContainer:
			return rc.lazyOrBitmap(c)
		case *runContainer16:
			return rc.lazyOrRunContainer16(c)
		}
		panic("unsupported container type")
	*/
}

func (rc *runContainer16) intersects(a container) bool {
	// TODO: optimize by doing inplace/less allocation, possibly?
	isect := rc.and(a)
	return isect.getCardinality() > 0
}

func (rc *runContainer16) xor(a container) container {
	switch c := a.(type) {
	case *arrayContainer:
		return rc.xorArray(c)
	case *bitmapContainer:
		return rc.xorBitmap(c)
	case *runContainer16:
		return rc.xorRunContainer16(c)
	}
	panic("unsupported container type")
}

func (rc *runContainer16) iandNot(a container) container {
	switch c := a.(type) {
	case *arrayContainer:
		return rc.iandNotArray(c)
	case *bitmapContainer:
		return rc.iandNotBitmap(c)
	case *runContainer16:
		return rc.iandNotRunContainer16(c)
	}
	panic("unsupported container type")
}

// flip the values in the range [firstOfRange,lastOfRange)
func (rc *runContainer16) inot(firstOfRange, lastOfRange int) container {
	// TODO: minimize copies, do it all inplace; not() makes a copy.
	rc = rc.Not(firstOfRange, lastOfRange)
	return rc
}

func (rc *runContainer16) getCardinality() int {
	return int(rc.cardinality())
}

func (rc *runContainer16) rank(x uint16) int {
	n := int64(len(rc.Iv))
	xx := int64(x)
	w, already, _ := rc.search(xx, nil)
	if w < 0 {
		return 0
	}
	if !already && w == n-1 {
		return rc.getCardinality()
	}
	var rnk int64
	if !already {
		for i := int64(0); i <= w; i++ {
			rnk += rc.Iv[i].runlen()
		}
		return int(rnk)
	}
	for i := int64(0); i < w; i++ {
		rnk += rc.Iv[i].runlen()
	}
	rnk += int64(x-rc.Iv[w].Start) + 1
	return int(rnk)
}

func (rc *runContainer16) selectInt(x uint16) int {
	return rc.selectInt16(x)
}

func (rc *runContainer16) andNotRunContainer16(b *runContainer16) container {
	return rc.AndNotRunContainer16(b)
}

func (rc *runContainer16) andNotArray(ac *arrayContainer) container {
	rcb := rc.toBitmapContainer()
	acb := ac.toBitmapContainer()
	return rcb.andNotBitmap(acb)
}

func (rc *runContainer16) andNotBitmap(bc *bitmapContainer) container {
	rcb := rc.toBitmapContainer()
	return rcb.andNotBitmap(bc)
}

func (rc *runContainer16) toBitmapContainer() *bitmapContainer {
	bc := newBitmapContainer()
	n := rc.getCardinality()
	bc.Cardinality = n
	it := rc.NewRunIterator16()
	for it.HasNext() {
		x := it.Next()
		i := int(x) / 64
		bc.Bitmap[i] |= (uint64(1) << uint(x%64))
	}
	return bc
}

func (rc *runContainer16) iandNotRunContainer16(x2 *runContainer16) container {
	rcb := rc.toBitmapContainer()
	x2b := x2.toBitmapContainer()
	rcb.iandNotBitmapSurely(x2b)
	// TODO: check size and optimize the return value
	// TODO: is inplace modification really required? If not, elide the copy.
	rc2 := newRunContainer16FromBitmapContainer(rcb)
	*rc = *rc2
	return rc
}

func (rc *runContainer16) iandNotArray(ac *arrayContainer) container {
	rcb := rc.toBitmapContainer()
	acb := ac.toBitmapContainer()
	rcb.iandNotBitmapSurely(acb)
	// TODO: check size and optimize the return value
	// TODO: is inplace modification really required? If not, elide the copy.
	rc2 := newRunContainer16FromBitmapContainer(rcb)
	*rc = *rc2
	return rc
}

func (rc *runContainer16) iandNotBitmap(bc *bitmapContainer) container {
	rcb := rc.toBitmapContainer()
	rcb.iandNotBitmapSurely(bc)
	// TODO: check size and optimize the return value
	// TODO: is inplace modification really required? If not, elide the copy.
	rc2 := newRunContainer16FromBitmapContainer(rcb)
	*rc = *rc2
	return rc
}

func (rc *runContainer16) xorRunContainer16(x2 *runContainer16) container {
	rcb := rc.toBitmapContainer()
	x2b := x2.toBitmapContainer()
	return rcb.xorBitmap(x2b)
}

func (rc *runContainer16) xorArray(ac *arrayContainer) container {
	rcb := rc.toBitmapContainer()
	acb := ac.toBitmapContainer()
	return rcb.xorBitmap(acb)
}

func (rc *runContainer16) xorBitmap(bc *bitmapContainer) container {
	rcb := rc.toBitmapContainer()
	return rcb.xorBitmap(bc)
}

// convert to bitmap or array *if needed*
func (rc *runContainer16) toEfficientContainer() container {

	// runContainer16SerializedSizeInBytes(numRuns)
	sizeAsRunContainer := rc.getSizeInBytes()
	sizeAsBitmapContainer := bitmapContainerSizeInBytes()
	card := int(rc.cardinality())
	sizeAsArrayContainer := arrayContainerSizeInBytes(card)
	if sizeAsRunContainer <= min(sizeAsBitmapContainer, sizeAsArrayContainer) {
		return rc
	}
	if card <= arrayDefaultMaxSize {
		ac := newArrayContainer()
		for i := range rc.Iv {
			ac.iaddRange(int(rc.Iv[i].Start), int(rc.Iv[i].Last+1))
		}
		return ac
	}
	bc := newBitmapContainerFromRun(rc)
	return bc
}

func newRunContainer16FromContainer(c container) *runContainer16 {

	switch x := c.(type) {
	case *runContainer16:
		return x.Clone()
	case *arrayContainer:
		return newRunContainer16FromArray(x)
	case *bitmapContainer:
		return newRunContainer16FromBitmapContainer(x)
	}
	panic("unsupported container type")
}
