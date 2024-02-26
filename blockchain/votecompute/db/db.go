package db

import (
	"bytes"
	"encoding/binary"
	"time"

	"github.com/pkt-cash/PKT-FullNode/btcutil"
	"github.com/pkt-cash/PKT-FullNode/btcutil/er"
	"github.com/pkt-cash/PKT-FullNode/database"
	"github.com/pkt-cash/PKT-FullNode/pktlog/log"
)

// This bucket contains both votes and balances.
// Balance structure is:  [address][0,0,0,0] => [blockn][balance]
// Vote structure is:     [address][blockn]  => [is_candidate][vote]
const BucketName = "votebalance"

const balanceInfoLen = 4 + 8

type BalanceInfo struct {
	// The balance info is valid as of this height
	BlockNum int32
	// The Balance at this height
	Balance uint64
}

func isBalanceKey(key []byte) bool {
	return len(key) >= 4 && bytes.Equal(key[len(key)-4:], []byte{0, 0, 0, 0})
}

func encodeBalanceKey(addr []byte) []byte {
	out := make([]byte, 0, len(addr)+4)
	out = append(out, addr...)
	return append(out, 0, 0, 0, 0)
}

func decodeBalanceKey(key []byte) ([]byte, er.R) {
	if len(key) < 4 {
		return nil, er.New("decodeBalanceKey: RUNT")
	}
	if !isBalanceKey(key) {
		return nil, er.New("decodeBalanceKey: Entry is a vote, not a balance")
	}
	return key[:len(key)-4], nil
}

func encodeBalanceVal(bi []BalanceInfo) []byte {
	out := make([]byte, len(bi)*balanceInfoLen)
	for i, b := range bi {
		idx := i * balanceInfoLen
		binary.LittleEndian.PutUint32(out[idx:idx+4], uint32(b.BlockNum))
		binary.LittleEndian.PutUint64(out[idx+4:idx+4+8], b.Balance)
	}
	return out
}

func decodeBalanceVal(b []byte) ([]BalanceInfo, er.R) {
	out := make([]BalanceInfo, 0, len(b)/balanceInfoLen)
	for i := 0; i < len(b); i += balanceInfoLen {
		sl := b[i : i+balanceInfoLen]
		if len(sl) < balanceInfoLen {
			return nil, er.Errorf("Failed to parse balanceVal, record length is [%d]", len(b))
		}
		out = append(out, BalanceInfo{
			BlockNum: int32(binary.LittleEndian.Uint32(sl[:4])),
			Balance:  binary.LittleEndian.Uint64(sl[4:]),
		})
	}
	return out, nil
}

type voteKey struct {
	address []byte
	blockn  int32
}

type voteVal struct {
	isCandidate   bool
	voteForScript []byte
}

func encodeVoteKey(vk *voteKey) []byte {
	out := make([]byte, 0, len(vk.address)+4)
	out = append(out, vk.address...)
	out = append(out, 0, 0, 0, 0)
	binary.BigEndian.PutUint32(out[len(out)-4:], uint32(vk.blockn))
	return out
}

func decodeVoteKey(k []byte) (*voteKey, er.R) {
	if len(k) < 4 {
		return nil, er.New("decodeVoteKey: RUNT")
	}
	blockn := binary.BigEndian.Uint32(k[len(k)-4:])
	if blockn == 0 {
		return nil, er.New("decodeVoteKey: entry is a a balance, not a vote")
	}
	return &voteKey{
		address: k[:len(k)-4],
		blockn:  int32(blockn),
	}, nil
}

func encodeVoteVal(v *voteVal) []byte {
	out := make([]byte, 0, len(v.voteForScript)+1)
	isCandidate := byte(0)
	if v.isCandidate {
		isCandidate = 1
	}
	out = append(out, isCandidate)
	out = append(out, v.voteForScript...)
	return out
}

func decodeVoteVal(v []byte) (*voteVal, er.R) {
	if len(v) < 1 {
		return nil, er.New("decodeVoteVal: RUNT")
	}
	if v[0] != 0 && v[0] != 1 {
		return nil, er.New("decodeVoteVal: entry with invalid prefix")
	}
	return &voteVal{
		isCandidate:   v[0] == 1,
		voteForScript: v[1:],
	}, nil
}

