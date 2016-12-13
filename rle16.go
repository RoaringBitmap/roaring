package roaring

//
// Copyright (c) 2016 by the roaring authors.
// Licensed under the Apache License, Version 2.0.
//
// We derive a few lines of code from the sort.Search
// function in the golang standard library. That function
// is Copyright 2009 The Go Authors, and licensed
// under the following BSD-style license.
/*
Copyright (c) 2009 The Go Authors. All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are
met:

   * Redistributions of source code must retain the above copyright
notice, this list of conditions and the following disclaimer.
   * Redistributions in binary form must reproduce the above
copyright notice, this list of conditions and the following disclaimer
in the documentation and/or other materials provided with the
distribution.
   * Neither the name of Google Inc. nor the names of its
contributors may be used to endorse or promote products derived from
this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
"AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

import (
	"fmt"
	"sort"
	"unsafe"
)

//go:generate msgp -unexported

// runContainer16 does run-length encoding of sets of
// uint16 integers.
type runContainer16 struct {
	iv   []interval16
	card int64

	// avoid allocation during search
	myOpts searchOptions `msg:"-"`
}

// interval16 is the internal to runContainer16
// structure that maintains the individual [Start, last]
// closed intervals.
type interval16 struct {
	start uint16
	last  uint16
}

// runlen returns the count of integers in the interval.
func (iv interval16) runlen() int64 {
	return 1 + int64(iv.last) - int64(iv.start)
}

// String produces a human viewable string of the contents.
func (iv interval16) String() string {
	return fmt.Sprintf("[%d, %d]", iv.start, iv.last)
}

func ivalString16(iv []interval16) string {
	var s string
	var j int
	var p interval16
	for j, p = range iv {
		s += fmt.Sprintf("%v:[%d, %d], ", j, p.start, p.last)
	}
	return s
}

// String produces a human viewable string of the contents.
func (rc *runContainer16) String() string {
	if len(rc.iv) == 0 {
		return "runContainer16{}"
	}
	is := ivalString16(rc.iv)
	return `runContainer16{` + is + `}`
}

// uint16Slice is a sort.Sort convenience method
type uint16Slice []uint16

// Len returns the length of p.
func (p uint16Slice) Len() int { return len(p) }

// Less returns p[i] < p[j]
func (p uint16Slice) Less(i, j int) bool { return p[i] < p[j] }

// Swap swaps elements i and j.
func (p uint16Slice) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

//msgp:ignore addHelper

// addHelper helps build a runContainer16.
type addHelper16 struct {
	runstart      uint16
	runlen        uint16
	actuallyAdded uint16
	m             []interval16
	rc            *runContainer16
}

func (ah *addHelper16) storeIval(runstart, runlen uint16) {
	mi := interval16{start: runstart, last: runstart + runlen}
	ah.m = append(ah.m, mi)
}

func (ah *addHelper16) add(cur, prev uint16, i int) {
	if cur == prev+1 {
		ah.runlen++
		ah.actuallyAdded++
	} else {
		if cur < prev {
			panic(fmt.Sprintf("newRunContainer16FromVals sees "+
				"unsorted vals; vals[%v]=cur=%v < prev=%v. Sort your vals"+
				" before calling us with alreadySorted == true.", i, cur, prev))
		}
		if cur == prev {
			// ignore duplicates
		} else {
			ah.actuallyAdded++
			ah.storeIval(ah.runstart, ah.runlen)
			ah.runstart = cur
			ah.runlen = 0
		}
	}
}

// newRunContainerRange makes a new container made of just the specified closed interval [rangestart,rangelast]
func newRunContainer16Range(rangestart uint16, rangelast uint16) *runContainer16 {
	rc := &runContainer16{}
	rc.iv = append(rc.iv, interval16{start: rangestart, last: rangelast})
	return rc
}

// newRunContainer16FromVals makes a new container from vals.
//
// For efficiency, vals should be sorted in ascending order.
// Ideally vals should not contain duplicates, but we detect and
// ignore them. If vals is already sorted in ascending order, then
// pass alreadySorted = true. Otherwise, for !alreadySorted,
// we will sort vals before creating a runContainer16 of them.
// We sort the original vals, so this will change what the
// caller sees in vals as a side effect.
func newRunContainer16FromVals(alreadySorted bool, vals ...uint16) *runContainer16 {
	// keep this in sync with newRunContainer16FromArray below

	rc := &runContainer16{}
	ah := addHelper16{rc: rc}

	if !alreadySorted {
		sort.Sort(uint16Slice(vals))
	}
	n := len(vals)
	var cur, prev uint16
	switch {
	case n == 0:
		// nothing more
	case n == 1:
		ah.m = append(ah.m, interval16{start: vals[0], last: vals[0]})
		ah.actuallyAdded++
	default:
		ah.runstart = vals[0]
		ah.actuallyAdded++
		for i := 1; i < n; i++ {
			prev = vals[i-1]
			cur = vals[i]
			ah.add(cur, prev, i)
		}
		ah.storeIval(ah.runstart, ah.runlen)
	}
	rc.iv = ah.m
	rc.card = int64(ah.actuallyAdded)
	return rc
}

// newRunContainer16FromBitmapContainer makes a new run container from bc.
func newRunContainer16FromBitmapContainer(bc *bitmapContainer) *runContainer16 {
	// todo: this could be optimized, see https://github.com/RoaringBitmap/RoaringBitmap/blob/master/src/main/java/org/roaringbitmap/RunContainer.java#L145-L192

	rc := &runContainer16{}
	ah := addHelper16{rc: rc}

	n := bc.getCardinality()
	it := bc.getShortIterator()
	var cur, prev, val uint16
	switch {
	case n == 0:
		// nothing more
	case n == 1:
		val = uint16(it.next())
		ah.m = append(ah.m, interval16{start: val, last: val})
		ah.actuallyAdded++
	default:
		prev = uint16(it.next())
		cur = uint16(it.next())
		ah.runstart = prev
		ah.actuallyAdded++
		for i := 1; i < n; i++ {
			ah.add(cur, prev, i)
			if it.hasNext() {
				prev = cur
				cur = uint16(it.next())
			}
		}
		ah.storeIval(ah.runstart, ah.runlen)
	}
	rc.iv = ah.m
	rc.card = int64(ah.actuallyAdded)
	return rc
}

//
// newRunContainer16FromArray populates a new
// runContainer16 from the contents of arr.
//
func newRunContainer16FromArray(arr *arrayContainer) *runContainer16 {
	// keep this in sync with newRunContainer16FromVals above

	rc := &runContainer16{}
	ah := addHelper16{rc: rc}

	n := arr.getCardinality()
	var cur, prev uint16
	switch {
	case n == 0:
		// nothing more
	case n == 1:
		ah.m = append(ah.m, interval16{start: uint16(arr.content[0]), last: uint16(arr.content[0])})
		ah.actuallyAdded++
	default:
		ah.runstart = uint16(arr.content[0])
		ah.actuallyAdded++
		for i := 1; i < n; i++ {
			prev = uint16(arr.content[i-1])
			cur = uint16(arr.content[i])
			ah.add(cur, prev, i)
		}
		ah.storeIval(ah.runstart, ah.runlen)
	}
	rc.iv = ah.m
	rc.card = int64(ah.actuallyAdded)
	return rc
}

// set adds the integers in vals to the set. Vals
// must be sorted in increasing order; if not, you should set
// alreadySorted to false, and we will sort them in place for you.
// (Be aware of this side effect -- it will affect the callers
// view of vals).
//
// If you have a small number of additions to an already
// big runContainer16, calling Add() may be faster.
func (rc *runContainer16) set(alreadySorted bool, vals ...uint16) {

	rc2 := newRunContainer16FromVals(alreadySorted, vals...)
	//p("set: rc2 is %s", rc2)
	un := rc.union(rc2)
	rc.iv = un.iv
	rc.card = 0
}

// canMerge returns true iff the intervals
// a and b either overlap or they are
// contiguous and so can be merged into
// a single interval.
func canMerge16(a, b interval16) bool {
	if int64(a.last)+1 < int64(b.start) {
		return false
	}
	return int64(b.last)+1 >= int64(a.start)
}

// haveOverlap differs from canMerge in that
// it tells you if the intersection of a
// and b would contain an element (otherwise
// it would be the empty set, and we return
// false).
func haveOverlap16(a, b interval16) bool {
	if int64(a.last)+1 <= int64(b.start) {
		return false
	}
	return int64(b.last)+1 > int64(a.start)
}

// mergeInterval16s joins a and b into a
// new interval, and panics if it cannot.
func mergeInterval16s(a, b interval16) (res interval16) {
	if !canMerge16(a, b) {
		panic(fmt.Sprintf("cannot merge %#v and %#v", a, b))
	}
	if b.start < a.start {
		res.start = b.start
	} else {
		res.start = a.start
	}
	if b.last > a.last {
		res.last = b.last
	} else {
		res.last = a.last
	}
	return
}

// intersectInterval16s returns the intersection
// of a and b. The isEmpty flag will be true if
// a and b were disjoint.
func intersectInterval16s(a, b interval16) (res interval16, isEmpty bool) {
	if !haveOverlap16(a, b) {
		isEmpty = true
		return
	}
	if b.start > a.start {
		res.start = b.start
	} else {
		res.start = a.start
	}
	if b.last < a.last {
		res.last = b.last
	} else {
		res.last = a.last
	}
	return
}

// union merges two runContainer16s, producing
// a new runContainer16 with the union of rc and b.
func (rc *runContainer16) union(b *runContainer16) *runContainer16 {

	// rc is also known as 'a' here, but golint insisted we
	// call it rc for consistency with the rest of the methods.

	var m []interval16

	alim := int64(len(rc.iv))
	blim := int64(len(b.iv))

	var na int64 // next from a
	var nb int64 // next from b

	// merged holds the current merge output, which might
	// get additional merges before being appended to m.
	var merged interval16
	var mergedUsed bool // is merged being used at the moment?

	var cura interval16 // currently considering this interval16 from a
	var curb interval16 // currently considering this interval16 from b

	pass := 0
	for na < alim && nb < blim {
		pass++
		cura = rc.iv[na]
		curb = b.iv[nb]

		//p("pass=%v, cura=%v, curb=%v, merged=%v, mergedUsed=%v m=%v", pass, cura, curb, merged, mergedUsed, m)

		if mergedUsed {
			//p("mergedUsed is true")
			mergedUpdated := false
			if canMerge16(cura, merged) {
				//p("canMerge16(cura=%s, merged=%s) is true", cura, merged)
				merged = mergeInterval16s(cura, merged)
				na = rc.indexOfIntervalAtOrAfter(int64(merged.last)+1, na+1)
				mergedUpdated = true
			}
			if canMerge16(curb, merged) {
				//p("canMerge16(curb=%s, merged=%s) is true", curb, merged)
				merged = mergeInterval16s(curb, merged)
				nb = b.indexOfIntervalAtOrAfter(int64(merged.last)+1, nb+1)
				mergedUpdated = true
			}
			if !mergedUpdated {
				//p("!mergedUpdated")
				// we know that merged is disjoint from cura and curb
				m = append(m, merged)
				mergedUsed = false
			}
			continue

		} else {
			//p("!mergedUsed")
			// !mergedUsed
			if !canMerge16(cura, curb) {
				if cura.start < curb.start {
					//p("cura is before curb")
					m = append(m, cura)
					na++
				} else {
					//p("curb is before cura")
					m = append(m, curb)
					nb++
				}
			} else {
				//p("intervals are not disjoint, we can merge them. cura=%s, curb=%s", cura, curb)
				merged = mergeInterval16s(cura, curb)
				mergedUsed = true
				na = rc.indexOfIntervalAtOrAfter(int64(merged.last)+1, na+1)
				nb = b.indexOfIntervalAtOrAfter(int64(merged.last)+1, nb+1)
			}
		}
	}
	var aDone, bDone bool
	if na >= alim {
		aDone = true
		//p("na(%v) >= alim=%v, the 'a' sequence is done, finish up on 'merged' and 'b'.", na, alim)
	}
	if nb >= blim {
		bDone = true
		//p("nb(%v) >= blim=%v, the 'b' sequence is done, finish up on 'merged' and 'a'.", nb, blim)
	}
	// finish by merging anything remaining into merged we can:
	if mergedUsed {
		if !aDone {
		aAdds:
			for na < alim {
				cura = rc.iv[na]
				if canMerge16(cura, merged) {
					//p("canMerge16(cura=%s, merged=%s) is true. na=%v", cura, merged, na)
					merged = mergeInterval16s(cura, merged)
					na = rc.indexOfIntervalAtOrAfter(int64(merged.last)+1, na+1)
				} else {
					break aAdds
				}
			}

		}

		if !bDone {
		bAdds:
			for nb < blim {
				curb = b.iv[nb]
				if canMerge16(curb, merged) {
					//p("canMerge16(curb=%s, merged=%s) is true. nb=%v", curb, merged, nb)
					merged = mergeInterval16s(curb, merged)
					nb = b.indexOfIntervalAtOrAfter(int64(merged.last)+1, nb+1)
				} else {
					break bAdds
				}
			}

		}

		//p("mergedUsed==true, before adding merged=%s, m=%v", merged, sliceToString16(m))
		m = append(m, merged)
		//p("added mergedUsed, m=%v", sliceToString16(m))
	}
	if na < alim {
		//p("adding the rest of a.vi[na:] = %v", sliceToString16(rc.iv[na:]))
		m = append(m, rc.iv[na:]...)
		//p("after the rest of a.vi[na:] to m, now m = %v", sliceToString16(m))
	}
	if nb < blim {
		//p("adding the rest of b.vi[nb:] = %v", sliceToString16(b.iv[nb:]))
		m = append(m, b.iv[nb:]...)
		//p("after the rest of a.vi[nb:] to m, now m = %v", sliceToString16(m))
	}

	//p("making res out of m = %v", sliceToString16(m))
	res := &runContainer16{iv: m}
	//p("union returning %s", res)
	return res
}

// indexOfIntervalAtOrAfter is a helper for union.
func (rc *runContainer16) indexOfIntervalAtOrAfter(key int64, startIndex int64) int64 {
	rc.myOpts.startIndex = startIndex
	rc.myOpts.endxIndex = 0

	w, already, _ := rc.search(key, &rc.myOpts)
	if already {
		return int64(w)
	}
	return int64(w) + 1
}

// intersect returns a new runContainer16 holding the
// intersection of rc (also known as 'a')  and b.
func (rc *runContainer16) intersect(b *runContainer16) *runContainer16 {

	a := rc
	numa := int64(len(a.iv))
	numb := int64(len(b.iv))
	res := &runContainer16{}
	if numa == 0 || numb == 0 {
		//p("intersection is empty, returning early")
		return res
	}

	if numa == 1 && numb == 1 {
		if !haveOverlap16(a.iv[0], b.iv[0]) {
			//p("intersection is empty, returning early")
			return res
		}
	}

	var output []interval16

	var acuri int64
	var bcuri int64

	astart := int64(a.iv[acuri].start)
	bstart := int64(b.iv[bcuri].start)

	var intersection interval16
	var leftoverstart int64
	var isOverlap, isLeftoverA, isLeftoverB bool
	var done bool
	pass := 0
toploop:
	for acuri < numa && bcuri < numb {
		//p("============     top of loop, pass = %v", pass)
		pass++

		isOverlap, isLeftoverA, isLeftoverB, leftoverstart, intersection = intersectWithLeftover16(astart, int64(a.iv[acuri].last), bstart, int64(b.iv[bcuri].last))

		//p("acuri=%v, astart=%v, a.iv[acuri].endx=%v,   bcuri=%v, bstart=%v, b.iv[bcuri].endx=%v, isOverlap=%v, isLeftoverA=%v, isLeftoverB=%v, leftoverstart=%v, intersection = %#v", acuri, astart, a.iv[acuri].endx, bcuri, bstart, b.iv[bcuri].endx, isOverlap, isLeftoverA, isLeftoverB, leftoverstart, intersection)

		if !isOverlap {
			switch {
			case astart < bstart:
				//p("no overlap, astart < bstart ... acuri = %v, key=bstart= %v", acuri, bstart)
				acuri, done = a.findNextIntervalThatIntersectsStartingFrom(acuri+1, bstart)
				//p("b.findNextIntervalThatIntersectsStartingFrom(startIndex=%v, key=%v) returned: acuri = %v, done=%v", acuri+1, bstart, acuri, done)
				if done {
					break toploop
				}
				astart = int64(a.iv[acuri].start)

			case astart > bstart:
				//p("no overlap, astart > bstart ... bcuri = %v, key=astart= %v", bcuri, astart)
				bcuri, done = b.findNextIntervalThatIntersectsStartingFrom(bcuri+1, astart)
				//p("b.findNextIntervalThatIntersectsStartingFrom(startIndex=%v, key=%v) returned: bcuri = %v, done=%v", bcuri+1, astart, bcuri, done)
				if done {
					break toploop
				}
				bstart = int64(b.iv[bcuri].start)

				//default:
				//	panic("impossible that astart == bstart, since !isOverlap")
			}

		} else {
			// isOverlap
			//p("isOverlap == true, intersection = %#v", intersection)
			output = append(output, intersection)
			switch {
			case isLeftoverA:
				//p("isLeftoverA true... new astart = leftoverstart = %v", leftoverstart)
				// note that we change astart without advancing acuri,
				// since we need to capture any 2ndary intersections with a.iv[acuri]
				astart = leftoverstart
				bcuri++
				if bcuri >= numb {
					break toploop
				}
				bstart = int64(b.iv[bcuri].start)
				//p("new bstart is %v", bstart)
			case isLeftoverB:
				//p("isLeftoverB true... new bstart = leftoverstart = %v", leftoverstart)
				// note that we change bstart without advancing bcuri,
				// since we need to capture any 2ndary intersections with b.iv[bcuri]
				bstart = leftoverstart
				acuri++
				if acuri >= numa {
					break toploop
				}
				astart = int64(a.iv[acuri].start)
				//p(" ... and new astart is %v", astart)
			default:
				//p("no leftovers after intersection")
				// neither had leftover, both completely consumed
				// optionally, assert for sanity:
				//if a.iv[acuri].endx != b.iv[bcuri].endx {
				//	panic("huh? should only be possible that endx agree now!")
				//}

				// advance to next a interval
				acuri++
				if acuri >= numa {
					//p("out of 'a' elements, breaking out of loop")
					break toploop
				}
				astart = int64(a.iv[acuri].start)

				// advance to next b interval
				bcuri++
				if bcuri >= numb {
					//p("out of 'b' elements, breaking out of loop")
					break toploop
				}
				bstart = int64(b.iv[bcuri].start)
				//p("no leftovers after intersection, new acuri=%v, astart=%v, bcuri=%v, bstart=%v", acuri, astart, bcuri, bstart)
			}
		}
	} // end for toploop

	if len(output) == 0 {
		return res
	}

	res.iv = output
	//p("intersect returning %#v", res)
	return res
}

// get returns true iff key is in the container.
func (rc *runContainer16) contains(key uint16) bool {
	_, in, _ := rc.search(int64(key), nil)
	return in
}

// numIntervals returns the count of intervals in the container.
func (rc *runContainer16) numIntervals() int {
	return len(rc.iv)
}

// search returns alreadyPresent to indicate if the
// key is already in one of our interval16s.
//
// If key is alreadyPresent, then whichInterval16 tells
// you where.
//
// If key is not already present, then whichInterval16 is
// set as follows:
//
//  a) whichInterval16 == len(rc.iv)-1 if key is beyond our
//     last interval16 in rc.iv;
//
//  b) whichInterval16 == -1 if key is before our first
//     interval16 in rc.iv;
//
//  c) whichInterval16 is set to the minimum index of rc.iv
//     which comes strictly before the key;
//     so  rc.iv[whichInterval16].last < key,
//     and  if whichInterval16+1 exists, then key < rc.iv[whichInterval16+1].start
//     (Note that whichInterval16+1 won't exist when
//     whichInterval16 is the last interval.)
//
// runContainer16.search always returns whichInterval16 < len(rc.iv).
//
// If not nil, opts can be used to further restrict
// the search space.
//
func (rc *runContainer16) search(key int64, opts *searchOptions) (whichInterval16 int64, alreadyPresent bool, numCompares int) {
	n := int64(len(rc.iv))
	if n == 0 {
		return -1, false, 0
	}

	startIndex := int64(0)
	endxIndex := int64(n)
	if opts != nil {
		startIndex = opts.startIndex

		// let endxIndex == 0 mean no effect
		if opts.endxIndex > 0 {
			endxIndex = opts.endxIndex
		}
	}

	// sort.Search returns the smallest index i
	// in [0, n) at which f(i) is true, assuming that on the range [0, n),
	// f(i) == true implies f(i+1) == true.
	// If there is no such index, Search returns n.

	// For correctness, this began as verbatim snippet from
	// sort.Search in the Go standard lib.
	// We inline our comparison function for speed, and
	// annotate with numCompares
	// to observe and test that extra bounds are utilized.
	i, j := startIndex, endxIndex
	for i < j {
		h := i + (j-i)/2 // avoid overflow when computing h as the bisector
		// i <= h < j
		numCompares++
		if !(key < int64(rc.iv[h].start)) {
			i = h + 1
		} else {
			j = h
		}
	}
	below := i
	// end std lib snippet.

	// The above is a simple in-lining and annotation of:
	/*	below := sort.Search(n,
		func(i int) bool {
			return key < rc.iv[i].start
		})
	*/
	whichInterval16 = int64(below) - 1

	if below == n {
		// all falses => key is >= start of all interval16s
		// ... so does it belong to the last interval16?
		if key < int64(rc.iv[n-1].last)+1 {
			// yes, it belongs to the last interval16
			alreadyPresent = true
			return
		}
		// no, it is beyond the last interval16.
		// leave alreadyPreset = false
		return
	}

	// INVAR: key is below rc.iv[below]
	if below == 0 {
		// key is before the first first interval16.
		// leave alreadyPresent = false
		return
	}

	// INVAR: key is >= rc.iv[below-1].start and
	//        key is <  rc.iv[below].start

	// is key in below-1 interval16?
	if key >= int64(rc.iv[below-1].start) && key < int64(rc.iv[below-1].last)+1 {
		// yes, it is. key is in below-1 interval16.
		alreadyPresent = true
		return
	}

	// INVAR: key >= rc.iv[below-1].endx && key < rc.iv[below].start
	//p("search, INVAR: key >= rc.iv[below-1].endx && key < rc.iv[below].start, where key=%v, below=%v, below-1=%v, rc.iv[below-1]=%v, rc.iv[below]=%v", key, below, below-1, rc.iv[below-1], rc.iv[below])
	// leave alreadyPresent = false
	return
}

