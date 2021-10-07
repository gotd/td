package peer

import (
	"golang.org/x/xerrors"

	"github.com/nnqq/td/tg"
)

// Kind represents peer kind.
type Kind int

const (
	// User is a private chat with user.
	User Kind = iota
	// Chat is a legacy chat.
	Chat
	// Channel is a supergroup/channel.
	Channel
)

// DialogKey is a generic peer key.
type DialogKey struct {
	Kind       Kind
	ID         int64
	AccessHash int64
}

// FromInputPeer fills key using given peer.
func (d *DialogKey) FromInputPeer(peer tg.InputPeerClass) error {
	switch v := peer.(type) {
	case *tg.InputPeerUser:
		d.Kind = User
		d.ID = v.UserID
		d.AccessHash = v.AccessHash
	case *tg.InputPeerChat:
		d.Kind = Chat
		d.ID = v.ChatID
	case *tg.InputPeerChannel:
		d.Kind = Channel
		d.ID = v.ChannelID
		d.AccessHash = v.AccessHash
	default:
		return xerrors.Errorf("unexpected type %T", peer)
	}

	return nil
}

// FromPeer fills key using given peer.
func (d *DialogKey) FromPeer(peer tg.PeerClass) error {
	switch v := peer.(type) {
	case *tg.PeerUser:
		d.Kind = User
		d.ID = v.UserID
	case *tg.PeerChat:
		d.Kind = Chat
		d.ID = v.ChatID
	case *tg.PeerChannel:
		d.Kind = Channel
		d.ID = v.ChannelID
	default:
		return xerrors.Errorf("unexpected type %T", peer)
	}

	return nil
}
