// Copyright (c) 2014-2016 The btcsuite developers
// Copyright (c) 2019 Caleb James DeLisle
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package chaincfg

import (
	"math"
	"math/big"
	"strings"
	"time"

	"github.com/pkt-cash/PKT-FullNode/blockchain/packetcrypt/difficulty"
	"github.com/pkt-cash/PKT-FullNode/btcutil/er"
	"github.com/pkt-cash/PKT-FullNode/wire/protocol"

	"github.com/pkt-cash/PKT-FullNode/chaincfg/chainhash"
	"github.com/pkt-cash/PKT-FullNode/chaincfg/globalcfg"
)

// These variables are the chain proof-of-work limit parameters for each default
// network.
var (
	// bigOne is 1 represented as a big.Int.  It is defined here to avoid
	// the overhead of creating it multiple times.
	bigOne = big.NewInt(1)

	// mainPowLimit is the highest proof of work value a Bitcoin block can
	// have for the main network.  It is the value 2^224 - 1.
	mainPowLimit = new(big.Int).Sub(new(big.Int).Lsh(bigOne, 224), bigOne)

	// regressionPowLimit is the highest proof of work value a Bitcoin block
	// can have for the regression test network.  It is the value 2^255 - 1.
	regressionPowLimit = new(big.Int).Sub(new(big.Int).Lsh(bigOne, 255), bigOne)

	// testNet3PowLimit is the highest proof of work value a Bitcoin block
	// can have for the test network (version 3).  It is the value
	// 2^224 - 1.
	testNet3PowLimit = new(big.Int).Sub(new(big.Int).Lsh(bigOne, 224), bigOne)

	// simNetPowLimit is the highest proof of work value a Bitcoin block
	// can have for the simulation test network.  It is the value 2^255 - 1.
	simNetPowLimit = new(big.Int).Sub(new(big.Int).Lsh(bigOne, 255), bigOne)
)

// Checkpoint identifies a known good point in the block chain.  Using
// checkpoints allows a few optimizations for old blocks during initial download
// and also prevents forks from old blocks.
//
// Each checkpoint is selected based upon several factors.  See the
// documentation for blockchain.IsCheckpointCandidate for details on the
// selection criteria.
type Checkpoint struct {
	Height int32
	Hash   *chainhash.Hash
}

// DNSSeed identifies a DNS seed.
type DNSSeed struct {
	// Host defines the hostname of the seed.
	Host string

	// HasFiltering defines whether the seed supports filtering
	// by service flags (wire.ServiceFlag).
	HasFiltering bool
}

// ConsensusDeployment defines details related to a specific consensus rule
// change that is voted in.  This is part of BIP0009.
type ConsensusDeployment struct {
	// BitNumber defines the specific bit number within the block version
	// this particular soft-fork deployment refers to.
	BitNumber uint8

	// StartTime is the median block time after which voting on the
	// deployment starts.
	StartTime uint64

	// ExpireTime is the median block time after which the attempted
	// deployment expires.
	ExpireTime uint64
}

// Constants that define the deployment offset in the deployments field of the
// parameters for each deployment.  This is useful to be able to get the details
// of a specific deployment by name.
const (
	// DeploymentTestDummy defines the rule change deployment ID for testing
	// purposes.
	DeploymentTestDummy = iota

	// DeploymentCSV defines the rule change deployment ID for the CSV
	// soft-fork package. The CSV package includes the deployment of BIPS
	// 68, 112, and 113.
	DeploymentCSV

	// DeploymentSegwit defines the rule change deployment ID for the
	// Segregated Witness (segwit) soft-fork package. The segwit package
	// includes the deployment of BIPS 141, 142, 144, 145, 147 and 173.
	DeploymentSegwit

	// NOTE: DefinedDeployments must always come last since it is used to
	// determine how many defined deployments there currently are.

	// DefinedDeployments is the number of currently defined deployments.
	DefinedDeployments
)

// Params defines a Bitcoin network by its parameters.  These parameters may be
// used by Bitcoin applications to differentiate networks as well as addresses
// and keys for one network from those intended for use on another network.
type Params struct {
	// Name defines a human-readable identifier for the network.
	Name string

	// Net defines the magic bytes used to identify the network.
	Net protocol.BitcoinNet

	// DefaultPort defines the default peer-to-peer port for the network.
	DefaultPort string

	// DNSSeeds defines a list of DNS seeds for the network that are used
	// as one method to discover peers.
	DNSSeeds []DNSSeed

	// The proof of work algorithm in use on this chain
	GlobalConf globalcfg.Config

	// The network network steward sig hash as of the first block
	InitialNetworkSteward []byte

	// GenesisHash is the starting block hash.
	GenesisHash *chainhash.Hash

	// PowLimit defines the highest allowed proof of work value for a block
	// as a uint256.
	PowLimit *big.Int

	// PowLimitBits defines the highest allowed proof of work value for a
	// block in compact form.
	PowLimitBits uint32

	// These fields define the block heights at which the specified softfork
	// BIP became active.
	BIP0034Height int32
	BIP0065Height int32
	BIP0066Height int32

	// CoinbaseMaturity is the number of blocks required before newly mined
	// coins (coinbase transactions) can be spent.
	CoinbaseMaturity uint16

	// SubsidyReductionInterval is the interval of blocks before the subsidy
	// is reduced.
	SubsidyReductionInterval int32

	// TargetTimespan is the desired amount of time that should elapse
	// before the block difficulty requirement is examined to determine how
	// it should be changed in order to maintain the desired block
	// generation rate.
	TargetTimespan time.Duration

	// TargetTimePerBlock is the desired amount of time to generate each
	// block.
	TargetTimePerBlock time.Duration

	// RetargetAdjustmentFactor is the adjustment factor used to limit
	// the minimum and maximum amount of adjustment that can occur between
	// difficulty retargets.
	RetargetAdjustmentFactor int64

	// ReduceMinDifficulty defines whether the network should reduce the
	// minimum required difficulty after a long enough period of time has
	// passed without finding a block.  This is really only useful for test
	// networks and should not be set on a main network.
	ReduceMinDifficulty bool

	// MinDiffReductionTime is the amount of time after which the minimum
	// required difficulty should be reduced when a block hasn't been found.
	//
	// NOTE: This only applies if ReduceMinDifficulty is true.
	MinDiffReductionTime time.Duration

	// GenerateSupported specifies whether or not CPU mining is allowed.
	GenerateSupported bool

	// Checkpoints ordered from oldest to newest.
	Checkpoints []Checkpoint

	// These fields are related to voting on consensus rule changes as
	// defined by BIP0009.
	//
	// RuleChangeActivationThreshold is the number of blocks in a threshold
	// state retarget window for which a positive vote for a rule change
	// must be cast in order to lock in a rule change. It should typically
	// be 95% for the main network and 75% for test networks.
	//
	// MinerConfirmationWindow is the number of blocks in each threshold
	// state retarget window.
	//
	// Deployments define the specific consensus rule changes to be voted
	// on.
	RuleChangeActivationThreshold uint32
	MinerConfirmationWindow       uint32
	Deployments                   [DefinedDeployments]ConsensusDeployment

	// Mempool parameters
	RelayNonStdTxs bool

	// Human-readable part for Bech32 encoded segwit addresses, as defined
	// in BIP 173.
	Bech32HRPSegwit string

	// Address encoding magics
	PubKeyHashAddrID        byte // First byte of a P2PKH address
	ScriptHashAddrID        byte // First byte of a P2SH address
	PrivateKeyID            byte // First byte of a WIF private key
	WitnessPubKeyHashAddrID byte // First byte of a P2WPKH address
	WitnessScriptHashAddrID byte // First byte of a P2WSH address

	// BIP32 hierarchical deterministic extended key magics
	HDPrivateKeyID [4]byte
	HDPublicKeyID  [4]byte

	// BIP44 coin type used in the hierarchical deterministic path for
	// address generation.
	HDCoinType uint32
}

