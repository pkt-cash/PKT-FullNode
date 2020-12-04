package contractcourt

import (
	"io"

	"github.com/pkt-cash/pktd/btcutil/er"
	"github.com/pkt-cash/pktd/lnd/channeldb"
	"github.com/pkt-cash/pktd/lnd/htlcswitch/hop"
	"github.com/pkt-cash/pktd/lnd/input"
	"github.com/pkt-cash/pktd/lnd/invoices"
	"github.com/pkt-cash/pktd/lnd/lntypes"
	"github.com/pkt-cash/pktd/lnd/lnwallet/chainfee"
	"github.com/pkt-cash/pktd/lnd/lnwire"
	"github.com/pkt-cash/pktd/lnd/sweep"
	"github.com/pkt-cash/pktd/wire"
)

// Registry is an interface which represents the invoice registry.
type Registry interface {
	// LookupInvoice attempts to look up an invoice according to its 32
	// byte payment hash.
	LookupInvoice(lntypes.Hash) (channeldb.Invoice, er.R)

	// NotifyExitHopHtlc attempts to mark an invoice as settled. If the
	// invoice is a debug invoice, then this method is a noop as debug
	// invoices are never fully settled. The return value describes how the
	// htlc should be resolved. If the htlc cannot be resolved immediately,
	// the resolution is sent on the passed in hodlChan later.
	NotifyExitHopHtlc(payHash lntypes.Hash, paidAmount lnwire.MilliSatoshi,
		expiry uint32, currentHeight int32,
		circuitKey channeldb.CircuitKey, hodlChan chan<- interface{},
		payload invoices.Payload) (invoices.HtlcResolution, er.R)

	// HodlUnsubscribeAll unsubscribes from all htlc resolutions.
	HodlUnsubscribeAll(subscriber chan<- interface{})
}

// OnionProcessor is an interface used to decode onion blobs.
type OnionProcessor interface {
	// ReconstructHopIterator attempts to decode a valid sphinx packet from
	// the passed io.Reader instance.
	ReconstructHopIterator(r io.Reader, rHash []byte) (hop.Iterator, er.R)
}

// UtxoSweeper defines the sweep functions that contract court requires.
type UtxoSweeper interface {
	// SweepInput sweeps inputs back into the wallet.
	SweepInput(input input.Input, params sweep.Params) (chan sweep.Result,
		er.R)

	// CreateSweepTx accepts a list of inputs and signs and generates a txn
	// that spends from them. This method also makes an accurate fee
	// estimate before generating the required witnesses.
	CreateSweepTx(inputs []input.Input, feePref sweep.FeePreference,
		currentBlockHeight uint32) (*wire.MsgTx, er.R)

	// RelayFeePerKW returns the minimum fee rate required for transactions
	// to be relayed.
	RelayFeePerKW() chainfee.SatPerKWeight

	// UpdateParams allows updating the sweep parameters of a pending input
	// in the UtxoSweeper. This function can be used to provide an updated
	// fee preference that will be used for a new sweep transaction of the
	// input that will act as a replacement transaction (RBF) of the
	// original sweeping transaction, if any.
	UpdateParams(input wire.OutPoint, params sweep.ParamsUpdate) (
		chan sweep.Result, er.R)
}
