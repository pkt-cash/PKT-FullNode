// Copyright (c) 2013-2017 The btcsuite developers
// Copyright (c) 2015-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wallet

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/pkt-cash/pktd/blockchain"
	"github.com/pkt-cash/pktd/btcec"
	"github.com/pkt-cash/pktd/btcutil"
	"github.com/pkt-cash/pktd/btcutil/er"
	"github.com/pkt-cash/pktd/pktwallet/waddrmgr"
	"github.com/pkt-cash/pktd/pktwallet/wallet/txauthor"
	"github.com/pkt-cash/pktd/pktwallet/walletdb"
	"github.com/pkt-cash/pktd/pktwallet/wtxmgr"
	"github.com/pkt-cash/pktd/txscript"
	"github.com/pkt-cash/pktd/wire"
)

// Maximum number of inputs which will be included in a transaction
const MaxInputsPerTx = 4500

// byAmount defines the methods needed to satisify sort.Interface to
// sort credits by their output amount.
// type byAmount []wtxmgr.Credit
// func (s byAmount) Len() int           { return len(s) }
// func (s byAmount) Less(i, j int) bool { return s[i].Amount < s[j].Amount }
// func (s byAmount) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

type byAge []wtxmgr.Credit

func (s byAge) Len() int           { return len(s) }
func (s byAge) Less(i, j int) bool { return s[i].Height < s[j].Height }
func (s byAge) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func makeInputSource(eligible []wtxmgr.Credit) txauthor.InputSource {
	// Sort outputs by age, oldest first
	sort.Sort(byAge(eligible))

	// Current inputs and their total value.  These are closed over by the
	// returned input source and reused across multiple calls.
	currentTotal := btcutil.Amount(0)
	currentInputs := make([]*wire.TxIn, 0, len(eligible))
	currentScripts := make([][]byte, 0, len(eligible))
	currentInputValues := make([]btcutil.Amount, 0, len(eligible))

	return func(target btcutil.Amount) (btcutil.Amount, []*wire.TxIn,
		[]btcutil.Amount, [][]byte, er.R) {

		for currentTotal < target && len(eligible) != 0 {
			nextCredit := &eligible[0]
			eligible = eligible[1:]
			nextInput := wire.NewTxIn(&nextCredit.OutPoint, nil, nil)
			currentTotal += nextCredit.Amount
			currentInputs = append(currentInputs, nextInput)
			currentScripts = append(currentScripts, nextCredit.PkScript)
			currentInputValues = append(currentInputValues, nextCredit.Amount)
		}
		return currentTotal, currentInputs, currentInputValues, currentScripts, nil
	}
}

// secretSource is an implementation of txauthor.SecretSource for the wallet's
// address manager.
type secretSource struct {
	*waddrmgr.Manager
	addrmgrNs walletdb.ReadBucket
}

func (s secretSource) GetKey(addr btcutil.Address) (*btcec.PrivateKey, bool, er.R) {
	ma, err := s.Address(s.addrmgrNs, addr)
	if err != nil {
		return nil, false, err
	}

	mpka, ok := ma.(waddrmgr.ManagedPubKeyAddress)
	if !ok {
		e := er.Errorf("managed address type for %v is `%T` but "+
			"want waddrmgr.ManagedPubKeyAddress", addr, ma)
		return nil, false, e
	}
	privKey, err := mpka.PrivKey()
	if err != nil {
		return nil, false, err
	}
	return privKey, ma.Compressed(), nil
}

func (s secretSource) GetScript(addr btcutil.Address) ([]byte, er.R) {
	ma, err := s.Address(s.addrmgrNs, addr)
	if err != nil {
		return nil, err
	}

	msa, ok := ma.(waddrmgr.ManagedScriptAddress)
	if !ok {
		e := er.Errorf("managed address type for %v is `%T` but "+
			"want waddrmgr.ManagedScriptAddress", addr, ma)
		return nil, e
	}
	return msa.Script()
}