// cardinality returns the count of the integers stored in the
// runContainer16.
func (rc *runContainer16) cardinality() int64 {
	if len(rc.iv) == 0 {
		rc.card = 0
		return 0
	}
	if rc.card > 0 {
		return rc.card // already cached
	}
	// have to compute it
	var n int64
	for _, p := range rc.iv {
		n += int64(p.runlen())
	}
	rc.card = n // cache it
	return n
}

// AsSlice decompresses the contents into a []uint16 slice.
func (rc *runContainer16) AsSlice() []uint16 {
	s := make([]uint16, rc.cardinality())
	j := 0
	for _, p := range rc.iv {
		for i := p.start; i <= p.last; i++ {
			s[j] = uint16(i)
			j++
		}
	}
	return s
}

// newRunContainer16 creates an empty run container.
func newRunContainer16() *runContainer16 {
	return &runContainer16{}
}

// newRunContainer16CopyIv creates a run container, initializing
// with a copy of the supplied iv slice.
//
func newRunContainer16CopyIv(iv []interval16) *runContainer16 {
	rc := &runContainer16{
		iv: make([]interval16, len(iv)),
	}
	copy(rc.iv, iv)
	return rc
}

func (rc *runContainer16) Clone() *runContainer16 {
	rc2 := newRunContainer16CopyIv(rc.iv)
	return rc2
}

