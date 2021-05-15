package migration_01_to_11

import (
	"fmt"
	"io"
	"net"

	"github.com/pkt-cash/pktd/btcec"
	"github.com/pkt-cash/pktd/btcutil"
	"github.com/pkt-cash/pktd/btcutil/er"
	"github.com/pkt-cash/pktd/btcutil/util"
	"github.com/pkt-cash/pktd/chaincfg/chainhash"
	"github.com/pkt-cash/pktd/lnd/keychain"
	"github.com/pkt-cash/pktd/lnd/lnwire"
	"github.com/pkt-cash/pktd/lnd/shachain"
	"github.com/pkt-cash/pktd/wire"
)

// writeOutpoint writes an outpoint to the passed writer using the minimal
// amount of bytes possible.
func writeOutpoint(w io.Writer, o *wire.OutPoint) er.R {
	if _, err := util.Write(w, o.Hash[:]); err != nil {
		return err
	}
	if err := util.WriteBin(w, byteOrder, o.Index); err != nil {
		return err
	}

	return nil
}

// readOutpoint reads an outpoint from the passed reader that was previously
// written using the writeOutpoint struct.
func readOutpoint(r io.Reader, o *wire.OutPoint) er.R {
	if _, err := util.ReadFull(r, o.Hash[:]); err != nil {
		return err
	}
	if err := util.ReadBin(r, byteOrder, &o.Index); err != nil {
		return err
	}

	return nil
}

// UnknownElementType is an error returned when the codec is unable to encode or
// decode a particular type.
type UnknownElementType struct {
	method  string
	element interface{}
}

// Error returns the name of the method that encountered the error, as well as
// the type that was unsupported.
func (e UnknownElementType) Error() string {
	return fmt.Sprintf("Unknown type in %s: %T", e.method, e.element)
}

// WriteElement is a one-stop shop to write the big endian representation of
// any element which is to be serialized for storage on disk. The passed
// io.Writer should be backed by an appropriately sized byte slice, or be able
// to dynamically expand to accommodate additional data.
func WriteElement(w io.Writer, element interface{}) er.R {
	switch e := element.(type) {
	case keychain.KeyDescriptor:
		if err := util.WriteBin(w, byteOrder, e.Family); err != nil {
			return err
		}
		if err := util.WriteBin(w, byteOrder, e.Index); err != nil {
			return err
		}

		if e.PubKey != nil {
			if err := util.WriteBin(w, byteOrder, true); err != nil {
				return er.Errorf("error writing serialized element: %s", err)
			}

			return WriteElement(w, e.PubKey)
		}

		return util.WriteBin(w, byteOrder, false)
	case ChannelType:
		if err := util.WriteBin(w, byteOrder, e); err != nil {
			return err
		}

	case chainhash.Hash:
		if _, err := util.Write(w, e[:]); err != nil {
			return err
		}

	case wire.OutPoint:
		return writeOutpoint(w, &e)

	case lnwire.ShortChannelID:
		if err := util.WriteBin(w, byteOrder, e.ToUint64()); err != nil {
			return err
		}

	case lnwire.ChannelID:
		if _, err := util.Write(w, e[:]); err != nil {
			return err
		}

	case int64, uint64:
		if err := util.WriteBin(w, byteOrder, e); err != nil {
			return err
		}

	case uint32:
		if err := util.WriteBin(w, byteOrder, e); err != nil {
			return err
		}

	case int32:
		if err := util.WriteBin(w, byteOrder, e); err != nil {
			return err
		}

	case uint16:
		if err := util.WriteBin(w, byteOrder, e); err != nil {
			return err
		}

	case uint8:
		if err := util.WriteBin(w, byteOrder, e); err != nil {
			return err
		}

	case bool:
		if err := util.WriteBin(w, byteOrder, e); err != nil {
			return err
		}

	case btcutil.Amount:
		if err := util.WriteBin(w, byteOrder, uint64(e)); err != nil {
			return err
		}

	case lnwire.MilliSatoshi:
		if err := util.WriteBin(w, byteOrder, uint64(e)); err != nil {
			return err
		}

	case *btcec.PrivateKey:
		b := e.Serialize()
		if _, err := util.Write(w, b); err != nil {
			return err
		}

	case *btcec.PublicKey:
		b := e.SerializeCompressed()
		if _, err := util.Write(w, b); err != nil {
			return err
		}

	case shachain.Producer:
		return e.Encode(w)

	case shachain.Store:
		return e.Encode(w)

	case *wire.MsgTx:
		return e.Serialize(w)

	case [32]byte:
		if _, err := util.Write(w, e[:]); err != nil {
			return err
		}

	case []byte:
		if err := wire.WriteVarBytes(w, 0, e); err != nil {
			return err
		}

	case lnwire.Message:
		if _, err := lnwire.WriteMessage(w, e, 0); err != nil {
			return err
		}

	case ChannelStatus:
		if err := util.WriteBin(w, byteOrder, e); err != nil {
			return err
		}

	case ClosureType:
		if err := util.WriteBin(w, byteOrder, e); err != nil {
			return err
		}

	case lnwire.FundingFlag:
		if err := util.WriteBin(w, byteOrder, e); err != nil {
			return err
		}

	case net.Addr:
		if err := serializeAddr(w, e); err != nil {
			return err
		}

	case []net.Addr:
		if err := WriteElement(w, uint32(len(e))); err != nil {
			return err
		}

		for _, addr := range e {
			if err := serializeAddr(w, addr); err != nil {
				return err
			}
		}

	default:
		return er.E(UnknownElementType{"WriteElement", e})
	}

	return nil
}

