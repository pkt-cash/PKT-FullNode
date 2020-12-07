package wtclient

import "github.com/pkt-cash/pktd/btcutil/er"

var (
	Err = er.NewErrorType("wtclient")

	// ErrClientExiting signals that the watchtower client is shutting down.
	ErrClientExiting = Err.CodeWithDetail("ErrClientExiting", "watchtower client shutting down")

	// ErrTowerCandidatesExhausted signals that a TowerCandidateIterator has
	// cycled through all available candidates.
	ErrTowerCandidatesExhausted = Err.CodeWithDetail("ErrTowerCandidatesExhausted", "exhausted all tower "+
		"candidates")

	// ErrPermanentTowerFailure signals that the tower has reported that it
	// has permanently failed or the client believes this has happened based
	// on the tower's behavior.
	ErrPermanentTowerFailure = Err.CodeWithDetail("ErrPermanentTowerFailure", "permanent tower failure")

	// ErrNegotiatorExiting signals that the SessionNegotiator is shutting
	// down.
	ErrNegotiatorExiting = Err.CodeWithDetail("ErrNegotiatorExiting", "negotiator exiting")

	// ErrNoTowerAddrs signals that the client could not be created because
	// we have no addresses with which we can reach a tower.
	ErrNoTowerAddrs = Err.CodeWithDetail("ErrNoTowerAddrs", "no tower addresses")

	// ErrFailedNegotiation signals that the session negotiator could not
	// acquire a new session as requested.
	ErrFailedNegotiation = Err.CodeWithDetail("ErrFailedNegotiation", "session negotiation unsuccessful")

	// ErrUnregisteredChannel signals that the client was unable to backup a
	// revoked state because the channel had not been previously registered
	// with the client.
	ErrUnregisteredChannel = Err.CodeWithDetail("ErrUnregisteredChannel", "channel is not registered")
)
