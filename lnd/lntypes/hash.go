package lntypes

import (
	"encoding/hex"

	"github.com/pkt-cash/pktd/btcutil/er"
	"github.com/pkt-cash/pktd/btcutil/util"
)

// HashSize of array used to store hashes.
const HashSize = 32

// ZeroHash is a predefined hash containing all zeroes.
var ZeroHash Hash

// Hash is used in several of the lightning messages and common structures. It
// typically represents a payment hash.
type Hash [HashSize]byte

// String returns the Hash as a hexadecimal string.
func (hash Hash) String() string {
	return hex.EncodeToString(hash[:])
}

// MakeHash returns a new Hash from a byte slice.  An error is returned if
// the number of bytes passed in is not HashSize.
func MakeHash(newHash []byte) (Hash, er.R) {
	nhlen := len(newHash)
	if nhlen != HashSize {
		return Hash{}, er.Errorf("invalid hash length of %v, want %v",
			nhlen, HashSize)
	}

	var hash Hash
	copy(hash[:], newHash)

	return hash, nil
}

// MakeHashFromStr creates a Hash from a hex hash string.
func MakeHashFromStr(newHash string) (Hash, er.R) {
	// Return error if hash string is of incorrect length.
	if len(newHash) != HashSize*2 {
		return Hash{}, er.Errorf("invalid hash string length of %v, "+
			"want %v", len(newHash), HashSize*2)
	}

	hash, err := util.DecodeHex(newHash)
	if err != nil {
		return Hash{}, err
	}

	return MakeHash(hash)
}
