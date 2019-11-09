// Copyright (c) 2019 Caleb James DeLisle
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package interpret

import (
	"fmt"
	"github.com/pkt-cash/pktd/blockchain/packetcrypt/randhash/opcodes"
	"github.com/pkt-cash/pktd/btcutil/er"
)

func doOp(ctx *Context, insn uint32, op opcodes.OpCode) {

	switch op {

	case opcodes.OpCode_POPCNT8:
		a := getA(ctx, insn)
		out := POPCNT8(a)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x -> %08x\n", "POPCNT8", a, out)
		}
		out1(ctx, out)
	case opcodes.OpCode_POPCNT16:
		a := getA(ctx, insn)
		out := POPCNT16(a)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x -> %08x\n", "POPCNT16", a, out)
		}
		out1(ctx, out)
	case opcodes.OpCode_POPCNT32:
		a := getA(ctx, insn)
		out := POPCNT32(a)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x -> %08x\n", "POPCNT32", a, out)
		}
		out1(ctx, out)
	case opcodes.OpCode_CLZ8:
		a := getA(ctx, insn)
		out := CLZ8(a)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x -> %08x\n", "CLZ8", a, out)
		}
		out1(ctx, out)
	case opcodes.OpCode_CLZ16:
		a := getA(ctx, insn)
		out := CLZ16(a)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x -> %08x\n", "CLZ16", a, out)
		}
		out1(ctx, out)
	case opcodes.OpCode_CLZ32:
		a := getA(ctx, insn)
		out := CLZ32(a)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x -> %08x\n", "CLZ32", a, out)
		}
		out1(ctx, out)
	case opcodes.OpCode_CTZ8:
		a := getA(ctx, insn)
		out := CTZ8(a)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x -> %08x\n", "CTZ8", a, out)
		}
		out1(ctx, out)
	case opcodes.OpCode_CTZ16:
		a := getA(ctx, insn)
		out := CTZ16(a)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x -> %08x\n", "CTZ16", a, out)
		}
		out1(ctx, out)
	case opcodes.OpCode_CTZ32:
		a := getA(ctx, insn)
		out := CTZ32(a)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x -> %08x\n", "CTZ32", a, out)
		}
		out1(ctx, out)

	case opcodes.OpCode_BSWAP16:
		a := getA(ctx, insn)
		out := BSWAP16(a)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x -> %08x\n", "BSWAP16", a, out)
		}
		out1(ctx, out)
	case opcodes.OpCode_BSWAP32:
		a := getA(ctx, insn)
		out := BSWAP32(a)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x -> %08x\n", "BSWAP32", a, out)
		}
		out1(ctx, out)

	case opcodes.OpCode_ADD8:
		a := getA(ctx, insn)
		b := getB(ctx, insn)
		out := ADD8(a, b)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x %08x -> %08x\n", "ADD8", a, b, out)
		}
		out1(ctx, out)
	case opcodes.OpCode_ADD16:
		a := getA(ctx, insn)
		b := getB(ctx, insn)
		out := ADD16(a, b)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x %08x -> %08x\n", "ADD16", a, b, out)
		}
		out1(ctx, out)
	case opcodes.OpCode_ADD32:
		a := getA(ctx, insn)
		b := getB(ctx, insn)
		out := ADD32(a, b)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x %08x -> %08x\n", "ADD32", a, b, out)
		}
		out1(ctx, out)
	case opcodes.OpCode_SUB8:
		a := getA(ctx, insn)
		b := getB(ctx, insn)
		out := SUB8(a, b)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x %08x -> %08x\n", "SUB8", a, b, out)
		}
		out1(ctx, out)
	case opcodes.OpCode_SUB16:
		a := getA(ctx, insn)
		b := getB(ctx, insn)
		out := SUB16(a, b)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x %08x -> %08x\n", "SUB16", a, b, out)
		}
		out1(ctx, out)
	case opcodes.OpCode_SUB32:
		a := getA(ctx, insn)
		b := getB(ctx, insn)
		out := SUB32(a, b)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x %08x -> %08x\n", "SUB32", a, b, out)
		}
		out1(ctx, out)
	case opcodes.OpCode_SHLL8:
		a := getA(ctx, insn)
		b := getB(ctx, insn)
		out := SHLL8(a, b)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x %08x -> %08x\n", "SHLL8", a, b, out)
		}
		out1(ctx, out)
	case opcodes.OpCode_SHLL16:
		a := getA(ctx, insn)
		b := getB(ctx, insn)
		out := SHLL16(a, b)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x %08x -> %08x\n", "SHLL16", a, b, out)
		}
		out1(ctx, out)
	case opcodes.OpCode_SHLL32:
		a := getA(ctx, insn)
		b := getB(ctx, insn)
		out := SHLL32(a, b)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x %08x -> %08x\n", "SHLL32", a, b, out)
		}
		out1(ctx, out)
	case opcodes.OpCode_SHRL8:
		a := getA(ctx, insn)
		b := getB(ctx, insn)
		out := SHRL8(a, b)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x %08x -> %08x\n", "SHRL8", a, b, out)
		}
		out1(ctx, out)
	case opcodes.OpCode_SHRL16:
		a := getA(ctx, insn)
		b := getB(ctx, insn)
		out := SHRL16(a, b)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x %08x -> %08x\n", "SHRL16", a, b, out)
		}
		out1(ctx, out)
	case opcodes.OpCode_SHRL32:
		a := getA(ctx, insn)
		b := getB(ctx, insn)
		out := SHRL32(a, b)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x %08x -> %08x\n", "SHRL32", a, b, out)
		}
		out1(ctx, out)
	case opcodes.OpCode_SHRA8:
		a := getA(ctx, insn)
		b := getB(ctx, insn)
		out := SHRA8(a, b)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x %08x -> %08x\n", "SHRA8", a, b, out)
		}
		out1(ctx, out)
	case opcodes.OpCode_SHRA16:
		a := getA(ctx, insn)
		b := getB(ctx, insn)
		out := SHRA16(a, b)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x %08x -> %08x\n", "SHRA16", a, b, out)
		}
		out1(ctx, out)
	case opcodes.OpCode_SHRA32:
		a := getA(ctx, insn)
		b := getB(ctx, insn)
		out := SHRA32(a, b)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x %08x -> %08x\n", "SHRA32", a, b, out)
		}
		out1(ctx, out)
	case opcodes.OpCode_ROTL8:
		a := getA(ctx, insn)
		b := getB(ctx, insn)
		out := ROTL8(a, b)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x %08x -> %08x\n", "ROTL8", a, b, out)
		}
		out1(ctx, out)
	case opcodes.OpCode_ROTL16:
		a := getA(ctx, insn)
		b := getB(ctx, insn)
		out := ROTL16(a, b)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x %08x -> %08x\n", "ROTL16", a, b, out)
		}
		out1(ctx, out)
	case opcodes.OpCode_ROTL32:
		a := getA(ctx, insn)
		b := getB(ctx, insn)
		out := ROTL32(a, b)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x %08x -> %08x\n", "ROTL32", a, b, out)
		}
		out1(ctx, out)
	case opcodes.OpCode_MUL8:
		a := getA(ctx, insn)
		b := getB(ctx, insn)
		out := MUL8(a, b)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x %08x -> %08x\n", "MUL8", a, b, out)
		}
		out1(ctx, out)
	case opcodes.OpCode_MUL16:
		a := getA(ctx, insn)
		b := getB(ctx, insn)
		out := MUL16(a, b)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x %08x -> %08x\n", "MUL16", a, b, out)
		}
		out1(ctx, out)
	case opcodes.OpCode_MUL32:
		a := getA(ctx, insn)
		b := getB(ctx, insn)
		out := MUL32(a, b)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x %08x -> %08x\n", "MUL32", a, b, out)
		}
		out1(ctx, out)

	case opcodes.OpCode_AND:
		a := getA(ctx, insn)
		b := getB(ctx, insn)
		out := AND(a, b)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x %08x -> %08x\n", "AND", a, b, out)
		}
		out1(ctx, out)
	case opcodes.OpCode_OR:
		a := getA(ctx, insn)
		b := getB(ctx, insn)
		out := OR(a, b)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x %08x -> %08x\n", "OR", a, b, out)
		}
		out1(ctx, out)
	case opcodes.OpCode_XOR:
		a := getA(ctx, insn)
		b := getB(ctx, insn)
		out := XOR(a, b)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x %08x -> %08x\n", "XOR", a, b, out)
		}
		out1(ctx, out)

	case opcodes.OpCode_ADD8C:
		a := getA(ctx, insn)
		b := getB(ctx, insn)
		out := ADD8C(a, b)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x %08x -> %08x %08x\n", "ADD8C", a, b, uint32(out), uint32(out>>32))
		}
		out2(ctx, out)
	case opcodes.OpCode_ADD16C:
		a := getA(ctx, insn)
		b := getB(ctx, insn)
		out := ADD16C(a, b)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x %08x -> %08x %08x\n", "ADD16C", a, b, uint32(out), uint32(out>>32))
		}
		out2(ctx, out)
	case opcodes.OpCode_ADD32C:
		a := getA(ctx, insn)
		b := getB(ctx, insn)
		out := ADD32C(a, b)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x %08x -> %08x %08x\n", "ADD32C", a, b, uint32(out), uint32(out>>32))
		}
		out2(ctx, out)
	case opcodes.OpCode_SUB8C:
		a := getA(ctx, insn)
		b := getB(ctx, insn)
		out := SUB8C(a, b)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x %08x -> %08x %08x\n", "SUB8C", a, b, uint32(out), uint32(out>>32))
		}
		out2(ctx, out)
	case opcodes.OpCode_SUB16C:
		a := getA(ctx, insn)
		b := getB(ctx, insn)
		out := SUB16C(a, b)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x %08x -> %08x %08x\n", "SUB16C", a, b, uint32(out), uint32(out>>32))
		}
		out2(ctx, out)
	case opcodes.OpCode_SUB32C:
		a := getA(ctx, insn)
		b := getB(ctx, insn)
		out := SUB32C(a, b)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x %08x -> %08x %08x\n", "SUB32C", a, b, uint32(out), uint32(out>>32))
		}
		out2(ctx, out)
	case opcodes.OpCode_MUL8C:
		a := getA(ctx, insn)
		b := getB(ctx, insn)
		out := MUL8C(a, b)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x %08x -> %08x %08x\n", "MUL8C", a, b, uint32(out), uint32(out>>32))
		}
		out2(ctx, out)
	case opcodes.OpCode_MUL16C:
		a := getA(ctx, insn)
		b := getB(ctx, insn)
		out := MUL16C(a, b)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x %08x -> %08x %08x\n", "MUL16C", a, b, uint32(out), uint32(out>>32))
		}
		out2(ctx, out)
	case opcodes.OpCode_MUL32C:
		a := getA(ctx, insn)
		b := getB(ctx, insn)
		out := MUL32C(a, b)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x %08x -> %08x %08x\n", "MUL32C", a, b, uint32(out), uint32(out>>32))
		}
		out2(ctx, out)
	case opcodes.OpCode_MULSU8C:
		a := getA(ctx, insn)
		b := getB(ctx, insn)
		out := MULSU8C(a, b)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x %08x -> %08x %08x\n", "MULSU8C", a, b, uint32(out), uint32(out>>32))
		}
		out2(ctx, out)
	case opcodes.OpCode_MULSU16C:
		a := getA(ctx, insn)
		b := getB(ctx, insn)
		out := MULSU16C(a, b)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x %08x -> %08x %08x\n", "MULSU16C", a, b, uint32(out), uint32(out>>32))
		}
		out2(ctx, out)
	case opcodes.OpCode_MULSU32C:
		a := getA(ctx, insn)
		b := getB(ctx, insn)
		out := MULSU32C(a, b)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x %08x -> %08x %08x\n", "MULSU32C", a, b, uint32(out), uint32(out>>32))
		}
		out2(ctx, out)
	case opcodes.OpCode_MULU8C:
		a := getA(ctx, insn)
		b := getB(ctx, insn)
		out := MULU8C(a, b)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x %08x -> %08x %08x\n", "MULU8C", a, b, uint32(out), uint32(out>>32))
		}
		out2(ctx, out)
	case opcodes.OpCode_MULU16C:
		a := getA(ctx, insn)
		b := getB(ctx, insn)
		out := MULU16C(a, b)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x %08x -> %08x %08x\n", "MULU16C", a, b, uint32(out), uint32(out>>32))
		}
		out2(ctx, out)
	case opcodes.OpCode_MULU32C:
		a := getA(ctx, insn)
		b := getB(ctx, insn)
		out := MULU32C(a, b)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x %08x -> %08x %08x\n", "MULU32C", a, b, uint32(out), uint32(out>>32))
		}
		out2(ctx, out)
	case opcodes.OpCode_ADD64:
		a := getA2(ctx, insn)
		b := getB2(ctx, insn)
		out := ADD64(a.lo, a.hi, b.lo, b.hi)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x %08x %08x %08x -> %08x %08x\n", "ADD64", a.lo, a.hi, b.lo, b.hi, uint32(out), uint32(out>>32))
		}
		out2(ctx, out)
	case opcodes.OpCode_SUB64:
		a := getA2(ctx, insn)
		b := getB2(ctx, insn)
		out := SUB64(a.lo, a.hi, b.lo, b.hi)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x %08x %08x %08x -> %08x %08x\n", "SUB64", a.lo, a.hi, b.lo, b.hi, uint32(out), uint32(out>>32))
		}
		out2(ctx, out)
	case opcodes.OpCode_SHLL64:
		a := getA2(ctx, insn)
		b := getB2(ctx, insn)
		out := SHLL64(a.lo, a.hi, b.lo, b.hi)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x %08x %08x %08x -> %08x %08x\n", "SHLL64", a.lo, a.hi, b.lo, b.hi, uint32(out), uint32(out>>32))
		}
		out2(ctx, out)
	case opcodes.OpCode_SHRL64:
		a := getA2(ctx, insn)
		b := getB2(ctx, insn)
		out := SHRL64(a.lo, a.hi, b.lo, b.hi)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x %08x %08x %08x -> %08x %08x\n", "SHRL64", a.lo, a.hi, b.lo, b.hi, uint32(out), uint32(out>>32))
		}
		out2(ctx, out)
	case opcodes.OpCode_SHRA64:
		a := getA2(ctx, insn)
		b := getB2(ctx, insn)
		out := SHRA64(a.lo, a.hi, b.lo, b.hi)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x %08x %08x %08x -> %08x %08x\n", "SHRA64", a.lo, a.hi, b.lo, b.hi, uint32(out), uint32(out>>32))
		}
		out2(ctx, out)
	case opcodes.OpCode_ROTL64:
		a := getA2(ctx, insn)
		b := getB2(ctx, insn)
		out := ROTL64(a.lo, a.hi, b.lo, b.hi)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x %08x %08x %08x -> %08x %08x\n", "ROTL64", a.lo, a.hi, b.lo, b.hi, uint32(out), uint32(out>>32))
		}
		out2(ctx, out)
	case opcodes.OpCode_ROTR64:
		a := getA2(ctx, insn)
		b := getB2(ctx, insn)
		out := ROTR64(a.lo, a.hi, b.lo, b.hi)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x %08x %08x %08x -> %08x %08x\n", "ROTR64", a.lo, a.hi, b.lo, b.hi, uint32(out), uint32(out>>32))
		}
		out2(ctx, out)
	case opcodes.OpCode_MUL64:
		a := getA2(ctx, insn)
		b := getB2(ctx, insn)
		out := MUL64(a.lo, a.hi, b.lo, b.hi)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x %08x %08x %08x -> %08x %08x\n", "MUL64", a.lo, a.hi, b.lo, b.hi, uint32(out), uint32(out>>32))
		}
		out2(ctx, out)
	case opcodes.OpCode_ADD64C:
		a := getA2(ctx, insn)
		b := getB2(ctx, insn)
		out := ADD64C(a.lo, a.hi, b.lo, b.hi)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x %08x %08x %08x -> %08x %08x %08x %08x\n", "ADD64C", a.lo, a.hi, b.lo, b.hi, U128_0(out), U128_1(out), U128_2(out), U128_3(out))
		}
		out4(ctx, out)
	case opcodes.OpCode_SUB64C:
		a := getA2(ctx, insn)
		b := getB2(ctx, insn)
		out := SUB64C(a.lo, a.hi, b.lo, b.hi)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x %08x %08x %08x -> %08x %08x %08x %08x\n", "SUB64C", a.lo, a.hi, b.lo, b.hi, U128_0(out), U128_1(out), U128_2(out), U128_3(out))
		}
		out4(ctx, out)
	case opcodes.OpCode_MUL64C:
		a := getA2(ctx, insn)
		b := getB2(ctx, insn)
		out := MUL64C(a.lo, a.hi, b.lo, b.hi)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x %08x %08x %08x -> %08x %08x %08x %08x\n", "MUL64C", a.lo, a.hi, b.lo, b.hi, U128_0(out), U128_1(out), U128_2(out), U128_3(out))
		}
		out4(ctx, out)
	case opcodes.OpCode_MULSU64C:
		a := getA2(ctx, insn)
		b := getB2(ctx, insn)
		out := MULSU64C(a.lo, a.hi, b.lo, b.hi)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x %08x %08x %08x -> %08x %08x %08x %08x\n", "MULSU64C", a.lo, a.hi, b.lo, b.hi, U128_0(out), U128_1(out), U128_2(out), U128_3(out))
		}
		out4(ctx, out)
	case opcodes.OpCode_MULU64C:
		a := getA2(ctx, insn)
		b := getB2(ctx, insn)
		out := MULU64C(a.lo, a.hi, b.lo, b.hi)
		if debug {
			pad(ctx)
			fmt.Printf("%s %08x %08x %08x %08x -> %08x %08x %08x %08x\n", "MULU64C", a.lo, a.hi, b.lo, b.hi, U128_0(out), U128_1(out), U128_2(out), U128_3(out))
		}
		out4(ctx, out)

	default:
		{
			fmt.Printf("Unexpected instruction %s", op.String())
			panic("unexpected instruction")
		}
	}
}

