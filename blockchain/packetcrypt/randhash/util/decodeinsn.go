// Copyright (c) 2019 Caleb James DeLisle
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package util

/*
 * Normal op
 *     3 3 2 2 2 2 2 2 2 2 2 2 1 1 1 1 1 1 1 1 1 1
 *     1 0 9 8 7 6 5 4 3 2 1 0 9 8 7 6 5 4 3 2 1 0 9 8 7 6 5 4 3 2 1 0
 *    +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 *  0 |  ?  |       regB      |?|0|      regA       |        op       |
 *    +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 *  4
 *
 * IMM op / IF / Loop / input
 *     3 3 2 2 2 2 2 2 2 2 2 2 1 1 1 1 1 1 1 1 1 1
 *     1 0 9 8 7 6 5 4 3 2 1 0 9 8 7 6 5 4 3 2 1 0 9 8 7 6 5 4 3 2 1 0
 *    +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 *  0 |          imm          |P|1|      regA       |        op       |
 *    +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 *  4
 *
 * JMP
 *     3 3 2 2 2 2 2 2 2 2 2 2 1 1 1 1 1 1 1 1 1 1
 *     1 0 9 8 7 6 5 4 3 2 1 0 9 8 7 6 5 4 3 2 1 0 9 8 7 6 5 4 3 2 1 0
 *    +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 *  0 |                     imm                     |        op       |
 *    +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 *  4
 *
 *  MEMORY
 *     3 3 2 2 2 2 2 2 2 2 2 2 1 1 1 1 1 1 1 1 1 1
 *     1 0 9 8 7 6 5 4 3 2 1 0 9 8 7 6 5 4 3 2 1 0 9 8 7 6 5 4 3 2 1 0
 *    +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 *  0 |            randloc          |  step | carry |        op       |
 *    +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 *  4
 */

func DecodeInsn_REGA(insn uint32) uint32         { return (insn >> 9) & 0x1ff }
func DecodeInsn_REGB(insn uint32) uint32         { return (insn >> 20) & 0x1ff }
func DecodeInsn_HAS_IMM(insn uint32) bool        { return ((insn >> 18) & 1) != 0 }
func DecodeInsn_MEMORY_CARRY(insn uint32) uint32 { return (insn >> 9) & 15 }
func DecodeInsn_MEMORY_WITH_CARRY(insn, carry uint32) uint32 {
	return (insn & ^(uint32(15) << 9)) | ((carry & 15) << 9)
}
func DecodeInsn_MEMORY_STEP(insn uint32) uint32 { return (insn >> 13) & 15 }
func DecodeInsn_MEMORY_BASE(insn uint32) uint32 { return insn >> 17 }
func DecodeInsn_OP(insn uint32) uint32          { return insn & 0xff }

func DecodeInsn_imm(insn uint32) int64 {
	if (insn & (1 << 19)) != 0 {
		// it's a pattern

		//     1 1
		//     1 0 9 8 7 6 5 4 3 2 1 0
		//    +-+-+-+-+-+-+-+-+-+-+-+-+
		//  0 |S|I|    B    |    A    |
		//    +-+-+-+-+-+-+-+-+-+-+-+-+
		imm := insn >> 20
		a := imm & ((1 << 5) - 1)
		imm >>= 5
		b := imm & ((1 << 5) - 1)
		imm >>= 5
		i := imm & 1
		imm >>= 1
		s := imm

		big1 := uint64(1)
		out := uint64(((uint64(i) << 63) - 1) ^ (big1 << b) ^ (big1 << a))

		// Drop the top bit
		imm <<= 1
		imm >>= 1

		big1 &= uint64(s)
		out |= big1 << 63
		return int64(out)
	}
	return int64(int32(insn) >> 20)
}
func DecodeInsn_immLo(insn uint32) int32 { return int32(DecodeInsn_imm(insn)) }
