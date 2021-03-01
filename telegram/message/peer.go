package message

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/gotd/td/tg"
)

func (s *Sender) builder(peer peerPromise) *Builder {
	return &Builder{
		sender: s,
		peer:   peer,
	}
}

// Peer uses given peer to create new message builder.
func (s *Sender) Peer(peer tg.InputPeerClass) *Builder {
	return s.builder(func(ctx context.Context) (tg.InputPeerClass, error) {
		return peer, nil
	})
}

// Self creates a new message builder to send it to yourself.
// It means that message will be sent to your Saved Messages folder.
func (s *Sender) Self() *Builder {
	return s.Peer(&tg.InputPeerSelf{})
}

// AnswerableMessageUpdate represents update which can be used to answer.
type AnswerableMessageUpdate interface {
	GetMessage() tg.MessageClass
	GetPts() int
}

type entities struct {
	Users    map[int]*tg.User
	Chats    map[int]*tg.Chat
	Channels map[int]*tg.Channel
}

func findPeer(uctx entities, peerID tg.PeerClass) (tg.InputPeerClass, error) {
	var peer tg.InputPeerClass
	switch p := peerID.(type) {
	case *tg.PeerUser: // peerUser#9db1bc6d
		dialog, ok := uctx.Users[p.UserID]
		if !ok {
			return nil, xerrors.Errorf("user %d not found in Update", p.UserID)
		}

		peer = &tg.InputPeerUser{
			UserID:     dialog.ID,
			AccessHash: dialog.AccessHash,
		}
	case *tg.PeerChat: // peerChat#bad0e5bb
		dialog, ok := uctx.Chats[p.ChatID]
		if !ok {
			return nil, xerrors.Errorf("chat %d not found in Update", p.ChatID)
		}

		peer = &tg.InputPeerChat{
			ChatID: dialog.ID,
		}
	case *tg.PeerChannel: // peerChannel#bddde532
		dialog, ok := uctx.Channels[p.ChannelID]
		if !ok {
			return nil, xerrors.Errorf("channel %d not found in Update", p.ChannelID)
		}

		peer = &tg.InputPeerChannel{
			ChannelID:  dialog.ID,
			AccessHash: dialog.AccessHash,
		}
	}

	return peer, nil
}

// Answer uses given message update to create message for same chat.
func (s *Sender) Answer(uctx tg.UpdateContext, upd AnswerableMessageUpdate) *Builder {
	return s.builder(func(ctx context.Context) (tg.InputPeerClass, error) {
		updMsg := upd.GetMessage()
		msg, ok := updMsg.AsNotEmpty()
		if !ok {
			emptyMsg, ok := updMsg.(*tg.MessageEmpty)
			if !ok {
				return nil, xerrors.Errorf("unexpected type %T", updMsg)
			}

			peer, ok := emptyMsg.GetPeerID()
			if !ok {
				return nil, xerrors.Errorf("got %T with empty PeerID", updMsg)
			}

			return findPeer(entities{
				Users:    uctx.Users,
				Chats:    uctx.Chats,
				Channels: uctx.Channels,
			}, peer)
		}

		return findPeer(entities{
			Users:    uctx.Users,
			Chats:    uctx.Chats,
			Channels: uctx.Channels,
		}, msg.GetPeerID())
	})
}

// Reply uses given message update to create message for same chat and create a reply.
// Shorthand for
//
// 	sender.Answer(uctx, upd).ReplyMsg(upd.GetMessage())
//
func (s *Sender) Reply(uctx tg.UpdateContext, upd AnswerableMessageUpdate) *Builder {
	return s.Answer(uctx, upd).ReplyMsg(upd.GetMessage())
}
