package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/pkt-cash/pktd/btcec"
	"github.com/pkt-cash/pktd/btcutil"
	"github.com/pkt-cash/pktd/chaincfg"
	"github.com/pkt-cash/pktd/chaincfg/chainhash"
	"github.com/pkt-cash/pktd/pktconfig/version"
	"github.com/pkt-cash/pktd/wire"
)

func usage() {
	fmt.Print("Usage: checksig <address> <signature> <message>\n")
}

func main() {
	version.SetUserAgentName("checksig")
	addrStr := ""
	sigStr := ""
	msg := ""
	if len(os.Args) == 2 && os.Args[1] == "-" {
		if _, err := fmt.Scanf("%s %s %s", &addrStr, &sigStr, &msg); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading in arguments from stdin")
			os.Exit(100)
		}
	} else if len(os.Args) != 4 {
		usage()
		os.Exit(100)
	} else {
		addrStr = os.Args[1]
		sigStr = os.Args[2]
		msg = os.Args[3]
	}

	// Decode the provided address.
	params := &chaincfg.PktMainNetParams
	addr, err := btcutil.DecodeAddress(addrStr, params)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Invalid address")
		os.Exit(100)
	}

	// Only P2PKH addresses are valid for signing.
	if _, ok := addr.(*btcutil.AddressPubKeyHash); !ok {
		fmt.Fprintln(os.Stderr, "Address is not a pay-to-pubkey-hash address")
		os.Exit(100)
	}

	sig, errr := base64.StdEncoding.DecodeString(sigStr)
	if errr != nil {
		fmt.Fprintln(os.Stderr, "Malformed base64 encoding")
		os.Exit(100)
	}

	var buf bytes.Buffer
	wire.WriteVarString(&buf, 0, "Bitcoin Signed Message:\n")
	wire.WriteVarString(&buf, 0, msg)
	expectedMessageHash := chainhash.DoubleHashB(buf.Bytes())
	pk, wasCompressed, err := btcec.RecoverCompact(btcec.S256(), sig, expectedMessageHash)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error in recoverCompact")
		os.Exit(100)
	}

	// Reconstruct the pubkey hash.
	var serializedPK []byte
	if wasCompressed {
		serializedPK = pk.SerializeCompressed()
	} else {
		serializedPK = pk.SerializeUncompressed()
	}
	address, err := btcutil.NewAddressPubKey(serializedPK, params)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error in NewAddressPubKey")
		os.Exit(100)
	}

	if address.EncodeAddress() == addrStr {
		fmt.Println("OK")
		os.Exit(0)
	}

	fmt.Fprintln(os.Stderr, "Signature mismatch")
	os.Exit(100)
}
