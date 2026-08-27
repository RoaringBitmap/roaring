//go:build amd64 && !appengine
// +build amd64,!appengine

#include "textflag.h"

// AVX2 implementation of the plain bitmap-container XOR operation.
//
// It computes dst[i] = a[i] ^ b[i] with four YMM registers per unrolled
// iteration. The destination may alias either source, as each source block is
// loaded before its corresponding stores. A vector tail and scalar tail keep
// the helper correct for arbitrary slice lengths.
//
// A []uint64 argument is a {ptr,len,cap} header. The second and third slices
// therefore start at +24(FP) and +48(FP).

// func xorSliceAVX2(dst, a, b []uint64)
TEXT ·xorSliceAVX2(SB), NOSPLIT, $0-72
	MOVQ dst_base+0(FP), DX
	MOVQ a_base+24(FP), SI
	MOVQ b_base+48(FP), DI
	MOVQ a_len+32(FP), BX
	MOVQ BX, CX
	SHRQ $4, CX // 16 words per unrolled iteration
	TESTQ CX, CX
	JZ xorsliceavx2tail

xorsliceavx2loop:
	VMOVDQU 0(SI), Y0
	VMOVDQU 32(SI), Y1
	VMOVDQU 64(SI), Y2
	VMOVDQU 96(SI), Y3
	VPXOR 0(DI), Y0, Y0
	VPXOR 32(DI), Y1, Y1
	VPXOR 64(DI), Y2, Y2
	VPXOR 96(DI), Y3, Y3
	VMOVDQU Y0, 0(DX)
	VMOVDQU Y1, 32(DX)
	VMOVDQU Y2, 64(DX)
	VMOVDQU Y3, 96(DX)
	ADDQ $128, SI
	ADDQ $128, DI
	ADDQ $128, DX
	DECQ CX
	JNZ xorsliceavx2loop

xorsliceavx2tail:
	ANDQ $15, BX
	MOVQ BX, CX
	SHRQ $2, CX // four words per vector tail iteration
	TESTQ CX, CX
	JZ xorsliceavx2scalartail

xorsliceavx2vectorloop:
	VMOVDQU (SI), Y0
	VPXOR (DI), Y0, Y0
	VMOVDQU Y0, (DX)
	ADDQ $32, SI
	ADDQ $32, DI
	ADDQ $32, DX
	DECQ CX
	JNZ xorsliceavx2vectorloop

xorsliceavx2scalartail:
	ANDQ $3, BX
	TESTQ BX, BX
	JZ xorsliceavx2done

xorsliceavx2scalarloop:
	MOVQ (SI), AX
	XORQ (DI), AX
	MOVQ AX, (DX)
	ADDQ $8, SI
	ADDQ $8, DI
	ADDQ $8, DX
	DECQ BX
	JNZ xorsliceavx2scalarloop

xorsliceavx2done:
	VZEROUPPER
	RET
