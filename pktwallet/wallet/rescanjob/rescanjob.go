package rescanjob

import (
	"sync"

	"github.com/pkt-cash/pktd/pktwallet/wallet/filterfetcher"
	"github.com/pkt-cash/pktd/pktwallet/wallet/watcher"
)

type RescanJob struct {
	Watch  *watcher.Watcher
	Height int32
	Quick  bool
	Name   string
	FF     *filterfetcher.FilterFetcher
	ctx    *RescanJobs
}

func (j *RescanJob) Delete() {
	j.ctx.lock.Lock()
	defer j.ctx.lock.Unlock()

	if j == j.ctx.activeJob {
		j.ctx.activeJob = nil
	}
	for i := range j.ctx.jobs {
		if j.ctx.jobs[i] == j {
			copy(j.ctx.jobs[i:], j.ctx.jobs[i+1:])
			j.ctx.jobs = j.ctx.jobs[:len(j.ctx.jobs)-1]
		}
	}
}

type RescanJobs struct {
	activeJob *RescanJob
	jobs      []*RescanJob
	lock      sync.Mutex
}

func New() RescanJobs {
	return RescanJobs{
		jobs: make([]*RescanJob, 0),
	}
}

func (r *RescanJobs) EnqueueJob(job *RescanJob) {
	r.lock.Lock()
	defer r.lock.Unlock()
	job.ctx = r
	r.jobs = append(r.jobs, job)
}

func (r *RescanJobs) GetJob() *RescanJob {
	r.lock.Lock()
	defer r.lock.Unlock()

	if len(r.jobs) == 0 {
	} else if r.activeJob == nil {
		r.activeJob = r.jobs[0]
	}
	return r.activeJob
}
