package lnpeer

import "github.com/pkt-cash/pktd/btcutil/er"

var (
	// ErrPeerExiting signals that the peer received a disconnect request.
	ErrPeerExiting = er.GenericErrorType.CodeWithDetail("ErrPeerExiting",
		"peer exiting")
)
