// Copyright (c) 2013-2017 The btcsuite developers
// Copyright (c) 2019 Caleb James DeLisle
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package txscript

import (
	"bytes"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/binary"
	"fmt"

	"github.com/pkt-cash/pktd/btcutil"
	"github.com/pkt-cash/pktd/btcutil/er"
	"github.com/pkt-cash/pktd/txscript/opcode"
	"github.com/pkt-cash/pktd/txscript/scriptnum"
	"github.com/pkt-cash/pktd/txscript/txscripterr"

	"github.com/pkt-cash/pktd/btcec"
	"github.com/pkt-cash/pktd/chaincfg/chainhash"
	"github.com/pkt-cash/pktd/wire"
)

// An opcodeT defines the information related to a txscript opcodeT.  opfunc, if
// present, is the function to call to perform the opcodeT on the script.  The
// current script is passed in as a slice with the first member being the opcodeT
// itself.
type opcodeT struct {
	value  byte
	name   string
	length int
	opfunc func(*parsedOpcode, *Engine) er.R
}

// Conditional execution constants.
const (
	OpCondFalse = 0
	OpCondTrue  = 1
	OpCondSkip  = 2
)

// opcodeArray holds details about all possible opcodes such as how many bytes
// the opcode and any associated data should take, its human-readable name, and
// the handler function.
var opcodeArray = [256]opcodeT{
	// Data push opcodes.
	opcode.OP_FALSE:     {opcode.OP_FALSE, "OP_0", 1, opcodeFalse},
	opcode.OP_DATA_1:    {opcode.OP_DATA_1, "OP_DATA_1", 2, opcodePushData},
	opcode.OP_DATA_2:    {opcode.OP_DATA_2, "OP_DATA_2", 3, opcodePushData},
	opcode.OP_DATA_3:    {opcode.OP_DATA_3, "OP_DATA_3", 4, opcodePushData},
	opcode.OP_DATA_4:    {opcode.OP_DATA_4, "OP_DATA_4", 5, opcodePushData},
	opcode.OP_DATA_5:    {opcode.OP_DATA_5, "OP_DATA_5", 6, opcodePushData},
	opcode.OP_DATA_6:    {opcode.OP_DATA_6, "OP_DATA_6", 7, opcodePushData},
	opcode.OP_DATA_7:    {opcode.OP_DATA_7, "OP_DATA_7", 8, opcodePushData},
	opcode.OP_DATA_8:    {opcode.OP_DATA_8, "OP_DATA_8", 9, opcodePushData},
	opcode.OP_DATA_9:    {opcode.OP_DATA_9, "OP_DATA_9", 10, opcodePushData},
	opcode.OP_DATA_10:   {opcode.OP_DATA_10, "OP_DATA_10", 11, opcodePushData},
	opcode.OP_DATA_11:   {opcode.OP_DATA_11, "OP_DATA_11", 12, opcodePushData},
	opcode.OP_DATA_12:   {opcode.OP_DATA_12, "OP_DATA_12", 13, opcodePushData},
	opcode.OP_DATA_13:   {opcode.OP_DATA_13, "OP_DATA_13", 14, opcodePushData},
	opcode.OP_DATA_14:   {opcode.OP_DATA_14, "OP_DATA_14", 15, opcodePushData},
	opcode.OP_DATA_15:   {opcode.OP_DATA_15, "OP_DATA_15", 16, opcodePushData},
	opcode.OP_DATA_16:   {opcode.OP_DATA_16, "OP_DATA_16", 17, opcodePushData},
	opcode.OP_DATA_17:   {opcode.OP_DATA_17, "OP_DATA_17", 18, opcodePushData},
	opcode.OP_DATA_18:   {opcode.OP_DATA_18, "OP_DATA_18", 19, opcodePushData},
	opcode.OP_DATA_19:   {opcode.OP_DATA_19, "OP_DATA_19", 20, opcodePushData},
	opcode.OP_DATA_20:   {opcode.OP_DATA_20, "OP_DATA_20", 21, opcodePushData},
	opcode.OP_DATA_21:   {opcode.OP_DATA_21, "OP_DATA_21", 22, opcodePushData},
	opcode.OP_DATA_22:   {opcode.OP_DATA_22, "OP_DATA_22", 23, opcodePushData},
	opcode.OP_DATA_23:   {opcode.OP_DATA_23, "OP_DATA_23", 24, opcodePushData},
	opcode.OP_DATA_24:   {opcode.OP_DATA_24, "OP_DATA_24", 25, opcodePushData},
	opcode.OP_DATA_25:   {opcode.OP_DATA_25, "OP_DATA_25", 26, opcodePushData},
	opcode.OP_DATA_26:   {opcode.OP_DATA_26, "OP_DATA_26", 27, opcodePushData},
	opcode.OP_DATA_27:   {opcode.OP_DATA_27, "OP_DATA_27", 28, opcodePushData},
	opcode.OP_DATA_28:   {opcode.OP_DATA_28, "OP_DATA_28", 29, opcodePushData},
	opcode.OP_DATA_29:   {opcode.OP_DATA_29, "OP_DATA_29", 30, opcodePushData},
	opcode.OP_DATA_30:   {opcode.OP_DATA_30, "OP_DATA_30", 31, opcodePushData},
	opcode.OP_DATA_31:   {opcode.OP_DATA_31, "OP_DATA_31", 32, opcodePushData},
	opcode.OP_DATA_32:   {opcode.OP_DATA_32, "OP_DATA_32", 33, opcodePushData},
	opcode.OP_DATA_33:   {opcode.OP_DATA_33, "OP_DATA_33", 34, opcodePushData},
	opcode.OP_DATA_34:   {opcode.OP_DATA_34, "OP_DATA_34", 35, opcodePushData},
	opcode.OP_DATA_35:   {opcode.OP_DATA_35, "OP_DATA_35", 36, opcodePushData},
	opcode.OP_DATA_36:   {opcode.OP_DATA_36, "OP_DATA_36", 37, opcodePushData},
	opcode.OP_DATA_37:   {opcode.OP_DATA_37, "OP_DATA_37", 38, opcodePushData},
	opcode.OP_DATA_38:   {opcode.OP_DATA_38, "OP_DATA_38", 39, opcodePushData},
	opcode.OP_DATA_39:   {opcode.OP_DATA_39, "OP_DATA_39", 40, opcodePushData},
	opcode.OP_DATA_40:   {opcode.OP_DATA_40, "OP_DATA_40", 41, opcodePushData},
	opcode.OP_DATA_41:   {opcode.OP_DATA_41, "OP_DATA_41", 42, opcodePushData},
	opcode.OP_DATA_42:   {opcode.OP_DATA_42, "OP_DATA_42", 43, opcodePushData},
	opcode.OP_DATA_43:   {opcode.OP_DATA_43, "OP_DATA_43", 44, opcodePushData},
	opcode.OP_DATA_44:   {opcode.OP_DATA_44, "OP_DATA_44", 45, opcodePushData},
	opcode.OP_DATA_45:   {opcode.OP_DATA_45, "OP_DATA_45", 46, opcodePushData},
	opcode.OP_DATA_46:   {opcode.OP_DATA_46, "OP_DATA_46", 47, opcodePushData},
	opcode.OP_DATA_47:   {opcode.OP_DATA_47, "OP_DATA_47", 48, opcodePushData},
	opcode.OP_DATA_48:   {opcode.OP_DATA_48, "OP_DATA_48", 49, opcodePushData},
	opcode.OP_DATA_49:   {opcode.OP_DATA_49, "OP_DATA_49", 50, opcodePushData},
	opcode.OP_DATA_50:   {opcode.OP_DATA_50, "OP_DATA_50", 51, opcodePushData},
	opcode.OP_DATA_51:   {opcode.OP_DATA_51, "OP_DATA_51", 52, opcodePushData},
	opcode.OP_DATA_52:   {opcode.OP_DATA_52, "OP_DATA_52", 53, opcodePushData},
	opcode.OP_DATA_53:   {opcode.OP_DATA_53, "OP_DATA_53", 54, opcodePushData},
	opcode.OP_DATA_54:   {opcode.OP_DATA_54, "OP_DATA_54", 55, opcodePushData},
	opcode.OP_DATA_55:   {opcode.OP_DATA_55, "OP_DATA_55", 56, opcodePushData},
	opcode.OP_DATA_56:   {opcode.OP_DATA_56, "OP_DATA_56", 57, opcodePushData},
	opcode.OP_DATA_57:   {opcode.OP_DATA_57, "OP_DATA_57", 58, opcodePushData},
	opcode.OP_DATA_58:   {opcode.OP_DATA_58, "OP_DATA_58", 59, opcodePushData},
	opcode.OP_DATA_59:   {opcode.OP_DATA_59, "OP_DATA_59", 60, opcodePushData},
	opcode.OP_DATA_60:   {opcode.OP_DATA_60, "OP_DATA_60", 61, opcodePushData},
	opcode.OP_DATA_61:   {opcode.OP_DATA_61, "OP_DATA_61", 62, opcodePushData},
	opcode.OP_DATA_62:   {opcode.OP_DATA_62, "OP_DATA_62", 63, opcodePushData},
	opcode.OP_DATA_63:   {opcode.OP_DATA_63, "OP_DATA_63", 64, opcodePushData},
	opcode.OP_DATA_64:   {opcode.OP_DATA_64, "OP_DATA_64", 65, opcodePushData},
	opcode.OP_DATA_65:   {opcode.OP_DATA_65, "OP_DATA_65", 66, opcodePushData},
	opcode.OP_DATA_66:   {opcode.OP_DATA_66, "OP_DATA_66", 67, opcodePushData},
	opcode.OP_DATA_67:   {opcode.OP_DATA_67, "OP_DATA_67", 68, opcodePushData},
	opcode.OP_DATA_68:   {opcode.OP_DATA_68, "OP_DATA_68", 69, opcodePushData},
	opcode.OP_DATA_69:   {opcode.OP_DATA_69, "OP_DATA_69", 70, opcodePushData},
	opcode.OP_DATA_70:   {opcode.OP_DATA_70, "OP_DATA_70", 71, opcodePushData},
	opcode.OP_DATA_71:   {opcode.OP_DATA_71, "OP_DATA_71", 72, opcodePushData},
	opcode.OP_DATA_72:   {opcode.OP_DATA_72, "OP_DATA_72", 73, opcodePushData},
	opcode.OP_DATA_73:   {opcode.OP_DATA_73, "OP_DATA_73", 74, opcodePushData},
	opcode.OP_DATA_74:   {opcode.OP_DATA_74, "OP_DATA_74", 75, opcodePushData},
	opcode.OP_DATA_75:   {opcode.OP_DATA_75, "OP_DATA_75", 76, opcodePushData},
	opcode.OP_PUSHDATA1: {opcode.OP_PUSHDATA1, "OP_PUSHDATA1", -1, opcodePushData},
	opcode.OP_PUSHDATA2: {opcode.OP_PUSHDATA2, "OP_PUSHDATA2", -2, opcodePushData},
	opcode.OP_PUSHDATA4: {opcode.OP_PUSHDATA4, "OP_PUSHDATA4", -4, opcodePushData},
	opcode.OP_1NEGATE:   {opcode.OP_1NEGATE, "OP_1NEGATE", 1, opcode1Negate},
	opcode.OP_RESERVED:  {opcode.OP_RESERVED, "OP_RESERVED", 1, opcodeReserved},
	opcode.OP_TRUE:      {opcode.OP_TRUE, "OP_1", 1, opcodeN},
	opcode.OP_2:         {opcode.OP_2, "OP_2", 1, opcodeN},
	opcode.OP_3:         {opcode.OP_3, "OP_3", 1, opcodeN},
	opcode.OP_4:         {opcode.OP_4, "OP_4", 1, opcodeN},
	opcode.OP_5:         {opcode.OP_5, "OP_5", 1, opcodeN},
	opcode.OP_6:         {opcode.OP_6, "OP_6", 1, opcodeN},
	opcode.OP_7:         {opcode.OP_7, "OP_7", 1, opcodeN},
	opcode.OP_8:         {opcode.OP_8, "OP_8", 1, opcodeN},
	opcode.OP_9:         {opcode.OP_9, "OP_9", 1, opcodeN},
	opcode.OP_10:        {opcode.OP_10, "OP_10", 1, opcodeN},
	opcode.OP_11:        {opcode.OP_11, "OP_11", 1, opcodeN},
	opcode.OP_12:        {opcode.OP_12, "OP_12", 1, opcodeN},
	opcode.OP_13:        {opcode.OP_13, "OP_13", 1, opcodeN},
	opcode.OP_14:        {opcode.OP_14, "OP_14", 1, opcodeN},
	opcode.OP_15:        {opcode.OP_15, "OP_15", 1, opcodeN},
	opcode.OP_16:        {opcode.OP_16, "OP_16", 1, opcodeN},

	// Control opcodes.
	opcode.OP_NOP:                 {opcode.OP_NOP, "OP_NOP", 1, opcodeNop},
	opcode.OP_VER:                 {opcode.OP_VER, "OP_VER", 1, opcodeReserved},
	opcode.OP_IF:                  {opcode.OP_IF, "OP_IF", 1, opcodeIf},
	opcode.OP_NOTIF:               {opcode.OP_NOTIF, "OP_NOTIF", 1, opcodeNotIf},
	opcode.OP_VERIF:               {opcode.OP_VERIF, "OP_VERIF", 1, opcodeReserved},
	opcode.OP_VOTE:                {opcode.OP_VOTE, "OP_VOTE", 1, opcode2Drop},
	opcode.OP_ELSE:                {opcode.OP_ELSE, "OP_ELSE", 1, opcodeElse},
	opcode.OP_ENDIF:               {opcode.OP_ENDIF, "OP_ENDIF", 1, opcodeEndif},
	opcode.OP_VERIFY:              {opcode.OP_VERIFY, "OP_VERIFY", 1, opcodeVerify},
	opcode.OP_RETURN:              {opcode.OP_RETURN, "OP_RETURN", 1, opcodeReturn},
	opcode.OP_CHECKLOCKTIMEVERIFY: {opcode.OP_CHECKLOCKTIMEVERIFY, "OP_CHECKLOCKTIMEVERIFY", 1, opcodeCheckLockTimeVerify},
	opcode.OP_CHECKSEQUENCEVERIFY: {opcode.OP_CHECKSEQUENCEVERIFY, "OP_CHECKSEQUENCEVERIFY", 1, opcodeCheckSequenceVerify},

	// Stack opcodes.
	opcode.OP_TOALTSTACK:   {opcode.OP_TOALTSTACK, "OP_TOALTSTACK", 1, opcodeToAltStack},
	opcode.OP_FROMALTSTACK: {opcode.OP_FROMALTSTACK, "OP_FROMALTSTACK", 1, opcodeFromAltStack},
	opcode.OP_2DROP:        {opcode.OP_2DROP, "OP_2DROP", 1, opcode2Drop},
	opcode.OP_2DUP:         {opcode.OP_2DUP, "OP_2DUP", 1, opcode2Dup},
	opcode.OP_3DUP:         {opcode.OP_3DUP, "OP_3DUP", 1, opcode3Dup},
	opcode.OP_2OVER:        {opcode.OP_2OVER, "OP_2OVER", 1, opcode2Over},
	opcode.OP_2ROT:         {opcode.OP_2ROT, "OP_2ROT", 1, opcode2Rot},
	opcode.OP_2SWAP:        {opcode.OP_2SWAP, "OP_2SWAP", 1, opcode2Swap},
	opcode.OP_IFDUP:        {opcode.OP_IFDUP, "OP_IFDUP", 1, opcodeIfDup},
	opcode.OP_DEPTH:        {opcode.OP_DEPTH, "OP_DEPTH", 1, opcodeDepth},
	opcode.OP_DROP:         {opcode.OP_DROP, "OP_DROP", 1, opcodeDrop},
	opcode.OP_DUP:          {opcode.OP_DUP, "OP_DUP", 1, opcodeDup},
	opcode.OP_NIP:          {opcode.OP_NIP, "OP_NIP", 1, opcodeNip},
	opcode.OP_OVER:         {opcode.OP_OVER, "OP_OVER", 1, opcodeOver},
	opcode.OP_PICK:         {opcode.OP_PICK, "OP_PICK", 1, opcodePick},
	opcode.OP_ROLL:         {opcode.OP_ROLL, "OP_ROLL", 1, opcodeRoll},
	opcode.OP_ROT:          {opcode.OP_ROT, "OP_ROT", 1, opcodeRot},
	opcode.OP_SWAP:         {opcode.OP_SWAP, "OP_SWAP", 1, opcodeSwap},
	opcode.OP_TUCK:         {opcode.OP_TUCK, "OP_TUCK", 1, opcodeTuck},

	// Splice opcodes.
	opcode.OP_CAT:    {opcode.OP_CAT, "OP_CAT", 1, opcodeDisabled},
	opcode.OP_SUBSTR: {opcode.OP_SUBSTR, "OP_SUBSTR", 1, opcodeDisabled},
	opcode.OP_LEFT:   {opcode.OP_LEFT, "OP_LEFT", 1, opcodeDisabled},
	opcode.OP_RIGHT:  {opcode.OP_RIGHT, "OP_RIGHT", 1, opcodeDisabled},
	opcode.OP_SIZE:   {opcode.OP_SIZE, "OP_SIZE", 1, opcodeSize},

	// Bitwise logic opcodes.
	opcode.OP_INVERT:      {opcode.OP_INVERT, "OP_INVERT", 1, opcodeDisabled},
	opcode.OP_AND:         {opcode.OP_AND, "OP_AND", 1, opcodeDisabled},
	opcode.OP_OR:          {opcode.OP_OR, "OP_OR", 1, opcodeDisabled},
	opcode.OP_XOR:         {opcode.OP_XOR, "OP_XOR", 1, opcodeDisabled},
	opcode.OP_EQUAL:       {opcode.OP_EQUAL, "OP_EQUAL", 1, opcodeEqual},
	opcode.OP_EQUALVERIFY: {opcode.OP_EQUALVERIFY, "OP_EQUALVERIFY", 1, opcodeEqualVerify},
	opcode.OP_RESERVED1:   {opcode.OP_RESERVED1, "OP_RESERVED1", 1, opcodeReserved},
	opcode.OP_RESERVED2:   {opcode.OP_RESERVED2, "OP_RESERVED2", 1, opcodeReserved},

	// Numeric related opcodes.
	opcode.OP_1ADD:               {opcode.OP_1ADD, "OP_1ADD", 1, opcode1Add},
	opcode.OP_1SUB:               {opcode.OP_1SUB, "OP_1SUB", 1, opcode1Sub},
	opcode.OP_2MUL:               {opcode.OP_2MUL, "OP_2MUL", 1, opcodeDisabled},
	opcode.OP_2DIV:               {opcode.OP_2DIV, "OP_2DIV", 1, opcodeDisabled},
	opcode.OP_NEGATE:             {opcode.OP_NEGATE, "OP_NEGATE", 1, opcodeNegate},
	opcode.OP_ABS:                {opcode.OP_ABS, "OP_ABS", 1, opcodeAbs},
	opcode.OP_NOT:                {opcode.OP_NOT, "OP_NOT", 1, opcodeNot},
	opcode.OP_0NOTEQUAL:          {opcode.OP_0NOTEQUAL, "OP_0NOTEQUAL", 1, opcode0NotEqual},
	opcode.OP_ADD:                {opcode.OP_ADD, "OP_ADD", 1, opcodeAdd},
	opcode.OP_SUB:                {opcode.OP_SUB, "OP_SUB", 1, opcodeSub},
	opcode.OP_MUL:                {opcode.OP_MUL, "OP_MUL", 1, opcodeDisabled},
	opcode.OP_DIV:                {opcode.OP_DIV, "OP_DIV", 1, opcodeDisabled},
	opcode.OP_MOD:                {opcode.OP_MOD, "OP_MOD", 1, opcodeDisabled},
	opcode.OP_LSHIFT:             {opcode.OP_LSHIFT, "OP_LSHIFT", 1, opcodeDisabled},
	opcode.OP_RSHIFT:             {opcode.OP_RSHIFT, "OP_RSHIFT", 1, opcodeDisabled},
	opcode.OP_BOOLAND:            {opcode.OP_BOOLAND, "OP_BOOLAND", 1, opcodeBoolAnd},
	opcode.OP_BOOLOR:             {opcode.OP_BOOLOR, "OP_BOOLOR", 1, opcodeBoolOr},
	opcode.OP_NUMEQUAL:           {opcode.OP_NUMEQUAL, "OP_NUMEQUAL", 1, opcodeNumEqual},
	opcode.OP_NUMEQUALVERIFY:     {opcode.OP_NUMEQUALVERIFY, "OP_NUMEQUALVERIFY", 1, opcodeNumEqualVerify},
	opcode.OP_NUMNOTEQUAL:        {opcode.OP_NUMNOTEQUAL, "OP_NUMNOTEQUAL", 1, opcodeNumNotEqual},
	opcode.OP_LESSTHAN:           {opcode.OP_LESSTHAN, "OP_LESSTHAN", 1, opcodeLessThan},
	opcode.OP_GREATERTHAN:        {opcode.OP_GREATERTHAN, "OP_GREATERTHAN", 1, opcodeGreaterThan},
	opcode.OP_LESSTHANOREQUAL:    {opcode.OP_LESSTHANOREQUAL, "OP_LESSTHANOREQUAL", 1, opcodeLessThanOrEqual},
	opcode.OP_GREATERTHANOREQUAL: {opcode.OP_GREATERTHANOREQUAL, "OP_GREATERTHANOREQUAL", 1, opcodeGreaterThanOrEqual},
	opcode.OP_MIN:                {opcode.OP_MIN, "OP_MIN", 1, opcodeMin},
	opcode.OP_MAX:                {opcode.OP_MAX, "OP_MAX", 1, opcodeMax},
	opcode.OP_WITHIN:             {opcode.OP_WITHIN, "OP_WITHIN", 1, opcodeWithin},

	// Crypto opcodes.
	opcode.OP_RIPEMD160:           {opcode.OP_RIPEMD160, "OP_RIPEMD160", 1, opcodeRipemd160},
	opcode.OP_SHA1:                {opcode.OP_SHA1, "OP_SHA1", 1, opcodeSha1},
	opcode.OP_SHA256:              {opcode.OP_SHA256, "OP_SHA256", 1, opcodeSha256},
	opcode.OP_HASH160:             {opcode.OP_HASH160, "OP_HASH160", 1, opcodeHash160},
	opcode.OP_HASH256:             {opcode.OP_HASH256, "OP_HASH256", 1, opcodeHash256},
	opcode.OP_CODESEPARATOR:       {opcode.OP_CODESEPARATOR, "OP_CODESEPARATOR", 1, opcodeCodeSeparator},
	opcode.OP_CHECKSIG:            {opcode.OP_CHECKSIG, "OP_CHECKSIG", 1, opcodeCheckSig},
	opcode.OP_CHECKSIGVERIFY:      {opcode.OP_CHECKSIGVERIFY, "OP_CHECKSIGVERIFY", 1, opcodeCheckSigVerify},
	opcode.OP_CHECKMULTISIG:       {opcode.OP_CHECKMULTISIG, "OP_CHECKMULTISIG", 1, opcodeCheckMultiSig},
	opcode.OP_CHECKMULTISIGVERIFY: {opcode.OP_CHECKMULTISIGVERIFY, "OP_CHECKMULTISIGVERIFY", 1, opcodeCheckMultiSigVerify},

	// Reserved opcodes.
	opcode.OP_NOP1:  {opcode.OP_NOP1, "OP_NOP1", 1, opcodeNop},
	opcode.OP_NOP4:  {opcode.OP_NOP4, "OP_NOP4", 1, opcodeNop},
	opcode.OP_NOP5:  {opcode.OP_NOP5, "OP_NOP5", 1, opcodeNop},
	opcode.OP_NOP6:  {opcode.OP_NOP6, "OP_NOP6", 1, opcodeNop},
	opcode.OP_NOP7:  {opcode.OP_NOP7, "OP_NOP7", 1, opcodeNop},
	opcode.OP_NOP8:  {opcode.OP_NOP8, "OP_NOP8", 1, opcodeNop},
	opcode.OP_NOP9:  {opcode.OP_NOP9, "OP_NOP9", 1, opcodeNop},
	opcode.OP_NOP10: {opcode.OP_NOP10, "OP_NOP10", 1, opcodeNop},

	// Undefined opcodes.
	opcode.OP_UNKNOWN186: {opcode.OP_UNKNOWN186, "OP_UNKNOWN186", 1, opcodeInvalid},
	opcode.OP_UNKNOWN187: {opcode.OP_UNKNOWN187, "OP_UNKNOWN187", 1, opcodeInvalid},
	opcode.OP_UNKNOWN188: {opcode.OP_UNKNOWN188, "OP_UNKNOWN188", 1, opcodeInvalid},
	opcode.OP_UNKNOWN189: {opcode.OP_UNKNOWN189, "OP_UNKNOWN189", 1, opcodeInvalid},
	opcode.OP_UNKNOWN190: {opcode.OP_UNKNOWN190, "OP_UNKNOWN190", 1, opcodeInvalid},
	opcode.OP_UNKNOWN191: {opcode.OP_UNKNOWN191, "OP_UNKNOWN191", 1, opcodeInvalid},
	opcode.OP_UNKNOWN192: {opcode.OP_UNKNOWN192, "OP_UNKNOWN192", 1, opcodeInvalid},
	opcode.OP_UNKNOWN193: {opcode.OP_UNKNOWN193, "OP_UNKNOWN193", 1, opcodeInvalid},
	opcode.OP_UNKNOWN194: {opcode.OP_UNKNOWN194, "OP_UNKNOWN194", 1, opcodeInvalid},
	opcode.OP_UNKNOWN195: {opcode.OP_UNKNOWN195, "OP_UNKNOWN195", 1, opcodeInvalid},
	opcode.OP_UNKNOWN196: {opcode.OP_UNKNOWN196, "OP_UNKNOWN196", 1, opcodeInvalid},
	opcode.OP_UNKNOWN197: {opcode.OP_UNKNOWN197, "OP_UNKNOWN197", 1, opcodeInvalid},
	opcode.OP_UNKNOWN198: {opcode.OP_UNKNOWN198, "OP_UNKNOWN198", 1, opcodeInvalid},
	opcode.OP_UNKNOWN199: {opcode.OP_UNKNOWN199, "OP_UNKNOWN199", 1, opcodeInvalid},
	opcode.OP_UNKNOWN200: {opcode.OP_UNKNOWN200, "OP_UNKNOWN200", 1, opcodeInvalid},
	opcode.OP_UNKNOWN201: {opcode.OP_UNKNOWN201, "OP_UNKNOWN201", 1, opcodeInvalid},
	opcode.OP_UNKNOWN202: {opcode.OP_UNKNOWN202, "OP_UNKNOWN202", 1, opcodeInvalid},
	opcode.OP_UNKNOWN203: {opcode.OP_UNKNOWN203, "OP_UNKNOWN203", 1, opcodeInvalid},
	opcode.OP_UNKNOWN204: {opcode.OP_UNKNOWN204, "OP_UNKNOWN204", 1, opcodeInvalid},
	opcode.OP_UNKNOWN205: {opcode.OP_UNKNOWN205, "OP_UNKNOWN205", 1, opcodeInvalid},
	opcode.OP_UNKNOWN206: {opcode.OP_UNKNOWN206, "OP_UNKNOWN206", 1, opcodeInvalid},
	opcode.OP_UNKNOWN207: {opcode.OP_UNKNOWN207, "OP_UNKNOWN207", 1, opcodeInvalid},
	opcode.OP_UNKNOWN208: {opcode.OP_UNKNOWN208, "OP_UNKNOWN208", 1, opcodeInvalid},
	opcode.OP_UNKNOWN209: {opcode.OP_UNKNOWN209, "OP_UNKNOWN209", 1, opcodeInvalid},
	opcode.OP_UNKNOWN210: {opcode.OP_UNKNOWN210, "OP_UNKNOWN210", 1, opcodeInvalid},
	opcode.OP_UNKNOWN211: {opcode.OP_UNKNOWN211, "OP_UNKNOWN211", 1, opcodeInvalid},
	opcode.OP_UNKNOWN212: {opcode.OP_UNKNOWN212, "OP_UNKNOWN212", 1, opcodeInvalid},
	opcode.OP_UNKNOWN213: {opcode.OP_UNKNOWN213, "OP_UNKNOWN213", 1, opcodeInvalid},
	opcode.OP_UNKNOWN214: {opcode.OP_UNKNOWN214, "OP_UNKNOWN214", 1, opcodeInvalid},
	opcode.OP_UNKNOWN215: {opcode.OP_UNKNOWN215, "OP_UNKNOWN215", 1, opcodeInvalid},
	opcode.OP_UNKNOWN216: {opcode.OP_UNKNOWN216, "OP_UNKNOWN216", 1, opcodeInvalid},
	opcode.OP_UNKNOWN217: {opcode.OP_UNKNOWN217, "OP_UNKNOWN217", 1, opcodeInvalid},
	opcode.OP_UNKNOWN218: {opcode.OP_UNKNOWN218, "OP_UNKNOWN218", 1, opcodeInvalid},
	opcode.OP_UNKNOWN219: {opcode.OP_UNKNOWN219, "OP_UNKNOWN219", 1, opcodeInvalid},
	opcode.OP_UNKNOWN220: {opcode.OP_UNKNOWN220, "OP_UNKNOWN220", 1, opcodeInvalid},
	opcode.OP_UNKNOWN221: {opcode.OP_UNKNOWN221, "OP_UNKNOWN221", 1, opcodeInvalid},
	opcode.OP_UNKNOWN222: {opcode.OP_UNKNOWN222, "OP_UNKNOWN222", 1, opcodeInvalid},
	opcode.OP_UNKNOWN223: {opcode.OP_UNKNOWN223, "OP_UNKNOWN223", 1, opcodeInvalid},
	opcode.OP_UNKNOWN224: {opcode.OP_UNKNOWN224, "OP_UNKNOWN224", 1, opcodeInvalid},
	opcode.OP_UNKNOWN225: {opcode.OP_UNKNOWN225, "OP_UNKNOWN225", 1, opcodeInvalid},
	opcode.OP_UNKNOWN226: {opcode.OP_UNKNOWN226, "OP_UNKNOWN226", 1, opcodeInvalid},
	opcode.OP_UNKNOWN227: {opcode.OP_UNKNOWN227, "OP_UNKNOWN227", 1, opcodeInvalid},
	opcode.OP_UNKNOWN228: {opcode.OP_UNKNOWN228, "OP_UNKNOWN228", 1, opcodeInvalid},
	opcode.OP_UNKNOWN229: {opcode.OP_UNKNOWN229, "OP_UNKNOWN229", 1, opcodeInvalid},
	opcode.OP_UNKNOWN230: {opcode.OP_UNKNOWN230, "OP_UNKNOWN230", 1, opcodeInvalid},
	opcode.OP_UNKNOWN231: {opcode.OP_UNKNOWN231, "OP_UNKNOWN231", 1, opcodeInvalid},
	opcode.OP_UNKNOWN232: {opcode.OP_UNKNOWN232, "OP_UNKNOWN232", 1, opcodeInvalid},
	opcode.OP_UNKNOWN233: {opcode.OP_UNKNOWN233, "OP_UNKNOWN233", 1, opcodeInvalid},
	opcode.OP_UNKNOWN234: {opcode.OP_UNKNOWN234, "OP_UNKNOWN234", 1, opcodeInvalid},
	opcode.OP_UNKNOWN235: {opcode.OP_UNKNOWN235, "OP_UNKNOWN235", 1, opcodeInvalid},
	opcode.OP_UNKNOWN236: {opcode.OP_UNKNOWN236, "OP_UNKNOWN236", 1, opcodeInvalid},
	opcode.OP_UNKNOWN237: {opcode.OP_UNKNOWN237, "OP_UNKNOWN237", 1, opcodeInvalid},
	opcode.OP_UNKNOWN238: {opcode.OP_UNKNOWN238, "OP_UNKNOWN238", 1, opcodeInvalid},
	opcode.OP_UNKNOWN239: {opcode.OP_UNKNOWN239, "OP_UNKNOWN239", 1, opcodeInvalid},
	opcode.OP_UNKNOWN240: {opcode.OP_UNKNOWN240, "OP_UNKNOWN240", 1, opcodeInvalid},
	opcode.OP_UNKNOWN241: {opcode.OP_UNKNOWN241, "OP_UNKNOWN241", 1, opcodeInvalid},
	opcode.OP_UNKNOWN242: {opcode.OP_UNKNOWN242, "OP_UNKNOWN242", 1, opcodeInvalid},
	opcode.OP_UNKNOWN243: {opcode.OP_UNKNOWN243, "OP_UNKNOWN243", 1, opcodeInvalid},
	opcode.OP_UNKNOWN244: {opcode.OP_UNKNOWN244, "OP_UNKNOWN244", 1, opcodeInvalid},
	opcode.OP_UNKNOWN245: {opcode.OP_UNKNOWN245, "OP_UNKNOWN245", 1, opcodeInvalid},
	opcode.OP_UNKNOWN246: {opcode.OP_UNKNOWN246, "OP_UNKNOWN246", 1, opcodeInvalid},
	opcode.OP_UNKNOWN247: {opcode.OP_UNKNOWN247, "OP_UNKNOWN247", 1, opcodeInvalid},
	opcode.OP_UNKNOWN248: {opcode.OP_UNKNOWN248, "OP_UNKNOWN248", 1, opcodeInvalid},
	opcode.OP_UNKNOWN249: {opcode.OP_UNKNOWN249, "OP_UNKNOWN249", 1, opcodeInvalid},

	// Bitcoin Core internal use opcode.  Defined here for completeness.
	opcode.OP_SMALLINTEGER: {opcode.OP_SMALLINTEGER, "OP_SMALLINTEGER", 1, opcodeInvalid},
	opcode.OP_PUBKEYS:      {opcode.OP_PUBKEYS, "OP_PUBKEYS", 1, opcodeInvalid},
	opcode.OP_UNKNOWN252:   {opcode.OP_UNKNOWN252, "OP_UNKNOWN252", 1, opcodeInvalid},
	opcode.OP_PUBKEYHASH:   {opcode.OP_PUBKEYHASH, "OP_PUBKEYHASH", 1, opcodeInvalid},
	opcode.OP_PUBKEY:       {opcode.OP_PUBKEY, "OP_PUBKEY", 1, opcodeInvalid},

	opcode.OP_INVALIDOPCODE: {opcode.OP_INVALIDOPCODE, "OP_INVALIDOPCODE", 1, opcodeInvalid},
}

