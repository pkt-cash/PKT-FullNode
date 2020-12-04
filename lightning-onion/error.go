package sphinx

import (
	"github.com/pkt-cash/pktd/btcutil/er"
)

var (
	Err = er.NewErrorType("lightning-onion")

	// ErrReplayedPacket is an error returned when a packet is rejected
	// during processing due to being an attempted replay or probing
	// attempt.
	ErrReplayedPacket = Err.CodeWithDetail("ErrReplayedPacket",
		"sphinx packet replay attempted")

	// ErrInvalidOnionVersion is returned during decoding of the onion
	// packet, when the received packet has an unknown version byte.
	ErrInvalidOnionVersion = Err.CodeWithDetail("ErrInvalidOnionVersion",
		"invalid onion packet version")

	// ErrInvalidOnionHMAC is returned during onion parsing process, when received
	// mac does not corresponds to the generated one.
	ErrInvalidOnionHMAC = Err.CodeWithDetail("ErrInvalidOnionHMAC",
		"invalid mismatched mac")

	// ErrInvalidOnionKey is returned during onion parsing process, when
	// onion key is invalid.
	ErrInvalidOnionKey = Err.CodeWithDetail("ErrInvalidOnionKey",
		"invalid onion key: pubkey isn't on secp256k1 curve")

	// ErrLogEntryNotFound is an error returned when a packet lookup in a replay
	// log fails because it is missing.
	ErrLogEntryNotFound = Err.CodeWithDetail("ErrLogEntryNotFound",
		"sphinx packet is not in log")
)
