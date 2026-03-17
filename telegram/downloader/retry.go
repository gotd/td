package downloader

import (
	"context"

	"github.com/go-faster/errors"

	"github.com/gotd/td/exchange"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

const maxRetryAttempts = 20

func retryLimitErr(op string, attempts int, err error) error {
	return errors.Wrapf(err, "%s: retry limit reached (%d)", op, attempts)
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
			if tgerr.Is(waitErr, tg.ErrTimeout) {
				if ctxErr := ctx.Err(); ctxErr != nil {
					return zero, ctxErr
				}
				// Timeout can happen on unstable proxy links; retry with bounded
				// attempts to avoid infinite tight loops.
				timeoutRetries++
				if timeoutRetries >= maxRetryAttempts {
					return zero, retryLimitErr(op, timeoutRetries, waitErr)
				}
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
