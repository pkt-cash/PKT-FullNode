package lnwire

import (
	"math/big"
	"testing"

	"github.com/pkt-cash/pktd/btcec"
	"github.com/pkt-cash/pktd/btcutil/er"
)

func TestSignatureSerializeDeserialize(t *testing.T) {
	t.Parallel()

	// Local-scoped closure to serialize and deserialize a Signature and
	// check for errors as well as check if the results are correct.
	signatureSerializeDeserialize := func(e btcec.Signature) er.R {
		sig, err := NewSigFromSignature(&e)
		if err != nil {
			return err
		}

		e2, err := sig.ToSignature()
		if err != nil {
			return err
		}

		if e.R.Cmp(e2.R) != 0 {
			return er.Errorf("Pre/post-serialize Rs don't match"+
				": %s, %s", e.R, e2.R)
		}
		if e.S.Cmp(e2.S) != 0 {
			return er.Errorf("Pre/post-serialize Ss don't match"+
				": %s, %s", e.S, e2.S)
		}
		return nil
	}

	sig := btcec.Signature{}

	// Check R = N-1, S = 128.
	sig.R = big.NewInt(1) // Allocate a big.Int before we call .Sub.
	sig.R.Sub(btcec.S256().N, sig.R)
	sig.S = big.NewInt(128)
	err := signatureSerializeDeserialize(sig)
	if err != nil {
		t.Fatalf("R = N-1, S = 128: %s", err.String())
	}

	// Check R = N-1, S = 127.
	sig.S = big.NewInt(127)
	err = signatureSerializeDeserialize(sig)
	if err != nil {
		t.Fatalf("R = N-1, S = 127: %s", err.String())
	}

	// Check R = N-1, S = N>>1.
	sig.S.Set(btcec.S256().N)
	sig.S.Rsh(sig.S, 1)
	err = signatureSerializeDeserialize(sig)
	if err != nil {
		t.Fatalf("R = N-1, S = N>>1: %s", err.String())
	}

	// Check R = N-1, S = N.
	sig.S.Set(btcec.S256().N)
	err = signatureSerializeDeserialize(sig)
	if err.Message() != "signature S isn't 1 or more" {
		t.Fatalf("R = N-1, S = N should become R = N-1, S = 0: %s",
			err.String())
	}

	// Check R = N-1, S = N-1.
	sig.S.Sub(sig.S, big.NewInt(1))
	err = signatureSerializeDeserialize(sig)
	if err.Message() != "Pre/post-serialize Ss don't match: 115792089237316"+
		"195423570985008687907852837564279074904382605163141518161494"+
		"336, 1" {
		t.Fatalf("R = N-1, S = N-1 should become R = N-1, S = 1: %s",
			err.String())
	}

	// Check R = 2N, S = 128
	sig.R.Mul(btcec.S256().N, big.NewInt(2))
	sig.S.Set(big.NewInt(127))
	err = signatureSerializeDeserialize(sig)
	if err.Message() != "R is over 32 bytes long without padding" {
		t.Fatalf("R = 2N, S = 128, R should be over 32 bytes: %s",
			err.String())
	}
}
