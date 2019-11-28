package testutil

import (
	"testing"

	"github.com/pkt-cash/pktd/btcutil/er"
)

// CheckDbError ensures the passed error is a database.Error with an error code
// that matches the passed  error code.
func CheckDbError(t *testing.T, testName string, gotErr er.R, wantErrCode *er.ErrorCode) bool {
	if !wantErrCode.Is(gotErr) {
		t.Errorf("%s: unexpected error code - got %s, want %s",
			testName, gotErr.Message(), wantErrCode.Default())
		return false
	}

	return true
}
