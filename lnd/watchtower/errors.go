package watchtower

import "github.com/pkt-cash/pktd/btcutil/er"

var (
	Err = er.NewErrorType("watchtower")
	// ErrNoListeners signals that no listening ports were provided,
	// rendering the tower unable to receive client requests.
	ErrNoListeners = Err.CodeWithDetail("ErrNoListeners", "no listening ports were specified")

	// ErrNoNetwork signals that no tor.Net is provided in the Config, which
	// prevents resolution of listening addresses.
	ErrNoNetwork = Err.CodeWithDetail("ErrNoNetwork", "no network specified, must be tor or clearnet")
)
