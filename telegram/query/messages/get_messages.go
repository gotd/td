package messages

import (
	"context"

	"github.com/go-faster/errors"

	"github.com/gotd/td/telegram/message/peer"
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

// GetMessagesWithEntities is like GetMessages, but returns non-empty
// messages as iterator elements, with users and chats of the response and
// input peers of message chats.
func (q *QueryBuilder) GetMessagesWithEntities(
	ctx context.Context,
	p tg.InputPeerClass,
	ids ...int,
) ([]Elem, error) {
	if len(ids) == 0 {
		return nil, errors.New("no message ids given")
	}

	input := make([]tg.InputMessageClass, 0, len(ids))
	for _, id := range ids {
		input = append(input, &tg.InputMessageID{ID: id})
	}

	res, err := q.getMessagesResult(ctx, p, input)
	if err != nil {
		return nil, errors.Wrap(err, "get messages")
	}
	entities := peer.EntitiesFromResult(&tg.MessagesMessages{
		Users: res.GetUsers(),
		Chats: res.GetChats(),
	})

	var elems []Elem
	for _, m := range res.GetMessages() {
		msg, ok := m.AsNotEmpty()
		if !ok {
			continue
		}
		msgPeer, err := entities.ExtractPeer(msg.GetPeerID())
		if err != nil {
			msgPeer = &tg.InputPeerEmpty{}
		}
		elems = append(elems, Elem{Msg: msg, Peer: msgPeer, Entities: entities})
	}
	return elems, nil
}
