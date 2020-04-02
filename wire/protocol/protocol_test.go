// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package protocol_test

import (
	"testing"

	"github.com/pkt-cash/pktd/wire/protocol"
)

// TestServiceFlagStringer tests the stringized output for service flag types.
func TestServiceFlagStringer(t *testing.T) {
	tests := []struct {
		in   protocol.ServiceFlag
		want string
	}{
		{0, "0x0"},
		{protocol.SFNodeNetwork, "SFNodeNetwork"},
		{protocol.SFNodeGetUTXO, "SFNodeGetUTXO"},
		{protocol.SFNodeBloom, "SFNodeBloom"},
		{protocol.SFNodeWitness, "SFNodeWitness"},
		{protocol.SFNodeXthin, "SFNodeXthin"},
		{protocol.SFNodeBit5, "SFNodeBit5"},
		{protocol.SFNodeCF, "SFNodeCF"},
		{protocol.SFNode2X, "SFNode2X"},
		{0xffffffff, "SFNodeNetwork|SFNodeGetUTXO|SFNodeBloom|SFNodeWitness|SFNodeXthin|SFNodeBit5|SFNodeCF|SFNode2X|0xffffff00"},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		result := test.in.String()
		if result != test.want {
			t.Errorf("String #%d\n got: %s want: %s", i, result,
				test.want)
			continue
		}
	}
}

// TestBitcoinNetStringer tests the stringized output for bitcoin net types.
func TestBitcoinNetStringer(t *testing.T) {
	tests := []struct {
		in   protocol.BitcoinNet
		want string
	}{
		{protocol.MainNet, "MainNet"},
		{protocol.TestNet, "TestNet"},
		{protocol.TestNet3, "TestNet3"},
		{protocol.SimNet, "SimNet"},
		{0xffffffff, "Unknown BitcoinNet (4294967295)"},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		result := test.in.String()
		if result != test.want {
			t.Errorf("String #%d\n got: %s want: %s", i, result,
				test.want)
			continue
		}
	}
}
