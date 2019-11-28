// Copyright (c) 2014-2017 The btcsuite developers
// Copyright (c) 2015-2017 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package rpcclient

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"

	"github.com/pkt-cash/pktd/btcjson"
	"github.com/pkt-cash/pktd/btcutil"
	"github.com/pkt-cash/pktd/btcutil/er"
	"github.com/pkt-cash/pktd/chaincfg/chainhash"
	"github.com/pkt-cash/pktd/wire"
)

// FutureDebugLevelResult is a future promise to deliver the result of a
// DebugLevelAsync RPC invocation (or an applicable error).
type FutureDebugLevelResult chan *response

// Receive waits for the response promised by the future and returns the result
// of setting the debug logging level to the passed level specification or the
// list of of the available subsystems for the special keyword 'show'.
func (r FutureDebugLevelResult) Receive() (string, er.R) {
	res, err := receiveFuture(r)
	if err != nil {
		return "", err
	}

	// Unmashal the result as a string.
	var result string
	errr := json.Unmarshal(res, &result)
	if errr != nil {
		return "", er.E(errr)
	}
	return result, nil
}

// DebugLevelAsync returns an instance of a type that can be used to get the
// result of the RPC at some future time by invoking the Receive function on
// the returned instance.
//
// See DebugLevel for the blocking version and more details.
//
// NOTE: This is a pktd extension.
func (c *Client) DebugLevelAsync(levelSpec string) FutureDebugLevelResult {
	cmd := btcjson.NewDebugLevelCmd(levelSpec)
	return c.sendCmd(cmd)
}

// DebugLevel dynamically sets the debug logging level to the passed level
// specification.
//
// The levelspec can be either a debug level or of the form:
// 	<subsystem>=<level>,<subsystem2>=<level2>,...
//
// Additionally, the special keyword 'show' can be used to get a list of the
// available subsystems.
//
// NOTE: This is a pktd extension.
func (c *Client) DebugLevel(levelSpec string) (string, er.R) {
	return c.DebugLevelAsync(levelSpec).Receive()
}

// FutureCreateEncryptedWalletResult is a future promise to deliver the error
// result of a CreateEncryptedWalletAsync RPC invocation.
type FutureCreateEncryptedWalletResult chan *response

// Receive waits for and returns the error response promised by the future.
func (r FutureCreateEncryptedWalletResult) Receive() er.R {
	_, err := receiveFuture(r)
	return err
}

// CreateEncryptedWalletAsync returns an instance of a type that can be used to
// get the result of the RPC at some future time by invoking the Receive
// function on the returned instance.
//
// See CreateEncryptedWallet for the blocking version and more details.
//
// NOTE: This is a pktwallet extension.
func (c *Client) CreateEncryptedWalletAsync(passphrase string) FutureCreateEncryptedWalletResult {
	cmd := btcjson.NewCreateEncryptedWalletCmd(passphrase)
	return c.sendCmd(cmd)
}

// CreateEncryptedWallet requests the creation of an encrypted wallet.  Wallets
// managed by pktwallet are only written to disk with encrypted private keys,
// and generating wallets on the fly is impossible as it requires user input for
// the encryption passphrase.  This RPC specifies the passphrase and instructs
// the wallet creation.  This may error if a wallet is already opened, or the
// new wallet cannot be written to disk.
//
// NOTE: This is a pktwallet extension.
func (c *Client) CreateEncryptedWallet(passphrase string) er.R {
	return c.CreateEncryptedWalletAsync(passphrase).Receive()
}

// FutureListAddressTransactionsResult is a future promise to deliver the result
// of a ListAddressTransactionsAsync RPC invocation (or an applicable error).
type FutureListAddressTransactionsResult chan *response