// MainNetParams defines the network parameters for the main Bitcoin network.
var MainNetParams = Params{
	Name:        "mainnet",
	Net:         protocol.MainNet,
	DefaultPort: "8333",
	DNSSeeds: []DNSSeed{
		{"seed.bitcoin.sipa.be", true},
		{"dnsseed.bluematt.me", true},
		{"dnsseed.bitcoin.dashjr.org", false},
		{"seed.bitcoinstats.com", true},
		{"seed.bitnodes.io", false},
		{"seed.bitcoin.jonasschnelli.ch", true},
	},

	// Chain parameters
	GlobalConf:               globalcfg.BitcoinDefaults(),
	GenesisHash:              newHashFromStr("000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f"),
	PowLimit:                 mainPowLimit,
	PowLimitBits:             0x1d00ffff,
	BIP0034Height:            227931, // 000000000000024b89b42a942fe0d9fea3bb44ab7bd1b19115dd6a759c0808b8
	BIP0065Height:            388381, // 000000000000000004c2b624ed5d7756c508d90fd0da2c7c679febfa6c4735f0
	BIP0066Height:            363725, // 00000000000000000379eaa19dce8c9b722d46ae6a57c2f1a988119488b50931
	CoinbaseMaturity:         100,
	SubsidyReductionInterval: 210000,
	TargetTimespan:           time.Hour * 24 * 14, // 14 days
	TargetTimePerBlock:       time.Minute * 10,    // 10 minutes
	RetargetAdjustmentFactor: 4,                   // 25% less, 400% more
	ReduceMinDifficulty:      false,
	MinDiffReductionTime:     0,
	GenerateSupported:        false,

	// Checkpoints ordered from oldest to newest.
	Checkpoints: []Checkpoint{
		{11111, newHashFromStr("0000000069e244f73d78e8fd29ba2fd2ed618bd6fa2ee92559f542fdb26e7c1d")},
		{33333, newHashFromStr("000000002dd5588a74784eaa7ab0507a18ad16a236e7b1ce69f00d7ddfb5d0a6")},
		{74000, newHashFromStr("0000000000573993a3c9e41ce34471c079dcf5f52a0e824a81e7f953b8661a20")},
		{105000, newHashFromStr("00000000000291ce28027faea320c8d2b054b2e0fe44a773f3eefb151d6bdc97")},
		{134444, newHashFromStr("00000000000005b12ffd4cd315cd34ffd4a594f430ac814c91184a0d42d2b0fe")},
		{168000, newHashFromStr("000000000000099e61ea72015e79632f216fe6cb33d7899acb35b75c8303b763")},
		{193000, newHashFromStr("000000000000059f452a5f7340de6682a977387c17010ff6e6c3bd83ca8b1317")},
		{210000, newHashFromStr("000000000000048b95347e83192f69cf0366076336c639f9b7228e9ba171342e")},
		{216116, newHashFromStr("00000000000001b4f4b433e81ee46494af945cf96014816a4e2370f11b23df4e")},
		{225430, newHashFromStr("00000000000001c108384350f74090433e7fcf79a606b8e797f065b130575932")},
		{250000, newHashFromStr("000000000000003887df1f29024b06fc2200b55f8af8f35453d7be294df2d214")},
		{267300, newHashFromStr("000000000000000a83fbd660e918f218bf37edd92b748ad940483c7c116179ac")},
		{279000, newHashFromStr("0000000000000001ae8c72a0b0c301f67e3afca10e819efa9041e458e9bd7e40")},
		{300255, newHashFromStr("0000000000000000162804527c6e9b9f0563a280525f9d08c12041def0a0f3b2")},
		{319400, newHashFromStr("000000000000000021c6052e9becade189495d1c539aa37c58917305fd15f13b")},
		{343185, newHashFromStr("0000000000000000072b8bf361d01a6ba7d445dd024203fafc78768ed4368554")},
		{352940, newHashFromStr("000000000000000010755df42dba556bb72be6a32f3ce0b6941ce4430152c9ff")},
		{382320, newHashFromStr("00000000000000000a8dc6ed5b133d0eb2fd6af56203e4159789b092defd8ab2")},
		{400000, newHashFromStr("000000000000000004ec466ce4732fe6f1ed1cddc2ed4b328fff5224276e3f6f")},
		{430000, newHashFromStr("000000000000000001868b2bb3a285f3cc6b33ea234eb70facf4dcdf22186b87")},
		{460000, newHashFromStr("000000000000000000ef751bbce8e744ad303c47ece06c8d863e4d417efc258c")},
		{490000, newHashFromStr("000000000000000000de069137b17b8d5a3dfbd5b145b2dcfb203f15d0c4de90")},
		{520000, newHashFromStr("0000000000000000000d26984c0229c9f6962dc74db0a6d525f2f1640396f69c")},
		{550000, newHashFromStr("000000000000000000223b7a2298fb1c6c75fb0efc28a4c56853ff4112ec6bc9")},
		{560000, newHashFromStr("0000000000000000002c7b276daf6efb2b6aa68e2ce3be67ef925b3264ae7122")},
	},

	// Consensus rule change deployments.
	//
	// The miner confirmation window is defined as:
	//   target proof of work timespan / target proof of work spacing
	RuleChangeActivationThreshold: 1916, // 95% of MinerConfirmationWindow
	MinerConfirmationWindow:       2016, //
	Deployments: [DefinedDeployments]ConsensusDeployment{
		DeploymentTestDummy: {
			BitNumber:  28,
			StartTime:  1199145601, // January 1, 2008 UTC
			ExpireTime: 1230767999, // December 31, 2008 UTC
		},
		DeploymentCSV: {
			BitNumber:  0,
			StartTime:  1462060800, // May 1st, 2016
			ExpireTime: 1493596800, // May 1st, 2017
		},
		DeploymentSegwit: {
			BitNumber:  1,
			StartTime:  1479168000, // November 15, 2016 UTC
			ExpireTime: 1510704000, // November 15, 2017 UTC.
		},
	},

	// Mempool parameters
	RelayNonStdTxs: false,

	// Human-readable part for Bech32 encoded segwit addresses, as defined in
	// BIP 173.
	Bech32HRPSegwit: "bc", // always bc for main net

	// Address encoding magics
	PubKeyHashAddrID:        0x00, // starts with 1
	ScriptHashAddrID:        0x05, // starts with 3
	PrivateKeyID:            0x80, // starts with 5 (uncompressed) or K (compressed)
	WitnessPubKeyHashAddrID: 0x06, // starts with p2
	WitnessScriptHashAddrID: 0x0A, // starts with 7Xh

	// BIP32 hierarchical deterministic extended key magics
	HDPrivateKeyID: [4]byte{0x04, 0x88, 0xad, 0xe4}, // starts with xprv
	HDPublicKeyID:  [4]byte{0x04, 0x88, 0xb2, 0x1e}, // starts with xpub

	// BIP44 coin type used in the hierarchical deterministic path for
	// address generation.
	HDCoinType: 0,
}

