// Binary bot-upload implements upload example for bot.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"

	"go.uber.org/zap"

	"github.com/nnqq/td/examples"
	"github.com/nnqq/td/telegram"
	"github.com/nnqq/td/telegram/message"
	"github.com/nnqq/td/telegram/message/html"
	"github.com/nnqq/td/telegram/uploader"
	"github.com/nnqq/td/tg"
)

func main() {
	// Environment variables:
	//	BOT_TOKEN:     token from BotFather
	// 	APP_ID:        app_id of Telegram app.
	// 	APP_HASH:      app_hash of Telegram app.
	// 	SESSION_FILE:  path to session file
	// 	SESSION_DIR:   path to session directory, if SESSION_FILE is not set
	filePath := flag.String("file", "", "file to upload")
	targetDomain := flag.String("target", "", "target to upload, e.g. channel name")
	flag.Parse()

	examples.Run(func(ctx context.Context, log *zap.Logger) error {
		if *filePath == "" || *targetDomain == "" {
			return errors.New("no --file or --target provided")
		}

		// The performUpload will be called after client initialization.
		performUpload := func(ctx context.Context, client *telegram.Client) error {
			// Raw MTProto API client, allows making raw RPC calls.
			api := tg.NewClient(client)

			// Helper for uploading. Automatically uses big file upload when needed.
			u := uploader.NewUploader(api)

			// Helper for sending messages.
			sender := message.NewSender(api).WithUploader(u)

			// Uploading directly from path. Note that you can do it from
			// io.Reader or buffer, see From* methods of uploader.
			log.Info("Uploading file")
			upload, err := u.FromPath(ctx, *filePath)
			if err != nil {
				return fmt.Errorf("upload %q: %w", *filePath, err)
			}

			// Now we have uploaded file handle, sending it as styled message.
			// First, preparing message.
			document := message.UploadedDocument(upload,
				html.String(nil, `Upload: <b>From bot</b>`),
			)

			// You can set MIME type, send file as video or audio by using
			// document builder:
			document.
				MIME("audio/mp3").
				Filename("some-audio.mp3").
				Audio()

			// Resolving target. Can be telephone number or @nickname of user,
			// group or channel.
			target := sender.Resolve(*targetDomain)

			// Sending message with media.
			log.Info("Sending file")
			if _, err := target.Media(ctx, document); err != nil {
				return fmt.Errorf("send: %w", err)
			}

			return nil
		}
		return telegram.BotFromEnvironment(ctx, telegram.Options{
			Logger:    log,
			NoUpdates: true, // don't subscribe to updates in one-shot mode
		}, nil, performUpload)
	})
}