// opcodeOnelineRepls defines opcode names which are replaced when doing a
// one-line disassembly.  This is done to match the output of the reference
// implementation while not changing the opcode names in the nicer full
// disassembly.
var opcodeOnelineRepls = map[string]string{
	"OP_1NEGATE": "-1",
	"OP_0":       "0",
	"OP_1":       "1",
	"OP_2":       "2",
	"OP_3":       "3",
	"OP_4":       "4",
	"OP_5":       "5",
	"OP_6":       "6",
	"OP_7":       "7",
	"OP_8":       "8",
	"OP_9":       "9",
	"OP_10":      "10",
	"OP_11":      "11",
	"OP_12":      "12",
	"OP_13":      "13",
	"OP_14":      "14",
	"OP_15":      "15",
	"OP_16":      "16",
}

// parsedOpcode represents an opcode that has been parsed and includes any
// potential data associated with it.
type parsedOpcode struct {
	opcode *opcodeT
	data   []byte
}

// isDisabled returns whether or not the opcode is disabled and thus is always
// bad to see in the instruction stream (even if turned off by a conditional).
func (pop *parsedOpcode) isDisabled() bool {
	switch pop.opcode.value {
	case opcode.OP_CAT:
		return true
	case opcode.OP_SUBSTR:
		return true
	case opcode.OP_LEFT:
		return true
	case opcode.OP_RIGHT:
		return true
	case opcode.OP_INVERT:
		return true
	case opcode.OP_AND:
		return true
	case opcode.OP_OR:
		return true
	case opcode.OP_XOR:
		return true
	case opcode.OP_2MUL:
		return true
	case opcode.OP_2DIV:
		return true
	case opcode.OP_MUL:
		return true
	case opcode.OP_DIV:
		return true
	case opcode.OP_MOD:
		return true
	case opcode.OP_LSHIFT:
		return true
	case opcode.OP_RSHIFT:
		return true
	default:
		return false
	}
}

