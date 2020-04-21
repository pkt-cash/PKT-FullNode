package parsescript

import (
	"fmt"

	"github.com/pkt-cash/pktd/btcutil/er"
	"github.com/pkt-cash/pktd/txscript/opcode"
	"github.com/pkt-cash/pktd/txscript/txscripterr"
)

// opcodeArray holds details about all possible opcodes such as how many bytes
// the opcode and any associated data should take, its human-readable name, and
// the handler function.
var opcodeArray [256]opcode.Opcode

func init() {
	for i := 0; i < 256; i++ {
		opcodeArray[i] = opcode.MkOpcode(byte(i))
	}
}

// ParsedOpcode represents an opcode that has been parsed and includes any
// potential data associated with it.
type ParsedOpcode struct {
	Opcode opcode.Opcode
	Data   []byte
}

// ParseScriptTemplate is the same as ParseScript but allows the passing of the
// template list for testing purposes.  When there are parse errors, it returns
// the list of parsed opcodes up to the point of failure along with the error.
func ParseScriptTemplate(script []byte, opcodes map[byte]opcode.Opcode) ([]ParsedOpcode, er.R) {
	retScript := make([]ParsedOpcode, 0, len(script))
	for i := 0; i < len(script); {
		instr := script[i]
		op := opcodeArray[instr]
		if opcodes == nil {
		} else if opp, ok := opcodes[instr]; ok {
			op = opp
		}
		pop := ParsedOpcode{Opcode: op}

		// Parse data out of instruction.
		switch {
		// No additional data.  Note that some of the opcodes, notably
		// OP_1NEGATE, OP_0, and OP_[1-16] represent the data
		// themselves.
		case op.Length == 1:
			i++

		// Data pushes of specific lengths -- OP_DATA_[1-75].
		case op.Length > 1:
			if len(script[i:]) < op.Length {
				str := fmt.Sprintf("opcode %s requires %d "+
					"bytes, but script only has %d remaining",
					opcode.OpcodeName(op.Value), op.Length, len(script[i:]))
				return retScript, txscripterr.ScriptError(txscripterr.ErrMalformedPush,
					str)
			}

			// Slice out the data.
			pop.Data = script[i+1 : i+op.Length]
			i += op.Length

		// Data pushes with parsed lengths -- OP_PUSHDATAP{1,2,4}.
		case op.Length < 0:
			var l uint
			off := i + 1

			if len(script[off:]) < -op.Length {
				str := fmt.Sprintf("opcode %s requires %d "+
					"bytes, but script only has %d remaining",
					opcode.OpcodeName(op.Value), -op.Length, len(script[off:]))
				return retScript, txscripterr.ScriptError(txscripterr.ErrMalformedPush,
					str)
			}

			// Next -length bytes are little endian length of data.
			switch op.Length {
			case -1:
				l = uint(script[off])
			case -2:
				l = ((uint(script[off+1]) << 8) |
					uint(script[off]))
			case -4:
				l = ((uint(script[off+3]) << 24) |
					(uint(script[off+2]) << 16) |
					(uint(script[off+1]) << 8) |
					uint(script[off]))
			default:
				str := fmt.Sprintf("invalid opcode length %d",
					op.Length)
				return retScript, txscripterr.ScriptError(txscripterr.ErrMalformedPush,
					str)
			}

			// Move offset to beginning of the data.
			off += -op.Length

			// Disallow entries that do not fit script or were
			// sign extended.
			if int(l) > len(script[off:]) || int(l) < 0 {
				str := fmt.Sprintf("opcode %s pushes %d bytes, "+
					"but script only has %d remaining",
					opcode.OpcodeName(op.Value), int(l), len(script[off:]))
				return retScript, txscripterr.ScriptError(txscripterr.ErrMalformedPush,
					str)
			}

			pop.Data = script[off : off+int(l)]
			i += 1 - op.Length + int(l)
		}

		retScript = append(retScript, pop)
	}

	return retScript, nil
}

// parseScript preparses the script in bytes into a list of parsedOpcodes while
// applying a number of sanity checks.
func ParseScript(script []byte) ([]ParsedOpcode, er.R) {
	return ParseScriptTemplate(script, nil)
}

// IsPushOnly returns true if the script only pushes data, false otherwise.
func IsPushOnly(pops []ParsedOpcode) bool {
	// NOTE: This function does NOT verify opcodes directly since it is
	// internal and is only called with parsed opcodes for scripts that did
	// not have any parse errors.  Thus, consensus is properly maintained.

	for _, pop := range pops {
		// All opcodes up to OP_16 are data push instructions.
		// NOTE: This does consider OP_RESERVED to be a data push
		// instruction, but execution of OP_RESERVED will fail anyways
		// and matches the behavior required by consensus.
		if pop.Opcode.Value > opcode.OP_16 {
			return false
		}
	}
	return true
}
