package updates

import (
	"context"

	"github.com/gotd/log"

	"github.com/gotd/td/tg"
)

// shortMessageOptional is implemented by both *tg.UpdateShortMessage and
// *tg.UpdateShortChatMessage; it exposes the optional fields that may reference
// additional user peers (forward origin, inline bot, mention entities).
type shortMessageOptional interface {
	GetFwdFrom() (tg.MessageFwdHeader, bool)
	GetViaBotID() (int64, bool)
	GetEntities() ([]tg.MessageEntityClass, bool)
}

// referencedUserIDs returns the user IDs a short message refers to whose access
// hash must be known before the message can be resolved: the primary
// peer/sender, the forward origin and its saved-from peer (when users), the
// inline bot, and mention-name entities. selfID and zero IDs are never returned.
//
// The primary is the counterpart user for UpdateShortMessage (u.UserID) and the
// sender for UpdateShortChatMessage (u.FromID); for an outgoing chat message the
// sender is selfID and is therefore excluded. The basic-group chat and channel
// peers carry no user access hash and are not checked here; the reply-to peer is
// likewise deliberately omitted (rare cross-thread case).
func referencedUserIDs(selfID, primary int64, opt shortMessageOptional) []int64 {
	ids := make([]int64, 0, 4)
	add := func(id int64) {
		if id != 0 && id != selfID {
			ids = append(ids, id)
		}
	}

	add(primary)

	if fwd, ok := opt.GetFwdFrom(); ok {
		if from, ok := fwd.GetFromID(); ok {
			if u, ok := from.(*tg.PeerUser); ok {
				add(u.UserID)
			}
		}
		if saved, ok := fwd.GetSavedFromPeer(); ok {
			if u, ok := saved.(*tg.PeerUser); ok {
				add(u.UserID)
			}
		}
	}
	if via, ok := opt.GetViaBotID(); ok {
		add(via)
	}
	if ents, ok := opt.GetEntities(); ok {
		for _, e := range ents {
			if m, ok := e.(*tg.MessageEntityMentionName); ok {
				add(m.UserID)
			}
		}
	}

	return ids
}

// shortMessagePeersKnown reports whether every user peer referenced by a private
// short message already has a known access hash.
func (s *internalState) shortMessagePeersKnown(ctx context.Context, u *tg.UpdateShortMessage) bool {
	return s.userPeersKnown(ctx, referencedUserIDs(s.selfID, u.UserID, u))
}

// shortChatMessagePeersKnown reports whether every user peer referenced by a
// basic-group short message already has a known access hash. The basic-group
// chat itself has no access hash and is intentionally not checked.
func (s *internalState) shortChatMessagePeersKnown(ctx context.Context, u *tg.UpdateShortChatMessage) bool {
	return s.userPeersKnown(ctx, referencedUserIDs(s.selfID, u.FromID, u))
}

// userPeersKnown reports whether every referenced user ID is known: a user is
// known iff it is selfID or a full (non-min, non-zero) access hash has been seen
// for it (see saveUserHashes). A hasher error is treated as "not known", so the
// caller falls back to getDifference.
func (s *internalState) userPeersKnown(ctx context.Context, ids []int64) bool {
	for _, id := range ids {
		if id == s.selfID {
			continue
		}
		if _, found, err := s.userHasher.GetUserAccessHash(ctx, s.selfID, id); err != nil || !found {
			s.log.Debug(ctx, "User access hash unknown, forcing getDifference",
				log.Int64("user_id", id),
				log.Bool("hasher_error", err != nil),
				log.Error(err),
			)
			return false
		}
	}
	return true
}
