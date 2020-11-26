package chainreg

import (
	"github.com/pkt-cash/pktd/pktlog"
	"github.com/pkt-cash/pktd/lnd/build"
)

// Subsystem defines the logging code for this subsystem.
const Subsystem = "CHRE"

// log is a logger that is initialized with the pktlog.Disabled logger.
var log pktlog.Logger

// The default amount of logging is none.
func init() {
	UseLogger(build.NewSubLogger(Subsystem, nil))
}

// DisableLog disables all logging output.
func DisableLog() {
	UseLogger(pktlog.Disabled)
}

// UseLogger uses a specified Logger to output package logging info.
func UseLogger(logger pktlog.Logger) {
	log = logger
}
