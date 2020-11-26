package kvdb

import "github.com/pkt-cash/pktd/pktlog"

// log is a logger that is initialized as disabled.  This means the package will
// not perform any logging by default until a logger is set.
var log = pktlog.Disabled

// UseLogger uses a specified Logger to output package logging info.
func UseLogger(logger pktlog.Logger) {
	log = logger
}
