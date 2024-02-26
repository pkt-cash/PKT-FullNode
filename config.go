// Copyright (c) 2013-2017 The btcsuite developers
// Copyright (c) 2019 Caleb James DeLisle
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	flags "github.com/jessevdk/go-flags"
	"github.com/pkt-cash/PKT-FullNode/blockchain"
	"github.com/pkt-cash/PKT-FullNode/btcutil"
	"github.com/pkt-cash/PKT-FullNode/btcutil/er"
	"github.com/pkt-cash/PKT-FullNode/chaincfg"
	"github.com/pkt-cash/PKT-FullNode/chaincfg/chainhash"
	"github.com/pkt-cash/PKT-FullNode/chaincfg/globalcfg"
	"github.com/pkt-cash/PKT-FullNode/database"
	_ "github.com/pkt-cash/PKT-FullNode/database/ffldb"
	"github.com/pkt-cash/PKT-FullNode/mempool"
	"github.com/pkt-cash/PKT-FullNode/mining"
	"github.com/pkt-cash/PKT-FullNode/peer"
	"github.com/pkt-cash/PKT-FullNode/pktconfig/version"
	"github.com/pkt-cash/PKT-FullNode/pktlog/log"
)

const (
	defaultConfigFilename        = "pktd.conf"
	defaultDataDirname           = "data"
	defaultLogLevel              = "info"
	defaultLogDirname            = "logs"
	defaultMaxPeers              = 2048
	defaultBanDuration           = time.Hour * 24
	defaultBanThreshold          = 120
	defaultConnectTimeout        = time.Second * 10
	defaultMaxRPCClients         = 10
	defaultMaxRPCWebsockets      = 25
	defaultMaxRPCConcurrentReqs  = 20
	defaultDbType                = "ffldb"
	defaultFreeTxRelayLimit      = 15.0
	defaultTrickleInterval       = peer.DefaultTrickleInterval
	defaultBlockMinSize          = 0
	defaultBlockMaxSize          = 750000
	defaultBlockMinWeight        = 0
	defaultBlockMaxWeight        = 3000000
	blockMaxSizeMin              = 1000
	blockMaxSizeMax              = blockchain.MaxBlockBaseSize - 1000
	blockMaxWeightMin            = 4000
	blockMaxWeightMax            = blockchain.MaxBlockWeight - 4000
	defaultGenerate              = false
	defaultMaxOrphanTransactions = 100
	defaultMaxOrphanTxSize       = 100000
	defaultSigCacheMaxSize       = 100000
	defaultTxIndex               = false
	defaultAddrIndex             = false
)

var (
	defaultHomeDir     = btcutil.AppDataDir("pktd", false)
	defaultConfigFile  = filepath.Join(defaultHomeDir, defaultConfigFilename)
	defaultDataDir     = filepath.Join(defaultHomeDir, defaultDataDirname)
	knownDbTypes       = database.SupportedDrivers()
	defaultRPCKeyFile  = filepath.Join(defaultHomeDir, "rpc.key")
	defaultRPCCertFile = filepath.Join(defaultHomeDir, "rpc.cert")
	defaultLogDir      = filepath.Join(defaultHomeDir, defaultLogDirname)
)

// runServiceCommand is only set to a real function on Windows.  It is used
// to parse and execute service commands specified via the -s flag.
var runServiceCommand func(string) er.R

// minUint32 is a helper function to return the minimum of two uint32s.
// This avoids a math import and the need to cast to floats.
func minUint32(a, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}

