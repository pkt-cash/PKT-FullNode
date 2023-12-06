//go:build goleveldbtests
// +build goleveldbtests

package leveldb

import (
	"testing"

	"github.com/pkt-cash/PKT-FullNode/goleveldb/leveldb/testutil"
)

func TestLevelDB(t *testing.T) {
	testutil.RunSuite(t, "LevelDB Suite")
}
