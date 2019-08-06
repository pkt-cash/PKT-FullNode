// Copyright (c) 2019 Caleb James DeLisle
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package nativetest

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"testing"
	"unsafe"

	"github.com/pkt-cash/pktd/blockchain/packetcrypt/pcutil"
	"github.com/pkt-cash/pktd/blockchain/packetcrypt/randhash/opcodes"
	"github.com/pkt-cash/pktd/blockchain/packetcrypt/randhash/util"

	"github.com/pkt-cash/pktd/chaincfg/chainhash"
	"github.com/pkt-cash/pktd/wire"
)

// #cgo CFLAGS:  -I/Users/user/wrk/play/PacketCrypt/include
// #cgo LDFLAGS: -L/Users/user/wrk/play/PacketCrypt -lpacketcrypt -lcrypto -L/usr/local/Cellar/libsodium/1.0.17/lib/ -lsodium
// #include "packetcrypt/AnnMiner.h"
// #include "packetcrypt/BlockMiner.h"
// #include "packetcrypt/PacketCrypt.h"
// #include "packetcrypt/Validate.h"
//
// void CryptoCycle_crypt(void* msg);
// void CryptoCycle_init(void* state, void* seed, uint64_t nonce);
// void CryptoCycle_update(void* state, void* item, int rhCycles, void* progBuf);
// void CryptoCycle_final(void* state);
//
// int RandGen_generate(void* buf, void* seed);
// void Hash_expand(void* buff, uint32_t len, void* seed, uint32_t num);
// int RandHash_interpret(void* prog, void* ccState, void* memory, int progLen, uint32_t memorySizeBytes, int cycles);
// void RandHashOps_doOp(void* inout, uint32_t op);
//
// #include <stdlib.h>
import "C"

func CryptoCycle(msg []byte) {
	ptr := C.CBytes(msg)
	C.CryptoCycle_crypt(ptr)
	msg2 := C.GoBytes(ptr, C.int(len(msg)))
	copy(msg, msg2)
	C.free(ptr)
}

type PcAnn struct {
	ch           chan wire.PacketCryptAnn
	pipeR, pipeW *os.File
	annMiner     *C.struct_AnnMiner_s
}

func freePcAnn(pc *PcAnn) {
	pc.pipeR.Close()
	pc.pipeW.Close()
	C.AnnMiner_free(pc.annMiner)
}

func (p *PcAnn) Start(
	contentHash []byte,
	contentType uint64,
	workTarget uint32,
	parentBlockHeight uint32,
	parentBlockHash *chainhash.Hash,
) {
	annHeader := make([]byte, 56)
	binary.LittleEndian.PutUint32(annHeader[8:12], workTarget)
	binary.LittleEndian.PutUint32(annHeader[12:16], parentBlockHeight)
	binary.LittleEndian.PutUint64(annHeader[16:24], contentType)
	copy(annHeader[24:], contentHash)

	ahP := C.CBytes(annHeader)
	pbhP := C.CBytes(parentBlockHash[:])

	ah := (*C.PacketCrypt_AnnounceHdr_t)(ahP)
	pbha := (*C.uchar)(pbhP)
	C.AnnMiner_start(p.annMiner, ah, pbha)

	C.free(ahP)
	C.free(pbhP)
}

func (p *PcAnn) Stop() {
	C.AnnMiner_stop(p.annMiner)
}

func getThing(f *os.File) *bytes.Buffer {
	var buf [16]byte
	if _, err := io.ReadFull(f, buf[:]); err != nil {
		if err == io.ErrClosedPipe {
			fmt.Printf("Pipe has closed")
		} else {
			fmt.Printf("Unexpected error reading from pipe %v", err)
		}
		return nil
	}
	ptru64 := binary.LittleEndian.Uint64(buf[0:8])
	ptr := unsafe.Pointer(uintptr(ptru64))
	length := C.int(binary.LittleEndian.Uint64(buf[8:16]))
	b := C.GoBytes(ptr, length)
	C.free(ptr)
	return bytes.NewBuffer(b)
}

