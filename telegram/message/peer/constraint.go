package peer

import (
	"context"
	"fmt"

	"golang.org/x/xerrors"

	"github.com/nnqq/td/tg"
)

// PromiseDecorator is a decorator of peer promise.
type PromiseDecorator = func(Promise) Promise

// ConstraintError is a peer resolve constraint error.
type ConstraintError struct {
	Expected string
	Got      tg.InputPeerClass
}

// Error implements error.
func (c *ConstraintError) Error() string {
	return fmt.Sprintf("expected %q, got %T", c.Expected, c.Got)
}

func tryUnpackConstraint(p tg.InputPeerClass, resolveErr error) (tg.InputPeerClass, error) {
	var constraintErr *ConstraintError
	if xerrors.As(resolveErr, &constraintErr) {
		return constraintErr.Got, nil
	}
	return p, resolveErr
}

// OnlyChannel returns Promise which returns error if resolved peer is not a channel.
func OnlyChannel(p Promise) Promise {
	return func(ctx context.Context) (tg.InputPeerClass, error) {
		resolved, err := tryUnpackConstraint(p(ctx))
		if err != nil {
			return nil, err
		}

		switch resolved.(type) {
		case *tg.InputPeerChannel, *tg.InputPeerChannelFromMessage:
			return resolved, nil
		default:
			return nil, &ConstraintError{
				Expected: "channel",
				Got:      resolved,
			}
		}
	}
}

// OnlyChat returns Promise which returns error if resolved peer is not a chat.
func OnlyChat(p Promise) Promise {
	return func(ctx context.Context) (tg.InputPeerClass, error) {
		resolved, err := tryUnpackConstraint(p(ctx))
		if err != nil {
			return nil, err
		}

		switch resolved.(type) {
		case *tg.InputPeerChat:
			return resolved, nil
		default:
			return nil, &ConstraintError{
				Expected: "chat",
				Got:      resolved,
			}
		}
	}
}

// OnlyUser returns Promise which returns error if resolved peer is not a user.
func OnlyUser(p Promise) Promise {
	return func(ctx context.Context) (tg.InputPeerClass, error) {
		resolved, err := tryUnpackConstraint(p(ctx))
		if err != nil {
			return nil, err
		}

		switch resolved.(type) {
		case *tg.InputPeerUser, *tg.InputPeerUserFromMessage, *tg.InputPeerSelf:
			return resolved, nil
		default:
			return nil, &ConstraintError{
				Expected: "user",
				Got:      resolved,
			}
		}
	}
}

// OnlyUserID returns Promise which returns error if resolved peer is not a user object with ID.
// Unlike OnlyUser, it returns error if resolved peer is tg.InputPeerSelf.
func OnlyUserID(p Promise) Promise {
	return func(ctx context.Context) (tg.InputPeerClass, error) {
		resolved, err := tryUnpackConstraint(p(ctx))
		if err != nil {
			return nil, err
		}

		switch resolved.(type) {
		case *tg.InputPeerUser, *tg.InputPeerUserFromMessage:
			return resolved, nil
		default:
			return nil, &ConstraintError{
				Expected: "userID",
				Got:      resolved,
			}
		}
	}
}
