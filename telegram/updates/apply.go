package updates

import (
	"context"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/tg"
)

type updateDiff interface {
	GetOtherUpdates() []tg.UpdateClass
	GetNewMessages() []tg.MessageClass
	GetUsers() []tg.UserClass
	GetChats() []tg.ChatClass
}

type channelTooLong struct {
	Pts     int
	Channel *tg.InputChannel
}

func (m *Manager) applyDiff(ctx context.Context, d updateDiff) []channelTooLong {
	// Adapting update to Handle() input.
	var updates []tg.UpdateClass
	var r []channelTooLong

	channels := tg.ChatClassArray(d.GetChats()).ChannelToMap()
	for _, u := range d.GetOtherUpdates() {
		updates = append(updates, u)

		if tooLong, ok := u.(*tg.UpdateChannelTooLong); ok {
			channel, ok := channels[tooLong.ChannelID]
			if !ok {
				continue
			}

			r = append(r, channelTooLong{
				Pts:     tooLong.Pts,
				Channel: channel.AsInput(),
			})
		}
	}
	for _, m := range d.GetNewMessages() {
		updates = append(updates, &tg.UpdateNewMessage{
			Message: m,
		})
	}

	m.apply(ctx, &tg.Updates{
		Updates: updates,
		Users:   d.GetUsers(),
		Chats:   d.GetChats(),
	})
	return r
}

func (m *Manager) apply(ctx context.Context, u *tg.Updates) {
	if err := m.handler.Handle(ctx, u); err != nil {
		m.log.Warn("Handle updates failed", zap.Error(err))
	}
}

func (m *Manager) applyUpdates(ctx context.Context, upds *tg.Updates) error { // nolint:gocognit
	type ptsUpdate interface {
		GetPts() int
		GetPtsCount() int
	}

	gapDetected := false
	// ID -> pts
	channels := map[int]ptsUpdate{}
	entities := peer.EntitiesFromResult(upds)
	// See https://core.telegram.org/api/updates#update-handling.
	if err := m.storage.Acquire(ctx, "pts", func(box Box) error {
		local, err := box.Load(ctx)
		if err != nil {
			return xerrors.Errorf("load pts: %w", err)
		}

		// Filter from SliceTricks.
		b := upds.Updates[:0]
		var smallestPts ptsUpdate
		for _, x := range upds.Updates {
			upd, ok := x.(ptsUpdate)
			if !ok {
				b = append(b, x)
				continue
			}
			pts := upd.GetPts()
			ptsCount := upd.GetPtsCount()

			channelID := -1
			switch upd := x.(type) {
			case interface{ GetChannelID() int }:
				channelID = upd.GetChannelID()
			case interface{ GetMessage() tg.MessageClass }:
				msg, ok := upd.GetMessage().AsNotEmpty()
				if ok {
					peerID, ok := msg.GetPeerID().(*tg.PeerChannel)
					if !ok {
						break
					}

					channelID = peerID.ChannelID
				}
			}

			if channelID >= 0 {
				v, ok := channels[channelID]
				if !ok || v.GetPts() > pts {
					channels[channelID] = upd
				}
			} else {
				// Skip already applied.
				if local+ptsCount > pts {
					continue
				}

				if smallestPts == nil || smallestPts.GetPts() > pts {
					smallestPts = upd
				}
			}

			b = append(b, x)
		}

		upds.Updates = b
		if smallestPts != nil {
			gapDetected = local+smallestPts.GetPtsCount() < smallestPts.GetPts()
		}

		return nil
	}); err != nil {
		return xerrors.Errorf("apply updates: %w", err)
	}

	if gapDetected {
		if err := m.gapCommon(ctx); err != nil {
			return xerrors.Errorf("sync gap: %w", err)
		}
	} else {
		m.apply(ctx, upds)
	}

	for id, update := range channels {
		channel, err := entities.ExtractChannel(&tg.PeerChannel{
			ChannelID: id,
		})
		if err != nil {
			m.log.Warn("Channel not found", zap.Int("id", id))
			continue
		}

		if err := m.syncChannel(ctx, &tg.InputChannel{
			ChannelID:  channel.ChannelID,
			AccessHash: channel.AccessHash,
		}, update.GetPts()); err != nil {
			return xerrors.Errorf("sync channel %d: %w", id, err)
		}
	}

	return nil
}
