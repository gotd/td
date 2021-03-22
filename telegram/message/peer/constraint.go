package peer

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/gotd/td/tg"
)

// PromiseDecorator is a decorator of peer promise.
type PromiseDecorator = func(Promise) Promise

// OnlyChannel returns Promise which returns error if resolved peer is not a channel.
func OnlyChannel(p Promise) Promise {
	return func(ctx context.Context) (tg.InputPeerClass, error) {
		resolved, err := p(ctx)
		if err != nil {
			return nil, err
		}

		switch resolved.(type) {
		case *tg.InputPeerChannel, *tg.InputPeerChannelFromMessage:
			return resolved, nil
		default:
			return nil, xerrors.Errorf("unexpected type %T", resolved)
		}
	}
}

// OnlyChat returns Promise which returns error if resolved peer is not a chat.
func OnlyChat(p Promise) Promise {
	return func(ctx context.Context) (tg.InputPeerClass, error) {
		resolved, err := p(ctx)
		if err != nil {
			return nil, err
		}

		switch resolved.(type) {
		case *tg.InputPeerChat:
			return resolved, nil
		default:
			return nil, xerrors.Errorf("unexpected type %T", resolved)
		}
	}
}

// OnlyUser returns Promise which returns error if resolved peer is not a user.
func OnlyUser(p Promise) Promise {
	return func(ctx context.Context) (tg.InputPeerClass, error) {
		resolved, err := p(ctx)
		if err != nil {
			return nil, err
		}

		switch resolved.(type) {
		case *tg.InputPeerUser, *tg.InputPeerUserFromMessage:
			return resolved, nil
		default:
			return nil, xerrors.Errorf("unexpected type %T", resolved)
		}
	}
}
