// Copyright (c) 2023 Caleb James DeLisle
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package indexers

import (
	"github.com/pkt-cash/PKT-FullNode/blockchain"
	"github.com/pkt-cash/PKT-FullNode/blockchain/votecompute"
	"github.com/pkt-cash/PKT-FullNode/blockchain/votecompute/balances"
	votedb "github.com/pkt-cash/PKT-FullNode/blockchain/votecompute/db"
	"github.com/pkt-cash/PKT-FullNode/blockchain/votecompute/votes"
	"github.com/pkt-cash/PKT-FullNode/blockchain/votecompute/votewinnerdb"
	"github.com/pkt-cash/PKT-FullNode/btcutil"
	"github.com/pkt-cash/PKT-FullNode/btcutil/er"
	"github.com/pkt-cash/PKT-FullNode/chaincfg"
	"github.com/pkt-cash/PKT-FullNode/database"
)

type VotesIndex struct {
	db database.DB
	vc *votecompute.VoteCompute
}

var _ Indexer = (*VotesIndex)(nil)

func (vi *VotesIndex) Key() []byte {
	return []byte(votedb.BucketName)
}

const votesIndexName = "votes"

func (vi *VotesIndex) Name() string {
	return votesIndexName
}

func (vi *VotesIndex) Create(dbTx database.Tx) er.R {
	return nil
}

func (vi *VotesIndex) Init() er.R {
	if err := vi.db.Update(func(tx database.Tx) er.R {
		if err := votedb.Init(tx); err != nil {
			return err
		} else {
			return votewinnerdb.Init(tx)
		}
	}); err != nil {
		return err
	}
	return vi.vc.Init()
}

func (vi *VotesIndex) ConnectBlock(dbTx database.Tx, block *btcutil.Block, stxo []blockchain.SpentTxOut) er.R {
	if err := balances.ConnectBlock(dbTx, block, stxo); err != nil {
		return err
	} else if err := votes.ConnectBlock(dbTx, block, stxo); err != nil {
		return err
	} else if err := vi.vc.UpdateHeight(block.Height()); err != nil {
		return err
	} else {
		return nil
	}
}

func (vi *VotesIndex) DisconnectBlock(dbTx database.Tx, block *btcutil.Block, stxo []blockchain.SpentTxOut) er.R {
	if err := balances.DisconnectBlock(dbTx, block, stxo); err != nil {
		return err
	} else if err := votes.DisconnectBlock(dbTx, block, stxo); err != nil {
		return err
	} else if err := vi.vc.UpdateHeight(block.Height() - 1); err != nil {
		return err
	} else {
		return nil
	}
}

func NewVotes(db database.DB, params *chaincfg.Params) (*VotesIndex, er.R) {
	if vc, err := votecompute.NewVoteCompute(db, params); err != nil {
		return nil, err
	} else {
		return &VotesIndex{
			db: db,
			vc: vc,
		}, nil
	}
}

func DropVotes(db database.DB, interrupt <-chan struct{}) er.R {
	if err := dropIndex(db, []byte(votedb.BucketName), votesIndexName, interrupt); err != nil {
		return err
	}
	return db.Update(func(tx database.Tx) er.R {
		return votewinnerdb.Destroy(tx)
	})
}

func (vi *VotesIndex) NeedsInputs() bool {
	return true
}

func (vi *VotesIndex) VoteCompute() *votecompute.VoteCompute {
	return vi.vc
}
