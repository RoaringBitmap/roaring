// +build amd64,!appengine

//go:build amd64 && !appengine

#include "textflag.h"

// AVX2 population-count routines using the Mula/Lemire VPSHUFB nibble-lookup
// kernel. Each 256-bit vector holds 4 uint64; the popcount of each byte is
// looked up via VPSHUFB over a 16-entry nibble table (replicated in both
// 128-bit lanes), the two nibble counts are added byte-wise, then VPSADBW sums
// each group of 8 bytes into a quadword that is accumulated. A scalar POPCNTQ
// tail handles the remaining (len % 4) words, so the routines are correct for
// any length. Loads/stores are unaligned (VMOVDQU) because the bitmap slices
// are only 8-byte aligned. VZEROUPPER precedes every return to avoid the
// AVX<->SSE transition penalty.

// 16-entry nibble popcount table, replicated across both 128-bit lanes.
DATA lutmask<>+0(SB)/8, $0x0302020102010100
DATA lutmask<>+8(SB)/8, $0x0403030203020201
DATA lutmask<>+16(SB)/8, $0x0302020102010100
DATA lutmask<>+24(SB)/8, $0x0403030203020201
// low-nibble mask, 0x0F in every byte.
DATA lutmask<>+32(SB)/8, $0x0f0f0f0f0f0f0f0f
DATA lutmask<>+40(SB)/8, $0x0f0f0f0f0f0f0f0f
DATA lutmask<>+48(SB)/8, $0x0f0f0f0f0f0f0f0f
DATA lutmask<>+56(SB)/8, $0x0f0f0f0f0f0f0f0f
GLOBL lutmask<>(SB), RODATA|NOPTR, $64

#define Ylut Y0
#define Ymask Y1
#define Yzero Y2
#define Yacc Y3
#define Ydata Y4
#define Yb Y5
#define Ylo Y6
#define Yhi Y7
#define Yc1 Y8
#define Yc2 Y9

// COUNTBLOCK adds the per-byte popcount of Ydata into the quadword accumulator
// Yacc, using the lookup table Ylut, the nibble mask Ymask and Yzero.
#define COUNTBLOCK \
	VPAND Ymask, Ydata, Ylo \
	VPSRLW $4, Ydata, Yhi \
	VPAND Ymask, Yhi, Yhi \
	VPSHUFB Ylo, Ylut, Yc1 \
	VPSHUFB Yhi, Ylut, Yc2 \
	VPADDB Yc2, Yc1, Yc1 \
	VPSADBW Yzero, Yc1, Yc1 \
	VPADDQ Yc1, Yacc, Yacc

// SETUP loads the constant table/mask and zeroes the accumulator registers.
#define SETUP \
	VMOVDQU lutmask<>+0(SB), Ylut \
	VMOVDQU lutmask<>+32(SB), Ymask \
	VPXOR Yzero, Yzero, Yzero \
	VPXOR Yacc, Yacc, Yacc

// HSUM horizontally sums the quadwords of Yacc into AX.
#define HSUM \
	VEXTRACTI128 $1, Yacc, X5 \
	VPADDQ X5, X3, X3 \
	VPEXTRQ $1, X3, DX \
	MOVQ X3, R9 \
	ADDQ R9, AX \
	ADDQ DX, AX

// func _popcntSliceAVX2(s []uint64) uint64
TEXT ·_popcntSliceAVX2(SB), NOSPLIT, $0-32
	MOVQ s_base+0(FP), SI
	MOVQ s_len+8(FP), CX
	XORQ AX, AX
	SETUP
	MOVQ CX, R8
	SHRQ $2, R8
	TESTQ R8, R8
	JZ slicetail
sliceloop:
	VMOVDQU (SI), Ydata
	COUNTBLOCK
	ADDQ $32, SI
	DECQ R8
	JNZ sliceloop
	HSUM
slicetail:
	ANDQ $3, CX
	TESTQ CX, CX
	JZ slicedone
slicetailloop:
	MOVQ (SI), DX
	POPCNTQ DX, DX
	ADDQ DX, AX
	ADDQ $8, SI
	DECQ CX
	JNZ slicetailloop
slicedone:
	VZEROUPPER
	MOVQ AX, ret+24(FP)
	RET

// func _popcntAndSliceAVX2(s, m []uint64) uint64
TEXT ·_popcntAndSliceAVX2(SB), NOSPLIT, $0-56
	MOVQ s_base+0(FP), SI
	MOVQ m_base+24(FP), DI
	MOVQ s_len+8(FP), CX
	XORQ AX, AX
	SETUP
	MOVQ CX, R8
	SHRQ $2, R8
	TESTQ R8, R8
	JZ andtail
andloop:
	VMOVDQU (SI), Ydata
	VMOVDQU (DI), Yb
	VPAND Yb, Ydata, Ydata
	COUNTBLOCK
	ADDQ $32, SI
	ADDQ $32, DI
	DECQ R8
	JNZ andloop
	HSUM
andtail:
	ANDQ $3, CX
	TESTQ CX, CX
	JZ anddone
