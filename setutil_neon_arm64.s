// +build arm64,!gccgo,!appengine

#include "textflag.h"

// NEON-vectorized main loop for union2by2 (issue #288).
// Port of CRoaring's union_vector16 with a transposed-bitonic merge network
// and a branchless main-loop block select. Processes full 8-element blocks from both
// inputs while both have blocks remaining; the Go wrapper handles tails.
//
// Register use:
//   R0/R1 cursors into set1/set2, R3/R4 their full-block end addresses,
//   R2 output cursor, R5 output base, R6 uniqshuf table, R7 leftover ptr,
//   R12 = 0x0101010101010101, R13 = 0x0102040810204080, R19 = 8.
//   V0 carry (8 largest so far), V1 last-stored vector, V2 fresh/min.

// Merge sorted V2 (fresh) with sorted V0 (carry):
// V2 <- 8 smallest (sorted), V0 <- 8 largest (sorted). Clobbers V3-V7.
// Transposed bitonic 16-merge. The reversal is applied to the fresh input
// rather than the carry so it stays off the loop-carried dependency chain.
#define MERGE \
	VREV64 V2.H8, V3.H8                \
	VEXT   $8, V3.B16, V3.B16, V3.B16  \
	VUMIN  V3.H8, V0.H8, V4.H8         \
	VUMAX  V3.H8, V0.H8, V5.H8         \
	VTRN1  V5.D2, V4.D2, V6.D2         \
	VTRN2  V5.D2, V4.D2, V7.D2         \
	VUMIN  V7.H8, V6.H8, V4.H8         \
	VUMAX  V7.H8, V6.H8, V5.H8         \
	VTRN1  V5.S4, V4.S4, V6.S4         \
	VTRN2  V5.S4, V4.S4, V7.S4         \
	VUMIN  V7.H8, V6.H8, V4.H8         \
	VUMAX  V7.H8, V6.H8, V5.H8         \
	VTRN1  V5.H8, V4.H8, V6.H8         \
	VTRN2  V5.H8, V4.H8, V7.H8         \
	VUMIN  V7.H8, V6.H8, V4.H8         \
	VUMAX  V7.H8, V6.H8, V5.H8         \
	VZIP1  V5.H8, V4.H8, V2.H8         \
	VZIP2  V5.H8, V4.H8, V0.H8

// Store the lanes of Vin that differ from their predecessor (previous lane,
// or last lane of V1 for lane 0) at (Rout), advancing Rout by 2 bytes per
// lane kept. Writes a full 16 bytes; the caller guarantees slack.
// Rcnt receives the number of lanes kept. Clobbers V3-V5, R14, R15, R16.
#define STOREUNIQ(Vin, Rout, Rcnt) \
	VEXT   $14, Vin.B16, V1.B16, V3.B16 \
	VCMEQ  V3.H8, Vin.H8, V4.H8         \
	VUZP1  V4.B16, V4.B16, V4.B16      \
	VMOV   V4.D[0], R14                \
	AND    R12, R14, R14               \
	MUL    R13, R14, R15               \
	LSR    $56, R15, R15               \
	MUL    R12, R14, R14               \
	LSR    $56, R14, R14               \
	SUB    R14, R19, Rcnt              \
	ADD    R15<<4, R6, R16             \
	VLD1   (R16), [V5.B16]             \
	VTBL   V5.B16, [Vin.B16], V3.B16   \
	VST1   [V3.B16], (Rout)            \
	ADD    Rcnt<<1, Rout, Rout

