package votecompute

import (
	"encoding/base64"
	"encoding/binary"
	"hash"
	"time"

	electorium "github.com/cjdelisle/Electorium_go"
	"github.com/dchest/blake2b"
	"github.com/pkt-cash/PKT-FullNode/blockchain/votecompute/candidatetree"
	"github.com/pkt-cash/PKT-FullNode/blockchain/votecompute/db"
	"github.com/pkt-cash/PKT-FullNode/blockchain/votecompute/votewinnerdb"
	"github.com/pkt-cash/PKT-FullNode/btcutil"
	"github.com/pkt-cash/PKT-FullNode/btcutil/er"
	"github.com/pkt-cash/PKT-FullNode/database"
	"github.com/pkt-cash/PKT-FullNode/pktlog/log"
	"github.com/pkt-cash/PKT-FullNode/txscript"
)

// CONSENSUS: This is the number of blocks which pass - from the vote epoch
// before the "inaguration" occurs - i.e. the NS begins receiving payment.
// This leaves 240 blocks for the vote counting to take place.
const InaugurationOffset = 360

// CONSENSUS: Limit on number of addresses who can have other addresses vote for them.
const limitCandidates = 100000

// This is the number of blocks which pass - from the vote epoch before the counting
// begins.
const startCountOffset = 60

// Limit how much time is spent in each db transaction
const timeLimit = time.Millisecond * 50

// If we have more than this number of expired entries, prune the db
const expiredToPrune = 50000

// Max number that fits in a signed int32 2**31 - 1
const int32Max = int32(^uint32(0) >> 1)

var abortErr = er.GenericErrorType.Code("abortErr flags that execution should stop")

func (vc *VoteCompute) scanBalances(blockHeight int32, handler func(ai *db.AddressInfo) er.R) er.R {
	var startFrom []byte
	effectiveHeight := db.LastEpochEnd(blockHeight)
	for {
		deadline := time.Now().Add(timeLimit)
		deadlineReached := false
		if err := vc.db.View(func(tx database.Tx) er.R {
			return db.ListAddressInfo(tx, startFrom, effectiveHeight, func(ai *db.AddressInfo) er.R {
				if time.Now().After(deadline) {
					deadlineReached = true
					startFrom = ai.AddressScript
					return er.LoopBreak
				}
				if err := handler(ai); err != nil {
					return err
				}
				return nil
			})
		}); err != nil {
			if er.IsLoopBreak(err) && deadlineReached {
				if vc.activeComputeJob.Load() < 0 {
					// Abort has been requested (rollback)
					return abortErr.Default()
				}
				continue
			}
			return err
		} else {
			// Reached the end
			return nil
		}
	}
}

func (vc *VoteCompute) addressPrinter(script []byte) log.LogClosure {
	return log.C(func() string {
		if len(script) > 0 {
			return txscript.PkScriptToAddress(script, vc.cp).EncodeAddress()
		} else {
			return "<nobody>"
		}
	})
}

func writeNum(h hash.Hash, num uint64) {
	var b [8]byte
	binary.LittleEndian.PutUint64(b[:], num)
	h.Write(b[:])
}