// config defines the configuration options for pktd.
//
// See loadConfig for details on the configuration load process.
type config struct {
	ShowVersion          bool          `short:"V" long:"version" description:"Display version information and exit"`
	ConfigFile           string        `short:"C" long:"configfile" description:"Path to configuration file"`
	DataDir              string        `short:"b" long:"datadir" description:"Directory to store data"`
	LogDir               string        `long:"logdir" description:"Directory to log output."`
	AddPeers             []string      `short:"a" long:"addpeer" description:"Add a peer to connect with at startup"`
	ConnectPeers         []string      `long:"connect" description:"Connect only to the specified peers at startup"`
	DisableListen        bool          `long:"nolisten" description:"Disable listening for incoming connections -- NOTE: Listening is automatically disabled if the --connect option is used without also specifying listening interfaces via --listen"`
	Listeners            []string      `long:"listen" description:"Add an interface/port to listen for connections (default all interfaces port: 8333, testnet: 18333)"`
	MaxPeers             int           `long:"maxpeers" description:"Max number of inbound and outbound peers"`
	DisableBanning       bool          `long:"nobanning" description:"Disable banning of misbehaving peers"`
	BanDuration          time.Duration `long:"banduration" description:"How long to ban misbehaving peers.  Valid time units are {s, m, h}.  Minimum 1 second"`
	BanThreshold         uint32        `long:"banthreshold" description:"Maximum allowed ban score before disconnecting and banning misbehaving peers."`
	Whitelists           []string      `long:"whitelist" description:"Add an IP network or IP that will not be banned. (eg. 192.168.1.0/24 or ::1)"`
	AgentBlacklist       []string      `long:"agentblacklist" description:"A comma separated list of user-agent substrings which will cause pktd to reject any peers whose user-agent contains any of the blacklisted substrings."`
	AgentWhitelist       []string      `long:"agentwhitelist" description:"A comma separated list of user-agent substrings which will cause pktd to require all peers' user-agents to contain one of the whitelisted substrings. The blacklist is applied before the blacklist, and an empty whitelist will allow all agents that do not fail the blacklist."`
	HomeDir              string        `long:"homedir" description:"Creates this directory at startup"`
	RPCUser              string        `short:"u" long:"rpcuser" description:"Username for RPC connections"`
	RPCPass              string        `short:"P" long:"rpcpass" default-mask:"-" description:"Password for RPC connections"`
	RPCLimitUser         string        `long:"rpclimituser" description:"Username for limited RPC connections"`
	RPCLimitPass         string        `long:"rpclimitpass" default-mask:"-" description:"Password for limited RPC connections"`
	RPCListeners         []string      `long:"rpclisten" description:"Add an interface/port to listen for RPC connections (default port: 8334, testnet: 18334)"`
	RPCCert              string        `long:"rpccert" description:"File containing the certificate file"`
	RPCKey               string        `long:"rpckey" description:"File containing the certificate key"`
	RPCMaxClients        int           `long:"rpcmaxclients" description:"Max number of RPC clients for standard connections"`
	RPCMaxWebsockets     int           `long:"rpcmaxwebsockets" description:"Max number of RPC websocket connections"`
	RPCMaxConcurrentReqs int           `long:"rpcmaxconcurrentreqs" description:"Max number of concurrent RPC requests that may be processed concurrently"`
	RPCQuirks            bool          `long:"rpcquirks" description:"Mirror some JSON-RPC quirks of Bitcoin Core -- NOTE: Discouraged unless interoperability issues need to be worked around"`
	DisableRPC           bool          `long:"norpc" description:"Disable built-in RPC server -- NOTE: The RPC server is disabled by default if no rpcuser/rpcpass or rpclimituser/rpclimitpass is specified"`
	DisableTLS           bool          `long:"notls" description:"Nolonger used, see --tls" hidden:"true"`
	EnableTLS            bool          `long:"tls" description:"Enable TLS for the RPC server -- default is disabled unless bound to non-localhost"`
	DisableDNSSeed       bool          `long:"nodnsseed" description:"Disable DNS seeding for peers"`
	ExternalIPs          []string      `long:"externalip" description:"Add an ip to the list of local addresses we claim to listen on to peers"`
	TestNet3             bool          `long:"testnet" description:"Use the test network"`
	PktTest              bool          `long:"pkttest" description:"Use the pkt.cash test network"`
	BtcMainNet           bool          `long:"btc" description:"Use the bitcoin main network"`
	PktMainNet           bool          `long:"pkt" description:"Use the pkt.cash main network"`
	RegressionTest       bool          `long:"regtest" description:"Use the regression test network"`
	SimNet               bool          `long:"simnet" description:"Use the simulation test network"`
	AddCheckpoints       []string      `long:"addcheckpoint" description:"Add a custom checkpoint.  Format: '<height>:<hash>'"`
	DisableCheckpoints   bool          `long:"nocheckpoints" description:"Disable built-in checkpoints.  Don't do this unless you know what you're doing."`
	DbType               string        `long:"dbtype" description:"Database backend to use for the Block Chain"`
	StatsViz             string        `long:"statsviz" description:"Enable StatsViz runtime visualization on given port -- NOTE port must be between 1024 and 65535"`
	Profile              string        `long:"profile" description:"Enable HTTP profiling on given port -- NOTE port must be between 1024 and 65535"`
	CPUProfile           string        `long:"cpuprofile" description:"Write CPU profile to the specified file"`
	DebugLevel           string        `short:"d" long:"debuglevel" description:"Logging level for all subsystems {trace, debug, info, warn, error, critical} -- You may also specify <subsystem>=<level>,<subsystem2>=<level>,... to set the log level for individual subsystems -- Use show to list available subsystems"`
	Upnp                 bool          `long:"upnp" description:"Use UPnP to map our listening port outside of NAT"`
	MinRelayTxFee        float64       `long:"minrelaytxfee" description:"The minimum transaction fee in BTC/kB to be considered a non-zero fee."`
	FreeTxRelayLimit     float64       `long:"limitfreerelay" description:"Limit relay of transactions with no transaction fee to the given amount in thousands of bytes per minute"`
	NoRelayPriority      bool          `long:"norelaypriority" description:"Do not require free or low-fee transactions to have high priority for relaying"`
	TrickleInterval      time.Duration `long:"trickleinterval" description:"Minimum time between attempts to send new inventory to a connected peer"`
	MaxOrphanTxs         int           `long:"maxorphantx" description:"Max number of orphan transactions to keep in memory"`
	Generate             bool          `long:"generate" hidden:"true" description:"Generate (mine) bitcoins using the CPU - doesn't work for PacketCrypt"`
	Coinbase             string        `long:"coinbase" description:"Include this message in generated coinbase"`
	MiningAddrs          []string      `long:"miningaddr" description:"Add the specified payment address to the list of addresses to use for generated blocks -- At least one address is required if the generate option is set"`
	BlockMinSize         uint32        `long:"blockminsize" description:"Mininum block size in bytes to be used when creating a block"`
	BlockMaxSize         uint32        `long:"blockmaxsize" description:"Maximum block size in bytes to be used when creating a block"`
	BlockMinWeight       uint32        `long:"blockminweight" description:"Mininum block weight to be used when creating a block"`
	BlockMaxWeight       uint32        `long:"blockmaxweight" description:"Maximum block weight to be used when creating a block"`
	BlockPrioritySize    uint32        `long:"blockprioritysize" description:"Size in bytes for high-priority/low-fee transactions when creating a block"`
	UserAgentComments    []string      `long:"uacomment" description:"Comment to add to the user agent -- See BIP 14 for more information."`
	NoPeerBloomFilters   bool          `long:"nopeerbloomfilters" description:"Disable bloom filtering support"`
	NoCFilters           bool          `long:"nocfilters" description:"Disable committed filtering (CF) support"`
	DropCfIndex          bool          `long:"dropcfindex" description:"Deletes the index used for committed filtering (CF) support from the database on start up and then exits."`
	SigCacheMaxSize      uint          `long:"sigcachemaxsize" description:"The maximum number of entries in the signature verification cache"`
	BlocksOnly           bool          `long:"blocksonly" description:"Do not accept transactions from remote peers."`
	TxIndex              bool          `long:"txindex" description:"Maintain a full hash-based transaction index which makes all transactions available via the getrawtransaction RPC"`
	DropTxIndex          bool          `long:"droptxindex" description:"Deletes the hash-based transaction index from the database on start up and then exits."`
	AddrIndex            bool          `long:"addrindex" description:"Maintain a full address-based transaction index which makes the searchrawtransactions RPC available"`
	DropAddrIndex        bool          `long:"dropaddrindex" description:"Deletes the address-based transaction index from the database on start up and then exits."`
	DropVotes            bool          `long:"dropvotes" description:"Delete the votes table and then exit, votes will be recomputed on next startup."`
	RelayNonStd          bool          `long:"relaynonstd" description:"Relay non-standard transactions regardless of the default settings for the active network."`
	RejectNonStd         bool          `long:"rejectnonstd" description:"Reject non-standard transactions regardless of the default settings for the active network."`
	RejectReplacement    bool          `long:"rejectreplacement" description:"Reject transactions that attempt to replace existing transactions within the mempool through the Replace-By-Fee (RBF) signaling policy."`
	MiningSkipChecks     string        `long:"miningskipchecks" description:"Either 'txns', 'template' or 'both', skips certain time-consuming checks during mining process, be careful as you might create invalid block templates!"`
	lookup               func(string) ([]net.IP, er.R)
	dial                 func(string, string, time.Duration) (net.Conn, er.R)
	addCheckpoints       []chaincfg.Checkpoint
	miningAddrs          map[btcutil.Address]float64
	minRelayTxFee        btcutil.Amount
	whitelists           []*net.IPNet
}