// RegressionNetParams defines the network parameters for the regression test
// Bitcoin network.  Not to be confused with the test Bitcoin network (version
// 3), this network is sometimes simply called "testnet".
var RegressionNetParams = Params{
	Name:        "regtest",
	Net:         protocol.TestNet,
	DefaultPort: "18444",
	DNSSeeds:    []DNSSeed{},

	// Chain parameters
	GlobalConf:               globalcfg.BitcoinDefaults(),
	GenesisHash:              newHashFromStr("0f9188f13cb7b2c71f2a335e3a4fc328bf5beb436012afca590b1a11466e2206"),
	PowLimit:                 regressionPowLimit,
	PowLimitBits:             0x207fffff,
	CoinbaseMaturity:         100,
	BIP0034Height:            100000000, // Not active - Permit ver 1 blocks
	BIP0065Height:            1351,      // Used by regression tests
	BIP0066Height:            1251,      // Used by regression tests
	SubsidyReductionInterval: 150,
	TargetTimespan:           time.Hour * 24 * 14, // 14 days
	TargetTimePerBlock:       time.Minute * 10,    // 10 minutes
	RetargetAdjustmentFactor: 4,                   // 25% less, 400% more
	ReduceMinDifficulty:      true,
	MinDiffReductionTime:     time.Minute * 20, // TargetTimePerBlock * 2
	GenerateSupported:        true,

	// Checkpoints ordered from oldest to newest.
	Checkpoints: nil,

	// Consensus rule change deployments.
	//
	// The miner confirmation window is defined as:
	//   target proof of work timespan / target proof of work spacing
	RuleChangeActivationThreshold: 108, // 75%  of MinerConfirmationWindow
	MinerConfirmationWindow:       144,
	Deployments: [DefinedDeployments]ConsensusDeployment{
		DeploymentTestDummy: {
			BitNumber:  28,
			StartTime:  0,             // Always available for vote
			ExpireTime: math.MaxInt64, // Never expires
		},
		DeploymentCSV: {
			BitNumber:  0,
			StartTime:  0,             // Always available for vote
			ExpireTime: math.MaxInt64, // Never expires
		},
		DeploymentSegwit: {
			BitNumber:  1,
			StartTime:  0,             // Always available for vote
			ExpireTime: math.MaxInt64, // Never expires.
		},
	},

	// Mempool parameters
	RelayNonStdTxs: true,

	// Human-readable part for Bech32 encoded segwit addresses, as defined in
	// BIP 173.
	Bech32HRPSegwit: "bcrt", // always bcrt for reg test net

	// Address encoding magics
	PubKeyHashAddrID: 0x6f, // starts with m or n
	ScriptHashAddrID: 0xc4, // starts with 2
	PrivateKeyID:     0xef, // starts with 9 (uncompressed) or c (compressed)

	// BIP32 hierarchical deterministic extended key magics
	HDPrivateKeyID: [4]byte{0x04, 0x35, 0x83, 0x94}, // starts with tprv
	HDPublicKeyID:  [4]byte{0x04, 0x35, 0x87, 0xcf}, // starts with tpub

	// BIP44 coin type used in the hierarchical deterministic path for
	// address generation.
	HDCoinType: 1,
}