// func unionKernelNEON(set1, set2, buffer []uint16, shuf *byte, leftover *[16]uint16) (outLen, pos1, pos2, leftoverLen int)
TEXT ·unionKernelNEON(SB), NOSPLIT, $0-120
	MOVD set1_base+0(FP), R0
	MOVD set2_base+24(FP), R1
	MOVD buffer_base+48(FP), R2
	MOVD set1_len+8(FP), R3
	MOVD set2_len+32(FP), R4
	MOVD shuf+72(FP), R6
	MOVD leftover+80(FP), R7

	// end pointers over whole blocks: base + (len &^ 7) * 2
	AND $~7, R3, R3
	AND $~7, R4, R4
	ADD R3<<1, R0, R3
	ADD R4<<1, R1, R4

	MOVD R2, R5

	MOVD $0x0101010101010101, R12
	MOVD $0x0102040810204080, R13
	MOVD $8, R19

	// prime: merge first block of each input
	VLD1.P 16(R0), [V2.H8]
	VLD1.P 16(R1), [V0.H8]
	MERGE
	// laststore = all ones (never equal to a first stored lane)
	VCMEQ V0.H8, V0.H8, V1.H8
	STOREUNIQ(V2, R2, R17)
	VORR V2.B16, V2.B16, V1.B16

loop:
	CMP R3, R0
	BHS done
	CMP R4, R1
	BHS done
	MOVHU (R0), R8
	MOVHU (R1), R9
	CMP  R9, R8
	CSEL LS, R0, R1, R10   // take set1's block on tie/less
	CSEL LS, R8, R9, R20
	VLD1 (R10), [V2.H8]
	CSET LS, R11
	ADD  R11<<4, R0, R0
	EOR  $1, R11, R11
	ADD  R11<<4, R1, R1
	// Disjoint fast path: when head >= carry's max the merged result is
	// just carry then fresh, so skip the network. Boundary equality falls
	// to the dedup chain, hence >= rather than >.
	VMOV V0.H[7], R21
	CMP  R21, R20
	BHS  disjoint
	MERGE
	STOREUNIQ(V2, R2, R17)
	VORR V2.B16, V2.B16, V1.B16
	B loop

// Fast-forward loop for consecutive disjoint blocks. The head select is a
// plain branch: on run-structured data it predicts, keeping the untaken
// cursor off the loop-carried chain that the main loop's CSEL select
// serializes. Interleaving input rejoins the main loop after one merge.
disjoint:
	STOREUNIQ(V0, R2, R17)
	VORR V0.B16, V0.B16, V1.B16    // laststore = old carry
	VORR V2.B16, V2.B16, V0.B16    // carry = fresh

ffloop:
	CMP R3, R0
	BHS done
	CMP R4, R1
	BHS done
	MOVHU (R0), R8
	MOVHU (R1), R9
	CMP  R9, R8
	BHI  fftake2
	VLD1.P 16(R0), [V2.H8]         // take set1's block on tie/less
	MOVD R8, R20
	B    ffcheck
fftake2:
	VLD1.P 16(R1), [V2.H8]
	MOVD R9, R20
ffcheck:
	VMOV V0.H[7], R21
	CMP  R21, R20
	BLO  ffmerge
	STOREUNIQ(V0, R2, R17)
	VORR V0.B16, V0.B16, V1.B16    // laststore = old carry
	VORR V2.B16, V2.B16, V0.B16    // carry = fresh
	B ffloop

ffmerge:
	MERGE
	STOREUNIQ(V2, R2, R17)
	VORR V2.B16, V2.B16, V1.B16
	B loop

done:
	// flush carry through the dedup into the leftover buffer
	STOREUNIQ(V0, R7, R17)
	MOVD R17, leftoverLen+112(FP)

	// outLen = (out - outBase) / 2
	SUB  R5, R2, R2
	LSR  $1, R2, R2
	MOVD R2, outLen+88(FP)

	// pos1/pos2 = blocks consumed
	MOVD set1_base+0(FP), R8
	SUB  R8, R0, R0
	LSR  $4, R0, R0
	MOVD R0, pos1+96(FP)
	MOVD set2_base+24(FP), R8
	SUB  R8, R1, R1
	LSR  $4, R1, R1
	MOVD R1, pos2+104(FP)
	RET
