package e2etest

import (
	"context"
	"time"

	"github.com/cenkalti/backoff/v4"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

type retryInvoker struct {
	prev tg.Invoker
}

func retryResult[T any](ctx context.Context, cb func() (T, error)) (T, error) {
	var zero T
	return backoff.RetryWithData[T](func() (T, error) {
		res, err := cb()
		if err != nil {
			if tgerr.IsCode(err, -500) {
				return zero, err
			}
			if tgerr.Is(err, "CONNECTION_NOT_INITED") {
				return zero, err
			}
			if ok, err := tgerr.FloodWait(ctx, err); ok {
				return zero, err
			}
			return zero, backoff.Permanent(err)
		}
		return res, nil
	}, backoff.WithContext(backoff.NewConstantBackOff(time.Millisecond*500), ctx))
}

func retry(ctx context.Context, cb func() error) error {
	return backoff.Retry(func() error {
		if err := cb(); err != nil {
			if tgerr.IsCode(err, -500) {
				return err
			}
			if tgerr.Is(err, "CONNECTION_NOT_INITED") {
				return err
			}
			if ok, err := tgerr.FloodWait(ctx, err); ok {
				return err
			}
			return backoff.Permanent(err)
		}

		return nil
	}, backoff.WithContext(backoff.NewExponentialBackOff(), ctx))
}

func (w retryInvoker) Invoke(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	return retry(ctx, func() error {
		return w.prev.Invoke(ctx, input, output)
	})
}
