package wallet

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/pkt-cash/pktd/btcutil/hdkeychain"
	"github.com/pkt-cash/pktd/chaincfg"
)

// defaultDBTimeout specifies the timeout value when opening the wallet
// database.
var defaultDBTimeout = 10 * time.Second

// testWallet creates a test wallet and unlocks it.
func testWallet(t *testing.T) (*Wallet, func()) {
	// Set up a wallet.
	dir, errr := ioutil.TempDir("", "test_wallet")
	if errr != nil {
		t.Fatalf("Failed to create db dir: %v", errr)
	}

	cleanup := func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatalf("could not cleanup test: %v", err)
		}
	}

	seed, err := hdkeychain.GenerateSeed(hdkeychain.MinSeedBytes)
	if err != nil {
		t.Fatalf("unable to create seed: %v", err)
	}

	pubPass := []byte("hello")
	privPass := []byte("world")

	loader := NewLoader(
		&chaincfg.TestNet3Params, dir, "wallet.db", true, 250,
	)
	w, err := loader.CreateNewWallet(pubPass, privPass, seed, time.Now(), nil)
	if err != nil {
		t.Fatalf("unable to create wallet: %v", err)
	}
	chainClient := &mockChainClient{}
	w.chainClient = chainClient
	if err := w.Unlock(privPass, time.After(10*time.Minute)); err != nil {
		t.Fatalf("unable to unlock wallet: %v", err)
	}

	return w, cleanup
}
