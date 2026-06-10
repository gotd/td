package updates

import (
	"context"

	"github.com/gotd/td/tg"
)

// messageUserIDs returns the user IDs a full (non-channel) message refers to
// whose access hash must be known before the message can be resolved: the
// dialog peer (when a user), the sender, the forward origin and its saved-from
// peer (when users), the inline bot, and mention-name entities. selfID and zero
// IDs are never returned. Mirrors referencedUserIDs for *tg.Message.
func messageUserIDs(selfID int64, msg *tg.Message) []int64 {
	ids := make([]int64, 0, 4)
	add := func(id int64) {
		if id != 0 && id != selfID {
			ids = append(ids, id)
		}
	}
	addPeer := func(p tg.PeerClass) {
		if u, ok := p.(*tg.PeerUser); ok {
			add(u.UserID)
		}
	}

	addPeer(msg.PeerID)
	if msg.FromID != nil {
		addPeer(msg.FromID)
	}
	if fwd, ok := msg.GetFwdFrom(); ok {
		if from, ok := fwd.GetFromID(); ok {
			addPeer(from)
		}
		if saved, ok := fwd.GetSavedFromPeer(); ok {
			addPeer(saved)
		}
	}
	if via, ok := msg.GetViaBotID(); ok {
		add(via)
	}
	if ents, ok := msg.GetEntities(); ok {
		for _, e := range ents {
			if m, ok := e.(*tg.MessageEntityMentionName); ok {
				add(m.UserID)
			}
		}
	}

	return ids
}

// messageUpdatesPeersKnown reports whether every non-channel message-bearing
// update in the batch references only user peers with a known access hash.
//
// It covers *tg.UpdateNewMessage and *tg.UpdateEditMessage (private/basic-group
// messages). Channel message updates (*tg.UpdateNewChannelMessage and friends)
// are intentionally excluded: their senders are routinely min in megagroups, so
// recovering them via a global getDifference would be prohibitively expensive;
// the channel-difference machinery owns that path. This mirrors TDLib
// is_acceptable_update restricted to the non-channel get_difference branch (see
// docs/research/telegram-min-updates-discipline.md §9).
func (s *internalState) messageUpdatesPeersKnown(ctx context.Context, updates []tg.UpdateClass) bool {
	for _, u := range updates {
		var msg tg.MessageClass
		switch upd := u.(type) {
		case *tg.UpdateNewMessage:
			msg = upd.Message
		case *tg.UpdateEditMessage:
			msg = upd.Message
		default:
			continue
		}
		m, ok := msg.(*tg.Message)
		if !ok {
			continue
		}
		if !s.userPeersKnown(ctx, messageUserIDs(s.selfID, m)) {
			return false
		}
	}
	return true
}
