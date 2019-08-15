// Copyright (c) 2019 Caleb James DeLisle
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package nativetest

import (
	"encoding/hex"
	"errors"
	"testing"

	"github.com/pkt-cash/pktd/blockchain/packetcrypt"
	"github.com/pkt-cash/pktd/chaincfg/chainhash"
	"github.com/pkt-cash/pktd/wire"
)

func validatePcBlock(mb *wire.MsgBlock, height int32, annParentHashes []*chainhash.Hash) error {
	if len(annParentHashes) != 4 {
		return errors.New("wrong number of annParentHashes")
	}
	if mb.Pcp == nil {
		return errors.New("missing packetcrypt proof")
	}
	coinbase := mb.Transactions[0]
	if coinbase == nil {
		return errors.New("missing coinbase")
	}
	cbc := packetcrypt.ExtractCoinbaseCommit(coinbase)
	if cbc == nil {
		return errors.New("missing packetcrypt commitment")
	}
	return validatePcProof(mb.Pcp, height, &mb.Header, cbc, annParentHashes)
}

func ValidatePcBlock(t *testing.T, mb *wire.MsgBlock, height int32, annParentHashes []*chainhash.Hash) error {
	err1 := validatePcBlock(mb, height, annParentHashes)
	_, err2 := packetcrypt.ValidatePcBlock(mb, height, 0, annParentHashes)
	if err1 != err2 {
		if err2 != nil {
			t.Errorf("go ValidatePcBlock() -> %v", err2)
		}
		if err1 != nil {
			t.Errorf("C ValidatePcBlock() -> %v", err1)
		}
	}
	return err2
}

func ValidatePcAnn(t *testing.T, p *wire.PacketCryptAnn, parentBlockHash *chainhash.Hash) (*chainhash.Hash, error) {
	hash1, err1 := validatePcAnn(p, parentBlockHash)
	hash2, err2 := packetcrypt.ValidatePcAnn(p, parentBlockHash)
	if err1 != err2 {
		t.Errorf("%v != %v", err1, err2)
	} else if hash1 != nil {
		if hash2 == nil {
			t.Errorf("hash mismatch")
		} else if !hash1.IsEqual(hash2) {
			t.Errorf("hash mismatch %v != %v",
				hex.EncodeToString(hash1[:]), hex.EncodeToString(hash2[:]))
		}
	} else if hash2 == nil {
		t.Errorf("hash mismatch")
	}
	return hash1, err1
}
