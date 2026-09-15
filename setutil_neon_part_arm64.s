// +build arm64,!gccgo,!appengine

#include "textflag.h"

// Block-maximum partition kernel for union2by2.
//
// Per iteration both inputs contribute a 16-element window:
//   A = set1[pa:pa+16], B = set2[pb:pb+16], am = A[15], bm = B[15]
//   cntA = #{i: A[i] <= bm}, cntB = #{i: B[i] <= am}, T = cntA + cntB
// With x = min(am,bm) those counts are exactly the elements <= x on each
// side, so the T smallest of the 32 merged values are the whole union up to
// x and nothing outside the two windows can fall below x. The merge network
// sorts all 32; the first T are emitted and the cursors advance by cntA and
// cntB. Nothing produced by the merge feeds the next iteration's cursors, so
// the loop-carried chain is only load -> compare -> count -> advance.
//
// Every iteration consumes all values <= x on both sides, so no duplicate
// crosses an iteration boundary. The first merged vector holds at most two
// copies of any value, hence comparing its first lane with its last lane
// cannot drop the first output.
//
// Below 64 elements on either side, or when the 16-lane loop reaches its
// reserve, an 8-lane partition stage continues with the same invariant.
// It halves the merge and compaction work and leaves a shorter scalar tail.
//
// Register use:
//   R0/R1 byte cursors into set1/set2, R3/R4 their loop end addresses,
//   R2 output cursor, R5 output base, R6 uniqshuf table,
//   R12 = 0x0101010101010101, R13 = 0x0102040810204080, R20 = 8.
//   V0/V1 A window, V2/V3 B window, V4 am, V5 bm, V6 x,
//   V16-V19 merged 32, V20-V23 compaction shuffles.

// Sort two 8-element bitonic sequences u,v (as produced by a compare-exchange
// pair) into lo,hi. Transposed network: one TRN pair per distance, both
// halves cleaned at once, ZIP to untranspose. Clobbers V14, V15, u, v.
#define CLEAN(u, v, lo, hi)         \
	VTRN1 v.D2, u.D2, V14.D2   \
	VTRN2 v.D2, u.D2, V15.D2   \
	VUMIN V15.H8, V14.H8, u.H8 \
	VUMAX V15.H8, V14.H8, v.H8 \
	VTRN1 v.S4, u.S4, V14.S4   \
	VTRN2 v.S4, u.S4, V15.S4   \
	VUMIN V15.H8, V14.H8, u.H8 \
	VUMAX V15.H8, V14.H8, v.H8 \
	VTRN1 v.H8, u.H8, V14.H8   \
	VTRN2 v.H8, u.H8, V15.H8   \
	VUMIN V15.H8, V14.H8, u.H8 \
	VUMAX V15.H8, V14.H8, v.H8 \
	VZIP1 v.H8, u.H8, lo.H8    \
	VZIP2 v.H8, u.H8, hi.H8

// Merge sorted (V0,V1) with sorted (V2,V3) into sorted V16..V19.
// Reversing B makes the concatenation bitonic; two compare-exchange stages
// split it into two bitonic 16-sequences, each cleaned by CLEAN.
#define MERGE32                            \
	VREV64 V3.H8, V8.H8                \
	VEXT   $8, V8.B16, V8.B16, V8.B16  \
	VREV64 V2.H8, V9.H8                \
	VEXT   $8, V9.B16, V9.B16, V9.B16  \
	VUMIN  V8.H8, V0.H8, V10.H8        \
	VUMAX  V8.H8, V0.H8, V12.H8        \
	VUMIN  V9.H8, V1.H8, V11.H8        \
	VUMAX  V9.H8, V1.H8, V13.H8        \
	VUMIN  V11.H8, V10.H8, V8.H8       \
	VUMAX  V11.H8, V10.H8, V9.H8       \
	VUMIN  V13.H8, V12.H8, V10.H8      \
	VUMAX  V13.H8, V12.H8, V11.H8      \
	CLEAN(V8, V9, V16, V17)            \
	CLEAN(V10, V11, V18, V19)

// Count a 16-lane prefix mask in t0,t1, then advance byte cursor p.
// UZP1 keeps one byte per predicate; SHRN #4 keeps four bits per
// predicate in a single doubleword. Thus the byte advance is 32-clz(mask)/2.
// nt0 is t0's register number for SHRN, which Go's assembler does not expose.
// Clobbers t0, t1, Ra, Rc.
#define COUNTADV(t0, t1, Ra, Rc, p, nt0) \
	VUZP1 t1.B16, t0.B16, t0.B16 \
	WORD $(0x0f0c8400 | (nt0 << 5) | nt0) \
	VMOV  t0.D[0], Ra          \
	CLZ   Ra, Ra               \
	ADD   $32, p, Rc           \
	SUB   Ra>>1, Rc, p

