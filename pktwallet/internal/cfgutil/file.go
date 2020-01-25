// Copyright (c) 2015 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package cfgutil

import (
	"os"

	"github.com/pkt-cash/pktd/btcutil/er"
)

// FileExists reports whether the named file or directory exists.
func FileExists(filePath string) (bool, er.R) {
	_, errr := os.Stat(filePath)
	if errr != nil {
		if os.IsNotExist(errr) {
			return false, nil
		}
		return false, er.E(errr)
	}
	return true, nil
}
