package uploader

import (
	"bytes"
	"context"
	"crypto/rand"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"strconv"
	"testing"
	"testing/iotest"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
	"golang.org/x/xerrors"

	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

type progressLogger struct {
	logger *zap.Logger
}

func (p progressLogger) Chunk(ctx context.Context, state ProgressState) error {
	p.logger.Sugar().Infof("Part uploaded %+v", state)
	return nil
}

type Image func() *image.RGBA

func testProfilePhotoUploader(gen Image) func(t *testing.T) {
	return func(t *testing.T) {
		a := require.New(t)
		logger := zaptest.NewLogger(t, zaptest.Level(zapcore.WarnLevel))

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		err := telegram.TestClient(ctx, telegram.Options{
			Logger: logger,
		}, func(ctx context.Context, client *telegram.Client) error {
			if _, err := client.Self(ctx); err != nil {
				return xerrors.Errorf("self: %w", err)
			}

			invoker, err := client.Pool(0 /* unlimited */)
			if err != nil {
				return xerrors.Errorf("pool: %w", err)
			}

			img := bytes.NewBuffer(nil)
			if err := png.Encode(img, gen()); err != nil {
				return xerrors.Errorf("png encode: %w", err)
			}
			t.Log("size of image", img.Len(), "bytes")

			raw := tg.NewClient(invoker)
			f, err := NewUploader(raw).
				WithPartSize(2048).
				WithProgress(progressLogger{logger.Named("progress")}).
				FromReader(ctx, "abc.jpg", iotest.HalfReader(img))
			if err != nil {
				return xerrors.Errorf("upload: %w", err)
			}

			req := &tg.PhotosUploadProfilePhotoRequest{}
			req.SetFile(f)
			res, err := raw.PhotosUploadProfilePhoto(ctx, req)
			if err != nil {
				return xerrors.Errorf("change profile photo: %w", err)
			}

			_, ok := res.Photo.(*tg.Photo)
			a.Truef(ok, "unexpected type %T", res.Photo)
			return nil
		})

		a.NoError(err)
	}
}

func generateImage(x, y int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, x, y))
	for x := 0; x < img.Bounds().Dx(); x++ {
		for y := 0; y < img.Bounds().Dy(); y++ {
			if (x+y)%2 == 0 || (x%2 != 0 && y%2 != 0) {
				img.SetRGBA(x, y, color.RGBA{
					R: 255,
					G: 255,
					B: 255,
					A: 255,
				})
			}
		}
	}
	return img
}

type generator struct {
	state int64
}

func (g *generator) Read(p []byte) (n int, err error) {
	if g.state <= 0 {
		return 0, io.EOF
	}

	for i := range p {
		p[i] = byte(g.state)
		g.state--
	}

	return len(p), nil
}

func TestExternalE2EDocUpload(t *testing.T) {
	if ok, _ := strconv.ParseBool(os.Getenv("GOTD_TEST_EXTERNAL")); !ok {
		t.Skip("Skipped. Set GOTD_TEST_EXTERNAL=1 to enable external e2e test.")
	}

	a := require.New(t)
	logger := zaptest.NewLogger(t, zaptest.Level(zapcore.InfoLevel))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	err := telegram.TestClient(ctx, telegram.Options{
		Logger: logger,
	}, func(ctx context.Context, client *telegram.Client) error {
		if _, err := client.Self(ctx); err != nil {
			return xerrors.Errorf("self: %w", err)
		}

		invoker, err := client.Pool(2)
		if err != nil {
			return xerrors.Errorf("pool: %w", err)
		}

		total := int64(20 * 1024 * 1024)
		g := &generator{state: total}

		raw := tg.NewClient(invoker)
		upld := NewUpload("devrandom.dat", iotest.HalfReader(g), total)
		f, err := NewUploader(raw).
			WithPartSize(MaximumPartSize).
			WithProgress(progressLogger{logger.Named("progress")}).
			Upload(ctx, upld)
		if err != nil {
			return xerrors.Errorf("upload: %w", err)
		}

		id, err := crypto.RandInt64(rand.Reader)
		if err != nil {
			return xerrors.Errorf("message id: %w", err)
		}

		_, err = raw.MessagesSendMedia(ctx, &tg.MessagesSendMediaRequest{
			Peer: &tg.InputPeerSelf{},
			Media: &tg.InputMediaUploadedDocument{
				File: f,
			},
			RandomID: id,
		})
		if err != nil {
			return xerrors.Errorf("send media: %w", err)
		}

		return nil
	})
	a.NoError(err)
}

func TestExternalE2EVideoUpload(t *testing.T) {
	if ok, _ := strconv.ParseBool(os.Getenv("GOTD_TEST_EXTERNAL")); !ok {
		t.Skip("Skipped. Set GOTD_TEST_EXTERNAL=1 to enable external e2e test.")
	}

	a := require.New(t)
	logger := zaptest.NewLogger(t, zaptest.Level(zapcore.InfoLevel))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	err := telegram.TestClient(ctx, telegram.Options{
		Logger: logger,
	}, func(ctx context.Context, client *telegram.Client) error {
		if _, err := client.Self(ctx); err != nil {
			return xerrors.Errorf("self: %w", err)
		}

		invoker, err := client.Pool(2)
		if err != nil {
			return xerrors.Errorf("pool: %w", err)
		}

		raw := tg.NewClient(invoker)
		f, err := NewUploader(raw).
			WithPartSize(MaximumPartSize).
			WithProgress(progressLogger{logger.Named("progress")}).
			FromPath(ctx, "./testdata/video.mp4")
		if err != nil {
			return xerrors.Errorf("upload: %w", err)
		}

		id, err := crypto.RandInt64(rand.Reader)
		if err != nil {
			return xerrors.Errorf("message id: %w", err)
		}

		_, err = raw.MessagesSendMedia(ctx, &tg.MessagesSendMediaRequest{
			Peer: &tg.InputPeerSelf{},
			Media: &tg.InputMediaUploadedDocument{
				File:     f,
				MimeType: "video/mp4",
				Attributes: []tg.DocumentAttributeClass{
					&tg.DocumentAttributeVideo{},
				},
			},
			RandomID: id,
		})
		if err != nil {
			return xerrors.Errorf("send media: %w", err)
		}

		return nil
	})
	a.NoError(err)
}

func TestExternalE2EProfilePhotoUpload(t *testing.T) {
	if ok, _ := strconv.ParseBool(os.Getenv("GOTD_TEST_EXTERNAL")); !ok {
		t.Skip("Skipped. Set GOTD_TEST_EXTERNAL=1 to enable external e2e test.")
	}

	t.Run("LessThanPart", testProfilePhotoUploader(func() *image.RGBA {
		return generateImage(255, 255)
	}))

	t.Run("BiggerThanPart", testProfilePhotoUploader(func() *image.RGBA {
		return generateImage(1024, 1024)
	}))
}