// alwaysIllegal returns whether or not the opcode is always illegal when passed
// over by the program counter even if in a non-executed branch (it isn't a
// coincidence that they are conditionals).
func (pop *parsedOpcode) alwaysIllegal() bool {
	switch pop.opcode.value {
	case opcode.OP_VERIF:
		return true
	default:
		return false
	}
}

// isConditional returns whether or not the opcode is a conditional opcode which
// changes the conditional execution stack when executed.
func (pop *parsedOpcode) isConditional() bool {
	switch pop.opcode.value {
	case opcode.OP_IF:
		return true
	case opcode.OP_NOTIF:
		return true
	case opcode.OP_ELSE:
		return true
	case opcode.OP_ENDIF:
		return true
	default:
		return false
	}
}

// checkMinimalDataPush returns whether or not the current data push uses the
// smallest possible opcode to represent it.  For example, the value 15 could
// be pushed with OP_DATA_1 15 (among other variations); however, OP_15 is a
// single opcode that represents the same value and is only a single byte versus
// two bytes.
func (pop *parsedOpcode) checkMinimalDataPush() er.R {
	data := pop.data
	dataLen := len(data)
	op := pop.opcode.value

	if dataLen == 0 && op != opcode.OP_0 {
		str := fmt.Sprintf("zero length data push is encoded with "+
			"opcode %s instead of OP_0", pop.opcode.name)
		return txscripterr.ScriptError(txscripterr.ErrMinimalData, str)
	} else if dataLen == 1 && data[0] >= 1 && data[0] <= 16 {
		if op != opcode.OP_1+data[0]-1 {
			// Should have used OP_1 .. OP_16
			str := fmt.Sprintf("data push of the value %d encoded "+
				"with opcode %s instead of OP_%d", data[0],
				pop.opcode.name, data[0])
			return txscripterr.ScriptError(txscripterr.ErrMinimalData, str)
		}
	} else if dataLen == 1 && data[0] == 0x81 {
		if op != opcode.OP_1NEGATE {
			str := fmt.Sprintf("data push of the value -1 encoded "+
				"with opcode %s instead of OP_1NEGATE",
				pop.opcode.name)
			return txscripterr.ScriptError(txscripterr.ErrMinimalData, str)
		}
	} else if dataLen <= 75 {
		if int(op) != dataLen {
			// Should have used a direct push
			str := fmt.Sprintf("data push of %d bytes encoded "+
				"with opcode %s instead of OP_DATA_%d", dataLen,
				pop.opcode.name, dataLen)
			return txscripterr.ScriptError(txscripterr.ErrMinimalData, str)
		}
	} else if dataLen <= 255 {
		if op != opcode.OP_PUSHDATA1 {
			str := fmt.Sprintf("data push of %d bytes encoded "+
				"with opcode %s instead of OP_PUSHDATA1",
				dataLen, pop.opcode.name)
			return txscripterr.ScriptError(txscripterr.ErrMinimalData, str)
		}
	} else if dataLen <= 65535 {
		if op != opcode.OP_PUSHDATA2 {
			str := fmt.Sprintf("data push of %d bytes encoded "+
				"with opcode %s instead of OP_PUSHDATA2",
				dataLen, pop.opcode.name)
			return txscripterr.ScriptError(txscripterr.ErrMinimalData, str)
		}
	}
	return nil
}