// Eight predicates fit in a doubleword after keeping one byte per lane.
#define COUNTADV8(t, Ra, Rc, p) \
	VUZP1 t.B16, t.B16, t.B16 \
	VMOV  t.D[0], Ra          \
	CLZ   Ra, Ra              \
	ADD   $16, p, Rc          \
	SUB   Ra>>2, Rc, p

// Build the compaction shuffle sv and kept-lane count Rcnt for cur, whose
// predecessor lane is lane 7 of prv. Clobbers t, R17, R19, R21.
#define PREP(cur, prv, t, sv, Rcnt) \
	VEXT  $14, cur.B16, prv.B16, t.B16 \
	VCMEQ t.H8, cur.H8, t.H8   \
	VUZP1 t.B16, t.B16, t.B16  \
	VMOV  t.D[0], R17          \
	AND   R12, R17, R17        \
	MUL   R13, R17, R19        \
	LSR   $56, R19, R19        \
	MUL   R12, R17, R17        \
	LSR   $56, R17, R17        \
	SUB   R17, R20, Rcnt       \
	ADD   R19<<4, R6, R21      \
	VLD1  (R21), [sv.B16]

// func unionPartKernelNEON(set1, set2, buffer []uint16, shuf *byte) (outLen, pos1, pos2 int)
TEXT ·unionPartKernelNEON(SB), NOSPLIT, $0-104
	MOVD set1_base+0(FP), R0
	MOVD set2_base+24(FP), R1
	MOVD buffer_base+48(FP), R2
	MOVD set1_len+8(FP), R3
	MOVD set2_len+32(FP), R4
	MOVD shuf+72(FP), R6
	MOVD R2, R5
	MOVD $0x0101010101010101, R12
	MOVD $0x0102040810204080, R13
	MOVD $8, R20
	CMP  $64, R3
	BLT  setup8
	CMP  $64, R4
	BLT  setup8

	// Stop 24 elements short of each whole-block end. Each iteration leaves
	// at least 9 elements per side unread; compaction stores extend at most
	// 8 elements past the output cursor. This also protects unread set1
	// when iorArray places it at buffer[len(set2):].
	AND  $~7, R3, R3
	SUB  $24, R3, R3
	ADD  R3<<1, R0, R3
	AND  $~7, R4, R4
	SUB  $24, R4, R4
	ADD  R4<<1, R1, R4

loop:
	CMP R3, R0
	BHS setup8
	CMP R4, R1
	BHS setup8

	// Non-overlapping windows: the lower one is the union up to its own
	// maximum, so it is emitted as loaded and the merge, the counts and the
	// compaction are all skipped. Scalar endpoint loads resolve the branch
	// without waiting on the vector loads. Boundary equality falls through
	// to the general path, which dedups it.
	MOVHU 30(R0), R8
	MOVHU (R1), R9
	MOVHU 30(R1), R10
	MOVHU (R0), R11
	CMP   R9, R8
	BLO   fast1
	CMP   R11, R10
	BLO   fast2

	// Half-window retry: ownership runs shorter than the window never let
	// the tests above fire, so they are repeated at 8 elements.
	MOVHU 14(R0), R16
	CMP   R9, R16
	BLO   fast3
	MOVHU 14(R1), R17
	CMP   R11, R17
	BLO   fast4

	VLD1 (R0), [V0.H8, V1.H8]
	VLD1 (R1), [V2.H8, V3.H8]
	VDUP V1.H[7], V4.H8
	VDUP V3.H[7], V5.H8

	WORD $0x6e603ca8 // CMHS V8.8H, V5.8H, V0.8H
	WORD $0x6e613ca9 // CMHS V9.8H, V5.8H, V1.8H
	COUNTADV(V8, V9, R8, R10, R0, 8)
	WORD $0x6e623c8a // CMHS V10.8H, V4.8H, V2.8H
	WORD $0x6e633c8b // CMHS V11.8H, V4.8H, V3.8H
	COUNTADV(V10, V11, R14, R16, R1, 10)

	VUMIN V5.H8, V4.H8, V6.H8

	MERGE32

	// Lanes above x are outside this iteration's emission. Clamping them to
	// x makes each a repeat of its predecessor, so the dedup compare drops
	// them. Only the last two vectors can hold such lanes (T >= 16).
	VUMIN V6.H8, V18.H8, V18.H8
	VUMIN V6.H8, V19.H8, V19.H8

	PREP(V16, V16, V8, V20, R22)
	PREP(V17, V16, V9, V21, R23)
	PREP(V18, V17, V10, V22, R24)
	PREP(V19, V18, V11, V23, R25)

	VTBL V20.B16, [V16.B16], V12.B16
	VST1 [V12.B16], (R2)
	ADD  R22<<1, R2, R2
	VTBL V21.B16, [V17.B16], V13.B16
	VST1 [V13.B16], (R2)
	ADD  R23<<1, R2, R2
	VTBL V22.B16, [V18.B16], V14.B16
	VST1 [V14.B16], (R2)
	ADD  R24<<1, R2, R2
	VTBL V23.B16, [V19.B16], V15.B16
	VST1 [V15.B16], (R2)
	ADD  R25<<1, R2, R2

	B    loop

