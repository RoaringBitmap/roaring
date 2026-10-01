// +build arm64,!gccgo,!appengine

#include "textflag.h"

// Deferred emission: V0's match masks accumulate in V2 and survivors are
// stored only when the set1 block retires; a lane unmatched so far may
// still match a later set2 block.
// Full stores are safe under aliasing: cap(buffer) >= len(set1), out <=
// pos1 at retirement, and each whole block is loaded before it is stored.
//
// Registers:
//   R0/R1 input cursors, R2 out cursor, R3/R4 full-block ends, R5 out
//   base, R6 shuffle table, R19 set1 base, R8-R11 block boundary values,
//   R12/R13 constants, R14-R17 scratch.
//   V0/V1 input blocks, V2 accumulated match mask, V3-V7 temps.

// Match mask of V0's lanes against all rotations of V1, result in V3.
#define MATCH8 \
	VCMEQ V1.H8, V0.H8, V3.H8          \
	VEXT  $2, V1.B16, V1.B16, V4.B16   \
	VCMEQ V4.H8, V0.H8, V4.H8          \
	VEXT  $4, V1.B16, V1.B16, V5.B16   \
	VCMEQ V5.H8, V0.H8, V5.H8          \
	VEXT  $6, V1.B16, V1.B16, V6.B16   \
	VCMEQ V6.H8, V0.H8, V6.H8          \
	VORR  V4.B16, V3.B16, V3.B16       \
	VORR  V6.B16, V5.B16, V5.B16       \
	VEXT  $8, V1.B16, V1.B16, V4.B16   \
	VCMEQ V4.H8, V0.H8, V4.H8          \
	VEXT  $10, V1.B16, V1.B16, V6.B16  \
	VCMEQ V6.H8, V0.H8, V6.H8          \
	VORR  V6.B16, V4.B16, V4.B16       \
	VEXT  $12, V1.B16, V1.B16, V6.B16  \
	VCMEQ V6.H8, V0.H8, V6.H8          \
	VEXT  $14, V1.B16, V1.B16, V7.B16  \
	VCMEQ V7.H8, V0.H8, V7.H8          \
	VORR  V7.B16, V6.B16, V6.B16       \
	VORR  V5.B16, V3.B16, V3.B16       \
	VORR  V6.B16, V4.B16, V4.B16       \
	VORR  V4.B16, V3.B16, V3.B16

// Retire V0: uniqshuf's clear-lane polarity compacts the unmatched lanes.
#define EMITBLOCK \
	VUZP1 V2.B16, V2.B16, V4.B16       \
	VMOV  V4.D[0], R14                 \
	AND   R12, R14, R14                \
	MUL   R12, R14, R16                \
	LSR   $56, R16, R16                \
	MUL   R13, R14, R15                \
	LSR   $56, R15, R15                \
	ADD   R15<<4, R6, R17              \
	VLD1  (R17), [V4.B16]              \
	VTBL  V4.B16, [V0.B16], V4.B16     \
	VST1  [V4.B16], (R2)               \
	ADD   $16, R2                      \
	SUB   R16<<1, R2, R2               \
	VEOR  V2.B16, V2.B16, V2.B16

// func andnotKernelNEON(set1, set2, buffer []uint16, shuf *byte, spill *[8]uint16) (outLen, pos1, pos2, spilled, accMask int)
TEXT ·andnotKernelNEON(SB), NOSPLIT, $0-128
	MOVD set1_base+0(FP), R0
	MOVD set2_base+24(FP), R1
	MOVD buffer_base+48(FP), R2
	MOVD set1_len+8(FP), R3
	MOVD set2_len+32(FP), R4
	MOVD shuf+72(FP), R6

	MOVD R0, R19
	MOVD R2, R5

	AND $~7, R3, R3
	AND $~7, R4, R4
	ADD R3<<1, R0, R3
	ADD R4<<1, R1, R4

	MOVD $0x0101010101010101, R12
	MOVD $0x0102040810204080, R13

	VEOR  V2.B16, V2.B16, V2.B16
	VLD1  (R0), [V0.H8]
	VLD1  (R1), [V1.H8]
	MOVHU (R0), R8
	MOVHU 14(R0), R9
	MOVHU (R1), R10
	MOVHU 14(R1), R11

