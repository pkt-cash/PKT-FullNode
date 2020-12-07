package record

import "github.com/pkt-cash/pktd/btcutil/er"

const (
	// CustomTypeStart is the start of the custom tlv type range as defined
	// in BOLT 01.
	CustomTypeStart = 65536
)

// CustomSet stores a set of custom key/value pairs.
type CustomSet map[uint64][]byte

// Validate checks that all custom records are in the custom type range.
func (c CustomSet) Validate() er.R {
	for key := range c {
		if key < CustomTypeStart {
			return er.Errorf("no custom records with types "+
				"below %v allowed", CustomTypeStart)
		}
	}

	return nil
}
