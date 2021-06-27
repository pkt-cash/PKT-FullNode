package headerfs

import (
	"bytes"
	"encoding/binary"
	"sort"

	"github.com/pkt-cash/pktd/btcutil/er"
	"github.com/pkt-cash/pktd/pktlog/log"

	"github.com/pkt-cash/pktd/chaincfg/chainhash"
	"github.com/pkt-cash/pktd/pktwallet/walletdb"
)

var (
	// indexBucket is the main top-level bucket for the header index.
	// Nothing is stored in this bucket other than the sub-buckets which
	// contains the indexes for the various header types.
	oldIndexBucket = []byte("header-index")

	// headersBucket is the top level, under this is each header type
	// e.g. "block", "filter", etc
	// under those are the following
	// * "tip" -> tip hash
	// * "hdr" -> headers by hash
	// * "byheight" -> header hash by height (main chain)
	headersBucket = []byte("headers")

	tipKey         = []byte("tip")
	hdrBucket      = []byte("hdr")
	byheightBucket = []byte("byheight")
)

var Err er.ErrorType = er.NewErrorType("headerfs.Err")

var (
	// ErrHeightNotFound is returned when a specified height isn't found in
	// a target index.
	ErrHeightNotFound = Err.CodeWithDetail("ErrHeightNotFound",
		"target height not found in index")

	// ErrHashNotFound is returned when a specified block hash isn't found
	// in a target index.
	ErrHashNotFound = Err.CodeWithDetail("ErrHashNotFound",
		"target hash not found in index")
)

// HeaderType is an enum-like type which defines the various header types that
// are stored within the index.
type HeaderType uint8

const (
	// Block is the header type that represents regular Bitcoin block
	// headers.
	Block HeaderType = iota

	// RegularFilter is a header type that represents the basic filter
	// header type for the filter header chain.
	RegularFilter
)

const (
	// BlockHeaderSize is the size in bytes of the Block header type.
	BlockHeaderSize = 80

	// RegularFilterHeaderSize is the size in bytes of the RegularFilter
	// header type.
	RegularFilterHeaderSize = 32
)

// headerIndex is an index stored within the database that allows for random
// access into the on-disk header file. This, in conjunction with a flat file
// of headers consists of header database. The keys have been specifically
// crafted in order to ensure maximum write performance during IBD, and also to
// provide the necessary indexing properties required.
type headerIndex struct {
	indexType []byte
}

// newHeaderIndex creates a new headerIndex given an already open database, and
// a particular header type.
func newHeaderIndex(tx walletdb.ReadWriteTx, indexType string) (*headerIndex, er.R) {
	// Drop the old bucket if it happens to exist
	if err := tx.DeleteTopLevelBucket(oldIndexBucket); err != nil && !walletdb.ErrBucketNotFound.Is(err) {
		return nil, err
	}

	hi := &headerIndex{
		indexType: []byte(indexType),
	}

	if err := hi.createBuckets(tx); err != nil {
		return nil, err
	}

	return hi, nil
}

func (h *headerIndex) createBuckets(tx walletdb.ReadWriteTx) er.R {
	if bkt, err := h.rwBucket(tx); err != nil {
		return err
	} else if _, err := bkt.CreateBucketIfNotExists(hdrBucket); err != nil {
		return err
	} else if _, err := bkt.CreateBucketIfNotExists(byheightBucket); err != nil {
		return err
	} else {
		return nil
	}
}

func rootRwBucket(tx walletdb.ReadWriteTx) (walletdb.ReadWriteBucket, er.R) {
	root := tx.ReadWriteBucket(headersBucket)
	if root == nil {
		if r, err := tx.CreateTopLevelBucket(headersBucket); err != nil {
			return nil, err
		} else {
			root = r
		}
	}
	return root, nil
}

func (h *headerIndex) deleteBuckets(tx walletdb.ReadWriteTx) er.R {
	root, err := rootRwBucket(tx)
	if err != nil {
		return err
	}
	if err := root.DeleteNestedBucket(h.indexType); err != nil && !walletdb.ErrBucketNotFound.Is(err) {
		return err
	}
	return nil
}

func (h *headerIndex) rwBucket(tx walletdb.ReadWriteTx) (walletdb.ReadWriteBucket, er.R) {
	root, err := rootRwBucket(tx)
	if err != nil {
		return nil, err
	}
	sub := root.NestedReadWriteBucket(h.indexType)
	if sub == nil {
		if s, err := root.CreateBucket(h.indexType); err != nil {
			return nil, err
		} else {
			sub = s
		}
	}
	return sub, nil
}

func (h *headerIndex) roBucket(tx walletdb.ReadTx) (walletdb.ReadBucket, er.R) {
	root := tx.ReadBucket(headersBucket)
	if root == nil {
		return nil, walletdb.ErrBucketNotFound.Default()
	}
	sub := root.NestedReadBucket(h.indexType)
	if sub == nil {
		return nil, walletdb.ErrBucketNotFound.Default()
	}
	return sub, nil
}

// headerEntry is an internal type that's used to quickly map a (height, hash)
// pair into the proper key that'll be stored within the database.
type headerEntry struct {
	hash   chainhash.Hash
	height uint32
	bytes  []byte
}

func heightBin(height uint32) []byte {
	heightBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(heightBytes[:], height)
	return heightBytes
}

// headerBatch is a batch of header entries to be written to disk.
type headerBatch []headerEntry

// Len returns the number of routes in the collection.
//
// NOTE: This is part of the sort.Interface implementation.
func (h headerBatch) Len() int {
	return len(h)
}