// Receive waits for the response promised by the future and returns information
// about all transactions associated with the provided addresses.
func (r FutureListAddressTransactionsResult) Receive() ([]btcjson.ListTransactionsResult, er.R) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal the result as an array of listtransactions objects.
	var transactions []btcjson.ListTransactionsResult
	errr := json.Unmarshal(res, &transactions)
	if errr != nil {
		return nil, er.E(errr)
	}
	return transactions, nil
}

// ListAddressTransactionsAsync returns an instance of a type that can be used
// to get the result of the RPC at some future time by invoking the Receive
// function on the returned instance.
//
// See ListAddressTransactions for the blocking version and more details.
//
// NOTE: This is a pktd extension.
func (c *Client) ListAddressTransactionsAsync(addresses []btcutil.Address, account string) FutureListAddressTransactionsResult {
	// Convert addresses to strings.
	addrs := make([]string, 0, len(addresses))
	for _, addr := range addresses {
		addrs = append(addrs, addr.EncodeAddress())
	}
	cmd := btcjson.NewListAddressTransactionsCmd(addrs, &account)
	return c.sendCmd(cmd)
}

// ListAddressTransactions returns information about all transactions associated
// with the provided addresses.
//
// NOTE: This is a pktwallet extension.
func (c *Client) ListAddressTransactions(addresses []btcutil.Address, account string) ([]btcjson.ListTransactionsResult, er.R) {
	return c.ListAddressTransactionsAsync(addresses, account).Receive()
}

// FutureGetBestBlockResult is a future promise to deliver the result of a
// GetBestBlockAsync RPC invocation (or an applicable error).
type FutureGetBestBlockResult chan *response

// Receive waits for the response promised by the future and returns the hash
// and height of the block in the longest (best) chain.
func (r FutureGetBestBlockResult) Receive() (*chainhash.Hash, int32, er.R) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, 0, err
	}

	// Unmarshal result as a getbestblock result object.
	var bestBlock btcjson.GetBestBlockResult
	errr := json.Unmarshal(res, &bestBlock)
	if errr != nil {
		return nil, 0, er.E(errr)
	}

	// Convert to hash from string.
	hash, err := chainhash.NewHashFromStr(bestBlock.Hash)
	if err != nil {
		return nil, 0, err
	}

	return hash, bestBlock.Height, nil
}

// GetBestBlockAsync returns an instance of a type that can be used to get the
// result of the RPC at some future time by invoking the Receive function on the
// returned instance.
//
// See GetBestBlock for the blocking version and more details.
//
// NOTE: This is a pktd extension.
func (c *Client) GetBestBlockAsync() FutureGetBestBlockResult {
	cmd := btcjson.NewGetBestBlockCmd()
	return c.sendCmd(cmd)
}

// GetBestBlock returns the hash and height of the block in the longest (best)
// chain.
//
// NOTE: This is a pktd extension.
func (c *Client) GetBestBlock() (*chainhash.Hash, int32, er.R) {
	return c.GetBestBlockAsync().Receive()
}

// FutureGetCurrentNetResult is a future promise to deliver the result of a
// GetCurrentNetAsync RPC invocation (or an applicable error).
type FutureGetCurrentNetResult chan *response

// Receive waits for the response promised by the future and returns the network
// the server is running on.
func (r FutureGetCurrentNetResult) Receive() (wire.BitcoinNet, er.R) {
	res, err := receiveFuture(r)
	if err != nil {
		return 0, err
	}

	// Unmarshal result as an int64.
	var net int64
	errr := json.Unmarshal(res, &net)
	if errr != nil {
		return 0, er.E(errr)
	}

	return wire.BitcoinNet(net), nil
}

// GetCurrentNetAsync returns an instance of a type that can be used to get the
// result of the RPC at some future time by invoking the Receive function on the
// returned instance.
//
// See GetCurrentNet for the blocking version and more details.
//
// NOTE: This is a pktd extension.
func (c *Client) GetCurrentNetAsync() FutureGetCurrentNetResult {
	cmd := btcjson.NewGetCurrentNetCmd()
	return c.sendCmd(cmd)
}

