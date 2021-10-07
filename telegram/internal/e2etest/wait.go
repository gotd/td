package e2etest

import (
	"context"

	"github.com/cenkalti/backoff/v4"

	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/tg"
	"github.com/nnqq/td/tgerr"
)

type waitInvoker struct {
	prev tg.Invoker
}

func retryFloodWait(ctx context.Context, cb func() error) error {
	return backoff.Retry(func() error {
		if err := cb(); err != nil {
			if ok, err := tgerr.FloodWait(ctx, err); ok {
				return err
			}

			return backoff.Permanent(err)
		}

		return nil
	}, backoff.WithContext(backoff.NewExponentialBackOff(), ctx))
}

func (w waitInvoker) Invoke(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	return retryFloodWait(ctx, func() error {
		return w.prev.Invoke(ctx, input, output)
	})
}
