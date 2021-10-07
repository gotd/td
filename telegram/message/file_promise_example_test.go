package message_test

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"golang.org/x/xerrors"

	"github.com/nnqq/td/telegram"
	"github.com/nnqq/td/telegram/message"
	"github.com/nnqq/td/tg"
)

func filePromiseResult(ctx context.Context) error {
	client, err := telegram.ClientFromEnvironment(telegram.Options{})
	if err != nil {
		return err
	}

	return client.Run(ctx, func(ctx context.Context) error {
		sender := message.NewSender(tg.NewClient(client))
		r := sender.Resolve("@durov")

		var result tg.InputFileClass
		_, err := r.Upload(message.Upload(func(ctx context.Context, b message.Uploader) (tg.InputFileClass, error) {
			r, err := b.FromPath(ctx, "file.jpg")
			if err != nil {
				return nil, err
			}

			result = r
			return r, nil
		})).Photo(ctx)
		if err != nil {
			return xerrors.Errorf("upload photo: %w", err)
		}

		_, err = r.Media(ctx, message.UploadedDocument(result))
		if err != nil {
			return xerrors.Errorf("upload document: %w", err)
		}

		return nil
	})
}

func ExampleUpload() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := filePromiseResult(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(2)
	}
}
