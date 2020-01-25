// Copyright (c) 2019 Caleb James DeLisle
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package pcutil

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
