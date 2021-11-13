package unpack

import (
	"github.com/go-faster/errors"

	"github.com/gotd/td/telegram/internal/upconv"
	"github.com/gotd/td/tg"
)

func extractClass(u tg.UpdateClass) (tg.MessageClass, bool) {
	switch v := u.(type) {
	case *tg.UpdateNewMessage:
		return v.Message, true
	case *tg.UpdateNewChannelMessage:
		return v.Message, true
	default:
		return nil, false
	}
}

// MessageClass tries to unpack sent message and returns it as MessageClass.
func MessageClass(u tg.UpdatesClass, err error) (tg.MessageClass, error) {
	if err != nil {
		return nil, err
	}

	var updates []tg.UpdateClass
	switch v := u.(type) {
	case *tg.UpdateShortMessage:
		short := upconv.ShortMessage(v)
		updates = []tg.UpdateClass{short.Update}
	case *tg.UpdateShortChatMessage:
		short := upconv.ShortChatMessage(v)
		updates = []tg.UpdateClass{short.Update}
	case *tg.UpdateShortSentMessage:
		short := upconv.ShortSentMessage(v)
		updates = []tg.UpdateClass{short.Update}
	case *tg.UpdateShort:
		updates = []tg.UpdateClass{v.Update}
	case *tg.UpdatesCombined:
		updates = v.GetUpdates()
	case *tg.Updates:
		updates = v.GetUpdates()
	default:
		return nil, errors.Errorf("unexpected type %T", u)
	}

	for _, update := range updates {
		if msg, ok := extractClass(update); ok {
			return msg, nil
		}
	}

	return nil, errors.Errorf("bad updates result %+v", updates)
}

// Message tries to unpack sent message and returns it as Message.
func Message(u tg.UpdatesClass, err error) (*tg.Message, error) {
	msg, err := MessageClass(u, err)
	if err != nil {
		return nil, err
	}

	m, ok := msg.(*tg.Message)
	if !ok {
		return nil, errors.Errorf("unexpected type %T", msg)
	}

	return m, nil
}

// MessageID tries to unpack sent message and returns message id.
func MessageID(u tg.UpdatesClass, err error) (int, error) {
	msg, err := MessageClass(u, err)
	if err != nil {
		return 0, err
	}

	return msg.GetID(), nil
}
