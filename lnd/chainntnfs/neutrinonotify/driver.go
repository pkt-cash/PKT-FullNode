package neutrinonotify

import (
	"fmt"

	"github.com/pkt-cash/pktd/btcutil/er"
	"github.com/pkt-cash/pktd/lnd/chainntnfs"
	"github.com/pkt-cash/pktd/neutrino"
)

// createNewNotifier creates a new instance of the ChainNotifier interface
// implemented by NeutrinoNotifier.
func createNewNotifier(args ...interface{}) (chainntnfs.ChainNotifier, er.R) {
	if len(args) != 3 {
		return nil, er.Errorf("incorrect number of arguments to "+
			".New(...), expected 3, instead passed %v", len(args))
	}

	config, ok := args[0].(*neutrino.ChainService)
	if !ok {
		return nil, er.New("first argument to neutrinonotify.New " +
			"is incorrect, expected a *neutrino.ChainService")
	}

	spendHintCache, ok := args[1].(chainntnfs.SpendHintCache)
	if !ok {
		return nil, er.New("second argument to neutrinonotify.New " +
			"is  incorrect, expected a chainntfs.SpendHintCache")
	}

	confirmHintCache, ok := args[2].(chainntnfs.ConfirmHintCache)
	if !ok {
		return nil, er.New("third argument to neutrinonotify.New " +
			"is  incorrect, expected a chainntfs.ConfirmHintCache")
	}

	return New(config, spendHintCache, confirmHintCache), nil
}

// init registers a driver for the NeutrinoNotify concrete implementation of
// the chainntnfs.ChainNotifier interface.
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
