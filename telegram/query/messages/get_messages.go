package messages

import (
	"context"

	"github.com/go-faster/errors"

	"github.com/gotd/td/tg"
)

// GetMessages fetches messages by their IDs in the given peer.
//
// It automatically uses channels.getMessages for channel peers and
// messages.getMessages otherwise, so the caller does not need to special-case
// channels (passing channel message IDs to messages.getMessages returns
// unrelated messages).
//
// The returned slice preserves the contents of the server response, which may
// include tg.MessageEmpty for IDs that do not exist or are inaccessible.
//
// See https://core.telegram.org/method/messages.getMessages
// and https://core.telegram.org/method/channels.getMessages.
func (q *QueryBuilder) GetMessages(
	ctx context.Context,
	peer tg.InputPeerClass,
	ids ...int,
) ([]tg.MessageClass, error) {
	if len(ids) == 0 {
		return nil, errors.New("no message ids given")
	}

	input := make([]tg.InputMessageClass, 0, len(ids))
	for _, id := range ids {
		input = append(input, &tg.InputMessageID{ID: id})
	}

	raw, err := q.getMessages(ctx, peer, input)
	if err != nil {
		return nil, errors.Wrap(err, "get messages")
	}
	return raw, nil
}
