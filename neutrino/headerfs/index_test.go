package headerfs

import (
	"bytes"
	"crypto/rand"
	"io/ioutil"
	"os"
	"testing"

	"github.com/pkt-cash/pktd/btcutil/er"

	"github.com/pkt-cash/pktd/pktwallet/walletdb"
	_ "github.com/pkt-cash/pktd/pktwallet/walletdb/bdb"
)

func createTestIndex(f func(tx walletdb.ReadWriteTx, hi *headerIndex) er.R) er.R {

	tempDir, errr := ioutil.TempDir("", "neutrino")
	if errr != nil {
		return er.E(errr)
	}

	db, err := walletdb.Create("bdb", tempDir+"/test.db", true)
	if err != nil {
		return err
	}

	defer func() {
		os.RemoveAll(tempDir)
		db.Close()
	}()

	return walletdb.Update(db, func(tx walletdb.ReadWriteTx) er.R {
		if filterDB, err := newHeaderIndex(tx, "block"); err != nil {
			return err
		} else {
			return f(tx, filterDB)
		}
	})
}

func TestAddHeadersIndexRetrieve(t *testing.T) {
	if err := createTestIndex(func(tx walletdb.ReadWriteTx, hIndex *headerIndex) er.R {

		// First, we'll create a a series of random headers that we'll use to
		// write into the database.
		const numHeaders = 100
		headerEntries := make(headerBatch, numHeaders)
		headerIndex := make(map[uint32]headerEntry)
		for i := uint32(0); i < numHeaders; i++ {
			var header headerEntry
			if _, err := rand.Read(header.hash[:]); err != nil {
				t.Fatalf("unable to read header: %v", err)
			}
			if _, err := rand.Read(header.bytes[:]); err != nil {
				t.Fatalf("unable to read header: %v", err)
			}
			header.height = i

			headerEntries[i] = header
			headerIndex[i] = header
		}

		// With the headers constructed, we'll write them to disk in a single
		// batch.
		if err := hIndex.addHeaders(tx, headerEntries, true); err != nil {
			t.Fatalf("unable to add headers: %v", err)
		}

		// Next, verify that the database tip matches the _final_ header
		// inserted.
		dbTip, err := hIndex.chainTip(tx)
		if err != nil {
			t.Fatalf("unable to obtain chain tip: %v", err)
		}
		lastEntry := headerIndex[numHeaders-1]
		if dbTip.height != lastEntry.height {
			t.Fatalf("height doesn't match: expected %v, got %v",
				lastEntry.height, dbTip.height)
		}
		if !bytes.Equal(dbTip.hash[:], lastEntry.hash[:]) {
			t.Fatalf("tip doesn't match: expected %x, got %x",
				lastEntry.hash[:], dbTip.hash[:])
		}

		// For each header written, check that we're able to retrieve the entry
		// both by hash and height.
		for i, headerEntry := range headerEntries {
			hdr, err := hIndex.headerByHash(tx, &headerEntry.hash)
			if err != nil {
				t.Fatalf("unable to retreive height(%v): %v", i, err)
			}
			if !hdr.hash.IsEqual(&headerEntry.hash) {
				t.Fatalf("height doesn't match: expected %v, got %v",
					hdr.hash, headerEntry.hash)
			}
		}

		// Next if we truncate the index by one, then we should end up at the
		// second to last entry for the tip.
		if _, err := hIndex.truncateIndex(tx, true); err != nil {
			t.Fatalf("unable to truncate index: %v", err)
		}

		// This time the database tip should be the _second_ to last entry
		// inserted.
		dbTip, err = hIndex.chainTip(tx)
		if err != nil {
			t.Fatalf("unable to obtain chain tip: %v", err)
		}
		lastEntry = headerIndex[numHeaders-2]
		if dbTip.height != lastEntry.height {
			t.Fatalf("height doesn't match: expected %v, got %v",
				lastEntry.height, dbTip.height)
		}
		if !bytes.Equal(dbTip.hash[:], lastEntry.hash[:]) {
			t.Fatalf("tip doesn't match: expected %x, got %x",
				lastEntry.hash[:], dbTip.hash[:])
		}
		return nil
	}); err != nil {
		t.Fatalf("unable to create test db: %v", err)
	}
}
