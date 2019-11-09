// Copyright (c) 2019 Caleb James DeLisle
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package nativetest_test

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/pkt-cash/pktd/blockchain"
	"github.com/pkt-cash/pktd/blockchain/packetcrypt"
	"github.com/pkt-cash/pktd/blockchain/packetcrypt/cryptocycle"
	"github.com/pkt-cash/pktd/blockchain/packetcrypt/nativetest"
	"github.com/pkt-cash/pktd/blockchain/packetcrypt/pcutil"
	"github.com/pkt-cash/pktd/blockchain/packetcrypt/randhash/interpret"
	"github.com/pkt-cash/pktd/blockchain/packetcrypt/randhash/opcodes"
	"github.com/pkt-cash/pktd/blockchain/packetcrypt/randhash/randgen"
	"github.com/pkt-cash/pktd/blockchain/testdata"
	"github.com/pkt-cash/pktd/btcutil"
	"github.com/pkt-cash/pktd/chaincfg"
	"github.com/pkt-cash/pktd/chaincfg/chainhash"
	"github.com/pkt-cash/pktd/chaincfg/globalcfg"
	"github.com/pkt-cash/pktd/txscript"
	"github.com/pkt-cash/pktd/wire"
)

func withPcPow(f func()) {
	if !globalcfg.SelectConfig(globalcfg.Config{
		ProofOfWorkAlgorithm: globalcfg.PowPacketCrypt,
	}) {
		panic("globalcfg already set")
	}
	f()
	if !globalcfg.RemoveConfig() {
		panic("globalcfg removed")
	}
}

func getAnnHashes() []*chainhash.Hash {
	hashes := make([]*chainhash.Hash, 4)
	for i := 0; i < 4; i++ {
		hashes[i] = &chainhash.Hash{}
	}
	return hashes
}
func TestValidate(t *testing.T) {
	withPcPow(func() {
		block := testdata.GetBlock("../../testdata/277647.withpcp.dat", t)
		if block == nil {
			return
		}
		hashes := getAnnHashes()
		if err := nativetest.ValidatePcBlock(t, block.MsgBlock(), 277647, hashes); err != nil {
			t.Errorf("ValidatePcProof() %v", err)
		}
	})
}

func TestSizes(t *testing.T) {
	nativetest.PcTestSizes(t)
}

func getAnnouncements(t *testing.T) []*wire.PacketCryptAnn {
	anns, err := testdata.LoadAnnouncements("../../testdata/announcements.277640.dat")
	if err != nil {
		t.Errorf("Error loading file: %v\n", err)
		return nil
	}
	return anns
}

func TestParseAnnouncements(t *testing.T) {
	getAnnouncements(t)
}

func TestMine(t *testing.T) {
	block := testdata.GetBlock("../../testdata/277647.dat.bz2", t)
	block.SetHeight(277647)

	anns := getAnnouncements(t)
	if block == nil || anns == nil {
		return
	}
	mb := block.MsgBlock()
	mb.Header.Bits = 0x207fffff

	ch := make(chan nativetest.PcBlockMineResult)
	pcBlk, err := nativetest.PcBlkNew(100000, ch, 4)
	if err != nil {
		t.Errorf("PcBlkNew [%v]", err)
		return
	}

	pcBlk.AddAnns(anns)
	cbc, err := pcBlk.LockForMining(block.Height(), mb.Header.Bits)
	if err != nil {
		t.Errorf("LockForMining [%v]", err)
		return
	}
	packetcrypt.InsertCoinbaseCommit(mb.Transactions[0], cbc)

	merkles := blockchain.BuildMerkleTreeStore(block.Transactions(), false)
	copy(mb.Header.MerkleRoot[:], merkles[len(merkles)-1].CloneBytes())

	pcBlk.Start(&mb.Header)

	for i := 0; i < 1; i++ {
		proof := <-ch

		mb.Header.Timestamp = proof.Header.Timestamp
		mb.Header.Nonce = proof.Header.Nonce
		mb.Pcp = &proof.Pcp

		hashes := getAnnHashes()
		for i, ann := range proof.Pcp.Announcements {
			_, err := nativetest.ValidatePcAnn(t, &ann, hashes[i])
			if err != nil {
				t.Errorf("ann content %v", hex.EncodeToString(ann.Header[:]))
				t.Errorf("%v", err)
			}
		}

		if err := nativetest.ValidatePcBlock(t, mb, block.Height(), hashes); err != nil {
			t.Errorf("ValidatePcProof() %v", err)
		}
	}

	pcBlk.Stop()

	withPcPow(func() {
		// for generating test data...
		return
		if err := testdata.OutputBlock(mb, "./277647.withpcp.dat"); err != nil {
			t.Errorf("OutputBlock() %v", err)
		}
	})
}