// newRunContainer16TakeOwnership returns a new runContainer16
// backed by the provided iv slice, which we will
// assume exclusive control over from now on.
//
func newRunContainer16TakeOwnership(iv []interval16) *runContainer16 {
	rc := &runContainer16{
		iv: iv,
	}
	return rc
}

const baseRc16Size = int(unsafe.Sizeof(runContainer16{}))
const perIntervalRc16Size = int(unsafe.Sizeof(interval16{}))

// serializedSizeInBytes returns the number of bytes of memory
// required by this runContainer16.
func (rc *runContainer16) serializedSizeInBytes() int {
	return rc.Msgsize()
}

// see also runContainer16SerializedSizeInBytes(numRuns int) int

// getSizeInBytes returns the number of bytes of memory
// required by this runContainer16.
func (rc *runContainer16) getSizeInBytes() int {
	return perIntervalRc16Size * len(rc.iv) // +  baseRc16Size
}

// runContainer16SerializedSizeInBytes returns the number of bytes of memory
// required to hold numRuns in a runContainer16.
func runContainer16SerializedSizeInBytes(numRuns int) int {
	return perIntervalRc16Size * numRuns // +  baseRc16Size
}

// Add adds a single value k to the set.
func (rc *runContainer16) Add(k uint16) (wasNew bool) {
	// TODO comment from runContainer16.java:
	// it might be better and simpler to do return
	// toBitmapOrArrayContainer(getCardinality()).add(k)
	// but note that some unit tests use this method to build up test
	// runcontainers without calling runOptimize

	k64 := int64(k)

	index, present, _ := rc.search(k64, nil)
	//p("search returned index=%v, present=%v", index, present)
	if present {
		return // already there
	}
	wasNew = true

	// increment card if it is cached already
	if rc.card > 0 {
		rc.card++
	}
	n := int64(len(rc.iv))
	if index == -1 {
		// we may need to extend the first run
		if n > 0 {
			if rc.iv[0].start == k+1 {
				rc.iv[0].start = k
				return
			}
		}
		// nope, k stands alone, starting the new first interval16.
		rc.iv = append([]interval16{interval16{start: k, last: k}}, rc.iv...)
		return
	}

	// are we off the end? handle both index == n and index == n-1:
	if index >= n-1 {
		if int64(rc.iv[n-1].last)+1 == k64 {
			rc.iv[n-1].last++
			return
		}
		rc.iv = append(rc.iv, interval16{start: k, last: k})
		return
	}

	// INVAR: index and index+1 both exist, and k goes between them.
	//
	// Now: add k into the middle,
	// possibly fusing with index or index+1 interval16
	// and possibly resulting in fusing of two interval16s
	// that had a one integer gap.

	left := index
	right := index + 1

	// are we fusing left and right by adding k?
	if int64(rc.iv[left].last)+1 == k64 && int64(rc.iv[right].start) == k64+1 {
		// fuse into left
		rc.iv[left].last = rc.iv[right].last
		// remove redundant right
		rc.iv = append(rc.iv[:left+1], rc.iv[right+1:]...)
		return
	}

	// are we an addition to left?
	if int64(rc.iv[left].last)+1 == k64 {
		// yes
		rc.iv[left].last++
		return
	}

	// are we an addition to right?
	if int64(rc.iv[right].start) == k64+1 {
		// yes
		rc.iv[right].start = k
		return
	}

	// k makes a standalone new interval16, inserted in the middle
	tail := append([]interval16{interval16{start: k, last: k}}, rc.iv[right:]...)
	rc.iv = append(rc.iv[:left+1], tail...)
	return
}

