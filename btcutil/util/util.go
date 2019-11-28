package util

import (
	"encoding/hex"
	"unsafe"

	"github.com/pkt-cash/pktd/btcutil/er"
)

func IsNil(i interface{}) bool {
	return (*[2]uintptr)(unsafe.Pointer(&i))[1] == 0
}

func DecodeHex(s string) ([]byte, er.R) {
	o, e := hex.DecodeString(s)
	return o, er.E(e)
}
