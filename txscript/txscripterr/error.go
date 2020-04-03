// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package txscripterr

import (
	"github.com/pkt-cash/pktd/btcutil/er"
)

// Err identifies a kind of script error.
var Err er.ErrorType = er.NewErrorType("txscript.Err")

// These constants are used to identify a specific Error.
var (
	// ErrInternal is returned if internal consistency checks fail.  In
	// practice this error should never be seen as it would mean there is an
	// error in the engine logic.
	ErrInternal = Err.Code("ErrInternal")

	// ---------------------------------------
	// Failures related to improper API usage.
	// ---------------------------------------

	// ErrInvalidFlags is returned when the passed flags to NewEngine
	// contain an invalid combination.
	ErrInvalidFlags = Err.Code("ErrInvalidFlags")

	// ErrInvalidIndex is returned when an out-of-bounds index is passed to
	// a function.
	ErrInvalidIndex = Err.Code("ErrInvalidIndex")

	// ErrUnsupportedAddress is returned when a concrete type that
	// implements a btcutil.Address is not a supported type.
	ErrUnsupportedAddress = Err.Code("ErrUnsupportedAddress")

	// ErrNotMultisigScript is returned from CalcMultiSigStats when the
	// provided script is not a multisig script.
	ErrNotMultisigScript = Err.Code("ErrNotMultisigScript")

	// ErrTooManyRequiredSigs is returned from MultiSigScript when the
	// specified number of required signatures is larger than the number of
	// provided public keys.
	ErrTooManyRequiredSigs = Err.Code("ErrTooManyRequiredSigs")

	// ErrTooMuchNullData is returned from NullDataScript when the length of
	// the provided data exceeds MaxDataCarrierSize.
	ErrTooMuchNullData = Err.Code("ErrTooMuchNullData")

	// ------------------------------------------
	// Failures related to final execution state.
	// ------------------------------------------

	// ErrEarlyReturn is returned when OP_RETURN is executed in the script.
	ErrEarlyReturn = Err.Code("ErrEarlyReturn")

	// ErrEmptyStack is returned when the script evaluated without error,
	// but terminated with an empty top stack element.
	ErrEmptyStack = Err.Code("ErrEmptyStack")

	// ErrEvalFalse is returned when the script evaluated without error but
	// terminated with a false top stack element.
	ErrEvalFalse = Err.Code("ErrEvalFalse")

	// ErrScriptUnfinished is returned when CheckErrorCondition is called on
	// a script that has not finished executing.
	ErrScriptUnfinished = Err.Code("ErrScriptUnfinished")

	// ErrScriptDone is returned when an attempt to execute an opcode is
	// made once all of them have already been executed.  This can happen
	// due to things such as a second call to Execute or calling Step after
	// all opcodes have already been executed.
	ErrInvalidProgramCounter = Err.Code("ErrInvalidProgramCounter")

	// -----------------------------------------------------
	// Failures related to exceeding maximum allowed limits.
	// -----------------------------------------------------

	// ErrScriptTooBig is returned if a script is larger than MaxScriptSize.
	ErrScriptTooBig = Err.Code("ErrScriptTooBig")

	// ErrElementTooBig is returned if the size of an element to be pushed
	// to the stack is over MaxScriptElementSize.
	ErrElementTooBig = Err.Code("ErrElementTooBig")

	// ErrTooManyOperations is returned if a script has more than
	// MaxOpsPerScript opcodes that do not push data.
	ErrTooManyOperations = Err.Code("ErrTooManyOperations")

	// ErrStackOverflow is returned when stack and altstack combined depth
	// is over the limit.
	ErrStackOverflow = Err.Code("ErrStackOverflow")

	// ErrInvalidPubKeyCount is returned when the number of public keys
	// specified for a multsig is either negative or greater than
	// MaxPubKeysPerMultiSig.
	ErrInvalidPubKeyCount = Err.Code("ErrInvalidPubKeyCount")

	// ErrInvalidSignatureCount is returned when the number of signatures
	// specified for a multisig is either negative or greater than the
	// number of public keys.
	ErrInvalidSignatureCount = Err.Code("ErrInvalidSignatureCount")

	// ErrNumberTooBig is returned when the argument for an opcode that
	// expects numeric input is larger than the expected maximum number of
	// bytes.  For the most part, opcodes that deal with stack manipulation
	// via offsets, arithmetic, numeric comparison, and boolean logic are
	// those that this applies to.  However, any opcode that expects numeric
	// input may fail with this code.
	ErrNumberTooBig = Err.Code("ErrNumberTooBig")

	// --------------------------------------------
	// Failures related to verification operations.
	// --------------------------------------------

	// ErrVerify is returned when OP_VERIFY is encountered in a script and
	// the top item on the data stack does not evaluate to true.
	ErrVerify = Err.Code("ErrVerify")

	// ErrEqualVerify is returned when OP_EQUALVERIFY is encountered in a
	// script and the top item on the data stack does not evaluate to true.
	ErrEqualVerify = Err.Code("ErrEqualVerify")

	// ErrNumEqualVerify is returned when OP_NUMEQUALVERIFY is encountered
	// in a script and the top item on the data stack does not evaluate to
	// true.
	ErrNumEqualVerify = Err.Code("ErrNumEqualVerify")

	// ErrCheckSigVerify is returned when OP_CHECKSIGVERIFY is encountered
	// in a script and the top item on the data stack does not evaluate to
	// true.
	ErrCheckSigVerify = Err.Code("ErrCheckSigVerify")

	// ErrCheckSigVerify is returned when OP_CHECKMULTISIGVERIFY is
	// encountered in a script and the top item on the data stack does not
	// evaluate to true.
	ErrCheckMultiSigVerify = Err.Code("ErrCheckMultiSigVerify")

	// --------------------------------------------
	// Failures related to improper use of opcodes.
	// --------------------------------------------

	// ErrDisabledOpcode is returned when a disabled opcode is encountered
	// in a script.
	ErrDisabledOpcode = Err.Code("ErrDisabledOpcode")

	// ErrReservedOpcode is returned when an opcode marked as reserved
	// is encountered in a script.
	ErrReservedOpcode = Err.Code("ErrReservedOpcode")

	// ErrMalformedPush is returned when a data push opcode tries to push
	// more bytes than are left in the script.
	ErrMalformedPush = Err.Code("ErrMalformedPush")

	// ErrInvalidStackOperation is returned when a stack operation is
	// attempted with a number that is invalid for the current stack size.
	ErrInvalidStackOperation = Err.Code("ErrInvalidStackOperation")

	// ErrUnbalancedConditional is returned when an OP_ELSE or OP_ENDIF is
	// encountered in a script without first having an OP_IF or OP_NOTIF or
	// the end of script is reached without encountering an OP_ENDIF when
	// an OP_IF or OP_NOTIF was previously encountered.
	ErrUnbalancedConditional = Err.Code("ErrUnbalancedConditional")

	// ---------------------------------
	// Failures related to malleability.
	// ---------------------------------

	// ErrMinimalData is returned when the ScriptVerifyMinimalData flag
	// is set and the script contains push operations that do not use
	// the minimal opcode required.
	ErrMinimalData = Err.Code("ErrMinimalData")

	// ErrInvalidSigHashType is returned when a signature hash type is not
	// one of the supported types.
	ErrInvalidSigHashType = Err.Code("ErrInvalidSigHashType")

	// ErrSigTooShort is returned when a signature that should be a
	// canonically-encoded DER signature is too short.
	ErrSigTooShort = Err.Code("ErrSigTooShort")

	// ErrSigTooLong is returned when a signature that should be a
	// canonically-encoded DER signature is too long.
	ErrSigTooLong = Err.Code("ErrSigTooLong")

	// ErrSigInvalidSeqID is returned when a signature that should be a
	// canonically-encoded DER signature does not have the expected ASN.1
	// sequence ID.
	ErrSigInvalidSeqID = Err.Code("ErrSigInvalidSeqID")

	// ErrSigInvalidDataLen is returned a signature that should be a
	// canonically-encoded DER signature does not specify the correct number
	// of remaining bytes for the R and S portions.
	ErrSigInvalidDataLen = Err.Code("ErrSigInvalidDataLen")

	// ErrSigMissingSTypeID is returned a signature that should be a
	// canonically-encoded DER signature does not provide the ASN.1 type ID
	// for S.
	ErrSigMissingSTypeID = Err.Code("ErrSigMissingSTypeID")

	// ErrSigMissingSLen is returned when a signature that should be a
	// canonically-encoded DER signature does not provide the length of S.
	ErrSigMissingSLen = Err.Code("ErrSigMissingSLen")

	// ErrSigInvalidSLen is returned a signature that should be a
	// canonically-encoded DER signature does not specify the correct number
	// of bytes for the S portion.
	ErrSigInvalidSLen = Err.Code("ErrSigInvalidSLen")

	// ErrSigInvalidRIntID is returned when a signature that should be a
	// canonically-encoded DER signature does not have the expected ASN.1
	// integer ID for R.
	ErrSigInvalidRIntID = Err.Code("ErrSigInvalidRIntID")

	// ErrSigZeroRLen is returned when a signature that should be a
	// canonically-encoded DER signature has an R length of zero.
	ErrSigZeroRLen = Err.Code("ErrSigZeroRLen")

	// ErrSigNegativeR is returned when a signature that should be a
	// canonically-encoded DER signature has a negative value for R.
	ErrSigNegativeR = Err.Code("ErrSigNegativeR")

	// ErrSigTooMuchRPadding is returned when a signature that should be a
	// canonically-encoded DER signature has too much padding for R.
	ErrSigTooMuchRPadding = Err.Code("ErrSigTooMuchRPadding")

	// ErrSigInvalidSIntID is returned when a signature that should be a
	// canonically-encoded DER signature does not have the expected ASN.1
	// integer ID for S.
	ErrSigInvalidSIntID = Err.Code("ErrSigInvalidSIntID")

	// ErrSigZeroSLen is returned when a signature that should be a
	// canonically-encoded DER signature has an S length of zero.
	ErrSigZeroSLen = Err.Code("ErrSigZeroSLen")

	// ErrSigNegativeS is returned when a signature that should be a
	// canonically-encoded DER signature has a negative value for S.
	ErrSigNegativeS = Err.Code("ErrSigNegativeS")

	// ErrSigTooMuchSPadding is returned when a signature that should be a
	// canonically-encoded DER signature has too much padding for S.
	ErrSigTooMuchSPadding = Err.Code("ErrSigTooMuchSPadding")

	// ErrSigHighS is returned when the ScriptVerifyLowS flag is set and the
	// script contains any signatures whose S values are higher than the
	// half order.
	ErrSigHighS = Err.Code("ErrSigHighS")

	// ErrNotPushOnly is returned when a script that is required to only
	// push data to the stack performs other operations.  A couple of cases
	// where this applies is for a pay-to-script-hash signature script when
	// bip16 is active and when the ScriptVerifySigPushOnly flag is set.
	ErrNotPushOnly = Err.Code("ErrNotPushOnly")

	// ErrSigNullDummy is returned when the ScriptStrictMultiSig flag is set
	// and a multisig script has anything other than 0 for the extra dummy
	// argument.
	ErrSigNullDummy = Err.Code("ErrSigNullDummy")

	// ErrPubKeyType is returned when the ScriptVerifyStrictEncoding
	// flag is set and the script contains invalid public keys.
	ErrPubKeyType = Err.Code("ErrPubKeyType")

	// ErrCleanStack is returned when the ScriptVerifyCleanStack flag
	// is set, and after evalution, the stack does not contain only a
	// single element.
	ErrCleanStack = Err.Code("ErrCleanStack")

	// ErrNullFail is returned when the ScriptVerifyNullFail flag is
	// set and signatures are not empty on failed checksig or checkmultisig
	// operations.
	ErrNullFail = Err.Code("ErrNullFail")

	// ErrWitnessMalleated is returned if ScriptVerifyWitness is set and a
	// native p2wsh program is encountered which has a non-empty sigScript.
	ErrWitnessMalleated = Err.Code("ErrWitnessMalleated")

	// ErrWitnessMalleatedP2SH is returned if ScriptVerifyWitness if set
	// and the validation logic for nested p2sh encounters a sigScript
	// which isn't *exactyl* a datapush of the witness program.
	ErrWitnessMalleatedP2SH = Err.Code("ErrWitnessMalleatedP2SH")

	// -------------------------------
	// Failures related to soft forks.
	// -------------------------------

	// ErrDiscourageUpgradableNOPs is returned when the
	// ScriptDiscourageUpgradableNops flag is set and a NOP opcode is
	// encountered in a script.
	ErrDiscourageUpgradableNOPs = Err.Code("ErrDiscourageUpgradableNOPs")

	// ErrNegativeLockTime is returned when a script contains an opcode that
	// interprets a negative lock time.
	ErrNegativeLockTime = Err.Code("ErrNegativeLockTime")

	// ErrUnsatisfiedLockTime is returned when a script contains an opcode
	// that involves a lock time and the required lock time has not been
	// reached.
	ErrUnsatisfiedLockTime = Err.Code("ErrUnsatisfiedLockTime")

	// ErrMinimalIf is returned if ScriptVerifyWitness is set and the
	// operand of an OP_IF/OP_NOF_IF are not either an empty vector or
	// [0x01].
	ErrMinimalIf = Err.Code("ErrMinimalIf")

	// ErrDiscourageUpgradableWitnessProgram is returned if
	// ScriptVerifyWitness is set and the versino of an executing witness
	// program is outside the set of currently defined witness program
	// vesions.
	ErrDiscourageUpgradableWitnessProgram = Err.Code("ErrDiscourageUpgradableWitnessProgram")

	// ----------------------------------------
	// Failures related to segregated witness.
	// ----------------------------------------

	// ErrWitnessProgramEmpty is returned if ScriptVerifyWitness is set and
	// the witness stack itself is empty.
	ErrWitnessProgramEmpty = Err.Code("ErrWitnessProgramEmpty")

	// ErrWitnessProgramMismatch is returned if ScriptVerifyWitness is set
	// and the witness itself for a p2wkh witness program isn't *exactly* 2
	// items or if the witness for a p2wsh isn't the sha255 of the witness
	// script.
	ErrWitnessProgramMismatch = Err.Code("ErrWitnessProgramMismatch")

	// ErrWitnessProgramWrongLength is returned if ScriptVerifyWitness is
	// set and the length of the witness program violates the length as
	// dictated by the current witness version.
	ErrWitnessProgramWrongLength = Err.Code("ErrWitnessProgramWrongLength")

	// ErrWitnessUnexpected is returned if ScriptVerifyWitness is set and a
	// transaction includes witness data but doesn't spend an which is a
	// witness program (nested or native).
	ErrWitnessUnexpected = Err.Code("ErrWitnessUnexpected")

	// ErrWitnessPubKeyType is returned if ScriptVerifyWitness is set and
	// the public key used in either a check-sig or check-multi-sig isn't
	// serialized in a compressed format.
	ErrWitnessPubKeyType = Err.Code("ErrWitnessPubKeyType")
)

// ScriptError creates an Error given a set of arguments.
func ScriptError(c *er.ErrorCode, desc string) er.R {
	return c.New(desc, nil)
}