nloop:
	// Range gates use strict <: equal boundary values may match.
	CMP  R10, R9
	BLO  nffA
	CMP  R8, R11
	BLO  nffB

	MATCH8
	VORR V3.B16, V2.B16, V2.B16

	CMP R11, R9
	BHI nadvB
	BLO nadvA
	EMITBLOCK
	ADD $16, R0
	ADD $16, R1
	CMP R3, R0
	BHS ndoneNoSpill
	CMP R4, R1
	BHS ndoneNoSpill
	VLD1  (R0), [V0.H8]
	MOVHU (R0), R8
	MOVHU 14(R0), R9
	VLD1  (R1), [V1.H8]
	MOVHU (R1), R10
	MOVHU 14(R1), R11
	B nloop

nadvA:
	EMITBLOCK
	ADD $16, R0
	CMP R3, R0
	BHS ntailA
	VLD1  (R0), [V0.H8]
	MOVHU (R0), R8
	MOVHU 14(R0), R9
	B nloop

nadvB:
	ADD $16, R1
	CMP R4, R1
	BHS ndoneSpill
	VLD1  (R1), [V1.H8]
	MOVHU (R1), R10
	MOVHU 14(R1), R11
	B nloop

nffA:
	// Nothing left in set2 can match blocks wholly below it: emit them whole.
	EMITBLOCK
nffAskip:
	ADD $16, R0
	CMP R3, R0
	BHS ndoneNoSpill
	MOVHU 14(R0), R9
	CMP   R10, R9
	BHS   nffAload
	VLD1  (R0), [V0.H8]
	VST1  [V0.H8], (R2)
	ADD   $16, R2
	B     nffAskip

nffAload:
	VLD1  (R0), [V0.H8]
	MOVHU (R0), R8
	B nloop

nffB:
	ADD $16, R1
	CMP R4, R1
	BHS ndoneSpill
	MOVHU 14(R1), R11
	CMP   R8, R11
	BLO   nffB
	VLD1  (R1), [V1.H8]
	MOVHU (R1), R10
	B nloop

ndoneSpill:
	MOVD  spill+80(FP), R16
	VST1  [V0.H8], (R16)
	VUZP1 V2.B16, V2.B16, V4.B16
	VMOV  V4.D[0], R14
	AND   R12, R14, R14
	MUL   R13, R14, R15
	LSR   $56, R15, R15
	MOVD  R15, accMask+120(FP)
	MOVD  $1, R16
	MOVD  R16, spilled+112(FP)
	ADD   $16, R0
	B     nout

ndoneNoSpill:
	MOVD ZR, spilled+112(FP)
	MOVD ZR, accMask+120(FP)

nout:
	SUB  R5, R2, R2
	LSR  $1, R2, R2
	MOVD R2, outLen+88(FP)
	SUB  R19, R0, R0
	LSR  $1, R0, R0
	MOVD R0, pos1+96(FP)
	MOVD set2_base+24(FP), R8
	SUB  R8, R1, R1
	LSR  $1, R1, R1
	MOVD R1, pos2+104(FP)
	RET

ntailA:
	// set2 lanes at or below the retired set1 maximum cannot match the
	// scalar tail; skipping them avoids rescanning up to seven values.
	VDUP  R9, V4.H8
	VUMIN V4.H8, V1.H8, V5.H8
	VCMEQ V5.H8, V1.H8, V5.H8
	VUZP1 V5.B16, V5.B16, V5.B16
	VMOV  V5.D[0], R14
	AND   R12, R14, R14
	MUL   R12, R14, R14
	LSR   $56, R14, R14
	ADD   R14<<1, R1, R1
	B     ndoneNoSpill
