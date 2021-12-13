package peers

import (
	"context"

	"github.com/go-faster/errors"

	"github.com/gotd/td/tg"
)

// Peer represents generic peer.
type Peer interface {
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

// Peer creates Peer using given tg.PeerClass.
func (m *Manager) Peer(ctx context.Context, p tg.PeerClass) (Peer, error) {
	switch p := p.(type) {
	case *tg.PeerUser:
		v, ok, err := m.storage.Find(ctx, usersPrefix, p.UserID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, &PeerNotFoundError{
				Peer: p,
			}
		}

		u, err := m.GetUser(ctx, &tg.InputUser{
			UserID:     p.UserID,
			AccessHash: v.AccessHash,
		})
		return u, err
	case *tg.PeerChat:
		c, err := m.GetChat(ctx, p.ChatID)
		return c, err
	case *tg.PeerChannel:
		v, ok, err := m.storage.Find(ctx, usersPrefix, p.ChannelID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, &PeerNotFoundError{
				Peer: p,
			}
		}

		c, err := m.GetChannel(ctx, &tg.InputChannel{
			ChannelID:  p.ChannelID,
			AccessHash: v.AccessHash,
		})
		return c, err
	default:
		return nil, errors.Errorf("unexpected type %T", p)
	}
}
