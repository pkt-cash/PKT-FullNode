package table

import (
	"testing"

	"github.com/pkt-cash/PKT-FullNode/goleveldb/leveldb/testutil"
)

func TestTable(t *testing.T) {
	testutil.RunSuite(t, "Table Suite")
}