// serviceOptions defines the configuration options for the daemon as a service on
// Windows.
type serviceOptions struct {
	ServiceCommand string `short:"s" long:"service" description:"Service command {install, remove, start, stop}"`
}

// cleanAndExpandPath expands environment variables and leading ~ in the
// passed path, cleans the result, and returns it.
func cleanAndExpandPath(path string) string {
	// Expand initial ~ to OS specific home directory.
	if strings.HasPrefix(path, "~") {
		homeDir := filepath.Dir(defaultHomeDir)
		path = strings.Replace(path, "~", homeDir, 1)
	}

	// NOTE: The os.ExpandEnv doesn't work with Windows-style %VARIABLE%,
	// but they variables can still be expanded via POSIX-style $VARIABLE.
	return filepath.Clean(os.ExpandEnv(path))
}

// validLogLevel returns whether or not logLevel is a valid debug log level.
func validLogLevel(logLevel string) bool {
	switch logLevel {
	case "trace":
		fallthrough
	case "debug":
		fallthrough
	case "info":
		fallthrough
	case "warn":
		fallthrough
	case "error":
		fallthrough
	case "critical":
		return true
	}
	return false
}

// validDbType returns whether or not dbType is a supported database type.
func validDbType(dbType string) bool {
	for _, knownType := range knownDbTypes {
		if dbType == knownType {
			return true
		}
	}

	return false
}

// removeDuplicateAddresses returns a new slice with all duplicate entries in
// addrs removed.
func removeDuplicateAddresses(addrs []string) []string {
	result := make([]string, 0, len(addrs))
	seen := map[string]struct{}{}
	for _, val := range addrs {
		if _, ok := seen[val]; !ok {
			result = append(result, val)
			seen[val] = struct{}{}
		}
	}
	return result
}

