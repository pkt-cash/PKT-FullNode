package balances

import (
	"bytes"

	"github.com/pkt-cash/PKT-FullNode/blockchain"
	"github.com/pkt-cash/PKT-FullNode/blockchain/votecompute/db"
	"github.com/pkt-cash/PKT-FullNode/btcutil"
	"github.com/pkt-cash/PKT-FullNode/btcutil/er"
	"github.com/pkt-cash/PKT-FullNode/btcutil/util/tmap"
	"github.com/pkt-cash/PKT-FullNode/chaincfg"
	"github.com/pkt-cash/PKT-FullNode/database"
	"github.com/pkt-cash/PKT-FullNode/pktlog/log"
	"github.com/pkt-cash/PKT-FullNode/txscript"
)

func addrC(pkScript []byte) log.LogClosure {
	return log.C(func() string {
		return txscript.PkScriptToAddress(pkScript, &chaincfg.PktMainNetParams).EncodeAddress()
	})
}

// applyBalanceChange updates the balance value to take into account the change specified
// in change for a change of balance which takes place in the block number blockNum.
func applyBalanceChange(
	balance *db.AddressBalance,
	change int64,
	blockNum int32,
) er.R {
	keep := make([]db.BalanceInfo, 0, len(balance.BalanceInfo))
	mostRecentBal := db.BalanceInfo{
		Balance:  0,
		BlockNum: 0,
	}
	for _, bal := range balance.BalanceInfo {
		if bal.BlockNum > mostRecentBal.BlockNum {
			// Track the most recent balance because this is the one we will base our new balance on
			mostRecentBal = bal
		}
		if bal.BlockNum >= blockNum {
			// Any balance entry which is *higher* than blockNum gets deleted (rollback)
		} else if db.EpochNum(int32(bal.BlockNum)) == db.EpochNum(blockNum) {
			// Same epoch, so we replace
		} else if db.EpochNum(blockNum)-db.EpochNum(bal.BlockNum) > 1 {
			// More than 1 epoch old, so we prune
		} else {
			keep = append(keep, bal)
		}
	}
	newEntry := db.BalanceInfo{
		Balance:  0,
		BlockNum: blockNum,
	}

	nb := int64(mostRecentBal.Balance) + change
	if nb < 0 {
		// TODO(cjd): Params should be a global so as not to pollute all of the code with passing them around.
		return er.Errorf("Impossible to apply balance change to [%s] at height [%d] "+
			"old balance is [%d @ %d] and change is [%d] which makes negative result [%d]",
			txscript.PkScriptToAddress(balance.AddressScript, &chaincfg.PktMainNetParams), blockNum,
			mostRecentBal.Balance, mostRecentBal.BlockNum, change, nb)
	}
	newEntry.Balance = uint64(nb)

	log.Tracef("Address [%s] changed by [%d] ([%d] -> [%d]) in block [%d]",
		addrC(balance.AddressScript), change, mostRecentBal.Balance,
		newEntry.Balance, blockNum)

	keep = append(keep, newEntry)
	balance.BalanceInfo = keep

	return nil
}

// A change of balance of an address
type balanceChange struct {
	// The address in pkScript format
	AddressScr []byte
	// The change of balance as a positive or negative number
	Diff int64
}

func newBalanceChanges() *tmap.Map[balanceChange, struct{}] {
	return tmap.New[balanceChange, struct{}](func(a, b *balanceChange) int {
		return bytes.Compare(a.AddressScr, b.AddressScr)
	})
}

func getBlockChanges(
	block *btcutil.Block,
	spent []blockchain.SpentTxOut,
) *tmap.Map[balanceChange, struct{}] {
	outCount := 0
	for _, tx := range block.Transactions() {
		outCount += len(tx.MsgTx().TxOut)
	}
	bcs := newBalanceChanges()
	quant := 0
	combined := 0
	insert := func(bc *balanceChange) {
		if old, _ := tmap.Insert(bcs, bc, &struct{}{}); old != nil {
			bc.Diff += old.Diff
			combined++
		}
		quant++
	}
	mints := 0
	for _, tx := range block.Transactions() {
		for _, out := range tx.MsgTx().TxOut {
			if out.Value > 0 {
				mints++
				log.Tracef("Address [%s] has acquired [%d] in block [%d]",
					addrC(out.PkScript), out.Value, block.Height())
				insert(&balanceChange{
					AddressScr: out.PkScript,
					Diff:       out.Value,
				})
			}
		}
	}
	spents := 0
	for _, sp := range spent {
		if sp.Amount > 0 {
			spents++
			log.Tracef("Address [%s] has spent [%d] in block [%d]", addrC(sp.PkScript), sp.Amount, block.Height())
			insert(&balanceChange{
				AddressScr: sp.PkScript,
				Diff:       -sp.Amount,
			})
		}
	}
	log.Tracef("In block [%d] there were [%d] balance changes ([%d] mints and [%d] spends), [%d] were combined",
		block.Height(), quant, mints, spents, combined)
	return bcs
}

func updateBalances(
	dbTx database.Tx,
	blockNum int32,
	changes *tmap.Map[balanceChange, struct{}],
) er.R {
	scr := make([][]byte, 0, tmap.Len(changes))
	amts := make([]int64, 0, tmap.Len(changes))
	tmap.ForEach(changes, func(c *balanceChange, _ *struct{}) er.R {
		scr = append(scr, c.AddressScr)
		amts = append(amts, c.Diff)
		return nil
	})
	log.Tracef("in block number [%d] there were [%d] = [%d] individual changes",
		blockNum, tmap.Len(changes), len(amts))
	balances, err := db.FetchBalances(dbTx, scr)
	if err != nil {
		return err
	}
	for i := 0; i < len(amts); i++ {
		if err := applyBalanceChange(&balances[i], amts[i], blockNum); err != nil {
			return err
		}
	}
	return db.PutBalances(dbTx, balances)
}