// TestNet3Params defines the network parameters for the test Bitcoin network
// (version 3).  Not to be confused with the regression test network, this
// network is sometimes simply called "testnet".
var TestNet3Params = Params{
	Name:        "testnet3",
	Net:         protocol.TestNet3,
	DefaultPort: "18333",
	DNSSeeds: []DNSSeed{
		{"testnet-seed.bitcoin.jonasschnelli.ch", true},
		{"testnet-seed.bitcoin.schildbach.de", false},
		{"seed.tbtc.petertodd.org", true},
		{"testnet-seed.bluematt.me", false},
	},

	// Chain parameters
	GlobalConf:               globalcfg.BitcoinDefaults(),
	GenesisHash:              newHashFromStr("000000000933ea01ad0ee984209779baaec3ced90fa3f408719526f8d77f4943"),
	PowLimit:                 testNet3PowLimit,
	PowLimitBits:             0x1d00ffff,
	BIP0034Height:            21111,  // 0000000023b3a96d3484e5abb3755c413e7d41500f8e2a5c3f0dd01299cd8ef8
	BIP0065Height:            581885, // 00000000007f6655f22f98e72ed80d8b06dc761d5da09df0fa1dc4be4f861eb6
	BIP0066Height:            330776, // 000000002104c8c45e99a8853285a3b592602a3ccde2b832481da85e9e4ba182
	CoinbaseMaturity:         100,
	SubsidyReductionInterval: 210000,
	TargetTimespan:           time.Hour * 24 * 14, // 14 days
	TargetTimePerBlock:       time.Minute * 10,    // 10 minutes
	RetargetAdjustmentFactor: 4,                   // 25% less, 400% more
	ReduceMinDifficulty:      true,
	MinDiffReductionTime:     time.Minute * 20, // TargetTimePerBlock * 2
	GenerateSupported:        false,

	// Checkpoints ordered from oldest to newest.
	Checkpoints: []Checkpoint{
		{546, newHashFromStr("000000002a936ca763904c3c35fce2f3556c559c0214345d31b1bcebf76acb70")},
		{100000, newHashFromStr("00000000009e2958c15ff9290d571bf9459e93b19765c6801ddeccadbb160a1e")},
		{200000, newHashFromStr("0000000000287bffd321963ef05feab753ebe274e1d78b2fd4e2bfe9ad3aa6f2")},
		{300001, newHashFromStr("0000000000004829474748f3d1bc8fcf893c88be255e6d7f571c548aff57abf4")},
		{400002, newHashFromStr("0000000005e2c73b8ecb82ae2dbc2e8274614ebad7172b53528aba7501f5a089")},
		{500011, newHashFromStr("00000000000929f63977fbac92ff570a9bd9e7715401ee96f2848f7b07750b02")},
		{600002, newHashFromStr("000000000001f471389afd6ee94dcace5ccc44adc18e8bff402443f034b07240")},
		{700000, newHashFromStr("000000000000406178b12a4dea3b27e13b3c4fe4510994fd667d7c1e6a3f4dc1")},
		{800010, newHashFromStr("000000000017ed35296433190b6829db01e657d80631d43f5983fa403bfdb4c1")},
		{900000, newHashFromStr("0000000000356f8d8924556e765b7a94aaebc6b5c8685dcfa2b1ee8b41acd89b")},
		{1000007, newHashFromStr("00000000001ccb893d8a1f25b70ad173ce955e5f50124261bbbc50379a612ddf")},
		{1100007, newHashFromStr("00000000000abc7b2cd18768ab3dee20857326a818d1946ed6796f42d66dd1e8")},
		{1200007, newHashFromStr("00000000000004f2dc41845771909db57e04191714ed8c963f7e56713a7b6cea")},
		{1300007, newHashFromStr("0000000072eab69d54df75107c052b26b0395b44f77578184293bf1bb1dbd9fa")},
	},

	// Consensus rule change deployments.
	//
	// The miner confirmation window is defined as:
	//   target proof of work timespan / target proof of work spacing
	RuleChangeActivationThreshold: 1512, // 75% of MinerConfirmationWindow
	MinerConfirmationWindow:       2016,
	Deployments: [DefinedDeployments]ConsensusDeployment{
		DeploymentTestDummy: {
			BitNumber:  28,
			StartTime:  1199145601, // January 1, 2008 UTC
			ExpireTime: 1230767999, // December 31, 2008 UTC
		},
		DeploymentCSV: {
			BitNumber:  0,
			StartTime:  1456790400, // March 1st, 2016
			ExpireTime: 1493596800, // May 1st, 2017
		},
		DeploymentSegwit: {
			BitNumber:  1,
			StartTime:  1462060800, // May 1, 2016 UTC
			ExpireTime: 1493596800, // May 1, 2017 UTC.
		},
	},

	// Mempool parameters
	RelayNonStdTxs: true,

	// Human-readable part for Bech32 encoded segwit addresses, as defined in
	// BIP 173.
	Bech32HRPSegwit: "tb", // always tb for test net

	// Address encoding magics
	PubKeyHashAddrID:        0x6f, // starts with m or n
	ScriptHashAddrID:        0xc4, // starts with 2
	WitnessPubKeyHashAddrID: 0x03, // starts with QW
	WitnessScriptHashAddrID: 0x28, // starts with T7n
	PrivateKeyID:            0xef, // starts with 9 (uncompressed) or c (compressed)

	// BIP32 hierarchical deterministic extended key magics
	HDPrivateKeyID: [4]byte{0x04, 0x35, 0x83, 0x94}, // starts with tprv
	HDPublicKeyID:  [4]byte{0x04, 0x35, 0x87, 0xcf}, // starts with tpub

	// BIP44 coin type used in the hierarchical deterministic path for
	// address generation.
	HDCoinType: 1,
}