// txToOutputs creates a signed transaction which includes each output from
// outputs.  Previous outputs to reedeem are chosen from the passed account's
// UTXO set and minconf policy. An additional output may be added to return
// change to the wallet.  An appropriate fee is included based on the wallet's
// current relay fee.  The wallet must be unlocked to create the transaction.
//
// NOTE: The dryRun argument can be set true to create a tx that doesn't alter
// the database. A tx created with this set to true will intentionally have no
// input scripts added and SHOULD NOT be broadcasted.
func (w *Wallet) txToOutputs(txr CreateTxReq) (tx *txauthor.AuthoredTx, err er.R) {

	chainClient, err := w.requireChainClient()
	if err != nil {
		return nil, err
	}

	dbtx, err := w.db.BeginReadWriteTx()
	if err != nil {
		return nil, err
	}
	defer dbtx.Rollback()

	addrmgrNs := dbtx.ReadWriteBucket(waddrmgrNamespaceKey)

	// Get current block's height and hash.
	bs, err := chainClient.BlockStamp()
	if err != nil {
		return nil, err
	}

	var sweepOutput *wire.TxOut
	var needAmount btcutil.Amount
	for _, out := range txr.Outputs {
		needAmount += btcutil.Amount(out.Value)
		if out.Value == 0 {
			sweepOutput = out
		}
	}
	if sweepOutput != nil {
		needAmount = 0
	}
	eligible, err := w.findEligibleOutputs(
		dbtx, needAmount, txr.InputAddresses, txr.Minconf, bs, txr.InputMinHeight)
	if err != nil {
		return nil, err
	}

	addrStr := "<all>"
	if txr.InputAddresses != nil {
		addrs := make([]string, 0, len(*txr.InputAddresses))
		for _, a := range *txr.InputAddresses {
			addrs = append(addrs, fmt.Sprintf("%s (%s)",
				a.EncodeAddress(), hex.EncodeToString(a.ScriptAddress())))
		}
		addrStr = strings.Join(addrs, ", ")
	}
	log.Debugf("Found [%d] eligable inputs from addresses including [%s]",
		len(eligible), addrStr)

	inputSource := makeInputSource(eligible)
	changeSource := func() ([]byte, er.R) {
		// Derive the change output script.  As a hack to allow
		// spending from the imported account, change addresses are
		// created from account 0.
		var changeAddr btcutil.Address
		var err er.R
		if txr.ChangeAddress != nil {
			changeAddr = *txr.ChangeAddress
		} else {
			for _, c := range eligible {
				_, addrs, _, _ := txscript.ExtractPkScriptAddrs(c.PkScript, w.chainParams)
				if len(addrs) == 1 {
					changeAddr = addrs[0]
				}
			}
			if changeAddr == nil {
				err = er.New("Unable to find qualifying change address")
			}
		}
		if err != nil {
			return nil, err
		}
		return txscript.PayToAddrScript(changeAddr)
	}
	tx, err = txauthor.NewUnsignedTransaction(txr.Outputs, txr.FeeSatPerKB,
		inputSource, changeSource)
	if err != nil {
		return nil, err
	}

	// Randomize change position, if change exists, before signing.  This
	// doesn't affect the serialize size, so the change amount will still
	// be valid.
	if tx.ChangeIndex >= 0 {
		tx.RandomizeChangePosition()
	}

	// If a dry run was requested, we return now before adding the input
	// scripts, and don't commit the database transaction. The DB will be
	// rolled back when this method returns to ensure the dry run didn't
	// alter the DB in any way.
	if txr.DryRun {
		return tx, nil
	}

	err = tx.AddAllInputScripts(secretSource{w.Manager, addrmgrNs})
	if err != nil {
		return nil, err
	}

	err = validateMsgTx(tx.Tx, tx.PrevScripts, tx.PrevInputValues)
	if err != nil {
		return nil, err
	}

	if err := dbtx.Commit(); err != nil {
		return nil, err
	}

	// Finally, we'll request the backend to notify us of the transaction
	// that pays to the change address, if there is one, when it confirms.
	if tx.ChangeIndex >= 0 {
		changePkScript := tx.Tx.TxOut[tx.ChangeIndex].PkScript
		_, addrs, _, err := txscript.ExtractPkScriptAddrs(
			changePkScript, w.chainParams,
		)
		if err != nil {
			return nil, err
		}
		if err := chainClient.NotifyReceived(addrs); err != nil {
			return nil, err
		}
	}

	return tx, nil
}

func addrMatch(w *Wallet, script []byte, fromAddresses *[]btcutil.Address) bool {
	_, addrs, _, err := txscript.ExtractPkScriptAddrs(script, w.chainParams)
	if err != nil || len(addrs) != 1 {
		// Can't use this address, lets continue
		return false
	}
	for _, addr := range *fromAddresses {
		if bytes.Equal(addrs[0].ScriptAddress(), addr.ScriptAddress()) {
			return true
		}
	}
	return false
}

