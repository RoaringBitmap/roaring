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
	"github.com/glycerine/roaring/serz"
	"sort"
	"unsafe"
)

//go:generate msgp -unexported

// runContainer32 does run-length encoding of sets of
// uint32 integers.
type runContainer32 struct {
	Iv   []interval32
	Card int64

	// avoid allocation during search
	myOpts searchOptions `msg:"-"`
}

// interval32 is the internal to runContainer32
// structure that maintains the individual [Start, Last]
// closed intervals.
type interval32 struct {
	Start uint32
	Last  uint32
}

// runlen returns the count of integers in the interval.
func (iv interval32) runlen() int64 {
	return 1 + int64(iv.Last) - int64(iv.Start)
}

func newRunContainer32FromSerz(sz *serz.RunContainer32) *runContainer32 {
	rc := &runContainer32{Card: sz.Card}
	for i := range sz.Iv {
		rc.Iv = append(rc.Iv, interval32{Start: sz.Iv[i].Start, Last: sz.Iv[i].Last})
	}
	return rc
}

func (rc *runContainer32) toSerz() *serz.RunContainer32 {
	sz := &serz.RunContainer32{Card: rc.Card}
	for i := range rc.Iv {
		sz.Iv = append(sz.Iv, serz.Interval32{Start: rc.Iv[i].Start, Last: rc.Iv[i].Last})
	}
	return sz

}

// String produces a human viewable string of the contents.
func (iv interval32) String() string {
	return fmt.Sprintf("[%d, %d]", iv.Start, iv.Last)
}

func ivalString32(iv []interval32) string {
	var s string
	var j int
	var p interval32
	for j, p = range iv {
		s += fmt.Sprintf("%v:[%d, %d], ", j, p.Start, p.Last)
	}
	return s
}

// String produces a human viewable string of the contents.
func (rc *runContainer32) String() string {
	if len(rc.Iv) == 0 {
		return "runContainer32{}"
	}
	is := ivalString32(rc.Iv)
	return `runContainer32{` + is + `}`
}

// uint32Slice is a sort.Sort convenience method
type uint32Slice []uint32

// Len returns the length of p.
func (p uint32Slice) Len() int { return len(p) }

// Less returns p[i] < p[j]
func (p uint32Slice) Less(i, j int) bool { return p[i] < p[j] }

// Swap swaps elements i and j.
func (p uint32Slice) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

//msgp:ignore addHelper

// addHelper helps build a runContainer32.
type addHelper32 struct {
	runstart      uint32
	runlen        uint32
	actuallyAdded uint32
	m             []interval32
	rc            *runContainer32
}

func (ah *addHelper32) storeIval(runstart, runlen uint32) {
	mi := interval32{Start: runstart, Last: runstart + runlen}
	ah.m = append(ah.m, mi)
}

