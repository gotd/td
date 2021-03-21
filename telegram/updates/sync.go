package updates

import (
	"context"
	"time"

	"golang.org/x/xerrors"

	"github.com/gotd/td/tg"
)

const commonStorageName = "pts"

func (m *Manager) syncCommon(ctx context.Context, remote int) error { // nolint:gocognit
	return m.storage.Acquire(ctx, commonStorageName, func(box Box) error {
		// Other start syncing.
		for {
			local, err := box.Load(ctx)
			if err != nil {
				return xerrors.Errorf("get local state: %w", err)
			}

			switch {
			case local == 0: // Initial state.
				return box.Commit(ctx, remote)
			case local >= remote: // Already synced.
				return nil
			}

			diff, err := m.raw.UpdatesGetDifference(ctx, &tg.UpdatesGetDifferenceRequest{
				Date: int(time.Now().Unix()),
				Pts:  local,
				Qts:  0, // No secret chats
			})
			if err != nil {
				return xerrors.Errorf("get difference: %w", err)
			}

			switch d := diff.(type) {
			case *tg.UpdatesDifference:
				channels := m.applyDiff(ctx, d)
				if err := box.Commit(ctx, remote); err != nil {
					return xerrors.Errorf("commit pts: %w", err)
				}

				for _, c := range channels {
					if err := m.syncChannel(ctx, c.Channel, c.Pts); err != nil {
						return xerrors.Errorf("sync channel %d", c.Channel.ChannelID)
					}
				}

				return nil
			case *tg.UpdatesDifferenceSlice:
				channels := m.applyDiff(ctx, d)
				if err := box.Commit(ctx, d.IntermediateState.Pts); err != nil {
					return xerrors.Errorf("commit pts: %w", err)
				}

				for _, c := range channels {
					if err := m.syncChannel(ctx, c.Channel, c.Pts); err != nil {
						return xerrors.Errorf("sync channel %d", c.Channel.ChannelID)
					}
				}

			case *tg.UpdatesDifferenceTooLong:
				m.log.Warn("Got UpdatesDifferenceTooLong, additional sync needed")
				return nil
			default:
				return nil
			}
		}
	})
}

func (m *Manager) syncChannel(ctx context.Context, channel *tg.InputChannel, remote int) error {
	return m.storage.Acquire(ctx, channelKey(*channel).String(), func(box Box) error {
		for {
			local, err := box.Load(ctx)
			if err != nil {
				return xerrors.Errorf("get local state: %w", err)
			}

			if remote > 0 {
				switch {
				case local == 0: // Initial state.
					return box.Commit(ctx, remote)
				case local >= remote: // Already synced.
					return nil
				}
			}

			diff, err := m.raw.UpdatesGetChannelDifference(ctx, &tg.UpdatesGetChannelDifferenceRequest{
				Channel: channel,
				Pts:     local,
				Filter:  &tg.ChannelMessagesFilterEmpty{},
			})
			if err != nil {
				return xerrors.Errorf("get difference: %w", err)
			}

			switch d := diff.(type) {
			case *tg.UpdatesChannelDifference:
				m.applyDiff(ctx, d)
				if err := box.Commit(ctx, d.Pts); err != nil {
					return xerrors.Errorf("commit channel %d pts: %w", channel.ChannelID, err)
				}
				remote = d.Pts

				if d.Timeout > 0 {
					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-time.After(time.Duration(d.Timeout) * time.Second):
					}
				}

				if d.Final {
					return nil
				}
			case *tg.UpdatesChannelDifferenceTooLong:
				m.log.Warn("Got UpdatesChannelDifferenceTooLong, additional sync needed")
				return nil
			default:
				return nil
			}
		}
	})
}