func mineGenesisAnnouncements(t *testing.T) []*wire.PacketCryptAnn {
	parentHash := chainhash.Hash{}
	parentBlockHeight := 0
	ch := make(chan wire.PacketCryptAnn)
	m, err := nativetest.PcAnnNew(ch, 4)
	if err != nil {
		t.Errorf("%v", err)
		return nil
	}
	m.Start(0, nil, 0x200fffff, uint32(parentBlockHeight), &parentHash, nil)
	var anns [4096]*wire.PacketCryptAnn
	for i := 0; i < 4096; i++ {
		ann := <-ch
		anns[i] = &ann
	}
	go func() { m.Stop() }()
	for len(ch) > 0 {
		<-ch
	}
	return anns[:]
}

func mkGenesis(t *testing.T) *wire.MsgBlock {
	params := chaincfg.PktTestNetParams
	sigScript, err := txscript.NewScriptBuilder().
		AddInt64(int64(0)).
		Script()
	if err != nil {
		t.Errorf("making genesis block script %v", err)
		return nil
	}
	tx := wire.NewMsgTx(wire.TxVersion)
	tx.AddTxIn(&wire.TxIn{
		// Coinbase transactions have no inputs, so previous outpoint is
		// zero hash and max index.
		PreviousOutPoint: *wire.NewOutPoint(&chainhash.Hash{}, wire.MaxPrevOutIndex),
		SignatureScript:  sigScript,
		Sequence:         wire.MaxTxInSequenceNum,
	})
	subsidy := blockchain.CalcBlockSubsidy(0, &params)
	tx.AddTxOut(&wire.TxOut{Value: subsidy, PkScript: params.InitialNetworkSteward})
	blk := wire.MsgBlock{}
	blk.Header.Bits = params.PowLimitBits
	blk.Transactions = []*wire.MsgTx{tx}
	return &blk
}
func TestMineGenesis(t *testing.T) {
	withPcPow(func() {
		//params := chaincfg.PktTestNetParams
		mb := mkGenesis(t)
		//mb := params.GenesisBlock
		fmt.Printf("Begin mining announcements\n")
		anns := mineGenesisAnnouncements(t)
		fmt.Printf("Mining announcements complete\n")
		if anns == nil {
			return
		}
		ch := make(chan nativetest.PcBlockMineResult)
		m, err := nativetest.PcBlkNew(10000, ch, 4)
		if err != nil {
			t.Errorf("%v", err)
			return
		}
		m.AddAnns(anns)
		cbc, err := m.LockForMining(0, mb.Header.Bits)
		if err != nil {
			t.Errorf("LockForMining() %v", err)
			return
		}

		packetcrypt.InsertCoinbaseCommit(mb.Transactions[0], cbc)
		block := btcutil.NewBlock(mb)
		merkles := blockchain.BuildMerkleTreeStore(block.Transactions(), false)
		copy(mb.Header.MerkleRoot[:], merkles[len(merkles)-1].CloneBytes())

		fmt.Printf("Begin mining block\n")
		if err := m.Start(&mb.Header); err != nil {
			t.Errorf("m.Start() %v", err)
			return
		}

		proof := <-ch

		go func() { m.Stop() }()
		for len(ch) > 0 {
			<-ch
		}

		mb.Header.Timestamp = proof.Header.Timestamp
		mb.Header.Nonce = proof.Header.Nonce
		mb.Pcp = &proof.Pcp

		if err := testdata.OutputBlock(mb, "./genesis.dat"); err != nil {
			t.Errorf("mb.Serialize() %v", err)
			return
		}
		fmt.Printf("Genesis block written to './genesis.dat'\n")
	})
}

