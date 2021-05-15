package headerfs

import (
	"bytes"
	"encoding/hex"
	"sync"

	"github.com/pkt-cash/pktd/btcutil/er"
	"github.com/pkt-cash/pktd/pktlog/log"

	"github.com/pkt-cash/pktd/blockchain"
	"github.com/pkt-cash/pktd/btcutil/gcs/builder"
	"github.com/pkt-cash/pktd/chaincfg"
	"github.com/pkt-cash/pktd/chaincfg/chainhash"
	"github.com/pkt-cash/pktd/chaincfg/genesis"
	"github.com/pkt-cash/pktd/pktwallet/waddrmgr"
	"github.com/pkt-cash/pktd/pktwallet/walletdb"
	"github.com/pkt-cash/pktd/wire"
)

// BlockHeaderStore is an interface that provides an abstraction for a generic
// store for block headers.
type BlockHeaderStore interface {
	// ChainTip returns the best known block header and height for the
	// BlockHeaderStore.
	ChainTip() (*wire.BlockHeader, uint32, er.R)
	ChainTip1(tx walletdb.ReadTx) (*wire.BlockHeader, uint32, er.R)

	// LatestBlockLocator returns the latest block locator object based on
	// the tip of the current main chain from the PoV of the
	// BlockHeaderStore.
	LatestBlockLocator() (blockchain.BlockLocator, er.R)

	// FetchHeaderByHeight attempts to retrieve a target block header based
	// on a block height.
	FetchHeaderByHeight(height uint32) (*wire.BlockHeader, er.R)
	FetchHeaderByHeight1(tx walletdb.ReadTx, height uint32) (*wire.BlockHeader, er.R)

	// FetchHeaderAncestors fetches the numHeaders block headers that are
	// the ancestors of the target stop hash. A total of numHeaders+1
	// headers will be returned, as we'll walk back numHeaders distance to
	// collect each header, then return the final header specified by the
	// stop hash. We'll also return the starting height of the header range
	// as well so callers can compute the height of each header without
	// knowing the height of the stop hash.
	FetchHeaderAncestors(uint32, *chainhash.Hash) ([]wire.BlockHeader,
		uint32, er.R)

	// HeightFromHash returns the height of a particular block header given
	// its hash.
	HeightFromHash(*chainhash.Hash) (uint32, er.R)

	// FetchHeader attempts to retrieve a block header determined by the
	// passed block height.
	FetchHeader(*chainhash.Hash) (*wire.BlockHeader, uint32, er.R)
	FetchHeader1(walletdb.ReadTx, *chainhash.Hash) (*wire.BlockHeader, uint32, er.R)

	// WriteHeaders adds a set of headers to the BlockHeaderStore in a
	// single atomic transaction.
	WriteHeaders(tx walletdb.ReadWriteTx, bh ...BlockHeader) er.R

	// RollbackLastBlock rolls back the BlockHeaderStore by a _single_
	// header. This method is meant to be used in the case of re-org which
	// disconnects the latest block header from the end of the main chain.
	// The information about the new header tip after truncation is
	// returned.
	RollbackLastBlock(tx walletdb.ReadWriteTx) (*waddrmgr.BlockStamp, er.R)
}

// headerBufPool is a pool of bytes.Buffer that will be re-used by the various
// headerStore implementations to batch their header writes to disk. By
// utilizing this variable we can minimize the total number of allocations when
// writing headers to disk.
var headerBufPool = sync.Pool{
	New: func() interface{} { return new(bytes.Buffer) },
}

// blockHeaderStore is an implementation of the BlockHeaderStore interface, a
// fully fledged database for Bitcoin block headers. The blockHeaderStore
// combines a flat file to store the block headers with a database instance for
// managing the index into the set of flat files.
type blockHeaderStore struct {
	Db walletdb.DB
	*headerIndex
}

// A compile-time check to ensure the blockHeaderStore adheres to the
// BlockHeaderStore interface.
var _ BlockHeaderStore = (*blockHeaderStore)(nil)

