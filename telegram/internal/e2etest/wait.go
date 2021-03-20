package e2etest

import (
	"context"
	"time"

	"github.com/cenkalti/backoff/v4"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

type waitInvoker struct {
	prev tg.Invoker
}

func (w waitInvoker) InvokeRaw(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	return backoff.Retry(func() error {
		if err := w.prev.InvokeRaw(ctx, input, output); err != nil {
			if timeout, ok := telegram.AsFloodWait(err); ok {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(timeout + 1*time.Second):
					return err
				}
			}

			return backoff.Permanent(err)
		}

		return nil
	}, backoff.WithContext(backoff.NewExponentialBackOff(), ctx))
}