type PcBlockMineResult struct {
	Header wire.BlockHeader
	Pcp    wire.PacketCryptProof
}

func blockReader(bm *PcBlk, ch chan PcBlockMineResult, f *os.File) {
	for {
		b := getThing(f)
		if b == nil {
			return
		}
		res := PcBlockMineResult{}
		if err := res.Header.BtcDecode(b, 0, wire.PcAnnNoContent); err != nil {
			panic("failed to decode block header")
		}
		var buf [8]byte
		if _, err := io.ReadFull(b, buf[:]); err != nil {
			panic("failed to read buffer")
		}
		res.Pcp.Nonce = binary.LittleEndian.Uint32(buf[4:])
		for i := 0; i < 4; i++ {
			if _, err := io.ReadFull(b, res.Pcp.Announcements[i].Header[:]); err != nil {
				panic("failed to read announcement")
			}
		}
		res.Pcp.AnnProof = make([]byte, b.Len())
		if _, err := io.ReadFull(b, res.Pcp.AnnProof[:]); err != nil {
			panic("failed to read AnnProof")
		}
		ch <- res
	}
}

func annReader(bm *PcAnn, ch chan wire.PacketCryptAnn, f *os.File) {
	for {
		b := getThing(f)
		if b == nil {
			return
		}
		res := wire.PacketCryptAnn{}
		if err := res.BtcDecode(b, 0, wire.PcAnnNoContent); err != nil {
			panic("failed to decode announcement")
		}
		ch <- res
	}
}

func validatePcAnn(p *wire.PacketCryptAnn, parentBlockHash *chainhash.Hash) (*chainhash.Hash, error) {
	annPtr := C.CBytes(p.Header[:])
	hashPtr := C.CBytes(parentBlockHash[:])
	vctx := C.malloc(C.sizeof_PacketCrypt_ValidateCtx_t)
	annHash := C.calloc(32, 1)
	ret := C.Validate_checkAnn(
		(*C.uint8_t)(annHash),
		(*C.PacketCrypt_Announce_t)(annPtr),
		(*C.uint8_t)(hashPtr),
		(*C.PacketCrypt_ValidateCtx_t)(vctx),
	)
	hash := C.GoBytes(annHash, 32)
	chh := chainhash.Hash{}
	copy(chh[:], hash)
	C.free(annHash)
	C.free(annPtr)
	C.free(hashPtr)
	C.free(vctx)
	if ret == 0 {
		return &chh, nil
	}
	if ret == 1 {
		return nil, errors.New("Validate_checkAnn_INVAL")
	}
	if ret == 2 {
		return nil, errors.New("  Validate_checkAnn_INVAL_ITEM4")
	}
	if ret == 3 {
		return &chh, errors.New("Validate_checkAnn_INSUF_POW")
	}
	return nil, errors.New("unknown error")
}

func PcAnnNew(ch chan wire.PacketCryptAnn, numWorkers uint32) (*PcAnn, error) {
	pipeR, pipeW, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	threads := C.int(numWorkers)
	outFiles := make([]byte, 4)
	binary.LittleEndian.PutUint32(outFiles, uint32(pipeW.Fd()))
	outFilesP := C.CBytes(outFiles)
	numOutFiles := C.int(1)
	sendPtr := C.int(1)
	am := C.AnnMiner_create(threads, (*C.int)(outFilesP), numOutFiles, sendPtr)
	C.free(outFilesP)
	out := PcAnn{
		ch:       ch,
		pipeR:    pipeR,
		pipeW:    pipeW,
		annMiner: am,
	}
	go annReader(&out, ch, pipeR)
	runtime.SetFinalizer(&out, freePcAnn)
	return &out, nil
}

/// Block miner

type PcBlk struct {
	ch           chan PcBlockMineResult
	pipeR, pipeW *os.File
	blkMiner     *C.struct_BlockMiner_s
}

