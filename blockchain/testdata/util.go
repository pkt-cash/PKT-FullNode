// Copyright (c) 2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

// The vast majority of the rules tested in this package were ported from the
// the original Java-based 'official' block acceptance tests at
// https://github.com/TheBlueMatt/test-scripts as well as some additional tests
// available in the Core python port of the same.

package testdata

import (
	"bytes"
	"compress/bzip2"
	"encoding/binary"
	"fmt"
	"github.com/pkt-cash/pktd/btcutil/er"
	"io"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/pkt-cash/pktd/btcutil"
	"github.com/pkt-cash/pktd/wire"
)

// LoadBlocks reads files containing bitcoin block data (gzipped but otherwise
// in the format bitcoind writes) from disk and returns them as an array of
// btcutil.Block.  This is largely borrowed from the test code in pktdb.
func LoadBlocks(filename string) (blocks []*btcutil.Block, err er.R) {
	var network = wire.MainNet
	var dr io.Reader
	var fi io.ReadCloser

	fi, err = os.Open(filename)
	if err != nil {
		return
	}

	if strings.HasSuffix(filename, ".bz2") {
		dr = bzip2.NewReader(fi)
	} else {
		dr = fi
	}
	defer fi.Close()

	var block *btcutil.Block

	err = nil
	for height := int64(1); err == nil; height++ {
		var rintbuf uint32
		err = binary.Read(dr, binary.LittleEndian, &rintbuf)
		if err == io.EOF {
			// hit end of file at expected offset: no warning
			height--
			err = nil
			break
		}
		if err != nil {
			break
		}
		if rintbuf != uint32(network) {
			break
		}
		err = binary.Read(dr, binary.LittleEndian, &rintbuf)
		blocklen := rintbuf

		rbytes := make([]byte, blocklen)

		// read block
		dr.Read(rbytes)

		block, err = btcutil.NewBlockFromBytes(rbytes)
		if err != nil {
			return
		}
		blocks = append(blocks, block)
	}

	return
}

func GetBlock(name string, t *testing.T) *btcutil.Block {
	runtime.GOMAXPROCS(runtime.NumCPU())

	blocks, err := LoadBlocks(name)
	if err != nil {
		t.Errorf("Error loading file: %v\n", err)
		return nil
	}
	if len(blocks) > 1 {
		t.Errorf("The test block file must only have one block in it")
		return nil
	}
	if len(blocks) == 0 {
		t.Errorf("The test block file seems empty")
		return nil
	}
	return blocks[0]
}

func OutputBlock(mb *wire.MsgBlock, name string) er.R {
	buf := bytes.NewBuffer(make([]byte, 0, mb.SerializeSize()))
	err := mb.Serialize(buf)
	b := buf.Bytes()
	if err != nil {
		return err
	}
	fmt.Printf("Creating file %v\n", name)
	fi, err := os.Create(name)
	if err != nil {
		return err
	}
	defer fi.Close()
	var header [8]byte
	binary.LittleEndian.PutUint32(header[:4], uint32(wire.MainNet))
	binary.LittleEndian.PutUint32(header[4:], uint32(len(b)))
	fi.Write(header[:])
	fi.Write(b)
	return nil
}

// LoadAnnouncements reads a file containing a sequence of announcements
// 1024 byte irreducible header followed by content
func LoadAnnouncements(filename string) ([]*wire.PacketCryptAnn, er.R) {
	var fi io.ReadCloser

	fi, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fi.Close()

	anns := make([]*wire.PacketCryptAnn, 0)
	for {
		ann := wire.PacketCryptAnn{}
		if err = ann.BtcDecode(fi, 0, 0); err != nil {
			if err == io.EOF {
				return anns, nil
			}
			return nil, err
		}
		anns = append(anns, &ann)
	}
}

func OutputAnnouncements(anns [][]byte) er.R {
	fi, err := os.Create("./outputanns.dat")
	if err != nil {
		return err
	}
	defer fi.Close()
	for _, annb := range anns {
		fi.Write(annb)
	}
	return nil
}
