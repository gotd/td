package downloader

import (
	"context"
	"net"

	"github.com/go-faster/errors"

	"github.com/gotd/td/exchange"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

const maxRetryAttempts = 20

func retryLimitErr(ctx context.Context, op string, attempts int, err error) error {
	err = errors.Wrapf(err, "%s: retry limit reached (%d)", op, attempts)
	observerFrom(ctx).exhausted(ctx, op, err)
	return err
}

func isCDNFingerprintErr(err error) bool {
	return errors.Is(err, exchange.ErrKeyFingerprintNotFound)
}

func isCDNMasterFallbackErr(err error) bool {
	// Token invalidation requires fetching fresh redirect/token window from
	// master DC.
	return tgerr.Is(
		err,
		"FILE_TOKEN_INVALID",
		"REQUEST_TOKEN_INVALID",
	)
}

func isRetryableTimeout(ctx context.Context, err error) bool {
	if err == nil {
		return false
	}
	if ctx.Err() != nil {
		return false
	}
	if tgerr.Is(err, tg.ErrTimeout) || errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}

func retryRequest[T any](
	ctx context.Context,
	op string,
	onRetry func(attempt int, err error),
	fn func() (T, error),
) (_ T, err error) {
	var zero T
	timeoutRetries := 0
	retryAttempt := 0
	for {
		if err := ctx.Err(); err != nil {
			return zero, err
		}

		result, err := fn()
		if _, ok := tgerr.AsFloodWait(err); ok {
			// Let Scheduler of the request shrink its window before sleeping,
			// as the request is retried here while still holding its slot.
			reportBackoff(ctx, err)
		}
		if flood, waitErr := tgerr.FloodWait(ctx, err); waitErr != nil {
			if flood {
				// FloodWait helper already slept required amount.
				if ctxErr := ctx.Err(); ctxErr != nil {
					return zero, ctxErr
				}
				retryAttempt++
				if onRetry != nil {
					onRetry(retryAttempt, waitErr)
				}
				continue
			}
			if isRetryableTimeout(ctx, waitErr) {
				// Timeout can happen on unstable proxy links; retry with bounded
				// attempts to avoid infinite tight loops.
				timeoutRetries++
				if timeoutRetries >= maxRetryAttempts {
					// Not retried: the error is reported by release of the
					// request, so it is not reported here.
					return zero, retryLimitErr(ctx, op, timeoutRetries, waitErr)
				}
				reportBackoff(ctx, waitErr)
				retryAttempt++
				if onRetry != nil {
					onRetry(retryAttempt, waitErr)
				}
				continue
			}
			return zero, waitErr
		}

		return result, nil
	}
}
