package headerfs

import (
	"bytes"
	"crypto/sha256"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/pkt-cash/pktd/btcutil/er"
	"github.com/pkt-cash/pktd/chaincfg"
	"github.com/pkt-cash/pktd/chaincfg/chainhash"
	"github.com/pkt-cash/pktd/chaincfg/genesis"
	"github.com/pkt-cash/pktd/pktwallet/walletdb"
	"github.com/pkt-cash/pktd/wire"
)

func createTestBlockHeaderStore() (func(), walletdb.DB, string,
	*blockHeaderStore, er.R) {
	tempDir, errr := ioutil.TempDir("", "store_test")
	if errr != nil {
		return nil, nil, "", nil, er.E(errr)
	}

	dbPath := filepath.Join(tempDir, "test.db")
	db, err := walletdb.Create("bdb", dbPath, true)
	if err != nil {
		return nil, nil, "", nil, err
	}

	hStore, err := NewBlockHeaderStore(db, &chaincfg.SimNetParams)
	if err != nil {
		return nil, nil, "", nil, err
	}

	cleanUp := func() {
		os.RemoveAll(tempDir)
		db.Close()
	}

	return cleanUp, db, tempDir, hStore.(*blockHeaderStore), nil
}

func createTestBlockHeaderChain(numHeaders uint32) []BlockHeader {
	blockHeaders := make([]BlockHeader, numHeaders)
	prevHeader := &genesis.Block(chaincfg.SimNetParams.GenesisHash).Header
	for i := uint32(1); i <= numHeaders; i++ {
		bitcoinHeader := &wire.BlockHeader{
			Bits:      uint32(rand.Int31()),
			Nonce:     uint32(rand.Int31()),
			Timestamp: prevHeader.Timestamp.Add(time.Minute * 1),
			PrevBlock: prevHeader.BlockHash(),
		}

		blockHeaders[i-1] = BlockHeader{
			BlockHeader: bitcoinHeader,
			Height:      i,
		}

		prevHeader = bitcoinHeader
	}

	return blockHeaders
}

