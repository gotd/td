package updates

import (
	"context"

	"github.com/go-faster/errors"
	"github.com/gotd/log"

	"github.com/gotd/td/tg"
)

// differenceContents are the contents of updates.difference or
// updates.differenceSlice.
type differenceContents struct {
	NewMessages          []tg.MessageClass
	NewEncryptedMessages []tg.EncryptedMessageClass
	OtherUpdates         []tg.UpdateClass
	Users                []tg.UserClass
	Chats                []tg.ChatClass
	State                tg.UpdatesState
}

// isDifferencePrelude reports whether u must be applied before the new
// messages of a difference, like TDLib process_get_difference_updates does:
// updateMessageID and updateStoryID resolve random IDs of sent messages and
// stories, updateEncryption and updateFolderPeers set up the chats the
// messages belong to.
func isDifferencePrelude(u tg.UpdateClass) bool {
	switch u.(type) {
	case *tg.UpdateMessageID,
		*tg.UpdateStoryID,
		*tg.UpdateEncryption,
		*tg.UpdateFolderPeers:
		return true
	default:
		return false
	}
}

// applyDifference applies the contents of a common difference and sets the
// local state to d.State.
//
// Updates of a difference are applied unconditionally ("force apply", TDLib
// process_updates(other_updates, force_apply=true)):
//
//   - common pts, qts and seq updates are dispatched directly, without the gap
//     check: the new messages of the difference occupy pts between the local
//     state and the other updates, so the gap check would postpone the other
//     updates and then drop them as outdated once the state is set;
//   - the known-peer check (messageUpdatesPeersKnown) is skipped: a difference
//     is the recovery path itself, re-requesting it for a min or missing user
//     returns the same difference and recurses without bound;
//   - channel updates are routed to their channel workers, which own the
//     channel pts sequence.
//
// Updates are dispatched in TDLib order: prelude updates (see
// isDifferencePrelude), new messages, new encrypted messages, then the rest of
// other updates, so that edits and reads of a message come after the message.
//
// It reports whether the difference contained updatePtsChanged: the caller
// must then refetch instead of trusting d.State (see TDLib on_update for
// updatePtsChanged, which forces getDifference).
func (s *internalState) applyDifference(
	ctx context.Context, d differenceContents, reason string,
) (ptsChanged bool, err error) {
	ctx, span := s.tracer.Start(ctx, "internalState.applyDifference")
	defer span.End()

	// Record entities before dispatching, so a user or channel recovered via
	// getDifference does not re-trigger getDifference for the next update.
	s.saveChannelHashes(ctx, d.Chats)
	s.saveUserHashes(ctx, d.Users)

	ents := entities{
		Users: d.Users,
		Chats: d.Chats,
	}
	updates := make([]tg.UpdateClass, 0,
		len(d.OtherUpdates)+len(d.NewMessages)+len(d.NewEncryptedMessages),
	)
	for _, u := range d.OtherUpdates {
		if isDifferencePrelude(u) {
			updates = append(updates, u)
		}
	}
	updates = append(updates, msgsToUpdates(d.NewMessages, false)...)
	updates = append(updates, encryptedMsgsToUpdates(d.NewEncryptedMessages)...)

	rest := make([]tg.UpdateClass, 0, len(d.OtherUpdates))
	for _, u := range d.OtherUpdates {
		if !isDifferencePrelude(u) {
			rest = append(rest, u)
		}
	}
	// Other updates of a difference are expected to be in pts order already,
	// sort them (stable) like any other batch to be safe.
	sortUpdatesByPts(rest)
	for _, u := range rest {
		if _, ok := u.(*tg.UpdatePtsChanged); ok {
			// Processed last, as in TDLib: the caller refetches.
			ptsChanged = true
			continue
		}
		if s.routeChannelUpdate(ctx, u, ents, d.State.Date) {
			continue
		}
		updates = append(updates, u)
	}

	if len(updates) > 0 {
		if err := s.dispatch(ctx, &tg.Updates{
			Updates: updates,
			Users:   d.Users,
			Chats:   d.Chats,
		}); err != nil {
			s.log.Error(ctx, "Handle updates error", log.Error(err))
		}
	}

	// Do not advance the state if the context is done: updates routed to
	// channel workers may have been lost, the next difference re-delivers them.
	if err := ctx.Err(); err != nil {
		return false, errors.Wrap(err, "apply difference")
	}

	if err := s.storage.SetState(ctx, s.selfID, State{}.fromRemote(&d.State)); err != nil {
		s.log.Warn(ctx, "SetState error", log.Error(err))
	}
	s.pts.SetState(d.State.Pts, reason)
	s.qts.SetState(d.State.Qts, reason)
	s.seq.SetState(d.State.Seq, reason)
	s.date = d.State.Date

	// Pending updates up to the new state were delivered by the difference.
	for _, box := range []*sequenceBox{s.pts, s.qts, s.seq} {
		if err := box.resyncPending(ctx); err != nil {
			return ptsChanged, errors.Wrap(err, "apply pending")
		}
	}
	return ptsChanged, nil
}