// print returns a human-readable string representation of the opcode for use
// in script disassembly.
func (pop *parsedOpcode) print(oneline bool) string {
	// The reference implementation one-line disassembly replaces opcodes
	// which represent values (e.g. OP_0 through OP_16 and OP_1NEGATE)
	// with the raw value.  However, when not doing a one-line dissassembly,
	// we prefer to show the actual opcode names.  Thus, only replace the
	// opcodes in question when the oneline flag is set.
	opcodeName := pop.opcode.name
	if oneline {
		if replName, ok := opcodeOnelineRepls[opcodeName]; ok {
			opcodeName = replName
		}

		// Nothing more to do for non-data push opcodes.
		if pop.opcode.length == 1 {
			return opcodeName
		}

		return fmt.Sprintf("%x", pop.data)
	}

	// Nothing more to do for non-data push opcodes.
	if pop.opcode.length == 1 {
		return opcodeName
	}

	// Add length for the OP_PUSHDATA# opcodes.
	retString := opcodeName
	switch pop.opcode.length {
	case -1:
		retString += fmt.Sprintf(" 0x%02x", len(pop.data))
	case -2:
		retString += fmt.Sprintf(" 0x%04x", len(pop.data))
	case -4:
		retString += fmt.Sprintf(" 0x%08x", len(pop.data))
	}

	return fmt.Sprintf("%s 0x%02x", retString, pop.data)
}

// bytes returns any data associated with the opcode encoded as it would be in
// a script.  This is used for unparsing scripts from parsed opcodes.
func (pop *parsedOpcode) bytes() ([]byte, er.R) {
	var retbytes []byte
	if pop.opcode.length > 0 {
		retbytes = make([]byte, 1, pop.opcode.length)
	} else {
		retbytes = make([]byte, 1, 1+len(pop.data)-
			pop.opcode.length)
	}

	retbytes[0] = pop.opcode.value
	if pop.opcode.length == 1 {
		if len(pop.data) != 0 {
			str := fmt.Sprintf("internal consistency error - "+
				"parsed opcode %s has data length %d when %d "+
				"was expected", pop.opcode.name, len(pop.data),
				0)
			return nil, txscripterr.ScriptError(txscripterr.ErrInternal, str)
		}
		return retbytes, nil
	}
	nbytes := pop.opcode.length
	if pop.opcode.length < 0 {
		l := len(pop.data)
		// tempting just to hardcode to avoid the complexity here.
		switch pop.opcode.length {
		case -1:
			retbytes = append(retbytes, byte(l))
			nbytes = int(retbytes[1]) + len(retbytes)
		case -2:
			retbytes = append(retbytes, byte(l&0xff),
				byte(l>>8&0xff))
			nbytes = int(binary.LittleEndian.Uint16(retbytes[1:])) +
				len(retbytes)
		case -4:
			retbytes = append(retbytes, byte(l&0xff),
				byte((l>>8)&0xff), byte((l>>16)&0xff),
				byte((l>>24)&0xff))
			nbytes = int(binary.LittleEndian.Uint32(retbytes[1:])) +
				len(retbytes)
		}
	}

	retbytes = append(retbytes, pop.data...)

	if len(retbytes) != nbytes {
		str := fmt.Sprintf("internal consistency error - "+
			"parsed opcode %s has data length %d when %d was "+
			"expected", pop.opcode.name, len(retbytes), nbytes)
		return nil, txscripterr.ScriptError(txscripterr.ErrInternal, str)
	}

	return retbytes, nil
}

// *******************************************
// Opcode implementation functions start here.
// *******************************************

// opcodeDisabled is a common handler for disabled opcodes.  It returns an
// appropriate error indicating the opcode is disabled.  While it would
// ordinarily make more sense to detect if the script contains any disabled
// opcodes before executing in an initial parse step, the consensus rules
// dictate the script doesn't fail until the program counter passes over a
// disabled opcode (even when they appear in a branch that is not executed).
func opcodeDisabled(op *parsedOpcode, vm *Engine) er.R {
	str := fmt.Sprintf("attempt to execute disabled opcode %s",
		op.opcode.name)
	return txscripterr.ScriptError(txscripterr.ErrDisabledOpcode, str)
}

// opcodeReserved is a common handler for all reserved opcodes.  It returns an
// appropriate error indicating the opcode is reserved.
func opcodeReserved(op *parsedOpcode, vm *Engine) er.R {
	str := fmt.Sprintf("attempt to execute reserved opcode %s",
		op.opcode.name)
	return txscripterr.ScriptError(txscripterr.ErrReservedOpcode, str)
}

// opcodeInvalid is a common handler for all invalid opcodes.  It returns an
// appropriate error indicating the opcode is invalid.
func opcodeInvalid(op *parsedOpcode, vm *Engine) er.R {
	str := fmt.Sprintf("attempt to execute invalid opcode %s",
		op.opcode.name)
	return txscripterr.ScriptError(txscripterr.ErrReservedOpcode, str)
}

// opcodeFalse pushes an empty array to the data stack to represent false.  Note
// that 0, when encoded as a number according to the numeric encoding consensus
// rules, is an empty array.
func opcodeFalse(op *parsedOpcode, vm *Engine) er.R {
	vm.dstack.PushByteArray(nil)
	return nil
}

// opcodePushData is a common handler for the vast majority of opcodes that push
// raw data (bytes) to the data stack.
func opcodePushData(op *parsedOpcode, vm *Engine) er.R {
	vm.dstack.PushByteArray(op.data)
	return nil
}

// opcode1Negate pushes -1, encoded as a number, to the data stack.
func opcode1Negate(op *parsedOpcode, vm *Engine) er.R {
	vm.dstack.PushInt(scriptnum.ScriptNum_GOMV_UNIQUE_STRING(-1))
	return nil
}

// opcodeN is a common handler for the small integer data push opcodes.  It
// pushes the numeric value the opcode represents (which will be from 1 to 16)
// onto the data stack.
func opcodeN(op *parsedOpcode, vm *Engine) er.R {
	// The opcodes are all defined consecutively, so the numeric value is
	// the difference.
	vm.dstack.PushInt(scriptnum.ScriptNum_GOMV_UNIQUE_STRING((op.opcode.value - (opcode.OP_1 - 1))))
	return nil
}

// opcodeNop is a common handler for the NOP family of opcodes.  As the name
// implies it generally does nothing, however, it will return an error when
// the flag to discourage use of NOPs is set for select opcodes.
func opcodeNop(op *parsedOpcode, vm *Engine) er.R {
	switch op.opcode.value {
	case opcode.OP_NOP1, opcode.OP_NOP4, opcode.OP_NOP5,
		opcode.OP_NOP6, opcode.OP_NOP7, opcode.OP_NOP8, opcode.OP_NOP9, opcode.OP_NOP10:
		if vm.hasFlag(ScriptDiscourageUpgradableNops) {
			str := fmt.Sprintf("OP_NOP%d reserved for soft-fork "+
				"upgrades", op.opcode.value-(opcode.OP_NOP1-1))
			return txscripterr.ScriptError(txscripterr.ErrDiscourageUpgradableNOPs, str)
		}
	}
	return nil
}

// popIfBool enforces the "minimal if" policy during script execution if the
// particular flag is set.  If so, in order to eliminate an additional source
// of nuisance malleability, post-segwit for version 0 witness programs, we now
// require the following: for OP_IF and OP_NOT_IF, the top stack item MUST
// either be an empty byte slice, or [0x01]. Otherwise, the item at the top of
// the stack will be popped and interpreted as a boolean.
func popIfBool(vm *Engine) (bool, er.R) {
	// When not in witness execution mode, not executing a v0 witness
	// program, or the minimal if flag isn't set pop the top stack item as
	// a normal bool.
	if !vm.isWitnessVersionActive(0) || !vm.hasFlag(ScriptVerifyMinimalIf) {
		return vm.dstack.PopBool()
	}

	// At this point, a v0 witness program is being executed and the minimal
	// if flag is set, so enforce additional constraints on the top stack
	// item.
	so, err := vm.dstack.PopByteArray()
	if err != nil {
		return false, err
	}

	// The top element MUST have a length of at least one.
	if len(so) > 1 {
		str := fmt.Sprintf("minimal if is active, top element MUST "+
			"have a length of at least, instead length is %v",
			len(so))
		return false, txscripterr.ScriptError(txscripterr.ErrMinimalIf, str)
	}

	// Additionally, if the length is one, then the value MUST be 0x01.
	if len(so) == 1 && so[0] != 0x01 {
		str := fmt.Sprintf("minimal if is active, top stack item MUST "+
			"be an empty byte array or 0x01, is instead: %v",
			so[0])
		return false, txscripterr.ScriptError(txscripterr.ErrMinimalIf, str)
	}

	return asBool(so), nil
}

