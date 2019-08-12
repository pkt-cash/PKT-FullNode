// Copyright (c) 2019 Caleb James DeLisle
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/dchest/blake2b"
	"github.com/pkt-cash/pktd/blockchain/packetcrypt/pcutil"
)

// PcCoinbaseCommitMagic is the first 4 bytes of the commitment
const PcCoinbaseCommitMagic uint32 = 0x0211f909

// PcCoinbaseCommit is the commmitment which is placed in the coinbase
// before beginning to calculate the PacketCrypt proof
type PcCoinbaseCommit struct {
	Bytes [48]byte
}

// NewPcCoinbaseCommit creates a new coinbase commit with the initial pattern (all fc)
func NewPcCoinbaseCommit() *PcCoinbaseCommit {
	out := PcCoinbaseCommit{}
	binary.LittleEndian.PutUint32(out.Bytes[:4], PcCoinbaseCommitMagic)
	for i := 4; i < 48; i++ {
		out.Bytes[i] = 0xfc
	}
	return &out
}

// AnnCount gets the number of announcements which were claimed in the coinbase commitment
func (c *PcCoinbaseCommit) AnnCount() uint64 {
	return binary.LittleEndian.Uint64(c.Bytes[40:])
}

// MerkleRoot gets the root of announcements claimed in the coinbase commitment
func (c *PcCoinbaseCommit) MerkleRoot() []byte {
	return c.Bytes[8:40]
}

// AnnMinDifficulty gets the claimed minimum target of any announcement claimed in the coinbase
// commitment, the format is in bitcoin compressed bignum form
func (c *PcCoinbaseCommit) AnnMinDifficulty() uint32 {
	return binary.LittleEndian.Uint32(c.Bytes[4:8])
}

// Magic gets the magic bytes from the coinbase commitment, they should match PcCoinbaseCommitMagic
func (c *PcCoinbaseCommit) Magic() uint32 {
	return binary.LittleEndian.Uint32(c.Bytes[:4])
}

const endType = 0
const pcpType = 1
const signaturesType = 2
const contentProofsType = 3

type PacketCryptEntity struct {
	Type    uint32
	Content []byte
}

// PacketCryptProof is the in-memory representation of the proof which sits between the header and
// the block content
type PacketCryptProof struct {
	Nonce           uint32
	Announcements   [4]PacketCryptAnn
	Signatures      [4][]byte
	ContentProof    []byte
	AnnProof        []byte
	UnknownEntities []PacketCryptEntity
}

// ContentProofIndex gets a uint32 based on the 4 announcements
// this is used to get the block of the content which will be proven in each of the
// content proofs for each of the announcements
func (h *PacketCryptProof) ContentProofIndex() uint32 {
	b2 := blake2b.New256()
	writeElement(b2, h.Nonce)
	for _, ann := range h.Announcements {
		b2.Write(ann.Header[:])
	}
	b2.Write(h.AnnProof)
	sum := b2.Sum(nil)
	return binary.LittleEndian.Uint32(sum[:4])
}

// SplitContentProof splits the content proof into the proofs for the
// 4 individual announcements.
func (h *PacketCryptProof) SplitContentProof(proofIdx uint32) ([][]byte, error) {
	if h.ContentProof == nil {
		return make([][]byte, 4), nil
	}
	cpb := bytes.NewBuffer(h.ContentProof)
	out := make([][]byte, 4)
	for i, ann := range h.Announcements {
		contentLength := ann.GetContentLength()
		if contentLength <= 32 {
			continue
		}
		totalBlocks := contentLength / 32
		if totalBlocks*32 < contentLength {
			totalBlocks++
		}
		blockToProve := proofIdx % totalBlocks
		depth := pcutil.Log2ceil(uint64(totalBlocks))
		length := 32
		blockSize := uint32(32)
		for i := 0; i < depth; i++ {
			if blockSize*(blockToProve^1) >= contentLength {
				blockToProve >>= 1
				blockSize <<= 1
				continue
			}
			length += 32
			blockToProve >>= 1
			blockSize <<= 1
		}
		b := make([]byte, length)
		if _, err := io.ReadFull(cpb, b[:]); err != nil {
			return nil, fmt.Errorf("SplitContentProof: unable to read ann content proof [%s]", err)
		}
		out[i] = b
	}
	return out, nil
}

// BtcDecode decodes a PacketCryptProof from a reader
func (h *PacketCryptProof) BtcDecode(r io.Reader, pver uint32, enc MessageEncoding) error {
	return readPacketCryptProof(r, pver, enc, h)
}

// BtcEncode encodes a PacketCryptProof to a writer
func (h *PacketCryptProof) BtcEncode(w io.Writer, pver uint32, enc MessageEncoding) error {
	return writePacketCryptProof(w, pver, enc, h)
}

// Deserialize reads a PacketCryptProof from the on-disk format
func (h *PacketCryptProof) Deserialize(r io.Reader) error {
	return readPacketCryptProof(r, 0, WitnessEncoding, h)
}

// Serialize writes a PacketCryptProof to the on-disk format
func (h *PacketCryptProof) Serialize(w io.Writer) error {
	return writePacketCryptProof(w, 0, WitnessEncoding, h)
}

// PcProofFromBytes is a helper function which reads a byte array into a PacketCryptProof
func PcProofFromBytes(b []byte) *PacketCryptProof {
	bb := bytes.NewBuffer(b)
	out := PacketCryptProof{}
	if err := out.Deserialize(bb); err != nil {
		panic("failed to decode pcp")
	}
	return &out
}