//msgp:ignore RunIterator

// RunIterator16 advice: you must call Next() at least once
// before calling Cur(); and you should call HasNext()
// before calling Next() to insure there are contents.
type RunIterator16 struct {
	rc            *runContainer16
	curIndex      int64
	curPosInIndex uint16
	curSeq        int64
}

// NewRunIterator16 returns a new empty run container.
func (rc *runContainer16) NewRunIterator16() *RunIterator16 {
	return &RunIterator16{rc: rc, curIndex: -1}
}

func (ri *RunIterator16) hasNext() bool {
	return ri.HasNext()
}
func (ri *RunIterator16) next() uint16 {
	return ri.Next()
}

// HasNext returns false if calling Next will panic. It
// returns true when there is at least one more value
// available in the iteration sequence.
func (ri *RunIterator16) HasNext() bool {
	if len(ri.rc.iv) == 0 {
		return false
	}
	if ri.curIndex == -1 {
		return true
	}
	return ri.curSeq+1 < ri.rc.cardinality()
}

// Cur returns the current value pointed to by the iterator.
func (ri *RunIterator16) Cur() uint16 {
	//p("in Cur, curIndex=%v, curPosInIndex=%v", ri.curIndex, ri.curPosInIndex)
	return ri.rc.iv[ri.curIndex].start + ri.curPosInIndex
}

