// Copyright (c) 2013-2014 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

// +build !windows,!plan9

package limits

import (
	"syscall"

	"github.com/pkt-cash/pktd/btcutil/er"
)

const (
	fileLimitWant = 2048
	fileLimitMin  = 1024
)

// SetLimits raises some process limits to values which allow pktd and
// associated utilities to run.
func SetLimits() er.R {
	var rLimit syscall.Rlimit

	err := er.E(syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit))
	if err != nil {
		return err
	}
	if rLimit.Cur > fileLimitWant {
		return nil
	}
	if rLimit.Max < fileLimitMin {
		return er.Errorf("need at least %v file descriptors", fileLimitMin)
	}
	if rLimit.Max < fileLimitWant {
		rLimit.Cur = rLimit.Max
	} else {
		rLimit.Cur = fileLimitWant
	}
	err = er.E(syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit))
	if err != nil {
		// try min value
		rLimit.Cur = fileLimitMin
		err = er.E(syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit))
		if err != nil {
			return err
		}
	}

	return nil
}