func TestBlockHeaderStoreOperations(t *testing.T) {
	cleanUp, _, _, bhs, err := createTestBlockHeaderStore()
	if cleanUp != nil {
		defer cleanUp()
	}
	if err != nil {
		t.Fatalf("unable to create new block header store: %v", err)
	}

	rand.Seed(time.Now().Unix())

	// With our test instance created, we'll now generate a series of
	// "fake" block headers to insert into the database.
	const numHeaders = 100
	blockHeaders := createTestBlockHeaderChain(numHeaders)

	// With all the headers inserted, we'll now insert them into the
	// database in a single batch.
	walletdb.Update(bhs.Db, func(tx walletdb.ReadWriteTx) er.R {
		if err := bhs.WriteHeaders(tx, blockHeaders...); err != nil {
			t.Fatalf("unable to write block headers: %v", err)
			return err
		}

		// At this point, the _tip_ of the chain from the PoV of the database
		// should be the very last header we inserted.
		lastHeader := blockHeaders[len(blockHeaders)-1]
		tipHeader, tipHeight, err := bhs.ChainTip1(tx)
		if err != nil {
			t.Fatalf("unable to fetch chain tip")
		}
		if !reflect.DeepEqual(lastHeader.BlockHeader, tipHeader) {
			t.Fatalf("tip height headers don't match up: "+
				"expected %v, got %v", spew.Sdump(lastHeader),
				spew.Sdump(tipHeader))
		}
		if tipHeight != lastHeader.Height {
			t.Fatalf("chain tip doesn't match: expected %v, got %v",
				lastHeader.Height, tipHeight)
		}

		// Ensure that from the PoV of the database, the headers perfectly
		// connect.
		if err := bhs.CheckConnectivity(tx); err != nil {
			t.Fatalf("bhs detects that headers don't connect: %v", err)
		}

		// With all the headers written, we should be able to retrieve each
		// header according to its hash _and_ height.
		for _, header := range blockHeaders {
			dbHeader, err := bhs.FetchHeaderByHeight1(tx, header.Height)
			if err != nil {
				t.Fatalf("unable to fetch header by height: %v", err)
			}
			if !reflect.DeepEqual(*header.BlockHeader, *dbHeader) {
				t.Fatalf("retrieved by height headers don't match up: "+
					"expected %v, got %v", spew.Sdump(*header.BlockHeader),
					spew.Sdump(*dbHeader))
			}

			blockHash := header.BlockHash()
			dbHeader, _, err = bhs.FetchHeader1(tx, &blockHash)
			if err != nil {
				t.Fatalf("unable to fetch header by hash: %v", err)
			}
			if !reflect.DeepEqual(*dbHeader, *header.BlockHeader) {
				t.Fatalf("retrieved by hash headers don't match up: "+
					"expected %v, got %v", spew.Sdump(header),
					spew.Sdump(dbHeader))
			}
		}

		// Finally, we'll test the roll back scenario. Roll back the chain by a
		// single block, the returned block stamp should exactly match the last
		// header inserted, and the current chain tip should be the second to
		// last header inserted.
		secondToLastHeader := blockHeaders[len(blockHeaders)-2]
		blockStamp, err := bhs.RollbackLastBlock(tx)
		if err != nil {
			t.Fatalf("unable to rollback chain: %v", err)
		}
		if secondToLastHeader.Height != uint32(blockStamp.Height) {
			t.Fatalf("chain tip doesn't match: expected %v, got %v",
				secondToLastHeader.Height, blockStamp.Height)
		}
		headerHash := secondToLastHeader.BlockHash()
		if !bytes.Equal(headerHash[:], blockStamp.Hash[:]) {
			t.Fatalf("header hashes don't match: expected %v, got %v",
				headerHash, blockStamp.Hash)
		}
		tipHeader, tipHeight, err = bhs.ChainTip1(tx)
		if err != nil {
			t.Fatalf("unable to fetch chain tip")
		}
		if !reflect.DeepEqual(secondToLastHeader.BlockHeader, tipHeader) {
			t.Fatalf("tip height headers don't match up: "+
				"expected %v, got %v", spew.Sdump(secondToLastHeader),
				spew.Sdump(tipHeader))
		}
		if tipHeight != secondToLastHeader.Height {
			t.Fatalf("chain tip doesn't match: expected %v, got %v",
				secondToLastHeader.Height, tipHeight)
		}
		return nil
	})
}

func TestBlockHeaderStoreRecovery(t *testing.T) {
	// In this test we want to exercise the ability of the block header
	// store to recover in the face of a partial batch write (the headers
	// were written, but the index wasn't updated).
	cleanUp, db, _, bhs, err := createTestBlockHeaderStore()
	if cleanUp != nil {
		defer cleanUp()
	}
	if err != nil {
		t.Fatalf("unable to create new block header store: %v", err)
	}

	blockHeaders := createTestBlockHeaderChain(10)
	walletdb.Update(bhs.Db, func(tx walletdb.ReadWriteTx) er.R {
		// First we'll generate a test header chain of length 10, inserting it
		// into the header store.
		if err := bhs.WriteHeaders(tx, blockHeaders...); err != nil {
			t.Fatalf("unable to write block headers: %v", err)
		}

		// Next, in order to simulate a partial write, we'll roll back the
		// internal index by 5 blocks.
		for i := 0; i < 5; i++ {
			newTip := blockHeaders[len(blockHeaders)-i-1].PrevBlock
			if he, err := bhs.truncateIndex(tx, true); err != nil {
				t.Fatalf("unable to truncate index: %v", err)
			} else if !newTip.IsEqual(&he.hash) {
				t.Fatalf("hash mismatch index: %v", err)
			}
		}
		return nil
	})

	// Next, we'll re-create the block header store in order to trigger the
	// recovery logic.
	hs, err := NewBlockHeaderStore(db, &chaincfg.SimNetParams)
	if err != nil {
		t.Fatalf("unable to re-create bhs: %v", err)
	}
	bhs = hs.(*blockHeaderStore)

	// The chain tip of this new instance should be of height 5, and match
	// the 5th to last block header.
	tipHash, tipHeight, err := bhs.ChainTip()
	if err != nil {
		t.Fatalf("unable to get chain tip: %v", err)
	}
	if tipHeight != 5 {
		t.Fatalf("tip height mismatch: expected %v, got %v", 5, tipHeight)
	}
	prevHeaderHash := blockHeaders[5].BlockHash()
	tipBlockHash := tipHash.BlockHash()
	if bytes.Equal(prevHeaderHash[:], tipBlockHash[:]) {
		t.Fatalf("block hash mismatch: expected %v, got %v",
			prevHeaderHash, tipBlockHash)
	}
}

