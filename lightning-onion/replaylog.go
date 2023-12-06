package sphinx

import (
	"crypto/sha256"

	"github.com/pkt-cash/PKT-FullNode/btcutil/er"
)

const (
	// HashPrefixSize is the size in bytes of the keys we will be storing
	// in the ReplayLog. It represents the first 20 bytes of a truncated
	// sha-256 hash of a secret generated by ECDH.
	HashPrefixSize = 20
)

// HashPrefix is a statically size, 20-byte array containing the prefix
// of a Hash256, and is used to detect duplicate sphinx packets.
type HashPrefix [HashPrefixSize]byte

// errReplayLogAlreadyStarted is an er.R returned when Start() is called on a
// ReplayLog after it is started and before it is stopped.
var errReplayLogAlreadyStarted *er.ErrorCode = Err.CodeWithDetail(
	"errReplayLogAlreadyStarted",
	"Replay log has already been started")

// errReplayLogNotStarted is an er.R returned when methods other than Start()
// are called on a ReplayLog before it is started or after it is stopped.
var errReplayLogNotStarted *er.ErrorCode = Err.CodeWithDetail(
	"errReplayLogNotStarted",
	"Replay log has not been started")

// hashSharedSecret Sha-256 hashes the shared secret and returns the first
// HashPrefixSize bytes of the hash.
func hashSharedSecret(sharedSecret *Hash256) *HashPrefix {
	// Sha256 hash of sharedSecret
	h := sha256.New()
	h.Write(sharedSecret[:])

	var sharedHash HashPrefix

	// Copy bytes to sharedHash
	copy(sharedHash[:], h.Sum(nil))
	return &sharedHash
}

// ReplayLog is an interface that defines a log of incoming sphinx packets,
// enabling strong replay protection. The interface is general to allow
// implementations near-complete autonomy. All methods must be safe for
// concurrent access.
type ReplayLog interface {
	// Start starts up the log. It returns an er.R if one occurs.
	Start() er.R

	// Stop safely stops the log. It returns an er.R if one occurs.
	Stop() er.R

	// Get retrieves an entry from the log given its hash prefix. It returns the
	// value stored and an er.R if one occurs. It returns ErrLogEntryNotFound
	// if the entry is not in the log.
	Get(*HashPrefix) (uint32, er.R)

	// Put stores an entry into the log given its hash prefix and an
	// accompanying purposefully general type. It returns ErrReplayedPacket if
	// the provided hash prefix already exists in the log.
	Put(*HashPrefix, uint32) er.R

	// Delete deletes an entry from the log given its hash prefix.
	Delete(*HashPrefix) er.R

	// PutBatch stores a batch of sphinx packets into the log given their hash
	// prefixes and accompanying values. Returns the set of entries in the batch
	// that are replays and an er.R if one occurs.
	PutBatch(*Batch) (*ReplaySet, er.R)
}

// MemoryReplayLog is a simple ReplayLog implementation that stores all added
// sphinx packets and processed batches in memory with no persistence.
//
// This is designed for use just in testing.
type MemoryReplayLog struct {
	batches map[string]*ReplaySet
	entries map[HashPrefix]uint32
}

// NewMemoryReplayLog constructs a new MemoryReplayLog.
func NewMemoryReplayLog() *MemoryReplayLog {
	return &MemoryReplayLog{}
}

// Start initializes the log and must be called before any other methods.
func (rl *MemoryReplayLog) Start() er.R {
	rl.batches = make(map[string]*ReplaySet)
	rl.entries = make(map[HashPrefix]uint32)
	return nil
}

// Stop wipes the state of the log.
func (rl *MemoryReplayLog) Stop() er.R {
	if rl.entries == nil || rl.batches == nil {
		return errReplayLogNotStarted.Default()
	}

	rl.batches = nil
	rl.entries = nil
	return nil
}

// Get retrieves an entry from the log given its hash prefix. It returns the
// value stored and an er.R if one occurs. It returns ErrLogEntryNotFound
// if the entry is not in the log.
func (rl *MemoryReplayLog) Get(hash *HashPrefix) (uint32, er.R) {
	if rl.entries == nil || rl.batches == nil {
		return 0, errReplayLogNotStarted.Default()
	}

	cltv, exists := rl.entries[*hash]
	if !exists {
		return 0, ErrLogEntryNotFound.Default()
	}

	return cltv, nil
}

// Put stores an entry into the log given its hash prefix and an accompanying
// purposefully general type. It returns ErrReplayedPacket if the provided hash
// prefix already exists in the log.
func (rl *MemoryReplayLog) Put(hash *HashPrefix, cltv uint32) er.R {
	if rl.entries == nil || rl.batches == nil {
		return errReplayLogNotStarted.Default()
	}

	_, exists := rl.entries[*hash]
	if exists {
		return ErrReplayedPacket.Default()
	}

	rl.entries[*hash] = cltv
	return nil
}

// Delete deletes an entry from the log given its hash prefix.
func (rl *MemoryReplayLog) Delete(hash *HashPrefix) er.R {
	if rl.entries == nil || rl.batches == nil {
		return errReplayLogNotStarted.Default()
	}

	delete(rl.entries, *hash)
	return nil
}

// PutBatch stores a batch of sphinx packets into the log given their hash
// prefixes and accompanying values. Returns the set of entries in the batch
// that are replays and an er.R if one occurs.
func (rl *MemoryReplayLog) PutBatch(batch *Batch) (*ReplaySet, er.R) {
	if rl.entries == nil || rl.batches == nil {
		return nil, errReplayLogNotStarted.Default()
	}

	// Return the result when the batch was first processed to provide
	// idempotence.
	replays, exists := rl.batches[string(batch.ID)]

	if !exists {
		replays = NewReplaySet()
		err := batch.ForEach(func(seqNum uint16, hashPrefix *HashPrefix, cltv uint32) er.R {
			err := rl.Put(hashPrefix, cltv)
			if ErrReplayedPacket.Is(err) {
				replays.Add(seqNum)
				return nil
			}

			// An er.R would be bad because we have already updated the entries
			// map, but no errors other than ErrReplayedPacket should occur.
			return err
		})
		if err != nil {
			return nil, err
		}

		replays.Merge(batch.ReplaySet)
		rl.batches[string(batch.ID)] = replays
	}

	batch.ReplaySet = replays
	batch.IsCommitted = true

	return replays, nil
}

// A compile time asserting *MemoryReplayLog implements the RelayLog interface.
var _ ReplayLog = (*MemoryReplayLog)(nil)
