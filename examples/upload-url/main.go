// Binary upload-url uploads a file fetched from an HTTP URL to Saved Messages,
// using the uploader/source helper as the remote source.
//
// uploader.FromSource streams the remote file straight into Telegram without
// buffering it on disk. The source.Source interface lets you customize how the
// bytes are fetched; here we use source.HTTPSource with a custom timeout.
package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-faster/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/gotd/log/logzap"

	"github.com/gotd/td/examples"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/html"
	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/telegram/uploader/source"
)

func run(ctx context.Context) error {
	rawURL := flag.String("url", "", "URL of the file to upload")
	flag.Parse()
	if *rawURL == "" {
		return errors.New("no --url provided")
	}

	log, _ := zap.NewDevelopment(zap.IncreaseLevel(zapcore.InfoLevel))
	defer func() { _ = log.Sync() }()

	// Initializing client from environment.
	// Available environment variables:
	// 	APP_ID:         app_id of Telegram app.
	// 	APP_HASH:       app_hash of Telegram app.
	// 	SESSION_FILE:   path to session file
	// 	SESSION_DIR:    path to session directory, if SESSION_FILE is not set
	client, err := telegram.ClientFromEnvironment(telegram.Options{
		Logger: logzap.New(log),
	})
	if err != nil {
		return err
	}

	// Reading phone, code and 2FA password from terminal when no session.
	flow := auth.NewFlow(examples.Terminal{}, auth.SendCodeOptions{})

	return client.Run(ctx, func(ctx context.Context) error {
		if err := client.Auth().IfNecessary(ctx, flow); err != nil {
			return errors.Wrap(err, "auth")
		}
		api := client.API()

		// Custom remote source: an HTTP fetcher with a 1-minute timeout.
		src := source.NewHTTPSource().WithClient(&http.Client{
			Timeout: time.Minute,
		})

		// Stream the remote file directly into Telegram.
		log.Info("Uploading from URL", zap.String("url", *rawURL))
		u := uploader.NewUploader(api)
		f, err := u.FromSource(ctx, src, *rawURL)
		if err != nil {
			return errors.Wrap(err, "upload from source")
		}

		// Send the uploaded file as a document to Saved Messages.
		sender := message.NewSender(api).WithUploader(u)
		doc := message.UploadedDocument(f, html.String(nil, "Uploaded from <b>URL</b>"))
		if _, err := sender.Self().Media(ctx, doc); err != nil {
			return errors.Wrap(err, "send")
		}
		log.Info("Sent to Saved Messages")
		return nil
	})
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	if err := run(ctx); err != nil {
		panic(err)
	}
}
