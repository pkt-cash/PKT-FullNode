package balances

import (
	"github.com/pkt-cash/PKT-FullNode/blockchain"
	"github.com/pkt-cash/PKT-FullNode/btcutil"
	"github.com/pkt-cash/PKT-FullNode/btcutil/er"
	"github.com/pkt-cash/PKT-FullNode/btcutil/util/tmap"
	"github.com/pkt-cash/PKT-FullNode/database"
)

func ConnectBlock(
	tx database.Tx,
	block *btcutil.Block,
	spent []blockchain.SpentTxOut,
) er.R {
	bc := getBlockChanges(block, spent)
	if err := updateBalances(tx, block.Height(), bc); err != nil {
		return err
	}
	return nil
}

func DisconnectBlock(
	tx database.Tx,
	block *btcutil.Block,
	spent []blockchain.SpentTxOut,
) er.R {
	bc := getBlockChanges(block, spent)
	// Invert everything since we're removing the block
	tmap.ForEach(bc, func(k *balanceChange, v *struct{}) er.R {
		k.Diff = -k.Diff
		return nil
	})
	if err := updateBalances(tx, block.Height()-1, bc); err != nil {
		return err
	}
	return nil
}