// normalizeAddress returns addr with the passed default port appended if
// there is not already a port specified.
func normalizeAddress(addr, defaultPort string) string {
	_, _, err := net.SplitHostPort(addr)
	if err != nil {
		return net.JoinHostPort(addr, defaultPort)
	}
	return addr
}

// normalizeAddresses returns a new slice with all the passed peer addresses
// normalized with the given default port, and all duplicates removed.
func normalizeAddresses(addrs []string, defaultPort string) []string {
	for i, addr := range addrs {
		addrs[i] = normalizeAddress(addr, defaultPort)
	}

	return removeDuplicateAddresses(addrs)
}

// newCheckpointFromStr parses checkpoints in the '<height>:<hash>' format.
func newCheckpointFromStr(checkpoint string) (chaincfg.Checkpoint, er.R) {
	parts := strings.Split(checkpoint, ":")
	if len(parts) != 2 {
		return chaincfg.Checkpoint{}, er.Errorf("unable to parse "+
			"checkpoint %q -- use the syntax <height>:<hash>",
			checkpoint)
	}

	height, errr := strconv.ParseInt(parts[0], 10, 32)
	if errr != nil {
		return chaincfg.Checkpoint{}, er.Errorf("unable to parse "+
			"checkpoint %q due to malformed height", checkpoint)
	}

	if len(parts[1]) == 0 {
		return chaincfg.Checkpoint{}, er.Errorf("unable to parse "+
			"checkpoint %q due to missing hash", checkpoint)
	}
	hash, err := chainhash.NewHashFromStr(parts[1])
	if err != nil {
		return chaincfg.Checkpoint{}, er.Errorf("unable to parse "+
			"checkpoint %q due to malformed hash", checkpoint)
	}

	return chaincfg.Checkpoint{
		Height: int32(height),
		Hash:   hash,
	}, nil
}

// parseCheckpoints checks the checkpoint strings for valid syntax
// ('<height>:<hash>') and parses them to chaincfg.Checkpoint instances.
func parseCheckpoints(checkpointStrings []string) ([]chaincfg.Checkpoint, er.R) {
	if len(checkpointStrings) == 0 {
		return nil, nil
	}
	checkpoints := make([]chaincfg.Checkpoint, len(checkpointStrings))
	for i, cpString := range checkpointStrings {
		checkpoint, err := newCheckpointFromStr(cpString)
		if err != nil {
			return nil, err
		}
		checkpoints[i] = checkpoint
	}
	return checkpoints, nil
}

// filesExists reports whether the named file or directory exists.
func fileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// newConfigParser returns a new command line flags parser.
func newConfigParser(cfg *config, so *serviceOptions, options flags.Options) *flags.Parser {
	parser := flags.NewParser(cfg, options)
	if runtime.GOOS == "windows" {
		parser.AddGroup("Service Options", "Service Options", so)
	}
	return parser
}

