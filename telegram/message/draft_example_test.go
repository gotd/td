package message_test

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"golang.org/x/xerrors"

	"github.com/nnqq/td/telegram"
	"github.com/nnqq/td/telegram/message"
	"github.com/nnqq/td/telegram/message/styling"
	"github.com/nnqq/td/tg"
)

func saveDraft(ctx context.Context) error {
	client, err := telegram.ClientFromEnvironment(telegram.Options{})
	if err != nil {
		return err
	}

	return client.Run(ctx, func(ctx context.Context) error {
		sender := message.NewSender(tg.NewClient(client))
		r := sender.Resolve("@durov")

		// Save draft message.
		if err := r.SaveDraft(ctx, "Hi!"); err != nil {
			return xerrors.Errorf("draft: %w", err)
		}

		// Save styled draft message.
		if err := r.SaveStyledDraft(ctx, styling.Bold("Hi!")); err != nil {
			return xerrors.Errorf("draft: %w", err)
		}

		// Clear draft for resolved @durov peer.
		if err := r.ClearDraft(ctx); err != nil {
			return xerrors.Errorf("draft: %w", err)
		}

		return nil
	})
}

func ExampleBuilder_SaveDraft() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := saveDraft(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(2)
	}
}