// NewBlockHeaderStore creates a new instance of the blockHeaderStore based on
// a target file path, an open database instance, and finally a set of
// parameters for the target chain. These parameters are required as if this is
// the initial start up of the blockHeaderStore, then the initial genesis
// header will need to be inserted.
func NewBlockHeaderStore(
	db walletdb.DB,
	netParams *chaincfg.Params,
) (BlockHeaderStore, er.R) {

	var bhs BlockHeaderStore
	retry := false
	err := walletdb.Update(db, func(tx walletdb.ReadWriteTx) er.R {
		hStore, err := newHeaderIndex(tx, "blocks")
		if err != nil {
			return err
		}
		bhs0 := &blockHeaderStore{headerIndex: hStore, Db: db}
		bhs = bhs0
		if _, err := bhs0.headerByHash(tx, netParams.GenesisHash); err != nil {
			genesisHeader := BlockHeader{
				BlockHeader: &genesis.Block(netParams.GenesisHash).Header,
				Height:      0,
			}
			gh := []headerEntry{genesisHeader.toIndexEntry()}
			if err := bhs0.addHeaders(tx, gh, true); err != nil {
				return err
			}
		}
		if err := bhs0.CheckConnectivity(tx); err != nil {
			log.Warnf("CheckConnectivity failed [%v] resyncing header chain", err)
			hStore.deleteBuckets(tx)
			retry = true
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if retry {
		return NewBlockHeaderStore(db, netParams)
	}
	return bhs, nil
}

// FetchHeader attempts to retrieve a block header determined by the passed
// block height.
//
// NOTE: Part of the BlockHeaderStore interface.
func (h *blockHeaderStore) FetchHeader(hash *chainhash.Hash) (*wire.BlockHeader, uint32, er.R) {
	var header *wire.BlockHeader
	var height uint32
	return header, height, walletdb.View(h.Db, func(tx walletdb.ReadTx) er.R {
		var err er.R
		header, height, err = h.FetchHeader1(tx, hash)
		return err
	})
}
func (h *blockHeaderStore) FetchHeader1(tx walletdb.ReadTx, hash *chainhash.Hash) (
	*wire.BlockHeader, uint32, er.R,
) {
	if he, err := h.headerByHash(tx, hash); err != nil {
		return nil, 0, err
	} else if hdr, err := blockHeaderFromHe(he); err != nil {
		return nil, 0, err
	} else {
		return hdr.BlockHeader, he.height, nil
	}
}

// FetchHeaderByHeight attempts to retrieve a target block header based on a
// block height.
//
// NOTE: Part of the BlockHeaderStore interface.
func (h *blockHeaderStore) FetchHeaderByHeight(height uint32) (*wire.BlockHeader, er.R) {
	var header *wire.BlockHeader
	return header, walletdb.View(h.Db, func(tx walletdb.ReadTx) er.R {
		var err er.R
		header, err = h.FetchHeaderByHeight1(tx, height)
		return err
	})
}

func (h *blockHeaderStore) FetchHeaderByHeight1(
	tx walletdb.ReadTx,
	height uint32,
) (*wire.BlockHeader, er.R) {
	if he, err := h.readHeader(tx, height); err != nil {
		return nil, err
	} else if hdr, err := blockHeaderFromHe(he); err != nil {
		return nil, err
	} else {
		return hdr.BlockHeader, nil
	}
}

// FetchHeaderAncestors fetches the numHeaders block headers that are the
// ancestors of the target stop hash. A total of numHeaders+1 headers will be
// returned, as we'll walk back numHeaders distance to collect each header,
// then return the final header specified by the stop hash. We'll also return
// the starting height of the header range as well so callers can compute the
// height of each header without knowing the height of the stop hash.
//
// NOTE: Part of the BlockHeaderStore interface.
func (h *blockHeaderStore) FetchHeaderAncestors(
	numHeaders uint32,
	stopHash *chainhash.Hash,
) ([]wire.BlockHeader, uint32, er.R) {
	var headers []wire.BlockHeader
	var startHeight uint32
	return headers, startHeight, walletdb.View(h.Db, func(tx walletdb.ReadTx) er.R {
		// First, we'll find the final header in the range, this will be the
		// ending height of our scan.
		endEntry, err := h.headerByHash(tx, stopHash)
		if err != nil {
			return err
		}
		startHeight = endEntry.height - numHeaders
		if headers, err = h.readHeaderRange(tx, startHeight, endEntry.height); err != nil {
			return err
		} else if realHash := headers[len(headers)-1].BlockHash(); realHash != endEntry.hash {
			return er.Errorf("Fetching %v headers up to %v - hash mismatch, got %v",
				numHeaders, stopHash, realHash)
		}
		return err
	})
}

// HeightFromHash returns the height of a particular block header given its
// hash.
//
// NOTE: Part of the BlockHeaderStore interface.
func (h *blockHeaderStore) HeightFromHash(hash *chainhash.Hash) (uint32, er.R) {
	var height uint32
	return height, walletdb.View(h.Db, func(tx walletdb.ReadTx) er.R {
		if he, err := h.headerByHash(tx, hash); err != nil {
			return err
		} else {
			height = he.height
			return nil
		}
	})
}

// RollbackLastBlock rollsback both the index, and on-disk header file by a
// _single_ header. This method is meant to be used in the case of re-org which
// disconnects the latest block header from the end of the main chain. The
// information about the new header tip after truncation is returned.
//
// NOTE: Part of the BlockHeaderStore interface.
func (h *blockHeaderStore) RollbackLastBlock(
	tx walletdb.ReadWriteTx,
) (*waddrmgr.BlockStamp, er.R) {
	if prev, err := h.truncateIndex(tx, true); err != nil {
		return nil, err
	} else {
		return &waddrmgr.BlockStamp{
			Height: int32(prev.height),
			Hash:   prev.hash,
		}, nil
	}
}

// BlockHeader is a Bitcoin block header that also has its height included.
type BlockHeader struct {
	*wire.BlockHeader

	// Height is the height of this block header within the current main
	// chain.
	Height uint32
}

// toIndexEntry converts the BlockHeader into a matching headerEntry. This
// method is used when a header is to be written to disk.
func (b *BlockHeader) toIndexEntry() headerEntry {
	var buf [80]byte
	hb := bytes.NewBuffer(buf[:])
	hb.Reset()

	// Finally, decode the raw bytes into a proper bitcoin header.
	if err := b.Serialize(hb); err != nil {
		panic(er.Errorf("Failed to serialize header %v", err))
	}
	return headerEntry{
		hash:   b.BlockHash(),
		height: b.Height,
		bytes:  hb.Bytes(),
	}
}

func blockHeaderFromHe(he *headerEntry) (*BlockHeader, er.R) {
	var ret wire.BlockHeader
	if err := ret.Deserialize(bytes.NewReader(he.bytes)); err != nil {
		return nil, err
	}
	return &BlockHeader{&ret, he.height}, nil
}

// WriteHeaders writes a set of headers to disk.
//
// NOTE: Part of the BlockHeaderStore interface.
func (h *blockHeaderStore) WriteHeaders(tx walletdb.ReadWriteTx, hdrs ...BlockHeader) er.R {
	headerLocs := make([]headerEntry, len(hdrs))
	for i, header := range hdrs {
		headerLocs[i] = header.toIndexEntry()
	}
	return h.addHeaders(tx, headerLocs, false)
}

// blockLocatorFromHash takes a given block hash and then creates a block
// locator using it as the root of the locator. We'll start by taking a single
// step backwards, then keep doubling the distance until genesis after we get
// 10 locators.
//
// TODO(roasbeef): make into single transaction.
func (h *blockHeaderStore) blockLocatorFromHash(tx walletdb.ReadTx, he *headerEntry) (
	blockchain.BlockLocator, er.R) {

	var locator blockchain.BlockLocator

	// Append the initial hash
	locator = append(locator, &he.hash)

	// If hash isn't found in DB or this is the genesis block, return the
	// locator as is
	height := he.height
	if height == 0 {
		return locator, nil
	}

	decrement := uint32(1)
	for height > 0 && len(locator) < wire.MaxBlockLocatorsPerMsg {
		// Decrement by 1 for the first 10 blocks, then double the jump
		// until we get to the genesis hash
		if len(locator) > 10 {
			decrement *= 2
		}

		if decrement > height {
			height = 0
		} else {
			height -= decrement
		}

		he, err := h.readHeader(tx, height)
		if err != nil {
			return locator, err
		}
		headerHash := he.hash

		locator = append(locator, &headerHash)
	}

	return locator, nil
}

// LatestBlockLocator returns the latest block locator object based on the tip
// of the current main chain from the PoV of the database and flat files.
//
// NOTE: Part of the BlockHeaderStore interface.
func (h *blockHeaderStore) LatestBlockLocator() (blockchain.BlockLocator, er.R) {
	var locator blockchain.BlockLocator
	return locator, walletdb.View(h.Db, func(tx walletdb.ReadTx) er.R {
		if ct, err := h.chainTip(tx); err != nil {
			return err
		} else {
			locator, err = h.blockLocatorFromHash(tx, ct)
			return err
		}
	})
}

// CheckConnectivity cycles through all of the block headers on disk, from last
// to first, and makes sure they all connect to each other. Additionally, at
// each block header, we also ensure that the index entry for that height and
// hash also match up properly.
func (h *blockHeaderStore) CheckConnectivity(tx walletdb.ReadTx) er.R {
	if he, err := h.chainTip(tx); err != nil {
		return err
	} else {
		for {
			if hdr, err := blockHeaderFromHe(he); err != nil {
				return err
			} else if bh := hdr.BlockHeader.BlockHash(); !he.hash.IsEqual(&bh) {
				return er.Errorf("hash mismatch at height %v, %v != %v",
					he.height, he.hash, bh)
			} else if he2, err := h.readHeader(tx, he.height); err != nil {
				return err
			} else if !he2.hash.IsEqual(&he.hash) {
				return er.Errorf("header with hash not equal to header at height "+
					" %v", he.height)
			} else if !bytes.Equal(he2.bytes, he.bytes) {
				return er.Errorf("header with bytes not equal to header at height "+
					" %v", he.height)
			} else if he.height > 0 {
				if he, err = h.headerByHash(tx, &hdr.BlockHeader.PrevBlock); err != nil {
					return err
				}
			} else {
				return nil
			}
		}
	}
}

// ChainTip returns the best known block header and height for the
// blockHeaderStore.
//
// NOTE: Part of the BlockHeaderStore interface.
func (h *blockHeaderStore) ChainTip() (*wire.BlockHeader, uint32, er.R) {
	var bh *wire.BlockHeader
	var height uint32
	return bh, height, walletdb.View(h.Db, func(tx walletdb.ReadTx) er.R {
		var err er.R
		bh, height, err = h.ChainTip1(tx)
		return err
	})
}

func (h *blockHeaderStore) ChainTip1(tx walletdb.ReadTx) (*wire.BlockHeader, uint32, er.R) {
	if ct, err := h.chainTip(tx); err != nil {
		return nil, 0, err
	} else if ch, err := blockHeaderFromHe(ct); err != nil {
		return nil, 0, err
	} else {
		return ch.BlockHeader, ct.height, nil
	}
}

// FilterHeaderStore is an implementation of a fully fledged database for any
// variant of filter headers.  The FilterHeaderStore combines a flat file to
// store the block headers with a database instance for managing the index into
// the set of flat files.
type FilterHeaderStore struct {
	Db walletdb.DB
	*headerIndex
}

// NewFilterHeaderStore returns a new instance of the FilterHeaderStore based
// on a target file path, filter type, and target net parameters. These
// parameters are required as if this is the initial start up of the
// FilterHeaderStore, then the initial genesis filter header will need to be
// inserted.
func NewFilterHeaderStore(
	db walletdb.DB,
	netParams *chaincfg.Params,
	headerStateAssertion *FilterHeader,
	bhs BlockHeaderStore,
) (*FilterHeaderStore, er.R) {

	var fhs *FilterHeaderStore
	resetState := false
	if err := walletdb.Update(db, func(tx walletdb.ReadWriteTx) er.R {
		fStore, err := newHeaderIndex(tx, "filters")
		if err != nil {
			return err
		}

		fhs = &FilterHeaderStore{
			db,
			fStore,
		}

		if _, err := fhs.headerByHash(tx, netParams.GenesisHash); err != nil {
			if basicFilter, err := builder.BuildBasicFilter(
				genesis.Block(netParams.GenesisHash), nil,
			); err != nil {
				return err
			} else if genesisFilterHash, err := builder.MakeHeaderForFilter(
				basicFilter,
				genesis.Block(netParams.GenesisHash).Header.PrevBlock,
			); err != nil {
				return err
			} else {
				fh := FilterHeader{
					HeaderHash: *netParams.GenesisHash,
					FilterHash: genesisFilterHash,
					Height:     0,
				}
				log.Debug("Inserting genesis block filter")
				return fhs.addHeaders(tx, []headerEntry{fh.toIndexEntry()}, true)
			}
		}

		// If we have a state assertion then we'll check it now to see if we
		// need to modify our filter header files before we proceed.
		if reset, err := fhs.maybeResetHeaderState(
			tx,
			headerStateAssertion,
			bhs,
		); err != nil {
			return err
		} else if reset {
			resetState = reset
			return nil
		}
		return nil
	}); err != nil {
		return nil, err
	}

	if resetState {
		log.Warn("Corrupt filter db, rebuilding")
		return NewFilterHeaderStore(db, netParams, nil, bhs)
	} else {
		return fhs, nil
	}
}

// maybeResetHeaderState will reset the header state if the header assertion
// fails, but only if the target height is found. The boolean returned indicates
// that header state was reset.
func (f *FilterHeaderStore) maybeResetHeaderState(
	tx walletdb.ReadWriteTx,
	headerStateAssertion *FilterHeader,
	bhs BlockHeaderStore,
) (bool, er.R) {

	failed := false

	if headerStateAssertion != nil {
		// First, we'll attempt to locate the header at this height. If no such
		// header is found, then we'll exit early.
		assertedHeader, err := f.FetchHeaderByHeight(headerStateAssertion.Height)
		if assertedHeader == nil {
			if !ErrHeaderNotFound.Is(err) {
				return false, err
			}
		} else if *assertedHeader != headerStateAssertion.FilterHash {
			log.Warnf("Filter header at height %v is not %v, assertion failed, resyncing filters",
				headerStateAssertion.Height, headerStateAssertion.HeaderHash)
			failed = true
		}
	}

	if !failed && bhs != nil {
		hdr, err := f.chainTip(tx)
		if err != nil {
			return false, err
		}
		for {
			if bh, err := bhs.FetchHeaderByHeight1(tx, hdr.height); err != nil {
				if ErrHashNotFound.Is(err) {
					log.Warnf("We have filter header number %v but no block header, "+
						"resetting filter headers", hdr.height)
					failed = true
					break
				}
				return false, err
			} else if bh := bh.BlockHash(); !hdr.hash.IsEqual(&bh) {
				log.Warnf("Filter header / block header mismatch at height %v: %v != %v",
					hdr.height, hdr.hash, bh)
				failed = true
				break
			} else if len(hdr.bytes) != 32 {
				log.Warnf("Filter header at height %v is not 32 bytes: %v",
					hdr.height, hex.EncodeToString(hdr.bytes))
				failed = true
				break
			} else if hdr.height == 0 {
				break
			}
			height := hdr.height - 1
			hdr, err = f.readHeader(tx, height)
			if err != nil {
				log.Warnf("Filter header missing at height %v (%v), resyncing filter headers",
					height, err)
				failed = true
				break
			}
		}
	}

	// If our on disk state and the provided header assertion don't match,
	// then we'll purge this state so we can sync it anew once we fully
	// start up.
	if failed {
		if err := f.deleteBuckets(tx); err != nil {
			return true, err
		} else {
			return true, f.createBuckets(tx)
		}
	}

	return false, nil
}

// FetchHeader returns the filter header that corresponds to the passed block
// height.
func (f *FilterHeaderStore) FetchHeader(hash *chainhash.Hash) (*chainhash.Hash, er.R) {
	var out *chainhash.Hash
	return out, walletdb.View(f.Db, func(tx walletdb.ReadTx) er.R {
		var err er.R
		out, err = f.FetchHeader1(tx, hash)
		return err
	})
}
func (f *FilterHeaderStore) FetchHeader1(tx walletdb.ReadTx, hash *chainhash.Hash) (*chainhash.Hash, er.R) {
	if hdr, err := f.headerByHash(tx, hash); err != nil {
		return nil, err
	} else if h, err := chainhash.NewHash(hdr.bytes); err != nil {
		return nil, err
	} else {
		return h, nil
	}
}

// FetchHeaderByHeight returns the filter header for a particular block height.
func (f *FilterHeaderStore) FetchHeaderByHeight(height uint32) (*chainhash.Hash, er.R) {
	var hash *chainhash.Hash
	return hash, walletdb.View(f.Db, func(tx walletdb.ReadTx) er.R {
		if hdr, err := f.readHeader(tx, height); err != nil {
			return err
		} else if h, err := chainhash.NewHash(hdr.bytes); err != nil {
			return err
		} else {
			hash = h
			return nil
		}
	})
}

// FetchHeaderAncestors fetches the numHeaders filter headers that are the
// ancestors of the target stop block hash. A total of numHeaders+1 headers will be
// returned, as we'll walk back numHeaders distance to collect each header,
// then return the final header specified by the stop hash. We'll also return
// the starting height of the header range as well so callers can compute the
// height of each header without knowing the height of the stop hash.
func (f *FilterHeaderStore) FetchHeaderAncestors(
	numHeaders uint32,
	stopHash *chainhash.Hash,
) ([]chainhash.Hash, uint32, er.R) {
	var hashes []chainhash.Hash
	var height uint32
	return hashes, height, walletdb.View(f.Db, func(tx walletdb.ReadTx) er.R {
		// First, we'll find the final header in the range, this will be the
		// ending height of our scan.
		endEntry, err := f.headerByHash(tx, stopHash)
		if err != nil {
			return err
		}
		startHeight := endEntry.height - numHeaders
		hashes, err = f.readHeaderRange(tx, startHeight, endEntry.height)
		if err != nil {
			return err
		}
		// for i, h := range hashes {
		// 	log.Debugf("Load filter header %d => [%s]", startHeight+uint32(i), h)
		// }
		if !bytes.Equal(hashes[len(hashes)-1][:], endEntry.bytes) {
			return er.Errorf("Hash mismatch on %v: %v %v", endEntry.height,
				hashes[len(hashes)-1], endEntry.bytes)
		}
		return nil
	})
}

// FilterHeader represents a filter header (basic or extended). The filter
// header itself is coupled with the block height and hash of the filter's
// block.
type FilterHeader struct {
	// HeaderHash is the hash of the block header that this filter header
	// corresponds to.
	HeaderHash chainhash.Hash

	// FilterHash is the filter header itself.
	FilterHash chainhash.Hash

	// Height is the block height of the filter header in the main chain.
	Height uint32
}

// toIndexEntry converts the filter header into a index entry to be stored
// within the database.
func (f *FilterHeader) toIndexEntry() headerEntry {
	return headerEntry{
		hash:   f.HeaderHash,
		height: f.Height,
		bytes:  f.FilterHash[:],
	}
}

// WriteHeaders writes a batch of filter headers to persistent storage. The
// headers themselves are appended to the flat file, and then the index updated
// to reflect the new entires.
func (f *FilterHeaderStore) WriteHeaders(tx walletdb.ReadWriteTx, hdrs ...FilterHeader) er.R {
	headerLocs := make([]headerEntry, len(hdrs))
	for i := range hdrs {
		headerLocs[i] = hdrs[i].toIndexEntry()
	}
	return f.addHeaders(tx, headerLocs, false)
}

// ChainTip returns the latest filter header and height known to the
// FilterHeaderStore.
func (f *FilterHeaderStore) ChainTip() (*chainhash.Hash, uint32, er.R) {
	var hash *chainhash.Hash
	var height uint32
	return hash, height, walletdb.View(f.Db, func(tx walletdb.ReadTx) er.R {
		var err er.R
		hash, height, err = f.ChainTip1(tx)
		return err
	})
}

func (f *FilterHeaderStore) ChainTip1(tx walletdb.ReadTx) (*chainhash.Hash, uint32, er.R) {
	if ct, err := f.chainTip(tx); err != nil {
		return nil, 0, err
	} else if ch, err := chainhash.NewHash(ct.bytes); err != nil {
		return nil, 0, err
	} else {
		return ch, ct.height, nil
	}
}

// RollbackLastBlock rollsback both the index, and on-disk header file by a
// _single_ filter header. This method is meant to be used in the case of
// re-org which disconnects the latest filter header from the end of the main
// chain. The information about the latest header tip after truncation is
// returned.
func (f *FilterHeaderStore) RollbackLastBlock(tx walletdb.ReadWriteTx) (*waddrmgr.BlockStamp, er.R) {
	if he, err := f.truncateIndex(tx, false); err != nil {
		return nil, err
	} else {
		// TODO(roasbeef): return chain hash also?
		return &waddrmgr.BlockStamp{
			Height: int32(he.height),
			Hash:   he.hash,
		}, nil
	}
}
