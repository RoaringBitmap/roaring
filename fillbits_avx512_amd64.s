//go:build amd64 && !appengine
// +build amd64,!appengine

#include "textflag.h"

// The vector contains the 16 possible bit positions in one 16-bit input
// block. VPCOMPRESSD writes only the lanes selected by the input mask, so its
// store is exactly the variable-sized output for that block.
DATA fillLeastSignificant16bitsOffsets<>+0(SB)/4, $0
DATA fillLeastSignificant16bitsOffsets<>+4(SB)/4, $1
DATA fillLeastSignificant16bitsOffsets<>+8(SB)/4, $2
DATA fillLeastSignificant16bitsOffsets<>+12(SB)/4, $3
DATA fillLeastSignificant16bitsOffsets<>+16(SB)/4, $4
DATA fillLeastSignificant16bitsOffsets<>+20(SB)/4, $5
DATA fillLeastSignificant16bitsOffsets<>+24(SB)/4, $6
DATA fillLeastSignificant16bitsOffsets<>+28(SB)/4, $7
DATA fillLeastSignificant16bitsOffsets<>+32(SB)/4, $8
DATA fillLeastSignificant16bitsOffsets<>+36(SB)/4, $9
DATA fillLeastSignificant16bitsOffsets<>+40(SB)/4, $10
DATA fillLeastSignificant16bitsOffsets<>+44(SB)/4, $11
DATA fillLeastSignificant16bitsOffsets<>+48(SB)/4, $12
DATA fillLeastSignificant16bitsOffsets<>+52(SB)/4, $13
DATA fillLeastSignificant16bitsOffsets<>+56(SB)/4, $14
DATA fillLeastSignificant16bitsOffsets<>+60(SB)/4, $15
GLOBL fillLeastSignificant16bitsOffsets<>(SB), RODATA|NOPTR, $64

// func fillLeastSignificant16bitsAVX512(bitmap []uint64, x []uint32, pos int, mask uint32) int
TEXT ·fillLeastSignificant16bitsAVX512(SB), NOSPLIT, $0-72
	MOVQ bitmap_base+0(FP), SI
	MOVQ bitmap_len+8(FP), CX
	MOVQ x_base+24(FP), DI
	MOVQ pos+48(FP), R8
	LEAQ (DI)(R8*4), DI
	MOVL mask+56(FP), R9
	VMOVDQU64 fillLeastSignificant16bitsOffsets<>(SB), Z0

	TESTQ CX, CX
	JZ fillLeastSignificant16bitsAVX512Done
fillLeastSignificant16bitsAVX512Loop:
	MOVQ (SI), AX

	MOVWQZX AX, R11
	KMOVW R11, K1
	VPBROADCASTD R9, Z1
	VPADDD Z1, Z0, Z2
	VPCOMPRESSD Z2, K1, (DI)
	POPCNTQ R11, R11
	ADDQ R11, R8
	SHLQ $2, R11
	ADDQ R11, DI
	ADDQ $16, R9
	SHRQ $16, AX

	MOVWQZX AX, R11
	KMOVW R11, K1
	VPBROADCASTD R9, Z1
	VPADDD Z1, Z0, Z2
	VPCOMPRESSD Z2, K1, (DI)
	POPCNTQ R11, R11
	ADDQ R11, R8
	SHLQ $2, R11
	ADDQ R11, DI
	ADDQ $16, R9
	SHRQ $16, AX

	MOVWQZX AX, R11
	KMOVW R11, K1
	VPBROADCASTD R9, Z1
	VPADDD Z1, Z0, Z2
	VPCOMPRESSD Z2, K1, (DI)
	POPCNTQ R11, R11
	ADDQ R11, R8
	SHLQ $2, R11
	ADDQ R11, DI
	ADDQ $16, R9
	SHRQ $16, AX

	MOVWQZX AX, R11
	KMOVW R11, K1
	VPBROADCASTD R9, Z1
	VPADDD Z1, Z0, Z2
	VPCOMPRESSD Z2, K1, (DI)
	POPCNTQ R11, R11
	ADDQ R11, R8
	SHLQ $2, R11
	ADDQ R11, DI
	ADDQ $16, R9

	ADDQ $8, SI
	DECQ CX
	JNZ fillLeastSignificant16bitsAVX512Loop
fillLeastSignificant16bitsAVX512Done:
	VZEROUPPER
	MOVQ R8, ret+64(FP)
	RET

// func _hasAVX512() bool
// Reports whether AVX-512 Foundation is available and the OS has enabled the
// complete state needed by ZMM and opmask registers. The check is done once at
// package initialization, so unsupported machines never execute the vector
// decoder.
TEXT ·_hasAVX512(SB), NOSPLIT, $0-1
	// CPUID leaf 1: require OSXSAVE (ECX bit 27) and AVX (ECX bit 28).
	MOVL $1, AX
	XORL CX, CX
	CPUID
	NOTL CX
	TESTL $0x18000000, CX
	JNE noavx512

	// XCR0: bits 1, 2, 5, 6, and 7 enable the SSE, YMM, opmask, ZMM_hi256,
	// and Hi16_ZMM state that AVX-512 instructions use.
	XORL CX, CX
	XGETBV
	NOTL AX
	TESTL $0xe6, AX
	JNE noavx512

	// CPUID leaf 7, sub-leaf 0: AVX-512 Foundation is EBX bit 16.
	MOVL $7, AX
	XORL CX, CX
	CPUID
	NOTL BX
	TESTL $0x10000, BX
noavx512:
	SETEQ ret+0(FP)
	RET
