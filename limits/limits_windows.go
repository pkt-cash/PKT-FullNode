// Copyright (c) 2013-2014 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package limits

import "github.com/pkt-cash/pktd/btcutil/er"

// SetLimits is a no-op on Windows since it's not required there.
func SetLimits() er.R {
	return nil
}