// opcodeIf treats the top item on the data stack as a boolean and removes it.
//
// An appropriate entry is added to the conditional stack depending on whether
// the boolean is true and whether this if is on an executing branch in order
// to allow proper execution of further opcodes depending on the conditional
// logic.  When the boolean is true, the first branch will be executed (unless
// this opcode is nested in a non-executed branch).
//
// <expression> if [statements] [else [statements]] endif
//
// Note that, unlike for all non-conditional opcodes, this is executed even when
// it is on a non-executing branch so proper nesting is maintained.
//
// Data stack transformation: [... bool] -> [...]
// Conditional stack transformation: [...] -> [... OpCondValue]
func opcodeIf(op *parsedOpcode, vm *Engine) er.R {
	condVal := OpCondFalse
	if vm.isBranchExecuting() {
		ok, err := popIfBool(vm)
		if err != nil {
			return err
		}

		if ok {
			condVal = OpCondTrue
		}
	} else {
		condVal = OpCondSkip
	}
	vm.condStack = append(vm.condStack, condVal)
	return nil
}

// opcodeNotIf treats the top item on the data stack as a boolean and removes
// it.
//
// An appropriate entry is added to the conditional stack depending on whether
// the boolean is true and whether this if is on an executing branch in order
// to allow proper execution of further opcodes depending on the conditional
// logic.  When the boolean is false, the first branch will be executed (unless
// this opcode is nested in a non-executed branch).
//
// <expression> notif [statements] [else [statements]] endif
//
// Note that, unlike for all non-conditional opcodes, this is executed even when
// it is on a non-executing branch so proper nesting is maintained.
//
// Data stack transformation: [... bool] -> [...]
// Conditional stack transformation: [...] -> [... OpCondValue]
func opcodeNotIf(op *parsedOpcode, vm *Engine) er.R {
	condVal := OpCondFalse
	if vm.isBranchExecuting() {
		ok, err := popIfBool(vm)
		if err != nil {
			return err
		}

		if !ok {
			condVal = OpCondTrue
		}
	} else {
		condVal = OpCondSkip
	}
	vm.condStack = append(vm.condStack, condVal)
	return nil
}

// opcodeElse inverts conditional execution for other half of if/else/endif.
//
// An error is returned if there has not already been a matching OP_IF.
//
// Conditional stack transformation: [... OpCondValue] -> [... !OpCondValue]
func opcodeElse(op *parsedOpcode, vm *Engine) er.R {
	if len(vm.condStack) == 0 {
		str := fmt.Sprintf("encountered opcode %s with no matching "+
			"opcode to begin conditional execution", op.opcode.name)
		return txscripterr.ScriptError(txscripterr.ErrUnbalancedConditional, str)
	}

	conditionalIdx := len(vm.condStack) - 1
	switch vm.condStack[conditionalIdx] {
	case OpCondTrue:
		vm.condStack[conditionalIdx] = OpCondFalse
	case OpCondFalse:
		vm.condStack[conditionalIdx] = OpCondTrue
	case OpCondSkip:
		// Value doesn't change in skip since it indicates this opcode
		// is nested in a non-executed branch.
	}
	return nil
}

// opcodeEndif terminates a conditional block, removing the value from the
// conditional execution stack.
//
// An error is returned if there has not already been a matching OP_IF.
//
// Conditional stack transformation: [... OpCondValue] -> [...]
func opcodeEndif(op *parsedOpcode, vm *Engine) er.R {
	if len(vm.condStack) == 0 {
		str := fmt.Sprintf("encountered opcode %s with no matching "+
			"opcode to begin conditional execution", op.opcode.name)
		return txscripterr.ScriptError(txscripterr.ErrUnbalancedConditional, str)
	}

	vm.condStack = vm.condStack[:len(vm.condStack)-1]
	return nil
}

// abstractVerify examines the top item on the data stack as a boolean value and
// verifies it evaluates to true.  An error is returned either when there is no
// item on the stack or when that item evaluates to false.  In the latter case
// where the verification fails specifically due to the top item evaluating
// to false, the returned error will use the passed error code.
func abstractVerify(op *parsedOpcode, vm *Engine, c *er.ErrorCode) er.R {
	verified, err := vm.dstack.PopBool()
	if err != nil {
		return err
	}

	if !verified {
		str := fmt.Sprintf("%s failed", op.opcode.name)
		return txscripterr.ScriptError(c, str)
	}
	return nil
}

// opcodeVerify examines the top item on the data stack as a boolean value and
// verifies it evaluates to true.  An error is returned if it does not.
func opcodeVerify(op *parsedOpcode, vm *Engine) er.R {
	return abstractVerify(op, vm, txscripterr.ErrVerify)
}

// opcodeReturn returns an appropriate error since it is always an error to
// return early from a script.
func opcodeReturn(op *parsedOpcode, vm *Engine) er.R {
	return txscripterr.ScriptError(txscripterr.ErrEarlyReturn, "script returned early")
}

// verifyLockTime is a helper function used to validate locktimes.
func verifyLockTime(txLockTime, threshold, lockTime int64) er.R {
	// The lockTimes in both the script and transaction must be of the same
	// type.
	if !((txLockTime < threshold && lockTime < threshold) ||
		(txLockTime >= threshold && lockTime >= threshold)) {
		str := fmt.Sprintf("mismatched locktime types -- tx locktime "+
			"%d, stack locktime %d", txLockTime, lockTime)
		return txscripterr.ScriptError(txscripterr.ErrUnsatisfiedLockTime, str)
	}

	if lockTime > txLockTime {
		str := fmt.Sprintf("locktime requirement not satisfied -- "+
			"locktime is greater than the transaction locktime: "+
			"%d > %d", lockTime, txLockTime)
		return txscripterr.ScriptError(txscripterr.ErrUnsatisfiedLockTime, str)
	}

	return nil
}

// opcodeCheckLockTimeVerify compares the top item on the data stack to the
// LockTime field of the transaction containing the script signature
// validating if the transaction outputs are spendable yet.  If flag
// ScriptVerifyCheckLockTimeVerify is not set, the code continues as if OP_NOP2
// were executed.
func opcodeCheckLockTimeVerify(op *parsedOpcode, vm *Engine) er.R {
	// If the ScriptVerifyCheckLockTimeVerify script flag is not set, treat
	// opcode as OP_NOP2 instead.
	if !vm.hasFlag(ScriptVerifyCheckLockTimeVerify) {
		if vm.hasFlag(ScriptDiscourageUpgradableNops) {
			return txscripterr.ScriptError(txscripterr.ErrDiscourageUpgradableNOPs,
				"OP_NOP2 reserved for soft-fork upgrades")
		}
		return nil
	}

	// The current transaction locktime is a uint32 resulting in a maximum
	// locktime of 2^32-1 (the year 2106).  However, scriptNums are signed
	// and therefore a standard 4-byte scriptNum would only support up to a
	// maximum of 2^31-1 (the year 2038).  Thus, a 5-byte scriptNum is used
	// here since it will support up to 2^39-1 which allows dates beyond the
	// current locktime limit.
	//
	// PeekByteArray is used here instead of PeekInt because we do not want
	// to be limited to a 4-byte integer for reasons specified above.
	so, err := vm.dstack.PeekByteArray(0)
	if err != nil {
		return err
	}
	lockTime, err := scriptnum.MakeScriptNum_GOMV_UNIQUE_STRING(so, vm.dstack.verifyMinimalData, 5)
	if err != nil {
		return err
	}

	// In the rare event that the argument needs to be < 0 due to some
	// arithmetic being done first, you can always use
	// 0 OP_MAX OP_CHECKLOCKTIMEVERIFY.
	if lockTime < 0 {
		str := fmt.Sprintf("negative lock time: %d", lockTime)
		return txscripterr.ScriptError(txscripterr.ErrNegativeLockTime, str)
	}

	// The lock time field of a transaction is either a block height at
	// which the transaction is finalized or a timestamp depending on if the
	// value is before the txscript.LockTimeThreshold.  When it is under the
	// threshold it is a block height.
	err = verifyLockTime(int64(vm.tx.LockTime), LockTimeThreshold,
		int64(lockTime))
	if err != nil {
		return err
	}

	// The lock time feature can also be disabled, thereby bypassing
	// OP_CHECKLOCKTIMEVERIFY, if every transaction input has been finalized by
	// setting its sequence to the maximum value (wire.MaxTxInSequenceNum).  This
	// condition would result in the transaction being allowed into the blockchain
	// making the opcode ineffective.
	//
	// This condition is prevented by enforcing that the input being used by
	// the opcode is unlocked (its sequence number is less than the max
	// value).  This is sufficient to prove correctness without having to
	// check every input.
	//
	// NOTE: This implies that even if the transaction is not finalized due to
	// another input being unlocked, the opcode execution will still fail when the
	// input being used by the opcode is locked.
	if vm.tx.TxIn[vm.txIdx].Sequence == wire.MaxTxInSequenceNum {
		return txscripterr.ScriptError(txscripterr.ErrUnsatisfiedLockTime,
			"transaction input is finalized")
	}

	return nil
}

// opcodeCheckSequenceVerify compares the top item on the data stack to the
// LockTime field of the transaction containing the script signature
// validating if the transaction outputs are spendable yet.  If flag
// ScriptVerifyCheckSequenceVerify is not set, the code continues as if OP_NOP3
// were executed.
func opcodeCheckSequenceVerify(op *parsedOpcode, vm *Engine) er.R {
	// If the ScriptVerifyCheckSequenceVerify script flag is not set, treat
	// opcode as OP_NOP3 instead.
	if !vm.hasFlag(ScriptVerifyCheckSequenceVerify) {
		if vm.hasFlag(ScriptDiscourageUpgradableNops) {
			return txscripterr.ScriptError(txscripterr.ErrDiscourageUpgradableNOPs,
				"OP_NOP3 reserved for soft-fork upgrades")
		}
		return nil
	}

	// The current transaction sequence is a uint32 resulting in a maximum
	// sequence of 2^32-1.  However, scriptNums are signed and therefore a
	// standard 4-byte scriptNum would only support up to a maximum of
	// 2^31-1.  Thus, a 5-byte scriptNum is used here since it will support
	// up to 2^39-1 which allows sequences beyond the current sequence
	// limit.
	//
	// PeekByteArray is used here instead of PeekInt because we do not want
	// to be limited to a 4-byte integer for reasons specified above.
	so, err := vm.dstack.PeekByteArray(0)
	if err != nil {
		return err
	}
	stackSequence, err := scriptnum.MakeScriptNum_GOMV_UNIQUE_STRING(so, vm.dstack.verifyMinimalData, 5)
	if err != nil {
		return err
	}

	// In the rare event that the argument needs to be < 0 due to some
	// arithmetic being done first, you can always use
	// 0 OP_MAX OP_CHECKSEQUENCEVERIFY.
	if stackSequence < 0 {
		str := fmt.Sprintf("negative sequence: %d", stackSequence)
		return txscripterr.ScriptError(txscripterr.ErrNegativeLockTime, str)
	}

	sequence := int64(stackSequence)

	// To provide for future soft-fork extensibility, if the
	// operand has the disabled lock-time flag set,
	// CHECKSEQUENCEVERIFY behaves as a NOP.
	if sequence&int64(wire.SequenceLockTimeDisabled) != 0 {
		return nil
	}

	// Transaction version numbers not high enough to trigger CSV rules must
	// fail.
	if vm.tx.Version < 2 {
		str := fmt.Sprintf("invalid transaction version: %d",
			vm.tx.Version)
		return txscripterr.ScriptError(txscripterr.ErrUnsatisfiedLockTime, str)
	}

	// Sequence numbers with their most significant bit set are not
	// consensus constrained. Testing that the transaction's sequence
	// number does not have this bit set prevents using this property
	// to get around a CHECKSEQUENCEVERIFY check.
	txSequence := int64(vm.tx.TxIn[vm.txIdx].Sequence)
	if txSequence&int64(wire.SequenceLockTimeDisabled) != 0 {
		str := fmt.Sprintf("transaction sequence has sequence "+
			"locktime disabled bit set: 0x%x", txSequence)
		return txscripterr.ScriptError(txscripterr.ErrUnsatisfiedLockTime, str)
	}

	// Mask off non-consensus bits before doing comparisons.
	lockTimeMask := int64(wire.SequenceLockTimeIsSeconds |
		wire.SequenceLockTimeMask)
	return verifyLockTime(txSequence&lockTimeMask,
		wire.SequenceLockTimeIsSeconds, sequence&lockTimeMask)
}

