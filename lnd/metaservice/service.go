package metaservice

import (
	"context"
	"net"
	"strconv"
	"time"

	"github.com/pkt-cash/pktd/btcjson"
	"github.com/pkt-cash/pktd/btcutil/er"
	"github.com/pkt-cash/pktd/lnd/lnrpc"
	"github.com/pkt-cash/pktd/neutrino"
	"github.com/pkt-cash/pktd/neutrino/banman"
	"github.com/pkt-cash/pktd/pktwallet/waddrmgr"
	"github.com/pkt-cash/pktd/pktwallet/wallet"
)

type MetaService struct {
	Neutrino *neutrino.ChainService
	Wallet   *wallet.Wallet
}

var _ lnrpc.MetaServiceServer = (*MetaService)(nil)

// New creates and returns a new MetaService
func NewMetaService(neutrino *neutrino.ChainService) *MetaService {
	return &MetaService{
		Neutrino: neutrino,
	}
}

func (m *MetaService) SetWallet(wallet *wallet.Wallet) {
	m.Wallet = wallet
}

func (m *MetaService) GetInfo2(ctx context.Context,
	in *lnrpc.GetInfo2Request) (*lnrpc.GetInfo2Response, error) {
	res, err := m.GetInfo20(ctx, in)
	return res, er.Native(err)
}

func (m *MetaService) GetInfo20(ctx context.Context,
	in *lnrpc.GetInfo2Request) (*lnrpc.GetInfo2Response, er.R) {

	var ni lnrpc.NeutrinoInfo
	neutrinoPeers := m.Neutrino.Peers()
	for i := range neutrinoPeers {
		var peerDesc lnrpc.PeerDesc
		neutrinoPeer := neutrinoPeers[i]

		peerDesc.BytesReceived = neutrinoPeer.BytesReceived()
		peerDesc.BytesSent = neutrinoPeer.BytesSent()
		peerDesc.LastRecv = neutrinoPeer.LastRecv().String()
		peerDesc.LastSend = neutrinoPeer.LastSend().String()
		peerDesc.Connected = neutrinoPeer.Connected()
		peerDesc.Addr = neutrinoPeer.Addr()
		peerDesc.Inbound = neutrinoPeer.Inbound()
		na := neutrinoPeer.NA()
		if na != nil {
			peerDesc.Na = na.IP.String() + ":" + strconv.Itoa(int(na.Port))
		}
		peerDesc.Id = neutrinoPeer.ID()
		peerDesc.UserAgent = neutrinoPeer.UserAgent()
		peerDesc.Services = neutrinoPeer.Services().String()
		peerDesc.VersionKnown = neutrinoPeer.VersionKnown()
		peerDesc.AdvertisedProtoVer = neutrinoPeer.Describe().AdvertisedProtoVer
		peerDesc.ProtocolVersion = neutrinoPeer.ProtocolVersion()
		peerDesc.SendHeadersPreferred = neutrinoPeer.Describe().SendHeadersPreferred
		peerDesc.VerAckReceived = neutrinoPeer.VerAckReceived()
		peerDesc.WitnessEnabled = neutrinoPeer.Describe().WitnessEnabled
		peerDesc.WireEncoding = strconv.Itoa(int(neutrinoPeer.Describe().WireEncoding))
		peerDesc.TimeOffset = neutrinoPeer.TimeOffset()
		peerDesc.TimeConnected = neutrinoPeer.Describe().TimeConnected.String()
		peerDesc.StartingHeight = neutrinoPeer.StartingHeight()
		peerDesc.LastBlock = neutrinoPeer.LastBlock()
		if neutrinoPeer.LastAnnouncedBlock() != nil {
			peerDesc.LastAnnouncedBlock = neutrinoPeer.LastAnnouncedBlock().CloneBytes()
		}
		peerDesc.LastPingNonce = neutrinoPeer.LastPingNonce()
		peerDesc.LastPingTime = neutrinoPeer.LastPingTime().String()
		peerDesc.LastPingMicros = neutrinoPeer.LastPingMicros()

		ni.Peers = append(ni.Peers, &peerDesc)
	}

	m.Neutrino.BanStore().ForEachBannedAddr(func(a *net.IPNet, r banman.Reason, t time.Time) er.R {
		ban := lnrpc.NeutrinoBan{}
		ban.Addr = a.String()
		ban.Reason = r.String()
		ban.EndTime = t.String()
		ni.Bans = append(ni.Bans, &ban)
		return nil
	})

	neutrionoQueries := m.Neutrino.GetActiveQueries()
	for i := range neutrionoQueries {
		nq := lnrpc.NeutrinoQuery{}
		query := neutrionoQueries[i]
		nq.Peer = query.Peer.String()
		nq.Command = query.Command
		nq.ReqNum = query.ReqNum
		nq.CreateTime = query.CreateTime
		nq.LastRequestTime = query.LastRequestTime
		nq.LastResponseTime = query.LastResponseTime

		ni.Queries = append(ni.Queries, &nq)
	}

	bb, err := m.Neutrino.BestBlock()
	if err != nil {
		return nil, err
	}
	ni.BlockHash = bb.Hash.String()
	ni.Height = bb.Height
	ni.BlockTimestamp = bb.Timestamp.String()
	ni.IsSyncing = !m.Neutrino.IsCurrent()

	mgrStamp := waddrmgr.BlockStamp{}
	walletInfo := &lnrpc.WalletInfo{}

	if m.Wallet != nil {
		mgrStamp = m.Wallet.Manager.SyncedTo()
		walletStats := &lnrpc.WalletStats{}
		m.Wallet.ReadStats(func(ws *btcjson.WalletStats) {
			walletStats.MaintenanceInProgress = ws.MaintenanceInProgress
			walletStats.MaintenanceName = ws.MaintenanceName
			walletStats.MaintenanceCycles = int32(ws.MaintenanceCycles)
			walletStats.MaintenanceLastBlockVisited = int32(ws.MaintenanceLastBlockVisited)
			walletStats.Syncing = ws.Syncing
			if ws.SyncStarted != nil {
				walletStats.SyncStarted = ws.SyncStarted.String()
			}
			walletStats.SyncRemainingSeconds = ws.SyncRemainingSeconds
			walletStats.SyncCurrentBlock = ws.SyncCurrentBlock
			walletStats.SyncFrom = ws.SyncFrom
			walletStats.SyncTo = ws.SyncTo
			walletStats.BirthdayBlock = ws.BirthdayBlock
		})
		walletInfo = &lnrpc.WalletInfo{
			CurrentBlockHash:      mgrStamp.Hash.String(),
			CurrentHeight:         mgrStamp.Height,
			CurrentBlockTimestamp: mgrStamp.Timestamp.String(),
			WalletVersion:         int32(waddrmgr.LatestMgrVersion),
			WalletStats:           walletStats,
		}
	} else {
		walletInfo = nil
	}

	return &lnrpc.GetInfo2Response{
		Neutrino:  &ni,
		Wallet:    walletInfo,
		Lightning: in.InfoResponse,
	}, nil
}
