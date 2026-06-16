package updates

import (
	"context"

	"github.com/gotd/log"

	"github.com/gotd/td/tg"
)

// Updates carry the raw entities from the server, which may be "min"
// constructors (Channel.Min / User.Min). A min entity's access hash is valid
// only for a narrow set of operations (e.g. inputPeerPhotoFileLocation) and
// fails normal RPCs such as messages.sendMessage with CHANNEL_INVALID. Since the
// manager already records full hashes from non-min observations
// (saveChannelHashes / saveUserHashes), we can swap a min hash for the known
// full one without any network round-trip, so user code that builds an
// InputPeer straight from update entities (e.g. e.Channels[id].AsInputPeer())
// works in every chat.
//
// The Min flag itself is left set: a min entity may still carry incomplete
// fields, and downstream consumers (e.g. peers.Manager.applyChats) rely on it to
// decide whether to persist the peer. Only the access hash is corrected, and
// only on a copy, so the original entity shared with other code is untouched.
//
// See https://github.com/gotd/td/issues/1553.

// dispatch hands an updates batch to the user handler after backfilling any min
// channel or user entities with the full access hash from the access-hash store.
func (s *internalState) dispatch(ctx context.Context, u *tg.Updates) error {
	u.Chats = backfillMinChats(ctx, s.log, s.hasher, s.selfID, u.Chats)
	u.Users = backfillMinUsers(ctx, s.log, s.userHasher, s.selfID, u.Users)
	return s.handler.Handle(ctx, u)
}

// dispatch mirrors internalState.dispatch for the per-channel worker, which owns
// delivery of channel message updates — the exact path in issue #1553.
func (s *channelState) dispatch(ctx context.Context, u *tg.Updates) error {
	u.Chats = backfillMinChats(ctx, s.log, s.hasher, s.selfID, u.Chats)
	u.Users = backfillMinUsers(ctx, s.log, s.userHasher, s.selfID, u.Users)
	return s.handler.Handle(ctx, u)
}

// backfillMinChats returns chats with the access hash of every min channel
// replaced by the full hash from the access-hash store, when one is known.
// Channels left unknown keep their min hash unchanged (no regression).
//
// The input slice and its entities are never mutated: the same slice may be
// shared with another goroutine (e.g. channelState.getDifference also hands
// diff.Chats to sendOut). Only when a replacement is needed is a fresh slice
// allocated and the affected entities copied; otherwise the original slice is
// returned as-is, keeping the common path allocation-free.
func backfillMinChats(ctx context.Context, lg log.Helper, hasher ChannelAccessHasher, selfID int64, chats []tg.ChatClass) []tg.ChatClass {
	out := chats
	for i, c := range chats {
		ch, ok := c.(*tg.Channel)
		if !ok || !ch.Min {
			continue
		}
		hash, found, err := hasher.GetChannelAccessHash(ctx, selfID, ch.ID)
		if err != nil {
			lg.Error(ctx, "GetChannelAccessHash error",
				log.Error(err), log.Int64("channel_id", ch.ID))
			continue
		}
		if !found {
			continue
		}
		if &out[0] == &chats[0] {
			out = make([]tg.ChatClass, len(chats))
			copy(out, chats)
		}
		cp := *ch
		cp.SetAccessHash(hash)
		out[i] = &cp
	}
	return out
}

// backfillMinUsers mirrors backfillMinChats for min users.
func backfillMinUsers(ctx context.Context, lg log.Helper, hasher UserAccessHasher, selfID int64, users []tg.UserClass) []tg.UserClass {
	out := users
	for i, u := range users {
		usr, ok := u.(*tg.User)
		if !ok || !usr.Min {
			continue
		}
		hash, found, err := hasher.GetUserAccessHash(ctx, selfID, usr.ID)
		if err != nil {
			lg.Error(ctx, "GetUserAccessHash error",
				log.Error(err), log.Int64("user_id", usr.ID))
			continue
		}
		if !found {
			continue
		}
		if &out[0] == &users[0] {
			out = make([]tg.UserClass, len(users))
			copy(out, users)
		}
		cp := *usr
		cp.SetAccessHash(hash)
		out[i] = &cp
	}
	return out
}
