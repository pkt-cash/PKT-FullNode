package main

import (
	"fmt"
	"math/big"
	"os"

	"github.com/pkt-cash/pktd/blockchain/packetcrypt/difficulty"
	"github.com/pkt-cash/pktd/pktconfig/version"
)

func usage() {
	fmt.Print("Usage: tar2work [-r] <target>")
}

func main() {
	version.SetUserAgentName("tar2work")
	if len(os.Args) < 2 {
		usage()
		return
	}
	arg := os.Args[1]
	reverse := false
	if arg == "-r" {
		if len(os.Args) < 3 {
			usage()
			return
		}
		arg = os.Args[1]
	}
	bigI := new(big.Int)
	bigI.SetString(arg, 16)
	var out *big.Int
	if reverse {
		out = difficulty.TargetForWork(bigI)
	} else {
		out = difficulty.WorkForTarget(bigI)
	}
	fmt.Printf("%s\n", out.Text(16))
}