// PktTestNetParams defines the network parameters for the test pkt.cash network
// (version 1).  Not to be confused with the regression test network, this
// network is sometimes simply called "testnet".
var PktTestNetParams = Params{
	Name:        "pkttest",
	Net:         protocol.PktTestNet,
	DefaultPort: "64512",
	DNSSeeds: []DNSSeed{
		{"testseed.cjd.li", false},
		{"testseed.anode.co", false},
	},

	// Chain parameters
	GlobalConf: globalcfg.PktDefaults(),
	InitialNetworkSteward: []byte{
		0x00, 0x20, 0xd5, 0xc1, 0x00, 0x5c, 0x0d, 0x40,
		0x12, 0xd3, 0xae, 0x26, 0x72, 0x31, 0x9e, 0x7f,
		0x9e, 0xb1, 0x5a, 0x57, 0x51, 0x6a, 0xee, 0xfa,
		0xbb, 0xbc, 0x06, 0x22, 0x65, 0xf6, 0x7e, 0x30,
		0x8f, 0x2b,
	},
	GenesisHash:              newHashFromStr("0bdc1712a46194e552cf417ab0439c2d4f456c35cf63a0a406964c6f93432d85"),
	PowLimitBits:             0x1f0fffff,
	BIP0034Height:            0,
	BIP0065Height:            0,
	BIP0066Height:            0,
	CoinbaseMaturity:         100,
	SubsidyReductionInterval: 2100000,
	TargetTimespan:           (time.Hour * 24 * 14) / 10, // 14 days
	TargetTimePerBlock:       time.Minute,                // 1 minute
	RetargetAdjustmentFactor: 4,                          // 25% less, 400% more
	ReduceMinDifficulty:      true,
	MinDiffReductionTime:     time.Minute * 2, // TargetTimePerBlock * 2
	GenerateSupported:        false,

	// Checkpoints ordered from oldest to newest.
	Checkpoints: []Checkpoint{},

	// Consensus rule change deployments.
	//
	// The miner confirmation window is defined as:
	//   target proof of work timespan / target proof of work spacing
	RuleChangeActivationThreshold: 1512, // 75% of MinerConfirmationWindow
	MinerConfirmationWindow:       2016,
	Deployments: [DefinedDeployments]ConsensusDeployment{
		DeploymentTestDummy: {
			BitNumber:  28,
			StartTime:  1199145601, // January 1, 2008 UTC
			ExpireTime: 1230767999, // December 31, 2008 UTC
		},
		DeploymentCSV: {
			StartTime:  math.MaxInt64,
			ExpireTime: math.MaxInt64,
		},
		DeploymentSegwit: {
			StartTime:  math.MaxInt64,
			ExpireTime: math.MaxInt64,
		},
	},

	// Mempool parameters
	RelayNonStdTxs: true,

	// Human-readable part for Bech32 encoded segwit addresses, as defined in
	// BIP 173.
	// https://github.com/satoshilabs/slips/blob/master/slip-0173.md
	Bech32HRPSegwit: "tpk",

	// Address encoding magics
	PubKeyHashAddrID:        0x6f, // starts with m or n
	ScriptHashAddrID:        0xc4, // starts with 2
	WitnessPubKeyHashAddrID: 0x03, // starts with QW
	WitnessScriptHashAddrID: 0x28, // starts with T7n
	PrivateKeyID:            0xef, // starts with 9 (uncompressed) or c (compressed)

	// BIP32 hierarchical deterministic extended key magics
	HDPrivateKeyID: [4]byte{0x04, 0x35, 0x83, 0x94}, // starts with tprv
	HDPublicKeyID:  [4]byte{0x04, 0x35, 0x87, 0xcf}, // starts with tpub

	// BIP44 coin type used in the hierarchical deterministic path for
	// address generation.
	HDCoinType: 1,
}

