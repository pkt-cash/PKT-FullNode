package watcher

import (
	"sort"
	"sync"

	"github.com/pkt-cash/pktd/btcutil"
	"github.com/pkt-cash/pktd/pktwallet/chain"
	"github.com/pkt-cash/pktd/pktwallet/waddrmgr"
	"github.com/pkt-cash/pktd/wire"
)

type OutPointWatch struct {
	BeginHeight int32
	OutPoint    wire.OutPoint
	Addr        btcutil.Address
}

type Watcher struct {
	watchAddrsLock sync.RWMutex
	watchAddrs     map[btcutil.Address]struct{}
	watchPoints    []OutPointWatch
}

func New() Watcher {
	return Watcher{
		watchAddrs:  make(map[btcutil.Address]struct{}),
		watchPoints: make([]OutPointWatch, 0),
	}
}

func (w *Watcher) watchStuff(
	addrs []btcutil.Address,
	ao []OutPointWatch,
) {
	w.watchAddrsLock.Lock()
	defer w.watchAddrsLock.Unlock()

	for _, addr := range addrs {
		w.watchAddrs[addr] = struct{}{}
	}
	if len(ao) > 0 {
		w.watchPoints = append(w.watchPoints, ao...)
		sort.Slice(w.watchPoints, func(i, j int) bool {
			return w.watchPoints[i].BeginHeight < w.watchPoints[j].BeginHeight
		})
	}
}

func (w *Watcher) WatchOutpoints(ao []OutPointWatch) {
	w.watchStuff(nil, ao)
}

func (w *Watcher) WatchAddrs(addrs []btcutil.Address) {
	w.watchStuff(addrs, nil)
}

func (w *Watcher) WatchAddr(addr btcutil.Address) {
	w.watchStuff([]btcutil.Address{addr}, nil)
}

func (w *Watcher) FilterReq(height int32) *chain.FilterBlocksRequest {
	w.watchAddrsLock.RLock()
	defer w.watchAddrsLock.RUnlock()
	filterReq := chain.FilterBlocksRequest{
		Blocks:           nil,
		ExternalAddrs:    make(map[waddrmgr.ScopedIndex]btcutil.Address),
		InternalAddrs:    make(map[waddrmgr.ScopedIndex]btcutil.Address),
		ImportedAddrs:    make([]btcutil.Address, 0, len(w.watchAddrs)),
		WatchedOutPoints: make(map[wire.OutPoint]btcutil.Address),
	}
	for wa := range w.watchAddrs {
		filterReq.ImportedAddrs = append(filterReq.ImportedAddrs, wa)
	}
	for _, opw := range w.watchPoints {
		if opw.BeginHeight > height {
			break
		}
		filterReq.WatchedOutPoints[opw.OutPoint] = opw.Addr
	}
	return &filterReq
}
