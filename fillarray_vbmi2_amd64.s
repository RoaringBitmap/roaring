//go:build amd64 && !appengine
// +build amd64,!appengine

#include "textflag.h"

// Decoding a bitmap container into a uint16 array container.
//
// These share the shape of fillbits_vbmi2_amd64.s: one VPCOMPRESSB per 64-bit
// word turns the whole word into 64 byte-sized bit positions, which are then
// widened 32 at a time with VPMOVZXBW and written with a masked store, so
// exactly popcount(word) uint16s are written and the output needs no slack.
// The compress-then-widen idea is taken from simdjson's bit_indexer::write
// (icelake kernel, Apache-2.0): https://github.com/simdjson/simdjson
//
// The AND, ANDNOT and XOR variants combine two containers word by word before
// emitting, which is what bitmapContainer.and/andNot/xor need when their result
// is small enough to become an array container.
//
// fillArraySkipVector additionally tests eight words at a time with VPTESTMQ
// and skips wholly empty groups. It wins only when the container is very
// sparse, where most 512-bit groups hold nothing at all.

DATA fa_bitPositions<>+0x00(SB)/8, $0x0706050403020100
DATA fa_bitPositions<>+0x08(SB)/8, $0x0f0e0d0c0b0a0908
DATA fa_bitPositions<>+0x10(SB)/8, $0x1716151413121110
DATA fa_bitPositions<>+0x18(SB)/8, $0x1f1e1d1c1b1a1918
DATA fa_bitPositions<>+0x20(SB)/8, $0x2726252423222120
DATA fa_bitPositions<>+0x28(SB)/8, $0x2f2e2d2c2b2a2928
DATA fa_bitPositions<>+0x30(SB)/8, $0x3736353433323130
DATA fa_bitPositions<>+0x38(SB)/8, $0x3f3e3d3c3b3a3938
GLOBL fa_bitPositions<>(SB), RODATA|NOPTR, $64

// Emit the set bits of AX as uint16 values at (DI), advancing DI.
// Z0 = byte positions, Z1 = running base (16-bit lanes). Clobbers R10, R11,
// K1, K2, Z2, Z4, Y5.
#define EMIT16(skip, adv)          \
	TESTQ AX, AX             \
	JZ    skip               \
	KMOVQ AX, K1             \
	VPCOMPRESSB.Z Z0, K1, Z2 \
	POPCNTQ AX, R10          \
	MOVQ $-1, R11            \
	BZHIQ R10, R11, R11      \
	KMOVD R11, K2            \
	VPMOVZXBW Y2, Z4         \
	VPADDW Z1, Z4, Z4        \
	VMOVDQU16 Z4, K2, (DI)   \
	CMPQ R10, $32            \
	JLE  adv          \
	SHRQ $32, R11            \
	KMOVD R11, K2            \
	VEXTRACTI64X4 $1, Z2, Y5 \
	VPMOVZXBW Y5, Z4         \
	VPADDW Z1, Z4, Z4        \
	VMOVDQU16 Z4, K2, 64(DI) \
adv:                         \
	LEAQ (DI)(R10*2), DI     \
skip:                        \
	VPADDW Z3, Z1, Z1

#define SETUP16              \
	VMOVDQU64 fa_bitPositions<>(SB), Z0 \
	VPXORQ Z1, Z1, Z1        \
	MOVL $64, AX             \
	VPBROADCASTW AX, Z3

// func fillArrayVector(bitmap []uint64, container []uint16)
TEXT ·fillArrayVector(SB), NOSPLIT, $0-48
	MOVQ bitmap_base+0(FP), SI
	MOVQ bitmap_len+8(FP), CX
	MOVQ container_base+24(FP), DI
	SETUP16
	TESTQ CX, CX
	JZ    fa_done
fa_loop:
	MOVQ (SI), AX
	EMIT16(fa_s, fa_a)
	ADDQ $8, SI
	DECQ CX
	JNZ  fa_loop
fa_done:
	VZEROUPPER
	RET

// func fillArrayANDVector(container []uint16, bitmap1, bitmap2 []uint64)
TEXT ·fillArrayANDVector(SB), NOSPLIT, $0-72
	MOVQ container_base+0(FP), DI
	MOVQ bitmap1_base+24(FP), SI
	MOVQ bitmap1_len+32(FP), CX
	MOVQ bitmap2_base+48(FP), BX
	SETUP16
	TESTQ CX, CX
	JZ    fand_done
