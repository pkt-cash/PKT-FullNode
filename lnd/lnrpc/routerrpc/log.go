package routerrpc

import (
	"github.com/pkt-cash/pktd/pktlog"
	"github.com/pkt-cash/pktd/lnd/build"
)

// log is a logger that is initialized with no output filters.  This
// means the package will not perform any logging by default until the caller
// requests it.
var log pktlog.Logger

// Subsystem defines the logging code for this subsystem.
const Subsystem = "RRPC"

// The default amount of logging is none.
func init() {
	UseLogger(build.NewSubLogger(Subsystem, nil))
}

// DisableLog disables all library log output.  Logging output is disabled
// by default until UseLogger is called.
func DisableLog() {
	UseLogger(pktlog.Disabled)
}

// UseLogger uses a specified Logger to output package logging info.
// This should be used in preference to SetLogWriter if the caller is also
// using pktlog.
func UseLogger(logger pktlog.Logger) {
	log = logger
}

// logClosure is used to provide a closure over expensive logging operations so
// don't have to be performed when the logging level doesn't warrant it.
type logClosure func() string // nolint:unused

// String invokes the underlying function and returns the result.
func (c logClosure) String() string {
	return c()
}

// newLogClosure returns a new closure over a function that returns a string
// which itself provides a Stringer interface so that it can be used with the
// logging system.
func newLogClosure(c func() string) logClosure { // nolint:unused
	return logClosure(c)
}
