package workqueue

import (
	"sync"
	"time"

	"github.com/pkt-cash/pktd/btcutil/er"
)

const DefaultWorkerCount = 6
const DefaultBacklog = 10

type threadCtx struct {
	ff  *WorkQueue
	job func(key uint64) er.R
}

type WorkQueue struct {
	lock        sync.Mutex
	nextNum     uint64
	maxNum      uint64
	resultCache map[uint64]er.R

	threads    []threadCtx
	callerLock sync.Mutex
}

type doNext uint8

const nextCont doNext = 0
const nextSleep doNext = 1
const nextQuit doNext = 2

const maxEntries = 20

type workTask struct {
	dn      doNext
	nextNum uint64
}

func work(tctx threadCtx, t workTask) {
	err := tctx.job(t.nextNum)
	tctx.ff.lock.Lock()
	defer tctx.ff.lock.Unlock()
	tctx.ff.resultCache[t.nextNum] = err
}

func task(tctx threadCtx) workTask {
	tctx.ff.lock.Lock()
	defer tctx.ff.lock.Unlock()
	if len(tctx.ff.resultCache) >= maxEntries {
		return workTask{dn: nextSleep}
	}
	num := tctx.ff.nextNum
	if num >= tctx.ff.maxNum {
		return workTask{dn: nextQuit}
	}
	tctx.ff.nextNum++
	return workTask{
		dn:      nextCont,
		nextNum: num,
	}
}

func threadLoop(tctx threadCtx) {
	for {
		t := task(tctx)
		if t.dn == nextSleep {
			time.Sleep(time.Millisecond * 500)
			continue
		} else if t.dn == nextQuit {
			return
		}
		work(tctx, t)
	}
}

func (ff *WorkQueue) Get(
	num uint64,
) er.R {
	var res er.R
	for {
		found := false
		ff.lock.Lock()
		if r, ok := ff.resultCache[num]; ok {
			delete(ff.resultCache, num)
			res = r
			found = true
		}
		ff.lock.Unlock()
		if found {
			return res
		}
		time.Sleep(time.Millisecond)
	}
}

func New(workerCount,
	maxResults int,
	rangeMin uint64,
	rangeMax uint64,
	job func(key uint64) er.R,
) *WorkQueue {
	out := WorkQueue{
		resultCache: make(map[uint64]er.R),
		threads:     make([]threadCtx, workerCount),
		nextNum:     rangeMin,
		maxNum:      rangeMax,
	}
	for i := 0; i < workerCount; i++ {
		out.threads[i].ff = &out
		out.threads[i].job = job
		go threadLoop(out.threads[i])
	}
	return &out
}