const checkBlockMask C.int = ^0 << 8
const checkBlock_OK C.int = 0
const checkBlock_RUNT C.int = 1 << 8
const checkBlock_ANN_INVALID C.int = 2 << 8
const checkBlock_ANN_INSUF_POW C.int = 3 << 8
const checkBlock_PCP_INVAL C.int = 4 << 8
const checkBlock_PCP_MISMATCH C.int = 5 << 8
const checkBlock_INSUF_POW C.int = 6 << 8
const checkBlock_BAD_COINBASE C.int = 7 << 8

// use ValidatePcBlock() instead
func validatePcProof(
	pcp *wire.PacketCryptProof,
	blockHeight int32,
	blockHeader *wire.BlockHeader,
	cb *wire.PcCoinbaseCommit,
	blockHashes []*chainhash.Hash,
) error {
	if len(blockHashes) != 4 {
		return errors.New("blockHashes invalid length")
	}

	hapLen := wire.MaxBlockHeaderPayload+4+4+(1024*4)+len(pcp.AnnProof)

	buf := bytes.NewBuffer(
		make([]byte, 0, hapLen+32*4))
	if err := blockHeader.BtcEncode(buf, 0, wire.PcAnnNoContent); err != nil {
		return err
	}
	hn := make([]byte, 8)
	binary.LittleEndian.PutUint32(hn[4:], pcp.Nonce)
	if _, err := buf.Write(hn); err != nil {
		return err;
	}
	for _, ann := range pcp.Announcements {
		if _, err := buf.Write(ann.Header[:]); err != nil {
			return err;
		}
	}
	if _, err := buf.Write(pcp.AnnProof); err != nil {
		return err;
	}

	for i := 0; i < len(blockHashes); i++ {
		if len(blockHashes[i][:]) != 32 {
			return errors.New("one of the blockHashes is not 32 bytes")
		}
		buf.Write(blockHashes[i][:])
	}
	cbcPtr := C.CBytes(cb.Bytes[:])
	ptr := C.CBytes(buf.Bytes())
	hashes := unsafe.Pointer(uintptr(ptr) + uintptr(hapLen))
	vctx := C.malloc(C.sizeof_PacketCrypt_ValidateCtx_t)
	workHashOut := C.malloc(32)
	ret := C.Validate_checkBlock(
		(*C.PacketCrypt_HeaderAndProof_t)(ptr),
		C.uint32_t(hapLen),
		C.uint32_t(blockHeight),
		C.uint32_t(blockHeader.Bits),
		(*C.PacketCrypt_Coinbase_t)(cbcPtr),
		(*C.uint8_t)(hashes),
		(*C.uchar)(workHashOut),
		(*C.PacketCrypt_ValidateCtx_t)(vctx),
	)
	C.free(cbcPtr)
	C.free(ptr)
	C.free(vctx)
	C.free(workHashOut)
	if ret == 0 {
		return nil
	} else if ret&checkBlockMask == checkBlock_RUNT {
		return errors.New("checkBlock_RUNT")
	} else if ret&checkBlockMask == checkBlock_ANN_INVALID {
		return errors.New("checkBlock_ANN_INVALID")
	} else if ret&checkBlockMask == checkBlock_ANN_INSUF_POW {
		return errors.New("checkBlock_ANN_INSUF_POW")
	} else if ret&checkBlockMask == checkBlock_PCP_INVAL {
		return errors.New("checkBlock_PCP_INVAL")
	} else if ret&checkBlockMask == checkBlock_PCP_MISMATCH {
		return errors.New("checkBlock_PCP_MISMATCH")
	} else if ret&checkBlockMask == checkBlock_INSUF_POW {
		return errors.New("checkBlock_INSUF_POW")
	} else if ret&checkBlockMask == checkBlock_BAD_COINBASE {
		return errors.New("checkBlock_BAD_COINBASE")
	} else {
		return errors.New("unknown error")
	}
}

