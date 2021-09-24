package metaservice

import (
	"context"

	"github.com/pkt-cash/pktd/btcjson"
	"github.com/pkt-cash/pktd/btcutil/er"
	"github.com/pkt-cash/pktd/lnd/lnrpc"
	"github.com/pkt-cash/pktd/neutrino"
)

var (
	// ErrUnlockTimeout signals that we did not get the expected unlock
	// message before the timeout occurred.
	ErrUnlockTimeout = er.GenericErrorType.CodeWithDetail("ErrUnlockTimeout",
		"got no unlock message before timeout")
)

const (
	// rpcAuthTimeoutSeconds is the number of seconds a connection to the
	// RPC server is allowed to stay open without authenticating before it
	// is closed.
	rpcAuthTimeoutSeconds = 10

	// gbtNonceRange is two 32-bit big-endian hexadecimal integers which
	// represent the valid ranges of nonces returned by the getblocktemplate
	// RPC.
	gbtNonceRange = "00000000ffffffff"

	// gbtRegenerateSeconds is the number of seconds that must pass before
	// a new template is generated when the previous block hash has not
	// changed and there have been changes to the available transactions
	// in the memory pool.
	gbtRegenerateSeconds = 60

	// maxProtocolVersion is the max protocol version the server supports.
	maxProtocolVersion = 70002
)

type getInfo2 struct {
	Message string
}

type MetaService struct {
	Neutrino *neutrino.ChainService
}

var _ lnrpc.MetaServiceServer = (*MetaService)(nil)

// New creates and returns a new MetaService
func NewMetaService(neutrino *neutrino.ChainService) *MetaService {

	return &MetaService{
		Neutrino: neutrino,
	}
}

func (m *MetaService) GetInfo2(ctx context.Context,
	in *lnrpc.GetInfo2Request) (*lnrpc.GetInfo2Responce, error) {
	res, err := u.GetInfo20(ctx, in)
	return res, er.Native(err)
}

func (m *MetaService) GetInfo20(_ context.Context,
	in *lnrpc.GetInfo2Request) (*lnrpc.GetInfo2Responce, er.R) {
	var NeutrinoInfo = btcjson.NeutrinoInfo{}
	m.Neutrino.
	var ni lnrpc.NeutrinoInfo
	ni.Bans[0].Addr = NeutrinoInfo.Bans[0].Addr
	ni.Bans[0].EndTime = NeutrinoInfo.Bans[0].EndTime
	ni.Bans[0].Reason = NeutrinoInfo.Bans[0].Reason
	ni.Peers[0].Addr = NeutrinoInfo.Peers[0].Addr
	ni.Peers[0].AdvertisedProtoVer = NeutrinoInfo.Peers[0].AdvertisedProtoVer
	ni.Peers[0].BytesReceived = NeutrinoInfo.Peers[0].BytesReceived
	
	//ni.BlockHash = &btcjson.WalletInfoResult.BackendBlockHash
	//ni.BlockHash = NeutrinoInfo.Peers

	var wallet lnrpc.WalletInfo
	//wallet.CurrentBlockHash = &btcjson.WalletInfoResult.CurrentBlockHash
	var lightninginfo lnrpc.GetInfoResponse
	return &lnrpc.GetInfo2Responce{
		Neutrino:  &ni,
		Wallet:    &wallet,
		Lightning: &lightninginfo,
	}, nil
}