// loadConfig initializes and parses the config using a config file and command
// line options.
//
// The configuration proceeds as follows:
// 	1) Start with a default config with sane settings
// 	2) Pre-parse the command line to check for an alternative config file
// 	3) Load configuration file overwriting defaults with any specified options
// 	4) Parse CLI options and overwrite/add any specified options
//
// The above results in pktd functioning properly without any config settings
// while still allowing the user to override settings with config files and
// command line options.  Command line options always take precedence.
func loadConfig() (*config, []string, er.R) {
	// Default config.
	cfg := config{
		ConfigFile:           defaultConfigFile,
		DebugLevel:           defaultLogLevel,
		MaxPeers:             defaultMaxPeers,
		BanDuration:          defaultBanDuration,
		BanThreshold:         defaultBanThreshold,
		RPCMaxClients:        defaultMaxRPCClients,
		RPCMaxWebsockets:     defaultMaxRPCWebsockets,
		RPCMaxConcurrentReqs: defaultMaxRPCConcurrentReqs,
		HomeDir:              defaultHomeDir,
		DataDir:              defaultDataDir,
		LogDir:               defaultLogDir,
		DbType:               defaultDbType,
		RPCKey:               defaultRPCKeyFile,
		RPCCert:              defaultRPCCertFile,
		MinRelayTxFee:        -1, // this gets configured later
		FreeTxRelayLimit:     defaultFreeTxRelayLimit,
		TrickleInterval:      defaultTrickleInterval,
		Coinbase:             mining.DefaultCoinbaseFlags,
		MiningAddrs:          []string{"pkt1q6hqsqhqdgqfd8t3xwgceulu7k9d9w5t2amath0qxyfjlvl3s3u4sjza2g2"},
		BlockMinSize:         defaultBlockMinSize,
		BlockMaxSize:         defaultBlockMaxSize,
		BlockMinWeight:       defaultBlockMinWeight,
		BlockMaxWeight:       defaultBlockMaxWeight,
		BlockPrioritySize:    mempool.DefaultBlockPrioritySize,
		MaxOrphanTxs:         defaultMaxOrphanTransactions,
		SigCacheMaxSize:      defaultSigCacheMaxSize,
		Generate:             defaultGenerate,
		TxIndex:              defaultTxIndex,
		AddrIndex:            defaultAddrIndex,
	}

	// Service options which are only added on Windows.
	serviceOpts := serviceOptions{}

	// Pre-parse the command line options to see if an alternative config
	// file or the version flag was specified.  Any errors aside from the
	// help message error can be ignored here since they will be caught by
	// the final parse below.
	preCfg := cfg
	preParser := newConfigParser(&preCfg, &serviceOpts, flags.HelpFlag)
	_, errr := preParser.Parse()
	if errr != nil {
		if e, ok := errr.(*flags.Error); ok && e.Type == flags.ErrHelp {
			fmt.Fprintln(os.Stderr, errr)
			return nil, nil, er.E(errr)
		}
	}

	// Show the version and exit if the version flag was specified.
	appName := filepath.Base(os.Args[0])
	appName = strings.TrimSuffix(appName, filepath.Ext(appName))
	usageMessage := fmt.Sprintf("Use %s -h to show usage", appName)
	if preCfg.ShowVersion {
		fmt.Println(appName, "version", version.Version())
		os.Exit(0)
	}

	// Perform service command and exit if specified.  Invalid service
	// commands show an appropriate error.  Only runs on Windows since
	// the runServiceCommand function will be nil when not on Windows.
	if serviceOpts.ServiceCommand != "" && runServiceCommand != nil {
		err := runServiceCommand(serviceOpts.ServiceCommand)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(0)
	}

	// Load additional config from file.
	configNotFound := false
	parser := newConfigParser(&cfg, &serviceOpts, flags.Default)
	if !(preCfg.RegressionTest || preCfg.SimNet) || preCfg.ConfigFile !=
		defaultConfigFile {

		errr := flags.NewIniParser(parser).ParseFile(preCfg.ConfigFile)
		if errr != nil {
			if _, ok := errr.(*os.PathError); !ok {
				fmt.Fprintf(os.Stderr, "Error parsing config "+
					"file: %v\n", errr)
				fmt.Fprintln(os.Stderr, usageMessage)
				return nil, nil, er.E(errr)
			}
			configNotFound = true
		}
	}

	// Don't add peers from the config file when in regression test mode.
	if preCfg.RegressionTest && len(cfg.AddPeers) > 0 {
		cfg.AddPeers = nil
	}

	// Parse command line options again to ensure they take precedence.
	remainingArgs, errr := parser.Parse()
	if errr != nil {
		if e, ok := errr.(*flags.Error); !ok || e.Type != flags.ErrHelp {
			fmt.Fprintln(os.Stderr, usageMessage)
		}
		return nil, nil, er.E(errr)
	}

	// Create the home directory if it doesn't already exist.
	funcName := "loadConfig"
	errr = os.MkdirAll(cfg.HomeDir, 0700)
	if errr != nil {
		// Show a nicer error message if it's because a symlink is
		// linked to a directory that does not exist (probably because
		// it's not mounted).
		if e, ok := errr.(*os.PathError); ok && os.IsExist(errr) {
			if link, lerr := os.Readlink(e.Path); lerr == nil {
				str := "is symlink %s -> %s mounted?"
				errr = fmt.Errorf(str, e.Path, link)
			}
		}

		str := "%s: Failed to create home directory: %v"
		err := er.Errorf(str, funcName, errr)
		fmt.Fprintln(os.Stderr, err)
		return nil, nil, err
	}

	// Multiple networks can't be selected simultaneously.
	numNets := 0
	// Count number of network flags passed; assign active network params
	// while we're at it
	if cfg.TestNet3 {
		numNets++
		activeNetParams = &testNet3Params
	}
	if cfg.PktTest {
		numNets++
		activeNetParams = &pktTestNetParams
	}
	if cfg.BtcMainNet {
		numNets++
		activeNetParams = &mainNetParams
	}
	if cfg.PktMainNet {
		numNets++
		activeNetParams = &pktMainNetParams
	}
	if cfg.RegressionTest {
		numNets++
		activeNetParams = &regressionNetParams
	}
	if cfg.SimNet {
		numNets++
		// Also disable dns seeding on the simulation test network.
		activeNetParams = &simNetParams
		cfg.DisableDNSSeed = true
	}
	if numNets > 1 {
		str := "%s: The testnet, regtest, segnet, and simnet params " +
			"can't be used together -- choose one of the four"
		err := er.Errorf(str, funcName)
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return nil, nil, err
	}

	// TODO(cjd): this is trash, but CompactToBig is a util function and it shouldn't
	// be in blockchain, but it is, and trying to call it from cfg is a dependency
	// loop. And duplicating the powlimit twice in the config is also trash...
	activeNetParams.PowLimit = blockchain.CompactToBig(activeNetParams.PowLimitBits)

	globalcfg.SelectConfig(activeNetParams.GlobalConf)

	// Set the default policy for relaying non-standard transactions
	// according to the default of the active network. The set
	// configuration value takes precedence over the default value for the
	// selected network.
	relayNonStd := activeNetParams.RelayNonStdTxs
	switch {
	case cfg.RelayNonStd && cfg.RejectNonStd:
		str := "%s: rejectnonstd and relaynonstd cannot be used " +
			"together -- choose only one"
		err := er.Errorf(str, funcName)
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return nil, nil, err
	case cfg.RejectNonStd:
		relayNonStd = false
	case cfg.RelayNonStd:
		relayNonStd = true
	}
	cfg.RelayNonStd = relayNonStd

	// Append the network type to the data directory so it is "namespaced"
	// per network.  In addition to the block database, there are other
	// pieces of data that are saved to disk such as address manager state.
	// All data is specific to a network, so namespacing the data directory
	// means each individual piece of serialized data does not have to
	// worry about changing names per network and such.
	cfg.DataDir = cleanAndExpandPath(cfg.DataDir)
	cfg.DataDir = filepath.Join(cfg.DataDir, netName(activeNetParams))

	// Append the network type to the log directory so it is "namespaced"
	// per network in the same fashion as the data directory.
	cfg.LogDir = cleanAndExpandPath(cfg.LogDir)
	cfg.LogDir = filepath.Join(cfg.LogDir, netName(activeNetParams))

	// Parse, validate, and set debug log level(s).
	if err := log.SetLogLevels(cfg.DebugLevel); err != nil {
		err := er.Errorf("%s: %v", funcName, err)
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return nil, nil, err
	}

	// Validate database type.
	if !validDbType(cfg.DbType) {
		str := "%s: The specified database type [%v] is invalid -- " +
			"supported types %v"
		err := er.Errorf(str, funcName, cfg.DbType, knownDbTypes)
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return nil, nil, err
	}

	// Validate profile port number
	if cfg.Profile != "" {
		profilePort, err := strconv.Atoi(cfg.Profile)
		if err != nil || profilePort < 1024 || profilePort > 65535 {
			str := "%s: The profile port must be between 1024 and 65535"
			err := er.Errorf(str, funcName)
			fmt.Fprintln(os.Stderr, err)
			fmt.Fprintln(os.Stderr, usageMessage)
			return nil, nil, err
		}
	}

	// Validate StatsViz port number
	if cfg.StatsViz != "" {
		statsvizPort, err := strconv.Atoi(cfg.StatsViz)
		if err != nil || statsvizPort < 1024 || statsvizPort > 65535 {
			str := "%s: The StatsViz port must be between 1024 and 65535"
			err := er.Errorf(str, funcName)
			fmt.Fprintln(os.Stderr, err)
			fmt.Fprintln(os.Stderr, usageMessage)
			return nil, nil, err
		}
	}

	// Don't allow ban durations that are too short.
	if cfg.BanDuration < time.Second {
		str := "%s: The banduration option may not be less than 1s -- parsed [%v]"
		err := er.Errorf(str, funcName, cfg.BanDuration)
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return nil, nil, err
	}

	// Validate any given whitelisted IP addresses and networks.
	if len(cfg.Whitelists) > 0 {
		var ip net.IP
		cfg.whitelists = make([]*net.IPNet, 0, len(cfg.Whitelists))

		for _, addr := range cfg.Whitelists {
			_, ipnet, err := net.ParseCIDR(addr)
			if err != nil {
				ip = net.ParseIP(addr)
				if ip == nil {
					str := "%s: The whitelist value of '%s' is invalid"
					err := er.Errorf(str, funcName, addr)
					fmt.Fprintln(os.Stderr, err)
					fmt.Fprintln(os.Stderr, usageMessage)
					return nil, nil, err
				}
				var bits int
				if ip.To4() == nil {
					// IPv6
					bits = 128
				} else {
					bits = 32
				}
				ipnet = &net.IPNet{
					IP:   ip,
					Mask: net.CIDRMask(bits, bits),
				}
			}
			cfg.whitelists = append(cfg.whitelists, ipnet)
		}
	}

	// --addPeer and --connect do not mix.
	if len(cfg.AddPeers) > 0 && len(cfg.ConnectPeers) > 0 {
		str := "%s: the --addpeer and --connect options can not be " +
			"mixed"
		err := er.Errorf(str, funcName)
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return nil, nil, err
	}

	// --connect without --listen disables listening.
	if len(cfg.ConnectPeers) > 0 &&
		len(cfg.Listeners) == 0 {
		cfg.DisableListen = true
	}

	// Connect means no DNS seeding.
	if len(cfg.ConnectPeers) > 0 {
		cfg.DisableDNSSeed = true
	}

	// Add the default listener if none were specified. The default
	// listener is all addresses on the listen port for the network
	// we are to connect to.
	if len(cfg.Listeners) == 0 {
		cfg.Listeners = []string{
			net.JoinHostPort("", activeNetParams.DefaultPort),
		}
	}

	// Check to make sure limited and admin users don't have the same username
	if cfg.RPCUser == cfg.RPCLimitUser && cfg.RPCUser != "" {
		str := "%s: --rpcuser and --rpclimituser must not specify the " +
			"same username"
		err := er.Errorf(str, funcName)
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return nil, nil, err
	}

	// Check to make sure limited and admin users don't have the same password
	if cfg.RPCPass == cfg.RPCLimitPass && cfg.RPCPass != "" {
		str := "%s: --rpcpass and --rpclimitpass must not specify the " +
			"same password"
		err := er.Errorf(str, funcName)
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return nil, nil, err
	}

	if cfg.RPCUser == "" || cfg.RPCPass == "" {
		log.Infof("Creating a .cookie file")
		cookiePath := filepath.Join(defaultHomeDir, ".cookie")
		var buf [32]byte
		if _, errr := rand.Read(buf[:]); errr != nil {
			err := er.E(errr)
			err.AddMessage("Unable to get random numbers")
			return nil, nil, err
		}
		cfg.RPCUser = "__PKT_COOKIE__"
		cfg.RPCPass = hex.EncodeToString(buf[:])
		cookie := cfg.RPCUser + ":" + cfg.RPCPass
		if errr := ioutil.WriteFile(cookiePath, []byte(cookie), 0600); errr != nil {
			err := er.E(errr)
			err.AddMessage("Could not write cookie")
			return nil, nil, err
		}
	}

	// The RPC server is disabled if no username or password is provided.
	if (cfg.RPCUser == "" || cfg.RPCPass == "") &&
		(cfg.RPCLimitUser == "" || cfg.RPCLimitPass == "") {
		cfg.DisableRPC = true
	}

	if cfg.DisableRPC {
		log.Infof("RPC service is disabled")
	}

	// Default RPC to listen on localhost only.
	if !cfg.DisableRPC && len(cfg.RPCListeners) == 0 {
		addrs, errr := net.LookupHost("localhost")
		if errr != nil {
			return nil, nil, er.E(errr)
		}
		cfg.RPCListeners = make([]string, 0, len(addrs))
		for _, addr := range addrs {
			addr = net.JoinHostPort(addr, activeNetParams.rpcPort)
			cfg.RPCListeners = append(cfg.RPCListeners, addr)
		}
	}

	if cfg.RPCMaxConcurrentReqs < 0 {
		str := "%s: The rpcmaxwebsocketconcurrentrequests option may " +
			"not be less than 0 -- parsed [%d]"
		err := er.Errorf(str, funcName, cfg.RPCMaxConcurrentReqs)
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return nil, nil, err
	}

	// Validate the the minrelaytxfee.
	if cfg.MinRelayTxFee >= 0 {
		mrf, err := globalcfg.NewAmount(cfg.MinRelayTxFee)
		if err != nil {
			str := "%s: invalid minrelaytxfee: %v"
			err := er.Errorf(str, funcName, err)
			fmt.Fprintln(os.Stderr, err)
			fmt.Fprintln(os.Stderr, usageMessage)
			return nil, nil, err
		}
		cfg.minRelayTxFee = btcutil.Amount(mrf)
	} else {
		cfg.minRelayTxFee = mempool.DefaultMinRelayTxFee
		cfg.MinRelayTxFee = btcutil.Amount(cfg.minRelayTxFee).ToBTC()
	}

	// Limit the max block size to a sane value.
	if cfg.BlockMaxSize < blockMaxSizeMin || cfg.BlockMaxSize >
		blockMaxSizeMax {

		str := "%s: The blockmaxsize option must be in between %d " +
			"and %d -- parsed [%d]"
		err := er.Errorf(str, funcName, blockMaxSizeMin,
			blockMaxSizeMax, cfg.BlockMaxSize)
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return nil, nil, err
	}

	// Limit the max block weight to a sane value.
	if cfg.BlockMaxWeight < blockMaxWeightMin ||
		cfg.BlockMaxWeight > blockMaxWeightMax {

		str := "%s: The blockmaxweight option must be in between %d " +
			"and %d -- parsed [%d]"
		err := er.Errorf(str, funcName, blockMaxWeightMin,
			blockMaxWeightMax, cfg.BlockMaxWeight)
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return nil, nil, err
	}

	// Limit the max orphan count to a sane vlue.
	if cfg.MaxOrphanTxs < 0 {
		str := "%s: The maxorphantx option may not be less than 0 " +
			"-- parsed [%d]"
		err := er.Errorf(str, funcName, cfg.MaxOrphanTxs)
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return nil, nil, err
	}

	// Limit the block priority and minimum block sizes to max block size.
	cfg.BlockPrioritySize = minUint32(cfg.BlockPrioritySize, cfg.BlockMaxSize)
	cfg.BlockMinSize = minUint32(cfg.BlockMinSize, cfg.BlockMaxSize)
	cfg.BlockMinWeight = minUint32(cfg.BlockMinWeight, cfg.BlockMaxWeight)

	switch {
	// If the max block size isn't set, but the max weight is, then we'll
	// set the limit for the max block size to a safe limit so weight takes
	// precedence.
	case cfg.BlockMaxSize == defaultBlockMaxSize &&
		cfg.BlockMaxWeight != defaultBlockMaxWeight:

		cfg.BlockMaxSize = blockchain.MaxBlockBaseSize - 1000

	// If the max block weight isn't set, but the block size is, then we'll
	// scale the set weight accordingly based on the max block size value.
	case cfg.BlockMaxSize != defaultBlockMaxSize &&
		cfg.BlockMaxWeight == defaultBlockMaxWeight:

		cfg.BlockMaxWeight = cfg.BlockMaxSize * blockchain.WitnessScaleFactor
	}

	// Look for illegal characters in the user agent comments.
	for _, uaComment := range cfg.UserAgentComments {
		if strings.ContainsAny(uaComment, "/:()") {
			err := er.Errorf("%s: The following characters must not "+
				"appear in user agent comments: '/', ':', '(', ')'",
				funcName)
			fmt.Fprintln(os.Stderr, err)
			fmt.Fprintln(os.Stderr, usageMessage)
			return nil, nil, err
		}
	}

	// --txindex and --droptxindex do not mix.
	if cfg.TxIndex && cfg.DropTxIndex {
		err := er.Errorf("%s: the --txindex and --droptxindex "+
			"options may  not be activated at the same time",
			funcName)
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return nil, nil, err
	}

	// --addrindex and --dropaddrindex do not mix.
	if cfg.AddrIndex && cfg.DropAddrIndex {
		err := er.Errorf("%s: the --addrindex and --dropaddrindex "+
			"options may not be activated at the same time",
			funcName)
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return nil, nil, err
	}

	// --addrindex and --droptxindex do not mix.
	if cfg.AddrIndex && cfg.DropTxIndex {
		err := er.Errorf("%s: the --addrindex and --droptxindex "+
			"options may not be activated at the same time "+
			"because the address index relies on the transaction "+
			"index",
			funcName)
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return nil, nil, err
	}

	// Check mining addresses are valid and saved parsed versions.
	cfg.miningAddrs = make(map[btcutil.Address]float64)
	for _, strAddr := range cfg.MiningAddrs {
		addr, err := btcutil.DecodeAddress(strAddr, activeNetParams.Params)
		if err != nil {
			str := "%s: mining address '%s' failed to decode: %v"
			err := er.Errorf(str, funcName, strAddr, err)
			fmt.Fprintln(os.Stderr, err)
			fmt.Fprintln(os.Stderr, usageMessage)
			return nil, nil, err
		}
		if !addr.IsForNet(activeNetParams.Params) {
			str := "%s: mining address '%s' is on the wrong network"
			err := er.Errorf(str, funcName, strAddr)
			fmt.Fprintln(os.Stderr, err)
			fmt.Fprintln(os.Stderr, usageMessage)
			return nil, nil, err
		}
		cfg.miningAddrs[addr] = float64(1)
	}

	// Ensure there is at least one mining address when the generate flag is
	// set.
	if cfg.Generate && len(cfg.MiningAddrs) == 0 {
		str := "%s: the generate flag is set, but there are no mining " +
			"addresses specified "
		err := er.Errorf(str, funcName)
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return nil, nil, err
	}

	// Add default port to all listener addresses if needed and remove
	// duplicate addresses.
	cfg.Listeners = normalizeAddresses(cfg.Listeners,
		activeNetParams.DefaultPort)

	// Add default port to all rpc listener addresses if needed and remove
	// duplicate addresses.
	cfg.RPCListeners = normalizeAddresses(cfg.RPCListeners,
		activeNetParams.rpcPort)

	// Add default port to all added peer addresses if needed and remove
	// duplicate addresses.
	cfg.AddPeers = normalizeAddresses(cfg.AddPeers,
		activeNetParams.DefaultPort)
	cfg.ConnectPeers = normalizeAddresses(cfg.ConnectPeers,
		activeNetParams.DefaultPort)

	// Check the checkpoints for syntax errors.
	var err er.R
	cfg.addCheckpoints, err = parseCheckpoints(cfg.AddCheckpoints)
	if err != nil {
		str := "%s: Error parsing checkpoints: %v"
		err := er.Errorf(str, funcName, err)
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return nil, nil, err
	}

	// Setup dial and DNS resolution (lookup) functions depending on the
	// specified options.  The default is to use the standard
	// net.DialTimeout function as well as the system DNS resolver.
	cfg.dial = func(n string, addr string, to time.Duration) (net.Conn, er.R) {
		ret, errr := net.DialTimeout(n, addr, to)
		return ret, er.E(errr)
	}
	cfg.lookup = func(host string) ([]net.IP, er.R) {
		out, errr := net.LookupIP(host)
		return out, er.E(errr)
	}

	// Warn about missing config file only after all other configuration is
	// done.  This prevents the warning on help messages and invalid
	// options.  Note this should go directly before the return.
	if configNotFound && preCfg.ConfigFile != defaultConfigFile {
		log.Warnf("Could not find config file [%s]", preCfg.ConfigFile)
	}

	return &cfg, remainingArgs, nil
}

// pktdDial connects to the address on the named network using the appropriate
// dial function depending on the address and configuration options.
func pktdDial(addr net.Addr) (net.Conn, er.R) {
	return cfg.dial(addr.Network(), addr.String(), defaultConnectTimeout)
}

// pktdLookup resolves the IP of the given host using the correct DNS lookup
// function depending on the configuration options.
func pktdLookup(host string) ([]net.IP, er.R) {
	return cfg.lookup(host)
}
