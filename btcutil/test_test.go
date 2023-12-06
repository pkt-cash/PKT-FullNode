package btcutil_test

import (
	"os"
	"testing"

	"github.com/pkt-cash/PKT-FullNode/chaincfg/globalcfg"
)

func TestMain(m *testing.M) {
	globalcfg.SelectConfig(globalcfg.BitcoinDefaults())
	os.Exit(m.Run())
}