// opcodeToAltStack removes the top item from the main data stack and pushes it
// onto the alternate data stack.
//
// Main data stack transformation: [... x1 x2 x3] -> [... x1 x2]
// Alt data stack transformation:  [... y1 y2 y3] -> [... y1 y2 y3 x3]
func opcodeToAltStack(op *parsedOpcode, vm *Engine) er.R {
	so, err := vm.dstack.PopByteArray()
	if err != nil {
		return err
	}
	vm.astack.PushByteArray(so)

	return nil
}

// opcodeFromAltStack removes the top item from the alternate data stack and
// pushes it onto the main data stack.
//
// Main data stack transformation: [... x1 x2 x3] -> [... x1 x2 x3 y3]
// Alt data stack transformation:  [... y1 y2 y3] -> [... y1 y2]
func opcodeFromAltStack(op *parsedOpcode, vm *Engine) er.R {
	so, err := vm.astack.PopByteArray()
	if err != nil {
		return err
	}
	vm.dstack.PushByteArray(so)

	return nil
}

// opcode2Drop removes the top 2 items from the data stack.
//
// Stack transformation: [... x1 x2 x3] -> [... x1]
func opcode2Drop(op *parsedOpcode, vm *Engine) er.R {
	return vm.dstack.DropN(2)
}

// opcode2Dup duplicates the top 2 items on the data stack.
//
// Stack transformation: [... x1 x2 x3] -> [... x1 x2 x3 x2 x3]
func opcode2Dup(op *parsedOpcode, vm *Engine) er.R {
	return vm.dstack.DupN(2)
}

// opcode3Dup duplicates the top 3 items on the data stack.
//
// Stack transformation: [... x1 x2 x3] -> [... x1 x2 x3 x1 x2 x3]
func opcode3Dup(op *parsedOpcode, vm *Engine) er.R {
	return vm.dstack.DupN(3)
}

// opcode2Over duplicates the 2 items before the top 2 items on the data stack.
//
// Stack transformation: [... x1 x2 x3 x4] -> [... x1 x2 x3 x4 x1 x2]
func opcode2Over(op *parsedOpcode, vm *Engine) er.R {
	return vm.dstack.OverN(2)
}

// opcode2Rot rotates the top 6 items on the data stack to the left twice.
//
// Stack transformation: [... x1 x2 x3 x4 x5 x6] -> [... x3 x4 x5 x6 x1 x2]
func opcode2Rot(op *parsedOpcode, vm *Engine) er.R {
	return vm.dstack.RotN(2)
}

// opcode2Swap swaps the top 2 items on the data stack with the 2 that come
// before them.
//
// Stack transformation: [... x1 x2 x3 x4] -> [... x3 x4 x1 x2]
func opcode2Swap(op *parsedOpcode, vm *Engine) er.R {
	return vm.dstack.SwapN(2)
}

// opcodeIfDup duplicates the top item of the stack if it is not zero.
//
// Stack transformation (x1==0): [... x1] -> [... x1]
// Stack transformation (x1!=0): [... x1] -> [... x1 x1]
func opcodeIfDup(op *parsedOpcode, vm *Engine) er.R {
	so, err := vm.dstack.PeekByteArray(0)
	if err != nil {
		return err
	}

	// Push copy of data iff it isn't zero
	if asBool(so) {
		vm.dstack.PushByteArray(so)
	}

	return nil
}

// opcodeDepth pushes the depth of the data stack prior to executing this
// opcode, encoded as a number, onto the data stack.
//
// Stack transformation: [...] -> [... <num of items on the stack>]
// Example with 2 items: [x1 x2] -> [x1 x2 2]
// Example with 3 items: [x1 x2 x3] -> [x1 x2 x3 3]
func opcodeDepth(op *parsedOpcode, vm *Engine) er.R {
	vm.dstack.PushInt(scriptnum.ScriptNum_GOMV_UNIQUE_STRING(vm.dstack.Depth()))
	return nil
}

// opcodeDrop removes the top item from the data stack.
//
// Stack transformation: [... x1 x2 x3] -> [... x1 x2]
func opcodeDrop(op *parsedOpcode, vm *Engine) er.R {
	return vm.dstack.DropN(1)
}

// opcodeDup duplicates the top item on the data stack.
//
// Stack transformation: [... x1 x2 x3] -> [... x1 x2 x3 x3]
func opcodeDup(op *parsedOpcode, vm *Engine) er.R {
	return vm.dstack.DupN(1)
}

// opcodeNip removes the item before the top item on the data stack.
//
// Stack transformation: [... x1 x2 x3] -> [... x1 x3]
func opcodeNip(op *parsedOpcode, vm *Engine) er.R {
	return vm.dstack.NipN(1)
}

// opcodeOver duplicates the item before the top item on the data stack.
//
// Stack transformation: [... x1 x2 x3] -> [... x1 x2 x3 x2]
func opcodeOver(op *parsedOpcode, vm *Engine) er.R {
	return vm.dstack.OverN(1)
}

// opcodePick treats the top item on the data stack as an integer and duplicates
// the item on the stack that number of items back to the top.
//
// Stack transformation: [xn ... x2 x1 x0 n] -> [xn ... x2 x1 x0 xn]
// Example with n=1: [x2 x1 x0 1] -> [x2 x1 x0 x1]
// Example with n=2: [x2 x1 x0 2] -> [x2 x1 x0 x2]
func opcodePick(op *parsedOpcode, vm *Engine) er.R {
	val, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	return vm.dstack.PickN(val.Int32())
}

// opcodeRoll treats the top item on the data stack as an integer and moves
// the item on the stack that number of items back to the top.
//
// Stack transformation: [xn ... x2 x1 x0 n] -> [... x2 x1 x0 xn]
// Example with n=1: [x2 x1 x0 1] -> [x2 x0 x1]
// Example with n=2: [x2 x1 x0 2] -> [x1 x0 x2]
func opcodeRoll(op *parsedOpcode, vm *Engine) er.R {
	val, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	return vm.dstack.RollN(val.Int32())
}

// opcodeRot rotates the top 3 items on the data stack to the left.
//
// Stack transformation: [... x1 x2 x3] -> [... x2 x3 x1]
func opcodeRot(op *parsedOpcode, vm *Engine) er.R {
	return vm.dstack.RotN(1)
}

// opcodeSwap swaps the top two items on the stack.
//
// Stack transformation: [... x1 x2] -> [... x2 x1]
func opcodeSwap(op *parsedOpcode, vm *Engine) er.R {
	return vm.dstack.SwapN(1)
}

// opcodeTuck inserts a duplicate of the top item of the data stack before the
// second-to-top item.
//
// Stack transformation: [... x1 x2] -> [... x2 x1 x2]
func opcodeTuck(op *parsedOpcode, vm *Engine) er.R {
	return vm.dstack.Tuck()
}

// opcodeSize pushes the size of the top item of the data stack onto the data
// stack.
//
// Stack transformation: [... x1] -> [... x1 len(x1)]
func opcodeSize(op *parsedOpcode, vm *Engine) er.R {
	so, err := vm.dstack.PeekByteArray(0)
	if err != nil {
		return err
	}

	vm.dstack.PushInt(scriptnum.ScriptNum_GOMV_UNIQUE_STRING(len(so)))
	return nil
}

// opcodeEqual removes the top 2 items of the data stack, compares them as raw
// bytes, and pushes the result, encoded as a boolean, back to the stack.
//
// Stack transformation: [... x1 x2] -> [... bool]
func opcodeEqual(op *parsedOpcode, vm *Engine) er.R {
	a, err := vm.dstack.PopByteArray()
	if err != nil {
		return err
	}
	b, err := vm.dstack.PopByteArray()
	if err != nil {
		return err
	}

	vm.dstack.PushBool(bytes.Equal(a, b))
	return nil
}

// opcodeEqualVerify is a combination of opcodeEqual and opcodeVerify.
// Specifically, it removes the top 2 items of the data stack, compares them,
// and pushes the result, encoded as a boolean, back to the stack.  Then, it
// examines the top item on the data stack as a boolean value and verifies it
// evaluates to true.  An error is returned if it does not.
//
// Stack transformation: [... x1 x2] -> [... bool] -> [...]
func opcodeEqualVerify(op *parsedOpcode, vm *Engine) er.R {
	err := opcodeEqual(op, vm)
	if err == nil {
		err = abstractVerify(op, vm, txscripterr.ErrEqualVerify)
	}
	return err
}

// opcode1Add treats the top item on the data stack as an integer and replaces
// it with its incremented value (plus 1).
//
// Stack transformation: [... x1 x2] -> [... x1 x2+1]
func opcode1Add(op *parsedOpcode, vm *Engine) er.R {
	m, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	vm.dstack.PushInt(m + 1)
	return nil
}

// opcode1Sub treats the top item on the data stack as an integer and replaces
// it with its decremented value (minus 1).
//
// Stack transformation: [... x1 x2] -> [... x1 x2-1]
func opcode1Sub(op *parsedOpcode, vm *Engine) er.R {
	m, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}
	vm.dstack.PushInt(m - 1)

	return nil
}

// opcodeNegate treats the top item on the data stack as an integer and replaces
// it with its negation.
//
// Stack transformation: [... x1 x2] -> [... x1 -x2]
func opcodeNegate(op *parsedOpcode, vm *Engine) er.R {
	m, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	vm.dstack.PushInt(-m)
	return nil
}

// opcodeAbs treats the top item on the data stack as an integer and replaces it
// it with its absolute value.
//
// Stack transformation: [... x1 x2] -> [... x1 abs(x2)]
func opcodeAbs(op *parsedOpcode, vm *Engine) er.R {
	m, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	if m < 0 {
		m = -m
	}
	vm.dstack.PushInt(m)
	return nil
}

// opcodeNot treats the top item on the data stack as an integer and replaces
// it with its "inverted" value (0 becomes 1, non-zero becomes 0).
//
// NOTE: While it would probably make more sense to treat the top item as a
// boolean, and push the opposite, which is really what the intention of this
// opcode is, it is extremely important that is not done because integers are
// interpreted differently than booleans and the consensus rules for this opcode
// dictate the item is interpreted as an integer.
//
// Stack transformation (x2==0): [... x1 0] -> [... x1 1]
// Stack transformation (x2!=0): [... x1 1] -> [... x1 0]
// Stack transformation (x2!=0): [... x1 17] -> [... x1 0]
func opcodeNot(op *parsedOpcode, vm *Engine) er.R {
	m, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	if m == 0 {
		vm.dstack.PushInt(scriptnum.ScriptNum_GOMV_UNIQUE_STRING(1))
	} else {
		vm.dstack.PushInt(scriptnum.ScriptNum_GOMV_UNIQUE_STRING(0))
	}
	return nil
}

// opcode0NotEqual treats the top item on the data stack as an integer and
// replaces it with either a 0 if it is zero, or a 1 if it is not zero.
//
// Stack transformation (x2==0): [... x1 0] -> [... x1 0]
// Stack transformation (x2!=0): [... x1 1] -> [... x1 1]
// Stack transformation (x2!=0): [... x1 17] -> [... x1 1]
func opcode0NotEqual(op *parsedOpcode, vm *Engine) er.R {
	m, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	if m != 0 {
		m = 1
	}
	vm.dstack.PushInt(m)
	return nil
}

// opcodeAdd treats the top two items on the data stack as integers and replaces
// them with their sum.
//
// Stack transformation: [... x1 x2] -> [... x1+x2]
func opcodeAdd(op *parsedOpcode, vm *Engine) er.R {
	v0, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	v1, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	vm.dstack.PushInt(v0 + v1)
	return nil
}

// opcodeSub treats the top two items on the data stack as integers and replaces
// them with the result of subtracting the top entry from the second-to-top
// entry.
//
// Stack transformation: [... x1 x2] -> [... x1-x2]
func opcodeSub(op *parsedOpcode, vm *Engine) er.R {
	v0, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	v1, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	vm.dstack.PushInt(v1 - v0)
	return nil
}

