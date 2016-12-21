package roaring

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCountTrailingZeros072(t *testing.T) {
	Convey("countTrailingZeros", t, func() {
		// undefined on older cpus, so skip this check on 0.
		//So(countTrailingZerosAsm(0), ShouldEqual, 64)

		So(countTrailingZerosAsm(8), ShouldEqual, 3)
		So(countTrailingZerosAsm(7), ShouldEqual, 0)
		So(countTrailingZerosAsm(1<<17), ShouldEqual, 17)
		So(countTrailingZerosAsm(7<<17), ShouldEqual, 17)
		So(countTrailingZerosAsm(255<<33), ShouldEqual, 33)

		So(countTrailingZerosDeBruijn(0), ShouldEqual, 64)
		So(countTrailingZerosDeBruijn(8), ShouldEqual, 3)
		So(countTrailingZerosDeBruijn(7), ShouldEqual, 0)
		So(countTrailingZerosDeBruijn(1<<17), ShouldEqual, 17)
		So(countTrailingZerosDeBruijn(7<<17), ShouldEqual, 17)
		So(countTrailingZerosDeBruijn(255<<33), ShouldEqual, 33)

	})
}

func getRandomUint64Set(n int) []uint64 {
	seed := int64(42)
	p("seed is %v", seed)
	rand.Seed(seed)

	var buf [8]byte
	var o []uint64
	for i := 0; i < n; i++ {
		rand.Read(buf[:])
		o = append(o, binary.LittleEndian.Uint64(buf[:]))
	}
	return o
}

func getAllOneBitUint64Set() []uint64 {
	var o []uint64
	for i := uint(0); i < 64; i++ {
		o = append(o, 1<<i)
	}
	return o
}

func Benchmark100OrigNumberOfTrailingZeros(b *testing.B) {
	b.StopTimer()

	r := getRandomUint64Set(64)
	r = append(r, getAllOneBitUint64Set()...)

	b.ResetTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		for i := range r {
			numberOfTrailingZeros(r[i])
		}
	}
}

func Benchmark100CountTrailingZerosDeBruijn(b *testing.B) {
	b.StopTimer()

	r := getRandomUint64Set(64)
	r = append(r, getAllOneBitUint64Set()...)

	b.ResetTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		for i := range r {
			countTrailingZerosDeBruijn(r[i])
		}
	}
}

func Benchmark100CountTrailingZerosAsm(b *testing.B) {
	b.StopTimer()

	r := getRandomUint64Set(64)
	r = append(r, getAllOneBitUint64Set()...)

	b.ResetTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		for i := range r {
			countTrailingZerosAsm(r[i])
		}
	}
}

func numberOfTrailingZeros(i uint64) int {
	if i == 0 {
		return 64
	}
	x := i
	n := int64(63)
	y := x << 32
	if y != 0 {
		n -= 32
		x = y
	}
	y = x << 16
	if y != 0 {
		n -= 16
		x = y
	}
	y = x << 8
	if y != 0 {
		n -= 8
		x = y
	}
	y = x << 4
	if y != 0 {
		n -= 4
		x = y
	}
	y = x << 2
	if y != 0 {
		n -= 2
		x = y
	}
	return int(n - int64(uint64(x<<1)>>63))
}

/*
//
// on an Intel(R) Core(TM) i7-5557U CPU @ 3.10GHz:
//
Benchmark100CountTrailingZerosDeBruijn-4   	10000000	       168 ns/op
Benchmark100CountTrailingZerosAsm-4        	 5000000	       278 ns/op
Benchmark100OrigNumberOfTrailingZeros-4    	 3000000	       592 ns/op

// and again:

Benchmark100CountTrailingZerosDeBruijn-4   	10000000	       168 ns/op
Benchmark100CountTrailingZerosAsm-4        	 5000000	       278 ns/op
Benchmark100OrigNumberOfTrailingZeros-4    	 3000000	       585 ns/op
*/
// go test -v -bench=100 -run 101
func Test101CountTrailingZerosCorrectness(t *testing.T) {
	r := getAllOneBitUint64Set()
	for i, v := range r {
		a := countTrailingZerosDeBruijn(v)
		b := countTrailingZerosAsm(v)
		if a != b {
			panic(fmt.Errorf("on r[%v]= v=%v,  a: %v, b:%v", i, v, a, b))
		}
	}
	// don't do zero checks, since the Asm version can be undefined
	// for older architectures
	/*
		related Intel spec:

		    LZCNT is an extension of the BSR instruction. The key difference
		    between LZCNT and BSR is that LZCNT provides operand size as output
		    when source operand is zero, while in the case of BSR instruction,
		    if source operand is zero, the content of destination operand are
		    undefined. On processors that do not support LZCNT, the instruction
		    byte encoding is executed as BSR.

				if countTrailingZerosAsm(0) != countTrailingZerosDeBruijn(0) {
					panic("disagree on zero value")
				}
	*/

}