// GetCurrentNet returns the network the server is running on.
//
// NOTE: This is a pktd extension.
func (c *Client) GetCurrentNet() (wire.BitcoinNet, er.R) {
	return c.GetCurrentNetAsync().Receive()
}

// FutureGetHeadersResult is a future promise to deliver the result of a
// getheaders RPC invocation (or an applicable error).
//
// NOTE: This is a btcsuite extension ported from
// github.com/decred/dcrrpcclient.
type FutureGetHeadersResult chan *response

// Receive waits for the response promised by the future and returns the
// getheaders result.
//
// NOTE: This is a btcsuite extension ported from
// github.com/decred/dcrrpcclient.
func (r FutureGetHeadersResult) Receive() ([]wire.BlockHeader, er.R) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal result as a slice of strings.
	var result []string
	errr := json.Unmarshal(res, &result)
	if errr != nil {
		return nil, er.E(errr)
	}

	// Deserialize the []string into []wire.BlockHeader.
	headers := make([]wire.BlockHeader, len(result))
	for i, headerHex := range result {
		serialized, errr := hex.DecodeString(headerHex)
		if errr != nil {
			return nil, er.E(errr)
		}
		err = headers[i].Deserialize(bytes.NewReader(serialized))
		if err != nil {
			return nil, err
		}
	}
	return headers, nil
}

// GetHeadersAsync returns an instance of a type that can be used to get the result
// of the RPC at some future time by invoking the Receive function on the returned instance.
//
// See GetHeaders for the blocking version and more details.
//
// NOTE: This is a btcsuite extension ported from
// github.com/decred/dcrrpcclient.
func (c *Client) GetHeadersAsync(blockLocators []chainhash.Hash, hashStop *chainhash.Hash) FutureGetHeadersResult {
	locators := make([]string, len(blockLocators))
	for i := range blockLocators {
		locators[i] = blockLocators[i].String()
	}
	hash := ""
	if hashStop != nil {
		hash = hashStop.String()
	}
	cmd := btcjson.NewGetHeadersCmd(locators, hash)
	return c.sendCmd(cmd)
}

// GetHeaders mimics the wire protocol getheaders and headers messages by
// returning all headers on the main chain after the first known block in the
// locators, up until a block hash matches hashStop.
//
// NOTE: This is a btcsuite extension ported from
// github.com/decred/dcrrpcclient.
func (c *Client) GetHeaders(blockLocators []chainhash.Hash, hashStop *chainhash.Hash) ([]wire.BlockHeader, er.R) {
	return c.GetHeadersAsync(blockLocators, hashStop).Receive()
}

// FutureExportWatchingWalletResult is a future promise to deliver the result of
// an ExportWatchingWalletAsync RPC invocation (or an applicable error).
type FutureExportWatchingWalletResult chan *response

// Receive waits for the response promised by the future and returns the
// exported wallet.
func (r FutureExportWatchingWalletResult) Receive() ([]byte, []byte, er.R) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, nil, err
	}

	// Unmarshal result as a JSON object.
	var obj map[string]interface{}
	errr := json.Unmarshal(res, &obj)
	if errr != nil {
		return nil, nil, er.E(errr)
	}

	// Check for the wallet and tx string fields in the object.
	base64Wallet, ok := obj["wallet"].(string)
	if !ok {
		return nil, nil, er.Errorf("unexpected response type for "+
			"exportwatchingwallet 'wallet' field: %T\n",
			obj["wallet"])
	}
	base64TxStore, ok := obj["tx"].(string)
	if !ok {
		return nil, nil, er.Errorf("unexpected response type for "+
			"exportwatchingwallet 'tx' field: %T\n",
			obj["tx"])
	}

	walletBytes, errr := base64.StdEncoding.DecodeString(base64Wallet)
	if errr != nil {
		return nil, nil, er.E(errr)
	}

	txStoreBytes, errr := base64.StdEncoding.DecodeString(base64TxStore)
	if errr != nil {
		return nil, nil, er.E(errr)
	}

	return walletBytes, txStoreBytes, nil

}