func (ah *addHelper32) add(cur, prev uint32, i int) {
	if cur == prev+1 {
		ah.runlen++
		ah.actuallyAdded++
	} else {
		if cur < prev {
			panic(fmt.Sprintf("newRunContainer32FromVals sees "+
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
func newRunContainer32Range(rangestart uint32, rangelast uint32) *runContainer32 {
	rc := &runContainer32{}
	rc.Iv = append(rc.Iv, interval32{Start: rangestart, Last: rangelast})
	return rc
}

// newRunContainer32FromVals makes a new container from vals.
//
// For efficiency, vals should be sorted in ascending order.
// Ideally vals should not contain duplicates, but we detect and
// ignore them. If vals is already sorted in ascending order, then
// pass alreadySorted = true. Otherwise, for !alreadySorted,
// we will sort vals before creating a runContainer32 of them.
// We sort the original vals, so this will change what the
// caller sees in vals as a side effect.
func newRunContainer32FromVals(alreadySorted bool, vals ...uint32) *runContainer32 {
	// keep this in sync with newRunContainer32FromArray below

	rc := &runContainer32{}
	ah := addHelper32{rc: rc}

	if !alreadySorted {
		sort.Sort(uint32Slice(vals))
	}
	n := len(vals)
	var cur, prev uint32
	switch {
	case n == 0:
		// nothing more
	case n == 1:
		ah.m = append(ah.m, interval32{Start: vals[0], Last: vals[0]})
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
	rc.Iv = ah.m
	rc.Card = int64(ah.actuallyAdded)
	return rc
}

// newRunContainer32FromBitmapContainer makes a new run container from bc.
func newRunContainer32FromBitmapContainer(bc *bitmapContainer) *runContainer32 {
	// todo: this could be optimized, see https://github.com/RoaringBitmap/RoaringBitmap/blob/master/src/main/java/org/roaringbitmap/RunContainer.java#L145-L192

	rc := &runContainer32{}
	ah := addHelper32{rc: rc}

	n := bc.getCardinality()
	it := bc.getShortIterator()
	var cur, prev, val uint32
	switch {
	case n == 0:
		// nothing more
	case n == 1:
		val = uint32(it.next())
		ah.m = append(ah.m, interval32{Start: val, Last: val})
		ah.actuallyAdded++
	default:
		prev = uint32(it.next())
		cur = uint32(it.next())
		ah.runstart = prev
		ah.actuallyAdded++
		for i := 1; i < n; i++ {
			ah.add(cur, prev, i)
			if it.hasNext() {
				prev = cur
				cur = uint32(it.next())
			}
		}
		ah.storeIval(ah.runstart, ah.runlen)
	}
	rc.Iv = ah.m
	rc.Card = int64(ah.actuallyAdded)
	return rc
}

//
// newRunContainer32FromArray populates a new
// runContainer32 from the contents of arr.
//
func newRunContainer32FromArray(arr *arrayContainer) *runContainer32 {
	// keep this in sync with newRunContainer32FromVals above

	rc := &runContainer32{}
	ah := addHelper32{rc: rc}

	n := arr.getCardinality()
	var cur, prev uint32
	switch {
	case n == 0:
		// nothing more
	case n == 1:
		ah.m = append(ah.m, interval32{Start: uint32(arr.Content[0]), Last: uint32(arr.Content[0])})
		ah.actuallyAdded++
	default:
		ah.runstart = uint32(arr.Content[0])
		ah.actuallyAdded++
		for i := 1; i < n; i++ {
			prev = uint32(arr.Content[i-1])
			cur = uint32(arr.Content[i])
			ah.add(cur, prev, i)
		}
		ah.storeIval(ah.runstart, ah.runlen)
	}
	rc.Iv = ah.m
	rc.Card = int64(ah.actuallyAdded)
	return rc
}

// set adds the integers in vals to the set. Vals
// must be sorted in increasing order; if not, you should set
// alreadySorted to false, and we will sort them in place for you.
// (Be aware of this side effect -- it will affect the callers
// view of vals).
//
// If you have a small number of additions to an already
// big runContainer32, calling Add() may be faster.
func (rc *runContainer32) set(alreadySorted bool, vals ...uint32) {

	rc2 := newRunContainer32FromVals(alreadySorted, vals...)
	//p("set: rc2 is %s", rc2)
	un := rc.union(rc2)
	rc.Iv = un.Iv
	rc.Card = 0
}

// canMerge returns true iff the intervals
// a and b either overlap or they are
// contiguous and so can be merged into
// a single interval.
func canMerge32(a, b interval32) bool {
	if int64(a.Last)+1 < int64(b.Start) {
		return false
	}
	return int64(b.Last)+1 >= int64(a.Start)
}

// haveOverlap differs from canMerge in that
// it tells you if the intersection of a
// and b would contain an element (otherwise
// it would be the empty set, and we return
// false).
func haveOverlap32(a, b interval32) bool {
	if int64(a.Last)+1 <= int64(b.Start) {
		return false
	}
	return int64(b.Last)+1 > int64(a.Start)
}

// mergeInterval32s joins a and b into a
// new interval, and panics if it cannot.
func mergeInterval32s(a, b interval32) (res interval32) {
	if !canMerge32(a, b) {
		panic(fmt.Sprintf("cannot merge %#v and %#v", a, b))
	}
	if b.Start < a.Start {
		res.Start = b.Start
	} else {
		res.Start = a.Start
	}
	if b.Last > a.Last {
		res.Last = b.Last
	} else {
		res.Last = a.Last
	}
	return
}

// intersectInterval32s returns the intersection
// of a and b. The isEmpty flag will be true if
// a and b were disjoint.
func intersectInterval32s(a, b interval32) (res interval32, isEmpty bool) {
	if !haveOverlap32(a, b) {
		isEmpty = true
		return
	}
	if b.Start > a.Start {
		res.Start = b.Start
	} else {
		res.Start = a.Start
	}
	if b.Last < a.Last {
		res.Last = b.Last
	} else {
		res.Last = a.Last
	}
	return
}

// union merges two runContainer32s, producing
// a new runContainer32 with the union of rc and b.
func (rc *runContainer32) union(b *runContainer32) *runContainer32 {

	// rc is also known as 'a' here, but golint insisted we
	// call it rc for consistency with the rest of the methods.

	var m []interval32

	alim := int64(len(rc.Iv))
	blim := int64(len(b.Iv))

	var na int64 // next from a
	var nb int64 // next from b

	// merged holds the current merge output, which might
	// get additional merges before being appended to m.
	var merged interval32
	var mergedUsed bool // is merged being used at the moment?

	var cura interval32 // currently considering this interval32 from a
	var curb interval32 // currently considering this interval32 from b

	pass := 0
	for na < alim && nb < blim {
		pass++
		cura = rc.Iv[na]
		curb = b.Iv[nb]

		//p("pass=%v, cura=%v, curb=%v, merged=%v, mergedUsed=%v m=%v", pass, cura, curb, merged, mergedUsed, m)

		if mergedUsed {
			//p("mergedUsed is true")
			mergedUpdated := false
			if canMerge32(cura, merged) {
				//p("canMerge32(cura=%s, merged=%s) is true", cura, merged)
				merged = mergeInterval32s(cura, merged)
				na = rc.indexOfIntervalAtOrAfter(int64(merged.Last)+1, na+1)
				mergedUpdated = true
			}
			if canMerge32(curb, merged) {
				//p("canMerge32(curb=%s, merged=%s) is true", curb, merged)
				merged = mergeInterval32s(curb, merged)
				nb = b.indexOfIntervalAtOrAfter(int64(merged.Last)+1, nb+1)
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
			if !canMerge32(cura, curb) {
				if cura.Start < curb.Start {
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
				merged = mergeInterval32s(cura, curb)
				mergedUsed = true
				na = rc.indexOfIntervalAtOrAfter(int64(merged.Last)+1, na+1)
				nb = b.indexOfIntervalAtOrAfter(int64(merged.Last)+1, nb+1)
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
				cura = rc.Iv[na]
				if canMerge32(cura, merged) {
					//p("canMerge32(cura=%s, merged=%s) is true. na=%v", cura, merged, na)
					merged = mergeInterval32s(cura, merged)
					na = rc.indexOfIntervalAtOrAfter(int64(merged.Last)+1, na+1)
				} else {
					break aAdds
				}
			}

		}

		if !bDone {
		bAdds:
			for nb < blim {
				curb = b.Iv[nb]
				if canMerge32(curb, merged) {
					//p("canMerge32(curb=%s, merged=%s) is true. nb=%v", curb, merged, nb)
					merged = mergeInterval32s(curb, merged)
					nb = b.indexOfIntervalAtOrAfter(int64(merged.Last)+1, nb+1)
				} else {
					break bAdds
				}
			}

		}

		//p("mergedUsed==true, before adding merged=%s, m=%v", merged, sliceToString32(m))
		m = append(m, merged)
		//p("added mergedUsed, m=%v", sliceToString32(m))
	}
	if na < alim {
		//p("adding the rest of a.vi[na:] = %v", sliceToString32(rc.Iv[na:]))
		m = append(m, rc.Iv[na:]...)
		//p("after the rest of a.vi[na:] to m, now m = %v", sliceToString32(m))
	}
	if nb < blim {
		//p("adding the rest of b.vi[nb:] = %v", sliceToString32(b.Iv[nb:]))
		m = append(m, b.Iv[nb:]...)
		//p("after the rest of a.vi[nb:] to m, now m = %v", sliceToString32(m))
	}

	//p("making res out of m = %v", sliceToString32(m))
	res := &runContainer32{Iv: m}
	//p("union returning %s", res)
	return res
}

// indexOfIntervalAtOrAfter is a helper for union.
func (rc *runContainer32) indexOfIntervalAtOrAfter(key int64, startIndex int64) int64 {
	rc.myOpts.StartIndex = startIndex
	rc.myOpts.EndxIndex = 0

	w, already, _ := rc.search(key, &rc.myOpts)
	if already {
		return int64(w)
	}
	return int64(w) + 1
}

// intersect returns a new runContainer32 holding the
// intersection of rc (also known as 'a')  and b.
func (rc *runContainer32) intersect(b *runContainer32) *runContainer32 {

	a := rc
	numa := int64(len(a.Iv))
	numb := int64(len(b.Iv))
	res := &runContainer32{}
	if numa == 0 || numb == 0 {
		//p("intersection is empty, returning early")
		return res
	}

	if numa == 1 && numb == 1 {
		if !haveOverlap32(a.Iv[0], b.Iv[0]) {
			//p("intersection is empty, returning early")
			return res
		}
	}

	var output []interval32

	var acuri int64
	var bcuri int64

	astart := int64(a.Iv[acuri].Start)
	bstart := int64(b.Iv[bcuri].Start)

	var intersection interval32
	var leftoverStart int64
	var isOverlap, isLeftoverA, isLeftoverB bool
	var done bool
	pass := 0
toploop:
	for acuri < numa && bcuri < numb {
		//p("============     top of loop, pass = %v", pass)
		pass++

		isOverlap, isLeftoverA, isLeftoverB, leftoverStart, intersection = intersectWithLeftover32(astart, int64(a.Iv[acuri].Last), bstart, int64(b.Iv[bcuri].Last))

		//p("acuri=%v, astart=%v, a.Iv[acuri].endx=%v,   bcuri=%v, bstart=%v, b.Iv[bcuri].endx=%v, isOverlap=%v, isLeftoverA=%v, isLeftoverB=%v, leftoverStart=%v, intersection = %#v", acuri, astart, a.Iv[acuri].endx, bcuri, bstart, b.Iv[bcuri].endx, isOverlap, isLeftoverA, isLeftoverB, leftoverStart, intersection)

		if !isOverlap {
			switch {
			case astart < bstart:
				//p("no overlap, astart < bstart ... acuri = %v, key=bstart= %v", acuri, bstart)
				acuri, done = a.findNextIntervalThatIntersectsStartingFrom(acuri+1, bstart)
				//p("b.findNextIntervalThatIntersectsStartingFrom(startIndex=%v, key=%v) returned: acuri = %v, done=%v", acuri+1, bstart, acuri, done)
				if done {
					break toploop
				}
				astart = int64(a.Iv[acuri].Start)

			case astart > bstart:
				//p("no overlap, astart > bstart ... bcuri = %v, key=astart= %v", bcuri, astart)
				bcuri, done = b.findNextIntervalThatIntersectsStartingFrom(bcuri+1, astart)
				//p("b.findNextIntervalThatIntersectsStartingFrom(startIndex=%v, key=%v) returned: bcuri = %v, done=%v", bcuri+1, astart, bcuri, done)
				if done {
					break toploop
				}
				bstart = int64(b.Iv[bcuri].Start)

				//default:
				//	panic("impossible that astart == bstart, since !isOverlap")
			}

		} else {
			// isOverlap
			//p("isOverlap == true, intersection = %#v", intersection)
			output = append(output, intersection)
			switch {
			case isLeftoverA:
				//p("isLeftoverA true... new astart = leftoverStart = %v", leftoverStart)
				// note that we change astart without advancing acuri,
				// since we need to capture any 2ndary intersections with a.Iv[acuri]
				astart = leftoverStart
				bcuri++
				if bcuri >= numb {
					break toploop
				}
				bstart = int64(b.Iv[bcuri].Start)
				//p("new bstart is %v", bstart)
			case isLeftoverB:
				//p("isLeftoverB true... new bstart = leftoverStart = %v", leftoverStart)
				// note that we change bstart without advancing bcuri,
				// since we need to capture any 2ndary intersections with b.Iv[bcuri]
				bstart = leftoverStart
				acuri++
				if acuri >= numa {
					break toploop
				}
				astart = int64(a.Iv[acuri].Start)
				//p(" ... and new astart is %v", astart)
			default:
				//p("no leftovers after intersection")
				// neither had leftover, both completely consumed
				// optionally, assert for sanity:
				//if a.Iv[acuri].endx != b.Iv[bcuri].endx {
				//	panic("huh? should only be possible that endx agree now!")
				//}

				// advance to next a interval
				acuri++
				if acuri >= numa {
					//p("out of 'a' elements, breaking out of loop")
					break toploop
				}
				astart = int64(a.Iv[acuri].Start)

				// advance to next b interval
				bcuri++
				if bcuri >= numb {
					//p("out of 'b' elements, breaking out of loop")
					break toploop
				}
				bstart = int64(b.Iv[bcuri].Start)
				//p("no leftovers after intersection, new acuri=%v, astart=%v, bcuri=%v, bstart=%v", acuri, astart, bcuri, bstart)
			}
		}
	} // end for toploop

	if len(output) == 0 {
		return res
	}

	res.Iv = output
	//p("intersect returning %#v", res)
	return res
}

// get returns true iff key is in the container.
func (rc *runContainer32) contains(key uint32) bool {
	_, in, _ := rc.search(int64(key), nil)
	return in
}

// numIntervals returns the count of intervals in the container.
func (rc *runContainer32) numIntervals() int {
	return len(rc.Iv)
}

// search returns alreadyPresent to indicate if the
// key is already in one of our interval32s.
//
// If key is alreadyPresent, then whichInterval32 tells
// you where.
//
// If key is not already present, then whichInterval32 is
// set as follows:
//
//  a) whichInterval32 == len(rc.Iv)-1 if key is beyond our
//     last interval32 in rc.Iv;
//
//  b) whichInterval32 == -1 if key is before our first
//     interval32 in rc.Iv;
//
//  c) whichInterval32 is set to the minimum index of rc.Iv
//     which comes strictly before the key;
//     so  rc.Iv[whichInterval32].Last < key,
//     and  if whichInterval32+1 exists, then key < rc.Iv[whichInterval32+1].Start
//     (Note that whichInterval32+1 won't exist when
//     whichInterval32 is the last interval.)
//
// runContainer32.search always returns whichInterval32 < len(rc.Iv).
//
// If not nil, opts can be used to further restrict
// the search space.
//
func (rc *runContainer32) search(key int64, opts *searchOptions) (whichInterval32 int64, alreadyPresent bool, numCompares int) {
	n := int64(len(rc.Iv))
	if n == 0 {
		return -1, false, 0
	}

	startIndex := int64(0)
	endxIndex := int64(n)
	if opts != nil {
		startIndex = opts.StartIndex

		// let EndxIndex == 0 mean no effect
		if opts.EndxIndex > 0 {
			endxIndex = opts.EndxIndex
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
		if !(key < int64(rc.Iv[h].Start)) {
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
			return key < rc.Iv[i].Start
		})
	*/
	whichInterval32 = int64(below) - 1

	if below == n {
		// all falses => key is >= start of all interval32s
		// ... so does it belong to the last interval32?
		if key < int64(rc.Iv[n-1].Last)+1 {
			// yes, it belongs to the last interval32
			alreadyPresent = true
			return
		}
		// no, it is beyond the last interval32.
		// leave alreadyPreset = false
		return
	}

	// INVAR: key is below rc.Iv[below]
	if below == 0 {
		// key is before the first first interval32.
		// leave alreadyPresent = false
		return
	}

	// INVAR: key is >= rc.Iv[below-1].Start and
	//        key is <  rc.Iv[below].Start

	// is key in below-1 interval32?
	if key >= int64(rc.Iv[below-1].Start) && key < int64(rc.Iv[below-1].Last)+1 {
		// yes, it is. key is in below-1 interval32.
		alreadyPresent = true
		return
	}

	// INVAR: key >= rc.Iv[below-1].endx && key < rc.Iv[below].Start
	//p("search, INVAR: key >= rc.Iv[below-1].endx && key < rc.Iv[below].Start, where key=%v, below=%v, below-1=%v, rc.Iv[below-1]=%v, rc.Iv[below]=%v", key, below, below-1, rc.Iv[below-1], rc.Iv[below])
	// leave alreadyPresent = false
	return
}

// cardinality returns the count of the integers stored in the
// runContainer32.
func (rc *runContainer32) cardinality() int64 {
	if len(rc.Iv) == 0 {
		rc.Card = 0
		return 0
	}
	if rc.Card > 0 {
		return rc.Card // already cached
	}
	// have to compute it
	var n int64
	for _, p := range rc.Iv {
		n += int64(p.runlen())
	}
	rc.Card = n // cache it
	return n
}

// AsSlice decompresses the contents into a []uint32 slice.
func (rc *runContainer32) AsSlice() []uint32 {
	s := make([]uint32, rc.cardinality())
	j := 0
	for _, p := range rc.Iv {
		for i := p.Start; i <= p.Last; i++ {
			s[j] = uint32(i)
			j++
		}
	}
	return s
}

// newRunContainer32 creates an empty run container.
func newRunContainer32() *runContainer32 {
	return &runContainer32{}
}

// newRunContainer32CopyIv creates a run container, initializing
// with a copy of the supplied iv slice.
//
func newRunContainer32CopyIv(iv []interval32) *runContainer32 {
	rc := &runContainer32{
		Iv: make([]interval32, len(iv)),
	}
	copy(rc.Iv, iv)
	return rc
}

func (rc *runContainer32) Clone() *runContainer32 {
	rc2 := newRunContainer32CopyIv(rc.Iv)
	return rc2
}

// newRunContainer32TakeOwnership returns a new runContainer32
// backed by the provided iv slice, which we will
// assume exclusive control over from now on.
//
func newRunContainer32TakeOwnership(iv []interval32) *runContainer32 {
	rc := &runContainer32{
		Iv: iv,
	}
	return rc
}

const baseRc32Size = int(unsafe.Sizeof(runContainer32{}))
const perIntervalRc32Size = int(unsafe.Sizeof(interval32{}))

// serializedSizeInBytes returns the number of bytes of memory
// required by this runContainer32.
func (rc *runContainer32) serializedSizeInBytes() int {
	return rc.Msgsize()
}

// see also runContainer32SerializedSizeInBytes(numRuns int) int

// getSizeInBytes returns the number of bytes of memory
// required by this runContainer32.
func (rc *runContainer32) getSizeInBytes() int {
	return perIntervalRc32Size * len(rc.Iv) // +  baseRc32Size
}

// runContainer32SerializedSizeInBytes returns the number of bytes of memory
// required to hold numRuns in a runContainer32.
func runContainer32SerializedSizeInBytes(numRuns int) int {
	return perIntervalRc32Size * numRuns // +  baseRc32Size
}

// Add adds a single value k to the set.
func (rc *runContainer32) Add(k uint32) (wasNew bool) {
	// TODO comment from runContainer32.java:
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

	// increment Card if it is cached already
	if rc.Card > 0 {
		rc.Card++
	}
	n := int64(len(rc.Iv))
	if index == -1 {
		// we may need to extend the first run
		if n > 0 {
			if rc.Iv[0].Start == k+1 {
				rc.Iv[0].Start = k
				return
			}
		}
		// nope, k stands alone, starting the new first interval32.
		rc.Iv = append([]interval32{interval32{Start: k, Last: k}}, rc.Iv...)
		return
	}

	// are we off the end? handle both index == n and index == n-1:
	if index >= n-1 {
		if int64(rc.Iv[n-1].Last)+1 == k64 {
			rc.Iv[n-1].Last++
			return
		}
		rc.Iv = append(rc.Iv, interval32{Start: k, Last: k})
		return
	}

	// INVAR: index and index+1 both exist, and k goes between them.
	//
	// Now: add k into the middle,
	// possibly fusing with index or index+1 interval32
	// and possibly resulting in fusing of two interval32s
	// that had a one integer gap.

	left := index
	right := index + 1

	// are we fusing left and right by adding k?
	if int64(rc.Iv[left].Last)+1 == k64 && int64(rc.Iv[right].Start) == k64+1 {
		// fuse into left
		rc.Iv[left].Last = rc.Iv[right].Last
		// remove redundant right
		rc.Iv = append(rc.Iv[:left+1], rc.Iv[right+1:]...)
		return
	}

	// are we an addition to left?
	if int64(rc.Iv[left].Last)+1 == k64 {
		// yes
		rc.Iv[left].Last++
		return
	}

	// are we an addition to right?
	if int64(rc.Iv[right].Start) == k64+1 {
		// yes
		rc.Iv[right].Start = k
		return
	}

	// k makes a standalone new interval32, inserted in the middle
	tail := append([]interval32{interval32{Start: k, Last: k}}, rc.Iv[right:]...)
	rc.Iv = append(rc.Iv[:left+1], tail...)
	return
}

//msgp:ignore RunIterator

// RunIterator32 advice: you must call Next() at least once
// before calling Cur(); and you should call HasNext()
// before calling Next() to insure there are contents.
type RunIterator32 struct {
	rc            *runContainer32
	curIndex      int64
	curPosInIndex uint32
	curSeq        int64
}

// NewRunIterator32 returns a new empty run container.
func (rc *runContainer32) NewRunIterator32() *RunIterator32 {
	return &RunIterator32{rc: rc, curIndex: -1}
}

func (ri *RunIterator32) hasNext() bool {
	return ri.HasNext()
}
func (ri *RunIterator32) next() uint32 {
	return ri.Next()
}

// HasNext returns false if calling Next will panic. It
// returns true when there is at least one more value
// available in the iteration sequence.
func (ri *RunIterator32) HasNext() bool {
	if len(ri.rc.Iv) == 0 {
		return false
	}
	if ri.curIndex == -1 {
		return true
	}
	return ri.curSeq+1 < ri.rc.cardinality()
}

// Cur returns the current value pointed to by the iterator.
func (ri *RunIterator32) Cur() uint32 {
	//p("in Cur, curIndex=%v, curPosInIndex=%v", ri.curIndex, ri.curPosInIndex)
	return ri.rc.Iv[ri.curIndex].Start + ri.curPosInIndex
}

// Next returns the next value in the iteration sequence.
func (ri *RunIterator32) Next() uint32 {
	if !ri.HasNext() {
		panic("no Next available")
	}
	if ri.curIndex >= int64(len(ri.rc.Iv)) {
		panic("RunIterator.Next() going beyond what is available")
	}
	if ri.curIndex == -1 {
		// first time is special
		ri.curIndex = 0
	} else {
		ri.curPosInIndex++
		if int64(ri.rc.Iv[ri.curIndex].Start)+int64(ri.curPosInIndex) == int64(ri.rc.Iv[ri.curIndex].Last)+1 {
			//p("rolling from ri.curIndex==%v to ri.curIndex=%v", ri.curIndex, ri.curIndex+1)
			ri.curPosInIndex = 0
			ri.curIndex++
		} else {
			//p("no roll ... ri.curPosInIndex is now %v, ri.rc.Iv[ri.curIndex].endx=%v", ri.curPosInIndex, ri.rc.Iv[ri.curIndex].endx)
		}
		ri.curSeq++
	}
	return ri.Cur()
}

// Remove removes the element that the iterator
// is on from the run container. You can use
// Cur if you want to double check what is about
// to be deleted.
func (ri *RunIterator32) Remove() uint32 {
	n := ri.rc.cardinality()
	if n == 0 {
		panic("RunIterator.Remove called on empty runContainer32")
	}
	cur := ri.Cur()

	ri.rc.deleteAt(&ri.curIndex, &ri.curPosInIndex, &ri.curSeq)
	return cur
}

// remove removes key from the container.
func (rc *runContainer32) removeKey(key uint32) (wasPresent bool) {

	var index int64
	var curSeq int64
	index, wasPresent, _ = rc.search(int64(key), nil)
	if !wasPresent {
		return // already removed, nothing to do.
	}
	pos := key - rc.Iv[index].Start
	rc.deleteAt(&index, &pos, &curSeq)
	return
}

// internal helper functions

func (rc *runContainer32) deleteAt(curIndex *int64, curPosInIndex *uint32, curSeq *int64) {
	rc.Card--
	(*curSeq)--
	ci := *curIndex
	pos := *curPosInIndex

	// are we first, last, or in the middle of our interval32?
	switch {
	case pos == 0:
		//p("pos == 0, first")
		if int64(rc.Iv[ci].Start) == int64(rc.Iv[ci].Last) {
			// our interval disappears
			rc.Iv = append(rc.Iv[:ci], rc.Iv[ci+1:]...)
			// curIndex stays the same, since the delete did
			// the advance for us.
			*curPosInIndex = 0
		} else {
			rc.Iv[ci].Start++ // no longer overflowable
		}
	case int64(pos) == rc.Iv[ci].runlen()-1:
		// last
		rc.Iv[ci].Last--
		// our interval32 cannot disappear, else we would have been pos == 0, case first above.
		//p("deleteAt: pos is last case, curIndex=%v, curPosInIndex=%v", *curIndex, *curPosInIndex)
		(*curPosInIndex)--
		// if we leave *curIndex alone, then Next() will work properly even after the delete.
		//p("deleteAt: pos is last case, after update: curIndex=%v, curPosInIndex=%v", *curIndex, *curPosInIndex)
	default:
		//p("middle...split")
		//middle
		// split into two, adding an interval32
		new0 := interval32{
			Start: rc.Iv[ci].Start,
			Last:  rc.Iv[ci].Start + *curPosInIndex - 1}

		new1start := int64(rc.Iv[ci].Start) + int64(*curPosInIndex) + 1
		if new1start > int64(MaxUint32) {
			panic("overflow?!?!")
		}
		new1 := interval32{
			Start: uint32(new1start),
			Last:  rc.Iv[ci].Last}

		//p("new0 = %#v", new0)
		//p("new1 = %#v", new1)

		tail := append([]interval32{new0, new1}, rc.Iv[ci+1:]...)
		rc.Iv = append(rc.Iv[:ci], tail...)
		// update curIndex and curPosInIndex
		(*curIndex)++
		*curPosInIndex = 0
	}

}

func have4Overlap32(astart, alast, bstart, blast int64) bool {
	if int64(alast)+1 <= bstart {
		return false
	}
	return int64(blast)+1 > astart
}

func intersectWithLeftover32(astart, alast, bstart, blast int64) (isOverlap, isLeftoverA, isLeftoverB bool, leftoverStart int64, intersection interval32) {
	if !have4Overlap32(astart, alast, bstart, blast) {
		return
	}
	isOverlap = true

	// do the intersection:
	if bstart > astart {
		intersection.Start = uint32(bstart)
	} else {
		intersection.Start = uint32(astart)
	}
	switch {
	case blast < alast:
		isLeftoverA = true
		leftoverStart = int64(blast) + 1
		intersection.Last = uint32(blast)
	case alast < blast:
		isLeftoverB = true
		leftoverStart = int64(alast) + 1
		intersection.Last = uint32(alast)
	default:
		// alast == blast
		intersection.Last = uint32(alast)
	}

	return
}

func (rc *runContainer32) findNextIntervalThatIntersectsStartingFrom(startIndex int64, key int64) (index int64, done bool) {

	rc.myOpts.StartIndex = startIndex
	rc.myOpts.EndxIndex = 0

	w, _, _ := rc.search(key, &rc.myOpts)
	// rc.search always returns w < len(rc.Iv)
	if w < startIndex {
		// not found and comes before lower bound startIndex,
		// so just use the lower bound.
		if startIndex == int64(len(rc.Iv)) {
			// also this bump up means that we are done
			return startIndex, true
		}
		return startIndex, false
	}

	return w, false
}

func sliceToString32(m []interval32) string {
	s := ""
	for i := range m {
		s += fmt.Sprintf("%v: %s, ", i, m[i])
	}
	return s
}

// selectInt32 returns the j-th value in the container.
// We panic of j is out of bounds.
func (rc *runContainer32) selectInt32(j uint32) int {
	n := rc.cardinality()
	if int64(j) > n {
		panic(fmt.Sprintf("Cannot select %v since Cardinality is %v", j, n))
	}

	var offset int64
	for k := range rc.Iv {
		nextOffset := offset + rc.Iv[k].runlen() + 1
		if nextOffset > int64(j) {
			return int(int64(rc.Iv[k].Start) + (int64(j) - offset))
		}
		offset = nextOffset
	}
	panic(fmt.Sprintf("Cannot select %v since Cardinality is %v", j, n))
}

// helper for invert
func (rc *runContainer32) invertLastInterval(origin uint32, lastIdx int) []interval32 {
	cur := rc.Iv[lastIdx]
	if cur.Last == MaxUint32 {
		if cur.Start == origin {
			return nil // empty container
		}
		return []interval32{{Start: origin, Last: cur.Start - 1}}
	}
	if cur.Start == origin {
		return []interval32{{Start: cur.Last + 1, Last: MaxUint32}}
	}
	// invert splits
	return []interval32{
		{Start: origin, Last: cur.Start - 1},
		{Start: cur.Last + 1, Last: MaxUint32},
	}
}

// invert returns a new container (not inplace), that is
// the inversion of rc. For each bit b in rc, the
// returned value has !b
func (rc *runContainer32) invert() *runContainer32 {
	ni := len(rc.Iv)
	var m []interval32
	switch ni {
	case 0:
		return &runContainer32{Iv: []interval32{{0, MaxUint32}}}
	case 1:
		return &runContainer32{Iv: rc.invertLastInterval(0, 0)}
	}
	var invStart int64
	ult := ni - 1
	for i, cur := range rc.Iv {
		if i == ult {
			// invertLastInteval will add both intervals (b) and (c) in
			// diagram below.
			m = append(m, rc.invertLastInterval(uint32(invStart), i)...)
			break
		}
		// INVAR: i and cur are not the last interval, there is a next at i+1
		//
		// ........[cur.Start, cur.Last] ...... [next.Start, next.Last]....
		//    ^                             ^                           ^
		//   (a)                           (b)                         (c)
		//
		// Now: we add interval (a); but if (a) is empty, for cur.Start==0, we skip it.
		if cur.Start > 0 {
			m = append(m, interval32{Start: uint32(invStart), Last: cur.Start - 1})
		}
		invStart = int64(cur.Last + 1)
	}
	return &runContainer32{Iv: m}
}

func (a interval32) equal(b interval32) bool {
	if a.Start == b.Start {
		return a.Last == b.Last
	}
	return false
}

func (a interval32) isSuperSetOf(b interval32) bool {
	return a.Start <= b.Start && b.Last <= a.Last
}

func (cur interval32) subtractInterval(del interval32) (left []interval32, delcount int64) {
	defer func() {
		//p("returning from subtractInterval of cur - del with cur=%s and del=%s, returning left=%s, delcount=%v", cur, del, ivalString32(left), delcount)
	}()
	isect, isEmpty := intersectInterval32s(cur, del)

	if isEmpty {
		//p("isEmpty")
		return nil, 0
	}
	if del.isSuperSetOf(cur) {
		return nil, cur.runlen()
	}

	switch {
	case isect.Start > cur.Start && isect.Last < cur.Last:
		//p("split into two")
		new0 := interval32{Start: cur.Start, Last: isect.Start - 1}
		new1 := interval32{Start: isect.Last + 1, Last: cur.Last}
		return []interval32{new0, new1}, isect.runlen()
	case isect.Start == cur.Start:
		//p("removal of only the first half or so of cur interval. isect: %s, cur: %s, del: %s", isect, cur, del)
		return []interval32{{Start: isect.Last + 1, Last: cur.Last}}, isect.runlen()
	default:
		//p("isect.end == cur.end")
		//p("removal of only the last half or so of cur interval")
		return []interval32{{Start: cur.Start, Last: isect.Start - 1}}, isect.runlen()
	}
}

func (rc *runContainer32) isubtract(del interval32) {
	origiv := make([]interval32, len(rc.Iv))
	copy(origiv, rc.Iv)
	//p("isubtract starting, with del = %s, and rc = %s", del, rc)
	n := int64(len(rc.Iv))
	if n == 0 {
		return // already done.
	}

	_, isEmpty := intersectInterval32s(
		interval32{
			Start: rc.Iv[0].Start,
			Last:  rc.Iv[n-1].Last,
		}, del)
	if isEmpty {
		//p("del=%v -> isEmpty, returning early from isubtract", del)
		return // done
	}
	// INVAR there is some intersection between rc and del
	istart, startAlready, _ := rc.search(int64(del.Start), nil)
	ilast, lastAlready, _ := rc.search(int64(del.Last), nil)
	rc.Card = -1

	//p("del=%v, istart = %v, startAlready = %v", del, istart, startAlready)
	//p("del=%v, ilast = %v, lastAlready = %v", del, ilast, lastAlready)

	if istart == -1 {
		if ilast == n-1 {
			//p("discard it all")
			rc.Iv = nil
			return
		}
	}
	// some intervals will remain
	//p("orig rc.Iv = '%s'", ivalString32(rc.Iv))
	switch {
	case startAlready && lastAlready:
		//p("case 1: startAlready && lastAlready; istart=%v, ilast=%v. staring rc.Iv='%s'", istart, ilast, ivalString32(rc.Iv))
		res0, _ := rc.Iv[istart].subtractInterval(del)

		//p("case 1 rc.Iv[:start] = '%s', while res0='%s'", ivalString32(rc.Iv[:istart]), ivalString32(res0))
		// would overwrite values in iv b/c res0 can have len 2. so
		// write to origiv instead.

		//p("case 1 pre = '%s'", ivalString32(pre))
		//p("orig rc.Iv = '%s'", ivalString32(rc.Iv))

		lost := 1 + ilast - istart
		changeSize := int64(len(res0)) - lost
		newSize := int64(len(rc.Iv)) + changeSize

		//p("case 1 before suffixing with: rc.Iv[ilast+1:] = '%s'", ivalString32(rc.Iv[ilast+1:]))
		//	rc.Iv = append(pre, caboose...)
		//	return

		if ilast != istart {
			res1, _ := rc.Iv[ilast].subtractInterval(del)
			res0 = append(res0, res1...)
			changeSize = int64(len(res0)) - lost
			newSize = int64(len(rc.Iv)) + changeSize
		}
		switch {
		case changeSize < 0:
			// shrink
			copy(rc.Iv[istart+int64(len(res0)):], rc.Iv[ilast+1:])
			copy(rc.Iv[istart:istart+int64(len(res0))], res0)
			rc.Iv = rc.Iv[:newSize]
			return
		case changeSize == 0:
			// stay the same
			copy(rc.Iv[istart:istart+int64(len(res0))], res0)
			return
		default:
			// changeSize > 0 is only possible when ilast == istart.
			// Hence we now know: changeSize == 1 and len(res0) == 2
			rc.Iv = append(rc.Iv, interval32{})
			// len(rc.Iv) is correct now, no need to rc.Iv = rc.Iv[:newSize]

			// copy the tail into place
			copy(rc.Iv[ilast+2:], rc.Iv[ilast+1:])
			// copy the new item(s) into place
			copy(rc.Iv[istart:istart+2], res0)
			return
		}

	case !startAlready && !lastAlready:
		//p("case 2: !startAlready && !lastAlready")
		// we get to discard whole intervals

		// from the search() definition:

		// if del.Start is not present, then istart is
		// set as follows:
		//
		//  a) istart == n-1 if del.Start is beyond our
		//     last interval32 in rc.Iv;
		//
		//  b) istart == -1 if del.Start is before our first
		//     interval32 in rc.Iv;
		//
		//  c) istart is set to the minimum index of rc.Iv
		//     which comes strictly before the del.Start;
		//     so  del.Start > rc.Iv[istart].Last,
		//     and  if istart+1 exists, then del.Start < rc.Iv[istart+1].Startx

		// if del.Last is not present, then ilast is
		// set as follows:
		//
		//  a) ilast == n-1 if del.Last is beyond our
		//     last interval32 in rc.Iv;
		//
		//  b) ilast == -1 if del.Last is before our first
		//     interval32 in rc.Iv;
		//
		//  c) ilast is set to the minimum index of rc.Iv
		//     which comes strictly before the del.Last;
		//     so  del.Last > rc.Iv[ilast].Last,
		//     and  if ilast+1 exists, then del.Last < rc.Iv[ilast+1].Start

		// INVAR: istart >= 0
		pre := rc.Iv[:istart+1]
		if ilast == n-1 {
			rc.Iv = pre
			return
		}
		// INVAR: ilast < n-1
		lost := ilast - istart
		changeSize := -lost
		newSize := int64(len(rc.Iv)) + changeSize
		if changeSize != 0 {
			copy(rc.Iv[ilast+1+changeSize:], rc.Iv[ilast+1:])
		}
		rc.Iv = rc.Iv[:newSize]
		return

	case startAlready && !lastAlready:
		// we can only shrink or stay the same size
		// i.e. we either eliminate the whole interval,
		// or just cut off the right side.
		//p("case 3: startAlready && !lastAlready, rc='%s', del='%s'", rc, del)
		res0, _ := rc.Iv[istart].subtractInterval(del)
		if len(res0) > 0 {
			// len(res) must be 1
			rc.Iv[istart] = res0[0]
		}
		lost := 1 + (ilast - istart)
		changeSize := int64(len(res0)) - lost
		newSize := int64(len(rc.Iv)) + changeSize
		if changeSize != 0 {
			copy(rc.Iv[ilast+1+changeSize:], rc.Iv[ilast+1:])
		}
		rc.Iv = rc.Iv[:newSize]
		return

	case !startAlready && lastAlready:
		// we can only shrink or stay the same size
		//p("case 4: !startAlready && lastAlready")
		res1, _ := rc.Iv[ilast].subtractInterval(del)
		lost := ilast - istart
		changeSize := int64(len(res1)) - lost
		newSize := int64(len(rc.Iv)) + changeSize
		if changeSize != 0 {
			// move the tail first to make room for res1
			copy(rc.Iv[ilast+1+changeSize:], rc.Iv[ilast+1:])
		}
		copy(rc.Iv[istart+1:], res1)
		rc.Iv = rc.Iv[:newSize]
		return
	}
}

// compute rc minus b, and return the result as a new value (not inplace).
// port of run_container_andnot from CRoaring...
// https://github.com/RoaringBitmap/CRoaring/blob/master/src/containers/run.c#L435-L496
func (rc *runContainer32) AndNotRunContainer32(b *runContainer32) *runContainer32 {

	if len(b.Iv) == 0 || len(rc.Iv) == 0 {
		return rc
	}

	dst := newRunContainer32()
	apos := 0
	bpos := 0

	a := rc

	astart := a.Iv[apos].Start
	alast := a.Iv[apos].Last
	bstart := b.Iv[bpos].Start
	blast := b.Iv[bpos].Last

	alen := len(a.Iv)
	blen := len(b.Iv)

	for apos < alen && bpos < blen {
		//p("top: apos = %v, alen=%v, bpos=%v, blen=%v", apos, alen, bpos, blen)
		switch {
		case alast < bstart:
			// output the first run
			dst.Iv = append(dst.Iv, interval32{Start: uint32(astart), Last: uint32(alast)})
			//p("alast(%v) < bstart(%v), dst after adding [astart, last] is: %s", alast, bstart, dst)
			apos++
			if apos < alen {
				astart = a.Iv[apos].Start
				alast = a.Iv[apos].Last
			}
		case blast < astart:
			// exit the second run
			bpos++
			if bpos < blen {
				bstart = b.Iv[bpos].Start
				blast = b.Iv[bpos].Last
			}
		default:
			//   a: [             ]
			//   b:            [    ]
			// alast >= bstart
			// blast >= astart
			if astart < bstart {
				dst.Iv = append(dst.Iv, interval32{Start: uint32(astart), Last: uint32(bstart - 1)})
				//p("astart(%v) < bstart(%v), dst after adding [astart, bstart] is: %s", astart, bstart, dst)
			}
			if alast > blast {
				astart = blast + 1
			} else {
				apos++
				if apos < alen {
					astart = a.Iv[apos].Start
					alast = a.Iv[apos].Last
				}
			}
		}
	}
	if apos < alen {
		dst.Iv = append(dst.Iv, interval32{Start: uint32(astart), Last: uint32(alast)})
		apos++
		if apos < alen {
			dst.Iv = append(dst.Iv, a.Iv[apos:]...)
		}
	}

	return dst
}

func (rc *runContainer32) numberOfRuns() (nr int) {
	return len(rc.Iv)
}

func (bc *runContainer32) containerType() contype {
	return run32Contype
}