// Next returns the next value in the iteration sequence.
func (ri *RunIterator16) Next() uint16 {
	if !ri.HasNext() {
		panic("no Next available")
	}
	if ri.curIndex >= int64(len(ri.rc.iv)) {
		panic("RunIterator.Next() going beyond what is available")
	}
	if ri.curIndex == -1 {
		// first time is special
		ri.curIndex = 0
	} else {
		ri.curPosInIndex++
		if int64(ri.rc.iv[ri.curIndex].start)+int64(ri.curPosInIndex) == int64(ri.rc.iv[ri.curIndex].last)+1 {
			//p("rolling from ri.curIndex==%v to ri.curIndex=%v", ri.curIndex, ri.curIndex+1)
			ri.curPosInIndex = 0
			ri.curIndex++
		} else {
			//p("no roll ... ri.curPosInIndex is now %v, ri.rc.iv[ri.curIndex].endx=%v", ri.curPosInIndex, ri.rc.iv[ri.curIndex].endx)
		}
		ri.curSeq++
	}
	return ri.Cur()
}

// Remove removes the element that the iterator
// is on from the run container. You can use
// Cur if you want to double check what is about
// to be deleted.
func (ri *RunIterator16) Remove() uint16 {
	n := ri.rc.cardinality()
	if n == 0 {
		panic("RunIterator.Remove called on empty runContainer16")
	}
	cur := ri.Cur()

	ri.rc.deleteAt(&ri.curIndex, &ri.curPosInIndex, &ri.curSeq)
	return cur
}