type amountCount struct {
	amount btcutil.Amount
	count  int
}

func (w *Wallet) findEligibleOutputs(
	dbtx walletdb.ReadTx,
	needAmount btcutil.Amount,
	fromAddresses *[]btcutil.Address,
	minconf int32,
	bs *waddrmgr.BlockStamp,
	inputMinHeight int,
) ([]wtxmgr.Credit, er.R) {
	txmgrNs := dbtx.ReadBucket(wtxmgrNamespaceKey)

	haveAmounts := make(map[string]*amountCount)
	var winningAddr []byte

	eligible := make([]wtxmgr.Credit, 0, 50)
	if err := w.TxStore.ForEachUnspentOutput(txmgrNs, func(output *wtxmgr.Credit) er.R {

		// Verify that the output is coming from one of the addresses which we accept to spend from
		// This is inherently expensive to filter at this level and ideally it would be moved into
		// the database by storing address->credit mappings directly, but after each transaction
		// is loaded, it's not much more effort to also extract the addresses each time.
		if fromAddresses != nil && !addrMatch(w, output.PkScript, fromAddresses) {
			return nil
		}

		// Only include this output if it meets the required number of
		// confirmations.  Coinbase transactions must have have reached
		// maturity before their outputs may be spent.
		if !confirmed(minconf, output.Height, bs.Height) {
			return nil
		}

		if output.Height < int32(inputMinHeight) {
			log.Debugf("Skipping output %s at height %d because it is below minimum %d\n",
				output.String(), output.Height, inputMinHeight)
			return nil
		}

		if output.FromCoinBase {
			target := int32(w.chainParams.CoinbaseMaturity)
			if !confirmed(target, output.Height, bs.Height) {
				return nil
			}
			if !w.chainParams.GlobalConf.HasNetworkSteward {
			} else if bs.Height-129600+1440 < output.Height {
			} else if int64(output.Amount) != blockchain.PktCalcNetworkStewardPayout(
				blockchain.CalcBlockSubsidy(output.Height, w.chainParams)) {
			} else {
				log.Debugf("Skipping burned output at height %d\n", output.Height)
				return nil
			}
		}

		// Locked unspent outputs are skipped.
		if w.LockedOutpoint(output.OutPoint) {
			return nil
		}

		eligible = append(eligible, *output)

		str := hex.EncodeToString(output.PkScript)
		ha := haveAmounts[str]
		if ha == nil {
			haa := amountCount{}
			ha = &haa
			haveAmounts[str] = ha
		}
		ha.amount += output.Amount
		ha.count += 1
		if (needAmount > 0 && ha.amount >= needAmount) || ha.count > MaxInputsPerTx {
			winningAddr = output.PkScript
			return er.LoopBreak
		}
		return nil
	}); err != nil && !er.IsLoopBreak(err) {
		return nil, err
	}

	// Lastly, we select one address from the elligable group (if possible) so as not
	// to mix addresses when it is not necessary to do so.
	if len(winningAddr) > 0 {
		n := 0
		for _, x := range eligible {
			if bytes.Equal(x.PkScript, winningAddr) {
				eligible[n] = x
				n++
			}
		}
		eligible = eligible[:n]
	} else if len(eligible) > MaxInputsPerTx {
		eligible = eligible[:MaxInputsPerTx]
	}

	return eligible, nil
}

// validateMsgTx verifies transaction input scripts for tx.  All previous output
// scripts from outputs redeemed by the transaction, in the same order they are
// spent, must be passed in the prevScripts slice.
func validateMsgTx(tx *wire.MsgTx, prevScripts [][]byte, inputValues []btcutil.Amount) er.R {
	hashCache := txscript.NewTxSigHashes(tx)
	for i, prevScript := range prevScripts {
		vm, err := txscript.NewEngine(prevScript, tx, i,
			txscript.StandardVerifyFlags, nil, hashCache, int64(inputValues[i]))
		if err != nil {
			return er.Errorf("cannot create script engine: %s", err)
		}
		err = vm.Execute()
		if err != nil {
			return er.Errorf("cannot validate transaction: %s", err)
		}
	}
	return nil
}
