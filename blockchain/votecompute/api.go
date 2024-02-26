package votecompute

import (
	"time"

	"github.com/pkt-cash/PKT-FullNode/blockchain/votecompute/db"
	"github.com/pkt-cash/PKT-FullNode/blockchain/votecompute/votewinnerdb"
	"github.com/pkt-cash/PKT-FullNode/btcutil/er"
	"github.com/pkt-cash/PKT-FullNode/btcutil/util/mailbox"
	"github.com/pkt-cash/PKT-FullNode/chaincfg"
	"github.com/pkt-cash/PKT-FullNode/database"
	"github.com/pkt-cash/PKT-FullNode/pktlog/log"
)

type VoteCompute struct {
	// Set from the outside
	currentHeight mailbox.Mailbox[int32]
	// Set from the vc thread
	stopAtHeight mailbox.Mailbox[int32] // NOTE: by default we need to set this to db.EpochBlocks + InaugurationOffset

	activeComputeJob mailbox.Mailbox[int32]

	db database.DB

	cp *chaincfg.Params
}

func NewVoteCompute(d database.DB, params *chaincfg.Params) (*VoteCompute, er.R) {
	return &VoteCompute{
		currentHeight:    mailbox.NewMailbox(int32(-1)),
		stopAtHeight:     mailbox.NewMailbox(int32(-1)),
		activeComputeJob: mailbox.NewMailbox(int32(-1)),
		db:               d,
		cp:               params,
	}, nil
}

func (vc *VoteCompute) Init() er.R {
	return vc.db.View(func(tx database.Tx) er.R {
		height := int32(0)
		if err := votewinnerdb.ListWinnersBefore(tx, int32Max, func(i int32, _, _ []byte) er.R {
			height = i
			return er.LoopBreak
		}); err != nil && !er.IsLoopBreak(err) {
			return er.Errorf("Votedb corruption, unable to load votewinnerdb: %v", err)
		}
		vc.currentHeight.Store(height)
		vc.stopAtHeight.Store(height + db.EpochBlocks + InaugurationOffset)
		go voteComputeThread(vc)
		return nil
	})
}

func (vc *VoteCompute) WaitUntilCanUpdateHeight(newHeight int32) {
	sah := vc.stopAtHeight.Load()
	for newHeight >= sah {
		log.Debugf("Pausing because new block height [%d] exceeds election result max height [%d]",
			newHeight, sah)
		t0 := time.Now()
		sah = vc.stopAtHeight.AwaitUpdate()
		log.Debugf("Pausing because new block height [%d] exceeds election result max height [%d] -> done in [%s]",
			newHeight, sah, time.Since(t0))
	}
}

func (vc *VoteCompute) UpdateHeight(newHeight int32) er.R {
	sah := vc.stopAtHeight.Load()
	if newHeight >= sah {
		return er.Errorf("Unable to insert block because vote computation is not complete")
	}
	acj := vc.activeComputeJob.Load()
	if acj >= newHeight {
		// In case of a rollback, kill the active job
		vc.activeComputeJob.Store(-1)
	}
	vc.currentHeight.Store(newHeight)
	return nil
}