// Sort by height
func (h headerBatch) Less(i, j int) bool {
	return h[i].height-h[j].height < 0
}

// Swap swaps the elements with indexes i and j.
//
// NOTE: This is part of the sort.Interface implementation.
func (h headerBatch) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

// addHeaders writes a batch of header entries in a single atomic batch
func (h *headerIndex) addHeaders(tx walletdb.ReadWriteTx, batch headerBatch, isGenesis bool) er.R {
	// If we're writing a 0-length batch, make no changes and return.
	if len(batch) == 0 {
		return nil
	}

	rootBucket, err := h.rwBucket(tx)
	if err != nil {
		return err
	}
	headerBucket := rootBucket.NestedReadWriteBucket(hdrBucket)
	byheight := rootBucket.NestedReadWriteBucket(byheightBucket)

	sort.Sort(batch)
	var tip *headerEntry
	if !isGenesis {
		tip, err = h.chainTip(tx)
		if err != nil {
			return err
		}
	} else {
		tip = &headerEntry{}
	}

	for _, header := range batch {
		if !isGenesis && header.height > tip.height+1 {
			log.Warnf("Unable to add header at height %v because tip is %v", header.height, tip.height)
			break
		}
		heightBytes := heightBin(header.height)
		headerBuf := bytes.NewBuffer(make([]byte, 0, len(heightBytes)+len(header.bytes)))
		headerBuf.Write(heightBytes[:])
		headerBuf.Write(header.bytes)
		content := headerBuf.Bytes()
		if err := headerBucket.Put(header.hash[:], content); err != nil {
			return err
		}
		if err := byheight.Put(heightBytes[:], header.hash[:]); err != nil {
			return err
		}

		tip.height = header.height
		tip.hash = header.hash
	}

	if tip == nil {
		return nil
	}
	return rootBucket.Put(tipKey, tip.hash[:])
}

func (h *headerIndex) headerByHash(tx walletdb.ReadTx, hash *chainhash.Hash) (*headerEntry, er.R) {
	rootBucket, err := h.roBucket(tx)
	if err != nil {
		return nil, err
	}
	headersBucket := rootBucket.NestedReadBucket(hdrBucket)
	if hdrBytes := headersBucket.Get(hash[:]); hdrBytes == nil {
		// If the hash wasn't found, then we don't know of this
		// hash within the index.
		return nil, ErrHashNotFound.New("", er.Errorf("With hash %v", hash))
	} else {
		return &headerEntry{
			hash:   *hash,
			height: binary.BigEndian.Uint32(hdrBytes[:4]),
			bytes:  hdrBytes[4:],
		}, nil
	}
}

func (h *headerIndex) readHeader(tx walletdb.ReadTx, height uint32) (*headerEntry, er.R) {
	rootBucket, err := h.roBucket(tx)
	if err != nil {
		return nil, err
	}
	byheight := rootBucket.NestedReadBucket(byheightBucket)
	hb := heightBin(height)
	if hash := byheight.Get(hb[:]); hash == nil {
		// If the hash wasn't found, then we don't know of this
		// hash within the index.
		return nil, ErrHashNotFound.New("", er.Errorf("height: %v", height))
	} else if ch, err := chainhash.NewHash(hash); err != nil {
		return nil, err
	} else if hbh, err := h.headerByHash(tx, ch); err != nil {
		return nil, err
	} else if hbh.height != height {
		return nil, er.Errorf("Db corruption, header %v at height %d is actually height %d",
			ch, height, hbh.height)
	} else {
		return hbh, nil
	}
}

// chainTip returns the best hash and height that the index knows of.
func (h *headerIndex) chainTip(tx walletdb.ReadTx) (*headerEntry, er.R) {
	rootBucket, err := h.roBucket(tx)
	if err != nil {
		return nil, err
	}
	if th, err := chainhash.NewHash(rootBucket.Get(tipKey)); err != nil {
		return nil, err
	} else {
		return h.headerByHash(tx, th)
	}
}

// truncateIndex truncates the index for a particluar header type by a single
// header entry. The passed newTip pointer should point to the hash of the new
// chain tip. Optionally, if the entry is to be deleted as well, deleteFlag
// should be set to true.
func (h *headerIndex) truncateIndex(
	tx walletdb.ReadWriteTx,
	deleteFlag bool,
) (*headerEntry, er.R) {
	rootBucket, err := h.rwBucket(tx)
	if err != nil {
		return nil, err
	}

	// If deleteFlag is set, then we'll also delete this entry
	// from the database as the primary index (block headers)
	// is being rolled back.
	if ct, err := h.chainTip(tx); err != nil {
		return nil, err
	} else if prev, err := h.readHeader(tx, ct.height-1); err != nil {
		return nil, err
	} else if err := rootBucket.Put(tipKey, prev.hash[:]); err != nil {
		return nil, err
	} else {
		if deleteFlag {
			hdrGroup := rootBucket.NestedReadWriteBucket(hdrBucket)
			byheight := rootBucket.NestedReadWriteBucket(byheightBucket)

			// Delete byheight but only if it's still the same
			hb := heightBin(ct.height)
			if !bytes.Equal(byheight.Get(hb[:]), ct.hash[:]) {
			} else if err := byheight.Delete(hb[:]); err != nil {
				return nil, err
			}

			if err := hdrGroup.Delete(ct.hash[:]); err != nil {
				return nil, err
			}
		}
		return prev, nil
	}
}