// remove removes key from the container.
func (rc *runContainer16) removeKey(key uint16) (wasPresent bool) {

	var index int64
	var curSeq int64
	index, wasPresent, _ = rc.search(int64(key), nil)
	if !wasPresent {
		return // already removed, nothing to do.
	}
	pos := key - rc.iv[index].start
	rc.deleteAt(&index, &pos, &curSeq)
	return
}

// internal helper functions

func (rc *runContainer16) deleteAt(curIndex *int64, curPosInIndex *uint16, curSeq *int64) {
	rc.card--
	(*curSeq)--
	ci := *curIndex
	pos := *curPosInIndex

	// are we first, last, or in the middle of our interval16?
	switch {
	case pos == 0:
		//p("pos == 0, first")
		if int64(rc.iv[ci].start) == int64(rc.iv[ci].last) {
			// our interval disappears
			rc.iv = append(rc.iv[:ci], rc.iv[ci+1:]...)
			// curIndex stays the same, since the delete did
			// the advance for us.
			*curPosInIndex = 0
		} else {
			rc.iv[ci].start++ // no longer overflowable
		}
	case int64(pos) == rc.iv[ci].runlen()-1:
		// last
		rc.iv[ci].last--
		// our interval16 cannot disappear, else we would have been pos == 0, case first above.
		//p("deleteAt: pos is last case, curIndex=%v, curPosInIndex=%v", *curIndex, *curPosInIndex)
		(*curPosInIndex)--
		// if we leave *curIndex alone, then Next() will work properly even after the delete.
		//p("deleteAt: pos is last case, after update: curIndex=%v, curPosInIndex=%v", *curIndex, *curPosInIndex)
	default:
		//p("middle...split")
		//middle
		// split into two, adding an interval16
		new0 := interval16{
			start: rc.iv[ci].start,
			last:  rc.iv[ci].start + *curPosInIndex - 1}

		new1start := int64(rc.iv[ci].start) + int64(*curPosInIndex) + 1
		if new1start > int64(MaxUint16) {
			panic("overflow?!?!")
		}
		new1 := interval16{
			start: uint16(new1start),
			last:  rc.iv[ci].last}

		//p("new0 = %#v", new0)
		//p("new1 = %#v", new1)

		tail := append([]interval16{new0, new1}, rc.iv[ci+1:]...)
		rc.iv = append(rc.iv[:ci], tail...)
		// update curIndex and curPosInIndex
		(*curIndex)++
		*curPosInIndex = 0
	}

}

func have4Overlap16(astart, alast, bstart, blast int64) bool {
	if int64(alast)+1 <= bstart {
		return false
	}
	return int64(blast)+1 > astart
}

func intersectWithLeftover16(astart, alast, bstart, blast int64) (isOverlap, isLeftoverA, isLeftoverB bool, leftoverstart int64, intersection interval16) {
	if !have4Overlap16(astart, alast, bstart, blast) {
		return
	}
	isOverlap = true

	// do the intersection:
	if bstart > astart {
		intersection.start = uint16(bstart)
	} else {
		intersection.start = uint16(astart)
	}
	switch {
	case blast < alast:
		isLeftoverA = true
		leftoverstart = int64(blast) + 1
		intersection.last = uint16(blast)
	case alast < blast:
		isLeftoverB = true
		leftoverstart = int64(alast) + 1
		intersection.last = uint16(alast)
	default:
		// alast == blast
		intersection.last = uint16(alast)
	}

	return
}

func (rc *runContainer16) findNextIntervalThatIntersectsStartingFrom(startIndex int64, key int64) (index int64, done bool) {

	rc.myOpts.startIndex = startIndex
	rc.myOpts.endxIndex = 0

	w, _, _ := rc.search(key, &rc.myOpts)
	// rc.search always returns w < len(rc.iv)
	if w < startIndex {
		// not found and comes before lower bound startIndex,
		// so just use the lower bound.
		if startIndex == int64(len(rc.iv)) {
			// also this bump up means that we are done
			return startIndex, true
		}
		return startIndex, false
	}

	return w, false
}

func sliceToString16(m []interval16) string {
	s := ""
	for i := range m {
		s += fmt.Sprintf("%v: %s, ", i, m[i])
	}
	return s
}

// selectInt16 returns the j-th value in the container.
// We panic of j is out of bounds.
func (rc *runContainer16) selectInt16(j uint16) int {
	n := rc.cardinality()
	if int64(j) > n {
		panic(fmt.Sprintf("Cannot select %v since Cardinality is %v", j, n))
	}

	var offset int64
	for k := range rc.iv {
		nextOffset := offset + rc.iv[k].runlen() + 1
		if nextOffset > int64(j) {
			return int(int64(rc.iv[k].start) + (int64(j) - offset))
		}
		offset = nextOffset
	}
	panic(fmt.Sprintf("Cannot select %v since Cardinality is %v", j, n))
}

