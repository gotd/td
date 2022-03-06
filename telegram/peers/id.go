package peers

import (
	"context"

	"github.com/go-faster/errors"

	"github.com/gotd/td/constant"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

// ResolveTDLibID creates Peer using given constant.TDLibPeerID.
func (m *Manager) ResolveTDLibID(ctx context.Context, peerID constant.TDLibPeerID) (p Peer, err error) {
	switch {
	case peerID.IsUser():
		p, err = m.ResolveUserID(ctx, peerID.ToPlain())
	case peerID.IsChat():
		p, err = m.ResolveChatID(ctx, peerID.ToPlain())
	case peerID.IsChannel():
		p, err = m.ResolveChannelID(ctx, peerID.ToPlain())
	default:
		return nil, errors.Errorf("invalid ID %d", peerID)
	}
	return p, err
}

// ResolveUserID creates User using given id.
func (m *Manager) ResolveUserID(ctx context.Context, id int64) (User, error) {
	v, ok, err := m.storage.Find(ctx, Key{
		Prefix: usersPrefix,
		ID:     id,
	})
	if err != nil {
		return User{}, err
	}
	u, err := m.GetUser(ctx, &tg.InputUser{
		UserID:     id,
		AccessHash: v.AccessHash,
	})
	if !ok && tgerr.Is(err, tg.ErrUserIDInvalid) {
		return User{}, &PeerNotFoundError{
			Peer: &tg.PeerUser{UserID: id},
		}
	}
	return u, err
}

// ResolveChatID creates Chat using given id.
func (m *Manager) ResolveChatID(ctx context.Context, id int64) (Chat, error) {
	c, err := m.GetChat(ctx, id)
	return c, err
}

// ResolveChannelID creates Channel using given id.
func (m *Manager) ResolveChannelID(ctx context.Context, id int64) (Channel, error) {
	v, ok, err := m.storage.Find(ctx, Key{
		Prefix: channelPrefix,
		ID:     id,
	})
	if err != nil {
		return Channel{}, err
	}
	c, err := m.GetChannel(ctx, &tg.InputChannel{
		ChannelID:  id,
		AccessHash: v.AccessHash,
	})
	if !ok && tgerr.Is(err, tg.ErrChannelInvalid) {
		return Channel{}, &PeerNotFoundError{
			Peer: &tg.PeerChannel{ChannelID: id},
		}
	}
	return c, err
}