// opcodeBoolAnd treats the top two items on the data stack as integers.  When
// both of them are not zero, they are replaced with a 1, otherwise a 0.
//
// Stack transformation (x1==0, x2==0): [... 0 0] -> [... 0]
// Stack transformation (x1!=0, x2==0): [... 5 0] -> [... 0]
// Stack transformation (x1==0, x2!=0): [... 0 7] -> [... 0]
// Stack transformation (x1!=0, x2!=0): [... 4 8] -> [... 1]
func opcodeBoolAnd(op *parsedOpcode, vm *Engine) er.R {
	v0, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	v1, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	if v0 != 0 && v1 != 0 {
		vm.dstack.PushInt(scriptnum.ScriptNum_GOMV_UNIQUE_STRING(1))
	} else {
		vm.dstack.PushInt(scriptnum.ScriptNum_GOMV_UNIQUE_STRING(0))
	}

	return nil
}

// opcodeBoolOr treats the top two items on the data stack as integers.  When
// either of them are not zero, they are replaced with a 1, otherwise a 0.
//
// Stack transformation (x1==0, x2==0): [... 0 0] -> [... 0]
// Stack transformation (x1!=0, x2==0): [... 5 0] -> [... 1]
// Stack transformation (x1==0, x2!=0): [... 0 7] -> [... 1]
// Stack transformation (x1!=0, x2!=0): [... 4 8] -> [... 1]
func opcodeBoolOr(op *parsedOpcode, vm *Engine) er.R {
	v0, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	v1, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	if v0 != 0 || v1 != 0 {
		vm.dstack.PushInt(scriptnum.ScriptNum_GOMV_UNIQUE_STRING(1))
	} else {
		vm.dstack.PushInt(scriptnum.ScriptNum_GOMV_UNIQUE_STRING(0))
	}

	return nil
}

// opcodeNumEqual treats the top two items on the data stack as integers.  When
// they are equal, they are replaced with a 1, otherwise a 0.
//
// Stack transformation (x1==x2): [... 5 5] -> [... 1]
// Stack transformation (x1!=x2): [... 5 7] -> [... 0]
func opcodeNumEqual(op *parsedOpcode, vm *Engine) er.R {
	v0, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	v1, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	if v0 == v1 {
		vm.dstack.PushInt(scriptnum.ScriptNum_GOMV_UNIQUE_STRING(1))
	} else {
		vm.dstack.PushInt(scriptnum.ScriptNum_GOMV_UNIQUE_STRING(0))
	}

	return nil
}

// opcodeNumEqualVerify is a combination of opcodeNumEqual and opcodeVerify.
//
// Specifically, treats the top two items on the data stack as integers.  When
// they are equal, they are replaced with a 1, otherwise a 0.  Then, it examines
// the top item on the data stack as a boolean value and verifies it evaluates
// to true.  An error is returned if it does not.
//
// Stack transformation: [... x1 x2] -> [... bool] -> [...]
func opcodeNumEqualVerify(op *parsedOpcode, vm *Engine) er.R {
	err := opcodeNumEqual(op, vm)
	if err == nil {
		err = abstractVerify(op, vm, txscripterr.ErrNumEqualVerify)
	}
	return err
}

// opcodeNumNotEqual treats the top two items on the data stack as integers.
// When they are NOT equal, they are replaced with a 1, otherwise a 0.
//
// Stack transformation (x1==x2): [... 5 5] -> [... 0]
// Stack transformation (x1!=x2): [... 5 7] -> [... 1]
func opcodeNumNotEqual(op *parsedOpcode, vm *Engine) er.R {
	v0, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	v1, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	if v0 != v1 {
		vm.dstack.PushInt(scriptnum.ScriptNum_GOMV_UNIQUE_STRING(1))
	} else {
		vm.dstack.PushInt(scriptnum.ScriptNum_GOMV_UNIQUE_STRING(0))
	}

	return nil
}

// opcodeLessThan treats the top two items on the data stack as integers.  When
// the second-to-top item is less than the top item, they are replaced with a 1,
// otherwise a 0.
//
// Stack transformation: [... x1 x2] -> [... bool]
func opcodeLessThan(op *parsedOpcode, vm *Engine) er.R {
	v0, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	v1, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	if v1 < v0 {
		vm.dstack.PushInt(scriptnum.ScriptNum_GOMV_UNIQUE_STRING(1))
	} else {
		vm.dstack.PushInt(scriptnum.ScriptNum_GOMV_UNIQUE_STRING(0))
	}

	return nil
}

// opcodeGreaterThan treats the top two items on the data stack as integers.
// When the second-to-top item is greater than the top item, they are replaced
// with a 1, otherwise a 0.
//
// Stack transformation: [... x1 x2] -> [... bool]
func opcodeGreaterThan(op *parsedOpcode, vm *Engine) er.R {
	v0, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	v1, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	if v1 > v0 {
		vm.dstack.PushInt(scriptnum.ScriptNum_GOMV_UNIQUE_STRING(1))
	} else {
		vm.dstack.PushInt(scriptnum.ScriptNum_GOMV_UNIQUE_STRING(0))
	}
	return nil
}

// opcodeLessThanOrEqual treats the top two items on the data stack as integers.
// When the second-to-top item is less than or equal to the top item, they are
// replaced with a 1, otherwise a 0.
//
// Stack transformation: [... x1 x2] -> [... bool]
func opcodeLessThanOrEqual(op *parsedOpcode, vm *Engine) er.R {
	v0, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	v1, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	if v1 <= v0 {
		vm.dstack.PushInt(scriptnum.ScriptNum_GOMV_UNIQUE_STRING(1))
	} else {
		vm.dstack.PushInt(scriptnum.ScriptNum_GOMV_UNIQUE_STRING(0))
	}
	return nil
}

// opcodeGreaterThanOrEqual treats the top two items on the data stack as
// integers.  When the second-to-top item is greater than or equal to the top
// item, they are replaced with a 1, otherwise a 0.
//
// Stack transformation: [... x1 x2] -> [... bool]
func opcodeGreaterThanOrEqual(op *parsedOpcode, vm *Engine) er.R {
	v0, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	v1, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	if v1 >= v0 {
		vm.dstack.PushInt(scriptnum.ScriptNum_GOMV_UNIQUE_STRING(1))
	} else {
		vm.dstack.PushInt(scriptnum.ScriptNum_GOMV_UNIQUE_STRING(0))
	}

	return nil
}

// opcodeMin treats the top two items on the data stack as integers and replaces
// them with the minimum of the two.
//
// Stack transformation: [... x1 x2] -> [... min(x1, x2)]
func opcodeMin(op *parsedOpcode, vm *Engine) er.R {
	v0, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	v1, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	if v1 < v0 {
		vm.dstack.PushInt(v1)
	} else {
		vm.dstack.PushInt(v0)
	}

	return nil
}

// opcodeMax treats the top two items on the data stack as integers and replaces
// them with the maximum of the two.
//
// Stack transformation: [... x1 x2] -> [... max(x1, x2)]
func opcodeMax(op *parsedOpcode, vm *Engine) er.R {
	v0, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	v1, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	if v1 > v0 {
		vm.dstack.PushInt(v1)
	} else {
		vm.dstack.PushInt(v0)
	}

	return nil
}

// opcodeWithin treats the top 3 items on the data stack as integers.  When the
// value to test is within the specified range (left inclusive), they are
// replaced with a 1, otherwise a 0.
//
// The top item is the max value, the second-top-item is the minimum value, and
// the third-to-top item is the value to test.
//
// Stack transformation: [... x1 min max] -> [... bool]
func opcodeWithin(op *parsedOpcode, vm *Engine) er.R {
	maxVal, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	minVal, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	x, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	if x >= minVal && x < maxVal {
		vm.dstack.PushInt(scriptnum.ScriptNum_GOMV_UNIQUE_STRING(1))
	} else {
		vm.dstack.PushInt(scriptnum.ScriptNum_GOMV_UNIQUE_STRING(0))
	}
	return nil
}

// opcodeRipemd160 treats the top item of the data stack as raw bytes and
// replaces it with ripemd160(data).
//
// Stack transformation: [... x1] -> [... ripemd160(x1)]
func opcodeRipemd160(op *parsedOpcode, vm *Engine) er.R {
	buf, err := vm.dstack.PopByteArray()
	if err != nil {
		return err
	}

	vm.dstack.PushByteArray(btcutil.Ripemd160(buf))
	return nil
}

// opcodeSha1 treats the top item of the data stack as raw bytes and replaces it
// with sha1(data).
//
// Stack transformation: [... x1] -> [... sha1(x1)]
func opcodeSha1(op *parsedOpcode, vm *Engine) er.R {
	buf, err := vm.dstack.PopByteArray()
	if err != nil {
		return err
	}

	hash := sha1.Sum(buf)
	vm.dstack.PushByteArray(hash[:])
	return nil
}

// opcodeSha256 treats the top item of the data stack as raw bytes and replaces
// it with sha256(data).
//
// Stack transformation: [... x1] -> [... sha256(x1)]
func opcodeSha256(op *parsedOpcode, vm *Engine) er.R {
	buf, err := vm.dstack.PopByteArray()
	if err != nil {
		return err
	}

	hash := sha256.Sum256(buf)
	vm.dstack.PushByteArray(hash[:])
	return nil
}

// opcodeHash160 treats the top item of the data stack as raw bytes and replaces
// it with ripemd160(sha256(data)).
//
// Stack transformation: [... x1] -> [... ripemd160(sha256(x1))]
func opcodeHash160(op *parsedOpcode, vm *Engine) er.R {
	buf, err := vm.dstack.PopByteArray()
	if err != nil {
		return err
	}

	vm.dstack.PushByteArray(btcutil.Hash160(buf))
	return nil
}

// opcodeHash256 treats the top item of the data stack as raw bytes and replaces
// it with sha256(sha256(data)).
//
// Stack transformation: [... x1] -> [... sha256(sha256(x1))]
func opcodeHash256(op *parsedOpcode, vm *Engine) er.R {
	buf, err := vm.dstack.PopByteArray()
	if err != nil {
		return err
	}

	vm.dstack.PushByteArray(chainhash.DoubleHashB(buf))
	return nil
}

// opcodeCodeSeparator stores the current script offset as the most recently
// seen OP_CODESEPARATOR which is used during signature checking.
//
// This opcode does not change the contents of the data stack.
func opcodeCodeSeparator(op *parsedOpcode, vm *Engine) er.R {
	vm.lastCodeSep = vm.scriptOff
	return nil
}

// opcodeCheckSig treats the top 2 items on the stack as a public key and a
// signature and replaces them with a bool which indicates if the signature was
// successfully verified.
//
// The process of verifying a signature requires calculating a signature hash in
// the same way the transaction signer did.  It involves hashing portions of the
// transaction based on the hash type byte (which is the final byte of the
// signature) and the portion of the script starting from the most recent
// OP_CODESEPARATOR (or the beginning of the script if there are none) to the
// end of the script (with any other OP_CODESEPARATORs removed).  Once this
// "script hash" is calculated, the signature is checked using standard
// cryptographic methods against the provided public key.
//
// Stack transformation: [... signature pubkey] -> [... bool]
func opcodeCheckSig(op *parsedOpcode, vm *Engine) er.R {
	pkBytes, err := vm.dstack.PopByteArray()
	if err != nil {
		return err
	}

	fullSigBytes, err := vm.dstack.PopByteArray()
	if err != nil {
		return err
	}

	// The signature actually needs needs to be longer than this, but at
	// least 1 byte is needed for the hash type below.  The full length is
	// checked depending on the script flags and upon parsing the signature.
	if len(fullSigBytes) < 1 {
		vm.dstack.PushBool(false)
		return nil
	}

	// Trim off hashtype from the signature string and check if the
	// signature and pubkey conform to the strict encoding requirements
	// depending on the flags.
	//
	// NOTE: When the strict encoding flags are set, any errors in the
	// signature or public encoding here result in an immediate script error
	// (and thus no result bool is pushed to the data stack).  This differs
	// from the logic below where any errors in parsing the signature is
	// treated as the signature failure resulting in false being pushed to
	// the data stack.  This is required because the more general script
	// validation consensus rules do not have the new strict encoding
	// requirements enabled by the flags.
	hashType := SigHashType(fullSigBytes[len(fullSigBytes)-1])
	sigBytes := fullSigBytes[:len(fullSigBytes)-1]
	if err := vm.checkHashTypeEncoding(hashType); err != nil {
		return err
	}
	if err := vm.checkSignatureEncoding(sigBytes); err != nil {
		return err
	}
	if err := vm.checkPubKeyEncoding(pkBytes); err != nil {
		return err
	}

	// Get script starting from the most recent OP_CODESEPARATOR.
	subScript := vm.subScript()

	// Generate the signature hash based on the signature hash type.
	var hash []byte
	if vm.isWitnessVersionActive(0) {
		var sigHashes *TxSigHashes
		if vm.hashCache != nil {
			sigHashes = vm.hashCache
		} else {
			sigHashes = NewTxSigHashes(&vm.tx)
		}

		hash, err = calcWitnessSignatureHash(subScript, sigHashes, hashType,
			&vm.tx, vm.txIdx, vm.inputAmount)
		if err != nil {
			return err
		}
	} else {
		// Remove the signature since there is no way for a signature
		// to sign itself.
		subScript = removeOpcodeByData(subScript, fullSigBytes)

		hash = calcSignatureHash(subScript, hashType, &vm.tx, vm.txIdx)
	}

	pubKey, err := btcec.ParsePubKey(pkBytes, btcec.S256())
	if err != nil {
		vm.dstack.PushBool(false)
		return nil
	}

	var signature *btcec.Signature
	if vm.hasFlag(ScriptVerifyStrictEncoding) ||
		vm.hasFlag(ScriptVerifyDERSignatures) {

		signature, err = btcec.ParseDERSignature(sigBytes, btcec.S256())
	} else {
		signature, err = btcec.ParseSignature(sigBytes, btcec.S256())
	}
	if err != nil {
		vm.dstack.PushBool(false)
		return nil
	}

	var valid bool
	if vm.sigCache != nil {
		var sigHash chainhash.Hash
		copy(sigHash[:], hash)

		valid = vm.sigCache.Exists(sigHash, signature, pubKey)
		if !valid && signature.Verify(hash, pubKey) {
			vm.sigCache.Add(sigHash, signature, pubKey)
			valid = true
		}
	} else {
		valid = signature.Verify(hash, pubKey)
	}

	if !valid && vm.hasFlag(ScriptVerifyNullFail) && len(sigBytes) > 0 {
		str := "signature not empty on failed checksig"
		return txscripterr.ScriptError(txscripterr.ErrNullFail, str)
	}

	vm.dstack.PushBool(valid)
	return nil
}

