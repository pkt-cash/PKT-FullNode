package votes

import (
	"github.com/pkt-cash/PKT-FullNode/blockchain"
	"github.com/pkt-cash/PKT-FullNode/blockchain/votecompute/db"
	"github.com/pkt-cash/PKT-FullNode/btcutil"
	"github.com/pkt-cash/PKT-FullNode/btcutil/er"
	"github.com/pkt-cash/PKT-FullNode/database"
	"github.com/pkt-cash/PKT-FullNode/pktlog/log"
)

func DisconnectBlock(
	tx database.Tx,
	block *btcutil.Block,
	spent []blockchain.SpentTxOut,
) er.R {
	votes, err := parseVotes(block, spent)
	if err != nil {
		return err
	}
	for _, v := range votes {
		if err := db.DeleteVote(tx, int32(v.VoteCastInBlock), v.VoterPkScript); err != nil {
			return err
		}
	}
	return nil
}

func ConnectBlock(dbTx database.Tx, block *btcutil.Block, stxo []blockchain.SpentTxOut) er.R {
	votes, err := parseVotes(block, stxo)
	if err != nil {
		log.Errorf("Unable to parse votes from block number [%d]: [%s]", block.Height(), err)
		return err
	}
	for _, v := range votes {
		if err := db.PutVote(
			dbTx,
			int32(v.VoteCastInBlock),
			v.VoterPkScript,
			v.VoterIsWillingCandidate,
			v.VoterPkScript,
		); err != nil {
			log.Errorf("Unable to store votes from block number [%d]: [%s]", block.Height(), err)
			return err
		}
	}
	return nil
}
