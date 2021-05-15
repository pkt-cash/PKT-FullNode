package fmgr

import (
	"github.com/pkt-cash/pktd/lnd/lnpeer"
	"github.com/pkt-cash/pktd/lnd/lnwire"
)

// Manager is an interface that describes the basic operation of a funding
// manager. It should at a minimum process a subset of lnwire messages that
// are denoted as funding messages.
type Manager interface {
	// ProcessFundingMsg processes a funding message represented by the
	// lnwire.Message parameter along with the Peer object representing a
	// connection to the counterparty.
	ProcessFundingMsg(lnwire.Message, lnpeer.Peer)

	// IsPendingChannel is used to determine whether to send an Error message
	// to the funding manager or not.
	IsPendingChannel([32]byte, lnpeer.Peer) bool
}