andtailloop:
	MOVQ (SI), DX
	ANDQ (DI), DX
	POPCNTQ DX, DX
	ADDQ DX, AX
	ADDQ $8, SI
	ADDQ $8, DI
	DECQ CX
	JNZ andtailloop
anddone:
	VZEROUPPER
	MOVQ AX, ret+48(FP)
	RET

// func _popcntOrSliceAVX2(s, m []uint64) uint64
TEXT ·_popcntOrSliceAVX2(SB), NOSPLIT, $0-56
	MOVQ s_base+0(FP), SI
	MOVQ m_base+24(FP), DI
	MOVQ s_len+8(FP), CX
	XORQ AX, AX
	SETUP
	MOVQ CX, R8
	SHRQ $2, R8
	TESTQ R8, R8
	JZ ortail
orloop:
	VMOVDQU (SI), Ydata
	VMOVDQU (DI), Yb
	VPOR Yb, Ydata, Ydata
	COUNTBLOCK
	ADDQ $32, SI
	ADDQ $32, DI
	DECQ R8
	JNZ orloop
	HSUM
ortail:
	ANDQ $3, CX
	TESTQ CX, CX
	JZ ordone
ortailloop:
	MOVQ (SI), DX
	ORQ (DI), DX
	POPCNTQ DX, DX
	ADDQ DX, AX
	ADDQ $8, SI
	ADDQ $8, DI
	DECQ CX
	JNZ ortailloop
ordone:
	VZEROUPPER
	MOVQ AX, ret+48(FP)
	RET

// func _popcntXorSliceAVX2(s, m []uint64) uint64
TEXT ·_popcntXorSliceAVX2(SB), NOSPLIT, $0-56
	MOVQ s_base+0(FP), SI
	MOVQ m_base+24(FP), DI
	MOVQ s_len+8(FP), CX
	XORQ AX, AX
	SETUP
	MOVQ CX, R8
	SHRQ $2, R8
	TESTQ R8, R8
	JZ xortail
xorloop:
	VMOVDQU (SI), Ydata
	VMOVDQU (DI), Yb
	VPXOR Yb, Ydata, Ydata
	COUNTBLOCK
	ADDQ $32, SI
	ADDQ $32, DI
	DECQ R8
	JNZ xorloop
	HSUM
xortail:
	ANDQ $3, CX
	TESTQ CX, CX
	JZ xordone
xortailloop:
	MOVQ (SI), DX
	XORQ (DI), DX
	POPCNTQ DX, DX
	ADDQ DX, AX
	ADDQ $8, SI
	ADDQ $8, DI
	DECQ CX
	JNZ xortailloop
xordone:
	VZEROUPPER
	MOVQ AX, ret+48(FP)
	RET

// func _popcntMaskSliceAVX2(s, m []uint64) uint64
// Computes sum of popcount(s[i] &^ m[i]) == popcount(s & ~m).
TEXT ·_popcntMaskSliceAVX2(SB), NOSPLIT, $0-56
	MOVQ s_base+0(FP), SI
	MOVQ m_base+24(FP), DI
	MOVQ s_len+8(FP), CX
	XORQ AX, AX
	SETUP
	MOVQ CX, R8
	SHRQ $2, R8
	TESTQ R8, R8
	JZ masktail
maskloop:
	VMOVDQU (SI), Ydata
	VMOVDQU (DI), Yb
	VPANDN Ydata, Yb, Ydata
	COUNTBLOCK
	ADDQ $32, SI
	ADDQ $32, DI
	DECQ R8
	JNZ maskloop
	HSUM
masktail:
	ANDQ $3, CX
	TESTQ CX, CX
	JZ maskdone
masktailloop:
	MOVQ (DI), R10
	NOTQ R10
	MOVQ (SI), DX
	ANDQ R10, DX
	POPCNTQ DX, DX
	ADDQ DX, AX
	ADDQ $8, SI
	ADDQ $8, DI
	DECQ CX
	JNZ masktailloop
maskdone:
	VZEROUPPER
	MOVQ AX, ret+48(FP)
	RET

// func _hasAVX2() bool
// Reports whether the CPU supports AVX2 and the OS has enabled YMM state.
TEXT ·_hasAVX2(SB), NOSPLIT, $0-1
	// CPUID leaf 1: require OSXSAVE (ECX bit 27) and AVX (ECX bit 28).
	MOVL $1, AX
	XORL CX, CX
	CPUID
	ANDL $0x18000000, CX
	CMPL CX, $0x18000000
	JNE noavx2
	// XGETBV(0): require XCR0 bits 1 (SSE) and 2 (AVX/YMM) set.
	XORL CX, CX
	XGETBV
	ANDL $0x6, AX
	CMPL AX, $0x6
	JNE noavx2
	// CPUID leaf 7, sub-leaf 0: require AVX2 (EBX bit 5).
	MOVL $7, AX
	XORL CX, CX
	CPUID
	ANDL $0x20, BX
	CMPL BX, $0x20
	JNE noavx2
	MOVB $1, ret+0(FP)
	RET
noavx2:
	MOVB $0, ret+0(FP)
	RET
