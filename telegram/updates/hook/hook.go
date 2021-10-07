// Package hook contains telegram update hook middleware.
package hook

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/telegram"
	"github.com/nnqq/td/tg"
)

// UpdateHook middleware is called on each tg.UpdatesClass method result.
//
// Function is called before invoker return. Returned error will be wrapped
// and returned as InvokeRaw result.
type UpdateHook func(ctx context.Context, u tg.UpdatesClass) error

// Handle implements telegram.Middleware.
func (h UpdateHook) Handle(next tg.Invoker) telegram.InvokeFunc {
	return func(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
		if err := next.Invoke(ctx, input, output); err != nil {
			return err
		}
		if u, ok := output.(*tg.UpdatesBox); ok {
			if err := h(ctx, u.Updates); err != nil {
				return xerrors.Errorf("hook: %w", err)
			}
		}

		return nil
	}
}
