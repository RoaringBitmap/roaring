// +build arm64,!gccgo,!appengine

#include "textflag.h"

// Block-maximum partition kernel for exclusiveUnion2by2.
//
// Each iteration takes a 16-element window from both inputs. With
// x = min(A[15], B[15]), the elements at or below x on each side are exactly
// the values this step settles, because nothing unread can fall below x. All
// 32 lanes go through a bitonic merge network; a value survives when it
// differs from both neighbours and is at or below x. The cursors advance by
// the two counts.
//
// Lanes above x are dropped by an explicit compare, never by clamping them to
// x, which would make a singleton x right before them read as a repeat.
//
// Entry needs 16 unread elements per side, so the output cursor stays at
// least 32 elements short of cap(buffer) = len1+len2 and the four full-width
// stores need no reserve. Output never aliases an input.
//
// Registers: R0/R1 input cursors, R3/R4 loop ends, R2 output cursor, R5 output
// base, R6 uniqshuf, R12/R13 movemask constants, R20 = 8. V0-V3 windows and
// then predecessor-equality masks, V4/V5 window maxima, V6 x, V16-V19 merged,
// V20-V23 shuffles, V29 all ones, V30 = [0, FFFF x7], V31 zero.

// Sort the bitonic pair u,v into lo,hi: transposed network, one TRN pair per
// distance, ZIP to untranspose. Clobbers V14, V15, u, v.
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

// Merge sorted (V0,V1) and (V2,V3) into sorted V16..V19: reverse B, two
// compare-exchange stages, then CLEAN each bitonic half.
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

// Advance byte cursor p by the lanes of window (v0,v1) that are <= m. The
// compare mask of a sorted window is a prefix, so the two byte-collapsed
// doublewords give cnt = (128 - clz_lo - clz_hi) >> 3. Clobbers t0, t1, Ra-Rc.
#define COUNTADV(v0, v1, m, t0, t1, Ra, Rb, Rc, p) \
	VUMIN v0.H8, m.H8, t0.H8     \
	VCMEQ t0.H8, v0.H8, t0.H8    \
	VUMIN v1.H8, m.H8, t1.H8     \
	VCMEQ t1.H8, v1.H8, t1.H8    \
	VUZP1 t1.B16, t0.B16, t0.B16 \
	VMOV  t0.D[0], Ra            \
	VMOV  t0.D[1], Rb            \
	CLZ   Ra, Ra                 \
	CLZ   Rb, Rb                 \
	ADD   Rb, Ra, Ra             \
	ADD   $32, p, Rc             \
	SUB   Ra>>2, Rc, p

// Turn the drop mask in V9 into the compaction shuffle sv and the kept count Rcnt. Clobbers R17, R19, R21.
#define EXTRACT(sv, Rcnt)            \
	VUZP1 V9.B16, V9.B16, V9.B16 \
	VMOV  V9.D[0], R17           \
	AND   R12, R17, R17          \
	MUL   R13, R17, R19          \
	LSR   $56, R19, R19          \
	MUL   R12, R17, R17          \
	LSR   $56, R17, R17          \
	SUB   R17, R20, Rcnt         \
	ADD   R19<<4, R6, R21        \
	VLD1  (R21), [sv.B16]

// Drop lanes equal to their predecessor (ep) or successor (ep shifted down, epNext feeding lane 7).
#define PREPDROP(ep, epNext, sv, Rcnt)       \
	VEXT $2, epNext.B16, ep.B16, V8.B16  \
	VORR V8.B16, ep.B16, V9.B16          \
	EXTRACT(sv, Rcnt)

// PREPDROP plus the lanes of cur above x: no unsigned greater-than mnemonic, so complement max(cur,x) == x.
#define PREPDROPX(cur, ep, epNext, sv, Rcnt)   \
	VEXT  $2, epNext.B16, ep.B16, V8.B16   \
	VORR  V8.B16, ep.B16, V9.B16           \
	VUMAX V6.H8, cur.H8, V10.H8            \
	VCMEQ V6.H8, V10.H8, V10.H8            \
	VEOR  V29.B16, V10.B16, V10.B16        \
	VORR  V10.B16, V9.B16, V9.B16          \
	EXTRACT(sv, Rcnt)

