// Copyright (c) 2014-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package mempool

import (
	"fmt"

	"github.com/pkt-cash/pktd/btcutil/er"
	"github.com/pkt-cash/pktd/wire"
	"github.com/pkt-cash/pktd/wire/ruleerror"
)

// txRuleError creates an underlying TxRuleError with the given a set of
// arguments and returns a RuleError that encapsulates it.
func txRuleError(c wire.RejectCode, desc string) er.R {
	code := ruleerror.Err.NumberToCode(int(c))
	if code == nil {
		return ruleerror.ErrUnknown.New(fmt.Sprintf("code: %d reason: %s", c, desc), nil)
	}
	return code.New(desc, nil)
}
