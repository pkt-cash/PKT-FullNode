package channeldb

import (
	"github.com/pkt-cash/pktd/btcutil/er"
	"github.com/pkt-cash/pktd/lnd/channeldb/kvdb"
)

var (
	// metaBucket stores all the meta information concerning the state of
	// the database.
	metaBucket = []byte("metadata")

	// dbVersionKey is a boltdb key and it's used for storing/retrieving
	// current database version.
	dbVersionKey = []byte("dbp")
)

// Meta structure holds the database meta information.
type Meta struct {
	// DbVersionNumber is the current schema version of the database.
	DbVersionNumber uint32
}

// FetchMeta fetches the meta data from boltdb and returns filled meta
// structure.
func (d *DB) FetchMeta(tx kvdb.RTx) (*Meta, er.R) {
	var meta *Meta

	err := kvdb.View(d, func(tx kvdb.RTx) er.R {
		return fetchMeta(meta, tx)
	}, func() {
		meta = &Meta{}
	})
	if err != nil {
		return nil, err
	}

	return meta, nil
}

// fetchMeta is an internal helper function used in order to allow callers to
// re-use a database transaction. See the publicly exported FetchMeta method
// for more information.
func fetchMeta(meta *Meta, tx kvdb.RTx) er.R {
	metaBucket := tx.ReadBucket(metaBucket)
	if metaBucket == nil {
		return ErrMetaNotFound.Default()
	}

	data := metaBucket.Get(dbVersionKey)
	if data == nil {
		meta.DbVersionNumber = getLatestDBVersion(dbVersions)
	} else {
		meta.DbVersionNumber = byteOrder.Uint32(data)
	}

	return nil
}

// PutMeta writes the passed instance of the database met-data struct to disk.
func (d *DB) PutMeta(meta *Meta) er.R {
	return kvdb.Update(d, func(tx kvdb.RwTx) er.R {
		return putMeta(meta, tx)
	}, func() {})
}

// putMeta is an internal helper function used in order to allow callers to
// re-use a database transaction. See the publicly exported PutMeta method for
// more information.
func putMeta(meta *Meta, tx kvdb.RwTx) er.R {
	metaBucket, err := tx.CreateTopLevelBucket(metaBucket)
	if err != nil {
		return err
	}

	return putDbVersion(metaBucket, meta)
}

func putDbVersion(metaBucket kvdb.RwBucket, meta *Meta) er.R {
	scratch := make([]byte, 4)
	byteOrder.PutUint32(scratch, meta.DbVersionNumber)
	return metaBucket.Put(dbVersionKey, scratch)
}
