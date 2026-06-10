package messages

import (
	"context"
	"sort"

	"github.com/go-faster/errors"

	"github.com/gotd/td/tg"
)

// maxMediaGroupSize is the maximum number of messages in a media group (album).
const maxMediaGroupSize = 10

// GetMediaGroup fetches all messages belonging to the same media group (album)
// as the message with the given id in peer, ordered by message ID.
//
// Album messages are consecutive and the anchor may be anywhere within the
// group, so a window of messages around msgID is fetched and filtered by the
// shared grouped_id.
//
// Returns an error if the message is not found or is not part of a media group.
//
// See https://core.telegram.org/api/files#albums-grouped-media.
func (q *QueryBuilder) GetMediaGroup(
	ctx context.Context,
	peer tg.InputPeerClass,
	msgID int,
) ([]*tg.Message, error) {
	ids := make([]tg.InputMessageClass, 0, 2*maxMediaGroupSize-1)
	for id := msgID - (maxMediaGroupSize - 1); id <= msgID+(maxMediaGroupSize-1); id++ {
		if id <= 0 {
			continue
		}
		ids = append(ids, &tg.InputMessageID{ID: id})
	}

	raw, err := q.getMessages(ctx, peer, ids)
	if err != nil {
		return nil, errors.Wrap(err, "get messages")
	}

	byID := make(map[int]*tg.Message, len(raw))
	for _, m := range raw {
		if msg, ok := m.(*tg.Message); ok {
			byID[msg.ID] = msg
		}
	}

	anchor, ok := byID[msgID]
	if !ok {
		return nil, errors.Errorf("message %d not found", msgID)
	}
	groupedID, ok := anchor.GetGroupedID()
	if !ok {
		return nil, errors.Errorf("message %d is not part of a media group", msgID)
	}

	var group []*tg.Message
	for _, msg := range byID {
		if id, ok := msg.GetGroupedID(); ok && id == groupedID {
			group = append(group, msg)
		}
	}
	sort.Slice(group, func(i, j int) bool {
		return group[i].ID < group[j].ID
	})

	return group, nil
}

func (q *QueryBuilder) getMessages(
	ctx context.Context,
	peer tg.InputPeerClass,
	ids []tg.InputMessageClass,
) ([]tg.MessageClass, error) {
	var (
		res tg.MessagesMessagesClass
		err error
	)
	if ch, ok := peer.(*tg.InputPeerChannel); ok {
		res, err = q.raw.ChannelsGetMessages(ctx, &tg.ChannelsGetMessagesRequest{
			Channel: &tg.InputChannel{
				ChannelID:  ch.ChannelID,
				AccessHash: ch.AccessHash,
			},
			ID: ids,
		})
	} else {
		res, err = q.raw.MessagesGetMessages(ctx, ids)
	}
	if err != nil {
		return nil, err
	}
	modified, ok := res.AsModified()
	if !ok {
		return nil, errors.Errorf("unexpected response %T", res)
	}
	return modified.GetMessages(), nil
}
