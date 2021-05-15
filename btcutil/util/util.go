package util

import (
	"encoding/binary"
	"encoding/hex"
	"io"
	"os"
	"testing"
	"unsafe"

	"github.com/pkt-cash/pktd/btcutil/er"
	"github.com/stretchr/testify/require"
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

func CloneBytes(b []byte) []byte {
	out := make([]byte, len(b))
	if copy(out, b) != len(b) {
		panic("copy not length of bytes")
	}
	return out
}

func WriteBin(w io.Writer, order binary.ByteOrder, data interface{}) er.R {
	return er.E(binary.Write(w, order, data))
}

func ReadBin(r io.Reader, order binary.ByteOrder, data interface{}) er.R {
	return er.E(binary.Read(r, order, data))
}

func ReadFull(r io.Reader, buf []byte) (int, er.R) {
	i, e := io.ReadFull(r, buf)
	return i, er.E(e)
}

func Write(w io.Writer, b []byte) (int, er.R) {
	i, e := w.Write(b)
	return i, er.E(e)
}

func RequireErr(t require.TestingT, err er.R, msgAndArgs ...interface{}) {
	require.Error(t, er.Native(err), msgAndArgs...)
}

func RequireNoErr(t require.TestingT, err er.R, msgAndArgs ...interface{}) {
	require.NoError(t, er.Native(err), msgAndArgs...)
}
