package message_test

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"golang.org/x/xerrors"

	"github.com/nnqq/td/telegram"
	"github.com/nnqq/td/telegram/message"
	"github.com/nnqq/td/telegram/uploader"
	"github.com/nnqq/td/tg"
)

func sendAudio(ctx context.Context) error {
	client, err := telegram.ClientFromEnvironment(telegram.Options{})
	if err != nil {
		return err
	}

	return client.Run(ctx, func(ctx context.Context) error {
		raw := tg.NewClient(client)
		// Upload file.
		f, err := uploader.NewUploader(raw).FromPath(ctx, "vsyo idyot po planu.mp3")
		if err != nil {
			return xerrors.Errorf("upload: %w", err)
		}

		sender := message.NewSender(raw)
		r := sender.Resolve("@durovschat")

		// Sends audio to the @durovschat.
		if _, err := r.Audio(ctx, f); err != nil {
			return err
		}

		// Sends audio with title to the @durovschat.
		if _, err := r.Media(ctx, message.Audio(f).
			Performer("Yegor Letov").
			Title("Everything is going according to plan")); err != nil {
			return err
		}

		// Sends voice message to the @durovschat.
		if _, err := r.Voice(ctx, f); err != nil {
			return err
		}

		return nil
	})
}

func ExampleAudio() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := sendAudio(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(2)
	}
}
