package util

import (
	"encoding/hex"
	"os"
	"testing"
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

// CheckError ensures the passed error has an error code that matches the passed  error code.
func CheckError(t *testing.T, testName string, gotErr er.R, wantErrCode *er.ErrorCode) bool {
	if !wantErrCode.Is(gotErr) {
		t.Errorf("%s: unexpected error code - got %s, want %s",
			testName, gotErr.Message(), wantErrCode.Default())
		return false
	}

	return true
}

func Exists(path string) bool {
	_, errr := os.Stat(path)
	return !os.IsNotExist(errr)
}
