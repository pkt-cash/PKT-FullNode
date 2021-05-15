// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package connmgr

import (
	"encoding/binary"
	"net"

	"github.com/pkt-cash/pktd/btcutil/er"
)

const (
	torSucceeded         = 0x00
	torGeneralError      = 0x01
	torNotAllowed        = 0x02
	torNetUnreachable    = 0x03
	torHostUnreachable   = 0x04
	torConnectionRefused = 0x05
	torTTLExpired        = 0x06
	torCmdNotSupported   = 0x07
	torAddrNotSupported  = 0x08
)

var (
	Err = er.NewErrorType("connmgr.Err")

	// ErrTorInvalidAddressResponse indicates an invalid address was
	// returned by the Tor DNS resolver.
	ErrTorInvalidAddressResponse = Err.CodeWithDetail("ErrTorInvalidAddressResponse",
		"invalid address response")

	// ErrTorInvalidProxyResponse indicates the Tor proxy returned a
	// response in an unexpected format.
	ErrTorInvalidProxyResponse = Err.CodeWithDetail("ErrTorInvalidProxyResponse",
		"invalid proxy response")

	// ErrTorUnrecognizedAuthMethod indicates the authentication method
	// provided is not recognized.
	ErrTorUnrecognizedAuthMethod = Err.CodeWithDetail("ErrTorUnrecognizedAuthMethod",
		"invalid proxy authentication method")

	torStatusErrors = map[byte]*er.ErrorCode{
		torSucceeded:         Err.CodeWithDetail("torSucceeded", "tor succeeded"),
		torGeneralError:      Err.CodeWithDetail("torGeneralError", "tor general error"),
		torNotAllowed:        Err.CodeWithDetail("torNotAllowed", "tor not allowed"),
		torNetUnreachable:    Err.CodeWithDetail("torNetUnreachable", "tor network is unreachable"),
		torHostUnreachable:   Err.CodeWithDetail("torHostUnreachable", "tor host is unreachable"),
		torConnectionRefused: Err.CodeWithDetail("torConnectionRefused", "tor connection refused"),
		torTTLExpired:        Err.CodeWithDetail("torTTLExpired", "tor TTL expired"),
		torCmdNotSupported:   Err.CodeWithDetail("torCmdNotSupported", "tor command not supported"),
		torAddrNotSupported:  Err.CodeWithDetail("torAddrNotSupported", "tor address type not supported"),
	}
)

// TorLookupIP uses Tor to resolve DNS via the SOCKS extension they provide for
// resolution over the Tor network. Tor itself doesn't support ipv6 so this
// doesn't either.
func TorLookupIP(host, proxy string) ([]net.IP, er.R) {
	conn, errr := net.Dial("tcp", proxy)
	if errr != nil {
		return nil, er.E(errr)
	}
	defer conn.Close()

	buf := []byte{'\x05', '\x01', '\x00'}
	_, errr = conn.Write(buf)
	if errr != nil {
		return nil, er.E(errr)
	}

	buf = make([]byte, 2)
	_, errr = conn.Read(buf)
	if errr != nil {
		return nil, er.E(errr)
	}
	if buf[0] != '\x05' {
		return nil, ErrTorInvalidProxyResponse.Default()
	}
	if buf[1] != '\x00' {
		return nil, ErrTorUnrecognizedAuthMethod.Default()
	}

	buf = make([]byte, 7+len(host))
	buf[0] = 5      // protocol version
	buf[1] = '\xF0' // Tor Resolve
	buf[2] = 0      // reserved
	buf[3] = 3      // Tor Resolve
	buf[4] = byte(len(host))
	copy(buf[5:], host)
	buf[5+len(host)] = 0 // Port 0

	_, errr = conn.Write(buf)
	if errr != nil {
		return nil, er.E(errr)
	}

	buf = make([]byte, 4)
	_, errr = conn.Read(buf)
	if errr != nil {
		return nil, er.E(errr)
	}
	if buf[0] != 5 {
		return nil, ErrTorInvalidProxyResponse.Default()
	}
	if buf[1] != 0 {
		if int(buf[1]) >= len(torStatusErrors) {
			return nil, ErrTorInvalidProxyResponse.Default()
		} else if erc := torStatusErrors[buf[1]]; erc != nil {
			return nil, erc.Default()
		}
		return nil, ErrTorInvalidProxyResponse.Default()
	}
	if buf[3] != 1 {
		erc := torStatusErrors[torGeneralError]
		return nil, erc.Default()
	}

	buf = make([]byte, 4)
	bytes, errr := conn.Read(buf)
	if errr != nil {
		return nil, er.E(errr)
	}
	if bytes != 4 {
		return nil, ErrTorInvalidAddressResponse.Default()
	}

	r := binary.BigEndian.Uint32(buf)

	addr := make([]net.IP, 1)
	addr[0] = net.IPv4(byte(r>>24), byte(r>>16), byte(r>>8), byte(r))

	return addr, nil
}