func createTestFilterHeaderStore() (func(), walletdb.DB, string, *FilterHeaderStore, er.R) {
	tempDir, errr := ioutil.TempDir("", "store_test")
	if errr != nil {
		return nil, nil, "", nil, er.E(errr)
	}

	dbPath := filepath.Join(tempDir, "test.db")
	db, err := walletdb.Create("bdb", dbPath, true)
	if err != nil {
		return nil, nil, "", nil, err
	}

	hStore, err := NewFilterHeaderStore(db, &chaincfg.SimNetParams, nil, nil)
	if err != nil {
		return nil, nil, "", nil, err
	}

	cleanUp := func() {
		os.RemoveAll(tempDir)
		db.Close()
	}

	return cleanUp, db, tempDir, hStore, nil
}

func createTestFilterHeaderChain(numHeaders uint32) []FilterHeader {
	filterHeaders := make([]FilterHeader, numHeaders)
	for i := uint32(1); i <= numHeaders; i++ {
		filterHeaders[i-1] = FilterHeader{
			HeaderHash: chainhash.DoubleHashH([]byte{byte(i)}),
			FilterHash: sha256.Sum256([]byte{byte(i)}),
			Height:     i,
		}
	}

	return filterHeaders
}

// TestBlockHeadersFetchHeaderAncestors tests that we're able to properly fetch
// the ancestors of a particular block, going from a set distance back to the
// target block.
func TestBlockHeadersFetchHeaderAncestors(t *testing.T) {

	cleanUp, _, _, bhs, err := createTestBlockHeaderStore()
	if cleanUp != nil {
		defer cleanUp()
	}
	if err != nil {
		t.Fatalf("unable to create new block header store: %v", err)
	}

	rand.Seed(time.Now().Unix())

	// With our test instance created, we'll now generate a series of
	// "fake" block headers to insert into the database.
	const numHeaders = 100
	blockHeaders := createTestBlockHeaderChain(numHeaders)

	walletdb.Update(bhs.Db, func(tx walletdb.ReadWriteTx) er.R {
		// With all the headers inserted, we'll now insert them into the
		// database in a single batch.
		if err := bhs.WriteHeaders(tx, blockHeaders...); err != nil {
			t.Fatalf("unable to write block headers: %v", err)
		}
		return nil
	})

	// Now that the headers have been written to disk, we'll attempt to
	// query for all the ancestors of the final header written, to query
	// the entire range.
	lastHeader := blockHeaders[numHeaders-1]
	lastHash := lastHeader.BlockHash()
	diskHeaders, startHeight, err := bhs.FetchHeaderAncestors(
		numHeaders-1, &lastHash,
	)
	if err != nil {
		t.Fatalf("unable to fetch headers: %v", err)
	}

	// Ensure that the first height of the block is height 1, and not the
	// genesis block.
	if startHeight != 1 {
		t.Fatalf("expected start height of %v got %v", 1, startHeight)
	}

	// Ensure that we retrieve the correct number of headers.
	if len(diskHeaders) != numHeaders {
		t.Fatalf("expected %v headers got %v headers",
			numHeaders, len(diskHeaders))
	}

	// We should get back the exact same set of headers that we inserted in
	// the first place.
	for i := 0; i < len(diskHeaders); i++ {
		diskHeader := diskHeaders[i]
		blockHeader := blockHeaders[i].BlockHeader
		if !reflect.DeepEqual(diskHeader, *blockHeader) {
			t.Fatalf("header mismatch, expected %v got %v",
				spew.Sdump(blockHeader), spew.Sdump(diskHeader))
		}
	}
}
