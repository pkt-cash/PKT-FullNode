// Copyright (c) 2015-2017 The btcsuite developers
// Copyright (c) 2015-2016 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wtxmgr

import (
	"github.com/pkt-cash/pktd/btcutil/er"
)

// Err identifies a category of error.
var Err er.ErrorType = er.NewErrorType("wtxmgr.Err")

// These constants are used to identify a specific Error.
var (
	// ErrDatabase indicates an error with the underlying database.  When
	// this error code is set, the Err field of the Error will be
	// set to the underlying error returned from the database.
	ErrDatabase = Err.Code("ErrDatabase")

	// ErrData describes an error where data stored in the transaction
	// database is incorrect.  This may be due to missing values, values of
	// wrong sizes, or data from different buckets that is inconsistent with
	// itself.  Recovering from an ErrData requires rebuilding all
	// transaction history or manual database surgery.  If the failure was
	// not due to data corruption, this error category indicates a
	// programming error in this package.
	ErrData = Err.Code("ErrData")

	// ErrInput describes an error where the variables passed into this
	// function by the caller are obviously incorrect.  Examples include
	// passing transactions which do not serialize, or attempting to insert
	// a credit at an index for which no transaction output exists.
	ErrInput = Err.Code("ErrInput")

	// ErrAlreadyExists describes an error where creating the store cannot
	// continue because a store already exists in the namespace.
	ErrAlreadyExists = Err.Code("ErrAlreadyExists")

	// ErrNoExists describes an error where the store cannot be opened due to
	// it not already existing in the namespace.  This error should be
	// handled by creating a new store.
	ErrNoExists = Err.Code("ErrNoExists")

	// ErrNeedsUpgrade describes an error during store opening where the
	// database contains an older version of the store.
	ErrNeedsUpgrade = Err.Code("ErrNeedsUpgrade")

	// ErrUnknownVersion describes an error where the store already exists
	// but the database version is newer than latest version known to this
	// software.  This likely indicates an outdated binary.
	ErrUnknownVersion = Err.Code("ErrUnknownVersion")
)

// This exists for compatibility to reduce how much code should be refactored.
// It's better to just use <errorcode>.New()
func storeError(c *er.ErrorCode, desc string, err er.R) er.R {
	return c.New(desc, err)
}
