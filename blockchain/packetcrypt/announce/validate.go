// Copyright (c) 2019 Caleb James DeLisle
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package announce

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/pkt-cash/pktd/blockchain/packetcrypt/cryptocycle"
	"github.com/pkt-cash/pktd/blockchain/packetcrypt/difficulty"
	"github.com/pkt-cash/pktd/blockchain/packetcrypt/pcutil"
	"github.com/pkt-cash/pktd/blockchain/packetcrypt/randhash/util"
	"github.com/pkt-cash/pktd/chaincfg/chainhash"
	"github.com/pkt-cash/pktd/wire"
)

const announceMerkleDepth int = 13
const announceTableSz uint64 = 1 << uint(announceMerkleDepth)

type context struct {
	itemBytes [1024]byte
	ann       wire.PacketCryptAnn
	annHash0  [64]byte
	annHash1  [64]byte
	ccState   cryptocycle.State
	progBuf   cryptocycle.Context
	item4Hash [64]byte
}

func memocycle(item *[1024]byte, bufcount, cycles int) {
	var tmpbuf [128]byte
	for cycle := 0; cycle < cycles; cycle++ {
		for i := 0; i < bufcount; i++ {
			p := (i - 1 + bufcount) % bufcount
			q := int(binary.LittleEndian.Uint32(item[64*p:][:4]) % uint32(bufcount-1))
			j := (i + q) % bufcount
			copy(tmpbuf[:64], item[64*p:][:64])
			copy(tmpbuf[64:], item[64*j:][:64])
			pcutil.HashCompress64(item[i*64:][:64], tmpbuf[:])
		}
	}
}

const announceItemHashcount int = 1024 / 64

func MkItem(itemNo int, item *[1024]byte, seed []byte) {
	pcutil.HashExpand(item[:64], seed, uint32(itemNo))
	for i := 1; i < announceItemHashcount; i++ {
		pcutil.HashCompress64(item[64*i:][:64], item[64*(i-1):][:64])
	}
	memocycle(item, announceItemHashcount, util.Conf_AnnHash_MEMOHASH_CYCLES)
}

func merkleIsValid(merkleProof []byte, item4Hash *[64]byte, itemNo int) bool {
	var buf [128]byte
	copy(buf[64*(itemNo&1):][:64], item4Hash[:])
	for i := 0; i < announceMerkleDepth; i++ {
		copy(buf[64*((^itemNo)&1):][:64], merkleProof[i*64:][:64])
		itemNo >>= 1
		pcutil.HashCompress64(buf[64*(itemNo&1):][:64], buf[:])
	}
	return bytes.Compare(buf[64*(itemNo&1):][:64], merkleProof[64*announceMerkleDepth:]) == 0
}

func CheckAnn(pcAnn *wire.PacketCryptAnn, parentBlockHash *chainhash.Hash) (*chainhash.Hash, error) {
	ctx := new(context)
	copy(ctx.ann.GetAnnounceHeader(), pcAnn.GetAnnounceHeader())
	copy(ctx.ann.GetMerkleProof()[:32], parentBlockHash[:])
	pcutil.Zero(ctx.ann.GetSoftNonce())
	pcutil.HashCompress64(ctx.annHash0[:], ctx.ann.Header[:wire.PcAnnHeaderLen+64])
	copy(ctx.ann.GetMerkleProof(), pcAnn.GetMerkleProof()[13*64:])
	pcutil.HashCompress64(ctx.annHash1[:], ctx.ann.Header[:wire.PcAnnHeaderLen+64])

	var softNonceBuf [4]byte
	copy(softNonceBuf[:], pcAnn.GetSoftNonce())
	softNonce := binary.LittleEndian.Uint32(softNonceBuf[:])
	cryptocycle.Init(&ctx.ccState, ctx.annHash1[:32], uint64(softNonce))
	itemNo := -1
	for i := 0; i < 4; i++ {
		itemNo = int(cryptocycle.GetItemNo(&ctx.ccState) % announceTableSz)
		// only half of the seed is used
		MkItem(itemNo, &ctx.itemBytes, ctx.annHash0[:32])
		if !cryptocycle.Update(
			&ctx.ccState, ctx.itemBytes[:], nil, util.Conf_AnnHash_RANDHASH_CYCLES, &ctx.progBuf) {
			return nil, errors.New("Validate_checkAnn_INVAL")
		}
	}

	if bytes.Compare(ctx.itemBytes[:wire.PcItem4PrefixLen], pcAnn.GetItem4Prefix()) != 0 {
		return nil, errors.New("Validate_checkAnn_INVAL_ITEM4")
	}

	pcutil.HashCompress64(ctx.item4Hash[:], ctx.itemBytes[:])
	if !merkleIsValid(pcAnn.GetMerkleProof(), &ctx.item4Hash, itemNo) {
		return nil, errors.New("Validate_checkAnn_INVAL_MERKLE")
	}

	target := pcAnn.GetWorkTarget()
	cryptocycle.Final(&ctx.ccState)

	h := chainhash.Hash{}
	copy(h[:], ctx.ccState.Bytes[:32])
	if !difficulty.IsOk(ctx.ccState.Bytes[:32], target) {
		return &h, fmt.Errorf("Validate_checkAnn_INSUF_POW need target [%x] "+
			"but ann work hash is [%s]", target, h.String())
	}

	return &h, nil
}
