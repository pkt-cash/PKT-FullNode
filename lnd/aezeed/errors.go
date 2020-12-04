package aezeed

import (
	"github.com/pkt-cash/pktd/btcutil/er"
)

var (
	Err = er.NewErrorType("aezeed")
	// ErrIncorrectVersion is returned if a seed bares a mismatched
	// external version to that of the package executing the aezeed scheme.
	ErrIncorrectVersion = Err.CodeWithDetail("ErrIncorrectVersion",
		"wrong seed version")

	// ErrInvalidPass is returned if the user enters an invalid passphrase
	// for a particular enciphered mnemonic.
	ErrInvalidPass = Err.CodeWithDetail("ErrInvalidPass", "invalid passphrase")

	// ErrIncorrectMnemonic is returned if we detect that the checksum of
	// the specified mnemonic doesn't match. This indicates the user input
	// the wrong mnemonic.
	ErrIncorrectMnemonic = Err.CodeWithDetail("ErrIncorrectMnemonic",
		"mnemonic phrase checksum doesn't match")

	ErrUnknownMnenomicWord = Err.Code("ErrUnknownMnenomicWord")
)