// -----------------------------------------------------------------------------
// The addressbalances bucket stores current and recent snapshot of address balances.
// The keys in this bucket are the addressScript and the values are the serialized
// balanceInfo.
// -----------------------------------------------------------------------------
type AddressBalance struct {
	// The address whose balance we are considering (in pkScript format)
	AddressScript []byte
	// The balance info for this address, if the address does not exist on chain
	// or the address has zero balance and has been pruned since the last checkpoint,
	// then this field will be nil
	BalanceInfo []BalanceInfo
}

// FetchBalances gets the balance info for a list of addresses
// dbTx: A read or read/write db transactin
// addressScripts: A list of addresses in pkScript form
// returns: A list of addressBalance entries for each address
func FetchBalances(dbTx database.Tx, addressScripts [][]byte) ([]AddressBalance, er.R) {
	balancesBucket := dbTx.Metadata().Bucket([]byte(BucketName))
	out := make([]AddressBalance, 0, len(addressScripts))
	for _, addressScript := range addressScripts {
		balances := balancesBucket.Get(encodeBalanceKey(addressScript))
		if balances != nil {
			if bi, err := decodeBalanceVal(balances); err != nil {
				return nil, err
			} else {
				out = append(out, AddressBalance{addressScript, bi})
			}
		} else {
			out = append(out, AddressBalance{addressScript, nil})
		}
	}
	return out, nil
}

// PutBalances stores a list of address balances.
// dbTx: A read/write transaction
// balances: A list of addressBalance objects
func PutBalances(dbTx database.Tx, balances []AddressBalance) er.R {
	balancesBucket := dbTx.Metadata().Bucket([]byte(BucketName))
	for _, bal := range balances {
		if bal.BalanceInfo == nil {
			if err := balancesBucket.Delete(encodeBalanceKey(bal.AddressScript)); err != nil {
				return err
			}
		} else {
			balancesBucket.Put(
				encodeBalanceKey(bal.AddressScript),
				encodeBalanceVal(bal.BalanceInfo),
			)
		}
	}
	return nil
}

func DeleteVote(
	dbTx database.Tx,
	blockNum int32,
	addrScript []byte,
) er.R {
	balancesBucket := dbTx.Metadata().Bucket([]byte(BucketName))
	if balancesBucket == nil {
		return er.New("Unable to delete vote, bucket not created")
	}
	return balancesBucket.Delete(
		encodeVoteKey(&voteKey{address: addrScript, blockn: blockNum}),
	)
}

func PutVote(
	dbTx database.Tx,
	blockNum int32,
	addrScript []byte,
	isCandidate bool,
	voteForScript []byte,
) er.R {
	balancesBucket := dbTx.Metadata().Bucket([]byte(BucketName))
	if balancesBucket == nil {
		return er.New("Unable to store vote, bucket not created")
	}
	return balancesBucket.Put(
		encodeVoteKey(&voteKey{address: addrScript, blockn: blockNum}),
		encodeVoteVal(&voteVal{isCandidate: isCandidate, voteForScript: voteForScript}),
	)
}

func Init(dbTx database.Tx) er.R {
	buck := dbTx.Metadata().Bucket([]byte(BucketName))
	if buck == nil {
		log.Infof("Creating address balances and votes in database")
		if b, err := dbTx.Metadata().CreateBucket([]byte(BucketName)); err != nil {
			return err
		} else {
			buck = b
		}
	}
	t0 := time.Now()
	addrs := 0
	votes := 0
	maxBlock := int32(0)
	if err := buck.ForEach(func(k, v []byte) er.R {
		if isBalanceKey(k) {
			if _, err := decodeBalanceKey(k); err != nil {
				return err
			} else if bi, err := decodeBalanceVal(v); err != nil {
				return err
			} else {
				for _, b := range bi {
					if b.BlockNum > maxBlock {
						maxBlock = b.BlockNum
					}
				}
				addrs++
			}
		} else {
			if _, err := decodeVoteKey(k); err != nil {
				return err
			} else if _, err := decodeVoteVal(v); err != nil {
				return err
			} else {
				votes++
			}
		}
		return nil
	}); err != nil {
		return err
	}
	log.Infof("Scanned [%d] address balances and [%d] votes in [%d] milliseconds",
		addrs, votes, time.Since(t0).Milliseconds())
	return nil
}

