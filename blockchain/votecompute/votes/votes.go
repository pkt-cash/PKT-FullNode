package votes

import (
	"bytes"

	"github.com/pkt-cash/PKT-FullNode/blockchain"
	"github.com/pkt-cash/PKT-FullNode/btcutil"
	"github.com/pkt-cash/PKT-FullNode/btcutil/er"
	"github.com/pkt-cash/PKT-FullNode/pktlog/log"
	"github.com/pkt-cash/PKT-FullNode/txscript/opcode"
	"github.com/pkt-cash/PKT-FullNode/txscript/parsescript"
)

const (
	VOTE      byte = 0x00
	CANDIDATE byte = 0x01
)

type NsVote struct {
	VoterPkScript           []byte
	VoterIsWillingCandidate bool
	VoteCastInBlock         uint32
	VoteForPkScript         []byte
}

func getVote(outputScript []byte) *NsVote {
	scr, err := parsescript.ParseScript(outputScript)
	if err != nil {
		return nil
	}
	if len(scr) < 1 || scr[0].Opcode.Value != opcode.OP_RETURN {
		// Normal script, does not begin with OP_RETURN
		return nil
	}
	if len(scr) < 2 || scr[1].Opcode.Value > opcode.OP_16 {
		// It's an op-return script which contains something other than a push
		return nil
	}
	if len(scr) > 2 {
		// it's an op-return script but it contains additional data after the push
		return nil
	}
	data := scr[1].Data
	if len(data) < 1 || (data[0] != VOTE && data[0] != CANDIDATE) {
		// Not a vote operation
		return nil
	}
	return &NsVote{
		VoterIsWillingCandidate: data[0] == CANDIDATE,
		VoteForPkScript:         data[1:],
	}
}

func parseVotes(block *btcutil.Block, stxo []blockchain.SpentTxOut) ([]NsVote, er.R) {
	stxoIdx := 0
	var blockVotes []NsVote
txns:
	for _, tx := range block.Transactions()[1:] {
		inputs := stxo[stxoIdx : stxoIdx+len(tx.MsgTx().TxIn)]
		stxoIdx += len(tx.MsgTx().TxIn)

		var vote *NsVote
		for _, out := range tx.MsgTx().TxOut {
			if out.Value != 0 {
				continue
			}
			v := getVote(out.PkScript)
			if v != nil {
				if vote != nil {
					log.Infof("Ignoring votes in transaction [%s@%d], a transaction can only have one vote",
						tx.Hash(), block.Height())
					continue txns
				}
				vote = v
			}
		}
		if vote == nil {
			// No votes
			continue txns
		}

		// There is no explicit mapping between the stxos and the block transactions, but they
		// are in the same order so we can walk through the stxos as we walk though the inputs
		if len(inputs) != len(tx.MsgTx().TxIn) {
			return nil, er.Errorf("Mismatch in number of spent txouts for txn [%s@%d] "+
				"expect [%d] got [%d]", tx.Hash(), block.Height(), len(tx.MsgTx().TxIn), len(inputs))
		}
		var addr []byte
		for _, inp := range inputs {
			if addr == nil {
				addr = inp.PkScript
			} else if !bytes.Equal(addr, inp.PkScript) {
				log.Infof("Ignoring vote in transaction [%s@%d], only one input address is allowed",
					tx.Hash(), block.Height())
				continue txns
			}
		}

		vote.VoteCastInBlock = uint32(block.Height())
		vote.VoterPkScript = addr

		blockVotes = append(blockVotes, *vote)
	}

	if stxoIdx != len(stxo) {
		return nil, er.Errorf("Transactions in block [%d] have a total of [%d] SpentTxOut but [%d] txins "+
			"this should not happen.", block.Height(), len(stxo), stxoIdx)
	}

	return blockVotes, nil
}
