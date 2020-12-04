package btcdnotify

import (
	"fmt"

	"github.com/pkt-cash/pktd/btcutil/er"
	"github.com/pkt-cash/pktd/chaincfg"
	"github.com/pkt-cash/pktd/lnd/chainntnfs"
	"github.com/pkt-cash/pktd/rpcclient"
)

// createNewNotifier creates a new instance of the ChainNotifier interface
// implemented by BtcdNotifier.
func createNewNotifier(args ...interface{}) (chainntnfs.ChainNotifier, er.R) {
	if len(args) != 4 {
		return nil, er.Errorf("incorrect number of arguments to "+
			".New(...), expected 4, instead passed %v", len(args))
	}

	config, ok := args[0].(*rpcclient.ConnConfig)
	if !ok {
		return nil, er.New("first argument to btcdnotify.New " +
			"is incorrect, expected a *rpcclient.ConnConfig")
	}

	chainParams, ok := args[1].(*chaincfg.Params)
	if !ok {
		return nil, er.New("second argument to btcdnotify.New " +
			"is incorrect, expected a *chaincfg.Params")
	}

	spendHintCache, ok := args[2].(chainntnfs.SpendHintCache)
	if !ok {
		return nil, er.New("third argument to btcdnotify.New " +
			"is incorrect, expected a chainntnfs.SpendHintCache")
	}

	confirmHintCache, ok := args[3].(chainntnfs.ConfirmHintCache)
	if !ok {
		return nil, er.New("fourth argument to btcdnotify.New " +
			"is incorrect, expected a chainntnfs.ConfirmHintCache")
	}

	return New(config, chainParams, spendHintCache, confirmHintCache)
}

// init registers a driver for the BtcdNotifier concrete implementation of the
// chainntnfs.ChainNotifier interface.
func init() {
	// Register the driver.
	notifier := &chainntnfs.NotifierDriver{
		NotifierType: notifierType,
		New:          createNewNotifier,
	}

	if err := chainntnfs.RegisterNotifier(notifier); err != nil {
		panic(fmt.Sprintf("failed to register notifier driver '%s': %v",
			notifierType, err))
	}
}
