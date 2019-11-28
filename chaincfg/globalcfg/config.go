// Copyright (c) 2019 Caleb James DeLisle
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

// Package globalcfg contains configuration which must be available
// anywhere in the project, do not import anything which is part of pktd.
package globalcfg

import (
	"fmt"
	"time"
)

const (
	// SatoshiPerBitcoin is the number of satoshi in one bitcoin (1 BTC).
	satoshiPerBitcoin = 1e8

	// MaxSatoshi is the maximum transaction amount allowed in satoshi.
	maxSatoshi = 21e6 * satoshiPerBitcoin
)

// ProofOfWork means the type of proof of work used on the chain
type ProofOfWork int

const (
	// PowSha256 is the original proof of work from satoshi.
	// This is the default value
	PowSha256 ProofOfWork = iota

	// PowPacketCrypt is the PoW used by chains such as pkt.cash
	PowPacketCrypt
)

// Config is the global config which is accessible anywhere in the app
type Config struct {
	ProofOfWorkAlgorithm ProofOfWork
	HasNetworkSteward    bool
	MaxSatoshi           int64
	SatoshiPerBitcoin    int64
	MaxTimeOffsetSeconds time.Duration
	MedianTimeBlocks     int
}

var gConf Config
var registered bool

// BitcoinDefaults creates a new config with the default values for bitcoin
func BitcoinDefaults() Config {
	return Config{
		ProofOfWorkAlgorithm: PowSha256,
		HasNetworkSteward:    false,
		MaxSatoshi:           maxSatoshi,
		SatoshiPerBitcoin:    satoshiPerBitcoin,
		MaxTimeOffsetSeconds: 2 * 60 * 60,
		MedianTimeBlocks:     11,
	}
}

// SelectConfig is used to register the blockchain parameters with globalcfg
func SelectConfig(conf Config) bool {
	if registered {
		return false
	}
	registered = true
	gConf = conf
	return true
}

// RemoveConfig deletes the config, used in tests
func RemoveConfig() bool {
	if !registered {
		return false
	}
	fmt.Printf("Configuration removed\n")
	registered = false
	gConf = Config{}
	return true
}

func checkRegistered() {
	if !registered {
		panic("globalcfg requested but not yet registered")
	}
}

// GetMaxTimeOffsetSeconds is the maximum number of seconds a block time
// is allowed to be ahead of the current time.
func GetMaxTimeOffsetSeconds() time.Duration {
	checkRegistered()
	return gConf.MaxTimeOffsetSeconds
}

// GetMedianTimeBlocks provides the number of previous blocks which should be
// used to calculate the median time used to validate block timestamps.
func GetMedianTimeBlocks() int {
	checkRegistered()
	return gConf.MedianTimeBlocks
}

// GetProofOfWorkAlgorithm tells whether the chain in use uses a custom proof
// of work algorithm or the normal sha256 proof of work.
func GetProofOfWorkAlgorithm() ProofOfWork {
	checkRegistered()
	return gConf.ProofOfWorkAlgorithm
}

// IsPacketCryptAllowedVersion tells whether the specified version of PacketCrypt proof is allowed.
func IsPacketCryptAllowedVersion(version int, blockHeight int32) bool {
	checkRegistered()
	return version <= 1 || blockHeight >= 113949
}

// HasNetworkSteward returns true for blockchains which require a network steward fee
func HasNetworkSteward() bool {
	checkRegistered()
	return gConf.HasNetworkSteward
}

// SatoshiPerBitcoin returns the number of atomic units per "coin"
func SatoshiPerBitcoin() int64 {
	checkRegistered()
	return gConf.SatoshiPerBitcoin
}

// MaxSatoshi returns the maximum number of atomic units of currency
func MaxSatoshi() int64 {
	checkRegistered()
	return gConf.MaxSatoshi
}
