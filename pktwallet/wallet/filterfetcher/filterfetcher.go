package filterfetcher

import (
	"sync"
	"time"

	"github.com/pkt-cash/pktd/btcutil/er"
	"github.com/pkt-cash/pktd/pktwallet/chain"
	"github.com/pkt-cash/pktd/pktwallet/wallet/watcher"
	"github.com/pkt-cash/pktd/pktwallet/wtxmgr"
)

const DefaultWorkerCount = 6
const DefaultBacklog = 10

func loadTransactions(
	chainClient chain.Interface,
	height int32,
	watch *watcher.Watcher,
) (*chain.FilterBlocksResponse, er.R) {
	if hash, err := chainClient.GetBlockHash(int64(height)); err != nil {
		return nil, err
	} else if header, err := chainClient.GetBlockHeader(hash); err != nil {
		return nil, err
	} else {
		filterReq := watch.FilterReq(height)
		filterReq.Blocks = []wtxmgr.BlockMeta{
			{
				Block: wtxmgr.Block{
					Hash:   header.BlockHash(),
					Height: height,
				},
				Time: header.Timestamp,
			},
		}
		return chainClient.FilterBlocks(filterReq)
	}
}

type filterFetcherThread struct {
	ff          *FilterFetcher
	chainClient chain.Interface
	watch       *watcher.Watcher
}

type FilterFetcher struct {
	lock        sync.Mutex
	nextHeight  int32
	stop        bool
	filterCache map[int32]*chain.FilterBlocksResponse

	threads []filterFetcherThread
	caller  filterFetcherThread
}

type doNext uint8

const nextCont doNext = 0
const nextSleep doNext = 1
const nextQuit doNext = 2

const maxEntries = 20

type Task struct {
	dn         doNext
	nextHeight int32
}

func work(fft filterFetcherThread, t Task) {
	res, err := loadTransactions(fft.chainClient, t.nextHeight, fft.watch)
	if err != nil {
		// do nothing, let it get requested again by the main thread
		return
	}
	fft.ff.lock.Lock()
	defer fft.ff.lock.Unlock()
	fft.ff.filterCache[t.nextHeight] = res
}

func task(fft filterFetcherThread) Task {
	fft.ff.lock.Lock()
	defer fft.ff.lock.Unlock()
	if fft.ff.stop {
		return Task{dn: nextQuit}
	}
	if len(fft.ff.filterCache) >= maxEntries {
		return Task{dn: nextSleep}
	}
	height := fft.ff.nextHeight
	fft.ff.nextHeight++
	return Task{
		dn:         nextCont,
		nextHeight: height,
	}
}

func threadLoop(fft filterFetcherThread) {
	for {
		t := task(fft)
		if t.dn == nextSleep {
			time.Sleep(time.Millisecond * 500)
			continue
		} else if t.dn == nextQuit {
			return
		}
		work(fft, t)
	}
}

func (ff *FilterFetcher) Fetch(
	height int32,
) (*chain.FilterBlocksResponse, er.R) {
	var res *chain.FilterBlocksResponse
	ff.lock.Lock()
	for h, f := range ff.filterCache {
		if h <= height {
			delete(ff.filterCache, h)
		}
		if h == height {
			res = f
		}
	}
	ff.lock.Unlock()
	if res != nil {
		return res, nil
	}
	return loadTransactions(ff.caller.ff.caller.chainClient, height, ff.caller.watch)
}

func New(workerCount,
	maxBacklog int,
	height int32,
	chainClient chain.Interface,
	watch *watcher.Watcher,
) *FilterFetcher {
	out := FilterFetcher{
		nextHeight:  height,
		filterCache: make(map[int32]*chain.FilterBlocksResponse),
		threads:     make([]filterFetcherThread, workerCount),
		caller:      filterFetcherThread{},
	}
	out.caller.ff = &out
	out.caller.chainClient = chainClient
	out.caller.watch = watch
	for i := 0; i < workerCount; i++ {
		out.threads[i].ff = &out
		out.threads[i].chainClient = chainClient
		out.threads[i].watch = watch
	}
	return &out
}
