// Package peers contains helpers to work with Telegram peers
package peers

import (
	"context"

	"github.com/go-faster/errors"

	"github.com/gotd/td/constant"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

// Peer represents generic peer.
type Peer interface {
	// ID returns entity ID.
	ID() int64
	// TDLibPeerID returns TDLibPeerID for this entity.
	TDLibPeerID() constant.TDLibPeerID

	// VisibleName returns visible name of peer.
	//
	// It returns FirstName + " " + LastName for users, and title for chats and channels.
	VisibleName() string
	// Username returns peer username, if any.
	Username() (string, bool)
	// Restricted whether this user/chat/channel is restricted.
	Restricted() ([]tg.RestrictionReason, bool)
	// Verified whether this user/chat/channel is verified by Telegram.
	Verified() bool
	// Scam whether this user/chat/channel is probably a scam.
	Scam() bool
	// Fake whether this user/chat/channel was reported by many users as a fake or scam: be
	// careful when interacting with it.
	Fake() bool

	// InputPeer returns input peer for this peer.
	InputPeer() tg.InputPeerClass
	// Sync updates current object.
	Sync(ctx context.Context) error

	// Report reports a peer for violation of telegram's Terms of Service.
	Report(ctx context.Context, reason tg.ReportReasonClass, message string) error
	// Photo returns peer photo, if any.
	Photo(ctx context.Context) (*tg.Photo, bool, error)
}

var _ = []Peer{
	User{},
	Chat{},
	Channel{},
}

// ResolveTDLibID creates Peer using given constant.TDLibPeerID.
func (m *Manager) ResolveTDLibID(ctx context.Context, peerID constant.TDLibPeerID) (Peer, error) {
	var p tg.PeerClass
	switch {
	case peerID.IsUser():
		p = &tg.PeerUser{UserID: peerID.ToPlain()}
	case peerID.IsChat():
		p = &tg.PeerChat{ChatID: peerID.ToPlain()}
	case peerID.IsChannel():
		p = &tg.PeerChannel{ChannelID: peerID.ToPlain()}
	default:
		return nil, errors.Errorf("invalid ID %d", peerID)
	}
	return m.ResolvePeer(ctx, p)
}

// ResolvePeer creates Peer using given tg.PeerClass.
func (m *Manager) ResolvePeer(ctx context.Context, p tg.PeerClass) (Peer, error) {
	switch p := p.(type) {
	case *tg.PeerUser:
		v, ok, err := m.storage.Find(ctx, Key{
			Prefix: usersPrefix,
			ID:     p.UserID,
		})
		if err != nil {
			return nil, err
		}
		u, err := m.GetUser(ctx, &tg.InputUser{
			UserID:     p.UserID,
			AccessHash: v.AccessHash,
		})
		if !ok && tgerr.Is(err, tg.ErrUserIDInvalid) {
			return nil, &PeerNotFoundError{
				Peer: p,
			}
		}
		return u, err
	case *tg.PeerChat:
		c, err := m.GetChat(ctx, p.ChatID)
		return c, err
	case *tg.PeerChannel:
		v, ok, err := m.storage.Find(ctx, Key{
			Prefix: channelPrefix,
			ID:     p.ChannelID,
		})
		if err != nil {
			return nil, err
		}
		c, err := m.GetChannel(ctx, &tg.InputChannel{
			ChannelID:  p.ChannelID,
			AccessHash: v.AccessHash,
		})
		if !ok && tgerr.Is(err, tg.ErrChannelInvalid) {
			return nil, &PeerNotFoundError{
				Peer: p,
			}
		}
		return c, err
	default:
		return nil, errors.Errorf("unexpected type %T", p)
	}
}

// FromInputPeer gets Peer from tg.InputPeerClass.
func (m *Manager) FromInputPeer(ctx context.Context, p tg.InputPeerClass) (Peer, error) {
	switch p := p.(type) {
	case *tg.InputPeerSelf:
		return m.Self(ctx)
	case *tg.InputPeerChat:
		return m.GetChat(ctx, p.ChatID)
	case *tg.InputPeerUser:
		return m.GetUser(ctx, &tg.InputUser{
			UserID:     p.UserID,
			AccessHash: p.AccessHash,
		})
	case *tg.InputPeerChannel:
		return m.GetChannel(ctx, &tg.InputChannel{
			ChannelID:  p.ChannelID,
			AccessHash: p.AccessHash,
		})
	case *tg.InputPeerUserFromMessage:
		return m.GetUser(ctx, &tg.InputUserFromMessage{
			Peer:   p.Peer,
			MsgID:  p.MsgID,
			UserID: p.UserID,
		})
	case *tg.InputPeerChannelFromMessage:
		return m.GetChannel(ctx, &tg.InputChannelFromMessage{
			Peer:      p.Peer,
			MsgID:     p.MsgID,
			ChannelID: p.ChannelID,
		})
	default:
		return nil, errors.Errorf("unexpected type %T", p)
	}
}