// PktMainNetParams defines the network parameters for the pkt.cash network.
var PktMainNetParams = Params{
	Name:        "pkt",
	Net:         protocol.PktMainNet,
	DefaultPort: "64764",
	DNSSeeds: []DNSSeed{
		{"seed.cjd.li", false},
		{"pktdseed.pkt.world", false},
		{"seed.pkt.pink", false},
		{"seed.pkt.ai", false},
	},

	// Chain parameters
	GlobalConf: globalcfg.PktDefaults(),
	InitialNetworkSteward: []byte{
		0x00, 0x20, 0xd5, 0xc1, 0x00, 0x5c, 0x0d, 0x40,
		0x12, 0xd3, 0xae, 0x26, 0x72, 0x31, 0x9e, 0x7f,
		0x9e, 0xb1, 0x5a, 0x57, 0x51, 0x6a, 0xee, 0xfa,
		0xbb, 0xbc, 0x06, 0x22, 0x65, 0xf6, 0x7e, 0x30,
		0x8f, 0x2b,
	},
	GenesisHash:              newHashFromStr("0bdc1712a46194e552cf417ab0439c2d4f456c35cf63a0a406964c6f93432d85"),
	PowLimitBits:             0x1f0fffff,
	PowLimit:                 difficulty.CompactToBig(0x1f0fffff),
	BIP0034Height:            0,
	BIP0065Height:            0,
	BIP0066Height:            0,
	CoinbaseMaturity:         100,
	SubsidyReductionInterval: -1,
	TargetTimespan:           (time.Hour * 24 * 14) / 10, // 1.4 days
	TargetTimePerBlock:       time.Minute,                // 1 minute
	RetargetAdjustmentFactor: 4,                          // 25% less, 400% more
	ReduceMinDifficulty:      false,
	MinDiffReductionTime:     time.Minute * 2, // TargetTimePerBlock * 2
	GenerateSupported:        false,

	// Checkpoints ordered from oldest to newest.
	Checkpoints: []Checkpoint{
		{1 << 15, newHashFromStr("9e2f78040c97b244fcbbf9dd528e02f3aaa54705d04408e3079ee98f427cb1fa")},
		{2 << 15, newHashFromStr("26e5ca803a949a5ee09926a1042408e9c1ca0cfc4950c134791fa5eb2624183c")},
		{3 << 15, newHashFromStr("8745c74b6a105b0173c076adc2728ffae48e6e1c8219e9ab7588e020a10e07cb")},
		{4 << 15, newHashFromStr("7194f5f6210489ecd6bd0feb84fc72c00a4b4aed33e6e2da5a92faec94d56879")},
		{5 << 15, newHashFromStr("3c8a1d44c35bdcde485e8e0238593c4228f40643a89ab02c8766dc837c24ae9d")},
		{6 << 15, newHashFromStr("594567c710572b7ee01a3013d54efb4ca1a283d05ebbc24d206decb77aeef942")},
		{7 << 15, newHashFromStr("28b4cdf10297b018acb97562dd985c7d5bdbdbb6121768e1cb9c9c7910193c41")},
		{8 << 15, newHashFromStr("1d2d6ece172b67878cdc308aa9d2d1cb684d60f663fd820dd24aaaebc92e1450")},
		{9 << 15, newHashFromStr("46ac9323f6a3dd96e69f9f49b26844054265d96265d8f5f7b45ad23ac042ca0f")},
		{10 << 15, newHashFromStr("abaff3c0e48b822aec3fa4b3e40ddc3ac7121e1efa96c6883bb7217b49df3d2e")},
		{11 << 15, newHashFromStr("c2ddcc7d981bca4a924e82b3dd54a05fd5f4aa72d9731e01b8a7ee570f128831")},
		{12 << 15, newHashFromStr("a55d3d59fab43d8c28784dbaee31d3710be3fef2cab0fa02c77079bc88d62c95")},
		{13 << 15, newHashFromStr("64ae244ff7ef2d6e907cb0cb953abda21087100e4450fc57e1166bf6961b7a04")},
		{14 << 15, newHashFromStr("3a0668eaa7a1f89e7b5c8c87805f25877fe0e48ff0d28039fb34dc0cdee664f7")},
		{15 << 15, newHashFromStr("74fb58760a49b5469aabc2b193d020d781ea170e4619e605d9dfd7214fa04f12")},
		{16 << 15, newHashFromStr("4acf5a9646f521bb00f568036a4e739e7ecc10fdc12edd8146ab96e00666e0b8")},
		{17 << 15, newHashFromStr("4dd9226548ddef3193c97da7ecbd1986a9d4d3a77a9fddbb38cc729c669e0e1a")},
		{18 << 15, newHashFromStr("b8d2b39e2b6e50e36fa3f3c7890ae27fe6f5363d219dfec7c88ce50cc3466a4c")},
		{19 << 15, newHashFromStr("e4b42f977d75346c1fd89cbf00feebeff3814a8d29dc96c0b137b6ede60431b6")},
		{20 << 15, newHashFromStr("38c7905f4dacbd8101614f19c2a86b5344e993483b3ba1bfa84ef1124e543457")},
		{21 << 15, newHashFromStr("91cb278fa7c0069d42e25785a78429bdbaceb19f753abdd63bff394dd2f04337")},
		{22 << 15, newHashFromStr("fd089c041c2cd79b9f10239310f2f28e2157f7a0c2015b7f696ea35cf6b20082")},
		{23 << 15, newHashFromStr("a4e1d2228a7e8acd2a6317389c70c09d475e334b3ad934d571d5bcb0253c03ab")},
		{24 << 15, newHashFromStr("7ae9e5f968d75dc6b3dcb78a310dfc0628c3a2b3d209826ae8b247bfc3e718c6")},
		{25 << 15, newHashFromStr("800f3cf026c650bd60b039f1d374a613c46eb283fe842b11c20c096af4a05174")},
		{26 << 15, newHashFromStr("502b930a3df5456b822b1aea93df9ff04d4c506efb1243880882bdd391ff6e74")},
		{27 << 15, newHashFromStr("96e707504ec446f8db5537f9570e097580f73776307714c984300433592f6c01")},
		{28 << 15, newHashFromStr("36b4ec111b3552309229b9e9346a926bd7575eaa99a6d1560d675d0de1cbd153")},
		{29 << 15, newHashFromStr("62c6b30afdf16b9bcf37f28f2cac5f2ced02dd29e523561ce244897e2f02cfa9")},
		{30 << 15, newHashFromStr("f201f7bd254441c96d2a6646ab2ab17aaa44813b4d777739e958b3682aa073da")},
		{31 << 15, newHashFromStr("31f6876a24b92b709b73e888a49565611b5fffd4a7b3c7d9e57567229f28e327")},
		{32 << 15, newHashFromStr("7c498d0ce0da699f812163a21732f23735f6ae7b907da533fc885dbb60b49d90")},
		{33 << 15, newHashFromStr("be78a867e38f605092ff792164e1a609714c0f439f0343f8fd7503c593eacf39")},
		{34 << 15, newHashFromStr("7e04e0dc8f0b7cc736884d78d5d768fd1c981e50a6aedacc05458ec274178d8c")},
		{35 << 15, newHashFromStr("fb2d9b65cd23f53ec0ff9102881f20f599ace34f9e7a2a6767a2f6127241bf5e")},
		{36 << 15, newHashFromStr("fbfa6217f048698b0853c48232510693742edd93fca9929fe0d3e2c4459ecf4b")},
		{37 << 15, newHashFromStr("ca081e2a8af161284e4436c591d1fdf75362e7ab44c6a9a42d22b3fc435ddfb0")},
		{38 << 15, newHashFromStr("2da4585a5f97bf8d9d7ce24de51b918666c48096d02e3ecaeb2c16a04b6938be")},
		{39 << 15, newHashFromStr("a468deb9c3f71ae85d5132dc2c18ffe1596e3aaeaabefc78274d29222589d9f0")},
		{40 << 15, newHashFromStr("e7e0d190fedf330131673e3a7c554fae17324f0a335e8c1c6ed30b9f8143ea25")},
		{41 << 15, newHashFromStr("caa62452841380e8d868e432d2d91765a83d2180a2e2ac14b154ca608df980d7")},
		{42 << 15, newHashFromStr("7c412619ed1c3b516d20c1d74b4b31060e1650d5e05ec9865aa32e1a87ca6817")},
		{43 << 15, newHashFromStr("09fbe7718260a4532e71c0ee385a8dac63c9ade79229e323123813644183dee3")},
		{44 << 15, newHashFromStr("513a45eb329e46a0a8bf1ef0ce8be43173295abb26b68651888bb6d7350c5510")},
		{45 << 15, newHashFromStr("2be0c1eb2bed850b75bc53c879d222abb78054e1a7c80dcf7619eaec4326a09f")},
		{46 << 15, newHashFromStr("eae42cd7a3e7d5c086b325a24992961b4b7f78ac96174d9f39edebf7ecc0e55d")},
		{47 << 15, newHashFromStr("361c5d7a76171e9caff6451d822c97256d68b7a87733bc95aa9cbe3aa3fd51ac")},
		{48 << 15, newHashFromStr("2ede3892f6ff4874fbd4b98e4bcf3d4b1d72ee4dd43d6b210a84c73f0f3194cd")},
		{49 << 15, newHashFromStr("d22f510e2494b97205e3e0c3f8340ed43b591d463789d2317c3043990ec42d7d")},
		{50 << 15, newHashFromStr("92d029520cc265d87ff5e0c057d456d49dfbce91c46381da62c21abd5150ace6")},
		{51 << 15, newHashFromStr("b7425861302ce3c984ae9fb9e264c001e126a30e2b5379fdc38605e1aeaec2b1")},
		{52 << 15, newHashFromStr("7f4460239c77affe0e134b0a62ddeeb575ae2cd2b449f81c409380b1ea383f50")},
		{53 << 15, newHashFromStr("418f787b9368fecb9d22378b5f75dd9e822487ee2e0f4f4831bb53273476b2f8")},
		{54 << 15, newHashFromStr("1f65bab0bb4e6eda0525d9055e87c1b143bdd14a0d09c3281d4b046ccf55ba6e")},
		{55 << 15, newHashFromStr("b9387767df9baf17f3c1ad99ed4fb3613b8bd863d4dcad70cbc7f967c335011e")},
		{56 << 15, newHashFromStr("dcb6b8d1a85fa89b494fefe3dd738a8caad452e0ec09527fc2ab38c7901912e7")},
		{57 << 15, newHashFromStr("85b45132326b3288abc29df1bf549679aa38731038abfc2e3be0725c35d17c5b")},
		{58 << 15, newHashFromStr("6ecfecdfb9c87332f8bbba8e5b5e2ac2b4dc741798e59f47481dc64e5d1c2f72")},
		{59 << 15, newHashFromStr("5ec4eef148936658209dc004362c051f2831195e14a1582423d7e2b82164a8d2")},
		{60 << 15, newHashFromStr("028c2c38497130cad7a02eb0386270d9fc9cee2369171628b18d73dc0c40ef01")},
		{61 << 15, newHashFromStr("a191a49bb963f4cbffabcded327c0bb8befcf9831be66460f1c250ca1fbec67e")},
		{62 << 15, newHashFromStr("81a7dd047af88a512c10b3e8b2159aedc7a047cf07a428762744539526c436e2")},
		{63 << 15, newHashFromStr("9e96647591bfddc8ec03aa1f9f885fc17fc56b319ad48d18079daffb5f689bab")},
		{64 << 15, newHashFromStr("f83dc987e6b081e403956d4a0faabe67a291aa8025a8d372749fee3cb7184b6b")},
		{65 << 15, newHashFromStr("689ef27b79fa2841d9cdc6a873d5c4ea77ae0488af33195f4bd213f391a0d135")},
		{66 << 15, newHashFromStr("b4b9d23e9acbfacc2b62e7abd28bcde97b419b86e9454731df1ec931a6272d2d")},
		{67 << 15, newHashFromStr("c39c8a3b140a3dc04ad6c824b0184aa95cc6fed6136deb5b0a1ee8a37f1e0418")},
		{68 << 15, newHashFromStr("b64073ea002446c0dbed664f6ba05b9801282669a43cf1261fc923c67eb7c860")},
		{69 << 15, newHashFromStr("3189252627c63e47790764157734d6badff2a628a099fea4b793c51adda4a114")},
	},

	// Consensus rule change deployments.
	//
	// The miner confirmation window is defined as:
	//   target proof of work timespan / target proof of work spacing
	RuleChangeActivationThreshold: 1512, // 75% of MinerConfirmationWindow
	MinerConfirmationWindow:       2016,
	Deployments: [DefinedDeployments]ConsensusDeployment{
		DeploymentTestDummy: {
			BitNumber:  28,
			StartTime:  1199145601, // January 1, 2008 UTC
			ExpireTime: 1230767999, // December 31, 2008 UTC
		},
		DeploymentCSV: {
			StartTime:  math.MaxInt64,
			ExpireTime: math.MaxInt64,
		},
		DeploymentSegwit: {
			StartTime:  math.MaxInt64,
			ExpireTime: math.MaxInt64,
		},
	},

	// Mempool parameters
	RelayNonStdTxs: false,

	// Human-readable part for Bech32 encoded segwit addresses, as defined in
	// BIP 173.
	// https://github.com/satoshilabs/slips/blob/master/slip-0173.md
	Bech32HRPSegwit: "pkt",

	// Address encoding magics
	PubKeyHashAddrID:        0x75, // starts with p
	ScriptHashAddrID:        0x38, // starts with P
	WitnessPubKeyHashAddrID: 0xa3, // starts with P2
	WitnessScriptHashAddrID: 0x22, // starts with PCZ
	PrivateKeyID:            0xe0, // starts with 8 (uncompressed) or a (compressed)

	// BIP32 hierarchical deterministic extended key magics
	HDPrivateKeyID: [4]byte{0x6b, 0x86, 0x3b, 0xed}, // starts with PpvtX
	HDPublicKeyID:  [4]byte{0x6b, 0x85, 0xc5, 0x3f}, // starts with PpubX

	// BIP44 coin type used in the hierarchical deterministic path for
	// address generation.
	// https://github.com/satoshilabs/slips/blob/master/slip-0044.md
	HDCoinType: 390,
}