func TestMineAnn(t *testing.T) {
	parentHash := chainhash.Hash{}
	copy(parentHash[:], []byte("abcdefghijklmnopqrstuvwxyz012345"))
	parentBlockHeight := 123
	ch := make(chan wire.PacketCryptAnn)
	m, err := nativetest.PcAnnNew(ch, 4)
	if err != nil {
		t.Errorf("%v", err)
	}
	m.Start(uint32(0), nil, 0x207fffff, uint32(parentBlockHeight), &parentHash, nil)
	ann := <-ch
	go func() { m.Stop() }()
	for len(ch) > 0 {
		<-ch
	}
	if _, err = nativetest.ValidatePcAnn(t, &ann, &parentHash); err != nil {
		t.Errorf("%v", hex.EncodeToString(ann.Header[:]))
		t.Errorf("Failed validation      %v", err)
	}
	//t.Errorf("%v", hex.EncodeToString(ann.Bytes))
}
func TestCryptoCycle(t *testing.T) {
	state1 := make([]byte, 2048)
	ccState := new(cryptocycle.State)
	item1 := make([]byte, 1024)
	seed := make([]byte, 32)
	ctx := new(cryptocycle.Context)
	for i := 0; i < 1000; i++ {
		//fmt.Printf("x\n")
		nativetest.CryptoCycleInit(state1, seed, uint64(i))
		cryptocycle.Init(ccState, seed, uint64(i))

		if bytes.Compare(ccState.Bytes[:], state1) != 0 {
			t.Errorf("init different with %v", hex.EncodeToString(seed))
			return
		}

		nativetest.CryptoCycleUpdate(state1, item1, 4)
		cryptocycle.Update(ccState, item1, nil, 4, ctx)

		if bytes.Compare(ccState.Bytes[:], state1) != 0 {
			t.Errorf("update different with %v", hex.EncodeToString(seed))
			t.Errorf("glang  %v", hex.EncodeToString(ccState.Bytes[0:64]))
			t.Errorf("native %v", hex.EncodeToString(state1[0:64]))
			t.Errorf("%v %v", ccState.GetAddLen(), ccState.GetLength())
			pcutil.Zero(ccState.Bytes[16:32])
			pcutil.Zero(state1[16:32])
			for i := 0; i < 2048; i++ {
				if ccState.Bytes[i] != state1[i] {
					fmt.Printf("%d %02x  %02x\n", i, ccState.Bytes[i], state1[i])
				}
			}

			return
		}

		nativetest.CryptoCycleFinal(state1)
		cryptocycle.Final(ccState)

		if bytes.Compare(ccState.Bytes[:], state1) != 0 {
			t.Errorf("final different with %v", hex.EncodeToString(seed))
			return
		}

		rand.Read(seed)
		rand.Read(state1)
		rand.Read(ccState.Bytes[:])
		rand.Read(item1)
	}
}

func TestValidateAnns(t *testing.T) {
	parentHash := chainhash.Hash{}
	copy(parentHash[:], []byte("abcdefghijklmnopqrstuvwxyz012345"))
	parentBlockHeight := 123
	ch := make(chan wire.PacketCryptAnn)
	m, err := nativetest.PcAnnNew(ch, 4)
	if err != nil {
		t.Errorf("%v", err)
	}
	m.Start(uint32(0), nil, 0x207fffff, uint32(parentBlockHeight), &parentHash, nil)
	for i := 0; i < 1000; i++ {
		ann := <-ch
		_, err := nativetest.ValidatePcAnn(t, &ann, &parentHash)
		if err != nil {
			t.Errorf("%v", hex.EncodeToString(ann.Header[:]))
			t.Errorf("Failed validation      %v", err)
			break
		}
	}
	go func() { m.Stop() }()
	for len(ch) > 0 {
		<-ch
	}
	//t.Errorf("%v", hex.EncodeToString(ann.Bytes))
}

func TestValidateAnn(t *testing.T) {
	anns := getAnnouncements(t)

	// This is the default parent block hash used for generating
	// dummy announcements in pcann
	parentHash := getAnnHashes()[0]

	for _, ann := range anns {
		_, err := nativetest.ValidatePcAnn(t, ann, parentHash)
		if err != nil {
			t.Errorf("%v", hex.EncodeToString(ann.Header[:]))
			t.Errorf("Failed validation      %v", err)
			break
		}
		break
	}
}

