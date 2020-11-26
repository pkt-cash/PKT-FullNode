package localchans

import (
	"github.com/pkt-cash/pktd/pktlog"
)

// log is a logger that is initialized with no output filters.  This
// means the package will not perform any logging by default until the caller
// requests it.
var log pktlog.Logger

// UseLogger uses a specified Logger to output package logging info. This
// function is called from the parent package htlcswitch logger initialization.
func UseLogger(logger pktlog.Logger) {
	log = logger
}
