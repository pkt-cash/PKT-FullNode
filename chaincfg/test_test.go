package chaincfg

import (
	"os"
	"testing"

	"github.com/pkt-cash/pktd/chaincfg/globalcfg"
)

func TestMain(m *testing.M) {
	globalcfg.SelectConfig(globalcfg.BitcoinDefaults())
	os.Exit(m.Run())
}
