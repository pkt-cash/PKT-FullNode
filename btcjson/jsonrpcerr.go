// Copyright (c) 2014 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package btcjson

import "github.com/pkt-cash/pktd/btcutil/er"

// Standard JSON-RPC 2.0 errors.
const (
	ErrRPCInvalidRequest RPCErrorCode = -32600
	ErrRPCMethodNotFound RPCErrorCode = -32601
	ErrRPCInvalidParams  RPCErrorCode = -32602
	ErrRPCInternal       RPCErrorCode = -32603
	ErrRPCParse          RPCErrorCode = -32700
)

func NewErrRPCInternal() er.R {
	return NewRPCError(ErrRPCInternal, "Internal error", nil)
}

// General application defined JSON errors.
const (
	ErrRPCMisc                RPCErrorCode = -1
	ErrRPCForbiddenBySafeMode RPCErrorCode = -2
	ErrRPCType                RPCErrorCode = -3
	ErrRPCInvalidAddressOrKey RPCErrorCode = -5
	ErrRPCOutOfMemory         RPCErrorCode = -7
	ErrRPCInvalidParameter    RPCErrorCode = -8
	ErrRPCDatabase            RPCErrorCode = -20
	ErrRPCDeserialization     RPCErrorCode = -22
	ErrRPCVerify              RPCErrorCode = -25
)

// Peer-to-peer client errors.
const (
	ErrRPCClientNotConnected      RPCErrorCode = -9
	ErrRPCClientInInitialDownload RPCErrorCode = -10
	ErrRPCClientNodeNotAdded      RPCErrorCode = -24
)

// Wallet JSON errors
const (
	ErrRPCWallet                    RPCErrorCode = -4
	ErrRPCWalletInsufficientFunds   RPCErrorCode = -6
	ErrRPCWalletInvalidAccountName  RPCErrorCode = -11
	ErrRPCWalletKeypoolRanOut       RPCErrorCode = -12
	ErrRPCWalletUnlockNeeded        RPCErrorCode = -13
	ErrRPCWalletPassphraseIncorrect RPCErrorCode = -14
	ErrRPCWalletWrongEncState       RPCErrorCode = -15
	ErrRPCWalletEncryptionFailed    RPCErrorCode = -16
	ErrRPCWalletAlreadyUnlocked     RPCErrorCode = -17
)

// Specific Errors related to commands.  These are the ones a user of the RPC
// server are most likely to see.  Generally, the codes should match one of the
// more general errors above.
const (
	ErrRPCBlockNotFound     RPCErrorCode = -5
	ErrRPCBlockCount        RPCErrorCode = -5
	ErrRPCBestBlockHash     RPCErrorCode = -5
	ErrRPCDifficulty        RPCErrorCode = -5
	ErrRPCOutOfRange        RPCErrorCode = -1
	ErrRPCNoTxInfo          RPCErrorCode = -5
	ErrRPCNoCFIndex         RPCErrorCode = -5
	ErrRPCNoNewestBlockInfo RPCErrorCode = -5
	ErrRPCInvalidTxVout     RPCErrorCode = -5
	ErrRPCRawTxString       RPCErrorCode = -32602
	ErrRPCDecodeHexString   RPCErrorCode = -22
	ErrRPCTxError           RPCErrorCode = -25
	ErrRPCTxRejected        RPCErrorCode = -26
	ErrRPCTxAlreadyInChain  RPCErrorCode = -27
)

// Errors that are specific to pktd.
const (
	ErrRPCNoWallet      RPCErrorCode = -1
	ErrRPCUnimplemented RPCErrorCode = -1
)
