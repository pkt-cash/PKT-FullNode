// Copyright (c) 2014-2016 The btcsuite developers
// Copyright (c) 2019 Caleb James DeLisle
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockchain

import (
	"fmt"

	"github.com/pkt-cash/pktd/btcutil/er"
	"github.com/pkt-cash/pktd/wire/ruleerror"
)

// DeploymentError identifies an error that indicates a deployment ID was
// specified that does not exist.
type deploymentError0 uint32

// Error returns the assertion error as a human-readable string and satisfies
// the error interface.
func (e deploymentError0) Error() string {
	return fmt.Sprintf("deployment ID %d does not exist", uint32(e))
}

func DeploymentError(i uint32) er.R {
	return er.E(deploymentError0(i))
}

// AssertError identifies an error that indicates an internal code consistency
// issue and should be treated as a critical and unrecoverable error.
func AssertError(s string) er.R {
	return er.New("assertion failed: " + s)
}

// These are just aliases to the relevant ruleerror
// TODO(cjd): These should all go away and the relevant code should be
// refactored to use the ruleerror
var (
	// ErrDuplicateBlock indicates a block with the same hash already
	// exists.
	ErrDuplicateBlock = ruleerror.ErrDuplicateBlock

	// ErrBlockTooBig indicates the serialized block size exceeds the
	// maximum allowed size.
	ErrBlockTooBig = ruleerror.ErrBlockTooBig

	// ErrBlockWeightTooHigh indicates that the block's computed weight
	// metric exceeds the maximum allowed value.
	ErrBlockWeightTooHigh = ruleerror.ErrBlockWeightTooHigh

	// ErrBlockVersionTooOld indicates the block version is too old and is
	// no longer accepted since the majority of the network has upgraded
	// to a newer version.
	ErrBlockVersionTooOld = ruleerror.ErrBlockVersionTooOld

	// ErrTimeTooOld indicates the time is either before the median time of
	// the last several blocks per the chain consensus rules or prior to the
	// most recent checkpoint.
	ErrTimeTooOld = ruleerror.ErrTimeTooOld

	// ErrTimeTooNew indicates the time is too far in the future as compared
	// the current time.
	ErrTimeTooNew = ruleerror.ErrTimeTooNew

	// ErrDifficultyTooLow indicates the difficulty for the block is lower
	// than the difficulty required by the most recent checkpoint.
	ErrDifficultyTooLow = ruleerror.ErrDifficultyTooLow

	// ErrUnexpectedDifficulty indicates specified bits do not align with
	// the expected value either because it doesn't match the calculated
	// valued based on difficulty regarted rules or it is out of the valid
	// range.
	ErrUnexpectedDifficulty = ruleerror.ErrUnexpectedDifficulty

	// ErrHighHash indicates the block does not hash to a value which is
	// lower than the required target difficultly.
	ErrHighHash = ruleerror.ErrHighHash

	// ErrBadMerkleRoot indicates the calculated merkle root does not match
	// the expected value.
	ErrBadMerkleRoot = ruleerror.ErrBadMerkleRoot

	// ErrBadCheckpoint indicates a block that is expected to be at a
	// checkpoint height does not match the expected one.
	ErrBadCheckpoint = ruleerror.ErrBadCheckpoint

	// ErrForkTooOld indicates a block is attempting to fork the block chain
	// before the most recent checkpoint.
	ErrForkTooOld = ruleerror.ErrForkTooOld

	// ErrCheckpointTimeTooOld indicates a block has a timestamp before the
	// most recent checkpoint.
	ErrCheckpointTimeTooOld = ruleerror.ErrCheckpointTimeTooOld

	// ErrNoTransactions indicates the block does not have a least one
	// transaction.  A valid block must have at least the coinbase
	// transaction.
	ErrNoTransactions = ruleerror.ErrNoTransactions

	// ErrNoTxInputs indicates a transaction does not have any inputs.  A
	// valid transaction must have at least one input.
	ErrNoTxInputs = ruleerror.ErrNoTxInputs

	// ErrNoTxOutputs indicates a transaction does not have any outputs.  A
	// valid transaction must have at least one output.
	ErrNoTxOutputs = ruleerror.ErrNoTxOutputs

	// ErrTxTooBig indicates a transaction exceeds the maximum allowed size
	// when serialized.
	ErrTxTooBig = ruleerror.ErrTxTooBig

	// ErrDuplicateTxInputs indicates a transaction references the same
	// input more than once.
	ErrDuplicateTxInputs = ruleerror.ErrDuplicateTxInputs

	// ErrMissingTxOut indicates a transaction output referenced by an input
	// either does not exist or has already been spent.
	ErrMissingTxOut = ruleerror.ErrMissingTxOut

	// ErrUnfinalizedTx indicates a transaction has not been finalized.
	// A valid block may only contain finalized transactions.
	ErrUnfinalizedTx = ruleerror.ErrUnfinalizedTx

	// ErrDuplicateTx indicates a block contains an identical transaction
	// (or at least two transactions which hash to the same value).  A
	// valid block may only contain unique transactions.
	ErrDuplicateTx = ruleerror.ErrDuplicateTx

	// ErrOverwriteTx indicates a block contains a transaction that has
	// the same hash as a previous transaction which has not been fully
	// spent.
	ErrOverwriteTx = ruleerror.ErrOverwriteTx

	// ErrImmatureSpend indicates a transaction is attempting to spend a
	// coinbase that has not yet reached the required maturity.
	ErrImmatureSpend = ruleerror.ErrImmatureSpend

	// ErrSpendTooHigh indicates a transaction is attempting to spend more
	// value than the sum of all of its inputs.
	ErrSpendTooHigh = ruleerror.ErrSpendTooHigh

	// ErrBadFees indicates the total fees for a block are invalid due to
	// exceeding the maximum possible value.
	ErrBadFees = ruleerror.ErrBadFees

	// ErrTooManySigOps indicates the total number of signature operations
	// for a transaction or block exceed the maximum allowed limits.
	ErrTooManySigOps = ruleerror.ErrTooManySigOps

	// ErrFirstTxNotCoinbase indicates the first transaction in a block
	// is not a coinbase transaction.
	ErrFirstTxNotCoinbase = ruleerror.ErrFirstTxNotCoinbase

	// ErrMultipleCoinbases indicates a block contains more than one
	// coinbase transaction.
	ErrMultipleCoinbases = ruleerror.ErrMultipleCoinbases

	// ErrBadCoinbaseScriptLen indicates the length of the signature script
	// for a coinbase transaction is not within the valid range.
	ErrBadCoinbaseScriptLen = ruleerror.ErrBadCoinbaseScriptLen

	// ErrBadCoinbaseValue indicates the amount of a coinbase value does
	// not match the expected value of the subsidy plus the sum of all fees.
	ErrBadCoinbaseValue = ruleerror.ErrBadCoinbaseValue

	// ErrMissingCoinbaseHeight indicates the coinbase transaction for a
	// block does not start with the serialized block block height as
	// required for version 2 and higher blocks.
	ErrMissingCoinbaseHeight = ruleerror.ErrMissingCoinbaseHeight

	// ErrBadCoinbaseHeight indicates the serialized block height in the
	// coinbase transaction for version 2 and higher blocks does not match
	// the expected value.
	ErrBadCoinbaseHeight = ruleerror.ErrBadCoinbaseHeight

	// ErrBadCoinbaseTax indicates that the tax was paid to the wrong
	// network steward key.
	ErrBadCoinbaseNetworkSteward = ruleerror.ErrBadCoinbaseNetworkSteward

	// ErrScriptMalformed indicates a transaction script is malformed in
	// some way.  For example, it might be longer than the maximum allowed
	// length or fail to parse.
	ErrScriptMalformed = ruleerror.ErrScriptMalformed

	// ErrScriptValidation indicates the result of executing transaction
	// script failed.  The error covers any failure when executing scripts
	// such signature verification failures and execution past the end of
	// the stack.
	ErrScriptValidation = ruleerror.ErrScriptValidation

	// ErrUnexpectedWitness indicates that a block includes transactions
	// with witness data, but doesn't also have a witness commitment within
	// the coinbase transaction.
	ErrUnexpectedWitness = ruleerror.ErrUnexpectedWitness

	// ErrInvalidWitnessCommitment indicates that a block's witness
	// commitment is not well formed.
	ErrInvalidWitnessCommitment = ruleerror.ErrInvalidWitnessCommitment

	// ErrWitnessCommitmentMismatch indicates that the witness commitment
	// included in the block's coinbase transaction doesn't match the
	// manually computed witness commitment.
	ErrWitnessCommitmentMismatch = ruleerror.ErrWitnessCommitmentMismatch

	// ErrPreviousBlockUnknown indicates that the previous block is not known.
	ErrPreviousBlockUnknown = ruleerror.ErrPreviousBlockUnknown

	// ErrInvalidAncestorBlock indicates that an ancestor of this block has
	// already failed validation.
	ErrInvalidAncestorBlock = ruleerror.ErrInvalidAncestorBlock

	// ErrPrevBlockNotBest indicates that the block's previous block is not the
	// current chain tip. This is not a block validation rule, but is required
	// for block proposals submitted via getblocktemplate RPC.
	ErrPrevBlockNotBest = ruleerror.ErrPrevBlockNotBest

	// ErrBadPow indicates that the proof of work was somehow malformed
	ErrBadPow = ruleerror.ErrBadPow

	// ErrPowCannotVerify indicates that the pow cannot be verified because there
	// is a missing block header required to check one of the announcements.
	ErrPowCannotVerify = ruleerror.ErrPowCannotVerify

	// ErrNetworkStewardOldSpend indicates that a transaction tried to spend
	// coins in the network steward wallet which are older than the age limit
	// for network steward payouts.
	ErrNetworkStewardOldSpend = ruleerror.ErrNetworkStewardOldSpend
)

// ruleError creates an RuleError given a set of arguments.
func ruleError(c *er.ErrorCode, desc string) er.R {
	return c.New(desc, nil)
}
