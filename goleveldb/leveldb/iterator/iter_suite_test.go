package iterator_test

import (
	"testing"

	"github.com/pkt-cash/PKT-FullNode/goleveldb/leveldb/testutil"
)

func TestIterator(t *testing.T) {
	testutil.RunSuite(t, "Iterator Suite")
}
