// Copyright (c) 2019 Caleb James DeLisle
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package pcutil

import (
	"encoding/binary"

	"github.com/aead/chacha20"
	"golang.org/x/crypto/blake2b"
)

func HashExpand(out, key []byte, counter uint32) {
	if len(key) != 32 {
		panic("unexpected key length")
	}
	nonce := []byte("____PC_EXPND")
	binary.LittleEndian.PutUint32(nonce[0:4], counter)
	for i := range out {
		out[i] = 0
	}
	//chacha20.XORKeyStream(out, out, &nonce, &key)
	chacha20.XORKeyStream(out, out, nonce, key)
}

func HashCompress(out, in []byte) {
	if len(out) < 32 {
		panic("need 32 byte output to place hash in")
	}
	hash := blake2b.Sum256(in)
	copy(out, hash[:])
}

func HashCompress64(out, in []byte) {
	if len(out) < 64 {
		panic("need 64 byte output to place hash in")
	}
	hash := blake2b.Sum512(in)
	copy(out, hash[:])
}