// helper for invert
func (rc *runContainer16) invertlastInterval(origin uint16, lastIdx int) []interval16 {
	cur := rc.iv[lastIdx]
	if cur.last == MaxUint16 {
		if cur.start == origin {
			return nil // empty container
		}
		return []interval16{{start: origin, last: cur.start - 1}}
	}
	if cur.start == origin {
		return []interval16{{start: cur.last + 1, last: MaxUint16}}
	}
	// invert splits
	return []interval16{
		{start: origin, last: cur.start - 1},
		{start: cur.last + 1, last: MaxUint16},
	}
}

// invert returns a new container (not inplace), that is
// the inversion of rc. For each bit b in rc, the
// returned value has !b
func (rc *runContainer16) invert() *runContainer16 {
	ni := len(rc.iv)
	var m []interval16
	switch ni {
	case 0:
		return &runContainer16{iv: []interval16{{0, MaxUint16}}}
	case 1:
		return &runContainer16{iv: rc.invertlastInterval(0, 0)}
	}
	var invstart int64
	ult := ni - 1
	for i, cur := range rc.iv {
		if i == ult {
			// invertlastInteval will add both intervals (b) and (c) in
			// diagram below.
			m = append(m, rc.invertlastInterval(uint16(invstart), i)...)
			break
		}
		// INVAR: i and cur are not the last interval, there is a next at i+1
		//
		// ........[cur.start, cur.last] ...... [next.start, next.last]....
		//    ^                             ^                           ^
		//   (a)                           (b)                         (c)
		//
		// Now: we add interval (a); but if (a) is empty, for cur.start==0, we skip it.
		if cur.start > 0 {
			m = append(m, interval16{start: uint16(invstart), last: cur.start - 1})
		}
		invstart = int64(cur.last + 1)
	}
	return &runContainer16{iv: m}
}

func (a interval16) equal(b interval16) bool {
	if a.start == b.start {
		return a.last == b.last
	}
	return false
}

func (a interval16) isSuperSetOf(b interval16) bool {
	return a.start <= b.start && b.last <= a.last
}

func (cur interval16) subtractInterval(del interval16) (left []interval16, delcount int64) {
	defer func() {
		//p("returning from subtractInterval of cur - del with cur=%s and del=%s, returning left=%s, delcount=%v", cur, del, ivalString16(left), delcount)
	}()
	isect, isEmpty := intersectInterval16s(cur, del)

	if isEmpty {
		//p("isEmpty")
		return nil, 0
	}
	if del.isSuperSetOf(cur) {
		return nil, cur.runlen()
	}

	switch {
	case isect.start > cur.start && isect.last < cur.last:
		//p("split into two")
		new0 := interval16{start: cur.start, last: isect.start - 1}
		new1 := interval16{start: isect.last + 1, last: cur.last}
		return []interval16{new0, new1}, isect.runlen()
	case isect.start == cur.start:
		//p("removal of only the first half or so of cur interval. isect: %s, cur: %s, del: %s", isect, cur, del)
		return []interval16{{start: isect.last + 1, last: cur.last}}, isect.runlen()
	default:
		//p("isect.end == cur.end")
		//p("removal of only the last half or so of cur interval")
		return []interval16{{start: cur.start, last: isect.start - 1}}, isect.runlen()
	}
}

