package channeldb

import (
	"testing"
	"time"

	"github.com/pkt-cash/pktd/btcutil/util"
	"github.com/pkt-cash/pktd/lnd/routing/route"
	"github.com/stretchr/testify/require"
)

// TestFlapCount tests lookup and writing of flap count to disk.
func TestFlapCount(t *testing.T) {
	db, cleanup, err := MakeTestDB()
	util.RequireNoErr(t, err)
	defer cleanup()

	// Try to read flap count for a peer that we have no records for.
	_, err = db.ReadFlapCount(testPub)
	require.True(t, ErrNoPeerBucket.Is(err))

	var (
		testPub2       = route.Vertex{2, 2, 2}
		peer1FlapCount = &FlapCount{
			Count:    20,
			LastFlap: time.Unix(100, 23),
		}
		peer2FlapCount = &FlapCount{
			Count:    39,
			LastFlap: time.Unix(200, 23),
		}
	)

	peers := map[route.Vertex]*FlapCount{
		testPub:  peer1FlapCount,
		testPub2: peer2FlapCount,
	}

	err = db.WriteFlapCounts(peers)
	util.RequireNoErr(t, err)

	// Lookup flap count for our first pubkey.
	count, err := db.ReadFlapCount(testPub)
	util.RequireNoErr(t, err)
	require.Equal(t, peer1FlapCount, count)

	// Lookup our flap count for the second peer.
	count, err = db.ReadFlapCount(testPub2)
	util.RequireNoErr(t, err)
	require.Equal(t, peer2FlapCount, count)
}
