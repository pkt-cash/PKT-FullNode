package banmgr

import (
	"net"
	"sync"
	"time"

	"github.com/pkt-cash/PKT-FullNode/btcutil/er"
	"github.com/pkt-cash/PKT-FullNode/pktlog/log"
)

type Config struct {
	DisableBanning bool
	IpWhiteList    []string
	BanThreashold  uint32
}

type BanInfo struct {
	Addr           string
	Reason         string
	BanScore       int32
	BanExpiresTime time.Time
}

type BannedPeers struct {
	time   time.Time
	reason string
}

type SuspiciousPeers struct {
	banReason       *string
	dynamicBanScore *DynamicBanScore
}

type BanMgr struct {
	config     *Config
	banned     map[string]BannedPeers
	m          sync.Mutex
	suspicious map[string]SuspiciousPeers
}

func TrimAddress(host string) string {
	address, _, err := net.SplitHostPort(host)
	if err != nil {
		log.Debugf("can't split hostport %v", err)
		return host
	}
	return address
}

func New(config *Config) *BanMgr {
	return &BanMgr{
		config:     config,
		suspicious: make(map[string]SuspiciousPeers),
		banned:     make(map[string]BannedPeers),
	}
}

func (b *BanMgr) BanScore(ip string) uint32 {
	addr := TrimAddress(ip)
	b.m.Lock()
	defer b.m.Unlock()
	if banned, ok := b.banned[addr]; ok {
		if time.Now().Before(banned.time) {
			return 9999
		}
	}
	if sus, ok := b.suspicious[addr]; ok {
		return sus.dynamicBanScore.Int()
	}
	return 0
}

func (b *BanMgr) IsBanned(ip string) bool {
	addr := TrimAddress(ip)
	b.m.Lock()
	if banned, ok := b.banned[addr]; ok {
		if time.Now().Before(banned.time) {
			log.Debugf("Peer %s is banned for another %v - disconnecting", addr, time.Until(banned.time))
			b.m.Unlock()
			return true
		}
		log.Infof("Peer %s is no longer banned", addr)
		delete(b.banned, addr)
	}
	b.m.Unlock()
	return false
}

func (b *BanMgr) ForEachIp(f func(bi BanInfo) er.R) er.R {
	b.m.Lock()
	var notExpired []BanInfo
	//Go through banned peers
	for ip, peer := range b.banned {
		if !time.Now().Before(peer.time) {
			delete(b.banned, ip)
		} else {
			notExpired = append(notExpired, BanInfo{Addr: ip, Reason: peer.reason, BanScore: -1, BanExpiresTime: peer.time})
		}
	}
	//Go through suspicious peers
	for ip, peer := range b.suspicious {
		score := peer.dynamicBanScore.Int()
		if score == 0 {
			delete(b.suspicious, ip)
		} else {
			notExpired = append(notExpired, BanInfo{Addr: ip, Reason: *peer.banReason, BanScore: int32(score), BanExpiresTime: time.Time{}})
		}
	}
	b.m.Unlock()
	for _, item := range notExpired {
		err := f(item)
		if err != nil {
			if er.IsLoopBreak(err) {
				return nil
			} else {
				return err
			}
		}
	}

	return nil
}

func (b *BanMgr) AddBanScore(host string, persistent, transient uint32, reason string) bool {
	b.m.Lock()
	defer b.m.Unlock()

	ip := TrimAddress(host)
	// No warning is logged and no score is calculated if banning is disabled.
	if b.config.DisableBanning {
		return false
	}

	for _, item := range b.config.IpWhiteList {
		if item == ip {
			log.Debugf("Misbehaving whitelisted peer %s: %s", ip, reason)
			return false
		}
	}

	if b.suspicious == nil {
		log.Debugf("Misbehaving peer %s: %s and no ban manager yet")
		return false
	}
	b.suspicious[ip] = SuspiciousPeers{
		banReason:       &reason,
		dynamicBanScore: &DynamicBanScore{},
	}
	warnThreshold := b.config.BanThreashold >> 1
	if transient == 0 && persistent == 0 {
		// The score is not being increased, but a warning message is still
		// logged if the score is above the warn threshold.
		score := b.suspicious[ip].dynamicBanScore.Int()
		if score > warnThreshold {
			log.Warnf("Misbehaving peer %s: %s -- ban score is %d, it was not increased this time", ip, reason, score)
		}
		return false
	}
	//Increase is safe for concurrent access
	score := b.suspicious[ip].dynamicBanScore.Increase(persistent, transient)
	log.Debugf("Suspicious peer ban score increased!")
	if score > warnThreshold {
		log.Warnf("Misbehaving peer %s: %s -- ban score increased to %d", ip, reason, score)
		if score > b.config.BanThreashold {
			log.Warnf("Misbehaving peer %s -- banning and disconnecting", ip)
			//add to banned
			b.banned[ip] = BannedPeers{time.Now(), reason}
			return true
			//Will be done by the server
			//sp.server.BanPeer(ip)
			//sp.Disconnect()
		}
	}
	return false
}