// SimNetParams defines the network parameters for the simulation test Bitcoin
// network.  This network is similar to the normal test network except it is
// intended for private use within a group of individuals doing simulation
// testing.  The functionality is intended to differ in that the only nodes
// which are specifically specified are used to create the network rather than
// following normal discovery rules.  This is important as otherwise it would
// just turn into another public testnet.
var SimNetParams = Params{
	Name:        "simnet",
	Net:         protocol.SimNet,
	DefaultPort: "18555",
	DNSSeeds:    []DNSSeed{}, // NOTE: There must NOT be any seeds.

	// Chain parameters
	GlobalConf:               globalcfg.BitcoinDefaults(),
	GenesisHash:              newHashFromStr("683e86bd5c6d110d91b94b97137ba6bfe02dbbdb8e3dff722a669b5d69d77af6"),
	PowLimit:                 simNetPowLimit,
	PowLimitBits:             0x207fffff,
	BIP0034Height:            0, // Always active on simnet
	BIP0065Height:            0, // Always active on simnet
	BIP0066Height:            0, // Always active on simnet
	CoinbaseMaturity:         100,
	SubsidyReductionInterval: 210000,
	TargetTimespan:           time.Hour * 24 * 14, // 14 days
	TargetTimePerBlock:       time.Minute * 10,    // 10 minutes
	RetargetAdjustmentFactor: 4,                   // 25% less, 400% more
	ReduceMinDifficulty:      true,
	MinDiffReductionTime:     time.Minute * 20, // TargetTimePerBlock * 2
	GenerateSupported:        true,

	// Checkpoints ordered from oldest to newest.
	Checkpoints: nil,

	// Consensus rule change deployments.
	//
	// The miner confirmation window is defined as:
	//   target proof of work timespan / target proof of work spacing
	RuleChangeActivationThreshold: 75, // 75% of MinerConfirmationWindow
	MinerConfirmationWindow:       100,
	Deployments: [DefinedDeployments]ConsensusDeployment{
		DeploymentTestDummy: {
			BitNumber:  28,
			StartTime:  0,             // Always available for vote
			ExpireTime: math.MaxInt64, // Never expires
		},
		DeploymentCSV: {
			BitNumber:  0,
			StartTime:  0,             // Always available for vote
			ExpireTime: math.MaxInt64, // Never expires
		},
		DeploymentSegwit: {
			BitNumber:  1,
			StartTime:  0,             // Always available for vote
			ExpireTime: math.MaxInt64, // Never expires.
		},
	},

	// Mempool parameters
	RelayNonStdTxs: true,

	// Human-readable part for Bech32 encoded segwit addresses, as defined in
	// BIP 173.
	Bech32HRPSegwit: "sb", // always sb for sim net

	// Address encoding magics
	PubKeyHashAddrID:        0x3f, // starts with S
	ScriptHashAddrID:        0x7b, // starts with s
	PrivateKeyID:            0x64, // starts with 4 (uncompressed) or F (compressed)
	WitnessPubKeyHashAddrID: 0x19, // starts with Gg
	WitnessScriptHashAddrID: 0x28, // starts with ?

	// BIP32 hierarchical deterministic extended key magics
	HDPrivateKeyID: [4]byte{0x04, 0x20, 0xb9, 0x00}, // starts with sprv
	HDPublicKeyID:  [4]byte{0x04, 0x20, 0xbd, 0x3a}, // starts with spub

	// BIP44 coin type used in the hierarchical deterministic path for
	// address generation.
	HDCoinType: 115, // ASCII for s
}