func TestOp(inout []uint32, op opcodes.OpCode) {
	switch op {

	case opcodes.OpCode_POPCNT8:
		out1x(inout, POPCNT8(inout[0]))
	case opcodes.OpCode_POPCNT16:
		out1x(inout, POPCNT16(inout[0]))
	case opcodes.OpCode_POPCNT32:
		out1x(inout, POPCNT32(inout[0]))
	case opcodes.OpCode_CLZ8:
		out1x(inout, CLZ8(inout[0]))
	case opcodes.OpCode_CLZ16:
		out1x(inout, CLZ16(inout[0]))
	case opcodes.OpCode_CLZ32:
		out1x(inout, CLZ32(inout[0]))
	case opcodes.OpCode_CTZ8:
		out1x(inout, CTZ8(inout[0]))
	case opcodes.OpCode_CTZ16:
		out1x(inout, CTZ16(inout[0]))
	case opcodes.OpCode_CTZ32:
		out1x(inout, CTZ32(inout[0]))

	case opcodes.OpCode_BSWAP16:
		out1x(inout, BSWAP16(inout[0]))
	case opcodes.OpCode_BSWAP32:
		out1x(inout, BSWAP32(inout[0]))

	case opcodes.OpCode_ADD8:
		out1x(inout, ADD8(inout[0], inout[1]))
	case opcodes.OpCode_ADD16:
		out1x(inout, ADD16(inout[0], inout[1]))
	case opcodes.OpCode_ADD32:
		out1x(inout, ADD32(inout[0], inout[1]))
	case opcodes.OpCode_SUB8:
		out1x(inout, SUB8(inout[0], inout[1]))
	case opcodes.OpCode_SUB16:
		out1x(inout, SUB16(inout[0], inout[1]))
	case opcodes.OpCode_SUB32:
		out1x(inout, SUB32(inout[0], inout[1]))
	case opcodes.OpCode_SHLL8:
		out1x(inout, SHLL8(inout[0], inout[1]))
	case opcodes.OpCode_SHLL16:
		out1x(inout, SHLL16(inout[0], inout[1]))
	case opcodes.OpCode_SHLL32:
		out1x(inout, SHLL32(inout[0], inout[1]))
	case opcodes.OpCode_SHRL8:
		out1x(inout, SHRL8(inout[0], inout[1]))
	case opcodes.OpCode_SHRL16:
		out1x(inout, SHRL16(inout[0], inout[1]))
	case opcodes.OpCode_SHRL32:
		out1x(inout, SHRL32(inout[0], inout[1]))
	case opcodes.OpCode_SHRA8:
		out1x(inout, SHRA8(inout[0], inout[1]))
	case opcodes.OpCode_SHRA16:
		out1x(inout, SHRA16(inout[0], inout[1]))
	case opcodes.OpCode_SHRA32:
		out1x(inout, SHRA32(inout[0], inout[1]))
	case opcodes.OpCode_ROTL8:
		out1x(inout, ROTL8(inout[0], inout[1]))
	case opcodes.OpCode_ROTL16:
		out1x(inout, ROTL16(inout[0], inout[1]))
	case opcodes.OpCode_ROTL32:
		out1x(inout, ROTL32(inout[0], inout[1]))
	case opcodes.OpCode_MUL8:
		out1x(inout, MUL8(inout[0], inout[1]))
	case opcodes.OpCode_MUL16:
		out1x(inout, MUL16(inout[0], inout[1]))
	case opcodes.OpCode_MUL32:
		out1x(inout, MUL32(inout[0], inout[1]))

	case opcodes.OpCode_AND:
		out1x(inout, AND(inout[0], inout[1]))
	case opcodes.OpCode_OR:
		out1x(inout, OR(inout[0], inout[1]))
	case opcodes.OpCode_XOR:
		out1x(inout, XOR(inout[0], inout[1]))

	case opcodes.OpCode_ADD8C:
		out2x(inout, ADD8C(inout[0], inout[1]))
	case opcodes.OpCode_ADD16C:
		out2x(inout, ADD16C(inout[0], inout[1]))
	case opcodes.OpCode_ADD32C:
		out2x(inout, ADD32C(inout[0], inout[1]))
	case opcodes.OpCode_SUB8C:
		out2x(inout, SUB8C(inout[0], inout[1]))
	case opcodes.OpCode_SUB16C:
		out2x(inout, SUB16C(inout[0], inout[1]))
	case opcodes.OpCode_SUB32C:
		out2x(inout, SUB32C(inout[0], inout[1]))
	case opcodes.OpCode_MUL8C:
		out2x(inout, MUL8C(inout[0], inout[1]))
	case opcodes.OpCode_MUL16C:
		out2x(inout, MUL16C(inout[0], inout[1]))
	case opcodes.OpCode_MUL32C:
		out2x(inout, MUL32C(inout[0], inout[1]))
	case opcodes.OpCode_MULSU8C:
		out2x(inout, MULSU8C(inout[0], inout[1]))
	case opcodes.OpCode_MULSU16C:
		out2x(inout, MULSU16C(inout[0], inout[1]))
	case opcodes.OpCode_MULSU32C:
		out2x(inout, MULSU32C(inout[0], inout[1]))
	case opcodes.OpCode_MULU8C:
		out2x(inout, MULU8C(inout[0], inout[1]))
	case opcodes.OpCode_MULU16C:
		out2x(inout, MULU16C(inout[0], inout[1]))
	case opcodes.OpCode_MULU32C:
		out2x(inout, MULU32C(inout[0], inout[1]))

	case opcodes.OpCode_ADD64:
		out2x(inout, ADD64(inout[0], inout[1], inout[2], inout[3]))
	case opcodes.OpCode_SUB64:
		out2x(inout, SUB64(inout[0], inout[1], inout[2], inout[3]))
	case opcodes.OpCode_SHLL64:
		out2x(inout, SHLL64(inout[0], inout[1], inout[2], inout[3]))
	case opcodes.OpCode_SHRL64:
		out2x(inout, SHRL64(inout[0], inout[1], inout[2], inout[3]))
	case opcodes.OpCode_SHRA64:
		out2x(inout, SHRA64(inout[0], inout[1], inout[2], inout[3]))
	case opcodes.OpCode_ROTL64:
		out2x(inout, ROTL64(inout[0], inout[1], inout[2], inout[3]))
	case opcodes.OpCode_ROTR64:
		out2x(inout, ROTR64(inout[0], inout[1], inout[2], inout[3]))
	case opcodes.OpCode_MUL64:
		out2x(inout, MUL64(inout[0], inout[1], inout[2], inout[3]))

	case opcodes.OpCode_ADD64C:
		out4x(inout, ADD64C(inout[0], inout[1], inout[2], inout[3]))
	case opcodes.OpCode_SUB64C:
		out4x(inout, SUB64C(inout[0], inout[1], inout[2], inout[3]))
	case opcodes.OpCode_MUL64C:
		out4x(inout, MUL64C(inout[0], inout[1], inout[2], inout[3]))
	case opcodes.OpCode_MULSU64C:
		out4x(inout, MULSU64C(inout[0], inout[1], inout[2], inout[3]))
	case opcodes.OpCode_MULU64C:
		out4x(inout, MULU64C(inout[0], inout[1], inout[2], inout[3]))
	}
}
