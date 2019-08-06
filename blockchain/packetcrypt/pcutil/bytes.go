// Copyright (c) 2019 Caleb James DeLisle
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package pcutil

import "encoding/binary"

func U32FromB(out []uint32, b []byte) []uint32 {
	if len(b)%4 != 0 {
		panic("length of b must be a multiple of 4")
	}
	if out == nil {
		out = make([]uint32, len(b)/4)
	} else {
		out = out[:len(b)/4]
	}
	for i := 0; i < len(out); i++ {
		out[i] = binary.LittleEndian.Uint32(b[i*4:][:4])
	}
	return out
}

func BFromU32(out []byte, in []uint32) []byte {
	if out == nil {
		out = make([]byte, len(in)*4)
	} else {
		out = out[:len(in)*4]
	}
	for i := 0; i < len(in); i++ {
		binary.LittleEndian.PutUint32(out[i*4:][:4], in[i])
	}
	return out
}

func Memset(x []byte, y byte) {
	for i := 0; i < len(x); i++ {
		x[i] = y
	}
}

func Zero(x []byte) {
	Memset(x, 0)
}

func IsZero(x []byte) bool {
	is := byte(0)
	for i := 0; i < len(x); i++ {
		is |= x[i]
	}
	return is == 0
}

func Bint(x bool) int {
	if x {
		return 1
	}
	return 0
}

// Reverse reverses the order of the bytes in a slice
func Reverse(x []byte) {
	l := len(x)
	for i := 0; i < l/2; i++ {
		x[i], x[l-1-i] = x[l-1-i], x[i]
	}
}
