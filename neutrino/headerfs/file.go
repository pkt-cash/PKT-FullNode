package headerfs

import (
	"github.com/pkt-cash/pktd/btcutil/er"
	"github.com/pkt-cash/pktd/pktwallet/walletdb"

	"github.com/pkt-cash/pktd/chaincfg/chainhash"
	"github.com/pkt-cash/pktd/wire"
)

// ErrHeaderNotFound is returned when a target header on disk (flat file) can't
// be found.
var ErrHeaderNotFound = er.GenericErrorType.Code("headerfs.ErrHeaderNotFound")

// readHeaderRange will attempt to fetch a series of block headers within the
// target height range.
//
// NOTE: The end height is _inclusive_ so we'll fetch all headers from the
// startHeight up to the end height, including the final header.
func (h *blockHeaderStore) readHeaderRange(
	tx walletdb.ReadTx,
	startHeight uint32,
	endHeight uint32,
) ([]wire.BlockHeader, er.R) {

	out := make([]wire.BlockHeader, 0, endHeight-startHeight+1)
	for i := startHeight; i <= endHeight; i++ {
		if he, err := h.readHeader(tx, i); err != nil {
			return nil, err
		} else if hdr, err := blockHeaderFromHe(he); err != nil {
			return nil, err
		} else {
			out = append(out, *hdr.BlockHeader)
		}
	}
	return out, nil
}

// readHeaderRange will attempt to fetch a series of filter headers within the
// target height range.
//
// NOTE: The end height is _inclusive_ so we'll fetch all headers from the
// startHeight up to the end height, including the final header.
func (f *FilterHeaderStore) readHeaderRange(
	tx walletdb.ReadTx,
	startHeight uint32,
	endHeight uint32,
) ([]chainhash.Hash, er.R) {
	out := make([]chainhash.Hash, 0, endHeight-startHeight+1)
	for i := startHeight; i <= endHeight; i++ {
		if ret, err := f.readHeader(tx, i); err != nil {
			return nil, err
		} else if hash, err := chainhash.NewHash(ret.bytes); err != nil {
			return nil, err
		} else {
			out = append(out, *hash)
		}
	}
	return out, nil
}
