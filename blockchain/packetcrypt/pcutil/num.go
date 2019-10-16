// Copyright (c) 2019 Caleb James DeLisle
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package pcutil

import "math/bits"

// Log2floor returns floor(log2(x))
func Log2floor(x uint64) int {
	if x == 0 {
		panic("log2floor called on 0")
	}
	return 63 - bits.LeadingZeros64(x)
}

// Log2ceil returns the ceiling(log2(x))
func Log2ceil(x uint64) int {
	out := Log2floor(x)
	if (x & (x - 1)) != 0 {
		out++
	}
	return out
}
