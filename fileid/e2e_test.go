package fileid_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/go-faster/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"github.com/gotd/td/fileid"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/testutil"
)

func runBot(ctx context.Context, token, fileID string, logger *zap.Logger) error {
	bot := telegram.NewClient(telegram.TestAppID, telegram.TestAppHash, telegram.Options{
		Logger: logger,
	})
	d := downloader.NewDownloader()

	return bot.Run(ctx, func(ctx context.Context) error {
		auth, err := bot.Auth().Bot(ctx, token)
		if err != nil {
			return errors.Wrap(err, "auth bot")
		}
		user, ok := auth.User.AsNotEmpty()
		if !ok {
			return errors.Errorf("unexpected type %T", auth.User)
		}
		_ = user

		decoded, err := fileid.DecodeFileID(fileID)
		if err != nil {
			return errors.Wrap(err, "decode FileID")
		}

		loc, ok := decoded.AsInputFileLocation()
		if !ok {
			return errors.Errorf("can't map %q", fileID)
		}

		filename := "file.dat"
		switch decoded.Type {
		case fileid.Thumbnail, fileid.ProfilePhoto, fileid.Photo:
			filename = "file.jpg"
		case fileid.Video,
			fileid.Animation,
			fileid.VideoNote:
			filename = "file.mp4"
		case fileid.Audio:
			filename = "file.mp3"
		case fileid.Voice:
			filename = "file.ogg"
		case fileid.Sticker:
			filename = "file.png"
		}

		if _, err := d.Download(bot.API(), loc).ToPath(ctx, filename); err != nil {
			return errors.Wrap(err, "download")
		}
		return nil
	})
}

func TestExternalE2ECheckFileID(t *testing.T) {
	testutil.SkipExternal(t)
	token := os.Getenv("GOTD_E2E_BOT_TOKEN")
	if token == "" {
		t.Skip("Set GOTD_E2E_BOT_TOKEN env to run test.")
	}
	fileID := os.Getenv("GOTD_E2E_FILE_ID")
	if fileID == "" {
		t.Skip("Set GOTD_E2E_FILE_ID env to run test.")
	}
	logger := zaptest.NewLogger(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	if err := runBot(ctx, token, fileID, logger.Named("bot")); err != nil {
		t.Error(err)
	}
}
