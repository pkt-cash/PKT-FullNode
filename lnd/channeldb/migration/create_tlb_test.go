package migration_test

import (
	"testing"

	"github.com/pkt-cash/pktd/btcutil/er"
	"github.com/pkt-cash/pktd/lnd/channeldb/kvdb"
	"github.com/pkt-cash/pktd/lnd/channeldb/migration"
	"github.com/pkt-cash/pktd/lnd/channeldb/migtest"
)

// TestCreateTLB asserts that a CreateTLB properly initializes a new top-level
// bucket, and that it succeeds even if the bucket already exists. It would
// probably be better if the latter failed, but the kvdb abstraction doesn't
// support this.
func TestCreateTLB(t *testing.T) {
	newBucket := []byte("hello")

	tests := []struct {
		name            string
		beforeMigration func(kvdb.RwTx) er.R
		shouldFail      bool
	}{
		{
			name: "already exists",
			beforeMigration: func(tx kvdb.RwTx) er.R {
				_, err := tx.CreateTopLevelBucket(newBucket)
				return err
			},
			shouldFail: true,
		},
		{
			name:            "does not exist",
			beforeMigration: func(_ kvdb.RwTx) er.R { return nil },
			shouldFail:      false,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			migtest.ApplyMigration(
				t,
				test.beforeMigration,
				func(tx kvdb.RwTx) er.R {
					if tx.ReadBucket(newBucket) != nil {
						return nil
					}
					return er.Errorf("bucket \"%s\" not "+
						"created", newBucket)
				},
				migration.CreateTLB(newBucket),
				test.shouldFail,
			)
		})
	}
}
