package tlv

import (
	"encoding/binary"
	"io"

	"github.com/pkt-cash/pktd/btcutil/er"
	"github.com/pkt-cash/pktd/btcutil/util"
	"github.com/pkt-cash/pktd/wire"
)

var Err = er.NewErrorType("lnd.tlv")

// ErrVarIntNotCanonical signals that the decoded varint was not minimally encoded.
var ErrVarIntNotCanonical = Err.CodeWithDetail("ErrVarIntNotCanonical", "decoded varint is not canonical")

// ReadVarInt reads a variable length integer from r and returns it as a uint64.
func ReadVarInt(r io.Reader, buf *[8]byte) (uint64, er.R) {
	_, err := util.ReadFull(r, buf[:1])
	if err != nil {
		return 0, err
	}
	discriminant := buf[0]

	var rv uint64
	switch {
	case discriminant < 0xfd:
		rv = uint64(discriminant)

	case discriminant == 0xfd:
		_, err := util.ReadFull(r, buf[:2])
		switch {
		case er.Wrapped(err) == io.EOF:
			return 0, er.E(io.ErrUnexpectedEOF)
		case err != nil:
			return 0, err
		}
		rv = uint64(binary.BigEndian.Uint16(buf[:2]))

		// The encoding is not canonical if the value could have been
		// encoded using fewer bytes.
		if rv < 0xfd {
			return 0, ErrVarIntNotCanonical.Default()
		}

	case discriminant == 0xfe:
		_, err := util.ReadFull(r, buf[:4])
		switch {
		case er.Wrapped(err) == io.EOF:
			return 0, er.E(io.ErrUnexpectedEOF)
		case err != nil:
			return 0, err
		}
		rv = uint64(binary.BigEndian.Uint32(buf[:4]))

		// The encoding is not canonical if the value could have been
		// encoded using fewer bytes.
		if rv <= 0xffff {
			return 0, ErrVarIntNotCanonical.Default()
		}

	default:
		_, err := util.ReadFull(r, buf[:])
		switch {
		case er.EOF.Is(err):
			return 0, er.ErrUnexpectedEOF.Default()
		case err != nil:
			return 0, err
		}
		rv = binary.BigEndian.Uint64(buf[:])

		// The encoding is not canonical if the value could have been
		// encoded using fewer bytes.
		if rv <= 0xffffffff {
			return 0, ErrVarIntNotCanonical.Default()
		}
	}

	return rv, nil
}

// WriteVarInt serializes val to w using a variable number of bytes depending
// on its value.
func WriteVarInt(w io.Writer, val uint64, buf *[8]byte) er.R {
	var length int
	switch {
	case val < 0xfd:
		buf[0] = uint8(val)
		length = 1

	case val <= 0xffff:
		buf[0] = uint8(0xfd)
		binary.BigEndian.PutUint16(buf[1:3], uint16(val))
		length = 3

	case val <= 0xffffffff:
		buf[0] = uint8(0xfe)
		binary.BigEndian.PutUint32(buf[1:5], uint32(val))
		length = 5

	default:
		buf[0] = uint8(0xff)
		_, err := util.Write(w, buf[:1])
		if err != nil {
			return err
		}

		binary.BigEndian.PutUint64(buf[:], uint64(val))
		length = 8
	}

	_, err := util.Write(w, buf[:length])
	return err
}

// VarIntSize returns the required number of bytes to encode a var int.
func VarIntSize(val uint64) uint64 {
	return uint64(wire.VarIntSerializeSize(val))
}
