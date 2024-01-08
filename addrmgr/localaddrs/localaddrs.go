package localaddrs

import (
	"net"
	"strings"
	"sync"

	"github.com/pkt-cash/PKT-FullNode/addrmgr/addrutil"
	"github.com/pkt-cash/PKT-FullNode/pktlog/log"
	"github.com/pkt-cash/PKT-FullNode/wire"
)

type LocalAddrs struct {
	m        sync.Mutex
	a        map[string]*wire.NetAddress
	wasTried bool
}

func New() LocalAddrs {
	return LocalAddrs{
		a: make(map[string]*wire.NetAddress),
	}
}

func (la *LocalAddrs) Referesh() {
	ifaces, errr := net.Interfaces()
	if errr != nil {
		log.Warnf("LocalAddrs.Referesh() failed: [%v]", errr.Error())
		la.m.Lock()
		defer la.m.Unlock()
		la.wasTried = true
		return
	}
	out := make(map[string]struct{})
	for _, i := range ifaces {
		addrs, errr := i.Addrs()
		if errr != nil {
			log.Warnf("LocalAddrs.Referesh(): [%s]", errr.Error())
			continue
		}
		for _, a := range addrs {
			out[a.String()] = struct{}{}
		}
	}
	la.m.Lock()
	la.wasTried = true
	for s := range la.a {
		if _, ok := out[s]; !ok {
			log.Infof("Local address gone [%s]", log.IpAddr(s))
			delete(la.a, s)
		}
	}
	for s := range out {
		if _, ok := la.a[s]; !ok {
			// drop the port
			spl := strings.Split(s, "/")
			ip := net.ParseIP(spl[0])
			if ip == nil {
				log.Warnf("LocalAddrs.Referesh(): unable to parse addr [%s]", s)
			} else {
				wip := wire.NewNetAddressIPPort(ip, 0, 0)
				if (addrutil.IsIPv4(wip) && !addrutil.IsLocal(wip)) || addrutil.IsRoutable(wip) {
					log.Infof("Local address detected [%s]", log.IpAddr(s))
					la.a[s] = wip
				} else {
					log.Debugf("Non-routable local address detected [%s]", s)
					la.a[s] = nil
				}
			}
		}
	}
	la.m.Unlock()
}

func (la *LocalAddrs) IsWorking() bool {
	la.m.Lock()
	defer la.m.Unlock()
	return len(la.a) > 0 || !la.wasTried
}

func (la *LocalAddrs) Reachable(na *wire.NetAddress) bool {
	out := false
	la.m.Lock()
	for _, localNa := range la.a {
		if localNa != nil && addrutil.Reachable(localNa, na) {
			log.Tracef("[%s] reachable via [%s]", na.IP.String(), localNa.IP.String())
			out = true
			break
		}
	}
	la.m.Unlock()
	return out
}
