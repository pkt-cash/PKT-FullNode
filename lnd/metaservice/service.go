package metaservice

import (
	"context"

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
	in *lnrpc.GetInfo2Request, inforesp *lnrpc.GetInfoResponse) (*lnrpc.GetInfo2Responce, error) {
	res, err := m.GetInfo20(ctx, in, inforesp)
	return res, er.Native(err)
}

func (m *MetaService) GetInfo20(ctx context.Context,
	in *lnrpc.GetInfo2Request, inforesp *lnrpc.GetInfoResponse) (*lnrpc.GetInfo2Responce, er.R) {

	var ni lnrpc.NeutrinoInfo
	neutrinoPeers := m.Neutrino.Peers()
	for i := range neutrinoPeers {
		var peerDesc lnrpc.PeerDesc
		neutrinoPeer := neutrinoPeers[i]
		peerDesc.Addr = neutrinoPeer.Addr()
		peerDesc.AdvertisedProtoVer = neutrinoPeer.Describe().AdvertisedProtoVer
		peerDesc.BytesReceived = neutrinoPeer.BytesReceived()
		peerDesc.BytesSent = neutrinoPeer.BytesSent()
		desc := neutrinoPeer.Describe()
		peerDesc.Cfg = &lnrpc.Config{}
		peerDesc.Cfg.UserAgentName = desc.Cfg.UserAgentName
		peerDesc.Cfg.UserAgentVersion = desc.Cfg.UserAgentVersion
		peerDesc.Cfg.UserAgentComments = desc.Cfg.UserAgentComments
		peerDesc.Cfg.DisableRelayTx = desc.Cfg.DisableRelayTx
		peerDesc.Cfg.Services = desc.Cfg.Services.String()
		peerDesc.Cfg.TrickleInterval = int64(desc.Cfg.TrickleInterval)
		peerDesc.Connected = neutrinoPeer.Connected()
		peerDesc.Id = neutrinoPeer.ID()
		peerDesc.Inbound = neutrinoPeer.Inbound()
		peerDesc.LastAnnouncedBlock = neutrinoPeer.LastAnnouncedBlock().CloneBytes()
		peerDesc.LastBlock = neutrinoPeer.LastBlock()
		peerDesc.LastPingMicros = neutrinoPeer.LastPingMicros()
		peerDesc.LastPingNonce = neutrinoPeer.LastPingNonce()
		peerDesc.LastPingTime = neutrinoPeer.LastPingTime().String()
		peerDesc.LastRecv = neutrinoPeer.LastRecv().String()
		peerDesc.LastSend = neutrinoPeer.LastSend().String()

		ni.Peers = append(ni.Peers, &peerDesc)
	}

	ni.Bans = nil
	neutrionoQueries := m.Neutrino.GetActiveQueries()
	for i := range neutrionoQueries {
		var nq lnrpc.NeutrinoQuery
		query := neutrionoQueries[i]
		nq.Command = query.Command
		nq.CreateTime = query.CreateTime
		nq.LastRequestTime = query.LastRequestTime
		nq.LastResponseTime = query.LastResponseTime
		nq.Peer = query.Peer.String()
		nq.ReqNum = query.ReqNum

		ni.Queries = append(ni.Queries, &nq)
	}

	ni.BlockHash = ""      //???
	ni.Height = 0          //???
	ni.BlockTimestamp = "" //???
	ni.IsSyncing = false

	var wallet lnrpc.WalletInfo
	
	return &lnrpc.GetInfo2Responce{
		Neutrino:  &ni,
		Wallet:    &wallet,
		Lightning: inforesp,
	}, nil
}
