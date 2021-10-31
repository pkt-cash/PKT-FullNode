package enough

import (
	"github.com/pkt-cash/pktd/btcutil"
	"github.com/pkt-cash/pktd/pktwallet/wallet/internal/txsizes"
	"github.com/pkt-cash/pktd/pktwallet/wallet/txrules"
	"github.com/pkt-cash/pktd/wire"
)

type IsEnough struct {
	sweeping      bool
	baseSize      int
	sizeOneSegwit int
	sizeOneLegacy int
	feePerKb      btcutil.Amount
	needed        btcutil.Amount
}

func MkIsEnough(txOutputs []*wire.TxOut, feePerKb btcutil.Amount) IsEnough {
	sweepOutput := GetSweepOutput(txOutputs)
	needed := btcutil.Amount(0)
	for _, o := range txOutputs {
		needed += btcutil.Amount(o.Value)
	}
	// We're going to start with the base tx size for fee estimation
	// Then we'll take the size with 1 input in order to be able to estimate fee per input
	baseSize := txsizes.EstimateVirtualSize(0, 0, 0, txOutputs, true)
	return IsEnough{
		sweeping:      sweepOutput != nil,
		sizeOneSegwit: txsizes.EstimateVirtualSize(0, 1, 0, txOutputs, true) - baseSize,
		sizeOneLegacy: txsizes.EstimateVirtualSize(1, 0, 0, txOutputs, true) - baseSize,
		baseSize:      baseSize,
		feePerKb:      feePerKb,
		needed:        needed,
	}
}
func (ii *IsEnough) IsSweeping() bool {
	return ii.sweeping
}
func (ii *IsEnough) WellIsIt(inputCount int, segwit bool, amt btcutil.Amount) bool {
	if ii.sweeping {
	} else if amt < ii.needed {
	} else {
		perInput := ii.sizeOneSegwit
		if !segwit {
			perInput = ii.sizeOneLegacy
		}
		size := ii.baseSize + perInput*inputCount
		fee := txrules.FeeForSerializeSize(ii.feePerKb, size)
		return amt > ii.needed+fee
	}
	return false
}

func GetSweepOutput(outs []*wire.TxOut) *wire.TxOut {
	var sweepOutput *wire.TxOut
	for _, out := range outs {
		if out.Value == 0 {
			sweepOutput = out
		}
	}
	return sweepOutput
}