fand_loop:
	MOVQ (SI), AX
	ANDQ (BX), AX
	EMIT16(fand_s, fand_a)
	ADDQ $8, SI
	ADDQ $8, BX
	DECQ CX
	JNZ  fand_loop
fand_done:
	VZEROUPPER
	RET

// func fillArrayANDNOTVector(container []uint16, bitmap1, bitmap2 []uint64)
TEXT ·fillArrayANDNOTVector(SB), NOSPLIT, $0-72
	MOVQ container_base+0(FP), DI
	MOVQ bitmap1_base+24(FP), SI
	MOVQ bitmap1_len+32(FP), CX
	MOVQ bitmap2_base+48(FP), BX
	SETUP16
	TESTQ CX, CX
	JZ    fandn_done
fandn_loop:
	MOVQ (BX), AX
	NOTQ AX
	ANDQ (SI), AX
	EMIT16(fandn_s, fandn_a)
	ADDQ $8, SI
	ADDQ $8, BX
	DECQ CX
	JNZ  fandn_loop
fandn_done:
	VZEROUPPER
	RET

// func fillArrayXORVector(container []uint16, bitmap1, bitmap2 []uint64)
TEXT ·fillArrayXORVector(SB), NOSPLIT, $0-72
	MOVQ container_base+0(FP), DI
	MOVQ bitmap1_base+24(FP), SI
	MOVQ bitmap1_len+32(FP), CX
	MOVQ bitmap2_base+48(FP), BX
	SETUP16
	TESTQ CX, CX
	JZ    fxor_done
fxor_loop:
	MOVQ (SI), AX
	XORQ (BX), AX
	EMIT16(fxor_s, fxor_a)
	ADDQ $8, SI
	ADDQ $8, BX
	DECQ CX
	JNZ  fxor_loop
fxor_done:
	VZEROUPPER
	RET

// func fillArraySkipVector(bitmap []uint64, container []uint16)
//
// Same as fillArrayVector but tests eight words at a time with VPTESTMQ and
// skips wholly empty groups. Wins when the container is very sparse.
TEXT ·fillArraySkipVector(SB), NOSPLIT, $0-48
	MOVQ bitmap_base+0(FP), SI
	MOVQ bitmap_len+8(FP), CX
	MOVQ container_base+24(FP), DI
	VMOVDQU64 fa_bitPositions<>(SB), Z0
	SHRQ $3, CX
	TESTQ CX, CX
	JZ    fs_done
	XORQ  DX, DX            // word index of the group
fs_loop:
	VMOVDQU64 (SI), Z6
	VPTESTMQ  Z6, Z6, K3
	KMOVB     K3, R9
	TESTQ     R9, R9
	JZ        fs_next
fs_word:
	TZCNTQ R9, R12
	BLSRQ  R9, R9
	MOVQ   (SI)(R12*8), AX
	KMOVQ  AX, K1
	VPCOMPRESSB.Z Z0, K1, Z2
	POPCNTQ AX, R10
	MOVQ   $-1, R11
	BZHIQ  R10, R11, R11
	LEAQ   (DX)(R12*1), R13
	SHLQ   $6, R13
	VPBROADCASTW R13, Z1
	KMOVD  R11, K2
	VPMOVZXBW Y2, Z4
	VPADDW Z1, Z4, Z4
	VMOVDQU16 Z4, K2, (DI)
	CMPQ   R10, $32
	JLE    fs_adv
	SHRQ   $32, R11
	KMOVD  R11, K2
	VEXTRACTI64X4 $1, Z2, Y5
	VPMOVZXBW Y5, Z4
	VPADDW Z1, Z4, Z4
	VMOVDQU16 Z4, K2, 64(DI)
fs_adv:
	LEAQ   (DI)(R10*2), DI
	TESTQ  R9, R9
	JNZ    fs_word
fs_next:
	ADDQ $64, SI
	ADDQ $8, DX
	DECQ CX
	JNZ  fs_loop
fs_done:
	VZEROUPPER
	RET
