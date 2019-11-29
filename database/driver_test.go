// Copyright (c) 2015-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package database_test

import (
	"fmt"
	"testing"

	"github.com/pkt-cash/pktd/btcutil/er"
	"github.com/pkt-cash/pktd/btcutil/util"
	"github.com/pkt-cash/pktd/database"
	_ "github.com/pkt-cash/pktd/database/ffldb"
)

// TestAddDuplicateDriver ensures that adding a duplicate driver does not
// overwrite an existing one.
func TestAddDuplicateDriver(t *testing.T) {
	supportedDrivers := database.SupportedDrivers()
	if len(supportedDrivers) == 0 {
		t.Errorf("no backends to test")
		return
	}
	dbType := supportedDrivers[0]

	// bogusCreateDB is a function which acts as a bogus create and open
	// driver function and intentionally returns a failure that can be
	// detected if the interface allows a duplicate driver to overwrite an
	// existing one.
	bogusCreateDB := func(args ...interface{}) (database.DB, er.R) {
		return nil, er.Errorf("duplicate driver allowed for database "+
			"type [%v]", dbType)
	}

	// Create a driver that tries to replace an existing one.  Set its
	// create and open functions to a function that causes a test failure if
	// they are invoked.
	driver := database.Driver{
		DbType: dbType,
		Create: bogusCreateDB,
		Open:   bogusCreateDB,
	}
	testName := "duplicate driver registration"
	err := database.RegisterDriver(driver)
	if !util.CheckError(t, testName, err, database.ErrDbTypeRegistered) {
		return
	}
}

// TestCreateOpenFail ensures that errors which occur while opening or closing
// a database are handled properly.
func TestCreateOpenFail(t *testing.T) {
	// bogusCreateDB is a function which acts as a bogus create and open
	// driver function that intentionally returns a failure which can be
	// detected.
	dbType := "createopenfail"
	openError := fmt.Errorf("failed to create or open database for "+
		"database type [%v]", dbType)
	bogusCreateDB := func(args ...interface{}) (database.DB, er.R) {
		return nil, er.E(openError)
	}

	// Create and add driver that intentionally fails when created or opened
	// to ensure errors on database open and create are handled properly.
	driver := database.Driver{
		DbType: dbType,
		Create: bogusCreateDB,
		Open:   bogusCreateDB,
	}
	database.RegisterDriver(driver)

	// Ensure creating a database with the new type fails with the expected
	// error.
	_, err := database.Create(dbType)
	if er.Wrapped(err) != openError {
		t.Errorf("expected error not received - got: %v, want %v", err,
			openError)
		return
	}

	// Ensure opening a database with the new type fails with the expected
	// error.
	_, err = database.Open(dbType)
	if er.Wrapped(err) != openError {
		t.Errorf("expected error not received - got: %v, want %v", err,
			openError)
		return
	}
}

// TestCreateOpenUnsupported ensures that attempting to create or open an
// unsupported database type is handled properly.
func TestCreateOpenUnsupported(t *testing.T) {
	// Ensure creating a database with an unsupported type fails with the
	// expected error.
	testName := "create with unsupported database type"
	dbType := "unsupported"
	_, err := database.Create(dbType)
	if !util.CheckError(t, testName, err, database.ErrDbUnknownType) {
		return
	}

	// Ensure opening a database with the an unsupported type fails with the
	// expected error.
	testName = "open with unsupported database type"
	_, err = database.Open(dbType)
	if !util.CheckError(t, testName, err, database.ErrDbUnknownType) {
		return
	}
}