// ExportWatchingWalletAsync returns an instance of a type that can be used to
// get the result of the RPC at some future time by invoking the Receive
// function on the returned instance.
//
// See ExportWatchingWallet for the blocking version and more details.
//
// NOTE: This is a pktwallet extension.
func (c *Client) ExportWatchingWalletAsync(account string) FutureExportWatchingWalletResult {
	cmd := btcjson.NewExportWatchingWalletCmd(&account, btcjson.Bool(true))
	return c.sendCmd(cmd)
}

// ExportWatchingWallet returns the raw bytes for a watching-only version of
// wallet.bin and tx.bin, respectively, for the specified account that can be
// used by pktwallet to enable a wallet which does not have the private keys
// necessary to spend funds.
//
// NOTE: This is a pktwallet extension.
func (c *Client) ExportWatchingWallet(account string) ([]byte, []byte, er.R) {
	return c.ExportWatchingWalletAsync(account).Receive()
}

// FutureSessionResult is a future promise to deliver the result of a
// SessionAsync RPC invocation (or an applicable error).
type FutureSessionResult chan *response

// Receive waits for the response promised by the future and returns the
// session result.
func (r FutureSessionResult) Receive() (*btcjson.SessionResult, er.R) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal result as a session result object.
	var session btcjson.SessionResult
	errr := json.Unmarshal(res, &session)
	if errr != nil {
		return nil, er.E(errr)
	}

	return &session, nil
}

// SessionAsync returns an instance of a type that can be used to get the result
// of the RPC at some future time by invoking the Receive function on the
// returned instance.
//
// See Session for the blocking version and more details.
//
// NOTE: This is a btcsuite extension.
func (c *Client) SessionAsync() FutureSessionResult {
	// Not supported in HTTP POST mode.
	if c.config.HTTPPostMode {
		return newFutureError(ErrWebsocketsRequired.Default())
	}

	cmd := btcjson.NewSessionCmd()
	return c.sendCmd(cmd)
}

// Session returns details regarding a websocket client's current connection.
//
// This RPC requires the client to be running in websocket mode.
//
// NOTE: This is a btcsuite extension.
func (c *Client) Session() (*btcjson.SessionResult, er.R) {
	return c.SessionAsync().Receive()
}

// FutureVersionResult is a future promise to deliver the result of a version
// RPC invocation (or an applicable error).
//
// NOTE: This is a btcsuite extension ported from
// github.com/decred/dcrrpcclient.
type FutureVersionResult chan *response

// Receive waits for the response promised by the future and returns the version
// result.
//
// NOTE: This is a btcsuite extension ported from
// github.com/decred/dcrrpcclient.
func (r FutureVersionResult) Receive() (map[string]btcjson.VersionResult,
	er.R) {
	res, err := receiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal result as a version result object.
	var vr map[string]btcjson.VersionResult
	errr := json.Unmarshal(res, &vr)
	if errr != nil {
		return nil, er.E(errr)
	}

	return vr, nil
}

// VersionAsync returns an instance of a type that can be used to get the result
// of the RPC at some future time by invoking the Receive function on the
// returned instance.
//
// See Version for the blocking version and more details.
//
// NOTE: This is a btcsuite extension ported from
// github.com/decred/dcrrpcclient.
func (c *Client) VersionAsync() FutureVersionResult {
	cmd := btcjson.NewVersionCmd()
	return c.sendCmd(cmd)
}

// Version returns information about the server's JSON-RPC API versions.
//
// NOTE: This is a btcsuite extension ported from
// github.com/decred/dcrrpcclient.
func (c *Client) Version() (map[string]btcjson.VersionResult, er.R) {
	return c.VersionAsync().Receive()
}
