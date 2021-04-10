package telegram

import (
	"context"
	"os"

	"golang.org/x/xerrors"
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
		return xerrors.Errorf("create client: %w", err)
	}

	if err := setup(ctx, client); err != nil {
		return xerrors.Errorf("setup: %w", err)
	}

	return client.Run(ctx, func(ctx context.Context) error {
		status, err := client.AuthStatus(ctx)
		if err != nil {
			return xerrors.Errorf("auth status: %w", err)
		}

		if !status.Authorized {
			if _, err := client.AuthBot(ctx, os.Getenv("BOT_TOKEN")); err != nil {
				return xerrors.Errorf("login: %w", err)
			}
		}

		return cb(ctx, client)
	})
}