func (vc *VoteCompute) compute(height int32) er.R {

	// 1. Load the winner for this height
	lastElectionHeight := int32(0)
	for {
		// log.Debugf("VoteCompute: Checking most recent winner for height [%d]", height)
		if err := vc.db.View(func(tx database.Tx) er.R {
			return votewinnerdb.ListWinnersBefore(tx, height, func(i int32, _, _ []byte) er.R {
				// log.Debugf("VoteCompute: Found last election at height [%d]", i)
				lastElectionHeight = i
				return er.LoopBreak
			})
		}); !er.IsLoopBreak(err) {
			if err != nil {
				return er.Errorf("Votedb corruption, unable to load votewinnerdb: %v", err)
			}
		}
		if height == 0 && lastElectionHeight == 0 {
			log.Debugf("VoteCompute: Height is zero and no election data, nothing to do")
			return nil
		}
		if lastElectionHeight >= height {
			// the win has been undermined by a rollback, destroy it
			log.Debugf("VoteCompute: Rolling back winner [%d] because there was a reorg", lastElectionHeight)
			if err := vc.db.Update(func(tx database.Tx) er.R {
				return votewinnerdb.RemoveWinner(tx, lastElectionHeight)
			}); err != nil {
				return er.Errorf("Votedb corruption, unable to load votewinnerdb: %v", err)
			}
		} else {
			break
		}
	}

	nextElectionHeight := lastElectionHeight + db.EpochBlocks
	if height > nextElectionHeight+db.EpochBlocks {
		panic("Block height exceeded 2 epochs without a vote computation")
	}
	if height < nextElectionHeight+startCountOffset {
		// We don't have enough data to run a vote computation
		return nil
	}

	// Launch the compute job, from now on, shouldAbort() will be false unless we need to stop
	vc.activeComputeJob.Store(nextElectionHeight)

	ct := candidatetree.NewCandidateTree(limitCandidates)

	// 2. Scan balances / votes to collect limitCandidates candidates
	log.Infof("VoteCompute: Starting vote computation for height [%d] (last vote at [%d])",
		nextElectionHeight, lastElectionHeight)
	addressCount := 0
	t0 := time.Now()
	hash := blake2b.New256()
	// Section 1: non-candidate non-voters by balance
	writeNum(hash, 1)
	i := 0
	if err := vc.scanBalances(nextElectionHeight, func(ai *db.AddressInfo) er.R {
		i++
		if i%100000 == 0 {
			log.Debugf("VoteCompute: Scanning address balances and votes [%d]", i)
		}
		if ai.Balance == 0 {
			// Zero balance addresses are excluded from computation. A non-voting
			// zero balance address will be completely pruned from the db and we do not
			// guarantee whether it has been pruned yet.
			return nil
		}
		addressCount++
		writeNum(hash, uint64(len(ai.AddressScript)))
		hash.Write(ai.AddressScript)
		writeNum(hash, uint64(ai.Balance))
		if ai.IsCandidate || len(ai.VoteFor) > 0 {
			ct.AddCandidate(&electorium.Vote{
				VoterId:          base64.StdEncoding.EncodeToString(ai.AddressScript),
				VoteFor:          base64.StdEncoding.EncodeToString(ai.VoteFor),
				NumberOfVotes:    uint64(ai.Balance),
				WillingCandidate: ai.IsCandidate,
			})
		}
		return nil
	}); err != nil {
		return err
	}

	// 3. If we can't store all candidates, run over again to assign additional votes
	if ct.OverLimit() {
		log.Infof("VoteCompute: ran over limit, those with less than [%f] coins cannot candidate",
			btcutil.Amount(int64(ct.GetWorst().NumberOfVotes)).ToBTC(),
		)
		byId := ct.NodesById()
		if err := vc.scanBalances(nextElectionHeight, func(ai *db.AddressInfo) er.R {
			if ai.Balance == 0 || len(ai.VoteFor) == 0 {
				// Zero balance and non-voting addresses are skipped to save computation.
				return nil
			}
			id := base64.StdEncoding.EncodeToString(ai.AddressScript)
			if _, ok := byId[id]; ok {
				// Nothing to do, we already got them
				return nil
			}
			vfId := base64.StdEncoding.EncodeToString(ai.VoteFor)
			if vf, ok := byId[vfId]; !ok {
				log.Tracef("Voter [%s] not counted because they voted for [%s] who "+
					"doesn't have enough coins to qualify",
					vc.addressPrinter(ai.AddressScript),
					vc.addressPrinter(ai.VoteFor))
			} else {
				log.Tracef("Voter [%s] doesn't have enough coins to qualify so "+
					"nobody can vote for them.",
					vc.addressPrinter(ai.AddressScript))
				vf.NumberOfVotes += uint64(ai.Balance)
			}
			return nil
		}); err != nil {
			return err
		}
	}

	// End section 1
	writeNum(hash, 0)
	// Section 2: Votes and balances
	writeNum(hash, 2)

	votes := *ct.Votes()
	for _, vote := range votes {
		writeNum(hash, uint64(len([]byte(vote.VoterId))))
		hash.Write([]byte(vote.VoterId))
		writeNum(hash, uint64(len([]byte(vote.VoteFor))))
		hash.Write([]byte(vote.VoteFor))
		writeNum(hash, vote.NumberOfVotes)
		candidate := uint64(0)
		if vote.WillingCandidate {
			candidate += 1
		}
		writeNum(hash, candidate)
	}
	voteHash := hash.Sum(nil)

	// 4. Run vote computation algorithm
	counter := electorium.MkVoteCounter(*ct.Votes(), false)
	winner := counter.FindWinner()
	var winScr []byte
	if winner != nil {
		id, err := base64.StdEncoding.DecodeString(winner.VoterId)
		if err != nil {
			return er.E(err)
		}
		winScr = id
	}

	// 5. Store the result
	if err := vc.db.Update(func(tx database.Tx) er.R {
		return votewinnerdb.PutWinner(tx, nextElectionHeight, winScr, voteHash)
	}); err != nil {
		return err
	}

	vc.activeComputeJob.Store(-1)
	nextStopAt := nextElectionHeight + db.EpochBlocks + InaugurationOffset
	vc.stopAtHeight.Store(nextStopAt)

	timeTaken := time.Since(t0)
	log.Infof("VoteCompute: Vote won by [%s] (found in [%s]) - [%d] balances",
		vc.addressPrinter(winScr), timeTaken, addressCount)

	var startFrom []byte
	var stats db.PruneExpiredStats
	t0 = time.Now()
	for {
		deadline := time.Now().Add(timeLimit)
		if err := vc.db.Update(func(tx database.Tx) er.R {
			if sf, err := db.PruneExpired(tx, startFrom, nextElectionHeight, deadline, &stats); err != nil {
				return err
			} else {
				startFrom = sf
				return nil
			}
		}); err != nil {
			return err
		}
		if len(startFrom) == 0 {
			break
		}
	}
	log.Infof("VoteCompute: Pruned [%d] of [%d] balances and [%d] of [%d] votes. Took [%s] time.",
		stats.BalancesDeleted, stats.BalancesVisited, stats.VotesDeleted, stats.VotesVisited, time.Since(t0))
	return nil
}

func voteComputeThread(vc *VoteCompute) {
	log.Infof("VoteCompute: Thread launched")
	height := int32(-100)
	for {
		height = vc.currentHeight.AwaitTrue(func(h int32) bool { return h != height })
		if err := vc.compute(height); err != nil {
			if abortErr.Is(err) {
				continue
			}
			log.Criticalf("Error in vote computation: %v", err)
		}
	}
}