type AddressInfo struct {
	AddressScript []byte
	Balance       btcutil.Amount
	IsCandidate   bool
	VoteFor       []byte
	BalanceBlock  int32
	VoteBlock     int32
	ExpiredCount  int32
	VoteCount     int32
	BalanceCount  int32
}

const EpochBlocks = 60 * 24 * 7

const VoteExpirationEpochs = 52

const VoteExpirationBlocks = VoteExpirationEpochs * EpochBlocks

func LastEpochEnd(currentBlockHeight int32) int32 {
	epochNum := EpochNum(currentBlockHeight)
	if epochNum == 0 {
		return 0
	} else {
		return epochLastBlock(epochNum - 1)
	}
}

func EpochNum(blockHeight int32) uint32 {
	return uint32(blockHeight / EpochBlocks)
}

func epochLastBlock(epochNum uint32) int32 {
	return int32((epochNum+1)*EpochBlocks - 1)
}

func isExpiredBalance(bi []BalanceInfo, currentBlockNum int32) bool {
	truncateLimit := currentBlockNum - 2*EpochBlocks
	if len(bi) >= 1 {
		last := bi[len(bi)-1]
		// Has zero balance and is older than the truncate limit
		return last.Balance == 0 && last.BlockNum < truncateLimit
	} else {
		log.Warnf("Balance entry with zero balances")
		return true
	}
}

// ListAddressInfo gets the balance and vote information for an address
// dbTx: a read database txn
// startFrom: The key to begin with, keys numerically less than this will be skipped
// currentBlock: The most recent block in the blockchain
// lastEpoch: If true, then we get the data for the finalization of the last epoch, otherwise we get current data.
// handler: This callback will be called with info for each address, it will drop if it returns an error.
func ListAddressInfo(
	dbTx database.Tx,
	startFrom []byte,
	effectiveBlock int32,
	handler func(*AddressInfo) er.R,
) er.R {
	//lastBlock := getLimitBlock(currentBlock, lastEpoch)
	buck := dbTx.Metadata().Bucket([]byte(BucketName))
	if buck == nil {
		return er.Errorf("Address balances not indexed")
	}
	c := buck.Cursor()
	if startFrom != nil {
		c.Seek(startFrom)
	} else {
		c.First()
	}

	currentBalance := AddressInfo{}
	for {
		if k := c.Key(); isBalanceKey(k) {
			if bi, err := decodeBalanceVal(c.Value()); err != nil {
				return err
			} else if bk, err := decodeBalanceKey(k); err != nil {
				return err
			} else {
				if len(currentBalance.AddressScript) > 0 {
					// send the previous AI to the handler
					if err := handler(&currentBalance); err != nil {
						return err
					}
				}
				currentBalance = AddressInfo{
					AddressScript: bk,
				}
				for _, b := range bi {
					if b.BlockNum <= effectiveBlock && b.BlockNum > currentBalance.BalanceBlock {
						currentBalance.BalanceBlock = int32(b.BlockNum)
						currentBalance.Balance = btcutil.Amount(b.Balance)
					}
				}
				if isExpiredBalance(bi, effectiveBlock) {
					currentBalance.ExpiredCount++
				}
			}
		} else {
			if vv, err := decodeVoteVal(c.Value()); err != nil {
				return err
			} else if vk, err := decodeVoteKey(k); err != nil {
				return err
			} else if vk.blockn < effectiveBlock-VoteExpirationBlocks {
				// The vote has passed into expiration, ignore it.
				if vk.blockn < effectiveBlock-VoteExpirationBlocks-(2*EpochBlocks) {
					// We don't consider it prunable until it goes 2 epoch blocks older than
					// the actual expiration - to guard against rollbacks.
					currentBalance.ExpiredCount++
				}
			} else {
				if len(currentBalance.AddressScript) > 0 && !bytes.Equal(currentBalance.AddressScript, vk.address) {
					if err := handler(&currentBalance); err != nil {
						return err
					}
					currentBalance = AddressInfo{}
				}
				currentBalance.AddressScript = vk.address
				if vk.blockn <= effectiveBlock && vk.blockn > currentBalance.VoteBlock {
					if vk.blockn < effectiveBlock-(2*EpochBlocks) {
						// If the NEW vote is old enough that it won't possibly be rolled back,
						// then the OLD vote can be pruned.
						currentBalance.ExpiredCount++
					}
					currentBalance.VoteBlock = int32(vk.blockn)
					currentBalance.IsCandidate = vv.isCandidate
					currentBalance.VoteFor = vv.voteForScript
				}
			}
		}
		if !c.Next() {
			break
		}
	}
	if len(currentBalance.AddressScript) > 0 {
		if err := handler(&currentBalance); err != nil {
			return err
		}
	}
	return nil
}