// SerializeSize gets the size of the PacketCryptProof when serialized
func (h *PacketCryptProof) SerializeSize() int {
	out := 4 + PcAnnSerializeSize*4
	{
		pcplen := 1024*4 + 4 + len(h.AnnProof)
		out += VarIntSerializeSize(pcpType)
		out += VarIntSerializeSize(uint64(pcplen))
		out += pcplen
	}
	{
		slen := 0
		for i := 0; i < 4; i++ {
			slen += len(h.Signatures[i])
		}
		if slen > 0 {
			out += VarIntSerializeSize(signaturesType)
			out += VarIntSerializeSize(uint64(slen))
			out += slen
		}
	}
	{
		clen := len(h.ContentProof)
		if clen > 0 {
			out += VarIntSerializeSize(contentProofsType)
			out += VarIntSerializeSize(uint64(clen))
			out += clen
		}
	}
	{
		for i := 0; i < len(h.UnknownEntities); i++ {
			out += VarIntSerializeSize(uint64(h.UnknownEntities[i].Type))
			out += VarIntSerializeSize(uint64(len(h.UnknownEntities[i].Content)))
			out += len(h.UnknownEntities[i].Content)
		}
	}
	out += VarIntSerializeSize(endType)
	out += VarIntSerializeSize(0)

	return out
}

func readPacketCryptProof(r io.Reader, pver uint32, enc MessageEncoding, pcp *PacketCryptProof) error {
	hasPcp := false
	for {
		t, err := ReadVarInt(r, 0)
		if err != nil {
			return err
		}
		length, err := ReadVarInt(r, 0)
		if err != nil {
			return err
		}
		switch t {
		case endType:
			{
				if !hasPcp {
					return messageError("readPacketCryptProof", "Missing PacketCrypt proof")
				}
				return nil
			}
		case pcpType:
			{
				readElement(r, &pcp.Nonce)
				for i := 0; i < 4; i++ {
					if err := pcp.Announcements[i].BtcDecode(r, pver, enc); err != nil {
						return err
					}
				}
				pcp.AnnProof = make([]byte, length-(1024*4)-4)
				if _, err := io.ReadFull(r, pcp.AnnProof); err != nil {
					return err
				}
				hasPcp = true
			}
		case signaturesType:
			{
				if !hasPcp {
					return messageError("readPacketCryptProof", "Signatures came before pcp type")
				}
				remainingBytes := int(length)
				for i := 0; i < 4; i++ {
					if !pcp.Announcements[i].HasSigningKey() {
						continue
					}
					pcp.Signatures[i] = make([]byte, 64)
					if _, err := io.ReadFull(r, pcp.Signatures[i]); err != nil {
						return err
					}
					remainingBytes -= 64
					if remainingBytes < 0 {
						return messageError("readPacketCryptProof",
							"Not enough remaining bytes in read announcement signature")
					}
				}
				if remainingBytes != 0 {
					return messageError("readPacketCryptProof",
						"Dangling bytes after announcement signatures")
				}
			}
		case contentProofsType:
			{
				if !hasPcp {
					return messageError("readPacketCryptProof", "ContentProofs came before pcp type")
				}
				pcp.ContentProof = make([]byte, length)
				if _, err := io.ReadFull(r, pcp.ContentProof); err != nil {
					return err
				}
			}
		default:
			{
				e := PacketCryptEntity{
					Type:    uint32(t),
					Content: make([]byte, length),
				}
				if _, err := io.ReadFull(r, e.Content); err != nil {
					return err
				}
				pcp.UnknownEntities = append(pcp.UnknownEntities, e)
			}
		}
	}
}

func writePacketCryptProof(w io.Writer, pver uint32, enc MessageEncoding, pcp *PacketCryptProof) error {

	if err := WriteVarInt(w, 0, pcpType); err != nil {
		return err
	}
	if err := WriteVarInt(w, 0, uint64(4+(1024*4)+len(pcp.AnnProof))); err != nil {
		return err
	}
	if err := writeElement(w, pcp.Nonce); err != nil {
		return err
	}
	for i := 0; i < 4; i++ {
		if err := pcp.Announcements[i].BtcEncode(w, pver, enc); err != nil {
			return err
		}
	}
	if _, err := w.Write(pcp.AnnProof); err != nil {
		return err
	}

	{
		sigLen := 0
		for _, sig := range pcp.Signatures {
			sigLen += len(sig)
		}
		if sigLen > 0 {
			if err := WriteVarInt(w, 0, signaturesType); err != nil {
				return err
			}
			if err := WriteVarInt(w, 0, uint64(sigLen)); err != nil {
				return err
			}
			for _, sig := range pcp.Signatures {
				if sig == nil {
				} else if _, err := w.Write(sig); err != nil {
					return err
				}
			}
		}
	}

	{
		if pcp.ContentProof != nil {
			if err := WriteVarInt(w, 0, contentProofsType); err != nil {
				return err
			}
			if err := WriteVarInt(w, 0, uint64(len(pcp.ContentProof))); err != nil {
				return err
			}
			if _, err := w.Write(pcp.ContentProof); err != nil {
				return err
			}
		}
	}

	for _, e := range pcp.UnknownEntities {
		if err := WriteVarInt(w, 0, uint64(e.Type)); err != nil {
			return err
		}
		if err := WriteVarInt(w, 0, uint64(len(e.Content))); err != nil {
			return err
		}
		if _, err := w.Write(e.Content); err != nil {
			return err
		}
	}

	WriteVarInt(w, 0, endType)
	WriteVarInt(w, 0, 0)

	return nil
}
