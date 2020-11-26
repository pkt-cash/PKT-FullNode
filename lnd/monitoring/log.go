package monitoring

import (
	"github.com/pkt-cash/pktd/pktlog"
	"github.com/pkt-cash/pktd/lnd/build"
)

// log is a logger that is initialized with no output filters.  This means the
// package will not perform any logging by default until the caller requests
// it.
var log pktlog.Logger

// The default amount of logging is none.
func init() {
	UseLogger(build.NewSubLogger("PROM", nil))
}

// DisableLog disables all library log output.  Logging output is disabled by
// default until UseLogger is called.
func DisableLog() {
	UseLogger(pktlog.Disabled)
}

// UseLogger uses a specified Logger to output package logging info.  This
// should be used in preference to SetLogWriter if the caller is also using
// pktlog.
func UseLogger(logger pktlog.Logger) {
	log = logger
}
