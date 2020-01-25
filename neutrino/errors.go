package neutrino

import (
	"github.com/pkt-cash/pktd/btcutil/er"
)

var Err er.ErrorType = er.NewErrorType("neutrino.Err")

var (
	// ErrGetUtxoCancelled signals that a GetUtxo request was cancelled.
	ErrGetUtxoCancelled = Err.CodeWithDetail("ErrGetUtxoCancelled",
		"get utxo request cancelled")

	// ErrShuttingDown signals that neutrino received a shutdown request.
	ErrShuttingDown = Err.CodeWithDetail("ErrShuttingDown",
		"neutrino shutting down")
)