fast1:
	VLD1 (R0), [V0.H8, V1.H8]
	VST1 [V0.H8, V1.H8], (R2)
	ADD  $32, R0, R0
	ADD  $32, R2, R2
	B    loop

fast2:
	VLD1 (R1), [V2.H8, V3.H8]
	VST1 [V2.H8, V3.H8], (R2)
	ADD  $32, R1, R1
	ADD  $32, R2, R2
	B    loop

fast3:
	VLD1 (R0), [V0.H8]
	VST1 [V0.H8], (R2)
	ADD  $16, R0, R0
	ADD  $16, R2, R2
	B    loop

fast4:
	VLD1 (R1), [V2.H8]
	VST1 [V2.H8], (R2)
	ADD  $16, R1, R1
	ADD  $16, R2, R2
	B    loop

setup8:
	// Two full compaction stores need 16 unread elements per input to
	// preserve the relocated-input alias. Cursors need not be aligned.
	MOVD set1_len+8(FP), R3
	SUB  $16, R3, R3
	MOVD set1_base+0(FP), R8
	ADD  R3<<1, R8, R3
	MOVD set2_len+32(FP), R4
	SUB  $16, R4, R4
	MOVD set2_base+24(FP), R8
	ADD  R4<<1, R8, R4

loop8:
	CMP R3, R0
	BHI done
	CMP R4, R1
	BHI done

	MOVHU 14(R0), R8
	MOVHU (R1), R9
	CMP R9, R8
	BLO fastA8
	MOVHU 14(R1), R10
	MOVHU (R0), R11
	CMP R11, R10
	BLO fastB8

	VLD1 (R0), [V0.H8]
	VLD1 (R1), [V1.H8]
	VDUP V0.H[7], V4.H8
	VDUP V1.H[7], V5.H8
	WORD $0x6e603ca8 // CMHS V8.8H, V5.8H, V0.8H
	COUNTADV8(V8, R8, R10, R0)
	WORD $0x6e613c89 // CMHS V9.8H, V4.8H, V1.8H
	COUNTADV8(V9, R14, R16, R1)
	VUMIN V5.H8, V4.H8, V6.H8

	VREV64 V1.H8, V8.H8
	VEXT   $8, V8.B16, V8.B16, V8.B16
	VUMIN  V8.H8, V0.H8, V9.H8
	VUMAX  V8.H8, V0.H8, V10.H8
	CLEAN(V9, V10, V16, V17)
	VUMIN V6.H8, V17.H8, V17.H8

	PREP(V16, V16, V8, V20, R22)
	PREP(V17, V16, V9, V21, R23)
	VTBL V20.B16, [V16.B16], V12.B16
	VST1 [V12.B16], (R2)
	ADD  R22<<1, R2, R2
	VTBL V21.B16, [V17.B16], V13.B16
	VST1 [V13.B16], (R2)
	ADD  R23<<1, R2, R2
	B loop8

fastA8:
	VLD1 (R0), [V0.H8]
	VST1 [V0.H8], (R2)
	ADD $16, R0, R0
	ADD $16, R2, R2
	B loop8

fastB8:
	VLD1 (R1), [V1.H8]
	VST1 [V1.H8], (R2)
	ADD $16, R1, R1
	ADD $16, R2, R2
	B loop8

done:
	SUB  R5, R2, R2
	LSR  $1, R2, R2
	MOVD R2, outLen+80(FP)

	MOVD set1_base+0(FP), R8
	SUB  R8, R0, R0
	LSR  $1, R0, R0
	MOVD R0, pos1+88(FP)
	MOVD set2_base+24(FP), R8
	SUB  R8, R1, R1
	LSR  $1, R1, R1
	MOVD R1, pos2+96(FP)
	RET