type PruneExpiredStats struct {
	BalancesVisited uint32
	BalancesDeleted uint32
	VotesVisited    uint32
	VotesDeleted    uint32
}

// Prune expired data from the db, this includes balances of addresses whose
// balance has been 0 for more than 1 epoch, and votes which have been replaced
// by another vote that is more than 1 epoch old, OR, are more than 52 epochs old
// i.e. more than a year old.
//
// dbTx: A writable transaction
// startFrom: Keys which are numerically greater than this will be skipped, this allows
//            you to continue where you left off.
// blockNum: The current block height at which we are pruning.
// deadline: A moment in time after which we should stop pruning. This can be used with
//           startFrom in order to avoid holding the database transaction for too long
//           and thus allow other processes to take place.
// stats: A pointer to a PruneExpiredStats which will accumulate info about how many
//        entries from the db were deleted.
// returns:
//   * If the process stopped because it reached the deadline, then the first return
//     will be the last key that was visited, this can be passed back to the startFrom
//     field. If the process has gone to completion then this will be nil.
//   * Error if any.
func PruneExpired(
	dbTx database.Tx,
	startFrom []byte,
	blockNum int32,
	deadline time.Time,
	stats *PruneExpiredStats,
) ([]byte, er.R) {
	// The prune limit is now - twice the epoch length
	truncateLimit := blockNum - (2 * EpochBlocks)
	// The vote expiration prune limit is now minus the vote expiration time (52 epochs)
	// minus twice the prune limit.
	voteExpireLimit := truncateLimit - VoteExpirationBlocks
	buck := dbTx.Metadata().Bucket([]byte(BucketName))
	if buck == nil {
		// No bucket, nothing to prune, don't bother returning an error
		return nil, nil
	}
	c := buck.Cursor()
	if startFrom != nil {
		c.Seek(startFrom)
	} else {
		c.Last()
	}

	var hasVote []byte
	for {
		if time.Now().After(deadline) {
			return c.Key(), nil
		}
		save := false
		k := c.Key()
		if k == nil {
			save = true
			// Do nothing and continue
		} else if isBalanceKey(k) {
			stats.BalancesVisited++
			if bi, err := decodeBalanceVal(c.Value()); err != nil {
				return nil, err
			} else if isExpiredBalance(bi, blockNum) {
				stats.BalancesDeleted++
			} else {
				save = true
			}
		} else {
			stats.VotesVisited++
			if true {
				// We are NOT pruning votes at this time, because to do so
				// imposes a rollback limit which does not otherwise exist.
				save = true
			} else if vk, err := decodeVoteKey(k); err != nil {
				return nil, err
			} else if bytes.Equal(vk.address, hasVote) {
				// Any vote older than the most recent one that is sufficiently
				// deep in the chain to not be reverted, is always deleted.
			} else if vk.blockn < voteExpireLimit {
				// The vote has expired by virtue of the 1 year expiration
			} else {
				save = true
				if vk.blockn < truncateLimit {
					// This vote is old enough it will never be rolled back
					// truncate all older votes by the same address
					hasVote = vk.address
				}
			}
			if !save {
				stats.VotesDeleted++
			}
		}
		if !save {
			if err := c.Delete(); err != nil {
				return nil, err
			}
		}
		if !c.Prev() {
			return nil, nil
		}
	}
}