// WriteElements is writes each element in the elements slice to the passed
// io.Writer using WriteElement.
func WriteElements(w io.Writer, elements ...interface{}) er.R {
	for _, element := range elements {
		err := WriteElement(w, element)
		if err != nil {
			return err
		}
	}
	return nil
}

// ReadElement is a one-stop utility function to deserialize any datastructure
// encoded using the serialization format of the database.
func ReadElement(r io.Reader, element interface{}) er.R {
	switch e := element.(type) {
	case *keychain.KeyDescriptor:
		if err := util.ReadBin(r, byteOrder, &e.Family); err != nil {
			return err
		}
		if err := util.ReadBin(r, byteOrder, &e.Index); err != nil {
			return err
		}

		var hasPubKey bool
		if err := util.ReadBin(r, byteOrder, &hasPubKey); err != nil {
			return err
		}

		if hasPubKey {
			return ReadElement(r, &e.PubKey)
		}

	case *ChannelType:
		if err := util.ReadBin(r, byteOrder, e); err != nil {
			return err
		}

	case *chainhash.Hash:
		if _, err := util.ReadFull(r, e[:]); err != nil {
			return err
		}

	case *wire.OutPoint:
		return readOutpoint(r, e)

	case *lnwire.ShortChannelID:
		var a uint64
		if err := util.ReadBin(r, byteOrder, &a); err != nil {
			return err
		}
		*e = lnwire.NewShortChanIDFromInt(a)

	case *lnwire.ChannelID:
		if _, err := util.ReadFull(r, e[:]); err != nil {
			return err
		}

	case *int64, *uint64:
		if err := util.ReadBin(r, byteOrder, e); err != nil {
			return err
		}

	case *uint32:
		if err := util.ReadBin(r, byteOrder, e); err != nil {
			return err
		}

	case *int32:
		if err := util.ReadBin(r, byteOrder, e); err != nil {
			return err
		}

	case *uint16:
		if err := util.ReadBin(r, byteOrder, e); err != nil {
			return err
		}

	case *uint8:
		if err := util.ReadBin(r, byteOrder, e); err != nil {
			return err
		}

	case *bool:
		if err := util.ReadBin(r, byteOrder, e); err != nil {
			return err
		}

	case *btcutil.Amount:
		var a uint64
		if err := util.ReadBin(r, byteOrder, &a); err != nil {
			return err
		}

		*e = btcutil.Amount(a)

	case *lnwire.MilliSatoshi:
		var a uint64
		if err := util.ReadBin(r, byteOrder, &a); err != nil {
			return err
		}

		*e = lnwire.MilliSatoshi(a)

	case **btcec.PrivateKey:
		var b [btcec.PrivKeyBytesLen]byte
		if _, err := util.ReadFull(r, b[:]); err != nil {
			return err
		}

		priv, _ := btcec.PrivKeyFromBytes(btcec.S256(), b[:])
		*e = priv

	case **btcec.PublicKey:
		var b [btcec.PubKeyBytesLenCompressed]byte
		if _, err := util.ReadFull(r, b[:]); err != nil {
			return err
		}

		pubKey, err := btcec.ParsePubKey(b[:], btcec.S256())
		if err != nil {
			return err
		}
		*e = pubKey

	case *shachain.Producer:
		var root [32]byte
		if _, err := util.ReadFull(r, root[:]); err != nil {
			return err
		}

		// TODO(roasbeef): remove
		producer, err := shachain.NewRevocationProducerFromBytes(root[:])
		if err != nil {
			return err
		}

		*e = producer

	case *shachain.Store:
		store, err := shachain.NewRevocationStoreFromBytes(r)
		if err != nil {
			return err
		}

		*e = store

	case **wire.MsgTx:
		tx := wire.NewMsgTx(2)
		if err := tx.Deserialize(r); err != nil {
			return err
		}

		*e = tx

	case *[32]byte:
		if _, err := util.ReadFull(r, e[:]); err != nil {
			return err
		}

	case *[]byte:
		bytes, err := wire.ReadVarBytes(r, 0, 66000, "[]byte")
		if err != nil {
			return err
		}

		*e = bytes

	case *lnwire.Message:
		msg, err := lnwire.ReadMessage(r, 0)
		if err != nil {
			return err
		}

		*e = msg

	case *ChannelStatus:
		if err := util.ReadBin(r, byteOrder, e); err != nil {
			return err
		}

	case *ClosureType:
		if err := util.ReadBin(r, byteOrder, e); err != nil {
			return err
		}

	case *lnwire.FundingFlag:
		if err := util.ReadBin(r, byteOrder, e); err != nil {
			return err
		}

	case *net.Addr:
		addr, err := deserializeAddr(r)
		if err != nil {
			return err
		}
		*e = addr

	case *[]net.Addr:
		var numAddrs uint32
		if err := ReadElement(r, &numAddrs); err != nil {
			return err
		}

		*e = make([]net.Addr, numAddrs)
		for i := uint32(0); i < numAddrs; i++ {
			addr, err := deserializeAddr(r)
			if err != nil {
				return err
			}
			(*e)[i] = addr
		}

	default:
		return er.E(UnknownElementType{"ReadElement", e})
	}

	return nil
}

// ReadElements deserializes a variable number of elements into the passed
// io.Reader, with each element being deserialized according to the ReadElement
// function.
func ReadElements(r io.Reader, elements ...interface{}) er.R {
	for _, element := range elements {
		err := ReadElement(r, element)
		if err != nil {
			return err
		}
	}
	return nil
}
