// Copyright (c) 2015 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

// NOTE: This file is intended to house the RPC commands that are supported by
// a wallet server with pktwallet extensions.

package btcjson

// CreateNewAccountCmd defines the createnewaccount JSON-RPC command.
type CreateNewAccountCmd struct {
	Account string
}

// NewCreateNewAccountCmd returns a new instance which can be used to issue a
// createnewaccount JSON-RPC command.
func NewCreateNewAccountCmd(account string) *CreateNewAccountCmd {
	return &CreateNewAccountCmd{
		Account: account,
	}
}

// RenameAccountCmd defines the renameaccount JSON-RPC command.
type RenameAccountCmd struct {
	OldAccount string
	NewAccount string
}

// NewRenameAccountCmd returns a new instance which can be used to issue a
// renameaccount JSON-RPC command.
func NewRenameAccountCmd(oldAccount, newAccount string) *RenameAccountCmd {
	return &RenameAccountCmd{
		OldAccount: oldAccount,
		NewAccount: newAccount,
	}
}

func init() {
	// The commands in this file are only usable with a wallet server.
	flags := UFWalletOnly

	MustRegisterCmd("createnewaccount", (*CreateNewAccountCmd)(nil), flags)
	MustRegisterCmd("renameaccount", (*RenameAccountCmd)(nil), flags)
}
