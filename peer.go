package dctk

import (
	"fmt"

	"github.com/aler9/go-dc/adc"
	atypes "github.com/aler9/go-dc/adc/types"
	"github.com/aler9/go-dc/nmdc"

	"github.com/aler9/dctk/pkg/log"
	"github.com/aler9/dctk/pkg/protoadc"
)

// Peer represents a remote client connected to a Hub.
type Peer struct {
	// peer nickname
	Nick string
	// peer description (if provided)
	Description string
	// peer email (if provided)
	Email string
	// whether peer is a bot
	IsBot bool
	// whether peer is a operator
	IsOperator bool
	// client used by peer (in NMDC this could be hidden)
	Client string
	// version of client (in NMDC this could be hidden)
	Version string
	// overall size of files shared by peer
	ShareSize uint64
	// whether peer is in passive mode (in NMDC this could be hidden)
	IsPassive bool
	// peer ip (if provided by both peer and hub)
	IP string

	adcSessionID   atypes.SID
	adcClientID    atypes.CID
	adcFingerprint string
	adcFeatures    adc.ExtFeatures
	adcUDPPort     uint
	nmdcConnection string
	nmdcFlag       nmdc.UserFlag
}

// Peers returns a map containing all the peers connected to current hub.
func (c *Client) Peers() map[string]*Peer {
	return c.peers
}

func (c *Client) peerByNick(nick string) *Peer {
	if p, ok := c.peers[nick]; ok {
		return p
	}
	return nil
}

func (c *Client) peerBySessionID(sessionID adc.SID) *Peer {
	for _, p := range c.peers {
		if p.adcSessionID == sessionID {
			return p
		}
	}
	return nil
}

func (c *Client) peerByClientID(clientID adc.CID) *Peer {
	for _, p := range c.peers {
		if p.adcClientID == clientID {
			return p
		}
	}
	return nil
}

func (c *Client) peerSupportsAdc(p *Peer, f adc.Feature) bool {
	return p.adcFeatures.Has(f)
}

func (c *Client) peerSupportsEncryption(p *Peer) bool {
	if c.protoIsAdc() {
		if p.adcFingerprint != "" {
			return true
		}
		if c.peerSupportsAdc(p, adc.FeaADCS) {
			return true
		}
		return false
	}

	// we check only for bit 4
	return (p.nmdcFlag & nmdc.FlagTLSDownload) != 0
}

func (c *Client) peerRequestConnection(peer *Peer, adcToken string) {
	if !c.conf.IsPassive {
		c.peerConnectToMe(peer, adcToken)
	} else {
		c.peerRevConnectToMe(peer, adcToken)
	}
}

func (c *Client) peerConnectToMe(peer *Peer, adcToken string) {
	if c.protoIsAdc() {
		c.hubConn.conn.Write(&protoadc.AdcDConnectToMe{ //nolint:govet
			&adc.DirectPacket{ID: c.adcSessionID, To: peer.adcSessionID},
			&adc.ConnectRequest{ //nolint:govet
				func() string {
					if c.conf.PeerEncryptionMode != DisableEncryption && c.peerSupportsEncryption(peer) {
						return adc.ProtoADCS
					}
					return adc.ProtoADC
				}(),
				func() int {
					if c.conf.PeerEncryptionMode != DisableEncryption && c.peerSupportsEncryption(peer) {
						return int(c.conf.TLSPort)
					}
					return int(c.conf.TCPPort)
				}(),
				adcToken,
			},
		})
	} else {
		c.hubConn.conn.Write(&nmdc.ConnectToMe{
			Targ: peer.Nick,
			Address: fmt.Sprintf("%s:%d", c.ip, func() uint {
				if c.conf.PeerEncryptionMode != DisableEncryption && c.peerSupportsEncryption(peer) {
					return c.conf.TLSPort
				}
				return c.conf.TCPPort
			}()),
			Secure: (c.conf.PeerEncryptionMode != DisableEncryption && c.peerSupportsEncryption(peer)),
		})
	}
}

func (c *Client) peerRevConnectToMe(peer *Peer, adcToken string) {
	if c.protoIsAdc() {
		c.hubConn.conn.Write(&protoadc.AdcDRevConnectToMe{ //nolint:govet
			&adc.DirectPacket{ID: c.adcSessionID, To: peer.adcSessionID},
			&adc.RevConnectRequest{ //nolint:govet
				func() string {
					if c.conf.PeerEncryptionMode != DisableEncryption && c.peerSupportsEncryption(peer) {
						return adc.ProtoADCS
					}
					return adc.ProtoADC
				}(),
				adcToken,
			},
		})
	} else {
		c.hubConn.conn.Write(&nmdc.RevConnectToMe{
			From: c.conf.Nick,
			To:   peer.Nick,
		})
	}
}

func (c *Client) handlePeerConnected(peer *Peer) {
	c.peers[peer.Nick] = peer
	log.Log(c.conf.LogLevel, log.LevelInfo, "[hub] [peer on] %s (%v)", peer.Nick, peer.ShareSize)
	if c.OnPeerConnected != nil {
		c.OnPeerConnected(peer)
	}
}

func (c *Client) handlePeerUpdated(peer *Peer) {
	if c.OnPeerUpdated != nil {
		c.OnPeerUpdated(peer)
	}
}

func (c *Client) handlePeerDisconnected(peer *Peer) {
	delete(c.peers, peer.Nick)
	log.Log(c.conf.LogLevel, log.LevelInfo, "[hub] [peer off] %s", peer.Nick)
	if c.OnPeerDisconnected != nil {
		c.OnPeerDisconnected(peer)
	}
}

func (c *Client) handlePeerRevConnectToMe(peer *Peer, adcToken string) {
	// we can process RevConnectToMe only in active mode
	if !c.conf.IsPassive {
		c.peerConnectToMe(peer, adcToken)
	}
}
