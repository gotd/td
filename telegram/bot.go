package telegram

import (
	"context"
	"os"

	"github.com/go-faster/errors"
)

// RunUntilCanceled is client callback which
// locks until client context is canceled.
func RunUntilCanceled(ctx context.Context, client *Client) error {
	<-ctx.Done()
	return ctx.Err()
}

// BotFromEnvironment creates bot client using ClientFromEnvironment
// connects to server and authenticates it.
//
// Variables:
// BOT_TOKEN — token from BotFather.
func BotFromEnvironment(
	ctx context.Context,
	opts Options,
	setup func(ctx context.Context, client *Client) error,
	cb func(ctx context.Context, client *Client) error,
) error {
	client, err := ClientFromEnvironment(opts)
	if err != nil {
		return errors.Wrap(err, "create client")
	}

	if setup != nil {
		if err := setup(ctx, client); err != nil {
			return errors.Wrap(err, "setup")
		}
	}

	return client.Run(ctx, func(ctx context.Context) error {
		status, err := client.Auth().Status(ctx)
		if err != nil {
			return errors.Wrap(err, "auth status")
		}

		if !status.Authorized {
			if _, err := client.Auth().Bot(ctx, os.Getenv("BOT_TOKEN")); err != nil {
				return errors.Wrap(err, "login")
			}
		}

		return cb(ctx, client)
	})
}