func (rc *runContainer16) isubtract(del interval16) {
	origiv := make([]interval16, len(rc.iv))
	copy(origiv, rc.iv)
	//p("isubtract starting, with del = %s, and rc = %s", del, rc)
	n := int64(len(rc.iv))
	if n == 0 {
		return // already done.
	}

	_, isEmpty := intersectInterval16s(
		interval16{
			start: rc.iv[0].start,
			last:  rc.iv[n-1].last,
		}, del)
	if isEmpty {
		//p("del=%v -> isEmpty, returning early from isubtract", del)
		return // done
	}
	// INVAR there is some intersection between rc and del
	istart, startAlready, _ := rc.search(int64(del.start), nil)
	ilast, lastAlready, _ := rc.search(int64(del.last), nil)
	rc.card = -1

	//p("del=%v, istart = %v, startAlready = %v", del, istart, startAlready)
	//p("del=%v, ilast = %v, lastAlready = %v", del, ilast, lastAlready)

	if istart == -1 {
		if ilast == n-1 {
			//p("discard it all")
			rc.iv = nil
			return
		}
	}
	// some intervals will remain
	//p("orig rc.iv = '%s'", ivalString16(rc.iv))
	switch {
	case startAlready && lastAlready:
		//p("case 1: startAlready && lastAlready; istart=%v, ilast=%v. staring rc.iv='%s'", istart, ilast, ivalString16(rc.iv))
		res0, _ := rc.iv[istart].subtractInterval(del)

		//p("case 1 rc.iv[:start] = '%s', while res0='%s'", ivalString16(rc.iv[:istart]), ivalString16(res0))
		// would overwrite values in iv b/c res0 can have len 2. so
		// write to origiv instead.

		//p("case 1 pre = '%s'", ivalString16(pre))
		//p("orig rc.iv = '%s'", ivalString16(rc.iv))

		lost := 1 + ilast - istart
		changeSize := int64(len(res0)) - lost
		newSize := int64(len(rc.iv)) + changeSize

		//p("case 1 before suffixing with: rc.iv[ilast+1:] = '%s'", ivalString16(rc.iv[ilast+1:]))
		//	rc.iv = append(pre, caboose...)
		//	return

		if ilast != istart {
			res1, _ := rc.iv[ilast].subtractInterval(del)
			res0 = append(res0, res1...)
			changeSize = int64(len(res0)) - lost
			newSize = int64(len(rc.iv)) + changeSize
		}
		switch {
		case changeSize < 0:
			// shrink
			copy(rc.iv[istart+int64(len(res0)):], rc.iv[ilast+1:])
			copy(rc.iv[istart:istart+int64(len(res0))], res0)
			rc.iv = rc.iv[:newSize]
			return
		case changeSize == 0:
			// stay the same
			copy(rc.iv[istart:istart+int64(len(res0))], res0)
			return
		default:
			// changeSize > 0 is only possible when ilast == istart.
			// Hence we now know: changeSize == 1 and len(res0) == 2
			rc.iv = append(rc.iv, interval16{})
			// len(rc.iv) is correct now, no need to rc.iv = rc.iv[:newSize]

			// copy the tail into place
			copy(rc.iv[ilast+2:], rc.iv[ilast+1:])
			// copy the new item(s) into place
			copy(rc.iv[istart:istart+2], res0)
			return
		}

	case !startAlready && !lastAlready:
		//p("case 2: !startAlready && !lastAlready")
		// we get to discard whole intervals

		// from the search() definition:

		// if del.start is not present, then istart is
		// set as follows:
		//
		//  a) istart == n-1 if del.start is beyond our
		//     last interval16 in rc.iv;
		//
		//  b) istart == -1 if del.start is before our first
		//     interval16 in rc.iv;
		//
		//  c) istart is set to the minimum index of rc.iv
		//     which comes strictly before the del.start;
		//     so  del.start > rc.iv[istart].last,
		//     and  if istart+1 exists, then del.start < rc.iv[istart+1].startx

		// if del.last is not present, then ilast is
		// set as follows:
		//
		//  a) ilast == n-1 if del.last is beyond our
		//     last interval16 in rc.iv;
		//
		//  b) ilast == -1 if del.last is before our first
		//     interval16 in rc.iv;
		//
		//  c) ilast is set to the minimum index of rc.iv
		//     which comes strictly before the del.last;
		//     so  del.last > rc.iv[ilast].last,
		//     and  if ilast+1 exists, then del.last < rc.iv[ilast+1].start

		// INVAR: istart >= 0
		pre := rc.iv[:istart+1]
		if ilast == n-1 {
			rc.iv = pre
			return
		}
		// INVAR: ilast < n-1
		lost := ilast - istart
		changeSize := -lost
		newSize := int64(len(rc.iv)) + changeSize
		if changeSize != 0 {
			copy(rc.iv[ilast+1+changeSize:], rc.iv[ilast+1:])
		}
		rc.iv = rc.iv[:newSize]
		return

	case startAlready && !lastAlready:
		// we can only shrink or stay the same size
		// i.e. we either eliminate the whole interval,
		// or just cut off the right side.
		//p("case 3: startAlready && !lastAlready, rc='%s', del='%s'", rc, del)
		res0, _ := rc.iv[istart].subtractInterval(del)
		if len(res0) > 0 {
			// len(res) must be 1
			rc.iv[istart] = res0[0]
		}
		lost := 1 + (ilast - istart)
		changeSize := int64(len(res0)) - lost
		newSize := int64(len(rc.iv)) + changeSize
		if changeSize != 0 {
			copy(rc.iv[ilast+1+changeSize:], rc.iv[ilast+1:])
		}
		rc.iv = rc.iv[:newSize]
		return

	case !startAlready && lastAlready:
		// we can only shrink or stay the same size
		//p("case 4: !startAlready && lastAlready")
		res1, _ := rc.iv[ilast].subtractInterval(del)
		lost := ilast - istart
		changeSize := int64(len(res1)) - lost
		newSize := int64(len(rc.iv)) + changeSize
		if changeSize != 0 {
			// move the tail first to make room for res1
			copy(rc.iv[ilast+1+changeSize:], rc.iv[ilast+1:])
		}
		copy(rc.iv[istart+1:], res1)
		rc.iv = rc.iv[:newSize]
		return
	}
}

// compute rc minus b, and return the result as a new value (not inplace).
// port of run_container_andnot from CRoaring...
// https://github.com/RoaringBitmap/CRoaring/blob/master/src/containers/run.c#L435-L496
func (rc *runContainer16) AndNotRunContainer16(b *runContainer16) *runContainer16 {

	if len(b.iv) == 0 || len(rc.iv) == 0 {
		return rc
	}

	dst := newRunContainer16()
	apos := 0
	bpos := 0

	a := rc

	astart := a.iv[apos].start
	alast := a.iv[apos].last
	bstart := b.iv[bpos].start
	blast := b.iv[bpos].last

	alen := len(a.iv)
	blen := len(b.iv)

	for apos < alen && bpos < blen {
		//p("top: apos = %v, alen=%v, bpos=%v, blen=%v", apos, alen, bpos, blen)
		switch {
		case alast < bstart:
			// output the first run
			dst.iv = append(dst.iv, interval16{start: uint16(astart), last: uint16(alast)})
			//p("alast(%v) < bstart(%v), dst after adding [astart, last] is: %s", alast, bstart, dst)
			apos++
			if apos < alen {
				astart = a.iv[apos].start
				alast = a.iv[apos].last
			}
		case blast < astart:
			// exit the second run
			bpos++
			if bpos < blen {
				bstart = b.iv[bpos].start
				blast = b.iv[bpos].last
			}
		default:
			//   a: [             ]
			//   b:            [    ]
			// alast >= bstart
			// blast >= astart
			if astart < bstart {
				dst.iv = append(dst.iv, interval16{start: uint16(astart), last: uint16(bstart - 1)})
				//p("astart(%v) < bstart(%v), dst after adding [astart, bstart] is: %s", astart, bstart, dst)
			}
			if alast > blast {
				astart = blast + 1
			} else {
				apos++
				if apos < alen {
					astart = a.iv[apos].start
					alast = a.iv[apos].last
				}
			}
		}
	}
	if apos < alen {
		dst.iv = append(dst.iv, interval16{start: uint16(astart), last: uint16(alast)})
		apos++
		if apos < alen {
			dst.iv = append(dst.iv, a.iv[apos:]...)
		}
	}

	return dst
}

func (rc *runContainer16) numberOfRuns() (nr int) {
	return len(rc.iv)
}

func (bc *runContainer16) containerType() contype {
	return run16Contype
}
