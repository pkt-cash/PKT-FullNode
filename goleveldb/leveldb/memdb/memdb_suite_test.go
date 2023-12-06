package memdb

import (
	"testing"

	"github.com/pkt-cash/PKT-FullNode/goleveldb/leveldb/testutil"
)

func TestMemDB(t *testing.T) {
	testutil.RunSuite(t, "MemDB Suite")
}
