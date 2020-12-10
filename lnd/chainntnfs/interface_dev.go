// +build dev

package chainntnfs

import (
	"github.com/pkt-cash/pktd/btcutil/er"
	"github.com/pkt-cash/pktd/chaincfg/chainhash"
)

// TestChainNotifier enables the use of methods that are only present during
// testing for ChainNotifiers.
type TestChainNotifier interface {
	ChainNotifier

	// UnsafeStart enables notifiers to start up with a specific best block.
	// Used for testing.
	UnsafeStart(int32, *chainhash.Hash, int32, func() er.R) er.R
}