func (bm *PcBlk) AddAnns(anns []*wire.PacketCryptAnn) {
	//ptr := C.malloc(C.size_t(len(anns) * 1024))
	//bytes := C.GoBytes(ptr, C.int(len(anns)*1024))
	bytes := make([]byte, len(anns)*1024)
	for i := 0; i < len(anns); i++ {
		if anns[i].MissingContent() {
			panic("AddAnns called on announcement with content missing")
		}
		if anns[i].GetWorkTarget() == 0 {
			panic("insane announcement passed to AddAnns")
		}
		copy(bytes[i*1024:(i+1)*1024], anns[i].Header[:])
	}
	ptr := C.CBytes(bytes)
	C.BlockMiner_addAnns(bm.blkMiner, (*C.PacketCrypt_Announce_t)(ptr), C.uint64_t(len(anns)), 1)
}

func (bm *PcBlk) Start(header *wire.BlockHeader) error {
	headerBuf := bytes.NewBuffer(make([]byte, 0, wire.MaxBlockHeaderPayload))
	_ = header.BtcEncode(headerBuf, 0, 0)
	headerBytes := headerBuf.Bytes()
	ptr := C.CBytes(headerBytes)
	res := C.BlockMiner_start(bm.blkMiner, (*C.PacketCrypt_BlockHeader_t)(ptr))
	C.free(ptr)
	if res == 0 {
		return nil
	}
	if res == 1 {
		return errors.New("Not yet locked for mining")
	}
	if res == 2 {
		return errors.New("Already mining")
	}
	return errors.New("unknown error")
}

func (bm *PcBlk) LockForMining(nextBlockHeight int32, nextBlockTarget uint32) (*wire.PcCoinbaseCommit, error) {
	ptr := C.calloc(C.sizeof_PacketCrypt_Coinbase_t+C.sizeof_uintptr_t, 1)
	ret := C.BlockMiner_lockForMining(
		bm.blkMiner,
		(*C.BlockMiner_Stats_t)(C.NULL),
		(*C.PacketCrypt_Coinbase_t)(ptr),
		C.uint32_t(nextBlockHeight),
		C.uint32_t(nextBlockTarget))
	if ret != 0 {
		C.free(ptr)
		if ret == 1 {
			return nil, errors.New("no anns")
		}
		return nil, errors.New("unknown error")
	}
	b := C.GoBytes(ptr, C.sizeof_PacketCrypt_Coinbase_t)
	out := &wire.PcCoinbaseCommit{}
	copy(out.Bytes[:], b)
	C.free(ptr)
	return out, nil

}

func PcTestSizes(t *testing.T) {
	cbc := wire.PcCoinbaseCommit{}
	if C.sizeof_PacketCrypt_Coinbase_t != len(cbc.Bytes) {
		t.Errorf(" C.sizeof_PacketCrypt_Coinbase_t != len(wire.PcCoinbaseCommit.Bytes)")
	}
}

func (bm *PcBlk) Stop() error {
	ret := C.BlockMiner_stop(bm.blkMiner)
	if ret == 0 {
		return nil
	}
	if ret == 1 {
		return errors.New("Not mining")
	}
	return errors.New("unknown error")
}

func freePcBlk(pc *PcBlk) {
	C.BlockMiner_free(pc.blkMiner)
	pc.pipeR.Close()
	pc.pipeW.Close()
}

func PcBlkNew(maxAnns uint64, ch chan PcBlockMineResult, numWorkers uint32) (*PcBlk, error) {
	pipeR, pipeW, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	ma := C.uint64_t(maxAnns)
	thr := C.int(numWorkers)
	pfd := C.int(pipeW.Fd())
	sendPtr := C._Bool(true)
	out := PcBlk{
		ch:         ch,
		pipeR:      pipeR,
		pipeW:      pipeW,
		blkMiner:   C.BlockMiner_create(ma, C.uint32_t(0), thr, pfd, sendPtr),
	}
	go blockReader(&out, ch, pipeR)
	runtime.SetFinalizer(&out, freePcBlk)
	return &out, nil
}

