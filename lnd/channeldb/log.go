package channeldb

import (
	"github.com/pkt-cash/pktd/pktlog"
	"github.com/pkt-cash/pktd/lnd/build"
	"github.com/pkt-cash/pktd/lnd/channeldb/kvdb"
	mig "github.com/pkt-cash/pktd/lnd/channeldb/migration"
	"github.com/pkt-cash/pktd/lnd/channeldb/migration12"
	"github.com/pkt-cash/pktd/lnd/channeldb/migration13"
	"github.com/pkt-cash/pktd/lnd/channeldb/migration16"
	"github.com/pkt-cash/pktd/lnd/channeldb/migration_01_to_11"
)

// log is a logger that is initialized with no output filters.  This
// means the package will not perform any logging by default until the caller
// requests it.
var log pktlog.Logger

func init() {
	UseLogger(build.NewSubLogger("CHDB", nil))
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
	mig.UseLogger(logger)
	migration_01_to_11.UseLogger(logger)
	migration12.UseLogger(logger)
	migration13.UseLogger(logger)
	migration16.UseLogger(logger)
	kvdb.UseLogger(logger)
}