// opcodeCheckSigVerify is a combination of opcodeCheckSig and opcodeVerify.
// The opcodeCheckSig function is invoked followed by opcodeVerify.  See the
// documentation for each of those opcodes for more details.
//
// Stack transformation: signature pubkey] -> [... bool] -> [...]
func opcodeCheckSigVerify(op *parsedOpcode, vm *Engine) er.R {
	err := opcodeCheckSig(op, vm)
	if err == nil {
		err = abstractVerify(op, vm, txscripterr.ErrCheckSigVerify)
	}
	return err
}

// parsedSigInfo houses a raw signature along with its parsed form and a flag
// for whether or not it has already been parsed.  It is used to prevent parsing
// the same signature multiple times when verifying a multisig.
type parsedSigInfo struct {
	signature       []byte
	parsedSignature *btcec.Signature
	parsed          bool
}

// opcodeCheckMultiSig treats the top item on the stack as an integer number of
// public keys, followed by that many entries as raw data representing the public
// keys, followed by the integer number of signatures, followed by that many
// entries as raw data representing the signatures.
//
// Due to a bug in the original Satoshi client implementation, an additional
// dummy argument is also required by the consensus rules, although it is not
// used.  The dummy value SHOULD be an OP_0, although that is not required by
// the consensus rules.  When the ScriptStrictMultiSig flag is set, it must be
// OP_0.
//
// All of the aforementioned stack items are replaced with a bool which
// indicates if the requisite number of signatures were successfully verified.
//
// See the opcodeCheckSigVerify documentation for more details about the process
// for verifying each signature.
//
// Stack transformation:
// [... dummy [sig ...] numsigs [pubkey ...] numpubkeys] -> [... bool]
func opcodeCheckMultiSig(op *parsedOpcode, vm *Engine) er.R {
	numKeys, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}

	numPubKeys := int(numKeys.Int32())
	if numPubKeys < 0 {
		str := fmt.Sprintf("number of pubkeys %d is negative",
			numPubKeys)
		return txscripterr.ScriptError(txscripterr.ErrInvalidPubKeyCount, str)
	}
	if numPubKeys > MaxPubKeysPerMultiSig {
		str := fmt.Sprintf("too many pubkeys: %d > %d",
			numPubKeys, MaxPubKeysPerMultiSig)
		return txscripterr.ScriptError(txscripterr.ErrInvalidPubKeyCount, str)
	}
	vm.numOps += numPubKeys
	if vm.numOps > MaxOpsPerScript {
		str := fmt.Sprintf("exceeded max operation limit of %d",
			MaxOpsPerScript)
		return txscripterr.ScriptError(txscripterr.ErrTooManyOperations, str)
	}

	pubKeys := make([][]byte, 0, numPubKeys)
	for i := 0; i < numPubKeys; i++ {
		pubKey, err := vm.dstack.PopByteArray()
		if err != nil {
			return err
		}
		pubKeys = append(pubKeys, pubKey)
	}

	numSigs, err := vm.dstack.PopInt()
	if err != nil {
		return err
	}
	numSignatures := int(numSigs.Int32())
	if numSignatures < 0 {
		str := fmt.Sprintf("number of signatures %d is negative",
			numSignatures)
		return txscripterr.ScriptError(txscripterr.ErrInvalidSignatureCount, str)

	}
	if numSignatures > numPubKeys {
		str := fmt.Sprintf("more signatures than pubkeys: %d > %d",
			numSignatures, numPubKeys)
		return txscripterr.ScriptError(txscripterr.ErrInvalidSignatureCount, str)
	}

	signatures := make([]*parsedSigInfo, 0, numSignatures)
	for i := 0; i < numSignatures; i++ {
		signature, err := vm.dstack.PopByteArray()
		if err != nil {
			return err
		}
		sigInfo := &parsedSigInfo{signature: signature}
		signatures = append(signatures, sigInfo)
	}

	// A bug in the original Satoshi client implementation means one more
	// stack value than should be used must be popped.  Unfortunately, this
	// buggy behavior is now part of the consensus and a hard fork would be
	// required to fix it.
	dummy, err := vm.dstack.PopByteArray()
	if err != nil {
		return err
	}

	// Since the dummy argument is otherwise not checked, it could be any
	// value which unfortunately provides a source of malleability.  Thus,
	// there is a script flag to force an error when the value is NOT 0.
	if vm.hasFlag(ScriptStrictMultiSig) && len(dummy) != 0 {
		str := fmt.Sprintf("multisig dummy argument has length %d "+
			"instead of 0", len(dummy))
		return txscripterr.ScriptError(txscripterr.ErrSigNullDummy, str)
	}

	// Get script starting from the most recent OP_CODESEPARATOR.
	script := vm.subScript()

	// Remove the signature in pre version 0 segwit scripts since there is
	// no way for a signature to sign itself.
	if !vm.isWitnessVersionActive(0) {
		for _, sigInfo := range signatures {
			script = removeOpcodeByData(script, sigInfo.signature)
		}
	}

	success := true
	numPubKeys++
	pubKeyIdx := -1
	signatureIdx := 0
	for numSignatures > 0 {
		// When there are more signatures than public keys remaining,
		// there is no way to succeed since too many signatures are
		// invalid, so exit early.
		pubKeyIdx++
		numPubKeys--
		if numSignatures > numPubKeys {
			success = false
			break
		}

		sigInfo := signatures[signatureIdx]
		pubKey := pubKeys[pubKeyIdx]

		// The order of the signature and public key evaluation is
		// important here since it can be distinguished by an
		// OP_CHECKMULTISIG NOT when the strict encoding flag is set.

		rawSig := sigInfo.signature
		if len(rawSig) == 0 {
			// Skip to the next pubkey if signature is empty.
			continue
		}

		// Split the signature into hash type and signature components.
		hashType := SigHashType(rawSig[len(rawSig)-1])
		signature := rawSig[:len(rawSig)-1]

		// Only parse and check the signature encoding once.
		var parsedSig *btcec.Signature
		if !sigInfo.parsed {
			if err := vm.checkHashTypeEncoding(hashType); err != nil {
				return err
			}
			if err := vm.checkSignatureEncoding(signature); err != nil {
				return err
			}

			// Parse the signature.
			var err er.R
			if vm.hasFlag(ScriptVerifyStrictEncoding) ||
				vm.hasFlag(ScriptVerifyDERSignatures) {

				parsedSig, err = btcec.ParseDERSignature(signature,
					btcec.S256())
			} else {
				parsedSig, err = btcec.ParseSignature(signature,
					btcec.S256())
			}
			sigInfo.parsed = true
			if err != nil {
				continue
			}
			sigInfo.parsedSignature = parsedSig
		} else {
			// Skip to the next pubkey if the signature is invalid.
			if sigInfo.parsedSignature == nil {
				continue
			}

			// Use the already parsed signature.
			parsedSig = sigInfo.parsedSignature
		}

		if err := vm.checkPubKeyEncoding(pubKey); err != nil {
			return err
		}

		// Parse the pubkey.
		parsedPubKey, err := btcec.ParsePubKey(pubKey, btcec.S256())
		if err != nil {
			continue
		}

		// Generate the signature hash based on the signature hash type.
		var hash []byte
		if vm.isWitnessVersionActive(0) {
			var sigHashes *TxSigHashes
			if vm.hashCache != nil {
				sigHashes = vm.hashCache
			} else {
				sigHashes = NewTxSigHashes(&vm.tx)
			}

			hash, err = calcWitnessSignatureHash(script, sigHashes, hashType,
				&vm.tx, vm.txIdx, vm.inputAmount)
			if err != nil {
				return err
			}
		} else {
			hash = calcSignatureHash(script, hashType, &vm.tx, vm.txIdx)
		}

		var valid bool
		if vm.sigCache != nil {
			var sigHash chainhash.Hash
			copy(sigHash[:], hash)

			valid = vm.sigCache.Exists(sigHash, parsedSig, parsedPubKey)
			if !valid && parsedSig.Verify(hash, parsedPubKey) {
				vm.sigCache.Add(sigHash, parsedSig, parsedPubKey)
				valid = true
			}
		} else {
			valid = parsedSig.Verify(hash, parsedPubKey)
		}

		if valid {
			// PubKey verified, move on to the next signature.
			signatureIdx++
			numSignatures--
		}
	}

	if !success && vm.hasFlag(ScriptVerifyNullFail) {
		for _, sig := range signatures {
			if len(sig.signature) > 0 {
				str := "not all signatures empty on failed checkmultisig"
				return txscripterr.ScriptError(txscripterr.ErrNullFail, str)
			}
		}
	}

	vm.dstack.PushBool(success)
	return nil
}

// opcodeCheckMultiSigVerify is a combination of opcodeCheckMultiSig and
// opcodeVerify.  The opcodeCheckMultiSig is invoked followed by opcodeVerify.
// See the documentation for each of those opcodes for more details.
//
// Stack transformation:
// [... dummy [sig ...] numsigs [pubkey ...] numpubkeys] -> [... bool] -> [...]
func opcodeCheckMultiSigVerify(op *parsedOpcode, vm *Engine) er.R {
	err := opcodeCheckMultiSig(op, vm)
	if err == nil {
		err = abstractVerify(op, vm, txscripterr.ErrCheckMultiSigVerify)
	}
	return err
}

// OpcodeByName is a map that can be used to lookup an opcode by its
// human-readable name (OP_CHECKMULTISIG, OP_CHECKSIG, etc).
var OpcodeByName = make(map[string]byte)

func init() {
	// Initialize the opcode name to value map using the contents of the
	// opcode array.  Also add entries for "OP_FALSE", "OP_TRUE", and
	// "OP_NOP2" since they are aliases for "OP_0", "OP_1",
	// and "OP_CHECKLOCKTIMEVERIFY" respectively.
	for _, op := range opcodeArray {
		OpcodeByName[op.name] = op.value
	}
	OpcodeByName["OP_FALSE"] = opcode.OP_FALSE
	OpcodeByName["OP_TRUE"] = opcode.OP_TRUE
	OpcodeByName["OP_NOP2"] = opcode.OP_CHECKLOCKTIMEVERIFY
	OpcodeByName["OP_NOP3"] = opcode.OP_CHECKSEQUENCEVERIFY
}