func CryptoCycleInit(state, seed []byte, nonce uint64) {
	if len(state) != 2048 || len(seed) != 32 {
		panic("bad lengths")
	}
	stateP := C.CBytes(state)
	seedP := C.CBytes(seed)
	C.CryptoCycle_init(stateP, seedP, C.uint64_t(nonce))
	state1 := C.GoBytes(stateP, C.int(len(state)))
	copy(state, state1)
	C.free(stateP)
	C.free(seedP)
}

func CryptoCycleUpdate(state, item []byte, rhCycles int) {
	if len(state) != 2048 || len(item) != 1024 {
		panic("bad lengths")
	}
	stateP := C.CBytes(state)
	itemP := C.CBytes(item)
	progBufP := C.malloc(util.Conf_RandGen_MAX_INSNS * 4)
	C.CryptoCycle_update(stateP, itemP, C.int(rhCycles), progBufP)
	state1 := C.GoBytes(stateP, C.int(len(state)))
	copy(state, state1)
	C.free(stateP)
	C.free(itemP)
	C.free(progBufP)
}

func CryptoCycleFinal(state []byte) {
	if len(state) != 2048 {
		panic("bad lengths")
	}
	stateP := C.CBytes(state)
	C.CryptoCycle_final(stateP)
	state1 := C.GoBytes(stateP, C.int(len(state)))
	copy(state, state1)
	C.free(stateP)
}

func Generate(seed []byte) ([]uint32, error) {
	out2C := C.malloc(4 * util.Conf_RandGen_MAX_INSNS)
	seedC := C.CBytes(seed)
	ret := C.RandGen_generate(out2C, seedC)
	if ret < 0 {
		if ret == -2 {
			return nil, errors.New("insn count < Conf_RandGen_MIN_INSNS")
		} else if ret == -1 {
			return nil, errors.New("insn count > Conf_RandGen_MAX_INSNS")
		} else {
			return nil, errors.New("unknown error")
		}
	}
	out2 := C.GoBytes(out2C, 4*ret)
	C.free(out2C)
	C.free(seedC)

	return pcutil.U32FromB(nil, out2), nil
}

func Interpret(prog []uint32, ccState, memory []byte, cycles int) error {

	ccStateP := C.CBytes(ccState)
	progP := C.CBytes(pcutil.BFromU32(nil, prog))
	memoryP := C.CBytes(memory)

	ret := C.RandHash_interpret(
		progP,
		ccStateP,
		memoryP,
		C.int(len(prog)),
		C.uint32_t(len(memory)),
		C.int(cycles))

	if ret != 0 {
		if ret == -1 {
			return errors.New("RandHash_TOO_BIG")
		} else if ret == -2 {
			return errors.New("RandHash_TOO_SMALL")
		} else if ret == -3 {
			return errors.New("RandHash_TOO_LONG")
		} else if ret == -4 {
			return errors.New("RandHash_TOO_SHORT")
		} else {
			return errors.New("unknown error")
		}
	}

	ccState2B := C.GoBytes(ccStateP, C.int(len(ccState)))
	copy(ccState, ccState2B)

	C.free(ccStateP)
	C.free(progP)
	C.free(memoryP)

	return nil
}

func HashExpand(len uint32, seed []uint8, num uint32) []byte {
	outC := C.malloc(C.size_t(len))
	seedC := C.CBytes(seed)
	C.Hash_expand(outC, C.uint32_t(len), seedC, C.uint32_t(num))
	out := C.GoBytes(outC, C.int(len))
	C.free(outC)
	C.free(seedC)
	return out
}

func RandHashDoOp(inout []uint32, op opcodes.OpCode) {
	inoutP := C.CBytes(pcutil.BFromU32(nil, inout))
	C.RandHashOps_doOp(inoutP, C.uint32_t(op))
	inoutB := C.GoBytes(inoutP, C.int(len(inout)*4))
	C.free(inoutP)
	pcutil.U32FromB(inout, inoutB)
}
