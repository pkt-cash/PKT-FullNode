package lnwallet

import (
	"fmt"

	"github.com/pkt-cash/pktd/btcutil"
	"github.com/pkt-cash/pktd/btcutil/er"
	"github.com/pkt-cash/pktd/chaincfg/chainhash"
	"github.com/pkt-cash/pktd/lnd/lnwire"
)

// er.R wraps certain errors returned during channel reservation
// that can be sent across the wire to the remote peer. Errors not being
// ReservationErrors will not be sent to the remote in case of a failed channel
// reservation, as they may contain private information.
var ErrReservation = Err.Code("ErrReservation")

func mkReservationError(e er.R) er.R {
	return ErrReservation.New("", e)
}

// ErrZeroCapacity returns an error indicating the funder attempted to put zero
// funds into the channel.
func ErrZeroCapacity() er.R {
	return mkReservationError(er.New("zero channel funds"))
}

// ErrChainMismatch returns an error indicating that the initiator tried to
// open a channel for an unknown chain.
func ErrChainMismatch(knownChain,
	unknownChain *chainhash.Hash) er.R {
	return mkReservationError(
		er.Errorf("unknown chain=%v, supported chain=%v",
			unknownChain, knownChain),
	)
}

// ErrFunderBalanceDust returns an error indicating the initial balance of the
// funder is considered dust at the current commitment fee.
func ErrFunderBalanceDust(commitFee, funderBalance,
	minBalance int64) er.R {
	return mkReservationError(
		er.Errorf("funder balance too small (%v) with fee=%v sat, "+
			"minimum=%v sat required", funderBalance,
			commitFee, minBalance),
	)
}

// ErrCsvDelayTooLarge returns an error indicating that the CSV delay was to
// large to be accepted, along with the current max.
func ErrCsvDelayTooLarge(remoteDelay, maxDelay uint16) er.R {
	return mkReservationError(
		er.Errorf("CSV delay too large: %v, max is %v",
			remoteDelay, maxDelay),
	)
}

// ErrChanReserveTooSmall returns an error indicating that the channel reserve
// the remote is requiring is too small to be accepted.
func ErrChanReserveTooSmall(reserve, dustLimit btcutil.Amount) er.R {
	return mkReservationError(
		er.Errorf("channel reserve of %v sat is too small, min is %v "+
			"sat", int64(reserve), int64(dustLimit)),
	)
}

// ErrChanReserveTooLarge returns an error indicating that the chan reserve the
// remote is requiring, is too large to be accepted.
func ErrChanReserveTooLarge(reserve,
	maxReserve btcutil.Amount) er.R {
	return mkReservationError(
		er.Errorf("channel reserve is too large: %v sat, max "+
			"is %v sat", int64(reserve), int64(maxReserve)),
	)
}

// ErrNonZeroPushAmount is returned by a remote peer that receives a
// FundingOpen request for a channel with non-zero push amount while
// they have 'rejectpush' enabled.
func ErrNonZeroPushAmount() er.R {
	return mkReservationError(er.New("non-zero push amounts are disabled"))
}

// ErrMinHtlcTooLarge returns an error indicating that the MinHTLC value the
// remote required is too large to be accepted.
func ErrMinHtlcTooLarge(minHtlc,
	maxMinHtlc lnwire.MilliSatoshi) er.R {
	return mkReservationError(
		er.Errorf("minimum HTLC value is too large: %v, max is %v",
			minHtlc, maxMinHtlc),
	)
}

// ErrMaxHtlcNumTooLarge returns an error indicating that the 'max HTLCs in
// flight' value the remote required is too large to be accepted.
func ErrMaxHtlcNumTooLarge(maxHtlc, maxMaxHtlc uint16) er.R {
	return mkReservationError(
		er.Errorf("maxHtlcs is too large: %d, max is %d",
			maxHtlc, maxMaxHtlc),
	)
}

// ErrMaxHtlcNumTooSmall returns an error indicating that the 'max HTLCs in
// flight' value the remote required is too small to be accepted.
func ErrMaxHtlcNumTooSmall(maxHtlc, minMaxHtlc uint16) er.R {
	return mkReservationError(
		er.Errorf("maxHtlcs is too small: %d, min is %d",
			maxHtlc, minMaxHtlc),
	)
}

// ErrMaxValueInFlightTooSmall returns an error indicating that the 'max HTLC
// value in flight' the remote required is too small to be accepted.
func ErrMaxValueInFlightTooSmall(maxValInFlight,
	minMaxValInFlight lnwire.MilliSatoshi) er.R {
	return mkReservationError(
		er.Errorf("maxValueInFlight too small: %v, min is %v",
			maxValInFlight, minMaxValInFlight),
	)
}

// ErrNumConfsTooLarge returns an error indicating that the number of
// confirmations required for a channel is too large.
func ErrNumConfsTooLarge(numConfs, maxNumConfs uint32) er.R {
	return mkReservationError(
		er.Errorf("minimum depth of %d is too large, max is %d",
			numConfs, maxNumConfs),
	)
}

// ErrChanTooSmall returns an error indicating that an incoming channel request
// was too small. We'll reject any incoming channels if they're below our
// configured value for the min channel size we'll accept.
func ErrChanTooSmall(chanSize, minChanSize btcutil.Amount) er.R {
	return mkReservationError(
		er.Errorf("chan size of %v is below min chan size of %v",
			chanSize, minChanSize),
	)
}

// ErrChanTooLarge returns an error indicating that an incoming channel request
// was too large. We'll reject any incoming channels if they're above our
// configured value for the max channel size we'll accept.
func ErrChanTooLarge(chanSize, maxChanSize btcutil.Amount) er.R {
	return mkReservationError(
		er.Errorf("chan size of %v exceeds maximum chan size of %v",
			chanSize, maxChanSize),
	)
}

// ErrHtlcIndexAlreadyFailed is returned when the HTLC index has already been
// failed, but has not been committed by our commitment state.
type ErrHtlcIndexAlreadyFailed uint64

// Error returns a message indicating the index that had already been failed.
func (e ErrHtlcIndexAlreadyFailed) Error() string {
	return fmt.Sprintf("HTLC with ID %d has already been failed", e)
}

// ErrHtlcIndexAlreadySettled is returned when the HTLC index has already been
// settled, but has not been committed by our commitment state.
type ErrHtlcIndexAlreadySettled uint64

// Error returns a message indicating the index that had already been settled.
func (e ErrHtlcIndexAlreadySettled) Error() string {
	return fmt.Sprintf("HTLC with ID %d has already been settled", e)
}

// ErrInvalidSettlePreimage is returned when trying to settle an HTLC, but the
// preimage does not correspond to the payment hash.
type ErrInvalidSettlePreimage struct {
	preimage []byte
	rhash    []byte
}

// Error returns an error message with the offending preimage and intended
// payment hash.
func (e ErrInvalidSettlePreimage) Error() string {
	return fmt.Sprintf("Invalid payment preimage %x for hash %x",
		e.preimage, e.rhash)
}

// ErrUnknownHtlcIndex is returned when locally settling or failing an HTLC, but
// the HTLC index is not known to the channel. This typically indicates that the
// HTLC was already settled in a prior commitment.
type ErrUnknownHtlcIndex struct {
	chanID lnwire.ShortChannelID
	index  uint64
}

// Error returns an error logging the channel and HTLC index that was unknown.
func (e ErrUnknownHtlcIndex) Error() string {
	return fmt.Sprintf("No HTLC with ID %d in channel %v",
		e.index, e.chanID)
}
