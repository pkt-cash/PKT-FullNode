// Copyright (c) 2016 The Decred developers
// Copyright (c) 2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wallet

import (
	"time"

	"github.com/pkt-cash/pktd/chaincfg/chainhash"
	"github.com/pkt-cash/pktd/wire"
)

// Note: The following common types should never reference the Wallet type.
// Long term goal is to move these to their own package so that the database
// access APIs can create them directly for the wallet to return.

// BlockIdentity identifies a block, or the lack of one (used to describe an
// unmined transaction).
type BlockIdentity struct {
	Hash   chainhash.Hash
	Height int32
}

// OutputKind describes a kind of transaction output.  This is used to
// differentiate between coinbase, stakebase, and normal outputs.
type OutputKind byte

// Defined OutputKind constants
const (
	OutputKindNormal OutputKind = iota
	OutputKindCoinbase
)

// TransactionOutput describes an output that was or is at least partially
// controlled by the wallet.  Depending on context, this could refer to an
// unspent output, or a spent one.
type TransactionOutput struct {
	OutPoint   wire.OutPoint
	Output     wire.TxOut
	OutputKind OutputKind
	// These should be added later when the DB can return them more
	// efficiently:
	//TxLockTime      uint32
	//TxExpiry        uint32
	ContainingBlock BlockIdentity
	ReceiveTime     time.Time
}