// func xorKernelNEON(set1, set2, buffer []uint16, shuf *byte) (outLen, pos1, pos2 int)
TEXT ·xorKernelNEON(SB), NOSPLIT, $0-104
	MOVD set1_base+0(FP), R0
	MOVD set2_base+24(FP), R1
	MOVD buffer_base+48(FP), R2
	MOVD set1_len+8(FP), R3
	MOVD set2_len+32(FP), R4
	MOVD shuf+72(FP), R6
	MOVD R2, R5

	// Loop ends are the last cursors that still leave 16 unread per side.
	ADD  R3<<1, R0, R3
	SUB  $32, R3, R3
	ADD  R4<<1, R1, R4
	SUB  $32, R4, R4

	MOVD $0x0101010101010101, R12
	MOVD $0x0102040810204080, R13
	MOVD $8, R20

	VMOVI $255, V29.B16
	VEOR  V31.B16, V31.B16, V31.B16
	VEXT  $14, V29.B16, V31.B16, V30.B16

loop:
	CMP R3, R0
	BHI done
	CMP R4, R1
	BHI done

	// A window wholly below the other side is copied as is; boundary equality
	// falls through to the general path.
	MOVHU 30(R0), R8
	MOVHU (R1), R9
	MOVHU 30(R1), R10
	MOVHU (R0), R11
	CMP   R9, R8
	BLO   copyA16
	CMP   R11, R10
	BLO   copyB16

	// Same gates at half a window, for ownership runs shorter than 16.
	MOVHU 14(R0), R16
	CMP   R9, R16
	BLO   copyA8
	MOVHU 14(R1), R17
	CMP   R11, R17
	BLO   copyB8

	VLD1 (R0), [V0.H8, V1.H8]
	VLD1 (R1), [V2.H8, V3.H8]

	// Identical windows cancel entirely; the head compare gates the test.
	CMP   R9, R11
	BNE   general
	VCMEQ V2.H8, V0.H8, V8.H8
	VCMEQ V3.H8, V1.H8, V9.H8
	VAND  V9.B16, V8.B16, V8.B16
	VMOV  V8.D[0], R14
	VMOV  V8.D[1], R15
	AND   R15, R14, R14
	CMN   $1, R14
	BEQ   annihilate

general:
	VDUP V1.H[7], V4.H8
	VDUP V3.H[7], V5.H8

	COUNTADV(V0, V1, V5, V8, V9, R8, R9, R10, R0)
	COUNTADV(V2, V3, V4, V10, V11, R14, R15, R16, R1)

	VUMIN V5.H8, V4.H8, V6.H8

	MERGE32

	// Predecessor equality on the sorted values; lane 0 has no predecessor
	// and everything retired earlier is strictly smaller, so it is cleared.
	VEXT  $14, V16.B16, V16.B16, V8.B16
	VCMEQ V8.H8, V16.H8, V0.H8
	VAND  V30.B16, V0.B16, V0.B16
	VEXT  $14, V17.B16, V16.B16, V8.B16
	VCMEQ V8.H8, V17.H8, V1.H8
	VEXT  $14, V18.B16, V17.B16, V8.B16
	VCMEQ V8.H8, V18.H8, V2.H8
	VEXT  $14, V19.B16, V18.B16, V8.B16
	VCMEQ V8.H8, V19.H8, V3.H8

	// T >= 16, so only the last two vectors can hold lanes above x.
	PREPDROP(V0, V1, V20, R22)
	PREPDROP(V1, V2, V21, R23)
	PREPDROPX(V18, V2, V3, V22, R24)
	PREPDROPX(V19, V3, V31, V23, R25)

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

annihilate:
	ADD $32, R0, R0
	ADD $32, R1, R1
	B   loop

copyA16:
	VLD1 (R0), [V0.H8, V1.H8]
	VST1 [V0.H8, V1.H8], (R2)
	ADD  $32, R0, R0
	ADD  $32, R2, R2
	B    loop

copyB16:
	VLD1 (R1), [V2.H8, V3.H8]
	VST1 [V2.H8, V3.H8], (R2)
	ADD  $32, R1, R1
	ADD  $32, R2, R2
	B    loop

copyA8:
	VLD1 (R0), [V0.H8]
	VST1 [V0.H8], (R2)
	ADD  $16, R0, R0
	ADD  $16, R2, R2
	B    loop

copyB8:
	VLD1 (R1), [V2.H8]
	VST1 [V2.H8], (R2)
	ADD  $16, R1, R1
	ADD  $16, R2, R2
	B    loop

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
