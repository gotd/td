package updates

import (
	"go.uber.org/zap"
	"golang.org/x/net/context"

	"github.com/gotd/td/tg"
)

func (s *internalState) saveChannelHashes(ctx context.Context, chats []tg.ChatClass) {
	ctx, span := s.tracer.Start(ctx, "updates.saveChannelHashes")
	defer span.End()

	for _, c := range chats {
		switch c := c.(type) {
		case *tg.Channel:
			if c.Min {
				continue
			}

			if hash, ok := c.GetAccessHash(); ok {
				if _, ok = s.channels[c.ID]; ok {
					continue
				}
				s.log.Debug("New channel access hash",
					zap.Int64("channel_id", c.ID),
					zap.String("title", c.Title),
				)
				if err := s.hasher.SetChannelAccessHash(ctx, s.selfID, c.ID, hash); err != nil {
					s.log.Error("SetChannelState error", zap.Error(err))
				}
			}
		case *tg.ChannelForbidden:
			if _, ok := s.channels[c.ID]; ok {
				continue
			}
			s.log.Debug("New channel access hash",
				zap.Int64("channel_id", c.ID),
				zap.String("title", c.Title),
			)
			if err := s.hasher.SetChannelAccessHash(ctx, s.selfID, c.ID, c.AccessHash); err != nil {
				s.log.Error("SetChannelState error", zap.Error(err))
			}
		}
	}
}

func (s *internalState) restoreAccessHash(ctx context.Context, channelID int64, date int) (accessHash int64, ok bool) {
	ctx, span := s.tracer.Start(ctx, "updates.restoreAccessHash")
	defer span.End()

	diff, err := s.client.UpdatesGetDifference(ctx, &tg.UpdatesGetDifferenceRequest{
		Pts:  s.pts.State(),
		Qts:  s.qts.State(),
		Date: date,
	})
	if err != nil {
		s.log.Error("UpdatesGetDifference error", zap.Error(err))
		return 0, false
	}

	var chats []tg.ChatClass
	switch diff := diff.(type) {
	case *tg.UpdatesDifference:
		chats = diff.Chats
	case *tg.UpdatesDifferenceSlice:
		chats = diff.Chats
	}

	s.saveChannelHashes(ctx, chats)
	for _, c := range chats {
		switch c := c.(type) {
		case *tg.Channel:
			if c.Min {
				continue
			}

			if c.ID != channelID {
				continue
			}

			if hash, ok := c.GetAccessHash(); ok {
				return hash, true
			}

		case *tg.ChannelForbidden:
			if c.ID != channelID {
				continue
			}

			return c.AccessHash, true
		}
	}

	return 0, false
}
