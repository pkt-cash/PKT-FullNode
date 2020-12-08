package migration

import (
	"github.com/pkt-cash/pktd/btcutil/er"
	"github.com/pkt-cash/pktd/lnd/channeldb/kvdb"
	"github.com/pkt-cash/pktd/pktlog/log"
)

// CreateTLB creates a new top-level bucket with the passed bucket identifier.
func CreateTLB(bucket []byte) func(kvdb.RwTx) er.R {
	return func(tx kvdb.RwTx) er.R {
		log.Infof("Creating top-level bucket: \"%s\" ...", bucket)

		if tx.ReadBucket(bucket) != nil {
			return er.Errorf("top-level bucket \"%s\" "+
				"already exists", bucket)
		}

		_, err := tx.CreateTopLevelBucket(bucket)
		if err != nil {
			return err
		}

		log.Infof("Created top-level bucket: \"%s\"", bucket)
		return nil
	}
}
