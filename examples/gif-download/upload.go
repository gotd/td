package main

import (
	"context"
	"os"
	"path"
	"path/filepath"

	"github.com/ogen-go/errors"
	"go.uber.org/zap"

	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/unpack"
	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"
)

// upload lists inputDir and uploads all ".mp4" files to saved gifs.
//
// NB: Uses "Saved Messages" as temporary place for uploads.
func upload(ctx context.Context, log *zap.Logger, api *tg.Client, inputDir string) error {
	// Upload all gifs from requested dir.
	entries, err := os.ReadDir(inputDir)
	if err != nil {
		return errors.Wrap(err, "dir")
	}

	var names []string
	for _, e := range entries {
		if path.Ext(e.Name()) != ".mp4" {
			continue
		}
		names = append(names, filepath.Join(inputDir, e.Name()))
	}
	log.Info("Uploading all gifs from directory",
		zap.String("path", inputDir),
		zap.Int("count", len(names)),
	)

	u := uploader.NewUploader(api)
	for _, name := range names {
		f, err := u.FromPath(ctx, name)
		if err != nil {
			return err
		}

		// Using "Saved messages" as upload buffer, because we can't directly
		// upload gifs to "saved gifs".
		sender := message.NewSender(api).Self()

		// To be valid, media should have "animated" attribute and video/mp4
		// MIME-type.
		msg, err := unpack.Message(sender.Media(ctx, message.UploadedDocument(f).
			Attributes(&tg.DocumentAttributeAnimated{}).
			MIME("video/mp4"),
		))
		if err != nil {
			return err
		}
		doc, ok := msg.Media.(*tg.MessageMediaDocument).Document.AsNotEmpty()
		if !ok {
			return errors.New("unexpected document")
		}

		// Actually saving GIF.
		_, saveErr := api.MessagesSaveGif(ctx, &tg.MessagesSaveGifRequest{
			ID:     doc.AsInput(),
			Unsave: false,
		})
		// Cleaning up "buffer" message.
		if _, deleteErr := sender.Revoke().Messages(ctx, msg.ID); deleteErr != nil {
			return errors.Wrap(deleteErr, "delete")
		}
		// Checking for actual save error.
		if saveErr != nil {
			return errors.Wrap(saveErr, "save")
		}
		log.Info("Saved", zap.String("name", name))
	}

	return nil
}
