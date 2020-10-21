package neutrino

import (
	"github.com/pkt-cash/pktd/btcutil/er"
)

var Err er.ErrorType = er.NewErrorType("neutrino.Err")

var (
	// ErrGetUtxoCanceled signals that a GetUtxo request was canceled.
	ErrGetUtxoCanceled = Err.CodeWithDetail("ErrGetUtxoCanceled",
		"getutxorequest cancellation")

	// ErrShuttingDown signals that neutrino received a shutdown request.
	ErrShuttingDown = Err.CodeWithDetail("ErrShuttingDown",
		"neutrino shutting down")
)