var (
	// ErrDuplicateNet describes an error where the parameters for a Bitcoin
	// network could not be set due to the network already being a standard
	// network or previously-registered into this package.
	ErrDuplicateNet = er.GenericErrorType.CodeWithDetail("ErrDuplicateNet",
		"duplicate Bitcoin network")

	// ErrUnknownHDKeyID describes an error where the provided id which
	// is intended to identify the network for a hierarchical deterministic
	// private extended key is not registered.
	ErrUnknownHDKeyID = er.GenericErrorType.CodeWithDetail("ErrUnknownHDKeyID",
		"unknown hd private extended key bytes")
)

var (
	registeredNets       = make(map[protocol.BitcoinNet]struct{})
	pubKeyHashAddrIDs    = make(map[byte]struct{})
	scriptHashAddrIDs    = make(map[byte]struct{})
	bech32SegwitPrefixes = make(map[string]struct{})
	hdPrivToPubKeyIDs    = make(map[[4]byte][]byte)
)

// String returns the hostname of the DNS seed in human-readable form.
func (d DNSSeed) String() string {
	return d.Host
}

// Register registers the network parameters for a Bitcoin network.  This may
// error with ErrDuplicateNet if the network is already registered (either
// due to a previous Register call, or the network being one of the default
// networks).
//
// Network parameters should be registered into this package by a main package
// as early as possible.  Then, library packages may lookup networks or network
// parameters based on inputs and work regardless of the network being standard
// or not.
func Register(params *Params) er.R {
	if _, ok := registeredNets[params.Net]; ok {
		return ErrDuplicateNet.Default()
	}
	registeredNets[params.Net] = struct{}{}
	pubKeyHashAddrIDs[params.PubKeyHashAddrID] = struct{}{}
	scriptHashAddrIDs[params.ScriptHashAddrID] = struct{}{}
	hdPrivToPubKeyIDs[params.HDPrivateKeyID] = params.HDPublicKeyID[:]

	// A valid Bech32 encoded segwit address always has as prefix the
	// human-readable part for the given net followed by '1'.
	bech32SegwitPrefixes[params.Bech32HRPSegwit+"1"] = struct{}{}
	return nil
}

// mustRegister performs the same function as Register except it panics if there
// is an error.  This should only be called from package init functions.
func mustRegister(params *Params) {
	if err := Register(params); err != nil {
		panic("failed to register network: " + err.String())
	}
}

// IsPubKeyHashAddrID returns whether the id is an identifier known to prefix a
// pay-to-pubkey-hash address on any default or registered network.  This is
// used when decoding an address string into a specific address type.  It is up
// to the caller to check both this and IsScriptHashAddrID and decide whether an
// address is a pubkey hash address, script hash address, neither, or
// undeterminable (if both return true).
func IsPubKeyHashAddrID(id byte) bool {
	_, ok := pubKeyHashAddrIDs[id]
	return ok
}

// IsScriptHashAddrID returns whether the id is an identifier known to prefix a
// pay-to-script-hash address on any default or registered network.  This is
// used when decoding an address string into a specific address type.  It is up
// to the caller to check both this and IsPubKeyHashAddrID and decide whether an
// address is a pubkey hash address, script hash address, neither, or
// undeterminable (if both return true).
func IsScriptHashAddrID(id byte) bool {
	_, ok := scriptHashAddrIDs[id]
	return ok
}

// IsBech32SegwitPrefix returns whether the prefix is a known prefix for segwit
// addresses on any default or registered network.  This is used when decoding
// an address string into a specific address type.
func IsBech32SegwitPrefix(prefix string) bool {
	prefix = strings.ToLower(prefix)
	_, ok := bech32SegwitPrefixes[prefix]
	return ok
}

// HDPrivateKeyToPublicKeyID accepts a private hierarchical deterministic
// extended key id and returns the associated public key id.  When the provided
// id is not registered, the ErrUnknownHDKeyID error will be returned.
func HDPrivateKeyToPublicKeyID(id []byte) ([]byte, er.R) {
	if len(id) != 4 {
		return nil, ErrUnknownHDKeyID.Default()
	}

	var key [4]byte
	copy(key[:], id)
	pubBytes, ok := hdPrivToPubKeyIDs[key]
	if !ok {
		return nil, ErrUnknownHDKeyID.Default()
	}

	return pubBytes, nil
}

// newHashFromStr converts the passed big-endian hex string into a
// chainhash.Hash.  It only differs from the one available in chainhash in that
// it panics on an error since it will only (and must only) be called with
// hard-coded, and therefore known good, hashes.
func newHashFromStr(hexStr string) *chainhash.Hash {
	hash, err := chainhash.NewHashFromStr(hexStr)
	if err != nil {
		// Ordinarily I don't like panics in library code since it
		// can take applications down without them having a chance to
		// recover which is extremely annoying, however an exception is
		// being made in this case because the only way this can panic
		// is if there is an error in the hard-coded hashes.  Thus it
		// will only ever potentially panic on init and therefore is
		// 100% predictable.
		panic(err)
	}
	return hash
}

func init() {
	// Register all default networks when the package is initialized.
	mustRegister(&MainNetParams)
	mustRegister(&TestNet3Params)
	mustRegister(&PktTestNetParams)
	mustRegister(&PktMainNetParams)
	mustRegister(&RegressionNetParams)
	mustRegister(&SimNetParams)
}
