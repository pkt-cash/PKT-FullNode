package sweep

import (
	"github.com/pkt-cash/pktd/btcutil/er"
	"github.com/pkt-cash/pktd/lnd/lnwallet"
	"github.com/pkt-cash/pktd/wire"
)

// Wallet contains all wallet related functionality required by sweeper.
type Wallet interface {
	// PublishTransaction performs cursory validation (dust checks, etc) and
	// broadcasts the passed transaction to the Bitcoin network.
	PublishTransaction(tx *wire.MsgTx, label string) er.R

	// ListUnspentWitness returns all unspent outputs which are version 0
	// witness programs. The 'minconfirms' and 'maxconfirms' parameters
	// indicate the minimum and maximum number of confirmations an output
	// needs in order to be returned by this method.
	ListUnspentWitness(minconfirms, maxconfirms int32) ([]*lnwallet.Utxo,
		er.R)

	// WithCoinSelectLock will execute the passed function closure in a
	// synchronized manner preventing any coin selection operations from
	// proceeding while the closure is executing. This can be seen as the
	// ability to execute a function closure under an exclusive coin
	// selection lock.
	WithCoinSelectLock(f func() er.R) er.R
}
