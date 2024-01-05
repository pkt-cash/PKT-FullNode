// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2015-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package externaladdrs

import (
	"net"
	"sync"

	"github.com/pkt-cash/PKT-FullNode/addrmgr/addrutil"
	"github.com/pkt-cash/PKT-FullNode/btcutil/er"
	"github.com/pkt-cash/PKT-FullNode/pktlog/log"
	"github.com/pkt-cash/PKT-FullNode/wire"
	"github.com/pkt-cash/PKT-FullNode/wire/protocol"
)

// AddressPriority type is used to describe the hierarchy of local address
// discovery methods.
type AddressPriority int

const (
	// InterfacePrio signifies the address is on a local interface
	InterfacePrio AddressPriority = iota

	// BoundPrio signifies the address has been explicitly bounded to.
	BoundPrio

	// UpnpPrio signifies the address was obtained from UPnP.
	UpnpPrio

	// HTTPPrio signifies the address was obtained from an external HTTP service.
	//lint:ignore U1001 removing this will cause ManualPrio to change number
	HTTPPrio

	// ManualPrio signifies the address was provided by --externalip.
	ManualPrio
)

type localAddress struct {
	na    *wire.NetAddress
	score AddressPriority
}

type ExternalLocalAddrs struct {
	localAddresses map[string]*localAddress
	lamtx          sync.Mutex
}

// AddLocalAddress adds na to the list of known local external addresses
// to advertise with the given priority.
func (a *ExternalLocalAddrs) Add(na *wire.NetAddress, priority AddressPriority) er.R {
	if !addrutil.IsRoutable(na) {
		return er.Errorf("address %s is not routable", na.IP)
	}

	a.lamtx.Lock()
	defer a.lamtx.Unlock()

	key := addrutil.NetAddressKey(na)
	if a.localAddresses == nil {
		a.localAddresses = make(map[string]*localAddress)
	}
	if la, ok := a.localAddresses[key]; ok {
		if la.score < priority {
			la.score = priority + 1
		}
	} else {
		a.localAddresses[key] = &localAddress{
			na:    na,
			score: priority,
		}
	}
	return nil
}

// GetBestLocalAddress returns the most appropriate local address to use
// for the given remote address.
func (a *ExternalLocalAddrs) GetBest(remoteAddr *wire.NetAddress) *wire.NetAddress {
	a.lamtx.Lock()
	defer a.lamtx.Unlock()

	if a.localAddresses == nil {
		log.Debugf("No local addresses defined")
		return nil
	}

	var bestscore AddressPriority
	var bestAddress *wire.NetAddress
	for _, la := range a.localAddresses {
		if !addrutil.Reachable(la.na, remoteAddr) {
			continue
		} else if bestAddress != nil && la.score < bestscore {
			continue
		} else {
			bestscore = la.score
			bestAddress = la.na
		}
	}
	if bestAddress != nil {
		log.Debugf("Suggesting address %s:%d for %s:%d", bestAddress.IP,
			bestAddress.Port, remoteAddr.IP, remoteAddr.Port)
	} else {
		log.Debugf("No worthy address for %s:%d", remoteAddr.IP,
			remoteAddr.Port)

		// Send something unroutable if nothing suitable.
		var ip net.IP
		if !addrutil.IsIPv4(remoteAddr) {
			ip = net.IPv6zero
		} else {
			ip = net.IPv4zero
		}
		services := protocol.SFNodeNetwork | protocol.SFNodeWitness | protocol.SFNodeBloom
		bestAddress = wire.NewNetAddressIPPort(ip, 0, services)
	}

	return bestAddress
}
