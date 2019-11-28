// Copyright (c) 2014 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package waddrmgr

import (
	"strconv"

	"github.com/pkt-cash/pktd/btcutil/er"
	"github.com/pkt-cash/pktd/btcutil/hdkeychain"
)

// ManagerErr provides a single type for errors that can happen during address
// manager operation.  It is used to indicate several types of failures
// including errors with caller requests such as invalid accounts or requesting
// private keys against a locked address manager, errors with the database
// (ErrDatabase), errors with key chain derivation (ErrKeyChain), and errors
// related to crypto (ErrCrypto).
//
// The caller can use type assertions to determine if an error is a ManagerError
// and access the ErrorCode field to ascertain the specific reason for the
// failure.
//
// The ErrDatabase, ErrKeyChain, and ErrCrypto error codes will also have the
// Err field set with the underlying error.
var ManagerErr er.ErrorType = er.NewErrorType("waddrmgr.ManagerErr")

var (
	// ErrDatabase indicates an error with the underlying database.  When
	// this error code is set, the Err field of the ManagerError will be
	// set to the underlying error returned from the database.
	ErrDatabase *er.ErrorCode = ManagerErr.Code("ErrDatabase")

	// ErrUpgrade indicates the manager needs to be upgraded.  This should
	// not happen in practice unless the version number has been increased
	// and there is not yet any code written to upgrade.
	ErrUpgrade = ManagerErr.Code("ErrUpgrade")

	// ErrKeyChain indicates an error with the key chain typically either
	// due to the inability to create an extended key or deriving a child
	// extended key.  When this error code is set, the Err field of the
	// ManagerError will be set to the underlying error.
	ErrKeyChain = ManagerErr.Code("ErrUpgrade")

	// ErrCrypto indicates an error with the cryptography related operations
	// such as decrypting or encrypting data, parsing an EC public key,
	// or deriving a secret key from a password.  When this error code is
	// set, the Err field of the ManagerError will be set to the underlying
	// error.
	ErrCrypto = ManagerErr.Code("ErrCrypto")

	// ErrInvalidKeyType indicates an error where an invalid crypto
	// key type has been selected.
	ErrInvalidKeyType = ManagerErr.Code("ErrInvalidKeyType")

	// ErrNoExist indicates that the specified database does not exist.
	ErrNoExist = ManagerErr.Code("ErrNoExist")

	// ErrAlreadyExists indicates that the specified database already exists.
	ErrAlreadyExists = ManagerErr.CodeWithDetail("ErrAlreadyExists",
		"the specified address manager already exists")

	// ErrCoinTypeTooHigh indicates that the coin type specified in the provided
	// network parameters is higher than the max allowed value as defined
	// by the maxCoinType constant.
	ErrCoinTypeTooHigh = ManagerErr.CodeWithDetail("ErrCoinTypeTooHigh",
		"coin type may not exceed "+strconv.FormatUint(hdkeychain.HardenedKeyStart-1, 10))

	// ErrAccountNumTooHigh indicates that the specified account number is higher
	// than the max allowed value as defined by the MaxAccountNum constant.
	ErrAccountNumTooHigh = ManagerErr.CodeWithDetail("ErrAccountNumTooHigh",
		"account number may not exceed "+strconv.FormatUint(hdkeychain.HardenedKeyStart-1, 10))

	// ErrLocked indicates that an operation, which requires the account
	// manager to be unlocked, was requested on a locked account manager.
	ErrLocked = ManagerErr.CodeWithDetail("ErrLocked",
		"address manager is locked")

	// ErrWatchingOnly indicates that an operation, which requires the
	// account manager to have access to private data, was requested on
	// a watching-only account manager.
	ErrWatchingOnly = ManagerErr.CodeWithDetail("ErrWatchingOnly",
		"address manager is watching-only")

	// ErrInvalidAccount indicates that the requested account is not valid.
	ErrInvalidAccount = ManagerErr.Code("ErrInvalidAccount")

	// ErrAddressNotFound indicates that the requested address is not known to
	// the account manager.
	ErrAddressNotFound = ManagerErr.Code("ErrAddressNotFound")

	// ErrAccountNotFound indicates that the requested account is not known to
	// the account manager.
	ErrAccountNotFound = ManagerErr.Code("ErrAccountNotFound")

	// ErrDuplicateAddress indicates an address already exists.
	ErrDuplicateAddress = ManagerErr.Code("ErrDuplicateAddress")

	// ErrDuplicateAccount indicates an account already exists.
	ErrDuplicateAccount = ManagerErr.Code("ErrDuplicateAccount")

	// ErrTooManyAddresses indicates that more than the maximum allowed number of
	// addresses per account have been requested.
	ErrTooManyAddresses = ManagerErr.Code("ErrTooManyAddresses")

	// ErrWrongPassphrase indicates that the specified passphrase is incorrect.
	// This could be for either public or private master keys.
	ErrWrongPassphrase = ManagerErr.Code("ErrWrongPassphrase")

	// ErrWrongNet indicates that the private key to be imported is not for the
	// the same network the account manager is configured for.
	ErrWrongNet = ManagerErr.Code("ErrWrongNet")

	// ErrEmptyPassphrase indicates that the private passphrase was refused
	// due to being empty.
	ErrEmptyPassphrase = ManagerErr.Code("ErrEmptyPassphrase")

	// ErrScopeNotFound is returned when a target scope cannot be found
	// within the database.
	ErrScopeNotFound = ManagerErr.Code("ErrScopeNotFound")

	// ErrBirthdayBlockNotSet is returned when we attempt to retrieve the
	// wallet's birthday but it has not been set yet.
	ErrBirthdayBlockNotSet = ManagerErr.Code("ErrBirthdayBlockNotSet")

	// ErrBlockNotFound is returned when we attempt to retrieve the hash for
	// a block that we do not know of.
	ErrBlockNotFound = ManagerErr.Code("ErrBlockNotFound")
)

// managerError creates a ManagerError given a set of arguments.
func managerError(c *er.ErrorCode, desc string, err er.R) er.R {
	return c.New(desc, err)
}
