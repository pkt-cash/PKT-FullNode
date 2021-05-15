package channeldb

import (
	"io"

	"github.com/pkt-cash/pktd/btcutil/er"
)

// deserializeCloseChannelSummaryV6 reads the v6 database format for
// ChannelCloseSummary.
//
// NOTE: deprecated, only for migration.
func deserializeCloseChannelSummaryV6(r io.Reader) (*ChannelCloseSummary, er.R) {
	c := &ChannelCloseSummary{}

	err := ReadElements(r,
		&c.ChanPoint, &c.ShortChanID, &c.ChainHash, &c.ClosingTXID,
		&c.CloseHeight, &c.RemotePub, &c.Capacity, &c.SettledBalance,
		&c.TimeLockedBalance, &c.CloseType, &c.IsPending,
	)
	if err != nil {
		return nil, err
	}

	// We'll now check to see if the channel close summary was encoded with
	// any of the additional optional fields.
	err = ReadElements(r, &c.RemoteCurrentRevocation)
	switch {
	case er.EOF.Is(err):
		return c, nil

	// If we got a non-eof error, then we know there's an actually issue.
	// Otherwise, it may have been the case that this summary didn't have
	// the set of optional fields.
	case err != nil:
		return nil, err
	}

	if err := readChanConfig(r, &c.LocalChanConfig); err != nil {
		return nil, err
	}

	// Finally, we'll attempt to read the next unrevoked commitment point
	// for the remote party. If we closed the channel before receiving a
	// funding locked message, then this can be nil. As a result, we'll use
	// the same technique to read the field, only if there's still data
	// left in the buffer.
	err = ReadElements(r, &c.RemoteNextRevocation)
	if err != nil && !er.EOF.Is(err) {
		// If we got a non-eof error, then we know there's an actually
		// issue. Otherwise, it may have been the case that this
		// summary didn't have the set of optional fields.
		return nil, err
	}

	return c, nil
}
