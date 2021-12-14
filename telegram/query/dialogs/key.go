package dialogs

import (
	"github.com/go-faster/errors"

	"github.com/gotd/td/tg"
)

// PeerKind represents peer kind.
type PeerKind int

const (
	// User is a private chat with user.
	User PeerKind = iota
	// Chat is a legacy chat.
	Chat
	// Channel is a supergroup/channel.
	Channel
)

// DialogKey is a generic peer key.
type DialogKey struct {
	Kind       PeerKind
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
		return errors.Errorf("unexpected type %T", peer)
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
		return errors.Errorf("unexpected type %T", peer)
	}

	return nil
}