// RandHash tests

func TestHashExpand(t *testing.T) {
	key := make([]byte, 32)
	for x := 0; x < 10000; x++ {
		rand.Read(key)
		a := nativetest.HashExpand(uint32(x%64), key, uint32(x))
		b := make([]byte, x%64)
		pcutil.HashExpand(b, key, uint32(x))
		if bytes.Compare(a, b) != 0 {
			t.Errorf("different %v %v %v", x, a, b)
			return
		}
	}
}

func generateProg(t *testing.T, key []byte) []uint32 {
	out1, err1 := randgen.Generate(key)
	out2, err2 := nativetest.Generate(key)
	if err1 != err2 {
		t.Errorf("err1 != err2:  %v != %v", err1, err2)
		return nil
	}
	if len(out1) != len(out2) {
		t.Errorf("len(out1) != len(out2):  %v != %v", len(out1), len(out2))
		return nil
	}
	for i := 0; i < len(out1); i++ {
		if out1[i] != out2[i] {
			t.Errorf("out1[%v] != out2[%v]:  %v != %v", i, i, out1[i], out2[i])
			return nil
		}
	}
	return out1
}

func TestGenerate(t *testing.T) {
	key := make([]byte, 32)
	for x := 0; x < 100; x++ {
		rand.Read(key)
		prog := generateProg(t, key)
		if prog == nil {
			t.Errorf("Offending key: [%v]", hex.EncodeToString(key))
			return
		}
	}
}

func compareState(t *testing.T, pcState []uint32, pcState2 []uint32) bool {
	if len(pcState) != len(pcState2) {
		t.Errorf("different length")
		return true
	}
	for i := 0; i < len(pcState); i++ {
		if pcState2[i] != pcState[i] {
			t.Errorf("different [%v]", i)
			return true
		}
	}
	return false
}

func TestExecute(t *testing.T) {
	key := make([]byte, 32)
	pcState := make([]byte, 2048)
	pcState2 := make([]byte, 2048)
	pcItem := make([]byte, 1024)

	err := false
	for x := 0; x < 1000; x++ {
		prog := generateProg(t, key)
		if prog == nil {
			err = true
			break
		}

		pcutil.HashExpand(pcItem, key, 0)
		pcutil.HashExpand(pcState, key, 1)
		copy(pcState2, pcState)

		if bytes.Compare(pcState, pcState2) != 0 {
			err = true
			break
		}
		interpret.Interpret(prog, pcState, pcItem, 4)
		nativetest.Interpret(prog, pcState2, pcItem, 4)
		if bytes.Compare(pcState, pcState2) != 0 {
			err = true
			break
		}
		rand.Read(key)
	}
	if err {
		t.Errorf("Offending key: [%v]", hex.EncodeToString(key))
		return
	}
}

func TestOps(t *testing.T) {
	inout := [8]uint32{}
	inout2 := [8]uint32{}
	b := [4 * 4]byte{}

	for i := 0; i < 100; i++ {
		for i := 0; i < 8; i++ {
			inout[i] = 0
			inout2[i] = 0
		}
		rand.Read(b[:])
		pcutil.U32FromB(inout[:4], b[:])
		for i := 0; i < 4; i++ {
			inout2[i] = inout[i]
		}
		op := opcodes.OpCode(inout[3] % opcodes.OpCode__MAX)

		if compareState(t, inout[:], inout2[:]) {
			break
		}
		interpret.TestOp(inout[:], op)
		nativetest.RandHashDoOp(inout2[:], op)
		if compareState(t, inout[:], inout2[:]) {
			fmt.Printf("%08x %08x %08x %08x %08x %08x %08x %08x\n",
				inout[0], inout[1], inout[2], inout[3], inout[4], inout[5], inout[6], inout[7])
			fmt.Printf("%08x %08x %08x %08x %08x %08x %08x %08x\n",
				inout2[0], inout2[1], inout2[2], inout2[3], inout2[4], inout2[5], inout2[6], inout2[7])
			break
		}
	}
}
